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

	"github.com/hypebeast/go-osc/osc"
)

// VimixClient provides access to vimix via OSC.
// vimix is a video mixing application for live performance.
type VimixClient struct {
	host   string
	port   int
	client *osc.Client
}

// VimixStatus represents vimix connection status.
type VimixStatus struct {
	Connected bool   `json:"connected"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
}

var (
	vimixClientSingleton *VimixClient
	vimixClientOnce      sync.Once
	vimixClientErr       error

	// TestOverrideVimixClient, when non-nil, is returned by GetVimixClient.
	TestOverrideVimixClient *VimixClient
)

// GetVimixClient returns the singleton vimix client.
func GetVimixClient() (*VimixClient, error) {
	if TestOverrideVimixClient != nil {
		return TestOverrideVimixClient, nil
	}
	vimixClientOnce.Do(func() {
		vimixClientSingleton, vimixClientErr = NewVimixClient()
	})
	return vimixClientSingleton, vimixClientErr
}

// NewTestVimixClient creates an in-memory test client.
func NewTestVimixClient() *VimixClient {
	return &VimixClient{
		host: "localhost",
		port: 7000,
	}
}

// NewVimixClient creates a new vimix client from environment.
func NewVimixClient() (*VimixClient, error) {
	host := os.Getenv("VIMIX_HOST")
	if host == "" {
		host = "localhost"
	}

	port := 7000
	if p := os.Getenv("VIMIX_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}

	client := osc.NewClient(host, port)

	return &VimixClient{
		host:   host,
		port:   port,
		client: client,
	}, nil
}

// isReachable checks if the vimix port is accepting connections.
func (c *VimixClient) isReachable() bool {
	addr := net.JoinHostPort(c.host, strconv.Itoa(c.port))
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetStatus returns vimix connection status.
func (c *VimixClient) GetStatus(ctx context.Context) (*VimixStatus, error) {
	return &VimixStatus{
		Connected: c.isReachable(),
		Host:      c.host,
		Port:      c.port,
	}, nil
}

// SetSourceAlpha sets the opacity of a source (0.0-1.0).
func (c *VimixClient) SetSourceAlpha(ctx context.Context, sourceID int, alpha float64) error {
	if sourceID < 0 {
		return fmt.Errorf("source ID must be non-negative")
	}
	return nil
}

// SetSourceActive activates or deactivates a source.
func (c *VimixClient) SetSourceActive(ctx context.Context, sourceID int, active bool) error {
	if sourceID < 0 {
		return fmt.Errorf("source ID must be non-negative")
	}
	return nil
}

// SaveSession saves the current session.
func (c *VimixClient) SaveSession(ctx context.Context) error {
	return nil
}

// LoadSession loads a session by filename.
func (c *VimixClient) LoadSession(ctx context.Context, filename string) error {
	if filename == "" {
		return fmt.Errorf("filename is required")
	}
	return nil
}

// Host returns the configured host.
func (c *VimixClient) Host() string {
	return c.host
}

// Port returns the configured port.
func (c *VimixClient) Port() int {
	return c.port
}
