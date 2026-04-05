// Package clients provides API clients for external services.
package clients

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// TriggerSyncClient provides unified scene/clip/cue triggering across systems
type TriggerSyncClient struct {
	mu          sync.RWMutex
	mappings    []TriggerMapping
	lastTrigger time.Time

	// Clients
	abletonClient  *AbletonClient
	resolumeClient *ResolumeClient
	grandma3Client *GrandMA3Client
	obsClient      *OBSClient
}

// TriggerMapping defines a cross-system trigger relationship
type TriggerMapping struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	Triggers    []TriggerTarget `json:"triggers"`
	Enabled     bool            `json:"enabled"`
	CreatedAt   time.Time       `json:"created_at"`
}

// TriggerTarget represents a target for a trigger action
type TriggerTarget struct {
	System     string `json:"system"`     // ableton, resolume, obs, grandma3
	Action     string `json:"action"`     // scene, clip, cue, executor
	Identifier string `json:"identifier"` // scene index, clip path, cue number, etc.
}

// TriggerResult represents the result of a trigger action
type TriggerResult struct {
	System  string `json:"system"`
	Action  string `json:"action"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Latency int64  `json:"latency_ms"`
}

// TriggerSyncStatus represents the current trigger sync state
type TriggerSyncStatus struct {
	MappingCount int             `json:"mapping_count"`
	LastTrigger  time.Time       `json:"last_trigger,omitempty"`
	Systems      []TriggerSystem `json:"systems"`
}

// TriggerSystem represents a system's trigger capabilities
type TriggerSystem struct {
	Name      string   `json:"name"`
	Connected bool     `json:"connected"`
	Actions   []string `json:"actions"`
}

// TriggerSyncHealth represents health status
type TriggerSyncHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	MappingCount    int      `json:"mapping_count"`
	ConnectedCount  int      `json:"connected_count"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// NewTriggerSyncClient creates a new trigger sync client
func NewTriggerSyncClient() (*TriggerSyncClient, error) {
	return &TriggerSyncClient{
		mappings: make([]TriggerMapping, 0),
	}, nil
}

// getAbletonClient lazily initializes the Ableton client
func (c *TriggerSyncClient) getAbletonClient() (*AbletonClient, error) {
	if c.abletonClient == nil {
		client, err := NewAbletonClient()
		if err != nil {
			return nil, err
		}
		c.abletonClient = client
	}
	return c.abletonClient, nil
}

// getResolumeClient lazily initializes the Resolume client
func (c *TriggerSyncClient) getResolumeClient() (*ResolumeClient, error) {
	if c.resolumeClient == nil {
		client, err := NewResolumeClient()
		if err != nil {
			return nil, err
		}
		c.resolumeClient = client
	}
	return c.resolumeClient, nil
}

// getGrandMA3Client lazily initializes the grandMA3 client
func (c *TriggerSyncClient) getGrandMA3Client() (*GrandMA3Client, error) {
	if c.grandma3Client == nil {
		client, err := NewGrandMA3Client()
		if err != nil {
			return nil, err
		}
		c.grandma3Client = client
	}
	return c.grandma3Client, nil
}

// getOBSClient lazily initializes the OBS client
func (c *TriggerSyncClient) getOBSClient() (*OBSClient, error) {
	if c.obsClient == nil {
		client, err := NewOBSClient()
		if err != nil {
			return nil, err
		}
		c.obsClient = client
	}
	return c.obsClient, nil
}

// GetStatus returns the current trigger sync status
func (c *TriggerSyncClient) GetStatus(ctx context.Context) (*TriggerSyncStatus, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status := &TriggerSyncStatus{
		MappingCount: len(c.mappings),
		LastTrigger:  c.lastTrigger,
		Systems:      make([]TriggerSystem, 0),
	}

	// Check Ableton
	sys := TriggerSystem{
		Name:    "ableton",
		Actions: []string{"scene", "clip"},
	}
	if ableton, err := c.getAbletonClient(); err == nil {
		if abletonStatus, err := ableton.GetStatus(ctx); err == nil {
			sys.Connected = abletonStatus.Connected
		}
	}
	status.Systems = append(status.Systems, sys)

	// Check Resolume
	sys = TriggerSystem{
		Name:    "resolume",
		Actions: []string{"clip", "column", "layer"},
	}
	if resolume, err := c.getResolumeClient(); err == nil {
		if resolumeStatus, err := resolume.GetStatus(ctx); err == nil {
			sys.Connected = resolumeStatus.Connected
		}
	}
	status.Systems = append(status.Systems, sys)

	// Check grandMA3
	sys = TriggerSystem{
		Name:    "grandma3",
		Actions: []string{"cue", "executor", "sequence"},
	}
	if grandma3, err := c.getGrandMA3Client(); err == nil {
		if grandma3Status, err := grandma3.GetStatus(ctx); err == nil {
			sys.Connected = grandma3Status.Connected
		}
	}
	status.Systems = append(status.Systems, sys)

	// Check OBS
	sys = TriggerSystem{
		Name:    "obs",
		Actions: []string{"scene"},
	}
	if obs, err := c.getOBSClient(); err == nil {
		if obsStatus, err := obs.GetStatus(ctx); err == nil {
			sys.Connected = obsStatus.Connected
		}
	}
	status.Systems = append(status.Systems, sys)

	return status, nil
}

// TriggerScene triggers a scene across all linked systems
func (c *TriggerSyncClient) TriggerScene(ctx context.Context, sceneIndex int, dryRun bool) ([]TriggerResult, error) {
	results := make([]TriggerResult, 0)

	// Ableton: Fire scene
	start := time.Now()
	result := TriggerResult{
		System: "ableton",
		Action: "scene",
	}
	if !dryRun {
		if ableton, err := c.getAbletonClient(); err == nil {
			if err := ableton.FireScene(ctx, sceneIndex); err != nil {
				result.Error = err.Error()
			} else {
				result.Success = true
			}
		} else {
			result.Error = "client not available"
		}
	} else {
		result.Success = true
	}
	result.Latency = time.Since(start).Milliseconds()
	results = append(results, result)

	// Resolume: Trigger column (scenes map to columns)
	start = time.Now()
	result = TriggerResult{
		System: "resolume",
		Action: "column",
	}
	if !dryRun {
		if resolume, err := c.getResolumeClient(); err == nil {
			if err := resolume.TriggerColumn(ctx, sceneIndex+1); err != nil {
				result.Error = err.Error()
			} else {
				result.Success = true
			}
		} else {
			result.Error = "client not available"
		}
	} else {
		result.Success = true
	}
	result.Latency = time.Since(start).Milliseconds()
	results = append(results, result)

	// grandMA3: Go to executor/cue
	start = time.Now()
	result = TriggerResult{
		System: "grandma3",
		Action: "cue",
	}
	if !dryRun {
		if grandma3, err := c.getGrandMA3Client(); err == nil {
			// Use sequence 1 and cue matching scene index
			if err := grandma3.GoCue(ctx, 1, float64(sceneIndex+1)); err != nil {
				result.Error = err.Error()
			} else {
				result.Success = true
			}
		} else {
			result.Error = "client not available"
		}
	} else {
		result.Success = true
	}
	result.Latency = time.Since(start).Milliseconds()
	results = append(results, result)

	// OBS: Set current scene
	start = time.Now()
	result = TriggerResult{
		System: "obs",
		Action: "scene",
	}
	if !dryRun {
		if obs, err := c.getOBSClient(); err == nil {
			// Get scenes and switch to matching index
			scenes, err := obs.GetScenes(ctx)
			if err == nil && sceneIndex < len(scenes) {
				if err := obs.SetCurrentScene(ctx, scenes[sceneIndex].Name); err != nil {
					result.Error = err.Error()
				} else {
					result.Success = true
				}
			} else {
				result.Error = fmt.Sprintf("scene index %d out of range", sceneIndex)
			}
		} else {
			result.Error = "client not available"
		}
	} else {
		result.Success = true
	}
	result.Latency = time.Since(start).Milliseconds()
	results = append(results, result)

	if !dryRun {
		c.mu.Lock()
		c.lastTrigger = time.Now()
		c.mu.Unlock()
	}

	return results, nil
}

// CreateMapping creates a new trigger mapping
func (c *TriggerSyncClient) CreateMapping(ctx context.Context, mapping TriggerMapping) error {
	if mapping.ID == "" {
		mapping.ID = fmt.Sprintf("mapping_%d", time.Now().UnixNano())
	}
	mapping.CreatedAt = time.Now()
	mapping.Enabled = true

	c.mu.Lock()
	c.mappings = append(c.mappings, mapping)
	c.mu.Unlock()

	return nil
}

// GetMappings returns all trigger mappings
func (c *TriggerSyncClient) GetMappings(ctx context.Context) ([]TriggerMapping, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	mappings := make([]TriggerMapping, len(c.mappings))
	copy(mappings, c.mappings)
	return mappings, nil
}

// DeleteMapping deletes a trigger mapping by ID
func (c *TriggerSyncClient) DeleteMapping(ctx context.Context, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i, m := range c.mappings {
		if m.ID == id {
			c.mappings = append(c.mappings[:i], c.mappings[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("mapping not found: %s", id)
}

// TriggerMapping triggers a specific mapping by ID
func (c *TriggerSyncClient) TriggerMappingByID(ctx context.Context, id string, dryRun bool) ([]TriggerResult, error) {
	c.mu.RLock()
	var mapping *TriggerMapping
	for _, m := range c.mappings {
		if m.ID == id && m.Enabled {
			mappingCopy := m
			mapping = &mappingCopy
			break
		}
	}
	c.mu.RUnlock()

	if mapping == nil {
		return nil, fmt.Errorf("mapping not found or disabled: %s", id)
	}

	return c.executeTriggers(ctx, mapping.Triggers, dryRun)
}

// executeTriggers executes a list of trigger targets
func (c *TriggerSyncClient) executeTriggers(ctx context.Context, targets []TriggerTarget, dryRun bool) ([]TriggerResult, error) {
	results := make([]TriggerResult, 0, len(targets))

	for _, target := range targets {
		start := time.Now()
		result := TriggerResult{
			System: target.System,
			Action: target.Action,
		}

		if dryRun {
			result.Success = true
			result.Latency = 0
			results = append(results, result)
			continue
		}

		var err error
		switch target.System {
		case "ableton":
			err = c.triggerAbleton(ctx, target)
		case "resolume":
			err = c.triggerResolume(ctx, target)
		case "grandma3":
			err = c.triggerGrandMA3(ctx, target)
		case "obs":
			err = c.triggerOBS(ctx, target)
		default:
			err = fmt.Errorf("unknown system: %s", target.System)
		}

		if err != nil {
			result.Error = err.Error()
		} else {
			result.Success = true
		}
		result.Latency = time.Since(start).Milliseconds()
		results = append(results, result)
	}

	if !dryRun {
		c.mu.Lock()
		c.lastTrigger = time.Now()
		c.mu.Unlock()
	}

	return results, nil
}

func (c *TriggerSyncClient) triggerAbleton(ctx context.Context, target TriggerTarget) error {
	ableton, err := c.getAbletonClient()
	if err != nil {
		return err
	}

	var idx int
	fmt.Sscanf(target.Identifier, "%d", &idx)

	switch target.Action {
	case "scene":
		return ableton.FireScene(ctx, idx)
	case "clip":
		// Format: "track:slot"
		var track, slot int
		fmt.Sscanf(target.Identifier, "%d:%d", &track, &slot)
		return ableton.FireClip(ctx, track, slot)
	default:
		return fmt.Errorf("unknown action: %s", target.Action)
	}
}

func (c *TriggerSyncClient) triggerResolume(ctx context.Context, target TriggerTarget) error {
	resolume, err := c.getResolumeClient()
	if err != nil {
		return err
	}

	switch target.Action {
	case "column":
		var idx int
		fmt.Sscanf(target.Identifier, "%d", &idx)
		return resolume.TriggerColumn(ctx, idx)
	case "clip":
		// Format: "layer:column"
		var layer, column int
		fmt.Sscanf(target.Identifier, "%d:%d", &layer, &column)
		return resolume.TriggerClip(ctx, layer, column)
	case "layer":
		// Toggle layer bypass
		var idx int
		fmt.Sscanf(target.Identifier, "%d", &idx)
		return resolume.SetLayerBypass(ctx, idx, false)
	default:
		return fmt.Errorf("unknown action: %s", target.Action)
	}
}

func (c *TriggerSyncClient) triggerGrandMA3(ctx context.Context, target TriggerTarget) error {
	grandma3, err := c.getGrandMA3Client()
	if err != nil {
		return err
	}

	switch target.Action {
	case "cue":
		// Format: "sequence:cue"
		var seq int
		var cue float64
		fmt.Sscanf(target.Identifier, "%d:%f", &seq, &cue)
		return grandma3.GoCue(ctx, seq, cue)
	case "executor":
		// Format: "page:executor"
		var page, exec int
		fmt.Sscanf(target.Identifier, "%d:%d", &page, &exec)
		return grandma3.GoExecutor(ctx, page, exec)
	default:
		return fmt.Errorf("unknown action: %s", target.Action)
	}
}

func (c *TriggerSyncClient) triggerOBS(ctx context.Context, target TriggerTarget) error {
	obs, err := c.getOBSClient()
	if err != nil {
		return err
	}

	switch target.Action {
	case "scene":
		return obs.SetCurrentScene(ctx, target.Identifier)
	default:
		return fmt.Errorf("unknown action: %s", target.Action)
	}
}

// GetHealth returns health status
func (c *TriggerSyncClient) GetHealth(ctx context.Context) (*TriggerSyncHealth, error) {
	health := &TriggerSyncHealth{
		Score:  100,
		Status: "healthy",
	}

	status, err := c.GetStatus(ctx)
	if err != nil {
		health.Score -= 30
		health.Issues = append(health.Issues, fmt.Sprintf("Failed to get status: %v", err))
	} else {
		health.MappingCount = status.MappingCount

		for _, sys := range status.Systems {
			if sys.Connected {
				health.ConnectedCount++
			}
		}

		if health.ConnectedCount == 0 {
			health.Score -= 50
			health.Issues = append(health.Issues, "No systems connected")
			health.Recommendations = append(health.Recommendations, "Ensure at least one system (Ableton, Resolume, OBS, grandMA3) is running")
		} else if health.ConnectedCount < 2 {
			health.Score -= 20
			health.Issues = append(health.Issues, "Only one system connected")
			health.Recommendations = append(health.Recommendations, "Connect additional systems for cross-system triggering")
		}

		if health.MappingCount == 0 {
			health.Score -= 10
			health.Issues = append(health.Issues, "No trigger mappings configured")
			health.Recommendations = append(health.Recommendations, "Create mappings using aftrs_trigger_link")
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
