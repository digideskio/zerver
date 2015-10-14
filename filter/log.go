package filter

import (
	"github.com/cosiner/gohper/time2"
	"github.com/cosiner/gohper/utils/defval"
	"github.com/cosiner/ygo/log"
	"github.com/cosiner/zerver"
)

type Log struct {
	logger    log.Logger
	Prefix    string
	CountTime bool
}

func (l *Log) Init(env zerver.Environment) error {
	defval.String(&l.Prefix, "[Access]")
	l.logger = env.Logger().Prefix(l.Prefix)
	return nil
}

func (l *Log) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	if l.CountTime {
		now := time2.Now()

		chain(req, resp)
		cost := time2.Now().Sub(now)
		l.logger.Info(
			cost.String(),
			resp.Status(),
			req.Method(),
			req.URL().Path,
			req.RemoteIP(),
			req.UserAgent())
	} else {
		chain(req, resp)
		l.logger.Info(
			resp.Status(),
			req.Method(),
			req.URL().Path,
			req.RemoteIP(),
			req.UserAgent())
	}
}

func (l *Log) Destroy() {}
