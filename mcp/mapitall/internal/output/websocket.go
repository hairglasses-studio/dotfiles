package output

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/hairglasses-studio/mapitall/internal/mapping"
	"nhooyr.io/websocket"
)

// WebSocketTarget sends messages over WebSocket connections.
type WebSocketTarget struct {
	mu    sync.Mutex
	conns map[string]*websocket.Conn // url -> conn
}

// NewWebSocketTarget creates a WebSocket output target.
func NewWebSocketTarget() *WebSocketTarget {
	return &WebSocketTarget{conns: make(map[string]*websocket.Conn)}
}

func (t *WebSocketTarget) Type() mapping.OutputType { return mapping.OutputWebSocket }

func (t *WebSocketTarget) Execute(action mapping.OutputAction, value float64) error {
	url := action.URL
	if url == "" {
		return fmt.Errorf("websocket: no URL specified")
	}

	conn, err := t.getConn(url)
	if err != nil {
		return err
	}

	msg := action.Message
	msg = strings.ReplaceAll(msg, "{value}", fmt.Sprintf("%.4f", value))
	msg = strings.ReplaceAll(msg, "{scaled}", fmt.Sprintf("%.0f", value))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return conn.Write(ctx, websocket.MessageText, []byte(msg))
}

func (t *WebSocketTarget) getConn(url string) (*websocket.Conn, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if c, ok := t.conns[url]; ok {
		return c, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		return nil, fmt.Errorf("websocket dial %s: %w", url, err)
	}
	t.conns[url] = c
	slog.Info("websocket connected", "url", url)
	return c, nil
}

func (t *WebSocketTarget) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	for url, c := range t.conns {
		c.Close(websocket.StatusNormalClosure, "mapitall shutting down")
		delete(t.conns, url)
	}
	return nil
}
