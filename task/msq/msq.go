package msq

import (
	"github.com/cosiner/gohper/errors"
	"github.com/cosiner/zerver"
)

// Processor is the real message processor
type Processor interface {
	Process(interface{})
	TypeChecking(interface{})
}

// Queue is a simple case of TaskHandler
type Queue struct {
	queue   chan interface{}
	signal  chan struct{}
	Bufsize uint
	Processor
	EnableTypeChecking bool
}

func (m *Queue) Init(zerver.Environment) error {
	if m.Processor == nil {
		return errors.Err("message processor shouldn't be nil")
	}
	if m.Bufsize == 0 {
		m.Bufsize = 1024
	}

	m.queue = make(chan interface{}, m.Bufsize)
	m.signal = make(chan struct{})

	go func() {
		for {
			select {
			case msg := <-m.queue:
				m.Process(msg)
			case <-m.signal:
				return
			}
		}
	}()

	return nil
}

func (m *Queue) Handle(msg interface{}) {
	if msg != nil {
		if m.EnableTypeChecking {
			m.TypeChecking(msg)
		}
		m.queue <- msg
	}
}

func (m *Queue) Destroy() {
	m.signal <- struct{}{}
}
