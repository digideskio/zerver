package zerver

import (
	"reflect"

	. "github.com/cosiner/gohper/lib/errors"

	ref "github.com/cosiner/gohper/lib/reflect"
)

type (
	// URLVarIndexer is a indexer for name to value
	URLVarIndexer interface {
		// URLVar return value of variable
		URLVar(name string) string
		URLVarDef(name string, defvalue string) string
		ScanURLVar(name string, addr interface{}) error
		destroySelf() // avoid confilict with Request interface
	}

	// urlVarIndexer is an implementation of URLVarIndexer
	urlVarIndexer struct {
		vars   map[string]int // url variables and indexs of sections splited by '/'
		values []string       // all url variable values
	}
)

func (v *urlVarIndexer) destroySelf() {
	v.values = v.values[:0]
	v.vars = nil
	Pool.recycleVarIndexer(v)
}

// URLVar return values of variable
func (v *urlVarIndexer) URLVar(name string) string {
	if index, has := v.vars[name]; has {
		return v.values[index]
	}
	return ""
}

// URLVar return values of variable
func (v *urlVarIndexer) URLVarDef(name string, defvalue string) string {
	if index, has := v.vars[name]; has {
		return v.values[index]
	}
	return defvalue
}

// ScanURLVars scan values into variable addresses
// if address is nil, skip it
func (v *urlVarIndexer) ScanURLVar(name string, addr interface{}) error {
	if index, has := v.vars[name]; has {
		return ref.UnmarshalPrimitive(v.values[index], reflect.ValueOf(addr))
	}
	return Err("No this variable: " + name)
}
