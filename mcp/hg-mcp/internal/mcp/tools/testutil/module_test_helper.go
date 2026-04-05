package testutil

import (
	"testing"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// ModuleInfoTest holds the configuration for AssertModuleInfo.
type ModuleInfoTest struct {
	Module       tools.ToolModule
	ExpectedName string
	MinTools     int
	CBGroup      string
}

// AssertModuleValid validates that a module has a non-empty name, non-empty description,
// at least one tool, and that every tool has a name, description, non-nil handler, and
// a CircuitBreakerGroup matching the module name. This is the standard one-liner
// replacement for the verbose ModuleInfoTest struct pattern.
func AssertModuleValid(t *testing.T, m tools.ToolModule) {
	t.Helper()
	name := m.Name()
	if name == "" {
		t.Fatal("module name should not be empty")
	}
	if m.Description() == "" {
		t.Error("module description should not be empty")
	}
	defs := m.Tools()
	if len(defs) == 0 {
		t.Fatal("module should have at least one tool")
	}
	for _, td := range defs {
		if td.Tool.Name == "" {
			t.Error("tool has empty name")
		}
		if td.Tool.Description == "" {
			t.Errorf("tool %s should have a description", td.Tool.Name)
		}
		if td.Handler == nil {
			t.Errorf("tool %s has nil handler", td.Tool.Name)
		}
		if td.CircuitBreakerGroup != name {
			t.Errorf("tool %s CBGroup = %q, want %q", td.Tool.Name, td.CircuitBreakerGroup, name)
		}
	}
}

// AssertModuleInfo validates common module invariants: name, non-empty description,
// minimum tool count, non-nil handlers, and circuit breaker group on every tool.
// Deprecated: Prefer AssertModuleValid for the common case where CBGroup == module name.
// Use AssertModuleInfo only when CBGroup differs from the module name.
func AssertModuleInfo(t *testing.T, cfg ModuleInfoTest) {
	t.Helper()
	if cfg.Module.Name() != cfg.ExpectedName {
		t.Errorf("Name() = %q, want %q", cfg.Module.Name(), cfg.ExpectedName)
	}
	if cfg.Module.Description() == "" {
		t.Error("Description() should not be empty")
	}
	defs := cfg.Module.Tools()
	if len(defs) < cfg.MinTools {
		t.Errorf("Tools() returned %d tools, want at least %d", len(defs), cfg.MinTools)
	}
	for _, td := range defs {
		if td.Tool.Name == "" {
			t.Error("tool has empty name")
		}
		if td.Handler == nil {
			t.Errorf("tool %s has nil handler", td.Tool.Name)
		}
		if cfg.CBGroup != "" && td.CircuitBreakerGroup != cfg.CBGroup {
			t.Errorf("tool %s CBGroup = %q, want %q", td.Tool.Name, td.CircuitBreakerGroup, cfg.CBGroup)
		}
	}
}
