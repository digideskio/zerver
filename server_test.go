package zerver

import (
	"strings"

	"github.com/cosiner/gohper/testing"

	"testing"
	"time"
)

func TestServer(t *testing.T) {
	s := NewServer()
	s.Get("/", func(_ Request, resp Response) {
		resp.WriteString("Hello World")
	})
	var err error
	go func(s *Server, err *error) {
		time.Sleep(time.Millisecond)
		if *err == nil {
			s.Destroy()
		}
	}(s, &err)

	err = s.Start(nil)
	test.True(t, strings.Contains(err.Error(), "closed"))
}
