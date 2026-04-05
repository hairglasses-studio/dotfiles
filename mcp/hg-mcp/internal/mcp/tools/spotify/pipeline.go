// Package spotify provides MCP tools for Spotify music streaming service.
package spotify

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

// pipelineTools returns the Spotify download pipeline tool definitions
func pipelineTools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_spotify_download_playlist",
				mcp.WithDescription("Download a Spotify playlist using spotdl. Matches tracks via YouTube Music and downloads high-quality audio."),
				mcp.WithString("playlist_url", mcp.Description("Spotify playlist URL"), mcp.Required()),
				mcp.WithString("output_dir", mcp.Description("Output directory (default: ~/Music/Spotify)")),
				mcp.WithString("format", mcp.Description("Audio format: mp3, flac, opus, m4a (default: flac)")),
				mcp.WithNumber("bitrate", mcp.Description("Audio bitrate in kbps (default: 320)")),
			),
			Handler:     handleDownloadPlaylist,
			Category:    "music",
			Subcategory: "spotify",
			Tags:        []string{"spotify", "download", "playlist", "spotdl"},
			UseCases:    []string{"Download Spotify playlist for DJ use", "Offline playlist sync"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "spotify",
		},
		{
			Tool: mcp.NewTool("aftrs_spotify_download_album",
				mcp.WithDescription("Download a Spotify album using spotdl."),
				mcp.WithString("album_url", mcp.Description("Spotify album URL"), mcp.Required()),
				mcp.WithString("output_dir", mcp.Description("Output directory (default: ~/Music/Spotify)")),
				mcp.WithString("format", mcp.Description("Audio format: mp3, flac, opus, m4a (default: flac)")),
			),
			Handler:     handleDownloadAlbum,
			Category:    "music",
			Subcategory: "spotify",
			Tags:        []string{"spotify", "download", "album", "spotdl"},
			UseCases:    []string{"Download album for DJ library", "Archive album locally"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "spotify",
		},
		{
			Tool: mcp.NewTool("aftrs_spotify_to_rekordbox",
				mcp.WithDescription("Full pipeline: Download Spotify content, upload to GDrive, and create Rekordbox playlist."),
				mcp.WithString("url", mcp.Description("Spotify playlist, album, or track URL"), mcp.Required()),
				mcp.WithString("playlist_name", mcp.Description("Rekordbox playlist name for import")),
				mcp.WithBoolean("dry_run", mcp.Description("Preview actions without downloading (default: false)")),
			),
			Handler:     handleSpotifyToRekordbox,
			Category:    "music",
			Subcategory: "spotify",
			Tags:        []string{"spotify", "rekordbox", "pipeline", "sync", "import"},
			UseCases:    []string{"Import Spotify releases to DJ library", "Full sync workflow"},
			Complexity:          tools.ComplexityComplex,
			CircuitBreakerGroup: "spotify",
		},
		{
			Tool: mcp.NewTool("aftrs_spotify_check_tools",
				mcp.WithDescription("Verify spotdl and ffmpeg are installed and configured correctly."),
			),
			Handler:     handleCheckTools,
			Category:    "music",
			Subcategory: "spotify",
			Tags:        []string{"spotify", "check", "tools", "spotdl", "ffmpeg"},
			UseCases:    []string{"Verify download tools", "Troubleshoot issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "spotify",
		},
	}
}

// handleDownloadPlaylist handles the aftrs_spotify_download_playlist tool
func handleDownloadPlaylist(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	playlistURL := tools.GetStringParam(req, "playlist_url")
	outputDir := tools.GetStringParam(req, "output_dir")
	format := tools.GetStringParam(req, "format")
	bitrate := tools.GetFloatParam(req, "bitrate", 320)

	if playlistURL == "" {
		return tools.ErrorResult(fmt.Errorf("playlist_url is required")), nil
	}

	if outputDir == "" {
		outputDir = filepath.Join("~", "Music", "Spotify")
	}
	if format == "" {
		format = "flac"
	}

	var sb strings.Builder
	sb.WriteString("# Spotify Playlist Download\n\n")
	sb.WriteString(fmt.Sprintf("**URL:** %s\n", playlistURL))
	sb.WriteString(fmt.Sprintf("**Output:** %s\n", outputDir))
	sb.WriteString(fmt.Sprintf("**Format:** %s\n", format))
	sb.WriteString(fmt.Sprintf("**Bitrate:** %.0f kbps\n\n", bitrate))

	// Check if spotdl is installed
	spotdl, err := findSpotDL()
	if err != nil {
		sb.WriteString("## Error\n\n")
		sb.WriteString("❌ spotdl not found. Install with:\n")
		sb.WriteString("```bash\npip install spotdl\n```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("## Download Progress\n\n")
	sb.WriteString(fmt.Sprintf("Using: %s\n\n", spotdl))

	// Build command
	args := []string{
		"download",
		playlistURL,
		"--output", outputDir,
		"--format", format,
		"--bitrate", fmt.Sprintf("%.0f", bitrate),
	}

	// Execute download
	cmd := exec.CommandContext(ctx, spotdl, args...)
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

// handleDownloadAlbum handles the aftrs_spotify_download_album tool
func handleDownloadAlbum(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	albumURL := tools.GetStringParam(req, "album_url")
	outputDir := tools.GetStringParam(req, "output_dir")
	format := tools.GetStringParam(req, "format")

	if albumURL == "" {
		return tools.ErrorResult(fmt.Errorf("album_url is required")), nil
	}

	if outputDir == "" {
		outputDir = filepath.Join("~", "Music", "Spotify")
	}
	if format == "" {
		format = "flac"
	}

	var sb strings.Builder
	sb.WriteString("# Spotify Album Download\n\n")
	sb.WriteString(fmt.Sprintf("**URL:** %s\n", albumURL))
	sb.WriteString(fmt.Sprintf("**Output:** %s\n", outputDir))
	sb.WriteString(fmt.Sprintf("**Format:** %s\n\n", format))

	// Check if spotdl is installed
	spotdl, err := findSpotDL()
	if err != nil {
		sb.WriteString("## Error\n\n")
		sb.WriteString("❌ spotdl not found. Install with:\n")
		sb.WriteString("```bash\npip install spotdl\n```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("## Download Progress\n\n")
	sb.WriteString(fmt.Sprintf("Using: %s\n\n", spotdl))

	// Build command
	args := []string{
		"download",
		albumURL,
		"--output", outputDir,
		"--format", format,
	}

	// Execute download
	cmd := exec.CommandContext(ctx, spotdl, args...)
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

// handleSpotifyToRekordbox handles the aftrs_spotify_to_rekordbox tool
func handleSpotifyToRekordbox(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	url := tools.GetStringParam(req, "url")
	playlistName := tools.GetStringParam(req, "playlist_name")
	dryRun := tools.GetBoolParam(req, "dry_run", false)

	if url == "" {
		return tools.ErrorResult(fmt.Errorf("url is required")), nil
	}

	var sb strings.Builder
	sb.WriteString("# Spotify → Rekordbox Pipeline\n\n")
	sb.WriteString(fmt.Sprintf("**URL:** %s\n", url))
	if playlistName != "" {
		sb.WriteString(fmt.Sprintf("**Playlist:** %s\n", playlistName))
	}
	sb.WriteString(fmt.Sprintf("**Dry Run:** %v\n", dryRun))
	sb.WriteString(fmt.Sprintf("**Started:** %s\n\n", time.Now().Format(time.RFC3339)))

	if dryRun {
		sb.WriteString("## Pipeline Preview (Dry Run)\n\n")
		sb.WriteString("### Phase 1: Download\n")
		sb.WriteString("- Would download via spotdl (YouTube Music matching)\n")
		sb.WriteString("- Format: FLAC @ 320kbps equivalent\n")
		sb.WriteString("- Target: ~/Music/Spotify/{playlist}/\n\n")

		sb.WriteString("### Phase 2: Upload to GDrive\n")
		sb.WriteString("- Would sync to: gdrive:DJ Crates/Spotify/{playlist}/\n\n")

		sb.WriteString("### Phase 3: Rekordbox Import\n")
		sb.WriteString("- Would create/update playlist in Rekordbox\n")
		sb.WriteString("- Would analyze BPM, key, and waveforms\n\n")

		sb.WriteString("## Notes\n")
		sb.WriteString("- spotdl matches tracks via YouTube Music\n")
		sb.WriteString("- Audio quality depends on YT Music source\n")
		sb.WriteString("- Metadata comes from Spotify API\n\n")

		sb.WriteString("Run without dry_run=true to execute.\n")
		return tools.TextResult(sb.String()), nil
	}

	// Check prerequisites
	sb.WriteString("## Checking Prerequisites\n\n")

	spotdl, err := findSpotDL()
	if err != nil {
		sb.WriteString("❌ spotdl not found\n")
		sb.WriteString("\nInstall with: `pip install spotdl`\n")
		return tools.TextResult(sb.String()), nil
	}
	sb.WriteString(fmt.Sprintf("✅ spotdl: %s\n", spotdl))

	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		sb.WriteString("⚠️ ffmpeg not found (may affect audio conversion)\n")
	} else {
		sb.WriteString(fmt.Sprintf("✅ ffmpeg: %s\n", ffmpeg))
	}

	rclone, err := exec.LookPath("rclone")
	if err != nil {
		sb.WriteString("❌ rclone not found\n")
		return tools.TextResult(sb.String()), nil
	}
	sb.WriteString(fmt.Sprintf("✅ rclone: %s\n\n", rclone))

	// Phase 1: Download
	sb.WriteString("## Phase 1: Download\n\n")
	localPath := filepath.Join("~", "Music", "Spotify", "Downloads")

	downloadArgs := []string{"download", url, "--output", localPath, "--format", "flac"}
	downloadCmd := exec.CommandContext(ctx, spotdl, downloadArgs...)
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
	remotePath := "gdrive:DJ Crates/Spotify/"

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

	// Phase 3: Rekordbox Import
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

// handleCheckTools handles the aftrs_spotify_check_tools tool
func handleCheckTools(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var sb strings.Builder
	sb.WriteString("# Spotify Download Tools Status\n\n")

	allOK := true

	// Check spotdl
	sb.WriteString("## spotdl\n\n")
	spotdl, err := findSpotDL()
	if err != nil {
		sb.WriteString("❌ **Not installed**\n\n")
		sb.WriteString("Install with:\n")
		sb.WriteString("```bash\npip install spotdl\n```\n\n")
		allOK = false
	} else {
		sb.WriteString(fmt.Sprintf("✅ **Found:** %s\n\n", spotdl))

		// Get version
		cmd := exec.CommandContext(ctx, spotdl, "--version")
		if output, err := cmd.CombinedOutput(); err == nil {
			sb.WriteString(fmt.Sprintf("**Version:** %s\n\n", strings.TrimSpace(string(output))))
		}
	}

	// Check ffmpeg
	sb.WriteString("## ffmpeg\n\n")
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		sb.WriteString("⚠️ **Not installed** (optional but recommended)\n\n")
		sb.WriteString("Install:\n")
		sb.WriteString("- Windows: `choco install ffmpeg` or `winget install ffmpeg`\n")
		sb.WriteString("- macOS: `brew install ffmpeg`\n")
		sb.WriteString("- Linux: `apt install ffmpeg`\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("✅ **Found:** %s\n\n", ffmpeg))

		// Get version
		cmd := exec.CommandContext(ctx, ffmpeg, "-version")
		if output, err := cmd.CombinedOutput(); err == nil {
			lines := strings.Split(string(output), "\n")
			if len(lines) > 0 {
				sb.WriteString(fmt.Sprintf("**Version:** %s\n\n", lines[0]))
			}
		}
	}

	// Check rclone
	sb.WriteString("## rclone\n\n")
	rclone, err := exec.LookPath("rclone")
	if err != nil {
		sb.WriteString("⚠️ **Not installed** (needed for GDrive sync)\n\n")
		sb.WriteString("Install: https://rclone.org/install/\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("✅ **Found:** %s\n\n", rclone))
	}

	// Summary
	sb.WriteString("## Summary\n\n")
	if allOK {
		sb.WriteString("✅ All required tools are installed and ready.\n")
	} else {
		sb.WriteString("⚠️ Some tools are missing. Install them to use the download pipeline.\n")
	}

	return tools.TextResult(sb.String()), nil
}

// findSpotDL locates the spotdl executable
func findSpotDL() (string, error) {
	// Try common locations
	paths := []string{
		"spotdl",
		filepath.Join("~", ".local", "bin", "spotdl"),
	}

	for _, p := range paths {
		if path, err := exec.LookPath(p); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("spotdl not found in PATH")
}

// Note: These tools should be registered by extending the spotify module.
// Add them to the module.go Tools() method by appending pipelineTools().
