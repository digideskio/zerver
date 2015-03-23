package xsrf

import (
	"sync"
	"time"

	"github.com/cosiner/zerver"

	"github.com/cosiner/gohper/lib/crypto"
)

const (
	HEADER_XSRFTOKEN  = "X-XSRFToken"
	HEADER_CSRFTOKEN  = "X-CSRFToken"
	XSRF_NAME         = "_xsrf"
	XSRF_ONERRORTOKEN = "dsajhdoqwARUH20174P2UAsdJDASDKJ"
	XSRF_DEFFLUSH     = 60
	XSRF_DEFLIFETIME  = 300
	XSRF_ERRORCODE    = 101
	XSRF_FORMHEAD     = `<input type="hidden" name="` + XSRF_NAME + `" value="`
	XSRF_FORMEND      = `"/>`
)

type (
	// XsrfErrorHandler is handler of xsrf token not matched
	XsrfErrorHandler interface {
		HandleXsrfError(req zerver.Request, resp zerver.Response)
	}

	// Xsrf is a xsrf processor, all work about xsrf token check and set
	// is done by it, it's not necessery do such work in other place
	Xsrf interface {
		// start process, every interval, change current xsrf token value
		Start(interval int) // interval is by seconds
		// Stop stop xsrf processor
		Stop()
		// Set setup xsrf token for later check and return it for user to
		// setup in post form
		// where to save xsrf token depend on implementations
		Set(zerver.Request, zerver.Response) string
		// IsValid check xsrf token from request
		IsValid(zerver.Request) bool
	}

	// emptyXsrf is a empty xsrf processor
	emptyXsrf struct{}

	// XsrfTokenGenerator generate xsrf token
	XsrfTokenGenerator func() string

	// xsrf implements interface Xsrf, it save xsrf token in client cookie
	// and use timing changed global xsrf token
	// default generate token through a random number generator,
	// which is a bit slow, if possible, change it when call NewXsrf
	xsrf struct {
		value         string
		valueGen      XsrfTokenGenerator
		valueLifetime int
		*sync.RWMutex
		running bool
		stop    chan bool
	}
)

// XsrfFormHTML return xsrf form html string with given token
func XsrfFormHTML(tok string) string {
	return XSRF_FORMHEAD + tok + XSRF_FORMEND
}

func (emptyXsrf) Set(zerver.Request, zerver.Response) string { return "" }
func (emptyXsrf) IsValid(zerver.Request) bool                { return true }
func (emptyXsrf) Start(int)                                  {}
func (emptyXsrf) Stop()                                      {}

// NewXsrf create a new xsrf processor
func NewXsrf(valueGen XsrfTokenGenerator, lifetime int) Xsrf {
	if valueGen == nil {
		valueGen = GenXsrfToken
	}
	return &xsrf{
		value:         valueGen(),
		valueLifetime: lifetime,
		RWMutex:       new(sync.RWMutex),
		stop:          make(chan bool, 1),
		running:       true,
	}
}

// stop stop xsrf processor
func (x *xsrf) Stop() {
	if x.running {
		x.running = false
		x.stop <- true
	}
}

// Start start xsrf processor
func (x *xsrf) Start(interval int) {
	go func() {
		c := time.NewTicker(time.Duration(interval) * time.Second)
		for {
			select {
			case <-c.C:
				x.Lock()
				x.value = x.valueGen()
				x.Unlock()
			case <-x.stop:
				return
			}
		}
	}()
}

// Set setup store token into response as cookie
func (x *xsrf) Set(_ zerver.Request, resp zerver.Response) (value string) {
	if x.running {
		x.RLock()
		value = x.value
		x.RUnlock()
		resp.SetCookie(XSRF_NAME, value, x.valueLifetime)
	}
	return
}

// IsValid check whether request xsrf token is valid
func (x *xsrf) IsValid(req zerver.Request) (value bool) {
	value = true
	if x.running {
		cookieValue := req.Cookie(XSRF_NAME)
		formValue := req.Param(XSRF_NAME)
		if formValue == "" {
			if formValue = req.Header(HEADER_XSRFTOKEN); formValue == "" {
				formValue = req.Header(HEADER_CSRFTOKEN)
			}
		}
		value = (cookieValue == formValue)
	}
	return
}

// GenXsrfToken generate xsrf token use random number generator
func GenXsrfToken() string {
	s, err := crypto.RandAlphanumeric(32)
	if err != nil {
		s = XSRF_ONERRORTOKEN
	}
	return s
}
