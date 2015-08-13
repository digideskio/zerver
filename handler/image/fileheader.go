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

type File interface {
	multipart.File
	Filename() string
	Size() int64
}

type file struct {
	header *FileHeader
	multipart.File

	size int64
}

func (f *file) Filename() string {
	return f.header.Filename
}

func (f *file) Size() int64 {
	if f.size == 0 {
		f.size = f.header.Size()
	}
	return f.size
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
func (fh *FileHeader) Open() (File, error) {
	var (
		fd  multipart.File
		err error
	)

	if b := fh.content; b != nil {
		r := io.NewSectionReader(bytes.NewReader(b), 0, int64(len(b)))
		fd = sectionReadCloser{r}
	} else {
		fd, err = os.Open(fh.tmpfile)
		if err != nil {
			return nil, err
		}
	}

	return &file{header: fh, File: fd}, nil
}

type sectionReadCloser struct {
	*io.SectionReader
}

func (rc sectionReadCloser) Close() error {
	return nil
}
