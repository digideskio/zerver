package filter

import (
	"github.com/cosiner/gohper/net2/http2"
	"github.com/cosiner/gohper/time2"
	"github.com/cosiner/ygo/log"
	"github.com/cosiner/zerver"
)

type Log struct {
	CountTime bool
}

func (l *Log) Init(env zerver.Env) error {
	return nil
}

func (l *Log) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	if l.CountTime {
		now := time2.Now()

		chain(req, resp)
		cost := time2.Now().Sub(now)
		log.Info(
			cost.String(),
			resp.StatusCode(0),
			req.ReqMethod(),
			req.URL().Path,
			http2.IpOfAddr(req.RemoteAddr()),
			req.GetHeader(zerver.HEADER_USERAGENT),
		)
	} else {
		chain(req, resp)
		log.Info(
			resp.StatusCode(0),
			req.ReqMethod(),
			req.URL().Path,
			http2.IpOfAddr(req.RemoteAddr()),
			req.GetHeader(zerver.HEADER_USERAGENT),
		)
	}
}

func (l *Log) Destroy() {}
