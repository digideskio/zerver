package zerver

import "sync"

const (
	_UNINITIALIZE initialState = iota
	_WAITING
	_INITIALIZED
)

type (
	initialState int

	// Component is a Object which will automaticlly initial/destroyed by server
	// if it's added to server, else it should initial manually
	Component interface {
		Init(Enviroment) error
		Destroy()
	}

	FakeComponent struct{}

	ComponentEnviroment interface {
		Name() string
		Attr(name string) interface{}
		SetAttr(name string, value interface{})
		GetSetAttr(name string, value interface{}) interface{}
		Enviroment
	}

	componentEnv struct {
		comp  Component
		value interface{}

		name string
		Enviroment

		initialState
	}

	componentManager struct {
		components  map[string]*componentEnv
		anonymouses []Component
		lock        sync.RWMutex

		initHook func(name string)
	}
)

func (s initialState) String() string {
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

func (FakeComponent) Init(Enviroment) error { return nil }

func (FakeComponent) Destroy() {}

// NewComponentEnv is only a quick way get/set component attributes
func NewComponentEnv(env Enviroment, name string) ComponentEnviroment {
	return newComponentEnv(env, name, nil)
}

func newComponentEnv(e Enviroment, name string, c interface{}) *componentEnv {
	env := &componentEnv{
		name:       name,
		Enviroment: e,
	}

	if c != nil {
		switch c := c.(type) {
		case Component:
			env.comp = c
			env.initialState = _UNINITIALIZE
		default:
			env.value = c
			env.initialState = _INITIALIZED
		}
	}

	return env
}

func (env *componentEnv) componentValue() interface{} {
	if env.value != nil {
		return env.value
	}

	return env.comp
}

func (env *componentEnv) Init(e Enviroment) error {
	if env.initialState == _INITIALIZED {
		return nil
	}

	if env.initialState == _WAITING {
		panic("Cycle dependence on " + env.name)
	}

	env.initialState = _WAITING
	err := env.comp.Init(env)
	env.initialState = _INITIALIZED

	return err
}

func (env *componentEnv) Destroy() {
	if env.value == nil && env.initialState == _INITIALIZED {
		env.comp.Destroy()
	}
}

func ComponentAttr(comp, attr string) string {
	return comp + ":" + attr
}

func (env *componentEnv) Name() string {
	return env.name
}

func (env *componentEnv) Attr(name string) interface{} {
	return env.Server().Attr(ComponentAttr(env.name, name))
}

func (env *componentEnv) SetAttr(name string, value interface{}) {
	env.Server().SetAttr(ComponentAttr(env.name, name), value)
}

func (env *componentEnv) GetSetAttr(name string, val interface{}) interface{} {
	return env.Server().GetSetAttr(ComponentAttr(env.name, name), val)
}

func (env *componentEnv) String() string {
	return env.name + ":" + env.initialState.String()
}

func newComponentManager() componentManager {
	return componentManager{
		components: make(map[string]*componentEnv),
	}
}

const (
	_GLOBAL_COMPONENT    = "_Global_"
	_ANONYMOUS_COMPONENT = ""
)

func (cm *componentManager) Init(env Enviroment) error {
	// initial named component first for anonymouses may depend on them
	hook := cm.initHook
	if hook == nil {
		hook = func(string) {}
	} else {
		defer func(cm *componentManager) {
			cm.initHook = nil
		}(cm)
	}

	hook(_GLOBAL_COMPONENT)
	for name, comp := range cm.components {
		hook(name)
		if err := comp.Init(env); err != nil {
			return err
		}
	}

	hook(_ANONYMOUS_COMPONENT)
	for _, c := range cm.anonymouses {
		if err := c.Init(env); err != nil {
			return err
		}
	}

	return nil
}

func (cm *componentManager) Destroy() {
	cm.lock.Lock()
	for _, cs := range cm.components {
		cs.Destroy()
	}

	for _, c := range cm.anonymouses {
		c.Destroy()
	}
	cm.lock.Unlock()
}

func (cm *componentManager) Component(name string) (interface{}, error) {
	cm.lock.RLock()
	env, has := cm.components[name]
	cm.lock.RUnlock()

	if !has {
		return nil, ComponentNotFoundError(name)
	}

	if err := env.Init(env); err != nil { // only first time will execute
		return nil, err
	}

	return env.componentValue(), nil
}

func (cm *componentManager) RegisterComponent(
	env Enviroment,
	name string,
	component interface{},
) ComponentEnviroment {

	if name == "" {
		if c, is := component.(Component); is {
			cm.anonymouses = append(cm.anonymouses, c)
		}
		return nil
	}

	cs := newComponentEnv(env, name, component)
	cm.lock.Lock()
	cm.components[name] = cs
	cm.lock.Unlock()

	return cs
}

func (cm *componentManager) RemoveComponent(name string) {
	cm.lock.Lock()
	cs, has := cm.components[name]
	if !has {
		cm.lock.Unlock()
		return
	}

	defer cm.lock.Unlock()
	cs.Destroy()
	delete(cm.components, name)
}
