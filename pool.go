package zerver

import (
	"sync"

	"github.com/cosiner/gohper/errors"
)

// global variables need to be initialed by ServerOption
var (
	// pathVarCount is common url path variable count
	// match functions of router will create a slice use it as capcity to store
	// all path variable values
	// to get best performance, it should commonly set to the average, default, it's 2
	pathVarCount int
	filterCount  int
)

type requestEnv struct {
	req  request
	resp response
}

type serverPool struct {
	requestEnvPool sync.Pool
	varIndexerPool sync.Pool
	filtersPool    sync.Pool
	otherPools     map[int]*sync.Pool
}

var _defaultPool *serverPool

func init() {
	_defaultPool = &serverPool{otherPools: make(map[int]*sync.Pool)}
	_defaultPool.requestEnvPool.New = func() interface{} {
		env := &requestEnv{}
		env.req.AttrContainer = NewAttrContainer()
		return env
	}
	_defaultPool.varIndexerPool.New = func() interface{} {
		return &urlVarIndexer{values: make([]string, 0, pathVarCount)}
	}
	_defaultPool.filtersPool.New = func() interface{} {
		return make([]Filter, 0, filterCount)
	}
}

func ReigisterPool(id int, newFunc func() interface{}) error {
	op := _defaultPool.otherPools
	if _, has := op[id]; has {
		return errors.New("Pool for ", id, " already exist")
	}
	op[id] = &sync.Pool{New: newFunc}
	return nil
}

func NewFrom(id int) interface{} {
	return _defaultPool.otherPools[id].Get()
}

func newRequestEnvFromPool() *requestEnv {
	return _defaultPool.requestEnvPool.Get().(*requestEnv)
}

func newVarIndexerFromPool() *urlVarIndexer {
	return _defaultPool.varIndexerPool.Get().(*urlVarIndexer)
}

func newFiltersFromPool() []Filter {
	return _defaultPool.filtersPool.Get().([]Filter)
}

func recycleRequestEnv(req *requestEnv) {
	_defaultPool.requestEnvPool.Put(req)
}

func recycleVarIndexer(indexer URLVarIndexer) {
	_defaultPool.varIndexerPool.Put(indexer)
}

func recycleFilters(filters []Filter) {
	if filters != nil {
		filters = filters[:0]
		_defaultPool.filtersPool.Put(filters)
	}
}

func RecycleTo(id int, value interface{}) {
	_defaultPool.otherPools[id].Put(value)
}
