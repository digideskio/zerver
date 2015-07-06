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
	switch err := err.(type) {
	case httperrs.Error:
		resp.ReportStatus(err.Code())
		if err.Code() < int(httperrs.Server) {
			OnErrLog(resp.Send(KeyError, err.Error()))
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

func OnErrLog(err error) {
	if err != nil {
		Logger.Errorln(err)
	}
}
