// Package bridge provides bridges for connecting Pro DJ Link to other systems.
package bridge

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
)

// ResolumeDisplayMode determines how track info is displayed
type ResolumeDisplayMode string

const (
	// ResolumeDisplaySeparate puts artist and title on separate dashboard strings
	ResolumeDisplaySeparate ResolumeDisplayMode = "separate"
	// ResolumeDisplayCombined puts "Artist - Title" on a single dashboard string
	ResolumeDisplayCombined ResolumeDisplayMode = "combined"
	// ResolumeDisplayFull shows artist, title, key, bpm, and genre
	ResolumeDisplayFull ResolumeDisplayMode = "full"
)

// ResolumeConfig configures the prolink-resolume bridge
type ResolumeConfig struct {
	// DisplayMode determines how track info is shown
	DisplayMode ResolumeDisplayMode `json:"display_mode"`
	// UpdateOnLoad if true, updates display when track loads (not just on play)
	UpdateOnLoad bool `json:"update_on_load"`
	// ClearOnStop if true, clears display when playback stops
	ClearOnStop bool `json:"clear_on_stop"`
	// PollIntervalMs is how often to poll for track changes (default 500)
	PollIntervalMs int `json:"poll_interval_ms"`
}

// ResolumeStatus represents the current state of the resolume bridge
type ResolumeStatus struct {
	Running       bool                `json:"running"`
	DisplayMode   ResolumeDisplayMode `json:"display_mode"`
	CurrentArtist string              `json:"current_artist,omitempty"`
	CurrentTitle  string              `json:"current_title,omitempty"`
	TrackChanges  int                 `json:"track_changes"`
	LastUpdate    time.Time           `json:"last_update,omitempty"`
	Error         string              `json:"error,omitempty"`
}

// ResolumeTrackEvent is emitted when a track changes
type ResolumeTrackEvent struct {
	PlayerID int
	Artist   string
	Title    string
	Key      string
	BPM      float64
	Genre    string
}

// ResolumeDisplayBridge syncs Pro DJ Link track info to Resolume
type ResolumeDisplayBridge struct {
	mu sync.RWMutex

	config   *ResolumeConfig
	prolink  *clients.ProlinkClient
	resolume *clients.ResolumeClient

	running       bool
	cancel        context.CancelFunc
	currentArtist string
	currentTitle  string
	trackChanges  int
	lastUpdate    time.Time
	lastError     error

	// Callbacks
	onTrackChange func(ResolumeTrackEvent)
}

var (
	resolumeDisplayBridge     *ResolumeDisplayBridge
	resolumeDisplayBridgeMu   sync.Mutex
	resolumeDisplayBridgeOnce sync.Once
)

// GetResolumeDisplayBridge returns the singleton resolume display bridge
func GetResolumeDisplayBridge() (*ResolumeDisplayBridge, error) {
	resolumeDisplayBridgeMu.Lock()
	defer resolumeDisplayBridgeMu.Unlock()

	if resolumeDisplayBridge != nil {
		return resolumeDisplayBridge, nil
	}

	prolink, err := clients.GetProlinkClient()
	if err != nil {
		return nil, fmt.Errorf("get prolink client: %w", err)
	}

	resolume, err := clients.NewResolumeClient()
	if err != nil {
		return nil, fmt.Errorf("get resolume client: %w", err)
	}

	resolumeDisplayBridge = &ResolumeDisplayBridge{
		prolink:  prolink,
		resolume: resolume,
		config: &ResolumeConfig{
			DisplayMode:    ResolumeDisplaySeparate,
			UpdateOnLoad:   true,
			ClearOnStop:    false,
			PollIntervalMs: 500,
		},
	}

	return resolumeDisplayBridge, nil
}

// Configure updates the bridge configuration
func (b *ResolumeDisplayBridge) Configure(config *ResolumeConfig) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if config.DisplayMode != "" {
		b.config.DisplayMode = config.DisplayMode
	}
	b.config.UpdateOnLoad = config.UpdateOnLoad
	b.config.ClearOnStop = config.ClearOnStop
	if config.PollIntervalMs > 0 {
		b.config.PollIntervalMs = config.PollIntervalMs
	}
}

// SetOnTrackChange sets the callback for track changes
func (b *ResolumeDisplayBridge) SetOnTrackChange(fn func(ResolumeTrackEvent)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.onTrackChange = fn
}

// Start begins syncing track info to Resolume
func (b *ResolumeDisplayBridge) Start(ctx context.Context) error {
	b.mu.Lock()
	if b.running {
		b.mu.Unlock()
		return fmt.Errorf("bridge already running")
	}

	// Connect to prolink if not already connected
	if err := b.prolink.Connect(ctx); err != nil {
		b.mu.Unlock()
		return fmt.Errorf("connect prolink: %w", err)
	}

	bridgeCtx, cancel := context.WithCancel(ctx)
	b.cancel = cancel
	b.running = true
	b.trackChanges = 0
	b.lastError = nil
	b.mu.Unlock()

	go b.syncLoop(bridgeCtx)

	return nil
}

// Stop stops the bridge
func (b *ResolumeDisplayBridge) Stop() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.cancel != nil {
		b.cancel()
		b.cancel = nil
	}
	b.running = false

	// Clear display if configured
	if b.config.ClearOnStop {
		_ = b.resolume.ClearTrackDisplay()
	}
}

// GetStatus returns the current bridge status
func (b *ResolumeDisplayBridge) GetStatus() ResolumeStatus {
	b.mu.RLock()
	defer b.mu.RUnlock()

	status := ResolumeStatus{
		Running:       b.running,
		DisplayMode:   b.config.DisplayMode,
		CurrentArtist: b.currentArtist,
		CurrentTitle:  b.currentTitle,
		TrackChanges:  b.trackChanges,
		LastUpdate:    b.lastUpdate,
	}

	if b.lastError != nil {
		status.Error = b.lastError.Error()
	}

	return status
}

// UpdateNowPlaying manually sets the now playing info
func (b *ResolumeDisplayBridge) UpdateNowPlaying(artist, title string) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	return b.updateDisplay(artist, title, "", 0, "")
}

// updateDisplay sends track info to Resolume based on display mode
func (b *ResolumeDisplayBridge) updateDisplay(artist, title, key string, bpm float64, genre string) error {
	var err error

	switch b.config.DisplayMode {
	case ResolumeDisplayCombined:
		err = b.resolume.SetFormattedNowPlaying(artist, title)
	case ResolumeDisplayFull:
		err = b.resolume.SetTrackInfo(clients.TrackDisplay{
			Artist: artist,
			Title:  title,
			Key:    key,
			BPM:    bpm,
			Genre:  genre,
		})
	default: // ResolumeDisplaySeparate
		err = b.resolume.SetNowPlaying(artist, title)
	}

	if err != nil {
		b.lastError = err
		return err
	}

	b.currentArtist = artist
	b.currentTitle = title
	b.lastUpdate = time.Now()
	return nil
}

// syncLoop polls for track changes and updates Resolume
func (b *ResolumeDisplayBridge) syncLoop(ctx context.Context) {
	interval := time.Duration(b.config.PollIntervalMs) * time.Millisecond
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var lastArtist, lastTitle string

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.pollAndUpdate(ctx, &lastArtist, &lastTitle)
		}
	}
}

// pollAndUpdate checks for track changes and updates display
func (b *ResolumeDisplayBridge) pollAndUpdate(ctx context.Context, lastArtist, lastTitle *string) {
	// Get full prolink data which includes track info
	fullData, err := b.prolink.GetFullData(ctx)
	if err != nil {
		b.mu.Lock()
		b.lastError = err
		b.mu.Unlock()
		return
	}

	if !fullData.Connected {
		return
	}

	// Find the master player or first playing device
	var artist, title, key, genre string
	var bpm float64
	var found bool

	for _, player := range fullData.Players {
		// Get track from this device if it's playing
		isPlaying := player.Status != nil && player.Status.IsPlaying
		hasTrack := player.Track != nil

		if hasTrack && (isPlaying || b.config.UpdateOnLoad) {
			artist = player.Track.Artist
			title = player.Track.Title
			key = player.Track.Key
			genre = player.Track.Genre
			// BPM comes from status, not track
			if player.Status != nil {
				bpm = float64(player.Status.EffectiveBPM)
			}
			found = true
			break
		}
	}

	if !found {
		return
	}

	// Check if track changed
	if artist == *lastArtist && title == *lastTitle {
		return
	}

	*lastArtist = artist
	*lastTitle = title

	b.mu.Lock()
	b.trackChanges++
	err = b.updateDisplay(artist, title, key, bpm, genre)
	onTrackChange := b.onTrackChange
	b.mu.Unlock()

	// Fire callback if set
	if onTrackChange != nil && err == nil {
		onTrackChange(ResolumeTrackEvent{
			Artist: artist,
			Title:  title,
			Key:    key,
			BPM:    bpm,
			Genre:  genre,
		})
	}
}
