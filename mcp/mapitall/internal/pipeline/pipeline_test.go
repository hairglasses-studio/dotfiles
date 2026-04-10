package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/hairglasses-studio/mapitall/internal/mapping"
	"github.com/hairglasses-studio/mapitall/internal/output"
	"github.com/hairglasses-studio/mcpkit/device"
)

type mockTarget struct {
	typ   mapping.OutputType
	calls []executeCall
}

type executeCall struct {
	action mapping.OutputAction
	value  float64
}

func (m *mockTarget) Type() mapping.OutputType { return m.typ }
func (m *mockTarget) Execute(action mapping.OutputAction, value float64) error {
	m.calls = append(m.calls, executeCall{action: action, value: value})
	return nil
}
func (m *mockTarget) Close() error { return nil }

func TestPipelineDispatch(t *testing.T) {
	state := mapping.NewEngineState()
	registry := output.NewRegistry()

	cmdTarget := &mockTarget{typ: mapping.OutputCommand}
	registry.Register(cmdTarget)

	rule := &mapping.MappingRule{
		Input: "BTN_SOUTH",
		Output: mapping.OutputAction{
			Type: mapping.OutputCommand,
			Exec: []string{"echo", "hello"},
		},
	}

	p := New(state, registry, func(deviceID string, source string) *mapping.MappingRule {
		if source == "BTN_SOUTH" {
			return rule
		}
		return nil
	}, nil)

	events := make(chan device.Event, 1)
	events <- device.Event{
		DeviceID: "xbox",
		Source:   "BTN_SOUTH",
		Value:    1.0,
	}
	close(events)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p.ProcessEvents(ctx, events)

	if len(cmdTarget.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(cmdTarget.calls))
	}
	if cmdTarget.calls[0].value != 1.0 {
		t.Errorf("expected value 1.0, got %f", cmdTarget.calls[0].value)
	}
	if cmdTarget.calls[0].action.Exec[0] != "echo" {
		t.Errorf("expected exec[0]='echo', got %q", cmdTarget.calls[0].action.Exec[0])
	}
}

func TestPipelineNoMatch(t *testing.T) {
	state := mapping.NewEngineState()
	registry := output.NewRegistry()

	cmdTarget := &mockTarget{typ: mapping.OutputCommand}
	registry.Register(cmdTarget)

	p := New(state, registry, func(deviceID string, source string) *mapping.MappingRule {
		return nil // never matches
	}, nil)

	events := make(chan device.Event, 1)
	events <- device.Event{
		DeviceID: "xbox",
		Source:   "BTN_NORTH",
		Value:    1.0,
	}
	close(events)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p.ProcessEvents(ctx, events)

	if len(cmdTarget.calls) != 0 {
		t.Fatalf("expected 0 calls for unmatched event, got %d", len(cmdTarget.calls))
	}
}

func TestPipelineValueTransform(t *testing.T) {
	state := mapping.NewEngineState()
	registry := output.NewRegistry()

	oscTarget := &mockTarget{typ: mapping.OutputOSC}
	registry.Register(oscTarget)

	rule := &mapping.MappingRule{
		Input: "ABS_X",
		Output: mapping.OutputAction{
			Type:    mapping.OutputOSC,
			Address: "/test",
		},
		Value: &mapping.ValueTransform{
			InputRange:  [2]float64{0, 65535},
			OutputRange: [2]float64{0, 1},
			Curve:       mapping.CurveLinear,
		},
	}

	p := New(state, registry, func(deviceID string, source string) *mapping.MappingRule {
		if source == "ABS_X" {
			return rule
		}
		return nil
	}, nil)

	events := make(chan device.Event, 1)
	events <- device.Event{
		DeviceID: "xbox",
		Source:   "ABS_X",
		Value:    32767,
	}
	close(events)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p.ProcessEvents(ctx, events)

	if len(oscTarget.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(oscTarget.calls))
	}

	got := oscTarget.calls[0].value
	// 32767/65535 ≈ 0.4999924
	if got < 0.49 || got > 0.51 {
		t.Errorf("expected ~0.5, got %f", got)
	}
}

func TestPipelineCustomModifier(t *testing.T) {
	state := mapping.NewEngineState()
	registry := output.NewRegistry()

	cmdTarget := &mockTarget{typ: mapping.OutputCommand}
	registry.Register(cmdTarget)

	// BTN_TL is a custom modifier — should not dispatch, only update state.
	p := New(state, registry, func(deviceID string, source string) *mapping.MappingRule {
		return &mapping.MappingRule{
			Input:  source,
			Output: mapping.OutputAction{Type: mapping.OutputCommand, Exec: []string{"echo"}},
		}
	}, []string{"BTN_TL"})

	events := make(chan device.Event, 2)
	events <- device.Event{DeviceID: "xbox", Source: "BTN_TL", Value: 1.0}
	events <- device.Event{DeviceID: "xbox", Source: "BTN_SOUTH", Value: 1.0}
	close(events)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p.ProcessEvents(ctx, events)

	// BTN_TL should have been intercepted, BTN_SOUTH should dispatch.
	if len(cmdTarget.calls) != 1 {
		t.Fatalf("expected 1 dispatch (BTN_SOUTH only), got %d", len(cmdTarget.calls))
	}

	// Verify the modifier was set in engine state.
	active := state.ActiveModifiers
	if !active["BTN_TL"] {
		t.Error("expected BTN_TL modifier to be active in engine state")
	}
}

func TestPipeline_LayerSwitch(t *testing.T) {
	state := mapping.NewEngineState()
	registry := output.NewRegistry()

	// Register a real LayerTarget so the pipeline dispatches to it.
	layerTarget := output.NewLayerTarget(state)
	registry.Register(layerTarget)

	rule := &mapping.MappingRule{
		Input: "BTN_SELECT",
		Output: mapping.OutputAction{
			Type:   mapping.OutputLayerSwitch,
			Layer:  "2",
			Target: "xbox",
		},
	}

	p := New(state, registry, func(deviceID string, source string) *mapping.MappingRule {
		if source == "BTN_SELECT" {
			return rule
		}
		return nil
	}, nil)

	events := make(chan device.Event, 1)
	events <- device.Event{
		DeviceID: "xbox",
		Source:   "BTN_SELECT",
		Value:    1.0,
	}
	close(events)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p.ProcessEvents(ctx, events)

	// Verify the engine state's active layer was updated for the device.
	got := state.GetActiveLayer("xbox")
	if got != 2 {
		t.Errorf("expected active layer 2 for device 'xbox', got %d", got)
	}
}

func TestPipeline_LayerSwitchGlobal(t *testing.T) {
	state := mapping.NewEngineState()
	registry := output.NewRegistry()

	layerTarget := output.NewLayerTarget(state)
	registry.Register(layerTarget)

	// When Target is empty, LayerTarget uses "__global".
	rule := &mapping.MappingRule{
		Input: "BTN_MODE",
		Output: mapping.OutputAction{
			Type:  mapping.OutputLayerSwitch,
			Layer: "1",
		},
	}

	p := New(state, registry, func(deviceID string, source string) *mapping.MappingRule {
		if source == "BTN_MODE" {
			return rule
		}
		return nil
	}, nil)

	events := make(chan device.Event, 1)
	events <- device.Event{
		DeviceID: "xbox",
		Source:   "BTN_MODE",
		Value:    1.0,
	}
	close(events)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p.ProcessEvents(ctx, events)

	got := state.GetActiveLayer("__global")
	if got != 1 {
		t.Errorf("expected active layer 1 for '__global', got %d", got)
	}
}

func TestPipeline_Sequence(t *testing.T) {
	state := mapping.NewEngineState()
	registry := output.NewRegistry()

	// Register mock targets for each step type in the sequence.
	cmdTarget := &mockTarget{typ: mapping.OutputCommand}
	oscTarget := &mockTarget{typ: mapping.OutputOSC}
	registry.Register(cmdTarget)
	registry.Register(oscTarget)

	// Register a real SequenceTarget that will dispatch steps through the registry.
	seqTarget := output.NewSequenceTarget(registry)
	registry.Register(seqTarget)

	rule := &mapping.MappingRule{
		Input: "BTN_EAST",
		Output: mapping.OutputAction{
			Type: mapping.OutputSequence,
			Steps: []mapping.SequenceStep{
				{
					Type: mapping.OutputCommand,
					Exec: []string{"echo", "step1"},
				},
				{
					Type:    mapping.OutputOSC,
					Address: "/fx/toggle",
					Port:    9000,
				},
			},
		},
	}

	p := New(state, registry, func(deviceID string, source string) *mapping.MappingRule {
		if source == "BTN_EAST" {
			return rule
		}
		return nil
	}, nil)

	events := make(chan device.Event, 1)
	events <- device.Event{
		DeviceID: "xbox",
		Source:   "BTN_EAST",
		Value:    1.0,
	}
	close(events)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p.ProcessEvents(ctx, events)

	// Verify both steps dispatched.
	if len(cmdTarget.calls) != 1 {
		t.Fatalf("expected 1 command call, got %d", len(cmdTarget.calls))
	}
	if cmdTarget.calls[0].action.Exec[0] != "echo" || cmdTarget.calls[0].action.Exec[1] != "step1" {
		t.Errorf("unexpected command exec: %v", cmdTarget.calls[0].action.Exec)
	}
	if cmdTarget.calls[0].value != 1.0 {
		t.Errorf("expected value 1.0 passed to command step, got %f", cmdTarget.calls[0].value)
	}

	if len(oscTarget.calls) != 1 {
		t.Fatalf("expected 1 OSC call, got %d", len(oscTarget.calls))
	}
	if oscTarget.calls[0].action.Address != "/fx/toggle" {
		t.Errorf("expected OSC address '/fx/toggle', got %q", oscTarget.calls[0].action.Address)
	}
	if oscTarget.calls[0].value != 1.0 {
		t.Errorf("expected value 1.0 passed to OSC step, got %f", oscTarget.calls[0].value)
	}
}

func TestPipeline_SequencePreservesOrder(t *testing.T) {
	state := mapping.NewEngineState()
	registry := output.NewRegistry()

	// Use a single mock target type for all steps so we can verify order.
	cmdTarget := &mockTarget{typ: mapping.OutputCommand}
	registry.Register(cmdTarget)

	seqTarget := output.NewSequenceTarget(registry)
	registry.Register(seqTarget)

	rule := &mapping.MappingRule{
		Input: "BTN_WEST",
		Output: mapping.OutputAction{
			Type: mapping.OutputSequence,
			Steps: []mapping.SequenceStep{
				{Type: mapping.OutputCommand, Exec: []string{"first"}},
				{Type: mapping.OutputCommand, Exec: []string{"second"}},
				{Type: mapping.OutputCommand, Exec: []string{"third"}},
			},
		},
	}

	p := New(state, registry, func(deviceID string, source string) *mapping.MappingRule {
		if source == "BTN_WEST" {
			return rule
		}
		return nil
	}, nil)

	events := make(chan device.Event, 1)
	events <- device.Event{
		DeviceID: "xbox",
		Source:   "BTN_WEST",
		Value:    1.0,
	}
	close(events)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	p.ProcessEvents(ctx, events)

	if len(cmdTarget.calls) != 3 {
		t.Fatalf("expected 3 calls, got %d", len(cmdTarget.calls))
	}

	expected := []string{"first", "second", "third"}
	for i, want := range expected {
		got := cmdTarget.calls[i].action.Exec[0]
		if got != want {
			t.Errorf("step %d: expected exec[0]=%q, got %q", i, want, got)
		}
	}
}
