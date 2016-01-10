package component

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"hash"
	"net/http"

	"github.com/cosiner/gohper/bytes2"
	"github.com/cosiner/gohper/errors"
	"github.com/cosiner/gohper/net2/http2"
	"github.com/cosiner/gohper/time2"
	"github.com/cosiner/gohper/unsafe2"
	"github.com/cosiner/gohper/utils/defval"
	log "github.com/cosiner/ygo/jsonlog"
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
	Token struct {
		Token string `json:"token"`
	}

	Xsrf struct {
		Timeout    int64            // seconds
		Secret     string           // secret key
		HashMethod func() hash.Hash // hash method for signing data
		Error      string           // error message for invalid token
		FilterGet  bool             // whether filter GET/HEAD/OPTIONS request

		UsePool bool // whether use sync.Pool for bytes allocation
		Pool    bytes2.Pool

		TokenInfo TokenInfo // marshal/unmarshal token info, default use jsonToken

		log *log.Logger
	}

	TokenInfo interface {
		Marshal(time int64, ip string) ([]byte, error)
		Unmarshal([]byte) (time int64, ip string)
	}

	jsonToken struct {
		Time int64  `json:"a"`
		IP   string `json:"b"`
	}
)

func (j jsonToken) Marshal(time int64, ip, agent string) ([]byte, error) {
	j.Time = time
	j.IP = ip

	return json.Marshal(&j)
}

func (j jsonToken) Unmarshal(bs []byte) (int64, string) {
	if err := json.Unmarshal(bs, &j); err != nil {
		return -1, ""
	}

	return j.Time, j.IP
}

func (x *Xsrf) Init(env zerver.Env) error {
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
	x.log = log.Derive("Component", "Xsrf")
	return nil
}

func (x *Xsrf) Destroy() {}

// Create xsrf token, used as zerver.HandleFunc
func (x *Xsrf) Create(req zerver.Request, resp zerver.Response) {
	tokBytes, err := x.CreateFor(req)
	if err == nil {
		resp.StatusCode(http.StatusServiceUnavailable)
		return
	}

	if req.ReqMethod() == "POST" {
		resp.StatusCode(http.StatusCreated)
	}

	defer x.Pool.Put(tokBytes)
	err = resp.Send(Token{string(tokBytes)})
	if err != nil {
		x.log.Error(log.M{"msg":"send xsrf token", "err":err.Error()})
	}
}

func (x *Xsrf) CreateFor(req zerver.Request) ([]byte, error) {
	bs, err := x.TokenInfo.Marshal(time2.Now().Unix(), http2.IpOfAddr(req.RemoteAddr()))
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
		resp.StatusCode(http.StatusBadRequest)
	}
}

func (x *Xsrf) VerifyFor(req zerver.Request) bool {
	m := req.ReqMethod()
	if !x.FilterGet && (m == zerver.METHOD_GET || m == zerver.METHOD_HEAD || m == zerver.METHOD_OPTIONS) {
		return true
	}

	token := req.GetHeader(_HEADER_XSRFTOKEN)
	if token == "" {
		token = req.GetHeader(_HEADER_CSRFTOKEN)
		if token == "" {
			token = req.Vars().QueryVar(_XSRF_PARAM_NAME)
			if token == "" {
				return false
			}
		}
	}

	data := x.verify(unsafe2.Bytes(token))
	if data != nil {
		x.Pool.Put(data)
		t, ip := x.TokenInfo.Unmarshal(data)
		return t != -1 &&
			t+x.Timeout >= time2.Now().Unix() &&
			ip == http2.IpOfAddr(req.RemoteAddr())
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
