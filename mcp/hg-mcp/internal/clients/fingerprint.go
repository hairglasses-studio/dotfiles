// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	httpclient "github.com/hairglasses-studio/mcpkit/client"
)

// FingerprintClient provides audio fingerprinting using Chromaprint and AcoustID
type FingerprintClient struct {
	acoustIDKey string
	fpcalcPath  string
	httpClient  *http.Client
}

// Fingerprint represents an audio fingerprint
type Fingerprint struct {
	Duration    float64 `json:"duration"`
	Fingerprint string  `json:"fingerprint"`
	FilePath    string  `json:"file_path,omitempty"`
}

// AcoustIDMatch represents a match from AcoustID
type AcoustIDMatch struct {
	ID         string              `json:"id"`
	Score      float64             `json:"score"`
	Recordings []AcoustIDRecording `json:"recordings,omitempty"`
}

// AcoustIDRecording represents a recording from AcoustID
type AcoustIDRecording struct {
	ID       string            `json:"id"`
	Title    string            `json:"title,omitempty"`
	Duration int               `json:"duration,omitempty"`
	Artists  []AcoustIDArtist  `json:"artists,omitempty"`
	Releases []AcoustIDRelease `json:"releasegroups,omitempty"`
}

// AcoustIDArtist represents an artist
type AcoustIDArtist struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AcoustIDRelease represents a release/album
type AcoustIDRelease struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type,omitempty"`
}

// MatchResult represents the result of a fingerprint match
type MatchResult struct {
	FilePath    string          `json:"file_path"`
	Duration    float64         `json:"duration"`
	Fingerprint string          `json:"fingerprint,omitempty"`
	Matches     []AcoustIDMatch `json:"matches"`
	BestMatch   *BestMatch      `json:"best_match,omitempty"`
}

// BestMatch represents the best matching track
type BestMatch struct {
	Score    float64 `json:"score"`
	Title    string  `json:"title"`
	Artist   string  `json:"artist"`
	Album    string  `json:"album,omitempty"`
	Duration int     `json:"duration,omitempty"`
	MBID     string  `json:"musicbrainz_id,omitempty"`
}

// DuplicateResult represents potential duplicates
type DuplicateResult struct {
	Original   string   `json:"original"`
	Duplicates []string `json:"duplicates"`
	Score      float64  `json:"score"`
}

// FingerprintHealth represents health status
type FingerprintHealth struct {
	Score           int      `json:"score"`
	Status          string   `json:"status"`
	FpcalcInstalled bool     `json:"fpcalc_installed"`
	AcoustIDKey     bool     `json:"acoustid_key_set"`
	Issues          []string `json:"issues,omitempty"`
	Recommendations []string `json:"recommendations,omitempty"`
}

// NewFingerprintClient creates a new fingerprint client
func NewFingerprintClient() (*FingerprintClient, error) {
	acoustIDKey := os.Getenv("ACOUSTID_API_KEY")

	// Find fpcalc binary
	fpcalcPath := os.Getenv("FPCALC_PATH")
	if fpcalcPath == "" {
		// Try to find in PATH
		if path, err := exec.LookPath("fpcalc"); err == nil {
			fpcalcPath = path
		}
	}

	return &FingerprintClient{
		acoustIDKey: acoustIDKey,
		fpcalcPath:  fpcalcPath,
		httpClient: httpclient.Standard(),
	}, nil
}

// GenerateFingerprint generates a fingerprint for an audio file
func (c *FingerprintClient) GenerateFingerprint(ctx context.Context, filePath string) (*Fingerprint, error) {
	if c.fpcalcPath == "" {
		return nil, fmt.Errorf("fpcalc not found - install chromaprint")
	}

	// Run fpcalc
	cmd := exec.CommandContext(ctx, c.fpcalcPath, "-json", filePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("fpcalc failed: %w", err)
	}

	var result struct {
		Duration    float64 `json:"duration"`
		Fingerprint string  `json:"fingerprint"`
	}
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("failed to parse fpcalc output: %w", err)
	}

	return &Fingerprint{
		Duration:    result.Duration,
		Fingerprint: result.Fingerprint,
		FilePath:    filePath,
	}, nil
}

// MatchFingerprint looks up a fingerprint in AcoustID
func (c *FingerprintClient) MatchFingerprint(ctx context.Context, fp *Fingerprint) (*MatchResult, error) {
	if c.acoustIDKey == "" {
		return nil, fmt.Errorf("ACOUSTID_API_KEY not set")
	}

	// Build request
	params := url.Values{}
	params.Set("client", c.acoustIDKey)
	params.Set("duration", fmt.Sprintf("%.0f", fp.Duration))
	params.Set("fingerprint", fp.Fingerprint)
	params.Set("meta", "recordings releasegroups")

	reqURL := "https://api.acoustid.org/v2/lookup?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("AcoustID request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var acoustIDResp struct {
		Status string `json:"status"`
		Error  *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		Results []struct {
			ID         string  `json:"id"`
			Score      float64 `json:"score"`
			Recordings []struct {
				ID       string `json:"id"`
				Title    string `json:"title"`
				Duration int    `json:"duration"`
				Artists  []struct {
					ID   string `json:"id"`
					Name string `json:"name"`
				} `json:"artists"`
				ReleaseGroups []struct {
					ID    string `json:"id"`
					Title string `json:"title"`
					Type  string `json:"type"`
				} `json:"releasegroups"`
			} `json:"recordings"`
		} `json:"results"`
	}

	if err := json.Unmarshal(body, &acoustIDResp); err != nil {
		return nil, fmt.Errorf("failed to parse AcoustID response: %w", err)
	}

	if acoustIDResp.Status == "error" && acoustIDResp.Error != nil {
		return nil, fmt.Errorf("AcoustID error: %s", acoustIDResp.Error.Message)
	}

	result := &MatchResult{
		FilePath:    fp.FilePath,
		Duration:    fp.Duration,
		Fingerprint: fp.Fingerprint,
		Matches:     make([]AcoustIDMatch, 0),
	}

	// Convert results
	for _, r := range acoustIDResp.Results {
		match := AcoustIDMatch{
			ID:         r.ID,
			Score:      r.Score,
			Recordings: make([]AcoustIDRecording, 0),
		}

		for _, rec := range r.Recordings {
			recording := AcoustIDRecording{
				ID:       rec.ID,
				Title:    rec.Title,
				Duration: rec.Duration,
				Artists:  make([]AcoustIDArtist, 0),
				Releases: make([]AcoustIDRelease, 0),
			}

			for _, a := range rec.Artists {
				recording.Artists = append(recording.Artists, AcoustIDArtist{
					ID:   a.ID,
					Name: a.Name,
				})
			}

			for _, rg := range rec.ReleaseGroups {
				recording.Releases = append(recording.Releases, AcoustIDRelease{
					ID:    rg.ID,
					Title: rg.Title,
					Type:  rg.Type,
				})
			}

			match.Recordings = append(match.Recordings, recording)
		}

		result.Matches = append(result.Matches, match)
	}

	// Find best match
	if len(result.Matches) > 0 && len(result.Matches[0].Recordings) > 0 {
		topMatch := result.Matches[0]
		topRec := topMatch.Recordings[0]

		artistNames := make([]string, 0)
		for _, a := range topRec.Artists {
			artistNames = append(artistNames, a.Name)
		}

		albumName := ""
		if len(topRec.Releases) > 0 {
			albumName = topRec.Releases[0].Title
		}

		result.BestMatch = &BestMatch{
			Score:    topMatch.Score,
			Title:    topRec.Title,
			Artist:   strings.Join(artistNames, ", "),
			Album:    albumName,
			Duration: topRec.Duration,
			MBID:     topRec.ID,
		}
	}

	return result, nil
}

// IdentifyTrack generates fingerprint and looks up in AcoustID
func (c *FingerprintClient) IdentifyTrack(ctx context.Context, filePath string) (*MatchResult, error) {
	fp, err := c.GenerateFingerprint(ctx, filePath)
	if err != nil {
		return nil, err
	}

	return c.MatchFingerprint(ctx, fp)
}

// FindDuplicates scans a directory for duplicate tracks using fingerprints
func (c *FingerprintClient) FindDuplicates(ctx context.Context, directory string, threshold float64) ([]DuplicateResult, error) {
	if threshold == 0 {
		threshold = 0.9 // 90% similarity
	}

	// Find audio files
	var audioFiles []string
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".mp3" || ext == ".wav" || ext == ".flac" || ext == ".m4a" || ext == ".aiff" {
			audioFiles = append(audioFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Generate fingerprints for all files
	fingerprints := make(map[string]*Fingerprint)
	for _, file := range audioFiles {
		fp, err := c.GenerateFingerprint(ctx, file)
		if err != nil {
			continue
		}
		fingerprints[file] = fp
	}

	// Compare fingerprints (simplified - real implementation would use acoustic similarity)
	// This is a placeholder - real duplicate detection would use proper fingerprint comparison
	var duplicates []DuplicateResult

	seen := make(map[string]bool)
	for path1, fp1 := range fingerprints {
		if seen[path1] {
			continue
		}

		var dups []string
		for path2, fp2 := range fingerprints {
			if path1 == path2 || seen[path2] {
				continue
			}

			// Simple duration-based matching (real implementation would compare fingerprints)
			if abs(fp1.Duration-fp2.Duration) < 1.0 {
				dups = append(dups, path2)
				seen[path2] = true
			}
		}

		if len(dups) > 0 {
			duplicates = append(duplicates, DuplicateResult{
				Original:   path1,
				Duplicates: dups,
				Score:      threshold,
			})
			seen[path1] = true
		}
	}

	return duplicates, nil
}

// abs returns absolute value of float64
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// BatchGenerate generates fingerprints for multiple files
func (c *FingerprintClient) BatchGenerate(ctx context.Context, files []string) ([]*Fingerprint, error) {
	var results []*Fingerprint

	for _, file := range files {
		fp, err := c.GenerateFingerprint(ctx, file)
		if err != nil {
			// Continue with other files
			continue
		}
		results = append(results, fp)
	}

	return results, nil
}

// GetHealth returns health status
func (c *FingerprintClient) GetHealth(ctx context.Context) (*FingerprintHealth, error) {
	health := &FingerprintHealth{
		Score:  100,
		Status: "healthy",
	}

	// Check fpcalc
	if c.fpcalcPath == "" {
		health.FpcalcInstalled = false
		health.Score -= 50
		health.Issues = append(health.Issues, "fpcalc not found")
		health.Recommendations = append(health.Recommendations,
			"Install chromaprint: brew install chromaprint (macOS) or apt install libchromaprint-tools (Linux)")
	} else {
		health.FpcalcInstalled = true

		// Test fpcalc
		cmd := exec.CommandContext(ctx, c.fpcalcPath, "-version")
		if err := cmd.Run(); err != nil {
			health.Score -= 25
			health.Issues = append(health.Issues, "fpcalc exists but failed to run")
		}
	}

	// Check AcoustID key
	if c.acoustIDKey == "" {
		health.AcoustIDKey = false
		health.Score -= 30
		health.Issues = append(health.Issues, "ACOUSTID_API_KEY not set")
		health.Recommendations = append(health.Recommendations,
			"Get API key from https://acoustid.org/")
	} else {
		health.AcoustIDKey = true

		// Test AcoustID API
		testURL := fmt.Sprintf("https://api.acoustid.org/v2/lookup?client=%s&duration=0&fingerprint=test", c.acoustIDKey)
		req, _ := http.NewRequestWithContext(ctx, "GET", testURL, nil)
		resp, err := c.httpClient.Do(req)
		if err != nil {
			health.Score -= 10
			health.Issues = append(health.Issues, "Cannot reach AcoustID API")
		} else {
			defer resp.Body.Close()
			if resp.StatusCode == 401 || resp.StatusCode == 403 {
				health.Score -= 20
				health.Issues = append(health.Issues, "AcoustID API key is invalid")
			}
		}
	}

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

// FpcalcPath returns the configured fpcalc path
func (c *FingerprintClient) FpcalcPath() string {
	return c.fpcalcPath
}

// submitFingerprint submits a fingerprint to AcoustID (requires submission key)
func (c *FingerprintClient) submitFingerprint(ctx context.Context, fp *Fingerprint, metadata map[string]string) error {
	// This would require a user submission key, not just the API key
	// Left as placeholder for future implementation
	return fmt.Errorf("fingerprint submission not yet implemented")
}

// CompareFingerprints compares two fingerprints and returns similarity score
func (c *FingerprintClient) CompareFingerprints(fp1, fp2 *Fingerprint) float64 {
	// Simplified comparison - real implementation would use proper fingerprint comparison algorithms
	// The chromaprint fingerprint is a compressed representation of audio features

	if abs(fp1.Duration-fp2.Duration) > 5.0 {
		return 0.0 // Different lengths, probably not duplicates
	}

	// Basic string comparison (not accurate for real audio comparison)
	// Real implementation would decode and compare the fingerprints properly
	if fp1.Fingerprint == fp2.Fingerprint {
		return 1.0
	}

	// Return similarity based on duration match as a simple heuristic
	durationDiff := abs(fp1.Duration - fp2.Duration)
	if durationDiff < 1.0 {
		return 0.8
	} else if durationDiff < 2.0 {
		return 0.5
	}

	return 0.0
}
