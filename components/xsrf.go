package components

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"hash"
	"sync"
	"time"

	"github.com/cosiner/gohper/lib/errors"
	"github.com/cosiner/zerver"
)

const (
	DEF_XSRF_TIMEOUT  = 10 * 60 // 10 minutes
	DEF_BUF_SIZE      = 256
	_HEADER_XSRFTOKEN = "X-XSRFToken"
	_HEADER_CSRFTOKEN = "X-CSRFToken"
	_XSRF_PARAM_NAME  = "_xsrf"
	_XSRF_FORMHEAD    = `<input type="hidden" name="` + _XSRF_PARAM_NAME + `" value="`
	_XSRF_FORMEND     = `"/>`

	COMP_XSRF = "XsrfComponent"
)

var _ENCODING = base64.URLEncoding

type (
	Xsrf struct {
		Timeout    int64            // seconds
		Secret     []byte           // secret key
		HashMethod func() hash.Hash // hash method for signing data
		Error      string           // error message for invalid token
		FilterGet  bool             // whether filter GET/HEAD/OPTIONS request
		UsePool    bool             // whether use sync.Pool for bytes allocation
		BufSize    int              // buffer size for pool
		TokenInfo  TokenInfo        // marshal/unmarshal token info, default use jsonToken

		pool sync.Pool
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
	if err := json.Unmarshal(bs, &j); err == nil {
		return j.Time, j.IP, j.Agent
	}
	return -1, "", ""

}

func (x *Xsrf) Init(zerver.Enviroment) error {
	if x.Secret == nil {
		return errors.Err("xsrf secret can't be empty")
	}
	if x.Timeout == 0 {
		x.Timeout = DEF_XSRF_TIMEOUT
	}
	if x.HashMethod == nil {
		x.HashMethod = sha256.New
	}
	if x.Error == "" {
		x.Error = "xsrf token is invalid or not found"
	}
	if x.UsePool && x.pool.New == nil {
		if x.BufSize == 0 {
			x.BufSize = DEF_BUF_SIZE
		}
		x.pool.New = func() interface{} {
			return make([]byte, x.BufSize)
		}
	}
	if x.TokenInfo == nil {
		x.TokenInfo = jsonToken{}
	}
	return nil
}

func (x *Xsrf) PoolBytes(bs []byte) {
	if x.UsePool {
		x.pool.Put(bs)
	}
}

func (x *Xsrf) GetBytes(size int, asLen bool) []byte {
	var bs []byte
	if x.UsePool {
		bs = x.pool.Get().([]byte)
		if cap(bs) < size {
			bs = make([]byte, size)
		}
	} else {
		bs = make([]byte, size)
	}
	if asLen {
		bs = bs[:size]
	} else {
		bs = bs[:0]
	}
	return bs
}

func (x *Xsrf) Destroy() {}

func (x *Xsrf) Create(req zerver.Request, resp zerver.Response) {
	tokBytes, err := x.CreateFor(req)
	if err == nil {
		if req.Method() == "POST" {
			resp.ReportCreated()
		}
		req.Server().OnErrorLog(resp.Send("tokBytes", tokBytes))
		x.PoolBytes(tokBytes)
	} else {
		resp.ReportServiceUnavailable()
	}
}

func (x *Xsrf) CreateFor(req zerver.Request) ([]byte, error) {
	bs, err := x.TokenInfo.Marshal(time.Now().Unix(),
		req.RemoteIP(),
		req.UserAgent())
	if err == nil {
		return x.sign(bs), nil
	}
	return nil, err
}

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

	data := x.verify(zerver.Bytes(token))
	if data != nil {
		x.PoolBytes(data)
		t, ip, agent := x.TokenInfo.Unmarshal(data)
		return t != -1 &&
			t+x.Timeout >= time.Now().Unix() &&
			ip == req.RemoteIP() &&
			agent == req.UserAgent()
	}
	return false
}

func (x *Xsrf) sign(data []byte) []byte {
	hash := hmac.New(x.HashMethod, x.Secret)
	hash.Write(data)
	signing := hash.Sum(nil)

	bs := x.GetBytes(len(data)+hash.Size(), false) // data+signature
	buf := bytes.NewBuffer(bs)
	buf.Write(data)
	buf.Write(signing)

	dst := x.GetBytes(_ENCODING.EncodedLen(buf.Len()), true)
	_ENCODING.Encode(dst, buf.Bytes())
	x.PoolBytes(bs)

	return dst
}

func (x *Xsrf) verify(signing []byte) []byte {
	dst := x.GetBytes(_ENCODING.DecodedLen(len(signing)), true)
	n, err := _ENCODING.Decode(dst, signing)
	if err == nil {
		dst = dst[:n]
		hash := hmac.New(x.HashMethod, x.Secret)
		sep := len(dst) - hash.Size()
		if sep > 0 {
			data := dst[:sep]
			hash.Write(data)
			if bytes.Equal(hash.Sum(nil), dst[sep:]) {
				return data
			}
		}
	}
	x.PoolBytes(dst)
	return nil
}

func XsrfHTML(token []byte) string {
	return _XSRF_FORMHEAD + string(token) + _XSRF_FORMEND
}
