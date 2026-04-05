// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

// SnapshotsClient manages show state snapshots across systems
type SnapshotsClient struct {
	mu          sync.RWMutex
	snapshots   map[string]*Snapshot
	snapshotDir string

	// Clients for capturing state
	abletonClient     *AbletonClient
	resolumeClient    *ResolumeClient
	grandma3Client    *GrandMA3Client
	obsClient         *OBSClient
	showkontrolClient *ShowkontrolClient
}

// Snapshot represents a complete show state capture
type Snapshot struct {
	ID          string                  `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description,omitempty"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
	Systems     map[string]*SystemState `json:"systems"`
	Metadata    map[string]interface{}  `json:"metadata,omitempty"`
	Tags        []string                `json:"tags,omitempty"`
}

// SystemState represents the captured state of a single system
type SystemState struct {
	System     string                 `json:"system"`
	Connected  bool                   `json:"connected"`
	CapturedAt time.Time              `json:"captured_at"`
	State      map[string]interface{} `json:"state"`
	Error      string                 `json:"error,omitempty"`
}

// SnapshotDiff represents differences between two snapshots
type SnapshotDiff struct {
	Snapshot1   string                 `json:"snapshot1"`
	Snapshot2   string                 `json:"snapshot2"`
	Differences map[string]*SystemDiff `json:"differences"`
	Summary     string                 `json:"summary"`
}

// SystemDiff represents differences in a single system
type SystemDiff struct {
	System  string                 `json:"system"`
	Added   map[string]interface{} `json:"added,omitempty"`
	Removed map[string]interface{} `json:"removed,omitempty"`
	Changed []FieldChange          `json:"changed,omitempty"`
}

// FieldChange represents a changed field
type FieldChange struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
}

// NewSnapshotsClient creates a new snapshots client
func NewSnapshotsClient() (*SnapshotsClient, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/tmp"
	}
	snapshotDir := filepath.Join(homeDir, ".aftrs", "snapshots")

	// Ensure directory exists
	if err := os.MkdirAll(snapshotDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create snapshot directory: %w", err)
	}

	client := &SnapshotsClient{
		snapshots:   make(map[string]*Snapshot),
		snapshotDir: snapshotDir,
	}

	// Load existing snapshots
	if err := client.loadSnapshots(); err != nil {
		// Non-fatal, just log
		fmt.Printf("Warning: failed to load existing snapshots: %v\n", err)
	}

	return client, nil
}

// loadSnapshots loads snapshots from disk
func (c *SnapshotsClient) loadSnapshots() error {
	files, err := os.ReadDir(c.snapshotDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(c.snapshotDir, file.Name()))
		if err != nil {
			continue
		}

		var snapshot Snapshot
		if err := json.Unmarshal(data, &snapshot); err != nil {
			continue
		}

		c.snapshots[snapshot.ID] = &snapshot
	}

	return nil
}

// saveSnapshot saves a snapshot to disk
func (c *SnapshotsClient) saveSnapshot(snapshot *Snapshot) error {
	data, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return err
	}

	filename := filepath.Join(c.snapshotDir, snapshot.ID+".json")
	return os.WriteFile(filename, data, 0644)
}

// Lazy client initialization
func (c *SnapshotsClient) getAbletonClient() (*AbletonClient, error) {
	if c.abletonClient == nil {
		client, err := NewAbletonClient()
		if err != nil {
			return nil, err
		}
		c.abletonClient = client
	}
	return c.abletonClient, nil
}

func (c *SnapshotsClient) getResolumeClient() (*ResolumeClient, error) {
	if c.resolumeClient == nil {
		client, err := NewResolumeClient()
		if err != nil {
			return nil, err
		}
		c.resolumeClient = client
	}
	return c.resolumeClient, nil
}

func (c *SnapshotsClient) getGrandMA3Client() (*GrandMA3Client, error) {
	if c.grandma3Client == nil {
		client, err := NewGrandMA3Client()
		if err != nil {
			return nil, err
		}
		c.grandma3Client = client
	}
	return c.grandma3Client, nil
}

func (c *SnapshotsClient) getOBSClient() (*OBSClient, error) {
	if c.obsClient == nil {
		client, err := NewOBSClient()
		if err != nil {
			return nil, err
		}
		c.obsClient = client
	}
	return c.obsClient, nil
}

func (c *SnapshotsClient) getShowkontrolClient() (*ShowkontrolClient, error) {
	if c.showkontrolClient == nil {
		client, err := NewShowkontrolClient()
		if err != nil {
			return nil, err
		}
		c.showkontrolClient = client
	}
	return c.showkontrolClient, nil
}

// CaptureSnapshot captures state from all available systems
func (c *SnapshotsClient) CaptureSnapshot(ctx context.Context, name, description string, systems []string, tags []string) (*Snapshot, error) {
	now := time.Now()
	id := fmt.Sprintf("snap_%d", now.UnixNano())

	snapshot := &Snapshot{
		ID:          id,
		Name:        name,
		Description: description,
		CreatedAt:   now,
		UpdatedAt:   now,
		Systems:     make(map[string]*SystemState),
		Tags:        tags,
		Metadata:    make(map[string]interface{}),
	}

	// If no systems specified, capture all
	if len(systems) == 0 {
		systems = []string{"ableton", "resolume", "grandma3", "obs", "showkontrol"}
	}

	// Capture each system in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, sys := range systems {
		wg.Add(1)
		go func(system string) {
			defer wg.Done()

			state := c.captureSystemState(ctx, system)

			mu.Lock()
			snapshot.Systems[system] = state
			mu.Unlock()
		}(sys)
	}

	wg.Wait()

	// Save snapshot
	c.mu.Lock()
	c.snapshots[id] = snapshot
	c.mu.Unlock()

	if err := c.saveSnapshot(snapshot); err != nil {
		return snapshot, fmt.Errorf("snapshot captured but failed to save: %w", err)
	}

	return snapshot, nil
}

// captureSystemState captures state from a single system
func (c *SnapshotsClient) captureSystemState(ctx context.Context, system string) *SystemState {
	state := &SystemState{
		System:     system,
		CapturedAt: time.Now(),
		State:      make(map[string]interface{}),
	}

	switch system {
	case "ableton":
		if client, err := c.getAbletonClient(); err == nil {
			if status, err := client.GetStatus(ctx); err == nil {
				state.Connected = status.Connected
				if status.State != nil {
					state.State["tempo"] = status.State.Tempo
					state.State["playing"] = status.State.Playing
					state.State["recording"] = status.State.Recording
					state.State["song_time"] = status.State.SongTime
					state.State["selected_track"] = status.State.SelectedTrack
					state.State["selected_scene"] = status.State.SelectedScene
				}
				// Capture track states
				if tracks, err := client.GetTracks(ctx); err == nil {
					trackStates := make([]map[string]interface{}, len(tracks))
					for i, t := range tracks {
						trackStates[i] = map[string]interface{}{
							"name":   t.Name,
							"mute":   t.Mute,
							"solo":   t.Solo,
							"arm":    t.Arm,
							"volume": t.Volume,
							"pan":    t.Pan,
						}
					}
					state.State["tracks"] = trackStates
				}
			} else {
				state.Error = err.Error()
			}
		} else {
			state.Error = err.Error()
		}

	case "resolume":
		if client, err := c.getResolumeClient(); err == nil {
			if status, err := client.GetStatus(ctx); err == nil {
				state.Connected = status.Connected
				state.State["composition"] = status.Composition
				state.State["bpm"] = status.BPM
				state.State["playing"] = status.Playing
				state.State["master_level"] = status.MasterLevel
				// Get layers if available
				if layers, err := client.GetLayers(ctx); err == nil && len(layers) > 0 {
					layerStates := make([]map[string]interface{}, len(layers))
					for i, l := range layers {
						layerStates[i] = map[string]interface{}{
							"name":     l.Name,
							"opacity":  l.Opacity,
							"bypassed": l.Bypassed,
							"solo":     l.Solo,
						}
					}
					state.State["layers"] = layerStates
				}
			} else {
				state.Error = err.Error()
			}
		} else {
			state.Error = err.Error()
		}

	case "grandma3":
		if client, err := c.getGrandMA3Client(); err == nil {
			if status, err := client.GetStatus(ctx); err == nil {
				state.Connected = status.Connected
				state.State["host"] = status.Host
				state.State["port"] = status.Port
				state.State["protocol"] = status.Protocol
				// Note: grandMA3 detailed state requires show-specific queries
			} else {
				state.Error = err.Error()
			}
		} else {
			state.Error = err.Error()
		}

	case "obs":
		if client, err := c.getOBSClient(); err == nil {
			if status, err := client.GetStatus(ctx); err == nil {
				state.Connected = status.Connected
				state.State["current_scene"] = status.CurrentScene
				state.State["streaming"] = status.Streaming
				state.State["recording"] = status.Recording
				state.State["virtual_cam"] = status.VirtualCam
				// Get scene list
				if scenes, err := client.GetScenes(ctx); err == nil {
					sceneNames := make([]string, len(scenes))
					for i, s := range scenes {
						sceneNames[i] = s.Name
					}
					state.State["scenes"] = sceneNames
				}
			} else {
				state.Error = err.Error()
			}
		} else {
			state.Error = err.Error()
		}

	case "showkontrol":
		if client, err := c.getShowkontrolClient(); err == nil {
			if status, err := client.GetStatus(ctx); err == nil {
				state.Connected = status.Connected
				if status.CurrentShow != nil {
					state.State["show_id"] = status.CurrentShow.ID
					state.State["show_name"] = status.CurrentShow.Name
				}
				if status.Timecode != nil {
					state.State["timecode_position"] = status.Timecode.PositionTC
					state.State["timecode_running"] = status.Timecode.Running
				}
			} else {
				state.Error = err.Error()
			}
		} else {
			state.Error = err.Error()
		}

	default:
		state.Error = fmt.Sprintf("unknown system: %s", system)
	}

	return state
}

// RecallSnapshot restores state from a snapshot
func (c *SnapshotsClient) RecallSnapshot(ctx context.Context, snapshotID string, systems []string) error {
	c.mu.RLock()
	snapshot, exists := c.snapshots[snapshotID]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("snapshot not found: %s", snapshotID)
	}

	// If no systems specified, recall all that were captured
	if len(systems) == 0 {
		for sys := range snapshot.Systems {
			systems = append(systems, sys)
		}
	}

	var errs []error
	for _, sys := range systems {
		state, exists := snapshot.Systems[sys]
		if !exists {
			errs = append(errs, fmt.Errorf("%s: not in snapshot", sys))
			continue
		}

		if err := c.recallSystemState(ctx, sys, state); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", sys, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("some systems failed to recall: %v", errs)
	}

	return nil
}

// recallSystemState restores state to a single system
func (c *SnapshotsClient) recallSystemState(ctx context.Context, system string, state *SystemState) error {
	switch system {
	case "ableton":
		client, err := c.getAbletonClient()
		if err != nil {
			return err
		}

		// Restore tempo
		if tempo, ok := state.State["tempo"].(float64); ok {
			if err := client.SetTempo(ctx, tempo); err != nil {
				return fmt.Errorf("failed to set tempo: %w", err)
			}
		}

		// Restore track states
		if tracks, ok := state.State["tracks"].([]interface{}); ok {
			for i, t := range tracks {
				if track, ok := t.(map[string]interface{}); ok {
					if mute, ok := track["mute"].(bool); ok {
						_ = client.SetTrackMute(ctx, i, mute)
					}
					if solo, ok := track["solo"].(bool); ok {
						_ = client.SetTrackSolo(ctx, i, solo)
					}
					if volume, ok := track["volume"].(float64); ok {
						_ = client.SetTrackVolume(ctx, i, volume)
					}
				}
			}
		}

	case "resolume":
		client, err := c.getResolumeClient()
		if err != nil {
			return err
		}

		// Restore BPM
		if bpm, ok := state.State["bpm"].(float64); ok {
			if err := client.SetBPM(ctx, bpm); err != nil {
				return fmt.Errorf("failed to set BPM: %w", err)
			}
		}

		// Restore master level
		if level, ok := state.State["master_level"].(float64); ok {
			_ = client.SetMasterLevel(ctx, level)
		}

	case "grandma3":
		// grandMA3 state recall is limited - mostly use cues
		return nil

	case "obs":
		client, err := c.getOBSClient()
		if err != nil {
			return err
		}

		// Restore scene
		if scene, ok := state.State["current_scene"].(string); ok {
			if err := client.SetCurrentScene(ctx, scene); err != nil {
				return fmt.Errorf("failed to set scene: %w", err)
			}
		}

	case "showkontrol":
		client, err := c.getShowkontrolClient()
		if err != nil {
			return err
		}

		// Restore timecode position
		if tc, ok := state.State["timecode_position"].(string); ok {
			if err := client.GotoTimecode(ctx, tc); err != nil {
				return fmt.Errorf("failed to goto timecode: %w", err)
			}
		}

	default:
		return fmt.Errorf("unknown system: %s", system)
	}

	return nil
}

// ListSnapshots returns all snapshots
func (c *SnapshotsClient) ListSnapshots(ctx context.Context) []*Snapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	snapshots := make([]*Snapshot, 0, len(c.snapshots))
	for _, s := range c.snapshots {
		snapshots = append(snapshots, s)
	}

	// Sort by creation time (newest first)
	sort.Slice(snapshots, func(i, j int) bool {
		return snapshots[i].CreatedAt.After(snapshots[j].CreatedAt)
	})

	return snapshots
}

// GetSnapshot returns a specific snapshot
func (c *SnapshotsClient) GetSnapshot(ctx context.Context, id string) (*Snapshot, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	snapshot, exists := c.snapshots[id]
	if !exists {
		return nil, fmt.Errorf("snapshot not found: %s", id)
	}

	return snapshot, nil
}

// DeleteSnapshot removes a snapshot
func (c *SnapshotsClient) DeleteSnapshot(ctx context.Context, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.snapshots[id]; !exists {
		return fmt.Errorf("snapshot not found: %s", id)
	}

	delete(c.snapshots, id)

	// Remove from disk
	filename := filepath.Join(c.snapshotDir, id+".json")
	_ = os.Remove(filename)

	return nil
}

// DiffSnapshots compares two snapshots
func (c *SnapshotsClient) DiffSnapshots(ctx context.Context, id1, id2 string) (*SnapshotDiff, error) {
	c.mu.RLock()
	snap1, exists1 := c.snapshots[id1]
	snap2, exists2 := c.snapshots[id2]
	c.mu.RUnlock()

	if !exists1 {
		return nil, fmt.Errorf("snapshot not found: %s", id1)
	}
	if !exists2 {
		return nil, fmt.Errorf("snapshot not found: %s", id2)
	}

	diff := &SnapshotDiff{
		Snapshot1:   id1,
		Snapshot2:   id2,
		Differences: make(map[string]*SystemDiff),
	}

	// Get all systems from both snapshots
	allSystems := make(map[string]bool)
	for sys := range snap1.Systems {
		allSystems[sys] = true
	}
	for sys := range snap2.Systems {
		allSystems[sys] = true
	}

	changedCount := 0
	for sys := range allSystems {
		state1, has1 := snap1.Systems[sys]
		state2, has2 := snap2.Systems[sys]

		sysDiff := &SystemDiff{
			System:  sys,
			Changed: make([]FieldChange, 0),
		}

		if !has1 && has2 {
			sysDiff.Added = state2.State
			changedCount++
		} else if has1 && !has2 {
			sysDiff.Removed = state1.State
			changedCount++
		} else if has1 && has2 {
			// Compare states
			for key, val1 := range state1.State {
				val2, exists := state2.State[key]
				if !exists {
					if sysDiff.Removed == nil {
						sysDiff.Removed = make(map[string]interface{})
					}
					sysDiff.Removed[key] = val1
					changedCount++
				} else if fmt.Sprintf("%v", val1) != fmt.Sprintf("%v", val2) {
					sysDiff.Changed = append(sysDiff.Changed, FieldChange{
						Field:    key,
						OldValue: val1,
						NewValue: val2,
					})
					changedCount++
				}
			}
			for key, val2 := range state2.State {
				if _, exists := state1.State[key]; !exists {
					if sysDiff.Added == nil {
						sysDiff.Added = make(map[string]interface{})
					}
					sysDiff.Added[key] = val2
					changedCount++
				}
			}
		}

		if len(sysDiff.Added) > 0 || len(sysDiff.Removed) > 0 || len(sysDiff.Changed) > 0 {
			diff.Differences[sys] = sysDiff
		}
	}

	if changedCount == 0 {
		diff.Summary = "Snapshots are identical"
	} else {
		diff.Summary = fmt.Sprintf("%d differences across %d systems", changedCount, len(diff.Differences))
	}

	return diff, nil
}
