package zerver

import (
	"testing"

	"github.com/cosiner/gohper/testing2"
)

type Dep string

func (d Dep) Init(env Environment) error {
	_, _ = env.Component(string(d))
	return nil
}

func (d Dep) Destroy() {}

func TestCycleDependenced(t *testing.T) {
	defer testing2.Recover(t)

	s := NewServer()

	s.RegisterComponent("Comp1", Dep("Comp2"))
	s.RegisterComponent("Comp2", Dep("Comp1"))
	for _, comp := range s.components {
		comp.Init(s)
	}
}
