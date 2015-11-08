package filter

import (
	"github.com/cosiner/gohper/errors"
	"github.com/cosiner/zerver"
	"github.com/ngaut/log"
)

// use as callback parameter name such as ?callback=xxx
type JSONP string

func (j JSONP) Init(zerver.Environment) error {
	if string(j) == "" {
		return errors.Err("callback name should not be empty")
	}
	return nil
}

func (j JSONP) Destroy() {}

func (j JSONP) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	if req.Method() != "GET" {
		chain(req, resp)
		return
	}

	callback := req.Param(string(j))
	if callback == "" {
		resp.ReportBadRequest()
		return
	}

	_, err := resp.WriteString(callback)
	if err != nil {
		goto ERROR
	}
	_, err = resp.WriteString("(")
	if err != nil {
		goto ERROR
	}
	chain(req, resp)
	_, err = resp.WriteString(")")
	if err == nil {
		return
	}
ERROR:
	log.Warn("write string", err)
}
