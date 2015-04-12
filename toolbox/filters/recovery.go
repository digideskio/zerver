package filters

import (
	"log"
	"runtime"

	"github.com/cosiner/zerver"
)

type Recovery struct {
	Logger  func(string)
	Bufsize int
}

func (r *Recovery) Init(zerver.Enviroment) error {
	if r.Logger == nil {
		r.Logger = func(s string) {
			log.Print(s)
		}
	}
	if r.Bufsize == 0 {
		r.Bufsize = 1024 * 8
	}
	return nil
}

func (r *Recovery) Destroy() {}

func (r *Recovery) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	defer func() {
		if e := recover(); e != nil {
			resp.ReportInternalServerError()
			buf := make([]byte, r.Bufsize)
			runtime.Stack(buf, false)
			r.Logger(zerver.String(buf))
		}
	}()
	chain(req, resp)
}
