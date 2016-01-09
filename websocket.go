package zerver

import (
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/cosiner/gohper/unsafe2"
	websocket "github.com/cosiner/zerver_websocket"
)

type (
	WsConn interface {
		io.ReadWriteCloser
		Env

		Vars() *ReqVars
		WriteString(string) (int, error)
		SetDeadline(t time.Time) error
		SetReadDeadline(t time.Time) error
		SetWriteDeadline(t time.Time) error
		RemoteAddr() string
		URL() *url.URL
	}

	wsConn struct {
		Env
		vars *ReqVars
		*websocket.Conn
		request *http.Request
	}

	WsHandlerFunc func(WsConn)

	WsHandler interface {
		Component
		Handle(WsConn)
	}
)

func newWsConn(e Env, conn *websocket.Conn, vars *ReqVars) *wsConn {
	return &wsConn{
		Env:     e,
		Conn:    conn,
		vars:    vars,
		request: conn.Request(),
	}
}

func (c *wsConn) Vars() *ReqVars {
	return c.vars
}

func (c *wsConn) WriteString(s string) (int, error) {
	return c.Write(unsafe2.Bytes(s))
}

func (c *wsConn) URL() *url.URL {
	return c.request.URL
}

func (c *wsConn) RemoteAddr() string {
	return c.request.RemoteAddr
}

// UserAgent return user's agent identify
func (c *wsConn) UserAgent() string {
	return c.request.Header.Get(HEADER_USERAGENT)
}

func (WsHandlerFunc) Init(Env) error        { return nil }
func (fn WsHandlerFunc) Handle(conn WsConn) { fn(conn) }
func (WsHandlerFunc) Destroy()              {}
