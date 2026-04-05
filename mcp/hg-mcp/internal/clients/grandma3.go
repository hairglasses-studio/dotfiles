// Package clients provides API clients for external services.
package clients

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/hypebeast/go-osc/osc"
)

// GrandMA3Client provides access to grandMA3 lighting console via OSC
type GrandMA3Client struct {
	host      string
	oscPort   int
	oscClient *osc.Client
}

// GrandMA3Status represents console connection status
type GrandMA3Status struct {
	Connected bool   `json:"connected"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	Protocol  string `json:"protocol"`
}

// GrandMA3Executor represents an executor (fader/button)
type GrandMA3Executor struct {
	Page      int     `json:"page"`
	Number    int     `json:"number"`
	Name      string  `json:"name,omitempty"`
	FaderVal  float32 `json:"fader_value,omitempty"`
	IsRunning bool    `json:"is_running,omitempty"`
}

// GrandMA3Sequence represents a sequence
type GrandMA3Sequence struct {
	Number     int    `json:"number"`
	Name       string `json:"name"`
	CueCount   int    `json:"cue_count,omitempty"`
	CurrentCue string `json:"current_cue,omitempty"`
}

// GrandMA3Cue represents a cue in a sequence
type GrandMA3Cue struct {
	Number  float64 `json:"number"`
	Name    string  `json:"name,omitempty"`
	Trigger string  `json:"trigger,omitempty"`
}

// GrandMA3Master represents a grand master or speed master
type GrandMA3Master struct {
	Type  string  `json:"type"`
	Name  string  `json:"name"`
	Value float32 `json:"value"`
}

// GrandMA3Health represents console health status
type GrandMA3Health struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Connected       bool     `json:"connected"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// NewGrandMA3Client creates a new grandMA3 client
func NewGrandMA3Client() (*GrandMA3Client, error) {
	host := os.Getenv("GRANDMA3_HOST")
	if host == "" {
		host = "localhost"
	}

	oscPort := 8000 // Default grandMA3 OSC port
	if portStr := os.Getenv("GRANDMA3_OSC_PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			oscPort = p
		}
	}

	oscClient := osc.NewClient(host, oscPort)

	return &GrandMA3Client{
		host:      host,
		oscPort:   oscPort,
		oscClient: oscClient,
	}, nil
}

// Host returns the configured host
func (c *GrandMA3Client) Host() string {
	return c.host
}

// Port returns the configured OSC port
func (c *GrandMA3Client) Port() int {
	return c.oscPort
}

// sendOSC sends an OSC message to grandMA3
func (c *GrandMA3Client) sendOSC(address string, args ...interface{}) error {
	msg := osc.NewMessage(address)
	for _, arg := range args {
		msg.Append(arg)
	}
	return c.oscClient.Send(msg)
}

// GetStatus returns console connection status
func (c *GrandMA3Client) GetStatus(ctx context.Context) (*GrandMA3Status, error) {
	status := &GrandMA3Status{
		Connected: false,
		Host:      c.host,
		Port:      c.oscPort,
		Protocol:  "OSC/UDP",
	}

	// Try to connect to OSC port
	addr := net.JoinHostPort(c.host, strconv.Itoa(c.oscPort))
	conn, err := net.DialTimeout("udp", addr, 2*time.Second)
	if err == nil {
		conn.Close()
		status.Connected = true
	}

	return status, nil
}

// SendCommand sends a command line instruction to grandMA3
func (c *GrandMA3Client) SendCommand(ctx context.Context, command string) error {
	return c.sendOSC("/cmd", command)
}

// GoExecutor triggers Go on an executor
func (c *GrandMA3Client) GoExecutor(ctx context.Context, page, executor int) error {
	cmd := fmt.Sprintf("Go+ Executor %d.%d", page, executor)
	return c.sendOSC("/cmd", cmd)
}

// StopExecutor stops an executor
func (c *GrandMA3Client) StopExecutor(ctx context.Context, page, executor int) error {
	cmd := fmt.Sprintf("Off Executor %d.%d", page, executor)
	return c.sendOSC("/cmd", cmd)
}

// SetExecutorFader sets fader value for an executor (0-100)
func (c *GrandMA3Client) SetExecutorFader(ctx context.Context, page, executor int, value float32) error {
	// Clamp value to 0-100
	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}
	cmd := fmt.Sprintf("Executor %d.%d At %.1f", page, executor, value)
	return c.sendOSC("/cmd", cmd)
}

// FlashExecutor triggers flash on an executor
func (c *GrandMA3Client) FlashExecutor(ctx context.Context, page, executor int, on bool) error {
	address := fmt.Sprintf("/Page%d/Key%d", page, executor)
	val := int32(0)
	if on {
		val = 1
	}
	return c.sendOSC(address, val)
}

// GoCue triggers a specific cue in a sequence
func (c *GrandMA3Client) GoCue(ctx context.Context, sequence int, cue float64) error {
	cmd := fmt.Sprintf("Go+ Cue %g Sequence %d", cue, sequence)
	return c.sendOSC("/cmd", cmd)
}

// GoNextCue triggers the next cue in a sequence
func (c *GrandMA3Client) GoNextCue(ctx context.Context, sequence int) error {
	cmd := fmt.Sprintf("Go+ Sequence %d", sequence)
	return c.sendOSC("/cmd", cmd)
}

// GoPrevCue triggers the previous cue in a sequence
func (c *GrandMA3Client) GoPrevCue(ctx context.Context, sequence int) error {
	cmd := fmt.Sprintf("Go- Sequence %d", sequence)
	return c.sendOSC("/cmd", cmd)
}

// PauseSequence pauses a sequence
func (c *GrandMA3Client) PauseSequence(ctx context.Context, sequence int) error {
	cmd := fmt.Sprintf("Pause Sequence %d", sequence)
	return c.sendOSC("/cmd", cmd)
}

// StopSequence stops a sequence
func (c *GrandMA3Client) StopSequence(ctx context.Context, sequence int) error {
	cmd := fmt.Sprintf("Off Sequence %d", sequence)
	return c.sendOSC("/cmd", cmd)
}

// SetGrandMaster sets the grand master value (0-100)
func (c *GrandMA3Client) SetGrandMaster(ctx context.Context, value float32) error {
	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}
	cmd := fmt.Sprintf("Master 1 At %.1f", value)
	return c.sendOSC("/cmd", cmd)
}

// SetSpeedMaster sets a speed master value (0-100)
func (c *GrandMA3Client) SetSpeedMaster(ctx context.Context, master int, value float32) error {
	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}
	cmd := fmt.Sprintf("SpeedMaster %d At %.1f", master, value)
	return c.sendOSC("/cmd", cmd)
}

// SetBPMMaster sets BPM for a speed master
func (c *GrandMA3Client) SetBPMMaster(ctx context.Context, master int, bpm float32) error {
	cmd := fmt.Sprintf("SpeedMaster %d Rate %.2f BPM", master, bpm)
	return c.sendOSC("/cmd", cmd)
}

// TapTempo taps tempo for a speed master
func (c *GrandMA3Client) TapTempo(ctx context.Context, master int) error {
	cmd := fmt.Sprintf("Learn SpeedMaster %d", master)
	return c.sendOSC("/cmd", cmd)
}

// ClearProgrammer clears the programmer
func (c *GrandMA3Client) ClearProgrammer(ctx context.Context) error {
	return c.sendOSC("/cmd", "Clear")
}

// ClearAll clears all running sequences and programmer
func (c *GrandMA3Client) ClearAll(ctx context.Context) error {
	return c.sendOSC("/cmd", "ClearAll")
}

// Blackout triggers blackout
func (c *GrandMA3Client) Blackout(ctx context.Context, on bool) error {
	if on {
		return c.sendOSC("/cmd", "Blackout On")
	}
	return c.sendOSC("/cmd", "Blackout Off")
}

// SelectFixtures selects fixtures by number range
func (c *GrandMA3Client) SelectFixtures(ctx context.Context, fixtureSpec string) error {
	cmd := fmt.Sprintf("Fixture %s", fixtureSpec)
	return c.sendOSC("/cmd", cmd)
}

// SetFixtureAttribute sets an attribute for selected fixtures
func (c *GrandMA3Client) SetFixtureAttribute(ctx context.Context, attribute string, value float32) error {
	cmd := fmt.Sprintf("Attribute \"%s\" At %.1f", attribute, value)
	return c.sendOSC("/cmd", cmd)
}

// SetDimmer sets dimmer for selected fixtures (0-100)
func (c *GrandMA3Client) SetDimmer(ctx context.Context, value float32) error {
	if value < 0 {
		value = 0
	}
	if value > 100 {
		value = 100
	}
	return c.sendOSC("/cmd", fmt.Sprintf("At %.1f", value))
}

// StorePreset stores current programmer to a preset
func (c *GrandMA3Client) StorePreset(ctx context.Context, presetType string, number int) error {
	cmd := fmt.Sprintf("Store Preset %d.%s", number, presetType)
	return c.sendOSC("/cmd", cmd)
}

// CallPreset recalls a preset
func (c *GrandMA3Client) CallPreset(ctx context.Context, presetType string, number int) error {
	cmd := fmt.Sprintf("At Preset %d.%s", number, presetType)
	return c.sendOSC("/cmd", cmd)
}

// CallMacro triggers a macro
func (c *GrandMA3Client) CallMacro(ctx context.Context, macro int) error {
	cmd := fmt.Sprintf("Go+ Macro %d", macro)
	return c.sendOSC("/cmd", cmd)
}

// SetTimecodeEnabled enables or disables timecode
func (c *GrandMA3Client) SetTimecodeEnabled(ctx context.Context, slot int, enabled bool) error {
	if enabled {
		return c.sendOSC("/cmd", fmt.Sprintf("Go+ Timecode %d", slot))
	}
	return c.sendOSC("/cmd", fmt.Sprintf("Off Timecode %d", slot))
}

// JumpTimecode jumps to a specific timecode position
func (c *GrandMA3Client) JumpTimecode(ctx context.Context, slot int, hours, minutes, seconds, frames int) error {
	tc := fmt.Sprintf("%02d:%02d:%02d.%02d", hours, minutes, seconds, frames)
	cmd := fmt.Sprintf("Goto Timecode %d Time %s", slot, tc)
	return c.sendOSC("/cmd", cmd)
}

// GetHealth returns console health status
func (c *GrandMA3Client) GetHealth(ctx context.Context) (*GrandMA3Health, error) {
	health := &GrandMA3Health{
		Score:  100,
		Status: "healthy",
	}

	status, _ := c.GetStatus(ctx)
	health.Connected = status.Connected

	if !status.Connected {
		health.Score -= 50
		health.Issues = append(health.Issues, "Cannot reach grandMA3 console via OSC")
		health.Recommendations = append(health.Recommendations, "Check console IP and enable OSC in In & Out menu")
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
