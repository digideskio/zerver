package component

import (
	"bytes"
	"testing"

	"github.com/cosiner/gohper/testing2"
)

func TestXsrf(t *testing.T) {
	tt := testing2.Wrap(t)
	data := []byte("123456789")
	xsrf := &Xsrf{
		Secret: data,
	}
	xsrf.Init(nil)
	signing := xsrf.sign(data)
	d := xsrf.verify(signing)
	tt.True(bytes.Equal(data, d))
}
