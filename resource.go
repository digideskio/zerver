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

	ResourceMaster struct {
		Marshaler
		Unmarshal UnmarshalFunc
	}
)

var DefaultResourceMaster = ResourceMaster{
	Marshaler: MarshalFunc(json.Marshal),
	Unmarshal: json.Unmarshal,
}

func (m MarshalFunc) Marshal(v interface{}) ([]byte, error) { return m(v) }
func (MarshalFunc) Pool([]byte)                             {}

var (
	_jsonHeadStart = []byte(`{"`)
	_jsonHeadEnd   = []byte(`":`)
	_jsonEnd       = []byte("}")
)

// Send send resource to client, if key is empty, just send marshaled value,
// otherwise send as json object
func (r ResourceMaster) Send(resp io.Writer, key string, value interface{}) (err error) {
	var bs []byte
	if key != "" {
		resp.Write(_jsonHeadStart)
		resp.Write(types.UnsafeBytes(key))
		resp.Write(_jsonHeadEnd)
		if bs, err = r.Marshal(value); err == nil {
			r.Pool(bs)
			resp.Write(bs)
			_, err = resp.Write(_jsonEnd)
		}
	} else if bs, err = r.Marshal(value); err == nil {
		r.Pool(bs)
		_, err = resp.Write(bs)
	}
	return err
}

func (r ResourceMaster) Recieve(req io.Reader, v interface{}) error {
	bs, err := ioutil.ReadAll(req)
	if err == nil {
		err = r.Unmarshal(bs, v)
	}
	return err
}
