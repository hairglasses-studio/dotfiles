package sync

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/config"
)

// Beatport timeouts
const (
	BeatportS3SyncTimeout   = 15 * time.Minute
	BeatportDownloadTimeout = 30 * time.Minute
)

// BeatportSyncer handles Beatport sync operations
type BeatportSyncer struct {
	config *Config
}

// NewBeatportSyncer creates a new Beatport syncer
func NewBeatportSyncer(config *Config) *BeatportSyncer {
	return &BeatportSyncer{config: config}
}

// Sync syncs Beatport playlists for a user
func (b *BeatportSyncer) Sync(ctx context.Context, username string, state *State) ([]SyncResult, error) {
	var results []SyncResult

	// For now, only sync the main Beatport likes playlist (4014989)
	result, err := b.syncPlaylist(ctx, username, "4014989", state)
	if err != nil {
		log.Printf("Beatport sync error for %s: %v", username, err)
	}
	results = append(results, result)

	return results, nil
}

// syncPlaylist syncs a specific Beatport playlist
func (b *BeatportSyncer) syncPlaylist(ctx context.Context, username, playlistID string, state *State) (SyncResult, error) {
	result := SyncResult{
		Service:   "beatport",
		User:      username,
		Playlist:  playlistID,
		StartTime: time.Now(),
		DryRun:    b.config.DryRun,
	}

	// Local path for Beatport downloads
	localPath := filepath.Join(b.config.LocalRoot, "Beatport")
	if err := os.MkdirAll(localPath, 0755); err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("create dir: %v", err))
		result.EndTime = time.Now()
		return result, err
	}

	// Step 1: Run beatport-sync to download new tracks
	if !b.config.DryRun {
		if err := b.runBeatportSync(ctx); err != nil {
			log.Printf("beatport-sync warning: %v", err)
			// Continue - sync from S3 anyway
		}
	}

	// Step 2: Sync from S3 to local
	s3Path := fmt.Sprintf("s3://%s/beatport", b.config.S3Bucket)
	if err := b.s3Sync(ctx, s3Path, localPath); err != nil {
		log.Printf("S3 sync warning: %v", err)
	}

	// Step 3: Count files synced
	files, err := filepath.Glob(filepath.Join(localPath, "*.aiff"))
	if err == nil {
		result.Synced = len(files)
	}

	// Update state
	state.UpdatePlaylistState("beatport", username, playlistID, func(ps *PlaylistSyncState) {
		ps.LastSync = time.Now()
		ps.Synced = result.Synced
	})

	result.Total = result.Synced
	result.EndTime = time.Now()
	return result, nil
}

// runBeatportSync runs the beatport-sync command with timeout and retry
func (b *BeatportSyncer) runBeatportSync(ctx context.Context) error {
	// Find the beatport-sync binary
	syncBin := filepath.Join(config.Get().Home, "hairglasses-studio", "hg-mcp", "bin", "beatport-sync")
	if _, err := os.Stat(syncBin); os.IsNotExist(err) {
		log.Printf("beatport-sync not found, skipping download step")
		return nil
	}

	// Check circuit breaker first
	cb := GlobalCircuitBreakers.Get("beatport")

	return cb.Execute(ctx, func(ctx context.Context) error {
		// Add timeout for download
		downloadCtx, cancel := context.WithTimeout(ctx, BeatportDownloadTimeout)
		defer cancel()

		// Rate limit Beatport API calls
		if err := GlobalRateLimiters.Wait(ctx, "beatport"); err != nil {
			return WrapRateLimit(err)
		}

		// Retry with exponential backoff
		retryConfig := RetryConfig{
			MaxRetries:   3,
			InitialDelay: 2 * time.Second,
			MaxDelay:     60 * time.Second,
			Multiplier:   2.0,
			Jitter:       0.1,
		}

		return Retry(downloadCtx, "beatport", "download", retryConfig, func(ctx context.Context) error {
			cmd := exec.CommandContext(ctx, syncBin)
			bpCfg := config.Get()
			cmd.Env = append(os.Environ(),
				"BEATPORT_USERNAME="+bpCfg.BeatportUsername,
				"BEATPORT_PASSWORD="+bpCfg.BeatportPassword,
			)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				if ctx.Err() != nil {
					return WrapTimeout(ctx.Err())
				}
				return WrapRetriable(err)
			}
			return nil
		})
	})
}

// s3Sync syncs from S3 to local using AWS CLI with timeout and retry
func (b *BeatportSyncer) s3Sync(ctx context.Context, s3Path, localPath string) error {
	if b.config.DryRun {
		log.Printf("[DRY-RUN] aws s3 sync %s -> %s", s3Path, localPath)
		return nil
	}

	// Check circuit breaker first
	cb := GlobalCircuitBreakers.Get("beatport")

	return cb.Execute(ctx, func(ctx context.Context) error {
		// Add timeout for S3 sync
		syncCtx, cancel := context.WithTimeout(ctx, BeatportS3SyncTimeout)
		defer cancel()

		// Rate limit S3 API calls
		if err := GlobalRateLimiters.Wait(ctx, "beatport"); err != nil {
			return WrapRateLimit(err)
		}

		// Retry with exponential backoff
		return Retry(syncCtx, "beatport", "s3_sync", DefaultRetryConfig(), func(ctx context.Context) error {
			args := []string{
				"s3", "sync", s3Path, localPath,
				"--profile", b.config.AWSProfile,
			}

			cmd := exec.CommandContext(ctx, "aws", args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			if err := cmd.Run(); err != nil {
				if ctx.Err() != nil {
					return WrapTimeout(ctx.Err())
				}
				return WrapRetriable(err)
			}
			return nil
		})
	})
}
