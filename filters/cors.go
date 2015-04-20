package filters

import (
	"strconv"
	"strings"

	"github.com/cosiner/gohper/lib/types"
	"github.com/cosiner/zerver"
)

const (
	// request header
	_CORS_ORIGIN         = "Origin"
	_CORS_REQUESTMETHOD  = "Access-Control-Request-Method"
	_CORS_REQUESTHEADERS = "Access-Control-Request-Headers"

	// response header
	_CORS_ALLOWORIGIN      = "Access-Control-Allow-Origin"
	_CORS_ALLOWCREDENTIALS = "Access-Control-Allow-Credentials"
	_CORS_ALLOWHEADERS     = "Access-Control-Allow-Headers"
	_CORS_ALLOWMETHODS     = "Access-Control-Allow-Methods"
	_CORS_EXPOSEHEADERS    = "Access-Control-Expose-Headers"
	_CORS_MAXAGE           = "Access-Control-Max-Age"
)

type CORS struct {
	Origins          []string
	Methods          []string
	Headers          []string
	ExposeHeaders    []string `json:"expose"` // these headers can be accessed by javascript
	PreflightMaxage  int      `json:"maxage"` // max efficient seconds of browser preflight
	AllowCredentials bool     `json:"allow_cred"`

	allowAll         bool
	methods          string
	headers          string
	exposeHeaders    string
	preflightMaxage  string
	allowCredentials string
}

var (
	defAllowHeaders = []string{"Origin", "Accept", "Content-Type", "Authorization"}
	defAllowMethods = []string{"GET", "POST", "PATCH", "PUT", "DELETE"}
)

func (c *CORS) Init(zerver.Enviroment) error {
	if l := len(c.Origins); l == 0 || (l == 1 && c.Origins[0] == "*") {
		c.allowAll = true
		c.Origins = nil
	}
	if len(c.Methods) == 0 {
		c.Methods = defAllowMethods
	}
	c.methods = strings.Join(c.Methods, ",")
	if len(c.Headers) == 0 {
		c.Headers = defAllowHeaders
	}
	c.headers = strings.Join(c.Headers, ",")
	for i := range c.Headers {
		c.Headers[i] = strings.ToLower(c.Headers[i]) // chrome browser will use lower header
	}
	c.exposeHeaders = strings.Join(c.ExposeHeaders, ",")

	if c.AllowCredentials {
		c.allowCredentials = "true"
	} else {
		c.allowCredentials = "false"
	}
	if c.PreflightMaxage != 0 {
		c.preflightMaxage = strconv.Itoa(c.PreflightMaxage)
	}
	return nil
}

func (c *CORS) Destroy() {}

func (c *CORS) allow(origin string) bool {
	for i := range c.Origins {
		if c.Origins[i] == origin {
			return true
		}
	}
	return false
}

func (c *CORS) preflight(req zerver.Request, resp zerver.Response, method, headers string) {
	origin := "*"
	if !c.allowAll {
		origin = req.Header(_CORS_ORIGIN)
		if !c.allow(origin) {
			goto END
		}
	}
	resp.SetHeader(_CORS_ALLOWORIGIN, origin)
	method = strings.ToUpper(method)
	for _, m := range c.Methods {
		if m == method {
			resp.AddHeader(_CORS_ALLOWMETHODS, m)
			break
		}
	}
	for _, h := range types.TrimSplit(headers, ",") {
		for _, ch := range c.Headers {
			if strings.ToLower(h) == ch { // c.Headers already ToLowered when Init
				resp.AddHeader(_CORS_ALLOWHEADERS, ch)
				break
			}
		}
	}
	resp.SetHeader(_CORS_ALLOWCREDENTIALS, c.allowCredentials)
	if c.exposeHeaders != "" {
		resp.SetHeader(_CORS_EXPOSEHEADERS, c.exposeHeaders)
	}
	if c.preflightMaxage != "" {
		resp.SetHeader(_CORS_MAXAGE, c.preflightMaxage)
	}
END:
	resp.ReportOK()
}

func (c *CORS) filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	origin := "*"
	if !c.allowAll {
		origin = req.Header(_CORS_ORIGIN)
		if !c.allow(origin) {
			resp.ReportForbidden()
			return
		}
	}
	resp.SetHeader(_CORS_ALLOWORIGIN, origin)

	resp.SetHeader(_CORS_ALLOWMETHODS, c.methods)
	resp.SetHeader(_CORS_ALLOWHEADERS, c.headers)

	resp.SetHeader(_CORS_ALLOWCREDENTIALS, c.allowCredentials)
	if c.exposeHeaders != "" {
		resp.SetHeader(_CORS_EXPOSEHEADERS, c.exposeHeaders)
	}
	if c.preflightMaxage != "" {
		resp.SetHeader(_CORS_MAXAGE, c.preflightMaxage)
	}
	chain(req, resp)
}

func (c *CORS) Filter(req zerver.Request, resp zerver.Response, chain zerver.FilterChain) {
	reqMethod := req.Header(_CORS_REQUESTMETHOD)
	reqHeaders := req.Header(_CORS_REQUESTHEADERS)
	if req.Method() == "OPTIONS" && (reqMethod != "" || reqHeaders != "") {
		c.preflight(req, resp, reqMethod, reqHeaders)
	} else {
		c.filter(req, resp, chain)
	}
}
