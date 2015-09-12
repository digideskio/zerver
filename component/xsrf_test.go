package component

import (
	"bytes"
	"testing"
	"time"

	"github.com/cosiner/gohper/testing2"
	"github.com/cosiner/zerver"
)

func TestXsrf(t *testing.T) {
	tt := testing2.Wrap(t)
	key := "123456789"
	data := []byte(key)
	xsrf := &Xsrf{
		Secret: key,
	}
	s := zerver.NewServer("")
	go s.Start(nil)
	time.Sleep(1 * time.Millisecond)

	xsrf.Init(s)
	signing := xsrf.sign(data)
	d := xsrf.verify(signing)
	tt.True(bytes.Equal(data, d))
}
