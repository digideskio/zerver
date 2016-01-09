package handler

import (
	"net/http"

	"github.com/cosiner/zerver"
)

type MethodHandler interface {
	Get(zerver.Request, zerver.Response)
	Post(zerver.Request, zerver.Response)
	Delete(zerver.Request, zerver.Response)
	Put(zerver.Request, zerver.Response)
	Patch(zerver.Request, zerver.Response)
}

type methodHandler struct {
	zerver.Component
	MethodHandler
}

func WrapMethodHandler(m MethodHandler) zerver.Handler {
	return &methodHandler{
		Component:     zerver.NopComponent{},
		MethodHandler: m,
	}
}

func (s methodHandler) Handler(method string) zerver.HandleFunc {
	switch method {
	case zerver.METHOD_GET:
		return s.Get
	case zerver.METHOD_POST:
		return s.Post
	case zerver.METHOD_DELETE:
		return s.Delete
	case zerver.METHOD_PUT:
		return s.Put
	case zerver.METHOD_PATCH:
		return s.Patch
	}

	return nil
}

type NopMethodHandler struct{}

func (NopMethodHandler) Get(_ zerver.Request, resp zerver.Response) {
	resp.StatusCode(http.StatusMethodNotAllowed)
}

func (NopMethodHandler) Post(_ zerver.Request, resp zerver.Response) {
	resp.StatusCode(http.StatusMethodNotAllowed)
}

func (NopMethodHandler) Delete(_ zerver.Request, resp zerver.Response) {
	resp.StatusCode(http.StatusMethodNotAllowed)
}

func (NopMethodHandler) Put(_ zerver.Request, resp zerver.Response) {
	resp.StatusCode(http.StatusMethodNotAllowed)
}

func (NopMethodHandler) Patch(_ zerver.Request, resp zerver.Response) {
	resp.StatusCode(http.StatusMethodNotAllowed)
}
