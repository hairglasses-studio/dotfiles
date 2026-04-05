// Package vj_clips provides VJ clip management MCP tools.
package vj_clips

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewVJClipsClient)

// Module implements the ToolModule interface for VJ clip management
type Module struct{}

func (m *Module) Name() string {
	return "vj_clips"
}

func (m *Module) Description() string {
	return "VJ video clip management: scan, sync, upload, and download clips via S3 with user/playlist organization"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		// Discovery Tools
		{
			Tool: mcp.NewTool("aftrs_vj_clips_list",
				mcp.WithDescription("List local VJ clips organized by playlist."),
				mcp.WithString("playlist", mcp.Description("Filter by playlist name")),
			),
			Handler:             handleList,
			Category:            "vj_clips",
			Subcategory:         "discovery",
			Tags:                []string{"vj", "clips", "list", "resolume", "media"},
			UseCases:            []string{"Browse local clips", "Inventory media"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "vj_clips",
		},
		{
			Tool: mcp.NewTool("aftrs_vj_clips_playlists",
				mcp.WithDescription("List all local playlists with clip counts and sizes."),
			),
			Handler:             handlePlaylists,
			Category:            "vj_clips",
			Subcategory:         "discovery",
			Tags:                []string{"vj", "clips", "playlists", "packs"},
			UseCases:            []string{"View playlist overview", "Check media organization"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "vj_clips",
		},
		{
			Tool: mcp.NewTool("aftrs_vj_clips_s3_list",
				mcp.WithDescription("List VJ clips in S3 bucket."),
				mcp.WithString("user", mcp.Description("Filter by user (default: current user)")),
				mcp.WithString("playlist", mcp.Description("Filter by playlist name")),
			),
			Handler:             handleS3List,
			Category:            "vj_clips",
			Subcategory:         "discovery",
			Tags:                []string{"vj", "clips", "s3", "cloud", "list"},
			UseCases:            []string{"Browse cloud clips", "See available downloads"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "vj_clips",
		},
		{
			Tool: mcp.NewTool("aftrs_vj_clips_s3_playlists",
				mcp.WithDescription("List playlists in S3 bucket."),
				mcp.WithString("user", mcp.Description("Filter by user (default: all users)")),
			),
			Handler:             handleS3Playlists,
			Category:            "vj_clips",
			Subcategory:         "discovery",
			Tags:                []string{"vj", "clips", "s3", "playlists"},
			UseCases:            []string{"View cloud playlists", "Browse shared content"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "vj_clips",
		},

		// Upload Tools
		{
			Tool: mcp.NewTool("aftrs_vj_clips_upload",
				mcp.WithDescription("Upload a clip to S3."),
				mcp.WithString("path", mcp.Required(), mcp.Description("Local path to the clip file")),
				mcp.WithString("playlist", mcp.Description("Playlist/pack name (default: from path or 'default')")),
				mcp.WithString("user", mcp.Description("User/owner (default: current user)")),
				mcp.WithString("tags", mcp.Description("Comma-separated tags")),
			),
			Handler:             handleUpload,
			Category:            "vj_clips",
			Subcategory:         "sync",
			Tags:                []string{"vj", "clips", "upload", "s3"},
			UseCases:            []string{"Backup clip to cloud", "Share clip"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "vj_clips",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_vj_clips_upload_playlist",
				mcp.WithDescription("Upload all clips in a playlist to S3."),
				mcp.WithString("playlist", mcp.Required(), mcp.Description("Playlist name to upload")),
				mcp.WithString("user", mcp.Description("User/owner (default: current user)")),
			),
			Handler:             handleUploadPlaylist,
			Category:            "vj_clips",
			Subcategory:         "sync",
			Tags:                []string{"vj", "clips", "upload", "playlist", "batch"},
			UseCases:            []string{"Backup entire playlist", "Share pack"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "vj_clips",
			IsWrite:             true,
		},

		// Download Tools
		{
			Tool: mcp.NewTool("aftrs_vj_clips_download",
				mcp.WithDescription("Download a clip from S3."),
				mcp.WithString("key", mcp.Required(), mcp.Description("S3 key of the clip")),
				mcp.WithString("playlist", mcp.Description("Local playlist to save to (default: from S3 path)")),
			),
			Handler:             handleDownload,
			Category:            "vj_clips",
			Subcategory:         "sync",
			Tags:                []string{"vj", "clips", "download", "s3"},
			UseCases:            []string{"Download cloud clip", "Restore from backup"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "vj_clips",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_vj_clips_download_playlist",
				mcp.WithDescription("Download all clips in an S3 playlist."),
				mcp.WithString("user", mcp.Required(), mcp.Description("User who owns the playlist")),
				mcp.WithString("playlist", mcp.Required(), mcp.Description("Playlist name to download")),
			),
			Handler:             handleDownloadPlaylist,
			Category:            "vj_clips",
			Subcategory:         "sync",
			Tags:                []string{"vj", "clips", "download", "playlist", "batch"},
			UseCases:            []string{"Download entire playlist", "Get shared pack"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "vj_clips",
			IsWrite:             true,
		},

		// Sync Tools
		{
			Tool: mcp.NewTool("aftrs_vj_clips_sync",
				mcp.WithDescription("Compare local and S3 clips to identify sync opportunities."),
				mcp.WithString("user", mcp.Description("User to compare against (default: current user)")),
			),
			Handler:             handleSync,
			Category:            "vj_clips",
			Subcategory:         "sync",
			Tags:                []string{"vj", "clips", "sync", "compare"},
			UseCases:            []string{"Check sync status", "Find missing clips"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "vj_clips",
		},

		// Management Tools
		{
			Tool: mcp.NewTool("aftrs_vj_clips_delete",
				mcp.WithDescription("Delete a clip from S3."),
				mcp.WithString("key", mcp.Required(), mcp.Description("S3 key of the clip to delete")),
			),
			Handler:             handleDelete,
			Category:            "vj_clips",
			Subcategory:         "management",
			Tags:                []string{"vj", "clips", "delete", "s3"},
			UseCases:            []string{"Remove cloud clip", "Clean up storage"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "vj_clips",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_vj_clips_health",
				mcp.WithDescription("Check VJ clips system health."),
			),
			Handler:             handleHealth,
			Category:            "vj_clips",
			Subcategory:         "health",
			Tags:                []string{"vj", "clips", "health", "status"},
			UseCases:            []string{"Diagnose issues", "System check"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "vj_clips",
		},
	}
}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Handlers

func handleList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	clips, err := client.ScanLocalClips(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	playlistFilter := tools.GetStringParam(req, "playlist")

	var filtered []clients.VJClipInfo
	for _, c := range clips {
		if playlistFilter != "" && c.Playlist != playlistFilter {
			continue
		}
		filtered = append(filtered, c)
	}

	if len(filtered) == 0 {
		return tools.TextResult("No clips found."), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Local VJ Clips (%d)\n\n", len(filtered)))
	sb.WriteString(fmt.Sprintf("**Media path:** `%s`\n\n", client.GetMediaPath()))

	// Group by playlist
	byPlaylist := make(map[string][]clients.VJClipInfo)
	for _, c := range filtered {
		byPlaylist[c.Playlist] = append(byPlaylist[c.Playlist], c)
	}

	for playlist, clips := range byPlaylist {
		var totalSize int64
		for _, c := range clips {
			totalSize += c.Size
		}

		sb.WriteString(fmt.Sprintf("## %s (%d clips, %s)\n\n", playlist, len(clips), formatSize(totalSize)))
		sb.WriteString("| Name | Format | Size | DXV |\n")
		sb.WriteString("|------|--------|------|-----|\n")

		for _, c := range clips {
			dxv := ""
			if c.IsDXV {
				dxv = "✅"
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
				c.Name, c.Format, formatSize(c.Size), dxv))
		}
		sb.WriteString("\n")
	}

	return tools.TextResult(sb.String()), nil
}

func handlePlaylists(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	playlists, err := client.ListPlaylists(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(playlists) == 0 {
		return tools.TextResult("No playlists found."), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Local VJ Playlists (%d)\n\n", len(playlists)))

	var totalClips int
	var totalSize int64
	for _, p := range playlists {
		totalClips += p.ClipCount
		totalSize += p.TotalSize
	}

	sb.WriteString(fmt.Sprintf("**Total:** %d clips, %s\n\n", totalClips, formatSize(totalSize)))

	sb.WriteString("| Playlist | Clips | Size | Modified |\n")
	sb.WriteString("|----------|-------|------|----------|\n")

	for _, p := range playlists {
		sb.WriteString(fmt.Sprintf("| %s | %d | %s | %s |\n",
			p.Name, p.ClipCount, formatSize(p.TotalSize), p.ModifiedAt.Format("2006-01-02")))
	}

	return tools.TextResult(sb.String()), nil
}

func handleS3List(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	user := tools.GetStringParam(req, "user")
	playlist := tools.GetStringParam(req, "playlist")

	clips, err := client.ListS3Clips(ctx, user, playlist)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(clips) == 0 {
		return tools.TextResult("No clips found in S3."), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# S3 VJ Clips (%d)\n\n", len(clips)))

	// Group by user/playlist
	byKey := make(map[string][]clients.VJClipInfo)
	for _, c := range clips {
		key := c.User + "/" + c.Playlist
		byKey[key] = append(byKey[key], c)
	}

	for key, clips := range byKey {
		var totalSize int64
		for _, c := range clips {
			totalSize += c.Size
		}

		sb.WriteString(fmt.Sprintf("## %s (%d clips, %s)\n\n", key, len(clips), formatSize(totalSize)))
		sb.WriteString("| Name | Format | Size | S3 Key |\n")
		sb.WriteString("|------|--------|------|--------|\n")

		for _, c := range clips {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | `%s` |\n",
				c.Name, c.Format, formatSize(c.Size), c.S3Key))
		}
		sb.WriteString("\n")
	}

	return tools.TextResult(sb.String()), nil
}

func handleS3Playlists(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	user := tools.GetStringParam(req, "user")

	playlists, err := client.ListS3Playlists(ctx, user)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(playlists) == 0 {
		return tools.TextResult("No playlists found in S3."), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# S3 VJ Playlists (%d)\n\n", len(playlists)))

	sb.WriteString("| User | Playlist | Clips | Size |\n")
	sb.WriteString("|------|----------|-------|------|\n")

	for _, p := range playlists {
		sb.WriteString(fmt.Sprintf("| %s | %s | %d | %s |\n",
			p.User, p.Name, p.ClipCount, formatSize(p.TotalSize)))
	}

	return tools.TextResult(sb.String()), nil
}

func handleUpload(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	path, errResult := tools.RequireStringParam(req, "path")
	if errResult != nil {
		return errResult, nil
	}

	var tags []string
	if tagsStr := tools.GetStringParam(req, "tags"); tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	opts := clients.ClipUploadOptions{
		User:     tools.GetStringParam(req, "user"),
		Playlist: tools.GetStringParam(req, "playlist"),
		Tags:     tags,
	}

	result, err := client.UploadClip(ctx, path, opts)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Clip Uploaded\n\n")
	sb.WriteString(fmt.Sprintf("**Name:** %s\n", result.Name))
	sb.WriteString(fmt.Sprintf("**User:** %s\n", result.User))
	sb.WriteString(fmt.Sprintf("**Playlist:** %s\n", result.Playlist))
	sb.WriteString(fmt.Sprintf("**S3 Key:** `%s`\n", result.S3Key))
	sb.WriteString(fmt.Sprintf("**Size:** %s\n", formatSize(result.Size)))

	return tools.TextResult(sb.String()), nil
}

func handleUploadPlaylist(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	playlist, errResult := tools.RequireStringParam(req, "playlist")
	if errResult != nil {
		return errResult, nil
	}

	user := tools.GetStringParam(req, "user")

	result, err := client.UploadPlaylist(ctx, playlist, user)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(result.Uploaded) == 0 && len(result.Errors) == 0 {
		return tools.TextResult(fmt.Sprintf("No clips found in playlist: %s", playlist)), nil
	}

	var totalSize int64
	for _, c := range result.Uploaded {
		totalSize += c.Size
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Playlist Uploaded: %s\n\n", playlist))
	sb.WriteString(fmt.Sprintf("**Clips uploaded:** %d\n", len(result.Uploaded)))
	sb.WriteString(fmt.Sprintf("**Total size:** %s\n\n", formatSize(totalSize)))

	for _, c := range result.Uploaded {
		sb.WriteString(fmt.Sprintf("- %s → `%s`\n", c.Name, c.S3Key))
	}

	if len(result.Errors) > 0 {
		sb.WriteString(fmt.Sprintf("\n## Errors (%d)\n\n", len(result.Errors)))
		for _, e := range result.Errors {
			sb.WriteString(fmt.Sprintf("- ⚠️ %s\n", e))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleDownload(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	key, errResult := tools.RequireStringParam(req, "key")
	if errResult != nil {
		return errResult, nil
	}

	playlist := tools.GetStringParam(req, "playlist")

	destPath, err := client.DownloadClip(ctx, key, playlist)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("**Downloaded:** `%s`", destPath)), nil
}

func handleDownloadPlaylist(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	user, errResult := tools.RequireStringParam(req, "user")
	if errResult != nil {
		return errResult, nil
	}

	playlist, errResult := tools.RequireStringParam(req, "playlist")
	if errResult != nil {
		return errResult, nil
	}

	result, err := client.DownloadPlaylist(ctx, user, playlist)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(result.Downloaded) == 0 && len(result.Errors) == 0 {
		return tools.TextResult(fmt.Sprintf("No clips found in %s/%s", user, playlist)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Playlist Downloaded: %s/%s\n\n", user, playlist))
	sb.WriteString(fmt.Sprintf("**Clips downloaded:** %d\n\n", len(result.Downloaded)))

	for _, path := range result.Downloaded {
		sb.WriteString(fmt.Sprintf("- `%s`\n", path))
	}

	if len(result.Errors) > 0 {
		sb.WriteString(fmt.Sprintf("\n## Errors (%d)\n\n", len(result.Errors)))
		for _, e := range result.Errors {
			sb.WriteString(fmt.Sprintf("- ⚠️ %s\n", e))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleSync(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	user := tools.GetStringParam(req, "user")

	result, err := client.SyncClips(ctx, user)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# VJ Clips Sync Status\n\n")

	if len(result.ToUpload) > 0 {
		var totalSize int64
		for _, c := range result.ToUpload {
			totalSize += c.Size
		}
		sb.WriteString(fmt.Sprintf("## Local Only (%d, %s) → Upload to S3\n\n", len(result.ToUpload), formatSize(totalSize)))
		for _, c := range result.ToUpload {
			sb.WriteString(fmt.Sprintf("- %s/%s (%s)\n", c.Playlist, c.Name, formatSize(c.Size)))
		}
		sb.WriteString("\n")
	}

	if len(result.ToDownload) > 0 {
		var totalSize int64
		for _, c := range result.ToDownload {
			totalSize += c.Size
		}
		sb.WriteString(fmt.Sprintf("## S3 Only (%d, %s) → Download to Local\n\n", len(result.ToDownload), formatSize(totalSize)))
		for _, c := range result.ToDownload {
			sb.WriteString(fmt.Sprintf("- %s/%s (%s)\n", c.Playlist, c.Name, formatSize(c.Size)))
		}
		sb.WriteString("\n")
	}

	if len(result.ToUpload) == 0 && len(result.ToDownload) == 0 {
		sb.WriteString("**All clips are in sync!**\n")
	}

	return tools.TextResult(sb.String()), nil
}

func handleDelete(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	key, errResult := tools.RequireStringParam(req, "key")
	if errResult != nil {
		return errResult, nil
	}

	if err := client.DeleteClip(ctx, key); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("**Deleted:** `%s`", key)), nil
}

func handleHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# VJ Clips Health\n\n")

	statusEmoji := "✅"
	if health.Status == "degraded" {
		statusEmoji = "⚠️"
	} else if health.Status == "critical" {
		statusEmoji = "❌"
	}

	sb.WriteString(fmt.Sprintf("**Health Score:** %d/100 %s\n", health.Score, statusEmoji))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n\n", health.Status))

	sb.WriteString("## Metrics\n\n")
	sb.WriteString(fmt.Sprintf("- **Local clips:** %d (%s)\n", health.LocalCount, formatSize(health.LocalSize)))
	sb.WriteString(fmt.Sprintf("- **S3 clips:** %d (%s)\n", health.S3Count, formatSize(health.S3Size)))

	if len(health.Issues) > 0 {
		sb.WriteString("\n## Issues\n\n")
		for _, issue := range health.Issues {
			sb.WriteString(fmt.Sprintf("- ❌ %s\n", issue))
		}
	}

	if len(health.Recommendations) > 0 {
		sb.WriteString("\n## Recommendations\n\n")
		for _, rec := range health.Recommendations {
			sb.WriteString(fmt.Sprintf("- %s\n", rec))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
