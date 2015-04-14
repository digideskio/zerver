package zerver

import (
	"encoding/json"
	"io"
)

const (
	COMP_RESOURCE = "ResourceComponent"
)

type (
	ResourceMaster interface {
		Component
		Marshal(interface{}) ([]byte, error)
		Pool([]byte)
		Unmarshal([]byte, interface{}) error
		Send(w io.Writer, key string, value interface{}) error
		Receive(r io.Reader, v interface{}) error
	}

	// if use your own
	JSONResource struct{}
)

func (JSONResource) Init(Enviroment) error { return nil }
func (JSONResource) Destroy()              {}

func (JSONResource) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func (JSONResource) Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func (JSONResource) Pool([]byte) {}

func (JSONResource) Send(w io.Writer, key string, value interface{}) error {
	if key != "" {
		w.Write(JSONObjStart)
		w.Write(Bytes(key))
		w.Write(JSONObjMid)
		json.NewEncoder(w).Encode(value)
		_, err := w.Write(JSONObjEnd)
		return err
	} else {
		return json.NewEncoder(w).Encode(value)
	}
}

func (JSONResource) Receive(r io.Reader, value interface{}) error {
	d := json.NewDecoder(r)
	d.UseNumber()
	return d.Decode(value)
}

var (
	JSONObjStart = []byte(`{"`)
	JSONObjMid   = []byte(`":`)
	JSONObjEnd   = []byte("}")
)
