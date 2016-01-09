package router

import (
	"net/url"

	"github.com/cosiner/zerver"
)

type HostRouter struct {
	// for interface compatible
	zerver.Router

	hosts   []string
	routers []zerver.Router
}

func NewHostRouter() *HostRouter {
	return &HostRouter{}
}

func (r *HostRouter) AddRouter(host string, rt zerver.Router) {
	l := len(r.hosts) + 1

	hosts, routers := make([]string, l), make([]zerver.Router, l)
	copy(hosts, r.hosts)
	copy(routers, r.routers)
	hosts[l], routers[l] = host, rt
	r.hosts, r.routers = hosts, routers
}

// Implement RouterMatcher

func (r *HostRouter) match(url *url.URL) zerver.Router {
	host, hosts := url.Host, r.hosts
	for i := range hosts {
		if hosts[i] == host {
			return r.routers[i]
		}
	}

	return nil
}

// Init init handlers and filters, websocket handlers
func (r *HostRouter) Init(env zerver.Env) (err error) {
	for i := 0; i < len(r.routers) && err == nil; i++ {
		err = r.routers[i].Init(env)
	}

	return
}

func (r *HostRouter) Destroy() {
	for i := 0; i < len(r.routers); i++ {
		r.routers[i].Destroy()
	}

	return
}

func (r *HostRouter) MatchHandlerFilters(url *url.URL) (zerver.Handler, zerver.ReqVars, []zerver.Filter) {
	if router := r.match(url); router != nil {
		return router.MatchHandlerFilters(url)
	}

	return nil, zerver.ReqVars{}, nil
}

func (r *HostRouter) MatchWebSocketHandler(url *url.URL) (zerver.WsHandler, zerver.ReqVars) {
	if router := r.match(url); router != nil {
		return router.MatchWebSocketHandler(url)
	}

	return nil, zerver.ReqVars{}
}

func (r *HostRouter) MatchTaskHandler(url *url.URL) zerver.TaskHandler {
	if router := r.match(url); router != nil {
		return router.MatchTaskHandler(url)
	}
	return nil
}
