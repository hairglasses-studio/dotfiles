package cr8

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var cr8MigrationsPath = func() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("Docs", "cr8-cli", "cr8-cli", "lib", "migrations")
	}
	return filepath.Join(home, "Docs", "cr8-cli", "cr8-cli", "lib", "migrations")
}()

const (
	awsProfile = "cr8"
	s3Bucket   = "cr8-music-storage"
)

// runPythonScript executes a Python script and returns its output
func runPythonScript(ctx context.Context, script string, args ...string) (string, error) {
	scriptPath := filepath.Join(cr8MigrationsPath, script)
	cmdArgs := append([]string{"-u", scriptPath}, args...)
	cmd := exec.CommandContext(ctx, "python", cmdArgs...)
	cmd.Env = append(os.Environ(), "AWS_PROFILE="+awsProfile)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("script error: %v\noutput: %s", err, output)
	}
	return string(output), nil
}

// handleMigrationStatus returns comprehensive migration status
func handleMigrationStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	output, err := runPythonScript(ctx, "migration_status.py")
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(output), nil
}

// handleVerifySync verifies S3 and DynamoDB consistency
func handleVerifySync(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	output, err := runPythonScript(ctx, "verify_sync.py")
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(output), nil
}

// handleAnalysisStats returns audio analysis coverage statistics
func handleAnalysisStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	output, err := runPythonScript(ctx, "analysis_trigger.py", "--stats")
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(output), nil
}

// handleQueueAnalysis queues unanalyzed tracks for processing
func handleQueueAnalysis(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{"--queue"}

	limit := tools.GetIntParam(request, "limit", 0)
	if limit > 0 {
		args = append(args, "--limit", strconv.Itoa(limit))
	}

	dryRun := tools.GetBoolParam(request, "dry_run", false)
	if dryRun {
		args = append(args, "--dry-run")
	}

	output, err := runPythonScript(ctx, "analysis_trigger.py", args...)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(output), nil
}

// handleUpdatePaths updates DynamoDB track paths to new structure
func handleUpdatePaths(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{}

	// Default to dry-run for safety
	dryRun := tools.GetBoolParam(request, "dry_run", true)
	if dryRun {
		args = append(args, "--dry-run")
	}

	verbose := tools.GetBoolParam(request, "verbose", false)
	if verbose {
		args = append(args, "--verbose")
	}

	output, err := runPythonScript(ctx, "batch_update_paths.py", args...)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(output), nil
}

// handleImportS3Tracks imports orphaned S3 files into DynamoDB
func handleImportS3Tracks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{}

	// Default to dry-run for safety
	dryRun := tools.GetBoolParam(request, "dry_run", true)
	if dryRun {
		args = append(args, "--dry-run")
	}

	limit := tools.GetIntParam(request, "limit", 0)
	if limit > 0 {
		args = append(args, "--limit", strconv.Itoa(limit))
	}

	output, err := runPythonScript(ctx, "import_s3_tracks.py", args...)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(output), nil
}

// handleS3Structure returns S3 bucket structure
func handleS3Structure(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cmd := exec.CommandContext(ctx, "aws", "s3", "ls", "s3://"+s3Bucket+"/", "--profile", awsProfile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to list S3: %v", err)), nil
	}

	// Get summary
	summaryCmd := exec.CommandContext(ctx, "aws", "s3", "ls", "s3://"+s3Bucket, "--recursive", "--summarize", "--profile", awsProfile)
	summaryOutput, _ := summaryCmd.CombinedOutput()

	// Extract summary lines
	lines := strings.Split(string(summaryOutput), "\n")
	var summary string
	for _, line := range lines {
		if strings.Contains(line, "Total") {
			summary += line + "\n"
		}
	}

	result := fmt.Sprintf("S3 Bucket: %s\n\nPrefixes:\n%s\nSummary:\n%s", s3Bucket, string(output), summary)
	return tools.TextResult(result), nil
}

// handleSyncMusic syncs local music to S3
func handleSyncMusic(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	source, errResult := tools.RequireStringParam(request, "source")
	if errResult != nil {
		return errResult, nil
	}

	// Map source names to paths
	var sourcePath, destPrefix string
	switch strings.ToLower(source) {
	case "music":
		sourcePath = os.ExpandEnv("$HOME/Music")
		destPrefix = "music-library/"
	case "gdrive":
		sourcePath = os.ExpandEnv("$HOME/Google Drive/My Drive/DJ Crates")
		destPrefix = "gdrive/"
	default:
		sourcePath = source
		destPrefix = "uploads/"
	}

	// Build rclone command
	args := []string{
		"sync", sourcePath,
		fmt.Sprintf("cr8:%s/%s", s3Bucket, destPrefix),
		"--max-size", "25M",
		"--include", "*.m4a",
		"--include", "*.mp3",
		"--include", "*.wav",
		"--include", "*.flac",
		"--progress",
	}

	// Default to dry-run
	dryRun := tools.GetBoolParam(request, "dry_run", true)
	if dryRun {
		args = append(args, "--dry-run")
	}

	cmd := exec.CommandContext(ctx, "rclone", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("sync failed: %v\n%s", err, output)), nil
	}

	result := fmt.Sprintf("Sync from %s to s3://%s/%s\n\n%s", sourcePath, s3Bucket, destPrefix, string(output))
	if dryRun {
		result = "[DRY RUN]\n" + result
	}
	return tools.TextResult(result), nil
}

// handleS3Reorganize reorganizes S3 bucket structure
func handleS3Reorganize(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{}

	dryRun := tools.GetBoolParam(request, "dry_run", true)
	if dryRun {
		args = append(args, "--dry-run")
	}

	output, err := runPythonScript(ctx, "s3_reorganize.py", args...)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(output), nil
}

// handleVerifyMigration verifies full migration integrity
func handleVerifyMigration(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	output, err := runPythonScript(ctx, "verify_migration.py")
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(output), nil
}

// handleDynamoDBImport imports JSON data into DynamoDB
func handleDynamoDBImport(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	table, errResult := tools.RequireStringParam(request, "table")
	if errResult != nil {
		return errResult, nil
	}

	args := []string{"--table", table}

	jsonFile := tools.GetStringParam(request, "json_file")
	if jsonFile != "" {
		args = append(args, "--file", jsonFile)
	}

	dryRun := tools.GetBoolParam(request, "dry_run", true)
	if dryRun {
		args = append(args, "--dry-run")
	}

	output, err := runPythonScript(ctx, "dynamodb_import.py", args...)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(output), nil
}

// handleSupabaseExport exports data from Supabase to JSON
func handleSupabaseExport(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{}

	table := tools.GetStringParam(request, "table")
	if table != "" && table != "all" {
		args = append(args, "--table", table)
	}

	output, err := runPythonScript(ctx, "supabase_export.py", args...)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(output), nil
}

// handleRunAnalysis runs the AWS analysis worker to process queued tracks
func handleRunAnalysis(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{}

	batch := tools.GetIntParam(request, "batch", 10)
	args = append(args, "--batch", strconv.Itoa(batch))

	dryRun := tools.GetBoolParam(request, "dry_run", false)
	if dryRun {
		args = append(args, "--dry-run")
	}

	output, err := runPythonScript(ctx, "analysis_worker.py", args...)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(output), nil
}

// handleCR8Status returns consolidated CR8 system status
func handleCR8Status(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	output, err := runPythonScript(ctx, "cr8_status.py")
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(output), nil
}

// handleQueueStatus returns detailed queue status
func handleQueueStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{}

	if tools.GetBoolParam(request, "failures", false) {
		args = append(args, "--failures")
	}

	output, err := runPythonScript(ctx, "queue_status.py", args...)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(output), nil
}

// handleSyncLikes syncs new tracks from playlists to S3
func handleSyncLikes(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{}

	playlistID := tools.GetIntParam(request, "playlist_id", 0)
	if playlistID > 0 {
		args = append(args, "--playlist-id", strconv.Itoa(playlistID))
	}

	username := tools.GetStringParam(request, "username")
	if username != "" {
		args = append(args, "--username", username)
	}

	service := tools.GetStringParam(request, "service")
	if service != "" {
		args = append(args, "--service", service)
	}

	limit := tools.GetIntParam(request, "limit", 0)
	if limit > 0 {
		args = append(args, "--limit", strconv.Itoa(limit))
	}

	if tools.GetBoolParam(request, "dry_run", false) {
		args = append(args, "--dry-run")
	}

	output, err := runPythonScript(ctx, "sync_playlist_likes.py", args...)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(output), nil
}

// handleSyncRekordbox syncs CR8 playlists to Rekordbox
func handleSyncRekordbox(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{"--json"} // Use JSON output for MCP integration

	playlistID := tools.GetIntParam(request, "playlist_id", 0)
	if playlistID > 0 {
		args = append(args, "--playlist-id", strconv.Itoa(playlistID))
	}

	// Default to dry-run for safety
	dryRun := tools.GetBoolParam(request, "dry_run", true)
	if dryRun {
		args = append(args, "--dry-run")
	}

	download := tools.GetBoolParam(request, "download", false)
	if download {
		args = append(args, "--download")
	}

	syncLocal := tools.GetBoolParam(request, "sync_local", false)
	if syncLocal {
		args = append(args, "--sync-local")
	}

	output, err := runPythonScript(ctx, "sync_to_rekordbox.py", args...)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(output), nil
}

// handleMigratePaths migrates tracks from flat paths to full DJ Crates structure
func handleMigratePaths(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{}

	// Default to dry-run for safety
	dryRun := tools.GetBoolParam(request, "dry_run", true)
	if dryRun {
		args = append(args, "--dry-run")
	} else {
		args = append(args, "--execute")
	}

	limit := tools.GetIntParam(request, "limit", 0)
	if limit > 0 {
		args = append(args, "--limit", strconv.Itoa(limit))
	}

	verbose := tools.GetBoolParam(request, "verbose", false)
	if verbose {
		args = append(args, "--verbose")
	}

	output, err := runPythonScript(ctx, "migrate_s3_paths.py", args...)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(output), nil
}

// handleRekordboxStatus returns Rekordbox library and sync status
func handleRekordboxStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	output, err := runPythonScript(ctx, "rekordbox_status.py")
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(output), nil
}

// handleS3ToLocal syncs S3 music files to local folder using rclone
func handleS3ToLocal(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dryRun := tools.GetBoolParam(request, "dry_run", true)
	prefix := tools.GetStringParam(request, "prefix")
	if prefix == "" {
		prefix = "downloads"
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("resolve home: %w", err)), nil
	}
	localRoot := filepath.Join(home, "Music", "CR8")
	localPath := localRoot + "/" + prefix
	s3Path := fmt.Sprintf("cr8-s3:%s/%s", s3Bucket, prefix)

	args := []string{
		"sync",
		s3Path,
		localPath,
		"--include", "*.m4a",
		"--include", "*.mp3",
		"--include", "*.wav",
		"--include", "*.flac",
		"--progress",
		"--stats-one-line",
	}

	if dryRun {
		args = append(args, "--dry-run")
	}

	cmd := exec.CommandContext(ctx, "rclone", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("rclone error: %v\noutput: %s", err, output)), nil
	}

	result := fmt.Sprintf("S3 to Local Sync (%s)\n", prefix)
	if dryRun {
		result += "[DRY RUN]\n"
	}
	result += fmt.Sprintf("Source: %s\n", s3Path)
	result += fmt.Sprintf("Target: %s\n\n", localPath)
	result += string(output)

	return tools.TextResult(result), nil
}

// handleSyncStatus shows sync state and history
func handleSyncStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{}

	playlistID := tools.GetIntParam(request, "playlist_id", 0)
	if playlistID > 0 {
		args = append(args, "--playlist-id", strconv.Itoa(playlistID))
	}

	if tools.GetBoolParam(request, "logs", false) {
		args = append(args, "--logs")
	}

	if tools.GetBoolParam(request, "json", false) {
		args = append(args, "--json")
	}

	reset := tools.GetIntParam(request, "reset", 0)
	if reset > 0 {
		args = append(args, "--reset", strconv.Itoa(reset))
	}

	output, err := runPythonScript(ctx, "sync_status.py", args...)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(output), nil
}

// handlePlaylistList lists playlists with minimal output
func handlePlaylistList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{}

	service := tools.GetStringParam(request, "service")
	if service != "" {
		args = append(args, "--service", service)
	}

	if tools.GetBoolParam(request, "all", false) {
		args = append(args, "--all")
	}

	if tools.GetBoolParam(request, "json", false) {
		args = append(args, "--json")
	}

	output, err := runPythonScript(ctx, "playlist_list.py", args...)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(output), nil
}

// handleTrackSearch searches tracks by various criteria
func handleTrackSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	args := []string{}

	query := tools.GetStringParam(request, "query")
	if query != "" {
		args = append(args, query)
	}

	bpm := tools.GetStringParam(request, "bpm")
	if bpm != "" {
		args = append(args, "--bpm", bpm)
	}

	key := tools.GetStringParam(request, "key")
	if key != "" {
		args = append(args, "--key", key)
	}

	camelot := tools.GetStringParam(request, "camelot")
	if camelot != "" {
		args = append(args, "--camelot", camelot)
	}

	if tools.GetBoolParam(request, "unanalyzed", false) {
		args = append(args, "--unanalyzed")
	}

	limit := tools.GetIntParam(request, "limit", 50)
	args = append(args, "--limit", strconv.Itoa(limit))

	output, err := runPythonScript(ctx, "track_search.py", args...)
	if err != nil {
		return tools.ErrorResult(err), nil
	}
	return tools.TextResult(output), nil
}
