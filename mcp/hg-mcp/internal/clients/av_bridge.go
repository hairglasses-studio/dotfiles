// Package clients provides API clients for external services.
package clients

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// AVBridgeClient bridges music metadata to visual systems (Resolume, TouchDesigner)
type AVBridgeClient struct {
	resolume      *ResolumeClient
	touchdesigner *TouchDesignerClient
	mappings      *AVMappingConfig
	lastSync      time.Time
	liveMode      bool
	liveModeStop  chan struct{}
	mu            sync.RWMutex
}

// AVBridgeStatus represents the bridge connection status
type AVBridgeStatus struct {
	Connected           bool            `json:"connected"`
	ResolumeStatus      *ResolumeStatus `json:"resolume_status,omitempty"`
	TouchDesignerStatus *TDStatus       `json:"touchdesigner_status,omitempty"`
	LiveMode            bool            `json:"live_mode"`
	LastSync            time.Time       `json:"last_sync"`
	MappingsLoaded      int             `json:"mappings_loaded"`
}

// AVMappingConfig holds all audio-visual mappings
type AVMappingConfig struct {
	KeyToColor     map[string]string `json:"key_to_color"`
	GenreToPreset  map[string]string `json:"genre_to_preset"`
	EnergyMappings []EnergyMapping   `json:"energy_mappings"`
	CustomMappings []CustomMapping   `json:"custom_mappings"`
}

// EnergyMapping maps energy levels to visual parameters
type EnergyMapping struct {
	Name      string  `json:"name"`
	MinEnergy float64 `json:"min_energy"`
	MaxEnergy float64 `json:"max_energy"`
	Target    string  `json:"target"`    // "resolume" or "touchdesigner"
	Parameter string  `json:"parameter"` // Parameter path
	MinValue  float64 `json:"min_value"`
	MaxValue  float64 `json:"max_value"`
}

// CustomMapping allows user-defined mappings
type CustomMapping struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Source    string                 `json:"source"` // "bpm", "key", "genre", "energy"
	Target    string                 `json:"target"` // "resolume" or "touchdesigner"
	Parameter string                 `json:"parameter"`
	ValueMap  map[string]interface{} `json:"value_map,omitempty"`
	Transform string                 `json:"transform,omitempty"` // "linear", "exponential", "step"
	Enabled   bool                   `json:"enabled"`
	CreatedAt time.Time              `json:"created_at"`
}

// TrackMetadata holds music track information for sync
type TrackMetadata struct {
	Artist       string  `json:"artist"`
	Title        string  `json:"title"`
	BPM          float64 `json:"bpm"`
	Key          string  `json:"key"` // Camelot notation (1A-12B)
	Genre        string  `json:"genre"`
	Energy       float64 `json:"energy"`       // 0.0-1.0
	Danceability float64 `json:"danceability"` // 0.0-1.0
	Duration     float64 `json:"duration"`     // seconds
}

// AVSyncResult represents the result of an AV sync operation
type AVSyncResult struct {
	Success       bool              `json:"success"`
	ResolumeSync  bool              `json:"resolume_sync"`
	TDSync        bool              `json:"td_sync"`
	BPMSet        float64           `json:"bpm_set,omitempty"`
	ColorSet      string            `json:"color_set,omitempty"`
	PresetLoaded  string            `json:"preset_loaded,omitempty"`
	ParamsUpdated map[string]string `json:"params_updated,omitempty"`
	Errors        []string          `json:"errors,omitempty"`
}

// SetlistVisual represents pre-loaded visuals for a setlist track
type SetlistVisual struct {
	TrackID     string   `json:"track_id"`
	Artist      string   `json:"artist"`
	Title       string   `json:"title"`
	BPM         float64  `json:"bpm"`
	Key         string   `json:"key"`
	Genre       string   `json:"genre"`
	PresetName  string   `json:"preset_name"`
	ColorScheme string   `json:"color_scheme"`
	ClipColumns []int    `json:"clip_columns,omitempty"` // Resolume columns to trigger
	TDPresets   []string `json:"td_presets,omitempty"`   // TouchDesigner presets
}

// Singleton pattern for AVBridgeClient
var (
	avBridgeInstance *AVBridgeClient
	avBridgeOnce     sync.Once
)

// GetAVBridgeClient returns the singleton AVBridgeClient
func GetAVBridgeClient() *AVBridgeClient {
	avBridgeOnce.Do(func() {
		avBridgeInstance = &AVBridgeClient{
			mappings: getDefaultMappings(),
		}
	})
	return avBridgeInstance
}

// NewAVBridgeClient creates a new AVBridgeClient (for testing)
func NewAVBridgeClient() (*AVBridgeClient, error) {
	return &AVBridgeClient{
		mappings: getDefaultMappings(),
	}, nil
}

// getDefaultMappings returns the default audio-visual mappings
func getDefaultMappings() *AVMappingConfig {
	return &AVMappingConfig{
		// Camelot Wheel to Color mapping
		// Based on color theory - musical keys mapped to hues
		KeyToColor: map[string]string{
			// Minor keys (A column) - cooler, darker tones
			"1A":  "#FF6B6B", // Abm - Red (passionate)
			"2A":  "#FFE66D", // Ebm - Yellow (bright)
			"3A":  "#88D8B0", // Bbm - Mint (fresh)
			"4A":  "#4ECDC4", // Fm  - Teal (cool)
			"5A":  "#45B7D1", // Cm  - Sky Blue
			"6A":  "#5D5D9E", // Gm  - Purple Blue
			"7A":  "#9B59B6", // Dm  - Purple
			"8A":  "#E74C3C", // Am  - Dark Red
			"9A":  "#F39C12", // Em  - Orange
			"10A": "#2ECC71", // Bm  - Green
			"11A": "#1ABC9C", // F#m - Cyan
			"12A": "#3498DB", // C#m - Blue

			// Major keys (B column) - warmer, brighter tones
			"1B":  "#E91E63", // B   - Pink
			"2B":  "#FFEB3B", // F#  - Bright Yellow
			"3B":  "#8BC34A", // Db  - Lime
			"4B":  "#00BCD4", // Ab  - Cyan
			"5B":  "#2196F3", // Eb  - Blue
			"6B":  "#673AB7", // Bb  - Deep Purple
			"7B":  "#9C27B0", // F   - Magenta
			"8B":  "#F44336", // C   - Red
			"9B":  "#FF9800", // G   - Orange
			"10B": "#4CAF50", // D   - Green
			"11B": "#009688", // A   - Teal
			"12B": "#03A9F4", // E   - Light Blue
		},

		// Genre to visual preset mapping
		GenreToPreset: map[string]string{
			"techno":        "dark-industrial",
			"tech-house":    "urban-minimal",
			"house":         "warm-organic",
			"deep-house":    "underwater-flow",
			"progressive":   "cosmic-journey",
			"trance":        "cosmic-flow",
			"psytrance":     "psychedelic-fractal",
			"drum-and-bass": "aggressive-glitch",
			"jungle":        "jungle-organic",
			"dubstep":       "heavy-bass",
			"ambient":       "ethereal-slow",
			"downtempo":     "chill-waves",
			"electro":       "retro-synth",
			"breakbeat":     "kinetic-motion",
			"hardcore":      "strobe-chaos",
			"minimal":       "geometric-clean",
			"acid":          "acid-reactive",
			"industrial":    "dark-machine",
			"garage":        "urban-groove",
			"uk-garage":     "london-underground",
			"disco":         "retro-disco",
			"nu-disco":      "modern-disco",
			"funk":          "funky-groove",
			"hip-hop":       "street-style",
			"trap":          "trap-visual",
			"bass":          "bass-reactive",
			"experimental":  "abstract-art",
			"idm":           "glitch-abstract",
		},

		// Energy to effect intensity mappings
		EnergyMappings: []EnergyMapping{
			{
				Name:      "Master Intensity",
				MinEnergy: 0.0,
				MaxEnergy: 1.0,
				Target:    "resolume",
				Parameter: "/composition/video/opacity",
				MinValue:  0.6,
				MaxValue:  1.0,
			},
			{
				Name:      "Effect Intensity",
				MinEnergy: 0.0,
				MaxEnergy: 1.0,
				Target:    "touchdesigner",
				Parameter: "/project1/effect_intensity",
				MinValue:  0.0,
				MaxValue:  1.0,
			},
			{
				Name:      "Strobe Threshold",
				MinEnergy: 0.7,
				MaxEnergy: 1.0,
				Target:    "resolume",
				Parameter: "/composition/layers/1/effects/strobe/rate",
				MinValue:  0.0,
				MaxValue:  0.8,
			},
		},

		CustomMappings: []CustomMapping{},
	}
}

// initClients lazily initializes the visual clients
func (c *AVBridgeClient) initClients() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.resolume == nil {
		c.resolume, _ = NewResolumeClient()
	}
	if c.touchdesigner == nil {
		c.touchdesigner, _ = NewTouchDesignerClient()
	}
}

// GetStatus returns the bridge connection status
func (c *AVBridgeClient) GetStatus(ctx context.Context) (*AVBridgeStatus, error) {
	c.initClients()

	status := &AVBridgeStatus{
		Connected:      false,
		LiveMode:       c.liveMode,
		LastSync:       c.lastSync,
		MappingsLoaded: len(c.mappings.KeyToColor) + len(c.mappings.GenreToPreset) + len(c.mappings.EnergyMappings),
	}

	// Check Resolume connection
	if c.resolume != nil {
		resStatus, err := c.resolume.GetStatus(ctx)
		if err == nil {
			status.ResolumeStatus = resStatus
			if resStatus.Connected {
				status.Connected = true
			}
		}
	}

	// Check TouchDesigner connection
	if c.touchdesigner != nil {
		tdStatus, err := c.touchdesigner.GetStatus(ctx)
		if err == nil {
			status.TouchDesignerStatus = tdStatus
			if tdStatus.Connected {
				status.Connected = true
			}
		}
	}

	return status, nil
}

// SyncBPMToResolume pushes BPM to Resolume's tempo controller
func (c *AVBridgeClient) SyncBPMToResolume(ctx context.Context, bpm float64) error {
	c.initClients()

	if c.resolume == nil {
		return fmt.Errorf("resolume client not available")
	}

	if bpm < 20 || bpm > 999 {
		return fmt.Errorf("BPM must be between 20 and 999, got %.1f", bpm)
	}

	if err := c.resolume.SetBPM(ctx, bpm); err != nil {
		return fmt.Errorf("set resolume BPM: %w", err)
	}

	c.lastSync = time.Now()
	return nil
}

// SyncBPMToTouchDesigner pushes BPM to TouchDesigner
func (c *AVBridgeClient) SyncBPMToTouchDesigner(ctx context.Context, bpm float64) error {
	c.initClients()

	if c.touchdesigner == nil {
		return fmt.Errorf("touchdesigner client not available")
	}

	// Set BPM as a variable that TD projects can reference
	if err := c.touchdesigner.SetVariable(ctx, "current_bpm", fmt.Sprintf("%.2f", bpm)); err != nil {
		return fmt.Errorf("set TD BPM variable: %w", err)
	}

	// Also try to set a common BPM parameter path
	_ = c.touchdesigner.SetParameter(ctx, "/project1/bpm", "value", bpm)

	c.lastSync = time.Now()
	return nil
}

// SyncBPMToBoth pushes BPM to both Resolume and TouchDesigner
func (c *AVBridgeClient) SyncBPMToBoth(ctx context.Context, bpm float64) (*AVSyncResult, error) {
	result := &AVSyncResult{
		Success:       true,
		BPMSet:        bpm,
		ParamsUpdated: make(map[string]string),
	}

	// Sync to Resolume
	if err := c.SyncBPMToResolume(ctx, bpm); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Resolume: %v", err))
	} else {
		result.ResolumeSync = true
		result.ParamsUpdated["resolume_bpm"] = fmt.Sprintf("%.2f", bpm)
	}

	// Sync to TouchDesigner
	if err := c.SyncBPMToTouchDesigner(ctx, bpm); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("TouchDesigner: %v", err))
	} else {
		result.TDSync = true
		result.ParamsUpdated["td_bpm"] = fmt.Sprintf("%.2f", bpm)
	}

	result.Success = result.ResolumeSync || result.TDSync
	return result, nil
}

// MapKeyToColor returns the color for a musical key (Camelot notation)
func (c *AVBridgeClient) MapKeyToColor(key string) string {
	if color, ok := c.mappings.KeyToColor[key]; ok {
		return color
	}
	// Default to white if key not found
	return "#FFFFFF"
}

// SyncKeyToColor pushes a key-based color palette to visual systems
func (c *AVBridgeClient) SyncKeyToColor(ctx context.Context, key string) (*AVSyncResult, error) {
	c.initClients()

	color := c.MapKeyToColor(key)
	result := &AVSyncResult{
		Success:       true,
		ColorSet:      color,
		ParamsUpdated: make(map[string]string),
	}

	// Send to Resolume via Dashboard String (for text display) and effect parameter
	if c.resolume != nil {
		// Set dashboard string with key info
		_ = c.resolume.SetDashboardString(3, key)
		result.ResolumeSync = true
		result.ParamsUpdated["resolume_key_display"] = key
		result.ParamsUpdated["resolume_color"] = color
	}

	// Send to TouchDesigner
	if c.touchdesigner != nil {
		// Set color as RGB components (0-1 range)
		r, g, b := hexToRGB(color)
		_ = c.touchdesigner.SetVariable(ctx, "key_color_r", fmt.Sprintf("%.3f", r))
		_ = c.touchdesigner.SetVariable(ctx, "key_color_g", fmt.Sprintf("%.3f", g))
		_ = c.touchdesigner.SetVariable(ctx, "key_color_b", fmt.Sprintf("%.3f", b))
		_ = c.touchdesigner.SetVariable(ctx, "current_key", key)
		result.TDSync = true
		result.ParamsUpdated["td_key"] = key
		result.ParamsUpdated["td_color"] = color
	}

	c.lastSync = time.Now()
	return result, nil
}

// MapGenreToPreset returns the visual preset for a genre
func (c *AVBridgeClient) MapGenreToPreset(genre string) string {
	if preset, ok := c.mappings.GenreToPreset[genre]; ok {
		return preset
	}
	// Default preset
	return "default-visual"
}

// SyncGenrePreset loads a genre-appropriate visual preset
func (c *AVBridgeClient) SyncGenrePreset(ctx context.Context, genre string) (*AVSyncResult, error) {
	c.initClients()

	preset := c.MapGenreToPreset(genre)
	result := &AVSyncResult{
		Success:       true,
		PresetLoaded:  preset,
		ParamsUpdated: make(map[string]string),
	}

	// Notify Resolume (via dashboard or effect selection)
	if c.resolume != nil {
		_ = c.resolume.SetDashboardString(5, genre)
		result.ResolumeSync = true
		result.ParamsUpdated["resolume_genre"] = genre
		result.ParamsUpdated["resolume_preset"] = preset
	}

	// Notify TouchDesigner
	if c.touchdesigner != nil {
		_ = c.touchdesigner.SetVariable(ctx, "current_genre", genre)
		_ = c.touchdesigner.SetVariable(ctx, "current_preset", preset)
		result.TDSync = true
		result.ParamsUpdated["td_genre"] = genre
		result.ParamsUpdated["td_preset"] = preset
	}

	c.lastSync = time.Now()
	return result, nil
}

// SyncEnergyIntensity maps energy level to visual effect intensities
func (c *AVBridgeClient) SyncEnergyIntensity(ctx context.Context, energy float64) (*AVSyncResult, error) {
	c.initClients()

	if energy < 0 {
		energy = 0
	}
	if energy > 1 {
		energy = 1
	}

	result := &AVSyncResult{
		Success:       true,
		ParamsUpdated: make(map[string]string),
	}

	// Apply all energy mappings
	for _, mapping := range c.mappings.EnergyMappings {
		// Check if energy is in this mapping's range
		if energy < mapping.MinEnergy || energy > mapping.MaxEnergy {
			continue
		}

		// Calculate interpolated value
		normalizedEnergy := (energy - mapping.MinEnergy) / (mapping.MaxEnergy - mapping.MinEnergy)
		value := mapping.MinValue + normalizedEnergy*(mapping.MaxValue-mapping.MinValue)

		switch mapping.Target {
		case "resolume":
			if c.resolume != nil {
				// Use OSC to set parameter
				err := c.resolume.sendOSC(mapping.Parameter, float32(value))
				if err == nil {
					result.ResolumeSync = true
					result.ParamsUpdated[fmt.Sprintf("resolume_%s", mapping.Name)] = fmt.Sprintf("%.3f", value)
				}
			}
		case "touchdesigner":
			if c.touchdesigner != nil {
				// Set as variable
				err := c.touchdesigner.SetVariable(ctx, mapping.Name, fmt.Sprintf("%.3f", value))
				if err == nil {
					result.TDSync = true
					result.ParamsUpdated[fmt.Sprintf("td_%s", mapping.Name)] = fmt.Sprintf("%.3f", value)
				}
			}
		}
	}

	// Also set raw energy value
	if c.touchdesigner != nil {
		_ = c.touchdesigner.SetVariable(ctx, "track_energy", fmt.Sprintf("%.3f", energy))
	}

	c.lastSync = time.Now()
	return result, nil
}

// SyncTrackCue triggers visual changes on track change
func (c *AVBridgeClient) SyncTrackCue(ctx context.Context, track *TrackMetadata) (*AVSyncResult, error) {
	c.initClients()

	result := &AVSyncResult{
		Success:       true,
		ParamsUpdated: make(map[string]string),
	}

	// Sync all track properties

	// 1. BPM
	if track.BPM > 0 {
		bpmResult, _ := c.SyncBPMToBoth(ctx, track.BPM)
		if bpmResult != nil {
			result.BPMSet = track.BPM
			result.ResolumeSync = result.ResolumeSync || bpmResult.ResolumeSync
			result.TDSync = result.TDSync || bpmResult.TDSync
			for k, v := range bpmResult.ParamsUpdated {
				result.ParamsUpdated[k] = v
			}
		}
	}

	// 2. Key to color
	if track.Key != "" {
		keyResult, _ := c.SyncKeyToColor(ctx, track.Key)
		if keyResult != nil {
			result.ColorSet = keyResult.ColorSet
			for k, v := range keyResult.ParamsUpdated {
				result.ParamsUpdated[k] = v
			}
		}
	}

	// 3. Genre preset
	if track.Genre != "" {
		genreResult, _ := c.SyncGenrePreset(ctx, track.Genre)
		if genreResult != nil {
			result.PresetLoaded = genreResult.PresetLoaded
			for k, v := range genreResult.ParamsUpdated {
				result.ParamsUpdated[k] = v
			}
		}
	}

	// 4. Energy intensity
	if track.Energy > 0 {
		energyResult, _ := c.SyncEnergyIntensity(ctx, track.Energy)
		if energyResult != nil {
			for k, v := range energyResult.ParamsUpdated {
				result.ParamsUpdated[k] = v
			}
		}
	}

	// 5. Set now playing info
	if c.resolume != nil {
		_ = c.resolume.SetNowPlaying(track.Artist, track.Title)
		result.ParamsUpdated["resolume_artist"] = track.Artist
		result.ParamsUpdated["resolume_title"] = track.Title
	}

	if c.touchdesigner != nil {
		_ = c.touchdesigner.SetVariable(ctx, "track_artist", track.Artist)
		_ = c.touchdesigner.SetVariable(ctx, "track_title", track.Title)
		result.ParamsUpdated["td_artist"] = track.Artist
		result.ParamsUpdated["td_title"] = track.Title
	}

	c.lastSync = time.Now()
	return result, nil
}

// LoadSetlistVisuals pre-loads visuals for an entire setlist
func (c *AVBridgeClient) LoadSetlistVisuals(ctx context.Context, tracks []TrackMetadata) ([]SetlistVisual, error) {
	visuals := make([]SetlistVisual, 0, len(tracks))

	for _, track := range tracks {
		// Generate track ID from artist+title
		trackID := generateTrackID(track.Artist, track.Title)

		// Map to visual settings
		visual := SetlistVisual{
			TrackID:     trackID,
			Artist:      track.Artist,
			Title:       track.Title,
			BPM:         track.BPM,
			Key:         track.Key,
			Genre:       track.Genre,
			ColorScheme: c.MapKeyToColor(track.Key),
			PresetName:  c.MapGenreToPreset(track.Genre),
		}

		visuals = append(visuals, visual)
	}

	return visuals, nil
}

// StartLiveMode starts real-time sync mode
func (c *AVBridgeClient) StartLiveMode(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.liveMode {
		return fmt.Errorf("live mode already active")
	}

	c.liveMode = true
	c.liveModeStop = make(chan struct{})

	// Note: In a full implementation, this would start a goroutine
	// that listens to DJ software (Rekordbox, Serato, etc.) for track changes
	// and automatically syncs to visual systems.

	return nil
}

// StopLiveMode stops real-time sync mode
func (c *AVBridgeClient) StopLiveMode() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.liveMode {
		return fmt.Errorf("live mode not active")
	}

	close(c.liveModeStop)
	c.liveMode = false

	return nil
}

// IsLiveModeActive returns whether live mode is active
func (c *AVBridgeClient) IsLiveModeActive() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.liveMode
}

// AddCustomMapping adds a user-defined mapping
func (c *AVBridgeClient) AddCustomMapping(mapping CustomMapping) error {
	if mapping.ID == "" {
		mapping.ID = generateMappingID()
	}
	mapping.CreatedAt = time.Now()
	mapping.Enabled = true

	c.mu.Lock()
	defer c.mu.Unlock()

	c.mappings.CustomMappings = append(c.mappings.CustomMappings, mapping)
	return nil
}

// RemoveCustomMapping removes a custom mapping by ID
func (c *AVBridgeClient) RemoveCustomMapping(mappingID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, m := range c.mappings.CustomMappings {
		if m.ID == mappingID {
			c.mappings.CustomMappings = append(
				c.mappings.CustomMappings[:i],
				c.mappings.CustomMappings[i+1:]...,
			)
			return nil
		}
	}

	return fmt.Errorf("mapping not found: %s", mappingID)
}

// GetMappings returns all current mappings
func (c *AVBridgeClient) GetMappings() *AVMappingConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy
	return &AVMappingConfig{
		KeyToColor:     c.mappings.KeyToColor,
		GenreToPreset:  c.mappings.GenreToPreset,
		EnergyMappings: c.mappings.EnergyMappings,
		CustomMappings: c.mappings.CustomMappings,
	}
}

// SaveMappings saves mappings to a file
func (c *AVBridgeClient) SaveMappings(filePath string) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := json.MarshalIndent(c.mappings, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal mappings: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("write mappings file: %w", err)
	}

	return nil
}

// LoadMappings loads mappings from a file
func (c *AVBridgeClient) LoadMappings(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("read mappings file: %w", err)
	}

	var mappings AVMappingConfig
	if err := json.Unmarshal(data, &mappings); err != nil {
		return fmt.Errorf("parse mappings: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.mappings = &mappings
	return nil
}

// GetKeyColorPalette returns all key-to-color mappings
func (c *AVBridgeClient) GetKeyColorPalette() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.mappings.KeyToColor
}

// GetGenrePresets returns all genre-to-preset mappings
func (c *AVBridgeClient) GetGenrePresets() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.mappings.GenreToPreset
}

// Helper functions

// hexToRGB converts a hex color to RGB values (0-1 range)
func hexToRGB(hex string) (r, g, b float64) {
	if len(hex) == 0 {
		return 1, 1, 1
	}

	// Remove # prefix if present
	if hex[0] == '#' {
		hex = hex[1:]
	}

	if len(hex) != 6 {
		return 1, 1, 1
	}

	var ri, gi, bi uint64
	fmt.Sscanf(hex, "%02x%02x%02x", &ri, &gi, &bi)

	return float64(ri) / 255.0, float64(gi) / 255.0, float64(bi) / 255.0
}

// generateTrackID generates a unique ID for a track
func generateTrackID(artist, title string) string {
	hash := sha256.Sum256([]byte(artist + "|" + title))
	return fmt.Sprintf("%x", hash[:8])
}

// generateMappingID generates a unique ID for a mapping
func generateMappingID() string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return fmt.Sprintf("mapping_%x", hash[:4])
}

// ============================================================================
// Retry Logic and Connection Management
// ============================================================================

// RetryConfig holds retry configuration
type RetryConfig struct {
	MaxRetries    int           `json:"max_retries"`
	InitialDelay  time.Duration `json:"initial_delay"`
	MaxDelay      time.Duration `json:"max_delay"`
	BackoffFactor float64       `json:"backoff_factor"`
}

// DefaultRetryConfig returns the default retry configuration
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:    3,
		InitialDelay:  500 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
	}
}

// WithRetry executes a function with configurable retries and exponential backoff
func WithRetry[T any](ctx context.Context, config *RetryConfig, operation func() (T, error)) (T, error) {
	if config == nil {
		config = DefaultRetryConfig()
	}

	var lastErr error
	var zero T
	delay := config.InitialDelay

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		result, err := operation()
		if err == nil {
			return result, nil
		}
		lastErr = err

		// Don't sleep on the last attempt
		if attempt < config.MaxRetries {
			select {
			case <-ctx.Done():
				return zero, ctx.Err()
			case <-time.After(delay):
			}

			// Exponential backoff with cap
			delay = time.Duration(float64(delay) * config.BackoffFactor)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}
	}

	return zero, fmt.Errorf("operation failed after %d retries: %w", config.MaxRetries+1, lastErr)
}

// SyncBPMToResolumeWithRetry pushes BPM to Resolume with retry logic
func (c *AVBridgeClient) SyncBPMToResolumeWithRetry(ctx context.Context, bpm float64, config *RetryConfig) error {
	_, err := WithRetry(ctx, config, func() (struct{}, error) {
		return struct{}{}, c.SyncBPMToResolume(ctx, bpm)
	})
	return err
}

// SyncBPMToTouchDesignerWithRetry pushes BPM to TouchDesigner with retry logic
func (c *AVBridgeClient) SyncBPMToTouchDesignerWithRetry(ctx context.Context, bpm float64, config *RetryConfig) error {
	_, err := WithRetry(ctx, config, func() (struct{}, error) {
		return struct{}{}, c.SyncBPMToTouchDesigner(ctx, bpm)
	})
	return err
}

// SyncTrackCueWithRetry triggers visual changes with retry logic
func (c *AVBridgeClient) SyncTrackCueWithRetry(ctx context.Context, track *TrackMetadata, config *RetryConfig) (*AVSyncResult, error) {
	return WithRetry(ctx, config, func() (*AVSyncResult, error) {
		return c.SyncTrackCue(ctx, track)
	})
}

// ============================================================================
// Enhanced Status Reporting
// ============================================================================

// ConnectionHealth represents detailed health information
type ConnectionHealth struct {
	Resolume      *SystemHealth `json:"resolume"`
	TouchDesigner *SystemHealth `json:"touchdesigner"`
	OverallHealth float64       `json:"overall_health"` // 0.0-1.0
	LastCheck     time.Time     `json:"last_check"`
	Issues        []string      `json:"issues,omitempty"`
}

// SystemHealth represents health for a single system
type SystemHealth struct {
	Connected       bool          `json:"connected"`
	Reachable       bool          `json:"reachable"`
	ResponseTime    time.Duration `json:"response_time_ms"`
	LastSuccessSync time.Time     `json:"last_success_sync,omitempty"`
	ErrorCount      int           `json:"error_count"`
	Status          string        `json:"status"` // healthy, degraded, offline
}

// GetConnectionHealth returns detailed health information for all systems
func (c *AVBridgeClient) GetConnectionHealth(ctx context.Context) *ConnectionHealth {
	c.initClients()

	health := &ConnectionHealth{
		LastCheck: time.Now(),
	}

	var resolumeHealth, tdHealth float64

	// Check Resolume
	health.Resolume = c.checkResolumeHealth(ctx)
	if health.Resolume.Connected {
		resolumeHealth = 1.0
	} else if health.Resolume.Reachable {
		resolumeHealth = 0.5
		health.Issues = append(health.Issues, "Resolume reachable but not fully connected")
	} else {
		health.Issues = append(health.Issues, "Resolume not reachable")
	}

	// Check TouchDesigner
	health.TouchDesigner = c.checkTouchDesignerHealth(ctx)
	if health.TouchDesigner.Connected {
		tdHealth = 1.0
	} else if health.TouchDesigner.Reachable {
		tdHealth = 0.5
		health.Issues = append(health.Issues, "TouchDesigner reachable but not fully connected")
	} else {
		health.Issues = append(health.Issues, "TouchDesigner not reachable")
	}

	// Calculate overall health (average of both systems)
	health.OverallHealth = (resolumeHealth + tdHealth) / 2.0

	return health
}

// checkResolumeHealth checks Resolume connection health
func (c *AVBridgeClient) checkResolumeHealth(ctx context.Context) *SystemHealth {
	health := &SystemHealth{
		Status: "offline",
	}

	if c.resolume == nil {
		return health
	}

	start := time.Now()

	// Try health check
	if err := c.resolume.HealthCheck(ctx); err == nil {
		health.Connected = true
		health.Reachable = true
		health.Status = "healthy"
		health.ResponseTime = time.Since(start)
		return health
	}

	// Try basic status
	status, err := c.resolume.GetStatus(ctx)
	health.ResponseTime = time.Since(start)

	if err == nil && status.Connected {
		health.Connected = true
		health.Reachable = true
		health.Status = "healthy"
	} else if status != nil {
		health.Reachable = true
		health.Status = "degraded"
	}

	return health
}

// checkTouchDesignerHealth checks TouchDesigner connection health
func (c *AVBridgeClient) checkTouchDesignerHealth(ctx context.Context) *SystemHealth {
	health := &SystemHealth{
		Status: "offline",
	}

	if c.touchdesigner == nil {
		return health
	}

	start := time.Now()

	status, err := c.touchdesigner.GetStatus(ctx)
	health.ResponseTime = time.Since(start)

	if err == nil && status.Connected {
		health.Connected = true
		health.Reachable = true
		health.Status = "healthy"
	} else if status != nil {
		health.Reachable = true
		health.Status = "degraded"
	}

	return health
}

// WaitForSystems waits until at least one visual system is available
func (c *AVBridgeClient) WaitForSystems(ctx context.Context, pollInterval time.Duration) error {
	c.initClients()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			health := c.GetConnectionHealth(ctx)
			if health.OverallHealth > 0 {
				return nil
			}
		}
	}
}

// WaitForAllSystems waits until both visual systems are available
func (c *AVBridgeClient) WaitForAllSystems(ctx context.Context, pollInterval time.Duration) error {
	c.initClients()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			health := c.GetConnectionHealth(ctx)
			if health.OverallHealth >= 1.0 {
				return nil
			}
		}
	}
}

// ============================================================================
// Batch Operations
// ============================================================================

// BatchSyncResult contains results from a batch sync operation
type BatchSyncResult struct {
	TotalTracks   int            `json:"total_tracks"`
	SuccessCount  int            `json:"success_count"`
	FailureCount  int            `json:"failure_count"`
	Results       []AVSyncResult `json:"results,omitempty"`
	TotalDuration time.Duration  `json:"total_duration"`
	AverageSyncMs float64        `json:"average_sync_ms"`
}

// BatchSyncTracks syncs multiple tracks with optional parallelism
func (c *AVBridgeClient) BatchSyncTracks(ctx context.Context, tracks []TrackMetadata, parallel bool) *BatchSyncResult {
	start := time.Now()
	result := &BatchSyncResult{
		TotalTracks: len(tracks),
		Results:     make([]AVSyncResult, 0, len(tracks)),
	}

	if parallel {
		// Parallel execution with sync
		var wg sync.WaitGroup
		var mu sync.Mutex
		results := make([]AVSyncResult, len(tracks))

		for i, track := range tracks {
			wg.Add(1)
			go func(idx int, t TrackMetadata) {
				defer wg.Done()
				syncResult, err := c.SyncTrackCue(ctx, &t)
				if err != nil {
					results[idx] = AVSyncResult{
						Success: false,
						Errors:  []string{err.Error()},
					}
				} else if syncResult != nil {
					results[idx] = *syncResult
				}
			}(i, track)
		}

		wg.Wait()

		mu.Lock()
		for _, r := range results {
			result.Results = append(result.Results, r)
			if r.Success {
				result.SuccessCount++
			} else {
				result.FailureCount++
			}
		}
		mu.Unlock()
	} else {
		// Sequential execution
		for _, track := range tracks {
			syncResult, err := c.SyncTrackCue(ctx, &track)
			if err != nil {
				result.Results = append(result.Results, AVSyncResult{
					Success: false,
					Errors:  []string{err.Error()},
				})
				result.FailureCount++
			} else if syncResult != nil {
				result.Results = append(result.Results, *syncResult)
				if syncResult.Success {
					result.SuccessCount++
				} else {
					result.FailureCount++
				}
			}
		}
	}

	result.TotalDuration = time.Since(start)
	if len(tracks) > 0 {
		result.AverageSyncMs = float64(result.TotalDuration.Milliseconds()) / float64(len(tracks))
	}

	return result
}

// ============================================================================
// Event Callbacks
// ============================================================================

// EventType represents the type of AV bridge event
type EventType string

const (
	EventTrackChange    EventType = "track_change"
	EventBPMSync        EventType = "bpm_sync"
	EventKeySync        EventType = "key_sync"
	EventGenreSync      EventType = "genre_sync"
	EventConnectionLost EventType = "connection_lost"
	EventReconnected    EventType = "reconnected"
)

// BridgeEvent represents an event from the AV bridge
type BridgeEvent struct {
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// EventCallback is a function that handles bridge events
type EventCallback func(event BridgeEvent)

// eventCallbacks stores registered callbacks
var eventCallbacks = make(map[EventType][]EventCallback)
var eventCallbacksMu sync.RWMutex

// OnEvent registers a callback for a specific event type
func (c *AVBridgeClient) OnEvent(eventType EventType, callback EventCallback) {
	eventCallbacksMu.Lock()
	defer eventCallbacksMu.Unlock()
	eventCallbacks[eventType] = append(eventCallbacks[eventType], callback)
}

// emitEvent fires an event to all registered callbacks
func (c *AVBridgeClient) emitEvent(eventType EventType, data map[string]interface{}) {
	eventCallbacksMu.RLock()
	callbacks := eventCallbacks[eventType]
	eventCallbacksMu.RUnlock()

	event := BridgeEvent{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}

	for _, cb := range callbacks {
		go cb(event) // Non-blocking
	}
}
