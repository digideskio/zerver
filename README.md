# Zerver
__Zerver__ is a simple, scalable, restful api framework for [golang](http://golang.org).

It's mainly designed for restful api service, without session, template, cookie support, etc.. But you can still use it as a web framework by easily hack it. Documentation can be found at [godoc.org](godoc.org/github.com/cosiner/zerver), and each file contains a component, all api about this component is defined there.

### Install
`go get github.com/cosiner/zerver`

### Features
* RESTFul Route
* Tree-based mux/router, support route group
* Filter(also known as middleware) Chain supported
* Interceptor supported
* Builtin WebSocket supported
* Builtin Task supported
* Resource Marshaler/Unmarshaler

### Getting Started
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


# Exampels
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
server.AddFilter("/", logger(fmt.Println))
```

* interceptor
```Go
func BasicAuthFilter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
    auth := req.Header("Authorization")
    if auth == "" {
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

server.AddHandleFunc("/auth", "POST", zerver.InterceptHandler(
    AuthHandler,  BasicAuthFilter,  AuthLogFilter,
))
```


### Config
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
}
```
Example:
```
server.Start(&ServerOption{
    ContentType:"text/plain;charset=utf-8",
    ListenAddr:":8000",
})
```

### ResourceMaster
ResourceMaster responsible for marshal/unmarshal data, it's indenpendent from any other components, create it yourself.
```
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
var DefaultResourceMaster = ResourceMaster{
    Marshaler: MarshalFunc(json.Marshal),
    Unmarshal: json.Unmarshal,
}
```

### AttrContainer
Store attribute, the server has a locked container, each request has a unlocked
container, response has only a `interface{}` to store value, both used to share attributes between components.  
The server's container should be used to store global attributes. The request's container should be used to pass down values between filter/filter or filter/handler. The response's container should be used to pass up value.

Example:
```
server.Attr(name)
server.SetAttr(name, value)

request.Attr(name)
request.SetAttr(name, value)

response.Value()
response.SetValue(value)
```

More detail please see [wiki page](https://github.com/cosiner/zerver/wiki).

### License
MIT.
