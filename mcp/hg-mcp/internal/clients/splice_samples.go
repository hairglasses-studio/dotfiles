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
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// SpliceSampleInfo represents metadata about a Splice sample
type SpliceSampleInfo struct {
	Name        string    `json:"name"`
	Filename    string    `json:"filename"`
	Pack        string    `json:"pack"`                  // Sample pack name
	Category    string    `json:"category"`              // drums, bass, synth, vocals, fx, etc.
	Subcategory string    `json:"subcategory,omitempty"` // kicks, snares, etc.
	Format      string    `json:"format"`                // wav, mp3, aiff, flac
	BPM         float64   `json:"bpm,omitempty"`
	Key         string    `json:"key,omitempty"` // Musical key (C, Am, etc.)
	LocalPath   string    `json:"local_path,omitempty"`
	GDrivePath  string    `json:"gdrive_path,omitempty"`
	SHA256      string    `json:"sha256,omitempty"`
	Size        int64     `json:"size"`
	Duration    float64   `json:"duration,omitempty"` // seconds
	SampleRate  int       `json:"sample_rate,omitempty"`
	BitDepth    int       `json:"bit_depth,omitempty"`
	Channels    int       `json:"channels,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	ModifiedAt  time.Time `json:"modified_at"`
	IsLoop      bool      `json:"is_loop,omitempty"`
	IsOneShot   bool      `json:"is_oneshot,omitempty"`
}

// SplicePack represents a sample pack
type SplicePack struct {
	Name        string             `json:"name"`
	Artist      string             `json:"artist,omitempty"`
	SampleCount int                `json:"sample_count"`
	TotalSize   int64              `json:"total_size"`
	Categories  map[string]int     `json:"categories"` // category -> count
	Samples     []SpliceSampleInfo `json:"samples,omitempty"`
	ModifiedAt  time.Time          `json:"modified_at"`
	LocalPath   string             `json:"local_path,omitempty"`
}

// SpliceCatalog represents the full sample catalog
type SpliceCatalog struct {
	Version      int            `json:"version"`
	LastUpdated  time.Time      `json:"last_updated"`
	Packs        []SplicePack   `json:"packs"`
	TotalSamples int            `json:"total_samples"`
	TotalSize    int64          `json:"total_size"`
	ByCategory   map[string]int `json:"by_category"`
	ByFormat     map[string]int `json:"by_format"`
}

// SpliceSamplesClient provides Splice sample management
type SpliceSamplesClient struct {
	s3Client   *s3.Client
	bucketName string
	homeDir    string
	splicePath string
	gdrivePath string
	mu         sync.RWMutex
}

var (
	spliceSamplesClient     *SpliceSamplesClient
	spliceSamplesClientOnce sync.Once
	spliceSamplesClientErr  error
)

// GetSpliceSamplesClient returns the singleton Splice samples client
func GetSpliceSamplesClient() (*SpliceSamplesClient, error) {
	spliceSamplesClientOnce.Do(func() {
		spliceSamplesClient, spliceSamplesClientErr = NewSpliceSamplesClient()
	})
	return spliceSamplesClient, spliceSamplesClientErr
}

// NewSpliceSamplesClient creates a new Splice samples client
func NewSpliceSamplesClient() (*SpliceSamplesClient, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	bucketName := os.Getenv("SPLICE_SAMPLES_BUCKET")
	if bucketName == "" {
		bucketName = "aftrs-splice-samples"
	}

	// Default Splice path (can be overridden)
	splicePath := os.Getenv("SPLICE_PATH")
	if splicePath == "" {
		splicePath = filepath.Join(homeDir, "Documents", "Splice")
	}

	// Google Drive rclone remote path
	gdrivePath := os.Getenv("SPLICE_GDRIVE_PATH")
	if gdrivePath == "" {
		gdrivePath = "gdrive:/Music Production/Splice"
	}

	return &SpliceSamplesClient{
		s3Client:   s3.NewFromConfig(cfg),
		bucketName: bucketName,
		homeDir:    homeDir,
		splicePath: splicePath,
		gdrivePath: gdrivePath,
	}, nil
}

// GetSplicePath returns the local Splice path
func (c *SpliceSamplesClient) GetSplicePath() string {
	return c.splicePath
}

// ScanLocalSamples scans local Splice sample directories
func (c *SpliceSamplesClient) ScanLocalSamples(ctx context.Context) ([]SpliceSampleInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	samples := []SpliceSampleInfo{}

	if _, err := os.Stat(c.splicePath); os.IsNotExist(err) {
		return samples, nil
	}

	err := filepath.Walk(c.splicePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !isAudioFormat(ext) {
			return nil
		}

		// Extract pack/category from path structure
		// Typical: Splice/sounds/packs/PackName/Category/Subcategory/sample.wav
		relPath, _ := filepath.Rel(c.splicePath, path)
		parts := strings.Split(relPath, string(filepath.Separator))

		pack := "Unknown"
		category := "Unknown"
		subcategory := ""

		// Parse path structure
		if len(parts) >= 2 {
			// Skip "sounds/packs" or similar prefixes
			startIdx := 0
			for i, p := range parts {
				if p == "packs" || p == "samples" {
					startIdx = i + 1
					break
				}
			}

			if startIdx < len(parts) {
				pack = parts[startIdx]
			}
			if startIdx+1 < len(parts)-1 {
				category = parts[startIdx+1]
			}
			if startIdx+2 < len(parts)-1 {
				subcategory = parts[startIdx+2]
			}
		}

		// Normalize category
		category = normalizeSampleCategory(category)

		sample := SpliceSampleInfo{
			Name:        strings.TrimSuffix(info.Name(), ext),
			Filename:    info.Name(),
			Pack:        pack,
			Category:    category,
			Subcategory: subcategory,
			Format:      strings.TrimPrefix(ext, "."),
			LocalPath:   path,
			Size:        info.Size(),
			ModifiedAt:  info.ModTime(),
			IsLoop:      detectIsLoop(info.Name()),
			IsOneShot:   detectIsOneShot(info.Name()),
		}

		// Try to detect BPM and key from filename
		sample.BPM = detectBPM(info.Name())
		sample.Key = detectKey(info.Name())

		// Calculate hash for smaller files (< 50MB)
		if info.Size() < 50*1024*1024 {
			if hash, err := calculateFileHash(path); err == nil {
				sample.SHA256 = hash
			}
		}

		samples = append(samples, sample)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return samples, nil
}

func isAudioFormat(ext string) bool {
	audioExts := map[string]bool{
		".wav": true, ".mp3": true, ".aiff": true, ".aif": true,
		".flac": true, ".ogg": true, ".m4a": true, ".opus": true,
	}
	return audioExts[ext]
}

func normalizeSampleCategory(cat string) string {
	cat = strings.ToLower(cat)
	switch {
	case strings.Contains(cat, "drum") || strings.Contains(cat, "beat"):
		return "Drums"
	case strings.Contains(cat, "kick"):
		return "Drums/Kicks"
	case strings.Contains(cat, "snare"):
		return "Drums/Snares"
	case strings.Contains(cat, "hihat") || strings.Contains(cat, "hi-hat") || strings.Contains(cat, "hat"):
		return "Drums/Hats"
	case strings.Contains(cat, "perc"):
		return "Drums/Percussion"
	case strings.Contains(cat, "bass"):
		return "Bass"
	case strings.Contains(cat, "synth") || strings.Contains(cat, "lead"):
		return "Synths"
	case strings.Contains(cat, "pad"):
		return "Pads"
	case strings.Contains(cat, "vocal") || strings.Contains(cat, "vox"):
		return "Vocals"
	case strings.Contains(cat, "fx") || strings.Contains(cat, "effect") || strings.Contains(cat, "sfx"):
		return "FX"
	case strings.Contains(cat, "loop"):
		return "Loops"
	case strings.Contains(cat, "oneshot") || strings.Contains(cat, "one-shot") || strings.Contains(cat, "one shot"):
		return "One-Shots"
	case strings.Contains(cat, "foley"):
		return "Foley"
	case strings.Contains(cat, "guitar"):
		return "Guitar"
	case strings.Contains(cat, "piano") || strings.Contains(cat, "keys"):
		return "Keys"
	case strings.Contains(cat, "string"):
		return "Strings"
	case strings.Contains(cat, "brass") || strings.Contains(cat, "horn"):
		return "Brass"
	case strings.Contains(cat, "texture") || strings.Contains(cat, "ambient"):
		return "Textures"
	default:
		if cat == "" || cat == "unknown" {
			return "Uncategorized"
		}
		// Capitalize first letter
		return strings.ToUpper(cat[:1]) + cat[1:]
	}
}

func detectIsLoop(filename string) bool {
	lower := strings.ToLower(filename)
	return strings.Contains(lower, "loop") ||
		strings.Contains(lower, "_lp") ||
		strings.Contains(lower, "-lp")
}

func detectIsOneShot(filename string) bool {
	lower := strings.ToLower(filename)
	return strings.Contains(lower, "oneshot") ||
		strings.Contains(lower, "one-shot") ||
		strings.Contains(lower, "one shot") ||
		strings.Contains(lower, "_os") ||
		strings.Contains(lower, "-os")
}

func detectBPM(filename string) float64 {
	lower := strings.ToLower(filename)
	// Common patterns: 120bpm, 120_bpm, 120-bpm, bpm120
	patterns := []string{"bpm", "_bpm", "-bpm"}
	for _, pattern := range patterns {
		if idx := strings.Index(lower, pattern); idx != -1 {
			// Try to extract number before or after
			var bpm float64
			// Check before pattern
			if idx > 0 {
				start := idx - 1
				for start > 0 && (lower[start-1] >= '0' && lower[start-1] <= '9') {
					start--
				}
				if start < idx {
					fmt.Sscanf(lower[start:idx], "%f", &bpm)
				}
			}
			// Check after pattern
			if bpm == 0 {
				after := lower[idx+len(pattern):]
				fmt.Sscanf(after, "%f", &bpm)
			}
			if bpm >= 60 && bpm <= 200 {
				return bpm
			}
		}
	}
	return 0
}

func detectKey(filename string) string {
	lower := strings.ToLower(filename)
	// Common key patterns
	keys := []string{
		"cmaj", "c maj", "c major", "cmin", "c min", "c minor",
		"dmaj", "d maj", "d major", "dmin", "d min", "d minor",
		"emaj", "e maj", "e major", "emin", "e min", "e minor",
		"fmaj", "f maj", "f major", "fmin", "f min", "f minor",
		"gmaj", "g maj", "g major", "gmin", "g min", "g minor",
		"amaj", "a maj", "a major", "amin", "a min", "a minor",
		"bmaj", "b maj", "b major", "bmin", "b min", "b minor",
	}

	for _, k := range keys {
		if strings.Contains(lower, k) {
			// Return normalized key
			note := strings.ToUpper(string(k[0]))
			if strings.Contains(k, "min") {
				return note + "m"
			}
			return note
		}
	}

	// Check for sharp/flat patterns like C#, Db
	sharpFlat := []string{"c#", "db", "d#", "eb", "f#", "gb", "g#", "ab", "a#", "bb"}
	for _, sf := range sharpFlat {
		if strings.Contains(lower, sf) {
			return strings.ToUpper(sf[:1]) + sf[1:]
		}
	}

	return ""
}

func calculateFileHash(path string) (string, error) {
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

// ListPacks lists all sample packs
func (c *SpliceSamplesClient) ListPacks(ctx context.Context) ([]SplicePack, error) {
	samples, err := c.ScanLocalSamples(ctx)
	if err != nil {
		return nil, err
	}

	// Group by pack
	packMap := make(map[string]*SplicePack)
	for _, sample := range samples {
		if pl, ok := packMap[sample.Pack]; ok {
			pl.SampleCount++
			pl.TotalSize += sample.Size
			pl.Categories[sample.Category]++
			if sample.ModifiedAt.After(pl.ModifiedAt) {
				pl.ModifiedAt = sample.ModifiedAt
			}
		} else {
			packMap[sample.Pack] = &SplicePack{
				Name:        sample.Pack,
				SampleCount: 1,
				TotalSize:   sample.Size,
				Categories:  map[string]int{sample.Category: 1},
				ModifiedAt:  sample.ModifiedAt,
			}
		}
	}

	packs := make([]SplicePack, 0, len(packMap))
	for _, pl := range packMap {
		packs = append(packs, *pl)
	}

	return packs, nil
}

// SearchSamples searches samples by query
func (c *SpliceSamplesClient) SearchSamples(ctx context.Context, query string, category string, limit int) ([]SpliceSampleInfo, error) {
	samples, err := c.ScanLocalSamples(ctx)
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []SpliceSampleInfo

	for _, sample := range samples {
		// Category filter
		if category != "" && !strings.EqualFold(sample.Category, category) &&
			!strings.HasPrefix(strings.ToLower(sample.Category), strings.ToLower(category)) {
			continue
		}

		// Query filter
		if query != "" {
			if !strings.Contains(strings.ToLower(sample.Name), query) &&
				!strings.Contains(strings.ToLower(sample.Pack), query) &&
				!strings.Contains(strings.ToLower(sample.Filename), query) {
				continue
			}
		}

		results = append(results, sample)
		if limit > 0 && len(results) >= limit {
			break
		}
	}

	return results, nil
}

// GetCatalog returns the full sample catalog
func (c *SpliceSamplesClient) GetCatalog(ctx context.Context) (*SpliceCatalog, error) {
	samples, err := c.ScanLocalSamples(ctx)
	if err != nil {
		return nil, err
	}

	packs, err := c.ListPacks(ctx)
	if err != nil {
		return nil, err
	}

	catalog := &SpliceCatalog{
		Version:      1,
		LastUpdated:  time.Now(),
		Packs:        packs,
		TotalSamples: len(samples),
		ByCategory:   make(map[string]int),
		ByFormat:     make(map[string]int),
	}

	for _, sample := range samples {
		catalog.TotalSize += sample.Size
		catalog.ByCategory[sample.Category]++
		catalog.ByFormat[sample.Format]++
	}

	return catalog, nil
}

// GetPackSamples returns all samples in a pack
func (c *SpliceSamplesClient) GetPackSamples(ctx context.Context, packName string) ([]SpliceSampleInfo, error) {
	samples, err := c.ScanLocalSamples(ctx)
	if err != nil {
		return nil, err
	}

	var packSamples []SpliceSampleInfo
	for _, sample := range samples {
		if strings.EqualFold(sample.Pack, packName) {
			packSamples = append(packSamples, sample)
		}
	}

	return packSamples, nil
}

// GetSamplesByCategory returns samples filtered by category
func (c *SpliceSamplesClient) GetSamplesByCategory(ctx context.Context, category string) ([]SpliceSampleInfo, error) {
	samples, err := c.ScanLocalSamples(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []SpliceSampleInfo
	for _, sample := range samples {
		if strings.EqualFold(sample.Category, category) ||
			strings.HasPrefix(strings.ToLower(sample.Category), strings.ToLower(category)+"/") {
			filtered = append(filtered, sample)
		}
	}

	return filtered, nil
}

// ExportCatalog exports the catalog to JSON
func (c *SpliceSamplesClient) ExportCatalog(ctx context.Context, outputPath string) error {
	catalog, err := c.GetCatalog(ctx)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(outputPath, data, 0644)
}

// GetHealth returns Splice samples system health
func (c *SpliceSamplesClient) GetHealth(ctx context.Context) (*SpliceSamplesHealth, error) {
	health := &SpliceSamplesHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check local samples
	samples, err := c.ScanLocalSamples(ctx)
	if err != nil {
		health.Score -= 30
		health.Issues = append(health.Issues, "Failed to scan local samples")
	} else {
		health.LocalCount = len(samples)
		for _, s := range samples {
			health.LocalSize += s.Size
		}
	}

	// Check Splice directory
	if _, err := os.Stat(c.splicePath); os.IsNotExist(err) {
		health.Score -= 20
		health.Recommendations = append(health.Recommendations,
			fmt.Sprintf("Create Splice directory: %s", c.splicePath))
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

// SpliceSamplesHealth represents Splice samples system health
type SpliceSamplesHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	LocalCount      int      `json:"local_count"`
	LocalSize       int64    `json:"local_size"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}
