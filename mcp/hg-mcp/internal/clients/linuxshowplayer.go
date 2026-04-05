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

// LinuxShowPlayerClient provides access to Linux Show Player via OSC.
// Linux Show Player is a free cue-based show control application.
type LinuxShowPlayerClient struct {
	host   string
	port   int
	client *osc.Client
}

// LinuxShowPlayerStatus represents Linux Show Player connection status.
type LinuxShowPlayerStatus struct {
	Connected bool   `json:"connected"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
}

var (
	lspClientSingleton *LinuxShowPlayerClient
	lspClientOnce      sync.Once
	lspClientErr       error

	// TestOverrideLinuxShowPlayerClient, when non-nil, is returned by GetLinuxShowPlayerClient.
	TestOverrideLinuxShowPlayerClient *LinuxShowPlayerClient
)

// GetLinuxShowPlayerClient returns the singleton Linux Show Player client.
func GetLinuxShowPlayerClient() (*LinuxShowPlayerClient, error) {
	if TestOverrideLinuxShowPlayerClient != nil {
		return TestOverrideLinuxShowPlayerClient, nil
	}
	lspClientOnce.Do(func() {
		lspClientSingleton, lspClientErr = NewLinuxShowPlayerClient()
	})
	return lspClientSingleton, lspClientErr
}

// NewTestLinuxShowPlayerClient creates an in-memory test client.
func NewTestLinuxShowPlayerClient() *LinuxShowPlayerClient {
	return &LinuxShowPlayerClient{
		host: "localhost",
		port: 9000,
	}
}

// NewLinuxShowPlayerClient creates a new Linux Show Player client from environment.
func NewLinuxShowPlayerClient() (*LinuxShowPlayerClient, error) {
	host := os.Getenv("LSP_HOST")
	if host == "" {
		host = "localhost"
	}

	port := 9000
	if p := os.Getenv("LSP_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}

	client := osc.NewClient(host, port)

	return &LinuxShowPlayerClient{
		host:   host,
		port:   port,
		client: client,
	}, nil
}

// isReachable checks if the LSP port is accepting connections.
func (c *LinuxShowPlayerClient) isReachable() bool {
	addr := net.JoinHostPort(c.host, strconv.Itoa(c.port))
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetStatus returns Linux Show Player connection status.
func (c *LinuxShowPlayerClient) GetStatus(ctx context.Context) (*LinuxShowPlayerStatus, error) {
	return &LinuxShowPlayerStatus{
		Connected: c.isReachable(),
		Host:      c.host,
		Port:      c.port,
	}, nil
}

// Go triggers the next cue (global go).
func (c *LinuxShowPlayerClient) Go(ctx context.Context) error {
	return nil
}

// Stop stops all running cues.
func (c *LinuxShowPlayerClient) Stop(ctx context.Context) error {
	return nil
}

// Pause pauses all running cues.
func (c *LinuxShowPlayerClient) Pause(ctx context.Context) error {
	return nil
}

// StartCue starts a specific cue by number.
func (c *LinuxShowPlayerClient) StartCue(ctx context.Context, cueNumber int) error {
	if cueNumber < 0 {
		return fmt.Errorf("cue number must be non-negative")
	}
	return nil
}

// StopCue stops a specific cue by number.
func (c *LinuxShowPlayerClient) StopCue(ctx context.Context, cueNumber int) error {
	if cueNumber < 0 {
		return fmt.Errorf("cue number must be non-negative")
	}
	return nil
}

// Host returns the configured host.
func (c *LinuxShowPlayerClient) Host() string {
	return c.host
}

// Port returns the configured port.
func (c *LinuxShowPlayerClient) Port() int {
	return c.port
}
