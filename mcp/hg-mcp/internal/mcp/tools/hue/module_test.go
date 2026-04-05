package hue

import (
	"testing"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

func TestModuleRegistration(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}

func TestToolNames(t *testing.T) {
	m := &Module{}
	expected := map[string]bool{
		"aftrs_hue_status":         true,
		"aftrs_hue_health":         true,
		"aftrs_hue_lights":         true,
		"aftrs_hue_light_control":  true,
		"aftrs_hue_rooms":          true,
		"aftrs_hue_room_control":   true,
		"aftrs_hue_scenes":         true,
		"aftrs_hue_scene_activate": true,
		"aftrs_hue_discover":       true,
		"aftrs_hue_entertainment":  true,
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
		if td.Category != "hue" {
			t.Errorf("tool %s has category %q, expected 'hue'", td.Tool.Name, td.Category)
		}
	}
}

func TestWriteToolsMarked(t *testing.T) {
	m := &Module{}
	writeTools := map[string]bool{
		"aftrs_hue_light_control":  true,
		"aftrs_hue_room_control":   true,
		"aftrs_hue_scene_activate": true,
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

func TestHexToHueSat(t *testing.T) {
	// Red: FF0000 → hue=0, sat=254
	h, s, err := hexToHueSat("FF0000")
	if err != nil {
		t.Fatal(err)
	}
	if h != 0 || s != 254 {
		t.Errorf("FF0000: got h=%d s=%d, want 0,254", h, s)
	}

	// Invalid
	_, _, err = hexToHueSat("GG")
	if err == nil {
		t.Error("expected error for invalid hex")
	}
}

func TestColorNameToHueSat(t *testing.T) {
	tests := []struct {
		name string
		ok   bool
	}{
		{"red", true},
		{"blue", true},
		{"unknown", false},
	}
	for _, tc := range tests {
		_, _, ok := colorNameToHueSat(tc.name)
		if ok != tc.ok {
			t.Errorf("colorNameToHueSat(%q): got ok=%v, want %v", tc.name, ok, tc.ok)
		}
	}
}

func TestRuntimeGroupAssignment(t *testing.T) {
	registry := tools.NewToolRegistry()
	registry.RegisterModule(&Module{})

	lighting := registry.ListToolsByRuntimeGroup("lighting")
	if len(lighting) != 10 {
		t.Errorf("expected 10 lighting tools from hue, got %d", len(lighting))
	}
}
