package zerver

import (
	"net/http"
	"testing"
	"time"

	"github.com/cosiner/gohper/testing2"
)

func TestServerDestroyTimeout(t *testing.T) {
	tt := testing2.Wrap(t)

	var handler = func(Request, Response) {
		time.Sleep(1000 * time.Millisecond)
	}

	s := NewServer()
	s.Get("/", handler)

	go s.Start(nil)
	go func() {
		http.Get("http://localhost:4000/")
	}()

	time.Sleep(10 * time.Millisecond)
	tt.False(s.Destroy(10 * time.Millisecond))
}
