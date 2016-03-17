package handle

import (
	"net/http"

	"github.com/cosiner/gohper/errors"
	"github.com/cosiner/gohper/utils/httperrs"
	log "github.com/cosiner/ygo/jsonlog"
	"github.com/cosiner/zerver"
)

type Error struct {
	Error string `json:"error"`
}

var (
	BadRequestError httperrs.Error
)

func Wrap(handle func(zerver.Request, zerver.Response) error) zerver.HandleFunc {
	return func(req zerver.Request, resp zerver.Response) {
		if err := handle(req, resp); err != nil {
			SendErr(resp, err)
		}
	}
}

func SendErr(resp zerver.Response, err error) {
	if err == nil {
		panic("there is no error occurred")
	}
	switch e := errors.Unwrap(err).(type) {
	case httperrs.Error:
		resp.StatusCode(e.Code())
		if e.Code() < int(httperrs.Server) {
			resp.Send(Error{e.Error()})
			return
		}
	default:
		resp.Logger().Error(log.M{"msg": "internal server error", "error": err.Error()})
		resp.StatusCode(http.StatusInternalServerError)
	}
}

func BadRequest(err error) error {
	if BadRequestError == nil {
		return err
	}

	return BadRequestError
}

func SendBadRequest(resp zerver.Response, err error) {
	SendErr(resp, BadRequest(err))
}

func Send(resp zerver.Response, value interface{}, err error) {
	if err != nil {
		SendErr(resp, err)
	} else {
		resp.Send(value)
	}
}

func ReportStatus(resp zerver.Response, status int, err error) {
	if err != nil {
		SendErr(resp, err)
	} else {
		resp.StatusCode(status)
	}
}

func SendStatus(resp zerver.Response, data interface{}, status int, err error) {
	if err != nil {
		SendErr(resp, err)
	} else {
		resp.StatusCode(status)
		resp.Send(data)
	}
}
