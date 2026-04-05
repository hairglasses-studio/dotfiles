// Package clients provides API clients for external services.
package clients

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

// OPCClient provides access to Open Pixel Control LED servers via TCP binary protocol.
// OPC is a simple protocol for controlling arrays of RGB LEDs.
type OPCClient struct {
	host string
	port int
}

// OPCStatus represents OPC server connection status.
type OPCStatus struct {
	Connected bool   `json:"connected"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
}

var (
	opcClientSingleton *OPCClient
	opcClientOnce      sync.Once
	opcClientErr       error

	// TestOverrideOPCClient, when non-nil, is returned by GetOPCClient.
	TestOverrideOPCClient *OPCClient
)

// GetOPCClient returns the singleton OPC client.
func GetOPCClient() (*OPCClient, error) {
	if TestOverrideOPCClient != nil {
		return TestOverrideOPCClient, nil
	}
	opcClientOnce.Do(func() {
		opcClientSingleton, opcClientErr = NewOPCClient()
	})
	return opcClientSingleton, opcClientErr
}

// NewTestOPCClient creates an in-memory test client.
func NewTestOPCClient() *OPCClient {
	return &OPCClient{
		host: "localhost",
		port: 7890,
	}
}

// NewOPCClient creates a new OPC client from environment.
func NewOPCClient() (*OPCClient, error) {
	host := os.Getenv("OPC_HOST")
	if host == "" {
		host = "localhost"
	}

	port := 7890
	if p := os.Getenv("OPC_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}

	return &OPCClient{
		host: host,
		port: port,
	}, nil
}

// isReachable checks if the OPC server is accepting TCP connections.
func (c *OPCClient) isReachable() bool {
	addr := net.JoinHostPort(c.host, strconv.Itoa(c.port))
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetStatus returns OPC server connection status.
func (c *OPCClient) GetStatus(ctx context.Context) (*OPCStatus, error) {
	return &OPCStatus{
		Connected: c.isReachable(),
		Host:      c.host,
		Port:      c.port,
	}, nil
}

// SetPixels sends RGB pixel data to a channel.
// OPC message format: channel(1) + command(1) + length_hi(1) + length_lo(1) + data
// Command 0 = set pixel colors. Data = R,G,B,R,G,B,...
func (c *OPCClient) SetPixels(ctx context.Context, channel int, pixels []byte) error {
	if channel < 0 || channel > 255 {
		return fmt.Errorf("channel must be 0-255")
	}
	if len(pixels) == 0 {
		return fmt.Errorf("pixel data is required")
	}
	if len(pixels)%3 != 0 {
		return fmt.Errorf("pixel data must be a multiple of 3 (RGB triplets)")
	}
	if len(pixels) > 65535 {
		return fmt.Errorf("pixel data exceeds maximum length (65535 bytes)")
	}
	// Stub: would build 4-byte header + data and send over TCP
	return nil
}

// Host returns the configured host.
func (c *OPCClient) Host() string {
	return c.host
}

// Port returns the configured port.
func (c *OPCClient) Port() int {
	return c.port
}
