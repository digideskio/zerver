package filter

import (
	"github.com/cosiner/gohper/errors"
	"github.com/cosiner/ygo/resource"
	"github.com/cosiner/zerver"
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

	res, _ := req.ResourceMaster().Resource(resource.RES_JSON)
	if res == nil {
		resp.ReportNotAcceptable()
		return
	}

	callback := req.Param(string(j))
	if callback == "" {
		resp.ReportBadRequest()
		resp.Send("error", "no callback function")
		return
	}

	resp.SetContentType(resource.RES_JSON, res)
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
	req.Logger().Warnln(err)
}
