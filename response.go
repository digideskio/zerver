package zerver

import (
	"bufio"
	"io"
	"net"
	"net/http"

	"github.com/cosiner/gohper/errors"
)

const (
	ErrHijack = errors.Err("Connection not support hijack")
)

type (
	// ResponseWrapper wrap response writer, return another writer and a flag specified
	// whether should close writer on response destroy, the returned writer must be
	// io.Closer for it should close original writer
	ResponseWrapper func(http.ResponseWriter, bool) (http.ResponseWriter, bool)

	Response interface {
		Env
		http.Hijacker
		http.Flusher
		io.Writer

		Wrap(ResponseWrapper)
		Headers() http.Header
		StatusCode(statusCode int) int
		Value() interface{}
		SetValue(interface{})
		Send(interface{}) error

		destroy()
	}

	// response represent a response of request to user
	response struct {
		Env
		http.ResponseWriter
		status       int
		statusWrited bool
		value        interface{}
		needClose    bool

		hijacked bool
	}
)

// newResponse create a new response, and set default content type to HTML
func (resp *response) init(env Env, w http.ResponseWriter) Response {
	resp.Env = env
	resp.ResponseWriter = w
	resp.status = http.StatusOK

	return resp
}

func (resp *response) destroy() {
	resp.flushHeader()
	resp.statusWrited = false
	resp.value = nil

	if resp.needClose && !resp.hijacked {
		resp.needClose = false
		resp.ResponseWriter.(io.Closer).Close()
	}
	resp.hijacked = false
	resp.ResponseWriter = nil
}

func (resp *response) Wrap(fn ResponseWrapper) {
	resp.ResponseWriter, resp.needClose = fn(resp.ResponseWriter, resp.needClose)
}

func (resp *response) flushHeader() {
	if !resp.statusWrited {
		resp.WriteHeader(resp.status)
		resp.statusWrited = true
	}
}

func (resp *response) StatusCode(statusCode int) int {
	if !resp.statusWrited && statusCode > 0 {
		resp.status = statusCode
	}
	return resp.status
}

func (resp *response) Status() int {
	return resp.status
}

// Hijack hijack response connection
func (resp *response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	hijacker, is := resp.ResponseWriter.(http.Hijacker)
	if !is {
		return nil, nil, ErrHijack
	}

	resp.hijacked = true
	return hijacker.Hijack()
}

// Flush flush response's output
func (resp *response) Flush() {
	if flusher, is := resp.ResponseWriter.(http.Flusher); is {
		flusher.Flush()
	}
}

func (resp *response) Headers() http.Header {
	return resp.ResponseWriter.Header()
}

func (resp *response) Value() interface{} {
	return resp.value
}

func (resp *response) SetValue(v interface{}) {
	resp.value = v
}

func (resp *response) Write(data []byte) (i int, err error) {
	resp.flushHeader()
	return resp.ResponseWriter.Write(data)
}

func (resp *response) Send(v interface{}) error {
	return resp.Codec().Encode(resp, v)
}
