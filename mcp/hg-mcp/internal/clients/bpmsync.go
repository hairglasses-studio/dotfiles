// Package clients provides API clients for external services.
package clients

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"
)

// BPMSyncClient provides unified BPM synchronization across multiple systems
type BPMSyncClient struct {
	mu            sync.RWMutex
	masterSource  string
	linkedSystems map[string]bool
	currentBPM    float64
	lastSync      time.Time
	tapTimes      []time.Time

	// Clients
	abletonClient  *AbletonClient
	resolumeClient *ResolumeClient
	grandma3Client *GrandMA3Client
}

// BPMSystem represents a system capable of BPM sync
type BPMSystem struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"` // master, slave, bidirectional
	Connected  bool    `json:"connected"`
	CurrentBPM float64 `json:"current_bpm"`
	CanRead    bool    `json:"can_read"`
	CanWrite   bool    `json:"can_write"`
	Linked     bool    `json:"linked"`
}

// BPMSyncStatus represents the current sync state
type BPMSyncStatus struct {
	MasterSource string      `json:"master_source"`
	CurrentBPM   float64     `json:"current_bpm"`
	LastSync     time.Time   `json:"last_sync"`
	Systems      []BPMSystem `json:"systems"`
	InSync       bool        `json:"in_sync"`
	DriftMS      float64     `json:"drift_ms,omitempty"`
}

// BPMSyncHealth represents health status
type BPMSyncHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	MasterConnected bool     `json:"master_connected"`
	LinkedCount     int      `json:"linked_count"`
	ConnectedCount  int      `json:"connected_count"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// NewBPMSyncClient creates a new BPM sync client
func NewBPMSyncClient() (*BPMSyncClient, error) {
	return &BPMSyncClient{
		masterSource:  "ableton",
		linkedSystems: make(map[string]bool),
		currentBPM:    120.0,
		tapTimes:      make([]time.Time, 0, 4),
	}, nil
}

// getAbletonClient lazily initializes the Ableton client
func (c *BPMSyncClient) getAbletonClient() (*AbletonClient, error) {
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
func (c *BPMSyncClient) getResolumeClient() (*ResolumeClient, error) {
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
func (c *BPMSyncClient) getGrandMA3Client() (*GrandMA3Client, error) {
	if c.grandma3Client == nil {
		client, err := NewGrandMA3Client()
		if err != nil {
			return nil, err
		}
		c.grandma3Client = client
	}
	return c.grandma3Client, nil
}

// GetStatus returns the current BPM sync status across all systems
func (c *BPMSyncClient) GetStatus(ctx context.Context) (*BPMSyncStatus, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status := &BPMSyncStatus{
		MasterSource: c.masterSource,
		CurrentBPM:   c.currentBPM,
		LastSync:     c.lastSync,
		Systems:      make([]BPMSystem, 0),
		InSync:       true,
	}

	// Check Ableton
	if ableton, err := c.getAbletonClient(); err == nil {
		sys := BPMSystem{
			Name:     "ableton",
			Type:     "bidirectional",
			CanRead:  true,
			CanWrite: true,
			Linked:   c.linkedSystems["ableton"],
		}
		if abletonStatus, err := ableton.GetStatus(ctx); err == nil {
			sys.Connected = abletonStatus.Connected
			if abletonStatus.State != nil {
				sys.CurrentBPM = abletonStatus.State.Tempo
			}
		}
		status.Systems = append(status.Systems, sys)
	}

	// Check Resolume
	if resolume, err := c.getResolumeClient(); err == nil {
		sys := BPMSystem{
			Name:     "resolume",
			Type:     "bidirectional",
			CanRead:  true,
			CanWrite: true,
			Linked:   c.linkedSystems["resolume"],
		}
		if resolumeStatus, err := resolume.GetStatus(ctx); err == nil {
			sys.Connected = resolumeStatus.Connected
			sys.CurrentBPM = resolumeStatus.BPM
		}
		status.Systems = append(status.Systems, sys)
	}

	// Check grandMA3
	if grandma3, err := c.getGrandMA3Client(); err == nil {
		sys := BPMSystem{
			Name:     "grandma3",
			Type:     "slave",
			CanRead:  false,
			CanWrite: true,
			Linked:   c.linkedSystems["grandma3"],
		}
		if grandma3Status, err := grandma3.GetStatus(ctx); err == nil {
			sys.Connected = grandma3Status.Connected
		}
		status.Systems = append(status.Systems, sys)
	}

	// Calculate drift and sync status
	var maxDrift float64
	for _, sys := range status.Systems {
		if sys.Connected && sys.CurrentBPM > 0 && sys.CurrentBPM != c.currentBPM {
			drift := math.Abs(sys.CurrentBPM - c.currentBPM)
			if drift > maxDrift {
				maxDrift = drift
			}
			if drift > 0.1 {
				status.InSync = false
			}
		}
	}
	status.DriftMS = maxDrift * 1000 / c.currentBPM // Rough drift in ms per beat

	return status, nil
}

// SetMaster sets the master BPM source
func (c *BPMSyncClient) SetMaster(ctx context.Context, source string) error {
	validSources := map[string]bool{
		"ableton":  true,
		"resolume": true,
		"manual":   true,
	}

	if !validSources[source] {
		return fmt.Errorf("invalid master source: %s (valid: ableton, resolume, manual)", source)
	}

	c.mu.Lock()
	c.masterSource = source
	c.mu.Unlock()

	// If switching to a new master, read its current BPM
	if source != "manual" {
		bpm, err := c.readBPMFromSource(ctx, source)
		if err != nil {
			return fmt.Errorf("failed to read BPM from new master: %w", err)
		}
		c.mu.Lock()
		c.currentBPM = bpm
		c.mu.Unlock()
	}

	return nil
}

// readBPMFromSource reads BPM from a specific source
func (c *BPMSyncClient) readBPMFromSource(ctx context.Context, source string) (float64, error) {
	switch source {
	case "ableton":
		if ableton, err := c.getAbletonClient(); err == nil {
			return ableton.GetTempo(ctx)
		}
		return 0, fmt.Errorf("ableton client not available")

	case "resolume":
		if resolume, err := c.getResolumeClient(); err == nil {
			return resolume.GetBPM(ctx)
		}
		return 0, fmt.Errorf("resolume client not available")

	default:
		return c.currentBPM, nil
	}
}

// LinkSystem links a system to receive BPM updates from master
func (c *BPMSyncClient) LinkSystem(ctx context.Context, system string) error {
	validSystems := map[string]bool{
		"ableton":  true,
		"resolume": true,
		"grandma3": true,
	}

	if !validSystems[system] {
		return fmt.Errorf("invalid system: %s", system)
	}

	c.mu.Lock()
	c.linkedSystems[system] = true
	c.mu.Unlock()

	// Push current BPM to newly linked system
	return c.pushBPMToSystem(ctx, system, c.currentBPM)
}

// UnlinkSystem removes a system from BPM sync
func (c *BPMSyncClient) UnlinkSystem(ctx context.Context, system string) error {
	c.mu.Lock()
	delete(c.linkedSystems, system)
	c.mu.Unlock()
	return nil
}

// PushBPM pushes BPM to all linked systems
func (c *BPMSyncClient) PushBPM(ctx context.Context, bpm float64) error {
	if bpm < 20 || bpm > 999 {
		return fmt.Errorf("BPM out of range: %f (valid: 20-999)", bpm)
	}

	c.mu.Lock()
	c.currentBPM = bpm
	c.lastSync = time.Now()
	linkedSystems := make(map[string]bool)
	for k, v := range c.linkedSystems {
		linkedSystems[k] = v
	}
	c.mu.Unlock()

	var errs []error
	for system, linked := range linkedSystems {
		if linked {
			if err := c.pushBPMToSystem(ctx, system, bpm); err != nil {
				errs = append(errs, fmt.Errorf("%s: %w", system, err))
			}
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("some systems failed to sync: %v", errs)
	}

	return nil
}

// pushBPMToSystem pushes BPM to a specific system
func (c *BPMSyncClient) pushBPMToSystem(ctx context.Context, system string, bpm float64) error {
	switch system {
	case "ableton":
		if ableton, err := c.getAbletonClient(); err == nil {
			return ableton.SetTempo(ctx, bpm)
		}
		return fmt.Errorf("ableton client not available")

	case "resolume":
		if resolume, err := c.getResolumeClient(); err == nil {
			return resolume.SetBPM(ctx, bpm)
		}
		return fmt.Errorf("resolume client not available")

	case "grandma3":
		if grandma3, err := c.getGrandMA3Client(); err == nil {
			// Use speed master 1 by default
			return grandma3.SetBPMMaster(ctx, 1, float32(bpm))
		}
		return fmt.Errorf("grandma3 client not available")

	default:
		return fmt.Errorf("unknown system: %s", system)
	}
}

// TapTempo records a tap and calculates BPM from recent taps
func (c *BPMSyncClient) TapTempo(ctx context.Context) (float64, error) {
	now := time.Now()

	c.mu.Lock()
	// Reset if last tap was more than 2 seconds ago
	if len(c.tapTimes) > 0 && now.Sub(c.tapTimes[len(c.tapTimes)-1]) > 2*time.Second {
		c.tapTimes = c.tapTimes[:0]
	}

	c.tapTimes = append(c.tapTimes, now)

	// Keep only last 4 taps
	if len(c.tapTimes) > 4 {
		c.tapTimes = c.tapTimes[1:]
	}

	// Need at least 2 taps to calculate BPM
	if len(c.tapTimes) < 2 {
		c.mu.Unlock()
		return c.currentBPM, nil
	}

	// Calculate average interval
	var totalInterval time.Duration
	for i := 1; i < len(c.tapTimes); i++ {
		totalInterval += c.tapTimes[i].Sub(c.tapTimes[i-1])
	}
	avgInterval := totalInterval / time.Duration(len(c.tapTimes)-1)

	// Convert to BPM
	bpm := 60.0 / avgInterval.Seconds()

	// Clamp to valid range
	if bpm < 20 {
		bpm = 20
	} else if bpm > 999 {
		bpm = 999
	}

	c.currentBPM = bpm
	c.mu.Unlock()

	// Push to linked systems
	return bpm, c.PushBPM(ctx, bpm)
}

// SyncFromMaster reads BPM from master and pushes to all linked systems
func (c *BPMSyncClient) SyncFromMaster(ctx context.Context) error {
	c.mu.RLock()
	master := c.masterSource
	c.mu.RUnlock()

	bpm, err := c.readBPMFromSource(ctx, master)
	if err != nil {
		return fmt.Errorf("failed to read from master: %w", err)
	}

	return c.PushBPM(ctx, bpm)
}

// GetHealth returns health status
func (c *BPMSyncClient) GetHealth(ctx context.Context) (*BPMSyncHealth, error) {
	health := &BPMSyncHealth{
		Score:  100,
		Status: "healthy",
	}

	status, err := c.GetStatus(ctx)
	if err != nil {
		health.Score -= 50
		health.Issues = append(health.Issues, fmt.Sprintf("Failed to get status: %v", err))
	} else {
		// Check master connection
		for _, sys := range status.Systems {
			if sys.Name == status.MasterSource {
				health.MasterConnected = sys.Connected
				if !sys.Connected {
					health.Score -= 30
					health.Issues = append(health.Issues, fmt.Sprintf("Master source '%s' not connected", status.MasterSource))
					health.Recommendations = append(health.Recommendations, fmt.Sprintf("Ensure %s is running and accessible", status.MasterSource))
				}
			}
			if sys.Connected {
				health.ConnectedCount++
			}
			if sys.Linked {
				health.LinkedCount++
			}
		}

		// Check sync status
		if !status.InSync {
			health.Score -= 20
			health.Issues = append(health.Issues, fmt.Sprintf("Systems out of sync (drift: %.2fms)", status.DriftMS))
			health.Recommendations = append(health.Recommendations, "Run aftrs_sync_bpm_push to resync all systems")
		}

		// Check if any systems are linked
		if health.LinkedCount == 0 {
			health.Score -= 10
			health.Issues = append(health.Issues, "No systems linked for sync")
			health.Recommendations = append(health.Recommendations, "Link systems using aftrs_sync_bpm_link")
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
