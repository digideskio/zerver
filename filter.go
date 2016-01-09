package zerver

type (
	FilterChain HandleFunc

	FilterFunc func(Request, Response, FilterChain)

	Filter interface {
		Component
		Filter(Request, Response, FilterChain)
	}
)

func NopFilterFunc(req Request, resp Response, chain FilterChain) {
	chain(req, resp)
}

func (FilterFunc) Init(Env) error { return nil }

func (FilterFunc) Destroy() {}

func (fn FilterFunc) Filter(req Request, resp Response, chain FilterChain) {
	fn(req, resp, chain)
}

type filterChain struct {
	filters []Filter
	handler FilterChain
}

func newFilterChain(handler FilterChain, filters ...Filter) FilterChain {
	if handler == nil {
		handler = NopHandleFunc
	}

	if l := len(filters); l == 0 {
		return handler
	}

	c := &filterChain{
		filters: filters,
		handler: handler,
	}

	return c.doChain
}

func (c *filterChain) doChain(req Request, resp Response) {
	filters := c.filters
	if len(filters) == 0 {
		c.handler(req, resp)
	} else {
		filter := filters[0]
		c.filters = filters[1:]
		filter.Filter(req, resp, c.doChain)
	}
}
