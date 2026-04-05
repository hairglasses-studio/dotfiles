package sync

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	gosync "sync"
	"time"
)

// Google Drive timeouts and config
const (
	GDriveSyncTimeout      = 10 * time.Minute
	GDriveParallelWorkers  = 3
	GDriveDefaultRemote    = "gdrive"
	GDriveDefaultMountPath = "~/GDrive"
	GDriveDefaultBasePath  = "DJ Crates/SoundCloud"
)

// GDriveSyncer handles Google Drive sync operations via rclone
type GDriveSyncer struct {
	config    *Config
	remote    string // rclone remote name (e.g., "gdrive")
	mountPath string // local mount path for rclone mount
	basePath  string // base path in GDrive (e.g., "DJ Crates/SoundCloud")
}

// GDriveSyncerOption configures the GDrive syncer
type GDriveSyncerOption func(*GDriveSyncer)

// WithRemote sets the rclone remote name
func WithRemote(remote string) GDriveSyncerOption {
	return func(s *GDriveSyncer) {
		s.remote = remote
	}
}

// WithMountPath sets the local mount path
func WithMountPath(path string) GDriveSyncerOption {
	return func(s *GDriveSyncer) {
		s.mountPath = expandTilde(path)
	}
}

// WithBasePath sets the base path in Google Drive
func WithBasePath(path string) GDriveSyncerOption {
	return func(s *GDriveSyncer) {
		s.basePath = path
	}
}

// NewGDriveSyncer creates a new Google Drive syncer
func NewGDriveSyncer(config *Config, opts ...GDriveSyncerOption) *GDriveSyncer {
	s := &GDriveSyncer{
		config:    config,
		remote:    GDriveDefaultRemote,
		mountPath: expandTilde(GDriveDefaultMountPath),
		basePath:  GDriveDefaultBasePath,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// GDriveSyncResult contains the result of a GDrive sync operation
type GDriveSyncResult struct {
	Playlist      string        `json:"playlist"`
	FilesUploaded int           `json:"files_uploaded"`
	FilesSkipped  int           `json:"files_skipped"`
	BytesUploaded int64         `json:"bytes_uploaded"`
	GDrivePath    string        `json:"gdrive_path"`
	LocalPath     string        `json:"local_path"`
	MountPath     string        `json:"mount_path"`
	Duration      time.Duration `json:"duration"`
	Error         string        `json:"error,omitempty"`
}

// SyncFolder syncs a local folder to Google Drive
func (s *GDriveSyncer) SyncFolder(ctx context.Context, localPath, username, playlist string) (*GDriveSyncResult, error) {
	startTime := time.Now()

	result := &GDriveSyncResult{
		Playlist:  playlist,
		LocalPath: localPath,
	}

	// Build GDrive path: gdrive:DJ Crates/SoundCloud/{username}/{playlist}/
	gdrivePath := fmt.Sprintf("%s:%s/%s/%s", s.remote, s.basePath, username, playlist)
	result.GDrivePath = gdrivePath

	// Build mount path for Rekordbox access
	result.MountPath = filepath.Join(s.mountPath, s.basePath, username, playlist)

	if s.config.DryRun {
		log.Printf("[DRY-RUN] rclone sync %s -> %s", localPath, gdrivePath)
		result.Duration = time.Since(startTime)
		return result, nil
	}

	// Check if local path exists and has files
	files, err := countAudioFiles(localPath)
	if err != nil {
		result.Error = fmt.Sprintf("count files: %v", err)
		result.Duration = time.Since(startTime)
		return result, err
	}

	if files == 0 {
		log.Printf("No audio files in %s, skipping GDrive sync", localPath)
		result.Duration = time.Since(startTime)
		return result, nil
	}

	// Run rclone sync with circuit breaker and retry
	cb := GlobalCircuitBreakers.Get("gdrive")

	err = cb.Execute(ctx, func(ctx context.Context) error {
		syncCtx, cancel := context.WithTimeout(ctx, GDriveSyncTimeout)
		defer cancel()

		// Rate limit rclone operations
		if err := GlobalRateLimiters.Wait(ctx, "gdrive"); err != nil {
			return WrapRateLimit(err)
		}

		return Retry(syncCtx, "gdrive", "rclone_sync", DefaultRetryConfig(), func(ctx context.Context) error {
			args := []string{
				"sync",
				localPath,
				gdrivePath,
				"--progress",
				"--stats-one-line",
				"--transfers", "4",
				"--checkers", "8",
			}

			cmd := exec.CommandContext(ctx, "rclone", args...)
			output, err := cmd.CombinedOutput()

			if err != nil {
				if ctx.Err() != nil {
					return WrapTimeout(ctx.Err())
				}
				return WrapRetriable(fmt.Errorf("rclone sync failed: %w - %s", err, string(output)))
			}

			// Parse output for stats
			result.FilesUploaded, result.BytesUploaded = parseRcloneOutput(string(output))
			return nil
		})
	})

	if err != nil {
		result.Error = err.Error()
	}

	result.Duration = time.Since(startTime)
	return result, err
}

// SyncAllPlaylists syncs all playlists for a user to Google Drive
func (s *GDriveSyncer) SyncAllPlaylists(ctx context.Context, username string, playlists []string, localRoot string) ([]GDriveSyncResult, error) {
	var results []GDriveSyncResult
	var resultsMu gosync.Mutex

	if len(playlists) == 0 {
		return results, nil
	}

	log.Printf("Syncing %d playlists to GDrive for %s with %d workers", len(playlists), username, GDriveParallelWorkers)

	playlistChan := make(chan string, len(playlists))
	var wg gosync.WaitGroup

	// Start workers
	for i := 0; i < GDriveParallelWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for playlist := range playlistChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				localPath := filepath.Join(localRoot, username, playlist)
				result, err := s.SyncFolder(ctx, localPath, username, playlist)
				if err != nil {
					log.Printf("Worker %d: GDrive sync error for %s/%s: %v", workerID, username, playlist, err)
				}

				resultsMu.Lock()
				results = append(results, *result)
				resultsMu.Unlock()
			}
		}(i)
	}

	// Send playlists to workers
	for _, playlist := range playlists {
		playlistChan <- playlist
	}
	close(playlistChan)

	wg.Wait()
	return results, nil
}

// GetMountPath returns the local mount path for a playlist
func (s *GDriveSyncer) GetMountPath(username, playlist string) string {
	return filepath.Join(s.mountPath, s.basePath, username, playlist)
}

// CheckRcloneRemote verifies the rclone remote is configured
func (s *GDriveSyncer) CheckRcloneRemote(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, "rclone", "listremotes")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("rclone not available: %w", err)
	}

	remotes := strings.Split(string(output), "\n")
	for _, remote := range remotes {
		if strings.TrimSpace(remote) == s.remote+":" {
			return nil
		}
	}

	return fmt.Errorf("rclone remote '%s' not configured", s.remote)
}

// CheckMountPath verifies the mount path exists (for Rekordbox access)
func (s *GDriveSyncer) CheckMountPath() error {
	if _, err := os.Stat(s.mountPath); os.IsNotExist(err) {
		return fmt.Errorf("GDrive mount path does not exist: %s (run: rclone mount %s: %s)", s.mountPath, s.remote, s.mountPath)
	}
	return nil
}

// Helper functions

// expandTilde expands ~ to user home directory
func expandTilde(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}

func countAudioFiles(dir string) (int, error) {
	var count int

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		switch ext {
		case ".mp3", ".m4a", ".aiff", ".wav", ".flac":
			count++
		}
	}

	return count, nil
}

func parseRcloneOutput(output string) (files int, bytes int64) {
	// Parse rclone stats output
	// Example: "Transferred:   	   10 / 10, 100%, 0 B/s, ETA -"
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Transferred:") {
			// Extract file count - simplified parsing
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "/" && i > 0 {
					fmt.Sscanf(parts[i-1], "%d", &files)
					break
				}
			}
		}
	}
	return files, bytes
}
