package zerver

import (
	"encoding/xml"
	"strings"
	"encoding/json"
	"io"
)

const (
	RES_JSON = "json"
	RES_XML = "xml"
	RES_HTML = "html"
	RES_PLAIN = "plain"
)

type (
	ResourceMaster struct {
		Resources map[string]Resource
		def string
		TypeOf func(Request) string
	}

	Resource interface {
		Component
		Marshal(interface{}) ([]byte, error)
		Pool([]byte)
		Unmarshal([]byte, interface{}) error
		Send(w io.Writer, key string, value interface{}) error
		Receive(r io.Reader, v interface{}) error
	}

	JSONResource struct{}
	XMLResource  struct{}
)

func ResourceType(req Request) string {
	typ := req.ContentType()
	if typ != "" {
		switch {
			case strings.Contains(typ, RES_JSON):return RES_JSON
			case strings.Contains(typ, RES_XML):return RES_XML
			case strings.Contains(typ, RES_HTML):return RES_HTML
			case strings.Contains(typ, RES_PLAIN):return RES_PLAIN
		}
	}
	return RES_JSON
}

func newResourceMaster() ResourceMaster {
	return ResourceMaster{
		Resources:make(map[string]Resource),
		def:RES_JSON,
		TypeOf: ResourceType,
	}
}

func (rm *ResourceMaster) Init(env Enviroment) error {
	for _, r := range rm.Resources {
		if err := r.Init(env); err != nil {
			return err
		}
	}
	return nil
}

func (rm *ResourceMaster) Destroy() {
	for _, r := range rm.Resources {
		r.Destroy()
	}
}

func (rm *ResourceMaster) Use(typ string, res Resource) {
	rm.Resources[typ] = res
}

func (rm *ResourceMaster) Default(typ string, res Resource) {
	if res != nil {
		rm.Use(typ, res)
	}
	rm.def = typ
}

func (rm *ResourceMaster) Resource(req Request) Resource {
	return rm.Resources[rm.TypeOf(req)]
}

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
		ErrorPtrWriteString(&err, w, key)
		switch s := value.(type) {
		case string:
			ErrorPtrWrite(&err, w, JSONQuoteMid)
			ErrorPtrWriteString(&err, w, s)
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

func (XMLResource) Init(Enviroment) error { return nil }
func (XMLResource) Destroy()              {}

func (XMLResource) Marshal(v interface{}) ([]byte, error) {
	return xml.Marshal(v)
}

func (XMLResource) Unmarshal(data []byte, v interface{}) error {
	return xml.Unmarshal(data, v)
}

func (XMLResource) Pool([]byte) {}

func (XMLResource) Send(w io.Writer, key string, value interface{}) error {
	var err error
	if key != "" {
		ErrorPtrWrite(&err, w, XMLTagStart)
		ErrorPtrWriteString(&err, w, key)
		ErrorPtrWrite(&err, w, XMLTagEnd)
		switch s := value.(type) {
		case string:
			ErrorPtrWriteString(&err, w, s)
		case []byte:
			ErrorPtrWrite(&err, w, s)
		default:
			if err == nil {
				err = xml.NewEncoder(w).Encode(value)
			}
		}
		ErrorPtrWrite(&err, w, XMLTagCloseStart)
		ErrorPtrWriteString(&err, w, key)
		ErrorPtrWrite(&err, w, XMLTagEnd)
	} else {
		err = xml.NewEncoder(w).Encode(value)
	}
	return err
}

func (XMLResource) Receive(r io.Reader, value interface{}) error {
	d := json.NewDecoder(r)
	d.UseNumber()
	return d.Decode(value)
}

var (
	XMLTagStart = []byte("<")
	XMLTagEnd = []byte(">")
	XMLTagCloseStart = []byte("</")
)
