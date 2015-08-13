package msq

import (
	"runtime"

	"github.com/cosiner/gohper/bytes2"
	"github.com/cosiner/gohper/errors"
	"github.com/cosiner/gohper/sync2"
	"github.com/cosiner/gohper/unsafe2"
	"github.com/cosiner/ygo/log"
	"github.com/cosiner/zerver"
)

// Processor is the real message processor
type Processor interface {
	Process(interface{}) error
	TypeChecking(interface{})
}

// Queue is a simple case of TaskHandler
type Queue struct {
	Bufsize uint
	Processor
	EnableTypeChecking bool
	ErrorLogger        log.Logger
	NoRecover          bool
	BytesPool          bytes2.Pool

	queue     chan zerver.Task
	closeCond *sync2.LockCond
}

func (m *Queue) Init(zerver.Environment) error {
	if m.Processor == nil {
		return errors.Err("message processor shouldn't be nil")
	}
	if m.Bufsize == 0 {
		m.Bufsize = 1024
	}
	if m.ErrorLogger != nil {
		m.ErrorLogger = m.ErrorLogger.Prefix("zerver/msq")
	}
	if m.BytesPool == nil {
		m.BytesPool = bytes2.NewFakePool()
	}

	m.queue = make(chan zerver.Task, m.Bufsize)

	go m.start()
	return nil
}

func (m *Queue) Handle(msg zerver.Task) {
	if m.closeCond != nil || msg == nil {
		return
	}

	if m.EnableTypeChecking {
		m.TypeChecking(msg)
	}
	m.queue <- msg
}

func (m *Queue) start() {
	defer func(m *Queue) {
		e := recover()
		if e == nil {
			return
		}
		if m.ErrorLogger != nil {
			m.ErrorLogger.Errorln(e)

			buf := m.BytesPool.Get(4096, true)

			index := runtime.Stack(buf, false)
			buf = buf[:index]
			m.ErrorLogger.Errorln(unsafe2.String(buf))

			m.BytesPool.Put(buf)
		}
		if !m.NoRecover {
			go m.start()
		}
	}(m)
	for {
		select {
		case msg, ok := <-m.queue:
			if !ok {
				return
			}
			err := m.Process(msg.Value())
			if err != nil {
				m.ErrorLogger.Errorln(msg.Pattern(), ":", err)
			}

			if len(m.queue) == 0 && m.closeCond != nil {
				m.closeCond.Signal()
			} else {
				break
			}
		}
	}
}

func (m *Queue) Destroy() {
	if m.closeCond != nil {
		return
	}

	m.closeCond = sync2.NewLockCond(nil)
	if len(m.queue) > 0 {
		m.closeCond.Wait()
	}
	close(m.queue)
}
