package soundcloud

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	musicsync "github.com/hairglasses-studio/hg-mcp/internal/sync"
)

// Pipeline steps for SoundCloud -> GDrive -> Rekordbox

// DiscoverPlaylistsStep discovers playlists from SoundCloud
type DiscoverPlaylistsStep struct {
	client *clients.SoundCloudClient
}

func (s *DiscoverPlaylistsStep) Name() string {
	return "Discover Playlists"
}

func (s *DiscoverPlaylistsStep) Phase() musicsync.PipelinePhase {
	return musicsync.PhaseDiscover
}

func (s *DiscoverPlaylistsStep) Execute(ctx context.Context, state *musicsync.PipelineState) error {
	client, err := getSoundCloudClient()
	if err != nil {
		return fmt.Errorf("soundcloud client: %w", err)
	}

	// Resolve username to user ID
	url := fmt.Sprintf("https://soundcloud.com/%s", state.Username)
	resource, resourceType, err := client.Resolve(ctx, url)
	if err != nil {
		return fmt.Errorf("resolve user: %w", err)
	}

	if resourceType != "user" {
		return fmt.Errorf("expected user, got %s", resourceType)
	}

	userMap, ok := resource.(map[string]any)
	if !ok {
		return fmt.Errorf("invalid user response")
	}

	userID, ok := userMap["id"].(float64)
	if !ok {
		return fmt.Errorf("missing user ID")
	}

	var playlists []string
	totalTracks := 0

	// Add likes if requested
	if state.IncludeLikes {
		playlists = append(playlists, "Likes")
		// Get likes count
		likes, err := client.GetUserLikes(ctx, int64(userID), 1)
		if err == nil {
			// Estimate from limited query - actual count may be higher
			totalTracks += len(likes) * 50 // rough estimate
		}
	}

	// Get user playlists
	userPlaylists, err := client.GetUserPlaylists(ctx, int64(userID), 200)
	if err != nil {
		log.Printf("Warning: could not fetch playlists: %v", err)
	} else {
		for _, pl := range userPlaylists {
			playlists = append(playlists, pl.Title)
			totalTracks += pl.TrackCount
		}
	}

	state.Playlists = playlists
	state.TotalTracks = totalTracks

	log.Printf("Discovered %d playlists with ~%d tracks for %s", len(playlists), totalTracks, state.Username)
	return nil
}

// DownloadTracksStep downloads tracks from SoundCloud
type DownloadTracksStep struct{}

func (s *DownloadTracksStep) Name() string {
	return "Download Tracks"
}

func (s *DownloadTracksStep) Phase() musicsync.PipelinePhase {
	return musicsync.PhaseDownload
}

func (s *DownloadTracksStep) Execute(ctx context.Context, state *musicsync.PipelineState) error {
	if len(state.Playlists) == 0 {
		return nil
	}

	// Ensure local root exists
	userRoot := filepath.Join(state.LocalRoot, state.Username)
	if err := os.MkdirAll(userRoot, 0755); err != nil {
		return fmt.Errorf("create user root: %w", err)
	}

	for _, playlist := range state.Playlists {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		playlistDir := filepath.Join(userRoot, playlist)
		if err := os.MkdirAll(playlistDir, 0755); err != nil {
			state.DownloadErrors = append(state.DownloadErrors, fmt.Sprintf("%s: create dir: %v", playlist, err))
			continue
		}

		// Build URL
		var url string
		if playlist == "Likes" {
			url = fmt.Sprintf("https://soundcloud.com/%s/likes", state.Username)
		} else {
			// For playlists, we need to resolve the playlist URL
			url = fmt.Sprintf("https://soundcloud.com/%s/sets/%s", state.Username, slugify(playlist))
		}

		// Count existing files before download
		existingCount, _ := countAudioFilesInDir(playlistDir)

		// Download using scdl with fallback to yt-dlp
		if state.DryRun {
			log.Printf("[DRY-RUN] Would download %s to %s", url, playlistDir)
			continue
		}

		err := downloadWithScdl(ctx, url, playlistDir, state.Format)
		if err != nil {
			// Try yt-dlp fallback
			err = downloadWithYtdlp(ctx, url, playlistDir, state.Format)
			if err != nil {
				state.DownloadErrors = append(state.DownloadErrors, fmt.Sprintf("%s: %v", playlist, err))
				continue
			}
		}

		// Count new files
		newCount, _ := countAudioFilesInDir(playlistDir)
		downloaded := newCount - existingCount
		if downloaded > 0 {
			state.DownloadedTracks += downloaded
		}
		state.SkippedTracks += existingCount
	}

	log.Printf("Downloaded %d tracks, skipped %d existing", state.DownloadedTracks, state.SkippedTracks)
	return nil
}

// UploadToGDriveStep syncs local files to Google Drive
type UploadToGDriveStep struct{}

func (s *UploadToGDriveStep) Name() string {
	return "Upload to Google Drive"
}

func (s *UploadToGDriveStep) Phase() musicsync.PipelinePhase {
	return musicsync.PhaseUpload
}

func (s *UploadToGDriveStep) Execute(ctx context.Context, state *musicsync.PipelineState) error {
	if state.SkipGDrive {
		log.Printf("Skipping Google Drive sync")
		return nil
	}

	config := musicsync.DefaultConfig()
	config.DryRun = state.DryRun

	syncer := musicsync.NewGDriveSyncer(config,
		musicsync.WithMountPath(state.GDriveMountPath),
		musicsync.WithBasePath(state.GDriveBasePath),
	)

	// Verify rclone remote is configured
	if err := syncer.CheckRcloneRemote(ctx); err != nil {
		return fmt.Errorf("rclone check: %w", err)
	}

	// Sync all playlists
	results, err := syncer.SyncAllPlaylists(ctx, state.Username, state.Playlists, state.LocalRoot)
	if err != nil {
		return fmt.Errorf("gdrive sync: %w", err)
	}

	state.GDriveResults = results

	totalUploaded := 0
	for _, r := range results {
		totalUploaded += r.FilesUploaded
		if r.Error != "" {
			log.Printf("GDrive sync error for %s: %s", r.Playlist, r.Error)
		}
	}

	log.Printf("Uploaded %d files to Google Drive across %d playlists", totalUploaded, len(results))
	return nil
}

// ImportToRekordboxStep imports tracks to Rekordbox
type ImportToRekordboxStep struct{}

func (s *ImportToRekordboxStep) Name() string {
	return "Import to Rekordbox"
}

func (s *ImportToRekordboxStep) Phase() musicsync.PipelinePhase {
	return musicsync.PhaseImport
}

func (s *ImportToRekordboxStep) Execute(ctx context.Context, state *musicsync.PipelineState) error {
	if state.SkipRekordbox {
		log.Printf("Skipping Rekordbox import")
		return nil
	}

	// Check if Rekordbox is running
	if isRekordboxRunning() {
		return fmt.Errorf("rekordbox is running - close it before importing")
	}

	config := musicsync.DefaultConfig()
	config.DryRun = state.DryRun

	// Build dynamic mappings based on GDrive mount path
	mappings := buildGDriveMappings(state)

	if len(mappings) == 0 {
		log.Printf("No mappings to import")
		return nil
	}

	if state.DryRun {
		log.Printf("[DRY-RUN] Would import %d playlist mappings to Rekordbox", len(mappings))
		for _, m := range mappings {
			log.Printf("[DRY-RUN]   %s -> %s", m.LocalPath, m.RekordboxPath)
		}
		return nil
	}

	// Run Python sync script
	importer := musicsync.NewRekordboxImporter(config)
	result, err := importer.ImportWithMappings(ctx, mappings)
	if err != nil {
		return fmt.Errorf("rekordbox import: %w", err)
	}

	state.ImportedTracks = result.Synced
	state.PlaylistsCreated = extractPlaylistNames(mappings)

	if len(result.Errors) > 0 {
		state.ImportErrors = result.Errors
	}

	log.Printf("Imported %d tracks to Rekordbox", state.ImportedTracks)
	return nil
}

// buildGDriveMappings creates Rekordbox playlist mappings from GDrive-synced folders
func buildGDriveMappings(state *musicsync.PipelineState) []musicsync.PlaylistMapping {
	var mappings []musicsync.PlaylistMapping

	// Get display name for user
	displayName := getDisplayName(state.Username)

	for _, playlist := range state.Playlists {
		// Path in GDrive mount
		localPath := filepath.Join(state.GDriveMountPath, state.GDriveBasePath, state.Username, playlist)

		// Rekordbox path: {DisplayName}/SoundCloud {Playlist}
		rbPath := fmt.Sprintf("%s/SoundCloud %s", displayName, playlist)
		if playlist == "Likes" {
			rbPath = fmt.Sprintf("%s/SoundCloud Likes", displayName)
		}

		mappings = append(mappings, musicsync.PlaylistMapping{
			LocalPath:     localPath,
			RekordboxPath: rbPath,
			PlaylistName:  fmt.Sprintf("SoundCloud %s", playlist),
			ParentFolder:  displayName,
		})
	}

	return mappings
}

// Helper functions

func countAudioFilesInDir(dir string) (int, error) {
	var count int
	entries, err := os.ReadDir(dir)
	if err != nil {
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

func slugify(s string) string {
	// Simple slugify for playlist names
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

func isRekordboxRunning() bool {
	cmd := exec.Command("pgrep", "-x", "rekordbox")
	return cmd.Run() == nil
}

func getDisplayName(username string) string {
	// Map usernames to display names
	displayNames := map[string]string{
		"hairglasses": "Hairglasses",
		"luke-lasley": "Luke",
	}

	if name, ok := displayNames[username]; ok {
		return name
	}
	// Title case the username
	return strings.Title(username)
}

func extractPlaylistNames(mappings []musicsync.PlaylistMapping) []string {
	var names []string
	for _, m := range mappings {
		names = append(names, m.PlaylistName)
	}
	return names
}

// handleSoundCloudToRekordbox is the unified pipeline handler
func handleSoundCloudToRekordbox(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Parse parameters
	username, errResult := tools.RequireStringParam(req, "username")
	if errResult != nil {
		return errResult, nil
	}

	format := tools.OptionalStringParam(req, "format", "mp3")

	dryRun := tools.GetBoolParam(req, "dry_run", false)
	includeLikes := tools.GetBoolParam(req, "include_likes", true)
	skipGDrive := tools.GetBoolParam(req, "skip_gdrive", false)
	skipRekordbox := tools.GetBoolParam(req, "skip_rekordbox", false)

	gdriveMountPath := tools.GetStringParam(req, "gdrive_mount_path")
	if gdriveMountPath == "" {
		gdriveMountPath = musicsync.GDriveDefaultMountPath
	}

	gdriveBasePath := tools.GetStringParam(req, "gdrive_base_path")
	if gdriveBasePath == "" {
		gdriveBasePath = musicsync.GDriveDefaultBasePath
	}

	// Determine local root
	home, _ := os.UserHomeDir()
	localRoot := filepath.Join(home, "Music", "CR8")

	// Initialize pipeline state
	state := &musicsync.PipelineState{
		Username:        username,
		Format:          format,
		DryRun:          dryRun,
		IncludeLikes:    includeLikes,
		SkipGDrive:      skipGDrive,
		SkipRekordbox:   skipRekordbox,
		LocalRoot:       localRoot,
		GDriveMountPath: expandPath(gdriveMountPath),
		GDriveBasePath:  gdriveBasePath,
	}

	// Create pipeline
	pipeline := musicsync.NewPipeline("soundcloud-to-rekordbox", state)

	// Add steps
	pipeline.AddStep(&DiscoverPlaylistsStep{})
	pipeline.AddStep(&DownloadTracksStep{})
	pipeline.AddStep(&UploadToGDriveStep{})
	pipeline.AddStep(&ImportToRekordboxStep{})

	// Execute pipeline
	result := pipeline.Execute(ctx)

	return tools.TextResult(result.JSON()), nil
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, path[2:])
	}
	return path
}
