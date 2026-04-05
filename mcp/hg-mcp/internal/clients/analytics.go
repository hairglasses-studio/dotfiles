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

// AnalyticsClient tracks performance metrics and show analytics
type AnalyticsClient struct {
	mu        sync.RWMutex
	sessions  map[string]*PerformanceSession
	configDir string
	current   *PerformanceSession
}

// PerformanceSession represents a DJ set or live performance session
type PerformanceSession struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time,omitempty"`
	Duration    time.Duration          `json:"duration,omitempty"`
	Venue       string                 `json:"venue,omitempty"`
	Type        string                 `json:"type"` // dj, live, hybrid
	Tracks      []*TrackPlay           `json:"tracks"`
	BPMHistory  []BPMDataPoint         `json:"bpm_history"`
	Transitions []*Transition          `json:"transitions"`
	Metrics     *SessionMetrics        `json:"metrics"`
	Tags        []string               `json:"tags,omitempty"`
	Notes       string                 `json:"notes,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TrackPlay represents a track played during a session
type TrackPlay struct {
	ID        string        `json:"id"`
	Title     string        `json:"title"`
	Artist    string        `json:"artist"`
	BPM       float64       `json:"bpm"`
	Key       string        `json:"key,omitempty"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time,omitempty"`
	Duration  time.Duration `json:"duration,omitempty"`
	Source    string        `json:"source"`           // rekordbox, serato, traktor, ableton
	Energy    int           `json:"energy,omitempty"` // 1-10
	Rating    int           `json:"rating,omitempty"` // 1-5 crowd response
}

// BPMDataPoint represents BPM at a point in time
type BPMDataPoint struct {
	Time time.Time `json:"time"`
	BPM  float64   `json:"bpm"`
}

// Transition represents a transition between tracks
type Transition struct {
	FromTrack string        `json:"from_track"`
	ToTrack   string        `json:"to_track"`
	Time      time.Time     `json:"time"`
	Duration  time.Duration `json:"duration"`
	Type      string        `json:"type"` // mix, cut, drop, loop
	BPMChange float64       `json:"bpm_change"`
	KeyChange string        `json:"key_change,omitempty"`
	Quality   int           `json:"quality,omitempty"` // 1-5 rating
}

// SessionMetrics contains aggregated session metrics
type SessionMetrics struct {
	TotalTracks       int           `json:"total_tracks"`
	TotalDuration     time.Duration `json:"total_duration"`
	AverageBPM        float64       `json:"average_bpm"`
	MinBPM            float64       `json:"min_bpm"`
	MaxBPM            float64       `json:"max_bpm"`
	BPMRange          float64       `json:"bpm_range"`
	AverageTrackTime  time.Duration `json:"average_track_time"`
	TransitionCount   int           `json:"transition_count"`
	UniqueArtists     int           `json:"unique_artists"`
	EnergyProgression []int         `json:"energy_progression,omitempty"`
	PeakTime          time.Time     `json:"peak_time,omitempty"`
}

// NewAnalyticsClient creates a new analytics client
func NewAnalyticsClient() (*AnalyticsClient, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "/tmp"
	}
	configDir := filepath.Join(homeDir, ".aftrs", "analytics")

	if err := os.MkdirAll(configDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	client := &AnalyticsClient{
		sessions:  make(map[string]*PerformanceSession),
		configDir: configDir,
	}

	if err := client.loadSessions(); err != nil {
		fmt.Printf("Warning: failed to load sessions: %v\n", err)
	}

	return client, nil
}

func (c *AnalyticsClient) loadSessions() error {
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

		var session PerformanceSession
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}

		c.sessions[session.ID] = &session
	}

	return nil
}

func (c *AnalyticsClient) saveSession(session *PerformanceSession) error {
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}

	filename := filepath.Join(c.configDir, session.ID+".json")
	return os.WriteFile(filename, data, 0644)
}

// StartSession starts a new performance session
func (c *AnalyticsClient) StartSession(ctx context.Context, name, venue, sessionType string) (*PerformanceSession, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.current != nil && c.current.EndTime.IsZero() {
		return nil, fmt.Errorf("session already in progress: %s", c.current.Name)
	}

	now := time.Now()
	id := fmt.Sprintf("session_%d", now.UnixNano())

	session := &PerformanceSession{
		ID:          id,
		Name:        name,
		StartTime:   now,
		Venue:       venue,
		Type:        sessionType,
		Tracks:      make([]*TrackPlay, 0),
		BPMHistory:  make([]BPMDataPoint, 0),
		Transitions: make([]*Transition, 0),
		Metrics:     &SessionMetrics{},
		Metadata:    make(map[string]interface{}),
	}

	c.sessions[id] = session
	c.current = session

	return session, nil
}

// EndSession ends the current session
func (c *AnalyticsClient) EndSession(ctx context.Context) (*PerformanceSession, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.current == nil {
		return nil, fmt.Errorf("no active session")
	}

	c.current.EndTime = time.Now()
	c.current.Duration = c.current.EndTime.Sub(c.current.StartTime)

	// Calculate final metrics
	c.calculateMetrics(c.current)

	if err := c.saveSession(c.current); err != nil {
		return c.current, fmt.Errorf("session ended but failed to save: %w", err)
	}

	session := c.current
	c.current = nil

	return session, nil
}

// LogTrack logs a track play
func (c *AnalyticsClient) LogTrack(ctx context.Context, title, artist string, bpm float64, key, source string) (*TrackPlay, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.current == nil {
		return nil, fmt.Errorf("no active session")
	}

	now := time.Now()

	// End previous track if any
	if len(c.current.Tracks) > 0 {
		lastTrack := c.current.Tracks[len(c.current.Tracks)-1]
		if lastTrack.EndTime.IsZero() {
			lastTrack.EndTime = now
			lastTrack.Duration = lastTrack.EndTime.Sub(lastTrack.StartTime)
		}
	}

	track := &TrackPlay{
		ID:        fmt.Sprintf("track_%d", now.UnixNano()),
		Title:     title,
		Artist:    artist,
		BPM:       bpm,
		Key:       key,
		StartTime: now,
		Source:    source,
	}

	c.current.Tracks = append(c.current.Tracks, track)

	// Log BPM
	c.current.BPMHistory = append(c.current.BPMHistory, BPMDataPoint{
		Time: now,
		BPM:  bpm,
	})

	// Auto-save
	_ = c.saveSession(c.current)

	return track, nil
}

// LogTransition logs a transition between tracks
func (c *AnalyticsClient) LogTransition(ctx context.Context, transType string, duration time.Duration, quality int) (*Transition, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.current == nil {
		return nil, fmt.Errorf("no active session")
	}

	if len(c.current.Tracks) < 2 {
		return nil, fmt.Errorf("need at least 2 tracks for a transition")
	}

	tracks := c.current.Tracks
	fromTrack := tracks[len(tracks)-2]
	toTrack := tracks[len(tracks)-1]

	transition := &Transition{
		FromTrack: fromTrack.Title,
		ToTrack:   toTrack.Title,
		Time:      time.Now(),
		Duration:  duration,
		Type:      transType,
		BPMChange: toTrack.BPM - fromTrack.BPM,
		Quality:   quality,
	}

	if fromTrack.Key != "" && toTrack.Key != "" {
		transition.KeyChange = fmt.Sprintf("%s → %s", fromTrack.Key, toTrack.Key)
	}

	c.current.Transitions = append(c.current.Transitions, transition)

	return transition, nil
}

// LogBPM logs current BPM
func (c *AnalyticsClient) LogBPM(ctx context.Context, bpm float64) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.current == nil {
		return fmt.Errorf("no active session")
	}

	c.current.BPMHistory = append(c.current.BPMHistory, BPMDataPoint{
		Time: time.Now(),
		BPM:  bpm,
	})

	return nil
}

// GetCurrentSession returns the current active session
func (c *AnalyticsClient) GetCurrentSession(ctx context.Context) (*PerformanceSession, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.current == nil {
		return nil, fmt.Errorf("no active session")
	}

	// Calculate live metrics
	session := *c.current
	c.calculateMetrics(&session)

	return &session, nil
}

// GetSession returns a session by ID
func (c *AnalyticsClient) GetSession(ctx context.Context, id string) (*PerformanceSession, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	session, exists := c.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", id)
	}

	return session, nil
}

// ListSessions returns all sessions
func (c *AnalyticsClient) ListSessions(ctx context.Context) []*PerformanceSession {
	c.mu.RLock()
	defer c.mu.RUnlock()

	sessions := make([]*PerformanceSession, 0, len(c.sessions))
	for _, s := range c.sessions {
		sessions = append(sessions, s)
	}

	// Sort by start time (newest first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartTime.After(sessions[j].StartTime)
	})

	return sessions
}

// GetAnalytics returns analytics across all sessions
func (c *AnalyticsClient) GetAnalytics(ctx context.Context) (*OverallAnalytics, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	analytics := &OverallAnalytics{
		TotalSessions:   len(c.sessions),
		TrackFrequency:  make(map[string]int),
		ArtistFrequency: make(map[string]int),
		VenueStats:      make(map[string]*VenueStats),
	}

	var totalDuration time.Duration
	var totalTracks int
	var totalBPM float64
	var bpmCount int

	for _, session := range c.sessions {
		if !session.EndTime.IsZero() {
			totalDuration += session.Duration
		}
		totalTracks += len(session.Tracks)

		for _, track := range session.Tracks {
			analytics.TrackFrequency[track.Title]++
			analytics.ArtistFrequency[track.Artist]++
			totalBPM += track.BPM
			bpmCount++
		}

		if session.Venue != "" {
			if _, exists := analytics.VenueStats[session.Venue]; !exists {
				analytics.VenueStats[session.Venue] = &VenueStats{Name: session.Venue}
			}
			analytics.VenueStats[session.Venue].SessionCount++
			analytics.VenueStats[session.Venue].TotalDuration += session.Duration
		}
	}

	analytics.TotalPlaytime = totalDuration
	analytics.TotalTracks = totalTracks
	if bpmCount > 0 {
		analytics.AverageBPM = totalBPM / float64(bpmCount)
	}

	// Find most played
	var maxPlays int
	for track, plays := range analytics.TrackFrequency {
		if plays > maxPlays {
			maxPlays = plays
			analytics.MostPlayedTrack = track
		}
	}

	maxPlays = 0
	for artist, plays := range analytics.ArtistFrequency {
		if plays > maxPlays {
			maxPlays = plays
			analytics.MostPlayedArtist = artist
		}
	}

	return analytics, nil
}

// OverallAnalytics contains aggregated analytics
type OverallAnalytics struct {
	TotalSessions    int                    `json:"total_sessions"`
	TotalPlaytime    time.Duration          `json:"total_playtime"`
	TotalTracks      int                    `json:"total_tracks"`
	AverageBPM       float64                `json:"average_bpm"`
	MostPlayedTrack  string                 `json:"most_played_track"`
	MostPlayedArtist string                 `json:"most_played_artist"`
	TrackFrequency   map[string]int         `json:"track_frequency"`
	ArtistFrequency  map[string]int         `json:"artist_frequency"`
	VenueStats       map[string]*VenueStats `json:"venue_stats"`
}

// VenueStats contains venue-specific statistics
type VenueStats struct {
	Name          string        `json:"name"`
	SessionCount  int           `json:"session_count"`
	TotalDuration time.Duration `json:"total_duration"`
}

// calculateMetrics calculates session metrics
func (c *AnalyticsClient) calculateMetrics(session *PerformanceSession) {
	if session.Metrics == nil {
		session.Metrics = &SessionMetrics{}
	}

	metrics := session.Metrics
	metrics.TotalTracks = len(session.Tracks)
	metrics.TransitionCount = len(session.Transitions)

	if metrics.TotalTracks == 0 {
		return
	}

	// BPM stats
	var totalBPM float64
	metrics.MinBPM = 999
	metrics.MaxBPM = 0

	artistSet := make(map[string]bool)
	var totalTrackDuration time.Duration

	for _, track := range session.Tracks {
		totalBPM += track.BPM
		if track.BPM < metrics.MinBPM {
			metrics.MinBPM = track.BPM
		}
		if track.BPM > metrics.MaxBPM {
			metrics.MaxBPM = track.BPM
		}
		artistSet[track.Artist] = true
		if track.Duration > 0 {
			totalTrackDuration += track.Duration
		}
		if track.Energy > 0 {
			metrics.EnergyProgression = append(metrics.EnergyProgression, track.Energy)
		}
	}

	metrics.AverageBPM = totalBPM / float64(metrics.TotalTracks)
	metrics.BPMRange = metrics.MaxBPM - metrics.MinBPM
	metrics.UniqueArtists = len(artistSet)

	if metrics.TotalTracks > 0 {
		metrics.AverageTrackTime = totalTrackDuration / time.Duration(metrics.TotalTracks)
	}

	// Calculate duration
	if !session.EndTime.IsZero() {
		metrics.TotalDuration = session.EndTime.Sub(session.StartTime)
	} else {
		metrics.TotalDuration = time.Since(session.StartTime)
	}
}

// DeleteSession removes a session
func (c *AnalyticsClient) DeleteSession(ctx context.Context, id string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.sessions[id]; !exists {
		return fmt.Errorf("session not found: %s", id)
	}

	delete(c.sessions, id)

	filename := filepath.Join(c.configDir, id+".json")
	_ = os.Remove(filename)

	return nil
}
