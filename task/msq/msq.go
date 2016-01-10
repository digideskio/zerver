package msq

import (
	"github.com/cosiner/gohper/bytes2"
	"github.com/cosiner/gohper/errors"
	"github.com/cosiner/gohper/sync2"
	log "github.com/cosiner/ygo/jsonlog"
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
	closeFlag sync2.Flag
	log       *log.Logger
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
	m.log = log.Derive("TaskHandler", "MessageQueue")
	go m.start()
	return nil
}

func (m *Queue) Handle(msg zerver.Task) {
	if m.closeFlag.IsTrue() || msg == nil {
		return
	}

	if m.EnableTypeChecking {
		m.TypeChecking(msg)
	}
	m.queue <- msg
}

func (m *Queue) process(msg zerver.Task) {
	err := m.Process(msg.Value())
	if err != nil {
		m.log.Error(log.M{"msg": "process message failed", "err": err.Error(), "pattern": msg.Pattern()})
	}
}

func (m *Queue) start() {
	for msg := range m.queue {
		m.process(msg)
	}
}

func (m *Queue) waitEmpty() {
	for {
		select {
		case msg, ok := <-m.queue:
			if !ok {
				return
			}
			m.process(msg)
		default:
			return
		}
	}
}

func (m *Queue) Destroy() {
	if !m.closeFlag.MakeTrue() {
		return
	}

	m.waitEmpty()
	close(m.queue)
}
