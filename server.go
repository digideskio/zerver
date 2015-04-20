package zerver

import (
	"log"
	"net/http"
	"net/url"
	"runtime"
	"sync"

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
		// resource marshal/pool/unmarshal
		// first search by Server.Component, if not found, use JSONResource
		ResourceMaster
		// error logger, default use log.Println
		ErrorLogger func(...interface{})
	}

	ComponentState struct {
		Initialized bool
		NoLazy      bool
		Component
	}
	components map[string]ComponentState

	// Server represent a web server
	Server struct {
		Router
		AttrContainer
		RootFilters RootFilters // Match Every Routes
		Errorln     func(...interface{})

		components
		managedComponents []Component
		sync.RWMutex
		checker     websocket.HandshakeChecker
		contentType string // default content type
		resMaster   ResourceMaster
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

func (s *Server) ResourceMaster() ResourceMaster {
	return s.resMaster
}

func (s *Server) Component(name string) (Component, error) {
	s.RLock()
	if cs, has := s.components[name]; has {
		s.RUnlock()
		var err error
		if !cs.Initialized {
			s.Lock()
			if !cs.Initialized {
				if err = cs.Component.Init(s); err == nil {
					cs.Initialized = true
				}
			}
			s.Unlock()
		}
		return cs.Component, err
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

// Start start server
func (s *Server) start(o *ServerOption) {
	o.init()
	s.Errorln = o.ErrorLogger
	s.Errorln("ContentType:", o.ContentType)
	s.contentType = o.ContentType
	s.checker = websocket.HeaderChecker(o.WebSocketChecker).HandshakeCheck
	s.resMaster = o.ResourceMaster
	if s.resMaster == nil {
		c, err := s.Component(COMP_RESOURCE)
		if err == nil {
			s.resMaster = c.(ResourceMaster)
			s.Errorln("Search ResourceMaster: customed")
		} else if err == ErrComponentNotFound {
			s.resMaster = JSONResource{}
			s.Errorln("Search ResourceMaster: default JSONResource")
		} else {
			panic(err)
		}
	}
	s.Errorln("VarCountPerRoute:", o.PathVarCount)
	pathVarCount = o.PathVarCount
	s.Errorln("FilterCountPerRoute:", o.FilterCount)
	filterCount = o.FilterCount
	s.Errorln("Init managed components")
	for i := range s.managedComponents {
		errors.OnErrPanic(s.managedComponents[i].Init(s))
	}
	s.Errorln("Init root filters")
	errors.OnErrPanic(s.RootFilters.Init(s))
	s.Errorln("Init Handlers and Filters")
	errors.OnErrPanic(s.Router.Init(s))
	s.Errorln("Server Start:", o.ListenAddr)
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
	s.components.destroy()
	for i := range s.managedComponents {
		s.managedComponents[i].Destroy()
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

// serveTask serve for task
func (s *Server) StartTask(path string, value interface{}) {
	if handler := s.MatchTaskHandler(&url.URL{Path: path}); handler != nil {
		handler.Handle(value)
		return
	}
	panic("No task handler found for " + path)
}

// ManageComponent manage those filters used in InterceptHandler, or those added to
// multiple routes, for first condition, ManageComponent used to Init them;
// for second condition, ManageComponent used to avoid multiple call of Init
func (s *Server) ManageComponent(c Component) {
	s.managedComponents = append(s.managedComponents, c)
}
