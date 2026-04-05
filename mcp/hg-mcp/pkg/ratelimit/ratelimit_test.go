package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestGet_ReturnsNonNil(t *testing.T) {
	limiter := Get("test-service")
	if limiter == nil {
		t.Fatal("Get returned nil")
	}
}

func TestGet_SameServiceReturnsSame(t *testing.T) {
	a := Get("svc-a")
	b := Get("svc-a")
	// They should be the same limiter instance
	if a != b {
		t.Error("Get should return the same limiter for the same service name")
	}
}

func TestGet_DifferentServicesReturnDifferent(t *testing.T) {
	a := Get("svc-x")
	b := Get("svc-y")
	if a == b {
		t.Error("Get should return different limiters for different services")
	}
}

func TestConfigure_DoesNotPanic(t *testing.T) {
	// Configure should work without panicking
	Configure("test-configured", 10.0, 5)

	limiter := Get("test-configured")
	if limiter == nil {
		t.Fatal("configured limiter is nil")
	}
}

func TestLimiter_Wait(t *testing.T) {
	Configure("test-wait", 100.0, 10)
	limiter := Get("test-wait")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Should not block with a high rate limit
	err := limiter.Wait(ctx)
	if err != nil {
		t.Errorf("Wait returned error with generous rate limit: %v", err)
	}
}
