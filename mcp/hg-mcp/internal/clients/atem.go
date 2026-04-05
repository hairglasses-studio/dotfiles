// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

// ATEMClient provides access to Blackmagic ATEM switchers
type ATEMClient struct {
	host      string
	port      string
	conn      *net.UDPConn
	sessionID uint16
	localSeq  uint16
	remoteSeq uint16
	connected bool
	mu        sync.RWMutex
	state     *ATEMState
}

// ATEMState represents the current switcher state
type ATEMState struct {
	Model           string            `json:"model"`
	ProtocolVersion string            `json:"protocol_version"`
	ProgramInput    int               `json:"program_input"`
	PreviewInput    int               `json:"preview_input"`
	TransitionPos   float64           `json:"transition_position"`
	TransitionStyle string            `json:"transition_style"`
	TransitionRate  int               `json:"transition_rate"`
	InTransition    bool              `json:"in_transition"`
	FadeToBlack     bool              `json:"fade_to_black"`
	Inputs          map[int]ATEMInput `json:"inputs"`
	AudioLevels     map[int]float64   `json:"audio_levels,omitempty"`
}

// ATEMInput represents an input source
type ATEMInput struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	ShortName string `json:"short_name"`
	Type      string `json:"type"`
	Available bool   `json:"available"`
}

// ATEMStatus represents connection status
type ATEMStatus struct {
	Connected       bool   `json:"connected"`
	Host            string `json:"host"`
	Model           string `json:"model"`
	ProtocolVersion string `json:"protocol_version"`
	ProgramInput    int    `json:"program_input"`
	PreviewInput    int    `json:"preview_input"`
}

// ATEMTransition represents transition settings
type ATEMTransition struct {
	Style      string  `json:"style"`
	Rate       int     `json:"rate_frames"`
	Position   float64 `json:"position"`
	InProgress bool    `json:"in_progress"`
}

// ATEMHealth represents health status
type ATEMHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Connected       bool     `json:"connected"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// NewATEMClient creates a new ATEM client
func NewATEMClient() (*ATEMClient, error) {
	host := os.Getenv("ATEM_HOST")
	if host == "" {
		host = "192.168.1.240" // Common default for ATEM Mini
	}

	port := os.Getenv("ATEM_PORT")
	if port == "" {
		port = "9910"
	}

	return &ATEMClient{
		host: host,
		port: port,
		state: &ATEMState{
			Inputs: make(map[int]ATEMInput),
		},
	}, nil
}

// Connect establishes connection to the ATEM switcher
func (c *ATEMClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%s", c.host, c.port))
	if err != nil {
		return fmt.Errorf("failed to resolve ATEM address: %w", err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return fmt.Errorf("failed to connect to ATEM: %w", err)
	}

	c.conn = conn
	c.sessionID = 0x1234 // Will be assigned by switcher
	c.localSeq = 0

	// Send connection request
	if err := c.sendConnectRequest(); err != nil {
		conn.Close()
		return fmt.Errorf("failed to send connect request: %w", err)
	}

	// Wait for response with timeout
	c.conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	buf := make([]byte, 2048)
	n, err := c.conn.Read(buf)
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to receive connect response: %w", err)
	}

	if n < 12 {
		conn.Close()
		return fmt.Errorf("invalid response from ATEM")
	}

	// Parse session ID from response
	c.sessionID = binary.BigEndian.Uint16(buf[2:4])
	c.connected = true

	// Start state receiver goroutine
	go c.receiveLoop(ctx)

	return nil
}

// sendConnectRequest sends the initial connection packet
func (c *ATEMClient) sendConnectRequest() error {
	// ATEM connection request packet
	packet := make([]byte, 20)
	packet[0] = 0x10                                  // Flags: connect request
	packet[1] = 0x14                                  // Length: 20 bytes
	binary.BigEndian.PutUint16(packet[2:4], 0x0000)   // Session ID (0 for new connection)
	binary.BigEndian.PutUint16(packet[10:12], 0x0001) // Connect request type

	_, err := c.conn.Write(packet)
	return err
}

// receiveLoop handles incoming packets from the ATEM
func (c *ATEMClient) receiveLoop(ctx context.Context) {
	buf := make([]byte, 2048)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			c.conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			n, err := c.conn.Read(buf)
			if err != nil {
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					continue
				}
				c.mu.Lock()
				c.connected = false
				c.mu.Unlock()
				return
			}

			if n >= 12 {
				c.handlePacket(buf[:n])
			}
		}
	}
}

// handlePacket processes an incoming ATEM packet
func (c *ATEMClient) handlePacket(data []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	flags := data[0] >> 3

	// Send ACK if required
	if flags&0x01 != 0 {
		c.sendAck(binary.BigEndian.Uint16(data[4:6]))
	}

	// Parse commands in packet
	if len(data) > 12 {
		c.parseCommands(data[12:])
	}
}

// sendAck sends an acknowledgment packet
func (c *ATEMClient) sendAck(remoteSeq uint16) {
	packet := make([]byte, 12)
	packet[0] = 0x80 // ACK flag
	packet[1] = 0x0c // Length: 12 bytes
	binary.BigEndian.PutUint16(packet[2:4], c.sessionID)
	binary.BigEndian.PutUint16(packet[4:6], remoteSeq)
	c.conn.Write(packet)
}

// parseCommands parses ATEM command data
func (c *ATEMClient) parseCommands(data []byte) {
	offset := 0
	for offset < len(data)-4 {
		cmdLen := int(binary.BigEndian.Uint16(data[offset : offset+2]))
		if cmdLen < 8 || offset+cmdLen > len(data) {
			break
		}

		cmdName := string(data[offset+4 : offset+8])
		cmdData := data[offset+8 : offset+cmdLen]

		switch cmdName {
		case "_ver":
			if len(cmdData) >= 4 {
				major := binary.BigEndian.Uint16(cmdData[0:2])
				minor := binary.BigEndian.Uint16(cmdData[2:4])
				c.state.ProtocolVersion = fmt.Sprintf("%d.%d", major, minor)
			}
		case "_pin":
			// Product ID / Model name
			if len(cmdData) >= 44 {
				c.state.Model = string(cmdData[:44])
			}
		case "PrgI":
			if len(cmdData) >= 2 {
				c.state.ProgramInput = int(binary.BigEndian.Uint16(cmdData[0:2]))
			}
		case "PrvI":
			if len(cmdData) >= 2 {
				c.state.PreviewInput = int(binary.BigEndian.Uint16(cmdData[0:2]))
			}
		case "TrSS":
			if len(cmdData) >= 4 {
				c.state.TransitionStyle = transitionStyleName(cmdData[1])
			}
		case "TrPs":
			if len(cmdData) >= 4 {
				c.state.InTransition = cmdData[0] != 0
				c.state.TransitionPos = float64(binary.BigEndian.Uint16(cmdData[2:4])) / 10000.0
			}
		case "InPr":
			if len(cmdData) >= 32 {
				inputID := int(binary.BigEndian.Uint16(cmdData[0:2]))
				input := ATEMInput{
					ID:        inputID,
					Name:      trimNull(string(cmdData[2:22])),
					ShortName: trimNull(string(cmdData[22:26])),
					Available: cmdData[26] != 0,
				}
				c.state.Inputs[inputID] = input
			}
		}

		offset += cmdLen
	}
}

// transitionStyleName returns the name of a transition style
func transitionStyleName(style byte) string {
	switch style {
	case 0:
		return "mix"
	case 1:
		return "dip"
	case 2:
		return "wipe"
	case 3:
		return "dve"
	case 4:
		return "sting"
	default:
		return "unknown"
	}
}

// trimNull removes null bytes from a string
func trimNull(s string) string {
	for i, c := range s {
		if c == 0 {
			return s[:i]
		}
	}
	return s
}

// Close closes the connection
func (c *ATEMClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.connected = false
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// sendCommand sends a command to the ATEM
func (c *ATEMClient) sendCommand(cmdName string, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return fmt.Errorf("not connected to ATEM")
	}

	cmdLen := 8 + len(data)
	packet := make([]byte, 12+cmdLen)

	// Header
	packet[0] = 0x08 // Data flag
	binary.BigEndian.PutUint16(packet[0:2], uint16(12+cmdLen)|0x0800)
	binary.BigEndian.PutUint16(packet[2:4], c.sessionID)
	binary.BigEndian.PutUint16(packet[10:12], c.localSeq)
	c.localSeq++

	// Command
	binary.BigEndian.PutUint16(packet[12:14], uint16(cmdLen))
	copy(packet[16:20], cmdName)
	copy(packet[20:], data)

	_, err := c.conn.Write(packet)
	return err
}

// GetStatus returns the current switcher status
func (c *ATEMClient) GetStatus(ctx context.Context) (*ATEMStatus, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return &ATEMStatus{
		Connected:       c.connected,
		Host:            fmt.Sprintf("%s:%s", c.host, c.port),
		Model:           c.state.Model,
		ProtocolVersion: c.state.ProtocolVersion,
		ProgramInput:    c.state.ProgramInput,
		PreviewInput:    c.state.PreviewInput,
	}, nil
}

// GetInputs returns all available inputs
func (c *ATEMClient) GetInputs(ctx context.Context) ([]ATEMInput, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	inputs := make([]ATEMInput, 0, len(c.state.Inputs))
	for _, input := range c.state.Inputs {
		inputs = append(inputs, input)
	}
	return inputs, nil
}

// SetProgram sets the program output
func (c *ATEMClient) SetProgram(ctx context.Context, inputID int) error {
	data := make([]byte, 4)
	data[0] = 0 // ME index
	binary.BigEndian.PutUint16(data[2:4], uint16(inputID))
	return c.sendCommand("CPgI", data)
}

// SetPreview sets the preview output
func (c *ATEMClient) SetPreview(ctx context.Context, inputID int) error {
	data := make([]byte, 4)
	data[0] = 0 // ME index
	binary.BigEndian.PutUint16(data[2:4], uint16(inputID))
	return c.sendCommand("CPvI", data)
}

// Cut performs a cut transition
func (c *ATEMClient) Cut(ctx context.Context) error {
	data := make([]byte, 4)
	data[0] = 0 // ME index
	return c.sendCommand("DCut", data)
}

// Auto performs an auto transition
func (c *ATEMClient) Auto(ctx context.Context) error {
	data := make([]byte, 4)
	data[0] = 0 // ME index
	return c.sendCommand("DAut", data)
}

// SetTransition configures the transition type and rate
func (c *ATEMClient) SetTransition(ctx context.Context, style string, rateFrames int) error {
	var styleID byte
	switch style {
	case "mix":
		styleID = 0
	case "dip":
		styleID = 1
	case "wipe":
		styleID = 2
	case "dve":
		styleID = 3
	case "sting":
		styleID = 4
	default:
		return fmt.Errorf("unknown transition style: %s", style)
	}

	// Set transition style
	styleData := make([]byte, 4)
	styleData[0] = 0x01 // Mask: style
	styleData[1] = 0    // ME index
	styleData[2] = styleID
	if err := c.sendCommand("CTSt", styleData); err != nil {
		return err
	}

	// Set transition rate based on style
	rateCmd := "CTMx" // Default to mix rate
	switch style {
	case "dip":
		rateCmd = "CTDp"
	case "wipe":
		rateCmd = "CTWp"
	case "dve":
		rateCmd = "CTDV"
	}

	rateData := make([]byte, 20)
	rateData[0] = 0x01 // Mask: rate
	rateData[1] = 0    // ME index
	rateData[2] = byte(rateFrames)

	return c.sendCommand(rateCmd, rateData)
}

// GetTransition returns current transition settings
func (c *ATEMClient) GetTransition(ctx context.Context) (*ATEMTransition, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return &ATEMTransition{
		Style:      c.state.TransitionStyle,
		Rate:       c.state.TransitionRate,
		Position:   c.state.TransitionPos,
		InProgress: c.state.InTransition,
	}, nil
}

// FadeToBlack toggles fade to black
func (c *ATEMClient) FadeToBlack(ctx context.Context) error {
	data := make([]byte, 4)
	data[0] = 0 // ME index
	return c.sendCommand("FtbA", data)
}

// GetHealth returns health status
func (c *ATEMClient) GetHealth(ctx context.Context) (*ATEMHealth, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	health := &ATEMHealth{
		Score:     100,
		Status:    "healthy",
		Connected: c.connected,
	}

	if !c.connected {
		health.Score = 0
		health.Status = "critical"
		health.Issues = append(health.Issues, "Not connected to ATEM switcher")
		health.Recommendations = append(health.Recommendations,
			fmt.Sprintf("Check network connectivity to %s:%s", c.host, c.port))
		health.Recommendations = append(health.Recommendations,
			"Verify ATEM is powered on and on the same network")
		health.Recommendations = append(health.Recommendations,
			"Set ATEM_HOST environment variable if using non-default IP")
	}

	return health, nil
}

// Host returns the configured host
func (c *ATEMClient) Host() string {
	return c.host
}

// IsConnected returns connection status
func (c *ATEMClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}
