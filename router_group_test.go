package zerver

import (
	"net/url"
	"os"
	"testing"

	"github.com/cosiner/gohper/lib/test"
)

var grt = func() Router {
	r := NewRouter()
	gr := NewGroupRouter(r, "/user")
	gr.Get("/:id", EmptyHandlerFunc)
	gr.Post("/:id", EmptyHandlerFunc)
	gr.Delete("/:id", EmptyHandlerFunc)
	gr.Post("/:id/exist", EmptyHandlerFunc)
	return r
}()

var u = &url.URL{Path: "/user/123/exist"}

func TestGroup(t *testing.T) {
	tt := test.Wrap(t)
	grt.PrintRouteTree(os.Stdout)
	h, _, _ := grt.MatchHandlerFilters(u)
	tt.True(h != nil)
}

func BenchmarkRouter(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _, _ = grt.MatchHandlerFilters(u)
	}
}
