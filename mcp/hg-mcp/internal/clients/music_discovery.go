// Package clients provides API clients for external services.
package clients

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

// MusicDiscoveryClient provides unified search across all music platforms
type MusicDiscoveryClient struct {
	// Platform clients
	beatport   *BeatportClient
	traxsource *TraxsourceClient
	juno       *JunoClient
	boomkat    *BoomkatClient
	spotify    *SpotifyClient
	soundcloud *SoundCloudClient
	bandcamp   *BandcampClient
	mixcloud   *MixcloudClient
	discogs    *DiscogsClient

	// Client initialization
	initOnce sync.Once
	mu       sync.RWMutex
}

// Platform identifiers
const (
	PlatformBeatport   = "beatport"
	PlatformTraxsource = "traxsource"
	PlatformJuno       = "juno"
	PlatformBoomkat    = "boomkat"
	PlatformSpotify    = "spotify"
	PlatformSoundCloud = "soundcloud"
	PlatformBandcamp   = "bandcamp"
	PlatformMixcloud   = "mixcloud"
	PlatformDiscogs    = "discogs"
)

// AllPlatforms lists all supported platforms
var AllPlatforms = []string{
	PlatformBeatport,
	PlatformTraxsource,
	PlatformJuno,
	PlatformBoomkat,
	PlatformSpotify,
	PlatformSoundCloud,
	PlatformBandcamp,
	PlatformMixcloud,
	PlatformDiscogs,
}

// PlatformInfo describes a music platform
type PlatformInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"` // "store", "streaming", "community"
	Features    []string `json:"features"`
	HasBPM      bool     `json:"has_bpm"`
	HasKey      bool     `json:"has_key"`
	HasDownload bool     `json:"has_download"`
	HasPricing  bool     `json:"has_pricing"`
}

// PlatformInfoMap provides metadata about each platform
var PlatformInfoMap = map[string]PlatformInfo{
	PlatformBeatport: {
		ID:          PlatformBeatport,
		Name:        "Beatport",
		Type:        "store",
		Features:    []string{"search", "charts", "genres", "releases", "artists", "labels", "download"},
		HasBPM:      true,
		HasKey:      true,
		HasDownload: true,
		HasPricing:  true,
	},
	PlatformTraxsource: {
		ID:          PlatformTraxsource,
		Name:        "Traxsource",
		Type:        "store",
		Features:    []string{"search", "charts", "genres", "releases", "artists", "labels", "download"},
		HasBPM:      true,
		HasKey:      true,
		HasDownload: true,
		HasPricing:  true,
	},
	PlatformJuno: {
		ID:          PlatformJuno,
		Name:        "Juno Download",
		Type:        "store",
		Features:    []string{"search", "charts", "genres", "releases", "artists", "labels", "download"},
		HasBPM:      true,
		HasKey:      true,
		HasDownload: true,
		HasPricing:  true,
	},
	PlatformBoomkat: {
		ID:          PlatformBoomkat,
		Name:        "Boomkat",
		Type:        "store",
		Features:    []string{"search", "genres", "releases", "artists", "labels", "download"},
		HasBPM:      false,
		HasKey:      false,
		HasDownload: true,
		HasPricing:  true,
	},
	PlatformSpotify: {
		ID:          PlatformSpotify,
		Name:        "Spotify",
		Type:        "streaming",
		Features:    []string{"search", "playlists", "artists", "albums", "audio-features"},
		HasBPM:      true,
		HasKey:      true,
		HasDownload: false,
		HasPricing:  false,
	},
	PlatformSoundCloud: {
		ID:          PlatformSoundCloud,
		Name:        "SoundCloud",
		Type:        "community",
		Features:    []string{"search", "playlists", "artists", "tracks", "download"},
		HasBPM:      false,
		HasKey:      false,
		HasDownload: true,
		HasPricing:  false,
	},
	PlatformBandcamp: {
		ID:          PlatformBandcamp,
		Name:        "Bandcamp",
		Type:        "community",
		Features:    []string{"search", "albums", "artists", "tags", "download"},
		HasBPM:      false,
		HasKey:      false,
		HasDownload: true,
		HasPricing:  true,
	},
	PlatformMixcloud: {
		ID:          PlatformMixcloud,
		Name:        "Mixcloud",
		Type:        "community",
		Features:    []string{"search", "shows", "artists", "tags"},
		HasBPM:      false,
		HasKey:      false,
		HasDownload: false,
		HasPricing:  false,
	},
	PlatformDiscogs: {
		ID:          PlatformDiscogs,
		Name:        "Discogs",
		Type:        "marketplace",
		Features:    []string{"search", "releases", "artists", "labels", "marketplace"},
		HasBPM:      false,
		HasKey:      false,
		HasDownload: false,
		HasPricing:  true,
	},
}

// UnifiedTrack represents a track normalized across all platforms
type UnifiedTrack struct {
	NormalizedID string                `json:"normalized_id"` // SHA256(lowercase(artist+title))
	Title        string                `json:"title"`
	Artists      []string              `json:"artists"`
	Mix          string                `json:"mix,omitempty"`
	BPM          float64               `json:"bpm,omitempty"` // Best available BPM
	Key          string                `json:"key,omitempty"` // Camelot notation preferred
	Genre        string                `json:"genre,omitempty"`
	Duration     int                   `json:"duration_ms,omitempty"` // Duration in milliseconds
	ReleaseDate  string                `json:"release_date,omitempty"`
	Label        string                `json:"label,omitempty"`
	ImageURL     string                `json:"image_url,omitempty"`
	PlatformIDs  map[string]string     `json:"platform_ids"`      // platform -> id
	PlatformURLs map[string]string     `json:"platform_urls"`     // platform -> url
	Prices       map[string]PriceInfo  `json:"prices,omitempty"`  // platform -> price
	Formats      map[string][]string   `json:"formats,omitempty"` // platform -> available formats
	Sources      []PlatformTrackSource `json:"sources"`           // Raw data from each platform
}

// PriceInfo represents pricing information
type PriceInfo struct {
	Currency string  `json:"currency"`
	Amount   float64 `json:"amount"`
	Format   string  `json:"format,omitempty"` // WAV, MP3, FLAC, etc.
}

// PlatformTrackSource represents track data from a specific platform
type PlatformTrackSource struct {
	Platform   string  `json:"platform"`
	ID         string  `json:"id"`
	Title      string  `json:"title"`
	Artists    string  `json:"artists"`
	BPM        float64 `json:"bpm,omitempty"`
	Key        string  `json:"key,omitempty"`
	Genre      string  `json:"genre,omitempty"`
	Price      string  `json:"price,omitempty"`
	URL        string  `json:"url"`
	Confidence float64 `json:"confidence"` // Match confidence 0-1
}

// UnifiedSearchResults contains aggregated search results
type UnifiedSearchResults struct {
	Query           string            `json:"query"`
	Platforms       []string          `json:"platforms_searched"`
	TotalResults    int               `json:"total_results"`
	Tracks          []UnifiedTrack    `json:"tracks"`
	PlatformResults map[string]int    `json:"platform_results"` // Results per platform
	Errors          map[string]string `json:"errors,omitempty"` // Platform errors
	SearchTime      string            `json:"search_time"`
}

// PlatformStatus represents connection status for a platform
type PlatformStatus struct {
	Platform    string `json:"platform"`
	Name        string `json:"name"`
	Connected   bool   `json:"connected"`
	Available   bool   `json:"available"`
	LastChecked string `json:"last_checked"`
	Error       string `json:"error,omitempty"`
}

// MusicDiscoveryStatus represents overall discovery service status
type MusicDiscoveryStatus struct {
	TotalPlatforms     int              `json:"total_platforms"`
	AvailablePlatforms int              `json:"available_platforms"`
	ConnectedPlatforms int              `json:"connected_platforms"`
	Platforms          []PlatformStatus `json:"platforms"`
}

// TrackMatch represents a potential match across platforms
type TrackMatch struct {
	Query            string         `json:"query"`
	BestMatch        *UnifiedTrack  `json:"best_match,omitempty"`
	AlternateMatches []UnifiedTrack `json:"alternate_matches,omitempty"`
	MatchConfidence  float64        `json:"match_confidence"`
	PlatformMatches  int            `json:"platform_matches"`
}

// PriceComparison represents price comparison across platforms
type PriceComparison struct {
	Track        UnifiedTrack        `json:"track"`
	BestPrice    *PlatformPrice      `json:"best_price,omitempty"`
	AllPrices    []PlatformPrice     `json:"all_prices"`
	Savings      string              `json:"savings,omitempty"`
	Availability map[string][]string `json:"availability"` // platform -> formats
}

// PlatformPrice represents a price from a specific platform
type PlatformPrice struct {
	Platform string  `json:"platform"`
	Price    float64 `json:"price"`
	Currency string  `json:"currency"`
	Format   string  `json:"format"`
	URL      string  `json:"url"`
}

// NewReleasesResult contains aggregated new releases
type NewReleasesResult struct {
	Platforms []string         `json:"platforms"`
	Releases  []UnifiedRelease `json:"releases"`
	Total     int              `json:"total"`
	FetchedAt string           `json:"fetched_at"`
}

// UnifiedRelease represents a release normalized across platforms
type UnifiedRelease struct {
	NormalizedID string            `json:"normalized_id"`
	Title        string            `json:"title"`
	Artists      []string          `json:"artists"`
	Label        string            `json:"label,omitempty"`
	ReleaseDate  string            `json:"release_date,omitempty"`
	TrackCount   int               `json:"track_count,omitempty"`
	Genre        string            `json:"genre,omitempty"`
	ImageURL     string            `json:"image_url,omitempty"`
	PlatformIDs  map[string]string `json:"platform_ids"`
	PlatformURLs map[string]string `json:"platform_urls"`
	Sources      []string          `json:"sources"` // Platforms where found
}

// LibraryStatusResult contains unified library view
type LibraryStatusResult struct {
	TotalTracks    int                     `json:"total_tracks"`
	TotalPlaylists int                     `json:"total_playlists"`
	PlatformStats  map[string]LibraryStats `json:"platform_stats"`
	RecentActivity []LibraryActivity       `json:"recent_activity,omitempty"`
}

// LibraryStats represents library statistics for a platform
type LibraryStats struct {
	Platform  string `json:"platform"`
	Tracks    int    `json:"tracks"`
	Playlists int    `json:"playlists"`
	Downloads int    `json:"downloads,omitempty"`
	LastSync  string `json:"last_sync,omitempty"`
}

// LibraryActivity represents recent library activity
type LibraryActivity struct {
	Platform  string `json:"platform"`
	Type      string `json:"type"` // "download", "playlist_add", "purchase"
	Item      string `json:"item"`
	Timestamp string `json:"timestamp"`
}

var (
	musicDiscoveryInstance *MusicDiscoveryClient
	musicDiscoveryOnce     sync.Once
)

// GetMusicDiscoveryClient returns the singleton music discovery client
func GetMusicDiscoveryClient() *MusicDiscoveryClient {
	musicDiscoveryOnce.Do(func() {
		musicDiscoveryInstance = &MusicDiscoveryClient{}
	})
	return musicDiscoveryInstance
}

// initClients initializes all platform clients lazily
func (c *MusicDiscoveryClient) initClients() {
	c.initOnce.Do(func() {
		// Initialize clients - errors are ignored for lazy init,
		// status checks will reveal connection issues
		c.beatport, _ = NewBeatportClient()
		c.traxsource, _ = NewTraxsourceClient()
		c.juno, _ = NewJunoClient()
		c.boomkat, _ = NewBoomkatClient()
		c.spotify, _ = NewSpotifyClient()
		c.soundcloud, _ = NewSoundCloudClient()
		c.bandcamp, _ = NewBandcampClient()
		c.mixcloud, _ = NewMixcloudClient()
		c.discogs, _ = NewDiscogsClient()
	})
}

// Status returns the connection status of all music platforms
func (c *MusicDiscoveryClient) Status(ctx context.Context) (*MusicDiscoveryStatus, error) {
	c.initClients()

	status := &MusicDiscoveryStatus{
		TotalPlatforms: len(AllPlatforms),
		Platforms:      make([]PlatformStatus, 0, len(AllPlatforms)),
	}

	// Check each platform in parallel
	var wg sync.WaitGroup
	var mu sync.Mutex
	results := make(chan PlatformStatus, len(AllPlatforms))

	for _, platform := range AllPlatforms {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			ps := c.checkPlatformStatus(ctx, p)
			results <- ps
		}(platform)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for ps := range results {
		mu.Lock()
		status.Platforms = append(status.Platforms, ps)
		if ps.Available {
			status.AvailablePlatforms++
		}
		if ps.Connected {
			status.ConnectedPlatforms++
		}
		mu.Unlock()
	}

	// Sort platforms by name for consistent output
	sort.Slice(status.Platforms, func(i, j int) bool {
		return status.Platforms[i].Platform < status.Platforms[j].Platform
	})

	return status, nil
}

// checkPlatformStatus checks the status of a single platform
func (c *MusicDiscoveryClient) checkPlatformStatus(ctx context.Context, platform string) PlatformStatus {
	info := PlatformInfoMap[platform]
	ps := PlatformStatus{
		Platform:    platform,
		Name:        info.Name,
		LastChecked: time.Now().UTC().Format(time.RFC3339),
	}

	switch platform {
	case PlatformBeatport:
		if c.beatport != nil {
			// Beatport uses IsAuthenticated() check
			ps.Connected = c.beatport.IsAuthenticated()
			ps.Available = true
		}
	case PlatformTraxsource:
		if c.traxsource != nil {
			_, err := c.traxsource.GetStatus(ctx)
			if err != nil {
				ps.Error = err.Error()
			} else {
				ps.Connected = true
				ps.Available = true
			}
		}
	case PlatformJuno:
		if c.juno != nil {
			_, err := c.juno.GetStatus(ctx)
			if err != nil {
				ps.Error = err.Error()
			} else {
				ps.Connected = true
				ps.Available = true
			}
		}
	case PlatformBoomkat:
		if c.boomkat != nil {
			_, err := c.boomkat.GetStatus(ctx)
			if err != nil {
				ps.Error = err.Error()
			} else {
				ps.Connected = true
				ps.Available = true
			}
		}
	case PlatformSpotify:
		if c.spotify != nil {
			status, err := c.spotify.GetStatus(ctx)
			if err != nil {
				ps.Error = err.Error()
			} else {
				ps.Connected = status.Connected
				ps.Available = true
			}
		}
	case PlatformSoundCloud:
		if c.soundcloud != nil {
			status, err := c.soundcloud.GetStatus(ctx)
			if err != nil {
				ps.Error = err.Error()
			} else {
				ps.Connected = status.Connected
				ps.Available = true
			}
		}
	case PlatformBandcamp:
		if c.bandcamp != nil {
			_, err := c.bandcamp.GetStatus(ctx)
			if err != nil {
				ps.Error = err.Error()
			} else {
				ps.Connected = true
				ps.Available = true
			}
		}
	case PlatformMixcloud:
		if c.mixcloud != nil {
			_, err := c.mixcloud.GetStatus(ctx)
			if err != nil {
				ps.Error = err.Error()
			} else {
				ps.Connected = true
				ps.Available = true
			}
		}
	case PlatformDiscogs:
		if c.discogs != nil {
			status, err := c.discogs.GetStatus(ctx)
			if err != nil {
				ps.Error = err.Error()
			} else {
				ps.Connected = status.Connected
				ps.Available = true
			}
		}
	}

	return ps
}

// Search performs a unified search across specified platforms
func (c *MusicDiscoveryClient) Search(ctx context.Context, query string, platforms []string, limit int) (*UnifiedSearchResults, error) {
	c.initClients()

	if len(platforms) == 0 {
		platforms = AllPlatforms
	}
	if limit <= 0 {
		limit = 20
	}

	startTime := time.Now()

	results := &UnifiedSearchResults{
		Query:           query,
		Platforms:       platforms,
		PlatformResults: make(map[string]int),
		Errors:          make(map[string]string),
	}

	// Search each platform in parallel
	var wg sync.WaitGroup
	tracksChan := make(chan []PlatformTrackSource, len(platforms))
	errorsChan := make(chan struct {
		platform string
		err      error
	}, len(platforms))

	for _, platform := range platforms {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			tracks, err := c.searchPlatform(ctx, p, query, limit)
			if err != nil {
				errorsChan <- struct {
					platform string
					err      error
				}{p, err}
				return
			}
			tracksChan <- tracks
		}(platform)
	}

	go func() {
		wg.Wait()
		close(tracksChan)
		close(errorsChan)
	}()

	// Collect results
	var allTracks []PlatformTrackSource
	for tracks := range tracksChan {
		allTracks = append(allTracks, tracks...)
	}

	// Collect errors
	for e := range errorsChan {
		results.Errors[e.platform] = e.err.Error()
	}

	// Normalize and deduplicate tracks
	results.Tracks = c.normalizeAndMerge(allTracks)
	results.TotalResults = len(results.Tracks)

	// Count results per platform
	for _, track := range results.Tracks {
		for platform := range track.PlatformIDs {
			results.PlatformResults[platform]++
		}
	}

	results.SearchTime = time.Since(startTime).String()

	return results, nil
}

// searchPlatform searches a single platform
func (c *MusicDiscoveryClient) searchPlatform(ctx context.Context, platform, query string, limit int) ([]PlatformTrackSource, error) {
	var tracks []PlatformTrackSource

	switch platform {
	case PlatformBeatport:
		if c.beatport != nil {
			results, err := c.beatport.SearchTracks(ctx, query, limit)
			if err != nil {
				return nil, err
			}
			for _, t := range results {
				artists := make([]string, len(t.Artists))
				for i, a := range t.Artists {
					artists[i] = a.Name
				}
				key := ""
				if t.Key != nil && t.Key.Camelot != nil {
					key = t.Key.Camelot.Key
				} else if t.Key != nil {
					key = t.Key.Name
				}
				genre := ""
				if t.Genre != nil {
					genre = t.Genre.Name
				}
				tracks = append(tracks, PlatformTrackSource{
					Platform:   PlatformBeatport,
					ID:         fmt.Sprintf("%d", t.ID),
					Title:      t.Name,
					Artists:    strings.Join(artists, ", "),
					BPM:        t.BPM,
					Key:        key,
					Genre:      genre,
					URL:        t.URL,
					Confidence: 1.0,
				})
			}
		}

	case PlatformTraxsource:
		if c.traxsource != nil {
			results, err := c.traxsource.Search(ctx, query, "tracks", 1, limit)
			if err != nil {
				return nil, err
			}
			for _, t := range results.Tracks {
				tracks = append(tracks, PlatformTrackSource{
					Platform:   PlatformTraxsource,
					ID:         fmt.Sprintf("%d", t.ID),
					Title:      t.Title,
					Artists:    strings.Join(t.Artists, ", "),
					BPM:        float64(t.BPM),
					Key:        t.Key,
					Genre:      t.Genre,
					Price:      t.Price,
					URL:        t.URL,
					Confidence: 1.0,
				})
			}
		}

	case PlatformJuno:
		if c.juno != nil {
			results, err := c.juno.Search(ctx, query, "tracks", 1, limit)
			if err != nil {
				return nil, err
			}
			for _, t := range results.Tracks {
				tracks = append(tracks, PlatformTrackSource{
					Platform:   PlatformJuno,
					ID:         fmt.Sprintf("%d", t.ID),
					Title:      t.Title,
					Artists:    strings.Join(t.Artists, ", "),
					BPM:        float64(t.BPM),
					Key:        t.Key,
					Genre:      t.Genre,
					Price:      t.Price,
					URL:        t.URL,
					Confidence: 1.0,
				})
			}
		}

	case PlatformBoomkat:
		if c.boomkat != nil {
			results, err := c.boomkat.Search(ctx, query, "releases", 1, limit)
			if err != nil {
				return nil, err
			}
			for _, r := range results.Releases {
				// Boomkat returns releases, extract tracks
				for _, t := range r.Tracks {
					tracks = append(tracks, PlatformTrackSource{
						Platform:   PlatformBoomkat,
						ID:         t.ID,
						Title:      t.Title,
						Artists:    strings.Join(r.Artists, ", "),
						Genre:      r.Genre,
						Price:      r.Price,
						URL:        r.URL,
						Confidence: 0.9,
					})
				}
				// If no tracks, add release as a single entry
				if len(r.Tracks) == 0 {
					tracks = append(tracks, PlatformTrackSource{
						Platform:   PlatformBoomkat,
						ID:         r.ID,
						Title:      r.Title,
						Artists:    strings.Join(r.Artists, ", "),
						Genre:      r.Genre,
						Price:      r.Price,
						URL:        r.URL,
						Confidence: 0.8,
					})
				}
			}
		}

	case PlatformSpotify:
		if c.spotify != nil {
			results, err := c.spotify.Search(ctx, query, []string{"track"}, limit)
			if err != nil {
				return nil, err
			}
			for _, t := range results.Tracks {
				artists := make([]string, len(t.Artists))
				for i, a := range t.Artists {
					artists[i] = a.Name
				}
				tracks = append(tracks, PlatformTrackSource{
					Platform:   PlatformSpotify,
					ID:         t.ID,
					Title:      t.Name,
					Artists:    strings.Join(artists, ", "),
					URL:        t.ExternalURL,
					Confidence: 1.0,
				})
			}
		}

	case PlatformSoundCloud:
		if c.soundcloud != nil {
			results, err := c.soundcloud.Search(ctx, query, "tracks", limit)
			if err != nil {
				return nil, err
			}
			for _, t := range results.Tracks {
				tracks = append(tracks, PlatformTrackSource{
					Platform:   PlatformSoundCloud,
					ID:         fmt.Sprintf("%d", t.ID),
					Title:      t.Title,
					Artists:    t.User.Username,
					Genre:      t.Genre,
					URL:        t.PermalinkURL,
					Confidence: 1.0,
				})
			}
		}

	case PlatformBandcamp:
		if c.bandcamp != nil {
			results, err := c.bandcamp.Search(ctx, query, "all")
			if err != nil {
				return nil, err
			}
			// Bandcamp search returns separate slices for artists, albums, tracks
			for _, t := range results.Tracks {
				tracks = append(tracks, PlatformTrackSource{
					Platform:   PlatformBandcamp,
					ID:         t.URL,
					Title:      t.Title,
					Artists:    t.Artist,
					Genre:      "", // Bandcamp tracks don't have genre, only tags
					URL:        t.URL,
					Confidence: 0.9,
				})
			}
			for _, a := range results.Albums {
				genre := ""
				if len(a.Tags) > 0 {
					genre = strings.Join(a.Tags, ", ")
				}
				tracks = append(tracks, PlatformTrackSource{
					Platform:   PlatformBandcamp,
					ID:         a.URL,
					Title:      a.Title,
					Artists:    a.Artist,
					Genre:      genre,
					URL:        a.URL,
					Confidence: 0.85,
				})
			}
		}

	case PlatformMixcloud:
		if c.mixcloud != nil {
			results, err := c.mixcloud.Search(ctx, query, "cloudcast")
			if err != nil {
				return nil, err
			}
			for _, show := range results.Cloudcasts {
				tracks = append(tracks, PlatformTrackSource{
					Platform:   PlatformMixcloud,
					ID:         show.Key,
					Title:      show.Name,
					Artists:    show.User.Username,
					URL:        show.URL,
					Confidence: 0.8, // Lower confidence for mixes
				})
			}
		}

	case PlatformDiscogs:
		if c.discogs != nil {
			results, err := c.discogs.Search(ctx, query, "", 1)
			if err != nil {
				return nil, err
			}
			for _, r := range results.Results {
				tracks = append(tracks, PlatformTrackSource{
					Platform:   PlatformDiscogs,
					ID:         fmt.Sprintf("%d", r.ID),
					Title:      r.Title,
					Artists:    "", // Discogs combines artist in title
					Genre:      strings.Join(r.Genre, ", "),
					URL:        r.URI,
					Confidence: 0.85,
				})
			}
		}
	}

	return tracks, nil
}

// normalizeAndMerge normalizes tracks and merges duplicates across platforms
func (c *MusicDiscoveryClient) normalizeAndMerge(tracks []PlatformTrackSource) []UnifiedTrack {
	// Group by normalized ID
	trackMap := make(map[string]*UnifiedTrack)

	for _, t := range tracks {
		normalizedID := c.generateNormalizedID(t.Artists, t.Title)

		if existing, ok := trackMap[normalizedID]; ok {
			// Merge with existing track
			existing.PlatformIDs[t.Platform] = t.ID
			existing.PlatformURLs[t.Platform] = t.URL
			existing.Sources = append(existing.Sources, t)

			// Update metadata if better source
			if t.BPM > 0 && existing.BPM == 0 {
				existing.BPM = t.BPM
			}
			if t.Key != "" && existing.Key == "" {
				existing.Key = t.Key
			}
			if t.Genre != "" && existing.Genre == "" {
				existing.Genre = t.Genre
			}
			if t.Price != "" {
				if existing.Prices == nil {
					existing.Prices = make(map[string]PriceInfo)
				}
				existing.Prices[t.Platform] = PriceInfo{
					Currency: "USD", // Default, would need parsing
					Format:   "MP3",
				}
			}
		} else {
			// Create new unified track
			artists := strings.Split(t.Artists, ", ")
			unified := &UnifiedTrack{
				NormalizedID: normalizedID,
				Title:        t.Title,
				Artists:      artists,
				BPM:          t.BPM,
				Key:          t.Key,
				Genre:        t.Genre,
				PlatformIDs:  map[string]string{t.Platform: t.ID},
				PlatformURLs: map[string]string{t.Platform: t.URL},
				Sources:      []PlatformTrackSource{t},
			}
			if t.Price != "" {
				unified.Prices = map[string]PriceInfo{
					t.Platform: {Currency: "USD", Format: "MP3"},
				}
			}
			trackMap[normalizedID] = unified
		}
	}

	// Convert map to slice and sort by number of platforms (most found first)
	result := make([]UnifiedTrack, 0, len(trackMap))
	for _, track := range trackMap {
		result = append(result, *track)
	}

	sort.Slice(result, func(i, j int) bool {
		// Sort by number of platforms found, then alphabetically
		if len(result[i].PlatformIDs) != len(result[j].PlatformIDs) {
			return len(result[i].PlatformIDs) > len(result[j].PlatformIDs)
		}
		return result[i].Title < result[j].Title
	})

	return result
}

// generateNormalizedID creates a unique ID from artist and title
func (c *MusicDiscoveryClient) generateNormalizedID(artist, title string) string {
	// Normalize: lowercase, remove special chars, trim whitespace
	normalized := strings.ToLower(artist + "|" + title)
	normalized = strings.TrimSpace(normalized)

	// Generate SHA256 hash
	hash := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(hash[:16]) // First 16 bytes = 32 hex chars
}

// MatchTrack finds the same track across all platforms
func (c *MusicDiscoveryClient) MatchTrack(ctx context.Context, artist, title string, platforms []string) (*TrackMatch, error) {
	c.initClients()

	if len(platforms) == 0 {
		platforms = AllPlatforms
	}

	query := fmt.Sprintf("%s %s", artist, title)
	results, err := c.Search(ctx, query, platforms, 10)
	if err != nil {
		return nil, err
	}

	match := &TrackMatch{
		Query: query,
	}

	if len(results.Tracks) == 0 {
		return match, nil
	}

	// Find best match by comparing normalized artist/title
	targetNorm := c.generateNormalizedID(artist, title)

	for i, track := range results.Tracks {
		if track.NormalizedID == targetNorm {
			match.BestMatch = &results.Tracks[i]
			match.MatchConfidence = 1.0
			match.PlatformMatches = len(track.PlatformIDs)
		} else if match.BestMatch == nil {
			// Use first result as fallback best match
			match.BestMatch = &results.Tracks[i]
			match.MatchConfidence = 0.8
			match.PlatformMatches = len(track.PlatformIDs)
		} else {
			// Add to alternates
			match.AlternateMatches = append(match.AlternateMatches, track)
		}
	}

	return match, nil
}

// ComparePrices compares prices for a track across all platforms
func (c *MusicDiscoveryClient) ComparePrices(ctx context.Context, artist, title string) (*PriceComparison, error) {
	c.initClients()

	// Only search platforms with pricing
	pricingPlatforms := []string{
		PlatformBeatport,
		PlatformTraxsource,
		PlatformJuno,
		PlatformBoomkat,
		PlatformBandcamp,
	}

	match, err := c.MatchTrack(ctx, artist, title, pricingPlatforms)
	if err != nil {
		return nil, err
	}

	if match.BestMatch == nil {
		return nil, fmt.Errorf("track not found: %s - %s", artist, title)
	}

	comparison := &PriceComparison{
		Track:        *match.BestMatch,
		AllPrices:    make([]PlatformPrice, 0),
		Availability: make(map[string][]string),
	}

	// Extract prices from each platform
	for platform, priceInfo := range match.BestMatch.Prices {
		pp := PlatformPrice{
			Platform: platform,
			Price:    priceInfo.Amount,
			Currency: priceInfo.Currency,
			Format:   priceInfo.Format,
			URL:      match.BestMatch.PlatformURLs[platform],
		}
		comparison.AllPrices = append(comparison.AllPrices, pp)

		if comparison.BestPrice == nil || pp.Price < comparison.BestPrice.Price {
			comparison.BestPrice = &pp
		}
	}

	// Sort by price
	sort.Slice(comparison.AllPrices, func(i, j int) bool {
		return comparison.AllPrices[i].Price < comparison.AllPrices[j].Price
	})

	// Calculate savings if multiple prices
	if len(comparison.AllPrices) >= 2 {
		highest := comparison.AllPrices[len(comparison.AllPrices)-1].Price
		lowest := comparison.AllPrices[0].Price
		if highest > 0 {
			savings := ((highest - lowest) / highest) * 100
			comparison.Savings = fmt.Sprintf("%.0f%% savings vs highest price", savings)
		}
	}

	return comparison, nil
}

// GetNewReleases fetches new releases from all platforms
func (c *MusicDiscoveryClient) GetNewReleases(ctx context.Context, platforms []string, limit int) (*NewReleasesResult, error) {
	c.initClients()

	if len(platforms) == 0 {
		platforms = AllPlatforms
	}
	if limit <= 0 {
		limit = 20
	}

	result := &NewReleasesResult{
		Platforms: platforms,
		Releases:  make([]UnifiedRelease, 0),
		FetchedAt: time.Now().UTC().Format(time.RFC3339),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	releasesChan := make(chan []UnifiedRelease, len(platforms))

	for _, platform := range platforms {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			releases := c.fetchNewReleases(ctx, p, limit)
			if len(releases) > 0 {
				releasesChan <- releases
			}
		}(platform)
	}

	go func() {
		wg.Wait()
		close(releasesChan)
	}()

	for releases := range releasesChan {
		mu.Lock()
		result.Releases = append(result.Releases, releases...)
		mu.Unlock()
	}

	result.Total = len(result.Releases)

	// Sort by release date (newest first)
	sort.Slice(result.Releases, func(i, j int) bool {
		return result.Releases[i].ReleaseDate > result.Releases[j].ReleaseDate
	})

	// Limit results
	if len(result.Releases) > limit {
		result.Releases = result.Releases[:limit]
	}

	return result, nil
}

// fetchNewReleases fetches new releases from a single platform
func (c *MusicDiscoveryClient) fetchNewReleases(ctx context.Context, platform string, limit int) []UnifiedRelease {
	var releases []UnifiedRelease

	switch platform {
	case PlatformBeatport:
		// Beatport doesn't have GetNewReleases - use GetCharts as proxy for new/popular releases
		if c.beatport != nil {
			charts, err := c.beatport.GetCharts(ctx, "")
			if err == nil && len(charts) > 0 {
				// Get tracks from the first chart
				chartTracks, err := c.beatport.GetChartTracks(ctx, charts[0].ID)
				if err == nil {
					for _, t := range chartTracks {
						artists := make([]string, len(t.Artists))
						for i, a := range t.Artists {
							artists[i] = a.Name
						}
						label := ""
						if t.Label != nil {
							label = t.Label.Name
						}
						releases = append(releases, UnifiedRelease{
							NormalizedID: c.generateNormalizedID(strings.Join(artists, ", "), t.Name),
							Title:        t.Name,
							Artists:      artists,
							Label:        label,
							ReleaseDate:  t.PublishDate,
							TrackCount:   1,
							PlatformIDs:  map[string]string{PlatformBeatport: fmt.Sprintf("%d", t.ID)},
							PlatformURLs: map[string]string{PlatformBeatport: t.URL},
							Sources:      []string{PlatformBeatport},
						})
					}
				}
			}
		}

	case PlatformTraxsource:
		// Traxsource doesn't have GetNewReleases - use GetTopCharts as proxy for new/hot releases
		if c.traxsource != nil {
			topTracks, err := c.traxsource.GetTopCharts(ctx, "", 1)
			if err == nil {
				// Group tracks by release (approximate since we only have track data)
				for _, t := range topTracks {
					releases = append(releases, UnifiedRelease{
						NormalizedID: c.generateNormalizedID(strings.Join(t.Artists, ", "), t.Title),
						Title:        t.Title,
						Artists:      t.Artists,
						Label:        t.Label,
						Genre:        t.Genre,
						ReleaseDate:  t.ReleaseDate,
						TrackCount:   1,
						PlatformIDs:  map[string]string{PlatformTraxsource: fmt.Sprintf("%d", t.ID)},
						PlatformURLs: map[string]string{PlatformTraxsource: t.URL},
						Sources:      []string{PlatformTraxsource},
					})
				}
			}
		}

	case PlatformJuno:
		if c.juno != nil {
			newReleases, err := c.juno.GetNewReleases(ctx, "", 1)
			if err == nil {
				for _, r := range newReleases {
					releases = append(releases, UnifiedRelease{
						NormalizedID: c.generateNormalizedID(strings.Join(r.Artists, ", "), r.Title),
						Title:        r.Title,
						Artists:      r.Artists,
						Label:        r.Label,
						Genre:        r.Genre,
						ReleaseDate:  r.ReleaseDate,
						TrackCount:   len(r.Tracks),
						PlatformIDs:  map[string]string{PlatformJuno: fmt.Sprintf("%d", r.ID)},
						PlatformURLs: map[string]string{PlatformJuno: r.URL},
						Sources:      []string{PlatformJuno},
					})
				}
			}
		}

	case PlatformBoomkat:
		if c.boomkat != nil {
			newReleases, err := c.boomkat.GetNewReleases(ctx, "", 1)
			if err == nil {
				for _, r := range newReleases {
					releases = append(releases, UnifiedRelease{
						NormalizedID: c.generateNormalizedID(strings.Join(r.Artists, ", "), r.Title),
						Title:        r.Title,
						Artists:      r.Artists,
						Label:        r.Label,
						Genre:        r.Genre,
						ReleaseDate:  r.ReleaseDate,
						TrackCount:   len(r.Tracks),
						ImageURL:     r.ImageURL,
						PlatformIDs:  map[string]string{PlatformBoomkat: r.ID},
						PlatformURLs: map[string]string{PlatformBoomkat: r.URL},
						Sources:      []string{PlatformBoomkat},
					})
				}
			}
		}
	}

	return releases
}

// GetLibraryStatus returns unified library view across platforms
func (c *MusicDiscoveryClient) GetLibraryStatus(ctx context.Context) (*LibraryStatusResult, error) {
	c.initClients()

	result := &LibraryStatusResult{
		PlatformStats: make(map[string]LibraryStats),
	}

	// This would need platform-specific library access
	// For now, return placeholder stats
	for _, platform := range AllPlatforms {
		info := PlatformInfoMap[platform]
		result.PlatformStats[platform] = LibraryStats{
			Platform: platform,
			LastSync: time.Now().UTC().Format(time.RFC3339),
		}
		_ = info // Use info for feature availability
	}

	return result, nil
}

// GetPlatforms returns information about all supported platforms
func (c *MusicDiscoveryClient) GetPlatforms(ctx context.Context) ([]PlatformInfo, error) {
	platforms := make([]PlatformInfo, 0, len(PlatformInfoMap))
	for _, info := range PlatformInfoMap {
		platforms = append(platforms, info)
	}

	sort.Slice(platforms, func(i, j int) bool {
		return platforms[i].Name < platforms[j].Name
	})

	return platforms, nil
}

// EnrichMetadata aggregates metadata from multiple platforms for a track
func (c *MusicDiscoveryClient) EnrichMetadata(ctx context.Context, artist, title string) (*UnifiedTrack, error) {
	c.initClients()

	// Search metadata-rich platforms first
	metadataPlatforms := []string{
		PlatformBeatport,   // Best for BPM, Key
		PlatformTraxsource, // Good for BPM, Key
		PlatformSpotify,    // Has audio features
		PlatformJuno,       // Has metadata
	}

	match, err := c.MatchTrack(ctx, artist, title, metadataPlatforms)
	if err != nil {
		return nil, err
	}

	if match.BestMatch == nil {
		return nil, fmt.Errorf("track not found: %s - %s", artist, title)
	}

	// The match already contains aggregated metadata from all platforms
	return match.BestMatch, nil
}
