package zerver

type groupRouter struct {
	prefix string
	Router
}

func NewGroupRouter(rt Router, prefix string) Router {
	return groupRouter{
		prefix: prefix,
		Router: rt,
	}
}

// AddHandleFunc add a function handler, method are defined as constant string
func (gr groupRouter) AddHandleFunc(pattern string, method string, handler HandleFunc) error {
	return gr.Router.AddHandleFunc(gr.prefix+pattern, method, handler)
}

// AddHandler add a handler
func (gr groupRouter) AddHandler(pattern string, handler interface{}) error {
	return gr.Router.AddHandler(gr.prefix+pattern, handler)
}

// AddFilter add a filter
func (gr groupRouter) AddFilter(pattern string, filter interface{}) error {
	return gr.Router.AddFilter(gr.prefix+pattern, filter)
}

// AddWebSocketHandler add a websocket handler
func (gr groupRouter) AddWebSocketHandler(pattern string, handler interface{}) error {
	return gr.Router.AddWebSocketHandler(gr.prefix+pattern, handler)
}

// AddTaskHandler
func (gr groupRouter) AddTaskHandler(pattern string, handler interface{}) error {
	return gr.Router.AddTaskHandler(gr.prefix+pattern, handler)
}

// Get register a function handler process GET request for given pattern
func (gr groupRouter) Get(pattern string, handleFunc HandleFunc) error {
	return gr.Router.Get(gr.prefix+pattern, handleFunc)
}

// Post register a function handler process POST request for given pattern
func (gr groupRouter) Post(pattern string, handleFunc HandleFunc) error {
	return gr.Router.Post(gr.prefix+pattern, handleFunc)
}

// Put register a function handler process PUT request for given pattern
func (gr groupRouter) Put(pattern string, handleFunc HandleFunc) error {
	return gr.Router.Put(gr.prefix+pattern, handleFunc)
}

// Delete register a function handler process DELETE request for given pattern
func (gr groupRouter) Delete(pattern string, handleFunc HandleFunc) error {
	return gr.Router.Delete(gr.prefix+pattern, handleFunc)
}

// Patch register a function handler process PATCH request for given pattern
func (gr groupRouter) Patch(pattern string, handleFunc HandleFunc) error {
	return gr.Router.Patch(gr.prefix+pattern, handleFunc)
}
