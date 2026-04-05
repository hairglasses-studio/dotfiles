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

// SoundCloud timeouts and config
const (
	SoundCloudS3SyncTimeout    = 10 * time.Minute
	SoundCloudDiscoveryTimeout = 30 * time.Second
	SoundCloudParallelWorkers  = 3 // Number of parallel playlist sync workers
)

// SoundCloudSyncer handles SoundCloud sync operations
type SoundCloudSyncer struct {
	config *Config
}

// NewSoundCloudSyncer creates a new SoundCloud syncer
func NewSoundCloudSyncer(config *Config) *SoundCloudSyncer {
	return &SoundCloudSyncer{config: config}
}

// Sync syncs SoundCloud playlists for a user with parallel workers
func (s *SoundCloudSyncer) Sync(ctx context.Context, username string, state *State) ([]SyncResult, error) {
	var results []SyncResult
	var resultsMu gosync.Mutex

	// Sync likes playlist first (always)
	result, err := s.syncLikes(ctx, username, state)
	if err != nil {
		log.Printf("SoundCloud likes sync error for %s: %v", username, err)
	}
	results = append(results, result)

	// Auto-discover playlists from S3
	playlists, err := s.discoverPlaylists(ctx, username)
	if err != nil {
		log.Printf("Playlist discovery warning for %s: %v", username, err)
		return results, nil
	}

	// Filter out likes (already synced)
	var playlistsToSync []string
	for _, playlist := range playlists {
		if playlist != "likes" {
			playlistsToSync = append(playlistsToSync, playlist)
		}
	}

	if len(playlistsToSync) == 0 {
		return results, nil
	}

	// Parallel sync with worker pool
	log.Printf("Syncing %d playlists for %s with %d workers", len(playlistsToSync), username, SoundCloudParallelWorkers)

	playlistChan := make(chan string, len(playlistsToSync))
	var wg gosync.WaitGroup

	// Start workers
	for i := 0; i < SoundCloudParallelWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for playlist := range playlistChan {
				select {
				case <-ctx.Done():
					return
				default:
				}

				result, err := s.syncPlaylist(ctx, username, playlist, state)
				if err != nil {
					log.Printf("Worker %d: SoundCloud playlist %s sync error for %s: %v", workerID, playlist, username, err)
				}

				resultsMu.Lock()
				results = append(results, result)
				resultsMu.Unlock()
			}
		}(i)
	}

	// Send playlists to workers
	for _, playlist := range playlistsToSync {
		playlistChan <- playlist
	}
	close(playlistChan)

	// Wait for all workers to complete
	wg.Wait()

	return results, nil
}

// discoverPlaylists lists all playlist folders for a user in S3
func (s *SoundCloudSyncer) discoverPlaylists(ctx context.Context, username string) ([]string, error) {
	// Check circuit breaker first
	cb := GlobalCircuitBreakers.Get("soundcloud")

	return ExecuteWithResult(cb, ctx, func(ctx context.Context) ([]string, error) {
		// Add timeout for discovery
		discoverCtx, cancel := context.WithTimeout(ctx, SoundCloudDiscoveryTimeout)
		defer cancel()

		// Rate limit S3 API calls
		if err := GlobalRateLimiters.Wait(ctx, "soundcloud"); err != nil {
			return nil, WrapRateLimit(err)
		}

		return RetryWithResult(discoverCtx, "soundcloud", "discover_playlists", DefaultRetryConfig(), func(ctx context.Context) ([]string, error) {
			s3Path := fmt.Sprintf("s3://%s/users/%s/soundcloud/", s.config.S3Bucket, username)
			cmd := exec.CommandContext(ctx, "aws", "s3", "ls", s3Path, "--profile", s.config.AWSProfile)
			output, err := cmd.Output()
			if err != nil {
				return nil, WrapRetriable(err)
			}

			var playlists []string
			for _, line := range strings.Split(string(output), "\n") {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "PRE ") {
					name := strings.TrimSuffix(strings.TrimPrefix(line, "PRE "), "/")
					playlists = append(playlists, name)
				}
			}
			return playlists, nil
		})
	})
}

// syncLikes syncs user's likes from SoundCloud
func (s *SoundCloudSyncer) syncLikes(ctx context.Context, username string, state *State) (SyncResult, error) {
	result := SyncResult{
		Service:   "soundcloud",
		User:      username,
		Playlist:  "likes",
		StartTime: time.Now(),
		DryRun:    s.config.DryRun,
	}

	// Local path for this user's likes
	localPath := filepath.Join(s.config.LocalRoot, "CR8", username, "Likes")
	if err := os.MkdirAll(localPath, 0755); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("create dir: %v", err))
		result.EndTime = time.Now()
		return result, err
	}

	// Step 1: Sync from S3 to local using AWS CLI
	s3Path := fmt.Sprintf("s3://%s/users/%s/soundcloud/likes", s.config.S3Bucket, username)
	if err := s.s3Sync(ctx, s3Path, localPath); err != nil {
		log.Printf("S3 sync warning for %s: %v", username, err)
		// Continue - may just be empty
	}

	// Step 2: Count files synced
	files, err := filepath.Glob(filepath.Join(localPath, "*.aiff"))
	if err == nil {
		result.Synced = len(files)
	}
	mp3Files, _ := filepath.Glob(filepath.Join(localPath, "*.mp3"))
	result.Synced += len(mp3Files)

	// Update state
	state.UpdatePlaylistState("soundcloud", username, "likes", func(ps *PlaylistSyncState) {
		ps.LastSync = time.Now()
		ps.Synced = result.Synced
	})

	result.Total = result.Synced
	result.EndTime = time.Now()
	return result, nil
}

// syncPlaylist syncs a specific playlist from SoundCloud
func (s *SoundCloudSyncer) syncPlaylist(ctx context.Context, username, playlist string, state *State) (SyncResult, error) {
	result := SyncResult{
		Service:   "soundcloud",
		User:      username,
		Playlist:  playlist,
		StartTime: time.Now(),
		DryRun:    s.config.DryRun,
	}

	// Local path for this playlist
	localPath := filepath.Join(s.config.LocalRoot, "CR8", username, playlist)
	if err := os.MkdirAll(localPath, 0755); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("create dir: %v", err))
		result.EndTime = time.Now()
		return result, err
	}

	// Sync from S3
	s3Path := fmt.Sprintf("s3://%s/users/%s/soundcloud/%s", s.config.S3Bucket, username, playlist)
	if err := s.s3Sync(ctx, s3Path, localPath); err != nil {
		log.Printf("S3 sync warning for %s/%s: %v", username, playlist, err)
	}

	// Count files
	files, _ := filepath.Glob(filepath.Join(localPath, "*.aiff"))
	result.Synced = len(files)
	mp3Files, _ := filepath.Glob(filepath.Join(localPath, "*.mp3"))
	result.Synced += len(mp3Files)
	m4aFiles, _ := filepath.Glob(filepath.Join(localPath, "*.m4a"))
	result.Synced += len(m4aFiles)

	state.UpdatePlaylistState("soundcloud", username, playlist, func(ps *PlaylistSyncState) {
		ps.LastSync = time.Now()
		ps.Synced = result.Synced
	})

	result.Total = result.Synced
	result.EndTime = time.Now()
	return result, nil
}

// s3Sync syncs from S3 to local using AWS CLI with timeout and retry
func (s *SoundCloudSyncer) s3Sync(ctx context.Context, s3Path, localPath string) error {
	if s.config.DryRun {
		log.Printf("[DRY-RUN] aws s3 sync %s -> %s", s3Path, localPath)
		return nil
	}

	// Check circuit breaker first
	cb := GlobalCircuitBreakers.Get("soundcloud")

	return cb.Execute(ctx, func(ctx context.Context) error {
		// Add timeout for S3 sync
		syncCtx, cancel := context.WithTimeout(ctx, SoundCloudS3SyncTimeout)
		defer cancel()

		// Rate limit S3 API calls
		if err := GlobalRateLimiters.Wait(ctx, "soundcloud"); err != nil {
			return WrapRateLimit(err)
		}

		// Retry with exponential backoff
		return Retry(syncCtx, "soundcloud", "s3_sync", DefaultRetryConfig(), func(ctx context.Context) error {
			args := []string{
				"s3", "sync", s3Path, localPath,
				"--profile", s.config.AWSProfile,
			}

			cmd := exec.CommandContext(ctx, "aws", args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				// Check if context was cancelled
				if ctx.Err() != nil {
					return WrapTimeout(ctx.Err())
				}
				return WrapRetriable(err)
			}
			return nil
		})
	})
}

// GetUserDisplayName returns the display name for a username
func (s *SoundCloudSyncer) GetUserDisplayName(username string) string {
	for _, user := range s.config.Users {
		if strings.EqualFold(user.Username, username) {
			return user.DisplayName
		}
	}
	return username
}
