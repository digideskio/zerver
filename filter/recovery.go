package filter

import (
	"runtime"

	"github.com/cosiner/zerver"
)

type Recovery struct {
	Bufsize int
}

func (r *Recovery) Init(env zerver.Enviroment) error {
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
			req.Logger().Warnln(zerver.String(buf))
		}
	}()
	chain(req, resp)
}
