package filter

import (
	"runtime"

	"github.com/cosiner/gohper/lib/defval"
	"github.com/cosiner/zerver"
)

type Recovery struct {
	Bufsize int
	NoStack bool
}

func (r *Recovery) Init(env zerver.Enviroment) error {
	defval.Int(&r.Bufsize, 1024*8)
	return nil
}

func (r *Recovery) Destroy() {}

func (r *Recovery) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	defer func() {
		if e := recover(); e != nil && !r.NoStack {
			resp.ReportInternalServerError()
			buf := make([]byte, r.Bufsize)
			runtime.Stack(buf, false)
			req.Logger().Warnln(zerver.String(buf))
		}
	}()
	chain(req, resp)
}
