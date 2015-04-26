package components

import (
	"bytes"
	"testing"

	"github.com/cosiner/gohper/lib/test"
)

func TestXsrf(t *testing.T) {
	tt := test.Wrap(t)
	data := []byte("123456789")
	xsrf := &Xsrf{
		Secret: data,
	}
	xsrf.Init(nil)
	signing := xsrf.sign(data)
	d := xsrf.verify(signing)
	tt.True(bytes.Equal(data, d))
}
