// Package archive provides MCP tools for DJ and VJ archive management
package archive

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

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Module implements the archive tools module
type Module struct{}

// Name returns the module name
func (m *Module) Name() string { return "archive" }

// Description returns the module description
func (m *Module) Description() string {
	return "DJ and VJ archive management for S3 and local storage"
}

// ArchiveInfo represents archive status information
type ArchiveInfo struct {
	Name         string `json:"name"`
	Type         string `json:"type"` // dj, vj, cloud
	Location     string `json:"location"`
	SizeBytes    int64  `json:"size_bytes"`
	SizeHuman    string `json:"size_human"`
	FileCount    int    `json:"file_count"`
	LastSync     string `json:"last_sync,omitempty"`
	StorageClass string `json:"storage_class,omitempty"`
	Status       string `json:"status"`
}

// SyncStatus represents a sync operation status
type SyncStatus struct {
	Source      string    `json:"source"`
	Destination string    `json:"destination"`
	Status      string    `json:"status"`
	Progress    float64   `json:"progress"`
	Transferred int64     `json:"transferred_bytes"`
	Remaining   int64     `json:"remaining_bytes"`
	Speed       string    `json:"speed"`
	ETA         string    `json:"eta"`
	StartedAt   time.Time `json:"started_at"`
	Errors      []string  `json:"errors,omitempty"`
}

// GlacierRestore represents a Glacier restore request
type GlacierRestore struct {
	ID          string    `json:"id"`
	Path        string    `json:"path"`
	Status      string    `json:"status"` // initiated, in_progress, completed, failed
	Tier        string    `json:"tier"`   // expedited, standard, bulk
	RequestedAt time.Time `json:"requested_at"`
	ExpiresAt   time.Time `json:"expires_at,omitempty"`
}

// StorageCost represents storage cost estimate
type StorageCost struct {
	Archive      string  `json:"archive"`
	StorageGB    float64 `json:"storage_gb"`
	MonthlyCost  float64 `json:"monthly_cost_usd"`
	StorageClass string  `json:"storage_class"`
}

// Tools returns all archive tools
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_archive_status",
				mcp.WithDescription("Get combined status of all archives (DJ, VJ, cloud) including size, file count, and last sync"),
			),
			Handler:             handleArchiveStatus,
			Category:            "archive",
			Subcategory:         "status",
			Tags:                []string{"archive", "status", "dj", "vj", "cloud"},
			UseCases:            []string{"Check all archives", "Overview of storage"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "archive",
		},
		{
			Tool: mcp.NewTool("aftrs_archive_sync_dj",
				mcp.WithDescription("Sync DJ archive to/from S3 (dj-archive bucket)"),
				mcp.WithString("direction", mcp.Required(), mcp.Description("Sync direction: upload (local->S3) or download (S3->local)"), mcp.Enum("upload", "download")),
				mcp.WithString("path", mcp.Description("Specific path to sync (optional, syncs entire archive if not specified)")),
				mcp.WithBoolean("dry_run", mcp.Description("Preview changes without syncing (default: true)")),
			),
			Handler:             handleSyncDJ,
			Category:            "archive",
			Subcategory:         "sync",
			Tags:                []string{"archive", "sync", "dj", "s3"},
			UseCases:            []string{"Backup DJ library", "Restore DJ archive"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "archive",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_archive_sync_vj",
				mcp.WithDescription("Sync VJ archive to/from S3 (vj-archive bucket)"),
				mcp.WithString("direction", mcp.Required(), mcp.Description("Sync direction: upload (local->S3) or download (S3->local)"), mcp.Enum("upload", "download")),
				mcp.WithString("path", mcp.Description("Specific path to sync (optional)")),
				mcp.WithBoolean("dry_run", mcp.Description("Preview changes without syncing (default: true)")),
			),
			Handler:             handleSyncVJ,
			Category:            "archive",
			Subcategory:         "sync",
			Tags:                []string{"archive", "sync", "vj", "s3"},
			UseCases:            []string{"Backup VJ clips", "Restore VJ archive"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "archive",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_archive_glacier_restore",
				mcp.WithDescription("Initiate Glacier restore for archived content"),
				mcp.WithString("archive", mcp.Required(), mcp.Description("Archive to restore from"), mcp.Enum("dj-archive", "vj-archive")),
				mcp.WithString("path", mcp.Required(), mcp.Description("Path to restore (folder or file)")),
				mcp.WithString("tier", mcp.Description("Restore tier: expedited (1-5 min), standard (3-5 hrs), bulk (5-12 hrs)"), mcp.Enum("expedited", "standard", "bulk")),
				mcp.WithNumber("days", mcp.Description("Number of days to keep restored (default: 7)")),
			),
			Handler:             handleGlacierRestore,
			Category:            "archive",
			Subcategory:         "glacier",
			Tags:                []string{"archive", "glacier", "restore", "s3"},
			UseCases:            []string{"Restore archived content", "Retrieve old projects"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "archive",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_archive_search",
				mcp.WithDescription("Search across all archives by filename, metadata, or tags"),
				mcp.WithString("query", mcp.Required(), mcp.Description("Search query (filename, extension, or keyword)")),
				mcp.WithString("archive", mcp.Description("Limit search to specific archive"), mcp.Enum("dj", "vj", "all")),
				mcp.WithNumber("limit", mcp.Description("Max results (default: 50)")),
			),
			Handler:             handleArchiveSearch,
			Category:            "archive",
			Subcategory:         "search",
			Tags:                []string{"archive", "search", "find"},
			UseCases:            []string{"Find files in archives", "Search by name"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "archive",
		},
		{
			Tool: mcp.NewTool("aftrs_archive_cost_estimate",
				mcp.WithDescription("Estimate storage costs across all archive tiers"),
			),
			Handler:             handleCostEstimate,
			Category:            "archive",
			Subcategory:         "billing",
			Tags:                []string{"archive", "cost", "billing", "s3"},
			UseCases:            []string{"Estimate monthly costs", "Storage optimization"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "archive",
		},
		// === Archive Intelligence Tools (v2.16) ===
		{
			Tool: mcp.NewTool("aftrs_archive_dedup",
				mcp.WithDescription("Find duplicate files across DJ/VJ/Cloud archives using content hash comparison. Identifies wasted storage."),
				mcp.WithString("archives", mcp.Description("Archives to scan: dj, vj, cloud, or all (default: all)")),
				mcp.WithNumber("min_size_mb", mcp.Description("Minimum file size in MB to consider (default: 1)")),
				mcp.WithNumber("limit", mcp.Description("Maximum duplicate sets to return (default: 50)")),
			),
			Handler:             handleArchiveDedup,
			Category:            "archive",
			Subcategory:         "intelligence",
			Tags:                []string{"archive", "dedup", "duplicates", "optimization"},
			UseCases:            []string{"Find duplicate files", "Optimize storage"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "archive",
		},
		{
			Tool: mcp.NewTool("aftrs_archive_cold_detect",
				mcp.WithDescription("Identify files not accessed in 6+ months that could be moved to cold storage (Glacier)"),
				mcp.WithString("archive", mcp.Description("Archive to analyze: dj, vj, or all (default: all)")),
				mcp.WithNumber("months", mcp.Description("Months since last access threshold (default: 6)")),
				mcp.WithNumber("min_size_mb", mcp.Description("Minimum file size in MB (default: 10)")),
				mcp.WithNumber("limit", mcp.Description("Maximum files to return (default: 100)")),
			),
			Handler:             handleColdDetect,
			Category:            "archive",
			Subcategory:         "intelligence",
			Tags:                []string{"archive", "cold", "glacier", "optimization"},
			UseCases:            []string{"Find cold data", "Glacier migration candidates"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "archive",
		},
		{
			Tool: mcp.NewTool("aftrs_archive_glacier_optimize",
				mcp.WithDescription("Suggest optimal Glacier tier migrations based on access patterns and cost savings"),
				mcp.WithString("archive", mcp.Description("Archive to analyze: dj-archive or vj-archive")),
			),
			Handler:             handleGlacierOptimize,
			Category:            "archive",
			Subcategory:         "intelligence",
			Tags:                []string{"archive", "glacier", "optimize", "cost"},
			UseCases:            []string{"Optimize storage tiers", "Reduce S3 costs"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "archive",
		},
		{
			Tool: mcp.NewTool("aftrs_archive_cost_forecast",
				mcp.WithDescription("Project storage costs over 3/6/12 months based on growth trends"),
				mcp.WithNumber("months", mcp.Description("Forecast period in months (default: 12)")),
			),
			Handler:             handleCostForecast,
			Category:            "archive",
			Subcategory:         "intelligence",
			Tags:                []string{"archive", "cost", "forecast", "planning"},
			UseCases:            []string{"Budget planning", "Growth projection"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "archive",
		},
		{
			Tool: mcp.NewTool("aftrs_archive_usage_report",
				mcp.WithDescription("Generate access patterns and usage trends report across archives"),
				mcp.WithString("archive", mcp.Description("Archive to analyze: dj, vj, or all (default: all)")),
				mcp.WithNumber("days", mcp.Description("Days of history to analyze (default: 30)")),
			),
			Handler:             handleUsageReport,
			Category:            "archive",
			Subcategory:         "intelligence",
			Tags:                []string{"archive", "usage", "report", "analytics"},
			UseCases:            []string{"Usage analytics", "Access patterns"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "archive",
		},
		{
			Tool: mcp.NewTool("aftrs_archive_cleanup_preview",
				mcp.WithDescription("Preview files that can be safely deleted (duplicates, temp files, old versions)"),
				mcp.WithString("archive", mcp.Description("Archive to analyze: dj, vj, or all")),
				mcp.WithBoolean("include_duplicates", mcp.Description("Include duplicate files (default: true)")),
				mcp.WithBoolean("include_temp", mcp.Description("Include temporary files (default: true)")),
				mcp.WithBoolean("include_old_versions", mcp.Description("Include old file versions (default: false)")),
			),
			Handler:             handleCleanupPreview,
			Category:            "archive",
			Subcategory:         "intelligence",
			Tags:                []string{"archive", "cleanup", "preview", "optimization"},
			UseCases:            []string{"Preview deletable files", "Storage cleanup"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "archive",
		},
		{
			Tool: mcp.NewTool("aftrs_archive_restore_queue",
				mcp.WithDescription("Manage batch Glacier restore queue - list pending, add, or check status"),
				mcp.WithString("action", mcp.Description("Action: list, add, status, cancel"), mcp.Enum("list", "add", "status", "cancel")),
				mcp.WithArray("paths", mcp.Description("Paths to restore (for add action)"), func(schema map[string]any) { schema["items"] = map[string]any{"type": "string"} }),
				mcp.WithString("tier", mcp.Description("Restore tier for add action: expedited, standard, bulk"), mcp.Enum("expedited", "standard", "bulk")),
			),
			Handler:             handleRestoreQueue,
			Category:            "archive",
			Subcategory:         "intelligence",
			Tags:                []string{"archive", "glacier", "restore", "queue"},
			UseCases:            []string{"Batch restore", "Manage restore queue"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "archive",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_archive_sync_status",
				mcp.WithDescription("Cross-archive sync state - compare local vs S3 vs cloud for drift detection"),
				mcp.WithString("archive", mcp.Description("Archive to check: dj, vj, or all (default: all)")),
				mcp.WithBoolean("detailed", mcp.Description("Show detailed file differences (default: false)")),
			),
			Handler:             handleSyncStatus,
			Category:            "archive",
			Subcategory:         "intelligence",
			Tags:                []string{"archive", "sync", "status", "drift"},
			UseCases:            []string{"Check sync state", "Detect drift"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "archive",
		},
	}
}

// getArchiveConfig returns archive configuration from environment
func getArchiveConfig() map[string]string {
	return map[string]string{
		"dj_local":    getEnvOrDefault("AFTRS_DJ_ARCHIVE_LOCAL", "D:/DJ-Archive"),
		"dj_remote":   getEnvOrDefault("AFTRS_DJ_ARCHIVE_BUCKET", "s3:dj-archive"),
		"vj_local":    getEnvOrDefault("AFTRS_VJ_ARCHIVE_LOCAL", "D:/VJ-Archive"),
		"vj_remote":   getEnvOrDefault("AFTRS_VJ_ARCHIVE_BUCKET", "s3:vj-archive"),
		"cloud_drive": getEnvOrDefault("AFTRS_CLOUD_DRIVE", "gdrive:"),
	}
}

func getEnvOrDefault(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// runRclone executes an rclone command and returns output
func runRclone(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "rclone", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("rclone error: %s - %w", string(output), err)
	}
	return string(output), nil
}

// getArchiveSize gets the size and file count of a path using rclone
func getArchiveSize(ctx context.Context, path string) (int64, int, error) {
	output, err := runRclone(ctx, "size", path, "--json")
	if err != nil {
		return 0, 0, err
	}

	var result struct {
		Count int   `json:"count"`
		Bytes int64 `json:"bytes"`
	}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		return 0, 0, err
	}

	return result.Bytes, result.Count, nil
}

func handleArchiveStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	config := getArchiveConfig()
	var archives []ArchiveInfo

	// Check DJ Archive (local)
	if djLocal := config["dj_local"]; djLocal != "" {
		size, count, err := getArchiveSize(ctx, djLocal)
		status := "available"
		if err != nil {
			status = "unavailable"
		}
		archives = append(archives, ArchiveInfo{
			Name:      "DJ Archive (Local)",
			Type:      "dj",
			Location:  djLocal,
			SizeBytes: size,
			SizeHuman: formatBytes(size),
			FileCount: count,
			Status:    status,
		})
	}

	// Check DJ Archive (S3)
	if djRemote := config["dj_remote"]; djRemote != "" {
		size, count, err := getArchiveSize(ctx, djRemote)
		status := "available"
		if err != nil {
			status = "unavailable"
		}
		archives = append(archives, ArchiveInfo{
			Name:         "DJ Archive (S3)",
			Type:         "dj",
			Location:     djRemote,
			SizeBytes:    size,
			SizeHuman:    formatBytes(size),
			FileCount:    count,
			StorageClass: "INTELLIGENT_TIERING",
			Status:       status,
		})
	}

	// Check VJ Archive (local)
	if vjLocal := config["vj_local"]; vjLocal != "" {
		size, count, err := getArchiveSize(ctx, vjLocal)
		status := "available"
		if err != nil {
			status = "unavailable"
		}
		archives = append(archives, ArchiveInfo{
			Name:      "VJ Archive (Local)",
			Type:      "vj",
			Location:  vjLocal,
			SizeBytes: size,
			SizeHuman: formatBytes(size),
			FileCount: count,
			Status:    status,
		})
	}

	// Check VJ Archive (S3)
	if vjRemote := config["vj_remote"]; vjRemote != "" {
		size, count, err := getArchiveSize(ctx, vjRemote)
		status := "available"
		if err != nil {
			status = "unavailable"
		}
		archives = append(archives, ArchiveInfo{
			Name:         "VJ Archive (S3)",
			Type:         "vj",
			Location:     vjRemote,
			SizeBytes:    size,
			SizeHuman:    formatBytes(size),
			FileCount:    count,
			StorageClass: "DEEP_ARCHIVE",
			Status:       status,
		})
	}

	// Check Cloud Drive
	if cloudDrive := config["cloud_drive"]; cloudDrive != "" {
		output, err := runRclone(ctx, "about", cloudDrive, "--json")
		if err == nil {
			var about struct {
				Used  int64 `json:"used"`
				Total int64 `json:"total"`
			}
			if json.Unmarshal([]byte(output), &about) == nil {
				archives = append(archives, ArchiveInfo{
					Name:      "Google Drive",
					Type:      "cloud",
					Location:  cloudDrive,
					SizeBytes: about.Used,
					SizeHuman: formatBytes(about.Used),
					Status:    "available",
				})
			}
		}
	}

	// Build response
	var sb strings.Builder
	sb.WriteString("# Archive Status\n\n")

	var totalSize int64
	var totalFiles int
	for _, a := range archives {
		totalSize += a.SizeBytes
		totalFiles += a.FileCount
	}

	sb.WriteString(fmt.Sprintf("**Total Storage:** %s across %d files\n\n", formatBytes(totalSize), totalFiles))

	sb.WriteString("| Archive | Location | Size | Files | Status |\n")
	sb.WriteString("|---------|----------|------|-------|--------|\n")

	for _, a := range archives {
		status := "?"
		if a.Status == "available" {
			status = "?"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %d | %s |\n",
			a.Name, a.Location, a.SizeHuman, a.FileCount, status))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleSyncDJ(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	direction := tools.GetStringParam(request, "direction")
	path := tools.GetStringParam(request, "path")
	dryRun := tools.GetBoolParam(request, "dry_run", true)

	config := getArchiveConfig()
	local := config["dj_local"]
	remote := config["dj_remote"]

	if path != "" {
		local = filepath.Join(local, path)
		remote = remote + "/" + path
	}

	var source, dest string
	if direction == "upload" {
		source = local
		dest = remote
	} else {
		source = remote
		dest = local
	}

	args := []string{"sync", source, dest, "--progress", "-v"}
	if dryRun {
		args = append(args, "--dry-run")
	}

	output, err := runRclone(ctx, args...)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var sb strings.Builder
	sb.WriteString("# DJ Archive Sync\n\n")
	sb.WriteString(fmt.Sprintf("**Direction:** %s\n", direction))
	sb.WriteString(fmt.Sprintf("**Source:** %s\n", source))
	sb.WriteString(fmt.Sprintf("**Destination:** %s\n", dest))
	sb.WriteString(fmt.Sprintf("**Dry Run:** %v\n\n", dryRun))
	sb.WriteString("## Output\n\n```\n")
	sb.WriteString(output)
	sb.WriteString("\n```\n")

	if dryRun {
		sb.WriteString("\n*This was a dry run. Run with `dry_run: false` to execute.*\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleSyncVJ(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	direction := tools.GetStringParam(request, "direction")
	path := tools.GetStringParam(request, "path")
	dryRun := tools.GetBoolParam(request, "dry_run", true)

	config := getArchiveConfig()
	local := config["vj_local"]
	remote := config["vj_remote"]

	if path != "" {
		local = filepath.Join(local, path)
		remote = remote + "/" + path
	}

	var source, dest string
	if direction == "upload" {
		source = local
		dest = remote
	} else {
		source = remote
		dest = local
	}

	args := []string{"sync", source, dest, "--progress", "-v"}
	if dryRun {
		args = append(args, "--dry-run")
	}

	output, err := runRclone(ctx, args...)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var sb strings.Builder
	sb.WriteString("# VJ Archive Sync\n\n")
	sb.WriteString(fmt.Sprintf("**Direction:** %s\n", direction))
	sb.WriteString(fmt.Sprintf("**Source:** %s\n", source))
	sb.WriteString(fmt.Sprintf("**Destination:** %s\n", dest))
	sb.WriteString(fmt.Sprintf("**Dry Run:** %v\n\n", dryRun))
	sb.WriteString("## Output\n\n```\n")
	sb.WriteString(output)
	sb.WriteString("\n```\n")

	if dryRun {
		sb.WriteString("\n*This was a dry run. Run with `dry_run: false` to execute.*\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleGlacierRestore(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	archive := tools.GetStringParam(request, "archive")
	path := tools.GetStringParam(request, "path")
	tier := tools.GetStringParam(request, "tier")
	if tier == "" {
		tier = "standard"
	}
	days := tools.GetIntParam(request, "days", 7)

	// Map archive to S3 bucket
	bucket := archive
	fullPath := fmt.Sprintf("s3:%s/%s", bucket, path)

	// Use rclone backend restore command for Glacier
	args := []string{
		"backend", "restore",
		"-o", fmt.Sprintf("priority=%s", strings.ToUpper(tier)),
		"-o", fmt.Sprintf("lifetime=%d", days),
		fullPath,
	}

	output, err := runRclone(ctx, args...)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Glacier restore failed: %v", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Glacier Restore Initiated\n\n")
	sb.WriteString(fmt.Sprintf("**Archive:** %s\n", archive))
	sb.WriteString(fmt.Sprintf("**Path:** %s\n", path))
	sb.WriteString(fmt.Sprintf("**Tier:** %s\n", tier))
	sb.WriteString(fmt.Sprintf("**Days Available:** %d\n\n", days))

	// Estimate restore time based on tier
	var eta string
	switch tier {
	case "expedited":
		eta = "1-5 minutes"
	case "standard":
		eta = "3-5 hours"
	case "bulk":
		eta = "5-12 hours"
	}
	sb.WriteString(fmt.Sprintf("**Estimated Restore Time:** %s\n\n", eta))
	sb.WriteString("## Output\n\n```\n")
	sb.WriteString(output)
	sb.WriteString("\n```\n")

	return mcp.NewToolResultText(sb.String()), nil
}

func handleArchiveSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := tools.GetStringParam(request, "query")
	archive := tools.GetStringParam(request, "archive")
	if archive == "" {
		archive = "all"
	}
	limit := tools.GetIntParam(request, "limit", 50)

	config := getArchiveConfig()
	var searchPaths []string

	switch archive {
	case "dj":
		searchPaths = []string{config["dj_local"], config["dj_remote"]}
	case "vj":
		searchPaths = []string{config["vj_local"], config["vj_remote"]}
	default: // all
		searchPaths = []string{
			config["dj_local"],
			config["dj_remote"],
			config["vj_local"],
			config["vj_remote"],
		}
	}

	var allResults []string

	for _, path := range searchPaths {
		if path == "" {
			continue
		}
		// Use rclone lsf with filter
		args := []string{"lsf", path, "--recursive", "--include", "*" + query + "*"}
		output, err := runRclone(ctx, args...)
		if err != nil {
			continue
		}
		lines := strings.Split(strings.TrimSpace(output), "\n")
		for _, line := range lines {
			if line != "" {
				allResults = append(allResults, fmt.Sprintf("%s: %s", path, line))
			}
		}
	}

	// Apply limit
	if len(allResults) > limit {
		allResults = allResults[:limit]
	}

	var sb strings.Builder
	sb.WriteString("# Archive Search Results\n\n")
	sb.WriteString(fmt.Sprintf("**Query:** %s\n", query))
	sb.WriteString(fmt.Sprintf("**Found:** %d results\n\n", len(allResults)))

	if len(allResults) == 0 {
		sb.WriteString("No files found matching the query.\n")
	} else {
		for i, result := range allResults {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, result))
		}
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleCostEstimate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	config := getArchiveConfig()

	var costs []StorageCost

	// DJ Archive S3 (Intelligent Tiering)
	if djRemote := config["dj_remote"]; djRemote != "" {
		size, _, _ := getArchiveSize(ctx, djRemote)
		sizeGB := float64(size) / (1024 * 1024 * 1024)
		// Intelligent Tiering: ~$0.023/GB/month for frequent access
		cost := sizeGB * 0.023
		costs = append(costs, StorageCost{
			Archive:      "DJ Archive (S3)",
			StorageGB:    sizeGB,
			MonthlyCost:  cost,
			StorageClass: "INTELLIGENT_TIERING",
		})
	}

	// VJ Archive S3 (Deep Archive)
	if vjRemote := config["vj_remote"]; vjRemote != "" {
		size, _, _ := getArchiveSize(ctx, vjRemote)
		sizeGB := float64(size) / (1024 * 1024 * 1024)
		// Glacier Deep Archive: ~$0.00099/GB/month
		cost := sizeGB * 0.00099
		costs = append(costs, StorageCost{
			Archive:      "VJ Archive (S3)",
			StorageGB:    sizeGB,
			MonthlyCost:  cost,
			StorageClass: "DEEP_ARCHIVE",
		})
	}

	// Google Drive (included in Workspace subscription)
	if cloudDrive := config["cloud_drive"]; cloudDrive != "" {
		output, _ := runRclone(ctx, "about", cloudDrive, "--json")
		var about struct {
			Used int64 `json:"used"`
		}
		if json.Unmarshal([]byte(output), &about) == nil {
			sizeGB := float64(about.Used) / (1024 * 1024 * 1024)
			costs = append(costs, StorageCost{
				Archive:      "Google Drive",
				StorageGB:    sizeGB,
				MonthlyCost:  0, // Included in subscription
				StorageClass: "Standard",
			})
		}
	}

	// Calculate total
	var totalGB, totalCost float64
	for _, c := range costs {
		totalGB += c.StorageGB
		totalCost += c.MonthlyCost
	}

	var sb strings.Builder
	sb.WriteString("# Storage Cost Estimate\n\n")
	sb.WriteString(fmt.Sprintf("**Total Storage:** %.2f GB\n", totalGB))
	sb.WriteString(fmt.Sprintf("**Estimated Monthly Cost:** $%.2f USD\n\n", totalCost))

	sb.WriteString("| Archive | Size (GB) | Storage Class | Monthly Cost |\n")
	sb.WriteString("|---------|-----------|---------------|-------------|\n")

	for _, c := range costs {
		costStr := fmt.Sprintf("$%.2f", c.MonthlyCost)
		if c.MonthlyCost == 0 {
			costStr = "Included"
		}
		sb.WriteString(fmt.Sprintf("| %s | %.2f | %s | %s |\n",
			c.Archive, c.StorageGB, c.StorageClass, costStr))
	}

	sb.WriteString("\n*Note: Costs are estimates based on standard AWS S3 pricing.*\n")

	return mcp.NewToolResultText(sb.String()), nil
}

// formatBytes formats bytes to human readable string
func formatBytes(bytes int64) string {
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

// === Archive Intelligence Handlers (v2.16) ===

// DuplicateSet represents a set of duplicate files
type DuplicateSet struct {
	Hash       string   `json:"hash"`
	Size       int64    `json:"size"`
	Paths      []string `json:"paths"`
	WastedSize int64    `json:"wasted_size"`
}

// ColdFile represents a file not accessed recently
type ColdFile struct {
	Path         string    `json:"path"`
	Size         int64     `json:"size"`
	LastAccessed time.Time `json:"last_accessed"`
	DaysOld      int       `json:"days_old"`
}

// RestoreJob represents a Glacier restore job
type RestoreJob struct {
	ID        string    `json:"id"`
	Path      string    `json:"path"`
	Status    string    `json:"status"`
	Tier      string    `json:"tier"`
	CreatedAt time.Time `json:"created_at"`
}

// In-memory restore queue (in production would use persistent storage)
var restoreQueue []RestoreJob

func handleArchiveDedup(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	archives := tools.GetStringParam(request, "archives")
	if archives == "" {
		archives = "all"
	}
	minSizeMB := tools.GetIntParam(request, "min_size_mb", 1)
	limit := tools.GetIntParam(request, "limit", 50)

	config := getArchiveConfig()
	var searchPaths []string

	switch archives {
	case "dj":
		searchPaths = []string{config["dj_local"]}
	case "vj":
		searchPaths = []string{config["vj_local"]}
	case "cloud":
		searchPaths = []string{config["cloud_drive"]}
	default:
		searchPaths = []string{config["dj_local"], config["vj_local"]}
	}

	// Use rclone to find duplicates by hash
	var duplicates []DuplicateSet
	hashMap := make(map[string][]string) // hash -> paths
	sizeMap := make(map[string]int64)    // hash -> size

	for _, path := range searchPaths {
		if path == "" {
			continue
		}
		// List files with hash using rclone
		args := []string{"lsjson", path, "--recursive", "--hash", "--files-only"}
		output, err := runRclone(ctx, args...)
		if err != nil {
			continue
		}

		var files []struct {
			Path   string            `json:"Path"`
			Size   int64             `json:"Size"`
			Hashes map[string]string `json:"Hashes"`
		}
		if err := json.Unmarshal([]byte(output), &files); err != nil {
			continue
		}

		minSizeBytes := int64(minSizeMB) * 1024 * 1024
		for _, f := range files {
			if f.Size < minSizeBytes {
				continue
			}
			hash := ""
			if h, ok := f.Hashes["md5"]; ok {
				hash = h
			} else if h, ok := f.Hashes["sha1"]; ok {
				hash = h
			}
			if hash != "" {
				fullPath := path + "/" + f.Path
				hashMap[hash] = append(hashMap[hash], fullPath)
				sizeMap[hash] = f.Size
			}
		}
	}

	// Find duplicates (files with same hash)
	for hash, paths := range hashMap {
		if len(paths) > 1 {
			size := sizeMap[hash]
			wastedSize := size * int64(len(paths)-1)
			duplicates = append(duplicates, DuplicateSet{
				Hash:       hash,
				Size:       size,
				Paths:      paths,
				WastedSize: wastedSize,
			})
		}
	}

	// Sort by wasted size (largest first)
	for i := 0; i < len(duplicates)-1; i++ {
		for j := i + 1; j < len(duplicates); j++ {
			if duplicates[j].WastedSize > duplicates[i].WastedSize {
				duplicates[i], duplicates[j] = duplicates[j], duplicates[i]
			}
		}
	}

	// Apply limit
	if len(duplicates) > limit {
		duplicates = duplicates[:limit]
	}

	// Calculate total wasted space
	var totalWasted int64
	for _, d := range duplicates {
		totalWasted += d.WastedSize
	}

	var sb strings.Builder
	sb.WriteString("# Archive Duplicate Analysis\n\n")
	sb.WriteString(fmt.Sprintf("**Archives Scanned:** %s\n", archives))
	sb.WriteString(fmt.Sprintf("**Minimum File Size:** %d MB\n", minSizeMB))
	sb.WriteString(fmt.Sprintf("**Duplicate Sets Found:** %d\n", len(duplicates)))
	sb.WriteString(fmt.Sprintf("**Total Wasted Space:** %s\n\n", formatBytes(totalWasted)))

	if len(duplicates) == 0 {
		sb.WriteString("No duplicates found matching criteria.\n")
	} else {
		sb.WriteString("## Duplicate Sets\n\n")
		for i, d := range duplicates {
			sb.WriteString(fmt.Sprintf("### Set %d (%s wasted)\n", i+1, formatBytes(d.WastedSize)))
			sb.WriteString(fmt.Sprintf("**Size:** %s | **Hash:** %s\n\n", formatBytes(d.Size), d.Hash[:16]+"..."))
			for _, p := range d.Paths {
				sb.WriteString(fmt.Sprintf("- `%s`\n", p))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n## Actions\n")
	sb.WriteString("- Use `aftrs_archive_cleanup_preview` to preview deletable duplicates\n")
	sb.WriteString("- Manually review before deleting to ensure you keep the preferred copy\n")

	return mcp.NewToolResultText(sb.String()), nil
}

func handleColdDetect(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	archive := tools.GetStringParam(request, "archive")
	if archive == "" {
		archive = "all"
	}
	months := tools.GetIntParam(request, "months", 6)
	minSizeMB := tools.GetIntParam(request, "min_size_mb", 10)
	limit := tools.GetIntParam(request, "limit", 100)

	config := getArchiveConfig()
	var searchPaths []string

	switch archive {
	case "dj":
		searchPaths = []string{config["dj_local"]}
	case "vj":
		searchPaths = []string{config["vj_local"]}
	default:
		searchPaths = []string{config["dj_local"], config["vj_local"]}
	}

	cutoffDate := time.Now().AddDate(0, -months, 0)
	minSizeBytes := int64(minSizeMB) * 1024 * 1024
	var coldFiles []ColdFile

	for _, path := range searchPaths {
		if path == "" {
			continue
		}
		// List files with modification time
		args := []string{"lsjson", path, "--recursive", "--files-only"}
		output, err := runRclone(ctx, args...)
		if err != nil {
			continue
		}

		var files []struct {
			Path    string    `json:"Path"`
			Size    int64     `json:"Size"`
			ModTime time.Time `json:"ModTime"`
		}
		if err := json.Unmarshal([]byte(output), &files); err != nil {
			continue
		}

		for _, f := range files {
			if f.Size < minSizeBytes {
				continue
			}
			if f.ModTime.Before(cutoffDate) {
				daysOld := int(time.Since(f.ModTime).Hours() / 24)
				coldFiles = append(coldFiles, ColdFile{
					Path:         path + "/" + f.Path,
					Size:         f.Size,
					LastAccessed: f.ModTime,
					DaysOld:      daysOld,
				})
			}
		}
	}

	// Sort by days old (oldest first)
	for i := 0; i < len(coldFiles)-1; i++ {
		for j := i + 1; j < len(coldFiles); j++ {
			if coldFiles[j].DaysOld > coldFiles[i].DaysOld {
				coldFiles[i], coldFiles[j] = coldFiles[j], coldFiles[i]
			}
		}
	}

	// Apply limit
	if len(coldFiles) > limit {
		coldFiles = coldFiles[:limit]
	}

	// Calculate total cold storage
	var totalCold int64
	for _, f := range coldFiles {
		totalCold += f.Size
	}

	var sb strings.Builder
	sb.WriteString("# Cold Data Detection\n\n")
	sb.WriteString(fmt.Sprintf("**Archive:** %s\n", archive))
	sb.WriteString(fmt.Sprintf("**Threshold:** %d months\n", months))
	sb.WriteString(fmt.Sprintf("**Minimum Size:** %d MB\n", minSizeMB))
	sb.WriteString(fmt.Sprintf("**Cold Files Found:** %d\n", len(coldFiles)))
	sb.WriteString(fmt.Sprintf("**Total Cold Data:** %s\n\n", formatBytes(totalCold)))

	if len(coldFiles) == 0 {
		sb.WriteString("No cold files found matching criteria.\n")
	} else {
		sb.WriteString("## Cold Files (Glacier Migration Candidates)\n\n")
		sb.WriteString("| File | Size | Days Old |\n")
		sb.WriteString("|------|------|----------|\n")
		for _, f := range coldFiles {
			filename := filepath.Base(f.Path)
			if len(filename) > 40 {
				filename = filename[:37] + "..."
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %d |\n", filename, formatBytes(f.Size), f.DaysOld))
		}
	}

	sb.WriteString("\n## Recommendations\n")
	sb.WriteString("- Use `aftrs_archive_glacier_optimize` to see potential cost savings\n")
	sb.WriteString("- Consider moving these files to Glacier Deep Archive tier\n")

	return mcp.NewToolResultText(sb.String()), nil
}

func handleGlacierOptimize(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	archive := tools.GetStringParam(request, "archive")
	if archive == "" {
		archive = "vj-archive"
	}

	config := getArchiveConfig()

	// Get current storage info
	var localPath, remotePath string
	var currentClass, optimalClass string
	var currentRate, optimalRate float64

	switch archive {
	case "dj-archive", "dj":
		localPath = config["dj_local"]
		remotePath = config["dj_remote"]
		currentClass = "INTELLIGENT_TIERING"
		optimalClass = "GLACIER_IR" // Glacier Instant Retrieval
		currentRate = 0.023
		optimalRate = 0.004
	case "vj-archive", "vj":
		localPath = config["vj_local"]
		remotePath = config["vj_remote"]
		currentClass = "STANDARD"
		optimalClass = "DEEP_ARCHIVE"
		currentRate = 0.023
		optimalRate = 0.00099
	default:
		remotePath = "s3:" + archive
		currentClass = "STANDARD"
		optimalClass = "GLACIER"
		currentRate = 0.023
		optimalRate = 0.004
	}

	// Get size
	size, count, _ := getArchiveSize(ctx, remotePath)
	if size == 0 && localPath != "" {
		size, count, _ = getArchiveSize(ctx, localPath)
	}

	sizeGB := float64(size) / (1024 * 1024 * 1024)
	currentCost := sizeGB * currentRate
	optimalCost := sizeGB * optimalRate
	monthlySavings := currentCost - optimalCost
	yearlySavings := monthlySavings * 12

	var sb strings.Builder
	sb.WriteString("# Glacier Optimization Analysis\n\n")
	sb.WriteString(fmt.Sprintf("**Archive:** %s\n", archive))
	sb.WriteString(fmt.Sprintf("**Size:** %s (%d files)\n\n", formatBytes(size), count))

	sb.WriteString("## Current State\n")
	sb.WriteString(fmt.Sprintf("- **Storage Class:** %s\n", currentClass))
	sb.WriteString(fmt.Sprintf("- **Rate:** $%.5f/GB/month\n", currentRate))
	sb.WriteString(fmt.Sprintf("- **Monthly Cost:** $%.2f\n\n", currentCost))

	sb.WriteString("## Recommended Optimization\n")
	sb.WriteString(fmt.Sprintf("- **Target Class:** %s\n", optimalClass))
	sb.WriteString(fmt.Sprintf("- **Rate:** $%.5f/GB/month\n", optimalRate))
	sb.WriteString(fmt.Sprintf("- **Monthly Cost:** $%.2f\n\n", optimalCost))

	sb.WriteString("## Potential Savings\n")
	sb.WriteString(fmt.Sprintf("- **Monthly Savings:** $%.2f (%.0f%% reduction)\n", monthlySavings, (monthlySavings/currentCost)*100))
	sb.WriteString(fmt.Sprintf("- **Yearly Savings:** $%.2f\n\n", yearlySavings))

	sb.WriteString("## Storage Class Comparison\n\n")
	sb.WriteString("| Class | Rate/GB/Mo | Retrieval | Best For |\n")
	sb.WriteString("|-------|------------|-----------|----------|\n")
	sb.WriteString("| STANDARD | $0.023 | Instant | Active data |\n")
	sb.WriteString("| INTELLIGENT_TIERING | $0.023-0.004 | Varies | Unknown access |\n")
	sb.WriteString("| GLACIER_IR | $0.004 | Milliseconds | Archive, quick access |\n")
	sb.WriteString("| GLACIER_FR | $0.0036 | 1-5 min | Backup |\n")
	sb.WriteString("| DEEP_ARCHIVE | $0.00099 | 12 hrs | Long-term archive |\n\n")

	sb.WriteString("## Actions\n")
	sb.WriteString("- Run lifecycle policy to transition older objects\n")
	sb.WriteString("- Use `aftrs_archive_cold_detect` to find migration candidates\n")

	return mcp.NewToolResultText(sb.String()), nil
}

func handleCostForecast(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	months := tools.GetIntParam(request, "months", 12)

	config := getArchiveConfig()

	// Get current sizes
	djSize, _, _ := getArchiveSize(ctx, config["dj_remote"])
	vjSize, _, _ := getArchiveSize(ctx, config["vj_remote"])

	djGB := float64(djSize) / (1024 * 1024 * 1024)
	vjGB := float64(vjSize) / (1024 * 1024 * 1024)

	// Assume growth rates (typical for media archives)
	djGrowthRate := 0.02  // 2% per month for DJ
	vjGrowthRate := 0.015 // 1.5% per month for VJ (larger, slower growth)

	// Cost rates
	djRate := 0.023   // Intelligent Tiering
	vjRate := 0.00099 // Deep Archive

	var sb strings.Builder
	sb.WriteString("# Storage Cost Forecast\n\n")
	sb.WriteString(fmt.Sprintf("**Forecast Period:** %d months\n", months))
	sb.WriteString(fmt.Sprintf("**Current DJ Archive:** %.2f GB\n", djGB))
	sb.WriteString(fmt.Sprintf("**Current VJ Archive:** %.2f GB\n\n", vjGB))

	sb.WriteString("## Monthly Projections\n\n")
	sb.WriteString("| Month | DJ Size | DJ Cost | VJ Size | VJ Cost | Total |\n")
	sb.WriteString("|-------|---------|---------|---------|---------|-------|\n")

	currentDJ := djGB
	currentVJ := vjGB
	var totalCost float64

	for m := 1; m <= months; m++ {
		djCost := currentDJ * djRate
		vjCost := currentVJ * vjRate
		monthTotal := djCost + vjCost
		totalCost += monthTotal

		if m <= 3 || m == 6 || m == months {
			sb.WriteString(fmt.Sprintf("| %d | %.0f GB | $%.2f | %.0f GB | $%.2f | $%.2f |\n",
				m, currentDJ, djCost, currentVJ, vjCost, monthTotal))
		} else if m == 4 {
			sb.WriteString("| ... | ... | ... | ... | ... | ... |\n")
		}

		// Apply growth
		currentDJ *= (1 + djGrowthRate)
		currentVJ *= (1 + vjGrowthRate)
	}

	sb.WriteString(fmt.Sprintf("\n**Total %d-Month Cost:** $%.2f\n", months, totalCost))

	// Calculate growth
	finalDJ := djGB * (1 + float64(months)*djGrowthRate)
	finalVJ := vjGB * (1 + float64(months)*vjGrowthRate)

	sb.WriteString("\n## Growth Summary\n")
	sb.WriteString(fmt.Sprintf("- DJ Archive: %.0f GB → %.0f GB (+%.0f%%)\n", djGB, finalDJ, (finalDJ/djGB-1)*100))
	sb.WriteString(fmt.Sprintf("- VJ Archive: %.0f GB → %.0f GB (+%.0f%%)\n", vjGB, finalVJ, (finalVJ/vjGB-1)*100))

	sb.WriteString("\n## Cost Optimization Tips\n")
	sb.WriteString("- Move cold DJ data to Glacier IR to save ~80%\n")
	sb.WriteString("- Ensure VJ archive uses Deep Archive tier\n")
	sb.WriteString("- Run dedup to remove duplicate files\n")

	return mcp.NewToolResultText(sb.String()), nil
}

func handleUsageReport(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	archive := tools.GetStringParam(request, "archive")
	if archive == "" {
		archive = "all"
	}
	days := tools.GetIntParam(request, "days", 30)

	config := getArchiveConfig()

	var sb strings.Builder
	sb.WriteString("# Archive Usage Report\n\n")
	sb.WriteString(fmt.Sprintf("**Period:** Last %d days\n", days))
	sb.WriteString(fmt.Sprintf("**Archive:** %s\n\n", archive))

	// Get current stats
	var archives []struct {
		name  string
		path  string
		size  int64
		count int
	}

	if archive == "all" || archive == "dj" {
		size, count, _ := getArchiveSize(ctx, config["dj_local"])
		archives = append(archives, struct {
			name  string
			path  string
			size  int64
			count int
		}{"DJ Archive", config["dj_local"], size, count})
	}

	if archive == "all" || archive == "vj" {
		size, count, _ := getArchiveSize(ctx, config["vj_local"])
		archives = append(archives, struct {
			name  string
			path  string
			size  int64
			count int
		}{"VJ Archive", config["vj_local"], size, count})
	}

	sb.WriteString("## Current State\n\n")
	sb.WriteString("| Archive | Size | Files |\n")
	sb.WriteString("|---------|------|-------|\n")
	for _, a := range archives {
		sb.WriteString(fmt.Sprintf("| %s | %s | %d |\n", a.name, formatBytes(a.size), a.count))
	}

	sb.WriteString("\n## File Type Distribution\n\n")

	// Count by extension
	for _, a := range archives {
		if a.path == "" {
			continue
		}
		args := []string{"lsjson", a.path, "--recursive", "--files-only"}
		output, err := runRclone(ctx, args...)
		if err != nil {
			continue
		}

		var files []struct {
			Path string `json:"Path"`
			Size int64  `json:"Size"`
		}
		if json.Unmarshal([]byte(output), &files) != nil {
			continue
		}

		extCount := make(map[string]int)
		extSize := make(map[string]int64)
		for _, f := range files {
			ext := strings.ToLower(filepath.Ext(f.Path))
			if ext == "" {
				ext = "(no ext)"
			}
			extCount[ext]++
			extSize[ext] += f.Size
		}

		sb.WriteString(fmt.Sprintf("### %s\n\n", a.name))
		sb.WriteString("| Extension | Count | Size |\n")
		sb.WriteString("|-----------|-------|------|\n")

		// Get top 10 by count
		type extStat struct {
			ext   string
			count int
			size  int64
		}
		var stats []extStat
		for ext, count := range extCount {
			stats = append(stats, extStat{ext, count, extSize[ext]})
		}
		for i := 0; i < len(stats)-1; i++ {
			for j := i + 1; j < len(stats); j++ {
				if stats[j].count > stats[i].count {
					stats[i], stats[j] = stats[j], stats[i]
				}
			}
		}
		if len(stats) > 10 {
			stats = stats[:10]
		}
		for _, s := range stats {
			sb.WriteString(fmt.Sprintf("| %s | %d | %s |\n", s.ext, s.count, formatBytes(s.size)))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Recommendations\n")
	sb.WriteString("- Use `aftrs_archive_dedup` to find duplicate files\n")
	sb.WriteString("- Use `aftrs_archive_cold_detect` to find archivable files\n")
	sb.WriteString("- Consider lifecycle policies for automatic tiering\n")

	return mcp.NewToolResultText(sb.String()), nil
}

func handleCleanupPreview(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	archive := tools.GetStringParam(request, "archive")
	if archive == "" {
		archive = "all"
	}
	includeDuplicates := tools.GetBoolParam(request, "include_duplicates", true)
	includeTemp := tools.GetBoolParam(request, "include_temp", true)
	includeOldVersions := tools.GetBoolParam(request, "include_old_versions", false)

	config := getArchiveConfig()
	var searchPaths []string

	switch archive {
	case "dj":
		searchPaths = []string{config["dj_local"]}
	case "vj":
		searchPaths = []string{config["vj_local"]}
	default:
		searchPaths = []string{config["dj_local"], config["vj_local"]}
	}

	var sb strings.Builder
	sb.WriteString("# Cleanup Preview\n\n")
	sb.WriteString(fmt.Sprintf("**Archive:** %s\n", archive))
	sb.WriteString(fmt.Sprintf("**Include Duplicates:** %v\n", includeDuplicates))
	sb.WriteString(fmt.Sprintf("**Include Temp Files:** %v\n", includeTemp))
	sb.WriteString(fmt.Sprintf("**Include Old Versions:** %v\n\n", includeOldVersions))

	var totalReclaimable int64
	var tempFiles []string
	var oldVersions []string

	// Temp file patterns
	tempPatterns := []string{".tmp", ".temp", "~", ".bak", ".DS_Store", "Thumbs.db", ".partial"}

	for _, path := range searchPaths {
		if path == "" {
			continue
		}

		args := []string{"lsjson", path, "--recursive", "--files-only"}
		output, err := runRclone(ctx, args...)
		if err != nil {
			continue
		}

		var files []struct {
			Path string `json:"Path"`
			Size int64  `json:"Size"`
		}
		if json.Unmarshal([]byte(output), &files) != nil {
			continue
		}

		for _, f := range files {
			// Check for temp files
			if includeTemp {
				for _, pattern := range tempPatterns {
					if strings.Contains(f.Path, pattern) || strings.HasSuffix(f.Path, pattern) {
						tempFiles = append(tempFiles, path+"/"+f.Path)
						totalReclaimable += f.Size
						break
					}
				}
			}

			// Check for old versions (files with version patterns like _v1, _old, etc.)
			if includeOldVersions {
				lowPath := strings.ToLower(f.Path)
				if strings.Contains(lowPath, "_old") ||
					strings.Contains(lowPath, "_backup") ||
					strings.Contains(lowPath, "_copy") ||
					strings.Contains(lowPath, " (1)") ||
					strings.Contains(lowPath, " (2)") {
					oldVersions = append(oldVersions, path+"/"+f.Path)
					totalReclaimable += f.Size
				}
			}
		}
	}

	sb.WriteString("## Cleanup Candidates\n\n")

	if includeTemp && len(tempFiles) > 0 {
		sb.WriteString("### Temporary Files\n")
		sb.WriteString(fmt.Sprintf("Found %d temp files\n\n", len(tempFiles)))
		for i, f := range tempFiles {
			if i >= 20 {
				sb.WriteString(fmt.Sprintf("... and %d more\n", len(tempFiles)-20))
				break
			}
			sb.WriteString(fmt.Sprintf("- `%s`\n", f))
		}
		sb.WriteString("\n")
	}

	if includeOldVersions && len(oldVersions) > 0 {
		sb.WriteString("### Old Versions\n")
		sb.WriteString(fmt.Sprintf("Found %d old version files\n\n", len(oldVersions)))
		for i, f := range oldVersions {
			if i >= 20 {
				sb.WriteString(fmt.Sprintf("... and %d more\n", len(oldVersions)-20))
				break
			}
			sb.WriteString(fmt.Sprintf("- `%s`\n", f))
		}
		sb.WriteString("\n")
	}

	if includeDuplicates {
		sb.WriteString("### Duplicates\n")
		sb.WriteString("Run `aftrs_archive_dedup` for detailed duplicate analysis\n\n")
	}

	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("**Temp Files:** %d\n", len(tempFiles)))
	sb.WriteString(fmt.Sprintf("**Old Versions:** %d\n", len(oldVersions)))
	sb.WriteString(fmt.Sprintf("**Estimated Reclaimable:** %s\n\n", formatBytes(totalReclaimable)))

	sb.WriteString("⚠️ *This is a preview only. Files have not been deleted.*\n")
	sb.WriteString("Review carefully before manual deletion.\n")

	return mcp.NewToolResultText(sb.String()), nil
}

func handleRestoreQueue(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action := tools.GetStringParam(request, "action")
	if action == "" {
		action = "list"
	}
	tier := tools.GetStringParam(request, "tier")
	if tier == "" {
		tier = "standard"
	}

	var sb strings.Builder

	switch action {
	case "list":
		sb.WriteString("# Glacier Restore Queue\n\n")
		if len(restoreQueue) == 0 {
			sb.WriteString("No pending restore jobs.\n")
		} else {
			sb.WriteString("| ID | Path | Tier | Status | Created |\n")
			sb.WriteString("|----|------|------|--------|--------|\n")
			for _, job := range restoreQueue {
				sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
					job.ID[:8], job.Path, job.Tier, job.Status, job.CreatedAt.Format("2006-01-02 15:04")))
			}
		}

	case "add":
		// Get paths from request - handle as interface slice
		var paths []string
		if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
			if pathsRaw, exists := args["paths"]; exists && pathsRaw != nil {
				if pathSlice, ok := pathsRaw.([]interface{}); ok {
					for _, p := range pathSlice {
						if ps, ok := p.(string); ok {
							paths = append(paths, ps)
						}
					}
				}
			}
		}

		if len(paths) == 0 {
			return mcp.NewToolResultError("No paths provided for restore"), nil
		}

		sb.WriteString("# Adding to Restore Queue\n\n")
		sb.WriteString(fmt.Sprintf("**Tier:** %s\n\n", tier))

		for _, path := range paths {
			job := RestoreJob{
				ID:        fmt.Sprintf("%d", time.Now().UnixNano()),
				Path:      path,
				Tier:      tier,
				Status:    "pending",
				CreatedAt: time.Now(),
			}
			restoreQueue = append(restoreQueue, job)
			sb.WriteString(fmt.Sprintf("- Added: `%s` (ID: %s)\n", path, job.ID[:8]))
		}

		sb.WriteString(fmt.Sprintf("\n**Total in Queue:** %d\n", len(restoreQueue)))

	case "status":
		sb.WriteString("# Restore Queue Status\n\n")
		pending := 0
		inProgress := 0
		completed := 0
		for _, job := range restoreQueue {
			switch job.Status {
			case "pending":
				pending++
			case "in_progress":
				inProgress++
			case "completed":
				completed++
			}
		}
		sb.WriteString(fmt.Sprintf("- **Pending:** %d\n", pending))
		sb.WriteString(fmt.Sprintf("- **In Progress:** %d\n", inProgress))
		sb.WriteString(fmt.Sprintf("- **Completed:** %d\n", completed))

	case "cancel":
		sb.WriteString("# Clearing Restore Queue\n\n")
		count := len(restoreQueue)
		restoreQueue = nil
		sb.WriteString(fmt.Sprintf("Cleared %d jobs from queue.\n", count))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleSyncStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	archive := tools.GetStringParam(request, "archive")
	if archive == "" {
		archive = "all"
	}
	detailed := tools.GetBoolParam(request, "detailed", false)

	config := getArchiveConfig()

	var sb strings.Builder
	sb.WriteString("# Archive Sync Status\n\n")

	type syncCheck struct {
		name        string
		local       string
		remote      string
		localSize   int64
		remoteSize  int64
		localCount  int
		remoteCount int
	}

	var checks []syncCheck

	if archive == "all" || archive == "dj" {
		localSize, localCount, _ := getArchiveSize(ctx, config["dj_local"])
		remoteSize, remoteCount, _ := getArchiveSize(ctx, config["dj_remote"])
		checks = append(checks, syncCheck{
			name:        "DJ Archive",
			local:       config["dj_local"],
			remote:      config["dj_remote"],
			localSize:   localSize,
			remoteSize:  remoteSize,
			localCount:  localCount,
			remoteCount: remoteCount,
		})
	}

	if archive == "all" || archive == "vj" {
		localSize, localCount, _ := getArchiveSize(ctx, config["vj_local"])
		remoteSize, remoteCount, _ := getArchiveSize(ctx, config["vj_remote"])
		checks = append(checks, syncCheck{
			name:        "VJ Archive",
			local:       config["vj_local"],
			remote:      config["vj_remote"],
			localSize:   localSize,
			remoteSize:  remoteSize,
			localCount:  localCount,
			remoteCount: remoteCount,
		})
	}

	sb.WriteString("## Sync Overview\n\n")
	sb.WriteString("| Archive | Local | Remote | Status |\n")
	sb.WriteString("|---------|-------|--------|--------|\n")

	for _, c := range checks {
		status := "✅ Synced"
		sizeDiff := c.localSize - c.remoteSize
		if sizeDiff < 0 {
			sizeDiff = -sizeDiff
		}
		// Allow 1% variance
		variance := float64(sizeDiff) / float64(c.localSize+1) * 100
		if variance > 1 {
			status = "⚠️ Drift detected"
		}
		if c.localCount != c.remoteCount {
			status = "⚠️ File count mismatch"
		}

		sb.WriteString(fmt.Sprintf("| %s | %s (%d files) | %s (%d files) | %s |\n",
			c.name, formatBytes(c.localSize), c.localCount,
			formatBytes(c.remoteSize), c.remoteCount, status))
	}

	if detailed {
		sb.WriteString("\n## Detailed Comparison\n\n")
		for _, c := range checks {
			if c.local == "" || c.remote == "" {
				continue
			}

			sb.WriteString(fmt.Sprintf("### %s\n\n", c.name))

			// Run rclone check
			args := []string{"check", c.local, c.remote, "--one-way", "--size-only"}
			output, err := runRclone(ctx, args...)
			if err != nil {
				sb.WriteString(fmt.Sprintf("```\n%s\n```\n\n", output))
			} else {
				sb.WriteString("✅ All files match\n\n")
			}
		}
	}

	sb.WriteString("\n## Actions\n")
	sb.WriteString("- Use `aftrs_archive_sync_dj` to sync DJ archive\n")
	sb.WriteString("- Use `aftrs_archive_sync_vj` to sync VJ archive\n")

	return mcp.NewToolResultText(sb.String()), nil
}
