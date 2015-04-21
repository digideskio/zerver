package components

import (
	"io"
	"io/ioutil"

	"github.com/cosiner/zerver"
	"github.com/pquerna/ffjson/ffjson"
)

// Ffjson is a implementation of zerver.ResourceMaster
type Ffjson struct{}

func NewFfjsonResource() zerver.ResourceMaster {
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
			_, err = w.Write(data)
			ffjson.Pool(data)
		}
		return err
	}

	_, err = zerver.ErrorWrite(err, w, zerver.JSONObjStart) // send key
	_, err = zerver.ErrorWrite(err, w, zerver.Bytes(key))
	if err != nil {
		return err
	}
	if s, is := value.(string); is { // send string value
		_, err = zerver.ErrorWrite(err, w, zerver.JSONQuoteMid)
		_, err = zerver.ErrorWrite(err, w, zerver.Bytes(s))
		_, err = zerver.ErrorWrite(err, w, zerver.JSONQuoteEnd)
	} else { // send other value
		if data, err = ffjson.Marshal(value); err == nil {
			_, err = zerver.ErrorWrite(err, w, zerver.JSONObjMid)
			_, err = zerver.ErrorWrite(err, w, data)
			_, err = w.Write(zerver.JSONObjEnd)
			ffjson.Pool(data)
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
