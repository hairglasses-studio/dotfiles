package targets

import (
	"context"
	"testing"
)

func TestOSCTarget_Actions(t *testing.T) {
	target := NewOSCTarget(OSCTargetConfig{
		Host: "127.0.0.1",
		Port: 7000,
	})

	actions := target.Actions(context.Background())
	if len(actions) != 4 { // send_float, send_int, send_trigger, send_string
		t.Fatalf("expected 4 default actions, got %d", len(actions))
	}

	names := map[string]bool{}
	for _, a := range actions {
		names[a.ID] = true
	}
	for _, want := range []string{"send_float", "send_int", "send_trigger", "send_string"} {
		if !names[want] {
			t.Errorf("missing action: %s", want)
		}
	}
}

func TestOSCTarget_CustomActions(t *testing.T) {
	one := 1.0
	zero := 0.0
	target := NewOSCTarget(OSCTargetConfig{
		ID:   "resolume",
		Name: "Resolume Arena",
		Host: "localhost",
		Port: 7000,
		Actions: []OSCActionConfig{
			{ID: "layer_opacity", Name: "Layer Opacity", Address: "/layer/opacity", Type: "set_value", ParamType: "float32", Min: &zero, Max: &one},
			{ID: "tap_tempo", Name: "Tap Tempo", Address: "/tempo/tap", Type: "trigger"},
		},
	})

	actions := target.Actions(context.Background())
	if len(actions) != 2 {
		t.Fatalf("expected 2 custom actions, got %d", len(actions))
	}
	if actions[0].ID != "layer_opacity" {
		t.Errorf("first action ID = %q, want layer_opacity", actions[0].ID)
	}
	if actions[0].Type != ActionSetValue {
		t.Errorf("first action type = %v, want set_value", actions[0].Type)
	}
	if actions[1].Type != ActionTrigger {
		t.Errorf("second action type = %v, want trigger", actions[1].Type)
	}
}

func TestOSCTarget_Health(t *testing.T) {
	target := NewOSCTarget(OSCTargetConfig{Host: "localhost", Port: 7000})

	// Not connected yet.
	h := target.Health(context.Background())
	if h.Connected {
		t.Error("should not be connected before Connect()")
	}

	// After connect.
	target.Connect(context.Background())
	h = target.Health(context.Background())
	if !h.Connected {
		t.Error("should be connected after Connect()")
	}

	// After disconnect.
	target.Disconnect(context.Background())
	h = target.Health(context.Background())
	if h.Connected {
		t.Error("should not be connected after Disconnect()")
	}
}

func TestOSCTarget_ExecuteNotConnected(t *testing.T) {
	target := NewOSCTarget(OSCTargetConfig{Host: "localhost", Port: 7000})
	result, err := target.Execute(context.Background(), "send_float", map[string]any{
		"address": "/test", "value": 0.5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Error("should fail when not connected")
	}
}

func TestOSCTarget_ExecuteMissingAddress(t *testing.T) {
	target := NewOSCTarget(OSCTargetConfig{Host: "localhost", Port: 7000})
	target.Connect(context.Background())
	defer target.Disconnect(context.Background())

	result, err := target.Execute(context.Background(), "send_float", map[string]any{
		"value": 0.5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Success {
		t.Error("should fail without address")
	}
}

func TestOSCTarget_IDGeneration(t *testing.T) {
	target := NewOSCTarget(OSCTargetConfig{Host: "192.168.1.50", Port: 7001})
	if target.ID() != "osc_192.168.1.50_7001" {
		t.Errorf("generated ID = %q", target.ID())
	}
}

func TestResolumeTarget(t *testing.T) {
	target := NewResolumeTarget("localhost", 7000)
	if target.ID() != "resolume" {
		t.Errorf("ID = %q, want resolume", target.ID())
	}
	if target.Name() != "Resolume Arena" {
		t.Errorf("Name = %q", target.Name())
	}

	actions := target.Actions(context.Background())
	if len(actions) < 9 {
		t.Errorf("expected at least 9 Resolume actions, got %d", len(actions))
	}

	// Check for expected actions.
	ids := map[string]bool{}
	for _, a := range actions {
		ids[a.ID] = true
	}
	for _, want := range []string{"layer_opacity", "crossfader", "tap_tempo", "set_bpm", "trigger_clip"} {
		if !ids[want] {
			t.Errorf("missing Resolume action: %s", want)
		}
	}
}

func TestTouchDesignerTarget(t *testing.T) {
	target := NewTouchDesignerTarget("localhost", 7001)
	if target.ID() != "touchdesigner" {
		t.Errorf("ID = %q, want touchdesigner", target.ID())
	}
	// TD target should have default generic actions.
	actions := target.Actions(context.Background())
	if len(actions) != 4 {
		t.Errorf("expected 4 default actions, got %d", len(actions))
	}
}
