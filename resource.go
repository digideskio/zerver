package zerver

import (
	"encoding/json"
	"io"
	"io/ioutil"

	"github.com/cosiner/gohper/lib/types"
)

type (
	MarshalFunc func(interface{}) ([]byte, error)

	UnmarshalFunc func([]byte, interface{}) error

	Marshaler interface {
		Marshal(interface{}) ([]byte, error)
		Pool([]byte)
	}

	// if use your own
	ResourceMaster struct {
		WriteKV func(w ErrorWriter, key, value []byte) error
		Marshaler
		Unmarshal UnmarshalFunc
	}
)

func (m MarshalFunc) Marshal(v interface{}) ([]byte, error) { return m(v) }
func (MarshalFunc) Pool([]byte)                             {}

var JSONResource = &ResourceMaster{
	WriteKV:   WriteJSONKV,
	Marshaler: MarshalFunc(json.Marshal),
	Unmarshal: json.Unmarshal,
}

func (r *ResourceMaster) Init() *ResourceMaster {
	if r == nil {
		return JSONResource
	}
	if r.WriteKV == nil {
		r.WriteKV = JSONResource.WriteKV
	}
	if r.Marshaler == nil {
		r.Marshaler = JSONResource.Marshaler
	}
	if r.Unmarshal == nil {
		r.Unmarshal = JSONResource.Unmarshal
	}
	return r
}

var (
	_jsonHeadStart = []byte(`{"`)
	_jsonHeadEnd   = []byte(`":`)
	_jsonEnd       = []byte("}")
)

func WriteJSONKV(w ErrorWriter, key, value []byte) (err error) {
	w.Write(_jsonHeadStart)
	w.Write(key)
	w.Write(_jsonHeadEnd)
	w.Write(value)
	_, err = w.Write(_jsonEnd)
	return
}

// Send send resource to client, if key is empty, just send marshaled value,
// otherwise send as json object
func (r *ResourceMaster) Send(w ErrorWriter, key string, value interface{}) (err error) {
	var bs []byte
	if key != "" {
		bs, err = r.Marshal(value)
		if err == nil {
			err = r.WriteKV(w, types.UnsafeBytes(key), bs)
			r.Pool(bs)
		}
	} else if bs, err = r.Marshal(value); err == nil {
		r.Pool(bs)
		_, err = w.Write(bs)
	}
	return err
}

func (r *ResourceMaster) Recieve(rd io.Reader, v interface{}) error {
	bs, err := ioutil.ReadAll(rd)
	if err == nil {
		err = r.Unmarshal(bs, v)
	}
	return err
}
