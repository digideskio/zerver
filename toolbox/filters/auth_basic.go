package filters

import (
	"strings"

	"github.com/cosiner/zerver"
)

type BasicAuth struct {
	AuthUserAttrName string
	AuthPassAttrName string
}

func (b *BasicAuth) Init(zerver.Enviroment) error {
	if b.AuthUserAttrName == "" {
		b.AuthUserAttrName = "AuthUser"
	}
	if b.AuthPassAttrName == "" {
		b.AuthPassAttrName = "AuthPass"
	}
	return nil
}

func (b *BasicAuth) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	auth := req.Authorization()
	index := strings.IndexByte(auth, ':')
	if index > 0 && index < len(auth)-1 {
		req.SetAttr(b.AuthUserAttrName, auth[:index])
		req.SetAttr(b.AuthPassAttrName, auth[index+1:])
		chain(req, resp)
		return
	}
	resp.ReportBadRequest()
}

func (b *BasicAuth) Destroy() {}
