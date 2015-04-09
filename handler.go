package zerver

import (
	"strings"
)

type (
	// HandleFunc is the common request handler function type
	HandleFunc func(Request, Response)

	// Handler is an common interface of request handler
	// it will be inited on server started, destroyed on server stopped
	// and use it's method to dispatch which function to handle the request
	//
	// Handler represent a resource, method menas operation,
	// if handler was not found, 404 is default status, if handler found, but
	// Handler.Handler return nil, 405 is default status.
	// The meaning of 'default status' is that filters can change the status
	Handler interface {
		ServerInitializer
		// Handler return an method handle function by method name
		// if nill returned, means method not allowed
		Handler(method string) HandleFunc
	}

	// HandlerFunc is a function handler, it's different with HandleFunc
	HandlerFunc func(method string) HandleFunc

	// MethodHandler will apply standard handler mapping rule,
	// each method is correspond to it's handler
	MethodHandler interface {
		ServerInitializer
		Get(Request, Response)
		Post(Request, Response)
		Delete(Request, Response)
		Put(Request, Response)
		Patch(Request, Response)
	}

	standardHandler struct {
		MethodHandler
	}

	// MapHandler is a handler that use user customed handler function.
	MapHandler map[string]HandleFunc
)

// EmptyHandlerFunc is a empty handler function, it do nothing
// it's useful for test, may be also other conditions
func EmptyHandlerFunc(Request, Response) {}

// convertHandler convert a interfae to Handler,
// only support Handler,MapHandler,MethodHandler, otherwise panic
func convertHandler(i interface{}) Handler {
	switch h := i.(type) {
	case Handler:
		return h
	case func(string) HandleFunc:
		return HandlerFunc(h)
	case map[string]HandleFunc:
		return MapHandler(h)
	case MethodHandler:
		return standardHandler{MethodHandler: h}
	}
	return nil
}

func (HandlerFunc) Init(*Server) error {
	return nil
}

func (HandlerFunc) Destroy() {}

func (h HandlerFunc) Handler(method string) HandleFunc {
	return h(method)
}

func (s standardHandler) Handler(method string) HandleFunc {
	switch method {
	case GET:
		return s.Get
	case POST:
		return s.Post
	case DELETE:
		return s.Delete
	case PUT:
		return s.Put
	case PATCH:
		return s.Patch
	}
	return nil
}

func (fh MapHandler) Init(*Server) error {
	for m, h := range fh {
		delete(fh, m)
		fh[strings.ToUpper(m)] = h
	}
	return nil
}

// MapHandler implements MethodIndicator interface for custom method handler
func (fh MapHandler) Handler(method string) (handleFunc HandleFunc) {
	return fh[method]
}

func (fh MapHandler) Destroy() {
	for m := range fh {
		delete(fh, m)
	}
}

// setMethodHandler setup method handler for MapHandler
func (fh MapHandler) setMethodHandler(method string, handleFunc HandleFunc) {
	fh[method] = handleFunc
}
