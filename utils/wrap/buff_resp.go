package wrap

import (
	"bytes"
	"net/http"
)

type BuffRespWriter struct {
	http.ResponseWriter
	Buffer *bytes.Buffer
}

func (w *BuffRespWriter) Write(b []byte) (int, error) {
	return w.Buffer.Write(b)
}
