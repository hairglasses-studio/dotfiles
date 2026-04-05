package sync

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
	service    string
}

// RateLimitConfig configures rate limiting
type RateLimitConfig struct {
	RequestsPerMinute int // Maximum requests per minute
	BurstSize         int // Maximum burst size (bucket capacity)
}

// DefaultRateLimitConfig returns sensible defaults
func DefaultRateLimitConfig() RateLimitConfig {
	return RateLimitConfig{
		RequestsPerMinute: 60, // 1 per second
		BurstSize:         10, // Allow bursts of 10
	}
}

// Prometheus metrics for rate limiting
var (
	RateLimitWaits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sync_rate_limit_waits_total",
			Help: "Total number of rate limit waits",
		},
		[]string{"service"},
	)

	RateLimitDelaySeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sync_rate_limit_delay_seconds",
			Help:    "Duration of rate limit delays",
			Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 2, 5, 10},
		},
		[]string{"service"},
	)
)

// NewRateLimiter creates a new rate limiter for a service
func NewRateLimiter(service string, config RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		tokens:     float64(config.BurstSize),
		maxTokens:  float64(config.BurstSize),
		refillRate: float64(config.RequestsPerMinute) / 60.0, // Convert to per-second
		lastRefill: time.Now(),
		service:    service,
	}
}

// Wait blocks until a token is available or context is cancelled
func (r *RateLimiter) Wait(ctx context.Context) error {
	for {
		r.mu.Lock()
		r.refill()

		if r.tokens >= 1 {
			r.tokens--
			r.mu.Unlock()
			return nil
		}

		// Calculate wait time for next token
		tokensNeeded := 1 - r.tokens
		waitTime := time.Duration(tokensNeeded / r.refillRate * float64(time.Second))
		r.mu.Unlock()

		// Record metrics
		RateLimitWaits.WithLabelValues(r.service).Inc()
		RateLimitDelaySeconds.WithLabelValues(r.service).Observe(waitTime.Seconds())

		// Wait for token or context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
			// Try again
		}
	}
}

// TryAcquire attempts to acquire a token without waiting
// Returns true if successful, false if rate limited
func (r *RateLimiter) TryAcquire() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.refill()

	if r.tokens >= 1 {
		r.tokens--
		return true
	}
	return false
}

// refill adds tokens based on elapsed time (must be called with lock held)
func (r *RateLimiter) refill() {
	now := time.Now()
	elapsed := now.Sub(r.lastRefill).Seconds()
	r.lastRefill = now

	r.tokens += elapsed * r.refillRate
	if r.tokens > r.maxTokens {
		r.tokens = r.maxTokens
	}
}

// Available returns the current number of available tokens
func (r *RateLimiter) Available() float64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.refill()
	return r.tokens
}

// RateLimiterRegistry manages rate limiters for multiple services
type RateLimiterRegistry struct {
	mu       sync.RWMutex
	limiters map[string]*RateLimiter
	configs  map[string]RateLimitConfig
}

// NewRateLimiterRegistry creates a new registry
func NewRateLimiterRegistry() *RateLimiterRegistry {
	return &RateLimiterRegistry{
		limiters: make(map[string]*RateLimiter),
		configs:  make(map[string]RateLimitConfig),
	}
}

// Configure sets the rate limit config for a service
func (r *RateLimiterRegistry) Configure(service string, config RateLimitConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.configs[service] = config
	// Reset the limiter with new config
	r.limiters[service] = NewRateLimiter(service, config)
}

// Get returns the rate limiter for a service, creating one if needed
func (r *RateLimiterRegistry) Get(service string) *RateLimiter {
	r.mu.RLock()
	limiter, exists := r.limiters[service]
	r.mu.RUnlock()

	if exists {
		return limiter
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists := r.limiters[service]; exists {
		return limiter
	}

	// Create with default config or existing config
	config, hasConfig := r.configs[service]
	if !hasConfig {
		config = DefaultRateLimitConfig()
	}

	limiter = NewRateLimiter(service, config)
	r.limiters[service] = limiter
	return limiter
}

// Wait waits for rate limit on a specific service
func (r *RateLimiterRegistry) Wait(ctx context.Context, service string) error {
	return r.Get(service).Wait(ctx)
}

// Global rate limiter registry
var GlobalRateLimiters = NewRateLimiterRegistry()

// RateLimitedFunc wraps a function with rate limiting
func RateLimited[T any](ctx context.Context, service string, fn func() (T, error)) (T, error) {
	var zero T
	if err := GlobalRateLimiters.Wait(ctx, service); err != nil {
		return zero, WrapRateLimit(err)
	}
	return fn()
}
