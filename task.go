package zerver

type (
	Task interface {
		patternKeeper
		Value() interface{}
	}

	TaskHandlerFunc func(Task)

	TaskHandler interface {
		Component
		Handle(Task)
	}

	task struct {
		patternString
		value interface{}
	}
)

func newTask(pattern string, value interface{}) Task {
	return task{
		patternString: patternString(pattern),
		value:         value,
	}
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

func (TaskHandlerFunc) Init(Env) error      { return nil }
func (fn TaskHandlerFunc) Handle(task Task) { fn(task) }
func (TaskHandlerFunc) Destroy()            {}
