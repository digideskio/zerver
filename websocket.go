package zerver

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	websocket "github.com/cosiner/zerver_websocket"
)

type (
	// WebSocketConn represent an websocket connection
	// WebSocket connection is not be managed in server,
	// it's handler's responsibility to close connection
	WebSocketConn interface {
		URLVarIndexer
		io.ReadWriteCloser
		SetDeadline(t time.Time) error
		SetReadDeadline(t time.Time) error
		SetWriteDeadline(t time.Time) error
		RemoteAddr() string
		RemoteIP() string
		UserAgent() string
		URL() *url.URL
		Enviroment
	}

	// webSocketConn is the actual websocket connection
	webSocketConn struct {
		Enviroment
		*websocket.Conn
		URLVarIndexer
		request *http.Request
	}

	// WebSocketHandlerFunc is the websocket connection handler
	WebSocketHandlerFunc func(WebSocketConn)

	// WebSocketHandler is the handler of websocket connection
	WebSocketHandler interface {
		ServerInitializer
		Destroy()
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

func (wsc *webSocketConn) URL() *url.URL {
	return wsc.request.URL
}

func (wsc *webSocketConn) RemoteAddr() string {
	return wsc.request.RemoteAddr
}

func (wsc *webSocketConn) RemoteIP() string {
	return strings.Split(wsc.RemoteAddr(), ":")[0]
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

// WebSocketHandlerFunc is a function WebSocketHandler
func (WebSocketHandlerFunc) Init(*Server) error           { return nil }
func (fn WebSocketHandlerFunc) Handle(conn WebSocketConn) { fn(conn) }
func (WebSocketHandlerFunc) Destroy()                     {}
