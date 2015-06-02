package route

import (
	"github.com/cosiner/zerver"
)

type Route struct {
	Method       string
	Pattern      string
	Handler      zerver.HandleFunc
	Interceptors []interface{}
}

type Routes []Route

func (r Routes) New(method, pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return append(r, Route{
		Method:       method,
		Pattern:      pattern,
		Handler:      handler,
		Interceptors: interceptors,
	})
}

func (r Routes) Apply(rt zerver.Router) error {
	var err error
	for i := 0; err == nil && i < len(r); i++ {
		var route = r[i]
		err = rt.HandleFunc(route.Pattern,
			route.Method,
			zerver.Intercept(route.Handler,
				route.Interceptors...))
	}

	return err
}

func (r Routes) Get(pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return r.New("GET", pattern, handler, interceptors...)
}

func (r Routes) Post(pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return r.New("POST", pattern, handler, interceptors...)
}

func (r Routes) Delete(pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return r.New("DELETE", pattern, handler, interceptors...)
}

func (r Routes) Put(pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return r.New("PUT", pattern, handler, interceptors...)
}

func (r Routes) Patch(pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return r.New("PATCH", pattern, handler, interceptors...)
}

func New(method, pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return Routes(nil).New(method, pattern, handler, interceptors...)
}

func Get(pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return New("GET", pattern, handler, interceptors...)
}

func Post(pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return New("POST", pattern, handler, interceptors...)
}

func Delete(pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return New("DELETE", pattern, handler, interceptors...)
}

func Put(pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return New("PUT", pattern, handler, interceptors...)
}

func Patch(pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return New("PATCH", pattern, handler, interceptors...)
}

func Apply(r zerver.Router, routes ...Route) error {
	return Routes(routes).Apply(r)
}
