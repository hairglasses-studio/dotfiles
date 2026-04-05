// Package clients provides API clients for external services.
package clients

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
)

// SACNClient provides sACN/E1.31 protocol support for DMX over IP.
// sACN (Streaming Architecture for Control Networks) uses UDP multicast
// or unicast to transmit DMX512 data. Each universe carries 512 channels.
type SACNClient struct {
	// bindAddr is the local interface address for multicast
	bindAddr string
	// sourceName identifies this sender in the sACN packets
	sourceName string
	// priority is the sACN source priority (0-200, default 100)
	priority int
}

// SACNStatus represents sACN system status.
type SACNStatus struct {
	Active     bool   `json:"active"`
	BindAddr   string `json:"bind_addr"`
	SourceName string `json:"source_name"`
	Priority   int    `json:"priority"`
}

// SACNUniverse represents a sACN universe with channel data.
type SACNUniverse struct {
	ID       int    `json:"id"`
	Channels []byte `json:"channels,omitempty"`
	Active   bool   `json:"active"`
}

// SACNSource represents a discovered sACN source.
type SACNSource struct {
	Name     string `json:"name"`
	CID      string `json:"cid"` // Component Identifier (UUID)
	Universe int    `json:"universe"`
	Priority int    `json:"priority"`
	IP       string `json:"ip"`
}

// SACNHealth represents sACN system health.
type SACNHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

var (
	sacnClientSingleton *SACNClient
	sacnClientOnce      sync.Once
	sacnClientErr       error

	// TestOverrideSACNClient, when non-nil, is returned by GetSACNClient.
	TestOverrideSACNClient *SACNClient
)

// GetSACNClient returns the singleton sACN client.
func GetSACNClient() (*SACNClient, error) {
	if TestOverrideSACNClient != nil {
		return TestOverrideSACNClient, nil
	}
	sacnClientOnce.Do(func() {
		sacnClientSingleton, sacnClientErr = NewSACNClient()
	})
	return sacnClientSingleton, sacnClientErr
}

// NewTestSACNClient creates an in-memory test client.
func NewTestSACNClient() *SACNClient {
	return &SACNClient{
		bindAddr:   "0.0.0.0",
		sourceName: "hg-mcp-test",
		priority:   100,
	}
}

// NewSACNClient creates a new sACN client from environment.
func NewSACNClient() (*SACNClient, error) {
	bindAddr := os.Getenv("SACN_BIND_ADDR")
	if bindAddr == "" {
		bindAddr = "0.0.0.0"
	}

	sourceName := os.Getenv("SACN_SOURCE_NAME")
	if sourceName == "" {
		sourceName = "hg-mcp"
	}

	priority := 100
	if p := os.Getenv("SACN_PRIORITY"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed >= 0 && parsed <= 200 {
			priority = parsed
		}
	}

	return &SACNClient{
		bindAddr:   bindAddr,
		sourceName: sourceName,
		priority:   priority,
	}, nil
}

// GetStatus returns sACN system status.
func (c *SACNClient) GetStatus(ctx context.Context) (*SACNStatus, error) {
	return &SACNStatus{
		Active:     true,
		BindAddr:   c.bindAddr,
		SourceName: c.sourceName,
		Priority:   c.priority,
	}, nil
}

// SendUniverse sends DMX data to a sACN universe.
// Universe range: 1-63999. Data is 1-512 bytes.
func (c *SACNClient) SendUniverse(ctx context.Context, universe int, data []byte) error {
	if universe < 1 || universe > 63999 {
		return fmt.Errorf("universe %d out of range (1-63999)", universe)
	}
	if len(data) == 0 || len(data) > 512 {
		return fmt.Errorf("data length %d out of range (1-512)", len(data))
	}
	// Stub — real implementation builds E1.31 packet and sends via UDP multicast
	// Multicast address: 239.255.UNIVhi.UNIVlo (universe mapped to IPv4 multicast)
	return nil
}

// SendChannel sets a single channel value in a universe.
func (c *SACNClient) SendChannel(ctx context.Context, universe, channel int, value byte) error {
	if channel < 1 || channel > 512 {
		return fmt.Errorf("channel %d out of range (1-512)", channel)
	}
	data := make([]byte, channel)
	data[channel-1] = value
	return c.SendUniverse(ctx, universe, data)
}

// DiscoverSources listens for sACN source discovery packets.
// Returns sources found within the timeout period.
func (c *SACNClient) DiscoverSources(ctx context.Context) ([]SACNSource, error) {
	// Stub — real implementation joins multicast group 239.255.250.0 (universe discovery)
	// and listens for E1.31 Universe Discovery packets
	return []SACNSource{}, nil
}

// GetHealth returns sACN system health.
func (c *SACNClient) GetHealth(ctx context.Context) (*SACNHealth, error) {
	health := &SACNHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check if we can bind to the configured address
	addr, err := net.ResolveUDPAddr("udp4", c.bindAddr+":0")
	if err != nil {
		health.Score -= 50
		health.Issues = append(health.Issues, fmt.Sprintf("Cannot resolve bind address: %v", err))
		health.Recommendations = append(health.Recommendations, "Check SACN_BIND_ADDR configuration")
	} else {
		conn, err := net.ListenUDP("udp4", addr)
		if err != nil {
			health.Score -= 30
			health.Issues = append(health.Issues, fmt.Sprintf("Cannot bind UDP socket: %v", err))
			health.Recommendations = append(health.Recommendations, "Check network interface and firewall settings")
		} else {
			conn.Close()
		}
	}

	switch {
	case health.Score >= 80:
		health.Status = "healthy"
	case health.Score >= 50:
		health.Status = "degraded"
	default:
		health.Status = "critical"
	}

	return health, nil
}

// multicastAddr returns the sACN multicast address for a universe.
// sACN uses 239.255.{high byte}.{low byte} for universe addressing.
func multicastAddr(universe int) string {
	hi := (universe >> 8) & 0xFF
	lo := universe & 0xFF
	return fmt.Sprintf("239.255.%d.%d", hi, lo)
}

// BindAddr returns the configured bind address.
func (c *SACNClient) BindAddr() string {
	return c.bindAddr
}

// SourceName returns the configured source name.
func (c *SACNClient) SourceName() string {
	return c.sourceName
}

// Priority returns the configured priority.
func (c *SACNClient) Priority() int {
	return c.priority
}
