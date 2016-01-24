package zerver

import (
	"sync"

	"github.com/cosiner/gohper/utils/attrs"
)

type requestEnv struct {
	req  request
	resp response
}

var reqEnvPool = &sync.Pool{
	New: func() interface{} {
		env := &requestEnv{}
		env.req.Attrs = attrs.New()
		return env
	},
}

func newRequestEnv() *requestEnv {
	return reqEnvPool.Get().(*requestEnv)
}

func recycleRequestEnv(req *requestEnv) {
	reqEnvPool.Put(req)
}
