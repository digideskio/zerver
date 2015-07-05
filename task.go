package zerver

type (
	TaskHandlerFunc func(interface{})

	TaskHandler interface {
		Component
		Handle(interface{})
	}
)

func convertTaskHandler(i interface{}) TaskHandler {
	switch t := i.(type) {
	case func(interface{}):
		return TaskHandlerFunc(t)
	case TaskHandler:
		return t
	}
	return nil
}

func (TaskHandlerFunc) Init(Environment) error     { return nil }
func (fn TaskHandlerFunc) Handle(task interface{}) { fn(task) }
func (TaskHandlerFunc) Destroy()                   {}
