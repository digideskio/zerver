package zerver

type (
	TaskHandlerFunc func(interface{})

	TaskHandler interface {
		ServerInitializer
		Destroy()
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

func (TaskHandlerFunc) Init(*Server) error         { return nil }
func (fn TaskHandlerFunc) Handle(task interface{}) { fn(task) }
func (TaskHandlerFunc) Destroy()                   {}
