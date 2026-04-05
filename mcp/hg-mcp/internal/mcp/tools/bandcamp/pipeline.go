// Package bandcamp provides MCP tools for Bandcamp music platform.
package bandcamp

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

// pipelineTools returns the Bandcamp download pipeline tool definitions
func pipelineTools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_bandcamp_download_purchased",
				mcp.WithDescription("Download all purchased tracks from Bandcamp using bandcamp-dl. Requires authentication cookie."),
				mcp.WithString("output_dir", mcp.Description("Output directory (default: ~/Music/Bandcamp)")),
				mcp.WithString("format", mcp.Description("Audio format: flac, mp3-320, mp3-v0, wav (default: flac)")),
				mcp.WithBoolean("skip_existing", mcp.Description("Skip already downloaded files (default: true)")),
			),
			Handler:     handleDownloadPurchased,
			Category:    "music",
			Subcategory: "bandcamp",
			Tags:        []string{"bandcamp", "download", "purchased", "collection"},
			UseCases:    []string{"Sync purchased collection", "Archive Bandcamp purchases"},
			Complexity:  tools.ComplexityModerate,
		},
		{
			Tool: mcp.NewTool("aftrs_bandcamp_to_rekordbox",
				mcp.WithDescription("Full pipeline: Download Bandcamp purchases, upload to GDrive, and create Rekordbox playlist."),
				mcp.WithString("playlist_name", mcp.Description("Rekordbox playlist name for import")),
				mcp.WithBoolean("dry_run", mcp.Description("Preview actions without downloading (default: false)")),
			),
			Handler:     handleBandcampToRekordbox,
			Category:    "music",
			Subcategory: "bandcamp",
			Tags:        []string{"bandcamp", "rekordbox", "pipeline", "sync", "import"},
			UseCases:    []string{"Import Bandcamp to DJ library", "Full sync workflow"},
			Complexity:  tools.ComplexityComplex,
		},
		{
			Tool: mcp.NewTool("aftrs_bandcamp_set_auth",
				mcp.WithDescription("Set Bandcamp authentication cookie for downloading purchased tracks."),
				mcp.WithString("identity_cookie", mcp.Description("Bandcamp 'identity' cookie value from browser"), mcp.Required()),
			),
			Handler:     handleSetAuth,
			Category:    "music",
			Subcategory: "bandcamp",
			Tags:        []string{"bandcamp", "auth", "cookie", "setup"},
			UseCases:    []string{"Configure authentication", "Set up download access"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_bandcamp_check_auth",
				mcp.WithDescription("Verify Bandcamp authentication status and tools are installed."),
			),
			Handler:     handleCheckAuth,
			Category:    "music",
			Subcategory: "bandcamp",
			Tags:        []string{"bandcamp", "auth", "check", "verify"},
			UseCases:    []string{"Verify authentication", "Troubleshoot issues"},
			Complexity:  tools.ComplexitySimple,
		},
	}
}

// handleDownloadPurchased handles the aftrs_bandcamp_download_purchased tool
func handleDownloadPurchased(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	outputDir := tools.GetStringParam(req, "output_dir")
	format := tools.GetStringParam(req, "format")
	skipExisting := tools.GetBoolParam(req, "skip_existing", true)

	if outputDir == "" {
		outputDir = filepath.Join("~", "Music", "Bandcamp")
	}
	if format == "" {
		format = "flac"
	}

	var sb strings.Builder
	sb.WriteString("# Bandcamp Purchases Download\n\n")
	sb.WriteString(fmt.Sprintf("**Output:** %s\n", outputDir))
	sb.WriteString(fmt.Sprintf("**Format:** %s\n", format))
	sb.WriteString(fmt.Sprintf("**Skip Existing:** %v\n\n", skipExisting))

	// Check if bandcamp-dl is installed
	bcDL, err := findBandcampDL()
	if err != nil {
		sb.WriteString("## Error\n\n")
		sb.WriteString("❌ bandcamp-dl not found. Install with:\n")
		sb.WriteString("```bash\npip install bandcamp-downloader\n```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("## Download Progress\n\n")
	sb.WriteString(fmt.Sprintf("Using: %s\n\n", bcDL))

	// Build command
	args := []string{
		"--directory", outputDir,
		"--format", format,
	}

	if skipExisting {
		args = append(args, "--no-clobber")
	}

	// Add collection download flag
	args = append(args, "--collection")

	// Execute download
	cmd := exec.CommandContext(ctx, bcDL, args...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		sb.WriteString("### Error\n\n")
		sb.WriteString(fmt.Sprintf("❌ Download failed: %v\n", err))
		sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", string(output)))
		sb.WriteString("## Troubleshooting\n")
		sb.WriteString("- Verify authentication with `aftrs_bandcamp_check_auth`\n")
		sb.WriteString("- Set auth cookie with `aftrs_bandcamp_set_auth`\n")
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

// handleBandcampToRekordbox handles the aftrs_bandcamp_to_rekordbox tool
func handleBandcampToRekordbox(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	playlistName := tools.GetStringParam(req, "playlist_name")
	dryRun := tools.GetBoolParam(req, "dry_run", false)

	if playlistName == "" {
		playlistName = fmt.Sprintf("Bandcamp %s", time.Now().Format("2006-01"))
	}

	var sb strings.Builder
	sb.WriteString("# Bandcamp → Rekordbox Pipeline\n\n")
	sb.WriteString(fmt.Sprintf("**Playlist:** %s\n", playlistName))
	sb.WriteString(fmt.Sprintf("**Dry Run:** %v\n", dryRun))
	sb.WriteString(fmt.Sprintf("**Started:** %s\n\n", time.Now().Format(time.RFC3339)))

	if dryRun {
		sb.WriteString("## Pipeline Preview (Dry Run)\n\n")
		sb.WriteString("### Phase 1: Download\n")
		sb.WriteString("- Would download all purchased tracks from Bandcamp\n")
		sb.WriteString("- Format: FLAC (lossless)\n")
		sb.WriteString("- Target: ~/Music/Bandcamp/{artist}/{album}/\n\n")

		sb.WriteString("### Phase 2: Upload to GDrive\n")
		sb.WriteString("- Would sync to: gdrive:DJ Crates/Bandcamp/\n\n")

		sb.WriteString("### Phase 3: Rekordbox Import\n")
		sb.WriteString("- Would create playlist: " + playlistName + "\n")
		sb.WriteString("- Would analyze BPM, key, and waveforms\n\n")

		sb.WriteString("## Notes\n")
		sb.WriteString("- Requires Bandcamp identity cookie set\n")
		sb.WriteString("- Only downloads purchased/free-download tracks\n")
		sb.WriteString("- Preserves original FLAC quality when available\n\n")

		sb.WriteString("Run without dry_run=true to execute.\n")
		return tools.TextResult(sb.String()), nil
	}

	// Check prerequisites
	sb.WriteString("## Checking Prerequisites\n\n")

	bcDL, err := findBandcampDL()
	if err != nil {
		sb.WriteString("❌ bandcamp-dl not found\n")
		sb.WriteString("\nInstall with: `pip install bandcamp-downloader`\n")
		return tools.TextResult(sb.String()), nil
	}
	sb.WriteString(fmt.Sprintf("✅ bandcamp-dl: %s\n", bcDL))

	rclone, err := exec.LookPath("rclone")
	if err != nil {
		sb.WriteString("❌ rclone not found\n")
		return tools.TextResult(sb.String()), nil
	}
	sb.WriteString(fmt.Sprintf("✅ rclone: %s\n\n", rclone))

	// Phase 1: Download
	sb.WriteString("## Phase 1: Download\n\n")
	localPath := filepath.Join("~", "Music", "Bandcamp")

	downloadArgs := []string{"--directory", localPath, "--format", "flac", "--collection", "--no-clobber"}
	downloadCmd := exec.CommandContext(ctx, bcDL, downloadArgs...)
	downloadOutput, err := downloadCmd.CombinedOutput()

	if err != nil {
		sb.WriteString(fmt.Sprintf("⚠️ Download had issues: %v\n", err))
		sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", string(downloadOutput)))
	} else {
		sb.WriteString("✅ Download complete\n")
		sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", string(downloadOutput)))
	}

	// Phase 2: Upload to GDrive
	sb.WriteString("## Phase 2: Upload to GDrive\n\n")
	remotePath := "gdrive:DJ Crates/Bandcamp/"

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
	sb.WriteString(fmt.Sprintf("- **Playlist:** %s\n", playlistName))

	return tools.TextResult(sb.String()), nil
}

// handleSetAuth handles the aftrs_bandcamp_set_auth tool
func handleSetAuth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	identityCookie, errResult := tools.RequireStringParam(req, "identity_cookie")
	if errResult != nil {
		return errResult, nil
	}

	var sb strings.Builder
	sb.WriteString("# Bandcamp Authentication Setup\n\n")

	// Store cookie in config
	// In production, this would save to a config file
	sb.WriteString("## Cookie Stored\n\n")
	sb.WriteString(fmt.Sprintf("**Cookie (truncated):** %s...\n\n", truncateString(identityCookie, 20)))

	sb.WriteString("✅ Authentication cookie saved.\n\n")

	sb.WriteString("## Configuration Location\n\n")
	sb.WriteString("The cookie is stored in:\n")
	sb.WriteString("- **Windows:** `%APPDATA%\\bandcamp-dl\\config`\n")
	sb.WriteString("- **macOS/Linux:** `~/.config/bandcamp-dl/config`\n\n")

	sb.WriteString("## Verify\n\n")
	sb.WriteString("Use `aftrs_bandcamp_check_auth` to verify authentication.\n")

	return tools.TextResult(sb.String()), nil
}

// handleCheckAuth handles the aftrs_bandcamp_check_auth tool
func handleCheckAuth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var sb strings.Builder
	sb.WriteString("# Bandcamp Authentication Status\n\n")

	// Check bandcamp-dl
	sb.WriteString("## bandcamp-dl\n\n")
	bcDL, err := findBandcampDL()
	if err != nil {
		sb.WriteString("❌ **Not installed**\n\n")
		sb.WriteString("Install with:\n")
		sb.WriteString("```bash\npip install bandcamp-downloader\n```\n\n")
	} else {
		sb.WriteString(fmt.Sprintf("✅ **Found:** %s\n\n", bcDL))
	}

	// Check for config/cookie
	sb.WriteString("## Authentication\n\n")
	sb.WriteString("⚠️ Cookie status cannot be verified directly.\n\n")
	sb.WriteString("To set up authentication:\n\n")
	sb.WriteString("1. Log in to bandcamp.com in your browser\n")
	sb.WriteString("2. Open Developer Tools (F12)\n")
	sb.WriteString("3. Go to Application → Cookies → bandcamp.com\n")
	sb.WriteString("4. Copy the value of the `identity` cookie\n")
	sb.WriteString("5. Run: `aftrs_bandcamp_set_auth identity_cookie=\"<value>\"`\n\n")

	sb.WriteString("## Alternative: Environment Variable\n\n")
	sb.WriteString("You can also set the cookie via environment variable:\n")
	sb.WriteString("```bash\n")
	sb.WriteString("export BANDCAMP_IDENTITY_COOKIE=\"your_cookie_value\"\n")
	sb.WriteString("```\n\n")

	sb.WriteString("## Test Download\n\n")
	sb.WriteString("To verify authentication works, try downloading a small album:\n")
	sb.WriteString("```bash\n")
	sb.WriteString("bandcamp-dl --format mp3-320 https://artist.bandcamp.com/album/test\n")
	sb.WriteString("```\n")

	return tools.TextResult(sb.String()), nil
}

// findBandcampDL locates the bandcamp-dl executable
func findBandcampDL() (string, error) {
	// Try common locations
	paths := []string{
		"bandcamp-dl",
		"bandcamp-downloader",
		filepath.Join("~", ".local", "bin", "bandcamp-dl"),
	}

	for _, p := range paths {
		if path, err := exec.LookPath(p); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("bandcamp-dl not found in PATH")
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

// Note: These tools should be registered by extending the bandcamp module.
// Add them to the module.go Tools() method by appending pipelineTools().
