package testutil

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestMockHTTPServerRouting(t *testing.T) {
	srv := NewMockHTTPServer(map[string]MockHTTPResponse{
		"GET /api/status": {
			StatusCode: 200,
			Body:       `{"connected":true}`,
			Headers:    map[string]string{"Content-Type": "application/json"},
		},
		"POST /api/action": {
			StatusCode: 201,
			Body:       `{"ok":true}`,
		},
	})
	defer srv.Close()

	// Test GET route
	resp, err := http.Get(srv.URL() + "/api/status")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != `{"connected":true}` {
		t.Errorf("body = %q, want %q", string(body), `{"connected":true}`)
	}

	// Test 404 for unmatched route
	resp2, err := http.Get(srv.URL() + "/unknown")
	if err != nil {
		t.Fatalf("GET /unknown failed: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != 404 {
		t.Errorf("unmatched route status = %d, want 404", resp2.StatusCode)
	}
}

func TestMockHTTPServerJSONRoute(t *testing.T) {
	srv := NewMockHTTPServer(nil)
	defer srv.Close()

	srv.JSONRoute("GET", "/plugins", 200, []map[string]string{
		{"name": "ArtNet", "active": "true"},
	})

	resp, err := http.Get(srv.URL() + "/plugins")
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("content-type = %q, want application/json", ct)
	}
}

func TestMockHTTPServerSetRoute(t *testing.T) {
	srv := NewMockHTTPServer(nil)
	defer srv.Close()

	// Initially no routes — should 404
	resp, _ := http.Get(srv.URL() + "/test")
	resp.Body.Close()
	if resp.StatusCode != 404 {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}

	// Add route dynamically
	srv.SetRoute("GET", "/test", MockHTTPResponse{StatusCode: 200, Body: "ok"})

	resp2, _ := http.Get(srv.URL() + "/test")
	body, _ := io.ReadAll(resp2.Body)
	resp2.Body.Close()
	if string(body) != "ok" {
		t.Errorf("body = %q, want %q", string(body), "ok")
	}
}

func TestMockTCPServer(t *testing.T) {
	srv, err := NewMockTCPServer([]byte("HELLO\n"))
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}
	defer srv.Close()

	if srv.Port <= 0 {
		t.Errorf("port = %d, want positive", srv.Port)
	}

	// Connect and read response
	conn, err := net.DialTimeout("tcp", srv.Addr, 2*time.Second)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	// Send something to trigger response
	fmt.Fprint(conn, "test\n")

	// Read response
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 100)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(buf[:n]) != "HELLO\n" {
		t.Errorf("response = %q, want %q", string(buf[:n]), "HELLO\n")
	}
}

func TestMockTCPServerSetResponse(t *testing.T) {
	srv, err := NewMockTCPServer([]byte("v1"))
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}
	defer srv.Close()

	srv.SetResponse([]byte("v2"))

	conn, err := net.DialTimeout("tcp", srv.Addr, 2*time.Second)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	fmt.Fprint(conn, "ping")
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 100)
	n, _ := conn.Read(buf)
	if string(buf[:n]) != "v2" {
		t.Errorf("response = %q, want %q", string(buf[:n]), "v2")
	}
}

func TestMockUDPServer(t *testing.T) {
	srv, err := NewMockUDPServer(func(data []byte) []byte {
		return []byte("echo:" + string(data))
	})
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}
	defer srv.Close()

	if srv.Port <= 0 {
		t.Errorf("port = %d, want positive", srv.Port)
	}

	// Send a UDP packet
	conn, err := net.DialTimeout("udp", srv.Addr, 2*time.Second)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	fmt.Fprint(conn, "hello")

	// Read echo response
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 100)
	n, err := conn.Read(buf)
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	if string(buf[:n]) != "echo:hello" {
		t.Errorf("response = %q, want %q", string(buf[:n]), "echo:hello")
	}

	// Verify received packets
	time.Sleep(10 * time.Millisecond)
	received := srv.Received()
	if len(received) != 1 {
		t.Errorf("received %d packets, want 1", len(received))
	}
}

func TestMockUDPServerRecordOnly(t *testing.T) {
	srv, err := NewMockUDPServer(nil) // no response function
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}
	defer srv.Close()

	conn, err := net.DialTimeout("udp", srv.Addr, 2*time.Second)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	fmt.Fprint(conn, "packet1")
	time.Sleep(10 * time.Millisecond)
	fmt.Fprint(conn, "packet2")
	time.Sleep(10 * time.Millisecond)

	received := srv.Received()
	if len(received) != 2 {
		t.Errorf("received %d packets, want 2", len(received))
	}
}
