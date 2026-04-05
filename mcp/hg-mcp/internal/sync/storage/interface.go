package storage

import (
	"context"
	"time"
)

// Storage defines the interface for sync state and playlist storage.
type Storage interface {
	// Sync state operations
	GetSyncState(ctx context.Context, service, user, playlist string) (*SyncStateRecord, error)
	UpdateSyncState(ctx context.Context, record *SyncStateRecord) error
	ListSyncStates(ctx context.Context, service, user string) ([]*SyncStateRecord, error)

	// Playlist operations
	GetPlaylist(ctx context.Context, id string) (*PlaylistRecord, error)
	UpdatePlaylist(ctx context.Context, record *PlaylistRecord) error
	ListPlaylists(ctx context.Context, userID string) ([]*PlaylistRecord, error)

	// Queue operations
	EnqueueTrack(ctx context.Context, item *QueueItem) error
	DequeueTrack(ctx context.Context) (*QueueItem, error)
	CompleteTrack(ctx context.Context, id string, err error) error
	GetQueueStats(ctx context.Context) (map[string]int, error)

	// History
	RecordSyncHistory(ctx context.Context, playlistID, status string, details map[string]interface{}) error

	// Health check
	Health(ctx context.Context) error
}

// PlaylistState is a sync-friendly wrapper for playlist operations.
type PlaylistState struct {
	ID               string
	Service          string
	Username         string
	URL              string
	TrackCount       int
	TracksDiscovered int
	TracksQueued     int
	TracksCompleted  int
	TracksFailed     int
	IndexStatus      string
	LastSyncedAt     *time.Time
}

// ToRecord converts PlaylistState to PlaylistRecord.
func (p *PlaylistState) ToRecord() *PlaylistRecord {
	return &PlaylistRecord{
		ID:               p.ID,
		Service:          p.Service,
		Username:         p.Username,
		URL:              p.URL,
		TrackCount:       p.TrackCount,
		TracksDiscovered: p.TracksDiscovered,
		TracksQueued:     p.TracksQueued,
		TracksCompleted:  p.TracksCompleted,
		TracksFailed:     p.TracksFailed,
		IndexStatus:      p.IndexStatus,
		LastSyncedAt:     p.LastSyncedAt,
	}
}

// FromRecord creates PlaylistState from PlaylistRecord.
func FromRecord(r *PlaylistRecord) *PlaylistState {
	return &PlaylistState{
		ID:               r.ID,
		Service:          r.Service,
		Username:         r.Username,
		URL:              r.URL,
		TrackCount:       r.TrackCount,
		TracksDiscovered: r.TracksDiscovered,
		TracksQueued:     r.TracksQueued,
		TracksCompleted:  r.TracksCompleted,
		TracksFailed:     r.TracksFailed,
		IndexStatus:      r.IndexStatus,
		LastSyncedAt:     r.LastSyncedAt,
	}
}
