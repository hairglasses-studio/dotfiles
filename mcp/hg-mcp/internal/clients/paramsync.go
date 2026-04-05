// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ParamSyncClient manages parameter mappings between systems
type ParamSyncClient struct {
	mu          sync.RWMutex
	mappings    map[string]*ParamMapping
	activeSyncs map[string]*ContinuousSyncState // Active continuous sync goroutines
	configDir   string

	// Clients for parameter access
	abletonClient  *AbletonClient
	resolumeClient *ResolumeClient
	grandma3Client *GrandMA3Client
}

// ParamMapping represents a parameter mapping between systems
type ParamMapping struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Source      ParamEndpoint   `json:"source"`
	Target      ParamEndpoint   `json:"target"`
	Transform   *ParamTransform `json:"transform,omitempty"`
	Enabled     bool            `json:"enabled"`
	CreatedAt   time.Time       `json:"created_at"`
	LastSync    time.Time       `json:"last_sync,omitempty"`
	SyncCount   int64           `json:"sync_count"`
}

// ParamEndpoint represents a parameter endpoint (source or target)
type ParamEndpoint struct {
	System    string                 `json:"system"`     // ableton, resolume, grandma3, touchdesigner
	Type      string                 `json:"type"`       // track, device, layer, clip, effect, executor
	Path      string                 `json:"path"`       // e.g., "track/0/device/1/param/3"
	ParamName string                 `json:"param_name"` // Human-readable param name
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// ParamTransform defines how to transform values between systems
type ParamTransform struct {
	InputMin  float64 `json:"input_min"`
	InputMax  float64 `json:"input_max"`
	OutputMin float64 `json:"output_min"`
	OutputMax float64 `json:"output_max"`
	Curve     string  `json:"curve,omitempty"` // linear, exponential, logarithmic
	Invert    bool    `json:"invert"`
}

// ParamValue represents a parameter value
type ParamValue struct {
	Endpoint  ParamEndpoint `json:"endpoint"`
	Value     float64       `json:"value"`
	Timestamp time.Time     `json:"timestamp"`
}

// ParamSyncStatus represents the current sync status
type ParamSyncStatus struct {
	TotalMappings   int             `json:"total_mappings"`
	EnabledMappings int             `json:"enabled_mappings"`
	Mappings        []*ParamMapping `json:"mappings"`
	LastActivity    time.Time       `json:"last_activity"`
}

// NewParamSyncClient creates a new parameter sync client
func NewParamSyncClient() (*ParamSyncClient, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/tmp"
	}
	configDir := filepath.Join(homeDir, ".aftrs", "paramsync")

	// Ensure directory exists
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	client := &ParamSyncClient{
		mappings:  make(map[string]*ParamMapping),
		configDir: configDir,
	}

	// Load existing mappings
	if err := client.loadMappings(); err != nil {
		fmt.Printf("Warning: failed to load existing mappings: %v\n", err)
	}

	return client, nil
}

// loadMappings loads mappings from disk
func (c *ParamSyncClient) loadMappings() error {
	configFile := filepath.Join(c.configDir, "mappings.json")

	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var mappings []*ParamMapping
	if err := json.Unmarshal(data, &mappings); err != nil {
		return err
	}

	for _, m := range mappings {
		c.mappings[m.ID] = m
	}

	return nil
}

// saveMappings saves mappings to disk
func (c *ParamSyncClient) saveMappings() error {
	c.mu.RLock()
	mappings := make([]*ParamMapping, 0, len(c.mappings))
	for _, m := range c.mappings {
		mappings = append(mappings, m)
	}
	c.mu.RUnlock()

	data, err := json.MarshalIndent(mappings, "", "  ")
	if err != nil {
		return err
	}

	configFile := filepath.Join(c.configDir, "mappings.json")
	return os.WriteFile(configFile, data, 0644)
}

// Lazy client initialization
func (c *ParamSyncClient) getAbletonClient() (*AbletonClient, error) {
	if c.abletonClient == nil {
		client, err := NewAbletonClient()
		if err != nil {
			return nil, err
		}
		c.abletonClient = client
	}
	return c.abletonClient, nil
}

func (c *ParamSyncClient) getResolumeClient() (*ResolumeClient, error) {
	if c.resolumeClient == nil {
		client, err := NewResolumeClient()
		if err != nil {
			return nil, err
		}
		c.resolumeClient = client
	}
	return c.resolumeClient, nil
}

func (c *ParamSyncClient) getGrandMA3Client() (*GrandMA3Client, error) {
	if c.grandma3Client == nil {
		client, err := NewGrandMA3Client()
		if err != nil {
			return nil, err
		}
		c.grandma3Client = client
	}
	return c.grandma3Client, nil
}

// CreateMapping creates a new parameter mapping
func (c *ParamSyncClient) CreateMapping(ctx context.Context, name string, source, target ParamEndpoint, transform *ParamTransform) (*ParamMapping, error) {
	now := time.Now()
	id := fmt.Sprintf("map_%d", now.UnixNano())

	mapping := &ParamMapping{
		ID:        id,
		Name:      name,
		Source:    source,
		Target:    target,
		Transform: transform,
		Enabled:   true,
		CreatedAt: now,
	}

	// Validate source and target systems
	validSystems := map[string]bool{
		"ableton":       true,
		"resolume":      true,
		"grandma3":      true,
		"touchdesigner": true,
	}

	if !validSystems[source.System] {
		return nil, fmt.Errorf("invalid source system: %s", source.System)
	}
	if !validSystems[target.System] {
		return nil, fmt.Errorf("invalid target system: %s", target.System)
	}

	c.mu.Lock()
	c.mappings[id] = mapping
	c.mu.Unlock()

	if err := c.saveMappings(); err != nil {
		return mapping, fmt.Errorf("mapping created but failed to save: %w", err)
	}

	return mapping, nil
}

// GetMapping returns a mapping by ID
func (c *ParamSyncClient) GetMapping(ctx context.Context, id string) (*ParamMapping, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	mapping, exists := c.mappings[id]
	if !exists {
		return nil, fmt.Errorf("mapping not found: %s", id)
	}

	return mapping, nil
}

// ListMappings returns all mappings
func (c *ParamSyncClient) ListMappings(ctx context.Context) []*ParamMapping {
	c.mu.RLock()
	defer c.mu.RUnlock()

	mappings := make([]*ParamMapping, 0, len(c.mappings))
	for _, m := range c.mappings {
		mappings = append(mappings, m)
	}

	return mappings
}

// DeleteMapping removes a mapping
func (c *ParamSyncClient) DeleteMapping(ctx context.Context, id string) error {
	c.mu.Lock()
	if _, exists := c.mappings[id]; !exists {
		c.mu.Unlock()
		return fmt.Errorf("mapping not found: %s", id)
	}
	delete(c.mappings, id)
	c.mu.Unlock()

	return c.saveMappings()
}

// SetMappingEnabled enables or disables a mapping
func (c *ParamSyncClient) SetMappingEnabled(ctx context.Context, id string, enabled bool) error {
	c.mu.Lock()
	mapping, exists := c.mappings[id]
	if !exists {
		c.mu.Unlock()
		return fmt.Errorf("mapping not found: %s", id)
	}
	mapping.Enabled = enabled
	c.mu.Unlock()

	return c.saveMappings()
}

// PushValue pushes a value through a mapping manually
func (c *ParamSyncClient) PushValue(ctx context.Context, mappingID string, value float64) error {
	c.mu.RLock()
	mapping, exists := c.mappings[mappingID]
	if !exists {
		c.mu.RUnlock()
		return fmt.Errorf("mapping not found: %s", mappingID)
	}
	c.mu.RUnlock()

	// Apply transform
	outputValue := c.transformValue(value, mapping.Transform)

	// Push to target
	if err := c.setParameterValue(ctx, mapping.Target, outputValue); err != nil {
		return fmt.Errorf("failed to push to target: %w", err)
	}

	// Update mapping stats
	c.mu.Lock()
	mapping.LastSync = time.Now()
	mapping.SyncCount++
	c.mu.Unlock()

	return nil
}

// SyncMapping reads source and pushes to target
func (c *ParamSyncClient) SyncMapping(ctx context.Context, mappingID string) error {
	c.mu.RLock()
	mapping, exists := c.mappings[mappingID]
	if !exists {
		c.mu.RUnlock()
		return fmt.Errorf("mapping not found: %s", mappingID)
	}
	c.mu.RUnlock()

	if !mapping.Enabled {
		return fmt.Errorf("mapping is disabled")
	}

	// Read source value
	value, err := c.getParameterValue(ctx, mapping.Source)
	if err != nil {
		return fmt.Errorf("failed to read source: %w", err)
	}

	// Apply transform
	outputValue := c.transformValue(value, mapping.Transform)

	// Push to target
	if err := c.setParameterValue(ctx, mapping.Target, outputValue); err != nil {
		return fmt.Errorf("failed to push to target: %w", err)
	}

	// Update mapping stats
	c.mu.Lock()
	mapping.LastSync = time.Now()
	mapping.SyncCount++
	c.mu.Unlock()

	return nil
}

// transformValue applies a transform to a value
func (c *ParamSyncClient) transformValue(value float64, transform *ParamTransform) float64 {
	if transform == nil {
		return value
	}

	// Normalize input to 0-1
	normalized := (value - transform.InputMin) / (transform.InputMax - transform.InputMin)
	if normalized < 0 {
		normalized = 0
	}
	if normalized > 1 {
		normalized = 1
	}

	// Apply curve
	switch transform.Curve {
	case "exponential":
		normalized = normalized * normalized
	case "logarithmic":
		if normalized > 0 {
			normalized = 1 + (normalized-1)*normalized
		}
	}

	// Invert if needed
	if transform.Invert {
		normalized = 1 - normalized
	}

	// Scale to output range
	return transform.OutputMin + normalized*(transform.OutputMax-transform.OutputMin)
}

// getParameterValue reads a parameter value from a system
func (c *ParamSyncClient) getParameterValue(ctx context.Context, endpoint ParamEndpoint) (float64, error) {
	switch endpoint.System {
	case "ableton":
		client, err := c.getAbletonClient()
		if err != nil {
			return 0, err
		}

		// Parse path: track/N/device/M/param/P
		var trackIdx, deviceIdx, paramIdx int
		if _, err := fmt.Sscanf(endpoint.Path, "track/%d/device/%d/param/%d", &trackIdx, &deviceIdx, &paramIdx); err == nil {
			params, err := client.GetDeviceParameters(ctx, trackIdx, deviceIdx)
			if err != nil {
				return 0, err
			}
			if paramIdx < len(params) {
				return params[paramIdx].Value, nil
			}
			return 0, fmt.Errorf("param index out of range")
		}
		return 0, fmt.Errorf("invalid ableton path: %s", endpoint.Path)

	case "resolume":
		client, err := c.getResolumeClient()
		if err != nil {
			return 0, err
		}

		// Parse path: layer/N/effect/M/param/P or master/level
		if endpoint.Path == "master/level" {
			return client.GetMasterLevel(ctx)
		}

		var layerIdx, effectIdx int
		if _, err := fmt.Sscanf(endpoint.Path, "layer/%d/effect/%d", &layerIdx, &effectIdx); err == nil {
			params, err := client.GetEffectParams(ctx, layerIdx, effectIdx)
			if err != nil {
				return 0, err
			}
			if val, ok := params[endpoint.ParamName]; ok {
				if f, ok := val.(float64); ok {
					return f, nil
				}
			}
		}

		// Layer opacity
		var layerNum int
		if _, err := fmt.Sscanf(endpoint.Path, "layer/%d/opacity", &layerNum); err == nil {
			layers, err := client.GetLayers(ctx)
			if err != nil {
				return 0, err
			}
			if layerNum < len(layers) {
				return layers[layerNum].Opacity, nil
			}
		}
		return 0, fmt.Errorf("invalid resolume path: %s", endpoint.Path)

	case "grandma3":
		// grandMA3 parameter reading is limited
		return 0, fmt.Errorf("grandMA3 parameter reading not implemented")

	case "touchdesigner":
		// TouchDesigner would need MCP or HTTP API
		return 0, fmt.Errorf("touchdesigner parameter reading not implemented")

	default:
		return 0, fmt.Errorf("unknown system: %s", endpoint.System)
	}
}

// setParameterValue sets a parameter value on a system
func (c *ParamSyncClient) setParameterValue(ctx context.Context, endpoint ParamEndpoint, value float64) error {
	switch endpoint.System {
	case "ableton":
		client, err := c.getAbletonClient()
		if err != nil {
			return err
		}

		// Parse path: track/N/device/M/param/P
		var trackIdx, deviceIdx, paramIdx int
		if _, err := fmt.Sscanf(endpoint.Path, "track/%d/device/%d/param/%d", &trackIdx, &deviceIdx, &paramIdx); err == nil {
			return client.SetDeviceParameter(ctx, trackIdx, deviceIdx, paramIdx, value)
		}

		// Track volume/pan
		var trackNum int
		if _, err := fmt.Sscanf(endpoint.Path, "track/%d/volume", &trackNum); err == nil {
			return client.SetTrackVolume(ctx, trackNum, value)
		}
		if _, err := fmt.Sscanf(endpoint.Path, "track/%d/pan", &trackNum); err == nil {
			return client.SetTrackPan(ctx, trackNum, value)
		}

		return fmt.Errorf("invalid ableton path: %s", endpoint.Path)

	case "resolume":
		client, err := c.getResolumeClient()
		if err != nil {
			return err
		}

		// Master level
		if endpoint.Path == "master/level" {
			return client.SetMasterLevel(ctx, value)
		}

		// Layer opacity
		var layerNum int
		if _, err := fmt.Sscanf(endpoint.Path, "layer/%d/opacity", &layerNum); err == nil {
			return client.SetLayerOpacity(ctx, layerNum, value)
		}

		// Effect parameter
		var layerIdx, effectIdx int
		if _, err := fmt.Sscanf(endpoint.Path, "layer/%d/effect/%d", &layerIdx, &effectIdx); err == nil {
			return client.SetEffectParam(ctx, layerIdx, effectIdx, endpoint.ParamName, value)
		}

		return fmt.Errorf("invalid resolume path: %s", endpoint.Path)

	case "grandma3":
		client, err := c.getGrandMA3Client()
		if err != nil {
			return err
		}

		// Executor fader
		var page, exec int
		if _, err := fmt.Sscanf(endpoint.Path, "executor/%d.%d", &page, &exec); err == nil {
			return client.SetExecutorFader(ctx, page, exec, float32(value))
		}

		return fmt.Errorf("invalid grandma3 path: %s", endpoint.Path)

	case "touchdesigner":
		// TouchDesigner would need MCP or HTTP API
		return fmt.Errorf("touchdesigner parameter setting not implemented")

	default:
		return fmt.Errorf("unknown system: %s", endpoint.System)
	}
}

// ContinuousSyncState tracks a running continuous sync goroutine.
type ContinuousSyncState struct {
	MappingID  string    `json:"mapping_id"`
	IntervalMs int       `json:"interval_ms"`
	StartedAt  time.Time `json:"started_at"`
	cancel     context.CancelFunc
}

// StartContinuousSync starts a background goroutine that continuously syncs
// a mapping at the given interval. Returns an error if the mapping doesn't exist
// or if continuous sync is already running for this mapping.
func (c *ParamSyncClient) StartContinuousSync(ctx context.Context, mappingID string, intervalMs int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.mappings[mappingID]; !exists {
		return fmt.Errorf("mapping not found: %s", mappingID)
	}

	if c.activeSyncs == nil {
		c.activeSyncs = make(map[string]*ContinuousSyncState)
	}

	if _, running := c.activeSyncs[mappingID]; running {
		return fmt.Errorf("continuous sync already running for mapping: %s", mappingID)
	}

	if intervalMs < 10 {
		intervalMs = 100 // Minimum 100ms interval
	}

	syncCtx, cancel := context.WithCancel(context.Background())
	state := &ContinuousSyncState{
		MappingID:  mappingID,
		IntervalMs: intervalMs,
		StartedAt:  time.Now(),
		cancel:     cancel,
	}
	c.activeSyncs[mappingID] = state

	go func() {
		ticker := time.NewTicker(time.Duration(intervalMs) * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-syncCtx.Done():
				return
			case <-ticker.C:
				_ = c.SyncMapping(syncCtx, mappingID)
			}
		}
	}()

	return nil
}

// StopContinuousSync stops a running continuous sync for the given mapping.
func (c *ParamSyncClient) StopContinuousSync(mappingID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.activeSyncs == nil {
		return fmt.Errorf("no continuous sync running for mapping: %s", mappingID)
	}

	state, ok := c.activeSyncs[mappingID]
	if !ok {
		return fmt.Errorf("no continuous sync running for mapping: %s", mappingID)
	}

	state.cancel()
	delete(c.activeSyncs, mappingID)
	return nil
}

// ListContinuousSyncs returns all active continuous sync states.
func (c *ParamSyncClient) ListContinuousSyncs() []*ContinuousSyncState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]*ContinuousSyncState, 0, len(c.activeSyncs))
	for _, s := range c.activeSyncs {
		result = append(result, s)
	}
	return result
}

// GetStatus returns the current param sync status
func (c *ParamSyncClient) GetStatus(ctx context.Context) *ParamSyncStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status := &ParamSyncStatus{
		TotalMappings: len(c.mappings),
		Mappings:      make([]*ParamMapping, 0, len(c.mappings)),
	}

	var lastActivity time.Time
	for _, m := range c.mappings {
		status.Mappings = append(status.Mappings, m)
		if m.Enabled {
			status.EnabledMappings++
		}
		if m.LastSync.After(lastActivity) {
			lastActivity = m.LastSync
		}
	}

	status.LastActivity = lastActivity
	return status
}
