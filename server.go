package zerver

import (
	"log"
	"net/http"
	"net/url"
	"runtime"

	. "github.com/cosiner/gohper/lib/errors"
	"github.com/cosiner/gohper/lib/types"
	"github.com/cosiner/zerver_websocket"
)

var (
	Bytes  func(string) []byte = types.UnsafeBytes
	String func([]byte) string = types.UnsafeString
)

type (
	ServerOption struct {
		// check websocket header, default nil
		WebSocketChecker HeaderChecker
		// content type for each request, default application/json;charset=utf-8
		ContentType string
		// path variables count, suggest set as max or average, default 3
		PathVarCount int
		// filters count for each route, RootFilters is not include, default 5
		FilterCount int
		// server listening address, default :4000
		ListenAddr string
		// ssl config, default not enable tls
		CertFile, KeyFile string
		// resource marshal/pool/unmarshal, default use JSONResource
		*ResourceMaster
	}

	components map[string]Component

	// Server represent a web server
	Server struct {
		Router
		AttrContainer
		RootFilters RootFilters // Match Every Routes
		components
		checker     websocket.HandshakeChecker
		contentType string // default content type
		resMaster   *ResourceMaster
	}

	// HeaderChecker is a http header checker, it accept a function which can get
	// http header's value by name , if there is something wrong, throw an error
	// to terminate this request
	HeaderChecker func(func(string) string) error

	// Component is a Object which will automaticlly initialed/destroyed by server
	// if it's added to server, else it should initialed manually
	Component interface {
		Init(Enviroment) error
		Destroy()
	}

	// Enviroment is a server enviroment, real implementation is the Server itself.
	// it can be accessed from Request/WebsocketConn
	Enviroment interface {
		Server() *Server
		StartTask(path string, value interface{})
		Component(name string) Component
		ResourceMaster() *ResourceMaster
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
}

func (is components) init(env Enviroment) error {
	for _, i := range is {
		if e := i.Init(env); e != nil {
			return e
		}
	}
	return nil
}

func (is components) destroy() {
	for _, i := range is {
		i.Destroy()
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
		components:    make(components),
	}
}

// ent ServerEnviroment
func (s *Server) Server() *Server {
	return s
}

func (s *Server) ResourceMaster() *ResourceMaster {
	return s.resMaster
}

func (s *Server) Component(name string) Component {
	return s.components[name]
}

func (s *Server) AddComponent(name string, c Component) {
	if name != "" && c != nil {
		s.components[name] = c
	}
}

// Start start server
func (s *Server) start(o *ServerOption) {
	o.init()
	s.contentType = o.ContentType
	s.checker = websocket.HeaderChecker(o.WebSocketChecker).HandshakeCheck
	s.resMaster = o.ResourceMaster.Init()
	pathVarCount = o.PathVarCount
	filterCount = o.FilterCount

	OnErrPanic(s.components.init(s))
	OnErrPanic(s.RootFilters.Init(s))
	log.Println("Init Handlers and Filters")
	OnErrPanic(s.Router.Init(s))
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
	if s.components != nil {
		s.components.destroy()
	}
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
	}
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
