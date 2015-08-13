package zerver

type (
	Task interface {
		Pattern() string
		Value() interface{}
	}

	TaskHandlerFunc func(Task)

	TaskHandler interface {
		Component
		Handle(Task)
	}

	task struct {
		pattern string
		value   interface{}
	}
)

func newTask(pattern string, value interface{}) Task {
	return task{
		pattern: pattern,
		value:   value,
	}
}

func (t task) Pattern() string {
	return t.pattern
}

func (t task) Value() interface{} {
	return t.value
}

func convertTaskHandler(i interface{}) TaskHandler {
	switch t := i.(type) {
	case func(Task):
		return TaskHandlerFunc(t)
	case TaskHandler:
		return t
	}
	return nil
}

func (TaskHandlerFunc) Init(Environment) error { return nil }
func (fn TaskHandlerFunc) Handle(task Task)    { fn(task) }
func (TaskHandlerFunc) Destroy()               {}
