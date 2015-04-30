package zerver

import (
	"log"

	"github.com/cosiner/gohper/lib/runtime"
)

// Tmp* provide a temporary data store, it should not be used after server start
// because of this, lazy-initialied component should not use these functions in
// in their Init method unless it was initialized by Handler/Filter...'s Init
var _tmp = make(map[interface{}]interface{})

func TmpSet(key, value interface{}) {
	_tmpCheck()
	_tmp[key] = value
}

func TmpHSet(key, key2, value interface{}) {
	_tmpCheck()
	if vs := _tmp[key]; vs == nil {
		vs := map[interface{}]interface{}{
			key2: value,
		}
		_tmp[key] = vs
	} else if values, ok := vs.(map[interface{}]interface{}); ok {
		values[key2] = value
	}
}

func TmpGet(key interface{}) interface{} {
	_tmpCheck()
	return _tmp[key]
}

func TmpHGet(key, key2 interface{}) interface{} {
	_tmpCheck()
	if values := _tmp[key]; values != nil {
		return values.(map[interface{}]interface{})[key2]
	}
	return nil
}

func tmpDestroy() {
	_tmp = nil
}

func _tmpCheck() {
	if _tmp == nil {
		log.Panicf("Temporary data store has been destroyed: %s\n", runtime.CallerPosition(2))
	}
}
