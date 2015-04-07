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

func (hr *HostRouter) AddRouter(host string, rt Router) {
	l := len(hr.hosts) + 1
	hosts, routers := make([]string, l), make([]Router, l)
	copy(hosts, hr.hosts)
	copy(routers, hr.routers)
	hosts[l], routers[l] = host, rt
	hr.hosts, hr.routers = hosts, routers
}

// Implement RouterMatcher

func (hr *HostRouter) match(url *url.URL) Router {
	host, hosts := url.Host, hr.hosts
	for i := range hosts {
		if hosts[i] == host {
			return hr.routers[i]
		}
	}
	return nil
}

// Init init handlers and filters, websocket handlers
func (hr *HostRouter) Init(s *Server) (err error) {
	for i := 0; i < len(hr.routers) && err == nil; i++ {
		err = hr.routers[i].Init(s)
	}
	return
}

// Destroy destroy router, also responsible for destroy all handlers and filters
func (hr *HostRouter) Destroy() {
	for i := 0; i < len(hr.routers); i++ {
		hr.routers[i].Destroy()
	}
	return
}

// MatchHandlerFilters match given url to find all matched filters and final handler
func (hr *HostRouter) MatchHandlerFilters(url *url.URL) (handler Handler, indexer URLVarIndexer, filters []Filter) {
	if router := hr.match(url); router != nil {
		handler, indexer, filters = router.MatchHandlerFilters(url)
	}
	return
}

// MatchWebSocketHandler match given url to find a matched websocket handler
func (hr *HostRouter) MatchWebSocketHandler(url *url.URL) (handler WebSocketHandler, indexer URLVarIndexer) {
	if router := hr.match(url); router != nil {
		handler, indexer = router.MatchWebSocketHandler(url)
	}
	return
}

// MatchTaskHandler
func (hr *HostRouter) MatchTaskHandler(url *url.URL) (handler TaskHandler) {
	if router := hr.match(url); router != nil {
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

func (hr *HostRouter) PrintRouteTree(w io.Writer) {
	for i := range hr.routers {
		sys.WriteStrln(w, hr.hosts[i])
		hr.routers[i].PrintRouteTree(w)
	}
}
