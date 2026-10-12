package mudserver

import (
	"errors"
	"time"

	"github.com/gorilla/websocket"
)

// Transport is how one player session sends and closes.
// WebSocket and SSH both implement it. Send must not be assumed to run on
// the game loop's goroutine; SSH implementations enqueue instead of writing.
type Transport interface {
	Send(v any) error
	Close(code int, reason string)
	RemoteIP() string
	Kind() string
}

type wsTransport struct {
	conn *websocket.Conn
	ip   string
}

func (w *wsTransport) Send(v any) error {
	if w == nil || w.conn == nil {
		return errors.New("websocket closed")
	}
	return w.conn.WriteJSON(v)
}

func (w *wsTransport) Close(code int, reason string) {
	if w == nil || w.conn == nil {
		return
	}
	if code != 0 {
		_ = w.conn.WriteControl(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(code, reason),
			time.Now().Add(time.Second),
		)
	}
	_ = w.conn.Close()
}

func (w *wsTransport) RemoteIP() string {
	if w == nil {
		return ""
	}
	return w.ip
}

func (w *wsTransport) Kind() string { return "ws" }
