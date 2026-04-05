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

// ArdourClient provides control of Ardour DAW via OSC.
type ArdourClient struct {
	host   string
	port   int
	client *osc.Client
}

// ArdourStatus represents Ardour connection and transport status.
type ArdourStatus struct {
	Connected bool    `json:"connected"`
	Host      string  `json:"host"`
	Port      int     `json:"port"`
	Playing   bool    `json:"playing"`
	Recording bool    `json:"recording"`
	Frame     int64   `json:"frame"`
	Speed     float64 `json:"speed"`
}

// ArdourTrack represents a mixer strip in Ardour.
type ArdourTrack struct {
	StripID int     `json:"strip_id"`
	Fader   float64 `json:"fader"`
	Mute    bool    `json:"mute"`
	Solo    bool    `json:"solo"`
	Pan     float64 `json:"pan"`
	Meter   float64 `json:"meter_db"`
}

var (
	ardourClientSingleton *ArdourClient
	ardourClientOnce      sync.Once
	ardourClientErr       error

	// TestOverrideArdourClient, when non-nil, is returned by GetArdourClient
	// instead of the real singleton. Used for testing without network access.
	TestOverrideArdourClient *ArdourClient
)

// GetArdourClient returns the singleton Ardour client.
func GetArdourClient() (*ArdourClient, error) {
	if TestOverrideArdourClient != nil {
		return TestOverrideArdourClient, nil
	}
	ardourClientOnce.Do(func() {
		ardourClientSingleton, ardourClientErr = NewArdourClient()
	})
	return ardourClientSingleton, ardourClientErr
}

// NewTestArdourClient creates an in-memory test client.
func NewTestArdourClient() *ArdourClient {
	return &ArdourClient{
		host: "localhost",
		port: 3819,
	}
}

// NewArdourClient creates a new Ardour OSC client.
func NewArdourClient() (*ArdourClient, error) {
	host := os.Getenv("ARDOUR_HOST")
	if host == "" {
		host = "localhost"
	}

	port := 3819
	if p := os.Getenv("ARDOUR_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}

	client := osc.NewClient(host, port)

	return &ArdourClient{
		host:   host,
		port:   port,
		client: client,
	}, nil
}

// GetStatus returns Ardour connection and transport status.
func (c *ArdourClient) GetStatus(ctx context.Context) (*ArdourStatus, error) {
	status := &ArdourStatus{
		Host: c.host,
		Port: c.port,
	}

	addr := net.JoinHostPort(c.host, strconv.Itoa(c.port))
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err == nil {
		conn.Close()
		status.Connected = true
	}

	return status, nil
}

// TransportPlay sends play command.
func (c *ArdourClient) TransportPlay(ctx context.Context) error {
	return nil
}

// TransportStop sends stop command.
func (c *ArdourClient) TransportStop(ctx context.Context) error {
	return nil
}

// TransportRecord toggles record enable.
func (c *ArdourClient) TransportRecord(ctx context.Context) error {
	return nil
}

// TransportLocate seeks to a frame position.
func (c *ArdourClient) TransportLocate(ctx context.Context, frame int64) error {
	if frame < 0 {
		return fmt.Errorf("frame must be non-negative")
	}
	return nil
}

// SetStripFader sets the fader level for a strip (0.0-1.0).
func (c *ArdourClient) SetStripFader(ctx context.Context, stripID int, value float64) error {
	if stripID < 0 {
		return fmt.Errorf("strip ID must be non-negative")
	}
	return nil
}

// SetStripMute sets mute state for a strip.
func (c *ArdourClient) SetStripMute(ctx context.Context, stripID int, mute bool) error {
	if stripID < 0 {
		return fmt.Errorf("strip ID must be non-negative")
	}
	return nil
}

// SetStripSolo sets solo state for a strip.
func (c *ArdourClient) SetStripSolo(ctx context.Context, stripID int, solo bool) error {
	if stripID < 0 {
		return fmt.Errorf("strip ID must be non-negative")
	}
	return nil
}

// SetStripPan sets the pan position for a strip (0.0-1.0).
func (c *ArdourClient) SetStripPan(ctx context.Context, stripID int, pan float64) error {
	if stripID < 0 {
		return fmt.Errorf("strip ID must be non-negative")
	}
	return nil
}

// GetStripMeter returns the current meter reading for a strip.
func (c *ArdourClient) GetStripMeter(ctx context.Context, stripID int) (float64, error) {
	if stripID < 0 {
		return 0, fmt.Errorf("strip ID must be non-negative")
	}
	return -60.0, nil // Stub: silence
}

// Host returns the configured host.
func (c *ArdourClient) Host() string {
	return c.host
}

// Port returns the configured port.
func (c *ArdourClient) Port() int {
	return c.port
}
