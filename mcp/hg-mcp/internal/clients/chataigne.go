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

// ChataigneClient provides access to Chataigne show control via HTTP API + OSC.
// Chataigne is an open source show control application that bridges
// OSC, MIDI, DMX, ArtNet, sACN, HTTP, WebSocket, MQTT, PJLink, and Ableton Link.
type ChataigneClient struct {
	host    string
	port    int
	oscPort int
}

// ChataigneStatus represents Chataigne connection status.
type ChataigneStatus struct {
	Connected bool   `json:"connected"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	OSCPort   int    `json:"osc_port"`
	URL       string `json:"url"`
}

// ChataigneModule represents a Chataigne module (protocol bridge).
type ChataigneModule struct {
	Name    string `json:"name"`
	Type    string `json:"type"` // OSC, MIDI, DMX, HTTP, Serial, etc.
	Enabled bool   `json:"enabled"`
	Status  string `json:"status"`
}

// ChataigneStateMachine represents the state machine status.
type ChataigneStateMachine struct {
	CurrentState string   `json:"current_state"`
	States       []string `json:"states"`
}

// ChataigneSequence represents a sequence/timeline.
type ChataigneSequence struct {
	Name     string  `json:"name"`
	Playing  bool    `json:"playing"`
	Position float64 `json:"position"` // seconds
	Duration float64 `json:"duration"` // seconds
}

// ChataigneHealth represents system health.
type ChataigneHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

var (
	chataigneClientSingleton *ChataigneClient
	chataigneClientOnce      sync.Once
	chataigneClientErr       error

	// TestOverrideChataigneClient, when non-nil, is returned by GetChataigneClient.
	TestOverrideChataigneClient *ChataigneClient
)

// GetChataigneClient returns the singleton Chataigne client.
func GetChataigneClient() (*ChataigneClient, error) {
	if TestOverrideChataigneClient != nil {
		return TestOverrideChataigneClient, nil
	}
	chataigneClientOnce.Do(func() {
		chataigneClientSingleton, chataigneClientErr = NewChataigneClient()
	})
	return chataigneClientSingleton, chataigneClientErr
}

// NewTestChataigneClient creates an in-memory test client.
func NewTestChataigneClient() *ChataigneClient {
	return &ChataigneClient{
		host:    "localhost",
		port:    9000,
		oscPort: 12000,
	}
}

// NewChataigneClient creates a new Chataigne client from environment.
func NewChataigneClient() (*ChataigneClient, error) {
	host := os.Getenv("CHATAIGNE_HOST")
	if host == "" {
		host = "localhost"
	}

	port := 9000 // Default Chataigne HTTP port
	if p := os.Getenv("CHATAIGNE_PORT"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			port = parsed
		}
	}

	oscPort := 12000 // Default Chataigne OSC port
	if p := os.Getenv("CHATAIGNE_OSC_PORT"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			oscPort = parsed
		}
	}

	return &ChataigneClient{
		host:    host,
		port:    port,
		oscPort: oscPort,
	}, nil
}

// baseURL returns the HTTP base URL.
func (c *ChataigneClient) baseURL() string {
	return fmt.Sprintf("http://%s:%d", c.host, c.port)
}

// isReachable checks connectivity.
func (c *ChataigneClient) isReachable() bool {
	addr := net.JoinHostPort(c.host, strconv.Itoa(c.port))
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetStatus returns Chataigne connection status.
func (c *ChataigneClient) GetStatus(ctx context.Context) (*ChataigneStatus, error) {
	return &ChataigneStatus{
		Connected: c.isReachable(),
		Host:      c.host,
		Port:      c.port,
		OSCPort:   c.oscPort,
		URL:       c.baseURL(),
	}, nil
}

// GetModules returns the list of protocol modules.
func (c *ChataigneClient) GetModules(ctx context.Context) ([]ChataigneModule, error) {
	if !c.isReachable() {
		return nil, fmt.Errorf("Chataigne not reachable at %s:%d", c.host, c.port)
	}
	// Stub — real implementation queries HTTP API
	return []ChataigneModule{}, nil
}

// GetStateMachine returns the state machine status.
func (c *ChataigneClient) GetStateMachine(ctx context.Context) (*ChataigneStateMachine, error) {
	if !c.isReachable() {
		return nil, fmt.Errorf("Chataigne not reachable at %s:%d", c.host, c.port)
	}
	// Stub
	return &ChataigneStateMachine{}, nil
}

// TriggerState transitions the state machine to a new state.
func (c *ChataigneClient) TriggerState(ctx context.Context, state string) error {
	if !c.isReachable() {
		return fmt.Errorf("Chataigne not reachable at %s:%d", c.host, c.port)
	}
	if state == "" {
		return fmt.Errorf("state name is required")
	}
	// Stub — real implementation sends OSC or HTTP command
	return nil
}

// GetSequences returns available sequences/timelines.
func (c *ChataigneClient) GetSequences(ctx context.Context) ([]ChataigneSequence, error) {
	if !c.isReachable() {
		return nil, fmt.Errorf("Chataigne not reachable at %s:%d", c.host, c.port)
	}
	// Stub
	return []ChataigneSequence{}, nil
}

// TriggerSequence starts or stops a sequence.
func (c *ChataigneClient) TriggerSequence(ctx context.Context, name string, play bool) error {
	if !c.isReachable() {
		return fmt.Errorf("Chataigne not reachable at %s:%d", c.host, c.port)
	}
	if name == "" {
		return fmt.Errorf("sequence name is required")
	}
	// Stub
	return nil
}

// GetHealth returns Chataigne system health.
func (c *ChataigneClient) GetHealth(ctx context.Context) (*ChataigneHealth, error) {
	health := &ChataigneHealth{
		Score:  100,
		Status: "healthy",
	}

	if !c.isReachable() {
		health.Score = 0
		health.Status = "critical"
		health.Issues = append(health.Issues, fmt.Sprintf("Chataigne not reachable at %s:%d", c.host, c.port))
		health.Recommendations = append(health.Recommendations,
			"Start Chataigne and enable HTTP/OSC control",
			fmt.Sprintf("Verify CHATAIGNE_HOST=%s and CHATAIGNE_PORT=%d", c.host, c.port),
		)
	}

	return health, nil
}

// Host returns the configured host.
func (c *ChataigneClient) Host() string { return c.host }

// Port returns the configured HTTP port.
func (c *ChataigneClient) Port() int { return c.port }

// OSCPort returns the configured OSC port.
func (c *ChataigneClient) OSCPort() int { return c.oscPort }
