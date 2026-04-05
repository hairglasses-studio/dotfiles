package sync

import (
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Module implements the ToolModule interface for music sync tools
type Module struct{}

func (m *Module) Name() string {
	return "sync"
}

func (m *Module) Description() string {
	return "Music sync tools for SoundCloud, Beatport, and Rekordbox"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("sync_all",
				mcp.WithDescription("Sync all music sources to Rekordbox. Downloads from SoundCloud and Beatport, then imports to Rekordbox playlists."),
				mcp.WithString("user",
					mcp.Description("Sync specific user only (hairglasses, freaq-show, rahul, luke-lasley, aidan, fogel, marissughdevelops)"),
				),
				mcp.WithBoolean("dry_run",
					mcp.Description("Preview without making changes (default: true)"),
				),
			),
			Handler:             handleSyncAll,
			Category:            "sync",
			Subcategory:         "all",
			Tags:                []string{"sync", "soundcloud", "beatport", "rekordbox", "music"},
			UseCases:            []string{"Full sync of all music sources", "Sync specific user"},
			Complexity:          tools.ComplexityComplex,
			CircuitBreakerGroup: "sync",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("sync_soundcloud",
				mcp.WithDescription("Sync SoundCloud likes and playlists to local storage."),
				mcp.WithString("user",
					mcp.Description("SoundCloud username to sync"),
				),
				mcp.WithBoolean("dry_run",
					mcp.Description("Preview without making changes"),
				),
			),
			Handler:             handleSyncSoundCloud,
			Category:            "sync",
			Subcategory:         "soundcloud",
			Tags:                []string{"sync", "soundcloud", "music", "playlists"},
			UseCases:            []string{"Sync SoundCloud likes", "Sync SoundCloud playlists"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "sync",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("sync_beatport",
				mcp.WithDescription("Sync Beatport playlists to local storage."),
				mcp.WithString("user",
					mcp.Description("Beatport username to sync"),
				),
				mcp.WithBoolean("dry_run",
					mcp.Description("Preview without making changes"),
				),
			),
			Handler:             handleSyncBeatport,
			Category:            "sync",
			Subcategory:         "beatport",
			Tags:                []string{"sync", "beatport", "music", "download"},
			UseCases:            []string{"Sync Beatport purchases", "Download purchased tracks"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "sync",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("sync_rekordbox",
				mcp.WithDescription("Import pending music files to Rekordbox playlists."),
				mcp.WithBoolean("dry_run",
					mcp.Description("Preview without making changes"),
				),
			),
			Handler:             handleSyncRekordbox,
			Category:            "sync",
			Subcategory:         "rekordbox",
			Tags:                []string{"sync", "rekordbox", "import", "dj"},
			UseCases:            []string{"Import tracks to Rekordbox", "Update Rekordbox playlists"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "sync",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("sync_status",
				mcp.WithDescription("Show current sync status for all services."),
			),
			Handler:             handleSyncStatus,
			Category:            "sync",
			Subcategory:         "status",
			Tags:                []string{"sync", "status", "monitoring"},
			UseCases:            []string{"Check sync status", "View last sync times"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "sync",
		},
		{
			Tool: mcp.NewTool("sync_health",
				mcp.WithDescription("Check health of all sync dependencies (AWS CLI, S3, DynamoDB, ffmpeg). Returns status for each service and circuit breaker states."),
				mcp.WithString("reset_circuit_breaker",
					mcp.Description("Reset circuit breaker for a specific service (soundcloud, beatport, rekordbox) or 'all' to reset all"),
				),
			),
			Handler:             handleSyncHealth,
			Category:            "sync",
			Subcategory:         "health",
			Tags:                []string{"sync", "health", "monitoring", "aws", "s3", "circuit-breaker"},
			UseCases:            []string{"Check service health", "Diagnose sync issues", "Reset circuit breakers"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "sync",
		},
		{
			Tool: mcp.NewTool("sync_add_user",
				mcp.WithDescription("Add a new SoundCloud user for conversion tracking. Downloads likes and all public playlists to S3."),
				mcp.WithString("username",
					mcp.Description("SoundCloud username (from URL like soundcloud.com/username)"),
					mcp.Required(),
				),
				mcp.WithString("display_name",
					mcp.Description("Display name for Rekordbox folders"),
				),
				mcp.WithBoolean("download",
					mcp.Description("Start downloading immediately (default: true)"),
				),
			),
			Handler:             handleAddUser,
			Category:            "sync",
			Subcategory:         "users",
			Tags:                []string{"sync", "soundcloud", "user", "add"},
			UseCases:            []string{"Add new SoundCloud user", "Track new artist"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "sync",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("sync_list_users",
				mcp.WithDescription("List all tracked SoundCloud users and their sync status."),
			),
			Handler:             handleListUsers,
			Category:            "sync",
			Subcategory:         "users",
			Tags:                []string{"sync", "users", "list"},
			UseCases:            []string{"View tracked users", "Check user sync status"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "sync",
		},
		{
			Tool: mcp.NewTool("sync_discover_playlists",
				mcp.WithDescription("Discover and list all public playlists for a SoundCloud user."),
				mcp.WithString("username",
					mcp.Description("SoundCloud username"),
					mcp.Required(),
				),
			),
			Handler:             handleDiscoverPlaylists,
			Category:            "sync",
			Subcategory:         "discovery",
			Tags:                []string{"sync", "soundcloud", "playlists", "discover"},
			UseCases:            []string{"Find user playlists", "Discover new content"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "sync",
		},
	}
}
