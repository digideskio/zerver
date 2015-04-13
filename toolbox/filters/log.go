package filters

import (
	"time"

	"github.com/cosiner/zerver"
)

type Log func(v ...interface{})

func (l Log) Init(zerver.Enviroment) error { return nil }

func (l Log) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	chain(req, resp)
	status := resp.Status()
	l("[AccessLog]", status, req.URL().String(), req.RemoteAddr(), req.UserAgent())
}

func (l Log) Destroy() {}

type TimeLog func(...interface{})

func (l TimeLog) Init(zerver.Enviroment) error { return nil }

func (l TimeLog) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	start := time.Now().UnixNano()
	chain(req, resp)
	nano := time.Now().UnixNano() - start
	status := resp.Status()
	l("[AccessLog]", nano, "ns ", status, req.URL().String(), req.RemoteAddr(), req.UserAgent())
}

func (l TimeLog) Destroy() {}
