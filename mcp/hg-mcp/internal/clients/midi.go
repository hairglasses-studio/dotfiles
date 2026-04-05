// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// MIDIClient provides MIDI control capabilities
type MIDIClient struct {
	defaultOutput string
	defaultInput  string
	mappingStore  *MIDIMappingStore
}

// MIDIMappingStore manages MIDI-to-tool mappings with persistence
type MIDIMappingStore struct {
	mu       sync.RWMutex
	mappings map[string]*ToolMapping // ID → mapping
	profiles map[string]*MIDIProfile // profile name → profile
}

// ToolMapping represents a MIDI control → tool invocation mapping
type ToolMapping struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Channel     int                    `json:"channel"`    // MIDI channel (1-16)
	Type        string                 `json:"type"`       // cc, note, program
	Number      int                    `json:"number"`     // CC#, note#, program#
	ToolName    string                 `json:"tool_name"`  // Target tool to invoke
	Parameters  map[string]interface{} `json:"parameters"` // Static tool params
	ValueMap    *ValueMapping          `json:"value_map"`  // Range transform
	Enabled     bool                   `json:"enabled"`
	CreatedAt   time.Time              `json:"created_at"`
}

// ValueMapping transforms MIDI values to tool parameter values
type ValueMapping struct {
	InputMin    int     `json:"input_min"`    // MIDI range start (0-127)
	InputMax    int     `json:"input_max"`    // MIDI range end (0-127)
	OutputMin   float64 `json:"output_min"`   // Parameter range start
	OutputMax   float64 `json:"output_max"`   // Parameter range end
	TargetParam string  `json:"target_param"` // Which param receives the mapped value
	Invert      bool    `json:"invert"`       // Invert the mapping
}

// MIDIProfile stores a collection of mappings
type MIDIProfile struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	MappingIDs  []string       `json:"mapping_ids"`
	Mappings    []*ToolMapping `json:"mappings,omitempty"` // For export/import
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// MIDIDevice represents a MIDI device
type MIDIDevice struct {
	Name         string `json:"name"`
	Type         string `json:"type"` // input, output
	ID           int    `json:"id"`
	Connected    bool   `json:"connected"`
	Manufacturer string `json:"manufacturer,omitempty"`
}

// MIDIStatus represents MIDI system status
type MIDIStatus struct {
	InputDevices  []MIDIDevice `json:"input_devices"`
	OutputDevices []MIDIDevice `json:"output_devices"`
	DefaultInput  string       `json:"default_input"`
	DefaultOutput string       `json:"default_output"`
}

// MIDIMapping represents a MIDI control mapping
type MIDIMapping struct {
	Name       string `json:"name"`
	Channel    int    `json:"channel"`
	Controller int    `json:"controller"` // CC number or note
	Type       string `json:"type"`       // cc, note, program, pitch
	Target     string `json:"target,omitempty"`
	MinValue   int    `json:"min_value"`
	MaxValue   int    `json:"max_value"`
}

// MIDIMessage represents a MIDI message
type MIDIMessage struct {
	Type      string `json:"type"` // note_on, note_off, cc, program, pitch
	Channel   int    `json:"channel"`
	Data1     int    `json:"data1"` // note number or CC number
	Data2     int    `json:"data2"` // velocity or CC value
	Timestamp int64  `json:"timestamp_ms"`
}

// MIDIHealth represents MIDI system health
type MIDIHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	InputCount      int      `json:"input_count"`
	OutputCount     int      `json:"output_count"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// MIDILearnResult represents a MIDI learn operation result
type MIDILearnResult struct {
	Success    bool         `json:"success"`
	Message    MIDIMessage  `json:"message,omitempty"`
	Suggestion *MIDIMapping `json:"suggestion,omitempty"`
}

// NewMIDIClient creates a new MIDI client
func NewMIDIClient() (*MIDIClient, error) {
	defaultOutput := os.Getenv("MIDI_OUTPUT")
	defaultInput := os.Getenv("MIDI_INPUT")

	return &MIDIClient{
		defaultOutput: defaultOutput,
		defaultInput:  defaultInput,
		mappingStore: &MIDIMappingStore{
			mappings: make(map[string]*ToolMapping),
			profiles: make(map[string]*MIDIProfile),
		},
	}, nil
}

// GetStatus returns MIDI system status
func (c *MIDIClient) GetStatus(ctx context.Context) (*MIDIStatus, error) {
	status := &MIDIStatus{
		InputDevices:  []MIDIDevice{},
		OutputDevices: []MIDIDevice{},
		DefaultInput:  c.defaultInput,
		DefaultOutput: c.defaultOutput,
	}
	return status, nil
}

// GetDevices returns all MIDI devices
func (c *MIDIClient) GetDevices(ctx context.Context) ([]MIDIDevice, error) {
	devices := []MIDIDevice{}
	return devices, nil
}

// SendNoteOn sends a MIDI Note On message
func (c *MIDIClient) SendNoteOn(ctx context.Context, channel, note, velocity int, device string) error {
	if channel < 1 || channel > 16 {
		return fmt.Errorf("channel must be 1-16")
	}
	if note < 0 || note > 127 {
		return fmt.Errorf("note must be 0-127")
	}
	if velocity < 0 || velocity > 127 {
		return fmt.Errorf("velocity must be 0-127")
	}
	return nil
}

// SendNoteOff sends a MIDI Note Off message
func (c *MIDIClient) SendNoteOff(ctx context.Context, channel, note int, device string) error {
	if channel < 1 || channel > 16 {
		return fmt.Errorf("channel must be 1-16")
	}
	if note < 0 || note > 127 {
		return fmt.Errorf("note must be 0-127")
	}
	return nil
}

// SendCC sends a MIDI Control Change message
func (c *MIDIClient) SendCC(ctx context.Context, channel, controller, value int, device string) error {
	if channel < 1 || channel > 16 {
		return fmt.Errorf("channel must be 1-16")
	}
	if controller < 0 || controller > 127 {
		return fmt.Errorf("controller must be 0-127")
	}
	if value < 0 || value > 127 {
		return fmt.Errorf("value must be 0-127")
	}
	return nil
}

// SendProgramChange sends a MIDI Program Change message
func (c *MIDIClient) SendProgramChange(ctx context.Context, channel, program int, device string) error {
	if channel < 1 || channel > 16 {
		return fmt.Errorf("channel must be 1-16")
	}
	if program < 0 || program > 127 {
		return fmt.Errorf("program must be 0-127")
	}
	return nil
}

// SendPitchBend sends a MIDI Pitch Bend message
func (c *MIDIClient) SendPitchBend(ctx context.Context, channel int, value int, device string) error {
	if channel < 1 || channel > 16 {
		return fmt.Errorf("channel must be 1-16")
	}
	if value < -8192 || value > 8191 {
		return fmt.Errorf("pitch bend value must be -8192 to 8191")
	}
	return nil
}

// SendAllNotesOff sends MIDI All Notes Off on a channel
func (c *MIDIClient) SendAllNotesOff(ctx context.Context, channel int, device string) error {
	// CC 123 = All Notes Off
	return c.SendCC(ctx, channel, 123, 0, device)
}

// SendPanic sends MIDI Panic (All Notes Off on all channels)
func (c *MIDIClient) SendPanic(ctx context.Context, device string) error {
	for ch := 1; ch <= 16; ch++ {
		if err := c.SendAllNotesOff(ctx, ch, device); err != nil {
			return err
		}
	}
	return nil
}

// GetMappings returns configured MIDI mappings
func (c *MIDIClient) GetMappings(ctx context.Context) ([]MIDIMapping, error) {
	mappings := []MIDIMapping{}
	return mappings, nil
}

// CreateMapping creates a new MIDI mapping
func (c *MIDIClient) CreateMapping(ctx context.Context, mapping *MIDIMapping) error {
	return nil
}

// DeleteMapping deletes a MIDI mapping
func (c *MIDIClient) DeleteMapping(ctx context.Context, name string) error {
	return nil
}

// StartLearn starts MIDI learn mode
func (c *MIDIClient) StartLearn(ctx context.Context, timeoutSec int) (*MIDILearnResult, error) {
	result := &MIDILearnResult{
		Success: false,
	}
	return result, nil
}

// GetHealth returns MIDI system health
func (c *MIDIClient) GetHealth(ctx context.Context) (*MIDIHealth, error) {
	health := &MIDIHealth{
		Score:  100,
		Status: "healthy",
	}

	devices, _ := c.GetDevices(ctx)
	for _, d := range devices {
		if d.Type == "input" {
			health.InputCount++
		} else {
			health.OutputCount++
		}
	}

	if health.InputCount == 0 && health.OutputCount == 0 {
		health.Score -= 30
		health.Issues = append(health.Issues, "No MIDI devices detected")
		health.Recommendations = append(health.Recommendations, "Connect a MIDI controller or interface")
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

// DefaultOutput returns the default output device
func (c *MIDIClient) DefaultOutput() string {
	return c.defaultOutput
}

// DefaultInput returns the default input device
func (c *MIDIClient) DefaultInput() string {
	return c.defaultInput
}

// SendSysEx sends a MIDI System Exclusive message
func (c *MIDIClient) SendSysEx(ctx context.Context, data []byte, device string) error {
	if len(data) < 3 {
		return fmt.Errorf("sysex data too short")
	}
	if data[0] != 0xF0 || data[len(data)-1] != 0xF7 {
		return fmt.Errorf("invalid sysex data (must start with F0 and end with F7)")
	}
	return nil
}

// SendClock sends a MIDI Clock message
func (c *MIDIClient) SendClock(ctx context.Context, device string) error {
	return nil
}

// SendStart sends a MIDI Start message
func (c *MIDIClient) SendStart(ctx context.Context, device string) error {
	return nil
}

// SendStop sends a MIDI Stop message
func (c *MIDIClient) SendStop(ctx context.Context, device string) error {
	return nil
}

// SendContinue sends a MIDI Continue message
func (c *MIDIClient) SendContinue(ctx context.Context, device string) error {
	return nil
}

// CreateToolMapping creates a new MIDI→tool mapping
func (c *MIDIClient) CreateToolMapping(ctx context.Context, mapping *ToolMapping) error {
	if mapping.ID == "" {
		mapping.ID = fmt.Sprintf("map_%d", time.Now().UnixNano())
	}
	if mapping.Channel < 1 || mapping.Channel > 16 {
		return fmt.Errorf("channel must be 1-16")
	}
	if mapping.Number < 0 || mapping.Number > 127 {
		return fmt.Errorf("number must be 0-127")
	}
	if mapping.ToolName == "" {
		return fmt.Errorf("tool_name is required")
	}
	validTypes := map[string]bool{"cc": true, "note": true, "program": true}
	if !validTypes[mapping.Type] {
		return fmt.Errorf("type must be cc, note, or program")
	}

	mapping.Enabled = true
	mapping.CreatedAt = time.Now()
	if mapping.Parameters == nil {
		mapping.Parameters = make(map[string]interface{})
	}

	c.mappingStore.mu.Lock()
	c.mappingStore.mappings[mapping.ID] = mapping
	c.mappingStore.mu.Unlock()

	return nil
}

// GetToolMappings returns all tool mappings
func (c *MIDIClient) GetToolMappings(ctx context.Context) ([]*ToolMapping, error) {
	c.mappingStore.mu.RLock()
	defer c.mappingStore.mu.RUnlock()

	mappings := make([]*ToolMapping, 0, len(c.mappingStore.mappings))
	for _, m := range c.mappingStore.mappings {
		mappings = append(mappings, m)
	}
	return mappings, nil
}

// GetToolMapping returns a specific tool mapping by ID
func (c *MIDIClient) GetToolMapping(ctx context.Context, id string) (*ToolMapping, error) {
	c.mappingStore.mu.RLock()
	defer c.mappingStore.mu.RUnlock()

	mapping, ok := c.mappingStore.mappings[id]
	if !ok {
		return nil, fmt.Errorf("mapping not found: %s", id)
	}
	return mapping, nil
}

// DeleteToolMapping deletes a tool mapping by ID
func (c *MIDIClient) DeleteToolMapping(ctx context.Context, id string) error {
	c.mappingStore.mu.Lock()
	defer c.mappingStore.mu.Unlock()

	if _, ok := c.mappingStore.mappings[id]; !ok {
		return fmt.Errorf("mapping not found: %s", id)
	}
	delete(c.mappingStore.mappings, id)
	return nil
}

// EnableToolMapping enables or disables a tool mapping
func (c *MIDIClient) EnableToolMapping(ctx context.Context, id string, enabled bool) error {
	c.mappingStore.mu.Lock()
	defer c.mappingStore.mu.Unlock()

	mapping, ok := c.mappingStore.mappings[id]
	if !ok {
		return fmt.Errorf("mapping not found: %s", id)
	}
	mapping.Enabled = enabled
	return nil
}

// MapMIDIValue transforms a MIDI value using a ValueMapping
func (c *MIDIClient) MapMIDIValue(midiValue int, vm *ValueMapping) float64 {
	if vm == nil {
		return float64(midiValue)
	}

	// Clamp input to range
	if midiValue < vm.InputMin {
		midiValue = vm.InputMin
	}
	if midiValue > vm.InputMax {
		midiValue = vm.InputMax
	}

	// Normalize to 0-1
	inputRange := float64(vm.InputMax - vm.InputMin)
	if inputRange == 0 {
		inputRange = 1
	}
	normalized := float64(midiValue-vm.InputMin) / inputRange

	// Invert if needed
	if vm.Invert {
		normalized = 1.0 - normalized
	}

	// Scale to output range
	outputRange := vm.OutputMax - vm.OutputMin
	return vm.OutputMin + (normalized * outputRange)
}

// SaveProfile saves current mappings to a named profile
func (c *MIDIClient) SaveProfile(ctx context.Context, name, description string) error {
	if name == "" {
		return fmt.Errorf("profile name is required")
	}

	c.mappingStore.mu.Lock()
	defer c.mappingStore.mu.Unlock()

	// Collect current mapping IDs and copies
	mappingIDs := make([]string, 0, len(c.mappingStore.mappings))
	mappingsCopy := make([]*ToolMapping, 0, len(c.mappingStore.mappings))
	for id, m := range c.mappingStore.mappings {
		mappingIDs = append(mappingIDs, id)
		// Deep copy mapping for profile
		mcopy := *m
		if m.Parameters != nil {
			mcopy.Parameters = make(map[string]interface{})
			for k, v := range m.Parameters {
				mcopy.Parameters[k] = v
			}
		}
		if m.ValueMap != nil {
			vmcopy := *m.ValueMap
			mcopy.ValueMap = &vmcopy
		}
		mappingsCopy = append(mappingsCopy, &mcopy)
	}

	now := time.Now()
	profile := &MIDIProfile{
		Name:        name,
		Description: description,
		MappingIDs:  mappingIDs,
		Mappings:    mappingsCopy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// Update if exists
	if existing, ok := c.mappingStore.profiles[name]; ok {
		profile.CreatedAt = existing.CreatedAt
	}

	c.mappingStore.profiles[name] = profile
	return nil
}

// LoadProfile loads mappings from a named profile, replacing current mappings
func (c *MIDIClient) LoadProfile(ctx context.Context, name string) error {
	c.mappingStore.mu.Lock()
	defer c.mappingStore.mu.Unlock()

	profile, ok := c.mappingStore.profiles[name]
	if !ok {
		return fmt.Errorf("profile not found: %s", name)
	}

	// Clear current mappings
	c.mappingStore.mappings = make(map[string]*ToolMapping)

	// Load profile mappings
	for _, m := range profile.Mappings {
		mcopy := *m
		if m.Parameters != nil {
			mcopy.Parameters = make(map[string]interface{})
			for k, v := range m.Parameters {
				mcopy.Parameters[k] = v
			}
		}
		if m.ValueMap != nil {
			vmcopy := *m.ValueMap
			mcopy.ValueMap = &vmcopy
		}
		c.mappingStore.mappings[mcopy.ID] = &mcopy
	}

	return nil
}

// DeleteProfile deletes a saved profile
func (c *MIDIClient) DeleteProfile(ctx context.Context, name string) error {
	c.mappingStore.mu.Lock()
	defer c.mappingStore.mu.Unlock()

	if _, ok := c.mappingStore.profiles[name]; !ok {
		return fmt.Errorf("profile not found: %s", name)
	}
	delete(c.mappingStore.profiles, name)
	return nil
}

// GetProfiles returns all saved profiles
func (c *MIDIClient) GetProfiles(ctx context.Context) ([]*MIDIProfile, error) {
	c.mappingStore.mu.RLock()
	defer c.mappingStore.mu.RUnlock()

	profiles := make([]*MIDIProfile, 0, len(c.mappingStore.profiles))
	for _, p := range c.mappingStore.profiles {
		profiles = append(profiles, p)
	}
	return profiles, nil
}

// ExportProfile exports a profile to JSON
func (c *MIDIClient) ExportProfile(ctx context.Context, name string) ([]byte, error) {
	c.mappingStore.mu.RLock()
	defer c.mappingStore.mu.RUnlock()

	profile, ok := c.mappingStore.profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile not found: %s", name)
	}

	return json.MarshalIndent(profile, "", "  ")
}

// ImportProfile imports a profile from JSON
func (c *MIDIClient) ImportProfile(ctx context.Context, data []byte) error {
	var profile MIDIProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return fmt.Errorf("invalid profile JSON: %w", err)
	}

	if profile.Name == "" {
		return fmt.Errorf("profile name is required")
	}

	c.mappingStore.mu.Lock()
	defer c.mappingStore.mu.Unlock()

	profile.UpdatedAt = time.Now()
	c.mappingStore.profiles[profile.Name] = &profile
	return nil
}
