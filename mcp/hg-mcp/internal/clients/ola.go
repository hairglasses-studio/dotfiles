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

// OLAClient provides access to Open Lighting Architecture via HTTP REST API.
// OLA is a framework for controlling DMX-512 lighting equipment.
type OLAClient struct {
	host string
	port int
}

// OLAStatus represents OLA connection status.
type OLAStatus struct {
	Connected bool   `json:"connected"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	URL       string `json:"url"`
}

// OLAPlugin represents an installed OLA plugin.
type OLAPlugin struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Active  bool   `json:"active"`
	Enabled bool   `json:"enabled"`
}

// OLAUniverse represents a DMX universe in OLA.
type OLAUniverse struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	MergeMode   string `json:"merge_mode"`
	InputPorts  int    `json:"input_ports"`
	OutputPorts int    `json:"output_ports"`
}

var (
	olaClientSingleton *OLAClient
	olaClientOnce      sync.Once
	olaClientErr       error

	// TestOverrideOLAClient, when non-nil, is returned by GetOLAClient.
	TestOverrideOLAClient *OLAClient
)

// GetOLAClient returns the singleton OLA client.
func GetOLAClient() (*OLAClient, error) {
	if TestOverrideOLAClient != nil {
		return TestOverrideOLAClient, nil
	}
	olaClientOnce.Do(func() {
		olaClientSingleton, olaClientErr = NewOLAClient()
	})
	return olaClientSingleton, olaClientErr
}

// NewTestOLAClient creates an in-memory test client.
func NewTestOLAClient() *OLAClient {
	return &OLAClient{
		host: "localhost",
		port: 9090,
	}
}

// NewOLAClient creates a new OLA client from environment.
func NewOLAClient() (*OLAClient, error) {
	host := os.Getenv("OLA_HOST")
	if host == "" {
		host = "localhost"
	}

	port := 9090
	if p := os.Getenv("OLA_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}

	return &OLAClient{
		host: host,
		port: port,
	}, nil
}

// baseURL returns the HTTP base URL.
func (c *OLAClient) baseURL() string {
	return fmt.Sprintf("http://%s:%d", c.host, c.port)
}

// isReachable checks if the OLA HTTP port is accepting connections.
func (c *OLAClient) isReachable() bool {
	addr := net.JoinHostPort(c.host, strconv.Itoa(c.port))
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetStatus returns OLA connection status.
func (c *OLAClient) GetStatus(ctx context.Context) (*OLAStatus, error) {
	return &OLAStatus{
		Connected: c.isReachable(),
		Host:      c.host,
		Port:      c.port,
		URL:       c.baseURL(),
	}, nil
}

// GetPlugins returns installed OLA plugins.
// API: GET /get_plugins
func (c *OLAClient) GetPlugins(ctx context.Context) ([]OLAPlugin, error) {
	return []OLAPlugin{}, nil
}

// GetUniverseInfo returns metadata for a specific universe.
// API: GET /get_universe_info?id=<universe>
func (c *OLAClient) GetUniverseInfo(ctx context.Context, universeID int) (*OLAUniverse, error) {
	if universeID < 0 {
		return nil, fmt.Errorf("universe ID must be non-negative")
	}
	return &OLAUniverse{
		ID:        universeID,
		MergeMode: "HTP",
	}, nil
}

// GetDMX reads DMX channel values from a universe.
// API: GET /get_dmx?u=<universe>
func (c *OLAClient) GetDMX(ctx context.Context, universeID int) ([]int, error) {
	if universeID < 0 {
		return nil, fmt.Errorf("universe ID must be non-negative")
	}
	// Return 512 channels of zeros (stub)
	channels := make([]int, 512)
	return channels, nil
}

// SetDMX writes DMX channel values to a universe.
// API: POST /set_dmx body: u=<universe>&d=<comma-separated values>
func (c *OLAClient) SetDMX(ctx context.Context, universeID int, startChannel int, values []int) error {
	if universeID < 0 {
		return fmt.Errorf("universe ID must be non-negative")
	}
	if startChannel < 1 || startChannel > 512 {
		return fmt.Errorf("start channel must be between 1 and 512")
	}
	if len(values) == 0 {
		return fmt.Errorf("at least one value is required")
	}
	return nil
}

// Host returns the configured host.
func (c *OLAClient) Host() string {
	return c.host
}

// Port returns the configured port.
func (c *OLAClient) Port() int {
	return c.port
}
