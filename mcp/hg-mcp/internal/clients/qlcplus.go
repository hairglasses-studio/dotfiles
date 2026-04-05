// Package clients provides API clients for external services.
package clients

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// QLCPlusClient provides access to QLC+ via WebSocket API.
// QLC+ exposes a WebSocket endpoint at ws://host:port/qlcplusWS
// with a string-based protocol: QLC+API|<command>|<args>
type QLCPlusClient struct {
	host string
	port int
}

// QLCPlusStatus represents QLC+ connection status.
type QLCPlusStatus struct {
	Connected bool   `json:"connected"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	URL       string `json:"url"`
}

// QLCPlusWidget represents a QLC+ virtual console widget.
type QLCPlusWidget struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
}

// QLCPlusFunction represents a QLC+ function.
type QLCPlusFunction struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Status string `json:"status"` // running, stopped
}

// QLCPlusUniverse represents DMX universe info.
type QLCPlusUniverse struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Channels []int  `json:"channels,omitempty"`
}

// QLCPlusHealth represents system health.
type QLCPlusHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

var (
	qlcplusClientSingleton *QLCPlusClient
	qlcplusClientOnce      sync.Once
	qlcplusClientErr       error

	// TestOverrideQLCPlusClient, when non-nil, is returned by GetQLCPlusClient
	// instead of the real singleton. Used for testing without network access.
	TestOverrideQLCPlusClient *QLCPlusClient
)

// GetQLCPlusClient returns the singleton QLC+ client.
func GetQLCPlusClient() (*QLCPlusClient, error) {
	if TestOverrideQLCPlusClient != nil {
		return TestOverrideQLCPlusClient, nil
	}
	qlcplusClientOnce.Do(func() {
		qlcplusClientSingleton, qlcplusClientErr = NewQLCPlusClient()
	})
	return qlcplusClientSingleton, qlcplusClientErr
}

// NewTestQLCPlusClient creates an in-memory test client.
func NewTestQLCPlusClient() *QLCPlusClient {
	return &QLCPlusClient{
		host: "localhost",
		port: 9999,
	}
}

// NewQLCPlusClient creates a new QLC+ client from environment.
func NewQLCPlusClient() (*QLCPlusClient, error) {
	host := os.Getenv("QLCPLUS_HOST")
	if host == "" {
		host = "localhost"
	}

	port := 9999 // Default QLC+ WebSocket port
	if p := os.Getenv("QLCPLUS_PORT"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil {
			port = parsed
		}
	}

	return &QLCPlusClient{
		host: host,
		port: port,
	}, nil
}

// wsURL returns the WebSocket URL.
func (c *QLCPlusClient) wsURL() string {
	return fmt.Sprintf("ws://%s:%d/qlcplusWS", c.host, c.port)
}

// isReachable checks if the QLC+ WebSocket port is accepting connections.
func (c *QLCPlusClient) isReachable() bool {
	addr := net.JoinHostPort(c.host, strconv.Itoa(c.port))
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// GetStatus returns QLC+ connection status.
func (c *QLCPlusClient) GetStatus(ctx context.Context) (*QLCPlusStatus, error) {
	return &QLCPlusStatus{
		Connected: c.isReachable(),
		Host:      c.host,
		Port:      c.port,
		URL:       c.wsURL(),
	}, nil
}

// GetWidgets returns the list of virtual console widgets.
// QLC+ API: QLC+API|getWidgetsList
func (c *QLCPlusClient) GetWidgets(ctx context.Context) ([]QLCPlusWidget, error) {
	if !c.isReachable() {
		return nil, fmt.Errorf("QLC+ not reachable at %s:%d", c.host, c.port)
	}
	// Stub — real implementation sends QLC+API|getWidgetsList via WebSocket
	return []QLCPlusWidget{}, nil
}

// GetWidgetValue returns the current value of a widget.
// QLC+ API: QLC+API|getWidgetStatus|<widgetID>
func (c *QLCPlusClient) GetWidgetValue(ctx context.Context, widgetID string) (string, error) {
	if !c.isReachable() {
		return "", fmt.Errorf("QLC+ not reachable at %s:%d", c.host, c.port)
	}
	if widgetID == "" {
		return "", fmt.Errorf("widget ID is required")
	}
	// Stub — real implementation sends QLC+API|getWidgetStatus|<widgetID>
	return "", nil
}

// SetWidgetValue sets the value of a virtual console widget.
// QLC+ API: QLC+API|setWidgetValue|<widgetID>|<value>
func (c *QLCPlusClient) SetWidgetValue(ctx context.Context, widgetID string, value string) error {
	if !c.isReachable() {
		return fmt.Errorf("QLC+ not reachable at %s:%d", c.host, c.port)
	}
	if widgetID == "" {
		return fmt.Errorf("widget ID is required")
	}
	// Stub — real implementation sends QLC+API|setWidgetValue|<widgetID>|<value>
	return nil
}

// GetFunctionStatus returns the status of a function.
// QLC+ API: QLC+API|getFunctionStatus|<functionID>
func (c *QLCPlusClient) GetFunctionStatus(ctx context.Context, functionID string) (*QLCPlusFunction, error) {
	if !c.isReachable() {
		return nil, fmt.Errorf("QLC+ not reachable at %s:%d", c.host, c.port)
	}
	if functionID == "" {
		return nil, fmt.Errorf("function ID is required")
	}
	// Stub — real implementation sends QLC+API|getFunctionStatus|<functionID>
	return &QLCPlusFunction{ID: functionID, Status: "unknown"}, nil
}

// SetFunctionStatus starts or stops a function.
// QLC+ API: QLC+API|setFunctionStatus|<functionID>|<1=start|0=stop>
func (c *QLCPlusClient) SetFunctionStatus(ctx context.Context, functionID string, running bool) error {
	if !c.isReachable() {
		return fmt.Errorf("QLC+ not reachable at %s:%d", c.host, c.port)
	}
	if functionID == "" {
		return fmt.Errorf("function ID is required")
	}
	// Stub — real implementation sends QLC+API|setFunctionStatus|<functionID>|<0|1>
	return nil
}

// GetChannelsValues returns DMX channel values for a universe.
// QLC+ API: QLC+API|getChannelsValues|<universe>|<startChannel>|<count>
func (c *QLCPlusClient) GetChannelsValues(ctx context.Context, universe, startChannel, count int) ([]int, error) {
	if !c.isReachable() {
		return nil, fmt.Errorf("QLC+ not reachable at %s:%d", c.host, c.port)
	}
	// Stub — real implementation parses response CSV
	return make([]int, count), nil
}

// SetChannelsValues sets DMX channel values.
// QLC+ API: QLC+API|setChannelsValues|<universe>|<ch1>,<val1>|<ch2>,<val2>...
func (c *QLCPlusClient) SetChannelsValues(ctx context.Context, universe int, channelValues map[int]int) error {
	if !c.isReachable() {
		return fmt.Errorf("QLC+ not reachable at %s:%d", c.host, c.port)
	}
	if len(channelValues) == 0 {
		return fmt.Errorf("at least one channel-value pair is required")
	}
	// Stub — real implementation sends QLC+API|setChannelsValues|<universe>|<pairs>
	return nil
}

// GetHealth returns QLC+ system health.
func (c *QLCPlusClient) GetHealth(ctx context.Context) (*QLCPlusHealth, error) {
	health := &QLCPlusHealth{
		Score:  100,
		Status: "healthy",
	}

	if !c.isReachable() {
		health.Score = 0
		health.Status = "critical"
		health.Issues = append(health.Issues, fmt.Sprintf("QLC+ not reachable at %s:%d", c.host, c.port))
		health.Recommendations = append(health.Recommendations,
			"Start QLC+ and enable the WebSocket server",
			fmt.Sprintf("Verify QLCPLUS_HOST=%s and QLCPLUS_PORT=%d are correct", c.host, c.port),
		)
		return health, nil
	}

	return health, nil
}

// Host returns the configured host.
func (c *QLCPlusClient) Host() string {
	return c.host
}

// Port returns the configured port.
func (c *QLCPlusClient) Port() int {
	return c.port
}

// parseChannelValues parses "ch,val|ch,val" format from QLC+ response.
func parseChannelValues(data string) map[int]int {
	result := make(map[int]int)
	pairs := strings.Split(data, "|")
	for _, pair := range pairs {
		parts := strings.SplitN(pair, ",", 2)
		if len(parts) == 2 {
			ch, err1 := strconv.Atoi(strings.TrimSpace(parts[0]))
			val, err2 := strconv.Atoi(strings.TrimSpace(parts[1]))
			if err1 == nil && err2 == nil {
				result[ch] = val
			}
		}
	}
	return result
}
