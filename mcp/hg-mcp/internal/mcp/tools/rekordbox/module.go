// Package rekordbox provides Rekordbox DJ library integration tools for hg-mcp.
package rekordbox

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/config"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for Rekordbox integration
type Module struct{}

var getRekordboxClient = tools.LazyClient(clients.GetRekordboxClient)

func (m *Module) Name() string {
	return "rekordbox"
}

func (m *Module) Description() string {
	return "Rekordbox DJ library integration with OneLibrary cloud sync support"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		// ============================================================================
		// Phase 1: Library Analysis Tools
		// ============================================================================
		{
			Tool: mcp.NewTool("aftrs_rekordbox_stats",
				mcp.WithDescription("Get Rekordbox library statistics: track count, playlists, cues, analysis data."),
			),
			Handler:             handleStats,
			Category:            "rekordbox",
			Subcategory:         "library",
			Tags:                []string{"rekordbox", "stats", "library", "dj"},
			UseCases:            []string{"View library overview", "Check collection size"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rekordbox",
		},
		{
			Tool: mcp.NewTool("aftrs_rekordbox_playlists",
				mcp.WithDescription("List all Rekordbox playlists and folders with track counts."),
				mcp.WithBoolean("tree", mcp.Description("Show as hierarchical tree (default: true)")),
			),
			Handler:             handlePlaylists,
			Category:            "rekordbox",
			Subcategory:         "library",
			Tags:                []string{"rekordbox", "playlists", "folders", "organization"},
			UseCases:            []string{"Browse playlists", "View folder structure"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rekordbox",
		},
		{
			Tool: mcp.NewTool("aftrs_rekordbox_search",
				mcp.WithDescription("Search tracks in Rekordbox library by title, artist, BPM, or key."),
				mcp.WithString("query", mcp.Description("Search query (title, artist)")),
				mcp.WithNumber("bpm_min", mcp.Description("Minimum BPM")),
				mcp.WithNumber("bpm_max", mcp.Description("Maximum BPM")),
				mcp.WithString("key", mcp.Description("Musical key (e.g., 8A, 1B, Am, C)")),
				mcp.WithNumber("limit", mcp.Description("Max results (default: 20)")),
			),
			Handler:             handleSearch,
			Category:            "rekordbox",
			Subcategory:         "library",
			Tags:                []string{"rekordbox", "search", "tracks", "find"},
			UseCases:            []string{"Find tracks for sets", "Filter by BPM/key"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "rekordbox",
		},
		{
			Tool: mcp.NewTool("aftrs_rekordbox_track_info",
				mcp.WithDescription("Get detailed track information including cue points, loops, and beat grid."),
				mcp.WithNumber("track_id", mcp.Description("Track ID from search results")),
				mcp.WithString("path", mcp.Description("Or path to audio file")),
			),
			Handler:             handleTrackInfo,
			Category:            "rekordbox",
			Subcategory:         "library",
			Tags:                []string{"rekordbox", "track", "cues", "beatgrid", "details"},
			UseCases:            []string{"View cue points", "Check beat grid", "See track metadata"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "rekordbox",
		},
		{
			Tool: mcp.NewTool("aftrs_rekordbox_recent",
				mcp.WithDescription("List recently added or modified tracks."),
				mcp.WithNumber("days", mcp.Description("Number of days to look back (default: 7)")),
				mcp.WithNumber("limit", mcp.Description("Max results (default: 50)")),
			),
			Handler:             handleRecent,
			Category:            "rekordbox",
			Subcategory:         "library",
			Tags:                []string{"rekordbox", "recent", "new", "tracks"},
			UseCases:            []string{"See new additions", "Track recent imports"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rekordbox",
		},

		// ============================================================================
		// Phase 2: USB Export Tools
		// ============================================================================
		{
			Tool: mcp.NewTool("aftrs_rekordbox_usb_status",
				mcp.WithDescription("Check status of connected USB drives with Rekordbox exports."),
			),
			Handler:             handleUSBStatus,
			Category:            "rekordbox",
			Subcategory:         "usb",
			Tags:                []string{"rekordbox", "usb", "export", "status"},
			UseCases:            []string{"Check USB exports", "Verify gig prep"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rekordbox",
		},
		{
			Tool: mcp.NewTool("aftrs_rekordbox_usb_validate",
				mcp.WithDescription("Validate USB structure for XDJ-1000 MK2 compatibility."),
				mcp.WithString("path", mcp.Description("USB mount path (e.g., /Volumes/DJ_USB)")),
			),
			Handler:             handleUSBValidate,
			Category:            "rekordbox",
			Subcategory:         "usb",
			Tags:                []string{"rekordbox", "usb", "validate", "xdj"},
			UseCases:            []string{"Verify USB compatibility", "Check for issues"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "rekordbox",
		},

		// ============================================================================
		// Phase 3: Cloud Sync Tools
		// ============================================================================
		{
			Tool: mcp.NewTool("aftrs_rekordbox_cloud_status",
				mcp.WithDescription("Check Google Drive cloud sync status for Rekordbox library."),
			),
			Handler:             handleCloudStatus,
			Category:            "rekordbox",
			Subcategory:         "cloud",
			Tags:                []string{"rekordbox", "cloud", "gdrive", "sync"},
			UseCases:            []string{"Check sync status", "Monitor cloud library"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rekordbox",
		},
		{
			Tool: mcp.NewTool("aftrs_rekordbox_s3_sync",
				mcp.WithDescription("Sync Rekordbox library to S3 bucket for multi-machine access."),
				mcp.WithBoolean("dry_run", mcp.Description("Preview changes without syncing")),
				mcp.WithString("direction", mcp.Description("Sync direction: up, down, both (default: up)")),
			),
			Handler:             handleS3Sync,
			Category:            "rekordbox",
			Subcategory:         "cloud",
			Tags:                []string{"rekordbox", "s3", "sync", "backup"},
			UseCases:            []string{"Backup library", "Multi-machine sync"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "rekordbox",
			IsWrite:             true,
		},

		// ============================================================================
		// Phase 4: Multi-User Tools
		// ============================================================================
		{
			Tool: mcp.NewTool("aftrs_rekordbox_share_playlist",
				mcp.WithDescription("Share a playlist with another user via S3. Exports playlist to XML and uploads."),
				mcp.WithString("playlist", mcp.Required(), mcp.Description("Playlist name to share")),
				mcp.WithString("to_user", mcp.Required(), mcp.Description("Target user (e.g., luke-lasley, hairglasses)")),
				mcp.WithBoolean("include_tracks", mcp.Description("Include track files in share (default: false, metadata only)")),
			),
			Handler:             handleSharePlaylist,
			Category:            "rekordbox",
			Subcategory:         "sharing",
			Tags:                []string{"rekordbox", "share", "playlist", "multi-user"},
			UseCases:            []string{"Share playlists between DJs", "Collaborate on sets"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "rekordbox",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_rekordbox_list_shared",
				mcp.WithDescription("List playlists shared with you from other users."),
				mcp.WithString("from_user", mcp.Description("Filter by source user (optional)")),
			),
			Handler:             handleListShared,
			Category:            "rekordbox",
			Subcategory:         "sharing",
			Tags:                []string{"rekordbox", "shared", "playlists", "import"},
			UseCases:            []string{"See available shared playlists", "Check for updates"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rekordbox",
		},
		{
			Tool: mcp.NewTool("aftrs_rekordbox_import_shared",
				mcp.WithDescription("Import a shared playlist from another user."),
				mcp.WithString("playlist", mcp.Required(), mcp.Description("Shared playlist name to import")),
				mcp.WithString("from_user", mcp.Required(), mcp.Description("Source user who shared the playlist")),
				mcp.WithBoolean("dry_run", mcp.Description("Preview import without making changes")),
			),
			Handler:             handleImportShared,
			Category:            "rekordbox",
			Subcategory:         "sharing",
			Tags:                []string{"rekordbox", "import", "shared", "playlist"},
			UseCases:            []string{"Import collaborator playlists", "Sync shared sets"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "rekordbox",
			IsWrite:             true,
		},

		// ============================================================================
		// Phase 5: Library Maintenance & Automation Tools
		// ============================================================================
		{
			Tool: mcp.NewTool("aftrs_rekordbox_maintenance_status",
				mcp.WithDescription("Get comprehensive Rekordbox library status including track count, analysis progress, and duplicate playlists."),
			),
			Handler:             handleMaintenanceStatus,
			Category:            "rekordbox",
			Subcategory:         "maintenance",
			Tags:                []string{"rekordbox", "status", "maintenance", "analysis"},
			UseCases:            []string{"Check library health", "Monitor analysis progress", "Find duplicate playlists"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rekordbox",
		},
		{
			Tool: mcp.NewTool("aftrs_rekordbox_usb_info",
				mcp.WithDescription("Get information about connected USB drives with Rekordbox exports, including audio file count and analysis data."),
			),
			Handler:             handleUSBInfo,
			Category:            "rekordbox",
			Subcategory:         "maintenance",
			Tags:                []string{"rekordbox", "usb", "import", "gig"},
			UseCases:            []string{"Check USB drive contents", "Verify gig prep", "Pre-import inspection"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rekordbox",
		},
		{
			Tool: mcp.NewTool("aftrs_rekordbox_import_usb",
				mcp.WithDescription("Import audio files and analysis data from USB drive to local Rekordbox library."),
				mcp.WithBoolean("copy_analysis", mcp.Description("Also copy analysis files (waveforms, beat grids) to avoid re-analysis (default: true)")),
			),
			Handler:             handleImportUSB,
			Category:            "rekordbox",
			Subcategory:         "maintenance",
			Tags:                []string{"rekordbox", "usb", "import", "tracks"},
			UseCases:            []string{"Import from USB", "Restore from backup", "Copy gig data"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "rekordbox",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_rekordbox_cleanup_duplicates",
				mcp.WithDescription("Remove duplicate playlists (those ending with ' (1)') from Rekordbox library."),
				mcp.WithBoolean("use_ui", mcp.Description("Use UI automation instead of direct database (safer but requires Rekordbox running)")),
			),
			Handler:             handleCleanupDuplicates,
			Category:            "rekordbox",
			Subcategory:         "maintenance",
			Tags:                []string{"rekordbox", "cleanup", "playlists", "duplicates"},
			UseCases:            []string{"Clean up after imports", "Remove duplicate playlists", "Library maintenance"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "rekordbox",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_rekordbox_trigger_analysis",
				mcp.WithDescription("Trigger track analysis in Rekordbox to calculate BPM, key, and generate waveforms for unanalyzed tracks."),
			),
			Handler:             handleTriggerAnalysis,
			Category:            "rekordbox",
			Subcategory:         "maintenance",
			Tags:                []string{"rekordbox", "analysis", "bpm", "key", "waveform"},
			UseCases:            []string{"Analyze new imports", "Fix missing analysis", "Prepare for gigs"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "rekordbox",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_rekordbox_health",
				mcp.WithDescription("Check Rekordbox integration health: database connectivity, Python scripts, and overall status."),
			),
			Handler:             handleHealth,
			Category:            "rekordbox",
			Subcategory:         "maintenance",
			Tags:                []string{"rekordbox", "health", "diagnostics", "status"},
			UseCases:            []string{"Troubleshoot issues", "Verify setup", "Check dependencies"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rekordbox",
		},

		// ============================================================================
		// Phase 6: Playback & History Tools (for Showkontrol integration)
		// ============================================================================
		{
			Tool: mcp.NewTool("aftrs_rekordbox_now_playing",
				mcp.WithDescription("Get the currently playing or most recently played track from Rekordbox. Essential for live DJ monitoring and Showkontrol integration."),
			),
			Handler:             handleNowPlaying,
			Category:            "rekordbox",
			Subcategory:         "playback",
			Tags:                []string{"rekordbox", "now_playing", "live", "showkontrol"},
			UseCases:            []string{"Live DJ monitoring", "Showkontrol integration", "Track current playback"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rekordbox",
		},
		{
			Tool: mcp.NewTool("aftrs_rekordbox_history",
				mcp.WithDescription("Get play history from Rekordbox - recently played tracks."),
				mcp.WithNumber("limit", mcp.Description("Max results (default: 10)")),
			),
			Handler:             handleHistory,
			Category:            "rekordbox",
			Subcategory:         "playback",
			Tags:                []string{"rekordbox", "history", "played", "tracks"},
			UseCases:            []string{"View set history", "Track what was played", "Generate setlists"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rekordbox",
		},
		{
			Tool: mcp.NewTool("aftrs_rekordbox_sessions",
				mcp.WithDescription("List DJ sessions from Rekordbox history."),
				mcp.WithNumber("limit", mcp.Description("Max sessions to return (default: 10)")),
			),
			Handler:             handleSessions,
			Category:            "rekordbox",
			Subcategory:         "playback",
			Tags:                []string{"rekordbox", "sessions", "sets", "history"},
			UseCases:            []string{"View past DJ sets", "Export setlists", "Session management"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rekordbox",
		},
		{
			Tool: mcp.NewTool("aftrs_rekordbox_session_tracks",
				mcp.WithDescription("Get tracks from a specific DJ session."),
				mcp.WithNumber("session_id", mcp.Required(), mcp.Description("Session ID from aftrs_rekordbox_sessions")),
			),
			Handler:             handleSessionTracks,
			Category:            "rekordbox",
			Subcategory:         "playback",
			Tags:                []string{"rekordbox", "session", "tracks", "setlist"},
			UseCases:            []string{"Export setlist", "View session details", "Generate playlist from session"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rekordbox",
		},
	}
}

// pythonScript runs the rekordbox Python script with given arguments
func pythonScript(ctx context.Context, action string, args map[string]interface{}) (map[string]interface{}, error) {
	// Find the script
	scriptPath := filepath.Join(config.Get().Home, "hairglasses-studio", "hg-mcp", "scripts", "rekordbox_query.py")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		// Try relative path
		scriptPath = "scripts/rekordbox_query.py"
	}

	// Build command
	argsJSON, _ := json.Marshal(args)
	cmd := exec.CommandContext(ctx, "python3", scriptPath, "--action", action, "--args", string(argsJSON))

	// Set timeout
	ctxTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	cmd = exec.CommandContext(ctxTimeout, cmd.Args[0], cmd.Args[1:]...)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("script error: %s", string(exitErr.Stderr))
		}
		return nil, fmt.Errorf("failed to run script: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse output: %w", err)
	}

	if errMsg, ok := result["error"].(string); ok {
		return nil, fmt.Errorf("%s", errMsg)
	}

	return result, nil
}

// handleStats returns library statistics
func handleStats(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	result, err := pythonScript(ctx, "stats", nil)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Rekordbox Library Stats\n\n")

	if tracks, ok := result["tracks"].(float64); ok {
		sb.WriteString(fmt.Sprintf("**Total Tracks:** %d\n", int(tracks)))
	}
	if playlists, ok := result["playlists"].(float64); ok {
		sb.WriteString(fmt.Sprintf("**Playlists:** %d\n", int(playlists)))
	}
	if folders, ok := result["folders"].(float64); ok {
		sb.WriteString(fmt.Sprintf("**Folders:** %d\n", int(folders)))
	}
	if cues, ok := result["cue_points"].(float64); ok {
		sb.WriteString(fmt.Sprintf("**Cue Points:** %d\n", int(cues)))
	}
	if hotCues, ok := result["hot_cues"].(float64); ok {
		sb.WriteString(fmt.Sprintf("**Hot Cues:** %d\n", int(hotCues)))
	}
	if loops, ok := result["loops"].(float64); ok {
		sb.WriteString(fmt.Sprintf("**Loops:** %d\n", int(loops)))
	}

	// Rekordbox version info
	if version, ok := result["version"].(string); ok {
		sb.WriteString(fmt.Sprintf("\n**Rekordbox Version:** %s\n", version))
	}

	// Database info
	if dbPath, ok := result["db_path"].(string); ok {
		sb.WriteString(fmt.Sprintf("**Database:** `%s`\n", dbPath))
	}

	return tools.TextResult(sb.String()), nil
}

// handlePlaylists lists all playlists
func handlePlaylists(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tree := tools.GetBoolParam(req, "tree", true)

	result, err := pythonScript(ctx, "playlists", map[string]interface{}{"tree": tree})
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Rekordbox Playlists\n\n")

	if playlists, ok := result["playlists"].([]interface{}); ok {
		for _, p := range playlists {
			if pl, ok := p.(map[string]interface{}); ok {
				name := pl["name"].(string)
				isFolder := false
				if f, ok := pl["is_folder"].(bool); ok {
					isFolder = f
				}
				depth := 0
				if d, ok := pl["depth"].(float64); ok {
					depth = int(d)
				}
				trackCount := 0
				if c, ok := pl["track_count"].(float64); ok {
					trackCount = int(c)
				}

				indent := strings.Repeat("  ", depth)
				if isFolder {
					sb.WriteString(fmt.Sprintf("%s📁 **%s**/\n", indent, name))
				} else {
					sb.WriteString(fmt.Sprintf("%s🎵 %s (%d tracks)\n", indent, name, trackCount))
				}
			}
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleSearch searches for tracks
func handleSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := map[string]interface{}{
		"query":   tools.GetStringParam(req, "query"),
		"bpm_min": tools.GetIntParam(req, "bpm_min", 0),
		"bpm_max": tools.GetIntParam(req, "bpm_max", 0),
		"key":     tools.GetStringParam(req, "key"),
		"limit":   tools.GetIntParam(req, "limit", 20),
	}

	result, err := pythonScript(ctx, "search", args)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Search Results\n\n")

	if tracks, ok := result["tracks"].([]interface{}); ok {
		if len(tracks) == 0 {
			sb.WriteString("*No tracks found*\n")
		} else {
			sb.WriteString("| ID | Title | Artist | BPM | Key | Duration |\n")
			sb.WriteString("|----|-------|--------|-----|-----|----------|\n")

			for _, t := range tracks {
				if track, ok := t.(map[string]interface{}); ok {
					id := int(track["id"].(float64))
					title := track["title"].(string)
					artist := ""
					if a, ok := track["artist"].(string); ok {
						artist = a
					}
					bpm := 0.0
					if b, ok := track["bpm"].(float64); ok {
						bpm = b
					}
					key := ""
					if k, ok := track["key"].(string); ok {
						key = k
					}
					duration := ""
					if d, ok := track["duration"].(string); ok {
						duration = d
					}

					sb.WriteString(fmt.Sprintf("| %d | %s | %s | %.1f | %s | %s |\n",
						id, title, artist, bpm, key, duration))
				}
			}

			sb.WriteString(fmt.Sprintf("\n*Found %d tracks*\n", len(tracks)))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleTrackInfo gets detailed track information
func handleTrackInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	trackID := tools.GetIntParam(req, "track_id", 0)
	path := tools.GetStringParam(req, "path")

	if trackID == 0 && path == "" {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("track_id or path required")), nil
	}

	args := map[string]interface{}{
		"track_id": trackID,
		"path":     path,
	}

	result, err := pythonScript(ctx, "track_info", args)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Track Details\n\n")

	if title, ok := result["title"].(string); ok {
		sb.WriteString(fmt.Sprintf("**Title:** %s\n", title))
	}
	if artist, ok := result["artist"].(string); ok {
		sb.WriteString(fmt.Sprintf("**Artist:** %s\n", artist))
	}
	if album, ok := result["album"].(string); ok {
		sb.WriteString(fmt.Sprintf("**Album:** %s\n", album))
	}
	if bpm, ok := result["bpm"].(float64); ok {
		sb.WriteString(fmt.Sprintf("**BPM:** %.1f\n", bpm))
	}
	if key, ok := result["key"].(string); ok {
		sb.WriteString(fmt.Sprintf("**Key:** %s\n", key))
	}
	if duration, ok := result["duration"].(string); ok {
		sb.WriteString(fmt.Sprintf("**Duration:** %s\n", duration))
	}

	// Cue points
	if cues, ok := result["cues"].([]interface{}); ok && len(cues) > 0 {
		sb.WriteString("\n## Cue Points\n\n")
		sb.WriteString("| # | Type | Name | Time |\n")
		sb.WriteString("|---|------|------|------|\n")
		for i, c := range cues {
			if cue, ok := c.(map[string]interface{}); ok {
				cueType := "Memory"
				if t, ok := cue["type"].(string); ok {
					cueType = t
				}
				name := ""
				if n, ok := cue["name"].(string); ok {
					name = n
				}
				time := ""
				if t, ok := cue["time"].(string); ok {
					time = t
				}
				sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n", i+1, cueType, name, time))
			}
		}
	}

	// File info
	if path, ok := result["path"].(string); ok {
		sb.WriteString(fmt.Sprintf("\n**File:** `%s`\n", path))
	}

	return tools.TextResult(sb.String()), nil
}

// handleRecent lists recently added tracks
func handleRecent(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	days := tools.GetIntParam(req, "days", 7)
	limit := tools.GetIntParam(req, "limit", 50)

	args := map[string]interface{}{
		"days":  days,
		"limit": limit,
	}

	result, err := pythonScript(ctx, "recent", args)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Recently Added Tracks (Last %d Days)\n\n", days))

	if tracks, ok := result["tracks"].([]interface{}); ok {
		if len(tracks) == 0 {
			sb.WriteString("*No recent tracks*\n")
		} else {
			sb.WriteString("| Date | Title | Artist | BPM |\n")
			sb.WriteString("|------|-------|--------|-----|\n")

			for _, t := range tracks {
				if track, ok := t.(map[string]interface{}); ok {
					title := track["title"].(string)
					artist := ""
					if a, ok := track["artist"].(string); ok {
						artist = a
					}
					bpm := 0.0
					if b, ok := track["bpm"].(float64); ok {
						bpm = b
					}
					date := ""
					if d, ok := track["date_added"].(string); ok {
						date = d
					}

					sb.WriteString(fmt.Sprintf("| %s | %s | %s | %.1f |\n",
						date, title, artist, bpm))
				}
			}

			sb.WriteString(fmt.Sprintf("\n*Found %d tracks*\n", len(tracks)))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleUSBStatus checks USB drives with Rekordbox exports
func handleUSBStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var sb strings.Builder
	sb.WriteString("# Rekordbox USB Drives\n\n")

	// Check /Volumes for drives with PIONEER folder
	volumes, err := os.ReadDir("/Volumes")
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	found := false
	for _, v := range volumes {
		if !v.IsDir() {
			continue
		}

		volumePath := filepath.Join("/Volumes", v.Name())
		pioneerPath := filepath.Join(volumePath, "PIONEER")

		if _, err := os.Stat(pioneerPath); err == nil {
			found = true
			sb.WriteString(fmt.Sprintf("## %s\n\n", v.Name()))
			sb.WriteString(fmt.Sprintf("**Path:** `%s`\n", volumePath))

			// Check for Device Library vs OneLibrary
			if _, err := os.Stat(filepath.Join(pioneerPath, "rekordbox")); err == nil {
				sb.WriteString("**Format:** Device Library (XDJ-1000 MK2 compatible)\n")
			}

			// Count tracks
			anlzPath := filepath.Join(pioneerPath, "USBANLZ")
			if entries, err := os.ReadDir(anlzPath); err == nil {
				sb.WriteString(fmt.Sprintf("**Analysis Files:** %d\n", len(entries)))
			}

			sb.WriteString("\n")
		}
	}

	if !found {
		sb.WriteString("*No Rekordbox USB drives detected*\n\n")
		sb.WriteString("Insert a USB drive with a Rekordbox export to see status.\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleUSBValidate validates USB structure for XDJ-1000 MK2
func handleUSBValidate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := tools.GetStringParam(req, "path")
	if path == "" {
		// Auto-detect
		volumes, _ := os.ReadDir("/Volumes")
		for _, v := range volumes {
			volumePath := filepath.Join("/Volumes", v.Name())
			if _, err := os.Stat(filepath.Join(volumePath, "PIONEER")); err == nil {
				path = volumePath
				break
			}
		}
	}

	if path == "" {
		return tools.CodedErrorResult(tools.ErrNotFound, fmt.Errorf("no USB path specified and none auto-detected")), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# USB Validation: %s\n\n", filepath.Base(path)))

	issues := []string{}
	warnings := []string{}

	// Required structure for XDJ-1000 MK2
	requiredPaths := []string{
		"PIONEER",
		"PIONEER/rekordbox",
		"PIONEER/USBANLZ",
	}

	for _, p := range requiredPaths {
		fullPath := filepath.Join(path, p)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			issues = append(issues, fmt.Sprintf("Missing: `%s`", p))
		}
	}

	// Check database
	dbPath := filepath.Join(path, "PIONEER", "rekordbox", "export.edb")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		issues = append(issues, "Missing export database (export.edb)")
	}

	// Check for analysis files
	anlzPath := filepath.Join(path, "PIONEER", "USBANLZ")
	if entries, err := os.ReadDir(anlzPath); err == nil && len(entries) == 0 {
		warnings = append(warnings, "USBANLZ folder is empty - tracks may not display waveforms")
	}

	// Report
	if len(issues) == 0 {
		sb.WriteString("✅ **Valid XDJ-1000 MK2 USB Export**\n\n")
	} else {
		sb.WriteString("❌ **Issues Found**\n\n")
		for _, issue := range issues {
			sb.WriteString(fmt.Sprintf("- %s\n", issue))
		}
		sb.WriteString("\n")
	}

	if len(warnings) > 0 {
		sb.WriteString("⚠️ **Warnings**\n\n")
		for _, warning := range warnings {
			sb.WriteString(fmt.Sprintf("- %s\n", warning))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleCloudStatus checks Google Drive sync status
func handleCloudStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var sb strings.Builder
	sb.WriteString("# Rekordbox Cloud Status\n\n")

	home := config.Get().Home

	// Check Google Drive paths
	gdrivePaths := []string{
		filepath.Join(home, "Google Drive", "My Drive", "rekordbox"),
		filepath.Join(home, "Library", "CloudStorage", "GoogleDrive-*", "My Drive", "rekordbox"),
	}

	found := false
	for _, pattern := range gdrivePaths {
		matches, _ := filepath.Glob(pattern)
		for _, match := range matches {
			if info, err := os.Stat(match); err == nil && info.IsDir() {
				found = true
				sb.WriteString(fmt.Sprintf("**Google Drive Path:** `%s`\n", match))

				// Count files
				fileCount := 0
				filepath.Walk(match, func(path string, info os.FileInfo, err error) error {
					if err == nil && !info.IsDir() {
						fileCount++
					}
					return nil
				})
				sb.WriteString(fmt.Sprintf("**Synced Files:** %d\n\n", fileCount))
				break
			}
		}
	}

	if !found {
		sb.WriteString("⚠️ **Google Drive Not Detected**\n\n")
		sb.WriteString("Rekordbox cloud sync uses Google Drive. Ensure:\n")
		sb.WriteString("1. Google Drive app is installed\n")
		sb.WriteString("2. Rekordbox cloud sync is enabled in Preferences\n")
		sb.WriteString("3. You have a Creative or Professional plan\n")
	}

	// Check Rekordbox database
	dbPath := filepath.Join(home, "Library", "Pioneer", "rekordbox", "master.db")
	if _, err := os.Stat(dbPath); err == nil {
		sb.WriteString(fmt.Sprintf("**Local Database:** `%s`\n", dbPath))
	}

	return tools.TextResult(sb.String()), nil
}

// handleS3Sync syncs library to S3
func handleS3Sync(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dryRun := tools.GetBoolParam(req, "dry_run", true)
	direction := tools.OptionalStringParam(req, "direction", "up")

	var sb strings.Builder
	sb.WriteString("# Rekordbox S3 Sync\n\n")

	// Export XML first
	home := config.Get().Home
	xmlPath := filepath.Join(home, "Library", "Pioneer", "rekordbox", "rekordbox.xml")

	if _, err := os.Stat(xmlPath); os.IsNotExist(err) {
		sb.WriteString("⚠️ **No XML Export Found**\n\n")
		sb.WriteString("Export your library from Rekordbox first:\n")
		sb.WriteString("File → Export Collection in XML format\n")
		return tools.TextResult(sb.String()), nil
	}

	// Use rclone to sync
	s3Path := "s3:cr8-music-storage/rekordbox/"
	var cmd *exec.Cmd

	if dryRun {
		sb.WriteString("**Mode:** Dry Run (preview only)\n\n")
	}

	switch direction {
	case "up":
		sb.WriteString(fmt.Sprintf("**Direction:** Local → S3\n"))
		sb.WriteString(fmt.Sprintf("**Source:** `%s`\n", xmlPath))
		sb.WriteString(fmt.Sprintf("**Destination:** `%s`\n\n", s3Path))

		args := []string{"copy", xmlPath, s3Path + "xml/", "-v"}
		if dryRun {
			args = append(args, "--dry-run")
		}
		cmd = exec.CommandContext(ctx, "rclone", args...)

	case "down":
		sb.WriteString(fmt.Sprintf("**Direction:** S3 → Local\n"))
		// Download from S3
		args := []string{"copy", s3Path + "xml/", filepath.Dir(xmlPath) + "/", "-v"}
		if dryRun {
			args = append(args, "--dry-run")
		}
		cmd = exec.CommandContext(ctx, "rclone", args...)

	default:
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("direction must be 'up' or 'down'")), nil
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		sb.WriteString(fmt.Sprintf("**Error:** %s\n", err))
		sb.WriteString(fmt.Sprintf("```\n%s\n```\n", string(output)))
	} else {
		if dryRun {
			sb.WriteString("✅ **Dry run complete**\n\n")
		} else {
			sb.WriteString("✅ **Sync complete**\n\n")
		}
		if len(output) > 0 {
			sb.WriteString(fmt.Sprintf("```\n%s\n```\n", string(output)))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleSharePlaylist shares a playlist with another user via S3
func handleSharePlaylist(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	playlistName, errResult := tools.RequireStringParam(req, "playlist")
	if errResult != nil {
		return errResult, nil
	}
	toUser, errResult := tools.RequireStringParam(req, "to_user")
	if errResult != nil {
		return errResult, nil
	}
	includeTracks := tools.GetBoolParam(req, "include_tracks", false)

	var sb strings.Builder
	sb.WriteString("# Share Playlist\n\n")
	sb.WriteString(fmt.Sprintf("**Playlist:** %s\n", playlistName))
	sb.WriteString(fmt.Sprintf("**To User:** %s\n", toUser))
	sb.WriteString(fmt.Sprintf("**Include Tracks:** %v\n\n", includeTracks))

	// Export playlist to XML via Python
	result, err := pythonScript(ctx, "export_playlist", map[string]interface{}{
		"playlist": playlistName,
	})
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to export playlist: %w", err)), nil
	}

	xmlPath, ok := result["xml_path"].(string)
	if !ok || xmlPath == "" {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get export path")), nil
	}

	trackCount := 0
	if tc, ok := result["track_count"].(float64); ok {
		trackCount = int(tc)
	}

	sb.WriteString(fmt.Sprintf("**Tracks:** %d\n", trackCount))
	sb.WriteString(fmt.Sprintf("**Export:** `%s`\n\n", xmlPath))

	// Upload to S3 shared folder
	// Use current user from environment or config
	currentUser := config.Get().AftrsUser
	if currentUser == "" {
		currentUser = "hairglasses" // Default
	}

	s3Path := fmt.Sprintf("s3:cr8-music-storage/rekordbox/shared/%s/to/%s/", currentUser, toUser)
	args := []string{"copy", xmlPath, s3Path, "-v"}
	cmd := exec.CommandContext(ctx, "rclone", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		sb.WriteString(fmt.Sprintf("**Upload Error:** %s\n", err))
		sb.WriteString(fmt.Sprintf("```\n%s\n```\n", string(output)))
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("✅ **Playlist shared successfully**\n\n")
	sb.WriteString(fmt.Sprintf("**S3 Path:** `%s`\n", s3Path))

	if includeTracks {
		sb.WriteString("\n⚠️ Track file sync not yet implemented - metadata only\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleListShared lists playlists shared with the current user
func handleListShared(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fromUser := tools.GetStringParam(req, "from_user")

	var sb strings.Builder
	sb.WriteString("# Shared Playlists\n\n")

	// Get current user
	currentUser := config.Get().AftrsUser
	if currentUser == "" {
		currentUser = "hairglasses"
	}

	// List from S3
	s3Path := fmt.Sprintf("s3:cr8-music-storage/rekordbox/shared/*/to/%s/", currentUser)
	if fromUser != "" {
		s3Path = fmt.Sprintf("s3:cr8-music-storage/rekordbox/shared/%s/to/%s/", fromUser, currentUser)
	}

	cmd := exec.CommandContext(ctx, "rclone", "lsf", s3Path, "--dirs-only")
	output, err := cmd.Output()
	if err != nil {
		// Try listing files instead
		cmd = exec.CommandContext(ctx, "rclone", "lsf", s3Path)
		output, err = cmd.Output()
		if err != nil {
			sb.WriteString("*No shared playlists found*\n\n")
			sb.WriteString("Ask collaborators to share playlists with you using `aftrs_rekordbox_share_playlist`\n")
			return tools.TextResult(sb.String()), nil
		}
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || (len(lines) == 1 && lines[0] == "") {
		sb.WriteString("*No shared playlists found*\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("| Playlist | From | Size |\n")
	sb.WriteString("|----------|------|------|\n")

	for _, line := range lines {
		if line == "" {
			continue
		}
		// Parse the path to extract user and playlist name
		name := strings.TrimSuffix(line, "/")
		name = strings.TrimSuffix(name, ".xml")
		sb.WriteString(fmt.Sprintf("| %s | %s | - |\n", name, fromUser))
	}

	sb.WriteString("\nUse `aftrs_rekordbox_import_shared` to import a playlist.\n")

	return tools.TextResult(sb.String()), nil
}

// handleImportShared imports a shared playlist from another user
func handleImportShared(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	playlistName, errResult := tools.RequireStringParam(req, "playlist")
	if errResult != nil {
		return errResult, nil
	}
	fromUser, errResult := tools.RequireStringParam(req, "from_user")
	if errResult != nil {
		return errResult, nil
	}
	dryRun := tools.GetBoolParam(req, "dry_run", false)

	var sb strings.Builder
	sb.WriteString("# Import Shared Playlist\n\n")
	sb.WriteString(fmt.Sprintf("**Playlist:** %s\n", playlistName))
	sb.WriteString(fmt.Sprintf("**From User:** %s\n", fromUser))
	if dryRun {
		sb.WriteString("**Mode:** Dry Run\n")
	}
	sb.WriteString("\n")

	// Get current user
	currentUser := config.Get().AftrsUser
	if currentUser == "" {
		currentUser = "hairglasses"
	}

	// Download from S3
	s3Path := fmt.Sprintf("s3:cr8-music-storage/rekordbox/shared/%s/to/%s/%s.xml", fromUser, currentUser, playlistName)
	localPath := filepath.Join(config.Get().Home, ".cache", "aftrs", "rekordbox", "imports", fromUser)

	// Create directory
	os.MkdirAll(localPath, 0755)

	localFile := filepath.Join(localPath, playlistName+".xml")

	cmd := exec.CommandContext(ctx, "rclone", "copy", s3Path, localPath, "-v")
	output, err := cmd.CombinedOutput()
	if err != nil {
		sb.WriteString(fmt.Sprintf("**Download Error:** %s\n", err))
		sb.WriteString(fmt.Sprintf("```\n%s\n```\n", string(output)))
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("**Downloaded to:** `%s`\n\n", localFile))

	// Check if file exists
	if _, err := os.Stat(localFile); os.IsNotExist(err) {
		sb.WriteString("❌ **Playlist not found in shared location**\n")
		return tools.TextResult(sb.String()), nil
	}

	if dryRun {
		sb.WriteString("✅ **Dry run complete** - playlist downloaded but not imported\n\n")
		sb.WriteString("Remove `dry_run` to import into Rekordbox.\n")
		return tools.TextResult(sb.String()), nil
	}

	// Import via Python
	result, err := pythonScript(ctx, "import_playlist", map[string]interface{}{
		"xml_path":    localFile,
		"folder_name": fmt.Sprintf("Shared/%s", fromUser),
	})
	if err != nil {
		sb.WriteString(fmt.Sprintf("**Import Error:** %s\n", err))
		return tools.TextResult(sb.String()), nil
	}

	imported := 0
	if i, ok := result["imported"].(float64); ok {
		imported = int(i)
	}
	skipped := 0
	if s, ok := result["skipped"].(float64); ok {
		skipped = int(s)
	}

	sb.WriteString("✅ **Import complete**\n\n")
	sb.WriteString(fmt.Sprintf("**Imported:** %d tracks\n", imported))
	sb.WriteString(fmt.Sprintf("**Skipped:** %d (already in library)\n", skipped))
	sb.WriteString(fmt.Sprintf("**Location:** Playlists → Shared → %s → %s\n", fromUser, playlistName))

	return tools.TextResult(sb.String()), nil
}

// ============================================================================
// Phase 5: Library Maintenance & Automation Handlers
// ============================================================================

// handleMaintenanceStatus returns comprehensive library status
func handleMaintenanceStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getRekordboxClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Rekordbox client: %w", err)), nil
	}

	status, err := clients.GetRekordboxStatusCached(ctx, client)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get status: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Rekordbox Library Status\n\n")

	if status.Connected {
		sb.WriteString("✅ **Connected**\n\n")
	} else {
		sb.WriteString("❌ **Not Connected**\n\n")
	}

	sb.WriteString(fmt.Sprintf("**Database:** `%s`\n", status.DatabasePath))
	sb.WriteString(fmt.Sprintf("**Total Tracks:** %d\n", status.TrackCount))
	sb.WriteString(fmt.Sprintf("**Total Playlists:** %d\n", status.PlaylistCount))
	sb.WriteString(fmt.Sprintf("**Duplicate Playlists:** %d\n", status.DuplicatePlaylists))
	sb.WriteString(fmt.Sprintf("**Analysis Progress:** %s\n", status.AnalysisProgress))

	if status.IsRunning {
		sb.WriteString("\n🎧 **Rekordbox is currently running**\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleUSBInfo returns information about connected USB drives
func handleUSBInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getRekordboxClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Rekordbox client: %w", err)), nil
	}

	info, err := client.GetUSBInfo(ctx)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get USB info: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# USB Drive Information\n\n")

	if !info.Connected {
		sb.WriteString("❌ **No Rekordbox USB drive detected**\n\n")
		sb.WriteString("Insert a USB drive with a PIONEER folder to see contents.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("✅ **USB Drive Connected**\n\n")
	sb.WriteString(fmt.Sprintf("**Drive:** %s\n", info.DriveLetter))
	sb.WriteString(fmt.Sprintf("**Audio Files:** %d\n", info.AudioFiles))
	sb.WriteString(fmt.Sprintf("**Analysis Files:** %d\n", info.AnalysisFiles))
	sb.WriteString(fmt.Sprintf("**Total Size:** %.2f GB\n", info.TotalSizeGB))

	return tools.TextResult(sb.String()), nil
}

// handleImportUSB imports tracks from USB drive
func handleImportUSB(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getRekordboxClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Rekordbox client: %w", err)), nil
	}

	copyAnalysis := tools.GetBoolParam(req, "copy_analysis", true)

	result, err := client.ImportUSB(ctx, copyAnalysis)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to import USB: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# USB Import Results\n\n")

	if result.Success {
		sb.WriteString("✅ **Import Successful**\n\n")
	} else {
		sb.WriteString("⚠️ **Import Completed with Issues**\n\n")
	}

	sb.WriteString(fmt.Sprintf("**Files Copied:** %d\n", result.Count))
	sb.WriteString(fmt.Sprintf("\n**Details:**\n```\n%s\n```\n", result.Details))

	return tools.TextResult(sb.String()), nil
}

// handleCleanupDuplicates removes duplicate playlists
func handleCleanupDuplicates(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getRekordboxClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Rekordbox client: %w", err)), nil
	}

	useUI := tools.GetBoolParam(req, "use_ui", false)

	result, err := client.CleanupDuplicates(ctx, useUI)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to cleanup duplicates: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Duplicate Playlist Cleanup\n\n")

	if result.Success {
		sb.WriteString("✅ **Cleanup Successful**\n\n")
	} else {
		sb.WriteString("⚠️ **Cleanup Completed with Issues**\n\n")
	}

	sb.WriteString(fmt.Sprintf("**Playlists Deleted:** %d\n", result.Count))
	sb.WriteString(fmt.Sprintf("\n**Details:**\n```\n%s\n```\n", result.Details))

	return tools.TextResult(sb.String()), nil
}

// handleTriggerAnalysis triggers track analysis in Rekordbox
func handleTriggerAnalysis(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getRekordboxClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Rekordbox client: %w", err)), nil
	}

	result, err := client.TriggerAnalysis(ctx)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to trigger analysis: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Track Analysis\n\n")

	if result.Success {
		sb.WriteString("✅ **Analysis Triggered**\n\n")
		sb.WriteString("Rekordbox is now analyzing tracks. This may take several hours for large libraries.\n\n")
		sb.WriteString("**Monitor progress with:** `aftrs_rekordbox_maintenance_status`\n")
	} else {
		sb.WriteString("⚠️ **Analysis Trigger Failed**\n\n")
		sb.WriteString("Make sure Rekordbox is running and visible.\n")
	}

	sb.WriteString(fmt.Sprintf("\n**Details:**\n```\n%s\n```\n", result.Details))

	return tools.TextResult(sb.String()), nil
}

// handleHealth returns health status for Rekordbox integration
func handleHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getRekordboxClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Rekordbox client: %w", err)), nil
	}

	health := client.GetHealth()

	var sb strings.Builder
	sb.WriteString("# Rekordbox Integration Health\n\n")

	// Status indicator
	switch health.Status {
	case "healthy":
		sb.WriteString(fmt.Sprintf("✅ **%s** (Score: %d/100)\n\n", strings.ToUpper(health.Status), health.Score))
	case "degraded":
		sb.WriteString(fmt.Sprintf("⚠️ **%s** (Score: %d/100)\n\n", strings.ToUpper(health.Status), health.Score))
	default:
		sb.WriteString(fmt.Sprintf("❌ **%s** (Score: %d/100)\n\n", strings.ToUpper(health.Status), health.Score))
	}

	// Connection status
	if health.Connected {
		sb.WriteString("**Connection:** ✅ Connected\n")
	} else {
		sb.WriteString("**Connection:** ❌ Not Connected\n")
	}

	if health.DatabaseExists {
		sb.WriteString("**Database:** ✅ Found\n")
	} else {
		sb.WriteString("**Database:** ❌ Not Found\n")
	}

	// Issues
	if len(health.Issues) > 0 {
		sb.WriteString("\n## Issues\n\n")
		for _, issue := range health.Issues {
			sb.WriteString(fmt.Sprintf("- %s\n", issue))
		}
	}

	// Recommendations
	if len(health.Recommendations) > 0 {
		sb.WriteString("\n## Recommendations\n\n")
		for _, rec := range health.Recommendations {
			sb.WriteString(fmt.Sprintf("- %s\n", rec))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// ============================================================================
// Phase 6: Playback & History Handlers (for Showkontrol integration)
// ============================================================================

// handleNowPlaying returns the currently playing or most recently played track
func handleNowPlaying(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getRekordboxClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Rekordbox client: %w", err)), nil
	}

	track, err := clients.GetRekordboxNowPlayingCached(ctx, client)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get now playing: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Now Playing\n\n")

	sb.WriteString(fmt.Sprintf("**Title:** %s\n", track.Title))
	sb.WriteString(fmt.Sprintf("**Artist:** %s\n", track.Artist))

	if track.Album != "" {
		sb.WriteString(fmt.Sprintf("**Album:** %s\n", track.Album))
	}
	if track.BPM > 0 {
		sb.WriteString(fmt.Sprintf("**BPM:** %.1f\n", track.BPM))
	}
	if track.Key != "" {
		sb.WriteString(fmt.Sprintf("**Key:** %s\n", track.Key))
	}
	if track.DurationMS > 0 {
		durationSec := track.DurationMS / 1000
		minutes := durationSec / 60
		seconds := durationSec % 60
		sb.WriteString(fmt.Sprintf("**Duration:** %d:%02d\n", minutes, seconds))
	}
	if track.Path != "" {
		sb.WriteString(fmt.Sprintf("**Path:** `%s`\n", track.Path))
	}
	if track.Timestamp != "" {
		sb.WriteString(fmt.Sprintf("\n*Updated: %s*\n", track.Timestamp))
	}

	return tools.TextResult(sb.String()), nil
}

// handleHistory returns play history from Rekordbox
func handleHistory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getRekordboxClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Rekordbox client: %w", err)), nil
	}

	limit := tools.GetIntParam(req, "limit", 10)

	history, err := client.GetHistory(ctx, limit)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get history: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Play History\n\n")

	if len(history) == 0 {
		sb.WriteString("*No play history found*\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("| # | Title | Artist | Played At |\n")
	sb.WriteString("|---|-------|--------|----------|\n")

	for i, entry := range history {
		playedAt := entry.PlayedAt
		if playedAt == "" {
			playedAt = "-"
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n",
			i+1, entry.Title, entry.Artist, playedAt))
	}

	sb.WriteString(fmt.Sprintf("\n*Showing %d tracks*\n", len(history)))

	return tools.TextResult(sb.String()), nil
}

// handleSessions lists DJ sessions from Rekordbox
func handleSessions(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getRekordboxClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Rekordbox client: %w", err)), nil
	}

	limit := tools.GetIntParam(req, "limit", 10)

	sessions, err := client.GetSessions(ctx, limit)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get sessions: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# DJ Sessions\n\n")

	if len(sessions) == 0 {
		sb.WriteString("*No sessions found*\n\n")
		sb.WriteString("Sessions are created when you play tracks in Rekordbox.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("| ID | Date | Tracks | Duration |\n")
	sb.WriteString("|----|------|--------|----------|\n")

	for _, session := range sessions {
		duration := session.Duration
		if duration == "" {
			duration = "-"
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %d | %s |\n",
			session.ID, session.Date, session.Tracks, duration))
	}

	sb.WriteString(fmt.Sprintf("\n*Showing %d sessions*\n\n", len(sessions)))
	sb.WriteString("Use `aftrs_rekordbox_session_tracks` with a session ID to see tracks.\n")

	return tools.TextResult(sb.String()), nil
}

// handleSessionTracks returns tracks from a specific DJ session
func handleSessionTracks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getRekordboxClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Rekordbox client: %w", err)), nil
	}

	sessionID, errResult := tools.RequireIntParam(req, "session_id")
	if errResult != nil {
		return errResult, nil
	}

	tracks, err := client.GetSessionTracks(ctx, sessionID)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrAPIError, fmt.Errorf("failed to get session tracks: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Session %d Tracks\n\n", sessionID))

	if len(tracks) == 0 {
		sb.WriteString("*No tracks found for this session*\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("| # | Title | Artist | Played At |\n")
	sb.WriteString("|---|-------|--------|----------|\n")

	for i, track := range tracks {
		playedAt := track.PlayedAt
		if playedAt == "" {
			playedAt = "-"
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n",
			i+1, track.Title, track.Artist, playedAt))
	}

	sb.WriteString(fmt.Sprintf("\n*Total: %d tracks*\n", len(tracks)))

	return tools.TextResult(sb.String()), nil
}

// init registers this module
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
