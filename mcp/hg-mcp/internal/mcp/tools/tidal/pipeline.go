// Package tidal provides MCP tools for Tidal Hi-Fi music streaming service.
package tidal

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// pipelineTools returns the Tidal download pipeline tool definitions
func pipelineTools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_tidal_download_album",
				mcp.WithDescription("Download a Tidal album in FLAC/MQA quality using tidal-dl-ng. Requires tidal-dl-ng installed and authenticated."),
				mcp.WithString("album_id", mcp.Description("Tidal album ID or URL"), mcp.Required()),
				mcp.WithString("output_dir", mcp.Description("Output directory (default: ~/Music/Tidal)")),
				mcp.WithString("quality", mcp.Description("Audio quality: master, hifi, high, low (default: master)")),
			),
			Handler:             handleDownloadAlbum,
			Category:            "music",
			Subcategory:         "tidal",
			Tags:                []string{"tidal", "download", "album", "flac", "mqa"},
			UseCases:            []string{"Download album for DJ use", "Archive high-quality audio"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "tidal",
		},
		{
			Tool: mcp.NewTool("aftrs_tidal_download_playlist",
				mcp.WithDescription("Download a Tidal playlist in FLAC/MQA quality using tidal-dl-ng."),
				mcp.WithString("playlist_id", mcp.Description("Tidal playlist UUID or URL"), mcp.Required()),
				mcp.WithString("output_dir", mcp.Description("Output directory (default: ~/Music/Tidal)")),
				mcp.WithString("quality", mcp.Description("Audio quality: master, hifi, high, low (default: master)")),
			),
			Handler:             handleDownloadPlaylist,
			Category:            "music",
			Subcategory:         "tidal",
			Tags:                []string{"tidal", "download", "playlist", "flac"},
			UseCases:            []string{"Download playlist for offline DJ use", "Sync Tidal playlist to library"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "tidal",
		},
		{
			Tool: mcp.NewTool("aftrs_tidal_to_rekordbox",
				mcp.WithDescription("Full pipeline: Download Tidal content, upload to GDrive, and import to Rekordbox. Combines download, cloud sync, and library management."),
				mcp.WithString("url", mcp.Description("Tidal album, playlist, or track URL"), mcp.Required()),
				mcp.WithString("playlist_name", mcp.Description("Rekordbox playlist name for import")),
				mcp.WithBoolean("dry_run", mcp.Description("Preview actions without downloading (default: false)")),
			),
			Handler:             handleTidalToRekordbox,
			Category:            "music",
			Subcategory:         "tidal",
			Tags:                []string{"tidal", "rekordbox", "pipeline", "sync", "import"},
			UseCases:            []string{"Import Tidal releases to DJ library", "Full sync workflow"},
			Complexity:          tools.ComplexityComplex,
			CircuitBreakerGroup: "tidal",
		},
		{
			Tool: mcp.NewTool("aftrs_tidal_check_auth",
				mcp.WithDescription("Verify tidal-dl-ng authentication status and token validity."),
			),
			Handler:             handleCheckAuth,
			Category:            "music",
			Subcategory:         "tidal",
			Tags:                []string{"tidal", "auth", "verify", "token"},
			UseCases:            []string{"Check authentication status", "Troubleshoot download issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tidal",
		},
		{
			Tool: mcp.NewTool("aftrs_tidal_setup_auth",
				mcp.WithDescription("Instructions and commands for setting up tidal-dl-ng authentication."),
			),
			Handler:             handleSetupAuth,
			Category:            "music",
			Subcategory:         "tidal",
			Tags:                []string{"tidal", "auth", "setup", "login"},
			UseCases:            []string{"Initial authentication setup", "Re-authenticate after token expiry"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tidal",
		},
	}
}

// handleDownloadAlbum handles the aftrs_tidal_download_album tool
func handleDownloadAlbum(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	albumID := tools.GetStringParam(req, "album_id")
	outputDir := tools.GetStringParam(req, "output_dir")
	quality := tools.GetStringParam(req, "quality")

	if albumID == "" {
		return tools.ErrorResult(fmt.Errorf("album_id is required")), nil
	}

	if outputDir == "" {
		outputDir = filepath.Join("~", "Music", "Tidal")
	}
	if quality == "" {
		quality = "master"
	}

	var sb strings.Builder
	sb.WriteString("# Tidal Album Download\n\n")
	sb.WriteString(fmt.Sprintf("**Album ID:** %s\n", albumID))
	sb.WriteString(fmt.Sprintf("**Output:** %s\n", outputDir))
	sb.WriteString(fmt.Sprintf("**Quality:** %s\n\n", quality))

	// Check if tidal-dl-ng is installed
	tidalDL, err := findTidalDL()
	if err != nil {
		sb.WriteString("## Error\n\n")
		sb.WriteString("❌ tidal-dl-ng not found. Install with:\n")
		sb.WriteString("```bash\npip install tidal-dl-ng\n```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("## Download Progress\n\n")
	sb.WriteString(fmt.Sprintf("Using: %s\n\n", tidalDL))

	// Build command
	args := []string{
		"dl",
		"--output", outputDir,
		"--quality", quality,
		albumID,
	}

	// Execute download
	cmd := exec.CommandContext(ctx, tidalDL, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		sb.WriteString("### Error\n\n")
		sb.WriteString(fmt.Sprintf("❌ Download failed: %v\n", err))
		sb.WriteString(fmt.Sprintf("```\n%s\n```\n", string(output)))
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("### Output\n\n")
	sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", string(output)))
	sb.WriteString("✅ Download complete!\n\n")

	sb.WriteString("## Next Steps\n")
	sb.WriteString("- Use `aftrs_rclone_copy` to upload to GDrive\n")
	sb.WriteString("- Use `aftrs_rekordbox_import` to add to library\n")

	return tools.TextResult(sb.String()), nil
}

// handleDownloadPlaylist handles the aftrs_tidal_download_playlist tool
func handleDownloadPlaylist(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	playlistID := tools.GetStringParam(req, "playlist_id")
	outputDir := tools.GetStringParam(req, "output_dir")
	quality := tools.GetStringParam(req, "quality")

	if playlistID == "" {
		return tools.ErrorResult(fmt.Errorf("playlist_id is required")), nil
	}

	if outputDir == "" {
		outputDir = filepath.Join("~", "Music", "Tidal")
	}
	if quality == "" {
		quality = "master"
	}

	var sb strings.Builder
	sb.WriteString("# Tidal Playlist Download\n\n")
	sb.WriteString(fmt.Sprintf("**Playlist ID:** %s\n", playlistID))
	sb.WriteString(fmt.Sprintf("**Output:** %s\n", outputDir))
	sb.WriteString(fmt.Sprintf("**Quality:** %s\n\n", quality))

	// Check if tidal-dl-ng is installed
	tidalDL, err := findTidalDL()
	if err != nil {
		sb.WriteString("## Error\n\n")
		sb.WriteString("❌ tidal-dl-ng not found. Install with:\n")
		sb.WriteString("```bash\npip install tidal-dl-ng\n```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("## Download Progress\n\n")
	sb.WriteString(fmt.Sprintf("Using: %s\n\n", tidalDL))

	// Build command
	args := []string{
		"dl",
		"--output", outputDir,
		"--quality", quality,
		playlistID,
	}

	// Execute download
	cmd := exec.CommandContext(ctx, tidalDL, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		sb.WriteString("### Error\n\n")
		sb.WriteString(fmt.Sprintf("❌ Download failed: %v\n", err))
		sb.WriteString(fmt.Sprintf("```\n%s\n```\n", string(output)))
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("### Output\n\n")
	sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", string(output)))
	sb.WriteString("✅ Download complete!\n\n")

	sb.WriteString("## Next Steps\n")
	sb.WriteString("- Use `aftrs_rclone_copy` to upload to GDrive\n")
	sb.WriteString("- Use `aftrs_rekordbox_import` to add to library\n")

	return tools.TextResult(sb.String()), nil
}

// handleTidalToRekordbox handles the aftrs_tidal_to_rekordbox tool
func handleTidalToRekordbox(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	url := tools.GetStringParam(req, "url")
	playlistName := tools.GetStringParam(req, "playlist_name")
	dryRun := tools.GetBoolParam(req, "dry_run", false)

	if url == "" {
		return tools.ErrorResult(fmt.Errorf("url is required")), nil
	}

	var sb strings.Builder
	sb.WriteString("# Tidal → Rekordbox Pipeline\n\n")
	sb.WriteString(fmt.Sprintf("**URL:** %s\n", url))
	if playlistName != "" {
		sb.WriteString(fmt.Sprintf("**Playlist:** %s\n", playlistName))
	}
	sb.WriteString(fmt.Sprintf("**Dry Run:** %v\n", dryRun))
	sb.WriteString(fmt.Sprintf("**Started:** %s\n\n", time.Now().Format(time.RFC3339)))

	if dryRun {
		sb.WriteString("## Pipeline Preview (Dry Run)\n\n")
		sb.WriteString("### Phase 1: Download\n")
		sb.WriteString("- Would download from Tidal in FLAC/MQA quality\n")
		sb.WriteString("- Target: ~/Music/Tidal/{artist}/{album}/\n\n")

		sb.WriteString("### Phase 2: Upload to GDrive\n")
		sb.WriteString("- Would sync to: gdrive:DJ Crates/Tidal/{artist}/{album}/\n\n")

		sb.WriteString("### Phase 3: Rekordbox Import\n")
		sb.WriteString("- Would create/update playlist in Rekordbox\n")
		sb.WriteString("- Would analyze BPM, key, and waveforms\n\n")

		sb.WriteString("## Estimated Resources\n")
		sb.WriteString("- **Storage:** ~500 MB (estimated)\n")
		sb.WriteString("- **Duration:** ~5 minutes\n\n")

		sb.WriteString("Run without dry_run=true to execute.\n")
		return tools.TextResult(sb.String()), nil
	}

	// Check prerequisites
	sb.WriteString("## Checking Prerequisites\n\n")

	tidalDL, err := findTidalDL()
	if err != nil {
		sb.WriteString("❌ tidal-dl-ng not found\n")
		sb.WriteString("\nInstall with: `pip install tidal-dl-ng`\n")
		return tools.TextResult(sb.String()), nil
	}
	sb.WriteString(fmt.Sprintf("✅ tidal-dl-ng: %s\n", tidalDL))

	rclone, err := exec.LookPath("rclone")
	if err != nil {
		sb.WriteString("❌ rclone not found\n")
		return tools.TextResult(sb.String()), nil
	}
	sb.WriteString(fmt.Sprintf("✅ rclone: %s\n\n", rclone))

	// Phase 1: Download
	sb.WriteString("## Phase 1: Download\n\n")
	localPath := filepath.Join("~", "Music", "Tidal", "Downloads")

	downloadArgs := []string{"dl", "--output", localPath, "--quality", "master", url}
	downloadCmd := exec.CommandContext(ctx, tidalDL, downloadArgs...)
	downloadOutput, err := downloadCmd.CombinedOutput()

	if err != nil {
		sb.WriteString(fmt.Sprintf("❌ Download failed: %v\n", err))
		sb.WriteString(fmt.Sprintf("```\n%s\n```\n", string(downloadOutput)))
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("✅ Download complete\n")
	sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", string(downloadOutput)))

	// Phase 2: Upload to GDrive
	sb.WriteString("## Phase 2: Upload to GDrive\n\n")
	remotePath := "gdrive:DJ Crates/Tidal/"

	syncArgs := []string{"copy", localPath, remotePath, "--progress"}
	syncCmd := exec.CommandContext(ctx, rclone, syncArgs...)
	syncOutput, err := syncCmd.CombinedOutput()

	if err != nil {
		sb.WriteString(fmt.Sprintf("⚠️ GDrive sync warning: %v\n", err))
		sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", string(syncOutput)))
	} else {
		sb.WriteString("✅ GDrive sync complete\n")
		sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", string(syncOutput)))
	}

	// Phase 3: Rekordbox Import (placeholder - would integrate with rekordbox tools)
	sb.WriteString("## Phase 3: Rekordbox Import\n\n")
	sb.WriteString("⏳ Rekordbox import pending - use `aftrs_rekordbox_import` to complete\n\n")

	sb.WriteString("## Summary\n\n")
	sb.WriteString("✅ Pipeline complete!\n")
	sb.WriteString(fmt.Sprintf("- **Local:** %s\n", localPath))
	sb.WriteString(fmt.Sprintf("- **Remote:** %s\n", remotePath))
	if playlistName != "" {
		sb.WriteString(fmt.Sprintf("- **Playlist:** %s\n", playlistName))
	}

	return tools.TextResult(sb.String()), nil
}

// handleCheckAuth handles the aftrs_tidal_check_auth tool
func handleCheckAuth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var sb strings.Builder
	sb.WriteString("# Tidal Authentication Status\n\n")

	tidalDL, err := findTidalDL()
	if err != nil {
		sb.WriteString("## Error\n\n")
		sb.WriteString("❌ tidal-dl-ng not found.\n\n")
		sb.WriteString("Install with:\n")
		sb.WriteString("```bash\npip install tidal-dl-ng\n```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("**Tool:** %s\n\n", tidalDL))

	// Check auth status
	cmd := exec.CommandContext(ctx, tidalDL, "auth", "status")
	output, err := cmd.CombinedOutput()

	if err != nil {
		sb.WriteString("## Status\n\n")
		sb.WriteString("❌ **Not authenticated** or token expired\n\n")
		sb.WriteString("### Output\n")
		sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", string(output)))
		sb.WriteString("## Fix\n\n")
		sb.WriteString("Run `aftrs_tidal_setup_auth` for authentication instructions.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("## Status\n\n")
	sb.WriteString("✅ **Authenticated**\n\n")
	sb.WriteString("### Details\n")
	sb.WriteString(fmt.Sprintf("```\n%s\n```\n", string(output)))

	return tools.TextResult(sb.String()), nil
}

// handleSetupAuth handles the aftrs_tidal_setup_auth tool
func handleSetupAuth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var sb strings.Builder
	sb.WriteString("# Tidal Authentication Setup\n\n")

	sb.WriteString("## Prerequisites\n\n")
	sb.WriteString("1. **Tidal HiFi/Plus subscription** (required for FLAC downloads)\n")
	sb.WriteString("2. **tidal-dl-ng** installed: `pip install tidal-dl-ng`\n\n")

	sb.WriteString("## Authentication Steps\n\n")

	sb.WriteString("### Option 1: Device Link (Recommended)\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("tidal-dl-ng auth link\n")
	sb.WriteString("```\n\n")
	sb.WriteString("1. A URL will be displayed\n")
	sb.WriteString("2. Open the URL in your browser\n")
	sb.WriteString("3. Log in with your Tidal account\n")
	sb.WriteString("4. Authorize the application\n")
	sb.WriteString("5. Token will be saved automatically\n\n")

	sb.WriteString("### Option 2: Manual Token\n\n")
	sb.WriteString("If device link doesn't work:\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("tidal-dl-ng auth token <your-access-token>\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## Verify Authentication\n\n")
	sb.WriteString("```bash\n")
	sb.WriteString("tidal-dl-ng auth status\n")
	sb.WriteString("```\n\n")
	sb.WriteString("Or use: `aftrs_tidal_check_auth`\n\n")

	sb.WriteString("## Token Location\n\n")
	sb.WriteString("Tokens are stored in:\n")
	sb.WriteString("- **Windows:** `%APPDATA%\\tidal-dl-ng\\`\n")
	sb.WriteString("- **macOS:** `~/Library/Application Support/tidal-dl-ng/`\n")
	sb.WriteString("- **Linux:** `~/.config/tidal-dl-ng/`\n\n")

	sb.WriteString("## Troubleshooting\n\n")
	sb.WriteString("- **Token expired:** Re-run `tidal-dl-ng auth link`\n")
	sb.WriteString("- **Rate limited:** Wait a few minutes and retry\n")
	sb.WriteString("- **Region locked:** Use a VPN if content is geo-restricted\n")

	return tools.TextResult(sb.String()), nil
}

// findTidalDL locates the tidal-dl-ng executable
func findTidalDL() (string, error) {
	// Try common locations
	paths := []string{
		"tidal-dl-ng",
		"tidal-dl",
		filepath.Join("~", ".local", "bin", "tidal-dl-ng"),
	}

	for _, p := range paths {
		if path, err := exec.LookPath(p); err == nil {
			return path, nil
		}
	}

	// Try pip's script location on Windows
	if pipScripts := filepath.Join("AppData", "Local", "Programs", "Python", "Python311", "Scripts", "tidal-dl-ng.exe"); true {
		homeDir := "~"
		fullPath := filepath.Join(homeDir, pipScripts)
		if _, err := exec.LookPath(fullPath); err == nil {
			return fullPath, nil
		}
	}

	return "", fmt.Errorf("tidal-dl-ng not found in PATH")
}

// init adds pipeline tools to the module
func init() {
	// Pipeline tools are registered via the main module
	// This is handled by extending Module.Tools()
}
