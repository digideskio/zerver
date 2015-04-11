package zerver

import (
	"encoding/json"
	"io"
)

type (
	ResourceMaster interface {
		Marshal(interface{}) ([]byte, error)
		Pool([]byte)
		Unmarshal([]byte, interface{}) error
		Send(w io.Writer, key string, value interface{}) error
		Recieve(r io.Reader, v interface{}) error
	}

	// if use your own
	JSONResource struct{}
)

func (JSONResource) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONResource) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (JSONResource) Pool([]byte) {}

func (JSONResource) Send(w io.Writer, key string, value interface{}) error {
	if key != "" {
		w.Write(_jsonHeadStart)
		w.Write(Bytes(key))
		w.Write(_jsonHeadEnd)
		json.NewEncoder(w).Encode(value)
		_, err := w.Write(_jsonEnd)
		return err
	} else {
		return json.NewEncoder(w).Encode(value)
	}
}

func (JSONResource) Recieve(r io.Reader, value interface{}) error {
	d := json.NewDecoder(r)
	d.UseNumber()
	return d.Decode(value)
}

var (
	_jsonHeadStart = []byte(`{"`)
	_jsonHeadEnd   = []byte(`":`)
	_jsonEnd       = []byte("}")
)
