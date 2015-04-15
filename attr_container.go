package zerver

import "sync"

type (
	// AttrContainer is a common container store attribute
	AttrContainer interface {
		Attr(name string) interface{}
		SetAttr(name string, value interface{})
		UpdateAttr(name string, value interface{}) bool
		RemoveAttr(name string)
		IsAttrExist(name string) bool
		Attrs() Values
		AccessAllAttrs(fn func(Values))
		Clear()
	}

	Values map[string]interface{}

	lockedValues struct {
		Values
		sync.RWMutex
	}
)

func NewValues() Values {
	return make(Values)
}

func NewLockedAttrContainer() AttrContainer {
	return &lockedValues{
		Values: NewValues(),
	}
}

func NewLockedAttrContainerWith(v Values) AttrContainer {
	return &lockedValues{
		Values: v,
	}
}

func NewAttrContainer() AttrContainer {
	return make(Values)
}

func NewAttrContainerWith(v Values) AttrContainer {
	return v
}

func (v Values) IsAttrExist(key string) bool {
	_, has := v[key]
	return has
}

func (v Values) Attr(key string) interface{} {
	val, has := v[key]
	if !has {
		return nil
	}
	return val
}

func (v Values) RemoveAttr(key string) {
	delete(v, key)
}

func (v Values) SetAttr(key string, val interface{}) {
	if val != nil {
		v[key] = val
	}
}

func (v Values) UpdateAttr(key string, val interface{}) (s bool) {
	if s = v.IsAttrExist(key); s {
		v[key] = val
	}
	return
}

func (v Values) Attrs() Values {
	return v
}

func (v Values) AccessAllAttrs(fn func(Values)) {
	fn(v)
}

func (v Values) Clear() {
	for k := range v {
		delete(v, k)
	}
}

func (lc *lockedValues) Attr(key string) (val interface{}) {
	lc.RLock()
	val = lc.Values.Attr(key)
	lc.RUnlock()
	return
}

func (lc *lockedValues) IsAttrExist(key string) bool {
	lc.RLock()
	has := lc.Values.IsAttrExist(key)
	lc.RUnlock()
	return has
}

func (lc *lockedValues) RemoveAttr(key string) {
	lc.Lock()
	lc.Values.RemoveAttr(key)
	lc.Unlock()
}

func (lc *lockedValues) SetAttr(key string, val interface{}) {
	lc.Lock()
	lc.Values.SetAttr(key, val)
	lc.Unlock()
}

func (lc *lockedValues) UpdateAttr(key string, val interface{}) (s bool) {
	lc.Lock()
	s = lc.Values.UpdateAttr(key, val)
	lc.Unlock()
	return
}

func (lc *lockedValues) Attrs() Values {
	return lc.Values
}

func (lc *lockedValues) AccessAllAttrs(fn func(Values)) {
	lc.RLock()
	lc.Values.AccessAllAttrs(fn)
	lc.RUnlock()
}

func (lc *lockedValues) Clear() {
	lc.Lock()
	lc.Values.Clear()
	lc.Unlock()
}
