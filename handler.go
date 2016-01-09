package zerver

type (
	HandleFunc func(Request, Response)

	Handler interface {
		Component
		Handler(method string) HandleFunc
	}

	HandlerFunc func(method string) HandleFunc

	interceptor struct {
		filter  Filter
		handler HandleFunc
	}
)

func NopHandleFunc(Request, Response) {}

func (HandlerFunc) Init(Env) error { return nil }

func (HandlerFunc) Destroy() {}

func (h HandlerFunc) Handler(method string) HandleFunc {
	return h(method)
}

// Intercept will create an immutable filter chain
// interceptors' lifetime should be managed by the Server
func Intercept(handler HandleFunc, interceptors ...Filter) HandleFunc {
	if handler == nil {
		handler = NopHandleFunc
	}
	if len(interceptors) == 0 {
		return handler
	}
	return newInterceptor(Intercept(handler, interceptors[1:]...), interceptors[0])
}

func newInterceptor(handler func(Request, Response), filter Filter) HandleFunc {
	return (&interceptor{
		filter:  filter,
		handler: handler,
	}).handle
}

func (i *interceptor) handle(req Request, resp Response) {
	i.filter.Filter(req, resp, FilterChain(i.handler))
}
