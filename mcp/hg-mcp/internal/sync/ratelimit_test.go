package sync

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestRateLimiter_TryAcquire(t *testing.T) {
	limiter := NewRateLimiter("test", RateLimitConfig{
		RequestsPerMinute: 60,
		BurstSize:         5,
	})

	// Should be able to acquire burst size tokens
	for i := 0; i < 5; i++ {
		if !limiter.TryAcquire() {
			t.Errorf("failed to acquire token %d", i)
		}
	}

	// Should fail after exhausting burst
	if limiter.TryAcquire() {
		t.Error("expected TryAcquire to fail after burst exhausted")
	}
}

func TestRateLimiter_Wait(t *testing.T) {
	limiter := NewRateLimiter("test", RateLimitConfig{
		RequestsPerMinute: 6000, // 100 per second
		BurstSize:         1,
	})

	ctx := context.Background()

	// Exhaust the token
	if err := limiter.Wait(ctx); err != nil {
		t.Errorf("first Wait failed: %v", err)
	}

	// Next wait should block briefly
	start := time.Now()
	if err := limiter.Wait(ctx); err != nil {
		t.Errorf("second Wait failed: %v", err)
	}
	elapsed := time.Since(start)

	// Should have waited approximately 10ms (1/100 second)
	if elapsed < 5*time.Millisecond {
		t.Errorf("Wait returned too quickly: %v", elapsed)
	}
}

func TestRateLimiter_WaitContextCancellation(t *testing.T) {
	limiter := NewRateLimiter("test", RateLimitConfig{
		RequestsPerMinute: 1, // Very slow
		BurstSize:         1,
	})

	ctx, cancel := context.WithCancel(context.Background())

	// Exhaust the token
	limiter.Wait(context.Background())

	// Cancel context after short delay
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	err := limiter.Wait(ctx)
	if err == nil {
		t.Error("expected error on cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestRateLimiter_Refill(t *testing.T) {
	limiter := NewRateLimiter("test", RateLimitConfig{
		RequestsPerMinute: 6000, // 100 per second
		BurstSize:         10,
	})

	// Exhaust all tokens
	for i := 0; i < 10; i++ {
		limiter.TryAcquire()
	}

	// Wait for refill (should refill ~10 tokens in 100ms)
	time.Sleep(100 * time.Millisecond)

	// Should have refilled
	available := limiter.Available()
	if available < 8 || available > 12 { // Allow some timing variance
		t.Errorf("expected ~10 tokens after refill, got %v", available)
	}
}

func TestRateLimiterRegistry(t *testing.T) {
	registry := NewRateLimiterRegistry()

	// Configure a service
	registry.Configure("service1", RateLimitConfig{
		RequestsPerMinute: 60,
		BurstSize:         5,
	})

	// Get the limiter
	limiter1 := registry.Get("service1")
	if limiter1 == nil {
		t.Error("expected limiter for service1")
	}

	// Get same limiter again
	limiter2 := registry.Get("service1")
	if limiter1 != limiter2 {
		t.Error("expected same limiter instance")
	}

	// Get limiter for unconfigured service (should use defaults)
	limiter3 := registry.Get("service2")
	if limiter3 == nil {
		t.Error("expected default limiter for service2")
	}
}

func TestRateLimiterRegistry_Wait(t *testing.T) {
	registry := NewRateLimiterRegistry()
	registry.Configure("test", RateLimitConfig{
		RequestsPerMinute: 60,
		BurstSize:         2,
	})

	ctx := context.Background()

	// Should succeed for burst
	if err := registry.Wait(ctx, "test"); err != nil {
		t.Errorf("Wait 1 failed: %v", err)
	}
	if err := registry.Wait(ctx, "test"); err != nil {
		t.Errorf("Wait 2 failed: %v", err)
	}
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	limiter := NewRateLimiter("test", RateLimitConfig{
		RequestsPerMinute: 6000,
		BurstSize:         100,
	})

	var wg sync.WaitGroup
	acquired := make(chan bool, 200)

	// Start 10 goroutines trying to acquire tokens
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				acquired <- limiter.TryAcquire()
			}
		}()
	}

	wg.Wait()
	close(acquired)

	// Count successful acquisitions
	success := 0
	for got := range acquired {
		if got {
			success++
		}
	}

	// Should have acquired exactly 100 (burst size)
	if success != 100 {
		t.Errorf("expected 100 successful acquisitions, got %d", success)
	}
}

func TestGlobalRateLimiters(t *testing.T) {
	// Global registry should be initialized
	if GlobalRateLimiters == nil {
		t.Error("GlobalRateLimiters is nil")
	}

	// Should be able to get limiters
	limiter := GlobalRateLimiters.Get("global-test")
	if limiter == nil {
		t.Error("expected limiter from global registry")
	}
}

func TestRateLimited(t *testing.T) {
	// Configure a test service
	GlobalRateLimiters.Configure("rate-test", RateLimitConfig{
		RequestsPerMinute: 60,
		BurstSize:         5,
	})

	calls := 0
	ctx := context.Background()

	// Should succeed
	result, err := RateLimited(ctx, "rate-test", func() (int, error) {
		calls++
		return 42, nil
	})

	if err != nil {
		t.Errorf("RateLimited failed: %v", err)
	}
	if result != 42 {
		t.Errorf("expected 42, got %d", result)
	}
	if calls != 1 {
		t.Errorf("expected 1 call, got %d", calls)
	}
}
