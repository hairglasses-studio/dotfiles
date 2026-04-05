// Package ratelimit re-exports mcpkit/resilience rate limiting with a global registry.
package ratelimit

import (
	"github.com/hairglasses-studio/mcpkit/resilience"
)

// global is the default rate limit registry.
var global = resilience.NewRateLimitRegistry()

// Get returns the rate limiter for the named service.
func Get(service string) *resilience.RateLimiter {
	return global.Get(service)
}

// Configure sets a custom rate limit for a service.
func Configure(service string, rate float64, burst int) {
	global.Configure(service, rate, burst)
}
