# Zerver [![wercker status](https://app.wercker.com/status/e28d0f3a9fd6a9f9142cb9199b1715e0/s "wercker status")](https://app.wercker.com/project/bykey/e28d0f3a9fd6a9f9142cb9199b1715e0) [![GoDoc](https://godoc.org/github.com/go-martini/martini?status.png)](http://godoc.org/github.com/cosiner/zerver)
__Zerver__ is a simple, scalable, restful api framework for [golang](http://golang.org).

[中文介绍](http://cosiner.github.io/zerver/2015/04/09/zerver.html)

It's mainly designed for restful api service, without session, template support, etc.. But you can still use it as a web framework by easily hack it. Documentation can be found at [godoc.org](https://godoc.org/github.com/cosiner/zerver), and each file contains a component, all api about this component is defined there.

##### Install
`go get github.com/cosiner/zerver`

##### Features
* RESTFul Route
* Tree-based mux/router, support route group, subrouter
* Helpful functions about request/response
* Filter(also known as middleware) Chain support
* Interceptor supported
* WebSocket support
* Task support
* Resource Marshal/Unmarshal, Pool marshaled bytes(if marshaler support)
* Request/Response Wrap
* Pluggable, lazy-initializable, removeable global components
* Predefined components/filters such as cors,compress,log,ffjson, redis etc..

### Getting Started
```Go
package main

import (
    "github.com/cosiner/zerver"
    "time"
)

func main() {
    server := zerver.NewServer()
    server.Get("/", func(req zerver.Request, resp zerver.Response) {
        resp.WriteString("Hello World!")
    })
    
    var err error
    go func(server *Server, err *error) {
        time.Sleep(10 * time.Millisecond)
        if *err == nil {
            server.Destroy()
        }
    }(server, &err)
    
    err = server.Start(nil) // default listen at ":4000"
}
```

### Exampels
* resource
```Go
type User struct {
    Id int      `json:"id"`
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
    resp.WriteString("Hello, " + req.URLVar("id"))
})
server.Get("/home/*subpath", func(req zerver.Request, resp zerver.Response) {
    resp.WriteString("You access " + req.URLVar("subpath"))
})
```

* filter
```Go
type logger func(v ...interface{}) // it can used as ServerOption.ErrorLogger
func (log logger) Init(zerver.Enviroment) error {return nil}
func (log logger) Destroy() {}
func (log logger) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
    log(req.RemoteIP(), req.UserAgent(), req.URL().Path)
    chain(req, resp) // continue the processing
}
server.Handle("/", logger(log.Println))
```

* interceptor
```Go
func BasicAuth(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
    user, pass := req.BasicAuth()
    if user != "abc" || pass != "123" {
        resp.ReportUnAuthorized()
        return
    }
    req.SetAttr("user", user) // do before
    chain(req, resp)
}

func AuthLog(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
    chain(req, resp)
    log.Println("Auth success: ", resp.Value()) // do after
}

func Auth(req zerver.Request, resp zerver.Response) {
    resp.WriteString(req.Attr("user"))
    resp.SetValue(true)
}

server.Post("/auth", zerver.Intercept(
    AuthHandler,  BasicAuthFilter,  AuthLogFilter,
))
```

* component
```Go
serer.AddComponent(zerver.CMP_RESOURCE, zerver.ComponentState{
    Initialized:true,
    Component:components.Ffjson{},
})
serer.AddComponent(components.CMP_REDIS, zerver.ComponentState{
    NoLazy:true,
    Component:&components.Redis{},
})

redis, err := server.Component(zerver.CMP_REDIS)

server.ManageComponent(customedComponent) // anonymous component
```

### Config
```Go
ServerOption struct {
    // server listening address, default :4000
    ListenAddr string
    // content type for each request, default application/json;charset=utf-8
    ContentType string

    // check websocket header, default nil
    WebSocketChecker HeaderChecker
    // logger, default use cosiner/gohper/log.Logger with ConsoleWriter
    Logger

    // path variables count, suggest set as max or average, default 3
    PathVarCount int
    // filters count for each route, RootFilters is not include, default 5
    FilterCount int

    // read timeout by millseconds
    ReadTimeout int
    // write timeout by millseconds
    WriteTimeout int
    // max header bytes
    MaxHeaderBytes int
    // tcp keep-alive period by minutes,
    // default 3, same as predefined in standard http package
    KeepAlivePeriod int
    // ssl config, default disable tls
    CertFile, KeyFile string
    // if not nil, cert and key will be ignored
    TLSConfig *tls.Config
}

server.Start(&ServerOption{
    ContentType:"text/plain;charset=utf-8",
    ListenAddr:":8000",
})
```

### Server
```Go
// NOTICE: Server only configured through ServerOption
Server struct {
    // exported fields
    Router
    AttrContainer
    RootFilters RootFilters // root filters, Match Every Routes
    ResourceMaster ResourceMaster // resource master, manage resource types
    Errorln     func(...interface{}) // log error message
}
```

### Enviroment
```Go
// Enviroment is a server enviroment, real implementation is the Server itself.
// it can be accessed from Request/WebsocketConn
Enviroment interface {
    Server() *Server
    Logger() Logger
    StartTask(path string, value interface{})
    Component(name string) (Component, error)
}
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

### ResourceMaster/Resource
`ResourceMaster` manage multiple resource types you added, it's stored in `Server`.
`Resource` responsible for marshal/unmarshal data. `JSONResource/XMLResource` already provided, and `Ffjson` is also provided under `components` package.

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
