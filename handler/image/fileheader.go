package image

import (
	"bytes"
	"io"
	"math"
	"mime/multipart"
	"net/textproto"
	"os"
	"unsafe"
)

type FileHeader struct {
	Filename string
	Header   textproto.MIMEHeader

	content []byte
	tmpfile string
}

func convertFileHandler(fh *multipart.FileHeader) *FileHeader {
	return (*FileHeader)(unsafe.Pointer(fh))
}

func (fh *FileHeader) Size() int64 {
	if fh.content != nil {
		return int64(len(fh.content))
	}

	stat, err := os.Stat(fh.tmpfile)
	if err != nil {
		return math.MaxInt64
	}

	return stat.Size()
}

// Open opens and returns the FileHeader's associated File.
func (fh *FileHeader) Open() (multipart.File, error) {
	if b := fh.content; b != nil {
		r := io.NewSectionReader(bytes.NewReader(b), 0, int64(len(b)))
		return sectionReadCloser{r}, nil
	}
	return os.Open(fh.tmpfile)
}

type sectionReadCloser struct {
	*io.SectionReader
}

func (rc sectionReadCloser) Close() error {
	return nil
}
