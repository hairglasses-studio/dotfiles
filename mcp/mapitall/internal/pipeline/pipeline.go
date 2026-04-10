package pipeline

import (
	"context"
	"log/slog"

	"github.com/hairglasses-studio/mapitall/internal/mapping"
	"github.com/hairglasses-studio/mapitall/internal/output"
	"github.com/hairglasses-studio/mcpkit/device"
)

// RuleResolver finds the best matching rule for a device event.
type RuleResolver func(deviceID string, source string) *mapping.MappingRule

// EventObserver is called for every device event before rule resolution.
// Useful for IPC event streaming and diagnostics.
type EventObserver func(ev device.Event)

// Pipeline processes device events through rule resolution and output dispatch.
type Pipeline struct {
	state           *mapping.EngineState
	registry        *output.Registry
	resolve         RuleResolver
	customModifiers map[string]bool
	observer        EventObserver
}

// New creates a pipeline. customModifiers is the set of input sources that act as
// modifiers (hold-without-emit): they update EngineState instead of dispatching.
func New(state *mapping.EngineState, registry *output.Registry, resolve RuleResolver, customModifiers []string) *Pipeline {
	mods := make(map[string]bool, len(customModifiers))
	for _, m := range customModifiers {
		mods[m] = true
	}
	return &Pipeline{
		state:           state,
		registry:        registry,
		resolve:         resolve,
		customModifiers: mods,
	}
}

// SetObserver registers a callback invoked for every device event.
// Set to nil to remove the observer.
func (p *Pipeline) SetObserver(obs EventObserver) {
	p.observer = obs
}

// ProcessEvents reads events from a channel and dispatches matched rules.
func (p *Pipeline) ProcessEvents(ctx context.Context, events <-chan device.Event) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-events:
			if !ok {
				return
			}
			p.handleEvent(ev)
		}
	}
}

// handleEvent is the hot path: resolve → transform → dispatch.
func (p *Pipeline) handleEvent(ev device.Event) {
	// Notify observer (IPC event streaming, diagnostics).
	if p.observer != nil {
		p.observer(ev)
	}

	// Custom modifier interception: update engine state, don't dispatch.
	if p.customModifiers[ev.Source] {
		p.state.SetModifier(ev.Source, ev.Value > 0.5)
		return
	}

	rule := p.resolve(string(ev.DeviceID), ev.Source)
	if rule == nil {
		return
	}

	// Apply value transform if present.
	value := ev.Value
	if rule.Value != nil {
		value = rule.Value.Transform(value)
	}

	// Dispatch to the output target.
	target := p.registry.Get(rule.Output.Type)
	if target == nil {
		slog.Debug("no output target", "type", rule.Output.Type)
		return
	}

	if err := target.Execute(rule.Output, value); err != nil {
		slog.Warn("output error",
			"type", rule.Output.Type,
			"input", ev.Source,
			"error", err,
		)
	}
}
