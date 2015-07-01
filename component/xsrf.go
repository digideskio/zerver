package component

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"hash"
	"time"

	"github.com/cosiner/gohper/bytes2"
	"github.com/cosiner/gohper/errors"
	"github.com/cosiner/gohper/unsafe2"
	"github.com/cosiner/gohper/utils/defval"
	"github.com/cosiner/ygo/log"
	"github.com/cosiner/zerver"
)

const (
	_DEF_XSRF_TIMEOUT = 10 * 60 // 10 minutes
	_DEF_BUF_SIZE     = 256
	_HEADER_XSRFTOKEN = "X-XSRFToken"
	_HEADER_CSRFTOKEN = "X-CSRFToken"
	_XSRF_PARAM_NAME  = "_xsrf"
	_XSRF_FORMHEAD    = `<input type="hidden" name="` + _XSRF_PARAM_NAME + `" value="`
	_XSRF_FORMEND     = `"/>`

	XSRF = "Xsrf"
)

var _ENCODING = base64.URLEncoding

type (
	Xsrf struct {
		Timeout    int64            // seconds
		Secret     string           // secret key
		HashMethod func() hash.Hash // hash method for signing data
		Error      string           // error message for invalid token
		FilterGet  bool             // whether filter GET/HEAD/OPTIONS request

		UsePool bool // whether use sync.Pool for bytes allocation
		Pool    bytes2.Pool

		TokenInfo TokenInfo // marshal/unmarshal token info, default use jsonToken

		logger log.Logger
	}

	TokenInfo interface {
		Marshal(time int64, ip, agent string) ([]byte, error)
		Unmarshal([]byte) (time int64, ip, agent string)
	}

	jsonToken struct {
		Time  int64  `json:"a"`
		IP    string `json:"b"`
		Agent string `json:"c"`
	}
)

func (j jsonToken) Marshal(time int64, ip, agent string) ([]byte, error) {
	j.Time = time
	j.IP = ip
	j.Agent = agent

	return json.Marshal(&j)
}

func (j jsonToken) Unmarshal(bs []byte) (int64, string, string) {
	if err := json.Unmarshal(bs, &j); err != nil {
		return -1, "", ""
	}

	return j.Time, j.IP, j.Agent
}

func (x *Xsrf) Init(env zerver.Environment) error {
	if x.Secret == "" {
		return errors.Err("xsrf secret can't be empty")
	}

	defval.Int64(&x.Timeout, _DEF_XSRF_TIMEOUT)
	defval.Nil(&x.HashMethod, sha256.New)
	defval.String(&x.Error, "xsrf token is invalid or not found")

	if x.UsePool {
		if x.Pool == nil {
			x.Pool = bytes2.NewSyncPool(0, true)
		}
	} else {
		x.Pool = bytes2.FakePool{}
	}
	defval.Nil(&x.TokenInfo, jsonToken{})

	x.logger = env.Logger().Prefix("[XSRF]")
	return nil
}

func (x *Xsrf) Destroy() {}

// Create xsrf token, used as zerver.HandleFunc
func (x *Xsrf) Create(req zerver.Request, resp zerver.Response) {
	tokBytes, err := x.CreateFor(req)
	if err == nil {
		resp.ReportServiceUnavailable()
		return
	}

	if req.Method() == "POST" {
		resp.ReportCreated()
	}

	defer x.Pool.Put(tokBytes)
	x.logger.Warnln(resp.Send("tokBytes", tokBytes))
}

func (x *Xsrf) CreateFor(req zerver.Request) ([]byte, error) {
	bs, err := x.TokenInfo.Marshal(time.Now().Unix(), req.RemoteIP(), req.UserAgent())
	if err == nil {
		return x.sign(bs), nil
	}

	return nil, err
}

// Verify xsrf token, used as zerver.FilterFunc
//
// The reason not use "Filter" as function name is to prevent the Xsrf from used as both Component and Filter
func (x *Xsrf) Verify(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	if x.VerifyFor(req) {
		chain(req, resp)
	} else {
		resp.ReportBadRequest()
		resp.Send("error", x.Error)
	}
}

func (x *Xsrf) VerifyFor(req zerver.Request) bool {
	m := req.Method()
	if !x.FilterGet && (m == "GET" || m == "HEAD" || m == "OPTIONS") {
		return true
	}

	token := req.Header(_HEADER_XSRFTOKEN)
	if token == "" {
		token = req.Header(_HEADER_CSRFTOKEN)
		if token == "" {
			token = req.Param(_XSRF_PARAM_NAME)
			if token == "" {
				return false
			}
		}
	}

	data := x.verify(unsafe2.Bytes(token))
	if data != nil {
		x.Pool.Put(data)
		t, ip, agent := x.TokenInfo.Unmarshal(data)
		return t != -1 &&
			t+x.Timeout >= time.Now().Unix() &&
			ip == req.RemoteIP() &&
			agent == req.UserAgent()
	}

	return false
}

func (x *Xsrf) hash() hash.Hash {
	return hmac.New(x.HashMethod, unsafe2.Bytes(x.Secret))
}

func (x *Xsrf) sign(data []byte) []byte {
	hash := x.hash()
	hash.Write(data)
	signing := hash.Sum(nil)

	bs := x.Pool.Get(len(data)+hash.Size(), false) // data+signature
	buf := bytes.NewBuffer(bs)
	buf.Write(data)
	buf.Write(signing)

	dst := x.Pool.Get(_ENCODING.EncodedLen(buf.Len()), true)
	_ENCODING.Encode(dst, buf.Bytes())
	x.Pool.Put(bs)

	return dst
}

func (x *Xsrf) verify(signing []byte) []byte {
	dst := x.Pool.Get(_ENCODING.DecodedLen(len(signing)), true)
	n, err := _ENCODING.Decode(dst, signing)
	if err == nil {
		dst = dst[:n]
		hash := x.hash()

		sep := len(dst) - hash.Size()
		if sep > 0 {
			data := dst[:sep]
			hash.Write(data)

			if bytes.Equal(hash.Sum(nil), dst[sep:]) {
				return data
			}
		}
	}

	x.Pool.Put(dst)

	return nil
}

func XsrfHTML(token []byte) string {
	return _XSRF_FORMHEAD + string(token) + _XSRF_FORMEND
}
