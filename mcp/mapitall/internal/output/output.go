package output

import (
	"sync"

	"github.com/hairglasses-studio/mapping"
)

// Target executes an output action.
type Target interface {
	// Type returns which OutputType this target handles.
	Type() mapping.OutputType

	// Execute performs the output action with the given transformed value.
	Execute(action mapping.OutputAction, value float64) error

	// Close releases any resources.
	Close() error
}

// Registry holds all registered output targets.
type Registry struct {
	mu      sync.RWMutex
	targets map[mapping.OutputType]Target
}

// NewRegistry creates an empty output registry.
func NewRegistry() *Registry {
	return &Registry{
		targets: make(map[mapping.OutputType]Target),
	}
}

// Register adds an output target to the registry.
func (r *Registry) Register(t Target) {
	r.mu.Lock()
	r.targets[t.Type()] = t
	r.mu.Unlock()
}

// Get returns the target for the given output type, or nil.
func (r *Registry) Get(typ mapping.OutputType) Target {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.targets[typ]
}

// Close shuts down all targets.
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, t := range r.targets {
		t.Close()
	}
	return nil
}
