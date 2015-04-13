package filters

import (
	"github.com/cosiner/zerver"
)

type Log func(v ...interface{})

func (l Log) Init(zerver.Enviroment) error { return nil }

func (l Log) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	status := resp.Status()
	chain(req, resp)
	newStatus := resp.Status()
	if status != newStatus {
		l("[AccessLog]", status, "->", newStatus, req.URL().String(), req.RemoteAddr(), req.UserAgent())
	} else {
		l("[AccessLog]", status, req.URL().String(), req.RemoteAddr(), req.UserAgent())
	}
}

func (l Log) Destroy() {}
