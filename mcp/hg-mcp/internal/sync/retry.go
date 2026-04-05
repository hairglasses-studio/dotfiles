package sync

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Error types for categorization
var (
	ErrRetryable = errors.New("retriable error")
	ErrPermanent = errors.New("permanent error")
	ErrTimeout   = errors.New("timeout error")
	ErrRateLimit = errors.New("rate limit exceeded")
	ErrNotFound  = errors.New("resource not found")
)

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxRetries   int           // Maximum number of retry attempts
	InitialDelay time.Duration // Initial delay before first retry
	MaxDelay     time.Duration // Maximum delay between retries
	Multiplier   float64       // Delay multiplier for exponential backoff
	Jitter       float64       // Random jitter factor (0-1)
}

// DefaultRetryConfig returns sensible defaults
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:   3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.1,
	}
}

// Prometheus metrics for retry tracking
var (
	RetryAttempts = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sync_retry_attempts_total",
			Help: "Total number of retry attempts",
		},
		[]string{"service", "operation", "attempt"},
	)

	RetrySuccess = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sync_retry_success_total",
			Help: "Total successful operations after retry",
		},
		[]string{"service", "operation"},
	)

	RetryExhausted = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sync_retry_exhausted_total",
			Help: "Total operations that exhausted all retries",
		},
		[]string{"service", "operation"},
	)
)

// RetryableFunc is a function that can be retried
type RetryableFunc func(ctx context.Context) error

// Retry executes fn with exponential backoff retry logic
func Retry(ctx context.Context, service, operation string, config RetryConfig, fn RetryableFunc) error {
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		// Record attempt metric
		if attempt > 0 {
			RetryAttempts.WithLabelValues(service, operation, fmt.Sprintf("%d", attempt)).Inc()
		}

		// Execute the function
		err := fn(ctx)
		if err == nil {
			if attempt > 0 {
				RetrySuccess.WithLabelValues(service, operation).Inc()
			}
			return nil
		}

		lastErr = err

		// Check if error is retriable
		if !IsRetriable(err) {
			return fmt.Errorf("%s/%s failed (non-retriable): %w", service, operation, err)
		}

		// Check if context is done
		if ctx.Err() != nil {
			return fmt.Errorf("%s/%s cancelled: %w", service, operation, ctx.Err())
		}

		// Don't sleep after the last attempt
		if attempt == config.MaxRetries {
			break
		}

		// Calculate delay with exponential backoff and jitter
		delay := calculateDelay(attempt, config)

		// Wait or context cancellation
		select {
		case <-ctx.Done():
			return fmt.Errorf("%s/%s cancelled during retry wait: %w", service, operation, ctx.Err())
		case <-time.After(delay):
			// Continue to next attempt
		}
	}

	RetryExhausted.WithLabelValues(service, operation).Inc()
	return fmt.Errorf("%s/%s failed after %d retries: %w", service, operation, config.MaxRetries, lastErr)
}

// RetryWithResult executes fn and returns a result with retry logic
func RetryWithResult[T any](ctx context.Context, service, operation string, config RetryConfig, fn func(ctx context.Context) (T, error)) (T, error) {
	var result T
	var lastErr error

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			RetryAttempts.WithLabelValues(service, operation, fmt.Sprintf("%d", attempt)).Inc()
		}

		var err error
		result, err = fn(ctx)
		if err == nil {
			if attempt > 0 {
				RetrySuccess.WithLabelValues(service, operation).Inc()
			}
			return result, nil
		}

		lastErr = err

		if !IsRetriable(err) {
			return result, fmt.Errorf("%s/%s failed (non-retriable): %w", service, operation, err)
		}

		if ctx.Err() != nil {
			return result, fmt.Errorf("%s/%s cancelled: %w", service, operation, ctx.Err())
		}

		if attempt == config.MaxRetries {
			break
		}

		delay := calculateDelay(attempt, config)
		select {
		case <-ctx.Done():
			return result, fmt.Errorf("%s/%s cancelled during retry wait: %w", service, operation, ctx.Err())
		case <-time.After(delay):
		}
	}

	RetryExhausted.WithLabelValues(service, operation).Inc()
	return result, fmt.Errorf("%s/%s failed after %d retries: %w", service, operation, config.MaxRetries, lastErr)
}

// calculateDelay computes the delay for a given attempt using exponential backoff with jitter
func calculateDelay(attempt int, config RetryConfig) time.Duration {
	// Exponential backoff: initialDelay * multiplier^attempt
	delay := float64(config.InitialDelay) * math.Pow(config.Multiplier, float64(attempt))

	// Apply max delay cap
	if delay > float64(config.MaxDelay) {
		delay = float64(config.MaxDelay)
	}

	// Apply jitter: delay * (1 - jitter + random(0, 2*jitter))
	if config.Jitter > 0 {
		jitterRange := delay * config.Jitter * 2
		delay = delay - (delay * config.Jitter) + (rand.Float64() * jitterRange)
	}

	return time.Duration(delay)
}

// IsRetriable checks if an error should be retried
func IsRetriable(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific error types
	if errors.Is(err, ErrPermanent) || errors.Is(err, ErrNotFound) {
		return false
	}

	if errors.Is(err, ErrRetryable) || errors.Is(err, ErrTimeout) || errors.Is(err, ErrRateLimit) {
		return true
	}

	// Check for context errors (not retriable)
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Default: retry on unknown errors
	return true
}

// WrapRetriable wraps an error as retriable
func WrapRetriable(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %v", ErrRetryable, err)
}

// WrapPermanent wraps an error as permanent (non-retriable)
func WrapPermanent(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %v", ErrPermanent, err)
}

// WrapTimeout wraps an error as a timeout error
func WrapTimeout(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %v", ErrTimeout, err)
}

// WrapRateLimit wraps an error as a rate limit error
func WrapRateLimit(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %v", ErrRateLimit, err)
}
