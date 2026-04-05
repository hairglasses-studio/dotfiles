// Package clients provides API clients for external services.
package clients

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hairglasses-studio/mcpkit/sanitize"
)

// RcloneClient provides rclone-based cloud sync capabilities
type RcloneClient struct {
	configPath     string
	transfers      int    // number of parallel transfers
	checkers       int    // number of parallel checkers
	bandwidth      string // bandwidth limit (e.g., "10M" for 10MB/s)
	driveChunkSize string // chunk size for Google Drive uploads (default 8M)
	fastList       bool   // use --fast-list for faster listing (20x speed improvement)
	progressChan   chan *SyncProgress
	mu             sync.RWMutex
	activeJobs     map[string]*SyncJob

	// Multi-threading for large files
	multiThreadStreams int    // number of streams for multi-threaded downloads
	multiThreadCutoff  string // file size threshold for multi-threading (e.g., "256M")

	// Rate limiting for Google Drive
	tpsLimit int // transactions per second limit
	tpsBurst int // TPS burst allowance

	// Memory and connection management
	maxBufferMemory string // max memory for transfer buffers
	useCookies      bool   // maintain persistent connections

	// Retry and resilience
	lowLevelRetries int    // retries for individual HTTP requests
	retriesSleep    string // sleep between retries
	resilient       bool   // continue on errors

	// Duration control for checkpointing
	maxDuration string // max duration before auto-stop

	// Bandwidth scheduling
	bandwidthSchedule []BandwidthSchedule

	// Speed tracking for better ETA
	speedHistory []SpeedSample
	speedMu      sync.RWMutex

	// Checkpoint and quota state files
	checkpointDir  string
	quotaStateFile string

	// Comparison mode for deduplication
	compareMode  ComparisonMode
	skipExisting bool   // --ignore-existing: skip files that exist on destination
	updateOnly   bool   // --update: skip files newer on destination
	noTraverse   bool   // --no-traverse: skip destination listing (faster for empty dest)
	orderBy      string // --order-by: file ordering (e.g., "size,desc")

	// Transfer history for analytics
	historyFile     string
	transferHistory []TransferRecord
	historyMu       sync.RWMutex
}

// SyncProfile defines preset configurations for different backup scenarios
type SyncProfile string

const (
	// ProfileDefault uses balanced settings for general use
	ProfileDefault SyncProfile = "default"
	// ProfileLargeFiles optimizes for video/media backup with large files
	ProfileLargeFiles SyncProfile = "large_files"
	// ProfileManySmallFiles optimizes for documents/code with many small files
	ProfileManySmallFiles SyncProfile = "many_small_files"
	// ProfileBackground uses conservative settings to minimize impact
	ProfileBackground SyncProfile = "background"
	// ProfileGoogleDrive optimizes for Google Drive with rate limit protection
	ProfileGoogleDrive SyncProfile = "google_drive"
)

// ComparisonMode defines how files are compared during sync
type ComparisonMode string

const (
	// CompareSizeOnly compares files by size only (fastest)
	CompareSizeOnly ComparisonMode = "size"
	// CompareSizeModTime compares by size and modification time (default, balanced)
	CompareSizeModTime ComparisonMode = "default"
	// CompareChecksum compares by hash/checksum (most accurate, slowest)
	CompareChecksum ComparisonMode = "checksum"
)

// TransferRecord stores completed sync records for history/analytics
type TransferRecord struct {
	JobID            string    `json:"job_id"`
	Source           string    `json:"source"`
	Destination      string    `json:"destination"`
	StartTime        time.Time `json:"start_time"`
	EndTime          time.Time `json:"end_time"`
	BytesTransferred int64     `json:"bytes_transferred"`
	FilesTransferred int       `json:"files_transferred"`
	AverageSpeed     float64   `json:"average_speed_bytes_per_sec"`
	Errors           int       `json:"errors"`
	Status           string    `json:"status"`
}

// SyncAnalysis contains pre-sync analysis results
type SyncAnalysis struct {
	Source          string             `json:"source"`
	Destination     string             `json:"destination"`
	TotalFiles      int                `json:"total_files"`
	FilesToTransfer int                `json:"files_to_transfer"`
	FilesToSkip     int                `json:"files_to_skip"`
	BytesToTransfer int64              `json:"bytes_to_transfer"`
	BytesToSkip     int64              `json:"bytes_to_skip"`
	EstimatedTime   string             `json:"estimated_time"`
	ByExtension     map[string]int64   `json:"by_extension"` // extension -> bytes
	LargestFiles    []FileTransferInfo `json:"largest_files"`
	AnalyzedAt      time.Time          `json:"analyzed_at"`
}

// FileTransferInfo contains info about a file to be transferred
type FileTransferInfo struct {
	Path      string    `json:"path"`
	Size      int64     `json:"size"`
	SizeHuman string    `json:"size_human"`
	ModTime   time.Time `json:"mod_time"`
}

// VerifyResult contains post-sync verification results
type VerifyResult struct {
	Source       string    `json:"source"`
	Destination  string    `json:"destination"`
	Verified     int       `json:"verified"`
	Mismatched   int       `json:"mismatched"`
	Missing      int       `json:"missing"`
	Extra        int       `json:"extra"`
	MismatchList []string  `json:"mismatch_list,omitempty"`
	MissingList  []string  `json:"missing_list,omitempty"`
	ExtraList    []string  `json:"extra_list,omitempty"`
	VerifiedAt   time.Time `json:"verified_at"`
	Duration     string    `json:"duration"`
}

// DuplicateGroup represents a group of duplicate files (uses HashIndex from data_migration.go)
type DuplicateGroup struct {
	Hash       string   `json:"hash"`
	Size       int64    `json:"size"`
	SizeHuman  string   `json:"size_human"`
	Paths      []string `json:"paths"`
	WastedSize int64    `json:"wasted_size"` // (count - 1) * size
}

// SyncDashboard provides comprehensive real-time status
type SyncDashboard struct {
	ActiveJobs    []*SyncJob         `json:"active_jobs"`
	TotalProgress *AggregateProgress `json:"total_progress"`
	SpeedMetrics  *SpeedMetrics      `json:"speed_metrics"`
	QuotaStatus   *DailyQuotaState   `json:"quota_status"`
	RecentErrors  []string           `json:"recent_errors"`
	GeneratedAt   time.Time          `json:"generated_at"`
}

// AggregateProgress combines progress across multiple jobs
type AggregateProgress struct {
	TotalBytes       int64   `json:"total_bytes"`
	TransferredBytes int64   `json:"transferred_bytes"`
	TotalFiles       int     `json:"total_files"`
	TransferredFiles int     `json:"transferred_files"`
	OverallPercent   float64 `json:"overall_percent"`
	CombinedETA      string  `json:"combined_eta"`
}

// SpeedMetrics provides speed statistics
type SpeedMetrics struct {
	CurrentSpeed float64 `json:"current_speed_mbps"`
	AverageSpeed float64 `json:"average_speed_mbps"`
	PeakSpeed    float64 `json:"peak_speed_mbps"`
	SpeedTrend   string  `json:"speed_trend"` // "increasing", "stable", "decreasing"
}

// CheckpointState tracks sync progress for resume capability
type CheckpointState struct {
	JobID          string           `json:"job_id"`
	Source         string           `json:"source"`
	Destination    string           `json:"destination"`
	CompletedFiles map[string]int64 `json:"completed_files"` // path -> size
	FailedFiles    []FailedFile     `json:"failed_files"`
	BytesUploaded  int64            `json:"bytes_uploaded"`
	LastCheckpoint time.Time        `json:"last_checkpoint"`
}

// FailedFile tracks a file that failed transfer
type FailedFile struct {
	Path       string    `json:"path"`
	Size       int64     `json:"size"`
	Error      string    `json:"error"`
	RetryCount int       `json:"retry_count"`
	LastTry    time.Time `json:"last_try"`
}

// DailyQuotaState tracks Google Drive uploads for quota management
type DailyQuotaState struct {
	Date          string    `json:"date"` // YYYY-MM-DD
	BytesUploaded int64     `json:"bytes_uploaded"`
	LastUpdated   time.Time `json:"last_updated"`
}

// SyncOptions extends sync behavior with verification and retry settings
type SyncOptions struct {
	AutoVerify      bool   `json:"auto_verify"`      // Run verification after sync completes
	VerifyMode      string `json:"verify_mode"`      // "size" or "checksum"
	RetryMismatches bool   `json:"retry_mismatches"` // Auto-retry files that fail verification
	MaxRetries      int    `json:"max_retries"`      // Maximum retry attempts (default: 3)
}

// VerifiedSyncResult combines sync result with verification
type VerifiedSyncResult struct {
	Job          *SyncJob      `json:"job"`
	VerifyResult *VerifyResult `json:"verify_result,omitempty"`
	RetryResult  *SyncJob      `json:"retry_result,omitempty"` // If mismatches were retried
	FinalStatus  string        `json:"final_status"`           // "verified", "mismatches", "failed"
}

// FailedFileInfo extends FailedFile with error categorization
type FailedFileInfo struct {
	Path        string    `json:"path"`
	Size        int64     `json:"size"`
	Error       string    `json:"error"`
	ErrorType   string    `json:"error_type"` // "network", "permission", "quota", "corrupt", "timeout", "unknown"
	RetryCount  int       `json:"retry_count"`
	LastAttempt time.Time `json:"last_attempt"`
	Extension   string    `json:"extension"`
}

// FailureAnalysis aggregates failure patterns for a job
type FailureAnalysis struct {
	JobID       string               `json:"job_id"`
	TotalFailed int                  `json:"total_failed"`
	TotalSize   int64                `json:"total_size"`
	ByErrorType map[string]int       `json:"by_error_type"`
	ByExtension map[string]int       `json:"by_extension"`
	Recoverable int                  `json:"recoverable"`
	Suggestions []RecoverySuggestion `json:"suggestions"`
	FailedFiles []FailedFileInfo     `json:"failed_files"`
	AnalyzedAt  time.Time            `json:"analyzed_at"`
}

// RecoverySuggestion provides actionable recovery advice
type RecoverySuggestion struct {
	ErrorType   string `json:"error_type"`
	Count       int    `json:"count"`
	Suggestion  string `json:"suggestion"`
	AutoFixable bool   `json:"auto_fixable"`
	Command     string `json:"command,omitempty"` // Suggested command to fix
}

// SpeedSample for rolling average speed calculation
type SpeedSample struct {
	Timestamp   time.Time `json:"timestamp"`
	BytesPerSec float64   `json:"bytes_per_sec"`
}

// BandwidthSchedule for time-based throttling
type BandwidthSchedule struct {
	StartHour int    `json:"start_hour"`
	EndHour   int    `json:"end_hour"`
	Limit     string `json:"limit"` // e.g., "20M", "off"
}

// SyncJob represents an active or completed sync job
type SyncJob struct {
	ID          string        `json:"id"`
	Source      string        `json:"source"`
	Destination string        `json:"destination"`
	Status      string        `json:"status"` // pending, running, completed, failed, cancelled
	StartTime   time.Time     `json:"start_time"`
	EndTime     *time.Time    `json:"end_time,omitempty"`
	Progress    *SyncProgress `json:"progress,omitempty"`
	Error       string        `json:"error,omitempty"`
	DryRun      bool          `json:"dry_run"`
	DeleteExtra bool          `json:"delete_extra"` // delete files on dest not in source
	Exclude     []string      `json:"exclude,omitempty"`
	cmd         *exec.Cmd
	cancel      context.CancelFunc
}

// SyncProgress represents progress of an ongoing sync
type SyncProgress struct {
	BytesTransferred int64         `json:"bytes_transferred"`
	BytesTotal       int64         `json:"bytes_total"`
	FilesTransferred int           `json:"files_transferred"`
	FilesTotal       int           `json:"files_total"`
	TransferSpeed    string        `json:"transfer_speed"`
	ETA              string        `json:"eta"`
	PercentComplete  float64       `json:"percent_complete"`
	CurrentFile      string        `json:"current_file,omitempty"`
	Errors           int           `json:"errors"`
	Checks           int           `json:"checks"`
	ElapsedTime      time.Duration `json:"elapsed_time"`
	LastUpdated      time.Time     `json:"last_updated"`
}

// RemoteInfo represents information about an rclone remote
type RemoteInfo struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Total     int64  `json:"total_bytes,omitempty"`
	Used      int64  `json:"used_bytes,omitempty"`
	Free      int64  `json:"free_bytes,omitempty"`
	Trashed   int64  `json:"trashed_bytes,omitempty"`
	Connected bool   `json:"connected"`
	Error     string `json:"error,omitempty"`
}

// FolderInfo represents a folder in local or remote storage
type FolderInfo struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	SizeBytes    int64     `json:"size_bytes"`
	FileCount    int       `json:"file_count"`
	FolderCount  int       `json:"folder_count"`
	LastModified time.Time `json:"last_modified"`
}

// CompareResult represents comparison between local and remote
type CompareResult struct {
	LocalPath       string    `json:"local_path"`
	RemotePath      string    `json:"remote_path"`
	LocalOnly       []string  `json:"local_only"`  // files only on local
	RemoteOnly      []string  `json:"remote_only"` // files only on remote
	Different       []string  `json:"different"`   // files that differ
	Matched         int       `json:"matched"`     // count of matching files
	TotalLocalSize  int64     `json:"total_local_size"`
	TotalRemoteSize int64     `json:"total_remote_size"`
	SizeToUpload    int64     `json:"size_to_upload"`
	EstimatedTime   string    `json:"estimated_time"` // based on transfer speed
	ScannedAt       time.Time `json:"scanned_at"`
}

// DriveInventory represents a full inventory of a local drive
type DriveInventory struct {
	DrivePath    string       `json:"drive_path"`
	TotalSize    int64        `json:"total_size_bytes"`
	UsedSize     int64        `json:"used_size_bytes"`
	FreeSize     int64        `json:"free_size_bytes"`
	Folders      []FolderInfo `json:"folders"`
	ScannedAt    time.Time    `json:"scanned_at"`
	ScanDuration string       `json:"scan_duration"`
	AccessErrors []string     `json:"access_errors,omitempty"`
}

// NewRcloneClient creates a new rclone client
func NewRcloneClient() (*RcloneClient, error) {
	configPath := os.Getenv("RCLONE_CONFIG")
	if configPath == "" {
		home, _ := os.UserHomeDir()
		configPath = filepath.Join(home, "AppData", "Roaming", "rclone", "rclone.conf")
	}

	// Get number of transfers from env or use default (8 for good bandwidth utilization)
	transfers := 8
	if t := os.Getenv("RCLONE_TRANSFERS"); t != "" {
		if n, err := strconv.Atoi(t); err == nil && n > 0 {
			transfers = n
		}
	}

	// Get number of checkers from env or use default
	checkers := 16
	if c := os.Getenv("RCLONE_CHECKERS"); c != "" {
		if n, err := strconv.Atoi(c); err == nil && n > 0 {
			checkers = n
		}
	}

	bandwidth := os.Getenv("RCLONE_BANDWIDTH")

	// Default chunk size optimized for large files (16MB is good for video)
	driveChunkSize := os.Getenv("RCLONE_DRIVE_CHUNK_SIZE")
	if driveChunkSize == "" {
		driveChunkSize = "16M" // 16MB chunks for better large file performance
	}

	// Set up state directories
	home, _ := os.UserHomeDir()
	checkpointDir := filepath.Join(home, ".hg-mcp", "checkpoints")
	quotaStateFile := filepath.Join(home, ".hg-mcp", "quota_state.json")
	historyFile := filepath.Join(home, ".hg-mcp", "transfer_history.json")

	// Create checkpoint directory if it doesn't exist
	os.MkdirAll(checkpointDir, 0755)

	return &RcloneClient{
		configPath:     configPath,
		transfers:      transfers,
		checkers:       checkers,
		bandwidth:      bandwidth,
		driveChunkSize: driveChunkSize,
		fastList:       true, // Enable by default for 20x faster listings
		activeJobs:     make(map[string]*SyncJob),

		// Multi-threading defaults for large files
		multiThreadStreams: 16,
		multiThreadCutoff:  "256M",

		// Google Drive rate limit protection
		tpsLimit: 3, // Google's write limit
		tpsBurst: 5, // Brief burst allowance

		// Memory management
		maxBufferMemory: "512M",
		useCookies:      true, // Persistent connections

		// Retry settings
		lowLevelRetries: 5,
		retriesSleep:    "30s",
		resilient:       true, // Continue on errors

		// State management
		checkpointDir:  checkpointDir,
		quotaStateFile: quotaStateFile,
		speedHistory:   make([]SpeedSample, 0, 60), // Keep 60 samples (30 sec at 0.5s intervals)

		// Comparison mode defaults
		compareMode:  CompareSizeModTime, // Balanced default
		skipExisting: false,
		updateOnly:   false,
		noTraverse:   false,
		orderBy:      "",

		// History tracking
		historyFile:     historyFile,
		transferHistory: make([]TransferRecord, 0),
	}, nil
}

// ApplyProfile applies a sync profile preset
func (c *RcloneClient) ApplyProfile(profile SyncProfile) {
	switch profile {
	case ProfileLargeFiles:
		// Optimized for VJ clips, movies, videos - large files
		c.transfers = 4           // Fewer transfers to maximize per-file bandwidth
		c.checkers = 8            // Moderate checking
		c.driveChunkSize = "128M" // Larger chunks for big files (increased from 64M)
		c.multiThreadStreams = 16 // Multi-threaded downloads for large files
		c.multiThreadCutoff = "256M"
		c.maxBufferMemory = "1G" // More buffer for large files
	case ProfileManySmallFiles:
		// Optimized for documents, code, many small files
		c.transfers = 16         // Many parallel transfers
		c.checkers = 32          // Many checkers for fast scanning
		c.driveChunkSize = "8M"  // Smaller chunks
		c.multiThreadStreams = 0 // Disable multi-threading for small files
		c.maxBufferMemory = "256M"
	case ProfileBackground:
		// Conservative settings to minimize impact on other work
		c.transfers = 2
		c.checkers = 4
		c.bandwidth = "5M" // 5MB/s limit
		c.driveChunkSize = "8M"
		c.multiThreadStreams = 4 // Limited multi-threading
		c.maxBufferMemory = "128M"
	case ProfileGoogleDrive:
		// Optimized specifically for Google Drive with rate limit protection
		c.transfers = 4 // Conservative transfers
		c.checkers = 8
		c.driveChunkSize = "64M"
		c.tpsLimit = 3 // Google's 3 req/sec write limit
		c.tpsBurst = 5
		c.multiThreadStreams = 8
		c.multiThreadCutoff = "256M"
		c.lowLevelRetries = 10 // More retries for transient errors
		c.retriesSleep = "30s"
		c.resilient = true
		c.useCookies = true
	default: // ProfileDefault
		c.transfers = 8
		c.checkers = 16
		c.driveChunkSize = "16M"
		c.multiThreadStreams = 16
		c.multiThreadCutoff = "256M"
	}
}

// ListRemotes returns all configured rclone remotes
func (c *RcloneClient) ListRemotes(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "rclone", "listremotes", "--config", c.configPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list remotes: %w", err)
	}

	var remotes []string
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			remotes = append(remotes, strings.TrimSuffix(line, ":"))
		}
	}
	return remotes, nil
}

// GetRemoteInfo returns information about a remote
func (c *RcloneClient) GetRemoteInfo(ctx context.Context, remote string) (*RemoteInfo, error) {
	info := &RemoteInfo{
		Name: remote,
	}

	// Get remote type from config
	cmd := exec.CommandContext(ctx, "rclone", "config", "show", remote, "--config", c.configPath)
	output, err := cmd.Output()
	if err == nil {
		for _, line := range strings.Split(string(output), "\n") {
			if strings.HasPrefix(line, "type = ") {
				info.Type = strings.TrimPrefix(line, "type = ")
				break
			}
		}
	}

	// Get space usage
	cmd = exec.CommandContext(ctx, "rclone", "about", remote+":", "--json", "--config", c.configPath)
	output, err = cmd.Output()
	if err != nil {
		info.Error = fmt.Sprintf("failed to get remote info: %v", err)
		return info, nil
	}

	var aboutData map[string]interface{}
	if err := json.Unmarshal(output, &aboutData); err == nil {
		if total, ok := aboutData["total"].(float64); ok {
			info.Total = int64(total)
		}
		if used, ok := aboutData["used"].(float64); ok {
			info.Used = int64(used)
		}
		if free, ok := aboutData["free"].(float64); ok {
			info.Free = int64(free)
		}
		if trashed, ok := aboutData["trashed"].(float64); ok {
			info.Trashed = int64(trashed)
		}
		info.Connected = true
	}

	return info, nil
}

// ListFolders lists top-level folders in a path (local or remote)
func (c *RcloneClient) ListFolders(ctx context.Context, path string) ([]FolderInfo, error) {
	cmd := exec.CommandContext(ctx, "rclone", "lsjson", path, "--dirs-only", "--config", c.configPath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list folders: %w", err)
	}

	var items []struct {
		Name    string    `json:"Name"`
		Size    int64     `json:"Size"`
		ModTime time.Time `json:"ModTime"`
		IsDir   bool      `json:"IsDir"`
	}
	if err := json.Unmarshal(output, &items); err != nil {
		return nil, fmt.Errorf("failed to parse folder list: %w", err)
	}

	var folders []FolderInfo
	for _, item := range items {
		if item.IsDir {
			folders = append(folders, FolderInfo{
				Name:         item.Name,
				Path:         filepath.Join(path, item.Name),
				LastModified: item.ModTime,
			})
		}
	}

	return folders, nil
}

// InventoryLocalDrive creates an inventory of a local drive
func (c *RcloneClient) InventoryLocalDrive(ctx context.Context, drivePath string) (*DriveInventory, error) {
	start := time.Now()
	inventory := &DriveInventory{
		DrivePath: drivePath,
		ScannedAt: start,
		Folders:   []FolderInfo{},
	}

	// Get drive space info using PowerShell (Windows)
	driveLetter := strings.TrimSuffix(drivePath, ":\\")
	if err := sanitize.DriveLetter(driveLetter); err != nil {
		return nil, fmt.Errorf("invalid drive path %q: %w", drivePath, err)
	}
	psCmd := fmt.Sprintf(`Get-Volume -DriveLetter '%s' | Select-Object Size, SizeRemaining | ConvertTo-Json`, driveLetter)
	cmd := exec.CommandContext(ctx, "powershell", "-Command", psCmd)
	output, err := cmd.Output()
	if err == nil {
		var volInfo struct {
			Size          int64 `json:"Size"`
			SizeRemaining int64 `json:"SizeRemaining"`
		}
		if json.Unmarshal(output, &volInfo) == nil {
			inventory.TotalSize = volInfo.Size
			inventory.FreeSize = volInfo.SizeRemaining
			inventory.UsedSize = volInfo.Size - volInfo.SizeRemaining
		}
	}

	// List top-level folders
	entries, err := os.ReadDir(drivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read drive: %w", err)
	}

	// Skip system folders
	skipFolders := map[string]bool{
		"$Recycle.Bin":              true,
		"$GetCurrent":               true,
		"$Windows.~WS":              true,
		"System Volume Information": true,
		"Recovery":                  true,
		"hiberfil.sys":              true,
		"pagefile.sys":              true,
		"swapfile.sys":              true,
	}

	for _, entry := range entries {
		if skipFolders[entry.Name()] {
			continue
		}

		folderPath := filepath.Join(drivePath, entry.Name())
		info, err := entry.Info()
		if err != nil {
			inventory.AccessErrors = append(inventory.AccessErrors, folderPath)
			continue
		}

		folderInfo := FolderInfo{
			Name:         entry.Name(),
			Path:         folderPath,
			LastModified: info.ModTime(),
		}

		if entry.IsDir() {
			// Get folder size using rclone (faster and handles permissions better)
			sizeCmd := exec.CommandContext(ctx, "rclone", "size", folderPath, "--json", "--config", c.configPath)
			sizeOut, err := sizeCmd.Output()
			if err == nil {
				var sizeInfo struct {
					Count int64 `json:"count"`
					Bytes int64 `json:"bytes"`
				}
				if json.Unmarshal(sizeOut, &sizeInfo) == nil {
					folderInfo.FileCount = int(sizeInfo.Count)
					folderInfo.SizeBytes = sizeInfo.Bytes
				}
			}
		} else {
			folderInfo.SizeBytes = info.Size()
			folderInfo.FileCount = 1
		}

		inventory.Folders = append(inventory.Folders, folderInfo)
	}

	inventory.ScanDuration = time.Since(start).Round(time.Millisecond).String()
	return inventory, nil
}

// Compare compares local path with remote path
func (c *RcloneClient) Compare(ctx context.Context, localPath, remotePath string) (*CompareResult, error) {
	result := &CompareResult{
		LocalPath:  localPath,
		RemotePath: remotePath,
		ScannedAt:  time.Now(),
		LocalOnly:  []string{},
		RemoteOnly: []string{},
		Different:  []string{},
	}

	// Use rclone check to compare
	cmd := exec.CommandContext(ctx, "rclone", "check", localPath, remotePath,
		"--combined", "-",
		"--config", c.configPath,
		"--checkers", strconv.Itoa(c.checkers))

	output, _ := cmd.Output() // rclone check returns non-zero if differences exist

	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}

		status := parts[0]
		path := parts[1]

		switch status {
		case "+": // local only
			result.LocalOnly = append(result.LocalOnly, path)
		case "-": // remote only
			result.RemoteOnly = append(result.RemoteOnly, path)
		case "*": // different
			result.Different = append(result.Different, path)
		case "=": // matched
			result.Matched++
		}
	}

	// Get sizes
	localSizeCmd := exec.CommandContext(ctx, "rclone", "size", localPath, "--json", "--config", c.configPath)
	if out, err := localSizeCmd.Output(); err == nil {
		var sizeInfo struct {
			Bytes int64 `json:"bytes"`
		}
		json.Unmarshal(out, &sizeInfo)
		result.TotalLocalSize = sizeInfo.Bytes
	}

	remoteSizeCmd := exec.CommandContext(ctx, "rclone", "size", remotePath, "--json", "--config", c.configPath)
	if out, err := remoteSizeCmd.Output(); err == nil {
		var sizeInfo struct {
			Bytes int64 `json:"bytes"`
		}
		json.Unmarshal(out, &sizeInfo)
		result.TotalRemoteSize = sizeInfo.Bytes
	}

	result.SizeToUpload = result.TotalLocalSize - result.TotalRemoteSize
	if result.SizeToUpload < 0 {
		result.SizeToUpload = 0
	}

	// Estimate time based on typical Google Drive speeds (10MB/s conservative)
	speedBytesPerSec := int64(10 * 1024 * 1024) // 10MB/s
	if result.SizeToUpload > 0 {
		seconds := result.SizeToUpload / speedBytesPerSec
		result.EstimatedTime = (time.Duration(seconds) * time.Second).String()
	}

	return result, nil
}

// StartSync starts a sync job in the background
func (c *RcloneClient) StartSync(ctx context.Context, source, dest string, dryRun, deleteExtra bool, exclude []string) (*SyncJob, error) {
	jobID := fmt.Sprintf("sync_%s", time.Now().Format("20060102_150405"))

	jobCtx, cancel := context.WithCancel(ctx)

	job := &SyncJob{
		ID:          jobID,
		Source:      source,
		Destination: dest,
		Status:      "pending",
		StartTime:   time.Now(),
		DryRun:      dryRun,
		DeleteExtra: deleteExtra,
		Exclude:     exclude,
		cancel:      cancel,
		Progress:    &SyncProgress{LastUpdated: time.Now()},
	}

	c.mu.Lock()
	c.activeJobs[jobID] = job
	c.mu.Unlock()

	// Start sync in background
	go c.runSync(jobCtx, job)

	return job, nil
}

// runSync executes the actual sync operation
func (c *RcloneClient) runSync(ctx context.Context, job *SyncJob) {
	job.Status = "running"

	args := []string{
		"sync",
		job.Source,
		job.Destination,
		"--config", c.configPath,
		"--transfers", strconv.Itoa(c.transfers),
		"--checkers", strconv.Itoa(c.checkers),
		"--progress",
		"--stats", "1s",
		"--stats-one-line",
		"-v",
	}

	// Add --fast-list for 20x faster directory listing (Google Drive optimization)
	if c.fastList {
		args = append(args, "--fast-list")
	}

	// Add chunk size for Google Drive (larger = better for big files)
	if c.driveChunkSize != "" {
		args = append(args, "--drive-chunk-size", c.driveChunkSize)
	}

	// Handle Google Drive 750GB daily upload limit gracefully
	args = append(args, "--drive-stop-on-upload-limit")

	// Multi-threading for large files
	if c.multiThreadStreams > 0 {
		args = append(args, "--multi-thread-streams", strconv.Itoa(c.multiThreadStreams))
	}
	if c.multiThreadCutoff != "" {
		args = append(args, "--multi-thread-cutoff", c.multiThreadCutoff)
	}

	// Rate limiting for Google Drive
	if c.tpsLimit > 0 {
		args = append(args, "--tpslimit", strconv.Itoa(c.tpsLimit))
	}
	if c.tpsBurst > 0 {
		args = append(args, "--tpslimit-burst", strconv.Itoa(c.tpsBurst))
	}

	// Memory management
	if c.maxBufferMemory != "" {
		args = append(args, "--buffer-size", c.maxBufferMemory)
	}

	// Persistent connections
	if c.useCookies {
		args = append(args, "--use-cookies")
	}

	// Retry settings for resilience
	if c.lowLevelRetries > 0 {
		args = append(args, "--low-level-retries", strconv.Itoa(c.lowLevelRetries))
	}
	if c.retriesSleep != "" {
		args = append(args, "--retries-sleep", c.retriesSleep)
	}
	if c.resilient {
		args = append(args, "--ignore-errors")
	}

	// Duration limit for controlled checkpointing
	if c.maxDuration != "" {
		args = append(args, "--max-duration", c.maxDuration)
	}

	// Bandwidth limit - check schedule first, then fall back to static limit
	currentLimit := c.GetCurrentBandwidthLimit()
	if currentLimit != "" && currentLimit != "off" {
		args = append(args, "--bwlimit", currentLimit)
	} else if c.bandwidth != "" {
		args = append(args, "--bwlimit", c.bandwidth)
	}

	if job.DryRun {
		args = append(args, "--dry-run")
	}

	if job.DeleteExtra {
		args = append(args, "--delete-during")
	}

	// Comparison mode flags for deduplication
	if c.skipExisting {
		args = append(args, "--ignore-existing")
	}
	if c.updateOnly {
		args = append(args, "--update")
	}
	switch c.compareMode {
	case CompareSizeOnly:
		args = append(args, "--size-only")
	case CompareChecksum:
		args = append(args, "--checksum")
		// CompareSizeModTime is the default, no flag needed
	}

	// Performance optimization flags
	if c.noTraverse {
		args = append(args, "--no-traverse")
	}
	if c.orderBy != "" {
		args = append(args, "--order-by", c.orderBy)
	}

	// Always exclude common problematic Windows patterns
	defaultExcludes := []string{
		"$RECYCLE.BIN/**",
		"$Recycle.Bin/**",
		"System Volume Information/**",
		"*.tmp",
		"Thumbs.db",
		"desktop.ini",
		"*.lnk",
		"hiberfil.sys",
		"pagefile.sys",
		"swapfile.sys",
	}

	allExcludes := append(defaultExcludes, job.Exclude...)
	for _, pattern := range allExcludes {
		args = append(args, "--exclude", pattern)
	}

	cmd := exec.CommandContext(ctx, "rclone", args...)
	job.cmd = cmd

	// Create pipe for output
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		job.Status = "failed"
		job.Error = fmt.Sprintf("failed to create pipe: %v", err)
		return
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		job.Status = "failed"
		job.Error = fmt.Sprintf("failed to start: %v", err)
		return
	}

	// Parse progress output
	scanner := bufio.NewScanner(stdout)
	progressRegex := regexp.MustCompile(`Transferred:\s+(\d+\.?\d*)\s*(\w+)\s*/\s*(\d+\.?\d*)\s*(\w+),\s*(\d+)%,\s*(\d+\.?\d*)\s*(\w+/s),\s*ETA\s+(.+)`)
	filesRegex := regexp.MustCompile(`Transferred:\s+(\d+)\s*/\s*(\d+)`)
	errorsRegex := regexp.MustCompile(`Errors:\s+(\d+)`)
	checksRegex := regexp.MustCompile(`Checks:\s+(\d+)`)

	for scanner.Scan() {
		line := scanner.Text()

		// Parse transferred bytes and speed
		if matches := progressRegex.FindStringSubmatch(line); len(matches) > 0 {
			job.Progress.BytesTransferred = parseSize(matches[1], matches[2])
			job.Progress.BytesTotal = parseSize(matches[3], matches[4])
			percent, _ := strconv.ParseFloat(matches[5], 64)
			job.Progress.PercentComplete = percent
			job.Progress.TransferSpeed = matches[6] + " " + matches[7]
			job.Progress.ETA = matches[8]
			job.Progress.LastUpdated = time.Now()
			job.Progress.ElapsedTime = time.Since(job.StartTime)
		}

		// Parse file counts
		if matches := filesRegex.FindStringSubmatch(line); len(matches) > 0 {
			job.Progress.FilesTransferred, _ = strconv.Atoi(matches[1])
			job.Progress.FilesTotal, _ = strconv.Atoi(matches[2])
		}

		// Parse errors
		if matches := errorsRegex.FindStringSubmatch(line); len(matches) > 0 {
			job.Progress.Errors, _ = strconv.Atoi(matches[1])
		}

		// Parse checks
		if matches := checksRegex.FindStringSubmatch(line); len(matches) > 0 {
			job.Progress.Checks, _ = strconv.Atoi(matches[1])
		}

		// Current file
		if strings.HasPrefix(line, "Transferring:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) > 1 {
				job.Progress.CurrentFile = strings.TrimSpace(parts[1])
			}
		}
	}

	// Wait for completion
	err = cmd.Wait()
	endTime := time.Now()
	job.EndTime = &endTime
	job.Progress.ElapsedTime = time.Since(job.StartTime)

	if ctx.Err() != nil {
		job.Status = "cancelled"
		job.Error = "sync was cancelled"
	} else if err != nil {
		job.Status = "failed"
		job.Error = fmt.Sprintf("sync failed: %v", err)
	} else {
		job.Status = "completed"
		job.Progress.PercentComplete = 100
	}
}

// GetJobStatus returns the current status of a sync job
func (c *RcloneClient) GetJobStatus(jobID string) (*SyncJob, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	job, exists := c.activeJobs[jobID]
	if !exists {
		return nil, fmt.Errorf("job not found: %s", jobID)
	}

	return job, nil
}

// ListJobs returns all sync jobs
func (c *RcloneClient) ListJobs() []*SyncJob {
	c.mu.RLock()
	defer c.mu.RUnlock()

	jobs := make([]*SyncJob, 0, len(c.activeJobs))
	for _, job := range c.activeJobs {
		jobs = append(jobs, job)
	}
	return jobs
}

// CancelJob cancels a running sync job
func (c *RcloneClient) CancelJob(jobID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	job, exists := c.activeJobs[jobID]
	if !exists {
		return fmt.Errorf("job not found: %s", jobID)
	}

	if job.Status != "running" && job.Status != "pending" {
		return fmt.Errorf("job is not running: %s", job.Status)
	}

	if job.cancel != nil {
		job.cancel()
	}
	if job.cmd != nil && job.cmd.Process != nil {
		job.cmd.Process.Kill()
	}

	return nil
}

// QuickSync performs a one-off sync operation (blocking)
func (c *RcloneClient) QuickSync(ctx context.Context, source, dest string, dryRun bool) error {
	args := []string{
		"sync",
		source,
		dest,
		"--config", c.configPath,
		"--transfers", strconv.Itoa(c.transfers),
		"--checkers", strconv.Itoa(c.checkers),
		"--progress",
	}

	if dryRun {
		args = append(args, "--dry-run")
	}

	cmd := exec.CommandContext(ctx, "rclone", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// SetTransfers sets the number of parallel transfers
func (c *RcloneClient) SetTransfers(n int) {
	if n > 0 {
		c.transfers = n
	}
}

// SetCheckers sets the number of parallel checkers
func (c *RcloneClient) SetCheckers(n int) {
	if n > 0 {
		c.checkers = n
	}
}

// SetBandwidth sets the bandwidth limit
func (c *RcloneClient) SetBandwidth(limit string) {
	c.bandwidth = limit
}

// SetDriveChunkSize sets the chunk size for Google Drive uploads
func (c *RcloneClient) SetDriveChunkSize(size string) {
	c.driveChunkSize = size
}

// SetFastList enables or disables fast listing
func (c *RcloneClient) SetFastList(enabled bool) {
	c.fastList = enabled
}

// SetMaxDuration sets the maximum duration for a sync operation
func (c *RcloneClient) SetMaxDuration(duration string) {
	c.maxDuration = duration
}

// GetConfig returns current configuration
func (c *RcloneClient) GetConfig() map[string]interface{} {
	return map[string]interface{}{
		"transfers":            c.transfers,
		"checkers":             c.checkers,
		"bandwidth":            c.bandwidth,
		"drive_chunk_size":     c.driveChunkSize,
		"fast_list":            c.fastList,
		"multi_thread_streams": c.multiThreadStreams,
		"multi_thread_cutoff":  c.multiThreadCutoff,
		"tps_limit":            c.tpsLimit,
		"tps_burst":            c.tpsBurst,
		"max_buffer_memory":    c.maxBufferMemory,
		"low_level_retries":    c.lowLevelRetries,
		"retries_sleep":        c.retriesSleep,
		"resilient":            c.resilient,
		"use_cookies":          c.useCookies,
		"max_duration":         c.maxDuration,
	}
}

// ==========================================
// Bandwidth Scheduling Methods
// ==========================================

// GetCurrentBandwidthLimit returns the current bandwidth limit based on schedule
func (c *RcloneClient) GetCurrentBandwidthLimit() string {
	if len(c.bandwidthSchedule) == 0 {
		return c.bandwidth
	}

	currentHour := time.Now().Hour()
	for _, schedule := range c.bandwidthSchedule {
		// Handle schedules that wrap around midnight
		if schedule.StartHour <= schedule.EndHour {
			if currentHour >= schedule.StartHour && currentHour < schedule.EndHour {
				return schedule.Limit
			}
		} else {
			// Wrap around midnight (e.g., 22:00 to 06:00)
			if currentHour >= schedule.StartHour || currentHour < schedule.EndHour {
				return schedule.Limit
			}
		}
	}

	return c.bandwidth // Fall back to default
}

// SetBandwidthSchedule sets time-based bandwidth limits
func (c *RcloneClient) SetBandwidthSchedule(schedules []BandwidthSchedule) {
	c.bandwidthSchedule = schedules
}

// GetBandwidthSchedule returns the current bandwidth schedule
func (c *RcloneClient) GetBandwidthSchedule() []BandwidthSchedule {
	return c.bandwidthSchedule
}

// ==========================================
// Speed Tracking Methods
// ==========================================

// RecordSpeed records a speed sample for rolling average calculation
func (c *RcloneClient) RecordSpeed(bytesPerSec float64) {
	c.speedMu.Lock()
	defer c.speedMu.Unlock()

	sample := SpeedSample{
		Timestamp:   time.Now(),
		BytesPerSec: bytesPerSec,
	}

	// Keep only last 60 samples (30 seconds at 0.5s intervals)
	if len(c.speedHistory) >= 60 {
		c.speedHistory = c.speedHistory[1:]
	}
	c.speedHistory = append(c.speedHistory, sample)
}

// GetRollingAverageSpeed returns the average speed over the last 30 seconds
func (c *RcloneClient) GetRollingAverageSpeed() float64 {
	c.speedMu.RLock()
	defer c.speedMu.RUnlock()

	if len(c.speedHistory) == 0 {
		return 0
	}

	// Filter samples from last 30 seconds
	cutoff := time.Now().Add(-30 * time.Second)
	var total float64
	var count int

	for _, sample := range c.speedHistory {
		if sample.Timestamp.After(cutoff) {
			total += sample.BytesPerSec
			count++
		}
	}

	if count == 0 {
		return 0
	}
	return total / float64(count)
}

// EstimateETAWithHistory estimates ETA using rolling average speed
func (c *RcloneClient) EstimateETAWithHistory(remainingBytes int64) string {
	avgSpeed := c.GetRollingAverageSpeed()
	if avgSpeed <= 0 {
		return "calculating..."
	}

	seconds := float64(remainingBytes) / avgSpeed
	duration := time.Duration(seconds) * time.Second

	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%dm %ds", int(duration.Minutes()), int(duration.Seconds())%60)
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%dh %dm", int(duration.Hours()), int(duration.Minutes())%60)
	} else {
		days := int(duration.Hours()) / 24
		hours := int(duration.Hours()) % 24
		return fmt.Sprintf("%dd %dh", days, hours)
	}
}

// ==========================================
// Quota Tracking Methods
// ==========================================

// TrackUpload records bytes uploaded to daily quota tracking
func (c *RcloneClient) TrackUpload(bytes int64) error {
	quota, err := c.GetDailyQuotaUsage()
	if err != nil {
		quota = &DailyQuotaState{
			Date:          time.Now().Format("2006-01-02"),
			BytesUploaded: 0,
		}
	}

	// Reset if new day
	today := time.Now().Format("2006-01-02")
	if quota.Date != today {
		quota.Date = today
		quota.BytesUploaded = 0
	}

	quota.BytesUploaded += bytes
	quota.LastUpdated = time.Now()

	// Save to file
	data, err := json.MarshalIndent(quota, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.quotaStateFile, data, 0644)
}

// GetDailyQuotaUsage returns the current daily quota usage
func (c *RcloneClient) GetDailyQuotaUsage() (*DailyQuotaState, error) {
	data, err := os.ReadFile(c.quotaStateFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &DailyQuotaState{
				Date:          time.Now().Format("2006-01-02"),
				BytesUploaded: 0,
				LastUpdated:   time.Now(),
			}, nil
		}
		return nil, err
	}

	var quota DailyQuotaState
	if err := json.Unmarshal(data, &quota); err != nil {
		return nil, err
	}

	// Reset if new day
	today := time.Now().Format("2006-01-02")
	if quota.Date != today {
		return &DailyQuotaState{
			Date:          today,
			BytesUploaded: 0,
			LastUpdated:   time.Now(),
		}, nil
	}

	return &quota, nil
}

// ShouldPauseForQuota checks if we should pause due to quota limits
// Returns (shouldPause, remainingBytes)
func (c *RcloneClient) ShouldPauseForQuota() (bool, int64) {
	quota, err := c.GetDailyQuotaUsage()
	if err != nil {
		return false, 0
	}

	// Google Drive daily limit is 750 GiB
	dailyLimit := int64(750 * 1024 * 1024 * 1024)
	// Pause at 90% (650 GiB) to be safe
	pauseThreshold := int64(650 * 1024 * 1024 * 1024)

	remaining := dailyLimit - quota.BytesUploaded
	return quota.BytesUploaded >= pauseThreshold, remaining
}

// ==========================================
// Checkpoint/Resume Methods
// ==========================================

// SaveCheckpoint saves the current job state for resume capability
func (c *RcloneClient) SaveCheckpoint(job *SyncJob) error {
	checkpoint := &CheckpointState{
		JobID:          job.ID,
		Source:         job.Source,
		Destination:    job.Destination,
		CompletedFiles: make(map[string]int64),
		FailedFiles:    []FailedFile{},
		LastCheckpoint: time.Now(),
	}

	if job.Progress != nil {
		checkpoint.BytesUploaded = job.Progress.BytesTransferred
	}

	checkpointFile := filepath.Join(c.checkpointDir, job.ID+".json")
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(checkpointFile, data, 0644)
}

// LoadCheckpoint loads a checkpoint for a job
func (c *RcloneClient) LoadCheckpoint(jobID string) (*CheckpointState, error) {
	checkpointFile := filepath.Join(c.checkpointDir, jobID+".json")
	data, err := os.ReadFile(checkpointFile)
	if err != nil {
		return nil, err
	}

	var checkpoint CheckpointState
	if err := json.Unmarshal(data, &checkpoint); err != nil {
		return nil, err
	}

	return &checkpoint, nil
}

// ListCheckpoints lists all available checkpoints
func (c *RcloneClient) ListCheckpoints() ([]string, error) {
	files, err := os.ReadDir(c.checkpointDir)
	if err != nil {
		return nil, err
	}

	var checkpoints []string
	for _, f := range files {
		if strings.HasSuffix(f.Name(), ".json") {
			checkpoints = append(checkpoints, strings.TrimSuffix(f.Name(), ".json"))
		}
	}
	return checkpoints, nil
}

// DeleteCheckpoint removes a checkpoint file
func (c *RcloneClient) DeleteCheckpoint(jobID string) error {
	checkpointFile := filepath.Join(c.checkpointDir, jobID+".json")
	return os.Remove(checkpointFile)
}

// ResumeFromCheckpoint resumes a sync from a checkpoint
func (c *RcloneClient) ResumeFromCheckpoint(ctx context.Context, checkpoint *CheckpointState) (*SyncJob, error) {
	// Start a new sync job from the checkpoint source/dest
	// The actual resume will use rclone's built-in --ignore-existing for completed files
	return c.StartSync(ctx, checkpoint.Source, checkpoint.Destination, false, false, nil)
}

// ==========================================
// Failed File Retry Methods
// ==========================================

// RecordFailedFile adds a file to the failed files list for a job
func (c *RcloneClient) RecordFailedFile(jobID string, path string, size int64, errMsg string) error {
	checkpoint, err := c.LoadCheckpoint(jobID)
	if err != nil {
		checkpoint = &CheckpointState{
			JobID:          jobID,
			CompletedFiles: make(map[string]int64),
			FailedFiles:    []FailedFile{},
		}
	}

	// Check if already in failed list and increment retry count
	found := false
	for i, f := range checkpoint.FailedFiles {
		if f.Path == path {
			checkpoint.FailedFiles[i].RetryCount++
			checkpoint.FailedFiles[i].LastTry = time.Now()
			checkpoint.FailedFiles[i].Error = errMsg
			found = true
			break
		}
	}

	if !found {
		checkpoint.FailedFiles = append(checkpoint.FailedFiles, FailedFile{
			Path:       path,
			Size:       size,
			Error:      errMsg,
			RetryCount: 1,
			LastTry:    time.Now(),
		})
	}

	// Save updated checkpoint
	checkpointFile := filepath.Join(c.checkpointDir, jobID+".json")
	data, err := json.MarshalIndent(checkpoint, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(checkpointFile, data, 0644)
}

// GetFailedFiles returns the list of failed files for a job
func (c *RcloneClient) GetFailedFiles(jobID string) ([]FailedFile, error) {
	checkpoint, err := c.LoadCheckpoint(jobID)
	if err != nil {
		return nil, err
	}
	return checkpoint.FailedFiles, nil
}

// RetryFailedFiles creates a new sync job to retry only the failed files
func (c *RcloneClient) RetryFailedFiles(ctx context.Context, jobID string) (*SyncJob, error) {
	checkpoint, err := c.LoadCheckpoint(jobID)
	if err != nil {
		return nil, fmt.Errorf("no checkpoint found for job: %s", jobID)
	}

	if len(checkpoint.FailedFiles) == 0 {
		return nil, fmt.Errorf("no failed files to retry")
	}

	// Create a new job for retrying
	newJobID := fmt.Sprintf("retry_%s_%s", jobID, time.Now().Format("150405"))

	// For retrying specific files, we need to use include patterns
	var includes []string
	for _, f := range checkpoint.FailedFiles {
		includes = append(includes, f.Path)
	}

	jobCtx, cancel := context.WithCancel(ctx)

	job := &SyncJob{
		ID:          newJobID,
		Source:      checkpoint.Source,
		Destination: checkpoint.Destination,
		Status:      "pending",
		StartTime:   time.Now(),
		DryRun:      false,
		DeleteExtra: false,
		cancel:      cancel,
		Progress:    &SyncProgress{LastUpdated: time.Now()},
	}

	c.mu.Lock()
	c.activeJobs[newJobID] = job
	c.mu.Unlock()

	// Start the retry sync
	go c.runRetrySync(jobCtx, job, includes)

	return job, nil
}

// runRetrySync runs a sync targeting only specific files
func (c *RcloneClient) runRetrySync(ctx context.Context, job *SyncJob, includes []string) {
	job.Status = "running"

	args := []string{
		"copy", // Use copy instead of sync for targeted retry
		job.Source,
		job.Destination,
		"--config", c.configPath,
		"--transfers", strconv.Itoa(c.transfers),
		"--checkers", strconv.Itoa(c.checkers),
		"--progress",
		"--stats", "1s",
		"-v",
	}

	// Add include patterns for specific files
	for _, include := range includes {
		args = append(args, "--include", include)
	}

	// Add retry-specific settings
	args = append(args, "--retries", "5")
	args = append(args, "--low-level-retries", "10")

	cmd := exec.CommandContext(ctx, "rclone", args...)
	job.cmd = cmd

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		job.Status = "failed"
		job.Error = fmt.Sprintf("failed to create pipe: %v", err)
		return
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		job.Status = "failed"
		job.Error = fmt.Sprintf("failed to start: %v", err)
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		// Basic progress tracking
		job.Progress.LastUpdated = time.Now()
		job.Progress.ElapsedTime = time.Since(job.StartTime)
	}

	err = cmd.Wait()
	endTime := time.Now()
	job.EndTime = &endTime

	if ctx.Err() != nil {
		job.Status = "cancelled"
	} else if err != nil {
		job.Status = "failed"
		job.Error = err.Error()
	} else {
		job.Status = "completed"
	}
}

// EstimateUploadTime estimates upload time based on size and typical Google Drive speeds
func EstimateUploadTime(bytes int64) string {
	// Google Drive typically achieves 10-50 MB/s depending on file size and connection
	// Use conservative estimate of 15 MB/s for planning
	speedBytesPerSec := int64(15 * 1024 * 1024)

	if bytes <= 0 {
		return "0s"
	}

	seconds := bytes / speedBytesPerSec

	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	} else if seconds < 3600 {
		return fmt.Sprintf("%dm %ds", seconds/60, seconds%60)
	} else if seconds < 86400 {
		hours := seconds / 3600
		minutes := (seconds % 3600) / 60
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		days := seconds / 86400
		hours := (seconds % 86400) / 3600
		return fmt.Sprintf("%dd %dh", days, hours)
	}
}

// CheckDailyQuota checks how much of the 750GB daily limit has been used
// Note: This is an estimate based on rclone logs, as Google doesn't expose this API
func (c *RcloneClient) CheckDailyQuota() (uploaded int64, limitGB int64, warning string) {
	limitGB = 750
	// Would need to track daily uploads in a state file to be accurate
	// For now, return advisory information
	warning = "Google Drive has a 750 GiB daily upload limit. Use --drive-stop-on-upload-limit to handle gracefully."
	return 0, limitGB, warning
}

// parseSize converts size string to bytes
func parseSize(value, unit string) int64 {
	v, _ := strconv.ParseFloat(value, 64)

	multipliers := map[string]float64{
		"B":   1,
		"KiB": 1024,
		"MiB": 1024 * 1024,
		"GiB": 1024 * 1024 * 1024,
		"TiB": 1024 * 1024 * 1024 * 1024,
		"KB":  1000,
		"MB":  1000 * 1000,
		"GB":  1000 * 1000 * 1000,
		"TB":  1000 * 1000 * 1000 * 1000,
		"kB":  1000,
		"k":   1000,
		"M":   1000 * 1000,
		"G":   1000 * 1000 * 1000,
		"T":   1000 * 1000 * 1000 * 1000,
	}

	if mult, ok := multipliers[unit]; ok {
		return int64(v * mult)
	}
	return int64(v)
}

// FormatBytes formats bytes to human readable string
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// ==========================================
// Phase 2: Comparison Mode Methods
// ==========================================

// SetCompareMode sets the file comparison mode
func (c *RcloneClient) SetCompareMode(mode ComparisonMode) {
	c.compareMode = mode
}

// GetCompareMode returns the current comparison mode
func (c *RcloneClient) GetCompareMode() ComparisonMode {
	return c.compareMode
}

// SetSkipExisting enables/disables skipping files that exist on destination
func (c *RcloneClient) SetSkipExisting(skip bool) {
	c.skipExisting = skip
}

// GetSkipExisting returns whether skip existing is enabled
func (c *RcloneClient) GetSkipExisting() bool {
	return c.skipExisting
}

// SetUpdateOnly enables/disables update-only mode (skip newer files on dest)
func (c *RcloneClient) SetUpdateOnly(update bool) {
	c.updateOnly = update
}

// GetUpdateOnly returns whether update-only mode is enabled
func (c *RcloneClient) GetUpdateOnly() bool {
	return c.updateOnly
}

// SetNoTraverse enables/disables no-traverse mode for faster sync to empty destinations
func (c *RcloneClient) SetNoTraverse(noTraverse bool) {
	c.noTraverse = noTraverse
}

// SetOrderBy sets the file ordering for transfers (e.g., "size,desc" for largest first)
func (c *RcloneClient) SetOrderBy(orderBy string) {
	c.orderBy = orderBy
}

// ==========================================
// Phase 2: Pre-Sync Analysis Methods
// ==========================================

// AnalyzeSync performs a dry-run analysis before actual sync
func (c *RcloneClient) AnalyzeSync(ctx context.Context, source, dest string) (*SyncAnalysis, error) {
	start := time.Now()

	analysis := &SyncAnalysis{
		Source:      source,
		Destination: dest,
		ByExtension: make(map[string]int64),
		AnalyzedAt:  start,
	}

	// Run rclone sync with --dry-run to see what would be transferred
	args := []string{
		"sync",
		source,
		dest,
		"--dry-run",
		"--config", c.configPath,
		"--checkers", strconv.Itoa(c.checkers),
		"--stats", "0",
		"-v",
	}

	if c.fastList {
		args = append(args, "--fast-list")
	}

	cmd := exec.CommandContext(ctx, "rclone", args...)
	output, _ := cmd.CombinedOutput()

	// Parse the output for files to be transferred
	lines := strings.Split(string(output), "\n")
	transferRegex := regexp.MustCompile(`^(\S+):\s+Copied \(new\)$|^(\S+):\s+Copied \(replaced existing\)$`)
	sizeRegex := regexp.MustCompile(`Transferred:\s+(\d+\.?\d*)\s*(\w+)\s*/\s*(\d+\.?\d*)\s*(\w+)`)

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Check for files that would be transferred
		if matches := transferRegex.FindStringSubmatch(line); len(matches) > 0 {
			var fileName string
			if matches[1] != "" {
				fileName = matches[1]
			} else if matches[2] != "" {
				fileName = matches[2]
			}

			if fileName != "" {
				analysis.FilesToTransfer++
				// Get file extension for categorization
				ext := strings.ToLower(filepath.Ext(fileName))
				if ext == "" {
					ext = "(no extension)"
				}
				// We'll get sizes from the summary
				analysis.ByExtension[ext]++
			}
		}

		// Parse size summary
		if matches := sizeRegex.FindStringSubmatch(line); len(matches) > 0 {
			analysis.BytesToTransfer = parseSize(matches[3], matches[4])
		}
	}

	// Get source size for total files calculation
	sizeCmd := exec.CommandContext(ctx, "rclone", "size", source, "--json", "--config", c.configPath)
	sizeOut, err := sizeCmd.Output()
	if err == nil {
		var sizeInfo struct {
			Count int64 `json:"count"`
			Bytes int64 `json:"bytes"`
		}
		if json.Unmarshal(sizeOut, &sizeInfo) == nil {
			analysis.TotalFiles = int(sizeInfo.Count)
			analysis.FilesToSkip = analysis.TotalFiles - analysis.FilesToTransfer
		}
	}

	// Estimate transfer time
	analysis.EstimatedTime = EstimateUploadTime(analysis.BytesToTransfer)

	// Get largest files (requires a separate listing)
	analysis.LargestFiles = c.getLargestFiles(ctx, source, 10)

	return analysis, nil
}

// getLargestFiles returns the N largest files in a path
func (c *RcloneClient) getLargestFiles(ctx context.Context, path string, n int) []FileTransferInfo {
	cmd := exec.CommandContext(ctx, "rclone", "lsjson", path,
		"--recursive", "--files-only",
		"--config", c.configPath)

	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	var files []struct {
		Path    string    `json:"Path"`
		Size    int64     `json:"Size"`
		ModTime time.Time `json:"ModTime"`
	}
	if err := json.Unmarshal(output, &files); err != nil {
		return nil
	}

	// Sort by size descending
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			if files[j].Size > files[i].Size {
				files[i], files[j] = files[j], files[i]
			}
		}
	}

	// Take top N
	if len(files) > n {
		files = files[:n]
	}

	result := make([]FileTransferInfo, len(files))
	for i, f := range files {
		result[i] = FileTransferInfo{
			Path:      f.Path,
			Size:      f.Size,
			SizeHuman: FormatBytes(f.Size),
			ModTime:   f.ModTime,
		}
	}
	return result
}

// ==========================================
// Phase 2: Post-Sync Verification Methods
// ==========================================

// VerifySync verifies files after sync using rclone check
func (c *RcloneClient) VerifySync(ctx context.Context, source, dest string, useChecksum bool) (*VerifyResult, error) {
	start := time.Now()

	result := &VerifyResult{
		Source:      source,
		Destination: dest,
		VerifiedAt:  start,
	}

	args := []string{
		"check",
		source,
		dest,
		"--one-way",
		"--config", c.configPath,
		"--checkers", strconv.Itoa(c.checkers),
	}

	if useChecksum {
		args = append(args, "--checksum")
	}
	if c.fastList {
		args = append(args, "--fast-list")
	}

	// Use combined output to capture differences
	args = append(args, "--combined", "-")

	cmd := exec.CommandContext(ctx, "rclone", args...)
	output, _ := cmd.CombinedOutput()

	// Parse the combined output
	for _, line := range strings.Split(string(output), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}

		status := parts[0]
		path := parts[1]

		switch status {
		case "=": // matched
			result.Verified++
		case "*": // different
			result.Mismatched++
			if len(result.MismatchList) < 100 {
				result.MismatchList = append(result.MismatchList, path)
			}
		case "-": // missing from dest (only in source)
			result.Missing++
			if len(result.MissingList) < 100 {
				result.MissingList = append(result.MissingList, path)
			}
		case "+": // extra on dest (only in dest, shouldn't happen with --one-way)
			result.Extra++
			if len(result.ExtraList) < 100 {
				result.ExtraList = append(result.ExtraList, path)
			}
		}
	}

	result.Duration = time.Since(start).Round(time.Millisecond).String()
	return result, nil
}

// ==========================================
// Phase 2: Dashboard & Monitoring Methods
// ==========================================

// GetDashboard returns comprehensive status of all sync operations
func (c *RcloneClient) GetDashboard() *SyncDashboard {
	c.mu.RLock()
	defer c.mu.RUnlock()

	dashboard := &SyncDashboard{
		ActiveJobs:    make([]*SyncJob, 0),
		TotalProgress: &AggregateProgress{},
		SpeedMetrics:  &SpeedMetrics{},
		GeneratedAt:   time.Now(),
	}

	// Collect active jobs
	var recentErrors []string
	for _, job := range c.activeJobs {
		dashboard.ActiveJobs = append(dashboard.ActiveJobs, job)

		if job.Progress != nil {
			dashboard.TotalProgress.TotalBytes += job.Progress.BytesTotal
			dashboard.TotalProgress.TransferredBytes += job.Progress.BytesTransferred
			dashboard.TotalProgress.TotalFiles += job.Progress.FilesTotal
			dashboard.TotalProgress.TransferredFiles += job.Progress.FilesTransferred
		}

		if job.Error != "" && len(recentErrors) < 10 {
			recentErrors = append(recentErrors, fmt.Sprintf("[%s] %s", job.ID, job.Error))
		}
	}

	// Calculate overall percentage
	if dashboard.TotalProgress.TotalBytes > 0 {
		dashboard.TotalProgress.OverallPercent = float64(dashboard.TotalProgress.TransferredBytes) / float64(dashboard.TotalProgress.TotalBytes) * 100
	}

	// Calculate combined ETA
	remaining := dashboard.TotalProgress.TotalBytes - dashboard.TotalProgress.TransferredBytes
	dashboard.TotalProgress.CombinedETA = c.EstimateETAWithHistory(remaining)

	// Speed metrics
	avgSpeed := c.GetRollingAverageSpeed()
	dashboard.SpeedMetrics.AverageSpeed = avgSpeed / (1024 * 1024) // Convert to MB/s

	// Get peak and current from history
	c.speedMu.RLock()
	if len(c.speedHistory) > 0 {
		dashboard.SpeedMetrics.CurrentSpeed = c.speedHistory[len(c.speedHistory)-1].BytesPerSec / (1024 * 1024)
		for _, s := range c.speedHistory {
			if s.BytesPerSec/(1024*1024) > dashboard.SpeedMetrics.PeakSpeed {
				dashboard.SpeedMetrics.PeakSpeed = s.BytesPerSec / (1024 * 1024)
			}
		}
	}
	c.speedMu.RUnlock()

	// Determine trend
	if len(c.speedHistory) >= 10 {
		recent := c.speedHistory[len(c.speedHistory)-5:]
		older := c.speedHistory[len(c.speedHistory)-10 : len(c.speedHistory)-5]
		var recentAvg, olderAvg float64
		for _, s := range recent {
			recentAvg += s.BytesPerSec
		}
		for _, s := range older {
			olderAvg += s.BytesPerSec
		}
		recentAvg /= 5
		olderAvg /= 5

		if recentAvg > olderAvg*1.1 {
			dashboard.SpeedMetrics.SpeedTrend = "increasing"
		} else if recentAvg < olderAvg*0.9 {
			dashboard.SpeedMetrics.SpeedTrend = "decreasing"
		} else {
			dashboard.SpeedMetrics.SpeedTrend = "stable"
		}
	} else {
		dashboard.SpeedMetrics.SpeedTrend = "calculating"
	}

	// Get quota status
	quota, _ := c.GetDailyQuotaUsage()
	dashboard.QuotaStatus = quota
	dashboard.RecentErrors = recentErrors

	return dashboard
}

// ==========================================
// Phase 2: Transfer History Methods
// ==========================================

// RecordTransfer saves a completed transfer to history
func (c *RcloneClient) RecordTransfer(job *SyncJob) error {
	if job.EndTime == nil {
		return fmt.Errorf("job not completed")
	}

	record := TransferRecord{
		JobID:       job.ID,
		Source:      job.Source,
		Destination: job.Destination,
		StartTime:   job.StartTime,
		EndTime:     *job.EndTime,
		Status:      job.Status,
	}

	if job.Progress != nil {
		record.BytesTransferred = job.Progress.BytesTransferred
		record.FilesTransferred = job.Progress.FilesTransferred
		record.Errors = job.Progress.Errors

		// Calculate average speed
		duration := job.EndTime.Sub(job.StartTime).Seconds()
		if duration > 0 {
			record.AverageSpeed = float64(record.BytesTransferred) / duration
		}
	}

	c.historyMu.Lock()
	c.transferHistory = append(c.transferHistory, record)
	c.historyMu.Unlock()

	return c.saveHistory()
}

// GetTransferHistory returns transfer history for the specified number of days
func (c *RcloneClient) GetTransferHistory(days int) ([]TransferRecord, error) {
	if err := c.loadHistory(); err != nil {
		return nil, err
	}

	c.historyMu.RLock()
	defer c.historyMu.RUnlock()

	cutoff := time.Now().AddDate(0, 0, -days)
	var results []TransferRecord

	for _, record := range c.transferHistory {
		if record.StartTime.After(cutoff) {
			results = append(results, record)
		}
	}

	return results, nil
}

// GetTransferStats returns aggregate transfer statistics
func (c *RcloneClient) GetTransferStats() (map[string]interface{}, error) {
	if err := c.loadHistory(); err != nil {
		return nil, err
	}

	c.historyMu.RLock()
	defer c.historyMu.RUnlock()

	stats := map[string]interface{}{
		"total_transfers": len(c.transferHistory),
		"total_bytes":     int64(0),
		"total_files":     0,
		"completed":       0,
		"failed":          0,
		"average_speed":   float64(0),
		"total_errors":    0,
	}

	var totalSpeed float64
	var speedCount int

	for _, record := range c.transferHistory {
		stats["total_bytes"] = stats["total_bytes"].(int64) + record.BytesTransferred
		stats["total_files"] = stats["total_files"].(int) + record.FilesTransferred
		stats["total_errors"] = stats["total_errors"].(int) + record.Errors

		if record.Status == "completed" {
			stats["completed"] = stats["completed"].(int) + 1
		} else if record.Status == "failed" {
			stats["failed"] = stats["failed"].(int) + 1
		}

		if record.AverageSpeed > 0 {
			totalSpeed += record.AverageSpeed
			speedCount++
		}
	}

	if speedCount > 0 {
		stats["average_speed"] = totalSpeed / float64(speedCount)
	}

	stats["total_bytes_human"] = FormatBytes(stats["total_bytes"].(int64))

	return stats, nil
}

// saveHistory saves transfer history to file
func (c *RcloneClient) saveHistory() error {
	c.historyMu.RLock()
	defer c.historyMu.RUnlock()

	data, err := json.MarshalIndent(c.transferHistory, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.historyFile, data, 0644)
}

// loadHistory loads transfer history from file
func (c *RcloneClient) loadHistory() error {
	data, err := os.ReadFile(c.historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // No history yet
		}
		return err
	}

	c.historyMu.Lock()
	defer c.historyMu.Unlock()

	return json.Unmarshal(data, &c.transferHistory)
}

// ==========================================
// Phase 2: Duplicate Detection Methods
// ==========================================

// FindDuplicatesInHashIndex finds duplicate files from a hash index
func (c *RcloneClient) FindDuplicatesInHashIndex(index *HashIndex) []DuplicateGroup {
	var duplicates []DuplicateGroup

	for hash, files := range index.Files {
		if len(files) > 1 {
			group := DuplicateGroup{
				Hash:  hash,
				Paths: make([]string, len(files)),
			}

			for i, f := range files {
				group.Paths[i] = f.Path
				if i == 0 {
					group.Size = f.Size
					group.SizeHuman = FormatBytes(f.Size)
				}
			}

			group.WastedSize = group.Size * int64(len(files)-1)
			duplicates = append(duplicates, group)
		}
	}

	// Sort by wasted size descending
	for i := 0; i < len(duplicates)-1; i++ {
		for j := i + 1; j < len(duplicates); j++ {
			if duplicates[j].WastedSize > duplicates[i].WastedSize {
				duplicates[i], duplicates[j] = duplicates[j], duplicates[i]
			}
		}
	}

	return duplicates
}

// ==========================================
// Phase 4: Verified Sync & Failure Analytics
// ==========================================

// SyncWithVerify performs a sync and automatically verifies the result
func (c *RcloneClient) SyncWithVerify(ctx context.Context, source, dest string, opts SyncOptions) (*VerifiedSyncResult, error) {
	result := &VerifiedSyncResult{
		FinalStatus: "pending",
	}

	// Set defaults
	if opts.MaxRetries <= 0 {
		opts.MaxRetries = 3
	}
	if opts.VerifyMode == "" {
		opts.VerifyMode = "size"
	}

	// Start the sync
	job, err := c.StartSync(ctx, source, dest, false, false, nil)
	if err != nil {
		result.FinalStatus = "failed"
		return result, err
	}
	result.Job = job

	// Wait for sync to complete
	for {
		select {
		case <-ctx.Done():
			result.FinalStatus = "cancelled"
			return result, ctx.Err()
		case <-time.After(2 * time.Second):
			status, err := c.GetJobStatus(job.ID)
			if err != nil {
				continue
			}
			if status.Status == "completed" || status.Status == "failed" || status.Status == "cancelled" {
				result.Job = status
				if status.Status != "completed" {
					result.FinalStatus = "failed"
					return result, nil
				}
				goto verify
			}
		}
	}

verify:
	if !opts.AutoVerify {
		result.FinalStatus = "completed"
		return result, nil
	}

	// Run verification
	useChecksum := opts.VerifyMode == "checksum"
	verifyResult, err := c.VerifySync(ctx, source, dest, useChecksum)
	if err != nil {
		result.FinalStatus = "verify_failed"
		return result, err
	}
	result.VerifyResult = verifyResult

	// Check if verification passed
	if verifyResult.Mismatched == 0 && verifyResult.Missing == 0 {
		result.FinalStatus = "verified"
		return result, nil
	}

	// Handle mismatches with retry if enabled
	if opts.RetryMismatches && (verifyResult.Mismatched > 0 || verifyResult.Missing > 0) {
		// Collect files to retry
		var filesToRetry []string
		filesToRetry = append(filesToRetry, verifyResult.MismatchList...)
		filesToRetry = append(filesToRetry, verifyResult.MissingList...)

		if len(filesToRetry) > 0 {
			// Create a retry sync for specific files
			retryJob, err := c.retrySpecificFiles(ctx, source, dest, filesToRetry, opts.MaxRetries)
			if err == nil {
				result.RetryResult = retryJob
			}
		}
	}

	result.FinalStatus = "mismatches"
	return result, nil
}

// retrySpecificFiles creates a sync job for specific files only
func (c *RcloneClient) retrySpecificFiles(ctx context.Context, source, dest string, files []string, maxRetries int) (*SyncJob, error) {
	// Use --files-from with a temporary file
	tmpFile, err := os.CreateTemp("", "rclone-retry-*.txt")
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	for _, f := range files {
		tmpFile.WriteString(f + "\n")
	}
	tmpFile.Close()

	// Build rclone command with --files-from
	args := []string{
		"copy",
		source,
		dest,
		"--files-from", tmpFile.Name(),
		"--retries", fmt.Sprintf("%d", maxRetries),
		"--low-level-retries", "10",
		"-v",
	}

	if c.configPath != "" {
		args = append(args, "--config", c.configPath)
	}

	job := &SyncJob{
		ID:          fmt.Sprintf("retry_%d", time.Now().UnixNano()),
		Source:      source,
		Destination: dest,
		Status:      "running",
		StartTime:   time.Now(),
		Progress:    &SyncProgress{},
	}

	cmd := exec.CommandContext(ctx, "rclone", args...)
	output, err := cmd.CombinedOutput()

	now := time.Now()
	job.EndTime = &now

	if err != nil {
		job.Status = "failed"
		job.Error = string(output)
	} else {
		job.Status = "completed"
	}

	return job, nil
}

// AnalyzeFailures categorizes failed files and provides recovery suggestions
func (c *RcloneClient) AnalyzeFailures(jobID string) (*FailureAnalysis, error) {
	checkpoint, err := c.LoadCheckpoint(jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to load checkpoint: %w", err)
	}

	analysis := &FailureAnalysis{
		JobID:       jobID,
		ByErrorType: make(map[string]int),
		ByExtension: make(map[string]int),
		AnalyzedAt:  time.Now(),
	}

	for _, f := range checkpoint.FailedFiles {
		// Categorize error
		errorType := categorizeError(f.Error)
		ext := strings.ToLower(filepath.Ext(f.Path))
		if ext == "" {
			ext = "(no extension)"
		}

		failedInfo := FailedFileInfo{
			Path:        f.Path,
			Size:        f.Size,
			Error:       f.Error,
			ErrorType:   errorType,
			RetryCount:  f.RetryCount,
			LastAttempt: f.LastTry,
			Extension:   ext,
		}

		analysis.FailedFiles = append(analysis.FailedFiles, failedInfo)
		analysis.TotalFailed++
		analysis.TotalSize += f.Size
		analysis.ByErrorType[errorType]++
		analysis.ByExtension[ext]++

		// Count recoverable errors
		if isRecoverableError(errorType) {
			analysis.Recoverable++
		}
	}

	// Generate suggestions
	analysis.Suggestions = generateRecoverySuggestions(analysis.ByErrorType)

	return analysis, nil
}

// categorizeError determines the error type from the error message
func categorizeError(errMsg string) string {
	errLower := strings.ToLower(errMsg)

	switch {
	case strings.Contains(errLower, "quota") || strings.Contains(errLower, "rate limit") || strings.Contains(errLower, "403"):
		return "quota"
	case strings.Contains(errLower, "permission") || strings.Contains(errLower, "access denied") || strings.Contains(errLower, "401"):
		return "permission"
	case strings.Contains(errLower, "timeout") || strings.Contains(errLower, "timed out"):
		return "timeout"
	case strings.Contains(errLower, "network") || strings.Contains(errLower, "connection") || strings.Contains(errLower, "dns"):
		return "network"
	case strings.Contains(errLower, "corrupt") || strings.Contains(errLower, "checksum") || strings.Contains(errLower, "hash mismatch"):
		return "corrupt"
	case strings.Contains(errLower, "not found") || strings.Contains(errLower, "404"):
		return "not_found"
	case strings.Contains(errLower, "space") || strings.Contains(errLower, "disk full"):
		return "disk_full"
	default:
		return "unknown"
	}
}

// isRecoverableError checks if an error type is recoverable with retry
func isRecoverableError(errorType string) bool {
	switch errorType {
	case "timeout", "network", "quota":
		return true
	default:
		return false
	}
}

// generateRecoverySuggestions creates actionable suggestions based on error types
func generateRecoverySuggestions(byErrorType map[string]int) []RecoverySuggestion {
	var suggestions []RecoverySuggestion

	for errorType, count := range byErrorType {
		var suggestion RecoverySuggestion
		suggestion.ErrorType = errorType
		suggestion.Count = count

		switch errorType {
		case "quota":
			suggestion.Suggestion = "Google Drive daily quota (750GB) exceeded. Wait until midnight Pacific time for quota reset, or continue tomorrow."
			suggestion.AutoFixable = true
			suggestion.Command = "aftrs_rclone_retry_failed"
		case "permission":
			suggestion.Suggestion = "Permission denied. Check if the destination path exists and you have write access. Re-authenticate if needed."
			suggestion.AutoFixable = false
			suggestion.Command = "rclone config reconnect remote:"
		case "timeout":
			suggestion.Suggestion = "Connection timed out. Check network stability. Consider reducing parallel transfers or increasing timeout."
			suggestion.AutoFixable = true
			suggestion.Command = "aftrs_rclone_retry_failed"
		case "network":
			suggestion.Suggestion = "Network error. Check internet connection and DNS resolution. Try again when connection is stable."
			suggestion.AutoFixable = true
			suggestion.Command = "aftrs_rclone_retry_failed"
		case "corrupt":
			suggestion.Suggestion = "File corruption detected. Re-download or restore from backup. Check source file integrity."
			suggestion.AutoFixable = false
		case "not_found":
			suggestion.Suggestion = "Source file not found. File may have been moved or deleted. Update source path or remove from retry list."
			suggestion.AutoFixable = false
		case "disk_full":
			suggestion.Suggestion = "Destination disk full. Free up space or use a different destination."
			suggestion.AutoFixable = false
		case "unknown":
			suggestion.Suggestion = "Unknown error. Review the error messages and retry manually."
			suggestion.AutoFixable = false
			suggestion.Command = "aftrs_rclone_retry_failed"
		}

		suggestions = append(suggestions, suggestion)
	}

	// Sort by count descending
	for i := 0; i < len(suggestions)-1; i++ {
		for j := i + 1; j < len(suggestions); j++ {
			if suggestions[j].Count > suggestions[i].Count {
				suggestions[i], suggestions[j] = suggestions[j], suggestions[i]
			}
		}
	}

	return suggestions
}

// GetFailureAnalysis is an alias for AnalyzeFailures for API consistency
func (c *RcloneClient) GetFailureAnalysis(jobID string) (*FailureAnalysis, error) {
	return c.AnalyzeFailures(jobID)
}
