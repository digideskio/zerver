package zerver

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

	FakeComponent struct{}
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

func convertComponentEnv(e Enviroment, name string, c interface{}) *componentEnv {
	env := &componentEnv{
		name:         name,
		Enviroment:   e,
		initialState: _UNINITIALIZE,
	}

	switch c := c.(type) {
	case Component:
		env.comp = c
	default:
		env.value = c
	}

	return env
}

// NewComponentEnv is only a quick way get/set component attributes
func NewComponentEnv(env Enviroment, name string) ComponentEnviroment {
	return &componentEnv{
		Enviroment: env,
		name:       name,
	}
}

func (env *componentEnv) componentValue() interface{} {
	if env.value != nil {
		return env.value
	}

	return env.comp
}

func (env *componentEnv) Init(e Enviroment) error {
	if env.value != nil || env.initialState == _INITIALIZED {
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
	if env.value == nil {
		env.comp.Destroy()
	}
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

func ComponentAttr(comp, attr string) string {
	return comp + ":" + attr
}
