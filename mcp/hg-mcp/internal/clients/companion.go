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

// CompanionClient provides access to Bitfocus Companion via HTTP REST API.
// Companion is a modular stream deck / button box controller that supports
// 400+ device modules for controlling professional AV equipment.
type CompanionClient struct {
	host string
	port int
}

// CompanionStatus represents Companion connection status.
type CompanionStatus struct {
	Connected bool   `json:"connected"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	URL       string `json:"url"`
}

// CompanionVariable represents a Companion custom variable.
type CompanionVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// CompanionInstance represents a connected module instance.
type CompanionInstance struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Module   string `json:"module"`
	Enabled  bool   `json:"enabled"`
	Status   string `json:"status"` // ok, warning, error, disabled
	Category string `json:"category,omitempty"`
}

// CompanionSurface represents a connected control surface (Stream Deck, etc).
type CompanionSurface struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Online bool   `json:"online"`
}

// CompanionHealth represents system health.
type CompanionHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

var (
	companionClientSingleton *CompanionClient
	companionClientOnce      sync.Once
	companionClientErr       error

	// TestOverrideCompanionClient, when non-nil, is returned by GetCompanionClient.
	TestOverrideCompanionClient *CompanionClient
)

// GetCompanionClient returns the singleton Companion client.
func GetCompanionClient() (*CompanionClient, error) {
	if TestOverrideCompanionClient != nil {
		return TestOverrideCompanionClient, nil
	}
	companionClientOnce.Do(func() {
		companionClientSingleton, companionClientErr = NewCompanionClient()
	})
	return companionClientSingleton, companionClientErr
}

// NewTestCompanionClient creates an in-memory test client.
func NewTestCompanionClient() *CompanionClient {
	return &CompanionClient{
		host: "localhost",
		port: 8000,
	}
}

// NewCompanionClient creates a new Companion client from environment.
func NewCompanionClient() (*CompanionClient, error) {
	host := os.Getenv("COMPANION_HOST")
	if host == "" {
		host = "localhost"
	}

	port := 8000 // Default Companion HTTP API port
	if p := os.Getenv("COMPANION_PORT"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			port = parsed
		}
	}

	return &CompanionClient{
		host: host,
		port: port,
	}, nil
}

// baseURL returns the HTTP base URL.
func (c *CompanionClient) baseURL() string {
	return fmt.Sprintf("http://%s:%d", c.host, c.port)
}

// isReachable checks if the Companion HTTP port is accepting connections.
func (c *CompanionClient) isReachable() bool {
	addr := net.JoinHostPort(c.host, strconv.Itoa(c.port))
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetStatus returns Companion connection status.
func (c *CompanionClient) GetStatus(ctx context.Context) (*CompanionStatus, error) {
	return &CompanionStatus{
		Connected: c.isReachable(),
		Host:      c.host,
		Port:      c.port,
		URL:       c.baseURL(),
	}, nil
}

// PressButton presses a button on a specific page and bank.
// API: GET /api/v1/press/bank/{page}/{bank}
func (c *CompanionClient) PressButton(ctx context.Context, page, bank int) error {
	if !c.isReachable() {
		return fmt.Errorf("Companion not reachable at %s:%d", c.host, c.port)
	}
	if page < 1 || page > 99 {
		return fmt.Errorf("page %d out of range (1-99)", page)
	}
	if bank < 1 || bank > 32 {
		return fmt.Errorf("bank %d out of range (1-32)", bank)
	}
	// Stub — real implementation: GET /api/v1/press/bank/{page}/{bank}
	return nil
}

// ReleaseButton releases a button.
// API: GET /api/v1/release/bank/{page}/{bank}
func (c *CompanionClient) ReleaseButton(ctx context.Context, page, bank int) error {
	if !c.isReachable() {
		return fmt.Errorf("Companion not reachable at %s:%d", c.host, c.port)
	}
	// Stub — real implementation: GET /api/v1/release/bank/{page}/{bank}
	return nil
}

// GetVariable gets a custom variable value.
// API: GET /api/v1/custom-variable/{name}
func (c *CompanionClient) GetVariable(ctx context.Context, name string) (string, error) {
	if !c.isReachable() {
		return "", fmt.Errorf("Companion not reachable at %s:%d", c.host, c.port)
	}
	if name == "" {
		return "", fmt.Errorf("variable name is required")
	}
	// Stub — real implementation: GET /api/v1/custom-variable/{name}
	return "", nil
}

// SetVariable sets a custom variable value.
// API: POST /api/v1/custom-variable/{name} with body = value
func (c *CompanionClient) SetVariable(ctx context.Context, name, value string) error {
	if !c.isReachable() {
		return fmt.Errorf("Companion not reachable at %s:%d", c.host, c.port)
	}
	if name == "" {
		return fmt.Errorf("variable name is required")
	}
	// Stub — real implementation: POST /api/v1/custom-variable/{name}
	return nil
}

// GetInstances returns connected module instances.
func (c *CompanionClient) GetInstances(ctx context.Context) ([]CompanionInstance, error) {
	if !c.isReachable() {
		return nil, fmt.Errorf("Companion not reachable at %s:%d", c.host, c.port)
	}
	// Stub — real implementation queries Companion API
	return []CompanionInstance{}, nil
}

// GetSurfaces returns connected control surfaces.
func (c *CompanionClient) GetSurfaces(ctx context.Context) ([]CompanionSurface, error) {
	if !c.isReachable() {
		return nil, fmt.Errorf("Companion not reachable at %s:%d", c.host, c.port)
	}
	// Stub
	return []CompanionSurface{}, nil
}

// SwitchPage switches the active page on a surface.
// API: GET /api/v1/set/page/{surface}/{page}
func (c *CompanionClient) SwitchPage(ctx context.Context, surfaceID string, page int) error {
	if !c.isReachable() {
		return fmt.Errorf("Companion not reachable at %s:%d", c.host, c.port)
	}
	if page < 1 || page > 99 {
		return fmt.Errorf("page %d out of range (1-99)", page)
	}
	// Stub
	return nil
}

// GetHealth returns Companion system health.
func (c *CompanionClient) GetHealth(ctx context.Context) (*CompanionHealth, error) {
	health := &CompanionHealth{
		Score:  100,
		Status: "healthy",
	}

	if !c.isReachable() {
		health.Score = 0
		health.Status = "critical"
		health.Issues = append(health.Issues, fmt.Sprintf("Companion not reachable at %s:%d", c.host, c.port))
		health.Recommendations = append(health.Recommendations,
			"Start Bitfocus Companion",
			fmt.Sprintf("Verify COMPANION_HOST=%s and COMPANION_PORT=%d", c.host, c.port),
		)
		return health, nil
	}

	return health, nil
}

// Host returns the configured host.
func (c *CompanionClient) Host() string {
	return c.host
}

// Port returns the configured port.
func (c *CompanionClient) Port() int {
	return c.port
}
