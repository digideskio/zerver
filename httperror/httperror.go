package httperror

import (
	"net/http"

	"github.com/cosiner/zerver"
)

const (
	SERVER  = iota // server error
	SERVICE        // temporary service error
	AUTH           // authorize error
	DATA           // data is not valid
	PARSE          // wrong data, unable to parse
)

type Error interface {
	Error() string
	Type() int
	Code() int
}

type httpError struct {
	err  error
	typ  int
	code int
}

func (e httpError) Error() string {
	return e.err.Error()
}

func (e httpError) Type() int {
	return e.typ
}

func (e httpError) Code() int {
	return e.code
}

func ServerError(e error) Error {
	if e == nil {
		return nil
	}

	return httpError{
		err:  e,
		typ:  SERVER,
		code: http.StatusInternalServerError,
	}
}

func ServiceError(e error) Error {
	if e == nil {
		return nil
	}

	return httpError{
		err:  e,
		typ:  SERVICE,
		code: http.StatusServiceUnavailable,
	}
}

func AuthError(e error) Error {
	if e == nil {
		return nil
	}

	return httpError{
		err:  e,
		typ:  AUTH,
		code: http.StatusUnauthorized,
	}
}

func DataError(e error) Error {
	if e == nil {
		return nil
	}

	return httpError{
		err:  e,
		typ:  DATA,
		code: http.StatusBadRequest,
	}
}

func ParseError(e error) Error {
	if e == nil {
		return nil
	}

	return httpError{
		err:  e,
		typ:  PARSE,
		code: http.StatusBadRequest,
	}
}

func Send(resp zerver.Response, err Error) error {
	resp.ReportStatus(err.Code())

	return resp.Send("error", err.Error())
}
