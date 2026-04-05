package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Rekordbox timeouts
const (
	RekordboxImportTimeout = 5 * time.Minute
)

// RekordboxImporter handles Rekordbox import operations
type RekordboxImporter struct {
	config *Config
}

// NewRekordboxImporter creates a new Rekordbox importer
func NewRekordboxImporter(config *Config) *RekordboxImporter {
	return &RekordboxImporter{config: config}
}

// Import imports pending files to Rekordbox
func (r *RekordboxImporter) Import(ctx context.Context, state *State) (SyncResult, error) {
	result := SyncResult{
		Service:   "rekordbox",
		User:      "all",
		Playlist:  "import",
		StartTime: time.Now(),
		DryRun:    r.config.DryRun,
	}

	// Check if Rekordbox is running
	if r.isRekordboxRunning() {
		result.Errors = append(result.Errors, "Rekordbox is running, skipping import")
		result.EndTime = time.Now()
		return result, fmt.Errorf("rekordbox is running")
	}

	// Build playlist mappings for each user
	mappings := r.buildPlaylistMappings()

	// Run the Python script
	output, err := r.runPythonSync(ctx, mappings)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("python sync: %v", err))
		result.EndTime = time.Now()
		return result, err
	}

	// Parse output
	if output != nil {
		result.Synced = output.Imported
		result.Skipped = output.Skipped
		result.Failed = output.Failed
		result.Total = output.Total
		if output.Error != "" {
			result.Errors = append(result.Errors, output.Error)
		}
	}

	// Update state
	state.UpdateRekordbox(func(rs *RekordboxState) {
		rs.LastImport = time.Now()
		rs.PendingFiles = nil
		if len(result.Errors) > 0 {
			rs.LastError = strings.Join(result.Errors, "; ")
		} else {
			rs.LastError = ""
		}
	})

	result.EndTime = time.Now()
	return result, nil
}

// PlaylistMapping defines how to map local folders to Rekordbox playlists
type PlaylistMapping struct {
	LocalPath     string `json:"local_path"`
	RekordboxPath string `json:"rekordbox_path"` // e.g., "Hairglasses/SoundCloud Likes"
	PlaylistName  string `json:"playlist_name"`
	ParentFolder  string `json:"parent_folder"`
}

// PythonSyncOutput is the output from the Python sync script
type PythonSyncOutput struct {
	Total    int    `json:"total"`
	Imported int    `json:"imported"`
	Skipped  int    `json:"skipped"`
	Failed   int    `json:"failed"`
	Error    string `json:"error,omitempty"`
}

// buildPlaylistMappings creates mappings from local folders to Rekordbox playlists
func (r *RekordboxImporter) buildPlaylistMappings() []PlaylistMapping {
	var mappings []PlaylistMapping

	for _, user := range r.config.Users {
		// SoundCloud Likes
		if user.SoundCloud {
			mappings = append(mappings, PlaylistMapping{
				LocalPath:     filepath.Join(r.config.LocalRoot, "CR8", user.Username, "Likes"),
				RekordboxPath: fmt.Sprintf("%s/SoundCloud Likes", user.DisplayName),
				PlaylistName:  "SoundCloud Likes",
				ParentFolder:  user.DisplayName,
			})
		}

		// Beatport Likes (only for hairglasses currently)
		if user.Beatport && user.Username == "hairglasses" {
			mappings = append(mappings, PlaylistMapping{
				LocalPath:     filepath.Join(r.config.LocalRoot, "Beatport"),
				RekordboxPath: fmt.Sprintf("%s/Beatport Likes", user.DisplayName),
				PlaylistName:  "Beatport Likes",
				ParentFolder:  user.DisplayName,
			})
		}
	}

	return mappings
}

// runPythonSync runs the Python Rekordbox sync script with timeout and retry
func (r *RekordboxImporter) runPythonSync(ctx context.Context, mappings []PlaylistMapping) (*PythonSyncOutput, error) {
	// Find the Python script
	scriptPath := filepath.Join(os.Getenv("HOME"), "aftrs-studio", "hg-mcp", "scripts", "rekordbox_sync.py")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return nil, WrapPermanent(fmt.Errorf("script not found: %s", scriptPath))
	}

	// Serialize mappings to JSON
	mappingsJSON, err := json.Marshal(mappings)
	if err != nil {
		return nil, WrapPermanent(fmt.Errorf("marshal mappings: %w", err))
	}

	// Check circuit breaker first
	cb := GlobalCircuitBreakers.Get("rekordbox")

	return ExecuteWithResult(cb, ctx, func(ctx context.Context) (*PythonSyncOutput, error) {
		// Add timeout for import
		importCtx, cancel := context.WithTimeout(ctx, RekordboxImportTimeout)
		defer cancel()

		// Retry with exponential backoff
		retryConfig := RetryConfig{
			MaxRetries:   2,
			InitialDelay: 5 * time.Second,
			MaxDelay:     30 * time.Second,
			Multiplier:   2.0,
			Jitter:       0.1,
		}

		return RetryWithResult(importCtx, "rekordbox", "python_sync", retryConfig, func(ctx context.Context) (*PythonSyncOutput, error) {
			args := []string{scriptPath, "--mappings", string(mappingsJSON)}
			if r.config.DryRun {
				args = append(args, "--dry-run")
			}

			// Use pyenv python which has pyrekordbox installed
			pythonPath := filepath.Join(os.Getenv("HOME"), ".pyenv", "shims", "python3")
			cmd := exec.CommandContext(ctx, pythonPath, args...)
			cmd.Stderr = os.Stderr

			output, err := cmd.Output()
			if err != nil {
				if ctx.Err() != nil {
					return nil, WrapTimeout(ctx.Err())
				}
				return nil, WrapRetriable(fmt.Errorf("run script: %w", err))
			}

			var result PythonSyncOutput
			if err := json.Unmarshal(output, &result); err != nil {
				log.Printf("Python output: %s", string(output))
				return nil, WrapPermanent(fmt.Errorf("parse output: %w", err))
			}

			return &result, nil
		})
	})
}

// isRekordboxRunning checks if Rekordbox is currently running
func (r *RekordboxImporter) isRekordboxRunning() bool {
	cmd := exec.Command("pgrep", "-x", "rekordbox")
	err := cmd.Run()
	return err == nil
}

// ImportWithMappings imports tracks using provided playlist mappings (for pipeline use)
func (r *RekordboxImporter) ImportWithMappings(ctx context.Context, mappings []PlaylistMapping) (SyncResult, error) {
	result := SyncResult{
		Service:   "rekordbox",
		User:      "pipeline",
		Playlist:  "import",
		StartTime: time.Now(),
		DryRun:    r.config.DryRun,
	}

	// Check if Rekordbox is running
	if r.isRekordboxRunning() {
		result.Errors = append(result.Errors, "Rekordbox is running, skipping import")
		result.EndTime = time.Now()
		return result, fmt.Errorf("rekordbox is running")
	}

	if len(mappings) == 0 {
		result.EndTime = time.Now()
		return result, nil
	}

	// Run the Python script with provided mappings
	output, err := r.runPythonSync(ctx, mappings)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("python sync: %v", err))
		result.EndTime = time.Now()
		return result, err
	}

	// Parse output
	if output != nil {
		result.Synced = output.Imported
		result.Skipped = output.Skipped
		result.Failed = output.Failed
		result.Total = output.Total
		if output.Error != "" {
			result.Errors = append(result.Errors, output.Error)
		}
	}

	result.EndTime = time.Now()
	return result, nil
}
