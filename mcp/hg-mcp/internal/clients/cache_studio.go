package clients

import (
	"context"
	"time"

	"github.com/hairglasses-studio/mcpkit/resilience"
)

// Studio client cache entries for reducing redundant API calls.
// These wrap the GetStatus methods of studio clients with TTL caching.

var (
	abletonStatusCache  = resilience.NewCache[*AbletonStatus](15 * time.Second)
	resolumeStatusCache = resilience.NewCache[*ResolumeStatus](10 * time.Second)
	obsStatusCache      = resilience.NewCache[*OBSStatus](10 * time.Second)
	grandma3StatusCache = resilience.NewCache[*GrandMA3Status](15 * time.Second)
	resolumeLayersCache = resilience.NewCache[[]ResolumeLayer](15 * time.Second)
)

// GetAbletonStatusCached returns Ableton status with 15s TTL caching.
func GetAbletonStatusCached(ctx context.Context, c *AbletonClient) (*AbletonStatus, error) {
	return abletonStatusCache.GetOrFetch(ctx, func(ctx context.Context) (*AbletonStatus, error) {
		return c.GetStatus(ctx)
	})
}

// GetResolumeStatusCached returns Resolume status with 10s TTL caching.
func GetResolumeStatusCached(ctx context.Context, c *ResolumeClient) (*ResolumeStatus, error) {
	return resolumeStatusCache.GetOrFetch(ctx, func(ctx context.Context) (*ResolumeStatus, error) {
		return c.GetStatus(ctx)
	})
}

// GetOBSStatusCached returns OBS status with 10s TTL caching.
func GetOBSStatusCached(ctx context.Context, c *OBSClient) (*OBSStatus, error) {
	return obsStatusCache.GetOrFetch(ctx, func(ctx context.Context) (*OBSStatus, error) {
		return c.GetStatus(ctx)
	})
}

// GetGrandMA3StatusCached returns grandMA3 status with 15s TTL caching.
func GetGrandMA3StatusCached(ctx context.Context, c *GrandMA3Client) (*GrandMA3Status, error) {
	return grandma3StatusCache.GetOrFetch(ctx, func(ctx context.Context) (*GrandMA3Status, error) {
		return c.GetStatus(ctx)
	})
}

// GetResolumeLayersCached returns Resolume layers with 15s TTL caching.
func GetResolumeLayersCached(ctx context.Context, c *ResolumeClient) ([]ResolumeLayer, error) {
	return resolumeLayersCache.GetOrFetch(ctx, func(ctx context.Context) ([]ResolumeLayer, error) {
		return c.GetLayers(ctx)
	})
}

// InvalidateStudioCaches clears all studio client caches.
func InvalidateStudioCaches() {
	abletonStatusCache.Invalidate()
	resolumeStatusCache.Invalidate()
	obsStatusCache.Invalidate()
	grandma3StatusCache.Invalidate()
	resolumeLayersCache.Invalidate()
}
