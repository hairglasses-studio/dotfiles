package prefetch

import (
	"testing"
)

func TestModuleInfo(t *testing.T) {
	m := &Module{}
	if m.Name() != "prefetch" {
		t.Errorf("expected name 'prefetch', got %q", m.Name())
	}
	if m.Description() == "" {
		t.Error("description should not be empty")
	}
}

func TestModuleTools(t *testing.T) {
	m := &Module{}
	defs := m.Tools()
	if len(defs) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(defs))
	}

	tool := defs[0]
	if tool.Tool.Name != "aftrs_prefetch_system_context" {
		t.Errorf("expected tool name 'aftrs_prefetch_system_context', got %q", tool.Tool.Name)
	}
	if tool.Category != "platform" {
		t.Errorf("expected category 'platform', got %q", tool.Category)
	}
	if tool.Handler == nil {
		t.Error("handler should not be nil")
	}
}

func TestRunCmdInvalid(t *testing.T) {
	// Test that runCmd returns empty string for nonexistent commands
	result := runCmd(t.Context(), "nonexistent_binary_12345")
	if result != "" {
		t.Errorf("expected empty string for invalid command, got %q", result)
	}
}
