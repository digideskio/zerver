package zerver

import (
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/cosiner/gohper/errors"
	"github.com/cosiner/gohper/testing2"
)

func TestCompile(t *testing.T) {
	tt := testing2.Wrap(t)
	tt.Log(len(strings.Split("", "/")))
	tt.Log(compile("/"))
	tt.Log(compile("/:user/:id/:a"))
	tt.Log(compile("/user/:a/:abc/"))
	tt.Log(compile("/user/:/:abc/"))
}

// routes is copy from github.com/julienschmidt/go-http-routing-benchmark
func rt() *router {
	node := &router{}
	node.addPath("/user/i")
	node.addPath("/user/ie")
	node.addPath("/user/ief")
	node.addPath("/user/ieg")
	node.addPath("/title/|")
	node.addPath("/title/id/|")
	node.addPath("/title/i/|")
	node.addPath("/title/id/12")
	node.addPath("/ti/id/12")
	node.addPath("/ti/|/12")

	// OAuth Authorizations
	node.addPath("/authorizations/|")
	node.addPath("/authorizations")

	node.addPath("/authorizations/|")
	node.addPath("/applications/|/tokens/|")
	node.addPath("/applications/|/tokens")
	node.addPath("/applications/|/tokens/|")
	// Activity
	node.addPath("/events")
	node.addPath("/repos/|/|/events")
	node.addPath("/networks/|/|/events")
	node.addPath("/orgs/|/events")
	node.addPath("/users/|/received_events")
	node.addPath("/users/|/received_events/public")
	node.addPath("/users/|/events")
	node.addPath("/users/|/events/public")
	node.addPath("/users/|/events/orgs/|")
	node.addPath("/feeds")
	node.addPath("/notifications")
	node.addPath("/repos/|/|/notifications")
	node.addPath("/notifications")
	node.addPath("/repos/|/|/notifications")
	node.addPath("/notifications/threads/|")

	node.addPath("/notifications/threads/|/subscription")
	node.addPath("/notifications/threads/|/subscription")
	node.addPath("/notifications/threads/|/subscription")
	node.addPath("/repos/|/|/stargazers")
	node.addPath("/users/|/starred")
	node.addPath("/user/starred")
	node.addPath("/user/starred/|/|")
	node.addPath("/user/starred/|/|")
	node.addPath("/user/starred/|/|")
	node.addPath("/repos/|/|/subscribers")
	node.addPath("/users/|/subscriptions")
	node.addPath("/user/subscriptions")
	node.addPath("/repos/|/|/subscription")
	node.addPath("/repos/|/|/subscription")
	node.addPath("/repos/|/|/subscription")
	node.addPath("/user/subscriptions/|/|")
	node.addPath("/user/subscriptions/|/|")
	node.addPath("/user/subscriptions/|/|")
	// Gists
	node.addPath("/users/|/gists")
	node.addPath("/gists")
	node.addPath("/gists/public")
	node.addPath("/gists/starred")
	node.addPath("/gists/|")
	node.addPath("/gists")

	node.addPath("/gists/|/star")
	node.addPath("/gists/|/star")
	node.addPath("/gists/|/star")
	node.addPath("/gists/|/forks")
	node.addPath("/gists/|")
	// Git Data
	node.addPath("/repos/|/|/git/blobs/|")
	node.addPath("/repos/|/|/git/blobs")
	node.addPath("/repos/|/|/git/commits/|")
	node.addPath("/repos/|/|/git/commits")
	node.addPath("/repos/|/|/git/refs/|ref")
	node.addPath("/repos/|/|/git/refs")
	node.addPath("/repos/|/|/git/refs")

	node.addPath("/repos/|/|/git/tags/|")
	node.addPath("/repos/|/|/git/tags")
	node.addPath("/repos/|/|/git/trees/|")
	node.addPath("/repos/|/|/git/trees")
	// Issues
	node.addPath("/issues")
	node.addPath("/user/issues")
	node.addPath("/orgs/|/issues")
	node.addPath("/repos/|/|/issues")
	node.addPath("/repos/|/|/issues/|")
	node.addPath("/repos/|/|/issues")

	node.addPath("/repos/|/|/assignees")
	node.addPath("/repos/|/|/assignees/|")
	node.addPath("/repos/|/|/issues/|/comments")
	node.addPath("/repos/|/|/issues/comments")
	node.addPath("/repos/|/|/issues/comments/|")
	node.addPath("/repos/|/|/issues/|/comments")

	node.addPath("/repos/|/|/issues/|/events")
	node.addPath("/repos/|/|/issues/events")
	node.addPath("/repos/|/|/issues/events/|")
	node.addPath("/repos/|/|/labels")
	node.addPath("/repos/|/|/labels/|")
	node.addPath("/repos/|/|/labels")

	node.addPath("/repos/|/|/labels/|")
	node.addPath("/repos/|/|/issues/|/labels")
	node.addPath("/repos/|/|/issues/|/labels")
	node.addPath("/repos/|/|/issues/|/labels/|")
	node.addPath("/repos/|/|/issues/|/labels")
	node.addPath("/repos/|/|/issues/|/labels")
	node.addPath("/repos/|/|/milestones/|/labels")
	node.addPath("/repos/|/|/milestones")
	node.addPath("/repos/|/|/milestones/|")
	node.addPath("/repos/|/|/milestones")

	node.addPath("/repos/|/|/milestones/|")
	// Miscellaneous
	node.addPath("/emojis")
	node.addPath("/gitignore/templates")
	node.addPath("/gitignore/templates/|")
	node.addPath("/markdown")
	node.addPath("/markdown/raw")
	node.addPath("/meta")
	node.addPath("/rate_limit")
	// Organizations
	node.addPath("/users/|/orgs")
	node.addPath("/user/orgs")
	node.addPath("/orgs/|")

	node.addPath("/orgs/|/members")
	node.addPath("/orgs/|/members/|")
	node.addPath("/orgs/|/members/|")
	node.addPath("/orgs/|/public_members")
	node.addPath("/orgs/|/public_members/|")
	node.addPath("/orgs/|/public_members/|")
	node.addPath("/orgs/|/public_members/|")
	node.addPath("/orgs/|/teams")
	node.addPath("/teams/|")
	node.addPath("/orgs/|/teams")

	node.addPath("/teams/|")
	node.addPath("/teams/|/members")
	node.addPath("/teams/|/members/|")
	node.addPath("/teams/|/members/|")
	node.addPath("/teams/|/members/|")
	node.addPath("/teams/|/repos")
	node.addPath("/teams/|/repos/|/|")
	node.addPath("/teams/|/repos/|/|")
	node.addPath("/teams/|/repos/|/|")
	node.addPath("/user/teams")
	// Pull Requests
	node.addPath("/repos/|/|/pulls")
	node.addPath("/repos/|/|/pulls/|")
	node.addPath("/repos/|/|/pulls")

	node.addPath("/repos/|/|/pulls/|/commits")
	node.addPath("/repos/|/|/pulls/|/files")
	node.addPath("/repos/|/|/pulls/|/merge")
	node.addPath("/repos/|/|/pulls/|/merge")
	node.addPath("/repos/|/|/pulls/|/comments")
	node.addPath("/repos/|/|/pulls/comments")
	node.addPath("/repos/|/|/pulls/comments/|")
	node.addPath("/repos/|/|/pulls/|/comments")

	// Repositories
	node.addPath("/user/repos")
	node.addPath("/users/|/repos")
	node.addPath("/orgs/|/repos")
	node.addPath("/repositories")
	node.addPath("/user/repos")
	node.addPath("/orgs/|/repos")
	node.addPath("/repos/|/|")

	node.addPath("/repos/|/|/contributors")
	node.addPath("/repos/|/|/languages")
	node.addPath("/repos/|/|/teams")
	node.addPath("/repos/|/|/tags")
	node.addPath("/repos/|/|/branches")
	node.addPath("/repos/|/|/branches/|")
	node.addPath("/repos/|/|")
	node.addPath("/repos/|/|/collaborators")
	node.addPath("/repos/|/|/collaborators/|")
	node.addPath("/repos/|/|/collaborators/|")
	node.addPath("/repos/|/|/collaborators/|")
	node.addPath("/repos/|/|/comments")
	node.addPath("/repos/|/|/commits/|/comments")
	node.addPath("/repos/|/|/commits/|/comments")
	node.addPath("/repos/|/|/comments/|")

	node.addPath("/repos/|/|/comments/|")
	node.addPath("/repos/|/|/commits")
	node.addPath("/repos/|/|/commits/|")
	node.addPath("/repos/|/|/readme")
	node.addPath("/repos/|/|/contents/|path")

	node.addPath("/repos/|/|/|/|")
	node.addPath("/repos/|/|/keys")
	node.addPath("/repos/|/|/keys/|")
	node.addPath("/repos/|/|/keys")

	node.addPath("/repos/|/|/keys/|")
	node.addPath("/repos/|/|/downloads")
	node.addPath("/repos/|/|/downloads/|")
	node.addPath("/repos/|/|/downloads/|")
	node.addPath("/repos/|/|/forks")
	node.addPath("/repos/|/|/forks")
	node.addPath("/repos/|/|/hooks")
	node.addPath("/repos/|/|/hooks/|")
	node.addPath("/repos/|/|/hooks")

	node.addPath("/repos/|/|/hooks/|/tests")
	node.addPath("/repos/|/|/hooks/|")
	node.addPath("/repos/|/|/merges")
	node.addPath("/repos/|/|/releases")
	node.addPath("/repos/|/|/releases/|")
	node.addPath("/repos/|/|/releases")

	node.addPath("/repos/|/|/releases/|")
	node.addPath("/repos/|/|/releases/|/assets")
	node.addPath("/repos/|/|/stats/contributors")
	node.addPath("/repos/|/|/stats/commit_activity")
	node.addPath("/repos/|/|/stats/code_frequency")
	node.addPath("/repos/|/|/stats/participation")
	node.addPath("/repos/|/|/stats/punch_card")
	node.addPath("/repos/|/|/statuses/|")
	node.addPath("/repos/|/|/statuses/|")
	// Search
	node.addPath("/search/repositories")
	node.addPath("/search/code")
	node.addPath("/search/issues")
	node.addPath("/search/users")
	node.addPath("/legacy/issues/search/|/|/|/|")
	node.addPath("/legacy/repos/search/|")
	node.addPath("/legacy/user/search/|")
	node.addPath("/legacy/user/email/|")
	// Users
	node.addPath("/users/|")
	node.addPath("/user")

	node.addPath("/users")
	node.addPath("/user/emails")
	node.addPath("/user/emails")
	node.addPath("/user/emails")
	node.addPath("/users/|/followers")
	node.addPath("/user/followers")
	node.addPath("/users/|/following")
	node.addPath("/user/following")
	node.addPath("/user/following/|")
	node.addPath("/users/|/following/|")
	node.addPath("/user/following/|")
	node.addPath("/user/following/|")
	node.addPath("/users/|/keys")
	node.addPath("/user/keys")
	node.addPath("/user/keys/|")
	node.addPath("/user/keys")

	node.addPath("/user/keys/|")
	return node
}

var r = rt()

func BenchmarkMatchRouteOne(b *testing.B) {
	// tt := testing2.Wrap(b)
	path := "/repos/cosiner/zerver/stargazers"
	// path := "/user/repos"
	// path := "/user/keys"
	// path := "/user/aa/exist"
	for i := 0; i < b.N; i++ {
		// pathIndex := 0
		// var vars []string
		// var continu = true
		// n := r
		// for continu {
		// 	pathIndex, vars, n, continu = n.matchMulti(path, pathIndex, vars)
		// }
		_, _ = r.matchOne(path, make([]string, 0, 2))
	}
}

func BenchmarkMatchRouteMultiple(b *testing.B) {
	// tt := testing2.Wrap(b)
	// path := "/legacy/issues/search/aaa/bbb/ccc/ddd"
	// path := "/user/repos"
	path := "/repos/cosiner/zerver/stargazers"
	// path := "/user/aa/exist"
	for i := 0; i < b.N; i++ {
		pathIndex := 0
		var vars = make([]string, 0, 2)
		var continu = true
		n := r
		for continu {
			pathIndex, vars, n, continu = n.matchMultiple(path, pathIndex, vars)
		}
		if n == nil {
			b.Fail()
		}
	}
}

func TestRoute(t *testing.T) {
	rt := new(router)
	errors.Fatal(rt.Handle("/user.:format", MapHandler{}))
	errors.Fatal(rt.Handle("/v:version", MapHandler{}))
	errors.Fatal(rt.Handle("/vaa/:id", MapHandler{}))
	// errors.Fatal(rt.Handle("/vba/:id", EmptyHandlerFunc))
	// errors.Fatal(rt.Handle("/v0a/:id", EmptyHandlerFunc))
	rt.PrintRouteTree(os.Stdout)
	_, value := rt.matchOne("/user.json", nil)
	t.Log(value)
	rt, value = rt.matchOne("/vbc", nil)
	testing2.True(t, rt != nil)
	t.Log(value)
}

func TestFilterHideHandler(t *testing.T) {
	tt := testing2.Wrap(t)
	rt := NewRouter()
	rt.Handle("/user/12:id", EmptyFilterFunc)
	rt.HandleFunc("/user/:id", GET, EmptyHandlerFunc)
	h, _, _ := rt.MatchHandlerFilters(&url.URL{Path: "/user/1234"})
	tt.True(h == nil)
	h, _, _ = rt.MatchHandlerFilters(&url.URL{Path: "/user/2234"})
	tt.True(h != nil)
}

func TestConflictWildcardCatchall(t *testing.T) {
	tt := testing2.Wrap(t)
	rt := new(router)
	tt.True(rt.HandleFunc("/:user/:id", GET, EmptyHandlerFunc) == nil)
	e := rt.HandleFunc("/*user", GET, EmptyHandlerFunc)
	tt.True(e != nil)
}

func TestSubRouter(t *testing.T) {
	tt := testing2.Wrap(t)
	userRt := NewRouter()
	userRt.HandleFunc("/info/:id", GET, EmptyHandlerFunc)
	userRt.HandleFunc("/posts/:id", GET, EmptyHandlerFunc)

	bookRt := NewRouter()
	bookRt.HandleFunc("/info/:id", GET, EmptyHandlerFunc)
	bookRt.HandleFunc("/count/:name", GET, EmptyHandlerFunc)

	rt := new(router)
	rt.Handle("/user", userRt)
	rt.Handle("/book", bookRt)

	tt.True(rt.matchOnly("/user/info/123") != nil)
	tt.True(rt.matchOnly("/bkko/info/123") == nil)
}
