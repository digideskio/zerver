package filter

import (
	"time"

	"github.com/cosiner/zerver"
)

type Log struct {
	logger    zerver.Logger
	CountTime bool
}

func (l *Log) Init(env zerver.Enviroment) error {
	l.logger = env.Logger()
	return nil
}

func (l *Log) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	if l.CountTime {
		nano := time.Now().UnixNano()
		chain(req, resp)
		nano = time.Now().UnixNano() - nano
		l.logger.Infoln("[AccessLog]",
			nano, "ns ",
			resp.Status(),
			req.URL().Path,
			req.RemoteIP(),
			req.UserAgent())
	} else {
		chain(req, resp)
		l.logger.Infoln("[AccessLog]",
			resp.Status(),
			req.URL().Path,
			req.RemoteIP(),
			req.UserAgent())
	}
}

func (l *Log) Destroy() {}
