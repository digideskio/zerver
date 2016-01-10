package filter

import (
	"net/http"
	"runtime"

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
			resp.StatusCode(http.StatusInternalServerError)
			buf := make([]byte, r.Bufsize)
			n := runtime.Stack(buf, false)
			buf = buf[:n]

			r.log.Error(log.M{"msg": "recover from panic", "error": err, "stack": string(buf)})
			return
		}
	}()

	chain(req, resp)
}
