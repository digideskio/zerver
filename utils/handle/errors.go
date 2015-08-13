package handle

import (
	"github.com/cosiner/gohper/utils/httperrs"
	"github.com/cosiner/ygo/log"
	"github.com/cosiner/zerver"
)

var (
	Logger          log.Logger
	KeyError        = "error"
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
	sendErrDepth(1, resp, err)
}

func sendErrDepth(depth int, resp zerver.Response, err error) {
	switch err := err.(type) {
	case httperrs.Error:
		resp.ReportStatus(err.Code())
		if err.Code() < int(httperrs.Server) {
			if err := resp.Send(KeyError, err.Error()); err != nil {
				Logger.ErrorDepth(depth+1, err)
			}
			OnErrLog(resp.Send(KeyError, err.Error()))
			return
		}
	default:
		resp.ReportInternalServerError()
	}

	Logger.ErrorDepth(depth+1, err.Error())
}

func BadRequest(err error) error {
	return badRequestDepth(1, err)
}

func badRequestDepth(depth int, err error) error {
	if BadRequestError == nil {
		return err
	}

	Logger.DebugDepth(depth+1, err.Error())
	return BadRequestError
}

func SendBadRequest(resp zerver.Response, err error) {
	sendErrDepth(1, resp, badRequestDepth(1, err))
}

func OnErrLog(err error) {
	if err != nil {
		Logger.ErrorDepth(1, err)
	}
}

func Send(resp zerver.Response, key string, value interface{}, err error) {
	if err != nil {
		sendErrDepth(1, resp, err)
	} else {
		resp.Send(key, value)
	}
}

func ReportStatus(resp zerver.Response, status int, err error) {
	if err != nil {
		sendErrDepth(1, resp, err)
	} else {
		resp.ReportStatus(status)
	}
}
