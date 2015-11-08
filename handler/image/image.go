package image

import (
	"net/http"
	"path/filepath"

	"github.com/cosiner/gohper/utils/attrs"
	"github.com/cosiner/gohper/utils/bytesize"
	"github.com/cosiner/gohper/utils/defval"
	"github.com/cosiner/gohper/utils/httperrs"
	"github.com/cosiner/ygo/log"
	"github.com/cosiner/zerver"
	"github.com/cosiner/zerver/utils/handle"
)

type Path struct {
	Path string `json:"path"`
}

type Handler struct {
	MaxMemory int64 // Parse Form
	// handle.BadRequestError

	FileKey   string // Search file
	Params    []string
	ErrNoFile httperrs.Error

	Suffixes    map[string]struct{} // Check file suffix
	ErrNonImage httperrs.Error

	MaxSize     int64 // Check file size
	ErrTooLarge httperrs.Error

	PreChecker func(zerver.Request) error
	SaveImage  func(File, attrs.Attrs) (path string, err error) // Save image file
	PostDo     func(zerver.Request) error
}

// Init must be called
func (h *Handler) Init(env zerver.Environment) error {
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

	return nil
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
	if len(h.Suffixes) == 0 {
		return true
	}

	_, has := h.Suffixes[suffix]
	return has
}

func (h *Handler) Handle(req zerver.Request, resp zerver.Response) {
	req.Wrap(func(requ *http.Request, shouldClose bool) (r *http.Request, c bool) {
		r, c = requ, shouldClose

		err := requ.ParseMultipartForm(h.MaxMemory)
		if err != nil {
			handle.SendBadRequest(resp, err)
			return
		}

		if requ.MultipartForm == nil || requ.MultipartForm.File == nil {
			handle.SendErr(resp, h.ErrNoFile)
			return
		}

		files, has := requ.MultipartForm.File[h.FileKey]
		if !has || len(files) == 0 {
			handle.SendErr(resp, h.ErrNoFile)
			return
		}

		if len(h.Params) > 0 {
			for _, param := range h.Params {
				if vals := requ.MultipartForm.Value[param]; len(vals) != 0 {
					req.SetAttr(param, vals[0])
				}
			}
		}
		if h.PreChecker != nil {
			if err = h.PreChecker(req); err != nil {
				handle.SendErr(resp, err)
				return
			}
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
			handle.SendBadRequest(resp, err)
			return
		}
		defer fd.Close()

		log.Debugf("upload file: %s, size: %s\n", fd.Filename(), bytesize.ToHuman(uint64(fd.Size())))
		path, err := h.SaveImage(fd, req)
		if err != nil {
			handle.SendErr(resp, err)
			return
		}

		if h.PostDo != nil {
			err := h.PostDo(req)
			if err != nil {
				log.Warn("PostDo", err)
			}
		}

		resp.Send(Path{path})
		return
	})
}
