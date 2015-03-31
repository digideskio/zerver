package zerver

import (
	"net/url"
)

type (
	// FilterFunc represent common filter function type,
	FilterFunc func(Request, Response, FilterChain)

	// Filter is an filter that run before or after handler,
	// to modify or check request and response
	// it will be inited on server started, destroyed on server stopped
	Filter interface {
		ServerInitializer
		Destroy()
		Filter(Request, Response, FilterChain)
	}

	// FilterChain represent a chain of filter, the last is final handler,
	// to continue the chain, must call chain(Request, Response)
	FilterChain HandleFunc

	// FilterChain interface {
	// 	Continue(Request, Response)
	// }

	// filterChain holds filters and handler
	// if there is no filters, the final handler will be called
	filterChain struct {
		filters []Filter
		handler HandleFunc
	}

	// handlerInterceptor is a interceptor
	handlerInterceptor struct {
		filter  Filter
		handler HandleFunc
	}

	// MatchAll routes
	RootFilters interface {
		ServerInitializer
		// Filters return all root filters
		Filters(url *url.URL) []Filter
		// AddFilter add root filter for "/"
		AddFilter(interface{})
		Destroy()
	}

	rootFilters []Filter
)

const (
	errNotFilter = "Not a filter"
)

// EmptyFilterFunc is a empty filter function, it simplely continue the filter chain
// it's useful for test, may be also other conditions
func EmptyFilterFunc(req Request, resp Response, chain FilterChain) {
	chain(req, resp)
}

func convertFilter(i interface{}) Filter {
	switch f := i.(type) {
	case func(Request, Response, FilterChain):
		return FilterFunc(f)
	case Filter:
		return f
	}
	panic(errNotFilter)
}

// FilterFunc is a function Filter
func (FilterFunc) Init(*Server) error { return nil }
func (FilterFunc) Destroy()           {}
func (fn FilterFunc) Filter(req Request, resp Response, chain FilterChain) {
	fn(req, resp, chain)
}

func NewRootFilters(filters []Filter) RootFilters {
	var rfs = rootFilters(filters)
	if filters == nil {
		rfs = make(rootFilters, 0)
		return &rfs
	}
	return &rfs
}

func (rfs *rootFilters) Init(s *Server) error {
	for _, f := range *rfs {
		if err := f.Init(s); err != nil {
			return err
		}
	}
	return nil
}

// Filters returl all root filters
func (rfs *rootFilters) Filters(*url.URL) []Filter {
	return *rfs
}

// AddFilter add root filter
func (rfs *rootFilters) AddFilter(filter interface{}) {
	*rfs = append(*rfs, convertFilter(filter))
}

func (rfs *rootFilters) Destroy() {
	for _, f := range *rfs {
		f.Destroy()
	}
}

// newFilterChain create a chain of filter
//
// NOTICE: FilterChain keeps some states, it should be used only once
// if need unlimit FilterChain, use InterceptHandler replace
// but it takes less memory space, InterceptHandler takes more, make your choice
func newFilterChain(filters []Filter, handler func(Request, Response)) FilterChain {
	if handler == nil {
		handler = EmptyHandlerFunc
	}
	if l := len(filters); l == 0 {
		return handler
	} else if l == 1 {
		return newInterceptHandler(handler, filters[0])
	}
	chain := newFilterChainFromPool()
	chain.filters = filters
	chain.handler = handler
	return chain.handleChain
}

// handleChain call next filter, if there is no next filter,then call final handler
// if there is no chains, filterChain will be recycle to pool
// if chain is not continue to last, chain will not be recycled
func (chain *filterChain) handleChain(req Request, resp Response) {
	filters := chain.filters
	if len(filters) == 0 {
		chain.handler(req, resp)
		chain.destroy()
	} else {
		filter := filters[0]
		chain.filters = filters[1:]
		filter.Filter(req, resp, chain.handleChain)
	}
}

// destroy destroy all reference hold by filterChain, then recycle it to pool
func (chain *filterChain) destroy() {
	chain.filters = nil
	chain.handler = nil
	recycleFilterChain(chain)
}

// InterceptHandler will create a permanent HandleFunc/FilterChain, there is no
// states keeps, it can be used without limitation.
//
// InterceptHandler wrap a handler with some filters as a interceptor
func InterceptHandler(handler func(Request, Response), filters ...interface{}) func(Request, Response) {
	if handler == nil {
		handler = EmptyHandlerFunc
	}
	if len(filters) == 0 {
		return handler
	}
	return newInterceptHandler(InterceptHandler(handler, filters[1:]...), convertFilter(filters[0]))
}

func newInterceptHandler(handler func(Request, Response), filter Filter) func(Request, Response) {
	return (&handlerInterceptor{
		filter:  filter,
		handler: handler,
	}).handle
}

func (interceptor *handlerInterceptor) handle(req Request, resp Response) {
	interceptor.filter.Filter(req, resp, FilterChain(interceptor.handler))
}
