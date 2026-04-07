// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

// RekordboxClient provides access to Rekordbox DJ library via Python scripts
type RekordboxClient struct {
	databasePath string
	musicPath    string
	playlistPath string
	pythonPath   string
	scriptsPath  string
}

// RekordboxStatus represents Rekordbox library status
type RekordboxStatus struct {
	Connected          bool   `json:"connected"`
	DatabasePath       string `json:"database_path"`
	TrackCount         int    `json:"track_count"`
	PlaylistCount      int    `json:"playlist_count"`
	AnalysisProgress   string `json:"analysis_progress"`
	AnalyzedCount      int    `json:"analyzed_count"`
	DuplicatePlaylists int    `json:"duplicate_playlists"`
	IsRunning          bool   `json:"is_running"`
}

// RekordboxPlaylist represents a playlist in Rekordbox
type RekordboxPlaylist struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	TrackCount  int    `json:"track_count"`
	IsDuplicate bool   `json:"is_duplicate"`
}

// RekordboxTrack represents a track in Rekordbox
type RekordboxTrack struct {
	ID         int     `json:"id"`
	Title      string  `json:"title"`
	Artist     string  `json:"artist"`
	Album      string  `json:"album,omitempty"`
	BPM        float64 `json:"bpm,omitempty"`
	Key        string  `json:"key,omitempty"`
	Path       string  `json:"path"`
	Filename   string  `json:"filename,omitempty"`
	DurationMS int     `json:"duration_ms,omitempty"`
	Analyzed   bool    `json:"analyzed"`
	Timestamp  string  `json:"timestamp,omitempty"`
	Source     string  `json:"source,omitempty"`
}

// RekordboxSession represents a DJ session/history entry
type RekordboxSession struct {
	ID       int    `json:"id"`
	Date     string `json:"date"`
	Tracks   int    `json:"tracks"`
	Duration string `json:"duration,omitempty"`
}

// RekordboxHistoryEntry represents a track played in history
type RekordboxHistoryEntry struct {
	ID       int    `json:"id"`
	TrackID  int    `json:"track_id,omitempty"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	PlayedAt string `json:"played_at,omitempty"`
	Path     string `json:"path,omitempty"`
}

// RekordboxUSBInfo represents USB drive info for Rekordbox
type RekordboxUSBInfo struct {
	Connected     bool    `json:"connected"`
	DriveLetter   string  `json:"drive_letter"`
	AudioFiles    int     `json:"audio_files"`
	AnalysisFiles int     `json:"analysis_files"`
	TotalSizeGB   float64 `json:"total_size_gb"`
}

// RekordboxHealth represents health status
type RekordboxHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	Connected       bool     `json:"connected"`
	DatabaseExists  bool     `json:"database_exists"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// MaintenanceResult represents the result of a maintenance operation
type MaintenanceResult struct {
	Operation string `json:"operation"`
	Success   bool   `json:"success"`
	Details   string `json:"details"`
	Count     int    `json:"count,omitempty"`
}

var (
	rekordboxClientSingleton *RekordboxClient
	rekordboxClientOnce      sync.Once
	rekordboxClientErr       error

	// TestOverrideRekordboxClient, when non-nil, is returned by GetRekordboxClient
	// instead of the real singleton. Used for testing without network access.
	TestOverrideRekordboxClient *RekordboxClient
)

// GetRekordboxClient returns the singleton Rekordbox client.
func GetRekordboxClient() (*RekordboxClient, error) {
	if TestOverrideRekordboxClient != nil {
		return TestOverrideRekordboxClient, nil
	}
	rekordboxClientOnce.Do(func() {
		rekordboxClientSingleton, rekordboxClientErr = NewRekordboxClient()
	})
	return rekordboxClientSingleton, rekordboxClientErr
}

// NewTestRekordboxClient creates an in-memory test client.
func NewTestRekordboxClient() *RekordboxClient {
	return &RekordboxClient{
		databasePath: "/tmp/test/rekordbox/master.db",
		musicPath:    "/tmp/test/rekordbox/music",
		playlistPath: "/tmp/test/rekordbox/playlists",
		pythonPath:   "python3",
		scriptsPath:  "/tmp/test/scripts",
	}
}

// NewRekordboxClient creates a new Rekordbox client
func NewRekordboxClient() (*RekordboxClient, error) {
	home, _ := os.UserHomeDir()

	databasePath := os.Getenv("REKORDBOX_DB_PATH")
	if databasePath == "" {
		databasePath = getDefaultRekordboxDBPath()
	}

	musicPath := os.Getenv("REKORDBOX_MUSIC_PATH")
	if musicPath == "" {
		musicPath = filepath.Join(home, "Music", "rekordbox")
	}

	playlistPath := os.Getenv("REKORDBOX_PLAYLIST_PATH")
	if playlistPath == "" {
		playlistPath = filepath.Join(musicPath, "playlists")
	}

	// Find Python - check common locations
	pythonPath := findPython()

	// Find scripts path
	scriptsPath := os.Getenv("CR8_SCRIPTS_PATH")
	if scriptsPath == "" {
		scriptsPath = filepath.Join(home, "Documents", "hairglasses-studio", "cr8-cli", "scripts")
	}

	return &RekordboxClient{
		databasePath: databasePath,
		musicPath:    musicPath,
		playlistPath: playlistPath,
		pythonPath:   pythonPath,
		scriptsPath:  scriptsPath,
	}, nil
}

// getDefaultRekordboxDBPath returns the default Rekordbox database path
func getDefaultRekordboxDBPath() string {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Pioneer", "rekordbox", "master.db")
	case "windows":
		return filepath.Join(home, "AppData", "Roaming", "Pioneer", "rekordbox", "master.db")
	default:
		return filepath.Join(home, ".pioneer", "rekordbox", "master.db")
	}
}

// findPython finds a working Python installation
func findPython() string {
	// Check environment variable first
	if p := os.Getenv("PYTHON_PATH"); p != "" {
		return p
	}

	// Try common locations
	candidates := []string{
		"python",
		"python3",
		"py",
	}

	for _, p := range candidates {
		if _, err := exec.LookPath(p); err == nil {
			return p
		}
	}

	return "python"
}

// GetStatus returns the current Rekordbox library status
func (c *RekordboxClient) GetStatus(ctx context.Context) (*RekordboxStatus, error) {
	script := filepath.Join(c.scriptsPath, "phase3_maintenance.py")

	// Run the status check script
	cmd := exec.CommandContext(ctx, c.pythonPath, script, "--status")
	cmd.Dir = filepath.Dir(c.scriptsPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Parse output even on error for status info
		status := c.parseStatusOutput(string(output))
		if status != nil {
			return status, nil
		}
		return nil, fmt.Errorf("failed to get status: %w\nOutput: %s", err, string(output))
	}

	return c.parseStatusOutput(string(output)), nil
}

// parseStatusOutput parses the status output from the Python script
func (c *RekordboxClient) parseStatusOutput(output string) *RekordboxStatus {
	status := &RekordboxStatus{
		DatabasePath: c.databasePath,
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Total Tracks:") {
			fmt.Sscanf(line, "Total Tracks: %d", &status.TrackCount)
			status.Connected = true
		} else if strings.HasPrefix(line, "Total Playlists:") {
			fmt.Sscanf(line, "Total Playlists: %d", &status.PlaylistCount)
		} else if strings.HasPrefix(line, "Duplicate Playlists:") {
			fmt.Sscanf(line, "Duplicate Playlists: %d", &status.DuplicatePlaylists)
		} else if strings.HasPrefix(line, "Analysis Progress:") {
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				status.AnalysisProgress = strings.TrimSpace(parts[1])
			}
		}
	}

	// Check if Rekordbox is running
	status.IsRunning = c.isRekordboxRunning()

	return status
}

// GetPlaylists returns all playlists in Rekordbox
func (c *RekordboxClient) GetPlaylists(ctx context.Context) ([]RekordboxPlaylist, error) {
	script := filepath.Join(c.scriptsPath, "cleanup_duplicate_playlists.py")

	cmd := exec.CommandContext(ctx, c.pythonPath, script)
	cmd.Dir = filepath.Dir(c.scriptsPath)

	output, err := cmd.CombinedOutput()
	if err != nil && !strings.Contains(string(output), "DUPLICATE PLAYLISTS") {
		return nil, fmt.Errorf("failed to get playlists: %w", err)
	}

	return c.parsePlaylistOutput(string(output)), nil
}

// parsePlaylistOutput parses playlist info from script output
func (c *RekordboxClient) parsePlaylistOutput(output string) []RekordboxPlaylist {
	var playlists []RekordboxPlaylist

	lines := strings.Split(output, "\n")
	inDuplicates := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "DUPLICATE PLAYLISTS") {
			inDuplicates = true
			continue
		}

		if inDuplicates && strings.HasPrefix(line, "-") {
			name := strings.TrimPrefix(line, "- ")
			name = strings.Trim(name, "'")
			if idx := strings.Index(name, "' ("); idx > 0 {
				name = name[:idx]
			}
			playlists = append(playlists, RekordboxPlaylist{
				Name:        name,
				IsDuplicate: strings.HasSuffix(name, " (1)"),
			})
		}
	}

	return playlists
}

// GetUSBInfo returns info about connected USB drive for Rekordbox
func (c *RekordboxClient) GetUSBInfo(ctx context.Context) (*RekordboxUSBInfo, error) {
	script := filepath.Join(c.scriptsPath, "import_usb_rekordbox.py")

	cmd := exec.CommandContext(ctx, c.pythonPath, script)
	cmd.Dir = filepath.Dir(c.scriptsPath)

	output, err := cmd.CombinedOutput()
	if err != nil && !strings.Contains(string(output), "USB DRIVE CONTENTS") {
		return &RekordboxUSBInfo{Connected: false}, nil
	}

	return c.parseUSBOutput(string(output)), nil
}

// parseUSBOutput parses USB info from script output
func (c *RekordboxClient) parseUSBOutput(output string) *RekordboxUSBInfo {
	info := &RekordboxUSBInfo{}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Drive:") {
			info.DriveLetter = strings.TrimPrefix(line, "Drive: ")
			info.Connected = true
		} else if strings.HasPrefix(line, "Audio files:") {
			fmt.Sscanf(line, "Audio files: %d", &info.AudioFiles)
		} else if strings.HasPrefix(line, "Analysis files:") {
			fmt.Sscanf(line, "Analysis files: %d", &info.AnalysisFiles)
		} else if strings.HasPrefix(line, "Total audio size:") {
			fmt.Sscanf(line, "Total audio size: %f GB", &info.TotalSizeGB)
		}
	}

	return info
}

// ImportUSB imports tracks from USB to local Rekordbox library
func (c *RekordboxClient) ImportUSB(ctx context.Context, copyAnalysis bool) (*MaintenanceResult, error) {
	script := filepath.Join(c.scriptsPath, "import_usb_rekordbox.py")

	args := []string{script, "--all"}
	if copyAnalysis {
		args = append(args, "--copy-anlz")
	}

	cmd := exec.CommandContext(ctx, c.pythonPath, args...)
	cmd.Dir = filepath.Dir(c.scriptsPath)

	output, err := cmd.CombinedOutput()
	result := &MaintenanceResult{
		Operation: "usb_import",
		Success:   err == nil,
		Details:   string(output),
	}

	// Parse count from output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "copied") {
			fmt.Sscanf(line, "Copy complete: %d copied", &result.Count)
		}
	}

	return result, err
}

// CleanupDuplicates removes duplicate playlists
func (c *RekordboxClient) CleanupDuplicates(ctx context.Context, useUI bool) (*MaintenanceResult, error) {
	script := filepath.Join(c.scriptsPath, "cleanup_duplicate_playlists.py")

	args := []string{script, "--backup", "--delete"}
	if useUI {
		args = append(args, "--ui")
	} else {
		args = append(args, "--db")
	}

	cmd := exec.CommandContext(ctx, c.pythonPath, args...)
	cmd.Dir = filepath.Dir(c.scriptsPath)

	output, err := cmd.CombinedOutput()
	result := &MaintenanceResult{
		Operation: "cleanup_duplicates",
		Success:   err == nil,
		Details:   string(output),
	}

	// Parse count from output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Deleted") {
			fmt.Sscanf(line, "CLEANUP COMPLETE: Deleted %d", &result.Count)
		}
	}

	return result, err
}

// TriggerAnalysis triggers track analysis in Rekordbox
func (c *RekordboxClient) TriggerAnalysis(ctx context.Context) (*MaintenanceResult, error) {
	script := filepath.Join(c.scriptsPath, "trigger_analysis.py")

	cmd := exec.CommandContext(ctx, c.pythonPath, script)
	cmd.Dir = filepath.Dir(c.scriptsPath)

	output, err := cmd.CombinedOutput()
	result := &MaintenanceResult{
		Operation: "trigger_analysis",
		Success:   err == nil,
		Details:   string(output),
	}

	return result, err
}

// GetHealth returns health status for Rekordbox integration
func (c *RekordboxClient) GetHealth() *RekordboxHealth {
	health := &RekordboxHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check database exists
	if _, err := os.Stat(c.databasePath); os.IsNotExist(err) {
		health.Score -= 50
		health.Issues = append(health.Issues, "Rekordbox database not found")
		health.Recommendations = append(health.Recommendations, "Run Rekordbox at least once to create database")
	} else {
		health.DatabaseExists = true
	}

	// Check scripts exist
	if _, err := os.Stat(c.scriptsPath); os.IsNotExist(err) {
		health.Score -= 30
		health.Issues = append(health.Issues, "CR8-CLI scripts not found")
		health.Recommendations = append(health.Recommendations, "Install cr8-cli scripts in "+c.scriptsPath)
	}

	// Check Python
	if _, err := exec.LookPath(c.pythonPath); err != nil {
		health.Score -= 20
		health.Issues = append(health.Issues, "Python not found")
		health.Recommendations = append(health.Recommendations, "Install Python 3.11+")
	}

	// Update status
	if health.Score >= 80 {
		health.Status = "healthy"
		health.Connected = true
	} else if health.Score >= 50 {
		health.Status = "degraded"
	} else {
		health.Status = "unhealthy"
	}

	return health
}

// isRekordboxRunning checks if Rekordbox is currently running
func (c *RekordboxClient) isRekordboxRunning() bool {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq rekordbox.exe")
		output, err := cmd.Output()
		if err != nil {
			return false
		}
		return strings.Contains(string(output), "rekordbox.exe")
	case "darwin":
		cmd := exec.Command("pgrep", "-x", "rekordbox")
		return cmd.Run() == nil
	default:
		return false
	}
}

// DatabasePath returns the configured database path
func (c *RekordboxClient) DatabasePath() string {
	return c.databasePath
}

// MusicPath returns the configured music path
func (c *RekordboxClient) MusicPath() string {
	return c.musicPath
}

// PlaylistPath returns the configured playlist path
func (c *RekordboxClient) PlaylistPath() string {
	return c.playlistPath
}

// GetNowPlaying returns the most recently played track from Rekordbox
func (c *RekordboxClient) GetNowPlaying(ctx context.Context) (*RekordboxTrack, error) {
	script := filepath.Join(c.scriptsPath, "rekordbox_now_playing.py")

	cmd := exec.CommandContext(ctx, c.pythonPath, script, "--json")
	cmd.Dir = filepath.Dir(c.scriptsPath)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get now playing: %w", err)
	}

	var track RekordboxTrack
	if err := json.Unmarshal(output, &track); err != nil {
		return nil, fmt.Errorf("failed to parse now playing: %w", err)
	}

	// Check for error in response
	if track.Title == "" && track.Source == "" {
		// Check if it's an error response
		var errResp map[string]string
		if err := json.Unmarshal(output, &errResp); err == nil {
			if errMsg, ok := errResp["error"]; ok {
				return nil, fmt.Errorf("rekordbox error: %s", errMsg)
			}
		}
	}

	return &track, nil
}

// GetHistory returns play history from Rekordbox
func (c *RekordboxClient) GetHistory(ctx context.Context, limit int) ([]RekordboxHistoryEntry, error) {
	script := filepath.Join(c.scriptsPath, "rekordbox_now_playing.py")

	args := []string{script, "--json", "--history"}
	if limit > 0 {
		args = append(args, "--limit", fmt.Sprintf("%d", limit))
	}

	cmd := exec.CommandContext(ctx, c.pythonPath, args...)
	cmd.Dir = filepath.Dir(c.scriptsPath)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}

	var history []RekordboxHistoryEntry
	if err := json.Unmarshal(output, &history); err != nil {
		return nil, fmt.Errorf("failed to parse history: %w", err)
	}

	return history, nil
}

// GetSessions returns list of DJ sessions from Rekordbox
func (c *RekordboxClient) GetSessions(ctx context.Context, limit int) ([]RekordboxSession, error) {
	script := filepath.Join(c.scriptsPath, "rekordbox_sessions.py")

	args := []string{script, "--json", "--list"}
	if limit > 0 {
		args = append(args, "--limit", fmt.Sprintf("%d", limit))
	}

	cmd := exec.CommandContext(ctx, c.pythonPath, args...)
	cmd.Dir = filepath.Dir(c.scriptsPath)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}

	var sessions []RekordboxSession
	if err := json.Unmarshal(output, &sessions); err != nil {
		return nil, fmt.Errorf("failed to parse sessions: %w", err)
	}

	return sessions, nil
}

// GetSessionTracks returns tracks from a specific DJ session
func (c *RekordboxClient) GetSessionTracks(ctx context.Context, sessionID int) ([]RekordboxHistoryEntry, error) {
	script := filepath.Join(c.scriptsPath, "rekordbox_sessions.py")

	args := []string{script, "--json", "--session", fmt.Sprintf("%d", sessionID)}

	cmd := exec.CommandContext(ctx, c.pythonPath, args...)
	cmd.Dir = filepath.Dir(c.scriptsPath)

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get session tracks: %w", err)
	}

	var tracks []RekordboxHistoryEntry
	if err := json.Unmarshal(output, &tracks); err != nil {
		return nil, fmt.Errorf("failed to parse session tracks: %w", err)
	}

	return tracks, nil
}
