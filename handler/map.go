package handler

import (
	"fmt"

	"github.com/cosiner/zerver"
)

type MapHandler map[string]zerver.HandleFunc

func (mh MapHandler) Init(zerver.Env) error {
	for m, h := range mh {
		delete(mh, m)
		mh[zerver.MethodName(m)] = h
	}

	return nil
}

func (mh MapHandler) Handler(method string) zerver.HandleFunc {
	return mh[method]
}

func (mh MapHandler) Destroy() {
	for m := range mh {
		delete(mh, m)
	}
}

func (mh MapHandler) Use(method string, handleFunc zerver.HandleFunc) {
	if _, has := mh[method]; has {
		panic(fmt.Errorf("handleFunc for method %s already exist.", method))
	}

	mh[method] = handleFunc
}
