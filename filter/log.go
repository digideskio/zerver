package filter

import (
	"time"

	"github.com/cosiner/gohper/time2"

	"github.com/cosiner/gohper/defval"
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
		nano := time.Now().UnixNano()

		chain(req, resp)
		nano = time.Now().UnixNano() - nano
		l.logger.Infoln(
			time2.ToHuman(nano),
			resp.Status(),
			req.Method(),
			req.URL().Path,
			req.RemoteIP(),
			req.UserAgent())
	} else {
		chain(req, resp)
		l.logger.Infoln(
			resp.Status(),
			req.Method(),
			req.URL().Path,
			req.RemoteIP(),
			req.UserAgent())
	}
}

func (l *Log) Destroy() {}
