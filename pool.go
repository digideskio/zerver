package zerver

import (
	. "github.com/cosiner/gohper/lib/errors"

	"sync"
)

type requestEnv struct {
	req  request
	resp response
}

type ServerPool struct {
	requestEnvPool  sync.Pool
	varIndexerPool  sync.Pool
	filtersPool     sync.Pool
	filterChainPool sync.Pool
	otherPools      map[string]*sync.Pool
}

var Pool *ServerPool

func init() {
	Pool = &ServerPool{otherPools: make(map[string]*sync.Pool)}
	Pool.requestEnvPool.New = func() interface{} {
		env := &requestEnv{}
		env.req.AttrContainer = NewAttrContainer()
		return env
	}
	Pool.varIndexerPool.New = func() interface{} {
		return &urlVarIndexer{values: make([]string, 0, PathVarCount)}
	}
	Pool.filtersPool.New = func() interface{} {
		return make([]Filter, 0, FilterCount)
	}
	Pool.filterChainPool.New = func() interface{} {
		return new(filterChain)
	}
}

func (pool *ServerPool) ReigisterPool(name string, newFunc func() interface{}) error {
	op := pool.otherPools
	if _, has := op[name]; has {
		return Err("Pool for " + name + " already exist")
	}
	op[name] = &sync.Pool{New: newFunc}
	return nil
}

func (pool *ServerPool) NewFrom(name string) interface{} {
	return pool.otherPools[name].Get()
}

func (pool *ServerPool) newRequestEnv() *requestEnv {
	return pool.requestEnvPool.Get().(*requestEnv)
}

func (pool *ServerPool) newVarIndexer() *urlVarIndexer {
	return pool.varIndexerPool.Get().(*urlVarIndexer)
}

func (pool *ServerPool) newFilters() []Filter {
	return pool.filtersPool.Get().([]Filter)
}

func (pool *ServerPool) newFilterChain() *filterChain {
	return pool.filterChainPool.Get().(*filterChain)
}

func (pool *ServerPool) recycleRequestEnv(req *requestEnv) {
	pool.requestEnvPool.Put(req)
}

func (pool *ServerPool) recycleVarIndexer(indexer URLVarIndexer) {
	pool.varIndexerPool.Put(indexer)
}

func (pool *ServerPool) recycleFilters(filters []Filter) {
	if filters != nil {
		filters = filters[:0]
		pool.filtersPool.Put(filters)
	}
}

func (pool *ServerPool) recycleFilterChain(chain *filterChain) {
	pool.filterChainPool.Put(chain)
}

func (pool *ServerPool) RecycleTo(name string, value interface{}) {
	pool.otherPools[name].Put(value)
}
