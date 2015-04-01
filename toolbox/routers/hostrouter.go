package routers

import (
	"io"
	"net/url"

	. "github.com/cosiner/zerver"

	"github.com/cosiner/gohper/lib/sys"
)

type (
	HostRouter struct {
		Router
		hosts   []string
		routers []Router
	}
)

func NewHostRouter() *HostRouter {
	return &HostRouter{}
}

func (hr HostRouter) AddRouter(host string, rt Router) {
	l := len(hr.hosts) + 1
	hosts, routers := make([]string, l), make([]Router, l)
	copy(hosts, hr.hosts)
	copy(routers, hr.routers)
	hosts[l], routers[l] = host, rt
	hr.hosts, hr.routers = hosts, routers
}

// Implement RouterMatcher

func (hr HostRouter) match(url *url.URL) Router {
	host, hosts := url.Host, hr.hosts
	for i := range hosts {
		if hosts[i] == host {
			return hr.routers[i]
		}
	}
	return nil
}

func (hr HostRouter) iterate(fn func(string, Router)) {
	hosts, routers := hr.hosts, hr.routers
	for i := range routers {
		fn(hosts[i], routers[i])
	}
}

// Init init handlers and filters, websocket handlers
func (hr HostRouter) Init(initHandler func(Handler) bool,
	initFilter func(Filter) bool,
	initWebSocket func(WebSocketHandler) bool,
	initTask func(TaskHandler) bool) {
	hr.iterate(func(_ string, rt Router) {
		rt.Init(initHandler, initFilter, initWebSocket, initTask)
	})
}

// Destroy destroy router, also responsible for destroy all handlers and filters
func (hr HostRouter) Destroy() {
	hr.iterate(func(_ string, rt Router) {
		rt.Destroy()
	})
}

// MatchHandlerFilters match given url to find all matched filters and final handler
func (hr HostRouter) MatchHandlerFilters(url *url.URL) (handler Handler, indexer URLVarIndexer, filters []Filter) {
	if router := hr.match(url); router != nil {
		handler, indexer, filters = router.MatchHandlerFilters(url)
	}
	return
}

// MatchWebSocketHandler match given url to find a matched websocket handler
func (hr HostRouter) MatchWebSocketHandler(url *url.URL) (handler WebSocketHandler, indexer URLVarIndexer) {
	if router := hr.match(url); router != nil {
		handler, indexer = router.MatchWebSocketHandler(url)
	}
	return
}

// MatchTaskHandler
func (hr HostRouter) MatchTaskHandler(url *url.URL) (handler TaskHandler, indexer URLVarIndexer) {
	if router := hr.match(url); router != nil {
		handler, indexer = router.MatchTaskHandler(url)
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

func (hr HostRouter) PrintRouteTree(w io.Writer) {
	hr.iterate(func(ident string, rt Router) {
		if _, e := sys.WriteStrln(w, ident); e == nil {
			rt.PrintRouteTree(indentWriter{w})
		}
	})
}
