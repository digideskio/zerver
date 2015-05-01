package zerver

import (
	"bufio"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/cosiner/gohper/lib/errors"
	"github.com/cosiner/gohper/resource"
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
		Wrap(ResponseWrapper)

		// These methods should be called before Write
		SetHeader(name, value string)
		AddHeader(name, value string)
		RemoveHeader(name string)

		SetContentEncoding(enc string)

		// SetContentType will cause Resource change
		SetContentType(typ string)

		SetAdvancedCookie(c *http.Cookie)
		SetCookie(name, value string, lifetime int)
		DeleteClientCookie(name string)

		CacheSeconds(secs int)
		CacheUntil(*time.Time)
		NoCache()

		Status() int
		// ReportStatus report status code, it will not immediately write the status
		// to response, unless response is destroyed or Write was called
		ReportStatus(statusCode int)
		StatusResponse

		http.Hijacker
		http.Flusher

		// Write will automicly write http status and header, any operations about
		// status and header should be performed before Write
		// if there is a previous error, operation will not be performed, just
		// return this error, and any error will be stored
		ErrorWriter
		WriteString(s string) (int, error)

		// Value/SetValue provide a approach to transmit value between filter/handler
		// there is only one instance, if necessary, save origin value first,
		// restore it when operation is done
		Value() interface{}
		SetValue(interface{})

		Resource() resource.Resource
		// Send send marshaled value to client
		Send(string, interface{}) error

		destroy()
	}

	ErrorWriter interface {
		io.Writer
		ClearError()
	}

	// response represent a response of request to user
	response struct {
		env Enviroment
		res resource.Resource
		http.ResponseWriter
		header       http.Header
		status       int
		statusWrited bool
		value        interface{}
		err          error
		needClose    bool

		hijacked bool
	}
)

// newResponse create a new response, and set default content type to HTML
func (resp *response) init(env Enviroment, r resource.Resource, w http.ResponseWriter) Response {
	resp.env = env
	resp.res = r
	resp.ResponseWriter = w
	resp.header = w.Header()
	resp.status = http.StatusOK
	return resp
}

func (resp *response) destroy() {
	resp.flushHeader()
	resp.statusWrited = false
	resp.header = nil
	resp.value = nil
	resp.err = nil
	if resp.needClose && !resp.hijacked {
		resp.needClose = false
		resp.ResponseWriter.(io.Closer).Close()
	}
	resp.hijacked = false
	resp.ResponseWriter = nil
}

func (resp *response) Wrap(fn ResponseWrapper) {
	resp.ResponseWriter, resp.needClose = fn(resp.ResponseWriter, resp.needClose)
	resp.header = resp.ResponseWriter.Header()
}

func (resp *response) flushHeader() {
	if !resp.statusWrited {
		resp.WriteHeader(resp.status)
		resp.statusWrited = true
	}
}

func (resp *response) Write(data []byte) (i int, err error) {
	resp.flushHeader()
	// inspired by official go blog "Errors are value: http://blog.golang.org/errors-are-values"
	if err = resp.err; err == nil {
		i, err = resp.ResponseWriter.Write(data)
		resp.err = err
	}
	return
}

func (resp *response) WriteString(s string) (int, error) {
	return resp.Write(Bytes(s))
}

func (resp *response) ClearError() {
	resp.err = nil
}

// ReportStatus report an http status with given status code
// only when response is not destroyed and Write was not called the status will
// be changed
func (resp *response) ReportStatus(statusCode int) {
	if !resp.statusWrited {
		resp.status = statusCode
	}
}

func (resp *response) Status() int {
	return resp.status
}

// Hijack hijack response connection
func (resp *response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, is := resp.ResponseWriter.(http.Hijacker); is {
		resp.hijacked = true
		return hijacker.Hijack()
	}
	return nil, nil, ErrHijack
}

// Flush flush response's output
func (resp *response) Flush() {
	if flusher, is := resp.ResponseWriter.(http.Flusher); is {
		flusher.Flush()
	}
}

// SetHeader setup response header
func (resp *response) SetHeader(name, value string) {
	resp.header.Set(name, value)
}

// AddHeader add a value to response header
func (resp *response) AddHeader(name, value string) {
	resp.header.Add(name, value)
}

// RemoveHeader remove response header by name
func (resp *response) RemoveHeader(name string) {
	resp.header.Del(name)
}

func (resp *response) SetContentType(typ string) {
	resp.res = resp.env.ResourceMaster().Resource(typ)
	resp.SetHeader(HEADER_CONTENTTYPE, typ)
}

// SetContentEncoding set content encoding of response
func (resp *response) SetContentEncoding(enc string) {
	resp.SetHeader(HEADER_CONTENTENCODING, enc)
}

func (resp *response) CacheSeconds(secs int) {
	resp.SetHeader(HEADER_CACHECONTROL, "max-age:"+strconv.Itoa(secs))
}

func (resp *response) CacheUntil(t *time.Time) {
	resp.SetHeader(HEADER_EXPIRES, t.Format(http.TimeFormat))
}

func (resp *response) NoCache() {
	resp.SetHeader(HEADER_CACHECONTROL, "no-cache")
}

func (resp *response) SetAdvancedCookie(c *http.Cookie) {
	resp.AddHeader(HEADER_SETCOOKIE, c.String())
}

// SetCookie setup response cookie
func (resp *response) SetCookie(name, value string, lifetime int) {
	resp.SetAdvancedCookie(&http.Cookie{
		Name:   name,
		Value:  value,
		MaxAge: lifetime,
	})
}

// DeleteClientCookie delete user browser's cookie by name
func (resp *response) DeleteClientCookie(name string) {
	resp.SetCookie(name, "", -1)
}

func (resp *response) Value() interface{} {
	return resp.value
}

func (resp *response) SetValue(v interface{}) {
	resp.value = v
}

func (resp *response) Resource() resource.Resource {
	return resp.res
}

func (resp *response) Send(key string, value interface{}) error {
	if resp.res == nil {
		resp.env.Logger().Panicln("There is no resource type match this request")
	}
	return resp.res.Send(resp, key, value)
}
