package component

import (
	"github.com/cosiner/zerver"
)

type ComponentInfo struct {
	Name        string
	Initialized bool
	NoLazy      bool
	zerver.Component

	OptionName string
	Option     interface{}
}

func Register(s *zerver.Server, info ComponentInfo) error {
	if info.OptionName != "" {
		s.SetAttr(info.OptionName, info.Option)
	}
	return s.AddComponent(info.Name, zerver.ComponentState{
		Initialized: info.Initialized,
		NoLazy:      info.NoLazy,
		Component:   info.Component,
	})
}
