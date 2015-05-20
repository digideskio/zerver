package host

import (
	"io"
	"net/url"

	"github.com/cosiner/gohper/unsafe2"
	"github.com/cosiner/zerver"
)

type (
	Router struct {
		zerver.Router
		hosts   []string
		routers []zerver.Router
	}
)

func NewRouter() *Router {
	return &Router{}
}

func (r *Router) AddRouter(host string, rt zerver.Router) {
	l := len(r.hosts) + 1

	hosts, routers := make([]string, l), make([]zerver.Router, l)
	copy(hosts, r.hosts)
	copy(routers, r.routers)
	hosts[l], routers[l] = host, rt
	r.hosts, r.routers = hosts, routers
}

// Implement RouterMatcher

func (r *Router) match(url *url.URL) zerver.Router {
	host, hosts := url.Host, r.hosts
	for i := range hosts {
		if hosts[i] == host {
			return r.routers[i]
		}
	}

	return nil
}

// Init init handlers and filters, websocket handlers
func (r *Router) Init(env zerver.Environment) (err error) {
	for i := 0; i < len(r.routers) && err == nil; i++ {
		err = r.routers[i].Init(env)
	}

	return
}

// Destroy destroy router, also responsible for destroy all handlers and filters
func (r *Router) Destroy() {
	for i := 0; i < len(r.routers); i++ {
		r.routers[i].Destroy()
	}

	return
}

// MatchHandlerFilters match given url to find all matched filters and final handler
func (r *Router) MatchHandlerFilters(url *url.URL) (handler zerver.Handler, indexer zerver.URLVarIndexer, filters []zerver.Filter) {
	if router := r.match(url); router != nil {
		handler, indexer, filters = router.MatchHandlerFilters(url)
	}

	return
}

// MatchWebSocketHandler match given url to find a matched websocket handler
func (r *Router) MatchWebSocketHandler(url *url.URL) (handler zerver.WebSocketHandler, indexer zerver.URLVarIndexer) {
	if router := r.match(url); router != nil {
		handler, indexer = router.MatchWebSocketHandler(url)
	}

	return
}

// MatchTaskHandler
func (r *Router) MatchTaskHandler(url *url.URL) (handler zerver.TaskHandler) {
	if router := r.match(url); router != nil {
		handler = router.MatchTaskHandler(url)
	}

	return
}

type indentWriter struct {
	io.Writer
}

var indentBytes = []byte("    ")
var indentBytesLen = len(indentBytes)

func (w indentWriter) Write(data []byte) (int, error) {
	c, e := w.Writer.Write(indentBytes)
	if e == nil {
		c1, e1 := w.Writer.Write(data)
		c, e = c+c1, e1
	}

	return c, e
}

func (r *Router) PrintRouteTree(w io.Writer) {
	for i := range r.routers {
		w.Write(unsafe2.Bytes(r.hosts[i] + "\n"))
		r.routers[i].PrintRouteTree(w)
	}
}
