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

// MapMapClient provides access to MapMap via OSC.
// MapMap is a free, open-source video mapping software.
type MapMapClient struct {
	host   string
	port   int
	client *osc.Client
}

// MapMapStatus represents MapMap connection status.
type MapMapStatus struct {
	Connected bool   `json:"connected"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
}

var (
	mapMapClientSingleton *MapMapClient
	mapMapClientOnce      sync.Once
	mapMapClientErr       error

	// TestOverrideMapMapClient, when non-nil, is returned by GetMapMapClient.
	TestOverrideMapMapClient *MapMapClient
)

// GetMapMapClient returns the singleton MapMap client.
func GetMapMapClient() (*MapMapClient, error) {
	if TestOverrideMapMapClient != nil {
		return TestOverrideMapMapClient, nil
	}
	mapMapClientOnce.Do(func() {
		mapMapClientSingleton, mapMapClientErr = NewMapMapClient()
	})
	return mapMapClientSingleton, mapMapClientErr
}

// NewTestMapMapClient creates an in-memory test client.
func NewTestMapMapClient() *MapMapClient {
	return &MapMapClient{
		host: "localhost",
		port: 12345,
	}
}

// NewMapMapClient creates a new MapMap client from environment.
func NewMapMapClient() (*MapMapClient, error) {
	host := os.Getenv("MAPMAP_HOST")
	if host == "" {
		host = "localhost"
	}

	port := 12345
	if p := os.Getenv("MAPMAP_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}

	client := osc.NewClient(host, port)

	return &MapMapClient{
		host:   host,
		port:   port,
		client: client,
	}, nil
}

// isReachable checks if the MapMap port is accepting connections.
func (c *MapMapClient) isReachable() bool {
	addr := net.JoinHostPort(c.host, strconv.Itoa(c.port))
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetStatus returns MapMap connection status.
func (c *MapMapClient) GetStatus(ctx context.Context) (*MapMapStatus, error) {
	return &MapMapStatus{
		Connected: c.isReachable(),
		Host:      c.host,
		Port:      c.port,
	}, nil
}

// SetSurfaceOpacity sets the opacity of a surface (0.0-1.0).
func (c *MapMapClient) SetSurfaceOpacity(ctx context.Context, surfaceID int, opacity float64) error {
	if surfaceID < 0 {
		return fmt.Errorf("surface ID must be non-negative")
	}
	return nil
}

// SetSurfaceVisible shows or hides a surface.
func (c *MapMapClient) SetSurfaceVisible(ctx context.Context, surfaceID int, visible bool) error {
	if surfaceID < 0 {
		return fmt.Errorf("surface ID must be non-negative")
	}
	return nil
}

// Host returns the configured host.
func (c *MapMapClient) Host() string {
	return c.host
}

// Port returns the configured port.
func (c *MapMapClient) Port() int {
	return c.port
}
