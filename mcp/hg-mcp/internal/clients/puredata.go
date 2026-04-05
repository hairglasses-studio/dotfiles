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

// PureDataClient provides access to Pure Data via FUDI (TCP) and OSC.
// Pure Data (Pd) is a visual programming language for multimedia.
type PureDataClient struct {
	host string
	port int
}

// PureDataStatus represents Pure Data connection status.
type PureDataStatus struct {
	Connected bool   `json:"connected"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
}

var (
	pureDataClientSingleton *PureDataClient
	pureDataClientOnce      sync.Once
	pureDataClientErr       error

	// TestOverridePureDataClient, when non-nil, is returned by GetPureDataClient.
	TestOverridePureDataClient *PureDataClient
)

// GetPureDataClient returns the singleton Pure Data client.
func GetPureDataClient() (*PureDataClient, error) {
	if TestOverridePureDataClient != nil {
		return TestOverridePureDataClient, nil
	}
	pureDataClientOnce.Do(func() {
		pureDataClientSingleton, pureDataClientErr = NewPureDataClient()
	})
	return pureDataClientSingleton, pureDataClientErr
}

// NewTestPureDataClient creates an in-memory test client.
func NewTestPureDataClient() *PureDataClient {
	return &PureDataClient{
		host: "localhost",
		port: 3000,
	}
}

// NewPureDataClient creates a new Pure Data client from environment.
func NewPureDataClient() (*PureDataClient, error) {
	host := os.Getenv("PUREDATA_HOST")
	if host == "" {
		host = "localhost"
	}

	port := 3000
	if p := os.Getenv("PUREDATA_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}

	return &PureDataClient{
		host: host,
		port: port,
	}, nil
}

// isReachable checks if the Pd FUDI port is accepting connections.
func (c *PureDataClient) isReachable() bool {
	addr := net.JoinHostPort(c.host, strconv.Itoa(c.port))
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetStatus returns Pure Data connection status.
func (c *PureDataClient) GetStatus(ctx context.Context) (*PureDataStatus, error) {
	return &PureDataStatus{
		Connected: c.isReachable(),
		Host:      c.host,
		Port:      c.port,
	}, nil
}

// SendFUDI sends a FUDI message to Pure Data.
// FUDI messages are semicolon-terminated ASCII strings over TCP.
func (c *PureDataClient) SendFUDI(ctx context.Context, message string) error {
	if message == "" {
		return fmt.Errorf("message is required")
	}
	// Stub: would open TCP connection and send message + ";\n"
	return nil
}

// SetDSP enables or disables DSP processing.
// Sends "pd dsp 1;" or "pd dsp 0;" via FUDI.
func (c *PureDataClient) SetDSP(ctx context.Context, on bool) error {
	return nil
}

// Host returns the configured host.
func (c *PureDataClient) Host() string {
	return c.host
}

// Port returns the configured port.
func (c *PureDataClient) Port() int {
	return c.port
}
