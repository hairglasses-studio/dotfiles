package clients

import (
	"context"
	"time"

	"github.com/hairglasses-studio/hg-mcp/pkg/cache"
)

// Google service cache entries for reducing redundant API calls.
// These wrap read-only list/get methods with TTL caching.

var (
	calendarListCache   = cache.New[[]CalendarInfo](60 * time.Second)
	calendarColorsCache = cache.New[map[string]string](24 * time.Hour)
	taskListsCache      = cache.New[[]TaskList](30 * time.Second)
)

// GetListCalendarsCached returns calendar list with 60s TTL caching.
func GetListCalendarsCached(ctx context.Context, c *CalendarClient) ([]CalendarInfo, error) {
	return calendarListCache.GetOrFetch(ctx, func(ctx context.Context) ([]CalendarInfo, error) {
		return c.ListCalendars(ctx)
	})
}

// GetCalendarColorsCached returns calendar colors with 24h TTL caching.
// Colors are static API configuration data that never changes.
func GetCalendarColorsCached(ctx context.Context, c *CalendarClient) (map[string]string, error) {
	return calendarColorsCache.GetOrFetch(ctx, func(ctx context.Context) (map[string]string, error) {
		return c.GetColors(ctx)
	})
}

// GetTaskListsCached returns Google Tasks lists with 30s TTL caching.
func GetTaskListsCached(ctx context.Context, c *TasksClient) ([]TaskList, error) {
	return taskListsCache.GetOrFetch(ctx, func(ctx context.Context) ([]TaskList, error) {
		return c.ListTaskLists(ctx)
	})
}

// InvalidateGoogleCaches clears all Google service caches.
func InvalidateGoogleCaches() {
	calendarListCache.Invalidate()
	calendarColorsCache.Invalidate()
	taskListsCache.Invalidate()
}
