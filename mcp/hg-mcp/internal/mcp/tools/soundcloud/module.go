// Package soundcloud provides MCP tools for SoundCloud API integration.
package soundcloud

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	musicsync "github.com/hairglasses-studio/hg-mcp/internal/sync"
)

// Module implements the SoundCloud tools module
type Module struct{}

var getSoundCloudClient = tools.LazyClient(clients.GetSoundCloudClient)

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "soundcloud"
}

// Description returns the module description
func (m *Module) Description() string {
	return "SoundCloud API integration for track discovery, playlists, and music sync"
}

// Tools returns all SoundCloud tools
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_soundcloud_status",
				mcp.WithDescription("Get SoundCloud API connection status and token info"),
			),
			Handler:             handleSoundCloudStatus,
			Category:            "soundcloud",
			Subcategory:         "status",
			Tags:                []string{"soundcloud", "music", "status", "api"},
			UseCases:            []string{"Check SoundCloud API connectivity", "Verify authentication"},
			Complexity:          "simple",
			CircuitBreakerGroup: "soundcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_soundcloud_health",
				mcp.WithDescription("Check SoundCloud API health and configuration"),
			),
			Handler:             handleSoundCloudHealth,
			Category:            "soundcloud",
			Subcategory:         "status",
			Tags:                []string{"soundcloud", "health", "diagnostics"},
			UseCases:            []string{"Diagnose API issues", "Check configuration"},
			Complexity:          "simple",
			CircuitBreakerGroup: "soundcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_soundcloud_resolve",
				mcp.WithDescription("Resolve any SoundCloud URL to get track, user, or playlist details"),
				mcp.WithString("url", mcp.Required(), mcp.Description("SoundCloud URL to resolve (track, user, or playlist)")),
			),
			Handler:             handleSoundCloudResolve,
			Category:            "soundcloud",
			Subcategory:         "utility",
			Tags:                []string{"soundcloud", "music", "resolve", "url"},
			UseCases:            []string{"Get details from SoundCloud link", "Identify resource type"},
			Complexity:          "simple",
			CircuitBreakerGroup: "soundcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_soundcloud_search",
				mcp.WithDescription("Search SoundCloud for tracks, users, or playlists"),
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query")),
				mcp.WithString("type", mcp.Description("Search type: track, user, playlist, or all (default: track)")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (default: 20, max: 200)")),
			),
			Handler:             handleSoundCloudSearch,
			Category:            "soundcloud",
			Subcategory:         "search",
			Tags:                []string{"soundcloud", "music", "search", "discovery"},
			UseCases:            []string{"Find tracks by name", "Search for artists", "Discover new music"},
			Complexity:          "simple",
			CircuitBreakerGroup: "soundcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_soundcloud_track",
				mcp.WithDescription("Get detailed information about a SoundCloud track"),
				mcp.WithNumber("track_id", mcp.Required(), mcp.Description("SoundCloud track ID")),
			),
			Handler:             handleSoundCloudTrack,
			Category:            "soundcloud",
			Subcategory:         "tracks",
			Tags:                []string{"soundcloud", "music", "track", "metadata"},
			UseCases:            []string{"Get track details", "View track metadata", "Get BPM and key"},
			Complexity:          "simple",
			CircuitBreakerGroup: "soundcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_soundcloud_user",
				mcp.WithDescription("Get detailed information about a SoundCloud user"),
				mcp.WithNumber("user_id", mcp.Required(), mcp.Description("SoundCloud user ID")),
			),
			Handler:             handleSoundCloudUser,
			Category:            "soundcloud",
			Subcategory:         "users",
			Tags:                []string{"soundcloud", "music", "user", "profile"},
			UseCases:            []string{"Get user profile", "View follower count"},
			Complexity:          "simple",
			CircuitBreakerGroup: "soundcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_soundcloud_user_tracks",
				mcp.WithDescription("Get tracks uploaded by a SoundCloud user"),
				mcp.WithNumber("user_id", mcp.Required(), mcp.Description("SoundCloud user ID")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (default: 50, max: 200)")),
			),
			Handler:             handleSoundCloudUserTracks,
			Category:            "soundcloud",
			Subcategory:         "users",
			Tags:                []string{"soundcloud", "music", "user", "tracks"},
			UseCases:            []string{"Browse user's music", "Get artist discography"},
			Complexity:          "simple",
			CircuitBreakerGroup: "soundcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_soundcloud_likes",
				mcp.WithDescription("Get a user's liked tracks on SoundCloud"),
				mcp.WithNumber("user_id", mcp.Required(), mcp.Description("SoundCloud user ID")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (default: 50, max: 200)")),
			),
			Handler:             handleSoundCloudLikes,
			Category:            "soundcloud",
			Subcategory:         "likes",
			Tags:                []string{"soundcloud", "music", "likes", "favorites"},
			UseCases:            []string{"View liked tracks", "Sync likes to library"},
			Complexity:          "simple",
			CircuitBreakerGroup: "soundcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_soundcloud_playlist",
				mcp.WithDescription("Get details and tracks from a SoundCloud playlist"),
				mcp.WithNumber("playlist_id", mcp.Required(), mcp.Description("SoundCloud playlist ID")),
			),
			Handler:             handleSoundCloudPlaylist,
			Category:            "soundcloud",
			Subcategory:         "playlists",
			Tags:                []string{"soundcloud", "music", "playlist"},
			UseCases:            []string{"View playlist tracks", "Get playlist metadata"},
			Complexity:          "simple",
			CircuitBreakerGroup: "soundcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_soundcloud_playlists",
				mcp.WithDescription("Get playlists created by a SoundCloud user"),
				mcp.WithNumber("user_id", mcp.Required(), mcp.Description("SoundCloud user ID")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (default: 50, max: 200)")),
			),
			Handler:             handleSoundCloudPlaylists,
			Category:            "soundcloud",
			Subcategory:         "playlists",
			Tags:                []string{"soundcloud", "music", "playlists", "user"},
			UseCases:            []string{"Browse user playlists", "Find curated collections"},
			Complexity:          "simple",
			CircuitBreakerGroup: "soundcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_soundcloud_followers",
				mcp.WithDescription("Get a user's followers or following list"),
				mcp.WithNumber("user_id", mcp.Required(), mcp.Description("SoundCloud user ID")),
				mcp.WithString("type", mcp.Description("Type: followers or following (default: followers)")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (default: 50)")),
			),
			Handler:             handleSoundCloudFollowers,
			Category:            "soundcloud",
			Subcategory:         "users",
			Tags:                []string{"soundcloud", "music", "followers", "social"},
			UseCases:            []string{"View user connections", "Discover related artists"},
			Complexity:          "simple",
			CircuitBreakerGroup: "soundcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_soundcloud_comments",
				mcp.WithDescription("Get comments on a SoundCloud track"),
				mcp.WithNumber("track_id", mcp.Required(), mcp.Description("SoundCloud track ID")),
				mcp.WithNumber("limit", mcp.Description("Maximum results (default: 50)")),
			),
			Handler:             handleSoundCloudComments,
			Category:            "soundcloud",
			Subcategory:         "tracks",
			Tags:                []string{"soundcloud", "music", "comments", "social"},
			UseCases:            []string{"View track feedback", "Read timed comments"},
			Complexity:          "simple",
			CircuitBreakerGroup: "soundcloud",
		},
		// Download & Sync tools
		{
			Tool: mcp.NewTool("aftrs_soundcloud_download",
				mcp.WithDescription("Download a SoundCloud track or playlist using scdl (with yt-dlp fallback). Downloads to local path and optionally uploads to S3."),
				mcp.WithString("url", mcp.Required(), mcp.Description("SoundCloud URL (track, playlist, or user likes)")),
				mcp.WithString("format", mcp.Description("Audio format: mp3 (default), aiff, wav, flac")),
				mcp.WithString("path", mcp.Description("Local download path (default: ~/Music/SoundCloud)")),
				mcp.WithBoolean("upload_s3", mcp.Description("Upload to S3 after download (default: false)")),
				mcp.WithBoolean("use_ytdlp", mcp.Description("Force use yt-dlp instead of scdl (default: false, uses scdl with yt-dlp fallback)")),
			),
			Handler:             handleSoundCloudDownload,
			Category:            "soundcloud",
			Subcategory:         "download",
			Tags:                []string{"soundcloud", "music", "download", "scdl", "yt-dlp"},
			UseCases:            []string{"Download track", "Download playlist", "Archive SoundCloud content"},
			Complexity:          "moderate",
			IsWrite:             true,
			CircuitBreakerGroup: "soundcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_soundcloud_download_likes",
				mcp.WithDescription("Download all liked tracks for a SoundCloud user using scdl"),
				mcp.WithString("username", mcp.Required(), mcp.Description("SoundCloud username (e.g., hairglasses)")),
				mcp.WithString("format", mcp.Description("Audio format: mp3 (default), aiff")),
				mcp.WithString("path", mcp.Description("Local download path (default: ~/Music/SoundCloud/{username}/likes)")),
				mcp.WithBoolean("upload_s3", mcp.Description("Upload to S3 after download (default: true)")),
			),
			Handler:             handleSoundCloudDownloadLikes,
			Category:            "soundcloud",
			Subcategory:         "download",
			Tags:                []string{"soundcloud", "music", "download", "likes", "scdl"},
			UseCases:            []string{"Backup liked tracks", "Sync likes to local library"},
			Complexity:          "moderate",
			IsWrite:             true,
			CircuitBreakerGroup: "soundcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_soundcloud_sync",
				mcp.WithDescription("Trigger full SoundCloud sync for a user (S3 → local). Uses the existing sync infrastructure with parallel workers."),
				mcp.WithString("username", mcp.Description("SoundCloud username to sync (default: hairglasses)")),
				mcp.WithBoolean("dry_run", mcp.Description("Preview without making changes (default: true)")),
			),
			Handler:             handleSoundCloudSync,
			Category:            "soundcloud",
			Subcategory:         "sync",
			Tags:                []string{"soundcloud", "music", "sync", "s3", "rekordbox"},
			UseCases:            []string{"Sync SoundCloud to Rekordbox", "Update local library from S3"},
			Complexity:          "complex",
			IsWrite:             true,
			CircuitBreakerGroup: "soundcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_soundcloud_check_tools",
				mcp.WithDescription("Check availability of download tools (scdl, yt-dlp, ffmpeg, aws) and their versions"),
			),
			Handler:             handleSoundCloudCheckTools,
			Category:            "soundcloud",
			Subcategory:         "status",
			Tags:                []string{"soundcloud", "tools", "diagnostics", "scdl", "yt-dlp"},
			UseCases:            []string{"Verify download tools are installed", "Debug download issues"},
			Complexity:          "simple",
			CircuitBreakerGroup: "soundcloud",
		},
		// Recent likes sync tools
		{
			Tool: mcp.NewTool("aftrs_soundcloud_recent_likes",
				mcp.WithDescription("Get the most recent SoundCloud likes for specified users from S3. Returns track names, dates, and paths for the latest synced likes."),
				mcp.WithArray("users", mcp.Required(), mcp.Description("List of SoundCloud usernames to check (e.g., [\"hairglasses\", \"luke-lasley\", \"freaq-show\"])"), func(schema map[string]any) { schema["items"] = map[string]any{"type": "string"} }),
				mcp.WithNumber("limit", mcp.Description("Number of recent tracks per user (default: 20, max: 100)")),
			),
			Handler:             handleSoundCloudRecentLikes,
			Category:            "soundcloud",
			Subcategory:         "sync",
			Tags:                []string{"soundcloud", "likes", "recent", "s3", "sync"},
			UseCases:            []string{"Check recent synced likes", "View latest downloads per user"},
			Complexity:          "simple",
			CircuitBreakerGroup: "soundcloud",
		},
		{
			Tool: mcp.NewTool("aftrs_soundcloud_import_recent_likes",
				mcp.WithDescription("Sync and import the most recent SoundCloud likes to Rekordbox as a dedicated playlist. Downloads from S3 to local, then imports to Rekordbox with a 'Recent Likes' playlist per user."),
				mcp.WithArray("users", mcp.Required(), mcp.Description("List of SoundCloud usernames (e.g., [\"hairglasses\", \"luke-lasley\", \"freaq-show\"])"), func(schema map[string]any) { schema["items"] = map[string]any{"type": "string"} }),
				mcp.WithNumber("limit", mcp.Description("Number of recent tracks per user to import (default: 20, max: 100)")),
				mcp.WithBoolean("dry_run", mcp.Description("Preview without making changes (default: false)")),
			),
			Handler:             handleSoundCloudImportRecentLikes,
			Category:            "soundcloud",
			Subcategory:         "sync",
			Tags:                []string{"soundcloud", "likes", "recent", "rekordbox", "import", "playlist"},
			UseCases:            []string{"Import recent likes to Rekordbox", "Create Recent Likes playlist", "Quick DJ prep"},
			Complexity:          "moderate",
			IsWrite:             true,
			CircuitBreakerGroup: "soundcloud",
		},
		// Unified pipeline tool
		{
			Tool: mcp.NewTool("aftrs_soundcloud_to_rekordbox",
				mcp.WithDescription("Complete pipeline: download all SoundCloud playlists, sync to Google Drive, and import to Rekordbox. Combines discovery, download, cloud sync, and DJ library import into a single automated workflow."),
				mcp.WithString("username", mcp.Required(), mcp.Description("SoundCloud username (e.g., hairglasses)")),
				mcp.WithString("format", mcp.Description("Audio format: mp3 (default), m4a, aiff, flac")),
				mcp.WithBoolean("dry_run", mcp.Description("Preview without making changes (default: false)")),
				mcp.WithBoolean("include_likes", mcp.Description("Include liked tracks in sync (default: true)")),
				mcp.WithBoolean("skip_gdrive", mcp.Description("Skip Google Drive upload step (default: false)")),
				mcp.WithBoolean("skip_rekordbox", mcp.Description("Skip Rekordbox import step (default: false)")),
				mcp.WithString("gdrive_mount_path", mcp.Description("Local path where Google Drive is mounted via rclone (default: ~/GDrive)")),
				mcp.WithString("gdrive_base_path", mcp.Description("Base path in Google Drive for DJ music (default: DJ Crates/SoundCloud)")),
			),
			Handler:             handleSoundCloudToRekordbox,
			Category:            "soundcloud",
			Subcategory:         "pipeline",
			Tags:                []string{"soundcloud", "rekordbox", "gdrive", "pipeline", "sync", "dj"},
			UseCases:            []string{"Sync SoundCloud to Rekordbox", "Automated DJ library update", "Cloud-backed music sync"},
			Complexity:          "complex",
			IsWrite:             true,
			CircuitBreakerGroup: "soundcloud",
		},
	}
}

func handleSoundCloudStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSoundCloudClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create SoundCloud client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get status: %w", err)), nil
	}

	return tools.JSONResult(status), nil
}

func handleSoundCloudHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSoundCloudClient()
	if err != nil {
		health := &clients.SoundCloudHealth{
			Score:          50,
			Status:         "degraded",
			Connected:      false,
			HasCredentials: false,
			Issues:         []string{err.Error()},
			Recommendations: []string{
				"Set SOUNDCLOUD_CLIENT_ID environment variable",
				"Optionally set SOUNDCLOUD_CLIENT_SECRET for OAuth access",
			},
		}
		return tools.JSONResult(health), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get health: %w", err)), nil
	}

	return tools.JSONResult(health), nil
}

func handleSoundCloudResolve(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSoundCloudClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create SoundCloud client: %w", err)), nil
	}

	url, errResult := tools.RequireStringParam(req, "url")
	if errResult != nil {
		return errResult, nil
	}

	resource, resourceType, err := client.Resolve(ctx, url)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to resolve URL: %w", err)), nil
	}

	result := map[string]any{
		"type":     resourceType,
		"resource": resource,
	}

	return tools.JSONResult(result), nil
}

func handleSoundCloudSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSoundCloudClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create SoundCloud client: %w", err)), nil
	}

	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	searchType := tools.OptionalStringParam(req, "type", "track")

	limit := tools.GetIntParam(req, "limit", 20)

	results, err := client.Search(ctx, query, searchType, limit)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("search failed: %w", err)), nil
	}

	// Add counts
	result := map[string]any{
		"query": query,
		"type":  searchType,
	}

	if len(results.Tracks) > 0 {
		result["tracks_count"] = len(results.Tracks)
		result["tracks"] = results.Tracks
	}
	if len(results.Users) > 0 {
		result["users_count"] = len(results.Users)
		result["users"] = results.Users
	}
	if len(results.Playlists) > 0 {
		result["playlists_count"] = len(results.Playlists)
		result["playlists"] = results.Playlists
	}

	return tools.JSONResult(result), nil
}

func handleSoundCloudTrack(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSoundCloudClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create SoundCloud client: %w", err)), nil
	}

	trackID := int64(tools.GetIntParam(req, "track_id", 0))
	if trackID == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("track_id is required")), nil
	}

	track, err := client.GetTrack(ctx, trackID)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get track: %w", err)), nil
	}

	// Add formatted duration
	result := map[string]any{
		"id":             track.ID,
		"title":          track.Title,
		"description":    track.Description,
		"duration":       clients.FormatDuration(track.Duration),
		"duration_ms":    track.Duration,
		"genre":          track.Genre,
		"bpm":            track.BPM,
		"key":            track.KeySignature,
		"waveform_url":   track.Waveform,
		"artwork_url":    track.ArtworkURL,
		"stream_url":     track.StreamURL,
		"download_url":   track.DownloadURL,
		"downloadable":   track.Downloadable,
		"user":           track.User,
		"playback_count": track.PlaybackCount,
		"likes_count":    track.LikesCount,
		"reposts_count":  track.RepostsCount,
		"comment_count":  track.CommentCount,
		"created_at":     track.CreatedAt,
		"permalink_url":  track.PermalinkURL,
		"label":          track.LabelName,
		"license":        track.License,
	}

	return tools.JSONResult(result), nil
}

func handleSoundCloudUser(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSoundCloudClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create SoundCloud client: %w", err)), nil
	}

	userID := int64(tools.GetIntParam(req, "user_id", 0))
	if userID == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("user_id is required")), nil
	}

	user, err := client.GetUser(ctx, userID)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get user: %w", err)), nil
	}

	return tools.JSONResult(user), nil
}

func handleSoundCloudUserTracks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSoundCloudClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create SoundCloud client: %w", err)), nil
	}

	userID := int64(tools.GetIntParam(req, "user_id", 0))
	if userID == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("user_id is required")), nil
	}

	limit := tools.GetIntParam(req, "limit", 50)

	tracks, err := client.GetUserTracks(ctx, userID, limit)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get user tracks: %w", err)), nil
	}

	result := map[string]any{
		"user_id": userID,
		"count":   len(tracks),
		"tracks":  tracks,
	}

	return tools.JSONResult(result), nil
}

func handleSoundCloudLikes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSoundCloudClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create SoundCloud client: %w", err)), nil
	}

	userID := int64(tools.GetIntParam(req, "user_id", 0))
	if userID == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("user_id is required")), nil
	}

	limit := tools.GetIntParam(req, "limit", 50)

	tracks, err := client.GetUserLikes(ctx, userID, limit)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get user likes: %w", err)), nil
	}

	result := map[string]any{
		"user_id": userID,
		"count":   len(tracks),
		"tracks":  tracks,
	}

	return tools.JSONResult(result), nil
}

func handleSoundCloudPlaylist(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSoundCloudClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create SoundCloud client: %w", err)), nil
	}

	playlistID := int64(tools.GetIntParam(req, "playlist_id", 0))
	if playlistID == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("playlist_id is required")), nil
	}

	playlist, err := client.GetPlaylist(ctx, playlistID)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get playlist: %w", err)), nil
	}

	result := map[string]any{
		"id":            playlist.ID,
		"title":         playlist.Title,
		"description":   playlist.Description,
		"duration":      clients.FormatDuration(playlist.Duration),
		"duration_ms":   playlist.Duration,
		"track_count":   playlist.TrackCount,
		"user":          playlist.User,
		"tracks":        playlist.Tracks,
		"artwork_url":   playlist.ArtworkURL,
		"likes_count":   playlist.LikesCount,
		"created_at":    playlist.CreatedAt,
		"permalink_url": playlist.PermalinkURL,
		"is_album":      playlist.IsAlbum,
		"genre":         playlist.Genre,
	}

	return tools.JSONResult(result), nil
}

func handleSoundCloudPlaylists(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSoundCloudClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create SoundCloud client: %w", err)), nil
	}

	userID := int64(tools.GetIntParam(req, "user_id", 0))
	if userID == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("user_id is required")), nil
	}

	limit := tools.GetIntParam(req, "limit", 50)

	playlists, err := client.GetUserPlaylists(ctx, userID, limit)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get user playlists: %w", err)), nil
	}

	result := map[string]any{
		"user_id":   userID,
		"count":     len(playlists),
		"playlists": playlists,
	}

	return tools.JSONResult(result), nil
}

func handleSoundCloudFollowers(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSoundCloudClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create SoundCloud client: %w", err)), nil
	}

	userID := int64(tools.GetIntParam(req, "user_id", 0))
	if userID == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("user_id is required")), nil
	}

	followType := tools.OptionalStringParam(req, "type", "followers")

	limit := tools.GetIntParam(req, "limit", 50)

	var users []clients.SoundCloudUser
	var err2 error

	if followType == "following" {
		users, err2 = client.GetUserFollowings(ctx, userID, limit)
	} else {
		users, err2 = client.GetUserFollowers(ctx, userID, limit)
	}

	if err2 != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get %s: %w", followType, err2)), nil
	}

	result := map[string]any{
		"user_id": userID,
		"type":    followType,
		"count":   len(users),
		"users":   users,
	}

	return tools.JSONResult(result), nil
}

func handleSoundCloudComments(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getSoundCloudClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create SoundCloud client: %w", err)), nil
	}

	trackID := int64(tools.GetIntParam(req, "track_id", 0))
	if trackID == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("track_id is required")), nil
	}

	limit := tools.GetIntParam(req, "limit", 50)

	comments, err := client.GetTrackComments(ctx, trackID, limit)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get comments: %w", err)), nil
	}

	result := map[string]any{
		"track_id": trackID,
		"count":    len(comments),
		"comments": comments,
	}

	return tools.JSONResult(result), nil
}

// Download handler - uses scdl with yt-dlp fallback
func handleSoundCloudDownload(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	scURL, errResult := tools.RequireStringParam(req, "url")
	if errResult != nil {
		return errResult, nil
	}

	format := tools.OptionalStringParam(req, "format", "mp3")

	path := tools.GetStringParam(req, "path")
	if path == "" {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, "Music", "SoundCloud")
	}

	uploadS3 := tools.GetBoolParam(req, "upload_s3", false)
	useYtdlp := tools.GetBoolParam(req, "use_ytdlp", false)

	// Ensure path exists
	if err := os.MkdirAll(path, 0755); err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to create directory: %w", err)), nil
	}

	result := map[string]any{
		"url":    scURL,
		"format": format,
		"path":   path,
	}

	var downloadErr error

	if useYtdlp {
		// Force yt-dlp
		downloadErr = downloadWithYtdlp(ctx, scURL, path, format)
		result["tool"] = "yt-dlp"
	} else {
		// Try scdl first, fallback to yt-dlp
		downloadErr = downloadWithScdl(ctx, scURL, path, format)
		result["tool"] = "scdl"

		if downloadErr != nil {
			// Fallback to yt-dlp
			downloadErr = downloadWithYtdlp(ctx, scURL, path, format)
			result["tool"] = "yt-dlp (fallback)"
		}
	}

	if downloadErr != nil {
		result["status"] = "failed"
		result["error"] = downloadErr.Error()
	} else {
		result["status"] = "success"

		// Upload to S3 if requested
		if uploadS3 {
			config := musicsync.DefaultConfig()
			s3Path := fmt.Sprintf("s3://%s/soundcloud/downloads/", config.S3Bucket)
			s3Err := uploadToS3(ctx, path, s3Path, config.AWSProfile)
			if s3Err != nil {
				result["s3_upload"] = "failed: " + s3Err.Error()
			} else {
				result["s3_upload"] = "success"
				result["s3_path"] = s3Path
			}
		}
	}

	return tools.JSONResult(result), nil
}

// Download likes handler
func handleSoundCloudDownloadLikes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	username, errResult := tools.RequireStringParam(req, "username")
	if errResult != nil {
		return errResult, nil
	}

	format := tools.OptionalStringParam(req, "format", "mp3")

	path := tools.GetStringParam(req, "path")
	if path == "" {
		home, _ := os.UserHomeDir()
		path = filepath.Join(home, "Music", "SoundCloud", username, "likes")
	}

	uploadS3 := tools.GetBoolParam(req, "upload_s3", true)

	// Ensure path exists
	if err := os.MkdirAll(path, 0755); err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to create directory: %w", err)), nil
	}

	likesURL := fmt.Sprintf("https://soundcloud.com/%s/likes", username)

	result := map[string]any{
		"username": username,
		"url":      likesURL,
		"format":   format,
		"path":     path,
	}

	// Use scdl for likes
	err := downloadWithScdl(ctx, likesURL, path, format)
	if err != nil {
		result["status"] = "failed"
		result["error"] = err.Error()
	} else {
		result["status"] = "success"

		// Count downloaded files
		files, _ := filepath.Glob(filepath.Join(path, "*.*"))
		result["files_count"] = len(files)

		// Upload to S3 if requested
		if uploadS3 {
			config := musicsync.DefaultConfig()
			s3Path := fmt.Sprintf("s3://%s/users/%s/soundcloud/likes/", config.S3Bucket, username)
			s3Err := uploadToS3(ctx, path, s3Path, config.AWSProfile)
			if s3Err != nil {
				result["s3_upload"] = "failed: " + s3Err.Error()
			} else {
				result["s3_upload"] = "success"
				result["s3_path"] = s3Path
			}
		}
	}

	return tools.JSONResult(result), nil
}

// Sync handler - wraps existing sync infrastructure
func handleSoundCloudSync(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	username := tools.OptionalStringParam(req, "username", "hairglasses")

	dryRun := tools.GetBoolParam(req, "dry_run", true)

	config := musicsync.DefaultConfig()
	config.DryRun = dryRun

	manager, err := musicsync.NewManager(config)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create sync manager: %w", err)), nil
	}

	results, err := manager.SyncSoundCloud(ctx, username)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("sync failed: %w", err)), nil
	}

	return tools.JSONResult(results), nil
}

// Check tools handler
func handleSoundCloudCheckTools(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	toolChecks := map[string]any{}

	// Check scdl
	scdlVersion, scdlErr := exec.CommandContext(ctx, "scdl", "--version").Output()
	if scdlErr != nil {
		toolChecks["scdl"] = map[string]any{"installed": false, "error": "scdl not found - install with: pip install scdl"}
	} else {
		toolChecks["scdl"] = map[string]any{"installed": true, "version": strings.TrimSpace(string(scdlVersion))}
	}

	// Check yt-dlp
	ytdlpVersion, ytdlpErr := exec.CommandContext(ctx, "yt-dlp", "--version").Output()
	if ytdlpErr != nil {
		toolChecks["yt-dlp"] = map[string]any{"installed": false, "error": "yt-dlp not found - install with: brew install yt-dlp"}
	} else {
		toolChecks["yt-dlp"] = map[string]any{"installed": true, "version": strings.TrimSpace(string(ytdlpVersion))}
	}

	// Check ffmpeg
	ffmpegVersion, ffmpegErr := exec.CommandContext(ctx, "ffmpeg", "-version").Output()
	if ffmpegErr != nil {
		toolChecks["ffmpeg"] = map[string]any{"installed": false, "error": "ffmpeg not found - install with: brew install ffmpeg"}
	} else {
		lines := strings.Split(string(ffmpegVersion), "\n")
		if len(lines) > 0 {
			toolChecks["ffmpeg"] = map[string]any{"installed": true, "version": lines[0]}
		}
	}

	// Check aws cli
	awsVersion, awsErr := exec.CommandContext(ctx, "aws", "--version").Output()
	if awsErr != nil {
		toolChecks["aws"] = map[string]any{"installed": false, "error": "aws cli not found - install with: brew install awscli"}
	} else {
		toolChecks["aws"] = map[string]any{"installed": true, "version": strings.TrimSpace(string(awsVersion))}
	}

	// Overall status
	allInstalled := true
	for _, check := range toolChecks {
		if checkMap, ok := check.(map[string]any); ok {
			if installed, ok := checkMap["installed"].(bool); ok && !installed {
				allInstalled = false
				break
			}
		}
	}

	result := map[string]any{
		"tools":         toolChecks,
		"all_installed": allInstalled,
	}

	if !allInstalled {
		result["recommendation"] = "Some tools are missing. Install them for full functionality."
	}

	return tools.JSONResult(result), nil
}

// Helper: download using scdl
func downloadWithScdl(ctx context.Context, url, path, format string) error {
	args := []string{"-l", url, "--path", path}

	switch format {
	case "mp3":
		args = append(args, "--onlymp3")
	case "aiff":
		args = append(args, "--original-art")
	}

	cmd := exec.CommandContext(ctx, "scdl", args...)
	cmd.Dir = path

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("scdl failed: %w - %s", err, string(output))
	}
	return nil
}

// Helper: download using yt-dlp
func downloadWithYtdlp(ctx context.Context, url, path, format string) error {
	outputTemplate := filepath.Join(path, "%(title)s.%(ext)s")

	args := []string{
		"-x", // Extract audio
		"-o", outputTemplate,
	}

	switch format {
	case "mp3":
		args = append(args, "--audio-format", "mp3", "--audio-quality", "0")
	case "aiff":
		args = append(args, "--audio-format", "aiff")
	case "wav":
		args = append(args, "--audio-format", "wav")
	case "flac":
		args = append(args, "--audio-format", "flac")
	default:
		args = append(args, "--audio-format", "mp3", "--audio-quality", "0")
	}

	args = append(args, url)

	cmd := exec.CommandContext(ctx, "yt-dlp", args...)
	cmd.Dir = path

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("yt-dlp failed: %w - %s", err, string(output))
	}
	return nil
}

// Helper: upload to S3
func uploadToS3(ctx context.Context, localPath, s3Path, awsProfile string) error {
	args := []string{"s3", "sync", localPath, s3Path, "--profile", awsProfile}

	cmd := exec.CommandContext(ctx, "aws", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("s3 sync failed: %w - %s", err, string(output))
	}
	return nil
}

// RecentTrack represents a recently synced track
type RecentTrack struct {
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
}

// UserRecentLikes represents recent likes for a user
type UserRecentLikes struct {
	Username string        `json:"username"`
	Count    int           `json:"count"`
	Tracks   []RecentTrack `json:"tracks"`
	Error    string        `json:"error,omitempty"`
}

// handleSoundCloudRecentLikes returns the most recent likes from S3 for specified users
func handleSoundCloudRecentLikes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Parse users array
	users := tools.GetStringArrayParam(req, "users")
	if len(users) == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("users is required (array of usernames)")), nil
	}

	limit := tools.GetIntParam(req, "limit", 20)
	if limit > 100 {
		limit = 100
	}

	config := musicsync.DefaultConfig()
	var results []UserRecentLikes

	for _, username := range users {
		userResult := UserRecentLikes{Username: username}

		// List files from S3 sorted by date
		s3Path := fmt.Sprintf("cr8-s3:cr8-music-storage/users/%s/soundcloud/likes/", username)
		args := []string{"lsl", s3Path, "--max-depth", "1"}

		cmd := exec.CommandContext(ctx, "rclone", args...)
		output, err := cmd.Output()
		if err != nil {
			userResult.Error = fmt.Sprintf("failed to list: %v", err)
			results = append(results, userResult)
			continue
		}

		// Parse output and sort by date (newest first)
		type fileInfo struct {
			size     int64
			date     string
			time     string
			name     string
			fullLine string
		}
		var files []fileInfo

		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			// Format: "  size date time name"
			parts := strings.Fields(line)
			if len(parts) < 4 {
				continue
			}
			// Skip non-audio files
			name := strings.Join(parts[3:], " ")
			if !strings.HasSuffix(strings.ToLower(name), ".m4a") &&
				!strings.HasSuffix(strings.ToLower(name), ".mp3") &&
				!strings.HasSuffix(strings.ToLower(name), ".aiff") {
				continue
			}
			// Skip partial files
			if strings.HasSuffix(name, ".part") {
				continue
			}

			var size int64
			fmt.Sscanf(parts[0], "%d", &size)

			files = append(files, fileInfo{
				size: size,
				date: parts[1],
				time: parts[2],
				name: name,
			})
		}

		// Sort by date+time descending (newest first)
		for i := 0; i < len(files)-1; i++ {
			for j := i + 1; j < len(files); j++ {
				if files[j].date+files[j].time > files[i].date+files[i].time {
					files[i], files[j] = files[j], files[i]
				}
			}
		}

		// Take top N
		if len(files) > limit {
			files = files[:limit]
		}

		for _, f := range files {
			userResult.Tracks = append(userResult.Tracks, RecentTrack{
				Name:     f.name,
				Path:     fmt.Sprintf("%s/soundcloud/likes/%s", config.S3Bucket, f.name),
				Size:     f.size,
				Modified: f.date + " " + f.time,
			})
		}
		userResult.Count = len(userResult.Tracks)
		results = append(results, userResult)
	}

	return tools.JSONResult(results), nil
}

// handleSoundCloudImportRecentLikes syncs recent likes from S3 and imports to Rekordbox
func handleSoundCloudImportRecentLikes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Parse users array
	users := tools.GetStringArrayParam(req, "users")
	if len(users) == 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("users is required (array of usernames)")), nil
	}

	limit := tools.GetIntParam(req, "limit", 20)
	if limit > 100 {
		limit = 100
	}
	dryRun := tools.GetBoolParam(req, "dry_run", false)

	home, _ := os.UserHomeDir()
	localRoot := filepath.Join(home, "Music", "rekordbox", "soundcloud")

	type ImportResult struct {
		Username     string   `json:"username"`
		TracksFound  int      `json:"tracks_found"`
		TracksSynced int      `json:"tracks_synced"`
		LocalPath    string   `json:"local_path"`
		Playlist     string   `json:"playlist"`
		Tracks       []string `json:"tracks"`
		Error        string   `json:"error,omitempty"`
	}

	var results []ImportResult

	for _, username := range users {
		result := ImportResult{
			Username: username,
			Playlist: fmt.Sprintf("Recent Likes - %s", getDisplayNameForUser(username)),
		}

		// Create local folder for this user's recent likes
		recentLikesPath := filepath.Join(localRoot, fmt.Sprintf("%s-recent-likes", username))
		result.LocalPath = recentLikesPath

		if !dryRun {
			if err := os.MkdirAll(recentLikesPath, 0755); err != nil {
				result.Error = fmt.Sprintf("create dir: %v", err)
				results = append(results, result)
				continue
			}
		}

		// Get list of recent files from S3
		s3Path := fmt.Sprintf("cr8-s3:cr8-music-storage/users/%s/soundcloud/likes/", username)
		args := []string{"lsl", s3Path, "--max-depth", "1"}

		cmd := exec.CommandContext(ctx, "rclone", args...)
		output, err := cmd.Output()
		if err != nil {
			result.Error = fmt.Sprintf("list s3: %v", err)
			results = append(results, result)
			continue
		}

		// Parse and sort by date
		type fileInfo struct {
			size int64
			date string
			time string
			name string
		}
		var files []fileInfo

		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			parts := strings.Fields(line)
			if len(parts) < 4 {
				continue
			}
			name := strings.Join(parts[3:], " ")
			// Only audio files, no partials
			if (!strings.HasSuffix(strings.ToLower(name), ".m4a") &&
				!strings.HasSuffix(strings.ToLower(name), ".mp3") &&
				!strings.HasSuffix(strings.ToLower(name), ".aiff")) ||
				strings.HasSuffix(name, ".part") {
				continue
			}

			var size int64
			fmt.Sscanf(parts[0], "%d", &size)
			files = append(files, fileInfo{size: size, date: parts[1], time: parts[2], name: name})
		}

		// Sort by date descending
		for i := 0; i < len(files)-1; i++ {
			for j := i + 1; j < len(files); j++ {
				if files[j].date+files[j].time > files[i].date+files[i].time {
					files[i], files[j] = files[j], files[i]
				}
			}
		}

		// Take top N
		if len(files) > limit {
			files = files[:limit]
		}
		result.TracksFound = len(files)

		// Build include filter for rclone
		var includeFiles []string
		for _, f := range files {
			includeFiles = append(includeFiles, f.name)
			result.Tracks = append(result.Tracks, f.name)
		}

		if dryRun {
			result.TracksSynced = len(files)
			results = append(results, result)
			continue
		}

		// Sync only these specific files from S3 to local
		if len(includeFiles) > 0 {
			// Use rclone copy with include filters
			copyArgs := []string{"copy", s3Path, recentLikesPath}
			for _, f := range includeFiles {
				copyArgs = append(copyArgs, "--include", f)
			}

			copyCmd := exec.CommandContext(ctx, "rclone", copyArgs...)
			if _, err := copyCmd.CombinedOutput(); err != nil {
				result.Error = fmt.Sprintf("sync: %v", err)
			} else {
				// Count synced files
				entries, _ := os.ReadDir(recentLikesPath)
				for _, e := range entries {
					if !e.IsDir() {
						result.TracksSynced++
					}
				}
			}
		}

		results = append(results, result)
	}

	// Build summary
	summary := map[string]any{
		"dry_run": dryRun,
		"users":   results,
	}

	if !dryRun {
		summary["next_step"] = "Files synced to local folders. Use aftrs_rekordbox_import_usb or open Rekordbox to import from these paths."
	}

	return tools.JSONResult(summary), nil
}

// getDisplayNameForUser returns display name for username
func getDisplayNameForUser(username string) string {
	displayNames := map[string]string{
		"hairglasses": "Hairglasses",
		"luke-lasley": "Luke",
		"freaq-show":  "Freaq Show",
	}
	if name, ok := displayNames[username]; ok {
		return name
	}
	return username
}
