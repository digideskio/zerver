package filters

import (
	"fmt"
	"net/http"

	"github.com/cosiner/gohper/lib/types"
	"github.com/cosiner/zerver"
)

var Bytes = types.UnsafeBytes
var notFound = []byte("<h1>404 Not Found</h1>")
var methodNotAllowed = []byte("<h1>405 Method Not Allowed</h1>")

func ErrorToBrowserFilter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	defer func() {
		if e := recover(); e != nil {
			resp.Write(Bytes(fmt.Sprint(e)))
		}
	}()
	var ret []byte
	if status := resp.Status(); status == http.StatusNotFound {
		ret = notFound
	} else if status == http.StatusMethodNotAllowed {
		ret = methodNotAllowed
	} else {
		chain(req, resp)
		return
	}
	resp.SetContentType(zerver.CONTENTTYPE_HTML)
	resp.Write(ret)
}
