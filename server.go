package zerver

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cosiner/gohper/crypto/tls2"
	"github.com/cosiner/gohper/encoding"
	"github.com/cosiner/gohper/utils/attrs"
	"github.com/cosiner/gohper/utils/defval"
	"github.com/cosiner/ygo/log"
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
		// server listening address, default :4000
		ListenAddr string

		// check websocket header, default nil
		WebSocketChecker HeaderChecker

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

		Headers map[string]string
		Codec   encoding.Codec
	}

	// Server represent a web server
	Server struct {
		RootPath string
		Router
		attrs.Attrs
		components CompManager

		checker websocket.HandshakeChecker

		listener    net.Listener
		state       int32          // destroy or normal running
		activeConns sync.WaitGroup // connections in service, don't include hijacked and websocket connections

		hooks map[string][]LifetimeHook

		headers map[string]string
		codec   encoding.Codec
	}

	// HeaderChecker is a http header checker, it accept a function which can get
	// http header's value by name , if there is something wrong, throw an error
	// to terminate this request
	HeaderChecker func(func(string) string) error
)

// NewServer create a new server with default router
func NewServer(rootPath string) *Server {
	return NewServerWith(rootPath, nil)
}

// NewServerWith create a new server with given router and root filters
func NewServerWith(rootPath string, rt Router) *Server {
	if rt == nil {
		rt = NewRouter()
	}

	return &Server{
		RootPath: rootPath,

		Router:     rt,
		Attrs:      attrs.NewLocked(),
		components: NewCompManager(),

		hooks: make(map[string][]LifetimeHook),
	}
}

func (s *Server) Server() *Server {
	return s
}

func (s *Server) Filepath(path string) string {
	sep := string(filepath.Separator)
	if path != "" && !strings.HasPrefix(path, sep) {
		path = strings.Replace(path, "/", sep, -1)
	}
	return filepath.Join(s.RootPath, path)
}

func (s *Server) Codec() encoding.Codec {
	return s.codec
}

// RegisterComponent let server manage this component and it's lifetime.
// If name is empty, component must implements Component, and it will initialized at
// server start and can't be accessed by name.
// Otherwise, it can be a Component, or others.
//
// When global component is initializing, the Environment passed to Init is exactly a
// CompEnv
func (s *Server) RegisterComponent(name string, component interface{}) *CompEnv {
	return s.components.Register(s, name, component)
}

func (s *Server) Component(name string) (interface{}, error) {
	return s.components.Get(name)
}

func (s *Server) RemoveComponent(name string) {
	s.components.Remove(name)
}

// StartTask start a task synchronously, the value will be passed to task handler
func (s *Server) StartTask(path string, value interface{}) {
	handler := s.MatchTaskHandler(&url.URL{Path: path})
	if handler == nil {
		log.Error("No task handler found for:", path)
		return
	}

	handler.Handle(newTask(path, value))
}

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

func (s *Server) serveWebSocket(w http.ResponseWriter, request *http.Request) {
	handler, vars := s.MatchWebSocketHandler(request.URL)
	if handler == nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		conn, err := websocket.UpgradeWebsocket(w, request, s.checker)
		if err == nil {
			handler.Handle(newWsConn(s, conn, &vars))
		} // else connecion will be auto-closed when error occoured,
	}
}

func (s *Server) serveHTTP(w http.ResponseWriter, request *http.Request) {
	url := request.URL
	url.Host = request.Host
	handler, vars, filters := s.MatchHandlerFilters(url)

	reqEnv := newRequestEnv()
	req := reqEnv.req.init(s, request, &vars)
	resp := reqEnv.resp.init(s, w)

	headers := resp.Headers()
	for k, v := range s.headers {
		headers.Set(k, v)
	}

	var chain FilterChain
	if handler == nil {
		resp.StatusCode(http.StatusNotFound)
	} else if chain = FilterChain(handler.Handler(req.ReqMethod())); chain == nil {
		resp.StatusCode(http.StatusMethodNotAllowed)
	}

	newFilterChain(chain, filters...)(req, resp)

	req.destroy()
	resp.destroy()
	recycleRequestEnv(reqEnv)
}

func (o *ServerOption) init() {
	defval.String(&o.ListenAddr, ":4000")
	if o.KeepAlivePeriod == 0 {
		o.KeepAlivePeriod = 3 * time.Minute // same as net/http/server.go:tcpKeepAliveListener
	}
	if o.Codec == nil {
		o.Codec = encoding.DefaultCodec
	}
}

func (o *ServerOption) TLSEnabled() bool {
	return o.CertFile != "" || o.TLSConfig != nil
}

func (s *Server) config(o *ServerOption) {
	o.init()

	var (
		errors []error
		logErr = func(err error) {
			if err != nil {
				errors = append(errors, err)
			}
		}
	)
	s.codec = o.Codec
	s.headers = o.Headers
	s.checker = websocket.HeaderChecker(o.WebSocketChecker).HandshakeCheck

	logErr(s.components.Init(s))

	log.Info("Execute registered init before routes funcs ")
	for _, f := range s.OnLoadRoutes() {
		logErr(f(s))
	}

	log.Info("Init Handlers and Filters")
	logErr(s.Router.Init(s))

	log.Info("Execute registered finial init funcs")
	for _, f := range s.OnStart() {
		logErr(f(s))
	}

	if len(errors) != 0 {
		log.Fatal(errors)
	}

	log.Info("Server Start: ", o.ListenAddr)
	runtime.GC()
}

// Start server as http server, if opt is nil, use default configurations
func (s *Server) Start(opt *ServerOption) error {
	runtime.GOMAXPROCS(runtime.NumCPU())

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
		log.Panic(err)
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
				log.Panic(e)
			}
		}
		log.Panic(err)
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
	err := s.listener.Close() // don't accept connections
	if err != nil {
		log.Warn("listener close", err)
	}

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

	s.Router.Destroy()
	s.components.Destroy()

	for _, fn := range s.OnDestroy() {
		err := fn(s)
		if err != nil {
			log.Error("destroy hook:", err)
		}
	}

	return !isTimeout
}

func (s *Server) appendHooks(typ string, fn ...LifetimeHook) []LifetimeHook {
	hooks := s.hooks[typ]
	hooks = append(hooks, fn...)
	s.hooks[typ] = hooks
	return hooks
}

func (s *Server) OnStart(fn ...LifetimeHook) []LifetimeHook {
	return s.appendHooks("onStart", fn...)
}

func (s *Server) OnLoadRoutes(fn ...LifetimeHook) []LifetimeHook {
	return s.appendHooks("onLoadRoutes", fn...)
}

func (s *Server) OnDestroy(fn ...LifetimeHook) []LifetimeHook {
	return s.appendHooks("onDestroy", fn...)
}
