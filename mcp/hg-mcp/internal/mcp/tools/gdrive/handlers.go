package gdrive

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
	"github.com/hairglasses-studio/mcpkit/sanitize"
)

// VJ Pack definitions for rclone sync
var vjPacks = map[string]string{
	"hackerglasses": "HACKERGLASSES",
	"hairglasses":   "Hairglasses Visuals",
	"fetz":          "Fetz VJ Storage",
	"masks":         "Masks",
	"algorave":      "Algorave",
	"relic":         "Relic VJ Clips",
	"mantissa":      "Mantissa",
	"tricky":        "TrickyFM Visuals",
	"pixabay":       "Pixabay",
	"resolume":      "Resolume",
	"recursive":     "Recursive",
	"milligram":     "Milligram",
	"church":        "Hairglasses at Church",
	"footage":       "Hairglasses Footage",
}

const vjBasePath = "gdrive:Video/VJ Tools (Laptop Salvage)/DXV3 1080p HQ Clips"

// handleList lists files in a folder
func handleList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetGDriveClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Google Drive not configured: %w", err)), nil
	}

	folderID := tools.OptionalStringParam(req, "folder_id", "root")
	limit := tools.GetIntParam(req, "limit", 100)
	pageToken := tools.GetStringParam(req, "page_token")

	files, nextToken, err := client.ListFiles(ctx, folderID, int64(limit), pageToken)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to list files: %w", err)), nil
	}

	// Format output
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Files in folder (showing %d)\n\n", len(files)))

	// Separate folders and files
	var folders, regularFiles []clients.GDriveFile
	for _, f := range files {
		if f.IsFolder {
			folders = append(folders, f)
		} else {
			regularFiles = append(regularFiles, f)
		}
	}

	// List folders first
	if len(folders) > 0 {
		sb.WriteString("### Folders\n")
		for _, f := range folders {
			sb.WriteString(fmt.Sprintf("- **%s/** `%s`\n", f.Name, f.ID))
		}
		sb.WriteString("\n")
	}

	// Then files
	if len(regularFiles) > 0 {
		sb.WriteString("### Files\n")
		for _, f := range regularFiles {
			size := clients.FormatFileSize(f.Size)
			sb.WriteString(fmt.Sprintf("- %s (%s) `%s`\n", f.Name, size, f.ID))
		}
	}

	if nextToken != "" {
		sb.WriteString(fmt.Sprintf("\n---\n*More files available. Use page_token: `%s`*", nextToken))
	}

	return tools.TextResult(sb.String()), nil
}

// handleSearch searches for files
func handleSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetGDriveClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Google Drive not configured: %w", err)), nil
	}

	query := tools.GetStringParam(req, "query")
	fileType := tools.GetStringParam(req, "file_type")
	folderID := tools.GetStringParam(req, "folder_id")
	limit := tools.GetIntParam(req, "limit", 50)

	files, err := client.SearchFiles(ctx, query, fileType, folderID, int64(limit))
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("search failed: %w", err)), nil
	}

	var sb strings.Builder
	searchDesc := "all files"
	if query != "" {
		searchDesc = fmt.Sprintf("'%s'", query)
	}
	if fileType != "" {
		searchDesc += fmt.Sprintf(" (%s only)", fileType)
	}

	sb.WriteString(fmt.Sprintf("## Search Results for %s\n\n", searchDesc))
	sb.WriteString(fmt.Sprintf("Found **%d** files\n\n", len(files)))

	for _, f := range files {
		icon := "📄"
		if f.IsFolder {
			icon = "📁"
		} else if strings.Contains(f.MimeType, "video") {
			icon = "🎬"
		} else if strings.Contains(f.MimeType, "image") {
			icon = "🖼️"
		} else if strings.Contains(f.MimeType, "audio") {
			icon = "🎵"
		}

		size := ""
		if !f.IsFolder && f.Size > 0 {
			size = fmt.Sprintf(" (%s)", clients.FormatFileSize(f.Size))
		}

		sb.WriteString(fmt.Sprintf("%s **%s**%s\n", icon, f.Name, size))
		sb.WriteString(fmt.Sprintf("   ID: `%s`\n", f.ID))
		if !f.ModifiedTime.IsZero() {
			sb.WriteString(fmt.Sprintf("   Modified: %s\n", f.ModifiedTime.Format("2006-01-02 15:04")))
		}
		sb.WriteString("\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleInfo gets file/folder info
func handleInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetGDriveClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Google Drive not configured: %w", err)), nil
	}

	fileID, errResult := tools.RequireStringParam(req, "file_id")
	if errResult != nil {
		return errResult, nil
	}
	includePath := tools.GetBoolParam(req, "include_path", false)

	file, err := client.GetFileInfo(ctx, fileID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get file info: %w", err)), nil
	}

	var sb strings.Builder
	icon := "📄"
	if file.IsFolder {
		icon = "📁"
	} else if strings.Contains(file.MimeType, "video") {
		icon = "🎬"
	} else if strings.Contains(file.MimeType, "image") {
		icon = "🖼️"
	}

	sb.WriteString(fmt.Sprintf("## %s %s\n\n", icon, file.Name))
	sb.WriteString(fmt.Sprintf("- **ID:** `%s`\n", file.ID))
	sb.WriteString(fmt.Sprintf("- **Type:** %s\n", file.MimeType))

	if !file.IsFolder {
		sb.WriteString(fmt.Sprintf("- **Size:** %s\n", clients.FormatFileSize(file.Size)))
	}

	if !file.CreatedTime.IsZero() {
		sb.WriteString(fmt.Sprintf("- **Created:** %s\n", file.CreatedTime.Format("2006-01-02 15:04:05")))
	}
	if !file.ModifiedTime.IsZero() {
		sb.WriteString(fmt.Sprintf("- **Modified:** %s\n", file.ModifiedTime.Format("2006-01-02 15:04:05")))
	}

	if file.WebViewLink != "" {
		sb.WriteString(fmt.Sprintf("- **Web Link:** %s\n", file.WebViewLink))
	}

	if includePath && len(file.Parents) > 0 {
		path, err := client.GetFolderPath(ctx, file.Parents[0])
		if err == nil {
			sb.WriteString(fmt.Sprintf("- **Path:** %s/%s\n", path, file.Name))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleSharedDrives lists shared drives
func handleSharedDrives(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetGDriveClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Google Drive not configured: %w", err)), nil
	}

	drives, err := client.ListSharedDrives(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to list shared drives: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("## Shared Drives\n\n")

	if len(drives) == 0 {
		sb.WriteString("No shared drives found.\n")
	} else {
		for _, d := range drives {
			sb.WriteString(fmt.Sprintf("- **%s** `%s`\n", d.Name, d.ID))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleDownload downloads a single file
func handleDownload(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetGDriveClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Google Drive not configured: %w", err)), nil
	}

	fileID, errResult := tools.RequireStringParam(req, "file_id")
	if errResult != nil {
		return errResult, nil
	}
	destination := tools.OptionalStringParam(req, "destination", ".")

	// Expand ~ in path
	if strings.HasPrefix(destination, "~") {
		home, _ := os.UserHomeDir()
		destination = filepath.Join(home, destination[1:])
	}

	if err := sanitize.RclonePath(destination); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid destination path: %w", err)), nil
	}

	progress, err := client.DownloadFile(ctx, fileID, destination)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("download failed: %w", err)), nil
	}

	return tools.TextResult(fmt.Sprintf("✅ Downloaded **%s** (%s)\nSaved to: %s",
		progress.FileName,
		clients.FormatFileSize(progress.DownloadedBytes),
		destination,
	)), nil
}

// handleDownloadFolder downloads all files from a folder
func handleDownloadFolder(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetGDriveClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Google Drive not configured: %w", err)), nil
	}

	folderID, errResult := tools.RequireStringParam(req, "folder_id")
	if errResult != nil {
		return errResult, nil
	}
	destination, errResult := tools.RequireStringParam(req, "destination")
	if errResult != nil {
		return errResult, nil
	}
	fileType := tools.GetStringParam(req, "file_type")
	recursive := tools.GetBoolParam(req, "recursive", false)

	// Expand ~ in path
	if strings.HasPrefix(destination, "~") {
		home, _ := os.UserHomeDir()
		destination = filepath.Join(home, destination[1:])
	}

	if err := sanitize.RclonePath(destination); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid destination path: %w", err)), nil
	}

	results, err := client.DownloadFolder(ctx, folderID, destination, fileType, recursive)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("download failed: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("## Download Complete\n\n")

	successCount := 0
	errorCount := 0
	var totalBytes int64

	for _, r := range results {
		if r.Status == "completed" {
			successCount++
			totalBytes += r.DownloadedBytes
			sb.WriteString(fmt.Sprintf("✅ %s (%s)\n", r.FileName, clients.FormatFileSize(r.DownloadedBytes)))
		} else {
			errorCount++
			sb.WriteString(fmt.Sprintf("❌ %s - %s\n", r.FileName, r.Status))
		}
	}

	sb.WriteString(fmt.Sprintf("\n---\n**Summary:** %d files downloaded (%s)", successCount, clients.FormatFileSize(totalBytes)))
	if errorCount > 0 {
		sb.WriteString(fmt.Sprintf(", %d errors", errorCount))
	}
	sb.WriteString(fmt.Sprintf("\n**Location:** %s", destination))

	return tools.TextResult(sb.String()), nil
}

// handleDownloadVideos is optimized for downloading video files only
func handleDownloadVideos(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetGDriveClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Google Drive not configured: %w", err)), nil
	}

	folderID, errResult := tools.RequireStringParam(req, "folder_id")
	if errResult != nil {
		return errResult, nil
	}
	destination, errResult := tools.RequireStringParam(req, "destination")
	if errResult != nil {
		return errResult, nil
	}
	recursive := tools.GetBoolParam(req, "recursive", false)

	// Expand ~ in path
	if strings.HasPrefix(destination, "~") {
		home, _ := os.UserHomeDir()
		destination = filepath.Join(home, destination[1:])
	}

	if err := sanitize.RclonePath(destination); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid destination path: %w", err)), nil
	}

	// Get folder info for context
	folderInfo, err := client.GetFileInfo(ctx, folderID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get folder info: %w", err)), nil
	}

	results, err := client.DownloadFolder(ctx, folderID, destination, "video", recursive)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("download failed: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## 🎬 VJ Clips Downloaded from '%s'\n\n", folderInfo.Name))

	successCount := 0
	errorCount := 0
	var totalBytes int64

	for _, r := range results {
		if r.Status == "completed" {
			successCount++
			totalBytes += r.DownloadedBytes
			sb.WriteString(fmt.Sprintf("✅ %s (%s)\n", r.FileName, clients.FormatFileSize(r.DownloadedBytes)))
		} else {
			errorCount++
			sb.WriteString(fmt.Sprintf("❌ %s - %s\n", r.FileName, r.Status))
		}
	}

	sb.WriteString(fmt.Sprintf("\n---\n🎬 **%d video clips** ready (%s total)\n", successCount, clients.FormatFileSize(totalBytes)))
	if errorCount > 0 {
		sb.WriteString(fmt.Sprintf("⚠️ %d files failed\n", errorCount))
	}
	sb.WriteString(fmt.Sprintf("📁 **Location:** %s", destination))

	return tools.TextResult(sb.String()), nil
}

// handleQuota checks storage quota
func handleQuota(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetGDriveClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Google Drive not configured: %w", err)), nil
	}

	quota, err := client.GetQuota(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get quota: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("## Google Drive Storage\n\n")

	if quota.Limit > 0 {
		sb.WriteString(fmt.Sprintf("- **Used:** %s / %s (%.1f%%)\n",
			clients.FormatFileSize(quota.Usage),
			clients.FormatFileSize(quota.Limit),
			quota.UsedPercent))
		sb.WriteString(fmt.Sprintf("- **Available:** %.2f GB\n", quota.AvailableGB))
	} else {
		sb.WriteString(fmt.Sprintf("- **Used:** %s (unlimited storage)\n", clients.FormatFileSize(quota.Usage)))
	}

	sb.WriteString(fmt.Sprintf("- **In Drive:** %s\n", clients.FormatFileSize(quota.UsageInDrive)))

	return tools.TextResult(sb.String()), nil
}

// handleVJSync syncs VJ clips to local Resolume folder
func handleVJSync(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := clients.GetGDriveClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("Google Drive not configured: %w", err)), nil
	}

	sourceFolderID, errResult := tools.RequireStringParam(req, "source_folder_id")
	if errResult != nil {
		return errResult, nil
	}

	destination := tools.GetStringParam(req, "destination")
	if destination == "" {
		// Default to Resolume media folder
		home, _ := os.UserHomeDir()
		destination = filepath.Join(home, "Documents", "Resolume Arena", "Media")
	}
	recursive := tools.GetBoolParam(req, "recursive", false)

	// Expand ~ in path
	if strings.HasPrefix(destination, "~") {
		home, _ := os.UserHomeDir()
		destination = filepath.Join(home, destination[1:])
	}

	if err := sanitize.RclonePath(destination); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid destination path: %w", err)), nil
	}

	// Create destination if it doesn't exist
	if err := os.MkdirAll(destination, 0755); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to create destination: %w", err)), nil
	}

	// Get source folder info
	folderInfo, err := client.GetFileInfo(ctx, sourceFolderID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get folder info: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## 🎬 VJ Sync: %s\n\n", folderInfo.Name))
	sb.WriteString("Syncing video clips to Resolume media folder...\n\n")

	// Download videos
	results, err := client.DownloadFolder(ctx, sourceFolderID, destination, "video", recursive)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("sync failed: %w", err)), nil
	}

	successCount := 0
	errorCount := 0
	var totalBytes int64

	for _, r := range results {
		if r.Status == "completed" {
			successCount++
			totalBytes += r.DownloadedBytes
		} else {
			errorCount++
		}
	}

	sb.WriteString(fmt.Sprintf("✅ **%d clips** synced (%s)\n", successCount, clients.FormatFileSize(totalBytes)))
	if errorCount > 0 {
		sb.WriteString(fmt.Sprintf("⚠️ %d files failed\n", errorCount))
	}
	sb.WriteString(fmt.Sprintf("\n📁 **Resolume Media:** %s\n", destination))
	sb.WriteString("\n*Clips are ready in Resolume Arena*")

	return tools.TextResult(sb.String()), nil
}

// handleVJPacks lists available VJ packs for sync
func handleVJPacks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var sb strings.Builder
	sb.WriteString("## Available VJ Packs\n\n")
	sb.WriteString("Use `aftrs_gdrive_rclone_sync` with pack name to sync.\n\n")

	// Check if rclone is available
	_, err := exec.LookPath("rclone")
	rcloneAvailable := err == nil

	if rcloneAvailable {
		sb.WriteString("| Pack | Folder |\n")
		sb.WriteString("|------|--------|\n")
		for key, folder := range vjPacks {
			sb.WriteString(fmt.Sprintf("| `%s` | %s |\n", key, folder))
		}
		sb.WriteString("\n### Quick Sync Commands\n")
		sb.WriteString("```\n")
		sb.WriteString("aftrs_gdrive_rclone_sync pack=\"hackerglasses\"\n")
		sb.WriteString("aftrs_gdrive_rclone_sync pack=\"hairglasses,fetz,masks\"\n")
		sb.WriteString("aftrs_gdrive_rclone_sync pack=\"all\"\n")
		sb.WriteString("```\n")
	} else {
		sb.WriteString("⚠️ **rclone not installed**\n\n")
		sb.WriteString("Install with: `brew install rclone`\n\n")
		sb.WriteString("Available packs:\n")
		for key, folder := range vjPacks {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n", key, folder))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleRcloneSync syncs VJ packs using rclone (faster than API)
func handleRcloneSync(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Check rclone is available
	rclonePath, err := exec.LookPath("rclone")
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("rclone not installed - run: brew install rclone")), nil
	}

	packInput, errResult := tools.RequireStringParam(req, "pack")
	if errResult != nil {
		return errResult, nil
	}

	destination := tools.GetStringParam(req, "destination")
	if destination == "" {
		home, _ := os.UserHomeDir()
		destination = filepath.Join(home, "Documents", "Resolume Arena", "Media")
	}

	// Expand ~ in path
	if strings.HasPrefix(destination, "~") {
		home, _ := os.UserHomeDir()
		destination = filepath.Join(home, destination[1:])
	}

	if err := sanitize.RclonePath(destination); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid destination path: %w", err)), nil
	}

	dryRun := tools.GetBoolParam(req, "dry_run", false)

	var sb strings.Builder
	sb.WriteString("## 🎬 VJ Pack Sync (rclone)\n\n")

	// Determine packs to sync
	var packsToSync []string
	if packInput == "all" {
		for key := range vjPacks {
			packsToSync = append(packsToSync, key)
		}
		sb.WriteString("Syncing **all packs**...\n\n")
	} else {
		packsToSync = strings.Split(packInput, ",")
		for i := range packsToSync {
			packsToSync[i] = strings.TrimSpace(packsToSync[i])
		}
	}

	// Validate packs
	for _, p := range packsToSync {
		if _, ok := vjPacks[p]; !ok {
			return tools.ErrorResult(fmt.Errorf("unknown pack: %s (use aftrs_gdrive_vj_packs to list)", p)), nil
		}
	}

	// Create destination
	if err := os.MkdirAll(destination, 0755); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to create destination: %w", err)), nil
	}

	// Sync each pack
	for _, packKey := range packsToSync {
		packFolder := vjPacks[packKey]
		source := fmt.Sprintf("%s/%s", vjBasePath, packFolder)
		dest := filepath.Join(destination, packFolder)

		sb.WriteString(fmt.Sprintf("### %s\n", packFolder))
		sb.WriteString(fmt.Sprintf("Source: `%s`\n", source))
		sb.WriteString(fmt.Sprintf("Dest: `%s`\n", dest))

		// Build rclone command
		args := []string{
			"sync", source, dest,
			"--transfers", "8",
			"--checkers", "16",
			"--buffer-size", "256M",
		}
		if dryRun {
			args = append(args, "--dry-run")
		}

		cmd := exec.CommandContext(ctx, rclonePath, args...)
		output, err := cmd.CombinedOutput()

		if err != nil {
			sb.WriteString(fmt.Sprintf("❌ **Error:** %v\n", err))
			if len(output) > 0 {
				sb.WriteString(fmt.Sprintf("```\n%s\n```\n", string(output)))
			}
		} else {
			if dryRun {
				sb.WriteString("✅ **Dry run complete**\n")
			} else {
				sb.WriteString("✅ **Synced**\n")
			}
			if len(output) > 0 && len(output) < 2000 {
				sb.WriteString(fmt.Sprintf("```\n%s\n```\n", string(output)))
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("---\n")
	if dryRun {
		sb.WriteString("*Dry run mode - no files transferred*\n")
	} else {
		sb.WriteString(fmt.Sprintf("📁 **Synced to:** %s\n", destination))
		sb.WriteString("*Clips ready in Resolume Arena*\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleVJZipPacks lists available VJ zip pack archives
func handleVJZipPacks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	packs := clients.VJZipPackCatalog()

	var sb strings.Builder
	sb.WriteString("## VJ Clip Zip Packs\n\n")
	sb.WriteString("Downloadable VJ clip archives from Google Drive.\n\n")

	sb.WriteString("| Key | Name | Size | Description |\n")
	sb.WriteString("|-----|------|------|-------------|\n")

	var totalSize float64
	for _, pack := range packs {
		parts := ""
		if len(pack.Parts) > 0 {
			parts = fmt.Sprintf(" (%d parts)", len(pack.Parts))
		}
		sb.WriteString(fmt.Sprintf("| `%s` | %s%s | %.2f GB | %s |\n",
			pack.Key, pack.Name, parts, pack.SizeGB, pack.Description))
		totalSize += pack.SizeGB
	}

	sb.WriteString(fmt.Sprintf("\n**Total:** %.2f GB across %d packs\n\n", totalSize, len(packs)))
	sb.WriteString("### Download Command\n```\naftrs_gdrive_vj_zip_download key=\"supreme_cyphers_v3\"\n```\n")

	return tools.TextResult(sb.String()), nil
}

// handleVJZipSearch searches VJ zip packs
func handleVJZipSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	packs := clients.SearchVJZipPacks(query)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## VJ Zip Search: '%s'\n\n", query))

	if len(packs) == 0 {
		sb.WriteString("No matching packs found.\n")
		sb.WriteString("\nUse `aftrs_gdrive_vj_zip_packs` to list all available packs.\n")
	} else {
		sb.WriteString(fmt.Sprintf("Found **%d** packs:\n\n", len(packs)))
		for _, pack := range packs {
			sb.WriteString(fmt.Sprintf("### %s (`%s`)\n", pack.Name, pack.Key))
			sb.WriteString(fmt.Sprintf("- **Size:** %.2f GB\n", pack.SizeGB))
			sb.WriteString(fmt.Sprintf("- **Description:** %s\n", pack.Description))
			if len(pack.Tags) > 0 {
				sb.WriteString(fmt.Sprintf("- **Tags:** %s\n", strings.Join(pack.Tags, ", ")))
			}
			sb.WriteString("\n")
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleVJZipDownload downloads and optionally extracts a VJ zip pack
func handleVJZipDownload(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	key, errResult := tools.RequireStringParam(req, "key")
	if errResult != nil {
		return errResult, nil
	}

	pack := clients.GetVJZipPack(key)
	if pack == nil {
		return tools.ErrorResult(fmt.Errorf("unknown pack: %s (use aftrs_gdrive_vj_zip_packs to list)", key)), nil
	}

	destination := tools.GetStringParam(req, "destination")
	if destination == "" {
		home, _ := os.UserHomeDir()
		destination = filepath.Join(home, "Documents", "Resolume Arena", "Media")
	}

	// Expand ~ in path
	if strings.HasPrefix(destination, "~") {
		home, _ := os.UserHomeDir()
		destination = filepath.Join(home, destination[1:])
	}

	if err := sanitize.RclonePath(destination); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid destination path: %w", err)), nil
	}

	extract := tools.GetBoolParam(req, "extract", true)
	dryRun := tools.GetBoolParam(req, "dry_run", false)

	// Check rclone is available
	rclonePath, err := exec.LookPath("rclone")
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("rclone not installed - run: brew install rclone")), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Downloading: %s\n\n", pack.Name))
	sb.WriteString(fmt.Sprintf("- **Size:** %.2f GB\n", pack.SizeGB))
	sb.WriteString(fmt.Sprintf("- **Destination:** %s\n\n", destination))

	// Create destination
	if err := os.MkdirAll(destination, 0755); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to create destination: %w", err)), nil
	}

	// Determine what to download
	var filesToDownload []string
	if len(pack.Parts) > 0 {
		// Multi-part archive - download all parts
		for _, part := range pack.Parts {
			filesToDownload = append(filesToDownload, filepath.Join(pack.Path, part))
		}
		sb.WriteString(fmt.Sprintf("Downloading %d parts...\n\n", len(pack.Parts)))
	} else if strings.HasSuffix(pack.Path, ".zip") {
		// Single zip file
		filesToDownload = append(filesToDownload, pack.Path)
	} else {
		// Directory - sync entire directory
		source := fmt.Sprintf("%s/%s", vjBasePath, pack.Path)
		args := []string{"sync", source, filepath.Join(destination, filepath.Base(pack.Path)), "--transfers", "8"}
		if dryRun {
			args = append(args, "--dry-run")
		}
		cmd := exec.CommandContext(ctx, rclonePath, args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("rclone failed: %w\n%s", err, string(output))), nil
		}
		sb.WriteString("Directory synced\n")
		if len(output) > 0 && len(output) < 1000 {
			sb.WriteString(fmt.Sprintf("```\n%s\n```\n", string(output)))
		}
		return tools.TextResult(sb.String()), nil
	}

	// Download zip files
	for _, file := range filesToDownload {
		source := fmt.Sprintf("%s/%s", vjBasePath, file)
		destFile := filepath.Join(destination, filepath.Base(file))

		sb.WriteString(fmt.Sprintf("### %s\n", filepath.Base(file)))

		args := []string{"copy", source, destination, "--progress", "--transfers", "4"}
		if dryRun {
			args = append(args, "--dry-run")
			sb.WriteString("*Dry run - no download*\n")
		}

		cmd := exec.CommandContext(ctx, rclonePath, args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			sb.WriteString(fmt.Sprintf("Error: %v\n", err))
			if len(output) > 0 {
				sb.WriteString(fmt.Sprintf("```\n%s\n```\n", string(output)))
			}
			continue
		}

		if dryRun {
			sb.WriteString("Would download\n")
		} else {
			sb.WriteString("Downloaded\n")

			// Extract if requested
			if extract && strings.HasSuffix(destFile, ".zip") {
				sb.WriteString("Extracting...\n")
				unzipCmd := exec.CommandContext(ctx, "unzip", "-o", destFile, "-d", destination)
				unzipOut, err := unzipCmd.CombinedOutput()
				if err != nil {
					sb.WriteString(fmt.Sprintf("Extract error: %v\n", err))
				} else {
					sb.WriteString("Extracted\n")
					// Count extracted files
					lines := strings.Split(string(unzipOut), "\n")
					fileCount := 0
					for _, line := range lines {
						if strings.Contains(line, "extracting:") || strings.Contains(line, "inflating:") {
							fileCount++
						}
					}
					if fileCount > 0 {
						sb.WriteString(fmt.Sprintf("   %d files extracted\n", fileCount))
					}
				}
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("---\n")
	sb.WriteString(fmt.Sprintf("**Location:** %s\n", destination))
	if !dryRun && extract {
		sb.WriteString("*Clips ready for Resolume Arena*\n")
	}

	return tools.TextResult(sb.String()), nil
}
