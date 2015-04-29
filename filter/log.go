package filter

import (
	"log"
	"time"

	"github.com/cosiner/zerver"
)

type Log struct {
	Printer   func(...interface{}) // default use log.Println
	CountTime bool
}

func (l Log) Init(zerver.Enviroment) error {
	if l.Printer == nil {
		l.Printer = log.Println
	}
	return nil
}

func (l Log) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	if l.CountTime {
		nano := time.Now().UnixNano()
		chain(req, resp)
		nano = time.Now().UnixNano() - nano
		l.Printer("[AccessLog]",
			nano, "ns ",
			resp.Status(),
			req.URL().Path,
			req.RemoteIP(),
			req.UserAgent())
	} else {
		chain(req, resp)
		l.Printer("[AccessLog]",
			resp.Status(),
			req.URL().Path,
			req.RemoteIP(),
			req.UserAgent())
	}
}

func (l Log) Destroy() {}
