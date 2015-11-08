package zerver

import (
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/cosiner/gohper/errors"
	"github.com/cosiner/gohper/strings2"
	"github.com/cosiner/gohper/utils/attrs"
	"github.com/cosiner/ygo/log"
)

const (
	ErrNoResourceType = errors.Err("there is no this resource type on server")
)

type (
	// RequestWrapper wrap a request, then return another one and a flag specified
	// whether should close request.Body on request destroy, it should close
	// original request.Body if need
	RequestWrapper func(*http.Request, bool) (*http.Request, bool)

	Request interface {
		Wrap(RequestWrapper)

		URL() *url.URL
		Method() string

		Header(name string) string

		RemoteAddr() string
		RemoteIP() string
		UserAgent() string
		Accepts() string
		AcceptEncodings() string
		Authorization() (string, bool)
		BasicAuth() (string, string)
		Cookie(name string) string

		Param(name string) string
		Params(name string) []string

		attrs.Attrs

		Environment

		io.Reader
		URLVarIndexer

		Receive(interface{}) error
		destroy()
	}

	// request represent an income request
	request struct {
		URLVarIndexer
		Environment
		attrs.Attrs

		request   *http.Request
		method    string
		header    http.Header
		params    url.Values
		needClose bool
	}
)

var (
	emptyParams = make(url.Values)
)

// newRequest create a new request
func (req *request) init(e Environment,  requ *http.Request, varIndexer URLVarIndexer) Request {
	req.Environment = e
	req.request = requ
	req.header = requ.Header
	req.URLVarIndexer = varIndexer

	method := requ.Method
	if method == POST {
		if m := requ.Header.Get(HEADER_METHODOVERRIDE); m != "" {
			method = m
		}
	}
	req.method = parseRequestMethod(method)

	return req
}

func (req *request) destroy() {
	req.Attrs.Clear()
	req.Environment = nil
	req.header = nil
	req.URLVarIndexer.destroySelf() // who owns resource, who releases resource
	req.URLVarIndexer = nil
	req.params = nil

	if req.needClose {
		req.needClose = false
		req.request.Body.Close()
	}
	req.request = nil
}

func (req *request) Wrap(fn RequestWrapper) {
	req.request, req.needClose = fn(req.request, req.needClose)
	req.header = req.request.Header
	req.method = req.request.Method
}

func (req *request) Read(data []byte) (int, error) {
	return req.request.Body.Read(data)
}

// Method return method of request
func (req *request) Method() string {
	return req.method
}

// Cookie return cookie value with given name
func (req *request) Cookie(name string) string {
	c, err := req.request.Cookie(name)
	if err != nil {
		return ""
	}

	return c.Value
}

// RemoteAddr return remote address
func (req *request) RemoteAddr() string {
	return req.request.RemoteAddr
}

func (req *request) RemoteIP() string {
	if ip := req.Header(HEADER_REALIP); ip != "" {
		return ip
	}

	addr := req.RemoteAddr()
	return addr[:strings2.LastIndexByte(addr, ':')]
}

// Param return request parameter with name
func (req *request) Param(name string) (value string) {
	params := req.Params(name)
	if len(params) > 0 {
		value = params[0]
	}

	return
}

// Params return request parameters with name,
// it only get params of url query, don't parse request body
func (req *request) Params(name string) []string {
	params, request := req.params, req.request
	if params == nil {
		switch req.method {
		case GET, HEAD, OPTIONS:
			params = request.URL.Query()
		default:
			err := request.ParseForm()
			if err == nil {
				params = request.PostForm
			} else {
				params = emptyParams
				log.Warn("parse form", err)
			}
		}
		req.params = params
	}

	return params[name]
}

// UserAgent return user's agent identify
func (req *request) UserAgent() string {
	return req.Header(HEADER_USERAGENT)
}

// Accepts extract content type form request header
func (req *request) Accepts() string {
	return req.Header(HEADER_ACCEPT)
}

func (req *request) AcceptEncodings() string {
	return req.Header(HEADER_ACCEPTENCODING)
}

// Authorization get authorization info from request header,
// if it isn't basic auth, auth info will be directly returned,
// otherwise, base64 decode will be peformed, empty string will be returned
// if error occured when decode
//
// the bool value identity whether it is basic auth
func (req *request) Authorization() (string, bool) {
	basic, auth := false, req.Header(HEADER_AUTHRIZATION)
	if basic = strings.HasPrefix(auth, "Basic "); basic {
		a, err := base64.StdEncoding.DecodeString(auth[len("Basic "):])
		if err == nil {
			auth = string(a)
		} else {
			auth = ""
		}
	}

	return auth, basic
}

// BasicAuth used for http basic auth, the returned value will be
// user,password, if any wrong, "", "" was returned
func (req *request) BasicAuth() (string, string) {
	if auth, basic := req.Authorization(); basic {
		return strings2.Separate(auth, ':')
	}

	return "", ""
}

// URL return request url
func (req *request) URL() *url.URL {
	return req.request.URL
}

// Header return header value with name
func (req *request) Header(name string) string {
	return req.header.Get(name)
}

func (req *request) Receive(v interface{}) error {
	return req.Codec().Decode(req, v)
}
