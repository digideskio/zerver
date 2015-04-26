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
	var err error
	if key != "" {
		ErrorPtrWrite(&err, w, JSONObjStart)
		ErrorPtrWrite(&err, w, Bytes(key))
		switch s := value.(type) {
		case string:
			ErrorPtrWrite(&err, w, JSONQuoteMid)
			ErrorPtrWrite(&err, w, Bytes(s))
			ErrorPtrWrite(&err, w, JSONQuoteEnd)
		case []byte:
			ErrorPtrWrite(&err, w, JSONQuoteMid)
			ErrorPtrWrite(&err, w, s)
			ErrorPtrWrite(&err, w, JSONQuoteEnd)
		default:
			ErrorPtrWrite(&err, w, JSONObjMid)
			if err == nil {
				err = json.NewEncoder(w).Encode(value)
			}
			ErrorPtrWrite(&err, w, JSONObjEnd)
		}
	} else {
		err = json.NewEncoder(w).Encode(value)
	}
	return err
}

func (JSONResource) Receive(r io.Reader, value interface{}) error {
	d := json.NewDecoder(r)
	d.UseNumber()
	return d.Decode(value)
}

var (
	JSONObjStart = []byte(`{"`)
	JSONObjMid   = []byte(`":`)
	JSONQuoteMid = []byte(`":"`)
	JSONObjEnd   = []byte("}")
	JSONQuoteEnd = []byte(`"}`)
)
