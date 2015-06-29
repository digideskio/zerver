package image

import (
	"mime/multipart"
	"net/http"
	"path/filepath"

	"github.com/cosiner/gohper/defval"
	"github.com/cosiner/gohper/errors/httperrs"
	"github.com/cosiner/zerver"
	"github.com/cosiner/zerver/util/handle"
)

type Handler struct {
	MaxMemory int64 // Parse Form
	// handle.BadRequestError

	FileKey   string // Search file
	ErrNoFile httperrs.Error

	Suffixes    map[string]struct{} // Check file suffix
	ErrNonImage httperrs.Error

	MaxSize     int64 // Check file size
	ErrTooLarge httperrs.Error

	SaveImage func(multipart.File) (path string, err error) // Save image file

	PathKey string // Response.Send(PathKey, path)
}

// Init should be called anyway
func (h *Handler) Init() *Handler {
	if h.SaveImage == nil {
		panic("the function to save image should not be nil")
	}

	defval.Int64(&h.MaxMemory, 1<<19) // 512K

	defval.String(&h.FileKey, "image")
	if h.ErrNoFile == nil {
		h.ErrNoFile = httperrs.BadRequest.NewS("upload file not exists")
	}

	if h.Suffixes == nil {
		h.AddSuffixes("png", "jpg")
	}

	if h.ErrNonImage == nil {
		h.ErrNonImage = httperrs.BadRequest.NewS("the upload file is not an image")
	}

	defval.Int64(&h.MaxSize, 1<<18) // 256K
	if h.ErrTooLarge == nil {
		h.ErrTooLarge = httperrs.BadRequest.NewS("the upload file size is too large")
	}

	defval.String(&h.PathKey, "path")

	return h
}

func (h *Handler) AddSuffixes(suffixes ...string) {
	exist := struct{}{}

	if h.Suffixes == nil {
		h.Suffixes = make(map[string]struct{})
	}

	for _, s := range suffixes {
		if s[0] != '.' {
			s = "." + s
		}

		h.Suffixes[s] = exist
	}
}

func (h *Handler) isSuffixSupported(suffix string) bool {
	_, has := h.Suffixes[suffix]
	return has
}

func (h *Handler) Handle(req zerver.Request, resp zerver.Response) {
	req.Wrap(func(req *http.Request, shouldClose bool) (r *http.Request, c bool) {
		r, c = req, shouldClose

		err := req.ParseMultipartForm(h.MaxMemory)
		if err != nil {
			handle.SendErr(resp, handle.BadRequest(err))
			return
		}

		if req.MultipartForm == nil || req.MultipartForm.File == nil {
			handle.SendErr(resp, h.ErrNoFile)
			return
		}

		files, has := req.MultipartForm.File[h.FileKey]
		if !has || len(files) == 0 {
			handle.SendErr(resp, h.ErrNoFile)
			return
		}

		file := convertFileHandler(files[0])
		if !h.isSuffixSupported(filepath.Ext(file.Filename)) {
			handle.SendErr(resp, h.ErrNonImage)
			return
		}

		if file.Size() > h.MaxSize {
			handle.SendErr(resp, h.ErrTooLarge)
			return
		}

		fd, err := file.Open()
		if err != nil {
			handle.SendErr(resp, handle.BadRequest(err))
			return
		}

		defer fd.Close()
		path, err := h.SaveImage(fd)
		if err != nil {
			handle.SendErr(resp, err)
			return
		}

		resp.Send(h.PathKey, path)
		return
	})
}
