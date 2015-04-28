package zerver

import (
	"strings"
	"testing"

	"github.com/cosiner/gohper/lib/test"
)

func resourceType(typ string) string {
	if typ != "" {
		switch {
		case strings.Contains(typ, RES_JSON):
			return RES_JSON
		case strings.Contains(typ, RES_XML):
			return RES_XML
		case strings.Contains(typ, RES_HTML):
			return RES_HTML
		case strings.Contains(typ, RES_PLAIN):
			return RES_PLAIN
		}
	}
	return RES_JSON
}

func TestResourceMaster(t *testing.T) {
	tt := test.Wrap(t)
	rm := newResourceMaster()
	rm.Default(RES_JSON, JSONResource{})
	rm.Use(RES_XML, XMLResource{})

	res := rm.Resources[resourceType(CONTENTTYPE_JSON)]
	_, is := res.(JSONResource)
	tt.True(is)

	res = rm.Resources[resourceType(CONTENTTYPE_XML)]
	_, is = res.(XMLResource)
	tt.True(is)

	res = rm.Resources[resourceType("abcdefghijklmn")]
	_, is = res.(JSONResource)
	tt.True(is)
}
