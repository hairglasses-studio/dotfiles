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

// SetlistClient manages DJ setlists and track planning
type SetlistClient struct {
	mu        sync.RWMutex
	setlists  map[string]*DJSetlist
	configDir string
}

// DJSetlist represents a planned set of tracks
type DJSetlist struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Venue       string                 `json:"venue,omitempty"`
	Date        time.Time              `json:"date,omitempty"`
	Duration    time.Duration          `json:"duration,omitempty"` // Target duration
	Tracks      []*SetlistTrack        `json:"tracks"`
	Tags        []string               `json:"tags,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// SetlistTrack represents a track in a setlist
type SetlistTrack struct {
	Position  int                    `json:"position"`
	Title     string                 `json:"title"`
	Artist    string                 `json:"artist"`
	BPM       float64                `json:"bpm,omitempty"`
	Key       string                 `json:"key,omitempty"`
	Duration  time.Duration          `json:"duration,omitempty"`
	Energy    int                    `json:"energy,omitempty"`    // 1-10
	Notes     string                 `json:"notes,omitempty"`     // Transition notes
	Source    string                 `json:"source,omitempty"`    // rekordbox, serato, etc
	SourceID  string                 `json:"source_id,omitempty"` // ID in source system
	CuePoints []SetlistCuePoint      `json:"cue_points,omitempty"`
	Played    bool                   `json:"played"`
	PlayedAt  time.Time              `json:"played_at,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// SetlistCuePoint represents a cue point in a setlist track
type SetlistCuePoint struct {
	Name     string        `json:"name"`
	Position time.Duration `json:"position"`
	Color    string        `json:"color,omitempty"`
}

// SetlistAnalysis provides analysis of a setlist
type SetlistAnalysis struct {
	TotalTracks    int           `json:"total_tracks"`
	TotalDuration  time.Duration `json:"total_duration"`
	AverageBPM     float64       `json:"average_bpm"`
	BPMRange       string        `json:"bpm_range"`
	MinBPM         float64       `json:"min_bpm"`
	MaxBPM         float64       `json:"max_bpm"`
	EnergyFlow     []int         `json:"energy_flow"`
	KeyProgression []string      `json:"key_progression"`
	UniqueArtists  int           `json:"unique_artists"`
	LargeBPMJumps  []string      `json:"large_bpm_jumps,omitempty"` // Transitions >10 BPM
	KeyClashes     []string      `json:"key_clashes,omitempty"`     // Non-harmonic transitions
}

// NewSetlistClient creates a new setlist client
func NewSetlistClient() (*SetlistClient, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/tmp"
	}
	configDir := filepath.Join(homeDir, ".aftrs", "setlists")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	client := &SetlistClient{
		setlists:  make(map[string]*DJSetlist),
		configDir: configDir,
	}

	if err := client.loadSetlists(); err != nil {
		fmt.Printf("Warning: failed to load setlists: %v\n", err)
	}

	return client, nil
}

func (c *SetlistClient) loadSetlists() error {
	files, err := os.ReadDir(c.configDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		data, err := os.ReadFile(filepath.Join(c.configDir, file.Name()))
		if err != nil {
			continue
		}

		var setlist DJSetlist
		if err := json.Unmarshal(data, &setlist); err != nil {
			continue
		}

		c.setlists[setlist.ID] = &setlist
	}

	return nil
}

func (c *SetlistClient) saveSetlist(setlist *DJSetlist) error {
	data, err := json.MarshalIndent(setlist, "", "  ")
	if err != nil {
		return err
	}

	filename := filepath.Join(c.configDir, setlist.ID+".json")
	return os.WriteFile(filename, data, 0644)
}

// CreateSetlist creates a new setlist
func (c *SetlistClient) CreateSetlist(ctx context.Context, name, description, venue string, date time.Time) (*DJSetlist, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	id := fmt.Sprintf("setlist_%d", now.UnixNano())

	setlist := &DJSetlist{
		ID:          id,
		Name:        name,
		Description: description,
		Venue:       venue,
		Date:        date,
		Tracks:      make([]*SetlistTrack, 0),
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    make(map[string]interface{}),
	}

	c.setlists[id] = setlist

	if err := c.saveSetlist(setlist); err != nil {
		return setlist, fmt.Errorf("setlist created but failed to save: %w", err)
	}

	return setlist, nil
}

// GetSetlist returns a setlist by ID
func (c *SetlistClient) GetSetlist(ctx context.Context, id string) (*DJSetlist, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	setlist, exists := c.setlists[id]
	if !exists {
		return nil, fmt.Errorf("setlist not found: %s", id)
	}

	return setlist, nil
}

// ListSetlists returns all setlists
func (c *SetlistClient) ListSetlists(ctx context.Context) []*DJSetlist {
	c.mu.RLock()
	defer c.mu.RUnlock()

	setlists := make([]*DJSetlist, 0, len(c.setlists))
	for _, s := range c.setlists {
		setlists = append(setlists, s)
	}

	// Sort by date (newest first)
	sort.Slice(setlists, func(i, j int) bool {
		return setlists[i].UpdatedAt.After(setlists[j].UpdatedAt)
	})

	return setlists
}

// AddTrack adds a track to a setlist
func (c *SetlistClient) AddTrack(ctx context.Context, setlistID string, track *SetlistTrack) (*SetlistTrack, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	setlist, exists := c.setlists[setlistID]
	if !exists {
		return nil, fmt.Errorf("setlist not found: %s", setlistID)
	}

	// Set position to end of list
	track.Position = len(setlist.Tracks) + 1
	setlist.Tracks = append(setlist.Tracks, track)
	setlist.UpdatedAt = time.Now()

	if err := c.saveSetlist(setlist); err != nil {
		return track, fmt.Errorf("track added but failed to save: %w", err)
	}

	return track, nil
}

// RemoveTrack removes a track from a setlist by position
func (c *SetlistClient) RemoveTrack(ctx context.Context, setlistID string, position int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	setlist, exists := c.setlists[setlistID]
	if !exists {
		return fmt.Errorf("setlist not found: %s", setlistID)
	}

	if position < 1 || position > len(setlist.Tracks) {
		return fmt.Errorf("invalid position: %d", position)
	}

	// Remove track at position (1-indexed)
	idx := position - 1
	setlist.Tracks = append(setlist.Tracks[:idx], setlist.Tracks[idx+1:]...)

	// Re-number positions
	for i := range setlist.Tracks {
		setlist.Tracks[i].Position = i + 1
	}

	setlist.UpdatedAt = time.Now()

	return c.saveSetlist(setlist)
}

// ReorderTrack moves a track to a new position
func (c *SetlistClient) ReorderTrack(ctx context.Context, setlistID string, fromPos, toPos int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	setlist, exists := c.setlists[setlistID]
	if !exists {
		return fmt.Errorf("setlist not found: %s", setlistID)
	}

	if fromPos < 1 || fromPos > len(setlist.Tracks) {
		return fmt.Errorf("invalid from position: %d", fromPos)
	}
	if toPos < 1 || toPos > len(setlist.Tracks) {
		return fmt.Errorf("invalid to position: %d", toPos)
	}

	// Extract track
	fromIdx := fromPos - 1
	track := setlist.Tracks[fromIdx]

	// Remove from old position
	setlist.Tracks = append(setlist.Tracks[:fromIdx], setlist.Tracks[fromIdx+1:]...)

	// Insert at new position
	toIdx := toPos - 1
	if toIdx > len(setlist.Tracks) {
		toIdx = len(setlist.Tracks)
	}
	setlist.Tracks = append(setlist.Tracks[:toIdx], append([]*SetlistTrack{track}, setlist.Tracks[toIdx:]...)...)

	// Re-number positions
	for i := range setlist.Tracks {
		setlist.Tracks[i].Position = i + 1
	}

	setlist.UpdatedAt = time.Now()

	return c.saveSetlist(setlist)
}

// AnalyzeSetlist analyzes a setlist for flow and compatibility
func (c *SetlistClient) AnalyzeSetlist(ctx context.Context, setlistID string) (*SetlistAnalysis, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	setlist, exists := c.setlists[setlistID]
	if !exists {
		return nil, fmt.Errorf("setlist not found: %s", setlistID)
	}

	analysis := &SetlistAnalysis{
		TotalTracks:    len(setlist.Tracks),
		EnergyFlow:     make([]int, 0),
		KeyProgression: make([]string, 0),
		LargeBPMJumps:  make([]string, 0),
		KeyClashes:     make([]string, 0),
	}

	if len(setlist.Tracks) == 0 {
		return analysis, nil
	}

	var totalBPM float64
	var totalDuration time.Duration
	artistSet := make(map[string]bool)
	analysis.MinBPM = 999
	analysis.MaxBPM = 0

	var prevTrack *SetlistTrack
	for _, track := range setlist.Tracks {
		// BPM stats
		if track.BPM > 0 {
			totalBPM += track.BPM
			if track.BPM < analysis.MinBPM {
				analysis.MinBPM = track.BPM
			}
			if track.BPM > analysis.MaxBPM {
				analysis.MaxBPM = track.BPM
			}
		}

		// Duration
		totalDuration += track.Duration

		// Artists
		artistSet[track.Artist] = true

		// Energy flow
		if track.Energy > 0 {
			analysis.EnergyFlow = append(analysis.EnergyFlow, track.Energy)
		}

		// Key progression
		if track.Key != "" {
			analysis.KeyProgression = append(analysis.KeyProgression, track.Key)
		}

		// Transition analysis
		if prevTrack != nil {
			// Large BPM jumps
			if prevTrack.BPM > 0 && track.BPM > 0 {
				bpmDiff := track.BPM - prevTrack.BPM
				if bpmDiff > 10 || bpmDiff < -10 {
					analysis.LargeBPMJumps = append(analysis.LargeBPMJumps,
						fmt.Sprintf("%d→%d: %.1f BPM jump", prevTrack.Position, track.Position, bpmDiff))
				}
			}

			// Key clashes (simplified - just flag different keys for now)
			if prevTrack.Key != "" && track.Key != "" && !isHarmonicKey(prevTrack.Key, track.Key) {
				analysis.KeyClashes = append(analysis.KeyClashes,
					fmt.Sprintf("%d→%d: %s → %s", prevTrack.Position, track.Position, prevTrack.Key, track.Key))
			}
		}

		prevTrack = track
	}

	analysis.TotalDuration = totalDuration
	analysis.UniqueArtists = len(artistSet)
	if len(setlist.Tracks) > 0 {
		analysis.AverageBPM = totalBPM / float64(len(setlist.Tracks))
	}
	if analysis.MinBPM < 999 {
		analysis.BPMRange = fmt.Sprintf("%.1f - %.1f", analysis.MinBPM, analysis.MaxBPM)
	}

	return analysis, nil
}

// isHarmonicKey checks if two keys are harmonically compatible
// Uses Camelot wheel logic for DJ mixing
func isHarmonicKey(key1, key2 string) bool {
	// Simplified: same key or adjacent on Camelot wheel
	// Full implementation would parse key notation and check wheel position
	if key1 == key2 {
		return true
	}
	// For now, allow any transition - full Camelot logic would go here
	return true
}

// DeleteSetlist removes a setlist
func (c *SetlistClient) DeleteSetlist(ctx context.Context, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.setlists[id]; !exists {
		return fmt.Errorf("setlist not found: %s", id)
	}

	delete(c.setlists, id)

	filename := filepath.Join(c.configDir, id+".json")
	_ = os.Remove(filename)

	return nil
}

// ExportSetlist exports a setlist to a format
func (c *SetlistClient) ExportSetlist(ctx context.Context, setlistID, format string) (string, error) {
	c.mu.RLock()
	setlist, exists := c.setlists[setlistID]
	c.mu.RUnlock()

	if !exists {
		return "", fmt.Errorf("setlist not found: %s", setlistID)
	}

	switch format {
	case "json":
		data, err := json.MarshalIndent(setlist, "", "  ")
		if err != nil {
			return "", err
		}
		return string(data), nil

	case "m3u":
		return c.exportM3U(setlist), nil

	case "text", "txt":
		return c.exportText(setlist), nil

	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}

func (c *SetlistClient) exportM3U(setlist *DJSetlist) string {
	var result string
	result += "#EXTM3U\n"
	result += fmt.Sprintf("#PLAYLIST:%s\n", setlist.Name)

	for _, track := range setlist.Tracks {
		duration := int(track.Duration.Seconds())
		result += fmt.Sprintf("#EXTINF:%d,%s - %s\n", duration, track.Artist, track.Title)
		if track.SourceID != "" {
			result += track.SourceID + "\n"
		}
	}

	return result
}

func (c *SetlistClient) exportText(setlist *DJSetlist) string {
	var result string
	result += fmt.Sprintf("# %s\n", setlist.Name)
	if setlist.Venue != "" {
		result += fmt.Sprintf("Venue: %s\n", setlist.Venue)
	}
	if !setlist.Date.IsZero() {
		result += fmt.Sprintf("Date: %s\n", setlist.Date.Format("2006-01-02"))
	}
	result += "\n"

	for _, track := range setlist.Tracks {
		bpm := ""
		if track.BPM > 0 {
			bpm = fmt.Sprintf(" [%.1f BPM]", track.BPM)
		}
		key := ""
		if track.Key != "" {
			key = fmt.Sprintf(" [%s]", track.Key)
		}
		result += fmt.Sprintf("%d. %s - %s%s%s\n", track.Position, track.Artist, track.Title, bpm, key)
		if track.Notes != "" {
			result += fmt.Sprintf("   → %s\n", track.Notes)
		}
	}

	return result
}

// MarkTrackPlayed marks a track as played
func (c *SetlistClient) MarkTrackPlayed(ctx context.Context, setlistID string, position int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	setlist, exists := c.setlists[setlistID]
	if !exists {
		return fmt.Errorf("setlist not found: %s", setlistID)
	}

	if position < 1 || position > len(setlist.Tracks) {
		return fmt.Errorf("invalid position: %d", position)
	}

	setlist.Tracks[position-1].Played = true
	setlist.Tracks[position-1].PlayedAt = time.Now()
	setlist.UpdatedAt = time.Now()

	return c.saveSetlist(setlist)
}
