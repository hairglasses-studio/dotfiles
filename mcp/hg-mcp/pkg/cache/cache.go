// Package cache re-exports mcpkit/resilience CacheEntry for TTL caching.
package cache

import (
	"time"

	"github.com/hairglasses-studio/mcpkit/resilience"
)

// Entry is a generic TTL cache entry with singleflight.
type Entry[T any] = resilience.CacheEntry[T]

// New creates a cache entry with the given TTL.
func New[T any](ttl time.Duration) *Entry[T] {
	return resilience.NewCache[T](ttl)
}
