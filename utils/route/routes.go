package route

import (
	"fmt"

	"github.com/cosiner/zerver"
	"github.com/cosiner/zerver/handler"
	"github.com/cosiner/zerver/utils/handle"
)

var globalInterceptors []zerver.Filter

func GlobalIntercept(inters ...interface{}) {
	globalInterceptors = append(globalInterceptors, convertFilters(inters)...)
}

type Route struct {
	zerver.Handler
	handler.MapHandler
}

type Routes map[string]Route

func convertHandleFunc(h interface{}) zerver.HandleFunc {
	switch h := h.(type) {
	case zerver.HandleFunc:
		return h
	case func(zerver.Request, zerver.Response):
		return h
	case func(zerver.Request, zerver.Response) error:
		return handle.Wrap(h)
	}

	panic("not a handle function")
}

func convertFilters(filters []interface{}) []zerver.Filter {
	res := make([]zerver.Filter, len(filters))
	for i, f := range filters {
		if ft, is := f.(zerver.Filter); is {
			res[i] = ft
		} else if fn, is := f.(zerver.FilterFunc); is {
			res[i] = fn
		} else if fn, is := f.(func(zerver.Request, zerver.Response, zerver.FilterChain)); is {
			res[i] = zerver.FilterFunc(fn)
		} else {
			panic(fmt.Errorf("interceptor at index %d is not a filter.", i))
		}
	}
	return res
}

func (r Routes) Apply(router zerver.Router) error {
	for pat, rt := range r {
		if rt.Handler == nil {
			rt.Handler = rt.MapHandler
		}
		err := router.Handler(pat, rt.Handler)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r Routes) HandleFunc(method, pattern string, fn zerver.HandleFunc, interceptors ...interface{}) Routes {
	rt, has := r[pattern]
	if has {
		if rt.Handler != nil {
			panic(fmt.Errorf("handler for patten %s already exists", pattern))
		}
	} else {
		rt = Route{
			MapHandler: make(handler.MapHandler),
		}
	}

	inters := convertFilters(interceptors)
	if len(globalInterceptors) != 0 {
		tmpInters := make([]zerver.Filter, 0, len(globalInterceptors)+len(interceptors))
		tmpInters = append(tmpInters, globalInterceptors...)
		inters = append(tmpInters, inters...)
	}
	rt.MapHandler.Use(method, zerver.Intercept(fn, inters...))
	r[pattern] = rt
	return r
}

func (r Routes) Handler(pattern string, handler zerver.Handler) Routes {
	_, has := r[pattern]
	if has {
		panic(fmt.Errorf("handler for patten %s already exists", pattern))
	}
	r[pattern] = Route{
		Handler: handler,
	}
	return r
}

func (r Routes) Get(pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return r.HandleFunc(zerver.METHOD_GET, pattern, handler, interceptors...)
}

func (r Routes) Post(pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return r.HandleFunc(zerver.METHOD_POST, pattern, handler, interceptors...)
}

func (r Routes) Delete(pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return r.HandleFunc(zerver.METHOD_DELETE, pattern, handler, interceptors...)
}

func (r Routes) Put(pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return r.HandleFunc(zerver.METHOD_PUT, pattern, handler, interceptors...)
}

func (r Routes) Patch(pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return r.HandleFunc(zerver.METHOD_PATCH, pattern, handler, interceptors...)
}

func New(method, pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return make(Routes).HandleFunc(method, pattern, handler, interceptors...)
}

func Handler(pattern string, handler zerver.Handler) Routes {
	return make(Routes).Handler(pattern, handler)
}

func Get(pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return New(zerver.METHOD_GET, pattern, handler, interceptors...)
}

func Post(pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return New(zerver.METHOD_POST, pattern, handler, interceptors...)
}

func Delete(pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return New(zerver.METHOD_DELETE, pattern, handler, interceptors...)
}

func Put(pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return New(zerver.METHOD_PUT, pattern, handler, interceptors...)
}

func Patch(pattern string, handler zerver.HandleFunc, interceptors ...interface{}) Routes {
	return New(zerver.METHOD_PATCH, pattern, handler, interceptors...)
}
