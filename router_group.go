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

// AddFuncHandler add a function handler, method are defined as constant string
func (gr groupRouter) AddFuncHandler(pattern string, method string, handler HandlerFunc) error {
	return gr.Router.AddFuncHandler(gr.prefix+pattern, method, handler)
}

// AddHandler add a handler
func (gr groupRouter) AddHandler(pattern string, handler Handler) error {
	return gr.Router.AddHandler(gr.prefix+pattern, handler)
}

func (gr groupRouter) AddOptionHandler(pattern string, o *OptionHandler) error {
	return gr.Router.AddOptionHandler(gr.prefix+pattern, o)
}

// AddFuncFilter add function filter
func (gr groupRouter) AddFuncFilter(pattern string, filter FilterFunc) error {
	return gr.Router.AddFuncFilter(gr.prefix+pattern, filter)
}

// AddFilter add a filter
func (gr groupRouter) AddFilter(pattern string, filter Filter) error {
	return gr.Router.AddFilter(gr.prefix+pattern, filter)
}

// AddFuncWebSocketHandler add a websocket functionhandler
func (gr groupRouter) AddFuncWebSocketHandler(pattern string, handler WebSocketHandlerFunc) error {
	return gr.Router.AddFuncWebSocketHandler(gr.prefix+pattern, handler)
}

// AddWebSocketHandler add a websocket handler
func (gr groupRouter) AddWebSocketHandler(pattern string, handler WebSocketHandler) error {
	return gr.Router.AddWebSocketHandler(gr.prefix+pattern, handler)
}

// AddFuncTaskHandler
func (gr groupRouter) AddFuncTaskHandler(pattern string, handler TaskHandlerFunc) error {
	return gr.Router.AddFuncTaskHandler(gr.prefix+pattern, handler)
}

// AddTaskHandler
func (gr groupRouter) AddTaskHandler(pattern string, handler TaskHandler) error {
	return gr.Router.AddTaskHandler(gr.prefix+pattern, handler)
}

// Get register a function handler process GET request for given pattern
func (gr groupRouter) Get(pattern string, handlerFunc HandlerFunc) error {
	return gr.Router.Get(gr.prefix+pattern, handlerFunc)
}

// Post register a function handler process POST request for given pattern
func (gr groupRouter) Post(pattern string, handlerFunc HandlerFunc) error {
	return gr.Router.Post(gr.prefix+pattern, handlerFunc)
}

// Put register a function handler process PUT request for given pattern
func (gr groupRouter) Put(pattern string, handlerFunc HandlerFunc) error {
	return gr.Router.Put(gr.prefix+pattern, handlerFunc)
}

// Delete register a function handler process DELETE request for given pattern
func (gr groupRouter) Delete(pattern string, handlerFunc HandlerFunc) error {
	return gr.Router.Delete(gr.prefix+pattern, handlerFunc)
}

// Patch register a function handler process PATCH request for given pattern
func (gr groupRouter) Patch(pattern string, handlerFunc HandlerFunc) error {
	return gr.Router.Patch(gr.prefix+pattern, handlerFunc)
}
