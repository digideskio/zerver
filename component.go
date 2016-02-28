package zerver

import (
	"sync"

	"github.com/cosiner/gohper/encoding"
	"github.com/cosiner/gohper/errors"
	log "github.com/cosiner/ygo/jsonlog"
)

// =============================================================================
//                                  Component State
// =============================================================================
type compState uint8

const (
	_UNINITIALIZE compState = iota
	_WAITING
	_INITIALIZED
)

func (s compState) String() string {
	switch s {
	case _UNINITIALIZE:
		return "Uninitialize"
	case _WAITING:
		return "Initializing"
	case _INITIALIZED:
		return "Initialized"
	}

	panic("unexpected initial state")
}

// =============================================================================
//                                  Component
// =============================================================================
type (
	// Env is a server environment, real implementation is the Server itself.
	Env interface {
		Server() *Server
		Filepath(path string) string
		StartTask(path string, value interface{})
		Component(name string) (interface{}, error)
		Codec() encoding.Codec
		Logger() *log.Logger
	}

	// Component is a Object which will automaticlly initial/destroyed by server
	// if it's added to server, else it should init manually
	Component interface {
		Init(Env) error
		Destroy()
	}

	NopComponent struct{}
)

func (NopComponent) Init(Env) error { return nil }

func (NopComponent) Destroy() {}

// =============================================================================
//                                  Component Environment
// =============================================================================
type CompEnv struct {
	comp  Component
	value interface{}

	name string
	Env

	state compState
}

func newCompEnv(env Env, name string, c interface{}) *CompEnv {
	e := &CompEnv{
		name: name,
		Env:  env,
	}

	switch c := c.(type) {
	case Component:
		e.comp = c
		e.state = _UNINITIALIZE
	default:
		e.value = c
		e.state = _INITIALIZED
	}

	return e
}

func (e *CompEnv) Name() string {
	return e.name
}

func (e *CompEnv) String() string {
	return e.name + ":" + e.state.String()
}

func ComponentAttr(compName, attr string) string {
	return compName + ":" + attr
}

func (e *CompEnv) Attr(name string) interface{} {
	return e.Server().Attr(ComponentAttr(e.name, name))
}

func (e *CompEnv) SetAttr(name string, value interface{}) {
	e.Server().SetAttr(ComponentAttr(e.name, name), value)
}

func (e *CompEnv) GetSetAttr(name string, val interface{}) interface{} {
	return e.Server().GetSetAttr(ComponentAttr(e.name, name), val)
}

func (e *CompEnv) Init(Env) error {
	if e.state == _INITIALIZED {
		return nil
	}

	if e.state == _WAITING {
		panic("Cycle dependence on " + e.name)
	}

	e.state = _WAITING
	err := e.comp.Init(e)
	e.state = _INITIALIZED

	return err
}

func (e *CompEnv) Destroy() {
	if e.value == nil && e.state == _INITIALIZED {
		e.comp.Destroy()
	}
}

func (e *CompEnv) underlay() interface{} {
	if e.value != nil {
		return e.value
	}

	return e.comp
}

// =============================================================================
//                                  Component Manager
// =============================================================================
var ErrCompNotFound = errors.New("component not found")

type CompManager struct {
	components map[string]*CompEnv
	anonymous  []Component
	mu         sync.RWMutex
}

func NewCompManager() CompManager {
	return CompManager{
		components: make(map[string]*CompEnv),
	}
}

func (m *CompManager) Get(name string) (interface{}, error) {
	m.mu.RLock()
	e, has := m.components[name]
	m.mu.RUnlock()

	if !has {
		return nil, ErrCompNotFound
	}

	if err := e.Init(e); err != nil { // only first time will execute
		return nil, err
	}

	return e.underlay(), nil
}

// Register make a component managed
func (m *CompManager) Register(env Env, name string, comp interface{}) *CompEnv {
	m.mu.Lock()
	defer m.mu.Unlock()

	if name == "" {
		if c, is := comp.(Component); is {
			m.anonymous = append(m.anonymous, c)
		} else {
			panic("non-component object shouldn't be add to manager anonymously")
		}
		return nil
	}

	cs := newCompEnv(env, name, comp)
	m.components[name] = cs
	return cs
}

// Remove will an component and Destroy it
func (m *CompManager) Remove(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	cs, has := m.components[name]
	if has {
		cs.Destroy()
		delete(m.components, name)
	}
}

func (m *CompManager) Init(e Env) error {
	// initial named component first for anonymous may depend on them
	for _, comp := range m.components {
		if err := comp.Init(e); err != nil {
			return err
		}
	}

	for _, c := range m.anonymous {
		if err := c.Init(e); err != nil {
			return err
		}
	}

	return nil
}

func (m *CompManager) Destroy() {
	m.mu.Lock()
	for _, cs := range m.components {
		cs.Destroy()
	}

	for _, c := range m.anonymous {
		c.Destroy()
	}
	m.mu.Unlock()
}
