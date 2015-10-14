package zerver

import (
	"crypto/tls"
	"log"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cosiner/gohper/crypto/tls2"
	"github.com/cosiner/gohper/utils/attrs"
	"github.com/cosiner/gohper/utils/defval"
	log2 "github.com/cosiner/ygo/log"
	"github.com/cosiner/ygo/resource"
	websocket "github.com/cosiner/zerver_websocket"
)

const (
	// server status
	_NORMAL    = 0
	_DESTROYED = 1
)

type (
	LifetimeHook func(*Server) error

	ServerOption struct {
		Logger log2.Logger

		// server listening address, default :4000
		ListenAddr string

		// check websocket header, default nil
		WebSocketChecker HeaderChecker

		// path variables count, suggest set as max or average, default 3
		PathVarCount int
		// filters count for each route, RootFilters is not include, default 5
		FilterCount int

		// whether process user request or not when
		// header value of "Accept" is not acceptable and there is no default
		// resource type
		ProcessNotAcceptable bool

		// read timeout
		ReadTimeout time.Duration
		// write timeout
		WriteTimeout time.Duration
		// max header bytes
		MaxHeaderBytes int
		// tcp keep-alive period by minutes,
		// default 3 minute, same as predefined in standard http package
		KeepAlivePeriod time.Duration

		// CA pem files to verify client certs
		CAs []string
		// ssl config, default disable tls
		CertFile, KeyFile string
		// if not nil, cert and key will be ignored
		TLSConfig *tls.Config
	}

	// Server represent a web server
	Server struct {
		RootPath string
		// public logger
		Log log2.Logger

		Router
		attrs.Attrs
		RootFilters RootFilters // Match Every Routes
		ResMaster   resource.Master
		componentManager
		// private logger
		log log2.Logger

		checker              websocket.HandshakeChecker
		processNotAcceptable bool

		listener    net.Listener
		state       int32          // destroy or normal running
		activeConns sync.WaitGroup // connections in service, don't include hijacked and websocket connections

		destroyHooks []LifetimeHook
	}

	// HeaderChecker is a http header checker, it accept a function which can get
	// http header's value by name , if there is something wrong, throw an error
	// to terminate this request
	HeaderChecker func(func(string) string) error

	// Environment is a server enviroment, real implementation is the Server itself.
	// it can be accessed from Request/WebsocketConn
	Environment interface {
		Server() *Server
		ResourceMaster() *resource.Master
		Logger() log2.Logger
		StartTask(path string, value interface{})
		Component(name string) (interface{}, error)
	}

	ComponentNotFoundError string
)

func (err ComponentNotFoundError) Name() string {
	return string(err)
}

func (err ComponentNotFoundError) Error() string {
	return "component \"" + string(err) + "\" is not found"
}

// NewServer create a new server with default router
func NewServer(rootPath string) *Server {
	return NewServerWith(rootPath, nil, nil)
}

// NewServerWith create a new server with given router and root filters
func NewServerWith(rootPath string, rt Router, filters RootFilters) *Server {
	if filters == nil {
		filters = NewRootFilters(nil)
	}
	if rt == nil {
		rt = NewRouter()
	}

	return &Server{
		RootPath: rootPath,

		Router:           rt,
		Attrs:            attrs.NewLocked(),
		RootFilters:      filters,
		ResMaster:        resource.NewMaster(),
		componentManager: newComponentManager(),
	}
}

func (s *Server) Server() *Server {
	return s
}

func (s *Server) Filepath(path string) string {
	return filepath.Join(s.RootPath, path)
}

func (s *Server) Logger() log2.Logger {
	return s.Log
}

func (s *Server) ResourceMaster() *resource.Master {
	return &s.ResMaster
}

// RegisterComponent let server manage this component and it's lifetime.
// If name is "", component must implements Component, and it will initialized at
// server start and can't be accessed by name.
// Otherwise, it can be a Component, or others.
//
// If component is added before server start, it will be initialized when server start,
// otherwise, initialized when first required by Server.Compoenent
//
// When global component is initializing, the Environment passed to Init is exactly a
// ComponentEnvironment
func (s *Server) RegisterComponent(name string, component interface{}) ComponentEnvironment {
	return s.componentManager.RegisterComponent(s, name, component)
}

// StartTask start a task synchronously, the value will be passed to task handler
func (s *Server) StartTask(path string, value interface{}) {
	handler := s.MatchTaskHandler(&url.URL{Path: path})
	if handler == nil {
		s.Log.Error("No task handler found for:", path)
		return
	}

	handler.Handle(newTask(path, value))
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
	} else {
		conn, err := websocket.UpgradeWebsocket(w, request, s.checker)
		if err == nil {
			handler.Handle(newWebSocketConn(s, conn, indexer))
			indexer.destroySelf()
		} // else connecion will be auto-closed when error occoured,
	}
}

// serveHTTP serve for http protocal
func (s *Server) serveHTTP(w http.ResponseWriter, request *http.Request) {
	url := request.URL
	url.Host = request.Host
	handler, indexer, filters := s.MatchHandlerFilters(url)

	res, resType := s.ResMaster.Resource(request.Header.Get(HEADER_ACCEPT))

	reqEnv := newRequestEnvFromPool()
	req := reqEnv.req.init(s, res, request, indexer)
	resp := reqEnv.resp.init(s, res, w)

	var chain FilterChain
	if handler == nil {
		resp.ReportNotFound()
	} else if res == nil && !s.processNotAcceptable {
		resp.ReportNotAcceptable()
	} else if chain = FilterChain(handler.Handler(req.Method())); chain == nil {
		resp.ReportMethodNotAllowed()
	} else {
		resp.SetContentType(resType, res)
	}

	newFilterChain(s.RootFilters.Filters(url),
		newFilterChain(filters, chain),
	)(req, resp)

	req.destroy()
	resp.destroy()

	recycleRequestEnv(reqEnv)
	recycleFilters(filters)
}

func (o *ServerOption) init() {
	defval.String(&o.ListenAddr, ":4000")
	defval.Int(&o.PathVarCount, 3)
	defval.Int(&o.FilterCount, 5)
	if o.KeepAlivePeriod == 0 {
		o.KeepAlivePeriod = 3 * time.Minute // same as net/http/server.go:tcpKeepAliveListener
	}
	if o.Logger == nil {
		o.Logger = log2.New(nil).AddWriter(log2.NewConsoleWriter(nil))
	}
}

func (o *ServerOption) TLSEnabled() bool {
	return o.CertFile != "" || o.TLSConfig != nil
}

// all log message before server start will use standard log package
func (s *Server) config(o *ServerOption) {
	o.init()

	s.Log = o.Logger
	s.log = s.Log.Prefix("[Server]")
	var (
		log    = s.log.Info
		errors []error
		logErr = func(err error) {
			if err != nil {
				errors = append(errors, err)
			}
		}
	)

	s.checker = websocket.HeaderChecker(o.WebSocketChecker).HandshakeCheck

	if len(s.ResMaster.Resources) == 0 {
		log("Use default resource:", resource.RES_JSON)
		s.ResMaster.DefUse(resource.RES_JSON, resource.JSON{})
	} else {
		log("Resource types:", s.ResMaster.Types, " Default:", s.ResMaster.Default)
	}

	s.processNotAcceptable = o.ProcessNotAcceptable
	log("Process non-acceptable request:", s.processNotAcceptable)

	log("VarCountPerRoute:", o.PathVarCount)
	pathVarCount = o.PathVarCount
	log("FilterCountPerRoute:", o.FilterCount)
	filterCount = o.FilterCount

	s.componentManager.initHook = func(name string) {
		switch name {
		case _GLOBAL_COMPONENT:
			log("Init global components")
		case _ANONYMOUS_COMPONENT:
			log("Init anonymous components")
		default:
			log("  " + name)
		}
	}
	logErr(s.componentManager.Init(s))

	log("Execute registered init before routes funcs ")
	for _, f := range s.beforeRoutes(nil) {
		logErr(f(s))
	}

	log("Init root filters")
	logErr(s.RootFilters.Init(s))

	log("Init Handlers and Filters")
	logErr(s.Router.Init(s))

	log("Execute registered finial init funcs")
	for _, f := range s.finalInits(nil) {
		logErr(f(s))
	}

	if len(errors) != 0 {
		s.Log.Fatal(errors)
	}

	// destroy temporary data store
	tmpDestroy()
	log("Server Start: ", o.ListenAddr)

	runtime.GC()
}

// Start server as http server, if opt is nil, use default configurations
func (s *Server) Start(opt *ServerOption) error {
	if opt == nil {
		opt = &ServerOption{}
	}
	s.config(opt)

	l := s.listen(opt)

	s.listener = l
	srv := &http.Server{
		ReadTimeout:  opt.ReadTimeout,
		WriteTimeout: opt.WriteTimeout,
		Handler:      s,
		ConnState:    s.connStateHook,
	}

	return srv.Serve(l)
}

// from net/http/server/go
type tcpKeepAliveListener struct {
	*net.TCPListener
	AlivePeriod time.Duration
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

func (s *Server) listen(opt *ServerOption) net.Listener {
	ln, err := net.Listen("tcp", opt.ListenAddr)
	if err != nil {
		log.Panicln(err)
	}

	ln = &tcpKeepAliveListener{
		TCPListener: ln.(*net.TCPListener),
		AlivePeriod: opt.KeepAlivePeriod,
	}

	if opt.TLSConfig != nil {
		return tls.NewListener(ln, opt.TLSConfig)
	}

	if opt.CertFile == "" {
		return ln
	}

	// from net/http/server.go.ListenAndServeTLS
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

	if err != nil {
		if ln != nil {
			if e := ln.Close(); e != nil {
				log.Panicln(e)
			}
		}
		log.Panicln(err)
	}

	return ln
}

func (s *Server) connStateHook(conn net.Conn, state http.ConnState) {
	switch state {
	case http.StateActive:
		if atomic.LoadInt32(&s.state) == _NORMAL {
			s.activeConns.Add(1)
		} else {
			// previous idle connections before call server.Destroy() becomes active, directly close it
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

// Destroy server, release all resources, if destroyed, server can't be reused
// It only wait for managed connections, hijacked/websocket connections will not waiting
// if timeout or server already destroyed, false was returned
func (s *Server) Destroy(timeout time.Duration) bool {
	if !atomic.CompareAndSwapInt32(&s.state, _NORMAL, _DESTROYED) { // signal close idle connections
		return false
	}

	var isTimeout = true
	s.warnLog(s.listener.Close()) // don't accept connections
	if timeout > 0 {
		c := make(chan struct{})
		go func(s *Server, c chan struct{}) {
			s.activeConns.Wait() // wait connections in service to be idle
			close(c)
		}(s, c)

		select {
		case <-time.NewTicker(timeout).C:
		case <-c:
			isTimeout = false
		}
	} else {
		s.activeConns.Wait() // wait connections in service to be idle
	}

	// release resources
	s.RootFilters.Destroy()
	s.Router.Destroy()
	s.componentManager.Destroy()
	for _, fn := range s.destroyHooks {
		err := fn(s)
		if err != nil {
			s.log.Error("Destroy:", err)
		}
	}

	return !isTimeout
}

func (s *Server) warnLog(err error) {
	if err != nil {
		s.log.Warn(err)
	}
}

func (s *Server) finalInits(fn []LifetimeHook) (funcs []LifetimeHook) {
	return s.inits("finalInits", fn)
}

func (s *Server) inits(typ string, fn []LifetimeHook) (funcs []LifetimeHook) {
	if val := TmpHGet(s, typ); val != nil {
		funcs = val.([]LifetimeHook)
	}

	if len(fn) > 0 {
		funcs = append(funcs, fn...)
		TmpHSet(s, typ, funcs)
	}

	return funcs
}

// FinalInit add functions to execute after all others done and before server start
// don't register component or add handler, filter in these functions unless you know
// what are you doing
func (s *Server) FinalInit(fn ...LifetimeHook) {
	s.finalInits(fn)
}

func (s *Server) beforeRoutes(fn []LifetimeHook) (funcs []LifetimeHook) {
	return s.inits("beforeRoutes", fn)
}

// BeforeRoutes add functions to execute before routes init
func (s *Server) BeforeRoutes(fn ...LifetimeHook) {
	s.beforeRoutes(fn)
}

func (s *Server) BeforeDestroy(fn ...LifetimeHook) {
	s.destroyHooks = append(s.destroyHooks, fn...)
}
