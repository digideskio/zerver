package zerver

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"net/url"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cosiner/gohper/lib/errors"
	"github.com/cosiner/gohper/lib/types"
	websocket "github.com/cosiner/zerver_websocket"
)

var (
	Bytes  = types.UnsafeBytes
	String = types.UnsafeString
)

const (
	ErrComponentNotFound = errors.Err("The required component is not found")
	_NORMAL              = 0
	_DESTROYED           = 1
)

type (
	ServerOption struct {
		// server listening address, default :4000
		ListenAddr string
		// content type for each request, default application/json;charset=utf-8
		ContentType string

		// check websocket header, default nil
		WebSocketChecker HeaderChecker
		// error logger, default use log.Println
		ErrorLogger func(...interface{})
		// resource marshal/pool/unmarshal
		// first search by Server.Component, if not found, use JSONResource
		ResourceMaster

		// path variables count, suggest set as max or average, default 3
		PathVarCount int
		// filters count for each route, RootFilters is not include, default 5
		FilterCount int

		// read timeout by millseconds
		ReadTimeout int
		// write timeout by millseconds
		WriteTimeout int
		// max header bytes
		MaxHeaderBytes int
		// tcp keep-alive period by minutes,
		// default 3, same as predefined in standard http package
		KeepAlivePeriod int
		// ssl config, default disable tls
		CertFile, KeyFile string
	}

	ComponentState struct {
		Initialized bool
		NoLazy      bool
		Component
	}

	// Server represent a web server
	Server struct {
		Router
		AttrContainer
		RootFilters RootFilters // Match Every Routes
		Errorln     func(...interface{})

		components        map[string]ComponentState
		managedComponents []Component
		sync.RWMutex

		checker     websocket.HandshakeChecker
		contentType string // default content type
		resMaster   ResourceMaster

		listener    net.Listener
		state       int32           // destroy or normal running
		activeConns *sync.WaitGroup // connections in service, don't include hijacked and websocket connections
	}

	// HeaderChecker is a http header checker, it accept a function which can get
	// http header's value by name , if there is something wrong, throw an error
	// to terminate this request
	HeaderChecker func(func(string) string) error

	// Component is a Object which will automaticlly initial/destroyed by server
	// if it's added to server, else it should initial manually
	Component interface {
		Init(Enviroment) error
		Destroy()
	}

	// Enviroment is a server enviroment, real implementation is the Server itself.
	// it can be accessed from Request/WebsocketConn
	Enviroment interface {
		Server() *Server
		StartTask(path string, value interface{})
		Component(name string) (Component, error)
		ResourceMaster() ResourceMaster
	}
)

func (o *ServerOption) init() {
	if o.ListenAddr == "" {
		o.ListenAddr = ":4000"
	}
	if o.ContentType == "" {
		o.ContentType = CONTENTTYPE_JSON
	}
	if o.PathVarCount == 0 {
		o.PathVarCount = 3
	}
	if o.FilterCount == 0 {
		o.FilterCount = 5
	}
	if o.ErrorLogger == nil {
		o.ErrorLogger = log.Println
	}
	if o.KeepAlivePeriod == 0 {
		o.KeepAlivePeriod = 3 // same as net/http/server.go:tcpKeepAliveListener
	}
}

// NewServer create a new server with default router
func NewServer() *Server {
	return NewServerWith(nil, nil)
}

// NewServerWith create a new server with given router and root filters
func NewServerWith(rt Router, filters RootFilters) *Server {
	if filters == nil {
		filters = NewRootFilters(nil)
	}
	if rt == nil {
		rt = NewRouter()
	}
	return &Server{
		Router:        rt,
		AttrContainer: NewLockedAttrContainer(),
		RootFilters:   filters,
		components:    make(map[string]ComponentState),
		activeConns:   new(sync.WaitGroup),
	}
}

// ent ServerEnviroment
func (s *Server) Server() *Server {
	return s
}

func (s *Server) ResourceMaster() ResourceMaster {
	return s.resMaster
}

func (s *Server) Component(name string) (Component, error) {
	s.RLock()
	if c, has := s.components[name]; has {
		s.RUnlock()
		var err error
		if !c.Initialized {
			s.Lock()
			if !c.Initialized {
				if err = c.Component.Init(s); err == nil {
					c.Initialized = true
				}
			}
			s.Unlock()
		}
		return c.Component, err
	}
	s.RUnlock()
	return nil, ErrComponentNotFound
}

func (s *Server) AddComponent(name string, c ComponentState) error {
	if name != "" && c.Component != nil {
		if !c.Initialized && c.NoLazy {
			if err := c.Init(s); err != nil {
				return err
			}
		}
		s.Lock()
		s.components[name] = c
		s.Unlock()
		return nil
	}
	panic("empty name or nil component is not allowed")
}

func (s *Server) RemoveComponent(name string) {
	if name != "" {
		s.Lock()
		if cs, has := s.components[name]; has {
			if cs.Initialized {
				defer s.Unlock()
				cs.Destroy()
				delete(s.components, name)
				return
			}
		}
		delete(s.components, name)
		s.Unlock()
	}
}

// ManageComponent manage those filters used in InterceptHandler, or those added to
// multiple routes, for first condition, ManageComponent used to Init them;
// for second condition, ManageComponent used to avoid multiple call of Init
func (s *Server) ManageComponent(c Component) {
	s.managedComponents = append(s.managedComponents, c)
}

// StartTask start a task synchronously, the value will be passed to task handler
func (s *Server) StartTask(path string, value interface{}) {
	if handler := s.MatchTaskHandler(&url.URL{Path: path}); handler != nil {
		handler.Handle(value)
		return
	}
	panic("No task handler found for " + path)
}

// ServHttp serve for http reuest
// find handler and resolve path, find filters, process
func (s *Server) ServeHTTP(w http.ResponseWriter, request *http.Request) {
	path := request.URL.Path
	if l := len(path); l > 1 && path[l-1] == '/' {
		request.URL.Path = path[:l-1]
	}
	if websocket.IsWebSocketRequest(request) {
		s.serveWebSocket(w, request)
	} else {
		s.serveHTTP(w, request)
	}
}

// serveWebSocket serve for websocket protocal
func (s *Server) serveWebSocket(w http.ResponseWriter, request *http.Request) {
	handler, indexer := s.MatchWebSocketHandler(request.URL)
	if handler == nil {
		w.WriteHeader(http.StatusNotFound)
	} else if conn, err := websocket.UpgradeWebsocket(w, request, s.checker); err == nil {
		handler.Handle(newWebSocketConn(s, conn, indexer))
		indexer.destroySelf()
	} // else connecion will be auto-closed when error occoured,
}

// serveHTTP serve for http protocal
func (s *Server) serveHTTP(w http.ResponseWriter, request *http.Request) {
	url := request.URL
	url.Host = request.Host
	handler, indexer, filters := s.MatchHandlerFilters(url)
	requestEnv := newRequestEnvFromPool()
	req := requestEnv.req.init(s, request, indexer)
	resp := requestEnv.resp.init(s.resMaster, w)
	resp.SetContentType(s.contentType)

	var chain FilterChain
	if handler == nil {
		resp.ReportNotFound()
	} else if chain = FilterChain(handler.Handler(req.Method())); chain == nil {
		resp.ReportMethodNotAllowed()
	}

	newFilterChain(s.RootFilters.Filters(url),
		newFilterChain(filters, chain),
	)(req, resp)

	req.destroy()
	resp.destroy()
	recycleRequestEnv(requestEnv)
	recycleFilters(filters)
}

// OnErrorLog help log error
func (s *Server) OnErrorLog(err error) {
	if err != nil {
		s.Errorln(err)
	}
}

// from net/http/server/go
type tcpKeepAliveListener struct {
	*net.TCPListener
	AlivePeriod int // alive period by minutes
}

func (ln *tcpKeepAliveListener) Accept() (c net.Conn, err error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(time.Duration(ln.AlivePeriod) * time.Minute)
	return tc, nil
}

func (s *Server) config(o *ServerOption) {
	o.init()
	s.Errorln = o.ErrorLogger
	log.Println("ContentType:", o.ContentType)
	s.contentType = o.ContentType
	s.checker = websocket.HeaderChecker(o.WebSocketChecker).HandshakeCheck
	s.resMaster = o.ResourceMaster
	if s.resMaster == nil {
		c, err := s.Component(COMP_RESOURCE)
		if err == nil {
			s.resMaster = c.(ResourceMaster)
			log.Println("ResourceMaster: customed")
		} else if err == ErrComponentNotFound {
			s.resMaster = JSONResource{}
			log.Println("ResourceMaster: default JSONResource")
		} else {
			panic(err)
		}
	}
	log.Println("VarCountPerRoute:", o.PathVarCount)
	pathVarCount = o.PathVarCount
	log.Println("FilterCountPerRoute:", o.FilterCount)
	filterCount = o.FilterCount
	log.Println("Init managed components")
	for i := range s.managedComponents {
		errors.OnErrPanic(s.managedComponents[i].Init(s))
	}
	log.Println("Init root filters")
	errors.OnErrPanic(s.RootFilters.Init(s))
	log.Println("Init Handlers and Filters")
	errors.OnErrPanic(s.Router.Init(s))
	log.Println("Server Start:", o.ListenAddr)
	// destroy temporary data store
	tmpDestroy()
	runtime.GC()
}

// Start start server as http server
func (s *Server) Start(options *ServerOption) error {
	if options == nil {
		options = &ServerOption{}
	}
	s.config(options)
	l, err := s.listen(options)
	if err == nil {
		s.listener = l
		srv := &http.Server{
			ReadTimeout:  time.Duration(options.ReadTimeout) * time.Millisecond,
			WriteTimeout: time.Duration(options.WriteTimeout) * time.Millisecond,
			Handler:      s,
			ConnState:    s.connStateHook,
		}
		err = srv.Serve(l)
	}
	return err
}

func (*Server) listen(options *ServerOption) (net.Listener, error) {
	ln, err := net.Listen("tcp", options.ListenAddr)
	if err == nil {
		ln = &tcpKeepAliveListener{
			TCPListener: ln.(*net.TCPListener),
			AlivePeriod: options.KeepAlivePeriod,
		}

		if options.CertFile != "" {
			// from net/http/server.go.ListenAndServeTLS
			config := &tls.Config{
				NextProtos:   []string{"http/1.1"},
				Certificates: make([]tls.Certificate, 1),
			}
			config.Certificates[0], err = tls.LoadX509KeyPair(options.CertFile, options.KeyFile)
			if err == nil {
				ln = tls.NewListener(ln, config)
			}
		}
	}
	if err != nil && ln != nil {
		ln.Close()
		return nil, err
	}
	return ln, err
}

func (s *Server) connStateHook(conn net.Conn, state http.ConnState) {
	switch state {
	case http.StateActive:
		if atomic.LoadInt32(&s.state) == _NORMAL {
			s.activeConns.Add(1)
		} else {
			// previous idle connections before server destroy to be active, directly close it
			conn.Close()
		}
	case http.StateIdle:
		if atomic.LoadInt32(&s.state) == _DESTROYED {
			conn.Close()
		}
		s.activeConns.Done()
	case http.StateHijacked:
		s.activeConns.Done()
	}
}

// Destroy stop server, release all resources, if destroyed, server can't be reused,
// instead, create a new one.
// It only wait for managed connections, hijacked/websocket connections is not
func (s *Server) Destroy() {
	if atomic.CompareAndSwapInt32(&s.state, _NORMAL, _DESTROYED) { // signal close idle connections
		s.listener.Close()   // don't accept connections
		s.activeConns.Wait() // wait connections in service to be idle

		// release resources
		s.RootFilters.Destroy()
		s.Router.Destroy()
		for _, c := range s.components {
			c.Destroy()
		}
		for i := range s.managedComponents {
			s.managedComponents[i].Destroy()
		}
	}
}
