package zerver

import (
	"github.com/cosiner/gohper/log"
)

type Logger interface {
	Flush()
	Close()

	Trace(...interface{})
	Info(...interface{})
	Warn(...interface{})
	Panic(...interface{}) // panic goroutine
	Fatal(...interface{}) // exit process

	Tracef(string, ...interface{})
	Infof(string, ...interface{})
	Warnf(string, ...interface{})
	Panicf(string, ...interface{})
	Fatalf(string, ...interface{})

	Traceln(...interface{})
	Infoln(...interface{})
	Warnln(...interface{})
	Panicln(...interface{})
	Fatalln(...interface{})

	// current function's depth is 0, parant's is 1, etc..
	TraceDepth(int, ...interface{})
	InfoDepth(int, ...interface{})
	WarnDepth(int, ...interface{})
	PanicDepth(int, ...interface{})
	FatalDepth(int, ...interface{})
}

func DefaultLogger() Logger {
	return log.Default()
}
