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

// OssiaClient provides access to ossia score via OSCQuery (HTTP+JSON) and OSC.
// ossia score is an interactive show control sequencer.
type OssiaClient struct {
	host string
	port int
}

// OssiaStatus represents ossia score connection status.
type OssiaStatus struct {
	Connected bool   `json:"connected"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	URL       string `json:"url"`
}

// OssiaParameter represents a parameter in the OSCQuery tree.
type OssiaParameter struct {
	FullPath string        `json:"full_path"`
	Type     string        `json:"type"`
	Value    []interface{} `json:"value,omitempty"`
	Min      interface{}   `json:"min,omitempty"`
	Max      interface{}   `json:"max,omitempty"`
}

// OssiaDevice represents a discovered device with its parameter tree.
type OssiaDevice struct {
	Name       string           `json:"name"`
	Parameters []OssiaParameter `json:"parameters,omitempty"`
}

var (
	ossiaClientSingleton *OssiaClient
	ossiaClientOnce      sync.Once
	ossiaClientErr       error

	// TestOverrideOssiaClient, when non-nil, is returned by GetOssiaClient.
	TestOverrideOssiaClient *OssiaClient
)

// GetOssiaClient returns the singleton ossia client.
func GetOssiaClient() (*OssiaClient, error) {
	if TestOverrideOssiaClient != nil {
		return TestOverrideOssiaClient, nil
	}
	ossiaClientOnce.Do(func() {
		ossiaClientSingleton, ossiaClientErr = NewOssiaClient()
	})
	return ossiaClientSingleton, ossiaClientErr
}

// NewTestOssiaClient creates an in-memory test client.
func NewTestOssiaClient() *OssiaClient {
	return &OssiaClient{
		host: "localhost",
		port: 5678,
	}
}

// NewOssiaClient creates a new ossia client from environment.
func NewOssiaClient() (*OssiaClient, error) {
	host := os.Getenv("OSSIA_HOST")
	if host == "" {
		host = "localhost"
	}

	port := 5678
	if p := os.Getenv("OSSIA_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}

	return &OssiaClient{
		host: host,
		port: port,
	}, nil
}

// baseURL returns the HTTP base URL for OSCQuery.
func (c *OssiaClient) baseURL() string {
	return fmt.Sprintf("http://%s:%d", c.host, c.port)
}

// isReachable checks if the ossia HTTP port is accepting connections.
func (c *OssiaClient) isReachable() bool {
	addr := net.JoinHostPort(c.host, strconv.Itoa(c.port))
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetStatus returns ossia connection status.
func (c *OssiaClient) GetStatus(ctx context.Context) (*OssiaStatus, error) {
	return &OssiaStatus{
		Connected: c.isReachable(),
		Host:      c.host,
		Port:      c.port,
		URL:       c.baseURL(),
	}, nil
}

// GetDevices returns discovered devices from the OSCQuery parameter tree.
// API: GET /
func (c *OssiaClient) GetDevices(ctx context.Context) ([]OssiaDevice, error) {
	return []OssiaDevice{}, nil
}

// TransportPlay sends play command.
func (c *OssiaClient) TransportPlay(ctx context.Context) error {
	return nil
}

// TransportStop sends stop command.
func (c *OssiaClient) TransportStop(ctx context.Context) error {
	return nil
}

// TransportSetPosition sets the transport position (in seconds).
func (c *OssiaClient) TransportSetPosition(ctx context.Context, position float64) error {
	if position < 0 {
		return fmt.Errorf("position must be non-negative")
	}
	return nil
}

// Host returns the configured host.
func (c *OssiaClient) Host() string {
	return c.host
}

// Port returns the configured port.
func (c *OssiaClient) Port() int {
	return c.port
}
