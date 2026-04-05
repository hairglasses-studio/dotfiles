// Package targets provides a unified output target abstraction for routing
// controller events to software applications (OBS, Resolume, Ableton, etc.).
//
// Each target implements the OutputTarget interface, with optional
// FeedbackTarget and DiscoverableTarget extensions for bidirectional
// communication and runtime parameter discovery.
package targets

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// Core interfaces
// ---------------------------------------------------------------------------

// OutputTarget is the core interface all software output targets implement.
type OutputTarget interface {
	// Identity
	ID() string       // Unique target ID: "obs", "resolume", "shell"
	Name() string     // Human-readable: "OBS Studio", "Resolume Arena"
	Protocol() string // "websocket", "osc", "http", "shell"

	// Connection lifecycle
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	Health(ctx context.Context) TargetHealth

	// Action discovery
	Actions(ctx context.Context) []ActionDescriptor

	// Execution
	Execute(ctx context.Context, actionID string, params map[string]any) (*ActionResult, error)

	// State queries
	State(ctx context.Context, path string) (*StateValue, error)
}

// FeedbackTarget extends OutputTarget with bidirectional feedback support.
type FeedbackTarget interface {
	OutputTarget
	Subscribe(ctx context.Context, paths []string, handler FeedbackHandler) (SubscriptionID, error)
	Unsubscribe(id SubscriptionID) error
}

// DiscoverableTarget extends OutputTarget with runtime parameter discovery.
type DiscoverableTarget interface {
	OutputTarget
	Discover(ctx context.Context) (*DiscoveryResult, error)
}

// ---------------------------------------------------------------------------
// Action types
// ---------------------------------------------------------------------------

// ActionType classifies how an action behaves.
type ActionType string

const (
	ActionTrigger  ActionType = "trigger"   // Fire and forget
	ActionToggle   ActionType = "toggle"    // On/off state
	ActionSetValue ActionType = "set_value" // Continuous parameter
	ActionGetValue ActionType = "get_value" // Query current value
	ActionSelect   ActionType = "select"    // Choose from list
)

// ActionDescriptor describes a single controllable action on a target.
type ActionDescriptor struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Category    string            `json:"category,omitempty"`
	Type        ActionType        `json:"type"`
	Parameters  []ParamDescriptor `json:"parameters,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
}

// ParamDescriptor describes a parameter for an action.
type ParamDescriptor struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`                     // "string", "number", "boolean", "select"
	Description string   `json:"description,omitempty"`
	Required    bool     `json:"required,omitempty"`
	Default     any      `json:"default,omitempty"`
	Min         *float64 `json:"min,omitempty"`
	Max         *float64 `json:"max,omitempty"`
	Step        *float64 `json:"step,omitempty"`
	Options     []string `json:"options,omitempty"`
}

// ActionResult is what Execute returns.
type ActionResult struct {
	Success bool           `json:"success"`
	Data    map[string]any `json:"data,omitempty"`
	Error   string         `json:"error,omitempty"`
}

// ---------------------------------------------------------------------------
// Health and state
// ---------------------------------------------------------------------------

// TargetHealth reports the health of an output target connection.
type TargetHealth struct {
	Connected bool          `json:"connected"`
	Latency   time.Duration `json:"latency_ms"`
	Status    string        `json:"status"` // "healthy", "degraded", "disconnected"
	Issues    []string      `json:"issues,omitempty"`
}

// StateValue represents a current parameter value from a target.
type StateValue struct {
	Path      string    `json:"path"`
	Value     any       `json:"value"`
	Type      string    `json:"type"` // "number", "boolean", "string"
	Timestamp time.Time `json:"timestamp"`
}

// ---------------------------------------------------------------------------
// Feedback
// ---------------------------------------------------------------------------

// FeedbackHandler is called when a target's state changes externally.
type FeedbackHandler func(event FeedbackEvent)

// FeedbackEvent represents an external state change from a target.
type FeedbackEvent struct {
	TargetID  string    `json:"target_id"`
	Path      string    `json:"path"`
	OldValue  any       `json:"old_value,omitempty"`
	NewValue  any       `json:"new_value"`
	Timestamp time.Time `json:"timestamp"`
}

// SubscriptionID identifies a feedback subscription.
type SubscriptionID string

// DiscoveryResult contains what was found when probing a target.
type DiscoveryResult struct {
	Actions      []ActionDescriptor     `json:"actions"`
	StateTree    map[string]*StateValue `json:"state_tree,omitempty"`
	Metadata     map[string]string      `json:"metadata,omitempty"`
	DiscoveredAt time.Time              `json:"discovered_at"`
}

// ---------------------------------------------------------------------------
// Qualified action (includes target prefix)
// ---------------------------------------------------------------------------

// QualifiedAction pairs an action with its target.
type QualifiedAction struct {
	TargetID string           `json:"target_id"`
	Action   ActionDescriptor `json:"action"`
	FullID   string           `json:"full_id"` // "obs.scene_switch"
}

// ---------------------------------------------------------------------------
// Feedback bus
// ---------------------------------------------------------------------------

// FeedbackBus aggregates feedback events from all targets into a single channel.
type FeedbackBus struct {
	mu       sync.RWMutex
	subs     map[SubscriptionID]*feedbackSub
	incoming chan FeedbackEvent
	nextID   int
}

type feedbackSub struct {
	id      SubscriptionID
	filter  FeedbackFilter
	handler FeedbackHandler
}

// FeedbackFilter restricts which events a subscriber receives.
type FeedbackFilter struct {
	TargetIDs  []string `json:"target_ids,omitempty"`
	PathPrefix string   `json:"path_prefix,omitempty"`
}

// NewFeedbackBus creates a new feedback bus.
func NewFeedbackBus() *FeedbackBus {
	fb := &FeedbackBus{
		subs:     make(map[SubscriptionID]*feedbackSub),
		incoming: make(chan FeedbackEvent, 256),
	}
	go fb.dispatch()
	return fb
}

// Publish sends an event to all matching subscribers.
func (fb *FeedbackBus) Publish(event FeedbackEvent) {
	select {
	case fb.incoming <- event:
	default: // Drop if full
	}
}

// Subscribe registers a handler for filtered events.
func (fb *FeedbackBus) Subscribe(filter FeedbackFilter, handler FeedbackHandler) SubscriptionID {
	fb.mu.Lock()
	defer fb.mu.Unlock()
	fb.nextID++
	id := SubscriptionID(json.Number(fmt.Sprintf("%d", fb.nextID)).String())
	fb.subs[id] = &feedbackSub{id: id, filter: filter, handler: handler}
	return id
}

// Unsubscribe removes a subscription.
func (fb *FeedbackBus) Unsubscribe(id SubscriptionID) {
	fb.mu.Lock()
	defer fb.mu.Unlock()
	delete(fb.subs, id)
}

func (fb *FeedbackBus) dispatch() {
	for event := range fb.incoming {
		fb.mu.RLock()
		for _, sub := range fb.subs {
			if matchesFilter(event, sub.filter) {
				sub.handler(event)
			}
		}
		fb.mu.RUnlock()
	}
}

func matchesFilter(event FeedbackEvent, filter FeedbackFilter) bool {
	if len(filter.TargetIDs) > 0 {
		found := false
		for _, id := range filter.TargetIDs {
			if id == event.TargetID {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	if filter.PathPrefix != "" {
		if len(event.Path) < len(filter.PathPrefix) {
			return false
		}
		if event.Path[:len(filter.PathPrefix)] != filter.PathPrefix {
			return false
		}
	}
	return true
}

