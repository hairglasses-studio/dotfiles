package sync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// State persists sync progress across runs
type State struct {
	SoundCloud map[string]*UserSyncState `json:"soundcloud"`
	Beatport   map[string]*UserSyncState `json:"beatport"`
	Rekordbox  *RekordboxState           `json:"rekordbox"`
	mu         sync.RWMutex
	path       string
}

// UserSyncState tracks sync state per user
type UserSyncState struct {
	Playlists map[string]*PlaylistSyncState `json:"playlists"`
}

// PlaylistSyncState tracks sync state per playlist
type PlaylistSyncState struct {
	LastSync    time.Time `json:"last_sync"`
	TrackCount  int       `json:"track_count"`
	Synced      int       `json:"synced"`
	Failed      int       `json:"failed"`
	LastTrackID string    `json:"last_track_id,omitempty"`
	Errors      []string  `json:"errors,omitempty"`
}

// RekordboxState tracks Rekordbox import state
type RekordboxState struct {
	LastImport   time.Time `json:"last_import"`
	PendingFiles []string  `json:"pending_files"`
	LastError    string    `json:"last_error,omitempty"`
}

// DefaultStatePath returns the default state file path
func DefaultStatePath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".cache", "aftrs", "sync-state.json")
}

// LoadState loads state from disk or creates a new one
func LoadState(path string) (*State, error) {
	if path == "" {
		path = DefaultStatePath()
	}

	state := &State{
		SoundCloud: make(map[string]*UserSyncState),
		Beatport:   make(map[string]*UserSyncState),
		Rekordbox:  &RekordboxState{},
		path:       path,
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return state, nil
		}
		return nil, err
	}

	if err := json.Unmarshal(data, state); err != nil {
		return nil, err
	}
	state.path = path

	return state, nil
}

// Save persists state to disk
func (s *State) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

// GetPlaylistState returns the sync state for a playlist
func (s *State) GetPlaylistState(service, user, playlist string) *PlaylistSyncState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var userStates map[string]*UserSyncState
	switch service {
	case "soundcloud":
		userStates = s.SoundCloud
	case "beatport":
		userStates = s.Beatport
	default:
		return nil
	}

	userState, ok := userStates[user]
	if !ok {
		return nil
	}

	return userState.Playlists[playlist]
}

// UpdatePlaylistState updates the sync state for a playlist
func (s *State) UpdatePlaylistState(service, user, playlist string, update func(*PlaylistSyncState)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var userStates map[string]*UserSyncState
	switch service {
	case "soundcloud":
		userStates = s.SoundCloud
	case "beatport":
		userStates = s.Beatport
	default:
		return
	}

	userState, ok := userStates[user]
	if !ok {
		userState = &UserSyncState{Playlists: make(map[string]*PlaylistSyncState)}
		userStates[user] = userState
	}

	playlistState, ok := userState.Playlists[playlist]
	if !ok {
		playlistState = &PlaylistSyncState{}
		userState.Playlists[playlist] = playlistState
	}

	update(playlistState)
}

// UpdateRekordbox updates the Rekordbox import state
func (s *State) UpdateRekordbox(update func(*RekordboxState)) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Rekordbox == nil {
		s.Rekordbox = &RekordboxState{}
	}
	update(s.Rekordbox)
}

// AddPendingFile adds a file to the Rekordbox import queue
func (s *State) AddPendingFile(path string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Rekordbox == nil {
		s.Rekordbox = &RekordboxState{}
	}

	// Check if already pending
	for _, p := range s.Rekordbox.PendingFiles {
		if p == path {
			return
		}
	}
	s.Rekordbox.PendingFiles = append(s.Rekordbox.PendingFiles, path)
}

// ClearPendingFiles removes all pending files
func (s *State) ClearPendingFiles() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Rekordbox != nil {
		s.Rekordbox.PendingFiles = nil
	}
}
