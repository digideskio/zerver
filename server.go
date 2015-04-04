package zerver

import (
	"log"
	"net/http"
	"net/url"
	"runtime"

	websocket "github.com/cosiner/zerver_websocket"

	. "github.com/cosiner/gohper/lib/errors"
)

type (
	ServerOption struct {
		WebSocketChecker HeaderChecker // default nil
		ContentType      string        // default application/json;charset=utf-8
		PathVarCount     int           // default 2
		FilterCount      int           // default 2
		ListenAddr       string        // default :4000
		CertFile         string        // default not enable tls
		KeyFile          string
	}

	// Server represent a web server
	Server struct {
		Router
		AttrContainer
		RootFilters RootFilters // Match Every Routes
		checker     websocket.HeaderChecker
		contentType string // default content type
	}

	// HeaderChecker is a http header checker, it accept a function which can get
	// any httper's value, it there is something wrong, throw an error
	HeaderChecker func(header func(string) string) error

	// ServerInitializer is a Object which will automaticlly initialed by server if
	// it's added to server, else it should initialed manually
	ServerInitializer interface {
		Init(s *Server) error
	}

	// serverGetter is a server getter
	serverGetter interface {
		Server() *Server
	}
)

func (o *ServerOption) init() {
	if o.ListenAddr == "" {
		o.ListenAddr = ":4000"
	}
	if o.ContentType == "" {
		o.ContentType = CONTENTTYPE_JSON
	}
	if o.FilterCount == 0 {
		o.FilterCount = 2
	}
	if o.PathVarCount == 0 {
		o.PathVarCount = 2
	}
}

// NewServer create a new server
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
	}
}

// Implement serverGetter
func (s *Server) Server() *Server {
	return s
}

// Start start server
func (s *Server) start(o *ServerOption) {
	o.init()
	s.contentType = o.ContentType
	pathVarCount = o.PathVarCount
	filterCount = o.FilterCount

	OnErrPanic(s.RootFilters.Init(s))
	log.Println("Init Handlers and Filters")
	s.Router.Init(func(handler Handler) bool {
		OnErrPanic(handler.Init(s))
		return true
	}, func(filter Filter) bool {
		OnErrPanic(filter.Init(s))
		return true
	}, func(websocketHandler WebSocketHandler) bool {
		OnErrPanic(websocketHandler.Init(s))
		return true
	}, func(taskHandler TaskHandler) bool {
		OnErrPanic(taskHandler.Init(s))
		return true
	})
	log.Println("Server Start")
	// destroy temporary data store
	tmpDestroy()
	runtime.GC()
}

// Start start server as http server
func (s *Server) Start(options *ServerOption) error {
	if options == nil {
		options = &ServerOption{}
	}
	s.start(options)
	if options.CertFile == "" {
		return http.ListenAndServe(options.ListenAddr, s)
	}
	return http.ListenAndServeTLS(options.ListenAddr, options.CertFile, options.KeyFile, s)
}

// Destroy is mainly designed to stop server, release all resources
// but it seems ther is no approach to stop a running golang but don't exit process
func (s *Server) Destroy() {
	s.RootFilters.Destroy()
	s.Router.Destroy()
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

// SetWebSocketHeaderChecker accept a checker function, checker can get an
func (s *Server) SetWebSocketHeaderChecker(checker HeaderChecker) {
	s.checker.Checker = checker
}

// serveWebSocket serve for websocket protocal
func (s *Server) serveWebSocket(w http.ResponseWriter, request *http.Request) {
	handler, indexer := s.MatchWebSocketHandler(request.URL)
	if handler == nil {
		w.WriteHeader(http.StatusNotFound)
	} else if conn, err := websocket.UpgradeWebsocket(w, request, s.checker.HandshakeCheck); err == nil {
		handler.Handle(newWebSocketConn(s, conn, indexer))
		indexer.destroySelf()
	}
}

// serveHTTP serve for http protocal
func (s *Server) serveHTTP(w http.ResponseWriter, request *http.Request) {
	url := request.URL
	url.Host = request.Host
	handler, indexer, filters := s.MatchHandlerFilters(url)
	requestEnv := newRequestEnvFromPool()
	req, resp := requestEnv.req.init(s, request, indexer), requestEnv.resp.init(w)
	resp.SetContentType(s.contentType)
	var chain FilterChain
	if handler == nil {
		resp.ReportNotFound()
	} else if chain = FilterChain(handler.Handler(req.Method())); chain == nil {
		resp.ReportMethodNotAllowed()
	}
	newFilterChain(s.RootFilters.Filters(url), newFilterChain(filters, chain))(req, resp)
	req.destroy()
	resp.destroy()
	recycleRequestEnv(requestEnv)
	recycleFilters(filters)
}

// serveTask serve for task
func (s *Server) StartTask(path string, value interface{}) {
	if handler := s.MatchTaskHandler(&url.URL{Path: path}); handler != nil {
		handler.Handle(value)
		return
	}
	panic("No task handler found for " + path)
}

// PanicServer create a new goroutine, it force panic whole process
func PanicServer(s string) {
	go panic(s)
}
