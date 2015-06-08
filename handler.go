package zerver

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
		Component
		// Handler return an method handle function by method name
		// if nill returned, means method not allowed
		Handler(method string) HandleFunc
	}

	// HandlerFunc is a function handler, it's different with HandleFunc
	HandlerFunc func(method string) HandleFunc

	// MethodHandler will apply standard handler mapping rule,
	// each method is correspond to it's handle function
	MethodHandler interface {
		Get(Request, Response)
		Post(Request, Response)
		Delete(Request, Response)
		Put(Request, Response)
		Patch(Request, Response)
	}

	// FakeMethodHandler's all method report 405(method not allowed)
	FakeMethodHandler struct{}

	standardHandler struct {
		Component
		MethodHandler
	}

	// MapHandler is a handler that use user customed handler function.
	MapHandler map[string]HandleFunc
)

// EmptyHandlerFunc is a empty handler function, it do nothing
// it's useful for test, may be also other conditions
func EmptyHandlerFunc(Request, Response) {}

// convertHandler convert a interfae to Handler,
// only support Handler,MapHandler,MethodHandler, otherwise return nil
func convertHandler(i interface{}) Handler {
	switch h := i.(type) {
	case Handler, MapHandler:
		return h.(Handler)
	case func(string) HandleFunc:
		return HandlerFunc(h)
	case map[string]HandleFunc:
		return MapHandler(h)

	case MethodHandler:
		var comp Component
		if c, is := h.(Component); is {
			comp = c
		} else {
			comp = FakeComponent{}
		}

		return standardHandler{
			Component:     comp,
			MethodHandler: h,
		}
	}

	return nil
}

func (HandlerFunc) Init(Environment) error { return nil }

func (HandlerFunc) Destroy() {}

func (h HandlerFunc) Handler(method string) HandleFunc { return h(method) }

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

func (mh MapHandler) Init(Environment) error {
	for m, h := range mh {
		delete(mh, m)
		mh[parseRequestMethod(m)] = h
	}

	return nil
}

// MapHandler implements MethodIndicator interface for custom method handler
func (mh MapHandler) Handler(method string) HandleFunc {
	return mh[method]
}

func (mh MapHandler) Destroy() {
	for m := range mh {
		delete(mh, m)
	}
}

// setMethodHandler setup method handler for MapHandler
func (mh MapHandler) setMethodHandler(method string, handleFunc HandleFunc) {
	mh[method] = handleFunc
}

func (FakeMethodHandler) Get(_ Request, resp Response) {
	resp.ReportMethodNotAllowed()
}

func (FakeMethodHandler) Post(_ Request, resp Response) {
	resp.ReportMethodNotAllowed()
}

func (FakeMethodHandler) Delete(_ Request, resp Response) {
	resp.ReportMethodNotAllowed()
}

func (FakeMethodHandler) Put(_ Request, resp Response) {
	resp.ReportMethodNotAllowed()
}

func (FakeMethodHandler) Patch(_ Request, resp Response) {
	resp.ReportMethodNotAllowed()
}
