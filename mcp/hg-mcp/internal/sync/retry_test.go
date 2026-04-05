package sync

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetry_Success(t *testing.T) {
	attempts := 0
	err := Retry(context.Background(), "test", "op", DefaultRetryConfig(), func(ctx context.Context) error {
		attempts++
		return nil
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt, got %d", attempts)
	}
}

func TestRetry_SuccessAfterRetries(t *testing.T) {
	attempts := 0
	config := RetryConfig{
		MaxRetries:   3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       0,
	}
	err := Retry(context.Background(), "test", "op", config, func(ctx context.Context) error {
		attempts++
		if attempts < 3 {
			return WrapRetriable(errors.New("temporary error"))
		}
		return nil
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetry_ExhaustedRetries(t *testing.T) {
	attempts := 0
	config := RetryConfig{
		MaxRetries:   2,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       0,
	}
	err := Retry(context.Background(), "test", "op", config, func(ctx context.Context) error {
		attempts++
		return WrapRetriable(errors.New("always fails"))
	})
	if err == nil {
		t.Error("expected error after exhausting retries")
	}
	if attempts != 3 { // Initial + 2 retries
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}

func TestRetry_PermanentError(t *testing.T) {
	attempts := 0
	err := Retry(context.Background(), "test", "op", DefaultRetryConfig(), func(ctx context.Context) error {
		attempts++
		return WrapPermanent(errors.New("permanent failure"))
	})
	if err == nil {
		t.Error("expected error on permanent failure")
	}
	if attempts != 1 {
		t.Errorf("expected 1 attempt for permanent error, got %d", attempts)
	}
}

func TestRetry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0
	config := RetryConfig{
		MaxRetries:   5,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       0,
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	err := Retry(ctx, "test", "op", config, func(ctx context.Context) error {
		attempts++
		return WrapRetriable(errors.New("retryable"))
	})

	if err == nil {
		t.Error("expected error on context cancellation")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
}

func TestRetryWithResult_Success(t *testing.T) {
	result, err := RetryWithResult(context.Background(), "test", "op", DefaultRetryConfig(), func(ctx context.Context) (int, error) {
		return 42, nil
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != 42 {
		t.Errorf("expected result 42, got %d", result)
	}
}

func TestRetryWithResult_SuccessAfterRetries(t *testing.T) {
	attempts := 0
	config := RetryConfig{
		MaxRetries:   3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       0,
	}
	result, err := RetryWithResult(context.Background(), "test", "op", config, func(ctx context.Context) (string, error) {
		attempts++
		if attempts < 2 {
			return "", WrapRetriable(errors.New("temporary"))
		}
		return "success", nil
	})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "success" {
		t.Errorf("expected 'success', got '%s'", result)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func TestIsRetriable(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"retriable error", WrapRetriable(errors.New("temp")), true},
		{"permanent error", WrapPermanent(errors.New("perm")), false},
		{"timeout error", WrapTimeout(errors.New("timeout")), true},
		{"rate limit error", WrapRateLimit(errors.New("rate")), true},
		{"not found error", errors.New("not found"), true}, // Unknown errors default to retriable
		{"context canceled", context.Canceled, false},
		{"context deadline", context.DeadlineExceeded, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsRetriable(tc.err)
			if result != tc.expected {
				t.Errorf("IsRetriable(%v) = %v, expected %v", tc.err, result, tc.expected)
			}
		})
	}
}

func TestCalculateDelay(t *testing.T) {
	config := RetryConfig{
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       0,
	}

	// Test exponential backoff
	delays := []time.Duration{
		1 * time.Second,  // attempt 0
		2 * time.Second,  // attempt 1
		4 * time.Second,  // attempt 2
		8 * time.Second,  // attempt 3
		16 * time.Second, // attempt 4
		30 * time.Second, // attempt 5 (capped at maxDelay)
	}

	for attempt, expected := range delays {
		actual := calculateDelay(attempt, config)
		if actual != expected {
			t.Errorf("attempt %d: expected %v, got %v", attempt, expected, actual)
		}
	}
}

func TestCalculateDelay_WithJitter(t *testing.T) {
	config := RetryConfig{
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.1, // 10% jitter
	}

	// Run multiple times to test jitter range
	for i := 0; i < 10; i++ {
		delay := calculateDelay(0, config)
		// With 10% jitter, delay should be between 0.9s and 1.1s
		if delay < 900*time.Millisecond || delay > 1100*time.Millisecond {
			t.Errorf("delay %v outside expected jitter range [900ms, 1100ms]", delay)
		}
	}
}
