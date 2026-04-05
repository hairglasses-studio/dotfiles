package targets

import (
	"context"
	"testing"
)

func TestDesktopTarget_Actions(t *testing.T) {
	target := NewDesktopTarget()
	actions := target.Actions(context.Background())

	// Should have at least 4 core actions (key_press, type_text, mouse_click, run_command).
	if len(actions) < 4 {
		t.Fatalf("expected at least 4 actions, got %d", len(actions))
	}

	ids := map[string]bool{}
	for _, a := range actions {
		ids[a.ID] = true
	}
	for _, want := range []string{"key_press", "type_text", "mouse_click", "run_command"} {
		if !ids[want] {
			t.Errorf("missing action: %s", want)
		}
	}
}

func TestDesktopTarget_Health(t *testing.T) {
	target := NewDesktopTarget()
	h := target.Health(context.Background())
	if !h.Connected {
		t.Error("desktop target should always be connected")
	}
	if h.Status != "healthy" {
		t.Errorf("status = %q, want healthy", h.Status)
	}
}

func TestDesktopTarget_ExecuteRunCommand(t *testing.T) {
	target := NewDesktopTarget()
	result, err := target.Execute(context.Background(), "run_command", map[string]any{
		"command": "echo hello_desktop",
	})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if !result.Success {
		t.Fatalf("expected success: %s", result.Error)
	}
	if output, ok := result.Data["output"].(string); !ok || output != "hello_desktop" {
		t.Errorf("output = %v, want 'hello_desktop'", result.Data["output"])
	}
}

func TestDesktopTarget_ExecuteUnknownAction(t *testing.T) {
	target := NewDesktopTarget()
	result, err := target.Execute(context.Background(), "nonexistent", nil)
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if result.Success {
		t.Error("expected failure for unknown action")
	}
}

func TestDesktopTarget_ExecuteMissingKey(t *testing.T) {
	target := NewDesktopTarget()
	result, err := target.Execute(context.Background(), "key_press", map[string]any{})
	if err != nil {
		t.Fatalf("Execute error: %v", err)
	}
	if result.Success {
		t.Error("expected failure for missing key")
	}
}

func TestDesktopTarget_Interface(t *testing.T) {
	// Verify it satisfies OutputTarget interface.
	var _ OutputTarget = (*DesktopTarget)(nil)

	target := NewDesktopTarget()
	if target.ID() != "desktop" {
		t.Errorf("ID = %q, want desktop", target.ID())
	}
	if target.Protocol() != "platform" {
		t.Errorf("Protocol = %q, want platform", target.Protocol())
	}
}
