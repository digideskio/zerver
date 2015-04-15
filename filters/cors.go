package filters

import (
	"github.com/cosiner/zerver"
)

const _HEADER_CORS_ALLOWSORIGIN = "Access-Control-Allow-Origin"

type CORS struct {
	Allows   []string
	allowAll bool
}

func (c *CORS) Init(zerver.Enviroment) error {
	l := len(c.Allows)
	if l == 0 || (l == 1 && c.Allows[0] == "*") {
		c.allowAll = true
		c.Allows = nil
	}
	return nil
}

func (c *CORS) Destroy() {}

func (c *CORS) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	if c.allowAll {
		resp.SetHeader(_HEADER_CORS_ALLOWSORIGIN, "*")
		chain(req, resp)
	} else {
		orgin := req.Header("Origin")
		for _, al := range c.Allows {
			if al == orgin {
				resp.SetHeader(_HEADER_CORS_ALLOWSORIGIN, orgin)
				chain(req, resp)
				return
			}
		}
		resp.ReportForbidden()
	}
}
