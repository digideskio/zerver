package request

import (
	"strings"

	"github.com/cosiner/zerver"
)

var ProxyHeaders = []string{"X-Forwarded-For"}

func RemoteAddr(req zerver.Request) string {
	for _, header := range ProxyHeaders {
		addr := req.GetHeader(header)
		if addr != "" {
			return strings.Split(addr, ";")[0]
		}
	}
	return req.RemoteAddr()
}

func ParseIp(addr string) string {
	index := strings.IndexByte(addr, ':')
	if index < 0 {
		return addr
	}
	if index == 0 {
		return "127.0.0.1"
	}
	return addr[:index]
}
