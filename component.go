package zerver

import (
	"sync"
)

type (
	// Component is a Object which will automaticlly initial/destroyed by server
	// if it's added to server, else it should initial manually
	Component interface {
		Init(Enviroment) error
		Destroy()
	}

	ComponentEnviroment interface {
		Name() string
		Attr(name string) interface{}
		RemoveAttr(name string)
		Enviroment
	}

	// ComponentState keeps states for global component, only Initialized, NoLazy, Comp
	// can setup
	ComponentState struct {
		Initialized bool
		NoLazy      bool
		Comp        Component
		value       interface{}

		name string
		Enviroment
		lock sync.RWMutex
	}

	FakeComponent struct{}
)

func (FakeComponent) Init(Enviroment) error { return nil }

func (FakeComponent) Destroy() {}

func convertComponentState(name string, c interface{}) *ComponentState {
	var cs *ComponentState

	switch c := c.(type) {
	case Component:
		cs = &ComponentState{
			Comp: c,
		}
	case ComponentState:
		cs = &c
	case *ComponentState:
		cs = c
	default:
		cs = &ComponentState{
			Initialized: true,
			value:       c,
		}
	}
	cs.name = name

	return cs
}

func (cs *ComponentState) Init(env Enviroment) (err error) {
	if cs.value != nil || cs.Initialized {
		return
	}

	cs.lock.Lock()

	if cs.Initialized {
		cs.lock.Unlock()
		return
	}

	cs.Enviroment = env
	defer cs.lock.Unlock()
	if err = cs.Comp.Init(cs); err == nil {
		cs.Initialized = true
	}
	cs.Enviroment = nil

	return
}

func (cs *ComponentState) Destroy() {
	if cs.value == nil && cs.Initialized {
		cs.lock.Lock()
		defer cs.lock.Unlock()
		cs.Comp.Destroy()
		cs.Initialized = false
	}
}

func (c *ComponentState) Name() string {
	return c.name
}

func (c *ComponentState) Attr(name string) interface{} {
	return c.Server().Attr(ComponentAttrName(c.name, name))
}

func (c *ComponentState) RemoveAttr(name string) {
	c.Server().RemoveAttr(ComponentAttrName(c.name, name))
}

func ComponentAttrName(comp, attr string) string {
	return comp + ":" + attr
}
