package clients

import (
	"context"
	"time"

	"github.com/hairglasses-studio/hg-mcp/pkg/cache"
)

// Rekordbox cache entries for reducing redundant Python script invocations.
// Python overhead makes caching especially valuable for these methods.

var (
	rekordboxStatusCache     = cache.New[*RekordboxStatus](15 * time.Second)
	rekordboxPlaylistsCache  = cache.New[[]RekordboxPlaylist](30 * time.Second)
	rekordboxNowPlayingCache = cache.New[*RekordboxTrack](5 * time.Second)
)

// GetRekordboxStatusCached returns Rekordbox status with 15s TTL caching.
func GetRekordboxStatusCached(ctx context.Context, c *RekordboxClient) (*RekordboxStatus, error) {
	return rekordboxStatusCache.GetOrFetch(ctx, func(ctx context.Context) (*RekordboxStatus, error) {
		return c.GetStatus(ctx)
	})
}

// GetRekordboxPlaylistsCached returns Rekordbox playlists with 30s TTL caching.
func GetRekordboxPlaylistsCached(ctx context.Context, c *RekordboxClient) ([]RekordboxPlaylist, error) {
	return rekordboxPlaylistsCache.GetOrFetch(ctx, func(ctx context.Context) ([]RekordboxPlaylist, error) {
		return c.GetPlaylists(ctx)
	})
}

// GetRekordboxNowPlayingCached returns the currently playing track with 5s TTL caching.
func GetRekordboxNowPlayingCached(ctx context.Context, c *RekordboxClient) (*RekordboxTrack, error) {
	return rekordboxNowPlayingCache.GetOrFetch(ctx, func(ctx context.Context) (*RekordboxTrack, error) {
		return c.GetNowPlaying(ctx)
	})
}

// InvalidateRekordboxCaches clears all Rekordbox caches.
func InvalidateRekordboxCaches() {
	rekordboxStatusCache.Invalidate()
	rekordboxPlaylistsCache.Invalidate()
	rekordboxNowPlayingCache.Invalidate()
}
