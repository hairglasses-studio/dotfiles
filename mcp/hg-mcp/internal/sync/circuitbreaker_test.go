package sync

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// shortConfig returns a CircuitBreakerConfig with very short timeouts for testing.
func shortConfig(failThreshold, successThreshold, halfOpenMax int) CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: failThreshold,
		SuccessThreshold: successThreshold,
		Timeout:          1 * time.Millisecond,
		HalfOpenMaxCalls: halfOpenMax,
	}
}

// openCircuit is a helper that drives a circuit breaker into the Open state.
func openCircuit(t *testing.T, cb *CircuitBreaker, failures int) {
	t.Helper()
	testErr := errors.New("forced failure")
	for i := 0; i < failures; i++ {
		_ = cb.Execute(context.Background(), func(ctx context.Context) error {
			return testErr
		})
	}
	if cb.State() != CircuitOpen {
		t.Fatalf("expected open state after %d failures, got %v", failures, cb.State())
	}
}

// transitionToHalfOpen opens the circuit then waits for the timeout to elapse.
func transitionToHalfOpen(t *testing.T, cb *CircuitBreaker, failures int) {
	t.Helper()
	openCircuit(t, cb, failures)
	time.Sleep(2 * time.Millisecond) // exceeds 1ms timeout
}

// --- 1. Full state machine: Closed -> Open -> Half-Open -> Closed ---

func TestCircuitBreaker_FullStateMachine(t *testing.T) {
	cb := NewCircuitBreaker("fsm", shortConfig(2, 2, 2))

	// Starts Closed
	if cb.State() != CircuitClosed {
		t.Fatalf("expected closed, got %v", cb.State())
	}

	// Closed -> Open after 2 failures
	openCircuit(t, cb, 2)

	// Open -> Half-Open after timeout
	time.Sleep(2 * time.Millisecond)

	// The next Execute triggers the transition to half-open internally
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// After one success, still half-open (need 2 successes)
	if cb.State() != CircuitHalfOpen {
		t.Fatalf("expected half_open after 1 success, got %v", cb.State())
	}

	// Half-Open -> Closed after second success
	err = cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cb.State() != CircuitClosed {
		t.Fatalf("expected closed after 2 successes in half-open, got %v", cb.State())
	}
}

// --- 2. Execute with success stays closed ---

func TestCircuitBreaker_SuccessStaysClosed(t *testing.T) {
	cb := NewCircuitBreaker("stay-closed", shortConfig(3, 2, 1))

	for i := 0; i < 10; i++ {
		err := cb.Execute(context.Background(), func(ctx context.Context) error {
			return nil
		})
		if err != nil {
			t.Errorf("iteration %d: unexpected error: %v", i, err)
		}
		if cb.State() != CircuitClosed {
			t.Errorf("iteration %d: expected closed, got %v", i, cb.State())
		}
	}
}

// --- 3. Repeated failures transition to open ---

func TestCircuitBreaker_RepeatedFailuresOpenCircuit(t *testing.T) {
	cb := NewCircuitBreaker("fail-open", shortConfig(3, 2, 1))
	testErr := errors.New("service down")

	// First 2 failures should keep it closed
	for i := 0; i < 2; i++ {
		_ = cb.Execute(context.Background(), func(ctx context.Context) error {
			return testErr
		})
		if cb.State() != CircuitClosed {
			t.Errorf("should still be closed after %d failures", i+1)
		}
	}

	// 3rd failure should open it
	_ = cb.Execute(context.Background(), func(ctx context.Context) error {
		return testErr
	})
	if cb.State() != CircuitOpen {
		t.Fatalf("expected open after 3 failures, got %v", cb.State())
	}
}

// --- 4. Open circuit rejects calls immediately ---

func TestCircuitBreaker_OpenRejectsImmediately(t *testing.T) {
	cb := NewCircuitBreaker("reject", CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 1,
		Timeout:          1 * time.Hour, // very long timeout so it stays open
		HalfOpenMaxCalls: 1,
	})

	openCircuit(t, cb, 2)

	called := false
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		called = true
		return nil
	})
	if called {
		t.Error("function should not have been called when circuit is open")
	}
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
}

// --- 5. Half-open allows limited calls ---

func TestCircuitBreaker_HalfOpenLimitsCalls(t *testing.T) {
	cb := NewCircuitBreaker("half-limit", shortConfig(2, 3, 1))

	transitionToHalfOpen(t, cb, 2)

	// First call should be allowed (transitions to half-open internally)
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatalf("first half-open call should succeed: %v", err)
	}

	// Second call in half-open should be rejected (max calls = 1, and we
	// already used the slot; need 3 successes to close, only have 1)
	if cb.State() != CircuitHalfOpen {
		t.Fatalf("expected half_open, got %v", cb.State())
	}

	err = cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected rejection in half-open when max calls exceeded, got %v", err)
	}
}

// --- 6. Half-open success transitions back to closed ---

func TestCircuitBreaker_HalfOpenSuccessCloses(t *testing.T) {
	cb := NewCircuitBreaker("ho-close", shortConfig(2, 2, 2))

	transitionToHalfOpen(t, cb, 2)

	// Two successes in half-open should close the circuit
	for i := 0; i < 2; i++ {
		err := cb.Execute(context.Background(), func(ctx context.Context) error {
			return nil
		})
		if err != nil {
			t.Fatalf("success %d: unexpected error: %v", i+1, err)
		}
	}

	if cb.State() != CircuitClosed {
		t.Errorf("expected closed after successes in half-open, got %v", cb.State())
	}
}

// --- 7. Half-open failure returns to open ---

func TestCircuitBreaker_HalfOpenFailureReopens(t *testing.T) {
	cb := NewCircuitBreaker("ho-reopen", shortConfig(2, 2, 1))

	transitionToHalfOpen(t, cb, 2)

	// A failure in half-open should return to open
	_ = cb.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("still broken")
	})

	if cb.State() != CircuitOpen {
		t.Errorf("expected open after failure in half-open, got %v", cb.State())
	}
}

// --- 8. Reset method ---

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker("reset-test", CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 2,
		Timeout:          1 * time.Hour,
		HalfOpenMaxCalls: 1,
	})

	// Open the circuit
	openCircuit(t, cb, 2)

	// Reset should return to closed
	cb.Reset()
	if cb.State() != CircuitClosed {
		t.Errorf("expected closed after reset, got %v", cb.State())
	}

	// Should accept calls again
	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Errorf("unexpected error after reset: %v", err)
	}
}

func TestCircuitBreaker_ResetWhenAlreadyClosed(t *testing.T) {
	cb := NewCircuitBreaker("reset-noop", shortConfig(3, 2, 1))

	if cb.State() != CircuitClosed {
		t.Fatal("should start closed")
	}

	// Reset on already-closed should be a no-op (does not call transition)
	cb.Reset()
	if cb.State() != CircuitClosed {
		t.Error("should still be closed after reset")
	}
}

// --- 9. CircuitBreakerRegistry ---

func TestCircuitBreakerRegistry_GetCreateAndReuse(t *testing.T) {
	reg := NewCircuitBreakerRegistry()

	cb1 := reg.Get("svc-a")
	if cb1 == nil {
		t.Fatal("expected non-nil circuit breaker")
	}

	// Same name returns same instance
	cb2 := reg.Get("svc-a")
	if cb1 != cb2 {
		t.Error("expected same instance for same service name")
	}

	// Different name returns different instance
	cb3 := reg.Get("svc-b")
	if cb1 == cb3 {
		t.Error("expected different instance for different service name")
	}
}

func TestCircuitBreakerRegistry_Configure_ExistingService(t *testing.T) {
	reg := NewCircuitBreakerRegistry()

	// Create a breaker first
	cb := reg.Get("svc-cfg")
	if cb.config.FailureThreshold != DefaultCircuitBreakerConfig().FailureThreshold {
		t.Fatal("expected default config")
	}

	// Reconfigure it
	newCfg := CircuitBreakerConfig{
		FailureThreshold: 10,
		SuccessThreshold: 5,
		Timeout:          2 * time.Second,
		HalfOpenMaxCalls: 3,
	}
	reg.Configure("svc-cfg", newCfg)

	// Verify it was updated (same instance)
	cb2 := reg.Get("svc-cfg")
	if cb2.config.FailureThreshold != 10 {
		t.Errorf("expected updated FailureThreshold=10, got %d", cb2.config.FailureThreshold)
	}
	if cb2.config.SuccessThreshold != 5 {
		t.Errorf("expected updated SuccessThreshold=5, got %d", cb2.config.SuccessThreshold)
	}
}

func TestCircuitBreakerRegistry_Configure_NewService(t *testing.T) {
	reg := NewCircuitBreakerRegistry()

	cfg := CircuitBreakerConfig{
		FailureThreshold: 7,
		SuccessThreshold: 3,
		Timeout:          5 * time.Second,
		HalfOpenMaxCalls: 2,
	}
	reg.Configure("new-svc", cfg)

	cb := reg.Get("new-svc")
	if cb == nil {
		t.Fatal("expected non-nil after Configure")
	}
	if cb.config.FailureThreshold != 7 {
		t.Errorf("expected FailureThreshold=7, got %d", cb.config.FailureThreshold)
	}
}

func TestCircuitBreakerRegistry_Status(t *testing.T) {
	reg := NewCircuitBreakerRegistry()

	reg.Get("alive")
	reg.Get("broken")

	// Open the "broken" one
	cb := reg.Get("broken")
	// Use a config with low threshold to open it
	reg.Configure("broken", shortConfig(1, 1, 1))
	_ = cb.Execute(context.Background(), func(ctx context.Context) error {
		return errors.New("fail")
	})

	status := reg.Status()
	if len(status) != 2 {
		t.Errorf("expected 2 entries, got %d", len(status))
	}
	if status["alive"] != "closed" {
		t.Errorf("expected alive=closed, got %s", status["alive"])
	}
	if status["broken"] != "open" {
		t.Errorf("expected broken=open, got %s", status["broken"])
	}
}

func TestCircuitBreakerRegistry_ConcurrentGet(t *testing.T) {
	reg := NewCircuitBreakerRegistry()
	const goroutines = 50

	var wg sync.WaitGroup
	results := make([]*CircuitBreaker, goroutines)

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(idx int) {
			defer wg.Done()
			results[idx] = reg.Get("concurrent-svc")
		}(i)
	}
	wg.Wait()

	// All goroutines should get the same instance
	for i := 1; i < goroutines; i++ {
		if results[i] != results[0] {
			t.Fatalf("goroutine %d got different instance", i)
		}
	}
}

// --- 10. ExecuteWithResult generic function ---

func TestExecuteWithResult_Success(t *testing.T) {
	cb := NewCircuitBreaker("ewr-ok", shortConfig(3, 2, 1))

	result, err := ExecuteWithResult(cb, context.Background(), func(ctx context.Context) (string, error) {
		return "hello", nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello" {
		t.Errorf("expected 'hello', got %q", result)
	}
	if cb.State() != CircuitClosed {
		t.Errorf("expected closed, got %v", cb.State())
	}
}

func TestExecuteWithResult_Error(t *testing.T) {
	cb := NewCircuitBreaker("ewr-err", shortConfig(5, 2, 1))

	_, err := ExecuteWithResult(cb, context.Background(), func(ctx context.Context) (int, error) {
		return 0, errors.New("compute failed")
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExecuteWithResult_OpenCircuitRejectsWithZeroValue(t *testing.T) {
	cb := NewCircuitBreaker("ewr-open", CircuitBreakerConfig{
		FailureThreshold: 2,
		SuccessThreshold: 1,
		Timeout:          1 * time.Hour,
		HalfOpenMaxCalls: 1,
	})

	// Open the circuit via Execute
	openCircuit(t, cb, 2)

	// ExecuteWithResult should be rejected and return zero value
	result, err := ExecuteWithResult(cb, context.Background(), func(ctx context.Context) (int, error) {
		t.Fatal("should not be called when circuit is open")
		return 99, nil
	})
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen, got %v", err)
	}
	if result != 0 {
		t.Errorf("expected zero value, got %d", result)
	}
}

func TestExecuteWithResult_StructType(t *testing.T) {
	type Data struct {
		Name  string
		Count int
	}

	cb := NewCircuitBreaker("ewr-struct", shortConfig(3, 2, 1))

	result, err := ExecuteWithResult(cb, context.Background(), func(ctx context.Context) (Data, error) {
		return Data{Name: "test", Count: 42}, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "test" || result.Count != 42 {
		t.Errorf("unexpected result: %+v", result)
	}
}

// --- Additional edge cases ---

func TestCircuitState_String(t *testing.T) {
	tests := []struct {
		state CircuitState
		want  string
	}{
		{CircuitClosed, "closed"},
		{CircuitOpen, "open"},
		{CircuitHalfOpen, "half_open"},
		{CircuitState(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.state.String(); got != tt.want {
			t.Errorf("CircuitState(%d).String() = %q, want %q", int(tt.state), got, tt.want)
		}
	}
}

func TestCircuitBreaker_SuccessResetsFailureCount(t *testing.T) {
	// Verify that a success in closed state resets the failure counter
	cb := NewCircuitBreaker("reset-failures", shortConfig(3, 2, 1))
	testErr := errors.New("fail")

	// 2 failures (just below threshold)
	for i := 0; i < 2; i++ {
		_ = cb.Execute(context.Background(), func(ctx context.Context) error {
			return testErr
		})
	}

	// 1 success should reset failures
	_ = cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})

	// 2 more failures should not open (counter was reset)
	for i := 0; i < 2; i++ {
		_ = cb.Execute(context.Background(), func(ctx context.Context) error {
			return testErr
		})
	}

	if cb.State() != CircuitClosed {
		t.Errorf("expected closed (failure counter should have been reset by success), got %v", cb.State())
	}

	// One more failure (now 3 total since last reset) should open it
	_ = cb.Execute(context.Background(), func(ctx context.Context) error {
		return testErr
	})
	if cb.State() != CircuitOpen {
		t.Errorf("expected open after 3 consecutive failures, got %v", cb.State())
	}
}

func TestCircuitBreaker_HalfOpenMaxCallsMultiple(t *testing.T) {
	// Test with HalfOpenMaxCalls > 1
	cb := NewCircuitBreaker("ho-multi", shortConfig(2, 3, 3))

	transitionToHalfOpen(t, cb, 2)

	// Should allow up to 3 calls in half-open
	// The first call is the one that triggers the transition; it counts as halfOpenCalls=1
	for i := 0; i < 3; i++ {
		err := cb.Execute(context.Background(), func(ctx context.Context) error {
			return nil
		})
		if err != nil {
			t.Errorf("call %d should be allowed in half-open: %v", i+1, err)
		}
	}

	// Should now be closed (3 successes >= successThreshold of 3)
	if cb.State() != CircuitClosed {
		t.Errorf("expected closed after 3 successes, got %v", cb.State())
	}
}

func TestCircuitBreaker_ErrorWrapping(t *testing.T) {
	cb := NewCircuitBreaker("wrap-test", CircuitBreakerConfig{
		FailureThreshold: 1,
		SuccessThreshold: 1,
		Timeout:          1 * time.Hour,
		HalfOpenMaxCalls: 1,
	})

	openCircuit(t, cb, 1)

	err := cb.Execute(context.Background(), func(ctx context.Context) error {
		return nil
	})

	// The error should contain the circuit breaker name
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("expected ErrCircuitOpen in chain, got %v", err)
	}
	// Error message should contain the name
	if got := err.Error(); got != "wrap-test: circuit breaker is open" {
		t.Errorf("unexpected error message: %q", got)
	}
}
