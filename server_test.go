package zerver

import (
	"strings"
	"testing"
	"time"

	"github.com/cosiner/gohper/testing2"
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
	testing2.True(t, strings.Contains(err.Error(), "closed"))
}
