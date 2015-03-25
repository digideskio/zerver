package zerver

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	. "github.com/cosiner/gohper/lib/errors"
)

type (
	Response interface {
		// These methods should be called before Write
		SetHeader(name, value string)
		AddHeader(name, value string)
		RemoveHeader(name string)
		SetContentEncoding(enc string)
		SetContentType(typ string)
		SetAdvancedCookie(c *http.Cookie)
		CacheSeconds(secs int)
		CacheUntil(*time.Time)
		NoCache()

		// SetCookie(name, value string, lifetime int)
		// SetSecureCookie(name, value string, lifetime int)
		// DeleteClientCookie(name string)
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
		io.Writer
		// ClearError clear error stored in response
		ClearError()

		destroy()
		// Value/SetValue provide a approach to transmit value between filter/handler
		// there is only one instance, if necessary first save origin value, after
		// operation, restore it
		Value() interface{}
		SetValue(interface{})
	}

	// response represent a response of request to user
	response struct {
		http.ResponseWriter
		header       http.Header
		status       int
		statusWrited bool
		value        interface{}
		err          error
	}
)

const (
	ErrHijack = Err("Connection not support hijack")
)

// newResponse create a new response, and set default content type to HTML
func (resp *response) init(w http.ResponseWriter) Response {
	resp.ResponseWriter = w
	resp.header = w.Header()
	resp.status = http.StatusOK
	return resp
}

func (resp *response) destroy() {
	resp.flushHeader()
	resp.statusWrited = false
	resp.ResponseWriter = nil
	resp.header = nil
	resp.err = nil
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

// SetContentType set content type of response
func (resp *response) SetContentType(typ string) {
	resp.SetHeader(HEADER_CONTENTTYPE, typ)
}

// SetContentEncoding set content encoding of response
func (resp *response) SetContentEncoding(enc string) {
	resp.SetHeader(HEADER_CONTENTENCODING, enc)
}

func (resp *response) CacheSeconds(secs int) {
	resp.SetHeader(HEADER_CACHECONTROL, fmt.Sprintf("max-age:%d", secs))
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

// // SetCookie setup response cookie
// func (resp *response) SetCookie(name, value string, lifetime int) {
// 	resp.SetAdvancedCookie(&http.Cookie{
// 		Name:   name,
// 		Value:  value,
// 		MaxAge: lifetime,
// 	})
// }

// // SetSecureCookie setup response cookie with secureity
// func (resp *response) SetSecureCookie(name, value string, lifetime int) {
// 	resp.SetCookie(name, value, lifetime)
// }

// // DeleteClientCookie delete user briwser's cookie by name
// func (resp *response) DeleteClientCookie(name string) {
// 	resp.SetCookie(name, "", -1)
// }

func (resp *response) Value() interface{} {
	return resp.value
}

func (resp *response) SetValue(v interface{}) {
	resp.value = v
}
