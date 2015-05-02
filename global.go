package zerver

import (
	"github.com/cosiner/gohper/resource"

	"strings"
)

const (
	// Http Header
	HEADER_CONTENTTYPE     = "Content-Type"
	HEADER_CONTENTLENGTH   = "Content-Length"
	HEADER_SETCOOKIE       = "Set-Cookie"
	HEADER_REFER           = "Referer"
	HEADER_CONTENTENCODING = "Content-Encoding"
	HEADER_USERAGENT       = "User-Agent"
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
	GET     = "GET"
	POST    = "POST"
	DELETE  = "DELETE"
	PUT     = "PUT"
	PATCH   = "PATCH"
	HEAD    = "HEAD"
	OPTIONS = "OPTIONS"
)

// parseRequestMethod convert a string to request method, default use GET
// if string is empty
func parseRequestMethod(s string) string {
	if s == "" {
		return GET
	}
	return strings.ToUpper(s)
}

// parseContentType parse content type
func parseContentType(str string) string {
	if str == "" {
		return resource.CONTENTTYPE_JSON
	}
	return strings.ToLower(strings.TrimSpace(str))
}
