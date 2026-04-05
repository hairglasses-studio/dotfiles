// Package data_migration provides MCP tools for drive scanning, data migration, and cloud sync.
package data_migration

import (
	"context"
	"fmt"
	"strings"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

var getClient = tools.LazyClient(clients.NewDataMigrationClient)

// Module implements the data migration tools module
type Module struct{}

// Name returns the module name
func (m *Module) Name() string {
	return "data_migration"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Drive scanning, data migration, deduplication, and cloud sync tools for NVME/UNRAID backup"
}

// Tools returns all tool definitions for this module
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		// Drive Scanning
		{
			Tool: mcp.NewTool("aftrs_migration_scan_drive",
				mcp.WithDescription("Scan a connected drive and categorize its content for migration. Returns folder list with sizes, categories, and target paths. Automatically excludes Kevin Archive, system folders, and movie content."),
				mcp.WithString("drive_path",
					mcp.Required(),
					mcp.Description("Drive letter or mount point (e.g., 'E:', 'E:\\', '/mnt/recovery')"),
				),
			),
			Handler:             handleScanDrive,
			Category:            "data_migration",
			Subcategory:         "scanning",
			Tags:                []string{"migration", "scan", "drive", "inventory", "categorize"},
			UseCases:            []string{"Scan external drive for content", "Categorize files for migration", "Pre-migration analysis"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "data_migration",
			IsWrite:             false,
		},
		{
			Tool: mcp.NewTool("aftrs_migration_drive_report",
				mcp.WithDescription("Generate a comprehensive markdown report for a drive including content breakdown, exclusions, and sync recommendations."),
				mcp.WithString("drive_path",
					mcp.Required(),
					mcp.Description("Drive letter or mount point"),
				),
			),
			Handler:             handleDriveReport,
			Category:            "data_migration",
			Subcategory:         "reporting",
			Tags:                []string{"migration", "report", "drive", "analysis"},
			UseCases:            []string{"Generate migration report", "Document drive contents", "Plan migration strategy"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "data_migration",
			IsWrite:             false,
		},
		// WSL2 Linux Drive Mounting
		{
			Tool: mcp.NewTool("aftrs_migration_list_disks",
				mcp.WithDescription("List physical USB/SATA disks available for WSL2 mounting. Use this to find disk numbers for BTRFS/XFS drives from UNRAID."),
			),
			Handler:             handleListDisks,
			Category:            "data_migration",
			Subcategory:         "wsl2",
			Tags:                []string{"migration", "disks", "wsl2", "btrfs", "xfs", "unraid"},
			UseCases:            []string{"List available disks for mounting", "Find UNRAID drive disk number"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "data_migration",
			IsWrite:             false,
		},
		{
			Tool: mcp.NewTool("aftrs_migration_mount_linux_drive",
				mcp.WithDescription("Mount a BTRFS or XFS drive from UNRAID via WSL2. Requires Windows 11 and WSL2 installed."),
				mcp.WithNumber("disk_number",
					mcp.Required(),
					mcp.Description("Physical disk number from list_disks (e.g., 2)"),
				),
				mcp.WithString("filesystem_type",
					mcp.Required(),
					mcp.Description("Filesystem type: 'btrfs' or 'xfs'"),
				),
				mcp.WithString("mount_point",
					mcp.Description("Mount point in WSL (default: /mnt/recovery)"),
				),
			),
			Handler:             handleMountLinuxDrive,
			Category:            "data_migration",
			Subcategory:         "wsl2",
			Tags:                []string{"migration", "mount", "wsl2", "btrfs", "xfs", "unraid"},
			UseCases:            []string{"Mount UNRAID cache drive", "Mount BTRFS drive on Windows", "Access XFS filesystem"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "data_migration",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_migration_mount_encrypted_drive",
				mcp.WithDescription("Mount a LUKS-encrypted drive from UNRAID via WSL2. Requires keyfile for decryption."),
				mcp.WithNumber("disk_number",
					mcp.Required(),
					mcp.Description("Physical disk number"),
				),
				mcp.WithString("keyfile_path",
					mcp.Required(),
					mcp.Description("Path to LUKS keyfile (Windows path, e.g., 'C:\\keys\\unraid.key')"),
				),
				mcp.WithString("mount_point",
					mcp.Description("Mount point in WSL (default: /mnt/recovery)"),
				),
			),
			Handler:             handleMountEncryptedDrive,
			Category:            "data_migration",
			Subcategory:         "wsl2",
			Tags:                []string{"migration", "mount", "wsl2", "luks", "encrypted", "unraid"},
			UseCases:            []string{"Mount encrypted UNRAID drive", "Recover data from encrypted disk"},
			Complexity:          tools.ComplexityComplex,
			CircuitBreakerGroup: "data_migration",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_migration_unmount_drive",
				mcp.WithDescription("Safely unmount a WSL2-mounted drive. Always unmount before disconnecting to prevent data corruption."),
				mcp.WithNumber("disk_number",
					mcp.Required(),
					mcp.Description("Physical disk number to unmount"),
				),
			),
			Handler:             handleUnmountDrive,
			Category:            "data_migration",
			Subcategory:         "wsl2",
			Tags:                []string{"migration", "unmount", "wsl2", "safety"},
			UseCases:            []string{"Safely disconnect drive", "Unmount before removal"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "data_migration",
			IsWrite:             true,
		},
		// Deduplication
		{
			Tool: mcp.NewTool("aftrs_migration_hash_index",
				mcp.WithDescription("Build a SHA256 hash index for deduplication. Indexes files for later duplicate detection."),
				mcp.WithString("source_path",
					mcp.Required(),
					mcp.Description("Path to index"),
				),
				mcp.WithString("index_name",
					mcp.Description("Name for the index (default: auto-generated)"),
				),
				mcp.WithString("file_types",
					mcp.Description("Comma-separated file extensions to index (e.g., '.mov,.mp4,.avi'). Leave empty for all files."),
				),
				mcp.WithNumber("min_size_mb",
					mcp.Description("Minimum file size in MB to index (default: 1)"),
				),
			),
			Handler:             handleHashIndex,
			Category:            "data_migration",
			Subcategory:         "deduplication",
			Tags:                []string{"migration", "hash", "index", "deduplication", "sha256"},
			UseCases:            []string{"Build dedup index", "Prepare for duplicate detection", "Index large files"},
			Complexity:          tools.ComplexityComplex,
			CircuitBreakerGroup: "data_migration",
			IsWrite:             false,
		},
		{
			Tool: mcp.NewTool("aftrs_migration_find_duplicates",
				mcp.WithDescription("Find duplicates between a local source and cloud destination. Reports unique files, duplicates, and conflicts."),
				mcp.WithString("source_path",
					mcp.Required(),
					mcp.Description("Local source path to check"),
				),
				mcp.WithString("destination_remote",
					mcp.Required(),
					mcp.Description("Rclone remote destination (e.g., 'gdrive:Visual Production')"),
				),
				mcp.WithBoolean("use_hash",
					mcp.Description("Use SHA256 hash comparison instead of size-only (slower but more accurate, default: false)"),
				),
			),
			Handler:             handleFindDuplicates,
			Category:            "data_migration",
			Subcategory:         "deduplication",
			Tags:                []string{"migration", "duplicates", "compare", "deduplication"},
			UseCases:            []string{"Find files to skip", "Calculate space savings", "Pre-sync duplicate check"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "data_migration",
			IsWrite:             false,
		},
		// Sync Operations
		{
			Tool: mcp.NewTool("aftrs_migration_sync_category",
				mcp.WithDescription("Sync a folder to its target category in Google Drive. Automatically applies exclusion patterns for Kevin Archive, movies, and system files."),
				mcp.WithString("source_path",
					mcp.Required(),
					mcp.Description("Local source path to sync"),
				),
				mcp.WithString("category",
					mcp.Required(),
					mcp.Description("Target category (e.g., 'Visual Production', 'Music Production', 'ROMs')"),
				),
				mcp.WithBoolean("dry_run",
					mcp.Description("Preview only, don't actually sync (default: true for safety)"),
				),
			),
			Handler:             handleSyncCategory,
			Category:            "data_migration",
			Subcategory:         "sync",
			Tags:                []string{"migration", "sync", "gdrive", "category"},
			UseCases:            []string{"Sync VJ clips to cloud", "Migrate music production files", "Backup ROMs"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "data_migration",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_migration_status",
				mcp.WithDescription("Check status of migration jobs including progress, transfer speed, and ETA."),
				mcp.WithString("job_id",
					mcp.Description("Specific job ID to check (leave empty for all active jobs)"),
				),
			),
			Handler:             handleMigrationStatus,
			Category:            "data_migration",
			Subcategory:         "status",
			Tags:                []string{"migration", "status", "progress", "jobs"},
			UseCases:            []string{"Monitor sync progress", "Check job status", "View transfer speeds"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "data_migration",
			IsWrite:             false,
		},
		// UNRAID Disaster Recovery Tools
		{
			Tool: mcp.NewTool("aftrs_unraid_scan_drives",
				mcp.WithDescription("Scan for UNRAID array drives connected to this system. Detects XFS/BTRFS/LUKS drives that may be from an UNRAID server."),
			),
			Handler:             handleUNRAIDScanDrives,
			Category:            "data_migration",
			Subcategory:         "unraid",
			Tags:                []string{"unraid", "drives", "scan", "recovery", "xfs", "btrfs"},
			UseCases:            []string{"Find UNRAID drives for recovery", "Detect array disks", "Disaster recovery"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "data_migration",
			IsWrite:             false,
		},
		{
			Tool: mcp.NewTool("aftrs_unraid_mount_drive",
				mcp.WithDescription("Mount an UNRAID XFS/BTRFS drive via WSL2 (Windows) or directly (Linux). Supports LUKS encryption."),
				mcp.WithString("device",
					mcp.Required(),
					mcp.Description("Device path (Windows: '\\\\.\\PHYSICALDRIVE1', Linux: '/dev/sdb1')"),
				),
				mcp.WithString("mount_point",
					mcp.Description("Mount point (default: /mnt/unraid)"),
				),
				mcp.WithString("filesystem",
					mcp.Description("Filesystem type: 'auto', 'xfs', 'btrfs' (default: auto-detect)"),
				),
				mcp.WithBoolean("read_only",
					mcp.Description("Mount as read-only for safety (recommended, default: true)"),
				),
				mcp.WithString("decryption_key",
					mcp.Description("Decryption key/passphrase for LUKS encrypted drives"),
				),
			),
			Handler:             handleUNRAIDMountDrive,
			Category:            "data_migration",
			Subcategory:         "unraid",
			Tags:                []string{"unraid", "mount", "xfs", "btrfs", "luks", "wsl2"},
			UseCases:            []string{"Mount UNRAID drive for recovery", "Access XFS filesystem on Windows", "Decrypt LUKS drive"},
			Complexity:          tools.ComplexityComplex,
			CircuitBreakerGroup: "data_migration",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_unraid_unmount_drive",
				mcp.WithDescription("Safely unmount an UNRAID drive. Always unmount before disconnecting physical drives."),
				mcp.WithString("mount_point",
					mcp.Required(),
					mcp.Description("Mount point to unmount (e.g., /mnt/unraid)"),
				),
			),
			Handler:             handleUNRAIDUnmountDrive,
			Category:            "data_migration",
			Subcategory:         "unraid",
			Tags:                []string{"unraid", "unmount", "safety"},
			UseCases:            []string{"Safely unmount drive", "Prepare for disconnect"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "data_migration",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_unraid_scan_appdata",
				mcp.WithDescription("Scan for Docker appdata on a mounted UNRAID drive. Finds container configs for backup."),
				mcp.WithString("mount_point",
					mcp.Required(),
					mcp.Description("Mount point of UNRAID drive (e.g., /mnt/unraid)"),
				),
			),
			Handler:             handleUNRAIDScanAppdata,
			Category:            "data_migration",
			Subcategory:         "unraid",
			Tags:                []string{"unraid", "appdata", "docker", "backup"},
			UseCases:            []string{"Find Docker configs", "Backup appdata", "Disaster recovery"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "data_migration",
			IsWrite:             false,
		},
		{
			Tool: mcp.NewTool("aftrs_unraid_backup_appdata",
				mcp.WithDescription("Backup UNRAID Docker appdata to a destination (local or remote)."),
				mcp.WithString("source",
					mcp.Required(),
					mcp.Description("Source appdata path (e.g., /mnt/unraid/appdata)"),
				),
				mcp.WithString("destination",
					mcp.Required(),
					mcp.Description("Destination path (local or remote:path)"),
				),
			),
			Handler:             handleUNRAIDBackupAppdata,
			Category:            "data_migration",
			Subcategory:         "unraid",
			Tags:                []string{"unraid", "appdata", "backup", "sync"},
			UseCases:            []string{"Backup Docker configs", "Migrate appdata", "Disaster recovery"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "data_migration",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_unraid_repair_filesystem",
				mcp.WithDescription("Check or repair XFS/BTRFS filesystem on UNRAID drive. Use dry_run first to check without modifying."),
				mcp.WithString("device",
					mcp.Required(),
					mcp.Description("Device path to check/repair (e.g., /dev/sdb1)"),
				),
				mcp.WithString("filesystem",
					mcp.Required(),
					mcp.Description("Filesystem type: 'xfs' or 'btrfs'"),
				),
				mcp.WithBoolean("dry_run",
					mcp.Description("If true, check only without repair (recommended first, default: true)"),
				),
			),
			Handler:             handleUNRAIDRepairFilesystem,
			Category:            "data_migration",
			Subcategory:         "unraid",
			Tags:                []string{"unraid", "repair", "xfs", "btrfs", "filesystem"},
			UseCases:            []string{"Check filesystem integrity", "Repair corrupted filesystem", "Disaster recovery"},
			Complexity:          tools.ComplexityComplex,
			CircuitBreakerGroup: "data_migration",
			IsWrite:             true,
		},

		// Cross-Drive Deduplication (Phase 3)
		{
			Tool: mcp.NewTool("aftrs_migration_merge_indexes",
				mcp.WithDescription("Merge multiple hash indexes from different drives into a master index for cross-drive duplicate detection."),
				mcp.WithString("index_ids",
					mcp.Required(),
					mcp.Description("Comma-separated list of hash index IDs to merge"),
				),
				mcp.WithString("name",
					mcp.Description("Name for the master index (optional)"),
				),
			),
			Handler:             handleMergeIndexes,
			Category:            "data_migration",
			Subcategory:         "deduplication",
			Tags:                []string{"dedup", "hash", "index", "merge", "cross-drive"},
			UseCases:            []string{"Combine indexes from multiple drives", "Prepare for cross-drive duplicate detection"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "data_migration",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_migration_cross_drive_dupes",
				mcp.WithDescription("Find files that exist on multiple drives using a merged master index."),
				mcp.WithString("master_index_id",
					mcp.Required(),
					mcp.Description("ID of the master index created from merged indexes"),
				),
				mcp.WithNumber("limit",
					mcp.Description("Maximum number of duplicates to return (default: 50)"),
				),
			),
			Handler:             handleCrossDriveDupes,
			Category:            "data_migration",
			Subcategory:         "deduplication",
			Tags:                []string{"dedup", "duplicates", "cross-drive", "analysis"},
			UseCases:            []string{"Find files duplicated across drives", "Identify space savings opportunities"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "data_migration",
			IsWrite:             false,
		},
		{
			Tool: mcp.NewTool("aftrs_migration_dedup_report",
				mcp.WithDescription("Generate a comprehensive deduplication report showing space savings, per-drive stats, and top duplicates."),
				mcp.WithString("master_index_id",
					mcp.Required(),
					mcp.Description("ID of the master index"),
				),
			),
			Handler:             handleDedupReport,
			Category:            "data_migration",
			Subcategory:         "deduplication",
			Tags:                []string{"dedup", "report", "analysis", "statistics"},
			UseCases:            []string{"Understand duplicate file distribution", "Calculate potential space savings", "Plan deduplication strategy"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "data_migration",
			IsWrite:             false,
		},
		{
			Tool: mcp.NewTool("aftrs_migration_update_index",
				mcp.WithDescription("Incrementally update an existing hash index with parallel hashing. Only rehashes new or modified files."),
				mcp.WithString("path",
					mcp.Required(),
					mcp.Description("Path to scan for files"),
				),
				mcp.WithString("existing_index_id",
					mcp.Description("ID of existing index to update (creates new if not provided)"),
				),
				mcp.WithBoolean("check_mod_time",
					mcp.Description("Skip files with unchanged modification time (default: true)"),
				),
				mcp.WithBoolean("new_files_only",
					mcp.Description("Only hash files not in existing index"),
				),
				mcp.WithNumber("workers",
					mcp.Description("Number of parallel hash workers (default: 4, max: 16)"),
				),
				mcp.WithString("file_types",
					mcp.Description("Comma-separated file extensions to include (e.g., 'mp4,mov,avi')"),
				),
				mcp.WithNumber("min_size_mb",
					mcp.Description("Minimum file size in MB (default: 1)"),
				),
			),
			Handler:             handleUpdateIndex,
			Category:            "data_migration",
			Subcategory:         "deduplication",
			Tags:                []string{"hash", "index", "incremental", "parallel"},
			UseCases:            []string{"Update index after adding files", "Fast re-index with parallel hashing"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "data_migration",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_migration_list_master_indexes",
				mcp.WithDescription("List all available master indexes created from merged hash indexes."),
			),
			Handler:             handleListMasterIndexes,
			Category:            "data_migration",
			Subcategory:         "deduplication",
			Tags:                []string{"index", "list", "master"},
			UseCases:            []string{"View available master indexes", "Find index IDs for dedup operations"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "data_migration",
			IsWrite:             false,
		},
	}
}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Handler implementations

func handleScanDrive(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	drivePath, errResult := tools.RequireStringParam(req, "drive_path")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	driveInfo, err := client.ScanDrive(ctx, drivePath)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Format output
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Drive Scan: %s\n\n", drivePath))
	sb.WriteString(fmt.Sprintf("**Label:** %s | **Filesystem:** %s\n", driveInfo.Label, driveInfo.FileSystem))
	sb.WriteString(fmt.Sprintf("**Size:** %s used of %s (%s free)\n",
		clients.FormatBytes(driveInfo.UsedBytes),
		clients.FormatBytes(driveInfo.TotalBytes),
		clients.FormatBytes(driveInfo.FreeBytes)))
	sb.WriteString(fmt.Sprintf("**Scan Time:** %s\n\n", driveInfo.ScanTime))

	sb.WriteString("### Categories\n\n")
	sb.WriteString("| Category | Size | Files | Target |\n")
	sb.WriteString("|----------|------|-------|--------|\n")

	var totalSync int64
	for _, cat := range driveInfo.Categories {
		sb.WriteString(fmt.Sprintf("| %s | %s | %d | `%s` |\n",
			cat.Category, clients.FormatBytes(cat.TotalSize), cat.FileCount, cat.TargetPath))
		totalSync += cat.TotalSize
	}

	sb.WriteString(fmt.Sprintf("\n**Total to sync:** %s\n", clients.FormatBytes(totalSync)))

	if driveInfo.ExcludedCount > 0 {
		sb.WriteString(fmt.Sprintf("\n**Excluded:** %d folders (%s)\n",
			driveInfo.ExcludedCount, clients.FormatBytes(driveInfo.ExcludedSize)))
	}

	return tools.TextResult(sb.String()), nil
}

func handleDriveReport(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	drivePath, errResult := tools.RequireStringParam(req, "drive_path")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	report, err := client.GenerateDriveReport(ctx, drivePath)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(report), nil
}

func handleListDisks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	disks, err := client.ListPhysicalDisks(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("## Physical Disks (USB/SATA)\n\n")
	sb.WriteString("| Disk # | Name | Size | Bus | Status |\n")
	sb.WriteString("|--------|------|------|-----|--------|\n")

	for _, disk := range disks {
		status := "Offline"
		if disk.IsOnline {
			status = "Online"
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s |\n",
			disk.Number, disk.FriendlyName, clients.FormatBytes(disk.SizeBytes), disk.BusType, status))
	}

	sb.WriteString("\nUse `aftrs_migration_mount_linux_drive` with the disk number to mount BTRFS/XFS drives.\n")

	return tools.TextResult(sb.String()), nil
}

func handleMountLinuxDrive(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	diskNum := tools.GetIntParam(req, "disk_number", -1)
	if diskNum < 0 {
		return tools.ErrorResult(fmt.Errorf("disk_number is required")), nil
	}

	fsType, errResult := tools.RequireStringParam(req, "filesystem_type")
	if errResult != nil {
		return errResult, nil
	}

	mountPoint := tools.GetStringParam(req, "mount_point")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.MountLinuxDrive(ctx, diskNum, fsType, mountPoint); err != nil {
		return tools.ErrorResult(err), nil
	}

	if mountPoint == "" {
		mountPoint = "/mnt/recovery"
	}

	return tools.TextResult(fmt.Sprintf("Disk %d mounted successfully at `%s`\n\nAccess via WSL: `wsl ls %s`\nOr via Windows: `\\\\wsl$\\Ubuntu%s`",
		diskNum, mountPoint, mountPoint, mountPoint)), nil
}

func handleMountEncryptedDrive(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	diskNum := tools.GetIntParam(req, "disk_number", -1)
	if diskNum < 0 {
		return tools.ErrorResult(fmt.Errorf("disk_number is required")), nil
	}

	keyfilePath, errResult := tools.RequireStringParam(req, "keyfile_path")
	if errResult != nil {
		return errResult, nil
	}

	mountPoint := tools.GetStringParam(req, "mount_point")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.MountEncryptedDrive(ctx, diskNum, keyfilePath, mountPoint); err != nil {
		return tools.ErrorResult(err), nil
	}

	if mountPoint == "" {
		mountPoint = "/mnt/recovery"
	}

	return tools.TextResult(fmt.Sprintf("Encrypted disk %d decrypted and mounted at `%s`", diskNum, mountPoint)), nil
}

func handleUnmountDrive(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	diskNum := tools.GetIntParam(req, "disk_number", -1)
	if diskNum < 0 {
		return tools.ErrorResult(fmt.Errorf("disk_number is required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.UnmountLinuxDrive(ctx, diskNum); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Disk %d unmounted successfully. Safe to disconnect.", diskNum)), nil
}

func handleHashIndex(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sourcePath, errResult := tools.RequireStringParam(req, "source_path")
	if errResult != nil {
		return errResult, nil
	}

	indexName := tools.GetStringParam(req, "index_name")
	fileTypesStr := tools.GetStringParam(req, "file_types")
	minSizeMB := tools.GetIntParam(req, "min_size_mb", 1)

	var fileTypes []string
	if fileTypesStr != "" {
		fileTypes = strings.Split(fileTypesStr, ",")
		for i := range fileTypes {
			fileTypes[i] = strings.TrimSpace(fileTypes[i])
		}
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	index, err := client.BuildHashIndex(ctx, sourcePath, indexName, fileTypes, minSizeMB)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Count duplicates
	duplicateGroups := 0
	for _, files := range index.Files {
		if len(files) > 1 {
			duplicateGroups++
		}
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Hash Index Created\n\n"))
	sb.WriteString(fmt.Sprintf("**ID:** `%s`\n", index.ID))
	sb.WriteString(fmt.Sprintf("**Path:** %s\n", index.Path))
	sb.WriteString(fmt.Sprintf("**Files Indexed:** %d\n", index.FileCount))
	sb.WriteString(fmt.Sprintf("**Total Size:** %s\n", clients.FormatBytes(index.TotalSize)))
	sb.WriteString(fmt.Sprintf("**Duplicate Groups:** %d\n", duplicateGroups))

	return tools.TextResult(sb.String()), nil
}

func handleFindDuplicates(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sourcePath, errResult := tools.RequireStringParam(req, "source_path")
	if errResult != nil {
		return errResult, nil
	}

	destRemote, errResult := tools.RequireStringParam(req, "destination_remote")
	if errResult != nil {
		return errResult, nil
	}

	useHash := tools.GetBoolParam(req, "use_hash", false)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	report, err := client.FindDuplicates(ctx, sourcePath, destRemote, useHash)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("## Duplicate Analysis\n\n")
	sb.WriteString(fmt.Sprintf("**Unique Files:** %d (%s)\n", len(report.UniqueFiles), clients.FormatBytes(report.UniqueSize)))
	sb.WriteString(fmt.Sprintf("**Already Synced:** %d (%s)\n", len(report.DuplicateFiles), clients.FormatBytes(report.DuplicateSize)))
	sb.WriteString(fmt.Sprintf("**Conflicts:** %d\n", len(report.Conflicts)))
	sb.WriteString(fmt.Sprintf("\n**Space Savings (skip duplicates):** %s\n", clients.FormatBytes(report.SpaceSavings)))

	return tools.TextResult(sb.String()), nil
}

func handleSyncCategory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sourcePath, errResult := tools.RequireStringParam(req, "source_path")
	if errResult != nil {
		return errResult, nil
	}

	category, errResult := tools.RequireStringParam(req, "category")
	if errResult != nil {
		return errResult, nil
	}

	dryRun := tools.GetBoolParam(req, "dry_run", true)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	job, err := client.SyncCategory(ctx, sourcePath, category, dryRun, true)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	modeStr := "DRY RUN"
	if !dryRun {
		modeStr = "LIVE"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Sync Started (%s)\n\n", modeStr))
	sb.WriteString(fmt.Sprintf("**Job ID:** `%s`\n", job.ID))
	sb.WriteString(fmt.Sprintf("**Source:** %s\n", job.Source))
	sb.WriteString(fmt.Sprintf("**Destination:** %s\n", job.Destination))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n", job.Status))
	sb.WriteString("\nUse `aftrs_migration_status` with the job ID to monitor progress.\n")

	return tools.TextResult(sb.String()), nil
}

func handleMigrationStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	jobID := tools.GetStringParam(req, "job_id")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder

	if jobID != "" {
		// Get specific job
		job, err := client.GetJobStatus(jobID)
		if err != nil {
			return tools.ErrorResult(err), nil
		}

		sb.WriteString(fmt.Sprintf("## Job Status: %s\n\n", job.ID))
		sb.WriteString(fmt.Sprintf("**Type:** %s\n", job.Type))
		sb.WriteString(fmt.Sprintf("**Status:** %s\n", job.Status))
		sb.WriteString(fmt.Sprintf("**Source:** %s\n", job.Source))
		sb.WriteString(fmt.Sprintf("**Destination:** %s\n", job.Destination))

		if job.Progress != nil {
			sb.WriteString(fmt.Sprintf("\n### Progress\n"))
			sb.WriteString(fmt.Sprintf("- **Files:** %d / %d\n", job.Progress.FilesTransferred, job.Progress.FilesTotal))
			sb.WriteString(fmt.Sprintf("- **Bytes:** %s / %s (%.1f%%)\n",
				clients.FormatBytes(job.Progress.BytesTransferred),
				clients.FormatBytes(job.Progress.BytesTotal),
				job.Progress.PercentComplete))
			sb.WriteString(fmt.Sprintf("- **Speed:** %s\n", job.Progress.TransferSpeed))
			sb.WriteString(fmt.Sprintf("- **ETA:** %s\n", job.Progress.ETA))
		}

		if job.Error != "" {
			sb.WriteString(fmt.Sprintf("\n**Error:** %s\n", job.Error))
		}
	} else {
		// List all active jobs
		jobs := client.ListActiveJobs()
		if len(jobs) == 0 {
			sb.WriteString("No active migration jobs.\n")
		} else {
			sb.WriteString("## Active Migration Jobs\n\n")
			sb.WriteString("| ID | Type | Status | Source | Progress |\n")
			sb.WriteString("|----|------|--------|--------|----------|\n")

			for _, job := range jobs {
				progress := "-"
				if job.Progress != nil {
					progress = fmt.Sprintf("%.1f%%", job.Progress.PercentComplete)
				}
				sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
					job.ID[:8], job.Type, job.Status, job.Source, progress))
			}
		}
	}

	return tools.TextResult(sb.String()), nil
}

// UNRAID Handler Functions

func handleUNRAIDScanDrives(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	drives, err := client.ScanUNRAIDDrives(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to scan UNRAID drives: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("## UNRAID Drives Detected\n\n")

	if len(drives) == 0 {
		sb.WriteString("No UNRAID drives detected. Make sure drives are connected and WSL2 is available.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("| Device | Size | Filesystem | Array Role | Mounted | Accessible |\n")
	sb.WriteString("|--------|------|------------|------------|---------|------------|\n")

	for _, drive := range drives {
		mounted := "No"
		if drive.IsMounted {
			mounted = "Yes"
		}
		accessible := "No"
		if drive.IsAccessible {
			accessible = "Yes"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
			drive.Device,
			clients.FormatBytes(drive.SizeBytes),
			drive.FileSystem,
			drive.ArrayRole,
			mounted,
			accessible))
	}

	sb.WriteString(fmt.Sprintf("\n**Total drives:** %d\n", len(drives)))
	sb.WriteString("\nUse `aftrs_unraid_mount_drive` to mount a drive for data recovery.\n")

	return tools.TextResult(sb.String()), nil
}

func handleUNRAIDMountDrive(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	device, errResult := tools.RequireStringParam(req, "device")
	if errResult != nil {
		return errResult, nil
	}

	mountPoint := tools.GetStringParam(req, "mount_point")
	fileSystem := tools.GetStringParam(req, "file_system")
	readOnly := tools.GetBoolParam(req, "read_only", true)
	decryptionKey := tools.GetStringParam(req, "decryption_key")
	if mountPoint == "" {
		mountPoint = "/mnt/unraid_recovery"
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	opts := clients.UNRAIDRecoveryOptions{
		DriveDevice:   device,
		MountPoint:    mountPoint,
		FileSystem:    fileSystem,
		ReadOnly:      readOnly,
		DecryptionKey: decryptionKey,
	}

	driveInfo, err := client.MountUNRAIDDrive(ctx, opts)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to mount UNRAID drive: %w", err)), nil
	}
	_ = driveInfo // Use for additional info if needed

	var sb strings.Builder
	sb.WriteString("## UNRAID Drive Mounted\n\n")
	sb.WriteString(fmt.Sprintf("**Device:** %s\n", device))
	sb.WriteString(fmt.Sprintf("**Mount Point:** %s\n", mountPoint))
	sb.WriteString(fmt.Sprintf("**Filesystem:** %s\n", fileSystem))
	if readOnly {
		sb.WriteString("**Mode:** Read-only (safe for recovery)\n")
	} else {
		sb.WriteString("**Mode:** Read-write\n")
	}
	sb.WriteString("\nDrive is now accessible. Use `aftrs_unraid_scan_appdata` to find Docker app data.\n")

	return tools.TextResult(sb.String()), nil
}

func handleUNRAIDUnmountDrive(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	mountPoint, errResult := tools.RequireStringParam(req, "mount_point")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.UnmountUNRAIDDrive(ctx, mountPoint)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to unmount UNRAID drive: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("## UNRAID Drive Unmounted\n\n")
	sb.WriteString(fmt.Sprintf("**Mount Point:** %s\n", mountPoint))
	sb.WriteString("\nDrive has been safely unmounted. It can now be disconnected.\n")

	return tools.TextResult(sb.String()), nil
}

func handleUNRAIDScanAppdata(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path := tools.GetStringParam(req, "path")

	if path == "" {
		path = "/mnt/unraid_recovery/appdata"
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	apps, err := client.ScanUNRAIDAppdata(ctx, path)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to scan appdata: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("## UNRAID Docker Appdata\n\n")

	if len(apps) == 0 {
		sb.WriteString("No Docker app data found at the specified path.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("| Application | Size | Files |\n")
	sb.WriteString("|-------------|------|-------|\n")

	var totalSize int64
	var totalFiles int
	for _, app := range apps {
		sb.WriteString(fmt.Sprintf("| %s | %s | %d |\n",
			app.ContainerName,
			clients.FormatBytes(app.SizeBytes),
			app.FileCount))
		totalSize += app.SizeBytes
		totalFiles += app.FileCount
	}

	sb.WriteString(fmt.Sprintf("\n**Total:** %d applications, %s, %d files\n",
		len(apps), clients.FormatBytes(totalSize), totalFiles))
	sb.WriteString("\nUse `aftrs_unraid_backup_appdata` to backup specific applications.\n")

	return tools.TextResult(sb.String()), nil
}

func handleUNRAIDBackupAppdata(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sourcePath, errResult := tools.RequireStringParam(req, "source_path")
	if errResult != nil {
		return errResult, nil
	}

	destPath, errResult := tools.RequireStringParam(req, "dest_path")
	if errResult != nil {
		return errResult, nil
	}

	appsParam := tools.GetStringParam(req, "apps")

	var apps []string
	if appsParam != "" {
		apps = strings.Split(appsParam, ",")
		for i := range apps {
			apps[i] = strings.TrimSpace(apps[i])
		}
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	job, err := client.BackupUNRAIDAppdata(ctx, sourcePath, destPath, apps)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to start appdata backup: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("## UNRAID Appdata Backup Started\n\n")
	sb.WriteString(fmt.Sprintf("**Job ID:** `%s`\n", job.ID))
	sb.WriteString(fmt.Sprintf("**Source:** %s\n", sourcePath))
	sb.WriteString(fmt.Sprintf("**Destination:** %s\n", destPath))
	if len(apps) > 0 {
		sb.WriteString(fmt.Sprintf("**Applications:** %s\n", strings.Join(apps, ", ")))
	} else {
		sb.WriteString("**Applications:** All\n")
	}
	sb.WriteString("\nUse `aftrs_migration_status` with the job ID to monitor progress.\n")

	return tools.TextResult(sb.String()), nil
}

func handleUNRAIDRepairFilesystem(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	device, errResult := tools.RequireStringParam(req, "device")
	if errResult != nil {
		return errResult, nil
	}

	fileSystem := tools.GetStringParam(req, "file_system")
	dryRun := tools.GetBoolParam(req, "dry_run", true)

	if fileSystem == "" {
		return tools.ErrorResult(fmt.Errorf("file_system is required (xfs or btrfs)")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var output string
	switch strings.ToLower(fileSystem) {
	case "xfs":
		output, err = client.RepairXFS(ctx, device, dryRun)
	case "btrfs":
		output, err = client.RepairBTRFS(ctx, device, dryRun)
	default:
		return tools.ErrorResult(fmt.Errorf("unsupported filesystem: %s (use xfs or btrfs)", fileSystem)), nil
	}

	if err != nil {
		return tools.ErrorResult(fmt.Errorf("filesystem repair failed: %w", err)), nil
	}

	var sb strings.Builder
	if dryRun {
		sb.WriteString("## Filesystem Check (Dry Run)\n\n")
	} else {
		sb.WriteString("## Filesystem Repair Complete\n\n")
	}
	sb.WriteString(fmt.Sprintf("**Device:** %s\n", device))
	sb.WriteString(fmt.Sprintf("**Filesystem:** %s\n", fileSystem))
	sb.WriteString("\n### Output\n```\n")
	sb.WriteString(output)
	sb.WriteString("\n```\n")

	if dryRun {
		sb.WriteString("\nThis was a read-only check. Set `dry_run: false` to perform actual repairs.\n")
	}

	return tools.TextResult(sb.String()), nil
}

// Cross-Drive Deduplication Handlers (Phase 3)

func handleMergeIndexes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	indexIDsStr, errResult := tools.RequireStringParam(req, "index_ids")
	if errResult != nil {
		return errResult, nil
	}

	name := tools.GetStringParam(req, "name")

	// Parse comma-separated index IDs
	indexIDs := strings.Split(indexIDsStr, ",")
	for i := range indexIDs {
		indexIDs[i] = strings.TrimSpace(indexIDs[i])
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	master, err := client.MergeHashIndexes(ctx, indexIDs, name)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to merge indexes: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("## Master Index Created\n\n")
	sb.WriteString(fmt.Sprintf("**ID:** `%s`\n", master.ID))
	sb.WriteString(fmt.Sprintf("**Name:** %s\n", master.Name))
	sb.WriteString(fmt.Sprintf("**Sources:** %d indexes merged\n", len(master.Sources)))
	sb.WriteString(fmt.Sprintf("**Total Files:** %d\n", master.TotalFiles))
	sb.WriteString(fmt.Sprintf("**Total Size:** %s\n", clients.FormatBytes(master.TotalSize)))
	sb.WriteString(fmt.Sprintf("**Unique Files:** %d (%s)\n", master.UniqueFiles, clients.FormatBytes(master.UniqueSize)))
	sb.WriteString(fmt.Sprintf("**Duplicate Files:** %d\n", master.DupeFiles))
	sb.WriteString(fmt.Sprintf("**Wasted Space:** %s\n", clients.FormatBytes(master.DupeSpace)))
	sb.WriteString("\nUse `aftrs_migration_cross_drive_dupes` or `aftrs_migration_dedup_report` with this master index ID.\n")

	return tools.TextResult(sb.String()), nil
}

func handleCrossDriveDupes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	masterIndexID, errResult := tools.RequireStringParam(req, "master_index_id")
	if errResult != nil {
		return errResult, nil
	}

	limit := tools.GetIntParam(req, "limit", 50)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	duplicates, err := client.FindCrossDriveDuplicates(ctx, masterIndexID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to find duplicates: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("## Cross-Drive Duplicates\n\n")

	if len(duplicates) == 0 {
		sb.WriteString("No cross-drive duplicates found.\n")
		return tools.TextResult(sb.String()), nil
	}

	// Calculate total wasted space
	var totalWasted int64
	for _, d := range duplicates {
		totalWasted += d.WastedSpace
	}

	sb.WriteString(fmt.Sprintf("**Total Duplicates:** %d file groups\n", len(duplicates)))
	sb.WriteString(fmt.Sprintf("**Total Wasted Space:** %s\n\n", clients.FormatBytes(totalWasted)))

	// Limit results
	shown := duplicates
	if len(shown) > limit {
		shown = shown[:limit]
	}

	sb.WriteString("| Size | Drives | Wasted | Sample Path |\n")
	sb.WriteString("|------|--------|--------|-------------|\n")

	for _, d := range shown {
		samplePath := d.Files[0].Path
		if len(samplePath) > 50 {
			samplePath = "..." + samplePath[len(samplePath)-47:]
		}
		sb.WriteString(fmt.Sprintf("| %s | %d | %s | %s |\n",
			d.SizeHuman,
			d.DriveCount,
			clients.FormatBytes(d.WastedSpace),
			samplePath))
	}

	if len(duplicates) > limit {
		sb.WriteString(fmt.Sprintf("\n*Showing %d of %d duplicates*\n", limit, len(duplicates)))
	}

	return tools.TextResult(sb.String()), nil
}

func handleDedupReport(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	masterIndexID, errResult := tools.RequireStringParam(req, "master_index_id")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	report, err := client.GenerateDeduplicationReport(ctx, masterIndexID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to generate report: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("## Deduplication Report\n\n")
	sb.WriteString(fmt.Sprintf("**Master Index:** `%s`\n", report.MasterIndexID))
	sb.WriteString(fmt.Sprintf("**Sources:** %d drives\n", report.SourceCount))
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", report.GeneratedAt.Format("2006-01-02 15:04:05")))

	sb.WriteString("### Summary\n\n")
	sb.WriteString(fmt.Sprintf("| Metric | Value |\n"))
	sb.WriteString(fmt.Sprintf("|--------|-------|\n"))
	sb.WriteString(fmt.Sprintf("| Total Files | %d |\n", report.TotalFiles))
	sb.WriteString(fmt.Sprintf("| Total Size | %s |\n", report.TotalSizeHuman))
	sb.WriteString(fmt.Sprintf("| Unique Files | %d |\n", report.UniqueFiles))
	sb.WriteString(fmt.Sprintf("| Unique Size | %s |\n", report.UniqueSizeHuman))
	sb.WriteString(fmt.Sprintf("| Duplicate Groups | %d |\n", report.DuplicateFiles))
	sb.WriteString(fmt.Sprintf("| Wasted Space | %s |\n", report.WastedSpaceHuman))
	sb.WriteString(fmt.Sprintf("| **Savings Potential** | **%.1f%%** |\n", report.SavingsPercent))

	sb.WriteString("\n### Per-Drive Statistics\n\n")
	sb.WriteString("| Drive | Files | Size | Unique | Shared |\n")
	sb.WriteString("|-------|-------|------|--------|--------|\n")
	for _, ds := range report.ByDrive {
		sb.WriteString(fmt.Sprintf("| %s | %d | %s | %d | %d |\n",
			ds.Drive,
			ds.TotalFiles,
			clients.FormatBytes(ds.TotalSize),
			ds.UniqueFiles,
			ds.SharedFiles))
	}

	if len(report.TopDuplicates) > 0 {
		sb.WriteString("\n### Top Duplicates (by wasted space)\n\n")
		sb.WriteString("| Size | Drives | Wasted | Sample Path |\n")
		sb.WriteString("|------|--------|--------|-------------|\n")

		maxShow := 10
		if len(report.TopDuplicates) < maxShow {
			maxShow = len(report.TopDuplicates)
		}
		for i := 0; i < maxShow; i++ {
			d := report.TopDuplicates[i]
			samplePath := d.Files[0].Path
			if len(samplePath) > 40 {
				samplePath = "..." + samplePath[len(samplePath)-37:]
			}
			sb.WriteString(fmt.Sprintf("| %s | %d | %s | %s |\n",
				d.SizeHuman,
				d.DriveCount,
				clients.FormatBytes(d.WastedSpace),
				samplePath))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleUpdateIndex(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, errResult := tools.RequireStringParam(req, "path")
	if errResult != nil {
		return errResult, nil
	}

	existingIndexID := tools.GetStringParam(req, "existing_index_id")
	checkModTime := tools.GetBoolParam(req, "check_mod_time", true)
	newFilesOnly := tools.GetBoolParam(req, "new_files_only", false)
	workers := tools.GetIntParam(req, "workers", 4)
	fileTypesStr := tools.GetStringParam(req, "file_types")
	minSizeMB := tools.GetIntParam(req, "min_size_mb", 1)

	var fileTypes []string
	if fileTypesStr != "" {
		fileTypes = strings.Split(fileTypesStr, ",")
		for i := range fileTypes {
			fileTypes[i] = strings.TrimSpace(fileTypes[i])
		}
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	opts := clients.IncrementalIndexOptions{
		ExistingIndexID: existingIndexID,
		CheckModTime:    checkModTime,
		NewFilesOnly:    newFilesOnly,
		ParallelWorkers: workers,
		FileTypes:       fileTypes,
		MinSizeMB:       minSizeMB,
	}

	index, err := client.UpdateHashIndex(ctx, path, opts)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to update index: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("## Hash Index Updated\n\n")
	sb.WriteString(fmt.Sprintf("**ID:** `%s`\n", index.ID))
	sb.WriteString(fmt.Sprintf("**Name:** %s\n", index.Name))
	sb.WriteString(fmt.Sprintf("**Path:** %s\n", index.Path))
	sb.WriteString(fmt.Sprintf("**Total Files:** %d\n", index.FileCount))
	sb.WriteString(fmt.Sprintf("**Total Size:** %s\n", clients.FormatBytes(index.TotalSize)))
	sb.WriteString(fmt.Sprintf("**Workers Used:** %d\n", workers))

	if existingIndexID != "" {
		sb.WriteString("\n*Incremental update completed*\n")
	} else {
		sb.WriteString("\n*New index created*\n")
	}

	return tools.TextResult(sb.String()), nil
}

func handleListMasterIndexes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	masters, err := client.ListMasterIndexes()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to list master indexes: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("## Master Indexes\n\n")

	if len(masters) == 0 {
		sb.WriteString("No master indexes found.\n")
		sb.WriteString("\nUse `aftrs_migration_merge_indexes` to create a master index from multiple hash indexes.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("| ID | Created |\n")
	sb.WriteString("|----|--------|\n")

	for _, id := range masters {
		// Try to load for more details
		master, err := client.LoadMasterIndex(id)
		if err != nil {
			sb.WriteString(fmt.Sprintf("| `%s` | (error loading) |\n", id))
		} else {
			sb.WriteString(fmt.Sprintf("| `%s` | %s |\n", id, master.CreatedAt.Format("2006-01-02 15:04")))
		}
	}

	sb.WriteString(fmt.Sprintf("\n**Total:** %d master indexes\n", len(masters)))

	return tools.TextResult(sb.String()), nil
}
