package router

import (
	"github.com/cosiner/zerver"
)

type GroupRouter struct {
	prefix string
	zerver.Router
}

func NewGroupRouter(rt zerver.Router, prefix string) zerver.Router {
	return GroupRouter{
		prefix: prefix,
		Router: rt,
	}
}

func (gr GroupRouter) Filter(pattern string, f zerver.Filter) error {
	return gr.Router.Filter(gr.prefix+pattern, f)
}

func (gr GroupRouter) Handler(pattern string, h zerver.Handler) error {
	return gr.Router.Handler(gr.prefix+pattern, h)
}

func (gr GroupRouter) TaskHandler(pattern string, th zerver.TaskHandler) error {
	return gr.Router.TaskHandler(gr.prefix+pattern, th)
}

func (gr GroupRouter) WsHandler(pattern string, th zerver.WsConn) error {
	return gr.Router.WsHandler(gr.prefix+pattern, th)
}
