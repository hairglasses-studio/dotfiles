package mapping

import (
	"testing"
)

// ---------------------------------------------------------------------------
// ParseLegacyInput
// ---------------------------------------------------------------------------

func TestParseLegacyInput(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantInput string
		wantMods  []string
	}{
		{
			name:      "simple button",
			input:     "BTN_SOUTH",
			wantInput: "BTN_SOUTH",
			wantMods:  nil,
		},
		{
			name:      "abs axis",
			input:     "ABS_Z",
			wantInput: "ABS_Z",
			wantMods:  nil,
		},
		{
			name:      "one modifier",
			input:     "BTN_TL-BTN_SOUTH",
			wantInput: "BTN_SOUTH",
			wantMods:  []string{"BTN_TL"},
		},
		{
			name:      "two modifiers",
			input:     "BTN_TL-BTN_TR-BTN_SOUTH",
			wantInput: "BTN_SOUTH",
			wantMods:  []string{"BTN_TL", "BTN_TR"},
		},
		{
			name:      "keyboard combo",
			input:     "KEY_LEFTCTRL-KEY_N",
			wantInput: "KEY_N",
			wantMods:  []string{"KEY_LEFTCTRL"},
		},
		{
			name:      "three modifiers",
			input:     "KEY_LEFTCTRL-KEY_LEFTSHIFT-KEY_LEFTALT-KEY_T",
			wantInput: "KEY_T",
			wantMods:  []string{"KEY_LEFTCTRL", "KEY_LEFTSHIFT", "KEY_LEFTALT"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := ParseLegacyInput(tt.input)
			if rule.Input != tt.wantInput {
				t.Errorf("Input = %q, want %q", rule.Input, tt.wantInput)
			}
			if len(rule.Modifiers) != len(tt.wantMods) {
				t.Fatalf("len(Modifiers) = %d, want %d", len(rule.Modifiers), len(tt.wantMods))
			}
			for i, mod := range rule.Modifiers {
				if mod != tt.wantMods[i] {
					t.Errorf("Modifiers[%d] = %q, want %q", i, mod, tt.wantMods[i])
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ConvertLegacyToUnified — remap section
// ---------------------------------------------------------------------------

func TestConvertLegacyToUnified_Remap(t *testing.T) {
	legacy := &MappingProfile{
		SourceName: "DualShock4",
		Remap: map[string][]string{
			"BTN_SOUTH": {"KEY_Z"},
			"BTN_EAST":  {"KEY_X"},
		},
	}

	unified := ConvertLegacyToUnified(legacy)

	if !unified.IsUnifiedFormat() {
		t.Fatal("converted profile should be unified format")
	}
	if unified.Profile.SchemaVersion != 1 {
		t.Errorf("SchemaVersion = %d, want 1", unified.Profile.SchemaVersion)
	}
	if unified.Profile.Device != "DualShock4" {
		t.Errorf("Device = %q, want %q", unified.Profile.Device, "DualShock4")
	}
	if len(unified.Mappings) != 2 {
		t.Fatalf("len(Mappings) = %d, want 2", len(unified.Mappings))
	}

	// Mappings should be sorted by input.
	if unified.Mappings[0].Input != "BTN_EAST" {
		t.Errorf("first mapping Input = %q, want BTN_EAST (sorted)", unified.Mappings[0].Input)
	}
	for _, m := range unified.Mappings {
		if m.Output.Type != OutputKey {
			t.Errorf("mapping %q output type = %q, want %q", m.Input, m.Output.Type, OutputKey)
		}
	}
}

// ---------------------------------------------------------------------------
// ConvertLegacyToUnified — commands section
// ---------------------------------------------------------------------------

func TestConvertLegacyToUnified_Commands(t *testing.T) {
	legacy := &MappingProfile{
		SourceName: "TestPad",
		Commands: map[string][]string{
			"BTN_NORTH": {"playerctl", "play-pause"},
		},
	}

	unified := ConvertLegacyToUnified(legacy)

	if len(unified.Mappings) != 1 {
		t.Fatalf("len(Mappings) = %d, want 1", len(unified.Mappings))
	}
	m := unified.Mappings[0]
	if m.Output.Type != OutputCommand {
		t.Errorf("output type = %q, want %q", m.Output.Type, OutputCommand)
	}
	if len(m.Output.Exec) != 2 || m.Output.Exec[0] != "playerctl" {
		t.Errorf("Exec = %v, want [playerctl play-pause]", m.Output.Exec)
	}
}

// ---------------------------------------------------------------------------
// ConvertLegacyToUnified — movements section
// ---------------------------------------------------------------------------

func TestConvertLegacyToUnified_Movements(t *testing.T) {
	legacy := &MappingProfile{
		SourceName: "TestPad",
		Movements: map[string]string{
			"ABS_X": "cursor_right",
			"ABS_Y": "cursor_down",
		},
	}

	unified := ConvertLegacyToUnified(legacy)

	if len(unified.Mappings) != 2 {
		t.Fatalf("len(Mappings) = %d, want 2", len(unified.Mappings))
	}
	for _, m := range unified.Mappings {
		if m.Output.Type != OutputMovement {
			t.Errorf("mapping %q output type = %q, want %q", m.Input, m.Output.Type, OutputMovement)
		}
		if m.Output.Target == "" {
			t.Errorf("mapping %q output target is empty", m.Input)
		}
	}
}

// ---------------------------------------------------------------------------
// ConvertLegacyToUnified — combined sections
// ---------------------------------------------------------------------------

func TestConvertLegacyToUnified_Combined(t *testing.T) {
	legacy := &MappingProfile{
		SourceName: "Pad",
		Remap: map[string][]string{
			"BTN_SOUTH": {"KEY_Z"},
		},
		Commands: map[string][]string{
			"BTN_NORTH": {"echo", "hello"},
		},
		Movements: map[string]string{
			"ABS_X": "cursor_right",
		},
	}

	unified := ConvertLegacyToUnified(legacy)
	if len(unified.Mappings) != 3 {
		t.Fatalf("len(Mappings) = %d, want 3", len(unified.Mappings))
	}

	// Check sorted order.
	inputs := make([]string, len(unified.Mappings))
	for i, m := range unified.Mappings {
		inputs[i] = m.Input
	}
	for i := 1; i < len(inputs); i++ {
		if inputs[i] < inputs[i-1] {
			t.Errorf("mappings not sorted: %v", inputs)
			break
		}
	}
}

// ---------------------------------------------------------------------------
// ConvertLegacyToUnified — modifier extraction
// ---------------------------------------------------------------------------

func TestConvertLegacyToUnified_WithModifiers(t *testing.T) {
	legacy := &MappingProfile{
		SourceName: "Pad",
		Remap: map[string][]string{
			"BTN_TL-BTN_SOUTH": {"KEY_LEFTCTRL", "KEY_Z"},
		},
	}

	unified := ConvertLegacyToUnified(legacy)
	if len(unified.Mappings) != 1 {
		t.Fatalf("len(Mappings) = %d, want 1", len(unified.Mappings))
	}

	m := unified.Mappings[0]
	if m.Input != "BTN_SOUTH" {
		t.Errorf("Input = %q, want BTN_SOUTH", m.Input)
	}
	if len(m.Modifiers) != 1 || m.Modifiers[0] != "BTN_TL" {
		t.Errorf("Modifiers = %v, want [BTN_TL]", m.Modifiers)
	}
}

// ---------------------------------------------------------------------------
// ConvertLegacyToUnified — app class from source name
// ---------------------------------------------------------------------------

func TestConvertLegacyToUnified_AppClass(t *testing.T) {
	legacy := &MappingProfile{
		SourceName: "DualShock4::firefox",
		Remap: map[string][]string{
			"BTN_SOUTH": {"KEY_ENTER"},
		},
	}

	unified := ConvertLegacyToUnified(legacy)

	if unified.Profile.AppClass != "firefox" {
		t.Errorf("AppClass = %q, want %q", unified.Profile.AppClass, "firefox")
	}
	if len(unified.AppOverrides) != 1 {
		t.Fatalf("len(AppOverrides) = %d, want 1", len(unified.AppOverrides))
	}
	if unified.AppOverrides[0].WindowClass != "firefox" {
		t.Errorf("WindowClass = %q, want %q", unified.AppOverrides[0].WindowClass, "firefox")
	}
	// Default mappings should be empty when app class is set.
	if len(unified.Mappings) != 0 {
		t.Errorf("len(Mappings) = %d, want 0 (rules go to AppOverrides)", len(unified.Mappings))
	}
}

// ---------------------------------------------------------------------------
// ConvertLegacyToUnified — layer from source name
// ---------------------------------------------------------------------------

func TestConvertLegacyToUnified_LayerFromName(t *testing.T) {
	legacy := &MappingProfile{
		SourceName: "DualShock4::2",
		Remap: map[string][]string{
			"BTN_SOUTH": {"KEY_A"},
		},
	}

	unified := ConvertLegacyToUnified(legacy)

	if unified.Profile.Layer != 2 {
		t.Errorf("Layer = %d, want 2", unified.Profile.Layer)
	}
	// Numeric suffix means it is a layer, not an app class.
	if unified.Profile.AppClass != "" {
		t.Errorf("AppClass = %q, want empty (numeric = layer)", unified.Profile.AppClass)
	}
}

// ---------------------------------------------------------------------------
// ConvertLegacySettings
// ---------------------------------------------------------------------------

func TestConvertLegacySettings(t *testing.T) {
	legacy := map[string]string{
		"GRAB_DEVICE":       "true",
		"16_BIT_AXIS":       "true",
		"CHAIN_ONLY":        "true",
		"INVERT_CURSOR_AXIS": "true",
		"INVERT_SCROLL_AXIS": "true",
		"LAYOUT_SWITCHER":    "sway-layout",
		"LSTICK":             "cursor",
		"LSTICK_SENSITIVITY": "50",
		"LSTICK_DEADZONE":    "10",
		"RSTICK":             "scroll",
		"RSTICK_SENSITIVITY": "30",
		"RSTICK_DEADZONE":    "5",
		"CURSOR_SPEED":       "1500",
		"CURSOR_ACCEL":       "1.5",
		"SCROLL_SPEED":       "800",
		"SCROLL_ACCEL":       "0.8",
		"CUSTOM_MODIFIERS":   "BTN_TL-BTN_TR",
	}

	s := &MappingSettings{}
	ConvertLegacySettings(s, legacy)

	if !s.GrabDevice {
		t.Error("expected GrabDevice = true")
	}
	if !s.Axis16Bit {
		t.Error("expected Axis16Bit = true")
	}
	if !s.ChainOnly {
		t.Error("expected ChainOnly = true")
	}
	if !s.InvertCursorAxis {
		t.Error("expected InvertCursorAxis = true")
	}
	if !s.InvertScrollAxis {
		t.Error("expected InvertScrollAxis = true")
	}
	if s.LayoutSwitcher != "sway-layout" {
		t.Errorf("LayoutSwitcher = %q, want %q", s.LayoutSwitcher, "sway-layout")
	}

	// LStick
	if s.LStick == nil {
		t.Fatal("LStick is nil")
	}
	if s.LStick.Function != "cursor" {
		t.Errorf("LStick.Function = %q, want %q", s.LStick.Function, "cursor")
	}
	if s.LStick.Sensitivity != 50 {
		t.Errorf("LStick.Sensitivity = %d, want 50", s.LStick.Sensitivity)
	}
	if s.LStick.Deadzone != 10 {
		t.Errorf("LStick.Deadzone = %d, want 10", s.LStick.Deadzone)
	}

	// RStick
	if s.RStick == nil {
		t.Fatal("RStick is nil")
	}
	if s.RStick.Function != "scroll" {
		t.Errorf("RStick.Function = %q, want %q", s.RStick.Function, "scroll")
	}
	if s.RStick.Sensitivity != 30 {
		t.Errorf("RStick.Sensitivity = %d, want 30", s.RStick.Sensitivity)
	}

	// Cursor
	if s.Cursor == nil {
		t.Fatal("Cursor is nil")
	}
	if s.Cursor.Speed != 1500 {
		t.Errorf("Cursor.Speed = %d, want 1500", s.Cursor.Speed)
	}
	if s.Cursor.Acceleration != 1.5 {
		t.Errorf("Cursor.Acceleration = %f, want 1.5", s.Cursor.Acceleration)
	}

	// Scroll
	if s.Scroll == nil {
		t.Fatal("Scroll is nil")
	}
	if s.Scroll.Speed != 800 {
		t.Errorf("Scroll.Speed = %d, want 800", s.Scroll.Speed)
	}
	if s.Scroll.Acceleration != 0.8 {
		t.Errorf("Scroll.Acceleration = %f, want 0.8", s.Scroll.Acceleration)
	}

	// Custom modifiers
	if len(s.CustomModifiers) != 2 {
		t.Fatalf("len(CustomModifiers) = %d, want 2", len(s.CustomModifiers))
	}
	if s.CustomModifiers[0] != "BTN_TL" || s.CustomModifiers[1] != "BTN_TR" {
		t.Errorf("CustomModifiers = %v, want [BTN_TL BTN_TR]", s.CustomModifiers)
	}
}

func TestConvertLegacySettings_Nil(t *testing.T) {
	s := &MappingSettings{}
	ConvertLegacySettings(s, nil) // should not panic
	if s.GrabDevice {
		t.Error("settings should remain zero-valued with nil legacy map")
	}
}

func TestConvertLegacySettings_Empty(t *testing.T) {
	s := &MappingSettings{}
	ConvertLegacySettings(s, map[string]string{})
	if s.GrabDevice {
		t.Error("settings should remain zero-valued with empty legacy map")
	}
}

// ---------------------------------------------------------------------------
// ConvertLegacyToUnified — empty legacy
// ---------------------------------------------------------------------------

func TestConvertLegacyToUnified_EmptyLegacy(t *testing.T) {
	legacy := &MappingProfile{
		SourceName: "empty",
	}
	unified := ConvertLegacyToUnified(legacy)

	if len(unified.Mappings) != 0 {
		t.Errorf("len(Mappings) = %d, want 0", len(unified.Mappings))
	}
	if !unified.IsUnifiedFormat() {
		t.Error("converted profile should still be unified format")
	}
}
