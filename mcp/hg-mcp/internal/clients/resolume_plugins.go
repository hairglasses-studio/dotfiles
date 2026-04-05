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
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// ResolumePluginType represents the type of Resolume plugin
type ResolumePluginType string

const (
	ResolumePluginTypeFFGLEffect ResolumePluginType = "ffgl_effect"
	ResolumePluginTypeFFGLSource ResolumePluginType = "ffgl_source"
	ResolumePluginTypeISF        ResolumePluginType = "isf_shader"
	ResolumePluginTypeJuiceBar   ResolumePluginType = "juicebar"
)

// ResolumePluginSource represents where a plugin came from
type ResolumePluginSource string

const (
	ResolumePluginSourceLocal     ResolumePluginSource = "local"
	ResolumePluginSourceS3        ResolumePluginSource = "s3"
	ResolumePluginSourceJuiceBar  ResolumePluginSource = "juicebar"
	ResolumePluginSourceSpackOMat ResolumePluginSource = "spack-o-mat"
	ResolumePluginSourceISF       ResolumePluginSource = "isf"
	ResolumePluginSourceCustom    ResolumePluginSource = "custom"
)

// ResolumePluginInfo represents metadata about a plugin
type ResolumePluginInfo struct {
	Name            string               `json:"name"`
	Filename        string               `json:"filename"`
	Type            ResolumePluginType   `json:"type"`
	Source          ResolumePluginSource `json:"source"`
	Version         string               `json:"version,omitempty"`
	Platform        string               `json:"platform"`
	Architecture    string               `json:"architecture,omitempty"`
	LocalPath       string               `json:"local_path,omitempty"`
	S3Key           string               `json:"s3_key,omitempty"`
	SHA256          string               `json:"sha256,omitempty"`
	Size            int64                `json:"size"`
	UploadedAt      *time.Time           `json:"uploaded_at,omitempty"`
	UploadedBy      string               `json:"uploaded_by,omitempty"`
	Tags            []string             `json:"tags,omitempty"`
	ResolumeVersion string               `json:"resolume_version,omitempty"`
	IsJuiceBar      bool                 `json:"juicebar"`
	ModifiedAt      time.Time            `json:"modified_at"`
}

// ResolumePluginCatalog represents the S3 plugin catalog
type ResolumePluginCatalog struct {
	Version     int                  `json:"version"`
	LastUpdated time.Time            `json:"last_updated"`
	Plugins     []ResolumePluginInfo `json:"plugins"`
}

// ResolumePluginsClient provides Resolume plugin management
type ResolumePluginsClient struct {
	s3Client   *s3.Client
	bucketName string
	homeDir    string
	platform   string
}

// NewResolumePluginsClient creates a new Resolume plugins client
func NewResolumePluginsClient() (*ResolumePluginsClient, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	bucketName := os.Getenv("RESOLUME_PLUGINS_BUCKET")
	if bucketName == "" {
		bucketName = "aftrs-resolume-plugins"
	}

	platform := runtime.GOOS
	if platform == "darwin" {
		platform = "macos"
	}

	return &ResolumePluginsClient{
		s3Client:   s3.NewFromConfig(cfg),
		bucketName: bucketName,
		homeDir:    homeDir,
		platform:   platform,
	}, nil
}

// GetPluginPaths returns all Resolume plugin directories
func (c *ResolumePluginsClient) GetPluginPaths() map[string]string {
	paths := map[string]string{}

	if c.platform == "macos" {
		paths["extra_effects"] = filepath.Join(c.homeDir, "Documents", "Resolume Arena", "Extra Effects")
		paths["juicebar"] = filepath.Join(c.homeDir, "Library", "Application Support", "Juicebar", "plugins")
		paths["app_support"] = filepath.Join(c.homeDir, "Library", "Application Support", "Resolume Arena")
		paths["builtin_vfx"] = "/Applications/Resolume Arena/plugins/vfx"
		paths["builtin_sources"] = "/Applications/Resolume Arena/plugins/sources"
	} else if c.platform == "windows" {
		paths["extra_effects"] = filepath.Join(c.homeDir, "Documents", "Resolume Arena", "Extra Effects")
		paths["juicebar"] = filepath.Join(c.homeDir, "AppData", "Local", "Juicebar", "plugins")
		paths["app_data"] = filepath.Join(c.homeDir, "AppData", "Local", "Resolume Arena")
		paths["builtin_vfx"] = "C:\\Program Files\\Resolume Arena\\plugins\\vfx"
		paths["builtin_sources"] = "C:\\Program Files\\Resolume Arena\\plugins\\sources"
	}

	return paths
}

// ScanLocalPlugins scans all local plugin directories
func (c *ResolumePluginsClient) ScanLocalPlugins(ctx context.Context) ([]ResolumePluginInfo, error) {
	plugins := []ResolumePluginInfo{}
	paths := c.GetPluginPaths()

	for location, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			continue
		}

		err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}

			ext := strings.ToLower(filepath.Ext(filePath))
			pluginType, isPlugin := c.getPluginType(ext)
			if !isPlugin {
				return nil
			}

			source := c.determineSource(location, filePath)

			plugin := ResolumePluginInfo{
				Name:       strings.TrimSuffix(info.Name(), ext),
				Filename:   info.Name(),
				Type:       pluginType,
				Source:     source,
				Platform:   c.platform,
				LocalPath:  filePath,
				Size:       info.Size(),
				IsJuiceBar: location == "juicebar",
				ModifiedAt: info.ModTime(),
			}

			// Calculate SHA256 for verification
			if hash, err := c.calculateSHA256(filePath); err == nil {
				plugin.SHA256 = hash
			}

			plugins = append(plugins, plugin)
			return nil
		})

		if err != nil {
			continue
		}
	}

	return plugins, nil
}

func (c *ResolumePluginsClient) getPluginType(ext string) (ResolumePluginType, bool) {
	switch ext {
	case ".bundle", ".dll":
		return ResolumePluginTypeFFGLEffect, true
	case ".isf", ".fs":
		return ResolumePluginTypeISF, true
	default:
		return "", false
	}
}

func (c *ResolumePluginsClient) determineSource(location, path string) ResolumePluginSource {
	if location == "juicebar" {
		return ResolumePluginSourceJuiceBar
	}

	lowerPath := strings.ToLower(path)
	if strings.Contains(lowerPath, "spack") {
		return ResolumePluginSourceSpackOMat
	}
	if strings.Contains(lowerPath, "isf") {
		return ResolumePluginSourceISF
	}

	return ResolumePluginSourceLocal
}

func (c *ResolumePluginsClient) calculateSHA256(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

// ListS3Plugins lists plugins in the S3 bucket
func (c *ResolumePluginsClient) ListS3Plugins(ctx context.Context, prefix string) ([]ResolumePluginInfo, error) {
	plugins := []ResolumePluginInfo{}

	if prefix == "" {
		prefix = "ffgl/"
	}

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(c.bucketName),
		Prefix: aws.String(prefix),
	}

	paginator := s3.NewListObjectsV2Paginator(c.s3Client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list S3 objects: %w", err)
		}

		for _, obj := range page.Contents {
			key := aws.ToString(obj.Key)

			// Skip metadata JSON files
			if strings.HasSuffix(key, ".json") {
				continue
			}

			// Skip directories
			if strings.HasSuffix(key, "/") {
				continue
			}

			ext := strings.ToLower(filepath.Ext(key))
			pluginType, isPlugin := c.getPluginType(ext)
			if !isPlugin {
				continue
			}

			plugin := ResolumePluginInfo{
				Name:       strings.TrimSuffix(filepath.Base(key), ext),
				Filename:   filepath.Base(key),
				Type:       pluginType,
				Source:     ResolumePluginSourceS3,
				S3Key:      key,
				Size:       aws.ToInt64(obj.Size),
				ModifiedAt: aws.ToTime(obj.LastModified),
			}

			// Try to load metadata
			metaKey := key + ".json"
			if meta, err := c.loadS3Metadata(ctx, metaKey); err == nil {
				plugin.Version = meta.Version
				plugin.Tags = meta.Tags
				plugin.SHA256 = meta.SHA256
				plugin.UploadedBy = meta.UploadedBy
				plugin.UploadedAt = meta.UploadedAt
				plugin.ResolumeVersion = meta.ResolumeVersion
				plugin.IsJuiceBar = meta.IsJuiceBar
			}

			plugins = append(plugins, plugin)
		}
	}

	return plugins, nil
}

func (c *ResolumePluginsClient) loadS3Metadata(ctx context.Context, key string) (*ResolumePluginInfo, error) {
	output, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer output.Body.Close()

	var meta ResolumePluginInfo
	if err := json.NewDecoder(output.Body).Decode(&meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

// UploadPlugin uploads a plugin to S3
func (c *ResolumePluginsClient) UploadPlugin(ctx context.Context, localPath string, opts UploadOptions) (*ResolumePluginInfo, error) {
	// Validate file exists
	info, err := os.Stat(localPath)
	if err != nil {
		return nil, fmt.Errorf("plugin not found: %w", err)
	}

	// Determine plugin type
	ext := strings.ToLower(filepath.Ext(localPath))
	pluginType, isPlugin := c.getPluginType(ext)
	if !isPlugin {
		return nil, fmt.Errorf("not a valid plugin file: %s", ext)
	}

	// Calculate SHA256
	hash, err := c.calculateSHA256(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate hash: %w", err)
	}

	// Determine S3 key
	s3Key := c.buildS3Key(pluginType, filepath.Base(localPath), opts)

	// Open file
	f, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open plugin: %w", err)
	}
	defer f.Close()

	// Upload plugin
	_, err = c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(s3Key),
		Body:   f,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload plugin: %w", err)
	}

	// Create metadata
	now := time.Now()
	uploadedBy := os.Getenv("USER")
	if uploadedBy == "" {
		uploadedBy = "unknown"
	}

	plugin := &ResolumePluginInfo{
		Name:            strings.TrimSuffix(filepath.Base(localPath), ext),
		Filename:        filepath.Base(localPath),
		Type:            pluginType,
		Source:          opts.Source,
		Version:         opts.Version,
		Platform:        c.platform,
		S3Key:           s3Key,
		SHA256:          hash,
		Size:            info.Size(),
		UploadedAt:      &now,
		UploadedBy:      uploadedBy,
		Tags:            opts.Tags,
		ResolumeVersion: opts.ResolumeVersion,
		IsJuiceBar:      opts.IsJuiceBar,
		ModifiedAt:      info.ModTime(),
	}

	// Upload metadata
	metaKey := s3Key + ".json"
	metaData, _ := json.MarshalIndent(plugin, "", "  ")

	_, err = c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucketName),
		Key:         aws.String(metaKey),
		Body:        strings.NewReader(string(metaData)),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload metadata: %w", err)
	}

	return plugin, nil
}

// UploadOptions configures plugin upload
type UploadOptions struct {
	Source          ResolumePluginSource
	Version         string
	Tags            []string
	ResolumeVersion string
	IsJuiceBar      bool
	Subfolder       string
}

func (c *ResolumePluginsClient) buildS3Key(pluginType ResolumePluginType, filename string, opts UploadOptions) string {
	var prefix string

	switch pluginType {
	case ResolumePluginTypeFFGLEffect:
		prefix = "ffgl/effects"
	case ResolumePluginTypeFFGLSource:
		prefix = "ffgl/sources"
	case ResolumePluginTypeISF:
		prefix = "isf"
	default:
		prefix = "other"
	}

	if opts.IsJuiceBar {
		prefix = "juicebar-backup"
	}

	if opts.Subfolder != "" {
		prefix = filepath.Join(prefix, opts.Subfolder)
	}

	return filepath.Join(prefix, filename)
}

// DownloadPlugin downloads a plugin from S3
func (c *ResolumePluginsClient) DownloadPlugin(ctx context.Context, s3Key string) (string, error) {
	// Download to Extra Effects folder
	paths := c.GetPluginPaths()
	destDir := paths["extra_effects"]

	// Ensure directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Sanitize filename to prevent path traversal
	filename := filepath.Base(s3Key)
	filename = strings.ReplaceAll(filename, "..", "")
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")
	if filename == "" || filename == "." {
		return "", fmt.Errorf("invalid plugin filename")
	}

	destPath := filepath.Join(destDir, filename)

	// Verify path is within destDir (defense in depth)
	absDestDir, _ := filepath.Abs(destDir)
	absDestPath, _ := filepath.Abs(destPath)
	if !strings.HasPrefix(absDestPath, absDestDir) {
		return "", fmt.Errorf("invalid destination path")
	}

	// Try to load metadata for hash verification
	var expectedHash string
	metaKey := s3Key + ".json"
	if meta, err := c.loadS3Metadata(ctx, metaKey); err == nil && meta.SHA256 != "" {
		expectedHash = meta.SHA256
	}

	// Download from S3
	output, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return "", fmt.Errorf("failed to download from S3: %w", err)
	}
	defer output.Body.Close()

	// Write to temp file first for atomic operation
	tmpPath := destPath + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}

	// Hash while writing for verification
	h := sha256.New()
	writer := io.MultiWriter(f, h)

	if _, err := io.Copy(writer, output.Body); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	f.Close()

	// Verify hash if available
	actualHash := hex.EncodeToString(h.Sum(nil))
	if expectedHash != "" && actualHash != expectedHash {
		os.Remove(tmpPath)
		return "", fmt.Errorf("hash mismatch: expected %s, got %s", expectedHash[:16]+"...", actualHash[:16]+"...")
	}

	// Atomic rename
	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to finalize file: %w", err)
	}

	// Set executable permission for bundles
	if strings.HasSuffix(destPath, ".bundle") {
		if err := os.Chmod(destPath, 0755); err != nil {
			// Log but don't fail - plugin is installed
			fmt.Fprintf(os.Stderr, "warning: failed to set executable permission: %v\n", err)
		}
	}

	return destPath, nil
}

// BackupResult contains the results of a backup operation
type BackupResult struct {
	Uploaded []ResolumePluginInfo `json:"uploaded"`
	Errors   []error              `json:"errors,omitempty"`
}

// BackupAllPlugins uploads all local plugins to S3
func (c *ResolumePluginsClient) BackupAllPlugins(ctx context.Context, includeJuiceBar bool) (*BackupResult, error) {
	localPlugins, err := c.ScanLocalPlugins(ctx)
	if err != nil {
		return nil, err
	}

	result := &BackupResult{
		Uploaded: []ResolumePluginInfo{},
		Errors:   []error{},
	}

	for _, plugin := range localPlugins {
		// Skip JuiceBar if not requested
		if plugin.IsJuiceBar && !includeJuiceBar {
			continue
		}

		// Skip built-in plugins
		if strings.Contains(plugin.LocalPath, "/Applications/") ||
			strings.Contains(plugin.LocalPath, "Program Files") {
			continue
		}

		opts := UploadOptions{
			Source:     plugin.Source,
			IsJuiceBar: plugin.IsJuiceBar,
		}

		uploaded, err := c.UploadPlugin(ctx, plugin.LocalPath, opts)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("%s: %w", plugin.Name, err))
			continue
		}

		result.Uploaded = append(result.Uploaded, *uploaded)
	}

	return result, nil
}

// SyncPlugins synchronizes local and S3 plugins
func (c *ResolumePluginsClient) SyncPlugins(ctx context.Context, direction string) (*SyncResult, error) {
	result := &SyncResult{}

	localPlugins, err := c.ScanLocalPlugins(ctx)
	if err != nil {
		return nil, err
	}

	s3Plugins, err := c.ListS3Plugins(ctx, "")
	if err != nil {
		return nil, err
	}

	// Build maps for comparison
	localMap := make(map[string]ResolumePluginInfo)
	for _, p := range localPlugins {
		localMap[p.Filename] = p
	}

	s3Map := make(map[string]ResolumePluginInfo)
	for _, p := range s3Plugins {
		s3Map[p.Filename] = p
	}

	// Find plugins to upload (local only)
	for filename, localPlugin := range localMap {
		if _, exists := s3Map[filename]; !exists {
			result.ToUpload = append(result.ToUpload, localPlugin)
		}
	}

	// Find plugins to download (S3 only)
	for filename, s3Plugin := range s3Map {
		if _, exists := localMap[filename]; !exists {
			result.ToDownload = append(result.ToDownload, s3Plugin)
		}
	}

	// Find updated plugins (different hash)
	for filename, localPlugin := range localMap {
		if s3Plugin, exists := s3Map[filename]; exists {
			if localPlugin.SHA256 != "" && s3Plugin.SHA256 != "" &&
				localPlugin.SHA256 != s3Plugin.SHA256 {
				if localPlugin.ModifiedAt.After(s3Plugin.ModifiedAt) {
					result.LocalNewer = append(result.LocalNewer, localPlugin)
				} else {
					result.S3Newer = append(result.S3Newer, s3Plugin)
				}
			}
		}
	}

	return result, nil
}

// SyncResult contains sync comparison results
type SyncResult struct {
	ToUpload   []ResolumePluginInfo `json:"to_upload"`
	ToDownload []ResolumePluginInfo `json:"to_download"`
	LocalNewer []ResolumePluginInfo `json:"local_newer"`
	S3Newer    []ResolumePluginInfo `json:"s3_newer"`
}

// InstallPlugin installs a plugin to the Extra Effects folder
func (c *ResolumePluginsClient) InstallPlugin(ctx context.Context, sourcePath string) (string, error) {
	paths := c.GetPluginPaths()
	destDir := paths["extra_effects"]

	// Ensure directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	destPath := filepath.Join(destDir, filepath.Base(sourcePath))

	// Copy file
	src, err := os.Open(sourcePath)
	if err != nil {
		return "", err
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	// Set executable permission for bundles
	if strings.HasSuffix(destPath, ".bundle") {
		os.Chmod(destPath, 0755)
	}

	return destPath, nil
}

// UninstallPlugin removes a plugin from Extra Effects
func (c *ResolumePluginsClient) UninstallPlugin(ctx context.Context, pluginName string) error {
	paths := c.GetPluginPaths()
	extraEffects := paths["extra_effects"]

	// Find the plugin
	var targetPath string
	filepath.Walk(extraEffects, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.Contains(strings.ToLower(info.Name()), strings.ToLower(pluginName)) {
			targetPath = path
			return filepath.SkipDir
		}
		return nil
	})

	if targetPath == "" {
		return fmt.Errorf("plugin not found: %s", pluginName)
	}

	return os.Remove(targetPath)
}

// GetHealth returns plugin system health
func (c *ResolumePluginsClient) GetHealth(ctx context.Context) (*ResolumePluginHealth, error) {
	health := &ResolumePluginHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check local plugins
	localPlugins, err := c.ScanLocalPlugins(ctx)
	if err != nil {
		health.Score -= 20
		health.Issues = append(health.Issues, "Failed to scan local plugins")
	} else {
		health.LocalCount = len(localPlugins)
	}

	// Check S3 connection
	s3Plugins, err := c.ListS3Plugins(ctx, "")
	if err != nil {
		health.Score -= 30
		health.Issues = append(health.Issues, "Failed to connect to S3 bucket")
	} else {
		health.S3Count = len(s3Plugins)
	}

	// Check directories
	paths := c.GetPluginPaths()
	for name, path := range paths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if name == "extra_effects" {
				health.Score -= 10
				health.Recommendations = append(health.Recommendations,
					fmt.Sprintf("Create missing directory: %s", path))
			}
		}
	}

	// Set status based on score
	switch {
	case health.Score >= 80:
		health.Status = "healthy"
	case health.Score >= 50:
		health.Status = "degraded"
	default:
		health.Status = "critical"
	}

	return health, nil
}

// ResolumePluginHealth represents plugin system health
type ResolumePluginHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	LocalCount      int      `json:"local_count"`
	S3Count         int      `json:"s3_count"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// SearchS3Plugins searches for plugins in S3 by name, tags, or type
func (c *ResolumePluginsClient) SearchS3Plugins(ctx context.Context, query string, pluginType string, tags []string) ([]ResolumePluginInfo, error) {
	// Get all plugins first
	allPlugins, err := c.ListS3Plugins(ctx, "")
	if err != nil {
		return nil, err
	}

	var results []ResolumePluginInfo
	queryLower := strings.ToLower(query)

	for _, plugin := range allPlugins {
		// Match by name
		if query != "" && !strings.Contains(strings.ToLower(plugin.Name), queryLower) {
			continue
		}

		// Match by type
		if pluginType != "" && string(plugin.Type) != pluginType {
			continue
		}

		// Match by tags
		if len(tags) > 0 {
			hasTag := false
			for _, searchTag := range tags {
				for _, pluginTag := range plugin.Tags {
					if strings.EqualFold(searchTag, pluginTag) {
						hasTag = true
						break
					}
				}
				if hasTag {
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		results = append(results, plugin)
	}

	return results, nil
}

// BatchUpload uploads multiple plugins to S3
func (c *ResolumePluginsClient) BatchUpload(ctx context.Context, paths []string, opts UploadOptions) ([]ResolumePluginInfo, []error) {
	var uploaded []ResolumePluginInfo
	var errors []error

	for _, path := range paths {
		result, err := c.UploadPlugin(ctx, path, opts)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", filepath.Base(path), err))
			continue
		}
		uploaded = append(uploaded, *result)
	}

	return uploaded, errors
}

// BatchDownload downloads multiple plugins from S3
func (c *ResolumePluginsClient) BatchDownload(ctx context.Context, keys []string) ([]string, []error) {
	var downloaded []string
	var errors []error

	for _, key := range keys {
		destPath, err := c.DownloadPlugin(ctx, key)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", filepath.Base(key), err))
			continue
		}
		downloaded = append(downloaded, destPath)
	}

	return downloaded, errors
}

// ImportFromGDrive downloads a plugin from Google Drive and optionally uploads to S3
func (c *ResolumePluginsClient) ImportFromGDrive(ctx context.Context, gdriveClient *GDriveClient, fileID string, uploadToS3 bool) (*ResolumePluginInfo, error) {
	// Get file info
	fileInfo, err := gdriveClient.GetFileInfo(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("get file info: %w", err)
	}

	// Validate it's a plugin file
	ext := strings.ToLower(filepath.Ext(fileInfo.Name))
	if _, isPlugin := c.getPluginType(ext); !isPlugin {
		return nil, fmt.Errorf("not a valid plugin file: %s", fileInfo.Name)
	}

	// Download to temp location
	tempDir := os.TempDir()
	tempPath := filepath.Join(tempDir, fileInfo.Name)

	progress, err := gdriveClient.DownloadFile(ctx, fileID, tempPath)
	if err != nil {
		return nil, fmt.Errorf("download from Drive: %w", err)
	}

	if progress.Status != "completed" {
		return nil, fmt.Errorf("download incomplete: %s", progress.Status)
	}

	// Install locally
	destPath, err := c.InstallPlugin(ctx, tempPath)
	if err != nil {
		os.Remove(tempPath)
		return nil, fmt.Errorf("install plugin: %w", err)
	}

	// Clean up temp file
	os.Remove(tempPath)

	// Scan the installed plugin for metadata
	plugins, _ := c.ScanLocalPlugins(ctx)
	var installed *ResolumePluginInfo
	for _, p := range plugins {
		if p.LocalPath == destPath {
			installed = &p
			break
		}
	}

	if installed == nil {
		installed = &ResolumePluginInfo{
			Name:      strings.TrimSuffix(fileInfo.Name, ext),
			Filename:  fileInfo.Name,
			LocalPath: destPath,
			Size:      fileInfo.Size,
		}
	}

	// Optionally upload to S3
	if uploadToS3 {
		opts := UploadOptions{
			Source: ResolumePluginSourceLocal,
		}
		result, err := c.UploadPlugin(ctx, destPath, opts)
		if err != nil {
			return installed, fmt.Errorf("plugin installed but S3 upload failed: %w", err)
		}
		installed.S3Key = result.S3Key
	}

	return installed, nil
}

// ImportFolderFromGDrive downloads all plugins from a Google Drive folder
func (c *ResolumePluginsClient) ImportFolderFromGDrive(ctx context.Context, gdriveClient *GDriveClient, folderID string, uploadToS3 bool) ([]ResolumePluginInfo, []error) {
	var imported []ResolumePluginInfo
	var errors []error

	// List files in folder
	files, _, err := gdriveClient.ListFiles(ctx, folderID, 1000, "")
	if err != nil {
		return nil, []error{fmt.Errorf("list folder: %w", err)}
	}

	for _, file := range files {
		// Skip folders
		if file.IsFolder {
			continue
		}

		// Check if it's a plugin file
		ext := strings.ToLower(filepath.Ext(file.Name))
		if _, isPlugin := c.getPluginType(ext); !isPlugin {
			continue
		}

		plugin, err := c.ImportFromGDrive(ctx, gdriveClient, file.ID, uploadToS3)
		if err != nil {
			errors = append(errors, fmt.Errorf("%s: %w", file.Name, err))
			continue
		}
		imported = append(imported, *plugin)
	}

	return imported, errors
}

// ISFShader represents an ISF shader from ISF.video
type ISFShader struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Author      string   `json:"author"`
	Description string   `json:"description"`
	Categories  []string `json:"categories"`
	DownloadURL string   `json:"download_url"`
	PreviewURL  string   `json:"preview_url"`
	Inputs      int      `json:"inputs"`
}

// BrowseISFShaders fetches ISF shaders from ISF.video community
func (c *ResolumePluginsClient) BrowseISFShaders(ctx context.Context, category string, limit int) ([]ISFShader, error) {
	// ISF.video doesn't have a public API, so we provide curated shader sources
	// This returns known good shader collections from GitHub
	shaders := []ISFShader{
		{ID: "isf-vidvox-1", Title: "Vidvox ISF Files", Author: "Vidvox", Description: "Official ISF shader collection from Vidvox", Categories: []string{"effects", "generators"}, DownloadURL: "https://github.com/Vidvox/ISF-Files/archive/refs/heads/master.zip"},
		{ID: "isf-mrfunk", Title: "Mr. Funk Collection", Author: "MrFunk", Description: "Creative visual effects collection", Categories: []string{"effects", "glitch"}, DownloadURL: "https://github.com/mrfunk/isf-shaders/archive/refs/heads/master.zip"},
		{ID: "isf-patriciogv", Title: "Patricio Gonzalez Vivo", Author: "patriciogv", Description: "Book of Shaders examples in ISF format", Categories: []string{"generators", "patterns"}, DownloadURL: "https://github.com/patriciogonzalezvivo/ISF-shaders/archive/refs/heads/master.zip"},
	}

	// Filter by category if specified
	if category != "" {
		var filtered []ISFShader
		for _, s := range shaders {
			for _, cat := range s.Categories {
				if strings.EqualFold(cat, category) {
					filtered = append(filtered, s)
					break
				}
			}
		}
		shaders = filtered
	}

	if limit > 0 && len(shaders) > limit {
		shaders = shaders[:limit]
	}

	return shaders, nil
}

// IsResolumeRunning checks if Resolume is currently running
func (c *ResolumePluginsClient) IsResolumeRunning(ctx context.Context) (bool, string, error) {
	// Try to create a Resolume client and check status
	resClient, err := NewResolumeClient()
	if err != nil {
		return false, "", nil
	}

	status, err := resClient.GetStatus(ctx)
	if err != nil {
		return false, "", nil
	}

	return status.Connected, status.Version, nil
}

// TriggerResolumeRescan sends a message to Resolume to rescan plugins
// Note: Resolume doesn't have a direct API for this, so we use AppleScript on macOS
func (c *ResolumePluginsClient) TriggerResolumeRescan(ctx context.Context) error {
	if c.platform != "macos" {
		return fmt.Errorf("plugin rescan is only supported on macOS")
	}

	// Check if Resolume is running
	running, _, _ := c.IsResolumeRunning(ctx)
	if !running {
		return fmt.Errorf("Resolume is not running")
	}

	// On macOS, the best we can do is suggest the user manually rescan
	// Resolume Arena 7+ doesn't have a documented OSC/API command for plugin rescan
	return fmt.Errorf("automatic rescan not available - please use Resolume menu: Arena > Preferences > Video > Rescan Plugins")
}

// GetResolumeInfo returns information about the connected Resolume instance
func (c *ResolumePluginsClient) GetResolumeInfo(ctx context.Context) (map[string]interface{}, error) {
	info := map[string]interface{}{
		"platform": c.platform,
		"paths":    c.GetPluginPaths(),
	}

	running, version, err := c.IsResolumeRunning(ctx)
	if err != nil {
		info["error"] = err.Error()
	}
	info["running"] = running
	info["version"] = version

	return info, nil
}
