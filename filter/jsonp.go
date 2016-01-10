package filter

import (
	"bytes"
	"net/http"

	"github.com/cosiner/gohper/errors"
	"github.com/cosiner/gohper/io2"
	log "github.com/cosiner/ygo/jsonlog"
	"github.com/cosiner/zerver"
	"github.com/cosiner/zerver/utils/wrap"
)

// use as callback parameter name such as ?callback=xxx
type JSONP struct {
	CallbackVar string
	log         *log.Logger
}

func (j *JSONP) Init(zerver.Env) error {
	if j.CallbackVar == "" {
		return errors.Err("callback name should not be empty")
	}
	j.log = log.Derive("Filter", "JSONP")
	return nil
}

func (j *JSONP) Destroy() {}

func (j *JSONP) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	if req.ReqMethod() != zerver.METHOD_GET {
		chain(req, resp)
		return
	}

	callback := req.Vars().QueryVar(j.CallbackVar)
	if callback == "" {
		chain(req, resp)
		return
	}

	buffer := bytes.NewBuffer(make([]byte, 0, 256))
	bw := wrap.BuffRespWriter{ // to avoid write header 200 first when write callback name
		Buffer: buffer,
	}
	resp.Wrap(func(w http.ResponseWriter, shouldClose bool) (http.ResponseWriter, bool) {
		bw.ResponseWriter = w
		bw.ShouldClose = shouldClose
		return &bw, shouldClose
	})
	chain(req, resp)
	bw.Buffer = nil

	_, err := io2.WriteString(resp, callback)
	if err == nil {
		_, err = io2.WriteString(resp, "(")
		if err == nil {
			_, err = resp.Write(buffer.Bytes())
			if err == nil {
				_, err = io2.WriteString(resp, ")")
			}
		}
	}
	if err != nil {
		j.log.Warn(log.M{"msg": "write jsonp response failed", "err": err.Error()})
	}
}
