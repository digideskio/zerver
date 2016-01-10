package filter

import (
	"github.com/cosiner/gohper/time2"
	log "github.com/cosiner/ygo/jsonlog"
	"github.com/cosiner/zerver"
)

type Log struct {
	log *log.Logger
}

func (l *Log) Init(env zerver.Env) error {
	l.log = log.Derive("Filter", "Log")
	return nil
}

func (l *Log) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	now := time2.Now()
	chain(req, resp)
	cost := time2.Now().Sub(now)

	l.log.Info(log.M{
		"method":     req.ReqMethod(),
		"url":        req.URL().String(),
		"remote":     req.RemoteAddr(),
		"userAgent":  req.GetHeader(zerver.HEADER_USERAGENT),
		"cost":       cost.String(),
		"statusCode": resp.StatusCode(0),
	})
}

func (l *Log) Destroy() {}
