package zerver

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"runtime"

	websocket "github.com/cosiner/zerver_websocket"

	. "github.com/cosiner/gohper/lib/errors"
)

type (
	// Server represent a web server
	Server struct {
		Router
		AttrContainer
		RootFilters RootFilters // Match Every Routes
		checker     websocket.HeaderChecker
		ContentType string // default content type
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
func (s *Server) start() {
	if s.ContentType == "" {
		s.ContentType = CONTENTTYPE_JSON
	}
	if Marshaler == nil {
		Marshaler = json.Marshal
	}
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
func (s *Server) Start(listenAddr string) error {
	s.start()
	return http.ListenAndServe(listenAddr, s)
}

// Destroy is mainly designed to stop server, release all resources
// but it seems ther is no approach to stop a running golang but don't exit process
func (s *Server) Destroy() {
	s.RootFilters.Destroy()
	s.Router.Destroy()
}

// StartTLS start server as https server
func (s *Server) StartTLS(listenAddr, certFile, keyFile string) error {
	s.start()
	return http.ListenAndServeTLS(listenAddr, certFile, keyFile, s)
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

// StartTask add a task
// Task must have corresponding handler, otherwise server will panic
func (s *Server) StartTask(async bool, path string, value interface{}) {
	if async {
		go s.serveTask(path, value)
	} else {
		s.serveTask(path, value)
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
	requestEnv := Pool.newRequestEnv()
	req, resp := requestEnv.req.init(s, request, indexer), requestEnv.resp.init(w)
	resp.SetContentType(s.ContentType)
	var chain FilterChain
	if handler == nil {
		resp.ReportNotFound()
	} else if chain = FilterChain(indicateHandler(req.Method(), handler)); chain == nil {
		resp.ReportMethodNotAllowed()
	}
	newFilterChain(s.RootFilters.Filters(url), newFilterChain(filters, chain))(req, resp)
	req.destroy()
	resp.destroy()
	Pool.recycleRequestEnv(requestEnv)
	Pool.recycleFilters(filters)
}

// serveTask serve for asynchronous task
func (s *Server) serveTask(path string, value interface{}) {
	u, err := url.Parse(path)
	if err != nil {
		PanicServer(err.Error())
	}
	handler, indexer := s.MatchTaskHandler(u)
	if handler == nil {
		PanicServer("No task handler found for " + path)
	}
	task := newTask(s, indexer, value)
	handler.Handle(task)
	task.destroy()
}

// PanicServer create a new goroutine, it force panic whole process
func PanicServer(s string) {
	go panic(s)
}
