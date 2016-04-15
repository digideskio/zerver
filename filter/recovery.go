package filter

import (
	"net/http"

	"github.com/cosiner/gohper/runtime2"
	"github.com/cosiner/gohper/utils/defval"
	log "github.com/cosiner/ygo/jsonlog"
	"github.com/cosiner/zerver"
)

type Recovery struct {
	Bufsize int
	log     *log.Logger
}

func (r *Recovery) Init(env zerver.Env) error {
	defval.Int(&r.Bufsize, 1024*4)
	r.log = log.Derive("Filter", "Recovery")
	return nil
}

func (r *Recovery) Destroy() {}

func (r *Recovery) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	defer func() {
		if err := recover(); err != nil {
			stack := runtime2.Stack(r.Bufsize, false)
			resp.StatusCode(http.StatusInternalServerError)
			r.log.Raw(0, log.LEVEL_ERROR, string(stack))
			return
		}
	}()
	chain(req, resp)
}
