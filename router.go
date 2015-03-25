package zerver

import (
	"io"
	"net/url"
	"strings"

	"github.com/cosiner/gohper/lib/sys"

	. "github.com/cosiner/gohper/lib/errors"
)

type (
	Router interface {
		RouteStore
		RouteMatcher
	}

	RouteStore interface {
		MethodRouter

		// Init init handlers and filters, websocket handlers
		Init(func(Handler) bool, func(Filter) bool, func(WebSocketHandler) bool, func(TaskHandler) bool)
		// Destroy destroy router, also responsible for destroy all handlers and filters
		Destroy()
		PrintRouteTree(w io.Writer)
		Group(prefix string, fn func(Router))

		// AddFuncHandler add a function handler, method are defined as constant string
		AddFuncHandler(pattern string, method string, handler HandlerFunc) error
		// AddHandler add a handler
		AddHandler(pattern string, handler Handler) error

		AddOptionHandler(pattern string, o *OptionHandler) error
		// AddFuncFilter add function filter
		AddFuncFilter(pattern string, filter FilterFunc) error
		// AddFilter add a filter
		AddFilter(pattern string, filter Filter) error
		// AddFuncWebSocketHandler add a websocket functionhandler
		AddFuncWebSocketHandler(pattern string, handler WebSocketHandlerFunc) error
		// AddWebSocketHandler add a websocket handler
		AddWebSocketHandler(pattern string, handler WebSocketHandler) error
		// AddFuncTaskHandler
		AddFuncTaskHandler(pattern string, handler TaskHandlerFunc) error
		// AddTaskHandler
		AddTaskHandler(pattern string, handler TaskHandler) error
	}

	RouteMatcher interface {
		// MatchHandlerFilters match given url to find all matched filters and final handler
		MatchHandlerFilters(url *url.URL) (Handler, URLVarIndexer, []Filter)
		// MatchWebSocketHandler match given url to find a matched websocket handler
		MatchWebSocketHandler(url *url.URL) (WebSocketHandler, URLVarIndexer)
		// MatchTaskHandler
		MatchTaskHandler(url *url.URL) (TaskHandler, URLVarIndexer)
	}

	MethodRouter interface {
		Get(string, HandlerFunc) error
		Post(string, HandlerFunc) error
		Put(string, HandlerFunc) error
		Delete(string, HandlerFunc) error
		Patch(string, HandlerFunc) error
	}

	// handlerProcessor keep handler and url variables of this route
	handlerProcessor struct {
		vars    map[string]int
		handler Handler
	}

	// wsHandlerProcessor keep websocket handler and url variables of this route
	wsHandlerProcessor struct {
		vars      map[string]int
		wsHandler WebSocketHandler
	}

	taskHandlerProcessor struct {
		vars        map[string]int
		taskHandler TaskHandler
	}

	// routeProcessor is processor of a route, it can hold handler, filters, and websocket handler
	routeProcessor struct {
		wsHandlerProcessor   *wsHandlerProcessor
		handlerProcessor     *handlerProcessor
		taskHandlerProcessor *taskHandlerProcessor
		filters              []Filter
	}

	// router is a actual url router, it only process path of url, other section is
	// not mentioned
	router struct {
		str       string          // path section hold by current route node
		chars     []byte          // all possible first characters of next route node
		childs    []*router       // child routers
		processor *routeProcessor // processor for current route node
		noFilter  bool
	}
)

const (
	ErrConflictPathVar = Err("There is a similar route pattern which use same wildcard" +
		" or catchall at the same position, " +
		"this means one of them will nerver be matched, " +
		"please check your routes")
	ErrHandlerExist = Err("handler for this route already exist")
)

// init init handler and filters hold by routeProcessor
func (rp *routeProcessor) init(initHandler func(Handler) bool,
	initFilter func(Filter) bool,
	initWebSocketHandler func(WebSocketHandler) bool,
	initTaskHandler func(TaskHandler) bool) bool {
	continu := true
	if rp.handlerProcessor != nil {
		continu = initHandler(rp.handlerProcessor.handler)
	}
	for _, f := range rp.filters {
		if !continu {
			break
		}
		continu = initFilter(f)
	}
	if continu && rp.wsHandlerProcessor != nil {
		continu = initWebSocketHandler(rp.wsHandlerProcessor.wsHandler)
	}
	if continu && rp.taskHandlerProcessor != nil {
		continu = initTaskHandler(rp.taskHandlerProcessor.taskHandler)
	}
	return continu
}

// destroy destroy handler and filters hold by routeProcessor
func (rp *routeProcessor) destroy() {
	if rp.handlerProcessor != nil {
		rp.handlerProcessor.handler.Destroy()
		rp.handlerProcessor = nil
	}
	for _, f := range rp.filters {
		f.Destroy()
	}
	rp.filters = nil
	if rp.wsHandlerProcessor != nil {
		rp.wsHandlerProcessor.wsHandler.Destroy()
		rp.wsHandlerProcessor = nil
	}
	if rp.taskHandlerProcessor != nil {
		rp.taskHandlerProcessor.taskHandler.Destroy()
		rp.taskHandlerProcessor = nil
	}
}

// NewRouter create a new Router
func NewRouter() Router {
	rt := new(router)
	rt.noFilter = true
	return rt
}

// Init init all handlers, filters, websocket handlers in route tree
func (rt *router) Init(initHandler func(Handler) bool,
	initFilter func(Filter) bool,
	initWebSocketHandler func(WebSocketHandler) bool,
	initTaskHandler func(TaskHandler) bool) {
	rt.init(initHandler, initFilter, initWebSocketHandler, initTaskHandler)
}

// Init init all handlers, filters, websocket handlers in route tree
func (rt *router) init(initHandler func(Handler) bool,
	initFilter func(Filter) bool,
	initWebSocketHandler func(WebSocketHandler) bool,
	initTaskHandler func(TaskHandler) bool) bool {
	continu := true
	if rt.processor != nil {
		continu = rt.processor.init(initHandler, initFilter, initWebSocketHandler, initTaskHandler)
	}
	for _, c := range rt.childs {
		if !continu {
			break
		}
		continu = c.init(initHandler, initFilter, initWebSocketHandler, initTaskHandler)
	}
	return continu
}

// Destroy destroy router and all handlers, filters, websocket handlers
func (rt *router) Destroy() {
	if rt.processor != nil {
		rt.processor.destroy()
	}
	for _, c := range rt.childs {
		c.Destroy()
	}
}

// routeProcessor return processor of current route node, if not exist
// then create a new one
func (rt *router) routeProcessor() *routeProcessor {
	if rt.processor == nil {
		rt.processor = new(routeProcessor)
	}
	return rt.processor
}

// Get register a function handler process GET request for given pattern
func (rt *router) Get(pattern string, handlerFunc HandlerFunc) error {
	return rt.AddFuncHandler(pattern, GET, handlerFunc)
}

// Post register a function handler process POST request for given pattern
func (rt *router) Post(pattern string, handlerFunc HandlerFunc) error {
	return rt.AddFuncHandler(pattern, POST, handlerFunc)
}

// Put register a function handler process PUT request for given pattern
func (rt *router) Put(pattern string, handlerFunc HandlerFunc) error {
	return rt.AddFuncHandler(pattern, PUT, handlerFunc)
}

// Delete register a function handler process DELETE request for given pattern
func (rt *router) Delete(pattern string, handlerFunc HandlerFunc) error {
	return rt.AddFuncHandler(pattern, DELETE, handlerFunc)
}

// Patch register a function handler process PATCH request for given pattern
func (rt *router) Patch(pattern string, handlerFunc HandlerFunc) error {
	return rt.AddFuncHandler(pattern, PATCH, handlerFunc)
}

func (rt *router) Group(prefix string, fn func(Router)) {
	fn(NewGroupRouter(rt, prefix))
}

// AddFuncHandler add function handler to router for given pattern and method
func (rt *router) AddFuncHandler(pattern, method string, handler HandlerFunc) (err error) {
	method = parseRequestMethod(method)
	if fHandler := _tmpGetFuncHandler(pattern); fHandler == nil {
		fHandler = newFuncHandler()
		fHandler.setMethodHandler(method, handler)
		if err = rt.AddHandler(pattern, fHandler); err == nil {
			_tmpSetFuncHandler(pattern, fHandler)
		}
	} else {
		fHandler.setMethodHandler(method, handler)
	}
	return
}

// AddHandler add handler to router for given pattern
func (rt *router) AddHandler(pattern string, handler Handler) error {
	return rt.addPattern(pattern, func(rp *routeProcessor, pathVars map[string]int) error {
		if rp.handlerProcessor != nil {
			return ErrHandlerExist
		}
		rp.handlerProcessor = &handlerProcessor{
			vars:    pathVars,
			handler: handler,
		}
		return nil
	})
}

func (rt *router) AddOptionHandler(pattern string, o *OptionHandler) error {
	return rt.AddHandler(pattern, newFuncHandlerFrom(o))
}

// AddFuncWebSocketHandler add funciton websocket handler to router for given pattern
func (rt *router) AddFuncWebSocketHandler(pattern string, handler WebSocketHandlerFunc) error {
	return rt.AddWebSocketHandler(pattern, handler)
}

// AddWebSocetHandler add websocket handler to router for given pattern
func (rt *router) AddWebSocketHandler(pattern string, handler WebSocketHandler) error {
	return rt.addPattern(pattern, func(rp *routeProcessor, pathVars map[string]int) error {
		if rp.wsHandlerProcessor != nil {
			return ErrHandlerExist
		}
		rp.wsHandlerProcessor = &wsHandlerProcessor{
			vars:      pathVars,
			wsHandler: handler,
		}
		return nil
	})
}

// AddFuncTaskHandler add funciton task handler to router for given pattern
func (rt *router) AddFuncTaskHandler(pattern string, handler TaskHandlerFunc) error {
	return rt.AddTaskHandler(pattern, handler)
}

// AddTaskHandler add task handler to router for given pattern
func (rt *router) AddTaskHandler(pattern string, handler TaskHandler) error {
	return rt.addPattern(pattern, func(rp *routeProcessor, pathVars map[string]int) error {
		if rp.taskHandlerProcessor != nil {
			return ErrHandlerExist
		}
		rp.taskHandlerProcessor = &taskHandlerProcessor{
			vars:        pathVars,
			taskHandler: handler,
		}
		return nil
	})
}

// AddFuncFilter add function filter to router
func (rt *router) AddFuncFilter(pattern string, filter FilterFunc) error {
	return rt.AddFilter(pattern, filter)
}

// AddFuncFilter add filter to router
func (rt *router) AddFilter(pattern string, filter Filter) error {
	rt.noFilter = false
	return rt.addPattern(pattern, func(rp *routeProcessor, _ map[string]int) error {
		rp.filters = append(rp.filters, filter)
		return nil
	})
}

// addPattern compile pattern, extract all variables, and add it to route tree
// setup by given function
func (rt *router) addPattern(pattern string, fn func(*routeProcessor, map[string]int) error) error {
	routePath, pathVars, err := compile(pattern)
	if err == nil {
		if r, success := rt.addPath(routePath); success {
			err = fn(r.routeProcessor(), pathVars)
		} else {
			err = ErrConflictPathVar
		}
	}
	return err
}

// MatchWebSockethandler match url to find final websocket handler
func (rt *router) MatchWebSocketHandler(url *url.URL) (WebSocketHandler, URLVarIndexer) {
	path := url.Path
	indexer := newVarIndexerFromPool()
	rt, values := rt.matchOne(path, indexer.values)
	indexer.values = values
	if rt != nil {
		if p := rt.processor; p != nil {
			if wsp := p.wsHandlerProcessor; wsp != nil {
				indexer.vars = wsp.vars
				return wsp.wsHandler, indexer
			}
		}
	}
	return nil, indexer
}

// MatchTaskhandler match url to find final websocket handler
func (rt *router) MatchTaskHandler(url *url.URL) (TaskHandler, URLVarIndexer) {
	path := url.Path
	indexer := newVarIndexerFromPool()
	rt, values := rt.matchOne(path, indexer.values)
	indexer.values = values
	if rt != nil {
		if p := rt.processor; p != nil {
			if thp := p.taskHandlerProcessor; thp != nil {
				indexer.vars = thp.vars
				return thp.taskHandler, indexer
			}
		}
	}
	return nil, indexer

}

// // MatchHandler match url to find final websocket handler
// func (rt *router) MatchHandler(url *url.URL) (handler Handler, indexer URLVarIndexer) {
// 	path := url.Path
// 	indexer = newVarIndexerFromPool()
// 	rt, values := rt.matchOne(path, indexer.values)
// 	indexer.values = values
// 	if rt != nil && rt.processor != nil {
// 		if hp := p.handlerProcessor; hp != nil {
// 			indexer.vars = hp.vars
// 			handler = hp.handler
// 		}
// 	}
// 	return
// }

// MatchHandlerFilters match url to fin final handler and each filters
func (rt *router) MatchHandlerFilters(url *url.URL) (Handler,
	URLVarIndexer, []Filter) {
	var (
		path    = url.Path
		p       *routeProcessor
		indexer = newVarIndexerFromPool()
		values  = indexer.values
		filters []Filter
	)
	if rt.noFilter {
		rt, values = rt.matchOne(path, indexer.values)
	} else {
		filters = newFiltersFromPool()
		pathIndex, continu := 0, true
		for continu {
			if p = rt.processor; p != nil {
				if pfs := p.filters; len(pfs) != 0 {
					filters = append(filters, pfs...)
				}
			}
			pathIndex, values, rt, continu = rt.matchMultiple(path, pathIndex, values)
		}
	}
	indexer.values = values
	if rt != nil {
		if p = rt.processor; p != nil {
			if hp := p.handlerProcessor; hp != nil {
				indexer.vars = hp.vars
				return hp.handler, indexer, filters
			}
		}
	}
	return nil, indexer, filters
}

// addPath add an new path to route, use given function to operate the final
// route node for this path
func (rt *router) addPath(path string) (*router, bool) {
	str := rt.str
	if str == "" && len(rt.chars) == 0 {
		rt.str = path
	} else {
		diff, pathLen, strLen := 0, len(path), len(str)
		for diff != pathLen && diff != strLen && path[diff] == str[diff] {
			diff++
		}
		if diff < pathLen {
			first := path[diff]
			if diff == strLen {
				for i, c := range rt.chars {
					if c == first {
						return rt.childs[i].addPath(path[diff:])
					}
				}
			} else { // diff < strLen
				rt.moveAllToChild(str[diff:], str[:diff])
			}
			newNode := &router{str: path[diff:]}
			if !rt.addChild(first, newNode) {
				return nil, false
			}
			rt = newNode
		} else if diff < strLen {
			rt.moveAllToChild(str[diff:], path)
		}
	}
	return rt, true
}

// moveAllToChild move all attributes to a new node, and make this new node
//  as one of it's child
func (rt *router) moveAllToChild(childStr string, newStr string) {
	rnCopy := &router{
		str:       childStr,
		chars:     rt.chars,
		childs:    rt.childs,
		processor: rt.processor,
	}
	rt.chars, rt.childs, rt.processor = nil, nil, nil
	rt.addChild(childStr[0], rnCopy)
	rt.str = newStr
}

// addChild add an child, all childs is sorted
func (rt *router) addChild(b byte, n *router) bool {
	chars, childs := rt.chars, rt.childs
	l := len(chars)
	if l > 0 && chars[l-1] >= _WILDCARD && b >= _WILDCARD {
		return false
	}
	chars, childs = make([]byte, l+1), make([]*router, l+1)
	copy(chars, rt.chars)
	copy(childs, rt.childs)
	for ; l > 0 && chars[l-1] > b; l-- {
		chars[l], childs[l] = chars[l-1], childs[l-1]
	}
	chars[l], childs[l] = b, n
	rt.chars, rt.childs = chars, childs
	return true
}

// path character < _WILDCARD < _REMAINSALL
const (
	_MATCH_WILDCARD = ':' // MUST BE:other character < _WILDCARD < _REMAINSALL
	// _WILDCARD is the replacement of named variable in compiled path
	_WILDCARD         = '|' // MUST BE:other character < _WILDCARD < _REMAINSALL
	_MATCH_REMAINSALL = '*'
	// _REMAINSALL is the replacement of catch remains all variable in compiled path
	_REMAINSALL = '~'
	// _PRINT_SEP is the seperator of tree level when print route tree
	_PRINT_SEP = "-"
)

// matchMultiple match multi route node
// returned value:(first:next path start index, second:if continue, it's next node to match,
// else it's final match node, last:whether continu match)
func (rt *router) matchMultiple(path string, pathIndex int, values []string) (int,
	[]string, *router, bool) {
	str, strIndex := rt.str, 0
	strLen, pathLen := len(str), len(path)
	for strIndex < strLen {
		if pathIndex != pathLen {
			c := str[strIndex]
			strIndex++
			switch c {
			case path[pathIndex]: // else check character MatchPath or not
				pathIndex++
			case _WILDCARD:
				// if read '*', MatchPath until next '/'
				start := pathIndex
				for pathIndex < pathLen && path[pathIndex] != '/' {
					pathIndex++
				}
				values = append(values, path[start:pathIndex])
			case _REMAINSALL: // parse end, full matched
				values = append(values, path[pathIndex:pathLen])
				pathIndex = pathLen
				strIndex = strLen
			default:
				return -1, nil, nil, false // not matched
			}
		} else {
			return -1, nil, nil, false // path parse end
		}
	}
	var continu bool
	if pathIndex != pathLen { // path not parse end, to find a child node to continue
		var node *router
		p := path[pathIndex]
		for i, c := range rt.chars {
			if c == p || c >= _WILDCARD {
				node = rt.childs[i] // child
				break
			}
		}
		rt = node
		if node != nil {
			continu = true
		}
	}
	return pathIndex, values, rt, continu
}

// matchOne match one longest route node and return values of path variable
func (rt *router) matchOne(path string, values []string) (*router, []string) {
	var (
		str                string
		strIndex, strLen   int
		pathIndex, pathLen = 0, len(path)
		node               = rt
	)
	for node != nil {
		str, strIndex = rt.str, 0
		strLen = len(str)
		for strIndex < strLen {
			if pathIndex != pathLen {
				c := str[strIndex]
				strIndex++
				switch c {
				case path[pathIndex]: // else check character MatchPath or not
					pathIndex++
				case _WILDCARD:
					// if read '*', MatchPath until next '/'
					start := pathIndex
					for pathIndex < pathLen && path[pathIndex] != '/' {
						pathIndex++
					}
					values = append(values, path[start:pathIndex])
				case _REMAINSALL: // parse end, full matched
					values = append(values, path[pathIndex:pathLen])
					return rt, values
				default:
					return nil, nil // not matched
				}
			} else {
				return nil, nil // path parse end
			}
		}
		node = nil
		if pathIndex != pathLen { // path not parse end, must find a child node to continue
			p := path[pathIndex]
			for i, c := range rt.chars {
				if c == p || c >= _WILDCARD {
					node = rt.childs[i] // child
					break
				}
			}
			rt = node // child to parse
		} /* else { path parse end, node is the last matched node }*/
	}
	return rt, values
}

// isInvalidSection check whether section has the predefined _WILDCARD and match
// all character
func isInvalidSection(s string) bool {
	for _, c := range s {
		if bc := byte(c); bc >= _WILDCARD {
			return true
		}
	}
	return false
}

var (
	// nilVars is empty variable map
	nilVars = make(map[string]int)
)

// compile compile a url path to a clean path that replace all named variable
// to _WILDCARD or _REMAINSALL and extract all variable names
// if just want to match and don't need variable value, only use ':' or '*'
// for ':', it will catch the single section of url path seperated by '/'
// for '*', it will catch all remains url path, it should appear in the last
// of pattern for variables behind it will all be ignored
func compile(path string) (newPath string, vars map[string]int, err error) {
	path = strings.TrimSpace(path)
	l := len(path)
	if l == 0 || path[0] != '/' {
		return "", nil, Errorf("Invalid url pattern: %s, must start with '/'", path)
	}
	if l != 1 && path[l-1] == '/' {
		path = path[:l-1]
	}
	sections := strings.Split(path[1:], "/")
	new := make([]byte, 0, len(path))
	varIndex := 0
	for _, s := range sections {
		new = append(new, '/')
		last := len(s)
		i := last - 1
		var c byte
		for ; i >= 0; i-- {
			if s[i] == _MATCH_WILDCARD {
				c = _WILDCARD
			} else if s[i] == _MATCH_REMAINSALL {
				c = _REMAINSALL
			} else {
				continue
			}
			if name := s[i+1:]; len(name) > 0 {
				if isInvalidSection(name) {
					goto ERROR
				}
				if vars == nil {
					vars = make(map[string]int)
				}
				vars[name] = varIndex
			}
			varIndex++
			last = i
			break
		}
		if last != 0 {
			new = append(new, []byte(s[:last])...)
		}
		if c != 0 {
			new = append(new, c)
		}
	}
	newPath = string(new)
	if vars == nil {
		vars = nilVars
	}
	return
ERROR:
	return "", nil,
		Errorf("path %s has pre-defined characters %c or %c",
			path, _WILDCARD, _REMAINSALL)
}

// PrintRouteTree print an route tree
// every level will be seperated by "-"
func (rt *router) PrintRouteTree(w io.Writer) {
	rt.printRouteTree(w, "")
}

// printRouteTree print route tree with given parent path
func (rt *router) printRouteTree(w io.Writer, parentPath string) {
	if parentPath != "" {
		parentPath = parentPath + _PRINT_SEP
	}
	s := []byte(rt.str)
	for i := range s {
		if s[i] == _WILDCARD {
			s[i] = _MATCH_WILDCARD
		} else if s[i] == _REMAINSALL {
			s[i] = _MATCH_REMAINSALL
		}
	}
	cur := parentPath + string(s)
	if _, e := sys.WriteStrln(w, cur); e == nil {
		rt.accessAllChilds(func(n *router) bool {
			n.printRouteTree(w, cur)
			return true
		})
	}
}

// accessAllChilds access all childs of node
func (rt *router) accessAllChilds(fn func(*router) bool) {
	for _, n := range rt.childs {
		if !fn(n) {
			break
		}
	}
}

func _tmpGetFuncHandler(pattern string) *funcHandler {
	h := TmpHGet("funcHandlers", pattern)
	if h == nil {
		return nil
	}
	return h.(*funcHandler)
}

func _tmpSetFuncHandler(pattern string, handler *funcHandler) {
	TmpHSet("funcHandlers", pattern, handler)
}
