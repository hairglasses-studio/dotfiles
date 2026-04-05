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

// ROMInfo represents metadata about a ROM file
type ROMInfo struct {
	Name       string    `json:"name"`
	Filename   string    `json:"filename"`
	System     string    `json:"system"` // PS2, PS1, GameCube, etc.
	Region     string    `json:"region"` // USA, EUR, JAP, etc.
	Format     string    `json:"format"` // iso, bin, chd, cue, etc.
	LocalPath  string    `json:"local_path,omitempty"`
	GDrivePath string    `json:"gdrive_path,omitempty"`
	SHA256     string    `json:"sha256,omitempty"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modified_at"`
	// Game metadata
	Title       string `json:"title,omitempty"`
	SerialID    string `json:"serial_id,omitempty"` // SLUS-12345
	Developer   string `json:"developer,omitempty"`
	Publisher   string `json:"publisher,omitempty"`
	Genre       string `json:"genre,omitempty"`
	ReleaseYear int    `json:"release_year,omitempty"`
	Players     int    `json:"players,omitempty"`
	// File info
	IsMultiDisc  bool     `json:"is_multi_disc,omitempty"`
	DiscNumber   int      `json:"disc_number,omitempty"`
	TotalDiscs   int      `json:"total_discs,omitempty"`
	IsCompressed bool     `json:"is_compressed,omitempty"`
	HasCheat     bool     `json:"has_cheat,omitempty"` // Has .pnach file
	CheatPath    string   `json:"cheat_path,omitempty"`
	Tags         []string `json:"tags,omitempty"`
}

// ROMCollection represents ROMs for a system
type ROMCollection struct {
	System    string         `json:"system"`
	ROMCount  int            `json:"rom_count"`
	TotalSize int64          `json:"total_size"`
	Regions   map[string]int `json:"regions"`
	Formats   map[string]int `json:"formats"`
	Genres    map[string]int `json:"genres"`
	ROMs      []ROMInfo      `json:"roms,omitempty"`
	LocalPath string         `json:"local_path,omitempty"`
}

// ROMCatalog represents the full ROM catalog
type ROMCatalog struct {
	Version     int             `json:"version"`
	LastUpdated time.Time       `json:"last_updated"`
	Collections []ROMCollection `json:"collections"`
	TotalROMs   int             `json:"total_roms"`
	TotalSize   int64           `json:"total_size"`
	BySystem    map[string]int  `json:"by_system"`
	ByRegion    map[string]int  `json:"by_region"`
	ByFormat    map[string]int  `json:"by_format"`
}

// ROMCollectionClient provides ROM collection management
type ROMCollectionClient struct {
	homeDir    string
	romsPath   string
	cheatsPath string
	gdrivePath string
	mu         sync.RWMutex
}

var (
	romCollectionClient     *ROMCollectionClient
	romCollectionClientOnce sync.Once
	romCollectionClientErr  error
)

// GetROMCollectionClient returns the singleton ROM collection client
func GetROMCollectionClient() (*ROMCollectionClient, error) {
	romCollectionClientOnce.Do(func() {
		romCollectionClient, romCollectionClientErr = NewROMCollectionClient()
	})
	return romCollectionClient, romCollectionClientErr
}

// NewROMCollectionClient creates a new ROM collection client
func NewROMCollectionClient() (*ROMCollectionClient, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	// Default ROMs path (can be overridden)
	romsPath := os.Getenv("ROMS_PATH")
	if romsPath == "" {
		// Try common locations
		possiblePaths := []string{
			filepath.Join(homeDir, "Documents", "ROMs"),
			filepath.Join(homeDir, "Documents", "All Sony ROMs"),
			"D:/ROMs",
		}
		for _, p := range possiblePaths {
			if _, err := os.Stat(p); err == nil {
				romsPath = p
				break
			}
		}
		if romsPath == "" {
			romsPath = filepath.Join(homeDir, "Documents", "ROMs")
		}
	}

	// Cheats path (PCSX2 .pnach files)
	cheatsPath := os.Getenv("PCSX2_CHEATS_PATH")
	if cheatsPath == "" {
		cheatsPath = filepath.Join(homeDir, "Documents", "PCSX2", "cheats")
	}

	// Google Drive rclone remote path
	gdrivePath := os.Getenv("ROMS_GDRIVE_PATH")
	if gdrivePath == "" {
		gdrivePath = "gdrive:/Gaming & Emulation/ROMs"
	}

	return &ROMCollectionClient{
		homeDir:    homeDir,
		romsPath:   romsPath,
		cheatsPath: cheatsPath,
		gdrivePath: gdrivePath,
	}, nil
}

// GetROMsPath returns the local ROMs path
func (c *ROMCollectionClient) GetROMsPath() string {
	return c.romsPath
}

// ScanLocalROMs scans local ROM directories
func (c *ROMCollectionClient) ScanLocalROMs(ctx context.Context) ([]ROMInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	roms := []ROMInfo{}

	if _, err := os.Stat(c.romsPath); os.IsNotExist(err) {
		return roms, nil
	}

	// Patterns for extracting info from filenames
	serialRegex := regexp.MustCompile(`([A-Z]{4}[-_]?\d{5})`)
	regionRegex := regexp.MustCompile(`\((USA|EUR|JAP|USA,\s*EUR|PAL|NTSC|World|Japan|Europe|Korea)\)`)
	discRegex := regexp.MustCompile(`(?i)(?:disc|disk|cd)\s*(\d+)`)

	err := filepath.Walk(c.romsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !isROMFormat(ext) {
			return nil
		}

		// Extract system from directory structure
		relPath, _ := filepath.Rel(c.romsPath, path)
		parts := strings.Split(relPath, string(filepath.Separator))

		system := detectSystem(parts, ext)
		baseName := strings.TrimSuffix(info.Name(), ext)

		// Parse region
		region := "Unknown"
		if matches := regionRegex.FindStringSubmatch(baseName); len(matches) >= 2 {
			region = normalizeRegion(matches[1])
		}

		// Parse serial ID
		serialID := ""
		if matches := serialRegex.FindStringSubmatch(baseName); len(matches) >= 2 {
			serialID = matches[1]
		}

		// Parse disc number
		isMultiDisc := false
		discNumber := 1
		if matches := discRegex.FindStringSubmatch(baseName); len(matches) >= 2 {
			isMultiDisc = true
			fmt.Sscanf(matches[1], "%d", &discNumber)
		}

		// Clean up title
		title := cleanROMTitle(baseName)

		rom := ROMInfo{
			Name:         baseName,
			Filename:     info.Name(),
			System:       system,
			Region:       region,
			Format:       strings.TrimPrefix(ext, "."),
			LocalPath:    path,
			Size:         info.Size(),
			ModifiedAt:   info.ModTime(),
			Title:        title,
			SerialID:     serialID,
			IsMultiDisc:  isMultiDisc,
			DiscNumber:   discNumber,
			IsCompressed: isCompressedFormat(ext),
		}

		// Check for corresponding cheat file
		if serialID != "" && c.cheatsPath != "" {
			cheatPath := filepath.Join(c.cheatsPath, serialID+".pnach")
			if _, err := os.Stat(cheatPath); err == nil {
				rom.HasCheat = true
				rom.CheatPath = cheatPath
			}
		}

		roms = append(roms, rom)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return roms, nil
}

func isROMFormat(ext string) bool {
	romExts := map[string]bool{
		// Disc images
		".iso": true, ".bin": true, ".cue": true, ".img": true,
		".mdf": true, ".mds": true, ".nrg": true,
		// Compressed
		".chd": true, ".cso": true, ".zso": true, ".gz": true,
		".rvz": true, ".wbfs": true,
		// Cartridge ROMs
		".z64": true, ".n64": true, ".v64": true, // N64
		".sfc": true, ".smc": true, // SNES
		".nes": true, ".fds": true, // NES
		".gba": true, ".gbc": true, ".gb": true, // Game Boy
		".nds": true, ".3ds": true, // DS/3DS
		".xci": true, ".nsp": true, // Switch
		".gcm": true, ".gcz": true, // GameCube
		".wad": true, // Wii
		// Other
		".pbp": true, ".elf": true,
	}
	return romExts[ext]
}

func isCompressedFormat(ext string) bool {
	compressed := map[string]bool{
		".chd": true, ".cso": true, ".zso": true, ".gz": true,
		".rvz": true, ".gcz": true, ".7z": true, ".zip": true,
	}
	return compressed[ext]
}

func detectSystem(pathParts []string, ext string) string {
	// Check path for system hints
	pathLower := strings.ToLower(strings.Join(pathParts, "/"))

	systems := map[string][]string{
		"PS2":       {"ps2", "playstation 2", "playstation2"},
		"PS1":       {"ps1", "psx", "playstation 1", "playstation1"},
		"PS3":       {"ps3", "playstation 3"},
		"PSP":       {"psp"},
		"GameCube":  {"gamecube", "gc", "ngc"},
		"Wii":       {"wii"},
		"WiiU":      {"wiiu", "wii u"},
		"Switch":    {"switch", "nx"},
		"N64":       {"n64", "nintendo 64"},
		"SNES":      {"snes", "super nintendo", "super nes"},
		"NES":       {"nes", "nintendo entertainment"},
		"GBA":       {"gba", "game boy advance", "gameboy advance"},
		"DS":        {"nds", "nintendo ds"},
		"3DS":       {"3ds", "nintendo 3ds"},
		"Xbox":      {"xbox"},
		"Xbox360":   {"xbox 360", "xbox360"},
		"Dreamcast": {"dreamcast", "dc"},
		"Saturn":    {"saturn"},
		"Genesis":   {"genesis", "mega drive", "megadrive"},
	}

	for system, keywords := range systems {
		for _, kw := range keywords {
			if strings.Contains(pathLower, kw) {
				return system
			}
		}
	}

	// Detect by extension
	extSystems := map[string]string{
		".z64": "N64", ".n64": "N64", ".v64": "N64",
		".sfc": "SNES", ".smc": "SNES",
		".nes": "NES", ".fds": "NES",
		".gba": "GBA", ".gbc": "GBC", ".gb": "GB",
		".nds": "DS", ".3ds": "3DS",
		".xci": "Switch", ".nsp": "Switch",
		".gcm": "GameCube", ".gcz": "GameCube",
		".wbfs": "Wii", ".rvz": "Wii",
		".pbp": "PSP",
	}

	if system, ok := extSystems[ext]; ok {
		return system
	}

	return "Unknown"
}

func normalizeRegion(region string) string {
	region = strings.ToUpper(region)
	switch {
	case strings.Contains(region, "USA") || region == "NTSC":
		return "USA"
	case strings.Contains(region, "EUR") || region == "PAL" || region == "EUROPE":
		return "EUR"
	case strings.Contains(region, "JAP") || region == "JAPAN":
		return "JAP"
	case region == "WORLD":
		return "World"
	case region == "KOREA":
		return "KOR"
	default:
		return region
	}
}

func cleanROMTitle(name string) string {
	// Remove common suffixes and tags
	patterns := []string{
		`\s*\([^)]*\)`,      // (USA), (v1.0), etc.
		`\s*\[[^\]]*\]`,     // [!], [SLUS-12345], etc.
		`\s*-\s*Disc\s*\d+`, // - Disc 1
		`\s*v\d+\.\d+`,      // v1.0
	}

	result := name
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		result = re.ReplaceAllString(result, "")
	}

	return strings.TrimSpace(result)
}

// ListCollections lists all ROM collections by system
func (c *ROMCollectionClient) ListCollections(ctx context.Context) ([]ROMCollection, error) {
	roms, err := c.ScanLocalROMs(ctx)
	if err != nil {
		return nil, err
	}

	// Group by system
	collectionMap := make(map[string]*ROMCollection)
	for _, rom := range roms {
		if col, ok := collectionMap[rom.System]; ok {
			col.ROMCount++
			col.TotalSize += rom.Size
			col.Regions[rom.Region]++
			col.Formats[rom.Format]++
			if rom.Genre != "" {
				col.Genres[rom.Genre]++
			}
		} else {
			collectionMap[rom.System] = &ROMCollection{
				System:    rom.System,
				ROMCount:  1,
				TotalSize: rom.Size,
				Regions:   map[string]int{rom.Region: 1},
				Formats:   map[string]int{rom.Format: 1},
				Genres:    make(map[string]int),
			}
			if rom.Genre != "" {
				collectionMap[rom.System].Genres[rom.Genre] = 1
			}
		}
	}

	collections := make([]ROMCollection, 0, len(collectionMap))
	for _, col := range collectionMap {
		collections = append(collections, *col)
	}

	return collections, nil
}

// SearchROMs searches ROMs by query
func (c *ROMCollectionClient) SearchROMs(ctx context.Context, query string, filters ROMSearchFilters) ([]ROMInfo, error) {
	roms, err := c.ScanLocalROMs(ctx)
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []ROMInfo

	for _, rom := range roms {
		// System filter
		if filters.System != "" && !strings.EqualFold(rom.System, filters.System) {
			continue
		}

		// Region filter
		if filters.Region != "" && !strings.EqualFold(rom.Region, filters.Region) {
			continue
		}

		// Format filter
		if filters.Format != "" && !strings.EqualFold(rom.Format, filters.Format) {
			continue
		}

		// Has cheat filter
		if filters.HasCheat && !rom.HasCheat {
			continue
		}

		// Query filter
		if query != "" {
			if !strings.Contains(strings.ToLower(rom.Title), query) &&
				!strings.Contains(strings.ToLower(rom.Name), query) &&
				!strings.Contains(strings.ToLower(rom.SerialID), query) {
				continue
			}
		}

		results = append(results, rom)
		if filters.Limit > 0 && len(results) >= filters.Limit {
			break
		}
	}

	return results, nil
}

// ROMSearchFilters contains search filter options
type ROMSearchFilters struct {
	System   string `json:"system,omitempty"`
	Region   string `json:"region,omitempty"`
	Format   string `json:"format,omitempty"`
	HasCheat bool   `json:"has_cheat,omitempty"`
	Limit    int    `json:"limit,omitempty"`
}

// GetCatalog returns the full ROM catalog
func (c *ROMCollectionClient) GetCatalog(ctx context.Context) (*ROMCatalog, error) {
	roms, err := c.ScanLocalROMs(ctx)
	if err != nil {
		return nil, err
	}

	collections, err := c.ListCollections(ctx)
	if err != nil {
		return nil, err
	}

	catalog := &ROMCatalog{
		Version:     1,
		LastUpdated: time.Now(),
		Collections: collections,
		TotalROMs:   len(roms),
		BySystem:    make(map[string]int),
		ByRegion:    make(map[string]int),
		ByFormat:    make(map[string]int),
	}

	for _, rom := range roms {
		catalog.TotalSize += rom.Size
		catalog.BySystem[rom.System]++
		catalog.ByRegion[rom.Region]++
		catalog.ByFormat[rom.Format]++
	}

	return catalog, nil
}

// GetROMsBySystem returns all ROMs for a system
func (c *ROMCollectionClient) GetROMsBySystem(ctx context.Context, system string) ([]ROMInfo, error) {
	roms, err := c.ScanLocalROMs(ctx)
	if err != nil {
		return nil, err
	}

	var systemROMs []ROMInfo
	for _, rom := range roms {
		if strings.EqualFold(rom.System, system) {
			systemROMs = append(systemROMs, rom)
		}
	}

	return systemROMs, nil
}

// GetROMBySerial finds a ROM by its serial ID
func (c *ROMCollectionClient) GetROMBySerial(ctx context.Context, serialID string) (*ROMInfo, error) {
	roms, err := c.ScanLocalROMs(ctx)
	if err != nil {
		return nil, err
	}

	serialID = strings.ToUpper(serialID)
	for _, rom := range roms {
		if strings.EqualFold(rom.SerialID, serialID) {
			return &rom, nil
		}
	}

	return nil, fmt.Errorf("ROM not found with serial: %s", serialID)
}

// CalculateROMHash calculates SHA256 hash for a ROM
func (c *ROMCollectionClient) CalculateROMHash(ctx context.Context, romPath string) (string, error) {
	f, err := os.Open(romPath)
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

// ExportCatalog exports the catalog to JSON
func (c *ROMCollectionClient) ExportCatalog(ctx context.Context, outputPath string) error {
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

// GetHealth returns ROM collection system health
func (c *ROMCollectionClient) GetHealth(ctx context.Context) (*ROMCollectionHealth, error) {
	health := &ROMCollectionHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check local ROMs
	roms, err := c.ScanLocalROMs(ctx)
	if err != nil {
		health.Score -= 30
		health.Issues = append(health.Issues, "Failed to scan local ROMs")
	} else {
		health.LocalCount = len(roms)
		for _, r := range roms {
			health.LocalSize += r.Size
		}

		// Count ROMs with cheats
		withCheats := 0
		for _, r := range roms {
			if r.HasCheat {
				withCheats++
			}
		}
		health.WithCheats = withCheats
	}

	// Check ROMs directory
	if _, err := os.Stat(c.romsPath); os.IsNotExist(err) {
		health.Score -= 20
		health.Recommendations = append(health.Recommendations,
			fmt.Sprintf("Create ROMs directory: %s", c.romsPath))
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

// ROMCollectionHealth represents ROM collection system health
type ROMCollectionHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	LocalCount      int      `json:"local_count"`
	LocalSize       int64    `json:"local_size"`
	WithCheats      int      `json:"with_cheats"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}
