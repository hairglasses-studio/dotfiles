package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/sync/storage"
)

// Config holds sync configuration
type Config struct {
	Users      []UserConfig   `json:"users"`
	LocalRoot  string         `json:"local_root"`  // ~/Music
	S3Bucket   string         `json:"s3_bucket"`   // cr8-music-storage
	AWSProfile string         `json:"aws_profile"` // cr8
	AWSRegion  string         `json:"aws_region"`  // us-east-1
	Backend    string         `json:"backend"`     // "local" or "aws"
	DryRun     bool           `json:"dry_run"`
	DynamoDB   DynamoDBConfig `json:"dynamodb"`
}

// DynamoDBConfig holds DynamoDB table names
type DynamoDBConfig struct {
	SyncStateTable   string `json:"sync_state_table"`
	PlaylistsTable   string `json:"playlists_table"`
	SyncQueueTable   string `json:"sync_queue_table"`
	SyncHistoryTable string `json:"sync_history_table"`
}

// UserConfig holds per-user configuration
type UserConfig struct {
	Username            string   `json:"username"`             // SoundCloud username
	DisplayName         string   `json:"display_name"`         // Rekordbox folder name
	SoundCloud          bool     `json:"soundcloud"`           // Enable SoundCloud sync
	SoundCloudPlaylists []string `json:"soundcloud_playlists"` // Additional SC playlists to sync (besides likes)
	Beatport            bool     `json:"beatport"`             // Enable Beatport sync
	BeatportUserID      string   `json:"beatport_userid"`      // Beatport user ID
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	home, _ := os.UserHomeDir()
	return &Config{
		Users: []UserConfig{
			{Username: "hairglasses", DisplayName: "Hairglasses", SoundCloud: true, Beatport: true},
			{Username: "freaq-show", DisplayName: "FreaQShoW", SoundCloud: true, Beatport: false, SoundCloudPlaylists: []string{}},
			{Username: "luke-lasley", DisplayName: "Luke", SoundCloud: true, Beatport: false},
			{Username: "fogel", DisplayName: "Fogel", SoundCloud: true, Beatport: false},
			{Username: "rahul", DisplayName: "Rahul", SoundCloud: true, Beatport: false},
			{Username: "aidan", DisplayName: "Aidan", SoundCloud: true, Beatport: false},
			{Username: "marissughdevelops", DisplayName: "Marissa", SoundCloud: true, Beatport: false, SoundCloudPlaylists: []string{}},
		},
		LocalRoot:  filepath.Join(home, "Music"),
		S3Bucket:   "cr8-music-storage",
		AWSProfile: "cr8",
		AWSRegion:  "us-east-1",
		Backend:    "aws",
		DryRun:     false,
		DynamoDB: DynamoDBConfig{
			SyncStateTable:   "cr8_sync_state",
			PlaylistsTable:   "cr8_playlists",
			SyncQueueTable:   "cr8_sync_queue",
			SyncHistoryTable: "cr8_sync_history",
		},
	}
}

// Manager orchestrates sync operations
type Manager struct {
	config     *Config
	state      *State
	storage    storage.Storage
	soundcloud *SoundCloudSyncer
	beatport   *BeatportSyncer
	rekordbox  *RekordboxImporter
}

// SyncResult holds the result of a sync operation
type SyncResult struct {
	Service   string    `json:"service"`
	User      string    `json:"user"`
	Playlist  string    `json:"playlist"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Total     int       `json:"total"`
	Synced    int       `json:"synced"`
	Skipped   int       `json:"skipped"`
	Failed    int       `json:"failed"`
	Errors    []string  `json:"errors,omitempty"`
	DryRun    bool      `json:"dry_run"`
}

// NewManager creates a new sync manager
func NewManager(config *Config) (*Manager, error) {
	return NewManagerWithContext(context.Background(), config)
}

// NewManagerWithContext creates a new sync manager with context
func NewManagerWithContext(ctx context.Context, config *Config) (*Manager, error) {
	if config == nil {
		config = DefaultConfig()
	}

	m := &Manager{
		config:     config,
		soundcloud: NewSoundCloudSyncer(config),
		beatport:   NewBeatportSyncer(config),
		rekordbox:  NewRekordboxImporter(config),
	}

	// Initialize storage backend
	if config.Backend == "aws" {
		awsStorage, err := storage.NewAWSStorage(ctx,
			storage.WithProfile(config.AWSProfile),
			storage.WithRegion(config.AWSRegion),
		)
		if err != nil {
			log.Printf("AWS storage init failed, falling back to local: %v", err)
			config.Backend = "local"
		} else {
			m.storage = awsStorage
			log.Printf("Using AWS DynamoDB storage backend")
		}
	}

	// Load local state (still needed for Rekordbox pending files)
	state, err := LoadState("")
	if err != nil {
		return nil, fmt.Errorf("load state: %w", err)
	}
	m.state = state

	return m, nil
}

// Storage returns the storage backend
func (m *Manager) Storage() storage.Storage {
	return m.storage
}

// SyncAll runs a full sync for all services and users
func (m *Manager) SyncAll(ctx context.Context) ([]SyncResult, error) {
	var results []SyncResult

	// Sync SoundCloud for all users
	for _, user := range m.config.Users {
		if !user.SoundCloud {
			continue
		}

		scResults, err := m.SyncSoundCloud(ctx, user.Username)
		if err != nil {
			log.Printf("SoundCloud sync error for %s: %v", user.Username, err)
		}
		results = append(results, scResults...)

		// Record to AWS storage
		for _, r := range scResults {
			m.recordSyncToAWS(ctx, r)
		}
	}

	// Sync Beatport for enabled users
	for _, user := range m.config.Users {
		if !user.Beatport {
			continue
		}

		bpResults, err := m.SyncBeatport(ctx, user.Username)
		if err != nil {
			log.Printf("Beatport sync error for %s: %v", user.Username, err)
		}
		results = append(results, bpResults...)

		// Record to AWS storage
		for _, r := range bpResults {
			m.recordSyncToAWS(ctx, r)
		}
	}

	// Import to Rekordbox
	rbResult, err := m.SyncRekordbox(ctx)
	if err != nil {
		log.Printf("Rekordbox import error: %v", err)
	}
	results = append(results, rbResult)
	m.recordSyncToAWS(ctx, rbResult)

	// Save local state (for Rekordbox pending files)
	if err := m.state.Save(); err != nil {
		log.Printf("Failed to save local state: %v", err)
	}

	return results, nil
}

// recordSyncToAWS records sync result to AWS DynamoDB
func (m *Manager) recordSyncToAWS(ctx context.Context, result SyncResult) {
	if m.storage == nil {
		return
	}

	// Update sync state
	record := &storage.SyncStateRecord{
		Service:    result.Service,
		User:       result.User,
		Playlist:   result.Playlist,
		LastSync:   result.EndTime,
		TrackCount: result.Total,
		Synced:     result.Synced,
		Failed:     result.Failed,
		Errors:     result.Errors,
	}
	if err := m.storage.UpdateSyncState(ctx, record); err != nil {
		log.Printf("Failed to update AWS sync state: %v", err)
	}

	// Record to history
	status := "completed"
	if result.Failed > 0 {
		status = "partial"
	}
	details := map[string]interface{}{
		"total":    result.Total,
		"synced":   result.Synced,
		"skipped":  result.Skipped,
		"failed":   result.Failed,
		"duration": result.EndTime.Sub(result.StartTime).String(),
		"dry_run":  result.DryRun,
	}
	playlistID := fmt.Sprintf("%s#%s#%s", result.Service, result.User, result.Playlist)
	if err := m.storage.RecordSyncHistory(ctx, playlistID, status, details); err != nil {
		log.Printf("Failed to record AWS sync history: %v", err)
	}
}

// SyncSoundCloud syncs SoundCloud playlists for a user
func (m *Manager) SyncSoundCloud(ctx context.Context, username string) ([]SyncResult, error) {
	StartSyncOp("soundcloud")
	defer EndSyncOp("soundcloud")

	results, err := m.soundcloud.Sync(ctx, username, m.state)
	for _, r := range results {
		RecordSyncResult(r)
	}
	return results, err
}

// SyncBeatport syncs Beatport playlists for a user
func (m *Manager) SyncBeatport(ctx context.Context, username string) ([]SyncResult, error) {
	StartSyncOp("beatport")
	defer EndSyncOp("beatport")

	results, err := m.beatport.Sync(ctx, username, m.state)
	for _, r := range results {
		RecordSyncResult(r)
	}
	return results, err
}

// SyncRekordbox imports pending files to Rekordbox
func (m *Manager) SyncRekordbox(ctx context.Context) (SyncResult, error) {
	StartSyncOp("rekordbox")
	defer EndSyncOp("rekordbox")

	result, err := m.rekordbox.Import(ctx, m.state)
	RecordSyncResult(result)
	return result, err
}

// Status returns the current sync status
func (m *Manager) Status() (*State, error) {
	return m.state, nil
}

// ToJSON converts results to JSON
func (r SyncResult) ToJSON() string {
	data, _ := json.MarshalIndent(r, "", "  ")
	return string(data)
}

// ListPlaylists returns all playlists from AWS storage
func (m *Manager) ListPlaylists(ctx context.Context, userID string) ([]*storage.PlaylistRecord, error) {
	if m.storage == nil {
		return nil, fmt.Errorf("AWS storage not initialized")
	}
	return m.storage.ListPlaylists(ctx, userID)
}

// GetPlaylist returns a single playlist from AWS storage
func (m *Manager) GetPlaylist(ctx context.Context, id string) (*storage.PlaylistRecord, error) {
	if m.storage == nil {
		return nil, fmt.Errorf("AWS storage not initialized")
	}
	return m.storage.GetPlaylist(ctx, id)
}

// UpdatePlaylist creates or updates a playlist in AWS storage
func (m *Manager) UpdatePlaylist(ctx context.Context, playlist *storage.PlaylistRecord) error {
	if m.storage == nil {
		return fmt.Errorf("AWS storage not initialized")
	}
	return m.storage.UpdatePlaylist(ctx, playlist)
}

// GetSyncState returns sync state from AWS storage
func (m *Manager) GetSyncState(ctx context.Context, service, user, playlist string) (*storage.SyncStateRecord, error) {
	if m.storage == nil {
		return nil, fmt.Errorf("AWS storage not initialized")
	}
	return m.storage.GetSyncState(ctx, service, user, playlist)
}

// GetQueueStats returns queue statistics from AWS storage
func (m *Manager) GetQueueStats(ctx context.Context) (map[string]int, error) {
	if m.storage == nil {
		return nil, fmt.Errorf("AWS storage not initialized")
	}
	return m.storage.GetQueueStats(ctx)
}

// EnqueueTrack adds a track to the sync queue in AWS
func (m *Manager) EnqueueTrack(ctx context.Context, playlistID, trackURL string, priority int) error {
	if m.storage == nil {
		return fmt.Errorf("AWS storage not initialized")
	}
	item := &storage.QueueItem{
		ID:         fmt.Sprintf("%s#%d", trackURL, time.Now().UnixNano()),
		PlaylistID: playlistID,
		TrackURL:   trackURL,
		Priority:   priority,
		Status:     "pending",
	}
	return m.storage.EnqueueTrack(ctx, item)
}

// StorageHealth checks AWS storage connectivity
func (m *Manager) StorageHealth(ctx context.Context) error {
	if m.storage == nil {
		return fmt.Errorf("AWS storage not initialized")
	}
	return m.storage.Health(ctx)
}
