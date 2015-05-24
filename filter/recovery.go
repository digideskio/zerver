package filter

import (
	"bytes"
	"fmt"
	"runtime"

	"github.com/cosiner/gohper/defval"
	"github.com/cosiner/gohper/unsafe2"
	"github.com/cosiner/ygo/log"
	"github.com/cosiner/zerver"
)

type Recovery struct {
	Bufsize int
	NoStack bool
	logger  log.Logger
}

func (r *Recovery) Init(env zerver.Environment) error {
	r.logger = env.Logger().Prefix("[Recovery]")
	defval.Int(&r.Bufsize, 1024*4)
	return nil
}

func (r *Recovery) Destroy() {}

func (r *Recovery) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	defer func() {
		e := recover()
		if e == nil || r.NoStack {
			return
		}

		resp.ReportInternalServerError()

		buffer := bytes.NewBuffer(make([]byte, 0, r.Bufsize))
		fmt.Fprint(buffer, e)
		buf := buffer.Bytes()

		runtime.Stack(buf[len(buf):cap(buf)], false)

		req.Logger().Errorln(unsafe2.String(buf[:cap(buf)]))
	}()

	chain(req, resp)
}
