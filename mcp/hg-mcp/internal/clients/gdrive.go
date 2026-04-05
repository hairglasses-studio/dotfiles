// Package clients provides client implementations for external services.
package clients

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/hairglasses-studio/hg-mcp/internal/config"
)

// GDriveClient provides Google Drive operations
type GDriveClient struct {
	service *drive.Service
	mu      sync.RWMutex
}

// GDriveFile represents a file or folder in Google Drive
type GDriveFile struct {
	ID           string            `json:"id"`
	Name         string            `json:"name"`
	MimeType     string            `json:"mime_type"`
	Size         int64             `json:"size"`
	CreatedTime  time.Time         `json:"created_time"`
	ModifiedTime time.Time         `json:"modified_time"`
	Parents      []string          `json:"parents,omitempty"`
	WebViewLink  string            `json:"web_view_link,omitempty"`
	IsFolder     bool              `json:"is_folder"`
	Path         string            `json:"path,omitempty"`
	Properties   map[string]string `json:"properties,omitempty"`
}

// GDriveQuota represents storage quota information
type GDriveQuota struct {
	Limit        int64   `json:"limit"`
	Usage        int64   `json:"usage"`
	UsageInDrive int64   `json:"usage_in_drive"`
	UsedPercent  float64 `json:"used_percent"`
	AvailableGB  float64 `json:"available_gb"`
}

// GDriveDownloadProgress tracks download progress
type GDriveDownloadProgress struct {
	FileName        string  `json:"file_name"`
	TotalBytes      int64   `json:"total_bytes"`
	DownloadedBytes int64   `json:"downloaded_bytes"`
	Percent         float64 `json:"percent"`
	Status          string  `json:"status"`
}

// Video file extensions for filtering
var videoExtensions = map[string]bool{
	".mp4": true, ".mov": true, ".avi": true, ".webm": true,
	".mkv": true, ".m4v": true, ".wmv": true, ".flv": true,
	".prores": true, ".mxf": true, ".ts": true, ".m2ts": true,
}

// Image file extensions for filtering
var imageExtensions = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true,
	".tiff": true, ".tif": true, ".bmp": true, ".webp": true,
	".psd": true, ".exr": true, ".hdr": true, ".tga": true,
}

var (
	gdriveClient     *GDriveClient
	gdriveClientOnce sync.Once
	gdriveClientErr  error
)

// GetGDriveClient returns the singleton Google Drive client
func GetGDriveClient() (*GDriveClient, error) {
	gdriveClientOnce.Do(func() {
		gdriveClient, gdriveClientErr = NewGDriveClient()
	})
	return gdriveClient, gdriveClientErr
}

// NewGDriveClient creates a new Google Drive client
func NewGDriveClient() (*GDriveClient, error) {
	ctx := context.Background()

	// Auth priority:
	// 1. GOOGLE_APPLICATION_CREDENTIALS (service account file)
	// 2. GOOGLE_API_KEY (API key — read-only)
	// 3. ADC fallback (gcloud auth application-default login)
	var opts []option.ClientOption

	cfg := config.Get()
	if cfg.GoogleApplicationCredentials != "" {
		opts = append(opts, option.WithCredentialsFile(cfg.GoogleApplicationCredentials))
	} else if cfg.GoogleAPIKey != "" {
		opts = append(opts, option.WithAPIKey(cfg.GoogleAPIKey))
	} else {
		// Fall back to ADC (same as Sheets client)
		opts = append(opts, option.WithScopes(drive.DriveReadonlyScope))
	}

	service, err := drive.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Drive service: %w", err)
	}

	return &GDriveClient{service: service}, nil
}

// IsConfigured returns true if the client is properly configured
func (c *GDriveClient) IsConfigured() bool {
	return c != nil && c.service != nil
}

// ListFiles lists files in a folder
func (c *GDriveClient) ListFiles(ctx context.Context, folderID string, pageSize int64, pageToken string) ([]GDriveFile, string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if folderID == "" {
		folderID = "root"
	}
	if pageSize <= 0 {
		pageSize = 100
	}

	query := fmt.Sprintf("'%s' in parents and trashed = false", folderID)

	call := c.service.Files.List().
		Context(ctx).
		Q(query).
		PageSize(pageSize).
		Fields("nextPageToken, files(id, name, mimeType, size, createdTime, modifiedTime, parents, webViewLink, properties)").
		OrderBy("name")

	if pageToken != "" {
		call = call.PageToken(pageToken)
	}

	result, err := call.Do()
	if err != nil {
		return nil, "", fmt.Errorf("failed to list files: %w", err)
	}

	files := make([]GDriveFile, 0, len(result.Files))
	for _, f := range result.Files {
		files = append(files, convertDriveFile(f))
	}

	return files, result.NextPageToken, nil
}

// SearchFiles searches for files matching a query
func (c *GDriveClient) SearchFiles(ctx context.Context, query string, fileType string, folderID string, limit int64) ([]GDriveFile, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	// Build query
	var queryParts []string
	queryParts = append(queryParts, "trashed = false")

	if query != "" {
		queryParts = append(queryParts, fmt.Sprintf("name contains '%s'", query))
	}

	if folderID != "" {
		queryParts = append(queryParts, fmt.Sprintf("'%s' in parents", folderID))
	}

	// Filter by file type
	switch strings.ToLower(fileType) {
	case "video":
		queryParts = append(queryParts, "(mimeType contains 'video/' or mimeType = 'application/mxf')")
	case "image":
		queryParts = append(queryParts, "mimeType contains 'image/'")
	case "folder":
		queryParts = append(queryParts, "mimeType = 'application/vnd.google-apps.folder'")
	case "audio":
		queryParts = append(queryParts, "mimeType contains 'audio/'")
	}

	fullQuery := strings.Join(queryParts, " and ")

	call := c.service.Files.List().
		Context(ctx).
		Q(fullQuery).
		PageSize(limit).
		Fields("files(id, name, mimeType, size, createdTime, modifiedTime, parents, webViewLink, properties)").
		OrderBy("modifiedTime desc")

	result, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to search files: %w", err)
	}

	files := make([]GDriveFile, 0, len(result.Files))
	for _, f := range result.Files {
		files = append(files, convertDriveFile(f))
	}

	return files, nil
}

// GetFileInfo gets metadata for a specific file
func (c *GDriveClient) GetFileInfo(ctx context.Context, fileID string) (*GDriveFile, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	file, err := c.service.Files.Get(fileID).
		Context(ctx).
		Fields("id, name, mimeType, size, createdTime, modifiedTime, parents, webViewLink, properties").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	result := convertDriveFile(file)
	return &result, nil
}

// DownloadFile downloads a file to the specified path
func (c *GDriveClient) DownloadFile(ctx context.Context, fileID string, destPath string) (*GDriveDownloadProgress, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Get file info first
	file, err := c.service.Files.Get(fileID).
		Context(ctx).
		Fields("id, name, mimeType, size").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Determine destination
	if destPath == "" {
		destPath = "."
	}

	// If destPath is a directory, append filename
	stat, err := os.Stat(destPath)
	if err == nil && stat.IsDir() {
		destPath = filepath.Join(destPath, file.Name)
	} else if os.IsNotExist(err) {
		// Create parent directory if needed
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	// Download the file
	resp, err := c.service.Files.Get(fileID).Context(ctx).Download()
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	// Create destination file
	out, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()

	// Copy with progress tracking
	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	progress := &GDriveDownloadProgress{
		FileName:        file.Name,
		TotalBytes:      file.Size,
		DownloadedBytes: written,
		Percent:         100.0,
		Status:          "completed",
	}

	return progress, nil
}

// DownloadFolder downloads all files in a folder
func (c *GDriveClient) DownloadFolder(ctx context.Context, folderID string, destPath string, fileType string, recursive bool) ([]GDriveDownloadProgress, error) {
	// Get folder info for name
	folderInfo, err := c.GetFileInfo(ctx, folderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get folder info: %w", err)
	}

	// Create destination directory
	folderPath := filepath.Join(destPath, folderInfo.Name)
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create folder: %w", err)
	}

	// List files
	files, _, err := c.ListFiles(ctx, folderID, 1000, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	var results []GDriveDownloadProgress

	for _, file := range files {
		// Handle folders recursively
		if file.IsFolder {
			if recursive {
				subResults, err := c.DownloadFolder(ctx, file.ID, folderPath, fileType, recursive)
				if err != nil {
					results = append(results, GDriveDownloadProgress{
						FileName: file.Name,
						Status:   fmt.Sprintf("error: %v", err),
					})
				} else {
					results = append(results, subResults...)
				}
			}
			continue
		}

		// Filter by file type if specified
		if fileType != "" {
			ext := strings.ToLower(filepath.Ext(file.Name))
			switch strings.ToLower(fileType) {
			case "video":
				if !videoExtensions[ext] && !strings.Contains(file.MimeType, "video/") {
					continue
				}
			case "image":
				if !imageExtensions[ext] && !strings.Contains(file.MimeType, "image/") {
					continue
				}
			}
		}

		// Download the file
		progress, err := c.DownloadFile(ctx, file.ID, folderPath)
		if err != nil {
			results = append(results, GDriveDownloadProgress{
				FileName: file.Name,
				Status:   fmt.Sprintf("error: %v", err),
			})
		} else {
			results = append(results, *progress)
		}
	}

	return results, nil
}

// GetQuota returns storage quota information
func (c *GDriveClient) GetQuota(ctx context.Context) (*GDriveQuota, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	about, err := c.service.About.Get().
		Context(ctx).
		Fields("storageQuota").
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get quota: %w", err)
	}

	q := about.StorageQuota
	usedPercent := 0.0
	availableGB := 0.0
	if q.Limit > 0 {
		usedPercent = float64(q.Usage) / float64(q.Limit) * 100
		availableGB = float64(q.Limit-q.Usage) / (1024 * 1024 * 1024)
	}

	return &GDriveQuota{
		Limit:        q.Limit,
		Usage:        q.Usage,
		UsageInDrive: q.UsageInDrive,
		UsedPercent:  usedPercent,
		AvailableGB:  availableGB,
	}, nil
}

// ListSharedDrives lists available shared drives
func (c *GDriveClient) ListSharedDrives(ctx context.Context) ([]GDriveFile, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result, err := c.service.Drives.List().
		Context(ctx).
		PageSize(100).
		Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list shared drives: %w", err)
	}

	drives := make([]GDriveFile, 0, len(result.Drives))
	for _, d := range result.Drives {
		drives = append(drives, GDriveFile{
			ID:       d.Id,
			Name:     d.Name,
			MimeType: "application/vnd.google-apps.folder",
			IsFolder: true,
		})
	}

	return drives, nil
}

// GetFolderPath builds the full path to a folder
func (c *GDriveClient) GetFolderPath(ctx context.Context, folderID string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var path []string
	currentID := folderID

	for currentID != "" && currentID != "root" {
		file, err := c.service.Files.Get(currentID).
			Context(ctx).
			Fields("id, name, parents").
			Do()
		if err != nil {
			return "", fmt.Errorf("failed to get parent: %w", err)
		}

		path = append([]string{file.Name}, path...)

		if len(file.Parents) > 0 {
			currentID = file.Parents[0]
		} else {
			break
		}
	}

	return "/" + strings.Join(path, "/"), nil
}

// convertDriveFile converts a Google Drive file to our format
func convertDriveFile(f *drive.File) GDriveFile {
	file := GDriveFile{
		ID:          f.Id,
		Name:        f.Name,
		MimeType:    f.MimeType,
		Size:        f.Size,
		Parents:     f.Parents,
		WebViewLink: f.WebViewLink,
		IsFolder:    f.MimeType == "application/vnd.google-apps.folder",
		Properties:  f.Properties,
	}

	if f.CreatedTime != "" {
		if t, err := time.Parse(time.RFC3339, f.CreatedTime); err == nil {
			file.CreatedTime = t
		}
	}
	if f.ModifiedTime != "" {
		if t, err := time.Parse(time.RFC3339, f.ModifiedTime); err == nil {
			file.ModifiedTime = t
		}
	}

	return file
}

// FormatFileSize formats bytes into human-readable size
func FormatFileSize(bytes int64) string {
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

// VJZipPack represents a downloadable VJ clip zip archive
type VJZipPack struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	SizeGB      float64  `json:"size_gb"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Parts       []string `json:"parts,omitempty"` // For multi-part archives
}

// VJZipPackCatalog returns the built-in catalog of VJ zip packs
func VJZipPackCatalog() []VJZipPack {
	return []VJZipPack{
		{
			Key:         "supreme_cyphers_v3",
			Name:        "SUPREME CYPHERS vol.3",
			Path:        "DXV3 1080p HQ Clips/Fetz VJ Storage/3D Rendered Beeple & SWIII LOOPS/COMPLETE PACK (zipped folder)/SUPREME CYPHERS vol.3 [complete pack].zip",
			SizeGB:      2.00,
			Description: "Supreme Cyphers visual pack volume 3",
			Tags:        []string{"3d", "beeple", "loops"},
		},
		{
			Key:         "supreme_cyphers_v2",
			Name:        "Supreme Cyphers vol.2",
			Path:        "DXV3 1080p HQ Clips/Fetz VJ Storage/3D Rendered Beeple & SWIII LOOPS/COMPLETE PACK ZIPPED FOLDER/_Supreme Cyphers (vol.2) complete pack.zip",
			SizeGB:      1.43,
			Description: "Supreme Cyphers visual pack volume 2",
			Tags:        []string{"3d", "beeple", "loops"},
		},
		{
			Key:         "nikita_vision_pack",
			Name:        "Nikita Vision Pack",
			Path:        "DXV3 1080p HQ Clips/Fetz VJ Storage/Homie VJ Packs/DXV3 Nikita Vision Pack (Get Permission Before Use)",
			SizeGB:      6.47,
			Description: "Nikita Vision VJ loop collection (4 parts, requires permission)",
			Tags:        []string{"loops", "dxv3", "nikita"},
			Parts: []string{
				"NIKITAVISIION VJ LOOP DUMP-20231209T073242Z-001.zip",
				"NIKITAVISIION VJ LOOP DUMP-20231209T073242Z-002.zip",
				"NIKITAVISIION VJ LOOP DUMP-20231209T073242Z-003.zip",
				"NIKITAVISIION VJ LOOP DUMP-20231209T073242Z-004.zip",
			},
		},
		{
			Key:         "beeple_brainfeeder",
			Name:        "Beeple Brainfeeder VJ Clips",
			Path:        "DXV3 1080p HQ Clips/Fetz VJ Storage/3D Rendered Beeple & SWIII LOOPS/brainfeeder_beeple-vjclips/vj clips.zip",
			SizeGB:      1.25,
			Description: "Beeple VJ clips from Brainfeeder collection",
			Tags:        []string{"beeple", "brainfeeder", "loops"},
		},
		{
			Key:         "ubersketch",
			Name:        "Ubersketch Project Files",
			Path:        "DXV3 1080p HQ Clips/Fetz VJ Storage/3D Rendered Beeple & SWIII LOOPS/Ubersketch/ubersketch project files.zip",
			SizeGB:      0.03,
			Description: "Ubersketch project files",
			Tags:        []string{"project", "ubersketch"},
		},
	}
}

// GetVJZipPack returns a specific zip pack by key
func GetVJZipPack(key string) *VJZipPack {
	for _, pack := range VJZipPackCatalog() {
		if pack.Key == key {
			return &pack
		}
	}
	return nil
}

// SearchVJZipPacks searches zip packs by name or tags
func SearchVJZipPacks(query string) []VJZipPack {
	query = strings.ToLower(query)
	var results []VJZipPack
	for _, pack := range VJZipPackCatalog() {
		// Search in name
		if strings.Contains(strings.ToLower(pack.Name), query) {
			results = append(results, pack)
			continue
		}
		// Search in description
		if strings.Contains(strings.ToLower(pack.Description), query) {
			results = append(results, pack)
			continue
		}
		// Search in tags
		for _, tag := range pack.Tags {
			if strings.Contains(strings.ToLower(tag), query) {
				results = append(results, pack)
				break
			}
		}
	}
	return results
}
