package zerver

import "sync"

type requestEnv struct {
	vars ReqVars
	req  request
	resp response
}

var reqEnvPool = &sync.Pool{
	New: func() interface{} {
		return &requestEnv{}
	},
}

func newRequestEnv() *requestEnv {
	return reqEnvPool.Get().(*requestEnv)
}

func recycleRequestEnv(req *requestEnv) {
	reqEnvPool.Put(req)
}
