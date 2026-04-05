package cache

import (
	"context"
	"testing"
	"time"
)

func TestNew_CreatesNonNil(t *testing.T) {
	c := New[string](1 * time.Minute)
	if c == nil {
		t.Fatal("New returned nil")
	}
}

func TestCacheEntry_GetOrFetch(t *testing.T) {
	c := New[string](1 * time.Minute)
	callCount := 0

	fetch := func(ctx context.Context) (string, error) {
		callCount++
		return "value", nil
	}

	// First call should invoke the fetch function
	val, err := c.GetOrFetch(context.Background(), fetch)
	if err != nil {
		t.Fatalf("first GetOrFetch: %v", err)
	}
	if val != "value" {
		t.Errorf("first val = %q, want %q", val, "value")
	}
	if callCount != 1 {
		t.Errorf("callCount = %d, want 1 after first fetch", callCount)
	}

	// Second call should return cached value
	val, err = c.GetOrFetch(context.Background(), fetch)
	if err != nil {
		t.Fatalf("second GetOrFetch: %v", err)
	}
	if val != "value" {
		t.Errorf("second val = %q, want %q", val, "value")
	}
	if callCount != 1 {
		t.Errorf("callCount = %d, want 1 (should be cached)", callCount)
	}
}

func TestCacheEntry_Expiry(t *testing.T) {
	c := New[int](50 * time.Millisecond)
	callCount := 0

	fetch := func(ctx context.Context) (int, error) {
		callCount++
		return callCount, nil
	}

	val, _ := c.GetOrFetch(context.Background(), fetch)
	if val != 1 {
		t.Errorf("first val = %d, want 1", val)
	}

	// Wait for cache to expire
	time.Sleep(60 * time.Millisecond)

	val, _ = c.GetOrFetch(context.Background(), fetch)
	if val != 2 {
		t.Errorf("after expiry val = %d, want 2 (re-fetched)", val)
	}
}
