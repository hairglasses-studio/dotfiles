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
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// VJClipInfo represents metadata about a VJ clip
type VJClipInfo struct {
	Name       string     `json:"name"`
	Filename   string     `json:"filename"`
	User       string     `json:"user"`     // Top-level user/owner
	Playlist   string     `json:"playlist"` // Playlist/pack name
	Format     string     `json:"format"`   // mov, mp4, avi, etc.
	Codec      string     `json:"codec,omitempty"`
	Resolution string     `json:"resolution,omitempty"`
	Duration   float64    `json:"duration,omitempty"`
	Framerate  float64    `json:"framerate,omitempty"`
	LocalPath  string     `json:"local_path,omitempty"`
	S3Key      string     `json:"s3_key,omitempty"`
	SHA256     string     `json:"sha256,omitempty"`
	Size       int64      `json:"size"`
	UploadedAt *time.Time `json:"uploaded_at,omitempty"`
	UploadedBy string     `json:"uploaded_by,omitempty"`
	Tags       []string   `json:"tags,omitempty"`
	ModifiedAt time.Time  `json:"modified_at"`
	// DXV codec info
	IsDXV bool `json:"is_dxv,omitempty"`
}

// VJPlaylist represents a collection of clips
type VJPlaylist struct {
	Name       string       `json:"name"`
	User       string       `json:"user"`
	ClipCount  int          `json:"clip_count"`
	TotalSize  int64        `json:"total_size"`
	Clips      []VJClipInfo `json:"clips,omitempty"`
	ModifiedAt time.Time    `json:"modified_at"`
}

// VJClipsCatalog represents the full S3 clip catalog
type VJClipsCatalog struct {
	Version     int          `json:"version"`
	LastUpdated time.Time    `json:"last_updated"`
	Users       []string     `json:"users"`
	Playlists   []VJPlaylist `json:"playlists"`
	TotalClips  int          `json:"total_clips"`
	TotalSize   int64        `json:"total_size"`
}

// VJClipsClient provides VJ clip management
type VJClipsClient struct {
	s3Client    *s3.Client
	bucketName  string
	homeDir     string
	mediaPath   string
	defaultUser string
}

// NewVJClipsClient creates a new VJ clips client
func NewVJClipsClient() (*VJClipsClient, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	bucketName := os.Getenv("VJ_CLIPS_BUCKET")
	if bucketName == "" {
		bucketName = "aftrs-vj-clips"
	}

	mediaPath := os.Getenv("VJ_CLIPS_PATH")
	if mediaPath == "" {
		mediaPath = filepath.Join(homeDir, "Documents", "Resolume Arena", "Media")
	}

	defaultUser := os.Getenv("USER")
	if defaultUser == "" {
		defaultUser = "default"
	}

	return &VJClipsClient{
		s3Client:    s3.NewFromConfig(cfg),
		bucketName:  bucketName,
		homeDir:     homeDir,
		mediaPath:   mediaPath,
		defaultUser: defaultUser,
	}, nil
}

// GetMediaPath returns the local media path
func (c *VJClipsClient) GetMediaPath() string {
	return c.mediaPath
}

// ScanLocalClips scans local VJ clip directories
func (c *VJClipsClient) ScanLocalClips(ctx context.Context) ([]VJClipInfo, error) {
	clips := []VJClipInfo{}

	if _, err := os.Stat(c.mediaPath); os.IsNotExist(err) {
		return clips, nil
	}

	err := filepath.Walk(c.mediaPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !isVideoFormat(ext) {
			return nil
		}

		// Extract playlist from path structure
		relPath, _ := filepath.Rel(c.mediaPath, path)
		parts := strings.Split(relPath, string(filepath.Separator))

		playlist := "default"
		if len(parts) > 1 {
			playlist = parts[0]
		}

		clip := VJClipInfo{
			Name:       strings.TrimSuffix(info.Name(), ext),
			Filename:   info.Name(),
			User:       c.defaultUser,
			Playlist:   playlist,
			Format:     strings.TrimPrefix(ext, "."),
			LocalPath:  path,
			Size:       info.Size(),
			ModifiedAt: info.ModTime(),
			IsDXV:      strings.Contains(strings.ToLower(info.Name()), "dxv"),
		}

		// Calculate hash for smaller files
		if info.Size() < 500*1024*1024 { // < 500MB
			if hash, err := c.calculateSHA256(path); err == nil {
				clip.SHA256 = hash
			}
		}

		clips = append(clips, clip)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return clips, nil
}

func isVideoFormat(ext string) bool {
	videoExts := map[string]bool{
		".mov": true, ".mp4": true, ".avi": true,
		".mkv": true, ".webm": true, ".m4v": true,
		".dxv": true, ".hap": true,
	}
	return videoExts[ext]
}

func (c *VJClipsClient) calculateSHA256(path string) (string, error) {
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

// ListPlaylists lists all local playlists
func (c *VJClipsClient) ListPlaylists(ctx context.Context) ([]VJPlaylist, error) {
	clips, err := c.ScanLocalClips(ctx)
	if err != nil {
		return nil, err
	}

	// Group by playlist
	playlistMap := make(map[string]*VJPlaylist)
	for _, clip := range clips {
		if pl, ok := playlistMap[clip.Playlist]; ok {
			pl.ClipCount++
			pl.TotalSize += clip.Size
			if clip.ModifiedAt.After(pl.ModifiedAt) {
				pl.ModifiedAt = clip.ModifiedAt
			}
		} else {
			playlistMap[clip.Playlist] = &VJPlaylist{
				Name:       clip.Playlist,
				User:       c.defaultUser,
				ClipCount:  1,
				TotalSize:  clip.Size,
				ModifiedAt: clip.ModifiedAt,
			}
		}
	}

	playlists := make([]VJPlaylist, 0, len(playlistMap))
	for _, pl := range playlistMap {
		playlists = append(playlists, *pl)
	}

	return playlists, nil
}

// ListS3Clips lists clips in S3
func (c *VJClipsClient) ListS3Clips(ctx context.Context, user, playlist string) ([]VJClipInfo, error) {
	clips := []VJClipInfo{}

	prefix := ""
	if user != "" {
		prefix = user + "/"
		if playlist != "" {
			prefix = user + "/" + playlist + "/"
		}
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

			// Skip metadata files
			if strings.HasSuffix(key, ".json") {
				continue
			}

			// Skip directories
			if strings.HasSuffix(key, "/") {
				continue
			}

			ext := strings.ToLower(filepath.Ext(key))
			if !isVideoFormat(ext) {
				continue
			}

			// Parse user/playlist from key
			parts := strings.Split(key, "/")
			clipUser := c.defaultUser
			clipPlaylist := "default"
			if len(parts) >= 2 {
				clipUser = parts[0]
				clipPlaylist = parts[1]
			}

			clip := VJClipInfo{
				Name:       strings.TrimSuffix(filepath.Base(key), ext),
				Filename:   filepath.Base(key),
				User:       clipUser,
				Playlist:   clipPlaylist,
				Format:     strings.TrimPrefix(ext, "."),
				S3Key:      key,
				Size:       aws.ToInt64(obj.Size),
				ModifiedAt: aws.ToTime(obj.LastModified),
			}

			// Try to load metadata
			metaKey := key + ".json"
			if meta, err := c.loadS3Metadata(ctx, metaKey); err == nil {
				clip.Tags = meta.Tags
				clip.SHA256 = meta.SHA256
				clip.Duration = meta.Duration
				clip.Resolution = meta.Resolution
				clip.Framerate = meta.Framerate
				clip.Codec = meta.Codec
			}

			clips = append(clips, clip)
		}
	}

	return clips, nil
}

// ListS3Playlists lists playlists in S3
func (c *VJClipsClient) ListS3Playlists(ctx context.Context, user string) ([]VJPlaylist, error) {
	clips, err := c.ListS3Clips(ctx, user, "")
	if err != nil {
		return nil, err
	}

	// Group by user/playlist
	playlistMap := make(map[string]*VJPlaylist)
	for _, clip := range clips {
		key := clip.User + "/" + clip.Playlist
		if pl, ok := playlistMap[key]; ok {
			pl.ClipCount++
			pl.TotalSize += clip.Size
			if clip.ModifiedAt.After(pl.ModifiedAt) {
				pl.ModifiedAt = clip.ModifiedAt
			}
		} else {
			playlistMap[key] = &VJPlaylist{
				Name:       clip.Playlist,
				User:       clip.User,
				ClipCount:  1,
				TotalSize:  clip.Size,
				ModifiedAt: clip.ModifiedAt,
			}
		}
	}

	playlists := make([]VJPlaylist, 0, len(playlistMap))
	for _, pl := range playlistMap {
		playlists = append(playlists, *pl)
	}

	return playlists, nil
}

func (c *VJClipsClient) loadS3Metadata(ctx context.Context, key string) (*VJClipInfo, error) {
	output, err := c.s3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	defer output.Body.Close()

	var meta VJClipInfo
	if err := json.NewDecoder(output.Body).Decode(&meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

// UploadClip uploads a clip to S3
func (c *VJClipsClient) UploadClip(ctx context.Context, localPath string, opts ClipUploadOptions) (*VJClipInfo, error) {
	info, err := os.Stat(localPath)
	if err != nil {
		return nil, fmt.Errorf("clip not found: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(localPath))
	if !isVideoFormat(ext) {
		return nil, fmt.Errorf("not a valid video file: %s", ext)
	}

	// Build S3 key: user/playlist/filename
	user := opts.User
	if user == "" {
		user = c.defaultUser
	}
	playlist := opts.Playlist
	if playlist == "" {
		playlist = "default"
	}
	s3Key := fmt.Sprintf("%s/%s/%s", user, playlist, filepath.Base(localPath))

	// Calculate SHA256 for smaller files
	var hash string
	if info.Size() < 500*1024*1024 {
		hash, _ = c.calculateSHA256(localPath)
	}

	// Open and upload file
	f, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open clip: %w", err)
	}
	defer f.Close()

	_, err = c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(s3Key),
		Body:   f,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload clip: %w", err)
	}

	// Create metadata
	now := time.Now()
	uploadedBy := os.Getenv("USER")

	clip := &VJClipInfo{
		Name:       strings.TrimSuffix(filepath.Base(localPath), ext),
		Filename:   filepath.Base(localPath),
		User:       user,
		Playlist:   playlist,
		Format:     strings.TrimPrefix(ext, "."),
		S3Key:      s3Key,
		SHA256:     hash,
		Size:       info.Size(),
		UploadedAt: &now,
		UploadedBy: uploadedBy,
		Tags:       opts.Tags,
		ModifiedAt: info.ModTime(),
	}

	// Upload metadata
	metaKey := s3Key + ".json"
	metaData, _ := json.MarshalIndent(clip, "", "  ")

	_, err = c.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(c.bucketName),
		Key:         aws.String(metaKey),
		Body:        strings.NewReader(string(metaData)),
		ContentType: aws.String("application/json"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload metadata: %w", err)
	}

	return clip, nil
}

// ClipUploadOptions configures clip upload
type ClipUploadOptions struct {
	User     string
	Playlist string
	Tags     []string
}

// DownloadClip downloads a clip from S3
func (c *VJClipsClient) DownloadClip(ctx context.Context, s3Key string, destPlaylist string) (string, error) {
	// Sanitize filename to prevent path traversal
	filename := filepath.Base(s3Key)
	filename = strings.ReplaceAll(filename, "..", "")
	filename = strings.ReplaceAll(filename, "/", "")
	filename = strings.ReplaceAll(filename, "\\", "")
	if filename == "" || filename == "." {
		return "", fmt.Errorf("invalid clip filename")
	}

	// Determine destination
	destDir := c.mediaPath
	if destPlaylist != "" {
		// Sanitize playlist name too
		safePlaylist := filepath.Base(destPlaylist)
		safePlaylist = strings.ReplaceAll(safePlaylist, "..", "")
		destDir = filepath.Join(c.mediaPath, safePlaylist)
	} else {
		// Use playlist from S3 key
		parts := strings.Split(s3Key, "/")
		if len(parts) >= 2 {
			safePlaylist := filepath.Base(parts[1])
			safePlaylist = strings.ReplaceAll(safePlaylist, "..", "")
			destDir = filepath.Join(c.mediaPath, safePlaylist)
		}
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	destPath := filepath.Join(destDir, filename)

	// Verify path is within media directory (defense in depth)
	absDestDir, _ := filepath.Abs(c.mediaPath)
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
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}

	// Calculate hash while writing
	h := sha256.New()
	writer := io.MultiWriter(f, h)

	if _, err := io.Copy(writer, output.Body); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	f.Close()

	// Verify hash if available
	if expectedHash != "" {
		actualHash := hex.EncodeToString(h.Sum(nil))
		if actualHash != expectedHash {
			os.Remove(tmpPath)
			return "", fmt.Errorf("hash mismatch: expected %s, got %s", expectedHash[:16]+"...", actualHash[:16]+"...")
		}
	}

	// Atomic rename
	if err := os.Rename(tmpPath, destPath); err != nil {
		os.Remove(tmpPath)
		return "", fmt.Errorf("failed to finalize file: %w", err)
	}

	return destPath, nil
}

// UploadPlaylist uploads all clips in a local playlist
func (c *VJClipsClient) UploadPlaylist(ctx context.Context, playlistName string, user string) (*PlaylistUploadResult, error) {
	clips, err := c.ScanLocalClips(ctx)
	if err != nil {
		return nil, err
	}

	result := &PlaylistUploadResult{
		Uploaded: []VJClipInfo{},
		Errors:   []string{},
	}

	for _, clip := range clips {
		if clip.Playlist != playlistName {
			continue
		}

		opts := ClipUploadOptions{
			User:     user,
			Playlist: playlistName,
		}

		uploaded, err := c.UploadClip(ctx, clip.LocalPath, opts)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", clip.Filename, err))
			continue
		}

		result.Uploaded = append(result.Uploaded, *uploaded)
	}

	return result, nil
}

// DownloadPlaylist downloads all clips in an S3 playlist
func (c *VJClipsClient) DownloadPlaylist(ctx context.Context, user, playlist string) (*PlaylistDownloadResult, error) {
	clips, err := c.ListS3Clips(ctx, user, playlist)
	if err != nil {
		return nil, err
	}

	result := &PlaylistDownloadResult{
		Downloaded: []string{},
		Errors:     []string{},
	}

	for _, clip := range clips {
		path, err := c.DownloadClip(ctx, clip.S3Key, playlist)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", clip.Filename, err))
			continue
		}
		result.Downloaded = append(result.Downloaded, path)
	}

	return result, nil
}

// SyncClips compares local and S3 clips
func (c *VJClipsClient) SyncClips(ctx context.Context, user string) (*ClipSyncResult, error) {
	result := &ClipSyncResult{}

	localClips, err := c.ScanLocalClips(ctx)
	if err != nil {
		return nil, err
	}

	s3Clips, err := c.ListS3Clips(ctx, user, "")
	if err != nil {
		return nil, err
	}

	// Build maps
	localMap := make(map[string]VJClipInfo)
	for _, c := range localClips {
		key := c.Playlist + "/" + c.Filename
		localMap[key] = c
	}

	s3Map := make(map[string]VJClipInfo)
	for _, c := range s3Clips {
		key := c.Playlist + "/" + c.Filename
		s3Map[key] = c
	}

	// Find differences
	for key, local := range localMap {
		if _, exists := s3Map[key]; !exists {
			result.ToUpload = append(result.ToUpload, local)
		}
	}

	for key, s3Clip := range s3Map {
		if _, exists := localMap[key]; !exists {
			result.ToDownload = append(result.ToDownload, s3Clip)
		}
	}

	return result, nil
}

// ClipSyncResult contains sync comparison results
type ClipSyncResult struct {
	ToUpload   []VJClipInfo `json:"to_upload"`
	ToDownload []VJClipInfo `json:"to_download"`
}

// PlaylistUploadResult contains results from uploading a playlist
type PlaylistUploadResult struct {
	Uploaded []VJClipInfo `json:"uploaded"`
	Errors   []string     `json:"errors,omitempty"`
}

// PlaylistDownloadResult contains results from downloading a playlist
type PlaylistDownloadResult struct {
	Downloaded []string `json:"downloaded"`
	Errors     []string `json:"errors,omitempty"`
}

// DeleteClip deletes a clip from S3
func (c *VJClipsClient) DeleteClip(ctx context.Context, s3Key string) error {
	// Delete clip
	_, err := c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(s3Key),
	})
	if err != nil {
		return err
	}

	// Delete metadata
	metaKey := s3Key + ".json"
	c.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(metaKey),
	})

	return nil
}

// GetHealth returns VJ clips system health
func (c *VJClipsClient) GetHealth(ctx context.Context) (*VJClipsHealth, error) {
	health := &VJClipsHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check local clips
	localClips, err := c.ScanLocalClips(ctx)
	if err != nil {
		health.Score -= 20
		health.Issues = append(health.Issues, "Failed to scan local clips")
	} else {
		health.LocalCount = len(localClips)
		for _, c := range localClips {
			health.LocalSize += c.Size
		}
	}

	// Check S3
	s3Clips, err := c.ListS3Clips(ctx, "", "")
	if err != nil {
		health.Score -= 30
		health.Issues = append(health.Issues, "Failed to connect to S3 bucket")
	} else {
		health.S3Count = len(s3Clips)
		for _, c := range s3Clips {
			health.S3Size += c.Size
		}
	}

	// Check media directory
	if _, err := os.Stat(c.mediaPath); os.IsNotExist(err) {
		health.Score -= 10
		health.Recommendations = append(health.Recommendations,
			fmt.Sprintf("Create media directory: %s", c.mediaPath))
	}

	// Set status
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

// VJClipsHealth represents VJ clips system health
type VJClipsHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	LocalCount      int      `json:"local_count"`
	LocalSize       int64    `json:"local_size"`
	S3Count         int      `json:"s3_count"`
	S3Size          int64    `json:"s3_size"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}
