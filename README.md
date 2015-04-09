# Zerver
__Zerver__ is a simple, scalable, restful api framework for [golang](http://golang.org).

[中文介绍](http://cosiner.github.io/zerver/2015/04/09/zerver.html)

It's mainly designed for restful api service, without session, template support, etc.. But you can still use it as a web framework by easily hack it. Documentation can be found at [godoc.org](godoc.org/github.com/cosiner/zerver), and each file contains a component, all api about this component is defined there.

##### Install
`go get github.com/cosiner/zerver`

##### Features
* RESTFul Route
* Tree-based mux/router, support route group, subrouter
* Filter(also known as middleware) Chain support
* Interceptor supported
* Builtin WebSocket support
* Builtin Task support
* Resource Marshal/Unmarshal, Pool marshaled bytes(if marshaler support)
* Request/Response Wrap

##### Getting Started
```Go
package main

import "github.com/cosiner/zerver"

func main() {
    server := zerver.NewServer()
    server.Get("/", func(req zerver.Request, resp zerver.Response) {
        resp.Write([]byte("Hello World!"))
    })
    server.Start(nil) // default listen at ":4000"
}
```

##### Config
```Go
ServerOption struct {
    // check for websocket header
    WebSocketChecker HeaderChecker // default nil
    // automiclly set for each response
    ContentType      string        // default application/json;charset=utf-8
    // average variable count at each route
    PathVarCount     int           // default 2
    // average filter count for of filters at each route
    FilterCount      int           // default 2
    // server listening address
    ListenAddr       string        // default :4000
    // TLS config
    CertFile         string        // default not enable tls
    KeyFile          string
    *ResourceMaster                // for resource marshal/unmarshal
                                   // used for Request.Recieve, Response.Send,
                                   // Request.ResourceMaster(), default use
                                   // zerver.JSONResource
}

server.Start(&ServerOption{
    ContentType:"text/plain;charset=utf-8",
    ListenAddr:":8000",
})
```

### Exampels
* resource
```Go
type User struct {
    Id int `json:"id"`
    Name string `json:"name"`
}
func Handle(req zerver.Request, resp zerver.Response) {
    u := &User{}
    req.Read(u)
    resp.Send("user", u)
}
```
* url variables
```Go
server.Get("/user/:id", func(req zerver.Request, resp zerver.Response) {
    resp.Write([]byte("Hello, " + req.URLVar("id")))
})
server.Get("/home/*subpath", func(req zerver.Request, resp zerver.Response) {
    resp.Write([]byte("You access " + req.URLVar("subpath")))
})
```

* filter
```Go
type logger func(v ...interface{})
func (l logger) Init(*zerver.Server) error {return nil}
func (l logger) Destroy() {}
func (l logger) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
    l(req.RemoteIP(), req.UserAgent(), req.URL())
    chain(req, resp)
}
server.Handle("/", logger(fmt.Println))
```

* interceptor
```Go
func BasicAuthFilter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
    auth := req.Authorization()
    if auth == "" || auth != "abc:123" {
        resp.ReportUnAuthorized()
        return
    }
    req.SetAttr("auth", auth) // do before
    chain(req, resp)
}

func AuthLogFilter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
    chain(req, resp)
    fmt.Println("Auth success: ", resp.Value()) // do after
}

func AuthHandler(req zerver.Request, resp zerver.Response) {
    resp.Write([]byte(req.Attr("auth")))
    resp.SetValue(true)
}

server.HandleFunc("/auth", "POST", zerver.InterceptHandler(
    AuthHandler,  BasicAuthFilter,  AuthLogFilter,
))
```




### Handler/Filter/WebSocketHandler/TaskHandler
There is only one method `Handle(pattern string, i interface{})` to add component 
to server(router), first parameter is the url pattern the handler process, second can be:
* `Router` (Created through `NewRouter()`)
* `Handler/HandlerFunc/Literal HandlerFunc/MapHandler/MethodHandler`
* `Filter/FilterFunc/Literal FilterFunc`
* `WebSocketHandler/WebSocketHandlerFunc/Literal WebSocketHandler`
* `TaskHandler/TaskHandlerFunc/Literal TaskHandlerFunc`  

For filter, it will be add to filters collection for this pattern.
For handlers, per __route__ should have only one handlers.
For router, all routes under this section should be managed by it, you can't use Handler/Router both with same prefix.

Note: in zerver, the pattern will compile to route, they are not equal. 
`/user/:id/info` and `/user/:name/info` is two pattern, but the same route.

### ResourceMaster
ResourceMaster responsible for marshal/unmarshal data, you can use it dependently or intergrate to Server.
```Go
MarshalFunc func(interface{}) ([]byte, error)
UnmarshalFunc func([]byte, interface{}) error
Marshaler interface {
    Marshal(interface{}) ([]byte, error)
    Pool([]byte) // pool marshaled buffer for reduce allocation
}
ResourceMaster struct {
    Marshaler
    Unmarshal UnmarshalFunc
}
func (m MarshalFunc) Marshal(v interface{}) ([]byte, error) { return m(v) }
func (MarshalFunc) Pool([]byte)                             {}

var JSONResource = ResourceMaster{
    Marshaler: MarshalFunc(json.Marshal),
    Unmarshal: json.Unmarshal,
}
```

### AttrContainer
Store attribute, the server has a locked container, each request has a unlocked
container, response has only a `interface{}` to store value, both used to share attributes between components.  
The server's container should be used to store global attributes. The request's container should be used to pass down values between filter/filter or filter/handler. The response's container should be used to pass up value.

Example:
```Go
server.Attr(name)
server.SetAttr(name, value)

request.Attr(name)
request.SetAttr(name, value)

response.Value()
response.SetValue(value)
```

### Contribution
Feedbacks or Pull Request is welcome.

### License
MIT.
