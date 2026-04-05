// Package rclone provides MCP tools for cloud storage sync and backup using rclone.
package rclone

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/hairglasses-studio/mcpkit/sanitize"
	"github.com/mark3labs/mcp-go/mcp"
)

// Module implements the rclone tools module
type Module struct{}

var getClient = tools.LazyClient(clients.NewRcloneClient)

// Name returns the module name
func (m *Module) Name() string {
	return "rclone"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Cloud storage sync and backup using rclone (Google Drive, S3, etc.)"
}

// Tools returns all tool definitions for this module
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		// Remote management
		{
			Tool: mcp.NewTool("aftrs_rclone_list_remotes",
				mcp.WithDescription("List all configured rclone remotes (cloud storage connections)."),
			),
			Handler:             handleListRemotes,
			Category:            "rclone",
			Subcategory:         "remotes",
			Tags:                []string{"rclone", "remotes", "cloud", "storage", "gdrive"},
			UseCases:            []string{"List cloud storage connections", "Check configured remotes"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rclone",
			IsWrite:             false,
		},
		{
			Tool: mcp.NewTool("aftrs_rclone_remote_info",
				mcp.WithDescription("Get information about a specific rclone remote including storage usage."),
				mcp.WithString("remote",
					mcp.Required(),
					mcp.Description("Name of the remote (e.g., 'gdrive' or 's3')"),
				),
			),
			Handler:             handleRemoteInfo,
			Category:            "rclone",
			Subcategory:         "remotes",
			Tags:                []string{"rclone", "remote", "info", "storage", "space"},
			UseCases:            []string{"Check cloud storage usage", "Verify remote connection"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rclone",
			IsWrite:             false,
		},
		// Inventory and comparison
		{
			Tool: mcp.NewTool("aftrs_rclone_inventory_drive",
				mcp.WithDescription("Create an inventory of a local drive, listing all folders with sizes and file counts."),
				mcp.WithString("drive_path",
					mcp.Required(),
					mcp.Description("Path to the drive (e.g., 'D:\\' or '/mnt/backup')"),
				),
			),
			Handler:             handleInventoryDrive,
			Category:            "rclone",
			Subcategory:         "inventory",
			Tags:                []string{"rclone", "inventory", "drive", "local", "scan"},
			UseCases:            []string{"Inventory external drive", "Scan local storage", "Pre-backup analysis"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "rclone",
			IsWrite:             false,
		},
		{
			Tool: mcp.NewTool("aftrs_rclone_list_folders",
				mcp.WithDescription("List folders in a local or remote path."),
				mcp.WithString("path",
					mcp.Required(),
					mcp.Description("Path to list (local path or remote:path format like 'gdrive:backups')"),
				),
			),
			Handler:             handleListFolders,
			Category:            "rclone",
			Subcategory:         "browse",
			Tags:                []string{"rclone", "list", "folders", "browse"},
			UseCases:            []string{"Browse cloud storage", "List remote folders"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rclone",
			IsWrite:             false,
		},
		{
			Tool: mcp.NewTool("aftrs_rclone_compare",
				mcp.WithDescription("Compare local path with remote path to find differences. Shows files that exist only locally, only remotely, or have differences."),
				mcp.WithString("local_path",
					mcp.Required(),
					mcp.Description("Local path to compare"),
				),
				mcp.WithString("remote_path",
					mcp.Required(),
					mcp.Description("Remote path to compare (e.g., 'gdrive:backups/myfiles')"),
				),
			),
			Handler:             handleCompare,
			Category:            "rclone",
			Subcategory:         "compare",
			Tags:                []string{"rclone", "compare", "diff", "sync", "check"},
			UseCases:            []string{"Pre-sync analysis", "Find files to upload", "Compare local vs cloud"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "rclone",
			IsWrite:             false,
		},
		// Sync operations
		{
			Tool: mcp.NewTool("aftrs_rclone_sync_start",
				mcp.WithDescription("Start a sync job to copy/sync files from source to destination. Uses multiple parallel transfers for speed. Returns job ID for monitoring."),
				mcp.WithString("source",
					mcp.Required(),
					mcp.Description("Source path (local or remote:path)"),
				),
				mcp.WithString("destination",
					mcp.Required(),
					mcp.Description("Destination path (local or remote:path)"),
				),
				mcp.WithBoolean("dry_run",
					mcp.Description("If true, show what would be done without actually doing it (default: false)"),
				),
				mcp.WithBoolean("delete_extra",
					mcp.Description("If true, delete files on destination that don't exist on source (default: false)"),
				),
				mcp.WithNumber("transfers",
					mcp.Description("Number of parallel file transfers (default: 8, max: 32)"),
				),
				mcp.WithString("exclude",
					mcp.Description("Comma-separated list of patterns to exclude (e.g., '*.tmp,node_modules/**')"),
				),
			),
			Handler:             handleSyncStart,
			Category:            "rclone",
			Subcategory:         "sync",
			Tags:                []string{"rclone", "sync", "backup", "upload", "transfer"},
			UseCases:            []string{"Backup to cloud", "Sync folders", "Upload files"},
			Complexity:          tools.ComplexityComplex,
			CircuitBreakerGroup: "rclone",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_rclone_sync_status",
				mcp.WithDescription("Get the status and progress of a sync job. Shows transfer speed, ETA, files completed, and errors."),
				mcp.WithString("job_id",
					mcp.Description("Job ID to check (if not provided, shows all active jobs)"),
				),
			),
			Handler:             handleSyncStatus,
			Category:            "rclone",
			Subcategory:         "sync",
			Tags:                []string{"rclone", "sync", "status", "progress", "monitor"},
			UseCases:            []string{"Monitor sync progress", "Check transfer status", "Estimate completion time"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rclone",
			IsWrite:             false,
		},
		{
			Tool: mcp.NewTool("aftrs_rclone_sync_cancel",
				mcp.WithDescription("Cancel a running sync job."),
				mcp.WithString("job_id",
					mcp.Required(),
					mcp.Description("Job ID to cancel"),
				),
			),
			Handler:             handleSyncCancel,
			Category:            "rclone",
			Subcategory:         "sync",
			Tags:                []string{"rclone", "sync", "cancel", "stop"},
			UseCases:            []string{"Cancel sync", "Stop transfer"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rclone",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_rclone_sync_list_jobs",
				mcp.WithDescription("List all sync jobs (active and completed)."),
			),
			Handler:             handleListJobs,
			Category:            "rclone",
			Subcategory:         "sync",
			Tags:                []string{"rclone", "sync", "jobs", "list"},
			UseCases:            []string{"View sync history", "List active syncs"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rclone",
			IsWrite:             false,
		},
		// Configuration
		{
			Tool: mcp.NewTool("aftrs_rclone_config",
				mcp.WithDescription("Configure rclone sync settings like number of parallel transfers and bandwidth limits."),
				mcp.WithNumber("transfers",
					mcp.Description("Number of parallel file transfers (1-32)"),
				),
				mcp.WithNumber("checkers",
					mcp.Description("Number of parallel file checkers (1-64)"),
				),
				mcp.WithString("bandwidth",
					mcp.Description("Bandwidth limit (e.g., '10M' for 10MB/s, '100M' for 100MB/s, empty for unlimited)"),
				),
				mcp.WithString("drive_chunk_size",
					mcp.Description("Chunk size for Google Drive uploads (e.g., '8M', '16M', '64M'). Larger = faster for big files."),
				),
				mcp.WithBoolean("fast_list",
					mcp.Description("Enable --fast-list for 20x faster directory listing (recommended for Google Drive)"),
				),
			),
			Handler:             handleConfig,
			Category:            "rclone",
			Subcategory:         "config",
			Tags:                []string{"rclone", "config", "settings", "bandwidth", "transfers"},
			UseCases:            []string{"Adjust sync speed", "Limit bandwidth", "Configure workers"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rclone",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_rclone_apply_profile",
				mcp.WithDescription("Apply a preset sync profile optimized for different backup scenarios."),
				mcp.WithString("profile",
					mcp.Required(),
					mcp.Description("Profile name: 'large_files' (VJ clips, videos, movies), 'many_small_files' (documents, code), 'background' (low-impact), 'default' (balanced)"),
				),
			),
			Handler:             handleApplyProfile,
			Category:            "rclone",
			Subcategory:         "config",
			Tags:                []string{"rclone", "profile", "preset", "optimize"},
			UseCases:            []string{"Optimize for video backup", "Configure for documents", "Low-impact background sync"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rclone",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_rclone_show_config",
				mcp.WithDescription("Show current rclone configuration including Google Drive best practices info."),
			),
			Handler:             handleShowConfig,
			Category:            "rclone",
			Subcategory:         "config",
			Tags:                []string{"rclone", "config", "show", "status"},
			UseCases:            []string{"View current settings", "Check configuration"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rclone",
			IsWrite:             false,
		},
		{
			Tool: mcp.NewTool("aftrs_rclone_estimate_time",
				mcp.WithDescription("Estimate upload/sync time based on data size and Google Drive typical speeds."),
				mcp.WithString("path",
					mcp.Required(),
					mcp.Description("Path to check size (local or remote:path)"),
				),
			),
			Handler:             handleEstimateTime,
			Category:            "rclone",
			Subcategory:         "estimate",
			Tags:                []string{"rclone", "estimate", "time", "upload"},
			UseCases:            []string{"Plan backup schedule", "Estimate completion time"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rclone",
			IsWrite:             false,
		},
		// Bandwidth scheduling
		{
			Tool: mcp.NewTool("aftrs_rclone_set_schedule",
				mcp.WithDescription("Configure time-based bandwidth scheduling. Set different bandwidth limits for different times of day."),
				mcp.WithString("schedules",
					mcp.Required(),
					mcp.Description("JSON array of schedules: [{\"start_hour\": 8, \"end_hour\": 18, \"limit\": \"20M\"}, {\"start_hour\": 18, \"end_hour\": 8, \"limit\": \"off\"}]"),
				),
			),
			Handler:             handleSetSchedule,
			Category:            "rclone",
			Subcategory:         "config",
			Tags:                []string{"rclone", "schedule", "bandwidth", "throttle"},
			UseCases:            []string{"Schedule low-impact backups", "Time-based throttling"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "rclone",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_rclone_get_schedule",
				mcp.WithDescription("Show the current bandwidth schedule and active limit."),
			),
			Handler:             handleGetSchedule,
			Category:            "rclone",
			Subcategory:         "config",
			Tags:                []string{"rclone", "schedule", "bandwidth", "status"},
			UseCases:            []string{"Check bandwidth schedule", "View current throttle"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rclone",
			IsWrite:             false,
		},
		// Quota tracking
		{
			Tool: mcp.NewTool("aftrs_rclone_quota_status",
				mcp.WithDescription("Check Google Drive daily quota usage (750 GiB limit). Shows amount uploaded today and remaining quota."),
			),
			Handler:             handleQuotaStatus,
			Category:            "rclone",
			Subcategory:         "quota",
			Tags:                []string{"rclone", "quota", "gdrive", "limit"},
			UseCases:            []string{"Check upload quota", "Monitor daily limit"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rclone",
			IsWrite:             false,
		},
		// Checkpoint/Resume
		{
			Tool: mcp.NewTool("aftrs_rclone_list_checkpoints",
				mcp.WithDescription("List all saved sync checkpoints that can be resumed."),
			),
			Handler:             handleListCheckpoints,
			Category:            "rclone",
			Subcategory:         "checkpoint",
			Tags:                []string{"rclone", "checkpoint", "resume", "recover"},
			UseCases:            []string{"List resumable jobs", "View saved checkpoints"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rclone",
			IsWrite:             false,
		},
		{
			Tool: mcp.NewTool("aftrs_rclone_resume_checkpoint",
				mcp.WithDescription("Resume a sync from a saved checkpoint."),
				mcp.WithString("job_id",
					mcp.Required(),
					mcp.Description("Job ID from checkpoint to resume"),
				),
			),
			Handler:             handleResumeCheckpoint,
			Category:            "rclone",
			Subcategory:         "checkpoint",
			Tags:                []string{"rclone", "resume", "checkpoint", "continue"},
			UseCases:            []string{"Resume interrupted sync", "Continue from checkpoint"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "rclone",
			IsWrite:             true,
		},
		// Failed file retry
		{
			Tool: mcp.NewTool("aftrs_rclone_retry_failed",
				mcp.WithDescription("Retry only the failed files from a previous sync job."),
				mcp.WithString("job_id",
					mcp.Required(),
					mcp.Description("Job ID to retry failed files from"),
				),
			),
			Handler:             handleRetryFailed,
			Category:            "rclone",
			Subcategory:         "sync",
			Tags:                []string{"rclone", "retry", "failed", "recover"},
			UseCases:            []string{"Retry failed uploads", "Recover from errors"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "rclone",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_rclone_list_failed",
				mcp.WithDescription("List files that failed during a sync job."),
				mcp.WithString("job_id",
					mcp.Required(),
					mcp.Description("Job ID to check for failed files"),
				),
			),
			Handler:             handleListFailed,
			Category:            "rclone",
			Subcategory:         "sync",
			Tags:                []string{"rclone", "failed", "errors", "list"},
			UseCases:            []string{"View failed transfers", "Diagnose sync issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rclone",
			IsWrite:             false,
		},
		// Phase 2: Comparison Mode Tools
		{
			Tool: mcp.NewTool("aftrs_rclone_set_compare_mode",
				mcp.WithDescription("Set the file comparison mode for sync operations. Controls how files are compared to determine if they need transfer."),
				mcp.WithString("mode",
					mcp.Required(),
					mcp.Description("Comparison mode: 'size' (fastest, size only), 'default' (balanced, size+modtime), 'checksum' (most accurate, hash comparison)"),
				),
				mcp.WithBoolean("skip_existing",
					mcp.Description("If true, skip files that already exist on destination (--ignore-existing)"),
				),
				mcp.WithBoolean("update_only",
					mcp.Description("If true, skip files that are newer on destination (--update)"),
				),
			),
			Handler:             handleSetCompareMode,
			Category:            "rclone",
			Subcategory:         "config",
			Tags:                []string{"rclone", "compare", "dedup", "mode", "checksum"},
			UseCases:            []string{"Enable deduplication", "Skip existing files", "Compare by checksum"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rclone",
			IsWrite:             true,
		},
		// Phase 2: Pre-Sync Analysis
		{
			Tool: mcp.NewTool("aftrs_rclone_analyze_sync",
				mcp.WithDescription("Analyze what would be transferred before actually syncing. Shows files to transfer, estimated time, and size breakdown by extension."),
				mcp.WithString("source",
					mcp.Required(),
					mcp.Description("Source path (local or remote:path)"),
				),
				mcp.WithString("destination",
					mcp.Required(),
					mcp.Description("Destination path (local or remote:path)"),
				),
			),
			Handler:             handleAnalyzeSync,
			Category:            "rclone",
			Subcategory:         "analyze",
			Tags:                []string{"rclone", "analyze", "preview", "dry-run", "estimate"},
			UseCases:            []string{"Pre-sync analysis", "Estimate transfer size", "Plan backup"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "rclone",
			IsWrite:             false,
		},
		// Phase 2: Post-Sync Verification
		{
			Tool: mcp.NewTool("aftrs_rclone_verify",
				mcp.WithDescription("Verify files after sync by comparing source and destination. Can use checksum for data integrity verification."),
				mcp.WithString("source",
					mcp.Required(),
					mcp.Description("Source path (local or remote:path)"),
				),
				mcp.WithString("destination",
					mcp.Required(),
					mcp.Description("Destination path (local or remote:path)"),
				),
				mcp.WithBoolean("checksum",
					mcp.Description("If true, compare by hash/checksum for maximum accuracy (slower)"),
				),
			),
			Handler:             handleVerify,
			Category:            "rclone",
			Subcategory:         "verify",
			Tags:                []string{"rclone", "verify", "check", "integrity", "checksum"},
			UseCases:            []string{"Verify backup integrity", "Check sync completeness", "Data validation"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "rclone",
			IsWrite:             false,
		},
		// Phase 2: Dashboard
		{
			Tool: mcp.NewTool("aftrs_rclone_dashboard",
				mcp.WithDescription("Get a comprehensive dashboard showing all active sync jobs, aggregate progress, speed metrics, and quota status."),
			),
			Handler:             handleDashboard,
			Category:            "rclone",
			Subcategory:         "monitor",
			Tags:                []string{"rclone", "dashboard", "monitor", "status", "overview"},
			UseCases:            []string{"Monitor all syncs", "View aggregate progress", "Check system status"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rclone",
			IsWrite:             false,
		},
		// Phase 2: Transfer History
		{
			Tool: mcp.NewTool("aftrs_rclone_history",
				mcp.WithDescription("View transfer history and statistics. Shows completed syncs, total bytes transferred, and average speeds."),
				mcp.WithNumber("days",
					mcp.Description("Number of days of history to show (default: 7)"),
				),
			),
			Handler:             handleHistory,
			Category:            "rclone",
			Subcategory:         "history",
			Tags:                []string{"rclone", "history", "stats", "analytics"},
			UseCases:            []string{"View transfer history", "Analyze sync patterns", "Track data moved"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rclone",
			IsWrite:             false,
		},
		// Phase 2: Find Duplicates
		{
			Tool: mcp.NewTool("aftrs_rclone_find_duplicates",
				mcp.WithDescription("Find duplicate files in a hash index. Shows groups of files with identical content and wasted space."),
				mcp.WithString("index_id",
					mcp.Required(),
					mcp.Description("Hash index ID to search for duplicates"),
				),
			),
			Handler:             handleFindDuplicates,
			Category:            "rclone",
			Subcategory:         "dedup",
			Tags:                []string{"rclone", "duplicates", "dedup", "cleanup"},
			UseCases:            []string{"Find duplicate files", "Identify wasted space", "Cleanup storage"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "rclone",
			IsWrite:             false,
		},

		// Phase 4: Verified Sync & Failure Analytics
		{
			Tool: mcp.NewTool("aftrs_rclone_sync_verified",
				mcp.WithDescription("Sync files with automatic post-sync verification. Optionally retries mismatched files."),
				mcp.WithString("source",
					mcp.Required(),
					mcp.Description("Source path (local or remote:path)"),
				),
				mcp.WithString("dest",
					mcp.Required(),
					mcp.Description("Destination path (local or remote:path)"),
				),
				mcp.WithBoolean("auto_verify",
					mcp.Description("Run verification after sync (default: true)"),
				),
				mcp.WithString("verify_mode",
					mcp.Description("Verification mode: 'size' (faster) or 'checksum' (more accurate, default: size)"),
				),
				mcp.WithBoolean("retry_mismatches",
					mcp.Description("Automatically retry files that fail verification (default: false)"),
				),
				mcp.WithNumber("max_retries",
					mcp.Description("Maximum retry attempts for mismatched files (default: 3)"),
				),
			),
			Handler:             handleSyncVerified,
			Category:            "rclone",
			Subcategory:         "sync",
			Tags:                []string{"rclone", "sync", "verify", "integrity"},
			UseCases:            []string{"Verified data transfer", "Ensure data integrity", "Reliable backup"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "rclone",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_rclone_failure_analysis",
				mcp.WithDescription("Analyze failed files from a sync job. Categorizes errors and provides recovery suggestions."),
				mcp.WithString("job_id",
					mcp.Required(),
					mcp.Description("Job ID to analyze failures for"),
				),
			),
			Handler:             handleFailureAnalysis,
			Category:            "rclone",
			Subcategory:         "monitoring",
			Tags:                []string{"rclone", "errors", "analysis", "recovery"},
			UseCases:            []string{"Diagnose sync failures", "Get recovery suggestions", "Understand error patterns"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "rclone",
			IsWrite:             false,
		},
	}
}

// Handler functions

func handleListRemotes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	remotes, err := client.ListRemotes(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(remotes) == 0 {
		return tools.TextResult("No rclone remotes configured. Use `rclone config` to add remotes."), nil
	}

	var sb strings.Builder
	sb.WriteString("## Configured Rclone Remotes\n\n")
	for _, remote := range remotes {
		sb.WriteString(fmt.Sprintf("- **%s**\n", remote))
	}
	sb.WriteString(fmt.Sprintf("\nTotal: %d remotes\n", len(remotes)))
	sb.WriteString("\nUse `aftrs_rclone_remote_info` to get details about a specific remote.")

	return tools.TextResult(sb.String()), nil
}

func handleRemoteInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	remote, errResult := tools.RequireStringParam(req, "remote")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	info, err := client.GetRemoteInfo(ctx, remote)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Remote: %s\n\n", info.Name))
	sb.WriteString(fmt.Sprintf("**Type:** %s\n", info.Type))
	sb.WriteString(fmt.Sprintf("**Connected:** %v\n\n", info.Connected))

	if info.Error != "" {
		sb.WriteString(fmt.Sprintf("**Error:** %s\n", info.Error))
	} else {
		sb.WriteString("### Storage Usage\n")
		sb.WriteString(fmt.Sprintf("- **Total:** %s\n", clients.FormatBytes(info.Total)))
		sb.WriteString(fmt.Sprintf("- **Used:** %s\n", clients.FormatBytes(info.Used)))
		sb.WriteString(fmt.Sprintf("- **Free:** %s\n", clients.FormatBytes(info.Free)))
		if info.Trashed > 0 {
			sb.WriteString(fmt.Sprintf("- **Trashed:** %s\n", clients.FormatBytes(info.Trashed)))
		}

		// Calculate percentage
		if info.Total > 0 {
			pct := float64(info.Used) / float64(info.Total) * 100
			sb.WriteString(fmt.Sprintf("\n**Usage:** %.1f%% used\n", pct))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleInventoryDrive(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	drivePath, errResult := tools.RequireStringParam(req, "drive_path")
	if errResult != nil {
		return errResult, nil
	}
	if err := sanitize.RclonePath(drivePath); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid drive_path: %w", err)), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	inventory, err := client.InventoryLocalDrive(ctx, drivePath)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Drive Inventory: %s\n\n", inventory.DrivePath))
	sb.WriteString(fmt.Sprintf("**Scanned:** %s\n", inventory.ScannedAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("**Scan Duration:** %s\n\n", inventory.ScanDuration))

	sb.WriteString("### Storage Summary\n")
	sb.WriteString(fmt.Sprintf("- **Total:** %s\n", clients.FormatBytes(inventory.TotalSize)))
	sb.WriteString(fmt.Sprintf("- **Used:** %s\n", clients.FormatBytes(inventory.UsedSize)))
	sb.WriteString(fmt.Sprintf("- **Free:** %s\n", clients.FormatBytes(inventory.FreeSize)))
	if inventory.TotalSize > 0 {
		pct := float64(inventory.UsedSize) / float64(inventory.TotalSize) * 100
		sb.WriteString(fmt.Sprintf("- **Usage:** %.1f%%\n", pct))
	}

	sb.WriteString("\n### Folders (sorted by size)\n\n")
	sb.WriteString("| Folder | Size | Files |\n")
	sb.WriteString("|--------|------|-------|\n")

	// Sort by size descending
	sort.Slice(inventory.Folders, func(i, j int) bool {
		return inventory.Folders[i].SizeBytes > inventory.Folders[j].SizeBytes
	})

	for _, folder := range inventory.Folders {
		sb.WriteString(fmt.Sprintf("| %s | %s | %d |\n",
			folder.Name,
			clients.FormatBytes(folder.SizeBytes),
			folder.FileCount))
	}

	if len(inventory.AccessErrors) > 0 {
		sb.WriteString(fmt.Sprintf("\n### Access Errors (%d)\n", len(inventory.AccessErrors)))
		for _, e := range inventory.AccessErrors[:min(5, len(inventory.AccessErrors))] {
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
		if len(inventory.AccessErrors) > 5 {
			sb.WriteString(fmt.Sprintf("- ... and %d more\n", len(inventory.AccessErrors)-5))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleListFolders(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, errResult := tools.RequireStringParam(req, "path")
	if errResult != nil {
		return errResult, nil
	}
	if err := sanitize.RclonePath(path); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid path: %w", err)), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	folders, err := client.ListFolders(ctx, path)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(folders) == 0 {
		return tools.TextResult(fmt.Sprintf("No folders found in: %s", path)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Folders in: %s\n\n", path))

	for _, folder := range folders {
		sb.WriteString(fmt.Sprintf("- **%s** (modified: %s)\n",
			folder.Name,
			folder.LastModified.Format("2006-01-02")))
	}
	sb.WriteString(fmt.Sprintf("\nTotal: %d folders\n", len(folders)))

	return tools.TextResult(sb.String()), nil
}

func handleCompare(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	localPath := tools.GetStringParam(req, "local_path")
	remotePath := tools.GetStringParam(req, "remote_path")

	if localPath == "" || remotePath == "" {
		return tools.ErrorResult(fmt.Errorf("both local_path and remote_path are required")), nil
	}
	if err := sanitize.RclonePath(localPath); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid local_path: %w", err)), nil
	}
	if err := sanitize.RclonePath(remotePath); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid remote_path: %w", err)), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.Compare(ctx, localPath, remotePath)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("## Comparison Results\n\n")
	sb.WriteString(fmt.Sprintf("**Local:** %s\n", result.LocalPath))
	sb.WriteString(fmt.Sprintf("**Remote:** %s\n\n", result.RemotePath))

	sb.WriteString("### Summary\n")
	sb.WriteString(fmt.Sprintf("- **Matching files:** %d\n", result.Matched))
	sb.WriteString(fmt.Sprintf("- **Local only (to upload):** %d\n", len(result.LocalOnly)))
	sb.WriteString(fmt.Sprintf("- **Remote only:** %d\n", len(result.RemoteOnly)))
	sb.WriteString(fmt.Sprintf("- **Different:** %d\n", len(result.Different)))

	sb.WriteString("\n### Size Analysis\n")
	sb.WriteString(fmt.Sprintf("- **Local size:** %s\n", clients.FormatBytes(result.TotalLocalSize)))
	sb.WriteString(fmt.Sprintf("- **Remote size:** %s\n", clients.FormatBytes(result.TotalRemoteSize)))
	sb.WriteString(fmt.Sprintf("- **To upload:** %s\n", clients.FormatBytes(result.SizeToUpload)))

	if result.EstimatedTime != "" {
		sb.WriteString(fmt.Sprintf("\n### Estimated Upload Time\n"))
		sb.WriteString(fmt.Sprintf("~%s (at 10 MB/s average)\n", result.EstimatedTime))
	}

	// Show sample files
	if len(result.LocalOnly) > 0 {
		sb.WriteString("\n### Files to Upload (sample)\n")
		for i, f := range result.LocalOnly {
			if i >= 10 {
				sb.WriteString(fmt.Sprintf("- ... and %d more\n", len(result.LocalOnly)-10))
				break
			}
			sb.WriteString(fmt.Sprintf("- %s\n", f))
		}
	}

	if len(result.Different) > 0 {
		sb.WriteString("\n### Different Files (sample)\n")
		for i, f := range result.Different {
			if i >= 10 {
				sb.WriteString(fmt.Sprintf("- ... and %d more\n", len(result.Different)-10))
				break
			}
			sb.WriteString(fmt.Sprintf("- %s\n", f))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleSyncStart(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	source := tools.GetStringParam(req, "source")
	destination := tools.GetStringParam(req, "destination")
	dryRun := tools.GetBoolParam(req, "dry_run", false)
	deleteExtra := tools.GetBoolParam(req, "delete_extra", false)
	transfers := tools.GetIntParam(req, "transfers", 0)
	excludeStr := tools.GetStringParam(req, "exclude")

	if source == "" || destination == "" {
		return tools.ErrorResult(fmt.Errorf("source and destination are required")), nil
	}
	if err := sanitize.RclonePath(source); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid source: %w", err)), nil
	}
	if err := sanitize.RclonePath(destination); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid destination: %w", err)), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Set transfers if specified
	if transfers > 0 {
		if transfers > 32 {
			transfers = 32
		}
		client.SetTransfers(transfers)
	}

	// Parse exclude patterns
	var exclude []string
	if excludeStr != "" {
		exclude = strings.Split(excludeStr, ",")
		for i := range exclude {
			exclude[i] = strings.TrimSpace(exclude[i])
		}
	}

	job, err := client.StartSync(ctx, source, destination, dryRun, deleteExtra, exclude)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("## Sync Job Started\n\n")
	sb.WriteString(fmt.Sprintf("**Job ID:** `%s`\n", job.ID))
	sb.WriteString(fmt.Sprintf("**Source:** %s\n", job.Source))
	sb.WriteString(fmt.Sprintf("**Destination:** %s\n", job.Destination))
	sb.WriteString(fmt.Sprintf("**Dry Run:** %v\n", job.DryRun))
	sb.WriteString(fmt.Sprintf("**Delete Extra:** %v\n", job.DeleteExtra))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n", job.Status))

	if len(job.Exclude) > 0 {
		sb.WriteString(fmt.Sprintf("**Excludes:** %s\n", strings.Join(job.Exclude, ", ")))
	}

	sb.WriteString("\nUse `aftrs_rclone_sync_status` with this job ID to monitor progress.")

	return tools.TextResult(sb.String()), nil
}

func handleSyncStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jobID := tools.GetStringParam(req, "job_id")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if jobID != "" {
		// Get specific job
		job, err := client.GetJobStatus(jobID)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(formatJobStatus(job)), nil
	}

	// List all jobs
	jobs := client.ListJobs()
	if len(jobs) == 0 {
		return tools.TextResult("No sync jobs found. Use `aftrs_rclone_sync_start` to start a sync."), nil
	}

	var sb strings.Builder
	sb.WriteString("## All Sync Jobs\n\n")

	for _, job := range jobs {
		sb.WriteString(formatJobStatus(job))
		sb.WriteString("\n---\n\n")
	}

	return tools.TextResult(sb.String()), nil
}

func handleSyncCancel(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jobID, errResult := tools.RequireStringParam(req, "job_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.CancelJob(jobID); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Sync job `%s` has been cancelled.", jobID)), nil
}

func handleListJobs(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	jobs := client.ListJobs()
	if len(jobs) == 0 {
		return tools.TextResult("No sync jobs found."), nil
	}

	var sb strings.Builder
	sb.WriteString("## Sync Jobs\n\n")
	sb.WriteString("| Job ID | Source | Destination | Status | Progress |\n")
	sb.WriteString("|--------|--------|-------------|--------|----------|\n")

	for _, job := range jobs {
		progress := "-"
		if job.Progress != nil && job.Progress.PercentComplete > 0 {
			progress = fmt.Sprintf("%.1f%%", job.Progress.PercentComplete)
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
			job.ID,
			truncate(job.Source, 20),
			truncate(job.Destination, 20),
			job.Status,
			progress))
	}

	return tools.TextResult(sb.String()), nil
}

func handleConfig(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	transfers := tools.GetIntParam(req, "transfers", 0)
	checkers := tools.GetIntParam(req, "checkers", 0)
	bandwidth := tools.GetStringParam(req, "bandwidth")
	driveChunkSize := tools.GetStringParam(req, "drive_chunk_size")
	fastList := tools.GetBoolParam(req, "fast_list", false)
	fastListProvided := tools.HasParam(req, "fast_list")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var changes []string

	if transfers > 0 {
		if transfers > 32 {
			transfers = 32
		}
		client.SetTransfers(transfers)
		changes = append(changes, fmt.Sprintf("Transfers set to %d", transfers))
	}

	if checkers > 0 {
		if checkers > 64 {
			checkers = 64
		}
		client.SetCheckers(checkers)
		changes = append(changes, fmt.Sprintf("Checkers set to %d", checkers))
	}

	if bandwidth != "" {
		client.SetBandwidth(bandwidth)
		changes = append(changes, fmt.Sprintf("Bandwidth limit set to %s", bandwidth))
	}

	if driveChunkSize != "" {
		client.SetDriveChunkSize(driveChunkSize)
		changes = append(changes, fmt.Sprintf("Google Drive chunk size set to %s", driveChunkSize))
	}

	if fastListProvided {
		client.SetFastList(fastList)
		changes = append(changes, fmt.Sprintf("Fast list mode: %v", fastList))
	}

	if len(changes) == 0 {
		return tools.TextResult("No configuration changes made. Specify transfers, checkers, bandwidth, drive_chunk_size, or fast_list to update."), nil
	}

	return tools.TextResult("## Configuration Updated\n\n" + strings.Join(changes, "\n")), nil
}

func handleApplyProfile(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	profileName, errResult := tools.RequireStringParam(req, "profile")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var profile clients.SyncProfile
	var description string

	switch profileName {
	case "large_files":
		profile = clients.ProfileLargeFiles
		description = "Optimized for VJ clips, movies, and large video files.\n- 4 parallel transfers\n- 64MB chunk size\n- 8 checkers"
	case "many_small_files":
		profile = clients.ProfileManySmallFiles
		description = "Optimized for documents, code, and many small files.\n- 16 parallel transfers\n- 8MB chunk size\n- 32 checkers"
	case "background":
		profile = clients.ProfileBackground
		description = "Low-impact mode for background syncing.\n- 2 parallel transfers\n- 5MB/s bandwidth limit\n- 4 checkers"
	case "default":
		profile = clients.ProfileDefault
		description = "Balanced settings for general use.\n- 8 parallel transfers\n- 16MB chunk size\n- 16 checkers"
	default:
		return tools.ErrorResult(fmt.Errorf("unknown profile: %s. Use 'large_files', 'many_small_files', 'background', or 'default'", profileName)), nil
	}

	client.ApplyProfile(profile)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Profile Applied: %s\n\n", profileName))
	sb.WriteString(description)
	sb.WriteString("\n\n### Google Drive Best Practices\n")
	sb.WriteString("- Daily upload limit: 750 GiB (auto-handled with --drive-stop-on-upload-limit)\n")
	sb.WriteString("- Fast listing enabled for 20x faster directory scans\n")
	sb.WriteString("- Automatic exclusion of Windows system files\n")

	return tools.TextResult(sb.String()), nil
}

func handleShowConfig(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	config := client.GetConfig()

	var sb strings.Builder
	sb.WriteString("## Current Rclone Configuration\n\n")
	sb.WriteString("### Transfer Settings\n")
	sb.WriteString(fmt.Sprintf("- **Parallel Transfers:** %v\n", config["transfers"]))
	sb.WriteString(fmt.Sprintf("- **Parallel Checkers:** %v\n", config["checkers"]))

	bw := config["bandwidth"]
	if bw == nil || bw == "" {
		bw = "unlimited"
	}
	sb.WriteString(fmt.Sprintf("- **Bandwidth Limit:** %v\n", bw))

	sb.WriteString("\n### Google Drive Settings\n")
	sb.WriteString(fmt.Sprintf("- **Chunk Size:** %v\n", config["drive_chunk_size"]))
	sb.WriteString(fmt.Sprintf("- **Fast List:** %v (20x faster directory listing)\n", config["fast_list"]))

	sb.WriteString("\n### Google Drive Limits\n")
	sb.WriteString("- **Daily Upload Limit:** 750 GiB\n")
	sb.WriteString("- **Daily Download Limit:** 10 TiB\n")
	sb.WriteString("- **Note:** Limits are per 24-hour rolling window\n")

	sb.WriteString("\n### Available Profiles\n")
	sb.WriteString("| Profile | Use Case | Transfers | Chunk Size |\n")
	sb.WriteString("|---------|----------|-----------|------------|\n")
	sb.WriteString("| large_files | VJ clips, movies, videos | 4 | 64M |\n")
	sb.WriteString("| many_small_files | Documents, code | 16 | 8M |\n")
	sb.WriteString("| background | Low-impact sync | 2 | 8M |\n")
	sb.WriteString("| default | Balanced | 8 | 16M |\n")

	return tools.TextResult(sb.String()), nil
}

func handleEstimateTime(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, errResult := tools.RequireStringParam(req, "path")
	if errResult != nil {
		return errResult, nil
	}
	if err := sanitize.RclonePath(path); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid path: %w", err)), nil
	}

	// Get size using rclone
	type sizeResult struct {
		Count int64 `json:"count"`
		Bytes int64 `json:"bytes"`
	}

	args := []string{"size", path, "--json"}
	cmd := exec.CommandContext(ctx, "rclone", args...)
	output, err := cmd.Output()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get size: %w", err)), nil
	}

	var size sizeResult
	if err := json.Unmarshal(output, &size); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to parse size: %w", err)), nil
	}

	// Estimate at different speeds
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Upload Time Estimate: %s\n\n", path))
	sb.WriteString(fmt.Sprintf("**Total Size:** %s (%d files)\n\n", clients.FormatBytes(size.Bytes), size.Count))

	sb.WriteString("### Estimated Upload Times\n")
	sb.WriteString("| Speed | Time | Notes |\n")
	sb.WriteString("|-------|------|-------|\n")

	// Conservative (10 MB/s)
	conservative := time.Duration(size.Bytes/(10*1024*1024)) * time.Second
	sb.WriteString(fmt.Sprintf("| 10 MB/s | %s | Conservative estimate |\n", formatDuration(conservative)))

	// Typical (25 MB/s)
	typical := time.Duration(size.Bytes/(25*1024*1024)) * time.Second
	sb.WriteString(fmt.Sprintf("| 25 MB/s | %s | Typical Google Drive |\n", formatDuration(typical)))

	// Fast (50 MB/s)
	fast := time.Duration(size.Bytes/(50*1024*1024)) * time.Second
	sb.WriteString(fmt.Sprintf("| 50 MB/s | %s | Fast connection |\n", formatDuration(fast)))

	// 750GB daily limit check
	dailyLimitBytes := int64(750 * 1024 * 1024 * 1024)
	if size.Bytes > dailyLimitBytes {
		days := (size.Bytes + dailyLimitBytes - 1) / dailyLimitBytes
		sb.WriteString(fmt.Sprintf("\n⚠️ **Note:** This exceeds Google Drive's 750 GiB daily limit.\n"))
		sb.WriteString(fmt.Sprintf("**Minimum days required:** %d days\n", days))
	}

	return tools.TextResult(sb.String()), nil
}

// Helper functions

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return d.Round(time.Second).String()
	} else if d < time.Hour {
		mins := int(d.Minutes())
		secs := int(d.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", mins, secs)
	} else if d < 24*time.Hour {
		hours := int(d.Hours())
		mins := int(d.Minutes()) % 60
		return fmt.Sprintf("%dh %dm", hours, mins)
	} else {
		days := int(d.Hours()) / 24
		hours := int(d.Hours()) % 24
		return fmt.Sprintf("%dd %dh", days, hours)
	}
}

func formatJobStatus(job *clients.SyncJob) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("### Job: %s\n\n", job.ID))
	sb.WriteString(fmt.Sprintf("**Source:** %s\n", job.Source))
	sb.WriteString(fmt.Sprintf("**Destination:** %s\n", job.Destination))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n", job.Status))
	sb.WriteString(fmt.Sprintf("**Started:** %s\n", job.StartTime.Format(time.RFC3339)))

	if job.EndTime != nil {
		sb.WriteString(fmt.Sprintf("**Ended:** %s\n", job.EndTime.Format(time.RFC3339)))
	}

	if job.DryRun {
		sb.WriteString("**Mode:** Dry Run\n")
	}

	if job.Error != "" {
		sb.WriteString(fmt.Sprintf("**Error:** %s\n", job.Error))
	}

	if job.Progress != nil {
		p := job.Progress
		sb.WriteString("\n**Progress:**\n")
		sb.WriteString(fmt.Sprintf("- Files: %d / %d\n", p.FilesTransferred, p.FilesTotal))
		sb.WriteString(fmt.Sprintf("- Bytes: %s / %s (%.1f%%)\n",
			clients.FormatBytes(p.BytesTransferred),
			clients.FormatBytes(p.BytesTotal),
			p.PercentComplete))
		sb.WriteString(fmt.Sprintf("- Speed: %s\n", p.TransferSpeed))
		sb.WriteString(fmt.Sprintf("- ETA: %s\n", p.ETA))
		sb.WriteString(fmt.Sprintf("- Elapsed: %s\n", p.ElapsedTime.Round(time.Second)))
		sb.WriteString(fmt.Sprintf("- Checks: %d\n", p.Checks))

		if p.Errors > 0 {
			sb.WriteString(fmt.Sprintf("- **Errors: %d**\n", p.Errors))
		}

		if p.CurrentFile != "" {
			sb.WriteString(fmt.Sprintf("- Current: %s\n", truncate(p.CurrentFile, 50)))
		}
	}

	return sb.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ==========================================
// New Handler Functions for Enhanced Features
// ==========================================

func handleSetSchedule(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	schedulesStr, errResult := tools.RequireStringParam(req, "schedules")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var schedules []clients.BandwidthSchedule
	if err := json.Unmarshal([]byte(schedulesStr), &schedules); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid schedule JSON: %w", err)), nil
	}

	// Validate schedules
	for i, s := range schedules {
		if s.StartHour < 0 || s.StartHour > 23 {
			return tools.ErrorResult(fmt.Errorf("schedule %d: start_hour must be 0-23", i)), nil
		}
		if s.EndHour < 0 || s.EndHour > 23 {
			return tools.ErrorResult(fmt.Errorf("schedule %d: end_hour must be 0-23", i)), nil
		}
		if s.Limit == "" {
			return tools.ErrorResult(fmt.Errorf("schedule %d: limit is required", i)), nil
		}
	}

	client.SetBandwidthSchedule(schedules)

	var sb strings.Builder
	sb.WriteString("## Bandwidth Schedule Updated\n\n")
	sb.WriteString("| Time Range | Limit |\n")
	sb.WriteString("|------------|-------|\n")
	for _, s := range schedules {
		sb.WriteString(fmt.Sprintf("| %02d:00 - %02d:00 | %s |\n", s.StartHour, s.EndHour, s.Limit))
	}

	currentLimit := client.GetCurrentBandwidthLimit()
	sb.WriteString(fmt.Sprintf("\n**Current active limit:** %s\n", currentLimit))

	return tools.TextResult(sb.String()), nil
}

func handleGetSchedule(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	schedules := client.GetBandwidthSchedule()

	var sb strings.Builder
	sb.WriteString("## Bandwidth Schedule\n\n")

	if len(schedules) == 0 {
		sb.WriteString("No bandwidth schedule configured. Using default bandwidth limit.\n")
	} else {
		sb.WriteString("| Time Range | Limit |\n")
		sb.WriteString("|------------|-------|\n")
		for _, s := range schedules {
			sb.WriteString(fmt.Sprintf("| %02d:00 - %02d:00 | %s |\n", s.StartHour, s.EndHour, s.Limit))
		}
	}

	currentLimit := client.GetCurrentBandwidthLimit()
	if currentLimit == "" {
		currentLimit = "unlimited"
	}
	sb.WriteString(fmt.Sprintf("\n**Current active limit:** %s\n", currentLimit))
	sb.WriteString(fmt.Sprintf("**Current hour:** %02d:00\n", time.Now().Hour()))

	return tools.TextResult(sb.String()), nil
}

func handleQuotaStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	quota, err := client.GetDailyQuotaUsage()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get quota: %w", err)), nil
	}

	dailyLimit := int64(750 * 1024 * 1024 * 1024)     // 750 GiB
	pauseThreshold := int64(650 * 1024 * 1024 * 1024) // 650 GiB (90%)
	usedPercent := float64(quota.BytesUploaded) / float64(dailyLimit) * 100
	remaining := dailyLimit - quota.BytesUploaded

	var sb strings.Builder
	sb.WriteString("## Google Drive Daily Quota Status\n\n")
	sb.WriteString(fmt.Sprintf("**Date:** %s\n", quota.Date))
	sb.WriteString(fmt.Sprintf("**Uploaded Today:** %s\n", clients.FormatBytes(quota.BytesUploaded)))
	sb.WriteString(fmt.Sprintf("**Daily Limit:** %s\n", clients.FormatBytes(dailyLimit)))
	sb.WriteString(fmt.Sprintf("**Remaining:** %s\n", clients.FormatBytes(remaining)))
	sb.WriteString(fmt.Sprintf("**Usage:** %.1f%%\n", usedPercent))
	sb.WriteString(fmt.Sprintf("**Last Updated:** %s\n\n", quota.LastUpdated.Format(time.RFC3339)))

	// Progress bar
	barWidth := 30
	filled := int(usedPercent / 100 * float64(barWidth))
	sb.WriteString("[")
	for i := 0; i < barWidth; i++ {
		if i < filled {
			sb.WriteString("█")
		} else {
			sb.WriteString("░")
		}
	}
	sb.WriteString(fmt.Sprintf("] %.1f%%\n\n", usedPercent))

	// Warning if approaching limit
	if quota.BytesUploaded >= pauseThreshold {
		sb.WriteString("⚠️ **Warning:** Approaching daily limit! Sync may be paused soon.\n")
	} else if usedPercent > 70 {
		sb.WriteString("⚡ Note: Over 70% of daily quota used.\n")
	}

	return tools.TextResult(sb.String()), nil
}

func handleListCheckpoints(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	checkpoints, err := client.ListCheckpoints()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to list checkpoints: %w", err)), nil
	}

	if len(checkpoints) == 0 {
		return tools.TextResult("No saved checkpoints found."), nil
	}

	var sb strings.Builder
	sb.WriteString("## Saved Checkpoints\n\n")
	sb.WriteString("| Job ID | Source | Destination | Last Checkpoint |\n")
	sb.WriteString("|--------|--------|-------------|----------------|\n")

	for _, jobID := range checkpoints {
		checkpoint, err := client.LoadCheckpoint(jobID)
		if err != nil {
			continue
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			jobID,
			truncate(checkpoint.Source, 25),
			truncate(checkpoint.Destination, 25),
			checkpoint.LastCheckpoint.Format("2006-01-02 15:04")))
	}

	sb.WriteString("\nUse `aftrs_rclone_resume_checkpoint` to resume a sync.")

	return tools.TextResult(sb.String()), nil
}

func handleResumeCheckpoint(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jobID, errResult := tools.RequireStringParam(req, "job_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	checkpoint, err := client.LoadCheckpoint(jobID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("checkpoint not found: %w", err)), nil
	}

	job, err := client.ResumeFromCheckpoint(ctx, checkpoint)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to resume: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("## Sync Resumed from Checkpoint\n\n")
	sb.WriteString(fmt.Sprintf("**Original Job:** %s\n", checkpoint.JobID))
	sb.WriteString(fmt.Sprintf("**New Job ID:** %s\n", job.ID))
	sb.WriteString(fmt.Sprintf("**Source:** %s\n", checkpoint.Source))
	sb.WriteString(fmt.Sprintf("**Destination:** %s\n", checkpoint.Destination))
	sb.WriteString(fmt.Sprintf("**Previous Progress:** %s uploaded\n", clients.FormatBytes(checkpoint.BytesUploaded)))
	sb.WriteString("\nUse `aftrs_rclone_sync_status` to monitor progress.")

	return tools.TextResult(sb.String()), nil
}

func handleRetryFailed(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jobID, errResult := tools.RequireStringParam(req, "job_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	job, err := client.RetryFailedFiles(ctx, jobID)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	failedFiles, _ := client.GetFailedFiles(jobID)

	var sb strings.Builder
	sb.WriteString("## Retrying Failed Files\n\n")
	sb.WriteString(fmt.Sprintf("**Original Job:** %s\n", jobID))
	sb.WriteString(fmt.Sprintf("**Retry Job ID:** %s\n", job.ID))
	sb.WriteString(fmt.Sprintf("**Files to Retry:** %d\n", len(failedFiles)))
	sb.WriteString("\nUse `aftrs_rclone_sync_status` with the retry job ID to monitor progress.")

	return tools.TextResult(sb.String()), nil
}

func handleListFailed(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jobID, errResult := tools.RequireStringParam(req, "job_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	failedFiles, err := client.GetFailedFiles(jobID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get failed files: %w", err)), nil
	}

	if len(failedFiles) == 0 {
		return tools.TextResult(fmt.Sprintf("No failed files recorded for job: %s", jobID)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Failed Files for Job: %s\n\n", jobID))
	sb.WriteString(fmt.Sprintf("**Total Failed:** %d files\n\n", len(failedFiles)))

	sb.WriteString("| File | Size | Error | Retries |\n")
	sb.WriteString("|------|------|-------|--------|\n")

	// Show up to 20 failed files
	shown := 0
	for _, f := range failedFiles {
		if shown >= 20 {
			sb.WriteString(fmt.Sprintf("\n... and %d more files\n", len(failedFiles)-20))
			break
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %d |\n",
			truncate(f.Path, 40),
			clients.FormatBytes(f.Size),
			truncate(f.Error, 30),
			f.RetryCount))
		shown++
	}

	sb.WriteString("\nUse `aftrs_rclone_retry_failed` to retry these files.")

	return tools.TextResult(sb.String()), nil
}

// ==========================================
// Phase 2 Handler Functions
// ==========================================

func handleSetCompareMode(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	mode, errResult := tools.RequireStringParam(req, "mode")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var compareMode clients.ComparisonMode
	switch mode {
	case "size":
		compareMode = clients.CompareSizeOnly
	case "default":
		compareMode = clients.CompareSizeModTime
	case "checksum":
		compareMode = clients.CompareChecksum
	default:
		return tools.ErrorResult(fmt.Errorf("invalid mode: %s. Use 'size', 'default', or 'checksum'", mode)), nil
	}

	client.SetCompareMode(compareMode)

	// Handle optional settings
	if tools.HasParam(req, "skip_existing") {
		skipExisting := tools.GetBoolParam(req, "skip_existing", false)
		client.SetSkipExisting(skipExisting)
	}

	if tools.HasParam(req, "update_only") {
		updateOnly := tools.GetBoolParam(req, "update_only", false)
		client.SetUpdateOnly(updateOnly)
	}

	var sb strings.Builder
	sb.WriteString("## Comparison Mode Updated\n\n")
	sb.WriteString(fmt.Sprintf("**Mode:** %s\n", mode))

	switch compareMode {
	case clients.CompareSizeOnly:
		sb.WriteString("- Compares files by size only (fastest)\n")
		sb.WriteString("- Best for: Large archives, encrypted backends\n")
	case clients.CompareSizeModTime:
		sb.WriteString("- Compares files by size and modification time (balanced)\n")
		sb.WriteString("- Best for: General use, most scenarios\n")
	case clients.CompareChecksum:
		sb.WriteString("- Compares files by hash/checksum (most accurate)\n")
		sb.WriteString("- Best for: Critical data, corruption detection\n")
	}

	sb.WriteString(fmt.Sprintf("\n**Skip Existing:** %v\n", client.GetSkipExisting()))
	sb.WriteString(fmt.Sprintf("**Update Only:** %v\n", client.GetUpdateOnly()))

	return tools.TextResult(sb.String()), nil
}

func handleAnalyzeSync(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	source := tools.GetStringParam(req, "source")
	destination := tools.GetStringParam(req, "destination")

	if source == "" || destination == "" {
		return tools.ErrorResult(fmt.Errorf("source and destination are required")), nil
	}
	if err := sanitize.RclonePath(source); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid source: %w", err)), nil
	}
	if err := sanitize.RclonePath(destination); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid destination: %w", err)), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	analysis, err := client.AnalyzeSync(ctx, source, destination)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("## Pre-Sync Analysis\n\n")
	sb.WriteString(fmt.Sprintf("**Source:** %s\n", analysis.Source))
	sb.WriteString(fmt.Sprintf("**Destination:** %s\n\n", analysis.Destination))

	sb.WriteString("### Transfer Summary\n")
	sb.WriteString(fmt.Sprintf("- **Total Files:** %d\n", analysis.TotalFiles))
	sb.WriteString(fmt.Sprintf("- **Files to Transfer:** %d\n", analysis.FilesToTransfer))
	sb.WriteString(fmt.Sprintf("- **Files to Skip:** %d\n", analysis.FilesToSkip))
	sb.WriteString(fmt.Sprintf("- **Data to Transfer:** %s\n", clients.FormatBytes(analysis.BytesToTransfer)))
	sb.WriteString(fmt.Sprintf("- **Estimated Time:** %s\n\n", analysis.EstimatedTime))

	if len(analysis.ByExtension) > 0 {
		sb.WriteString("### By File Type\n")
		sb.WriteString("| Extension | Count |\n")
		sb.WriteString("|-----------|-------|\n")
		for ext, count := range analysis.ByExtension {
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", ext, count))
		}
		sb.WriteString("\n")
	}

	if len(analysis.LargestFiles) > 0 {
		sb.WriteString("### Largest Files\n")
		sb.WriteString("| File | Size |\n")
		sb.WriteString("|------|------|\n")
		for _, f := range analysis.LargestFiles {
			sb.WriteString(fmt.Sprintf("| %s | %s |\n", truncate(f.Path, 50), f.SizeHuman))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleVerify(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	source := tools.GetStringParam(req, "source")
	destination := tools.GetStringParam(req, "destination")
	useChecksum := tools.GetBoolParam(req, "checksum", false)

	if source == "" || destination == "" {
		return tools.ErrorResult(fmt.Errorf("source and destination are required")), nil
	}
	if err := sanitize.RclonePath(source); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid source: %w", err)), nil
	}
	if err := sanitize.RclonePath(destination); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid destination: %w", err)), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result, err := client.VerifySync(ctx, source, destination, useChecksum)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("## Sync Verification Results\n\n")
	sb.WriteString(fmt.Sprintf("**Source:** %s\n", result.Source))
	sb.WriteString(fmt.Sprintf("**Destination:** %s\n", result.Destination))
	sb.WriteString(fmt.Sprintf("**Mode:** %s\n", map[bool]string{true: "Checksum", false: "Size+ModTime"}[useChecksum]))
	sb.WriteString(fmt.Sprintf("**Duration:** %s\n\n", result.Duration))

	sb.WriteString("### Summary\n")
	sb.WriteString(fmt.Sprintf("- ✓ **Verified:** %d files\n", result.Verified))
	sb.WriteString(fmt.Sprintf("- ✗ **Mismatched:** %d files\n", result.Mismatched))
	sb.WriteString(fmt.Sprintf("- ? **Missing:** %d files\n", result.Missing))
	sb.WriteString(fmt.Sprintf("- + **Extra on dest:** %d files\n\n", result.Extra))

	// Overall status
	if result.Mismatched == 0 && result.Missing == 0 {
		sb.WriteString("✓ **Verification PASSED** - All files match!\n")
	} else {
		sb.WriteString("✗ **Verification FAILED** - Differences found!\n")
	}

	// Show details if there are issues
	if len(result.MismatchList) > 0 {
		sb.WriteString("\n### Mismatched Files (sample)\n")
		for i, f := range result.MismatchList {
			if i >= 10 {
				sb.WriteString(fmt.Sprintf("- ... and %d more\n", len(result.MismatchList)-10))
				break
			}
			sb.WriteString(fmt.Sprintf("- %s\n", f))
		}
	}

	if len(result.MissingList) > 0 {
		sb.WriteString("\n### Missing Files (sample)\n")
		for i, f := range result.MissingList {
			if i >= 10 {
				sb.WriteString(fmt.Sprintf("- ... and %d more\n", len(result.MissingList)-10))
				break
			}
			sb.WriteString(fmt.Sprintf("- %s\n", f))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleDashboard(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	dashboard := client.GetDashboard()

	var sb strings.Builder
	sb.WriteString("## Rclone Sync Dashboard\n\n")
	sb.WriteString(fmt.Sprintf("*Generated: %s*\n\n", dashboard.GeneratedAt.Format(time.RFC3339)))

	// Active Jobs
	sb.WriteString("### Active Jobs\n")
	if len(dashboard.ActiveJobs) == 0 {
		sb.WriteString("No active sync jobs.\n\n")
	} else {
		sb.WriteString("| Job ID | Source | Status | Progress |\n")
		sb.WriteString("|--------|--------|--------|----------|\n")
		for _, job := range dashboard.ActiveJobs {
			progress := "-"
			if job.Progress != nil {
				progress = fmt.Sprintf("%.1f%%", job.Progress.PercentComplete)
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
				job.ID, truncate(job.Source, 25), job.Status, progress))
		}
		sb.WriteString("\n")
	}

	// Aggregate Progress
	if dashboard.TotalProgress != nil && dashboard.TotalProgress.TotalBytes > 0 {
		sb.WriteString("### Total Progress\n")
		sb.WriteString(fmt.Sprintf("- **Files:** %d / %d\n",
			dashboard.TotalProgress.TransferredFiles, dashboard.TotalProgress.TotalFiles))
		sb.WriteString(fmt.Sprintf("- **Bytes:** %s / %s (%.1f%%)\n",
			clients.FormatBytes(dashboard.TotalProgress.TransferredBytes),
			clients.FormatBytes(dashboard.TotalProgress.TotalBytes),
			dashboard.TotalProgress.OverallPercent))
		sb.WriteString(fmt.Sprintf("- **ETA:** %s\n\n", dashboard.TotalProgress.CombinedETA))
	}

	// Speed Metrics
	if dashboard.SpeedMetrics != nil {
		sb.WriteString("### Speed Metrics\n")
		sb.WriteString(fmt.Sprintf("- **Current:** %.2f MB/s\n", dashboard.SpeedMetrics.CurrentSpeed))
		sb.WriteString(fmt.Sprintf("- **Average:** %.2f MB/s\n", dashboard.SpeedMetrics.AverageSpeed))
		sb.WriteString(fmt.Sprintf("- **Peak:** %.2f MB/s\n", dashboard.SpeedMetrics.PeakSpeed))
		sb.WriteString(fmt.Sprintf("- **Trend:** %s\n\n", dashboard.SpeedMetrics.SpeedTrend))
	}

	// Quota Status
	if dashboard.QuotaStatus != nil {
		dailyLimit := int64(750 * 1024 * 1024 * 1024)
		usedPercent := float64(dashboard.QuotaStatus.BytesUploaded) / float64(dailyLimit) * 100
		sb.WriteString("### Daily Quota\n")
		sb.WriteString(fmt.Sprintf("- **Used:** %s / 750 GiB (%.1f%%)\n",
			clients.FormatBytes(dashboard.QuotaStatus.BytesUploaded), usedPercent))
		sb.WriteString(fmt.Sprintf("- **Date:** %s\n\n", dashboard.QuotaStatus.Date))
	}

	// Recent Errors
	if len(dashboard.RecentErrors) > 0 {
		sb.WriteString("### Recent Errors\n")
		for _, err := range dashboard.RecentErrors {
			sb.WriteString(fmt.Sprintf("- %s\n", err))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleHistory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	days := tools.GetIntParam(req, "days", 7)
	if days <= 0 {
		days = 7
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	history, err := client.GetTransferHistory(days)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	stats, err := client.GetTransferStats()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Transfer History (Last %d Days)\n\n", days))

	// Stats summary
	sb.WriteString("### Overall Statistics\n")
	sb.WriteString(fmt.Sprintf("- **Total Transfers:** %v\n", stats["total_transfers"]))
	sb.WriteString(fmt.Sprintf("- **Total Data:** %v\n", stats["total_bytes_human"]))
	sb.WriteString(fmt.Sprintf("- **Total Files:** %v\n", stats["total_files"]))
	sb.WriteString(fmt.Sprintf("- **Completed:** %v\n", stats["completed"]))
	sb.WriteString(fmt.Sprintf("- **Failed:** %v\n", stats["failed"]))
	if avgSpeed, ok := stats["average_speed"].(float64); ok && avgSpeed > 0 {
		sb.WriteString(fmt.Sprintf("- **Average Speed:** %.2f MB/s\n", avgSpeed/(1024*1024)))
	}
	sb.WriteString("\n")

	// Recent transfers
	if len(history) == 0 {
		sb.WriteString("No transfers in the selected period.\n")
	} else {
		sb.WriteString("### Recent Transfers\n")
		sb.WriteString("| Job ID | Source | Status | Size | Duration |\n")
		sb.WriteString("|--------|--------|--------|------|----------|\n")
		for i, record := range history {
			if i >= 20 {
				sb.WriteString(fmt.Sprintf("\n*Showing 20 of %d records*\n", len(history)))
				break
			}
			duration := record.EndTime.Sub(record.StartTime).Round(time.Second)
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
				record.JobID,
				truncate(record.Source, 25),
				record.Status,
				clients.FormatBytes(record.BytesTransferred),
				duration))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleFindDuplicates(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	indexID, errResult := tools.RequireStringParam(req, "index_id")
	if errResult != nil {
		return errResult, nil
	}

	// Get the DataMigrationClient to access hash indexes
	dmClient, err := clients.NewDataMigrationClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	index, err := dmClient.LoadIndex(indexID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("hash index not found: %w", err)), nil
	}

	rclClient, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	duplicates := rclClient.FindDuplicatesInHashIndex(index)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Duplicate Files in Index: %s\n\n", indexID))

	if len(duplicates) == 0 {
		sb.WriteString("No duplicate files found.\n")
		return tools.TextResult(sb.String()), nil
	}

	// Calculate total wasted space
	var totalWasted int64
	for _, d := range duplicates {
		totalWasted += d.WastedSize
	}

	sb.WriteString("### Summary\n")
	sb.WriteString(fmt.Sprintf("- **Duplicate Groups:** %d\n", len(duplicates)))
	sb.WriteString(fmt.Sprintf("- **Total Wasted Space:** %s\n\n", clients.FormatBytes(totalWasted)))

	sb.WriteString("### Top Duplicate Groups\n")
	sb.WriteString("| Size | Count | Wasted | Sample Path |\n")
	sb.WriteString("|------|-------|--------|-------------|\n")

	for i, d := range duplicates {
		if i >= 20 {
			sb.WriteString(fmt.Sprintf("\n*Showing 20 of %d groups*\n", len(duplicates)))
			break
		}
		sb.WriteString(fmt.Sprintf("| %s | %d | %s | %s |\n",
			d.SizeHuman,
			len(d.Paths),
			clients.FormatBytes(d.WastedSize),
			truncate(d.Paths[0], 40)))
	}

	return tools.TextResult(sb.String()), nil
}

// Phase 4 Handlers

func handleSyncVerified(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	source, errResult := tools.RequireStringParam(req, "source")
	if errResult != nil {
		return errResult, nil
	}
	dest, errResult := tools.RequireStringParam(req, "dest")
	if errResult != nil {
		return errResult, nil
	}
	autoVerify := tools.GetBoolParam(req, "auto_verify", true)
	verifyMode := tools.GetStringParam(req, "verify_mode")
	retryMismatches := tools.GetBoolParam(req, "retry_mismatches", false)
	maxRetries := tools.GetIntParam(req, "max_retries", 3)
	if err := sanitize.RclonePath(source); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid source: %w", err)), nil
	}
	if err := sanitize.RclonePath(dest); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid dest: %w", err)), nil
	}

	if verifyMode == "" {
		verifyMode = "size"
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	opts := clients.SyncOptions{
		AutoVerify:      autoVerify,
		VerifyMode:      verifyMode,
		RetryMismatches: retryMismatches,
		MaxRetries:      maxRetries,
	}

	result, err := client.SyncWithVerify(ctx, source, dest, opts)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("sync failed: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("## Verified Sync Result\n\n")
	sb.WriteString(fmt.Sprintf("**Source:** %s\n", source))
	sb.WriteString(fmt.Sprintf("**Destination:** %s\n", dest))
	sb.WriteString(fmt.Sprintf("**Final Status:** %s\n\n", result.FinalStatus))

	if result.Job != nil {
		sb.WriteString("### Sync Job\n")
		sb.WriteString(fmt.Sprintf("- **Job ID:** `%s`\n", result.Job.ID))
		sb.WriteString(fmt.Sprintf("- **Status:** %s\n", result.Job.Status))
		if result.Job.Progress != nil {
			sb.WriteString(fmt.Sprintf("- **Files:** %d\n", result.Job.Progress.FilesTransferred))
			sb.WriteString(fmt.Sprintf("- **Bytes:** %s\n", clients.FormatBytes(result.Job.Progress.BytesTransferred)))
		}
		if result.Job.Error != "" {
			sb.WriteString(fmt.Sprintf("- **Error:** %s\n", result.Job.Error))
		}
	}

	if result.VerifyResult != nil {
		sb.WriteString("\n### Verification\n")
		sb.WriteString(fmt.Sprintf("- **Verified:** %d files\n", result.VerifyResult.Verified))
		sb.WriteString(fmt.Sprintf("- **Mismatched:** %d files\n", result.VerifyResult.Mismatched))
		sb.WriteString(fmt.Sprintf("- **Missing:** %d files\n", result.VerifyResult.Missing))
		sb.WriteString(fmt.Sprintf("- **Extra:** %d files\n", result.VerifyResult.Extra))
		sb.WriteString(fmt.Sprintf("- **Duration:** %s\n", result.VerifyResult.Duration))
	}

	if result.RetryResult != nil {
		sb.WriteString("\n### Retry Result\n")
		sb.WriteString(fmt.Sprintf("- **Retry Status:** %s\n", result.RetryResult.Status))
		if result.RetryResult.Error != "" {
			sb.WriteString(fmt.Sprintf("- **Error:** %s\n", result.RetryResult.Error))
		}
	}

	// Status-specific messages
	switch result.FinalStatus {
	case "verified":
		sb.WriteString("\n**All files verified successfully.**\n")
	case "mismatches":
		sb.WriteString("\n**Some files have mismatches.** Use `aftrs_rclone_failure_analysis` to investigate.\n")
	case "failed":
		sb.WriteString("\n**Sync failed.** Check error messages above.\n")
	}

	return tools.TextResult(sb.String()), nil
}

func handleFailureAnalysis(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jobID, errResult := tools.RequireStringParam(req, "job_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	analysis, err := client.GetFailureAnalysis(jobID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to analyze failures: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("## Failure Analysis\n\n")
	sb.WriteString(fmt.Sprintf("**Job ID:** `%s`\n", analysis.JobID))
	sb.WriteString(fmt.Sprintf("**Analyzed:** %s\n\n", analysis.AnalyzedAt.Format("2006-01-02 15:04:05")))

	if analysis.TotalFailed == 0 {
		sb.WriteString("No failed files found for this job.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("### Summary\n")
	sb.WriteString(fmt.Sprintf("- **Total Failed:** %d files\n", analysis.TotalFailed))
	sb.WriteString(fmt.Sprintf("- **Total Size:** %s\n", clients.FormatBytes(analysis.TotalSize)))
	sb.WriteString(fmt.Sprintf("- **Recoverable:** %d files\n\n", analysis.Recoverable))

	sb.WriteString("### By Error Type\n")
	sb.WriteString("| Error Type | Count |\n")
	sb.WriteString("|------------|-------|\n")
	for errType, count := range analysis.ByErrorType {
		sb.WriteString(fmt.Sprintf("| %s | %d |\n", errType, count))
	}

	if len(analysis.ByExtension) > 0 {
		sb.WriteString("\n### By File Extension\n")
		sb.WriteString("| Extension | Count |\n")
		sb.WriteString("|-----------|-------|\n")
		for ext, count := range analysis.ByExtension {
			sb.WriteString(fmt.Sprintf("| %s | %d |\n", ext, count))
		}
	}

	if len(analysis.Suggestions) > 0 {
		sb.WriteString("\n### Recovery Suggestions\n\n")
		for _, s := range analysis.Suggestions {
			icon := "⚠️"
			if s.AutoFixable {
				icon = "✅"
			}
			sb.WriteString(fmt.Sprintf("**%s %s** (%d files)\n", icon, s.ErrorType, s.Count))
			sb.WriteString(fmt.Sprintf("> %s\n", s.Suggestion))
			if s.Command != "" {
				sb.WriteString(fmt.Sprintf("> Command: `%s`\n", s.Command))
			}
			sb.WriteString("\n")
		}
	}

	// Show sample failed files
	if len(analysis.FailedFiles) > 0 {
		sb.WriteString("### Sample Failed Files\n")
		sb.WriteString("| Path | Size | Error Type |\n")
		sb.WriteString("|------|------|------------|\n")

		maxShow := 10
		if len(analysis.FailedFiles) < maxShow {
			maxShow = len(analysis.FailedFiles)
		}
		for i := 0; i < maxShow; i++ {
			f := analysis.FailedFiles[i]
			path := f.Path
			if len(path) > 40 {
				path = "..." + path[len(path)-37:]
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", path, clients.FormatBytes(f.Size), f.ErrorType))
		}
		if len(analysis.FailedFiles) > maxShow {
			sb.WriteString(fmt.Sprintf("\n*Showing %d of %d failed files*\n", maxShow, len(analysis.FailedFiles)))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
