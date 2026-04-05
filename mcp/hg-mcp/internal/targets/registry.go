package targets

import (
	"context"
	"fmt"
	"strings"
	"sync"
)

// Registry manages output target instances and their lifecycle.
type Registry struct {
	mu       sync.RWMutex
	targets  map[string]OutputTarget
	bus      *FeedbackBus
	actions  *ActionRegistry
}

// NewRegistry creates a new target registry.
func NewRegistry() *Registry {
	return &Registry{
		targets: make(map[string]OutputTarget),
		bus:     NewFeedbackBus(),
		actions: NewActionRegistry(),
	}
}

// Register adds an output target. Does NOT connect it.
func (r *Registry) Register(target OutputTarget) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := target.ID()
	if _, exists := r.targets[id]; exists {
		return fmt.Errorf("target %q already registered", id)
	}
	r.targets[id] = target
	return nil
}

// Unregister disconnects and removes a target.
func (r *Registry) Unregister(ctx context.Context, id string) error {
	r.mu.Lock()
	target, ok := r.targets[id]
	if !ok {
		r.mu.Unlock()
		return fmt.Errorf("target %q not found", id)
	}
	delete(r.targets, id)
	r.mu.Unlock()

	r.actions.RemoveTarget(id)
	return target.Disconnect(ctx)
}

// Get returns a target by ID.
func (r *Registry) Get(id string) (OutputTarget, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.targets[id]
	return t, ok
}

// List returns all registered target IDs with their health.
func (r *Registry) List(ctx context.Context) map[string]TargetHealth {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]TargetHealth, len(r.targets))
	for id, t := range r.targets {
		result[id] = t.Health(ctx)
	}
	return result
}

// Connect connects a specific target.
func (r *Registry) Connect(ctx context.Context, id string) error {
	r.mu.RLock()
	target, ok := r.targets[id]
	r.mu.RUnlock()

	if !ok {
		return fmt.Errorf("target %q not found", id)
	}
	if err := target.Connect(ctx); err != nil {
		return err
	}

	// Rebuild action index for this target.
	r.actions.IndexTarget(id, target.Actions(ctx))
	return nil
}

// ConnectAll connects all registered targets concurrently.
func (r *Registry) ConnectAll(ctx context.Context) error {
	r.mu.RLock()
	targets := make(map[string]OutputTarget, len(r.targets))
	for id, t := range r.targets {
		targets[id] = t
	}
	r.mu.RUnlock()

	var wg sync.WaitGroup
	for id, t := range targets {
		wg.Add(1)
		go func(id string, t OutputTarget) {
			defer wg.Done()
			if err := t.Connect(ctx); err == nil {
				r.actions.IndexTarget(id, t.Actions(ctx))
			}
		}(id, t)
	}
	wg.Wait()
	return nil
}

// Execute dispatches an action to the appropriate target.
func (r *Registry) Execute(ctx context.Context, targetID, actionID string, params map[string]any) (*ActionResult, error) {
	r.mu.RLock()
	target, ok := r.targets[targetID]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("target %q not found", targetID)
	}
	return target.Execute(ctx, actionID, params)
}

// AllActions returns all actions across all connected targets.
func (r *Registry) AllActions() []QualifiedAction {
	return r.actions.All()
}

// SearchActions finds actions matching a query.
func (r *Registry) SearchActions(query string) []QualifiedAction {
	return r.actions.Search(query)
}

// FeedbackBus returns the centralized feedback bus.
func (r *Registry) FeedbackBus() *FeedbackBus {
	return r.bus
}

// ---------------------------------------------------------------------------
// Action registry
// ---------------------------------------------------------------------------

// ActionRegistry indexes all actions from all targets for fast lookup.
type ActionRegistry struct {
	mu      sync.RWMutex
	actions map[string]*QualifiedAction // "obs.scene_switch" -> action
	byTag   map[string][]string         // "audio" -> ["obs.source_volume", ...]
}

// NewActionRegistry creates a new action registry.
func NewActionRegistry() *ActionRegistry {
	return &ActionRegistry{
		actions: make(map[string]*QualifiedAction),
		byTag:   make(map[string][]string),
	}
}

// IndexTarget adds or replaces all actions from a target.
func (r *ActionRegistry) IndexTarget(targetID string, actions []ActionDescriptor) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Remove old entries for this target.
	r.removeTargetLocked(targetID)

	// Add new entries.
	for _, a := range actions {
		fullID := targetID + "." + a.ID
		qa := &QualifiedAction{
			TargetID: targetID,
			Action:   a,
			FullID:   fullID,
		}
		r.actions[fullID] = qa
		for _, tag := range a.Tags {
			r.byTag[tag] = append(r.byTag[tag], fullID)
		}
	}
}

// RemoveTarget removes all actions from a target.
func (r *ActionRegistry) RemoveTarget(targetID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.removeTargetLocked(targetID)
}

func (r *ActionRegistry) removeTargetLocked(targetID string) {
	prefix := targetID + "."
	for fullID := range r.actions {
		if strings.HasPrefix(fullID, prefix) {
			delete(r.actions, fullID)
		}
	}
	for tag, ids := range r.byTag {
		var filtered []string
		for _, id := range ids {
			if !strings.HasPrefix(id, prefix) {
				filtered = append(filtered, id)
			}
		}
		if len(filtered) == 0 {
			delete(r.byTag, tag)
		} else {
			r.byTag[tag] = filtered
		}
	}
}

// All returns all registered actions.
func (r *ActionRegistry) All() []QualifiedAction {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]QualifiedAction, 0, len(r.actions))
	for _, qa := range r.actions {
		result = append(result, *qa)
	}
	return result
}

// Search finds actions matching a query string (fuzzy name/tag/description).
func (r *ActionRegistry) Search(query string) []QualifiedAction {
	r.mu.RLock()
	defer r.mu.RUnlock()

	lower := strings.ToLower(query)
	var result []QualifiedAction
	for _, qa := range r.actions {
		if strings.Contains(strings.ToLower(qa.Action.Name), lower) ||
			strings.Contains(strings.ToLower(qa.Action.Description), lower) ||
			strings.Contains(strings.ToLower(qa.FullID), lower) {
			result = append(result, *qa)
		}
	}
	return result
}

// ByTag returns actions tagged with the given tag.
func (r *ActionRegistry) ByTag(tag string) []QualifiedAction {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := r.byTag[tag]
	result := make([]QualifiedAction, 0, len(ids))
	for _, id := range ids {
		if qa, ok := r.actions[id]; ok {
			result = append(result, *qa)
		}
	}
	return result
}
