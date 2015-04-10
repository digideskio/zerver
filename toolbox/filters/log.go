package filters

import (
	"github.com/cosiner/zerver"
)

type LogFilter func(v ...interface{})

func (l LogFilter) Init(zerver.Enviroment) error { return nil }

func (l LogFilter) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	status := resp.Status()
	chain(req, resp)
	newStatus := status
	if status != status {
		l("[AccessLog]", status, "->", newStatus, req.URL().String(), req.RemoteAddr(), req.UserAgent())
	} else {
		l("[AccessLog]", status, req.URL().String(), req.RemoteAddr(), req.UserAgent())
	}
}

func (l LogFilter) Destroy() {}
