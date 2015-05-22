package zerver

import (
	"log"
	"net/url"
)

type (
	// FilterFunc represent common filter function type,
	FilterFunc func(Request, Response, FilterChain)

	// FilterChain represent a chain of filter, the last is final handler,
	// to continue the chain, must call chain(Request, Response)
	FilterChain HandleFunc

	// Filter is an filter that run before or after handler,
	// to modify or check request and response
	// it will be inited on server started, destroyed on server stopped
	Filter interface {
		Component
		Filter(Request, Response, FilterChain)
	}

	// filterChain holds filters and handler
	// if there is no filters, the final handler will be called
	filterChain struct {
		filters []Filter
		handler HandleFunc
	}

	interceptor struct {
		filter  Filter
		handler HandleFunc
	}

	// MatchAll routes
	RootFilters interface {
		Component
		// Filters return all root filters
		Filters(url *url.URL) []Filter
		// Add add root filter for "/"
		Add(interface{})
	}

	rootFilters []Filter
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

	return nil
}

func panicConvertFilter(i interface{}) Filter {
	f := convertFilter(i)
	if f == nil {
		log.Panicln("Not a filter")
	}

	return f
}

// FilterFunc is a function Filter
func (FilterFunc) Init(Environment) error { return nil }
func (FilterFunc) Destroy()              {}
func (fn FilterFunc) Filter(req Request, resp Response, chain FilterChain) {
	fn(req, resp, chain)
}

func NewRootFilters(filters []Filter) RootFilters {
	rfs := rootFilters(filters)
	if filters == nil {
		rfs = make(rootFilters, 0)
	}

	return &rfs
}

func (rfs *rootFilters) Init(e Environment) error {
	var err error
	for i := 0; i < len(*rfs) && err == nil; i++ {
		err = (*rfs)[i].Init(e)
	}

	return err
}

// Filters returl all root filters
func (rfs *rootFilters) Filters(*url.URL) []Filter {
	return *rfs
}

// Add add root filter
func (rfs *rootFilters) Add(filter interface{}) {
	*rfs = append(*rfs, panicConvertFilter(filter))
}

func (rfs *rootFilters) Destroy() {
	for _, f := range *rfs {
		f.Destroy()
	}
}

// newFilterChain create a chain of filter
//
// NOTICE: FilterChain keeps some states, it should be used only once
// if need unlimit FilterChain, use Intercept replace
// but it takes less memory space, Intercept takes more, make your choice
func newFilterChain(filters []Filter, handler func(Request, Response)) FilterChain {
	if handler == nil {
		handler = EmptyHandlerFunc
	}

	if l := len(filters); l == 0 {
		return handler
	} else if l == 1 {
		return FilterChain(newInterceptor(handler, filters[0]))
	}

	return (&filterChain{
		filters: filters,
		handler: handler,
	}).handle
}

// handle call next filter, if there is no next filter,then call final handler
// if there is no chains, filterChain will be recycle to pool
// if chain is not continue to last, chain will not be recycled
func (chain *filterChain) handle(req Request, resp Response) {
	filters := chain.filters
	if len(filters) == 0 {
		chain.handler(req, resp)
	} else {
		filter := filters[0]
		chain.filters = filters[1:]
		filter.Filter(req, resp, chain.handle)
	}
}

// Intercept wrap a handler with some filters as a interceptor,
// it will create a permanent HandleFunc/FilterChain, there is no
// states keeps, it can be used without limitation.
func Intercept(handler HandleFunc, filters ...interface{}) HandleFunc {
	if handler == nil {
		handler = EmptyHandlerFunc
	}

	if len(filters) == 0 {
		return handler
	}

	return newInterceptor(Intercept(handler, filters[1:]...), panicConvertFilter(filters[0]))
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
