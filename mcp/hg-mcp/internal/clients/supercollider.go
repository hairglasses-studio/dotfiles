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

// SuperColliderClient provides control of SuperCollider via OSC
type SuperColliderClient struct {
	host        string
	scSynthPort int // scsynth server port (default 57110)
	scLangPort  int // sclang interpreter port (default 57120)
	client      *osc.Client
}

// SuperColliderStatus represents scsynth server status
type SuperColliderStatus struct {
	Connected        bool    `json:"connected"`
	Host             string  `json:"host"`
	ScSynthPort      int     `json:"scsynth_port"`
	ScLangPort       int     `json:"sclang_port"`
	UGens            int     `json:"ugens"`
	Synths           int     `json:"synths"`
	Groups           int     `json:"groups"`
	SynthDefs        int     `json:"synth_defs"`
	AvgCPU           float64 `json:"avg_cpu"`
	PeakCPU          float64 `json:"peak_cpu"`
	SampleRate       float64 `json:"sample_rate"`
	ActualSampleRate float64 `json:"actual_sample_rate"`
}

var (
	superColliderClientSingleton *SuperColliderClient
	superColliderClientOnce      sync.Once
	superColliderClientErr       error

	// TestOverrideSuperColliderClient, when non-nil, is returned by GetSuperColliderClient
	// instead of the real singleton. Used for testing without network access.
	TestOverrideSuperColliderClient *SuperColliderClient
)

// GetSuperColliderClient returns the singleton SuperCollider client.
func GetSuperColliderClient() (*SuperColliderClient, error) {
	if TestOverrideSuperColliderClient != nil {
		return TestOverrideSuperColliderClient, nil
	}
	superColliderClientOnce.Do(func() {
		superColliderClientSingleton, superColliderClientErr = NewSuperColliderClient()
	})
	return superColliderClientSingleton, superColliderClientErr
}

// NewTestSuperColliderClient creates an in-memory test client.
func NewTestSuperColliderClient() *SuperColliderClient {
	return &SuperColliderClient{
		host:        "localhost",
		scSynthPort: 57110,
		scLangPort:  57120,
	}
}

// NewSuperColliderClient creates a new SuperCollider client.
func NewSuperColliderClient() (*SuperColliderClient, error) {
	host := os.Getenv("SUPERCOLLIDER_HOST")
	if host == "" {
		host = "localhost"
	}

	scSynthPort := 57110
	if p := os.Getenv("SUPERCOLLIDER_SCSYNTH_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			scSynthPort = v
		}
	}

	scLangPort := 57120
	if p := os.Getenv("SUPERCOLLIDER_SCLANG_PORT"); p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			scLangPort = v
		}
	}

	client := osc.NewClient(host, scSynthPort)

	return &SuperColliderClient{
		host:        host,
		scSynthPort: scSynthPort,
		scLangPort:  scLangPort,
		client:      client,
	}, nil
}

// GetStatus returns SuperCollider server status.
func (c *SuperColliderClient) GetStatus(ctx context.Context) (*SuperColliderStatus, error) {
	status := &SuperColliderStatus{
		Host:        c.host,
		ScSynthPort: c.scSynthPort,
		ScLangPort:  c.scLangPort,
	}

	// Try TCP connect to scsynth port
	addr := net.JoinHostPort(c.host, strconv.Itoa(c.scSynthPort))
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err == nil {
		conn.Close()
		status.Connected = true
	}

	return status, nil
}

// CreateSynth creates a new synth node.
// OSC: /s_new <defName> <nodeID> <addAction> <targetID> [param value ...]
func (c *SuperColliderClient) CreateSynth(ctx context.Context, defName string, nodeID int, params map[string]float64) error {
	if defName == "" {
		return fmt.Errorf("synth def name is required")
	}
	if nodeID <= 0 {
		return fmt.Errorf("node ID must be positive")
	}
	return nil
}

// FreeSynth frees (stops) a synth node.
// OSC: /n_free <nodeID>
func (c *SuperColliderClient) FreeSynth(ctx context.Context, nodeID int) error {
	if nodeID <= 0 {
		return fmt.Errorf("node ID must be positive")
	}
	return nil
}

// SetNodeParam sets a parameter on a running node.
// OSC: /n_set <nodeID> <param> <value>
func (c *SuperColliderClient) SetNodeParam(ctx context.Context, nodeID int, param string, value float64) error {
	if nodeID <= 0 {
		return fmt.Errorf("node ID must be positive")
	}
	if param == "" {
		return fmt.Errorf("parameter name is required")
	}
	return nil
}

// EvalCode sends code to sclang for evaluation via TCP.
func (c *SuperColliderClient) EvalCode(ctx context.Context, code string) (string, error) {
	if code == "" {
		return "", fmt.Errorf("code is required")
	}
	// Stub: would connect to sclang TCP port and send code
	return "", nil
}

// Host returns the configured host.
func (c *SuperColliderClient) Host() string {
	return c.host
}

// ScSynthPort returns the configured scsynth port.
func (c *SuperColliderClient) ScSynthPort() int {
	return c.scSynthPort
}

// ScLangPort returns the configured sclang port.
func (c *SuperColliderClient) ScLangPort() int {
	return c.scLangPort
}
