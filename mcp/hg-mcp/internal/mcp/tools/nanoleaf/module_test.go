package nanoleaf

import (
	"testing"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

func TestModuleRegistration(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}

func TestToolCount(t *testing.T) {
	m := &Module{}
	defs := m.Tools()
	if len(defs) != 8 {
		t.Errorf("expected 8 tools, got %d", len(defs))
	}
}

func TestToolNames(t *testing.T) {
	m := &Module{}
	expected := map[string]bool{
		"aftrs_nanoleaf_status":     true,
		"aftrs_nanoleaf_health":     true,
		"aftrs_nanoleaf_power":      true,
		"aftrs_nanoleaf_brightness": true,
		"aftrs_nanoleaf_color":      true,
		"aftrs_nanoleaf_effect":     true,
		"aftrs_nanoleaf_panels":     true,
		"aftrs_nanoleaf_discover":   true,
	}
	for _, td := range m.Tools() {
		if !expected[td.Tool.Name] {
			t.Errorf("unexpected tool name: %s", td.Tool.Name)
		}
		delete(expected, td.Tool.Name)
	}
	for name := range expected {
		t.Errorf("missing tool: %s", name)
	}
}

func TestAllToolsHaveCategory(t *testing.T) {
	m := &Module{}
	for _, td := range m.Tools() {
		if td.Category != "nanoleaf" {
			t.Errorf("tool %s has category %q, expected 'nanoleaf'", td.Tool.Name, td.Category)
		}
	}
}

func TestAllToolsHaveCircuitBreaker(t *testing.T) {
	m := &Module{}
	for _, td := range m.Tools() {
		if td.CircuitBreakerGroup != "nanoleaf" {
			t.Errorf("tool %s missing CircuitBreakerGroup", td.Tool.Name)
		}
	}
}

func TestWriteToolsMarked(t *testing.T) {
	m := &Module{}
	writeTools := map[string]bool{
		"aftrs_nanoleaf_power":      true,
		"aftrs_nanoleaf_brightness": true,
		"aftrs_nanoleaf_color":      true,
		"aftrs_nanoleaf_effect":     true,
	}
	for _, td := range m.Tools() {
		if writeTools[td.Tool.Name] && !td.IsWrite {
			t.Errorf("tool %s should be marked IsWrite", td.Tool.Name)
		}
		if !writeTools[td.Tool.Name] && td.IsWrite {
			t.Errorf("tool %s should not be marked IsWrite", td.Tool.Name)
		}
	}
}

func TestAllToolsHaveHandlers(t *testing.T) {
	m := &Module{}
	for _, td := range m.Tools() {
		if td.Handler == nil {
			t.Errorf("tool %s has nil handler", td.Tool.Name)
		}
	}
}

func TestColorNameToHSB(t *testing.T) {
	tests := []struct {
		name string
		ok   bool
	}{
		{"red", true},
		{"blue", true},
		{"green", true},
		{"white", true},
		{"unknown", false},
	}
	for _, tc := range tests {
		_, _, _, ok := colorNameToHSB(tc.name)
		if ok != tc.ok {
			t.Errorf("colorNameToHSB(%q): got ok=%v, want %v", tc.name, ok, tc.ok)
		}
	}
}

func TestHexToHSB(t *testing.T) {
	// Red: FF0000 → hue=0, sat=100, bri=100
	h, s, b, err := hexToHSB("FF0000")
	if err != nil {
		t.Fatal(err)
	}
	if h != 0 || s != 100 || b != 100 {
		t.Errorf("FF0000: got h=%d s=%d b=%d, want 0,100,100", h, s, b)
	}

	// Invalid
	_, _, _, err = hexToHSB("GG")
	if err == nil {
		t.Error("expected error for invalid hex")
	}
}

func TestRuntimeGroupAssignment(t *testing.T) {
	// Verify that when registered, nanoleaf tools get RuntimeGroup=lighting
	registry := tools.NewToolRegistry()
	registry.RegisterModule(&Module{})

	lighting := registry.ListToolsByRuntimeGroup("lighting")
	if len(lighting) != 8 {
		t.Errorf("expected 8 lighting tools from nanoleaf, got %d", len(lighting))
	}
}
