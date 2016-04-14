package zerver

import (
	"encoding/base64"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/cosiner/gohper/errors"
	"github.com/cosiner/gohper/utils/attrs"
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
		patternKeeper

		ReqMethod() string
		URL() *url.URL
		GetHeader(name string) string
		RemoteAddr() string
		Authorization() (string, bool)

		Vars() *ReqVars
		attrs.Attrs
		Env
		io.Reader

		Receive(interface{}) error
		destroy()
	}

	// request represent an income request
	request struct {
		patternString
		Env
		attrs.Attrs
		*http.Request

		vars      *ReqVars
		needClose bool
	}
)

var (
	emptyParams = make(url.Values)
)

// newRequest create a new request
func (req *request) init(e Env, requ *http.Request, pattern string, reqVars *ReqVars) Request {
	req.patternString = patternString(pattern)
	req.Env = e
	req.Request = requ

	requ.ParseForm()
	reqVars.queryVars = requ.Form
	reqVars.formVars = requ.PostForm
	req.vars = reqVars

	method := requ.Method
	if method == METHOD_POST {
		if m := requ.Header.Get(HEADER_METHODOVERRIDE); m != "" {
			method = m
		}
	}
	requ.Method = MethodName(method)
	return req
}

func (req *request) destroy() {
	req.Attrs.Clear()
	req.Env = nil
	req.vars = nil

	if req.needClose {
		req.needClose = false
		req.Body.Close()
	}
	req.Request = nil
}

func (req *request) Wrap(fn RequestWrapper) {
	req.Request, req.needClose = fn(req.Request, req.needClose)
	req.Method = MethodName(req.Method)
}
func (req *request) ReqMethod() string {
	return req.Method
}

// URL return request url
func (req *request) URL() *url.URL {
	return req.Request.URL
}

// Header return header value with name
func (req *request) GetHeader(name string) string {
	return req.Header.Get(name)
}

func (req *request) RemoteAddr() string {
	return req.Request.RemoteAddr
}

func (req *request) Vars() *ReqVars {
	return req.vars
}

func (req *request) Authorization() (string, bool) {
	basic, auth := false, req.GetHeader(HEADER_AUTHRIZATION)
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

func (req *request) Read(data []byte) (int, error) {
	return req.Body.Read(data)
}

func (req *request) Receive(v interface{}) error {
	return req.Codec().Decode(req, v)
}
