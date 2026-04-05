package sync

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Circuit breaker states
type CircuitState int

const (
	CircuitClosed   CircuitState = iota // Normal operation
	CircuitOpen                         // Failing fast
	CircuitHalfOpen                     // Testing recovery
)

func (s CircuitState) String() string {
	switch s {
	case CircuitClosed:
		return "closed"
	case CircuitOpen:
		return "open"
	case CircuitHalfOpen:
		return "half_open"
	default:
		return "unknown"
	}
}

// Circuit breaker errors
var (
	ErrCircuitOpen = errors.New("circuit breaker is open")
)

// CircuitBreakerConfig configures circuit breaker behavior
type CircuitBreakerConfig struct {
	FailureThreshold int           // Number of failures before opening
	SuccessThreshold int           // Number of successes in half-open before closing
	Timeout          time.Duration // Time to wait before transitioning from open to half-open
	HalfOpenMaxCalls int           // Max concurrent calls in half-open state
}

// DefaultCircuitBreakerConfig returns sensible defaults
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		FailureThreshold: 5,
		SuccessThreshold: 2,
		Timeout:          30 * time.Second,
		HalfOpenMaxCalls: 1,
	}
}

// Prometheus metrics for circuit breaker
var (
	CircuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sync_circuit_breaker_state",
			Help: "Circuit breaker state (0=closed, 1=open, 2=half_open)",
		},
		[]string{"service"},
	)

	CircuitBreakerTransitions = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sync_circuit_breaker_transitions_total",
			Help: "Total circuit breaker state transitions",
		},
		[]string{"service", "from", "to"},
	)

	CircuitBreakerRejections = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sync_circuit_breaker_rejections_total",
			Help: "Total requests rejected by open circuit",
		},
		[]string{"service"},
	)
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	name          string
	config        CircuitBreakerConfig
	mu            sync.RWMutex
	state         CircuitState
	failures      int
	successes     int
	lastFailure   time.Time
	halfOpenCalls int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(name string, config CircuitBreakerConfig) *CircuitBreaker {
	cb := &CircuitBreaker{
		name:   name,
		config: config,
		state:  CircuitClosed,
	}
	CircuitBreakerState.WithLabelValues(name).Set(float64(CircuitClosed))
	return cb
}

// Execute runs the function through the circuit breaker
func (cb *CircuitBreaker) Execute(ctx context.Context, fn func(context.Context) error) error {
	if !cb.canExecute() {
		CircuitBreakerRejections.WithLabelValues(cb.name).Inc()
		return fmt.Errorf("%s: %w", cb.name, ErrCircuitOpen)
	}

	err := fn(ctx)
	cb.recordResult(err)
	return err
}

// ExecuteWithResult runs a function that returns a result through the circuit breaker
func ExecuteWithResult[T any](cb *CircuitBreaker, ctx context.Context, fn func(context.Context) (T, error)) (T, error) {
	var zero T
	if !cb.canExecute() {
		CircuitBreakerRejections.WithLabelValues(cb.name).Inc()
		return zero, fmt.Errorf("%s: %w", cb.name, ErrCircuitOpen)
	}

	result, err := fn(ctx)
	cb.recordResult(err)
	return result, err
}

// canExecute checks if a request can proceed
func (cb *CircuitBreaker) canExecute() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true

	case CircuitOpen:
		// Check if timeout has passed to transition to half-open
		if time.Since(cb.lastFailure) >= cb.config.Timeout {
			cb.transition(CircuitHalfOpen)
			cb.halfOpenCalls = 1
			return true
		}
		return false

	case CircuitHalfOpen:
		// Allow limited concurrent calls in half-open
		if cb.halfOpenCalls < cb.config.HalfOpenMaxCalls {
			cb.halfOpenCalls++
			return true
		}
		return false
	}

	return false
}

// recordResult records the result of an execution
func (cb *CircuitBreaker) recordResult(err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.recordFailure()
	} else {
		cb.recordSuccess()
	}
}

// recordFailure handles a failed execution
func (cb *CircuitBreaker) recordFailure() {
	cb.failures++
	cb.successes = 0
	cb.lastFailure = time.Now()

	switch cb.state {
	case CircuitClosed:
		if cb.failures >= cb.config.FailureThreshold {
			cb.transition(CircuitOpen)
		}
	case CircuitHalfOpen:
		// Any failure in half-open returns to open
		cb.transition(CircuitOpen)
	}
}

// recordSuccess handles a successful execution
func (cb *CircuitBreaker) recordSuccess() {
	switch cb.state {
	case CircuitClosed:
		cb.failures = 0
	case CircuitHalfOpen:
		cb.successes++
		if cb.successes >= cb.config.SuccessThreshold {
			cb.transition(CircuitClosed)
		}
	}
}

// transition changes the circuit breaker state
func (cb *CircuitBreaker) transition(newState CircuitState) {
	oldState := cb.state
	cb.state = newState
	cb.failures = 0
	cb.successes = 0
	cb.halfOpenCalls = 0

	// Record metrics
	CircuitBreakerState.WithLabelValues(cb.name).Set(float64(newState))
	CircuitBreakerTransitions.WithLabelValues(cb.name, oldState.String(), newState.String()).Inc()
}

// State returns the current circuit breaker state
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	if cb.state != CircuitClosed {
		cb.transition(CircuitClosed)
	}
}

// CircuitBreakerRegistry manages multiple circuit breakers
type CircuitBreakerRegistry struct {
	mu       sync.RWMutex
	breakers map[string]*CircuitBreaker
	config   CircuitBreakerConfig
}

// NewCircuitBreakerRegistry creates a new registry with default config
func NewCircuitBreakerRegistry() *CircuitBreakerRegistry {
	return &CircuitBreakerRegistry{
		breakers: make(map[string]*CircuitBreaker),
		config:   DefaultCircuitBreakerConfig(),
	}
}

// Get returns or creates a circuit breaker for the given service
func (r *CircuitBreakerRegistry) Get(service string) *CircuitBreaker {
	r.mu.RLock()
	cb, exists := r.breakers[service]
	r.mu.RUnlock()

	if exists {
		return cb
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Double-check after acquiring write lock
	if cb, exists = r.breakers[service]; exists {
		return cb
	}

	cb = NewCircuitBreaker(service, r.config)
	r.breakers[service] = cb
	return cb
}

// Configure sets the config for a specific service
func (r *CircuitBreakerRegistry) Configure(service string, config CircuitBreakerConfig) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if cb, exists := r.breakers[service]; exists {
		cb.config = config
	} else {
		r.breakers[service] = NewCircuitBreaker(service, config)
	}
}

// Status returns the state of all circuit breakers
func (r *CircuitBreakerRegistry) Status() map[string]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	status := make(map[string]string)
	for name, cb := range r.breakers {
		status[name] = cb.State().String()
	}
	return status
}

// GlobalCircuitBreakers is the default registry
var GlobalCircuitBreakers = NewCircuitBreakerRegistry()
