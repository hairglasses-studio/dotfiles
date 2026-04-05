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
	"regexp"
	"strings"
	"sync"
	"time"
)

// MilkDropPresetInfo represents metadata about a MilkDrop preset
type MilkDropPresetInfo struct {
	Name       string    `json:"name"`
	Filename   string    `json:"filename"`
	Author     string    `json:"author,omitempty"`
	Pack       string    `json:"pack"`     // Preset pack/collection name
	Category   string    `json:"category"` // abstract, geometric, trippy, etc.
	Format     string    `json:"format"`   // milk (MilkDrop preset)
	LocalPath  string    `json:"local_path,omitempty"`
	GDrivePath string    `json:"gdrive_path,omitempty"`
	SHA256     string    `json:"sha256,omitempty"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modified_at"`
	HasImage   bool      `json:"has_image,omitempty"` // Has accompanying .jpg
	ImagePath  string    `json:"image_path,omitempty"`
	Complexity string    `json:"complexity,omitempty"` // simple, moderate, complex
	Tags       []string  `json:"tags,omitempty"`
	// Shader info
	UsesShader    bool   `json:"uses_shader,omitempty"`
	ShaderType    string `json:"shader_type,omitempty"` // comp, warp, or both
	EquationCount int    `json:"equation_count,omitempty"`
}

// MilkDropPack represents a preset pack/collection
type MilkDropPack struct {
	Name        string               `json:"name"`
	Author      string               `json:"author,omitempty"`
	PresetCount int                  `json:"preset_count"`
	TotalSize   int64                `json:"total_size"`
	Categories  map[string]int       `json:"categories"`
	Presets     []MilkDropPresetInfo `json:"presets,omitempty"`
	ModifiedAt  time.Time            `json:"modified_at"`
	LocalPath   string               `json:"local_path,omitempty"`
}

// MilkDropCatalog represents the full preset catalog
type MilkDropCatalog struct {
	Version      int            `json:"version"`
	LastUpdated  time.Time      `json:"last_updated"`
	Packs        []MilkDropPack `json:"packs"`
	TotalPresets int            `json:"total_presets"`
	TotalSize    int64          `json:"total_size"`
	ByAuthor     map[string]int `json:"by_author"`
	ByCategory   map[string]int `json:"by_category"`
	WithShaders  int            `json:"with_shaders"`
	WithImages   int            `json:"with_images"`
}

// MilkDropPresetsClient provides MilkDrop preset management
type MilkDropPresetsClient struct {
	homeDir      string
	nestDropPath string
	gdrivePath   string
	mu           sync.RWMutex
}

var (
	milkDropPresetsClient     *MilkDropPresetsClient
	milkDropPresetsClientOnce sync.Once
	milkDropPresetsClientErr  error
)

// GetMilkDropPresetsClient returns the singleton MilkDrop presets client
func GetMilkDropPresetsClient() (*MilkDropPresetsClient, error) {
	milkDropPresetsClientOnce.Do(func() {
		milkDropPresetsClient, milkDropPresetsClientErr = NewMilkDropPresetsClient()
	})
	return milkDropPresetsClient, milkDropPresetsClientErr
}

// NewMilkDropPresetsClient creates a new MilkDrop presets client
func NewMilkDropPresetsClient() (*MilkDropPresetsClient, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	// Default NestDrop path (can be overridden)
	nestDropPath := os.Getenv("NESTDROP_PATH")
	if nestDropPath == "" {
		nestDropPath = filepath.Join(homeDir, "Documents", "NestDropProV2")
	}

	// Google Drive rclone remote path
	gdrivePath := os.Getenv("NESTDROP_GDRIVE_PATH")
	if gdrivePath == "" {
		gdrivePath = "gdrive:/Visual Production/NestDropProV2"
	}

	return &MilkDropPresetsClient{
		homeDir:      homeDir,
		nestDropPath: nestDropPath,
		gdrivePath:   gdrivePath,
	}, nil
}

// GetNestDropPath returns the local NestDrop path
func (c *MilkDropPresetsClient) GetNestDropPath() string {
	return c.nestDropPath
}

// ScanLocalPresets scans local MilkDrop preset directories
func (c *MilkDropPresetsClient) ScanLocalPresets(ctx context.Context) ([]MilkDropPresetInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	presets := []MilkDropPresetInfo{}

	// Look for presets in Plugins/Milkdrop2/Presets
	presetsPath := filepath.Join(c.nestDropPath, "Plugins", "Milkdrop2", "Presets")
	if _, err := os.Stat(presetsPath); os.IsNotExist(err) {
		// Try alternate path
		presetsPath = filepath.Join(c.nestDropPath, "Presets")
		if _, err := os.Stat(presetsPath); os.IsNotExist(err) {
			return presets, nil
		}
	}

	// Author regex patterns
	authorRegex := regexp.MustCompile(`^([^-_]+)[-_]`)

	err := filepath.Walk(presetsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".milk" {
			return nil
		}

		// Extract pack from directory structure
		relPath, _ := filepath.Rel(presetsPath, path)
		parts := strings.Split(relPath, string(filepath.Separator))

		pack := "Default"
		if len(parts) > 1 {
			pack = parts[0]
		}

		// Try to extract author from filename
		author := ""
		baseName := strings.TrimSuffix(info.Name(), ext)
		if matches := authorRegex.FindStringSubmatch(baseName); len(matches) > 1 {
			author = matches[1]
		}

		// Check for accompanying image
		imagePath := strings.TrimSuffix(path, ext) + ".jpg"
		hasImage := false
		if _, err := os.Stat(imagePath); err == nil {
			hasImage = true
		}

		// Read preset to analyze shader usage
		usesShader := false
		shaderType := ""
		equationCount := 0
		if data, err := os.ReadFile(path); err == nil {
			content := string(data)
			if strings.Contains(content, "comp_") || strings.Contains(content, "warp_") {
				usesShader = true
				if strings.Contains(content, "comp_") && strings.Contains(content, "warp_") {
					shaderType = "both"
				} else if strings.Contains(content, "comp_") {
					shaderType = "comp"
				} else {
					shaderType = "warp"
				}
			}
			// Count equations (per_frame_ and per_pixel_ lines)
			equationCount = strings.Count(content, "per_frame_") + strings.Count(content, "per_pixel_")
		}

		// Determine complexity
		complexity := "simple"
		if equationCount > 50 || usesShader {
			complexity = "complex"
		} else if equationCount > 20 {
			complexity = "moderate"
		}

		// Categorize based on author/pack name patterns
		category := categorizeMilkDropPreset(baseName, pack)

		preset := MilkDropPresetInfo{
			Name:          baseName,
			Filename:      info.Name(),
			Author:        author,
			Pack:          pack,
			Category:      category,
			Format:        "milk",
			LocalPath:     path,
			Size:          info.Size(),
			ModifiedAt:    info.ModTime(),
			HasImage:      hasImage,
			UsesShader:    usesShader,
			ShaderType:    shaderType,
			EquationCount: equationCount,
			Complexity:    complexity,
		}

		if hasImage {
			preset.ImagePath = imagePath
		}

		// Calculate hash for presets (they're small)
		if hash, err := calculatePresetHash(path); err == nil {
			preset.SHA256 = hash
		}

		presets = append(presets, preset)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return presets, nil
}

func categorizeMilkDropPreset(name, pack string) string {
	lower := strings.ToLower(name + " " + pack)

	switch {
	case strings.Contains(lower, "abstract"):
		return "Abstract"
	case strings.Contains(lower, "geometric") || strings.Contains(lower, "geometry"):
		return "Geometric"
	case strings.Contains(lower, "trippy") || strings.Contains(lower, "psychedelic"):
		return "Psychedelic"
	case strings.Contains(lower, "tunnel") || strings.Contains(lower, "wormhole"):
		return "Tunnels"
	case strings.Contains(lower, "fractal") || strings.Contains(lower, "mandelbrot"):
		return "Fractals"
	case strings.Contains(lower, "space") || strings.Contains(lower, "star") || strings.Contains(lower, "galaxy"):
		return "Space"
	case strings.Contains(lower, "water") || strings.Contains(lower, "ocean") || strings.Contains(lower, "liquid"):
		return "Liquid"
	case strings.Contains(lower, "fire") || strings.Contains(lower, "flame"):
		return "Fire"
	case strings.Contains(lower, "particle") || strings.Contains(lower, "dust"):
		return "Particles"
	case strings.Contains(lower, "wave") || strings.Contains(lower, "audio"):
		return "Audio Reactive"
	case strings.Contains(lower, "mirror") || strings.Contains(lower, "kaleidoscope"):
		return "Kaleidoscope"
	case strings.Contains(lower, "classic") || strings.Contains(lower, "retro"):
		return "Classic"
	default:
		return "General"
	}
}

func calculatePresetHash(path string) (string, error) {
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

// ListPacks lists all preset packs
func (c *MilkDropPresetsClient) ListPacks(ctx context.Context) ([]MilkDropPack, error) {
	presets, err := c.ScanLocalPresets(ctx)
	if err != nil {
		return nil, err
	}

	// Group by pack
	packMap := make(map[string]*MilkDropPack)
	for _, preset := range presets {
		if p, ok := packMap[preset.Pack]; ok {
			p.PresetCount++
			p.TotalSize += preset.Size
			p.Categories[preset.Category]++
			if preset.Author != "" && p.Author == "" {
				p.Author = preset.Author
			}
			if preset.ModifiedAt.After(p.ModifiedAt) {
				p.ModifiedAt = preset.ModifiedAt
			}
		} else {
			packMap[preset.Pack] = &MilkDropPack{
				Name:        preset.Pack,
				Author:      preset.Author,
				PresetCount: 1,
				TotalSize:   preset.Size,
				Categories:  map[string]int{preset.Category: 1},
				ModifiedAt:  preset.ModifiedAt,
			}
		}
	}

	packs := make([]MilkDropPack, 0, len(packMap))
	for _, p := range packMap {
		packs = append(packs, *p)
	}

	return packs, nil
}

// SearchPresets searches presets by query
func (c *MilkDropPresetsClient) SearchPresets(ctx context.Context, query string, filters PresetSearchFilters) ([]MilkDropPresetInfo, error) {
	presets, err := c.ScanLocalPresets(ctx)
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []MilkDropPresetInfo

	for _, preset := range presets {
		// Category filter
		if filters.Category != "" && !strings.EqualFold(preset.Category, filters.Category) {
			continue
		}

		// Author filter
		if filters.Author != "" && !strings.EqualFold(preset.Author, filters.Author) {
			continue
		}

		// Shader filter
		if filters.HasShader && !preset.UsesShader {
			continue
		}

		// Complexity filter
		if filters.Complexity != "" && preset.Complexity != filters.Complexity {
			continue
		}

		// Query filter
		if query != "" {
			if !strings.Contains(strings.ToLower(preset.Name), query) &&
				!strings.Contains(strings.ToLower(preset.Author), query) &&
				!strings.Contains(strings.ToLower(preset.Pack), query) {
				continue
			}
		}

		results = append(results, preset)
		if filters.Limit > 0 && len(results) >= filters.Limit {
			break
		}
	}

	return results, nil
}

// PresetSearchFilters contains search filter options
type PresetSearchFilters struct {
	Category   string `json:"category,omitempty"`
	Author     string `json:"author,omitempty"`
	HasShader  bool   `json:"has_shader,omitempty"`
	Complexity string `json:"complexity,omitempty"`
	Limit      int    `json:"limit,omitempty"`
}

// GetCatalog returns the full preset catalog
func (c *MilkDropPresetsClient) GetCatalog(ctx context.Context) (*MilkDropCatalog, error) {
	presets, err := c.ScanLocalPresets(ctx)
	if err != nil {
		return nil, err
	}

	packs, err := c.ListPacks(ctx)
	if err != nil {
		return nil, err
	}

	catalog := &MilkDropCatalog{
		Version:      1,
		LastUpdated:  time.Now(),
		Packs:        packs,
		TotalPresets: len(presets),
		ByAuthor:     make(map[string]int),
		ByCategory:   make(map[string]int),
	}

	for _, preset := range presets {
		catalog.TotalSize += preset.Size
		if preset.Author != "" {
			catalog.ByAuthor[preset.Author]++
		}
		catalog.ByCategory[preset.Category]++
		if preset.UsesShader {
			catalog.WithShaders++
		}
		if preset.HasImage {
			catalog.WithImages++
		}
	}

	return catalog, nil
}

// GetPresetsByAuthor returns all presets by an author
func (c *MilkDropPresetsClient) GetPresetsByAuthor(ctx context.Context, author string) ([]MilkDropPresetInfo, error) {
	presets, err := c.ScanLocalPresets(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []MilkDropPresetInfo
	for _, preset := range presets {
		if strings.EqualFold(preset.Author, author) {
			filtered = append(filtered, preset)
		}
	}

	return filtered, nil
}

// GetPresetsByCategory returns presets filtered by category
func (c *MilkDropPresetsClient) GetPresetsByCategory(ctx context.Context, category string) ([]MilkDropPresetInfo, error) {
	presets, err := c.ScanLocalPresets(ctx)
	if err != nil {
		return nil, err
	}

	var filtered []MilkDropPresetInfo
	for _, preset := range presets {
		if strings.EqualFold(preset.Category, category) {
			filtered = append(filtered, preset)
		}
	}

	return filtered, nil
}

// GetRandomPresets returns random presets for DJ mode
func (c *MilkDropPresetsClient) GetRandomPresets(ctx context.Context, count int, filters PresetSearchFilters) ([]MilkDropPresetInfo, error) {
	presets, err := c.SearchPresets(ctx, "", filters)
	if err != nil {
		return nil, err
	}

	if len(presets) <= count {
		return presets, nil
	}

	// Simple shuffle using time-based seed
	seed := time.Now().UnixNano()
	shuffled := make([]MilkDropPresetInfo, len(presets))
	copy(shuffled, presets)

	for i := len(shuffled) - 1; i > 0; i-- {
		j := int(seed) % (i + 1)
		if j < 0 {
			j = -j
		}
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		seed = seed*1103515245 + 12345
	}

	return shuffled[:count], nil
}

// ExportCatalog exports the catalog to JSON
func (c *MilkDropPresetsClient) ExportCatalog(ctx context.Context, outputPath string) error {
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

// GetHealth returns MilkDrop presets system health
func (c *MilkDropPresetsClient) GetHealth(ctx context.Context) (*MilkDropPresetsHealth, error) {
	health := &MilkDropPresetsHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check local presets
	presets, err := c.ScanLocalPresets(ctx)
	if err != nil {
		health.Score -= 30
		health.Issues = append(health.Issues, "Failed to scan local presets")
	} else {
		health.LocalCount = len(presets)
		for _, p := range presets {
			health.LocalSize += p.Size
			if p.UsesShader {
				health.ShaderCount++
			}
		}
	}

	// Check NestDrop directory
	if _, err := os.Stat(c.nestDropPath); os.IsNotExist(err) {
		health.Score -= 20
		health.Recommendations = append(health.Recommendations,
			fmt.Sprintf("Create NestDrop directory: %s", c.nestDropPath))
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

// MilkDropPresetsHealth represents MilkDrop presets system health
type MilkDropPresetsHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	LocalCount      int      `json:"local_count"`
	LocalSize       int64    `json:"local_size"`
	ShaderCount     int      `json:"shader_count"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}
