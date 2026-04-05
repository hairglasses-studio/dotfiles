// Package testutil provides test utilities for hg-mcp tool module testing.
// It includes mock servers for HTTP, OSC (UDP), and TCP protocols
// that return canned responses for integration testing without real services.
package testutil

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
)

// MockHTTPServer wraps httptest.Server with canned response routing.
// Routes are matched by method+path and return pre-configured responses.
type MockHTTPServer struct {
	Server *httptest.Server
	mu     sync.RWMutex
	routes map[string]MockHTTPResponse
}

// MockHTTPResponse defines a canned HTTP response.
type MockHTTPResponse struct {
	StatusCode int
	Body       string
	Headers    map[string]string
}

// NewMockHTTPServer creates a mock HTTP server with pre-configured routes.
// Routes are keyed by "METHOD /path" (e.g., "GET /api/status").
func NewMockHTTPServer(routes map[string]MockHTTPResponse) *MockHTTPServer {
	if routes == nil {
		routes = make(map[string]MockHTTPResponse)
	}
	m := &MockHTTPServer{
		routes: routes,
	}

	m.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Method + " " + r.URL.Path

		m.mu.RLock()
		resp, ok := m.routes[key]
		m.mu.RUnlock()

		if !ok {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintf(w, `{"error":"no mock for %s"}`, key)
			return
		}

		for k, v := range resp.Headers {
			w.Header().Set(k, v)
		}
		if resp.StatusCode > 0 {
			w.WriteHeader(resp.StatusCode)
		}
		fmt.Fprint(w, resp.Body)
	}))

	return m
}

// URL returns the mock server's base URL.
func (m *MockHTTPServer) URL() string {
	return m.Server.URL
}

// Close shuts down the mock server.
func (m *MockHTTPServer) Close() {
	m.Server.Close()
}

// SetRoute adds or updates a route at runtime.
func (m *MockHTTPServer) SetRoute(method, path string, resp MockHTTPResponse) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.routes[method+" "+path] = resp
}

// JSONRoute is a convenience for adding a route that returns JSON.
func (m *MockHTTPServer) JSONRoute(method, path string, statusCode int, body interface{}) {
	data, _ := json.Marshal(body)
	m.SetRoute(method, path, MockHTTPResponse{
		StatusCode: statusCode,
		Body:       string(data),
		Headers:    map[string]string{"Content-Type": "application/json"},
	})
}

// MockTCPServer listens on a random TCP port and returns canned responses.
// It accepts connections and echoes a fixed response for each connection.
type MockTCPServer struct {
	Listener net.Listener
	Addr     string
	Port     int
	mu       sync.RWMutex
	response []byte
	done     chan struct{}
}

// NewMockTCPServer creates a mock TCP server that returns a fixed response.
func NewMockTCPServer(response []byte) (*MockTCPServer, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	addr := listener.Addr().(*net.TCPAddr)

	m := &MockTCPServer{
		Listener: listener,
		Addr:     addr.String(),
		Port:     addr.Port,
		response: response,
		done:     make(chan struct{}),
	}

	go m.acceptLoop()
	return m, nil
}

func (m *MockTCPServer) acceptLoop() {
	for {
		conn, err := m.Listener.Accept()
		if err != nil {
			select {
			case <-m.done:
				return
			default:
				return
			}
		}
		go m.handleConn(conn)
	}
}

func (m *MockTCPServer) handleConn(conn net.Conn) {
	defer conn.Close()

	// Read whatever the client sends (up to 4KB)
	buf := make([]byte, 4096)
	conn.Read(buf) //nolint:errcheck

	// Send canned response
	m.mu.RLock()
	resp := m.response
	m.mu.RUnlock()

	if len(resp) > 0 {
		conn.Write(resp) //nolint:errcheck
	}
}

// SetResponse changes the canned response at runtime.
func (m *MockTCPServer) SetResponse(response []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.response = response
}

// Close shuts down the mock server.
func (m *MockTCPServer) Close() {
	close(m.done)
	m.Listener.Close()
}

// MockUDPServer listens on a random UDP port and returns canned responses.
// Useful for testing OSC and other UDP-based protocols.
type MockUDPServer struct {
	Conn *net.UDPConn
	Addr string
	Port int
	mu   sync.RWMutex
	// ResponseFunc is called for each received packet to generate a response.
	// If nil, no response is sent.
	ResponseFunc func(data []byte) []byte
	done         chan struct{}
	received     [][]byte
	receivedMu   sync.Mutex
}

// NewMockUDPServer creates a mock UDP server.
// If responseFunc is nil, the server only records received packets.
func NewMockUDPServer(responseFunc func([]byte) []byte) (*MockUDPServer, error) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to resolve addr: %w", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("failed to listen: %w", err)
	}

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	m := &MockUDPServer{
		Conn:         conn,
		Addr:         localAddr.String(),
		Port:         localAddr.Port,
		ResponseFunc: responseFunc,
		done:         make(chan struct{}),
	}

	go m.readLoop()
	return m, nil
}

func (m *MockUDPServer) readLoop() {
	buf := make([]byte, 65535)
	for {
		n, remoteAddr, err := m.Conn.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-m.done:
				return
			default:
				return
			}
		}

		data := make([]byte, n)
		copy(data, buf[:n])

		m.receivedMu.Lock()
		m.received = append(m.received, data)
		m.receivedMu.Unlock()

		m.mu.RLock()
		fn := m.ResponseFunc
		m.mu.RUnlock()

		if fn != nil {
			resp := fn(data)
			if len(resp) > 0 {
				m.Conn.WriteToUDP(resp, remoteAddr) //nolint:errcheck
			}
		}
	}
}

// Received returns all packets received so far.
func (m *MockUDPServer) Received() [][]byte {
	m.receivedMu.Lock()
	defer m.receivedMu.Unlock()
	result := make([][]byte, len(m.received))
	copy(result, m.received)
	return result
}

// Close shuts down the mock server.
func (m *MockUDPServer) Close() {
	close(m.done)
	m.Conn.Close()
}
