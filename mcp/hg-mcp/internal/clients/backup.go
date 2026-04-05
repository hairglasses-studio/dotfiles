// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// BackupClient provides backup automation capabilities
type BackupClient struct {
	configPath string
	backupRoot string
}

// BackupConfig represents backup configuration
type BackupConfig struct {
	BackupRoot  string             `json:"backup_root"`
	Retention   int                `json:"retention_days"`
	Compression bool               `json:"compression"`
	Projects    []BackupProjectDef `json:"projects"`
}

// BackupProjectDef defines a project to back up
type BackupProjectDef struct {
	Name       string   `json:"name"`
	SourcePath string   `json:"source_path"`
	Exclude    []string `json:"exclude,omitempty"`
	Schedule   string   `json:"schedule,omitempty"` // daily, weekly, manual
}

// BackupInfo represents information about a backup
type BackupInfo struct {
	ID          string    `json:"id"`
	ProjectName string    `json:"project_name"`
	SourcePath  string    `json:"source_path"`
	BackupPath  string    `json:"backup_path"`
	SizeBytes   int64     `json:"size_bytes"`
	FileCount   int       `json:"file_count"`
	Timestamp   time.Time `json:"timestamp"`
	Duration    string    `json:"duration"`
	Status      string    `json:"status"` // completed, failed, in_progress
	Error       string    `json:"error,omitempty"`
}

// BackupStatus represents overall backup status
type BackupStatus struct {
	TotalBackups   int             `json:"total_backups"`
	TotalSizeBytes int64           `json:"total_size_bytes"`
	LastBackup     *BackupInfo     `json:"last_backup,omitempty"`
	Projects       []ProjectStatus `json:"projects"`
}

// ProjectStatus represents backup status for a project
type ProjectStatus struct {
	Name        string      `json:"name"`
	LastBackup  *BackupInfo `json:"last_backup,omitempty"`
	BackupCount int         `json:"backup_count"`
	TotalSize   int64       `json:"total_size_bytes"`
}

// RestoreResult represents the result of a restore operation
type RestoreResult struct {
	BackupID    string `json:"backup_id"`
	RestorePath string `json:"restore_path"`
	FileCount   int    `json:"file_count"`
	SizeBytes   int64  `json:"size_bytes"`
	Duration    string `json:"duration"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
}

// NewBackupClient creates a new backup client
func NewBackupClient() (*BackupClient, error) {
	configPath := os.Getenv("BACKUP_CONFIG_PATH")
	if configPath == "" {
		home, _ := os.UserHomeDir()
		configPath = filepath.Join(home, ".aftrs", "backup.json")
	}

	backupRoot := os.Getenv("BACKUP_ROOT")
	if backupRoot == "" {
		home, _ := os.UserHomeDir()
		backupRoot = filepath.Join(home, "Backups", "aftrs-projects")
	}

	return &BackupClient{
		configPath: configPath,
		backupRoot: backupRoot,
	}, nil
}

// loadConfig loads the backup configuration
func (c *BackupClient) loadConfig() (*BackupConfig, error) {
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return default config
			return &BackupConfig{
				BackupRoot:  c.backupRoot,
				Retention:   30,
				Compression: true,
				Projects:    []BackupProjectDef{},
			}, nil
		}
		return nil, err
	}

	var config BackupConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// saveConfig saves the backup configuration
func (c *BackupClient) saveConfig(config *BackupConfig) error {
	if err := os.MkdirAll(filepath.Dir(c.configPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(c.configPath, data, 0644)
}

// BackupProject creates a backup of a project
func (c *BackupClient) BackupProject(ctx context.Context, projectName, sourcePath string, exclude []string) (*BackupInfo, error) {
	info := &BackupInfo{
		ProjectName: projectName,
		SourcePath:  sourcePath,
		Status:      "in_progress",
		Timestamp:   time.Now(),
	}

	// Validate source path exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		info.Status = "failed"
		info.Error = fmt.Sprintf("source path does not exist: %s", sourcePath)
		return info, nil
	}

	// Create backup directory
	backupDir := filepath.Join(c.backupRoot, projectName)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		info.Status = "failed"
		info.Error = fmt.Sprintf("failed to create backup directory: %v", err)
		return info, nil
	}

	// Generate backup ID and path
	info.ID = time.Now().Format("20060102-150405")
	backupPath := filepath.Join(backupDir, fmt.Sprintf("%s_%s.tar.gz", projectName, info.ID))
	info.BackupPath = backupPath

	start := time.Now()

	// Build tar command
	args := []string{"-czf", backupPath}

	// Add excludes
	for _, pattern := range exclude {
		args = append(args, "--exclude", pattern)
	}

	// Add source directory
	args = append(args, "-C", filepath.Dir(sourcePath), filepath.Base(sourcePath))

	// Execute tar
	cmd := exec.CommandContext(ctx, "tar", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		info.Status = "failed"
		info.Error = fmt.Sprintf("backup failed: %v - %s", err, string(output))
		return info, nil
	}

	info.Duration = time.Since(start).Round(time.Millisecond).String()

	// Get backup file info
	if stat, err := os.Stat(backupPath); err == nil {
		info.SizeBytes = stat.Size()
	}

	// Count files (approximate from source)
	info.FileCount = countFiles(sourcePath, exclude)

	info.Status = "completed"
	return info, nil
}

// GetBackupStatus returns overall backup status
func (c *BackupClient) GetBackupStatus(ctx context.Context) (*BackupStatus, error) {
	status := &BackupStatus{
		Projects: []ProjectStatus{},
	}

	// List project directories
	entries, err := os.ReadDir(c.backupRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return status, nil
		}
		return nil, err
	}

	var lastBackup *BackupInfo
	var lastBackupTime time.Time

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectName := entry.Name()
		projectDir := filepath.Join(c.backupRoot, projectName)

		projStatus := ProjectStatus{
			Name: projectName,
		}

		// List backups for this project
		backups, _ := os.ReadDir(projectDir)
		for _, backup := range backups {
			if backup.IsDir() || !strings.HasSuffix(backup.Name(), ".tar.gz") {
				continue
			}

			projStatus.BackupCount++

			backupPath := filepath.Join(projectDir, backup.Name())
			info, _ := backup.Info()
			if info != nil {
				projStatus.TotalSize += info.Size()
				status.TotalSizeBytes += info.Size()

				// Parse backup info
				backupInfo := &BackupInfo{
					ProjectName: projectName,
					BackupPath:  backupPath,
					SizeBytes:   info.Size(),
					Timestamp:   info.ModTime(),
					Status:      "completed",
				}

				// Extract ID from filename
				name := strings.TrimSuffix(backup.Name(), ".tar.gz")
				parts := strings.Split(name, "_")
				if len(parts) >= 2 {
					backupInfo.ID = parts[len(parts)-1]
				}

				if info.ModTime().After(lastBackupTime) {
					lastBackupTime = info.ModTime()
					lastBackup = backupInfo
				}

				if projStatus.LastBackup == nil || info.ModTime().After(projStatus.LastBackup.Timestamp) {
					projStatus.LastBackup = backupInfo
				}
			}

			status.TotalBackups++
		}

		status.Projects = append(status.Projects, projStatus)
	}

	status.LastBackup = lastBackup

	return status, nil
}

// ListBackups returns list of backups for a project
func (c *BackupClient) ListBackups(ctx context.Context, projectName string) ([]BackupInfo, error) {
	backups := []BackupInfo{}

	projectDir := filepath.Join(c.backupRoot, projectName)
	entries, err := os.ReadDir(projectDir)
	if err != nil {
		if os.IsNotExist(err) {
			return backups, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tar.gz") {
			continue
		}

		backupPath := filepath.Join(projectDir, entry.Name())
		info, err := entry.Info()
		if err != nil {
			continue
		}

		backupInfo := BackupInfo{
			ProjectName: projectName,
			BackupPath:  backupPath,
			SizeBytes:   info.Size(),
			Timestamp:   info.ModTime(),
			Status:      "completed",
		}

		// Extract ID from filename
		name := strings.TrimSuffix(entry.Name(), ".tar.gz")
		parts := strings.Split(name, "_")
		if len(parts) >= 2 {
			backupInfo.ID = parts[len(parts)-1]
		}

		backups = append(backups, backupInfo)
	}

	// Sort by timestamp descending
	sort.Slice(backups, func(i, j int) bool {
		return backups[i].Timestamp.After(backups[j].Timestamp)
	})

	return backups, nil
}

// RestoreBackup restores a backup to a target path
func (c *BackupClient) RestoreBackup(ctx context.Context, projectName, backupID, targetPath string) (*RestoreResult, error) {
	// Validate targetPath — must be absolute, no traversal
	if targetPath == "" {
		return nil, fmt.Errorf("target path is required")
	}
	if !filepath.IsAbs(targetPath) {
		return nil, fmt.Errorf("target path must be absolute: %s", targetPath)
	}
	if strings.Contains(targetPath, "..") {
		return nil, fmt.Errorf("target path must not contain '..': %s", targetPath)
	}

	result := &RestoreResult{
		BackupID:    backupID,
		RestorePath: targetPath,
	}

	// Find the backup file
	projectDir := filepath.Join(c.backupRoot, projectName)
	var backupPath string

	entries, err := os.ReadDir(projectDir)
	if err != nil {
		result.Error = fmt.Sprintf("project not found: %s", projectName)
		return result, nil
	}

	for _, entry := range entries {
		if strings.Contains(entry.Name(), backupID) && strings.HasSuffix(entry.Name(), ".tar.gz") {
			backupPath = filepath.Join(projectDir, entry.Name())
			break
		}
	}

	if backupPath == "" {
		result.Error = fmt.Sprintf("backup not found: %s", backupID)
		return result, nil
	}

	// Get backup size
	if info, err := os.Stat(backupPath); err == nil {
		result.SizeBytes = info.Size()
	}

	// Create target directory
	if err := os.MkdirAll(targetPath, 0755); err != nil {
		result.Error = fmt.Sprintf("failed to create target directory: %v", err)
		return result, nil
	}

	start := time.Now()

	// Extract backup
	cmd := exec.CommandContext(ctx, "tar", "-xzf", backupPath, "-C", targetPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Error = fmt.Sprintf("restore failed: %v - %s", err, string(output))
		return result, nil
	}

	result.Duration = time.Since(start).Round(time.Millisecond).String()
	result.Success = true

	// Count restored files
	result.FileCount = countFiles(targetPath, nil)

	return result, nil
}

// DeleteBackup deletes a specific backup
func (c *BackupClient) DeleteBackup(ctx context.Context, projectName, backupID string) error {
	projectDir := filepath.Join(c.backupRoot, projectName)

	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return fmt.Errorf("project not found: %s", projectName)
	}

	for _, entry := range entries {
		if strings.Contains(entry.Name(), backupID) && strings.HasSuffix(entry.Name(), ".tar.gz") {
			backupPath := filepath.Join(projectDir, entry.Name())
			return os.Remove(backupPath)
		}
	}

	return fmt.Errorf("backup not found: %s", backupID)
}

// CleanupOldBackups removes backups older than retention days
func (c *BackupClient) CleanupOldBackups(ctx context.Context, retentionDays int) (int, error) {
	if retentionDays <= 0 {
		retentionDays = 30
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	deleted := 0

	entries, err := os.ReadDir(c.backupRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectDir := filepath.Join(c.backupRoot, entry.Name())
		backups, _ := os.ReadDir(projectDir)

		for _, backup := range backups {
			if backup.IsDir() || !strings.HasSuffix(backup.Name(), ".tar.gz") {
				continue
			}

			info, err := backup.Info()
			if err != nil {
				continue
			}

			if info.ModTime().Before(cutoff) {
				backupPath := filepath.Join(projectDir, backup.Name())
				if err := os.Remove(backupPath); err == nil {
					deleted++
				}
			}
		}
	}

	return deleted, nil
}

// AddProject adds a project to the backup configuration
func (c *BackupClient) AddProject(projectName, sourcePath string, exclude []string, schedule string) error {
	config, err := c.loadConfig()
	if err != nil {
		return err
	}

	// Check if project already exists
	for i, p := range config.Projects {
		if p.Name == projectName {
			// Update existing
			config.Projects[i] = BackupProjectDef{
				Name:       projectName,
				SourcePath: sourcePath,
				Exclude:    exclude,
				Schedule:   schedule,
			}
			return c.saveConfig(config)
		}
	}

	// Add new project
	config.Projects = append(config.Projects, BackupProjectDef{
		Name:       projectName,
		SourcePath: sourcePath,
		Exclude:    exclude,
		Schedule:   schedule,
	})

	return c.saveConfig(config)
}

// RemoveProject removes a project from the backup configuration
func (c *BackupClient) RemoveProject(projectName string) error {
	config, err := c.loadConfig()
	if err != nil {
		return err
	}

	for i, p := range config.Projects {
		if p.Name == projectName {
			config.Projects = append(config.Projects[:i], config.Projects[i+1:]...)
			return c.saveConfig(config)
		}
	}

	return fmt.Errorf("project not found: %s", projectName)
}

// GetConfig returns the current backup configuration
func (c *BackupClient) GetConfig() (*BackupConfig, error) {
	return c.loadConfig()
}

// countFiles counts files in a directory
func countFiles(dir string, exclude []string) int {
	count := 0
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		// Check excludes
		for _, pattern := range exclude {
			if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
				return nil
			}
		}

		count++
		return nil
	})
	return count
}
