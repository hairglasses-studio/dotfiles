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
	"strconv"
	"strings"
	"sync"
	"time"
)

// DJTrackInfo represents metadata about a DJ track
type DJTrackInfo struct {
	Title      string    `json:"title"`
	Artist     string    `json:"artist"`
	Filename   string    `json:"filename"`
	Crate      string    `json:"crate"` // Collection/crate name
	Genre      string    `json:"genre,omitempty"`
	Format     string    `json:"format"` // mp3, wav, flac, aiff
	BPM        float64   `json:"bpm,omitempty"`
	Key        string    `json:"key,omitempty"`      // Musical key (1A-12B Camelot or standard)
	Energy     int       `json:"energy,omitempty"`   // 1-10 energy level
	Duration   float64   `json:"duration,omitempty"` // seconds
	LocalPath  string    `json:"local_path,omitempty"`
	GDrivePath string    `json:"gdrive_path,omitempty"`
	SHA256     string    `json:"sha256,omitempty"`
	Size       int64     `json:"size"`
	Bitrate    int       `json:"bitrate,omitempty"` // kbps
	SampleRate int       `json:"sample_rate,omitempty"`
	ModifiedAt time.Time `json:"modified_at"`
	// rekordbox fields
	RekordboxID string     `json:"rekordbox_id,omitempty"`
	PlayCount   int        `json:"play_count,omitempty"`
	Rating      int        `json:"rating,omitempty"` // 0-5 stars
	Comment     string     `json:"comment,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	CuePoints   []CuePoint `json:"cue_points,omitempty"`
	LastPlayed  *time.Time `json:"last_played,omitempty"`
}

// CuePoint represents a cue/hot cue point
type CuePoint struct {
	Name     string  `json:"name,omitempty"`
	Position float64 `json:"position"` // milliseconds
	Type     string  `json:"type"`     // cue, loop, memory
	Color    string  `json:"color,omitempty"`
}

// DJCrate represents a crate/playlist
type DJCrate struct {
	Name       string         `json:"name"`
	TrackCount int            `json:"track_count"`
	TotalSize  int64          `json:"total_size"`
	Duration   float64        `json:"total_duration"` // seconds
	Genres     map[string]int `json:"genres"`
	BPMRange   [2]float64     `json:"bpm_range"` // [min, max]
	Tracks     []DJTrackInfo  `json:"tracks,omitempty"`
	ModifiedAt time.Time      `json:"modified_at"`
	LocalPath  string         `json:"local_path,omitempty"`
}

// DJCatalog represents the full DJ music catalog
type DJCatalog struct {
	Version         int            `json:"version"`
	LastUpdated     time.Time      `json:"last_updated"`
	Crates          []DJCrate      `json:"crates"`
	TotalTracks     int            `json:"total_tracks"`
	TotalSize       int64          `json:"total_size"`
	ByGenre         map[string]int `json:"by_genre"`
	ByFormat        map[string]int `json:"by_format"`
	ByKey           map[string]int `json:"by_key"`
	BPMDistribution map[string]int `json:"bpm_distribution"` // "120-130": count
}

// DJCratesClient provides DJ crate management
type DJCratesClient struct {
	homeDir       string
	djCratesPath  string
	rekordboxPath string
	gdrivePath    string
	mu            sync.RWMutex
}

var (
	djCratesClient     *DJCratesClient
	djCratesClientOnce sync.Once
	djCratesClientErr  error
)

// GetDJCratesClient returns the singleton DJ crates client
func GetDJCratesClient() (*DJCratesClient, error) {
	djCratesClientOnce.Do(func() {
		djCratesClient, djCratesClientErr = NewDJCratesClient()
	})
	return djCratesClient, djCratesClientErr
}

// NewDJCratesClient creates a new DJ crates client
func NewDJCratesClient() (*DJCratesClient, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	// Default DJ Crates path (can be overridden)
	djCratesPath := os.Getenv("DJ_CRATES_PATH")
	if djCratesPath == "" {
		// Try common locations
		possiblePaths := []string{
			filepath.Join(homeDir, "Music", "DJ Crates"),
			filepath.Join(homeDir, "Documents", "DJ Crates"),
			"D:/Music/DJ Crates",
		}
		for _, p := range possiblePaths {
			if _, err := os.Stat(p); err == nil {
				djCratesPath = p
				break
			}
		}
		if djCratesPath == "" {
			djCratesPath = filepath.Join(homeDir, "Music", "DJ Crates")
		}
	}

	// rekordbox library path
	rekordboxPath := os.Getenv("REKORDBOX_PATH")
	if rekordboxPath == "" {
		rekordboxPath = filepath.Join(homeDir, "Documents", "rekordbox")
	}

	// Google Drive rclone remote path
	gdrivePath := os.Getenv("DJ_CRATES_GDRIVE_PATH")
	if gdrivePath == "" {
		gdrivePath = "gdrive:/Music Production/DJ Crates"
	}

	return &DJCratesClient{
		homeDir:       homeDir,
		djCratesPath:  djCratesPath,
		rekordboxPath: rekordboxPath,
		gdrivePath:    gdrivePath,
	}, nil
}

// GetDJCratesPath returns the local DJ crates path
func (c *DJCratesClient) GetDJCratesPath() string {
	return c.djCratesPath
}

// ScanLocalTracks scans local DJ track directories
func (c *DJCratesClient) ScanLocalTracks(ctx context.Context) ([]DJTrackInfo, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	tracks := []DJTrackInfo{}

	if _, err := os.Stat(c.djCratesPath); os.IsNotExist(err) {
		return tracks, nil
	}

	// Patterns for extracting info from filenames
	// Common: "Artist - Title.mp3" or "Artist - Title (Key BPM).mp3"
	artistTitleRegex := regexp.MustCompile(`^(.+?)\s*[-–]\s*(.+?)(?:\s*\(.*\))?$`)
	bpmRegex := regexp.MustCompile(`(\d{2,3})\s*(?:bpm)?`)
	keyRegex := regexp.MustCompile(`([0-9]{1,2}[AB]|[A-G][#b]?(?:m|maj|min)?)`)

	err := filepath.Walk(c.djCratesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !isDJAudioFormat(ext) {
			return nil
		}

		// Extract crate from directory structure
		relPath, _ := filepath.Rel(c.djCratesPath, path)
		parts := strings.Split(relPath, string(filepath.Separator))

		crate := "Unsorted"
		if len(parts) > 1 {
			crate = parts[0]
		}

		baseName := strings.TrimSuffix(info.Name(), ext)

		// Try to parse artist and title
		artist := ""
		title := baseName
		if matches := artistTitleRegex.FindStringSubmatch(baseName); len(matches) >= 3 {
			artist = strings.TrimSpace(matches[1])
			title = strings.TrimSpace(matches[2])
		}

		// Try to extract BPM and key from filename
		var bpm float64
		if matches := bpmRegex.FindStringSubmatch(baseName); len(matches) >= 2 {
			if b, err := strconv.ParseFloat(matches[1], 64); err == nil && b >= 60 && b <= 200 {
				bpm = b
			}
		}

		key := ""
		if matches := keyRegex.FindStringSubmatch(baseName); len(matches) >= 2 {
			key = normalizeMusicalKey(matches[1])
		}

		track := DJTrackInfo{
			Title:      title,
			Artist:     artist,
			Filename:   info.Name(),
			Crate:      crate,
			Format:     strings.TrimPrefix(ext, "."),
			BPM:        bpm,
			Key:        key,
			LocalPath:  path,
			Size:       info.Size(),
			ModifiedAt: info.ModTime(),
		}

		// Calculate hash for smaller files (< 100MB)
		if info.Size() < 100*1024*1024 {
			if hash, err := calculateTrackHash(path); err == nil {
				track.SHA256 = hash
			}
		}

		tracks = append(tracks, track)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return tracks, nil
}

func isDJAudioFormat(ext string) bool {
	audioExts := map[string]bool{
		".mp3": true, ".wav": true, ".flac": true, ".aiff": true,
		".aif": true, ".m4a": true, ".ogg": true, ".opus": true,
		".alac": true,
	}
	return audioExts[ext]
}

func normalizeMusicalKey(key string) string {
	key = strings.ToUpper(key)
	// Already in Camelot notation
	if len(key) == 2 || len(key) == 3 {
		if key[len(key)-1] == 'A' || key[len(key)-1] == 'B' {
			return key
		}
	}
	// Convert standard notation to Camelot
	camelotMap := map[string]string{
		"C": "8B", "CM": "5A", "CMAJ": "8B", "CMIN": "5A",
		"C#": "3B", "C#M": "12A", "DB": "3B", "DBM": "12A",
		"D": "10B", "DM": "7A", "DMAJ": "10B", "DMIN": "7A",
		"D#": "5B", "D#M": "2A", "EB": "5B", "EBM": "2A",
		"E": "12B", "EM": "9A", "EMAJ": "12B", "EMIN": "9A",
		"F": "7B", "FM": "4A", "FMAJ": "7B", "FMIN": "4A",
		"F#": "2B", "F#M": "11A", "GB": "2B", "GBM": "11A",
		"G": "9B", "GM": "6A", "GMAJ": "9B", "GMIN": "6A",
		"G#": "4B", "G#M": "1A", "AB": "4B", "ABM": "1A",
		"A": "11B", "AM": "8A", "AMAJ": "11B", "AMIN": "8A",
		"A#": "6B", "A#M": "3A", "BB": "6B", "BBM": "3A",
		"B": "1B", "BM": "10A", "BMAJ": "1B", "BMIN": "10A",
	}
	if camelot, ok := camelotMap[key]; ok {
		return camelot
	}
	return key
}

func calculateTrackHash(path string) (string, error) {
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

// ListCrates lists all crates
func (c *DJCratesClient) ListCrates(ctx context.Context) ([]DJCrate, error) {
	tracks, err := c.ScanLocalTracks(ctx)
	if err != nil {
		return nil, err
	}

	// Group by crate
	crateMap := make(map[string]*DJCrate)
	for _, track := range tracks {
		if cr, ok := crateMap[track.Crate]; ok {
			cr.TrackCount++
			cr.TotalSize += track.Size
			cr.Duration += track.Duration
			if track.Genre != "" {
				cr.Genres[track.Genre]++
			}
			// Update BPM range
			if track.BPM > 0 {
				if cr.BPMRange[0] == 0 || track.BPM < cr.BPMRange[0] {
					cr.BPMRange[0] = track.BPM
				}
				if track.BPM > cr.BPMRange[1] {
					cr.BPMRange[1] = track.BPM
				}
			}
			if track.ModifiedAt.After(cr.ModifiedAt) {
				cr.ModifiedAt = track.ModifiedAt
			}
		} else {
			crateMap[track.Crate] = &DJCrate{
				Name:       track.Crate,
				TrackCount: 1,
				TotalSize:  track.Size,
				Duration:   track.Duration,
				Genres:     make(map[string]int),
				BPMRange:   [2]float64{track.BPM, track.BPM},
				ModifiedAt: track.ModifiedAt,
			}
			if track.Genre != "" {
				crateMap[track.Crate].Genres[track.Genre] = 1
			}
		}
	}

	crates := make([]DJCrate, 0, len(crateMap))
	for _, cr := range crateMap {
		crates = append(crates, *cr)
	}

	return crates, nil
}

// SearchTracks searches tracks by query
func (c *DJCratesClient) SearchTracks(ctx context.Context, query string, filters TrackSearchFilters) ([]DJTrackInfo, error) {
	tracks, err := c.ScanLocalTracks(ctx)
	if err != nil {
		return nil, err
	}

	query = strings.ToLower(query)
	var results []DJTrackInfo

	for _, track := range tracks {
		// Crate filter
		if filters.Crate != "" && !strings.EqualFold(track.Crate, filters.Crate) {
			continue
		}

		// Genre filter
		if filters.Genre != "" && !strings.EqualFold(track.Genre, filters.Genre) {
			continue
		}

		// BPM range filter
		if filters.MinBPM > 0 && track.BPM < filters.MinBPM {
			continue
		}
		if filters.MaxBPM > 0 && track.BPM > filters.MaxBPM {
			continue
		}

		// Key filter (harmonic mixing)
		if filters.Key != "" && track.Key != "" {
			if !isHarmonicMatch(filters.Key, track.Key) {
				continue
			}
		}

		// Query filter
		if query != "" {
			if !strings.Contains(strings.ToLower(track.Title), query) &&
				!strings.Contains(strings.ToLower(track.Artist), query) &&
				!strings.Contains(strings.ToLower(track.Filename), query) {
				continue
			}
		}

		results = append(results, track)
		if filters.Limit > 0 && len(results) >= filters.Limit {
			break
		}
	}

	return results, nil
}

// TrackSearchFilters contains search filter options
type TrackSearchFilters struct {
	Crate  string  `json:"crate,omitempty"`
	Genre  string  `json:"genre,omitempty"`
	MinBPM float64 `json:"min_bpm,omitempty"`
	MaxBPM float64 `json:"max_bpm,omitempty"`
	Key    string  `json:"key,omitempty"`
	Limit  int     `json:"limit,omitempty"`
}

// isHarmonicMatch checks if two keys are harmonically compatible
func isHarmonicMatch(key1, key2 string) bool {
	if key1 == key2 {
		return true
	}

	// Parse Camelot keys
	parse := func(k string) (int, rune) {
		if len(k) < 2 {
			return 0, 0
		}
		mode := rune(k[len(k)-1])
		numStr := k[:len(k)-1]
		num, _ := strconv.Atoi(numStr)
		return num, mode
	}

	num1, mode1 := parse(key1)
	num2, mode2 := parse(key2)

	if num1 == 0 || num2 == 0 {
		return false
	}

	// Same key
	if num1 == num2 && mode1 == mode2 {
		return true
	}

	// Adjacent on wheel (+/- 1, wrapping 12 to 1)
	diff := (num1 - num2 + 12) % 12
	if (diff == 1 || diff == 11) && mode1 == mode2 {
		return true
	}

	// Mode switch (A/B at same number)
	if num1 == num2 && mode1 != mode2 {
		return true
	}

	return false
}

// GetCatalog returns the full DJ music catalog
func (c *DJCratesClient) GetCatalog(ctx context.Context) (*DJCatalog, error) {
	tracks, err := c.ScanLocalTracks(ctx)
	if err != nil {
		return nil, err
	}

	crates, err := c.ListCrates(ctx)
	if err != nil {
		return nil, err
	}

	catalog := &DJCatalog{
		Version:         1,
		LastUpdated:     time.Now(),
		Crates:          crates,
		TotalTracks:     len(tracks),
		ByGenre:         make(map[string]int),
		ByFormat:        make(map[string]int),
		ByKey:           make(map[string]int),
		BPMDistribution: make(map[string]int),
	}

	for _, track := range tracks {
		catalog.TotalSize += track.Size
		if track.Genre != "" {
			catalog.ByGenre[track.Genre]++
		}
		catalog.ByFormat[track.Format]++
		if track.Key != "" {
			catalog.ByKey[track.Key]++
		}
		// BPM distribution in 10 BPM buckets
		if track.BPM > 0 {
			bucket := int(track.BPM/10) * 10
			bucketKey := fmt.Sprintf("%d-%d", bucket, bucket+10)
			catalog.BPMDistribution[bucketKey]++
		}
	}

	return catalog, nil
}

// GetHarmonicMatches returns tracks that mix harmonically with the given key
func (c *DJCratesClient) GetHarmonicMatches(ctx context.Context, key string, bpmRange [2]float64) ([]DJTrackInfo, error) {
	tracks, err := c.ScanLocalTracks(ctx)
	if err != nil {
		return nil, err
	}

	var matches []DJTrackInfo
	for _, track := range tracks {
		if track.Key == "" {
			continue
		}

		// Check BPM range if specified
		if bpmRange[0] > 0 && track.BPM < bpmRange[0] {
			continue
		}
		if bpmRange[1] > 0 && track.BPM > bpmRange[1] {
			continue
		}

		if isHarmonicMatch(key, track.Key) {
			matches = append(matches, track)
		}
	}

	return matches, nil
}

// GetCrateTracks returns all tracks in a crate
func (c *DJCratesClient) GetCrateTracks(ctx context.Context, crateName string) ([]DJTrackInfo, error) {
	tracks, err := c.ScanLocalTracks(ctx)
	if err != nil {
		return nil, err
	}

	var crateTracks []DJTrackInfo
	for _, track := range tracks {
		if strings.EqualFold(track.Crate, crateName) {
			crateTracks = append(crateTracks, track)
		}
	}

	return crateTracks, nil
}

// ExportCatalog exports the catalog to JSON
func (c *DJCratesClient) ExportCatalog(ctx context.Context, outputPath string) error {
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

// GetHealth returns DJ crates system health
func (c *DJCratesClient) GetHealth(ctx context.Context) (*DJCratesHealth, error) {
	health := &DJCratesHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check local tracks
	tracks, err := c.ScanLocalTracks(ctx)
	if err != nil {
		health.Score -= 30
		health.Issues = append(health.Issues, "Failed to scan local tracks")
	} else {
		health.LocalCount = len(tracks)
		for _, t := range tracks {
			health.LocalSize += t.Size
		}
		// Check metadata coverage
		withBPM := 0
		withKey := 0
		for _, t := range tracks {
			if t.BPM > 0 {
				withBPM++
			}
			if t.Key != "" {
				withKey++
			}
		}
		if len(tracks) > 0 {
			bpmCoverage := float64(withBPM) / float64(len(tracks))
			keyCoverage := float64(withKey) / float64(len(tracks))
			if bpmCoverage < 0.5 {
				health.Recommendations = append(health.Recommendations,
					fmt.Sprintf("Only %.0f%% of tracks have BPM info. Consider analyzing with rekordbox.", bpmCoverage*100))
			}
			if keyCoverage < 0.5 {
				health.Recommendations = append(health.Recommendations,
					fmt.Sprintf("Only %.0f%% of tracks have key info. Consider analyzing with Mixed In Key.", keyCoverage*100))
			}
		}
	}

	// Check DJ Crates directory
	if _, err := os.Stat(c.djCratesPath); os.IsNotExist(err) {
		health.Score -= 20
		health.Recommendations = append(health.Recommendations,
			fmt.Sprintf("Create DJ Crates directory: %s", c.djCratesPath))
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

// DJCratesHealth represents DJ crates system health
type DJCratesHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	LocalCount      int      `json:"local_count"`
	LocalSize       int64    `json:"local_size"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}
