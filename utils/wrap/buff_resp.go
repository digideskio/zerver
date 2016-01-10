package wrap

import (
	"bytes"
	"net/http"
	"io"
)

type BuffRespWriter struct {
	http.ResponseWriter
	Buffer *bytes.Buffer
	ShouldClose bool
}

func (w *BuffRespWriter) Write(b []byte) (int, error) {
	if w.Buffer == nil {
		return w.ResponseWriter.Write(b)
	}
	return w.Buffer.Write(b)
}

func (w *BuffRespWriter) Close() error {
	if w.ShouldClose {
		return w.ResponseWriter.(io.Closer).Close()
	}
	return nil
}
