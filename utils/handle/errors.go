package handle

import (
	"github.com/cosiner/gohper/errors"
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
	switch e := errors.Unwrap(err).(type) {
	case httperrs.Error:
		resp.ReportStatus(e.Code())
		if e.Code() < int(httperrs.Server) {
			Logger.Debugln(err)
			if err := resp.Send(KeyError, e.Error()); err != nil {
				Logger.Errorln(err.Error())
			}
			return
		}
	default:
		resp.ReportInternalServerError()
	}

	Logger.Errorln(err.Error())
}

func BadRequest(err error) error {
	if BadRequestError == nil {
		return err
	}

	Logger.Debugln(err.Error())
	return BadRequestError
}

func SendBadRequest(resp zerver.Response, err error) {
	SendErr(resp, BadRequest(err))
}

func OnErrLog(err error) {
	if err != nil {
		Logger.Errorln(err)
	}
}

func Send(resp zerver.Response, key string, value interface{}, err error) {
	if err != nil {
		SendErr(resp, err)
	} else {
		resp.Send(key, value)
	}
}

func ReportStatus(resp zerver.Response, status int, err error) {
	if err != nil {
		SendErr(resp, err)
	} else {
		resp.ReportStatus(status)
	}
}
