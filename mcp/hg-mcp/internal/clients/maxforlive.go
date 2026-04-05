// Package clients provides API clients for external services.
package clients

import (
	"context"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/hypebeast/go-osc/osc"
)

// MaxForLiveClient provides control of Max for Live devices via OSC
type MaxForLiveClient struct {
	host      string
	port      int
	replyPort int
	client    *osc.Client
	mu        sync.RWMutex
	devices   map[string]*M4LDevice
	connected bool
}

// M4LDevice represents a Max for Live device
type M4LDevice struct {
	ID         string                   `json:"id"`
	Name       string                   `json:"name"`
	TrackName  string                   `json:"track_name,omitempty"`
	TrackIndex int                      `json:"track_index,omitempty"`
	Parameters map[string]*M4LParameter `json:"parameters,omitempty"`
	LastSeen   time.Time                `json:"last_seen"`
}

// M4LParameter represents a parameter on an M4L device
type M4LParameter struct {
	Name    string   `json:"name"`
	Value   float64  `json:"value"`
	Min     float64  `json:"min"`
	Max     float64  `json:"max"`
	Type    string   `json:"type,omitempty"`    // float, int, enum
	Options []string `json:"options,omitempty"` // for enum type
}

// M4LMapping represents a parameter mapping
type M4LMapping struct {
	SourceDevice string  `json:"source_device"`
	SourceParam  string  `json:"source_param"`
	TargetDevice string  `json:"target_device"`
	TargetParam  string  `json:"target_param"`
	Scale        float64 `json:"scale"`
	Offset       float64 `json:"offset"`
}

// M4LMacro represents a macro/preset
type M4LMacro struct {
	ID          string             `json:"id"`
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Parameters  map[string]float64 `json:"parameters"`
}

// M4LStatus represents connection status
type M4LStatus struct {
	Connected   bool         `json:"connected"`
	Host        string       `json:"host"`
	Port        int          `json:"port"`
	ReplyPort   int          `json:"reply_port"`
	DeviceCount int          `json:"device_count"`
	Devices     []*M4LDevice `json:"devices,omitempty"`
}

// M4LHealth represents health status
type M4LHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Connected       bool     `json:"connected"`
	DevicesFound    int      `json:"devices_found"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// NewMaxForLiveClient creates a new Max for Live client
func NewMaxForLiveClient() (*MaxForLiveClient, error) {
	host := os.Getenv("M4L_OSC_HOST")
	if host == "" {
		host = "localhost"
	}

	port := 11002
	if p := os.Getenv("M4L_OSC_PORT"); p != "" {
		fmt.Sscanf(p, "%d", &port)
	}

	replyPort := 11003
	if p := os.Getenv("M4L_OSC_REPLY_PORT"); p != "" {
		fmt.Sscanf(p, "%d", &replyPort)
	}

	client := osc.NewClient(host, port)

	return &MaxForLiveClient{
		host:      host,
		port:      port,
		replyPort: replyPort,
		client:    client,
		devices:   make(map[string]*M4LDevice),
	}, nil
}

// GetStatus returns connection status
func (c *MaxForLiveClient) GetStatus(ctx context.Context) (*M4LStatus, error) {
	status := &M4LStatus{
		Host:      c.host,
		Port:      c.port,
		ReplyPort: c.replyPort,
	}

	// Check connection
	conn, err := net.DialTimeout("udp", fmt.Sprintf("%s:%d", c.host, c.port), 2*time.Second)
	if err != nil {
		status.Connected = false
		return status, nil
	}
	conn.Close()
	status.Connected = true

	// Request device list
	c.sendMessage("/m4l/devices/list")

	c.mu.RLock()
	status.DeviceCount = len(c.devices)
	for _, dev := range c.devices {
		status.Devices = append(status.Devices, dev)
	}
	c.mu.RUnlock()

	return status, nil
}

// sendMessage sends an OSC message
func (c *MaxForLiveClient) sendMessage(address string, args ...interface{}) error {
	msg := osc.NewMessage(address)
	for _, arg := range args {
		msg.Append(arg)
	}
	return c.client.Send(msg)
}

// GetDevices returns all known M4L devices
func (c *MaxForLiveClient) GetDevices(ctx context.Context) ([]*M4LDevice, error) {
	// Request device list from M4L bridge
	c.sendMessage("/m4l/devices/list")
	time.Sleep(100 * time.Millisecond)

	c.mu.RLock()
	defer c.mu.RUnlock()

	var devices []*M4LDevice
	for _, dev := range c.devices {
		devices = append(devices, dev)
	}

	return devices, nil
}

// GetDevice returns a specific device by ID
func (c *MaxForLiveClient) GetDevice(ctx context.Context, deviceID string) (*M4LDevice, error) {
	c.mu.RLock()
	dev, ok := c.devices[deviceID]
	c.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("device not found: %s", deviceID)
	}

	// Request fresh parameter values
	c.sendMessage("/m4l/device/get", deviceID)
	time.Sleep(50 * time.Millisecond)

	return dev, nil
}

// GetParameters returns parameters for a device
func (c *MaxForLiveClient) GetParameters(ctx context.Context, deviceID string) (map[string]*M4LParameter, error) {
	c.sendMessage("/m4l/device/parameters", deviceID)
	time.Sleep(50 * time.Millisecond)

	c.mu.RLock()
	defer c.mu.RUnlock()

	dev, ok := c.devices[deviceID]
	if !ok {
		return nil, fmt.Errorf("device not found: %s", deviceID)
	}

	return dev.Parameters, nil
}

// SetParameter sets a parameter value
func (c *MaxForLiveClient) SetParameter(ctx context.Context, deviceID, paramName string, value float64) error {
	return c.sendMessage("/m4l/device/set", deviceID, paramName, float32(value))
}

// SetParameterNormalized sets a parameter using normalized 0-1 value
func (c *MaxForLiveClient) SetParameterNormalized(ctx context.Context, deviceID, paramName string, value float64) error {
	if value < 0 || value > 1 {
		return fmt.Errorf("normalized value must be between 0 and 1")
	}
	return c.sendMessage("/m4l/device/set/normalized", deviceID, paramName, float32(value))
}

// SendCustomMessage sends a custom OSC message to an M4L device
func (c *MaxForLiveClient) SendCustomMessage(ctx context.Context, address string, args ...interface{}) error {
	return c.sendMessage(address, args...)
}

// GetMappings returns all parameter mappings
func (c *MaxForLiveClient) GetMappings(ctx context.Context) ([]M4LMapping, error) {
	c.sendMessage("/m4l/mappings/list")
	time.Sleep(50 * time.Millisecond)

	// Would need response handling to populate this
	var mappings []M4LMapping
	return mappings, nil
}

// CreateMapping creates a parameter mapping
func (c *MaxForLiveClient) CreateMapping(ctx context.Context, mapping *M4LMapping) error {
	return c.sendMessage("/m4l/mapping/create",
		mapping.SourceDevice, mapping.SourceParam,
		mapping.TargetDevice, mapping.TargetParam,
		float32(mapping.Scale), float32(mapping.Offset))
}

// DeleteMapping deletes a parameter mapping
func (c *MaxForLiveClient) DeleteMapping(ctx context.Context, sourceDevice, sourceParam string) error {
	return c.sendMessage("/m4l/mapping/delete", sourceDevice, sourceParam)
}

// GetMacros returns available macros
func (c *MaxForLiveClient) GetMacros(ctx context.Context, deviceID string) ([]M4LMacro, error) {
	c.sendMessage("/m4l/macros/list", deviceID)
	time.Sleep(50 * time.Millisecond)

	var macros []M4LMacro
	return macros, nil
}

// TriggerMacro triggers a macro/preset
func (c *MaxForLiveClient) TriggerMacro(ctx context.Context, deviceID, macroID string) error {
	return c.sendMessage("/m4l/macro/trigger", deviceID, macroID)
}

// StoreMacro saves current parameter values as a macro
func (c *MaxForLiveClient) StoreMacro(ctx context.Context, deviceID, macroName string) error {
	return c.sendMessage("/m4l/macro/store", deviceID, macroName)
}

// Recall recalls a stored macro
func (c *MaxForLiveClient) RecallMacro(ctx context.Context, deviceID, macroID string) error {
	return c.sendMessage("/m4l/macro/recall", deviceID, macroID)
}

// Bang sends a bang to an M4L device
func (c *MaxForLiveClient) Bang(ctx context.Context, deviceID, inlet string) error {
	return c.sendMessage("/m4l/bang", deviceID, inlet)
}

// SendList sends a list to an M4L device
func (c *MaxForLiveClient) SendList(ctx context.Context, deviceID, inlet string, values []interface{}) error {
	args := []interface{}{deviceID, inlet}
	args = append(args, values...)
	return c.sendMessage("/m4l/list", args...)
}

// RegisterDevice registers an M4L device (called by the M4L device on load)
func (c *MaxForLiveClient) RegisterDevice(ctx context.Context, deviceID, deviceName string, trackIndex int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.devices[deviceID] = &M4LDevice{
		ID:         deviceID,
		Name:       deviceName,
		TrackIndex: trackIndex,
		Parameters: make(map[string]*M4LParameter),
		LastSeen:   time.Now(),
	}

	return nil
}

// UnregisterDevice removes an M4L device
func (c *MaxForLiveClient) UnregisterDevice(ctx context.Context, deviceID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.devices, deviceID)
	return nil
}

// GetHealth returns health status
func (c *MaxForLiveClient) GetHealth(ctx context.Context) (*M4LHealth, error) {
	health := &M4LHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check UDP connectivity
	conn, err := net.DialTimeout("udp", fmt.Sprintf("%s:%d", c.host, c.port), 2*time.Second)
	if err != nil {
		health.Connected = false
		health.Score -= 50
		health.Issues = append(health.Issues, fmt.Sprintf("Cannot connect to %s:%d", c.host, c.port))
		health.Recommendations = append(health.Recommendations,
			"Ensure Ableton Live is running with M4L bridge device")
	} else {
		conn.Close()
		health.Connected = true

		// Check for devices
		c.sendMessage("/m4l/devices/list")
		time.Sleep(100 * time.Millisecond)

		c.mu.RLock()
		health.DevicesFound = len(c.devices)
		c.mu.RUnlock()

		if health.DevicesFound == 0 {
			health.Score -= 30
			health.Issues = append(health.Issues, "No M4L bridge devices found")
			health.Recommendations = append(health.Recommendations,
				"Add hg-mcp-bridge.amxd to your Live set")
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

// Close closes the client
func (c *MaxForLiveClient) Close() error {
	return nil
}
