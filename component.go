package zerver

type (
	// Component is a Object which will automaticlly initial/destroyed by server
	// if it's added to server, else it should initial manually
	Component interface {
		Init(Enviroment) error
		Destroy()
	}

	ComponentState struct {
		Initialized bool
		NoLazy      bool
		Component

		value interface{}
	}

	FakeComponent struct{}
)

func (FakeComponent) Init(Enviroment) error { return nil }

func (FakeComponent) Destroy() {}

func convertComponentState(c interface{}) ComponentState {
	switch c := c.(type) {
	case Component:
		return ComponentState{
			Component: c,
		}
	case ComponentState:
		return c
	case *ComponentState:
		return *c
	default:
		return ComponentState{
			Initialized: true,
			value:       c,
		}
	}
}
