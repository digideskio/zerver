package filter

import (
	"github.com/cosiner/gohper/lib/errors"
	"github.com/cosiner/zerver"
)

// use as callback parameter name such as ?callback=xxx
type JSONP string

func (j JSONP) Init(zerver.Enviroment) error {
	if string(j) == "" {
		return errors.Err("callback name should not be empty")
	}
	return nil
}

func (j JSONP) Destroy() {}

func (j JSONP) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	var (
		end string
		err error
	)
	if req.Method() == "GET" {
		callback := req.Param(string(j))
		if callback != "" {
			_, err = resp.WriteString(callback)
			if err == nil {
				_, err = resp.WriteString("(")

			}
			end = ")"
		}
	}
	chain(req, resp)
	if end != "" {
		_, err = resp.WriteString(end)
	}
	if err != nil {
		req.Server().Errorln(err)
	}
}
