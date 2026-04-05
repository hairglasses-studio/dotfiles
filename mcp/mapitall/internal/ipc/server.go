package ipc

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"path/filepath"
)

// Request is a JSON-RPC 2.0 request.
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
	ID      any             `json:"id"`
}

// Response is a JSON-RPC 2.0 response.
type Response struct {
	JSONRPC string `json:"jsonrpc"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
	ID      any    `json:"id"`
}

// Error is a JSON-RPC 2.0 error object.
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Notification is a JSON-RPC 2.0 notification (no ID, no response expected).
type Notification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

// Handler processes a JSON-RPC method call.
type Handler func(params json.RawMessage) (any, error)

// StreamHandler processes a JSON-RPC method call and streams responses.
// It receives the request params, a JSON encoder for writing notifications,
// and a done channel that is closed when the connection drops.
// It should block until streaming is complete.
type StreamHandler func(params json.RawMessage, enc *json.Encoder, done <-chan struct{}) error

// Server listens on a Unix socket and dispatches JSON-RPC calls.
type Server struct {
	path           string
	handlers       map[string]Handler
	streamHandlers map[string]StreamHandler
	listener       net.Listener
}

// NewServer creates an IPC server at the given socket path.
func NewServer(path string) *Server {
	return &Server{
		path:           path,
		handlers:       make(map[string]Handler),
		streamHandlers: make(map[string]StreamHandler),
	}
}

// Handle registers a method handler.
func (s *Server) Handle(method string, h Handler) {
	s.handlers[method] = h
}

// HandleStream registers a streaming method handler.
// The handler writes JSON-RPC notifications to the connection until done.
func (s *Server) HandleStream(method string, h StreamHandler) {
	s.streamHandlers[method] = h
}

// Start begins listening. Call in a goroutine.
func (s *Server) Start(ctx context.Context) error {
	// Ensure parent directory exists.
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return fmt.Errorf("create socket dir: %w", err)
	}

	// Remove stale socket.
	os.Remove(s.path)

	ln, err := net.Listen("unix", s.path)
	if err != nil {
		return fmt.Errorf("listen %s: %w", s.path, err)
	}
	s.listener = ln

	go func() {
		<-ctx.Done()
		ln.Close()
	}()

	slog.Info("IPC server listening", "path", s.path)

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				slog.Warn("accept error", "error", err)
				continue
			}
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	defer conn.Close()

	dec := json.NewDecoder(conn)
	enc := json.NewEncoder(conn)

	var req Request
	if err := dec.Decode(&req); err != nil {
		return
	}

	// Check for streaming handlers first.
	if sh, ok := s.streamHandlers[req.Method]; ok {
		// Send an initial ack response so the client knows the stream started.
		enc.Encode(Response{
			JSONRPC: "2.0",
			Result:  map[string]any{"streaming": true},
			ID:      req.ID,
		})

		// Create a done channel that closes when the connection is lost.
		done := make(chan struct{})
		go func() {
			// Read from the connection until it's closed (client disconnect).
			buf := make([]byte, 1)
			for {
				if _, err := conn.Read(buf); err != nil {
					close(done)
					return
				}
			}
		}()

		if err := sh(req.Params, enc, done); err != nil {
			slog.Debug("stream handler ended", "method", req.Method, "error", err)
		}
		return
	}

	h, ok := s.handlers[req.Method]
	if !ok {
		enc.Encode(Response{
			JSONRPC: "2.0",
			Error:   &Error{Code: -32601, Message: "method not found: " + req.Method},
			ID:      req.ID,
		})
		return
	}

	result, err := h(req.Params)
	if err != nil {
		enc.Encode(Response{
			JSONRPC: "2.0",
			Error:   &Error{Code: -32000, Message: err.Error()},
			ID:      req.ID,
		})
		return
	}

	enc.Encode(Response{
		JSONRPC: "2.0",
		Result:  result,
		ID:      req.ID,
	})
}

// Close shuts down the server and removes the socket.
func (s *Server) Close() error {
	if s.listener != nil {
		s.listener.Close()
	}
	os.Remove(s.path)
	return nil
}
