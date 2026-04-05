// Package clients provides API clients for external services.
package clients

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/hairglasses-studio/mcpkit/sanitize"
)

// DataMigrationClient handles drive scanning, categorization, and cloud migration
type DataMigrationClient struct {
	rcloneClient     *RcloneClient
	hashIndexes      map[string]*HashIndex
	activeJobs       map[string]*MigrationJob
	categoryMappings map[string]string
	excludePatterns  []string
	mu               sync.RWMutex
	indexDir         string
}

// HashIndex stores file hashes for deduplication
type HashIndex struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Path      string                 `json:"path"`
	Files     map[string][]FileEntry `json:"files"` // hash -> files with that hash
	CreatedAt time.Time              `json:"created_at"`
	FileCount int                    `json:"file_count"`
	TotalSize int64                  `json:"total_size"`
}

// FileEntry represents a single file in the hash index
type FileEntry struct {
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	Hash    string    `json:"hash"`
	ModTime time.Time `json:"mod_time"`
}

// MigrationJob represents a migration operation
type MigrationJob struct {
	ID          string        `json:"id"`
	Type        string        `json:"type"` // "scan", "sync", "hash", "mount"
	Source      string        `json:"source"`
	Destination string        `json:"destination,omitempty"`
	Status      string        `json:"status"` // pending, running, completed, failed, cancelled
	Progress    *SyncProgress `json:"progress,omitempty"`
	StartTime   time.Time     `json:"start_time"`
	EndTime     *time.Time    `json:"end_time,omitempty"`
	Error       string        `json:"error,omitempty"`
	Result      interface{}   `json:"result,omitempty"`
	cancel      context.CancelFunc
}

// DriveInfo represents information about a scanned drive
type DriveInfo struct {
	Path          string            `json:"path"`
	Label         string            `json:"label,omitempty"`
	FileSystem    string            `json:"filesystem"`
	TotalBytes    int64             `json:"total_bytes"`
	UsedBytes     int64             `json:"used_bytes"`
	FreeBytes     int64             `json:"free_bytes"`
	Categories    []CategorySummary `json:"categories"`
	Folders       []FolderScan      `json:"folders"`
	TotalFiles    int               `json:"total_files"`
	TotalFolders  int               `json:"total_folders"`
	ScanTime      time.Duration     `json:"scan_time"`
	ExcludedSize  int64             `json:"excluded_size"`
	ExcludedCount int               `json:"excluded_count"`
}

// CategorySummary groups folders by category
type CategorySummary struct {
	Category    string       `json:"category"`
	TargetPath  string       `json:"target_path"`
	TotalSize   int64        `json:"total_size"`
	FileCount   int          `json:"file_count"`
	FolderCount int          `json:"folder_count"`
	Folders     []FolderScan `json:"folders"`
}

// FolderScan represents a scanned folder with categorization
type FolderScan struct {
	Name          string `json:"name"`
	Path          string `json:"path"`
	SizeBytes     int64  `json:"size_bytes"`
	FileCount     int    `json:"file_count"`
	Category      string `json:"category"`
	TargetPath    string `json:"target_path"`
	IsExcluded    bool   `json:"is_excluded"`
	ExcludeReason string `json:"exclude_reason,omitempty"`
}

// PhysicalDisk represents a physical disk for WSL2 mounting
type PhysicalDisk struct {
	Number         int    `json:"number"`
	FriendlyName   string `json:"friendly_name"`
	SizeBytes      int64  `json:"size_bytes"`
	PartitionStyle string `json:"partition_style"`
	BusType        string `json:"bus_type"`
	IsOnline       bool   `json:"is_online"`
}

// DuplicateReport contains duplicate detection results
type DuplicateReport struct {
	UniqueFiles    []FileEntry `json:"unique_files"`
	DuplicateFiles []FileEntry `json:"duplicate_files"`
	Conflicts      []FileEntry `json:"conflicts"` // Same name, different content
	SpaceSavings   int64       `json:"space_savings"`
	UniqueSize     int64       `json:"unique_size"`
	DuplicateSize  int64       `json:"duplicate_size"`
}

// Default category mappings
var defaultCategoryMappings = map[string]string{
	// Music Production
	"ableton":            "Music Production/Ableton",
	"ableton live":       "Music Production/Ableton",
	"splice":             "Music Production/Splice",
	"serum presets":      "Music Production/VST Presets/Serum",
	"native instruments": "Music Production/VST Presets/Native Instruments",
	"kontakt":            "Music Production/VST Presets/Kontakt",
	"massive":            "Music Production/VST Presets/Massive",
	"rekordbox":          "Music Production/DJ/rekordbox",
	"serato":             "Music Production/DJ/Serato",
	"traktor":            "Music Production/DJ/Traktor",
	"dj crates":          "Music Production/DJ Crates",

	// Visual Production
	"resolume arena": "Visual Production/Resolume Arena",
	"resolume wire":  "Visual Production/Resolume Wire",
	"resolume":       "Visual Production/Resolume Arena",
	"vj clips":       "Visual Production/VJ Clips",
	"vjloops":        "Visual Production/VJ Clips",
	"baeblade":       "Visual Production/VJ Clips/BAEBLADE",
	"nestdropprov2":  "Visual Production/NestDropProV2",
	"nestdrop":       "Visual Production/NestDropProV2",
	"ledfx":          "Visual Production/LedFx",
	"milkdrop":       "Visual Production/MilkDrop Presets",
	"touchdesigner":  "Visual Production/TouchDesigner",

	// Gaming
	"roms":        "Gaming & Emulation/ROMs",
	"pcsx2":       "Gaming & Emulation/PCSX2",
	"retroarch":   "Gaming & Emulation/RetroArch",
	"save games":  "Gaming & Emulation/Saves",
	"saved games": "Gaming & Emulation/Saves",
	"my games":    "Gaming & Emulation/Saves",

	// Development
	"projects":        "Development/Projects",
	"repos":           "Development/Repos",
	"unreal projects": "Development/Unreal Projects",
	"github":          "Development/Repos",

	// Documents
	"documents": "Documents & Notes/Documents",
	"obsidian":  "Documents & Notes/Obsidian",

	// Media
	"pictures":  "Media/Photos",
	"photos":    "Media/Photos",
	"videos":    "Media/Videos",
	"music":     "Media/Music",
	"downloads": "Downloads",
}

// Default exclude patterns
var defaultExcludePatterns = []string{
	// Manual archive - never sync
	"**/Kevin Archive/**",
	"**/kevin's Archive/**",
	"**/Kevin's Archive/**",
	"**/kevin archive/**",

	// Windows system
	"$Recycle.Bin/**",
	"$RECYCLE.BIN/**",
	"System Volume Information/**",
	"Windows/**",
	"Program Files/**",
	"Program Files (x86)/**",
	"ProgramData/**",
	"Recovery/**",
	"hiberfil.sys",
	"pagefile.sys",
	"swapfile.sys",
	"$WINDOWS.~BT/**",
	"$Windows.~WS/**",
	"ESD/**",
	"PerfLogs/**",
	"Intel/**",
	"inetpub/**",

	// Movie/TV patterns (skip pirated content)
	"**/*[0-9][0-9][0-9][0-9]p*YTS*/**",
	"**/*BluRay*/**",
	"**/*WEB-DL*/**",
	"**/*YIFY*/**",
	"**/*HDRip*/**",
	"**/*BRRip*/**",
	"**/*.S[0-9][0-9]E[0-9][0-9].*",

	// Temp files
	"**/*.tmp",
	"**/*.temp",
	"**/Thumbs.db",
	"**/.DS_Store",
	"**/desktop.ini",
}

// Movie/TV folder patterns for detection
var moviePatterns = []string{
	`\[\d{4}p\]`,
	`\[4K\]`,
	`BluRay`,
	`WEB-DL`,
	`YIFY`,
	`YTS`,
	`HDRip`,
	`BRRip`,
	`\.S\d{2}E\d{2}\.`,
	`Season\.\d+`,
	`S\d+-\d+`,
}

// NewDataMigrationClient creates a new data migration client
func NewDataMigrationClient() (*DataMigrationClient, error) {
	rclone, err := NewRcloneClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create rclone client: %w", err)
	}

	// Set up index directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	indexDir := filepath.Join(homeDir, ".aftrs", "migration", "indexes")
	if err := os.MkdirAll(indexDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create index directory: %w", err)
	}

	return &DataMigrationClient{
		rcloneClient:     rclone,
		hashIndexes:      make(map[string]*HashIndex),
		activeJobs:       make(map[string]*MigrationJob),
		categoryMappings: defaultCategoryMappings,
		excludePatterns:  defaultExcludePatterns,
		indexDir:         indexDir,
	}, nil
}

// ScanDrive scans a drive and categorizes its content
func (c *DataMigrationClient) ScanDrive(ctx context.Context, drivePath string) (*DriveInfo, error) {
	startTime := time.Now()

	// Normalize path
	drivePath = filepath.Clean(drivePath)
	if runtime.GOOS == "windows" && len(drivePath) == 1 {
		drivePath = drivePath + ":\\"
	}

	// Get drive info
	driveInfo := &DriveInfo{
		Path:       drivePath,
		FileSystem: "unknown",
		Categories: []CategorySummary{},
		Folders:    []FolderScan{},
	}

	// Get volume info on Windows
	if runtime.GOOS == "windows" {
		volumeInfo, err := c.getWindowsVolumeInfo(ctx, drivePath)
		if err == nil {
			driveInfo.Label = volumeInfo.Label
			driveInfo.FileSystem = volumeInfo.FileSystem
			driveInfo.TotalBytes = volumeInfo.TotalBytes
			driveInfo.FreeBytes = volumeInfo.FreeBytes
			driveInfo.UsedBytes = volumeInfo.TotalBytes - volumeInfo.FreeBytes
		}
	}

	// Scan top-level folders
	entries, err := os.ReadDir(drivePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	categoryMap := make(map[string]*CategorySummary)

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		folderName := entry.Name()
		folderPath := filepath.Join(drivePath, folderName)

		// Check if excluded
		excluded, reason := c.isExcluded(folderName, folderPath)

		// Get folder size using rclone
		sizeBytes, fileCount, err := c.getFolderSize(ctx, folderPath)
		if err != nil {
			// Skip folders we can't access
			continue
		}

		// Categorize folder
		category, targetPath := c.categorizeFolder(folderName)

		folder := FolderScan{
			Name:          folderName,
			Path:          folderPath,
			SizeBytes:     sizeBytes,
			FileCount:     fileCount,
			Category:      category,
			TargetPath:    targetPath,
			IsExcluded:    excluded,
			ExcludeReason: reason,
		}

		driveInfo.Folders = append(driveInfo.Folders, folder)
		driveInfo.TotalFiles += fileCount
		driveInfo.TotalFolders++

		if excluded {
			driveInfo.ExcludedSize += sizeBytes
			driveInfo.ExcludedCount++
			continue
		}

		// Add to category summary
		if _, exists := categoryMap[category]; !exists {
			categoryMap[category] = &CategorySummary{
				Category:   category,
				TargetPath: targetPath,
				Folders:    []FolderScan{},
			}
		}
		categoryMap[category].TotalSize += sizeBytes
		categoryMap[category].FileCount += fileCount
		categoryMap[category].FolderCount++
		categoryMap[category].Folders = append(categoryMap[category].Folders, folder)
	}

	// Convert map to slice and sort by size
	for _, summary := range categoryMap {
		driveInfo.Categories = append(driveInfo.Categories, *summary)
	}
	sort.Slice(driveInfo.Categories, func(i, j int) bool {
		return driveInfo.Categories[i].TotalSize > driveInfo.Categories[j].TotalSize
	})

	// Sort folders by size
	sort.Slice(driveInfo.Folders, func(i, j int) bool {
		return driveInfo.Folders[i].SizeBytes > driveInfo.Folders[j].SizeBytes
	})

	driveInfo.ScanTime = time.Since(startTime)
	return driveInfo, nil
}

// WindowsVolumeInfo holds Windows volume information
type WindowsVolumeInfo struct {
	Label      string
	FileSystem string
	TotalBytes int64
	FreeBytes  int64
}

func (c *DataMigrationClient) getWindowsVolumeInfo(ctx context.Context, drivePath string) (*WindowsVolumeInfo, error) {
	// Extract drive letter
	driveLetter := strings.TrimSuffix(strings.TrimSuffix(drivePath, "\\"), ":")
	if len(driveLetter) > 1 {
		driveLetter = string(driveLetter[0])
	}

	psCmd := fmt.Sprintf(`Get-Volume -DriveLetter '%s' | Select-Object FileSystemLabel, FileSystem, Size, SizeRemaining | ConvertTo-Json`, driveLetter)
	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", psCmd)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var result struct {
		FileSystemLabel string `json:"FileSystemLabel"`
		FileSystem      string `json:"FileSystem"`
		Size            int64  `json:"Size"`
		SizeRemaining   int64  `json:"SizeRemaining"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, err
	}

	return &WindowsVolumeInfo{
		Label:      result.FileSystemLabel,
		FileSystem: result.FileSystem,
		TotalBytes: result.Size,
		FreeBytes:  result.SizeRemaining,
	}, nil
}

func (c *DataMigrationClient) getFolderSize(ctx context.Context, folderPath string) (int64, int, error) {
	cmd := exec.CommandContext(ctx, "rclone", "size", folderPath, "--json")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	var result struct {
		Count int64 `json:"count"`
		Bytes int64 `json:"bytes"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return 0, 0, err
	}

	return result.Bytes, int(result.Count), nil
}

func (c *DataMigrationClient) isExcluded(folderName, folderPath string) (bool, string) {
	lowerName := strings.ToLower(folderName)

	// Check Kevin Archive
	if strings.Contains(lowerName, "kevin") && strings.Contains(lowerName, "archive") {
		return true, "Manual archive (Kevin Archive)"
	}

	// Check system folders
	systemFolders := []string{
		"$recycle.bin", "$windows.~bt", "$windows.~ws", "system volume information",
		"windows", "program files", "program files (x86)", "programdata",
		"recovery", "perflogs", "intel", "inetpub", "esd",
	}
	for _, sys := range systemFolders {
		if lowerName == sys {
			return true, "Windows system folder"
		}
	}

	// Check movie/TV patterns
	for _, pattern := range moviePatterns {
		if matched, _ := regexp.MatchString(pattern, folderName); matched {
			return true, "Movie/TV content"
		}
	}

	return false, ""
}

func (c *DataMigrationClient) categorizeFolder(folderName string) (string, string) {
	lowerName := strings.ToLower(folderName)

	// Check exact matches first
	if target, ok := c.categoryMappings[lowerName]; ok {
		parts := strings.Split(target, "/")
		return parts[0], target
	}

	// Check partial matches
	for key, target := range c.categoryMappings {
		if strings.Contains(lowerName, key) {
			parts := strings.Split(target, "/")
			return parts[0], target
		}
	}

	// Default categorization based on common patterns
	if strings.Contains(lowerName, "video") || strings.Contains(lowerName, "clip") || strings.Contains(lowerName, "loop") {
		return "Visual Production", "Visual Production/VJ Clips"
	}
	if strings.Contains(lowerName, "music") || strings.Contains(lowerName, "audio") || strings.Contains(lowerName, "sample") {
		return "Music Production", "Music Production/Samples"
	}
	if strings.Contains(lowerName, "rom") || strings.Contains(lowerName, "game") || strings.Contains(lowerName, "emulat") {
		return "Gaming & Emulation", "Gaming & Emulation"
	}
	if strings.Contains(lowerName, "project") || strings.Contains(lowerName, "dev") || strings.Contains(lowerName, "code") {
		return "Development", "Development/Projects"
	}
	if strings.Contains(lowerName, "doc") || strings.Contains(lowerName, "note") {
		return "Documents & Notes", "Documents & Notes"
	}
	if strings.Contains(lowerName, "photo") || strings.Contains(lowerName, "picture") || strings.Contains(lowerName, "image") {
		return "Media", "Media/Photos"
	}

	return "Uncategorized", "Uncategorized/" + folderName
}

// ListPhysicalDisks returns USB-connected physical disks for WSL2 mounting
func (c *DataMigrationClient) ListPhysicalDisks(ctx context.Context) ([]PhysicalDisk, error) {
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("WSL2 disk mounting only available on Windows")
	}

	psCmd := `Get-Disk | Where-Object { $_.BusType -eq 'USB' -or $_.BusType -eq 'SATA' } | Select-Object Number, FriendlyName, Size, PartitionStyle, BusType, IsOnline | ConvertTo-Json`
	cmd := exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", psCmd)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list disks: %w", err)
	}

	// Handle single result vs array
	var disks []PhysicalDisk
	if len(output) > 0 && output[0] == '[' {
		if err := json.Unmarshal(output, &disks); err != nil {
			return nil, fmt.Errorf("failed to parse disk list: %w", err)
		}
	} else {
		var disk PhysicalDisk
		if err := json.Unmarshal(output, &disk); err != nil {
			return nil, fmt.Errorf("failed to parse disk: %w", err)
		}
		disks = append(disks, disk)
	}

	return disks, nil
}

// MountLinuxDrive mounts a BTRFS/XFS drive via WSL2
func (c *DataMigrationClient) MountLinuxDrive(ctx context.Context, diskNum int, fsType string, mountPoint string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("WSL2 mounting only available on Windows")
	}

	// Validate filesystem type
	if err := sanitize.FileSystemType(fsType); err != nil {
		return fmt.Errorf("invalid filesystem type: %w", err)
	}

	if mountPoint == "" {
		mountPoint = "/mnt/recovery"
	}
	if err := sanitize.MountPoint(mountPoint); err != nil {
		return fmt.Errorf("invalid mount point: %w", err)
	}

	// Create mount point in WSL
	mkdirCmd := exec.CommandContext(ctx, "wsl", "-u", "root", "mkdir", "-p", mountPoint)
	if err := mkdirCmd.Run(); err != nil {
		return fmt.Errorf("failed to create mount point: %w", err)
	}

	// Mount the disk — diskNum is an int so injection-safe
	devicePath := fmt.Sprintf("\\\\.\\PHYSICALDRIVE%d", diskNum)
	mountCmd := exec.CommandContext(ctx, "wsl", "--mount", devicePath, "--type", fsType)
	if output, err := mountCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to mount disk: %w - %s", err, string(output))
	}

	return nil
}

// UnmountLinuxDrive safely unmounts a WSL2-mounted drive
func (c *DataMigrationClient) UnmountLinuxDrive(ctx context.Context, diskNum int) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("WSL2 unmounting only available on Windows")
	}

	devicePath := fmt.Sprintf("\\\\.\\PHYSICALDRIVE%d", diskNum)
	cmd := exec.CommandContext(ctx, "wsl", "--unmount", devicePath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to unmount disk: %w - %s", err, string(output))
	}

	return nil
}

// MountEncryptedDrive mounts a LUKS-encrypted drive via WSL2
func (c *DataMigrationClient) MountEncryptedDrive(ctx context.Context, diskNum int, keyfilePath string, mountPoint string) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("WSL2 mounting only available on Windows")
	}

	if mountPoint == "" {
		mountPoint = "/mnt/recovery"
	}
	if err := sanitize.MountPoint(mountPoint); err != nil {
		return fmt.Errorf("invalid mount point: %w", err)
	}

	// Attach disk to WSL2 in bare mode
	devicePath := fmt.Sprintf("\\\\.\\PHYSICALDRIVE%d", diskNum)
	attachCmd := exec.CommandContext(ctx, "wsl", "--mount", devicePath, "--bare")
	if output, err := attachCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to attach disk: %w - %s", err, string(output))
	}

	// Find the device in WSL (usually /dev/sdX)
	lsblkCmd := exec.CommandContext(ctx, "wsl", "-u", "root", "lsblk", "-o", "NAME,SIZE", "-n")
	output, err := lsblkCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list block devices: %w", err)
	}

	// Parse lsblk output to find the new device
	lines := strings.Split(string(output), "\n")
	var deviceName string
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "sd") {
			parts := strings.Fields(line)
			if len(parts) > 0 {
				deviceName = "/dev/" + parts[0]
				break
			}
		}
	}

	if deviceName == "" {
		return fmt.Errorf("could not find attached device")
	}

	// Convert Windows keyfile path to WSL path
	wslKeyfile := strings.Replace(keyfilePath, "\\", "/", -1)
	if len(wslKeyfile) > 1 && wslKeyfile[1] == ':' {
		wslKeyfile = "/mnt/" + strings.ToLower(string(wslKeyfile[0])) + wslKeyfile[2:]
	}

	// Open LUKS container
	luksCmd := exec.CommandContext(ctx, "wsl", "-u", "root", "cryptsetup", "luksOpen", deviceName+"1", "recovery", "--key-file", wslKeyfile)
	if output, err := luksCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to open LUKS container: %w - %s", err, string(output))
	}

	// Create mount point
	mkdirCmd := exec.CommandContext(ctx, "wsl", "-u", "root", "mkdir", "-p", mountPoint)
	mkdirCmd.Run()

	// Mount decrypted device
	mountCmd := exec.CommandContext(ctx, "wsl", "-u", "root", "mount", "/dev/mapper/recovery", mountPoint)
	if output, err := mountCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to mount decrypted device: %w - %s", err, string(output))
	}

	return nil
}

// BuildHashIndex creates a hash index for deduplication
func (c *DataMigrationClient) BuildHashIndex(ctx context.Context, sourcePath string, indexName string, fileTypes []string, minSizeMB int) (*HashIndex, error) {
	if indexName == "" {
		indexName = fmt.Sprintf("index_%d", time.Now().Unix())
	}

	index := &HashIndex{
		ID:        fmt.Sprintf("%s_%d", indexName, time.Now().UnixNano()),
		Name:      indexName,
		Path:      sourcePath,
		Files:     make(map[string][]FileEntry),
		CreatedAt: time.Now(),
	}

	minSize := int64(minSizeMB) * 1024 * 1024

	// Walk the directory tree
	err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip files we can't access
		}

		if info.IsDir() {
			return nil
		}

		// Check file size
		if info.Size() < minSize {
			return nil
		}

		// Check file type filter
		if len(fileTypes) > 0 {
			ext := strings.ToLower(filepath.Ext(path))
			matched := false
			for _, ft := range fileTypes {
				if ext == strings.ToLower(ft) || ext == "."+strings.ToLower(ft) {
					matched = true
					break
				}
			}
			if !matched {
				return nil
			}
		}

		// Calculate hash
		hash, err := c.calculateFileHash(path)
		if err != nil {
			return nil // Skip files we can't hash
		}

		entry := FileEntry{
			Path:    path,
			Size:    info.Size(),
			Hash:    hash,
			ModTime: info.ModTime(),
		}

		index.Files[hash] = append(index.Files[hash], entry)
		index.FileCount++
		index.TotalSize += info.Size()

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	// Save index
	c.mu.Lock()
	c.hashIndexes[index.ID] = index
	c.mu.Unlock()

	if err := c.saveIndex(index); err != nil {
		return nil, fmt.Errorf("failed to save index: %w", err)
	}

	return index, nil
}

func (c *DataMigrationClient) calculateFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (c *DataMigrationClient) saveIndex(index *HashIndex) error {
	indexPath := filepath.Join(c.indexDir, index.ID+".json")
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(indexPath, data, 0644)
}

// LoadIndex loads a previously saved hash index
func (c *DataMigrationClient) LoadIndex(indexID string) (*HashIndex, error) {
	c.mu.RLock()
	if index, ok := c.hashIndexes[indexID]; ok {
		c.mu.RUnlock()
		return index, nil
	}
	c.mu.RUnlock()

	indexPath := filepath.Join(c.indexDir, indexID+".json")
	data, err := os.ReadFile(indexPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read index: %w", err)
	}

	var index HashIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse index: %w", err)
	}

	c.mu.Lock()
	c.hashIndexes[index.ID] = &index
	c.mu.Unlock()

	return &index, nil
}

// FindDuplicates compares source with destination and reports duplicates
func (c *DataMigrationClient) FindDuplicates(ctx context.Context, sourcePath string, destRemote string, useHash bool) (*DuplicateReport, error) {
	report := &DuplicateReport{
		UniqueFiles:    []FileEntry{},
		DuplicateFiles: []FileEntry{},
		Conflicts:      []FileEntry{},
	}

	// Use rclone check for comparison
	args := []string{"check", sourcePath, destRemote, "--combined", "-"}
	if !useHash {
		args = append(args, "--size-only")
	}

	cmd := exec.CommandContext(ctx, "rclone", args...)
	output, _ := cmd.CombinedOutput()

	// Parse output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if len(line) < 2 {
			continue
		}

		status := line[0]
		filePath := strings.TrimSpace(line[2:])

		// Get file info
		info, err := os.Stat(filepath.Join(sourcePath, filePath))
		if err != nil {
			continue
		}

		entry := FileEntry{
			Path:    filePath,
			Size:    info.Size(),
			ModTime: info.ModTime(),
		}

		switch status {
		case '+': // Only in source (unique)
			report.UniqueFiles = append(report.UniqueFiles, entry)
			report.UniqueSize += entry.Size
		case '-': // Only in destination (already synced)
			// Don't add to report
		case '*': // Different content (conflict)
			report.Conflicts = append(report.Conflicts, entry)
		case '=': // Same (duplicate)
			report.DuplicateFiles = append(report.DuplicateFiles, entry)
			report.DuplicateSize += entry.Size
			report.SpaceSavings += entry.Size
		}
	}

	return report, nil
}

// SyncCategory syncs a categorized folder to Google Drive
func (c *DataMigrationClient) SyncCategory(ctx context.Context, sourcePath string, category string, dryRun bool, skipDuplicates bool) (*MigrationJob, error) {
	// Get target path from category
	targetPath, ok := c.categoryMappings[strings.ToLower(category)]
	if !ok {
		targetPath = category
	}

	destination := "gdrive:" + targetPath

	// Build exclude patterns
	excludes := append([]string{}, c.excludePatterns...)

	// Start sync via rclone client
	job, err := c.rcloneClient.StartSync(ctx, sourcePath, destination, dryRun, false, excludes)
	if err != nil {
		return nil, err
	}

	// Track as migration job
	migrationJob := &MigrationJob{
		ID:          job.ID,
		Type:        "sync",
		Source:      sourcePath,
		Destination: destination,
		Status:      job.Status,
		StartTime:   job.StartTime,
		Progress:    job.Progress,
	}

	c.mu.Lock()
	c.activeJobs[migrationJob.ID] = migrationJob
	c.mu.Unlock()

	return migrationJob, nil
}

// GetJobStatus returns the status of a migration job
func (c *DataMigrationClient) GetJobStatus(jobID string) (*MigrationJob, error) {
	c.mu.RLock()
	job, ok := c.activeJobs[jobID]
	c.mu.RUnlock()

	if !ok {
		// Try to get from rclone client
		syncJob, err := c.rcloneClient.GetJobStatus(jobID)
		if err != nil {
			return nil, fmt.Errorf("job not found: %s", jobID)
		}

		return &MigrationJob{
			ID:          syncJob.ID,
			Type:        "sync",
			Source:      syncJob.Source,
			Destination: syncJob.Destination,
			Status:      syncJob.Status,
			StartTime:   syncJob.StartTime,
			EndTime:     syncJob.EndTime,
			Progress:    syncJob.Progress,
			Error:       syncJob.Error,
		}, nil
	}

	// Update progress from rclone
	syncJob, err := c.rcloneClient.GetJobStatus(jobID)
	if err == nil && syncJob != nil {
		job.Status = syncJob.Status
		job.Progress = syncJob.Progress
		job.EndTime = syncJob.EndTime
		job.Error = syncJob.Error
	}

	return job, nil
}

// ListActiveJobs returns all active migration jobs
func (c *DataMigrationClient) ListActiveJobs() []*MigrationJob {
	c.mu.RLock()
	defer c.mu.RUnlock()

	jobs := make([]*MigrationJob, 0, len(c.activeJobs))
	for _, job := range c.activeJobs {
		jobs = append(jobs, job)
	}

	// Sort by start time
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].StartTime.After(jobs[j].StartTime)
	})

	return jobs
}

// GenerateDriveReport creates a comprehensive migration report for a drive
func (c *DataMigrationClient) GenerateDriveReport(ctx context.Context, drivePath string) (string, error) {
	driveInfo, err := c.ScanDrive(ctx, drivePath)
	if err != nil {
		return "", err
	}

	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Drive Migration Report: %s\n\n", drivePath))
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format(time.RFC3339)))

	// Drive Summary
	sb.WriteString("## Drive Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Label:** %s\n", driveInfo.Label))
	sb.WriteString(fmt.Sprintf("- **Filesystem:** %s\n", driveInfo.FileSystem))
	sb.WriteString(fmt.Sprintf("- **Total Size:** %s\n", FormatBytes(driveInfo.TotalBytes)))
	sb.WriteString(fmt.Sprintf("- **Used:** %s\n", FormatBytes(driveInfo.UsedBytes)))
	sb.WriteString(fmt.Sprintf("- **Free:** %s\n", FormatBytes(driveInfo.FreeBytes)))
	sb.WriteString(fmt.Sprintf("- **Total Folders:** %d\n", driveInfo.TotalFolders))
	sb.WriteString(fmt.Sprintf("- **Total Files:** %d\n", driveInfo.TotalFiles))
	sb.WriteString(fmt.Sprintf("- **Scan Time:** %s\n\n", driveInfo.ScanTime))

	// Category Breakdown
	sb.WriteString("## Content by Category\n\n")
	sb.WriteString("| Category | Size | Files | Folders | Target |\n")
	sb.WriteString("|----------|------|-------|---------|--------|\n")

	var totalSyncSize int64
	for _, cat := range driveInfo.Categories {
		sb.WriteString(fmt.Sprintf("| %s | %s | %d | %d | `%s` |\n",
			cat.Category, FormatBytes(cat.TotalSize), cat.FileCount, cat.FolderCount, cat.TargetPath))
		totalSyncSize += cat.TotalSize
	}
	sb.WriteString("\n")

	// Excluded Content
	if driveInfo.ExcludedCount > 0 {
		sb.WriteString("## Excluded Content\n\n")
		sb.WriteString(fmt.Sprintf("- **Excluded Folders:** %d\n", driveInfo.ExcludedCount))
		sb.WriteString(fmt.Sprintf("- **Excluded Size:** %s\n\n", FormatBytes(driveInfo.ExcludedSize)))

		sb.WriteString("| Folder | Size | Reason |\n")
		sb.WriteString("|--------|------|--------|\n")
		for _, folder := range driveInfo.Folders {
			if folder.IsExcluded {
				sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
					folder.Name, FormatBytes(folder.SizeBytes), folder.ExcludeReason))
			}
		}
		sb.WriteString("\n")
	}

	// Sync Recommendations
	sb.WriteString("## Sync Recommendations\n\n")
	sb.WriteString(fmt.Sprintf("**Total to sync:** %s\n\n", FormatBytes(totalSyncSize)))
	sb.WriteString("**Recommended order (largest first):**\n\n")

	for i, cat := range driveInfo.Categories {
		if i >= 5 {
			break
		}
		sb.WriteString(fmt.Sprintf("%d. %s (%s)\n", i+1, cat.Category, FormatBytes(cat.TotalSize)))
	}

	return sb.String(), nil
}

// Note: FormatBytes is defined in rclone.go and shared across the package

// ==========================================
// UNRAID Disaster Recovery Support
// ==========================================

// UNRAIDDriveInfo extends DriveInfo for UNRAID-specific information
type UNRAIDDriveInfo struct {
	Device         string `json:"device"`          // e.g., "/dev/sdb1"
	WindowsDevice  string `json:"windows_device"`  // e.g., "\\.\PHYSICALDRIVE1"
	MountPoint     string `json:"mount_point"`     // e.g., "/mnt/unraid"
	ArrayRole      string `json:"array_role"`      // "parity", "data", "cache"
	FileSystem     string `json:"file_system"`     // "xfs", "btrfs"
	EncryptionType string `json:"encryption_type"` // "none", "luks"
	SizeBytes      int64  `json:"size_bytes"`
	UsedBytes      int64  `json:"used_bytes"`
	SerialNumber   string `json:"serial_number"`
	IsMounted      bool   `json:"is_mounted"`
	IsAccessible   bool   `json:"is_accessible"`
	WSLMounted     bool   `json:"wsl_mounted"`
}

// UNRAIDRecoveryOptions for disaster recovery operations
type UNRAIDRecoveryOptions struct {
	DriveDevice    string `json:"drive_device"`    // e.g., "/dev/sdb1" or "\\.\PHYSICALDRIVE1"
	MountPoint     string `json:"mount_point"`     // e.g., "/mnt/unraid"
	FileSystem     string `json:"file_system"`     // auto-detect, "xfs", or "btrfs"
	ReadOnly       bool   `json:"read_only"`       // recommended for recovery
	EncryptionType string `json:"encryption_type"` // "none", "luks"
	DecryptionKey  string `json:"decryption_key"`  // for LUKS encrypted drives
}

// UNRAIDAppdataInfo represents a Docker appdata backup
type UNRAIDAppdataInfo struct {
	ContainerName string    `json:"container_name"`
	AppdataPath   string    `json:"appdata_path"`
	SizeBytes     int64     `json:"size_bytes"`
	LastModified  time.Time `json:"last_modified"`
	FileCount     int       `json:"file_count"`
}

// ScanUNRAIDDrives detects physical drives that may be UNRAID array disks
func (c *DataMigrationClient) ScanUNRAIDDrives(ctx context.Context) ([]UNRAIDDriveInfo, error) {
	var drives []UNRAIDDriveInfo

	if runtime.GOOS == "windows" {
		// Use PowerShell to get physical disks
		psCmd := `Get-PhysicalDisk | Select-Object DeviceId, FriendlyName, Size, MediaType, BusType, SerialNumber | ConvertTo-Json`
		cmd := exec.CommandContext(ctx, "powershell", "-Command", psCmd)
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to scan physical disks: %w", err)
		}

		var disks []struct {
			DeviceId     int    `json:"DeviceId"`
			FriendlyName string `json:"FriendlyName"`
			Size         int64  `json:"Size"`
			MediaType    string `json:"MediaType"`
			BusType      string `json:"BusType"`
			SerialNumber string `json:"SerialNumber"`
		}

		if err := json.Unmarshal(output, &disks); err != nil {
			// Try as single object
			var disk struct {
				DeviceId     int    `json:"DeviceId"`
				FriendlyName string `json:"FriendlyName"`
				Size         int64  `json:"Size"`
				MediaType    string `json:"MediaType"`
				BusType      string `json:"BusType"`
				SerialNumber string `json:"SerialNumber"`
			}
			if err := json.Unmarshal(output, &disk); err != nil {
				return nil, fmt.Errorf("failed to parse disk info: %w", err)
			}
			disks = append(disks, disk)
		}

		for _, disk := range disks {
			drive := UNRAIDDriveInfo{
				WindowsDevice: fmt.Sprintf(`\\.\PHYSICALDRIVE%d`, disk.DeviceId),
				SizeBytes:     disk.Size,
				SerialNumber:  disk.SerialNumber,
				ArrayRole:     "unknown", // Will be determined later
				FileSystem:    "unknown", // XFS/BTRFS detection requires mounting
			}

			// Check if this looks like an UNRAID drive (typically HDD/SSD via USB/SATA)
			if disk.BusType == "USB" || disk.BusType == "SATA" {
				drives = append(drives, drive)
			}
		}
	} else {
		// Linux: use lsblk
		cmd := exec.CommandContext(ctx, "lsblk", "-J", "-o", "NAME,SIZE,FSTYPE,MOUNTPOINT,SERIAL")
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("failed to scan drives: %w", err)
		}

		var lsblkOutput struct {
			Blockdevices []struct {
				Name       string `json:"name"`
				Size       string `json:"size"`
				Fstype     string `json:"fstype"`
				Mountpoint string `json:"mountpoint"`
				Serial     string `json:"serial"`
			} `json:"blockdevices"`
		}

		if err := json.Unmarshal(output, &lsblkOutput); err != nil {
			return nil, fmt.Errorf("failed to parse lsblk output: %w", err)
		}

		for _, dev := range lsblkOutput.Blockdevices {
			if dev.Fstype == "xfs" || dev.Fstype == "btrfs" || dev.Fstype == "crypto_LUKS" {
				drive := UNRAIDDriveInfo{
					Device:       fmt.Sprintf("/dev/%s", dev.Name),
					FileSystem:   dev.Fstype,
					MountPoint:   dev.Mountpoint,
					SerialNumber: dev.Serial,
					IsMounted:    dev.Mountpoint != "",
				}

				if dev.Fstype == "crypto_LUKS" {
					drive.EncryptionType = "luks"
				} else {
					drive.EncryptionType = "none"
				}

				drives = append(drives, drive)
			}
		}
	}

	return drives, nil
}

// MountUNRAIDDrive mounts an UNRAID drive via WSL2 (Windows) or directly (Linux)
func (c *DataMigrationClient) MountUNRAIDDrive(ctx context.Context, opts UNRAIDRecoveryOptions) (*UNRAIDDriveInfo, error) {
	if runtime.GOOS == "windows" {
		return c.mountUNRAIDDriveWSL(ctx, opts)
	}
	return c.mountUNRAIDDriveLinux(ctx, opts)
}

// mountUNRAIDDriveWSL mounts an UNRAID drive using WSL2.
// All exec calls use argument arrays (never shell string interpolation) to prevent injection.
func (c *DataMigrationClient) mountUNRAIDDriveWSL(ctx context.Context, opts UNRAIDRecoveryOptions) (*UNRAIDDriveInfo, error) {
	// Validate device path
	if err := sanitize.DevicePath(opts.DriveDevice); err != nil {
		return nil, fmt.Errorf("invalid device path: %w", err)
	}

	// Extract disk number from device path (e.g., \\.\PHYSICALDRIVE1 -> 1)
	diskNum := ""
	if strings.Contains(opts.DriveDevice, "PHYSICALDRIVE") {
		diskNum = strings.TrimPrefix(opts.DriveDevice, `\\.\PHYSICALDRIVE`)
	}

	if diskNum == "" {
		return nil, fmt.Errorf("invalid device path: %s (expected format: \\\\.\\PHYSICALDRIVE#)", opts.DriveDevice)
	}

	// Validate mount point
	if opts.MountPoint == "" {
		opts.MountPoint = "/mnt/unraid"
	}
	if err := sanitize.MountPoint(opts.MountPoint); err != nil {
		return nil, fmt.Errorf("invalid mount point: %w", err)
	}

	// Step 1: Mount physical disk to WSL in bare mode
	cmd := exec.CommandContext(ctx, "wsl", "--mount", opts.DriveDevice, "--bare")
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to mount to WSL: %w (output: %s)", err, string(output))
	}

	// Step 2: Create mount point in WSL
	cmd = exec.CommandContext(ctx, "wsl", "-u", "root", "mkdir", "-p", opts.MountPoint)
	cmd.Run() // Ignore error if exists

	// Step 3: Determine the device path in WSL (typically /dev/sdX1)
	// After --bare mount, the disk appears as /dev/sdX in WSL
	wslDevice := fmt.Sprintf("/dev/sd%c1", 'b'+diskNum[0]-'0') // Approximate mapping

	// Step 4: Detect filesystem if not specified
	fsType := opts.FileSystem
	if fsType == "" || fsType == "auto" {
		// Try to detect using blkid in WSL
		cmd = exec.CommandContext(ctx, "wsl", "-u", "root", "blkid", wslDevice, "-s", "TYPE", "-o", "value")
		output, err := cmd.Output()
		if err == nil {
			fsType = strings.TrimSpace(string(output))
		}
		if fsType == "" {
			fsType = "xfs" // Default for UNRAID
		}
	}

	// Validate filesystem type
	if err := sanitize.FileSystemType(fsType); err != nil {
		return nil, fmt.Errorf("unsupported filesystem: %w", err)
	}

	// Step 5: Handle LUKS encryption if needed
	if opts.EncryptionType == "luks" || fsType == "crypto_LUKS" {
		if opts.DecryptionKey == "" {
			return nil, fmt.Errorf("decryption key required for LUKS encrypted drive")
		}

		// Open LUKS container — pass key via stdin to avoid shell injection
		cmd = exec.CommandContext(ctx, "wsl", "-u", "root", "cryptsetup", "luksOpen", wslDevice, "unraid_crypt")
		cmd.Stdin = strings.NewReader(opts.DecryptionKey)
		if output, err := cmd.CombinedOutput(); err != nil {
			return nil, fmt.Errorf("failed to open LUKS: %w (output: %s)", err, string(output))
		}
		wslDevice = "/dev/mapper/unraid_crypt"
		fsType = "xfs" // Assume XFS inside LUKS
	}

	// Step 6: Mount the filesystem using argument array
	mountArgs := []string{"-u", "root", "mount", "-t", fsType}
	if opts.ReadOnly {
		mountArgs = append(mountArgs, "-o", "ro")
	}
	mountArgs = append(mountArgs, wslDevice, opts.MountPoint)

	cmd = exec.CommandContext(ctx, "wsl", mountArgs...)
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to mount filesystem: %w (output: %s)", err, string(output))
	}

	// Return drive info
	return &UNRAIDDriveInfo{
		Device:         wslDevice,
		WindowsDevice:  opts.DriveDevice,
		MountPoint:     opts.MountPoint,
		FileSystem:     fsType,
		EncryptionType: opts.EncryptionType,
		IsMounted:      true,
		WSLMounted:     true,
		IsAccessible:   true,
	}, nil
}

// mountUNRAIDDriveLinux mounts an UNRAID drive on native Linux.
// All inputs are validated before passing to exec.
func (c *DataMigrationClient) mountUNRAIDDriveLinux(ctx context.Context, opts UNRAIDRecoveryOptions) (*UNRAIDDriveInfo, error) {
	// Validate device path
	if err := sanitize.DevicePath(opts.DriveDevice); err != nil {
		return nil, fmt.Errorf("invalid device path: %w", err)
	}

	// Validate and default mount point
	if opts.MountPoint == "" {
		opts.MountPoint = "/mnt/unraid"
	}
	if err := sanitize.MountPoint(opts.MountPoint); err != nil {
		return nil, fmt.Errorf("invalid mount point: %w", err)
	}

	cmd := exec.CommandContext(ctx, "sudo", "mkdir", "-p", opts.MountPoint)
	cmd.Run()

	// Detect filesystem
	fsType := opts.FileSystem
	if fsType == "" || fsType == "auto" {
		cmd = exec.CommandContext(ctx, "blkid", opts.DriveDevice, "-s", "TYPE", "-o", "value")
		output, err := cmd.Output()
		if err == nil {
			fsType = strings.TrimSpace(string(output))
		}
		if fsType == "" {
			fsType = "xfs"
		}
	}

	// Validate filesystem type
	if err := sanitize.FileSystemType(fsType); err != nil {
		return nil, fmt.Errorf("unsupported filesystem: %w", err)
	}

	// Handle LUKS
	mountDevice := opts.DriveDevice
	if opts.EncryptionType == "luks" || fsType == "crypto_LUKS" {
		if opts.DecryptionKey == "" {
			return nil, fmt.Errorf("decryption key required for LUKS encrypted drive")
		}

		// Open LUKS — key via stdin
		cmd = exec.CommandContext(ctx, "sudo", "cryptsetup", "luksOpen", opts.DriveDevice, "unraid_crypt")
		cmd.Stdin = strings.NewReader(opts.DecryptionKey)
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to open LUKS: %w", err)
		}
		mountDevice = "/dev/mapper/unraid_crypt"
		fsType = "xfs"
	}

	// Mount
	mountArgs := []string{"mount", "-t", fsType}
	if opts.ReadOnly {
		mountArgs = append(mountArgs, "-o", "ro")
	}
	mountArgs = append(mountArgs, mountDevice, opts.MountPoint)

	cmd = exec.CommandContext(ctx, "sudo", mountArgs...)
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to mount: %w", err)
	}

	return &UNRAIDDriveInfo{
		Device:         opts.DriveDevice,
		MountPoint:     opts.MountPoint,
		FileSystem:     fsType,
		EncryptionType: opts.EncryptionType,
		IsMounted:      true,
		IsAccessible:   true,
	}, nil
}

// UnmountUNRAIDDrive unmounts an UNRAID drive
func (c *DataMigrationClient) UnmountUNRAIDDrive(ctx context.Context, mountPoint string) error {
	// Validate mount point
	if err := sanitize.MountPoint(mountPoint); err != nil {
		return fmt.Errorf("invalid mount point: %w", err)
	}

	if runtime.GOOS == "windows" {
		// Unmount in WSL — using argument array, not shell string
		cmd := exec.CommandContext(ctx, "wsl", "-u", "root", "umount", mountPoint)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to unmount: %w", err)
		}

		// Close LUKS if open
		exec.CommandContext(ctx, "wsl", "-u", "root", "cryptsetup", "luksClose", "unraid_crypt").Run()

		return nil
	}

	// Linux unmount
	cmd := exec.CommandContext(ctx, "sudo", "umount", mountPoint)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to unmount: %w", err)
	}

	// Close LUKS if open
	exec.CommandContext(ctx, "sudo", "cryptsetup", "luksClose", "unraid_crypt").Run()

	return nil
}

// ScanUNRAIDAppdata scans for Docker appdata on a mounted UNRAID drive
func (c *DataMigrationClient) ScanUNRAIDAppdata(ctx context.Context, mountPoint string) ([]UNRAIDAppdataInfo, error) {
	var appdata []UNRAIDAppdataInfo

	// UNRAID typically stores appdata in /mnt/user/appdata or /mnt/cache/appdata
	appdataPaths := []string{
		filepath.Join(mountPoint, "appdata"),
		filepath.Join(mountPoint, "mnt", "user", "appdata"),
		filepath.Join(mountPoint, "mnt", "cache", "appdata"),
	}

	for _, appdataPath := range appdataPaths {
		entries, err := os.ReadDir(appdataPath)
		if err != nil {
			continue // Path doesn't exist
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			containerPath := filepath.Join(appdataPath, entry.Name())
			info, err := entry.Info()
			if err != nil {
				continue
			}

			// Get size using rclone for accuracy
			var sizeBytes int64
			var fileCount int

			sizeCmd := exec.CommandContext(ctx, "rclone", "size", containerPath, "--json")
			if output, err := sizeCmd.Output(); err == nil {
				var sizeInfo struct {
					Count int64 `json:"count"`
					Bytes int64 `json:"bytes"`
				}
				if json.Unmarshal(output, &sizeInfo) == nil {
					sizeBytes = sizeInfo.Bytes
					fileCount = int(sizeInfo.Count)
				}
			}

			appdata = append(appdata, UNRAIDAppdataInfo{
				ContainerName: entry.Name(),
				AppdataPath:   containerPath,
				SizeBytes:     sizeBytes,
				LastModified:  info.ModTime(),
				FileCount:     fileCount,
			})
		}
	}

	return appdata, nil
}

// BackupUNRAIDAppdata backs up UNRAID Docker appdata to a destination
func (c *DataMigrationClient) BackupUNRAIDAppdata(ctx context.Context, source string, dest string, containerFilter []string) (*SyncJob, error) {
	// Use rclone client for the actual sync
	exclude := []string{
		"*.log",
		"*.tmp",
		"cache/**",
		"Cache/**",
	}

	return c.rcloneClient.StartSync(ctx, source, dest, false, false, exclude)
}

// RepairXFS attempts to repair an XFS filesystem (read-only check first)
func (c *DataMigrationClient) RepairXFS(ctx context.Context, device string, dryRun bool) (string, error) {
	if err := sanitize.DevicePath(device); err != nil {
		return "", fmt.Errorf("invalid device path: %w", err)
	}

	var args []string
	if dryRun {
		args = []string{"xfs_repair", "-n", device} // -n = no modify (dry run)
	} else {
		args = []string{"xfs_repair", device}
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Use argument array via wsl, not shell string interpolation
		wslArgs := append([]string{"-u", "root"}, args...)
		cmd = exec.CommandContext(ctx, "wsl", wslArgs...)
	} else {
		cmd = exec.CommandContext(ctx, "sudo", args...)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("xfs_repair failed: %w", err)
	}

	return string(output), nil
}

// RepairBTRFS attempts to repair a BTRFS filesystem (check mode)
func (c *DataMigrationClient) RepairBTRFS(ctx context.Context, device string, dryRun bool) (string, error) {
	if err := sanitize.DevicePath(device); err != nil {
		return "", fmt.Errorf("invalid device path: %w", err)
	}

	var args []string
	if dryRun {
		args = []string{"btrfs", "check", device}
	} else {
		args = []string{"btrfs", "check", "--repair", device}
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		// Use argument array via wsl, not shell string interpolation
		wslArgs := append([]string{"-u", "root"}, args...)
		cmd = exec.CommandContext(ctx, "wsl", wslArgs...)
	} else {
		cmd = exec.CommandContext(ctx, "sudo", args...)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("btrfs check failed: %w", err)
	}

	return string(output), nil
}

// ==========================================
// Cross-Drive Hash Index (Phase 3)
// ==========================================

// MasterHashIndex aggregates hash indexes from multiple drives
type MasterHashIndex struct {
	ID          string                   `json:"id"`
	Name        string                   `json:"name"`
	Sources     []string                 `json:"sources"` // List of source index IDs
	Files       map[string][]IndexedFile `json:"files"`   // hash -> files from all sources
	CreatedAt   time.Time                `json:"created_at"`
	TotalFiles  int                      `json:"total_files"`
	TotalSize   int64                    `json:"total_size"`
	UniqueFiles int                      `json:"unique_files"` // Files with unique hash
	UniqueSize  int64                    `json:"unique_size"`
	DupeFiles   int                      `json:"dupe_files"` // Total duplicate files
	DupeSpace   int64                    `json:"dupe_space"` // Space wasted by duplicates
}

// IndexedFile extends FileEntry with source drive information
type IndexedFile struct {
	Path        string    `json:"path"`
	SourceDrive string    `json:"source_drive"` // Drive letter or index ID
	Size        int64     `json:"size"`
	Hash        string    `json:"hash"`
	ModTime     time.Time `json:"mod_time"`
}

// CrossDriveDuplicate represents files duplicated across drives
type CrossDriveDuplicate struct {
	Hash        string        `json:"hash"`
	Size        int64         `json:"size"`
	SizeHuman   string        `json:"size_human"`
	Files       []IndexedFile `json:"files"`
	DriveCount  int           `json:"drive_count"`  // Number of drives with this file
	WastedSpace int64         `json:"wasted_space"` // (count - 1) * size
}

// DeduplicationReport summarizes cross-drive deduplication analysis
type DeduplicationReport struct {
	MasterIndexID      string                `json:"master_index_id"`
	SourceCount        int                   `json:"source_count"`
	TotalFiles         int                   `json:"total_files"`
	TotalSize          int64                 `json:"total_size"`
	TotalSizeHuman     string                `json:"total_size_human"`
	UniqueFiles        int                   `json:"unique_files"`
	UniqueSize         int64                 `json:"unique_size"`
	UniqueSizeHuman    string                `json:"unique_size_human"`
	DuplicateFiles     int                   `json:"duplicate_files"`
	DuplicateSize      int64                 `json:"duplicate_size"`
	DuplicateSizeHuman string                `json:"duplicate_size_human"`
	WastedSpace        int64                 `json:"wasted_space"`
	WastedSpaceHuman   string                `json:"wasted_space_human"`
	SavingsPercent     float64               `json:"savings_percent"`
	TopDuplicates      []CrossDriveDuplicate `json:"top_duplicates"` // Top 20 by wasted space
	ByDrive            map[string]DriveStats `json:"by_drive"`
	GeneratedAt        time.Time             `json:"generated_at"`
}

// DriveStats shows per-drive statistics
type DriveStats struct {
	Drive       string `json:"drive"`
	TotalFiles  int    `json:"total_files"`
	TotalSize   int64  `json:"total_size"`
	UniqueFiles int    `json:"unique_files"` // Files only on this drive
	UniqueSize  int64  `json:"unique_size"`
	SharedFiles int    `json:"shared_files"` // Files also on other drives
}

// MergeHashIndexes combines multiple hash indexes into a master index
func (c *DataMigrationClient) MergeHashIndexes(ctx context.Context, indexIDs []string, masterName string) (*MasterHashIndex, error) {
	if len(indexIDs) == 0 {
		return nil, fmt.Errorf("no index IDs provided")
	}

	if masterName == "" {
		masterName = fmt.Sprintf("master_%d", time.Now().Unix())
	}

	master := &MasterHashIndex{
		ID:        fmt.Sprintf("master_%d", time.Now().UnixNano()),
		Name:      masterName,
		Sources:   indexIDs,
		Files:     make(map[string][]IndexedFile),
		CreatedAt: time.Now(),
	}

	// Load and merge each source index
	for _, indexID := range indexIDs {
		index, err := c.LoadIndex(indexID)
		if err != nil {
			return nil, fmt.Errorf("failed to load index %s: %w", indexID, err)
		}

		// Add files from this index
		for hash, entries := range index.Files {
			for _, entry := range entries {
				indexedFile := IndexedFile{
					Path:        entry.Path,
					SourceDrive: index.Path, // Use source path as drive identifier
					Size:        entry.Size,
					Hash:        entry.Hash,
					ModTime:     entry.ModTime,
				}
				master.Files[hash] = append(master.Files[hash], indexedFile)
				master.TotalFiles++
				master.TotalSize += entry.Size
			}
		}
	}

	// Calculate unique vs duplicate stats
	for hash, files := range master.Files {
		if len(files) == 1 {
			master.UniqueFiles++
			master.UniqueSize += files[0].Size
		} else {
			// Multiple files with same hash = duplicates
			master.DupeFiles += len(files)
			// Wasted space is (count - 1) * size (keeping one copy)
			master.DupeSpace += int64(len(files)-1) * files[0].Size
		}
		_ = hash // silence unused warning
	}

	// Save master index
	if err := c.saveMasterIndex(master); err != nil {
		return nil, fmt.Errorf("failed to save master index: %w", err)
	}

	return master, nil
}

// FindCrossDriveDuplicates finds files that exist on multiple drives
func (c *DataMigrationClient) FindCrossDriveDuplicates(ctx context.Context, masterIndexID string) ([]CrossDriveDuplicate, error) {
	master, err := c.LoadMasterIndex(masterIndexID)
	if err != nil {
		return nil, fmt.Errorf("failed to load master index: %w", err)
	}

	var duplicates []CrossDriveDuplicate

	for hash, files := range master.Files {
		if len(files) <= 1 {
			continue
		}

		// Count unique drives
		drives := make(map[string]bool)
		for _, f := range files {
			drives[f.SourceDrive] = true
		}

		// Only include if on multiple drives
		if len(drives) > 1 {
			dupe := CrossDriveDuplicate{
				Hash:        hash,
				Size:        files[0].Size,
				SizeHuman:   FormatBytes(files[0].Size),
				Files:       files,
				DriveCount:  len(drives),
				WastedSpace: int64(len(files)-1) * files[0].Size,
			}
			duplicates = append(duplicates, dupe)
		}
	}

	// Sort by wasted space (largest first)
	for i := 0; i < len(duplicates)-1; i++ {
		for j := i + 1; j < len(duplicates); j++ {
			if duplicates[j].WastedSpace > duplicates[i].WastedSpace {
				duplicates[i], duplicates[j] = duplicates[j], duplicates[i]
			}
		}
	}

	return duplicates, nil
}

// GenerateDeduplicationReport creates a comprehensive dedup report
func (c *DataMigrationClient) GenerateDeduplicationReport(ctx context.Context, masterIndexID string) (*DeduplicationReport, error) {
	master, err := c.LoadMasterIndex(masterIndexID)
	if err != nil {
		return nil, fmt.Errorf("failed to load master index: %w", err)
	}

	duplicates, err := c.FindCrossDriveDuplicates(ctx, masterIndexID)
	if err != nil {
		return nil, fmt.Errorf("failed to find duplicates: %w", err)
	}

	// Calculate per-drive stats
	driveStats := make(map[string]*DriveStats)
	for _, files := range master.Files {
		isUnique := len(files) == 1
		for _, f := range files {
			if driveStats[f.SourceDrive] == nil {
				driveStats[f.SourceDrive] = &DriveStats{Drive: f.SourceDrive}
			}
			ds := driveStats[f.SourceDrive]
			ds.TotalFiles++
			ds.TotalSize += f.Size
			if isUnique {
				ds.UniqueFiles++
				ds.UniqueSize += f.Size
			} else {
				ds.SharedFiles++
			}
		}
	}

	// Convert to non-pointer map
	byDrive := make(map[string]DriveStats)
	for k, v := range driveStats {
		byDrive[k] = *v
	}

	// Calculate totals
	var totalDupeSize int64
	for _, d := range duplicates {
		totalDupeSize += d.Size * int64(len(d.Files))
	}

	// Top 20 duplicates
	topDupes := duplicates
	if len(topDupes) > 20 {
		topDupes = topDupes[:20]
	}

	report := &DeduplicationReport{
		MasterIndexID:      masterIndexID,
		SourceCount:        len(master.Sources),
		TotalFiles:         master.TotalFiles,
		TotalSize:          master.TotalSize,
		TotalSizeHuman:     FormatBytes(master.TotalSize),
		UniqueFiles:        master.UniqueFiles,
		UniqueSize:         master.UniqueSize,
		UniqueSizeHuman:    FormatBytes(master.UniqueSize),
		DuplicateFiles:     len(duplicates),
		DuplicateSize:      totalDupeSize,
		DuplicateSizeHuman: FormatBytes(totalDupeSize),
		WastedSpace:        master.DupeSpace,
		WastedSpaceHuman:   FormatBytes(master.DupeSpace),
		TopDuplicates:      topDupes,
		ByDrive:            byDrive,
		GeneratedAt:        time.Now(),
	}

	if master.TotalSize > 0 {
		report.SavingsPercent = float64(master.DupeSpace) / float64(master.TotalSize) * 100
	}

	return report, nil
}

// saveMasterIndex persists a master index to disk
func (c *DataMigrationClient) saveMasterIndex(master *MasterHashIndex) error {
	indexDir := filepath.Join(c.indexDir, "indexes")
	if err := os.MkdirAll(indexDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(master, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(indexDir, master.ID+".json"), data, 0644)
}

// LoadMasterIndex loads a master index from disk
func (c *DataMigrationClient) LoadMasterIndex(masterID string) (*MasterHashIndex, error) {
	indexFile := filepath.Join(c.indexDir, "indexes", masterID+".json")
	data, err := os.ReadFile(indexFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read master index file: %w", err)
	}

	var master MasterHashIndex
	if err := json.Unmarshal(data, &master); err != nil {
		return nil, fmt.Errorf("failed to parse master index: %w", err)
	}

	return &master, nil
}

// ListMasterIndexes returns all available master indexes
func (c *DataMigrationClient) ListMasterIndexes() ([]string, error) {
	indexDir := filepath.Join(c.indexDir, "indexes")
	files, err := os.ReadDir(indexDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, err
	}

	var masters []string
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "master_") && strings.HasSuffix(f.Name(), ".json") {
			masters = append(masters, strings.TrimSuffix(f.Name(), ".json"))
		}
	}
	return masters, nil
}

// ==========================================
// Incremental Hash Indexing (Phase 3)
// ==========================================

// IncrementalIndexOptions for efficient re-indexing
type IncrementalIndexOptions struct {
	ExistingIndexID string   `json:"existing_index_id"` // Index to update
	CheckModTime    bool     `json:"check_mod_time"`    // Skip if mod time unchanged
	NewFilesOnly    bool     `json:"new_files_only"`    // Only hash files not in index
	ParallelWorkers int      `json:"parallel_workers"`  // Concurrent hashing (default: 4)
	FileTypes       []string `json:"file_types"`        // Filter by extension
	MinSizeMB       int      `json:"min_size_mb"`       // Minimum file size
}

// HashResult for parallel hashing
type HashResult struct {
	Entry FileEntry
	Err   error
}

// UpdateHashIndex incrementally updates an existing hash index
func (c *DataMigrationClient) UpdateHashIndex(ctx context.Context, sourcePath string, opts IncrementalIndexOptions) (*HashIndex, error) {
	// Load existing index or create new
	var index *HashIndex
	var existingFiles map[string]FileEntry // path -> entry for quick lookup

	if opts.ExistingIndexID != "" {
		var err error
		index, err = c.LoadIndex(opts.ExistingIndexID)
		if err != nil {
			return nil, fmt.Errorf("failed to load existing index: %w", err)
		}

		// Build lookup map from existing entries
		existingFiles = make(map[string]FileEntry)
		for _, entries := range index.Files {
			for _, e := range entries {
				existingFiles[e.Path] = e
			}
		}
	} else {
		index = &HashIndex{
			ID:        fmt.Sprintf("index_%d", time.Now().UnixNano()),
			Name:      fmt.Sprintf("incremental_%d", time.Now().Unix()),
			Path:      sourcePath,
			Files:     make(map[string][]FileEntry),
			CreatedAt: time.Now(),
		}
		existingFiles = make(map[string]FileEntry)
	}

	workers := opts.ParallelWorkers
	if workers <= 0 {
		workers = 4
	}
	if workers > 16 {
		workers = 16
	}

	minSize := int64(opts.MinSizeMB) * 1024 * 1024

	// Collect files to hash
	var filesToHash []string
	var skippedCount int

	err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		// Check file size
		if info.Size() < minSize {
			return nil
		}

		// Check file type filter
		if len(opts.FileTypes) > 0 {
			ext := strings.ToLower(filepath.Ext(path))
			matched := false
			for _, ft := range opts.FileTypes {
				if ext == strings.ToLower(ft) || ext == "."+strings.ToLower(ft) {
					matched = true
					break
				}
			}
			if !matched {
				return nil
			}
		}

		// Check if we can skip
		if existing, ok := existingFiles[path]; ok {
			if opts.CheckModTime && existing.ModTime.Equal(info.ModTime()) && existing.Size == info.Size() {
				skippedCount++
				return nil // Unchanged, skip rehashing
			}
			if opts.NewFilesOnly {
				skippedCount++
				return nil // Already indexed, skip
			}
		}

		filesToHash = append(filesToHash, path)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	// Parallel hashing
	jobs := make(chan string, len(filesToHash))
	results := make(chan HashResult, len(filesToHash))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range jobs {
				info, err := os.Stat(path)
				if err != nil {
					results <- HashResult{Err: err}
					continue
				}

				hash, err := c.calculateFileHash(path)
				if err != nil {
					results <- HashResult{Err: err}
					continue
				}

				results <- HashResult{
					Entry: FileEntry{
						Path:    path,
						Size:    info.Size(),
						Hash:    hash,
						ModTime: info.ModTime(),
					},
				}
			}
		}()
	}

	// Send jobs
	go func() {
		for _, path := range filesToHash {
			jobs <- path
		}
		close(jobs)
	}()

	// Collect results
	go func() {
		wg.Wait()
		close(results)
	}()

	// Process results
	var newCount, errorCount int
	for result := range results {
		if result.Err != nil {
			errorCount++
			continue
		}

		// Remove old entry if exists (rehashing)
		if existing, ok := existingFiles[result.Entry.Path]; ok {
			// Remove from old hash bucket
			oldEntries := index.Files[existing.Hash]
			for i, e := range oldEntries {
				if e.Path == result.Entry.Path {
					index.Files[existing.Hash] = append(oldEntries[:i], oldEntries[i+1:]...)
					break
				}
			}
		}

		// Add new/updated entry
		index.Files[result.Entry.Hash] = append(index.Files[result.Entry.Hash], result.Entry)
		newCount++
	}

	// Recalculate totals
	index.FileCount = 0
	index.TotalSize = 0
	for _, entries := range index.Files {
		index.FileCount += len(entries)
		for _, e := range entries {
			index.TotalSize += e.Size
		}
	}

	// Save updated index
	c.mu.Lock()
	c.hashIndexes[index.ID] = index
	c.mu.Unlock()

	if err := c.saveIndex(index); err != nil {
		return nil, fmt.Errorf("failed to save index: %w", err)
	}

	return index, nil
}
