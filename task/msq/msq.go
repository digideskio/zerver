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
	TaskBufsize uint
	Processor
	EnableTypeChecking bool
	NoRecover          bool
	BytesPool          bytes2.Pool

	queue     chan zerver.Task
	closeCond *sync2.LockCond
}

func (m *Queue) Init(env zerver.Env) error {
	if m.Processor == nil {
		return errors.Err("message processor shouldn't be nil")
	}
	if m.TaskBufsize == 0 {
		m.TaskBufsize = 256
	}
	if m.BytesPool == nil {
		m.BytesPool = bytes2.NewFakePool()
	}

	m.queue = make(chan zerver.Task, m.TaskBufsize)

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
		log.Error(e)

		buf := m.BytesPool.Get(4096, true)

		index := runtime.Stack(buf, false)
		buf = buf[:index]
		log.Error(unsafe2.String(buf))

		m.BytesPool.Put(buf)

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
				log.Error("msq "+msg.Pattern(), err)
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
