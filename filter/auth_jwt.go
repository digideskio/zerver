package filter

import (
	"github.com/cosiner/gohper/errors"
	"github.com/cosiner/gohper/utils/defval"
	"github.com/cosiner/zerver"
	jwt "github.com/cosiner/zerver_jwt"
)

const (
	ErrNilJWT     = errors.Err("jwt token generator/validator can't be nil")
	ErrNilKeyFunc = errors.Err("jwt secret key getter can't be nil")
)

type JWTAuth struct {
	JWT               *jwt.JWT
	AuthTokenAttrName string
}

func (j *JWTAuth) Init(s *zerver.Server) error {
	if j.JWT == nil {
		return ErrNilJWT
	}
	if j.JWT.Keyfunc == nil {
		return ErrNilKeyFunc
	}
	defval.String(&j.AuthTokenAttrName, "AuthToken")
	defval.Nil(&j.JWT.SigningMethod, jwt.SigningMethodHS256)

	return nil
}

func (j *JWTAuth) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	if tokstr, basic := req.Authorization(); !basic && tokstr != "" {
		if tok, err := j.JWT.Parse(tokstr); err == nil {
			req.SetAttr(j.AuthTokenAttrName, tok)
			chain(req, resp)
			return
		}
	}

	resp.ReportUnauthorized()
}

func (j *JWTAuth) Destroy() {}
