package zerver

type Logger interface {
	Flush()
	Close()

	Debug(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Error(...interface{}) // panic goroutine
	Fatal(...interface{}) // exit process

	Debugf(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Errorf(string, ...interface{})
	Fatalf(string, ...interface{})

	Debugln(...interface{})
	Infoln(...interface{})
	Warnln(...interface{})
	Errorln(...interface{})
	Fatalln(...interface{})

	// current function's depth is 0, parant's is 1, etc..
	DebugDepth(int, ...interface{})
	InfoDepth(int, ...interface{})
	WarnDepth(int, ...interface{})
	ErrorDepth(int, ...interface{})
	FatalDepth(int, ...interface{})
}
