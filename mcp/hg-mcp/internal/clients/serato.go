// Package clients provides API clients for external services.
package clients

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
	"unicode/utf16"
)

// SeratoClient provides access to Serato DJ library
type SeratoClient struct {
	libraryPath string
}

// SeratoStatus represents Serato library status
type SeratoStatus struct {
	Connected    bool   `json:"connected"`
	LibraryPath  string `json:"library_path"`
	HasDatabase  bool   `json:"has_database"`
	CrateCount   int    `json:"crate_count"`
	HistoryCount int    `json:"history_count"`
}

// SeratoCrate represents a crate
type SeratoCrate struct {
	Name       string   `json:"name"`
	Path       string   `json:"path"`
	TrackCount int      `json:"track_count"`
	Tracks     []string `json:"tracks,omitempty"`
}

// SeratoTrack represents a track in the library
type SeratoTrack struct {
	Path     string `json:"path"`
	Filename string `json:"filename"`
	Title    string `json:"title,omitempty"`
	Artist   string `json:"artist,omitempty"`
	Album    string `json:"album,omitempty"`
	Genre    string `json:"genre,omitempty"`
	BPM      string `json:"bpm,omitempty"`
	Key      string `json:"key,omitempty"`
	Duration string `json:"duration,omitempty"`
}

// SeratoHistory represents a history session
type SeratoHistory struct {
	Name       string        `json:"name"`
	Path       string        `json:"path"`
	Date       string        `json:"date,omitempty"`
	Tracks     []SeratoTrack `json:"tracks,omitempty"`
	TrackCount int           `json:"track_count"`
}

// SeratoHealth represents health status
type SeratoHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Connected       bool     `json:"connected"`
	LibraryExists   bool     `json:"library_exists"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// NewSeratoClient creates a new Serato client
func NewSeratoClient() (*SeratoClient, error) {
	libraryPath := os.Getenv("SERATO_LIBRARY_PATH")
	if libraryPath == "" {
		libraryPath = getDefaultSeratoPath()
	}

	return &SeratoClient{
		libraryPath: libraryPath,
	}, nil
}

// getDefaultSeratoPath returns the default Serato library path
func getDefaultSeratoPath() string {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Music", "_Serato_")
	case "windows":
		return filepath.Join(home, "Music", "_Serato_")
	default:
		return filepath.Join(home, "Music", "_Serato_")
	}
}

// LibraryPath returns the configured library path
func (c *SeratoClient) LibraryPath() string {
	return c.libraryPath
}

// GetStatus returns library status
func (c *SeratoClient) GetStatus(ctx context.Context) (*SeratoStatus, error) {
	status := &SeratoStatus{
		Connected:   false,
		LibraryPath: c.libraryPath,
	}

	// Check if library exists
	if _, err := os.Stat(c.libraryPath); err == nil {
		status.Connected = true
	}

	// Check for database
	dbPath := filepath.Join(c.libraryPath, "database V2")
	if _, err := os.Stat(dbPath); err == nil {
		status.HasDatabase = true
	}

	// Count crates
	cratesPath := filepath.Join(c.libraryPath, "Subcrates")
	if entries, err := os.ReadDir(cratesPath); err == nil {
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".crate") {
				status.CrateCount++
			}
		}
	}

	// Count history sessions
	historyPath := filepath.Join(c.libraryPath, "History")
	if entries, err := os.ReadDir(historyPath); err == nil {
		for _, e := range entries {
			if strings.HasSuffix(e.Name(), ".session") {
				status.HistoryCount++
			}
		}
	}

	return status, nil
}

// GetCrates returns all crates
func (c *SeratoClient) GetCrates(ctx context.Context) ([]SeratoCrate, error) {
	cratesPath := filepath.Join(c.libraryPath, "Subcrates")
	entries, err := os.ReadDir(cratesPath)
	if err != nil {
		return nil, fmt.Errorf("reading crates directory: %w", err)
	}

	crates := make([]SeratoCrate, 0)
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".crate") {
			continue
		}

		cratePath := filepath.Join(cratesPath, e.Name())
		name := strings.TrimSuffix(e.Name(), ".crate")
		// Handle nested crates (%%% delimiter)
		name = strings.ReplaceAll(name, "%%%", "/")

		tracks, err := c.parseCrateFile(cratePath)
		if err != nil {
			continue
		}

		crates = append(crates, SeratoCrate{
			Name:       name,
			Path:       cratePath,
			TrackCount: len(tracks),
		})
	}

	return crates, nil
}

// GetCrate returns a specific crate with tracks
func (c *SeratoClient) GetCrate(ctx context.Context, name string) (*SeratoCrate, error) {
	// Convert path separators to Serato format
	filename := strings.ReplaceAll(name, "/", "%%%") + ".crate"
	cratePath := filepath.Join(c.libraryPath, "Subcrates", filename)

	if _, err := os.Stat(cratePath); err != nil {
		return nil, fmt.Errorf("crate not found: %s", name)
	}

	tracks, err := c.parseCrateFile(cratePath)
	if err != nil {
		return nil, fmt.Errorf("parsing crate: %w", err)
	}

	return &SeratoCrate{
		Name:       name,
		Path:       cratePath,
		TrackCount: len(tracks),
		Tracks:     tracks,
	}, nil
}

// GetHistory returns history sessions
func (c *SeratoClient) GetHistory(ctx context.Context, limit int) ([]SeratoHistory, error) {
	historyPath := filepath.Join(c.libraryPath, "History")
	entries, err := os.ReadDir(historyPath)
	if err != nil {
		return nil, fmt.Errorf("reading history directory: %w", err)
	}

	sessions := make([]SeratoHistory, 0)
	count := 0
	// Read in reverse order (newest first)
	for i := len(entries) - 1; i >= 0 && (limit <= 0 || count < limit); i-- {
		e := entries[i]
		if !strings.HasSuffix(e.Name(), ".session") {
			continue
		}

		sessionPath := filepath.Join(historyPath, e.Name())
		name := strings.TrimSuffix(e.Name(), ".session")

		// Parse date from filename if possible
		date := ""
		if len(name) >= 10 {
			date = name[:10]
		}

		tracks, err := c.parseSessionFile(sessionPath)
		if err != nil {
			continue
		}

		sessions = append(sessions, SeratoHistory{
			Name:       name,
			Path:       sessionPath,
			Date:       date,
			TrackCount: len(tracks),
		})
		count++
	}

	return sessions, nil
}

// GetHistorySession returns a specific history session with tracks
func (c *SeratoClient) GetHistorySession(ctx context.Context, name string) (*SeratoHistory, error) {
	sessionPath := filepath.Join(c.libraryPath, "History", name+".session")

	if _, err := os.Stat(sessionPath); err != nil {
		return nil, fmt.Errorf("session not found: %s", name)
	}

	tracks, err := c.parseSessionFile(sessionPath)
	if err != nil {
		return nil, fmt.Errorf("parsing session: %w", err)
	}

	date := ""
	if len(name) >= 10 {
		date = name[:10]
	}

	trackList := make([]SeratoTrack, len(tracks))
	for i, path := range tracks {
		trackList[i] = SeratoTrack{
			Path:     path,
			Filename: filepath.Base(path),
		}
	}

	return &SeratoHistory{
		Name:       name,
		Path:       sessionPath,
		Date:       date,
		Tracks:     trackList,
		TrackCount: len(tracks),
	}, nil
}

// SearchLibrary searches for tracks by query
func (c *SeratoClient) SearchLibrary(ctx context.Context, query string, limit int) ([]SeratoTrack, error) {
	// Search through all crates for matching tracks
	crates, err := c.GetCrates(ctx)
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	results := make([]SeratoTrack, 0)
	seen := make(map[string]bool)

	for _, crate := range crates {
		fullCrate, err := c.GetCrate(ctx, crate.Name)
		if err != nil {
			continue
		}

		for _, trackPath := range fullCrate.Tracks {
			if seen[trackPath] {
				continue
			}

			filename := strings.ToLower(filepath.Base(trackPath))
			if strings.Contains(filename, query) {
				results = append(results, SeratoTrack{
					Path:     trackPath,
					Filename: filepath.Base(trackPath),
				})
				seen[trackPath] = true

				if limit > 0 && len(results) >= limit {
					return results, nil
				}
			}
		}
	}

	return results, nil
}

// GetHealth returns health status
func (c *SeratoClient) GetHealth(ctx context.Context) (*SeratoHealth, error) {
	health := &SeratoHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check library exists
	if _, err := os.Stat(c.libraryPath); err != nil {
		health.Score -= 50
		health.LibraryExists = false
		health.Issues = append(health.Issues, fmt.Sprintf("Library not found at: %s", c.libraryPath))
		health.Recommendations = append(health.Recommendations, "Set SERATO_LIBRARY_PATH environment variable")
		health.Recommendations = append(health.Recommendations, "Ensure Serato DJ Pro has been run at least once")
	} else {
		health.LibraryExists = true
		health.Connected = true
	}

	// Check database
	dbPath := filepath.Join(c.libraryPath, "database V2")
	if _, err := os.Stat(dbPath); err != nil {
		health.Score -= 20
		health.Issues = append(health.Issues, "Database V2 not found")
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

// parseCrateFile parses a Serato crate file and returns track paths
func (c *SeratoClient) parseCrateFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return parseSeratoFile(data)
}

// parseSessionFile parses a Serato session file and returns track paths
func (c *SeratoClient) parseSessionFile(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return parseSeratoFile(data)
}

// parseSeratoFile parses Serato binary format (crate or session)
func parseSeratoFile(data []byte) ([]string, error) {
	tracks := make([]string, 0)
	reader := bytes.NewReader(data)

	for reader.Len() >= 8 {
		// Read 4-byte tag
		tag := make([]byte, 4)
		if _, err := reader.Read(tag); err != nil {
			break
		}

		// Read 4-byte length (big-endian)
		var length uint32
		if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
			break
		}

		if length == 0 || int(length) > reader.Len() {
			break
		}

		// Read field data
		fieldData := make([]byte, length)
		if _, err := reader.Read(fieldData); err != nil {
			break
		}

		tagStr := string(tag)

		// Handle nested track records
		if tagStr == "otrk" {
			// Parse nested fields
			nestedTracks := parseNestedFields(fieldData)
			tracks = append(tracks, nestedTracks...)
		} else if tagStr == "ptrk" {
			// Direct path field
			path := decodeUTF16BE(fieldData)
			if path != "" {
				tracks = append(tracks, path)
			}
		}
	}

	return tracks, nil
}

// parseNestedFields parses nested Serato fields
func parseNestedFields(data []byte) []string {
	tracks := make([]string, 0)
	reader := bytes.NewReader(data)

	for reader.Len() >= 8 {
		tag := make([]byte, 4)
		if _, err := reader.Read(tag); err != nil {
			break
		}

		var length uint32
		if err := binary.Read(reader, binary.BigEndian, &length); err != nil {
			break
		}

		if length == 0 || int(length) > reader.Len() {
			break
		}

		fieldData := make([]byte, length)
		if _, err := reader.Read(fieldData); err != nil {
			break
		}

		tagStr := string(tag)

		// Path fields start with 'p'
		if len(tagStr) > 0 && tagStr[0] == 'p' {
			path := decodeUTF16BE(fieldData)
			if path != "" {
				tracks = append(tracks, path)
			}
		}
	}

	return tracks
}

// decodeUTF16BE decodes UTF-16 big-endian bytes to string
func decodeUTF16BE(data []byte) string {
	if len(data) < 2 {
		return ""
	}

	// Convert to uint16 slice
	u16 := make([]uint16, len(data)/2)
	for i := 0; i < len(u16); i++ {
		u16[i] = binary.BigEndian.Uint16(data[i*2:])
	}

	// Decode UTF-16 to runes
	runes := utf16.Decode(u16)

	// Convert to string, removing null terminators
	result := string(runes)
	if idx := strings.IndexByte(result, 0); idx >= 0 {
		result = result[:idx]
	}

	return result
}

// GetNowPlaying attempts to get currently playing track (requires Serato Live or external integration)
func (c *SeratoClient) GetNowPlaying(ctx context.Context) (*SeratoTrack, error) {
	// Check for Serato Live history file (most recent)
	historyPath := filepath.Join(c.libraryPath, "History")
	entries, err := os.ReadDir(historyPath)
	if err != nil {
		return nil, fmt.Errorf("no history available")
	}

	// Find most recent session file modified today
	today := time.Now().Format("2006-01-02")
	var latestSession string
	var latestTime time.Time

	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".session") {
			continue
		}
		if !strings.HasPrefix(e.Name(), today) {
			continue
		}

		info, err := e.Info()
		if err != nil {
			continue
		}

		if info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latestSession = e.Name()
		}
	}

	if latestSession == "" {
		return nil, fmt.Errorf("no active session found")
	}

	// Get last track from session
	sessionPath := filepath.Join(historyPath, latestSession)
	tracks, err := c.parseSessionFile(sessionPath)
	if err != nil || len(tracks) == 0 {
		return nil, fmt.Errorf("no tracks in session")
	}

	lastTrack := tracks[len(tracks)-1]
	return &SeratoTrack{
		Path:     lastTrack,
		Filename: filepath.Base(lastTrack),
	}, nil
}
