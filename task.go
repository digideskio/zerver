package zerver

type (
	// Task
	Task interface {
		URLVarIndexer
		Value() interface{}
		serverGetter
		destroy()
	}

	TaskHandlerFunc func(Task)

	TaskHandler interface {
		ServerInitializer
		Destroy()
		Handle(Task)
	}

	task struct {
		serverGetter
		URLVarIndexer
		value interface{}
	}
)

func newTask(s serverGetter, indexer URLVarIndexer, value interface{}) Task {
	return &task{
		serverGetter:  s,
		URLVarIndexer: indexer,
		value:         value,
	}
}

func (t *task) destroy() {
	t.URLVarIndexer.destroySelf()
}

func (t *task) Value() interface{} {
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

func (TaskHandlerFunc) Init(*Server) error  { return nil }
func (fn TaskHandlerFunc) Handle(task Task) { fn(task) }
func (TaskHandlerFunc) Destroy()            {}
