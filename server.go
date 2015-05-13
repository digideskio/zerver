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

	"github.com/cosiner/gohper/crypto/tls2"
	"github.com/cosiner/gohper/defval"
	"github.com/cosiner/gohper/errors"
	"github.com/cosiner/ygo/resource"
	websocket "github.com/cosiner/zerver_websocket"
)

const (
	ErrComponentNotFound = errors.Err("The required component is not found")
	// server status
	_NORMAL    = 0
	_DESTROYED = 1

	_CONTENTTYPE_DISABLE = "-"
)

type (
	ServerOption struct {
		// server listening address, default :4000
		ListenAddr string
		// content type for each request, default application/json;charset=utf-8,
		// use "-" to disable the automation
		ContentType string

		// check websocket header, default nil
		WebSocketChecker HeaderChecker
		// logger, default use cosiner/gohper/log.Logger with ConsoleWriter
		Logger

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

		// CA pem files to verify client certs
		CAs []string
		// ssl config, default disable tls
		CertFile, KeyFile string
		// if not nil, cert and key will be ignored
		TLSConfig *tls.Config
	}

	// Server represent a web server
	Server struct {
		Router
		AttrContainer
		RootFilters RootFilters // Match Every Routes
		ResMaster   resource.Master
		Log         Logger

		components        map[string]ComponentState
		managedComponents []Component
		sync.RWMutex

		checker     websocket.HandshakeChecker
		contentType string // default content type

		listener    net.Listener
		state       int32          // destroy or normal running
		activeConns sync.WaitGroup // connections in service, don't include hijacked and websocket connections
	}

	// HeaderChecker is a http header checker, it accept a function which can get
	// http header's value by name , if there is something wrong, throw an error
	// to terminate this request
	HeaderChecker func(func(string) string) error

	// Enviroment is a server enviroment, real implementation is the Server itself.
	// it can be accessed from Request/WebsocketConn
	Enviroment interface {
		Server() *Server
		ResourceMaster() *resource.Master
		Logger() Logger
		StartTask(path string, value interface{})
		Component(name string) (interface{}, error)
	}
)

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
		ResMaster:     resource.NewMaster(),
	}
}

func (s *Server) Server() *Server {
	return s
}

func (s *Server) Logger() Logger {
	return s.Log
}

func (s *Server) ResourceMaster() *resource.Master {
	return &s.ResMaster
}

func (s *Server) Component(name string) (interface{}, error) {
	s.RLock()
	c, has := s.components[name]
	if !has {
		s.RUnlock()
		return nil, ErrComponentNotFound
	}

	s.RUnlock()
	if c.value != nil {
		return c.value, nil
	}

	var err error
	if !c.Initialized {
		s.Lock()
		if !c.Initialized {
			if err = c.Component.(Component).Init(s); err == nil { // must be a Component
				c.Initialized = true
			}
		}
		s.Unlock()
	}
	return c.Component, err
}

// AddComponent let server manage this component and it's lifetime.
//
// If name is "", component must implements Component, and it will initialized at
// server start and can't be accessed by name.
//
// Otherwise, it can be a Component, ComponentState, or others.
// Component is treat as Not Initialized, and is not NoLazy.
// ComponentState use it's own way.
// Others treat as Initialized.
func (s *Server) AddComponent(name string, component interface{}) error {
	if name == "" {
		if c, is := component.(Component); is {
			s.managedComponents = append(s.managedComponents, c)
		}
		return nil
	}

	comp := convertComponentState(component)
	if !comp.Initialized && comp.NoLazy {
		if err := comp.Init(s); err != nil {
			return err
		} else {
			comp.Initialized = true
		}
	}

	s.Lock()
	s.components[name] = comp
	s.Unlock()
	return nil
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

// StartTask start a task synchronously, the value will be passed to task handler
func (s *Server) StartTask(path string, value interface{}) {
	if handler := s.MatchTaskHandler(&url.URL{Path: path}); handler != nil {
		handler.Handle(value)
		return
	}
	s.Log.Panicln("No task handler found for:", path)
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
	reqEnv := newRequestEnvFromPool()
	res := s.ResMaster.Resource(reqEnv.req.ContentType())
	req := reqEnv.req.init(s, res, request, indexer)
	resp := reqEnv.resp.init(s, res, w)

	if s.contentType != _CONTENTTYPE_DISABLE {
		resp.SetContentType(s.contentType)
	}

	var chain FilterChain
	if handler == nil {
		resp.ReportNotFound()
	} else if chain = FilterChain(handler.Handler(req.Method())); chain == nil {
		resp.ReportMethodNotAllowed()
	}

	newFilterChain(s.RootFilters.Filters(url),
		newFilterChain(filters, chain),
	)(req, resp)

	s.warnLog(req.destroy())
	s.warnLog(resp.destroy())

	recycleRequestEnv(reqEnv)
	recycleFilters(filters)
}

func (o *ServerOption) init() {
	defval.String(&o.ListenAddr, ":4000")
	defval.String(&o.ContentType, resource.CONTENTTYPE_JSON)
	defval.Int(&o.PathVarCount, 3)
	defval.Int(&o.FilterCount, 5)
	defval.Int(&o.KeepAlivePeriod, 3) // same as net/http/server.go:tcpKeepAliveListener
	if o.Logger == nil {
		o.Logger = DefaultLogger()
	}
}

// all log message before server start will use standard log package
func (s *Server) config(o *ServerOption) {
	o.init()
	var panicError = func(err error) {
		if err != nil {
			log.Panicln(err)
		}
	}

	s.Log = o.Logger

	log.Println("ContentType:", o.ContentType)
	s.contentType = o.ContentType
	s.checker = websocket.HeaderChecker(o.WebSocketChecker).HandshakeCheck

	if len(s.ResMaster.Resources) == 0 {
		s.ResMaster.DefUse(resource.RES_JSON, resource.JSON{})
	}

	log.Println("VarCountPerRoute:", o.PathVarCount)
	pathVarCount = o.PathVarCount
	log.Println("FilterCountPerRoute:", o.FilterCount)
	filterCount = o.FilterCount

	log.Println("Init managed components")
	for i := range s.managedComponents {
		panicError(s.managedComponents[i].Init(s))
	}

	log.Println("Init root filters")
	panicError(s.RootFilters.Init(s))
	log.Println("Init Handlers and Filters")
	panicError(s.Router.Init(s))

	// destroy temporary data store
	tmpDestroy()
	log.Println("Server Start:", o.ListenAddr)

	runtime.GC()
}

// PanicLog will panic goroutine, be care to call this and note to relase resource
// with 'defer'
func (s *Server) PanicLog(err error) {
	if err != nil {
		s.Log.Panicln(err)
	}
}

func (s *Server) warnLog(err error) {
	if err != nil {
		s.Log.Warnln(err)
	}
}

// Start start server as http server, if opt is nil, use default configurations
func (s *Server) Start(opt *ServerOption) error {
	if opt == nil {
		opt = &ServerOption{}
	}
	s.config(opt)
	l, err := s.listen(opt)
	if err == nil {
		s.listener = l
		srv := &http.Server{
			ReadTimeout:  time.Duration(opt.ReadTimeout) * time.Millisecond,
			WriteTimeout: time.Duration(opt.WriteTimeout) * time.Millisecond,
			Handler:      s,
			ConnState:    s.connStateHook,
		}
		err = srv.Serve(l)
	}
	return err
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

	// if keep-alive fail, don't care
	_ = tc.SetKeepAlive(true)
	_ = tc.SetKeepAlivePeriod(time.Duration(ln.AlivePeriod) * time.Minute)
	return tc, nil
}

func (s *Server) listen(opt *ServerOption) (net.Listener, error) {
	ln, err := net.Listen("tcp", opt.ListenAddr)
	if err == nil {
		ln = &tcpKeepAliveListener{
			TCPListener: ln.(*net.TCPListener),
			AlivePeriod: opt.KeepAlivePeriod,
		}

		if opt.TLSConfig != nil {
			ln = tls.NewListener(ln, opt.TLSConfig)
		} else if opt.CertFile != "" {
			// from net/http/server.go.ListenAndServeTLS
			// TODO: support verify client certs
			tc := &tls.Config{
				NextProtos:   []string{"http/1.1"},
				Certificates: make([]tls.Certificate, 1),
			}
			tc.Certificates[0], err = tls.LoadX509KeyPair(opt.CertFile, opt.KeyFile)
			if err == nil {
				if opt.CAs != nil {
					tc.ClientCAs, err = tls2.CAPool(opt.CAs...)
					if err == nil {
						tc.ClientAuth = tls.RequireAndVerifyClientCert
					}
				}
				if err == nil {
					ln = tls.NewListener(ln, tc)
				}
			}
		}
	}

	if err != nil && ln != nil {
		s.warnLog(ln.Close())
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
			// previous idle connections before call server.Destroy() becomes active, directly close it
			s.warnLog(conn.Close())
		}
	case http.StateIdle:
		if atomic.LoadInt32(&s.state) == _DESTROYED {
			s.warnLog(conn.Close())
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
		s.warnLog(s.listener.Close()) // don't accept connections
		s.activeConns.Wait()          // wait connections in service to be idle

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
