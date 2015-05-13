package zerver

import (
	"github.com/cosiner/gohper/errors"
)

type (
	TaskHandlerFunc func(interface{})

	TaskHandler interface {
		Component
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

func (TaskHandlerFunc) Init(Enviroment) error      { return nil }
func (fn TaskHandlerFunc) Handle(task interface{}) { fn(task) }
func (TaskHandlerFunc) Destroy()                   {}

// MessageProcessor is the real message processor
type MessageProcessor func(interface{})

// MessageQueue is a simple case of TaskHandler
type MessageQueue struct {
	queue   chan interface{}
	signal  chan byte
	Bufsize uint
	MessageProcessor
}

func (m *MessageQueue) Init(Enviroment) error {
	if m.MessageProcessor == nil {
		return errors.Err("message processor shouldn't be nil")
	}
	if m.Bufsize == 0 {
		m.Bufsize = 1024
	}

	m.queue = make(chan interface{}, m.Bufsize)
	m.signal = make(chan byte)

	go func() {
		for {
			select {
			case msg := <-m.queue:
				m.MessageProcessor(msg)
			case <-m.signal:
				return
			}
		}
	}()

	return nil
}

func (m *MessageQueue) Handle(msg interface{}) {
	if msg != nil {
		m.queue <- msg
	}
}

func (m *MessageQueue) Destroy() {
	m.signal <- 1
}
