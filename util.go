package zerver

import "strings"

const (
	// Http Header
	HEADER_CONTENTTYPE     = "Content-Type"
	HEADER_CONTENTLENGTH   = "Content-Length"
	HEADER_SETCOOKIE       = "Set-Cookie"
	HEADER_REFER           = "Referer"
	HEADER_CONTENTENCODING = "Content-Encoding"
	HEADER_USERAGENT       = "User-Agent"
	HEADER_ACCEPT          = "Accept"
	HEADER_ACCEPTENCODING  = "Accept-Encoding"
	HEADER_CACHECONTROL    = "Cache-Control"
	HEADER_EXPIRES         = "Expires"
	HEADER_AUTHRIZATION    = "Authorization"
	HEADER_METHODOVERRIDE  = "X-HTTP-Method-Override"
	HEADER_REALIP          = "X-Real-IP"

	// ContentEncoding
	ENCODING_GZIP    = "gzip"
	ENCODING_DEFLATE = "deflate"

	// Request Method
	METHOD_GET = "GET"
	METHOD_POST = "POST"
	METHOD_DELETE = "DELETE"
	METHOD_PUT = "PUT"
	METHOD_PATCH = "PATCH"
	METHOD_HEAD = "HEAD"
	METHOD_OPTIONS = "OPTIONS"
)

func MethodName(s string) string {
	if s == "" {
		return METHOD_GET
	}

	return strings.ToUpper(s)
}

type errResp struct {
	Error interface{} `json:"error"`
}

var NewError = func(s interface{}) interface{} {
	return errResp{s}
}
