package filter

import (
	"net/http"

	"github.com/cosiner/gohper/errors"
	"github.com/cosiner/gohper/io2"
	"github.com/cosiner/zerver"
	"github.com/ngaut/log"
)

// use as callback parameter name such as ?callback=xxx
type JSONP string

func (j JSONP) Init(zerver.Env) error {
	if string(j) == "" {
		return errors.Err("callback name should not be empty")
	}
	return nil
}

func (j JSONP) Destroy() {}

func (j JSONP) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	if req.ReqMethod() != "GET" {
		chain(req, resp)
		return
	}

	callback := req.Vars().QueryVar(string(j))
	if callback == "" {
		resp.StatusCode(http.StatusBadRequest)
		return
	}

	_, err := io2.WriteString(resp, callback)
	if err != nil {
		goto ERROR
	}
	_, err = io2.WriteString(resp, "(")
	if err != nil {
		goto ERROR
	}
	chain(req, resp)
	_, err = io2.WriteString(resp, ")")
	if err == nil {
		return
	}
ERROR:
	log.Warn("write string", err)
}
