package zerver

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/cosiner/gohper/testing2"
	"github.com/cosiner/ygo/resource"
)

type MockWriter struct {
	Status  int
	Headers http.Header
	io.Writer
}

func NewMockWriter(w io.Writer) *MockWriter {
	return &MockWriter{
		Headers: make(http.Header),
		Writer:  w,
	}
}

func (m *MockWriter) WriteHeader(status int) {
	m.Status = status
}

func (m *MockWriter) Header() http.Header {
	return m.Headers
}

func TestFilter(t *testing.T) {
	tt := testing2.Wrap(t)
	s := NewServer()

	var n int
	ft := func(req Request, resp Response, chain FilterChain) {
		n++
		chain(req, resp)
	}

	s.Handle("/", ft)
	s.Handle("/user", ft)
	s.Handle("/user/info", ft)

	s.Get("/user/info/aaa", Intercept(func(Request, Response) {
		n++
	}, ft, ft, ft))

	s.ResMaster.DefUse(resource.RES_JSON, resource.JSON{})

	s.ServeHTTP(NewMockWriter(os.Stdout), &http.Request{
		Method: "Get",
		URL: &url.URL{
			Path: "/user/info/aaa",
		},
	})

	tt.Eq(7, n)
}
