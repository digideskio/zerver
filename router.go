package zerver

import (
	"io"
	"log"
	"net/url"
	"strings"

	"github.com/cosiner/gohper/errors"
	"github.com/cosiner/gohper/strings2"
	"github.com/cosiner/gohper/unsafe2"
)

var (
	ErrConflictPathVar = errors.New("There is a similar route pattern which use same wildcard" +
		" or catchall at the same position, " +
		"this means one of them will nerver be matched, " +
		"please check your routes")
	ErrHandlerExists = errors.New("pattern handler already exists.")
)

type (
	Router interface {
		Component

		PrintRouteTree(w io.Writer)

		Filter(pattern string, f Filter) error
		FilterFunc(pattern string, f FilterFunc) error
		Handler(pattern string, h Handler) error
		TaskHandler(pattern string, th TaskHandler) error
		WsHandler(pattern string, th WsConn) error

		MatchHandlerFilters(url *url.URL) (Handler, string, ReqVars, []Filter)
		MatchWebSocketHandler(url *url.URL) (WsHandler, string, ReqVars)
		MatchTaskHandler(url *url.URL) (TaskHandler, string)
	}

	routeProcessor struct {
		handlerPattern string
		handlerVars    map[string]int
		handler        Handler

		wsHandlerPattern string
		wsHandlerVars    map[string]int
		wsHandler        WsHandler

		taskHandlerPattern string
		taskHandlerVars    map[string]int
		taskHandler        TaskHandler

		filters []Filter
	}

	router struct {
		str      string    // path section hold by current route node
		chars    []byte    // all possible first characters of next route node
		children []*router // child routers
		noFilter bool
		routeProcessor
	}
)

// NewRouter create a new Router
func NewRouter() Router {
	rt := new(router)
	rt.noFilter = true

	return rt
}

func (rt *router) Init(env Env) (err error) {
	if rt.handler != nil {
		err = rt.handler.Init(env)
	}

	for i := 0; i < len(rt.filters) && err == nil; i++ {
		err = rt.filters[i].Init(env)
	}
	if err == nil && rt.wsHandler != nil {
		err = rt.wsHandler.Init(env)
	}
	if err == nil && rt.taskHandler != nil {
		err = rt.taskHandler.Init(env)
	}
	for i := 0; i < len(rt.children) && err == nil; i++ {
		err = rt.children[i].Init(env)
	}

	return
}

func (rt *router) Destroy() {
	if rt.handler != nil {
		rt.handler.Destroy()
	}

	for _, f := range rt.filters {
		f.Destroy()
	}
	if rt.wsHandler != nil {
		rt.wsHandler.Destroy()
	}
	if rt.taskHandler != nil {
		rt.taskHandler.Destroy()
	}
	for _, c := range rt.children {
		c.Destroy()
	}
}

func (rt *router) FilterFunc(pattern string, f FilterFunc) error {
	return rt.register(pattern, f)
}

func (rt *router) Filter(pattern string, f Filter) error {
	return rt.register(pattern, f)
}

func (rt *router) Handler(pattern string, h Handler) error {
	return rt.register(pattern, h)
}

func (rt *router) TaskHandler(pattern string, th TaskHandler) error {
	return rt.register(pattern, th)
}

func (rt *router) WsHandler(pattern string, ws WsConn) error {
	return rt.register(pattern, ws)
}

func (rt *router) register(pattern string, processor interface{}) error {
	if processor == nil || pattern == "" {
		panic("nil handler or empty pattern is not allowed")
	}

	routePath, pathVars := compile(pattern)
	if r, is := processor.(*router); is {
		if !rt.addPathRouter(routePath, r) {
			return ErrHandlerExists
		}
		return nil
	}

	nrt, success := rt.addPath(routePath)
	if !success {
		return ErrConflictPathVar
	}
	if h, is := processor.(Handler); is {
		if nrt.handler != nil {
			return ErrHandlerExists
		}
		nrt.handler = h
		nrt.handlerVars = pathVars
		nrt.handlerPattern = pattern
		return nil
	}
	if f, is := processor.(Filter); is {
		rt.noFilter = false
		nrt.filters = append(nrt.filters, f)
		return nil
	}
	if ws, is := processor.(WsHandler); is {
		if nrt.wsHandler != nil {
			return ErrHandlerExists
		}
		nrt.wsHandler = ws
		nrt.wsHandlerVars = pathVars
		nrt.wsHandlerPattern = pattern
		return nil
	}
	if th, is := processor.(TaskHandler); is {
		if nrt.taskHandler != nil {
			return ErrHandlerExists
		}
		nrt.taskHandler = th
		nrt.taskHandlerVars = pathVars
		return nil
	}
	panic("unreachable")
}

func (rt *router) MatchWebSocketHandler(url *url.URL) (WsHandler, string, ReqVars) {
	path := url.Path
	vars := ReqVars{}
	rt, vars.urlVals = rt.matchOne(path, vars.urlVals)
	if rt == nil {
		return nil, "", vars
	}
	if rt.wsHandler == nil {
		return nil, rt.wsHandlerPattern, vars
	}
	vars.urlVars = rt.wsHandlerVars
	return rt.wsHandler, rt.wsHandlerPattern, vars
}

func (rt *router) MatchTaskHandler(url *url.URL) (TaskHandler, string) {
	if rt = rt.matchOnly(url.Path); rt == nil {
		return nil, ""
	}

	return rt.taskHandler, rt.taskHandlerPattern
}

// func (rt *router) MatchHandler(url *url.URL) (handler Handler, indexer URLVarIndexer) {
//  path := url.Path
//  indexer = newVarIndexerFromPool()
//  rt, values := rt.matchOne(path, indexer.values)
//  indexer.values = values
//  if rt != nil && rt.processor != nil {
//      if hp := p.handler; hp != nil {
//          indexer.vars = hp.vars
//          handler = hp.handler
//      }
//  }
//  return
// }

func (rt *router) MatchHandlerFilters(url *url.URL) (Handler, string, ReqVars, []Filter) {
	var (
		path    = url.Path
		vars    ReqVars
		filters []Filter
	)

	if rt.noFilter {
		rt, vars.urlVals = rt.matchOne(path, vars.urlVals)
	} else {
		pathIndex, continu := 0, true
		for continu {
			if fs := rt.filters; len(fs) != 0 {
				if filters == nil {
					filters = make([]Filter, 0, 3)
				}
				filters = append(filters, fs...)
			}
			pathIndex, vars.urlVals, rt, continu = rt.matchMultiple(path, pathIndex, vars.urlVals)
		}
	}
	if rt == nil {
		return nil, "", vars, filters
	}
	if rt.handler == nil {
		return nil, rt.handlerPattern, vars, filters
	}
	vars.urlVars = rt.handlerVars
	return rt.handler, rt.handlerPattern, vars, filters
}

// addPath add an new path to route, use given function to operate the final
// route node for this path
func (rt *router) addPath(path string) (*router, bool) {
	str := rt.str
	if str == "" && len(rt.chars) == 0 {
		rt.str = path
		return rt, true
	}

	diff, pathLen, strLen := 0, len(path), len(str)
	for diff != pathLen && diff != strLen && path[diff] == str[diff] {
		diff++
	}

	if diff < pathLen {
		first := path[diff]
		if diff == strLen {
			for i, c := range rt.chars {
				if c == first {
					return rt.children[i].addPath(path[diff:])
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

	return rt, true
}

// addPath add an new path to route, use given function to operate the final
// route node for this path
func (rt *router) addPathRouter(path string, r *router) bool {
	first, str := byte('/'), rt.str
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
						return rt.children[i].addPathRouter(path[diff:], r)
					}
				}
			} else { // diff < strLen
				rt.moveAllToChild(str[diff:], str[:diff])
			}

			r.str = path[diff:] + r.str
		} else if diff < strLen {
			if str[diff] == '/' {
				return false
			}

			rt.moveAllToChild(str[diff:], path)
		} else {
			for _, c := range rt.chars {
				if c == '/' {
					return false
				}
			}
		}
	}

	rt.addChild(first, r)
	return true
}

// moveAllToChild move all attributes to a new node, and make this new node
//  as one of it's child
func (rt *router) moveAllToChild(childStr string, newStr string) {
	rnCopy := &router{
		str:            childStr,
		chars:          rt.chars,
		children:       rt.children,
		routeProcessor: rt.routeProcessor,
	}

	rt.chars, rt.children, rt.routeProcessor = nil, nil, routeProcessor{}
	rt.addChild(childStr[0], rnCopy)
	rt.str = newStr
}

// addChild add an child, all childs is sorted
func (rt *router) addChild(b byte, n *router) bool {
	chars, childs := rt.chars, rt.children
	l := len(chars)
	if l > 0 && chars[l-1] >= _WILDCARD && b >= _WILDCARD {
		return false
	}

	chars, childs = make([]byte, l+1), make([]*router, l+1)
	copy(chars, rt.chars)
	copy(childs, rt.children)
	for ; l > 0 && chars[l-1] > b; l-- {
		chars[l], childs[l] = chars[l-1], childs[l-1]
	}
	chars[l], childs[l] = b, n
	rt.chars, rt.children = chars, childs

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

	if pathIndex != pathLen { // path not parse end, to find a child node to continue
		p := path[pathIndex]
		for i, c := range rt.chars {
			if c == p || c >= _WILDCARD {
				return pathIndex, values, rt.children[i], true
			}
		}
		rt = nil
	}

	return pathIndex, values, rt, false
}

// matchOne match one longest route node and return values of path variable
func (rt *router) matchOne(path string, values []string) (*router, []string) {
	var (
		str                string
		strIndex, strLen   int
		pathIndex, pathLen = 0, len(path)
	)

AGAIN:
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

	if pathIndex != pathLen { // path not parse end, must find a child node to continue
		p := path[pathIndex]
		for i, c := range rt.chars {
			if c == p || c >= _WILDCARD {
				rt = rt.children[i] // child
				goto AGAIN
			}
		}

		rt = nil // child to parse
	} /* else { path parse end, node is the last matched node }*/

	return rt, values
}

// matchOnly match one longest route node without parameter values
func (rt *router) matchOnly(path string) *router {
	var (
		str                string
		strIndex, strLen   int
		pathIndex, pathLen = 0, len(path)
	)

AGAIN:
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
				for pathIndex < pathLen && path[pathIndex] != '/' {
					pathIndex++
				}
			case _REMAINSALL: // parse end, full matched
				return rt
			default:
				return nil // not matched
			}
		} else {
			return nil // path parse end
		}
	}

	if pathIndex != pathLen { // path not parse end, must find a child node to continue
		p := path[pathIndex]
		for i, c := range rt.chars {
			if c == p || c >= _WILDCARD {
				rt = rt.children[i] // found child
				goto AGAIN
			}
		}
		rt = nil // not found
	} /* else { path parse end, node is the last matched node }*/

	return rt
}

// isInvalidSection check whether section has the predefined _WILDCARD and match
// all character
func isInvalidSection(s string) bool {
	var invalid bool

	for i := 0; i < len(s) && !invalid; i++ {
		invalid = s[i] >= _WILDCARD
	}

	return invalid
}

var (
	// emptyVars is empty variable map
	emptyVars = make(map[string]int)
)

// compile compile a url path to a clean path that replace all named variable
// to _WILDCARD or _REMAINSALL and extract all variable names
// if just want to match and don't need variable value, only use ':' or '*'
// for ':', it will catch the single section of url path seperated by '/'
// for '*', it will catch all remains url path, it should appear in the last
// of pattern for variables behind it will all be ignored
//
// the query portion will be trimmed
func compile(path string) (newPath string, vars map[string]int) {
	path = strings2.TrimAfter(path, "?")
	l := len(path)

	if path[0] != '/' {
		log.Panicln("Invalid pattern, must start with '/': " + path)
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
					log.Panicf("path %s has pre-defined characters %c or %c\n",
						path, _WILDCARD, _REMAINSALL)
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
		vars = emptyVars
	}

	return
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
	if _, e := w.Write(unsafe2.Bytes(cur + "\n")); e == nil {
		rt.accessAllChildren(func(n *router) bool {
			n.printRouteTree(w, cur)
			return true
		})
	}
}

// accessAllChildren access all children of node
func (rt *router) accessAllChildren(fn func(*router) bool) {
	for _, n := range rt.children {
		if !fn(n) {
			break
		}
	}
}
