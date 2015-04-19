package zerver

import "testing"

type StringComponent string

func (t StringComponent) Init(e Enviroment) error {
	_, _ = e.Component(string(t))
	return nil
}

func (t StringComponent) Destroy() {}

type initial struct {
	MapHandler
}

func (i initial) Init(e Enviroment) error {
	e.Component("foo")
	return nil
}

func TestCircularDependence(t *testing.T) {
	s := NewServer()
	s.AddComponent("foo",
		ComponentState{
			Component: StringComponent("bar"),
		})
	s.AddComponent("bar",
		ComponentState{
			Component: StringComponent("foo"),
		})
	s.Handle("/", initial{})
	s.Start(nil)
}
