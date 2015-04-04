package zerver

import (
	"fmt"

	"github.com/cosiner/gohper/lib/runtime"
)

// Tmp* provide a temporary data store, it should not be used after server start
var _tmp = make(map[string]interface{})

func TmpSet(key string, value interface{}) {
	_tmpCheck()
	_tmp[key] = value
}

func TmpHSet(key, key2 string, value interface{}) {
	_tmpCheck()
	if vs := _tmp[key]; vs == nil {
		vs := map[string]interface{}{
			key2: value,
		}
		_tmp[key] = vs
	} else if values, ok := vs.(map[string]interface{}); ok {
		values[key2] = value
	}
}

func TmpGet(key string) interface{} {
	_tmpCheck()
	return _tmp[key]
}

func TmpHGet(key, key2 string) interface{} {
	_tmpCheck()
	if values := _tmp[key]; values != nil {
		return values.(map[string]interface{})[key2]
	}
	return nil
}

func tmpDestroy() {
	_tmp = nil
}

func _tmpCheck() {
	if _tmp == nil {
		PanicServer(fmt.Sprintf("Temporary data store has been destroyed: %s", runtime.CallerPosition(2)))
	}
}
