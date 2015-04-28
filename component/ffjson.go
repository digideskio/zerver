package component

import (
	"io"
	"io/ioutil"

	"github.com/cosiner/zerver"
	"github.com/pquerna/ffjson/ffjson"
)

// use COMP_RESOURCE

// Ffjson is a implementation of zerver.Resource
type Ffjson struct{}

func NewFfjsonResource() zerver.Resource {
	return Ffjson{}
}

func (Ffjson) Init(zerver.Enviroment) error { return nil }
func (Ffjson) Destroy()                     {}

func (Ffjson) Marshal(v interface{}) ([]byte, error) {
	return ffjson.Marshal(v)
}

func (Ffjson) Pool(data []byte) {
	ffjson.Pool(data)
}

func (Ffjson) Send(w io.Writer, key string, value interface{}) error {
	var (
		data []byte
		err  error
	)
	if key == "" { // send value only
		// alwayse use json marshal, simple string use Response.WriteString
		data, err = ffjson.Marshal(value)
		if err == nil {
			zerver.ErrorPtrWrite(&err, w, data)
			ffjson.Pool(data)
		}
		return err
	}

	zerver.ErrorPtrWrite(&err, w, zerver.JSONObjStart) // send key
	zerver.ErrorPtrWrite(&err, w, zerver.Bytes(key))
	if err == nil {
		if s, is := value.(string); is { // send string value
			zerver.ErrorPtrWrite(&err, w, zerver.JSONQuoteMid)
			zerver.ErrorPtrWrite(&err, w, zerver.Bytes(s))
			zerver.ErrorPtrWrite(&err, w, zerver.JSONQuoteEnd)
		} else { // send other value
			if data, err = ffjson.Marshal(value); err == nil {
				zerver.ErrorPtrWrite(&err, w, zerver.JSONObjMid)
				zerver.ErrorPtrWrite(&err, w, data)
				zerver.ErrorPtrWrite(&err, w, zerver.JSONObjEnd)
				ffjson.Pool(data)
			}
		}
	}
	return err
}

func (Ffjson) Unmarshal(data []byte, v interface{}) error {
	return ffjson.Unmarshal(data, v)
}

func (Ffjson) Receive(r io.Reader, v interface{}) error {
	data, err := ioutil.ReadAll(r)
	if err == nil {
		err = ffjson.Unmarshal(data, v)
	}
	return err
}
