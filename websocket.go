package zerver

import (
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/cosiner/gohper/strings2"
	"github.com/cosiner/gohper/unsafe2"
	websocket "github.com/cosiner/zerver_websocket"
)

type (
	// WebSocketConn represent an websocket connection
	// WebSocket connection is not be managed in server,
	// it's handler's responsibility to close connection
	WebSocketConn interface {
		URLVarIndexer
		io.ReadWriteCloser
		Enviroment

		WriteString(string) (int, error)
		SetDeadline(t time.Time) error
		SetReadDeadline(t time.Time) error
		SetWriteDeadline(t time.Time) error
		RemoteAddr() string
		RemoteIP() string
		UserAgent() string
		URL() *url.URL
	}

	// webSocketConn is the actual websocket connection
	webSocketConn struct {
		Enviroment
		URLVarIndexer
		*websocket.Conn
		request *http.Request
	}

	// WebSocketHandlerFunc is the websocket connection handler
	WebSocketHandlerFunc func(WebSocketConn)

	// WebSocketHandler is the handler of websocket connection
	WebSocketHandler interface {
		Component
		Handle(WebSocketConn)
	}
)

// newWebSocketConn wrap a exist websocket connection and url variables to a
// new webSocketConn
func newWebSocketConn(e Enviroment, conn *websocket.Conn, varIndexer URLVarIndexer) *webSocketConn {
	return &webSocketConn{
		Enviroment:    e,
		Conn:          conn,
		URLVarIndexer: varIndexer,
		request:       conn.Request(),
	}
}

func (wsc *webSocketConn) WriteString(s string) (int, error) {
	return wsc.Write(unsafe2.Bytes(s))
}

func (wsc *webSocketConn) URL() *url.URL {
	return wsc.request.URL
}

func (wsc *webSocketConn) RemoteAddr() string {
	return wsc.request.RemoteAddr
}

func (wsc *webSocketConn) RemoteIP() string {
	if ip := wsc.request.Header.Get(HEADER_REALIP); ip != "" {
		return ip
	}
	addr := wsc.RemoteAddr()
	return addr[:strings2.LastIndexByte(addr, ':')]
}

// UserAgent return user's agent identify
func (wsc *webSocketConn) UserAgent() string {
	return wsc.request.Header.Get(HEADER_USERAGENT)
}

func convertWebSocketHandler(i interface{}) WebSocketHandler {
	switch w := i.(type) {
	case func(WebSocketConn):
		return WebSocketHandlerFunc(w)
	case WebSocketHandler:
		return w
	}
	return nil
}

func (WebSocketHandlerFunc) Init(Enviroment) error        { return nil }
func (fn WebSocketHandlerFunc) Handle(conn WebSocketConn) { fn(conn) }
func (WebSocketHandlerFunc) Destroy()                     {}
