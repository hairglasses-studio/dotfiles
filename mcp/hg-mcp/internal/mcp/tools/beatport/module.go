// Package beatport provides MCP tools for Beatport API integration.
package beatport

import (
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for Beatport tools
type Module struct{}

func (m *Module) Name() string {
	return "beatport"
}

func (m *Module) Description() string {
	return "Beatport API tools for track search, metadata enrichment, and chart browsing"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		// Search
		{
			Tool: mcp.NewTool("beatport_search",
				mcp.WithDescription("Search Beatport for tracks by artist and/or title. Returns track metadata including BPM, key, genre, label."),
				mcp.WithString("query",
					mcp.Required(),
					mcp.Description("Search query (artist name, track title, or both)"),
				),
				mcp.WithNumber("limit",
					mcp.Description("Maximum results to return (default: 10, max: 50)"),
				),
			),
			Handler:             handleSearch,
			Category:            "beatport",
			Subcategory:         "search",
			Tags:                []string{"beatport", "search", "tracks", "metadata", "music"},
			UseCases:            []string{"Search for track metadata", "Find BPM and key", "Lookup by artist/title"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "beatport",
		},

		// Track Metadata
		{
			Tool: mcp.NewTool("beatport_track",
				mcp.WithDescription("Get detailed metadata for a specific Beatport track by ID"),
				mcp.WithNumber("track_id",
					mcp.Required(),
					mcp.Description("Beatport track ID"),
				),
			),
			Handler:             handleGetTrack,
			Category:            "beatport",
			Subcategory:         "metadata",
			Tags:                []string{"beatport", "track", "metadata", "details"},
			UseCases:            []string{"Get full track details", "Lookup by Beatport ID"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "beatport",
		},

		// Charts
		{
			Tool: mcp.NewTool("beatport_charts",
				mcp.WithDescription("List Beatport charts, optionally filtered by genre"),
				mcp.WithString("genre",
					mcp.Description("Genre slug to filter by (e.g., 'techno', 'house', 'drum-and-bass')"),
				),
			),
			Handler:             handleCharts,
			Category:            "beatport",
			Subcategory:         "charts",
			Tags:                []string{"beatport", "charts", "trending", "top"},
			UseCases:            []string{"Browse genre charts", "Find trending tracks"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "beatport",
		},

		// Chart Tracks
		{
			Tool: mcp.NewTool("beatport_chart_tracks",
				mcp.WithDescription("Get tracks from a specific Beatport chart"),
				mcp.WithNumber("chart_id",
					mcp.Required(),
					mcp.Description("Beatport chart ID"),
				),
			),
			Handler:             handleChartTracks,
			Category:            "beatport",
			Subcategory:         "charts",
			Tags:                []string{"beatport", "chart", "tracks", "top"},
			UseCases:            []string{"Get chart tracks", "View top tracks in genre"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "beatport",
		},

		// Genres
		{
			Tool: mcp.NewTool("beatport_genres",
				mcp.WithDescription("List all available Beatport genres"),
			),
			Handler:             handleGenres,
			Category:            "beatport",
			Subcategory:         "metadata",
			Tags:                []string{"beatport", "genres", "categories"},
			UseCases:            []string{"Browse genres", "Get genre list for filtering"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "beatport",
		},

		// Enrichment
		{
			Tool: mcp.NewTool("beatport_enrich",
				mcp.WithDescription("Enrich a CR8 track with Beatport metadata. Searches by artist/title and updates DynamoDB with BPM, key, genre, label, etc."),
				mcp.WithString("track_id",
					mcp.Required(),
					mcp.Description("CR8 track ID to enrich"),
				),
				mcp.WithBoolean("dry_run",
					mcp.Description("Preview without saving to database (default: true)"),
				),
			),
			Handler:             handleEnrich,
			Category:            "beatport",
			Subcategory:         "enrichment",
			Tags:                []string{"beatport", "enrich", "metadata", "cr8", "dynamodb"},
			UseCases:            []string{"Add BPM/key to tracks", "Enrich library metadata"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "beatport",
			IsWrite:             true,
		},

		// Batch Enrichment
		{
			Tool: mcp.NewTool("beatport_batch_enrich",
				mcp.WithDescription("Batch enrich multiple CR8 tracks with Beatport metadata. Processes tracks missing BPM/key data."),
				mcp.WithNumber("limit",
					mcp.Description("Maximum tracks to process (default: 10)"),
				),
				mcp.WithBoolean("dry_run",
					mcp.Description("Preview without saving to database (default: true)"),
				),
				mcp.WithString("filter",
					mcp.Description("Filter tracks: 'missing_bpm', 'missing_key', 'missing_any' (default: missing_any)"),
				),
			),
			Handler:             handleBatchEnrich,
			Category:            "beatport",
			Subcategory:         "enrichment",
			Tags:                []string{"beatport", "enrich", "batch", "cr8", "metadata"},
			UseCases:            []string{"Bulk enrich library", "Fill missing metadata"},
			Complexity:          tools.ComplexityComplex,
			CircuitBreakerGroup: "beatport",
			IsWrite:             true,
		},

		// Auth Status
		{
			Tool: mcp.NewTool("beatport_auth_status",
				mcp.WithDescription("Check Beatport authentication status and token validity"),
			),
			Handler:             handleAuthStatus,
			Category:            "beatport",
			Subcategory:         "status",
			Tags:                []string{"beatport", "auth", "status", "tokens"},
			UseCases:            []string{"Check auth status", "Verify API access"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "beatport",
		},

		// Playlist Sync
		{
			Tool: mcp.NewTool("beatport_sync_playlist",
				mcp.WithDescription("Sync a Beatport playlist to CR8 library. Downloads tracks as FLAC, converts to AIFF (Rekordbox-compatible), uploads to S3, and records in DynamoDB."),
				mcp.WithNumber("playlist_id",
					mcp.Required(),
					mcp.Description("Beatport playlist ID to sync"),
				),
				mcp.WithString("quality",
					mcp.Description("Download quality: 'lossless' (FLAC), 'medium' (256 AAC), 'low' (128 AAC). Default: lossless"),
				),
				mcp.WithString("format",
					mcp.Description("Output format: 'flac', 'aiff', 'wav'. Default: aiff (best for Rekordbox)"),
				),
				mcp.WithBoolean("dry_run",
					mcp.Description("Preview sync without downloading (default: true)"),
				),
				mcp.WithNumber("limit",
					mcp.Description("Maximum tracks to sync (default: all)"),
				),
			),
			Handler:             handleSyncPlaylist,
			Category:            "beatport",
			Subcategory:         "sync",
			Tags:                []string{"beatport", "sync", "playlist", "download", "cr8", "rekordbox"},
			UseCases:            []string{"Sync Beatport playlist to library", "Download tracks for DJing"},
			Complexity:          tools.ComplexityComplex,
			CircuitBreakerGroup: "beatport",
			IsWrite:             true,
		},

		// Get Playlist Info
		{
			Tool: mcp.NewTool("beatport_playlist",
				mcp.WithDescription("Get information about a Beatport playlist"),
				mcp.WithNumber("playlist_id",
					mcp.Required(),
					mcp.Description("Beatport playlist ID"),
				),
				mcp.WithBoolean("include_tracks",
					mcp.Description("Include track listing (default: false)"),
				),
			),
			Handler:             handleGetPlaylist,
			Category:            "beatport",
			Subcategory:         "playlist",
			Tags:                []string{"beatport", "playlist", "info"},
			UseCases:            []string{"View playlist details", "Preview before sync"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "beatport",
		},

		// Download Single Track
		{
			Tool: mcp.NewTool("beatport_download",
				mcp.WithDescription("Download a single track from Beatport to local storage or S3"),
				mcp.WithNumber("track_id",
					mcp.Required(),
					mcp.Description("Beatport track ID"),
				),
				mcp.WithString("quality",
					mcp.Description("Download quality: 'lossless', 'medium', 'low'. Default: lossless"),
				),
				mcp.WithString("format",
					mcp.Description("Output format: 'flac', 'aiff', 'wav'. Default: aiff"),
				),
				mcp.WithBoolean("upload_s3",
					mcp.Description("Upload to S3 after download (default: true)"),
				),
			),
			Handler:             handleDownloadTrack,
			Category:            "beatport",
			Subcategory:         "download",
			Tags:                []string{"beatport", "download", "track", "flac", "aiff"},
			UseCases:            []string{"Download single track", "Add track to library"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "beatport",
			IsWrite:             true,
		},

		// Sync Likes/Follows
		{
			Tool: mcp.NewTool("beatport_sync_likes",
				mcp.WithDescription("Like all tracks and follow all artists in a Beatport playlist using browser automation. Runs in background with resume support."),
				mcp.WithString("playlist_id",
					mcp.Required(),
					mcp.Description("Beatport playlist ID"),
				),
				mcp.WithString("mode",
					mcp.Description("Sync mode: 'both' (default), 'tracks' (likes only), 'artists' (follows only), 'status' (check progress)"),
				),
			),
			Handler:             handleSyncLikes,
			Category:            "beatport",
			Subcategory:         "sync",
			Tags:                []string{"beatport", "likes", "follows", "playlist", "automation"},
			UseCases:            []string{"Like all playlist tracks", "Follow all artists", "Build Beatport recommendations"},
			Complexity:          tools.ComplexityComplex,
			CircuitBreakerGroup: "beatport",
			IsWrite:             true,
		},
	}
}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
