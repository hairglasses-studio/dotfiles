package mapping

import (
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// Unified format TOML fixtures
// ---------------------------------------------------------------------------

const validUnifiedTOML = `
[profile]
schema_version = 1
type = "mapping"
device = "DualSense"
tags = ["gamepad", "ps5"]
description = "PS5 DualSense profile"

[device]
name = "DualSense"
type = "gamepad"
vendor_id = "054c"
product_id = "0ce6"

[settings]
grab_device = true
custom_modifiers = ["BTN_TL", "BTN_TR"]

[[mapping]]
input = "BTN_SOUTH"
description = "Cross button"

[mapping.output]
type = "key"
keys = ["KEY_ENTER"]

[[mapping]]
input = "ABS_Z"
description = "L2 trigger"

[mapping.output]
type = "osc"
address = "/fader/1"
host = "127.0.0.1"
port = 9000

[mapping.value]
input_range = [0.0, 255.0]
output_range = [0.0, 1.0]
curve = "logarithmic"
`

const validLegacyTOML = `
[remap]
BTN_SOUTH = ["KEY_Z"]
BTN_EAST = ["KEY_X"]
BTN_TL-BTN_SOUTH = ["KEY_LEFTCTRL", "KEY_Z"]

[commands]
BTN_NORTH = ["playerctl", "play-pause"]

[movements]
ABS_X = "cursor_right"
ABS_Y = "cursor_down"

[settings]
GRAB_DEVICE = "true"
LSTICK = "cursor"
LSTICK_SENSITIVITY = "50"
CURSOR_SPEED = "1500"
`

const malformedTOML = `
[profile
schema_version = 1
this is not valid TOML at all
`

const emptyProfileTOML = `
# just a comment, no sections
`

// ---------------------------------------------------------------------------
// LoadMappingProfile
// ---------------------------------------------------------------------------

func TestLoadMappingProfile_ValidUnified(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "dualsense.toml")
	if err := os.WriteFile(path, []byte(validUnifiedTOML), 0644); err != nil {
		t.Fatal(err)
	}

	p, err := LoadMappingProfile(path)
	if err != nil {
		t.Fatalf("LoadMappingProfile: %v", err)
	}

	if !p.IsUnifiedFormat() {
		t.Error("expected unified format, got legacy")
	}
	if p.SourcePath != path {
		t.Errorf("SourcePath = %q, want %q", p.SourcePath, path)
	}
	if p.SourceName != "dualsense" {
		t.Errorf("SourceName = %q, want %q", p.SourceName, "dualsense")
	}
	if p.DeviceName() != "DualSense" {
		t.Errorf("DeviceName = %q, want %q", p.DeviceName(), "DualSense")
	}
	if len(p.Mappings) != 2 {
		t.Fatalf("len(Mappings) = %d, want 2", len(p.Mappings))
	}
	if p.Mappings[0].Output.Type != OutputKey {
		t.Errorf("first mapping output type = %q, want %q", p.Mappings[0].Output.Type, OutputKey)
	}
}

func TestLoadMappingProfile_ValidLegacy(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "DualShock4.toml")
	if err := os.WriteFile(path, []byte(validLegacyTOML), 0644); err != nil {
		t.Fatal(err)
	}

	p, err := LoadMappingProfile(path)
	if err != nil {
		t.Fatalf("LoadMappingProfile: %v", err)
	}

	if !p.IsLegacyFormat() {
		t.Error("expected legacy format")
	}
	if p.IsUnifiedFormat() {
		t.Error("should not be unified format")
	}
	if len(p.Remap) != 3 {
		t.Errorf("len(Remap) = %d, want 3", len(p.Remap))
	}
	if len(p.Commands) != 1 {
		t.Errorf("len(Commands) = %d, want 1", len(p.Commands))
	}
	if len(p.Movements) != 2 {
		t.Errorf("len(Movements) = %d, want 2", len(p.Movements))
	}
}

func TestLoadMappingProfile_FileNotFound(t *testing.T) {
	_, err := LoadMappingProfile("/nonexistent/path/to/profile.toml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadMappingProfile_MalformedTOML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.toml")
	if err := os.WriteFile(path, []byte(malformedTOML), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadMappingProfile(path)
	if err == nil {
		t.Fatal("expected error for malformed TOML, got nil")
	}
}

// ---------------------------------------------------------------------------
// ParseMappingProfile
// ---------------------------------------------------------------------------

func TestParseMappingProfile_UnifiedMetadata(t *testing.T) {
	p, err := ParseMappingProfile(validUnifiedTOML, "dualsense.toml")
	if err != nil {
		t.Fatalf("ParseMappingProfile: %v", err)
	}

	if p.Profile == nil {
		t.Fatal("Profile metadata is nil")
	}
	if p.Profile.SchemaVersion != 1 {
		t.Errorf("SchemaVersion = %d, want 1", p.Profile.SchemaVersion)
	}
	if p.Profile.Type != "mapping" {
		t.Errorf("Type = %q, want %q", p.Profile.Type, "mapping")
	}
	if p.Profile.Device != "DualSense" {
		t.Errorf("Device = %q, want %q", p.Profile.Device, "DualSense")
	}
	if len(p.Profile.Tags) != 2 {
		t.Errorf("len(Tags) = %d, want 2", len(p.Profile.Tags))
	}
	if p.Profile.Description != "PS5 DualSense profile" {
		t.Errorf("Description = %q", p.Profile.Description)
	}
}

func TestParseMappingProfile_UnifiedDeviceConfig(t *testing.T) {
	p, err := ParseMappingProfile(validUnifiedTOML, "dualsense.toml")
	if err != nil {
		t.Fatalf("ParseMappingProfile: %v", err)
	}

	if p.Device == nil {
		t.Fatal("Device config is nil")
	}
	if p.Device.Name != "DualSense" {
		t.Errorf("Device.Name = %q, want %q", p.Device.Name, "DualSense")
	}
	if p.Device.Type != InputClassGamepad {
		t.Errorf("Device.Type = %q, want %q", p.Device.Type, InputClassGamepad)
	}
	if p.Device.VendorID != "054c" {
		t.Errorf("Device.VendorID = %q, want %q", p.Device.VendorID, "054c")
	}
}

func TestParseMappingProfile_UnifiedSettings(t *testing.T) {
	p, err := ParseMappingProfile(validUnifiedTOML, "dualsense.toml")
	if err != nil {
		t.Fatalf("ParseMappingProfile: %v", err)
	}

	if p.Settings == nil {
		t.Fatal("Settings is nil")
	}
	if !p.Settings.GrabDevice {
		t.Error("expected GrabDevice = true")
	}
	if len(p.Settings.CustomModifiers) != 2 {
		t.Errorf("len(CustomModifiers) = %d, want 2", len(p.Settings.CustomModifiers))
	}
}

func TestParseMappingProfile_UnifiedValueTransform(t *testing.T) {
	p, err := ParseMappingProfile(validUnifiedTOML, "dualsense.toml")
	if err != nil {
		t.Fatalf("ParseMappingProfile: %v", err)
	}

	// Second mapping has a value transform.
	if len(p.Mappings) < 2 {
		t.Fatalf("need at least 2 mappings, got %d", len(p.Mappings))
	}
	vt := p.Mappings[1].Value
	if vt == nil {
		t.Fatal("second mapping Value is nil")
	}
	if vt.InputRange != [2]float64{0, 255} {
		t.Errorf("InputRange = %v, want [0, 255]", vt.InputRange)
	}
	if vt.OutputRange != [2]float64{0, 1} {
		t.Errorf("OutputRange = %v, want [0, 1]", vt.OutputRange)
	}
	if vt.Curve != CurveLogarithmic {
		t.Errorf("Curve = %q, want %q", vt.Curve, CurveLogarithmic)
	}
}

func TestParseMappingProfile_LegacyAutoDetect(t *testing.T) {
	p, err := ParseMappingProfile(validLegacyTOML, "DualShock4.toml")
	if err != nil {
		t.Fatalf("ParseMappingProfile: %v", err)
	}

	if p.IsUnifiedFormat() {
		t.Error("legacy profile should not be unified format")
	}
	if !p.IsLegacyFormat() {
		t.Error("expected legacy format detection")
	}
	if p.SourceName != "DualShock4" {
		t.Errorf("SourceName = %q, want %q", p.SourceName, "DualShock4")
	}
}

func TestParseMappingProfile_LegacySettings(t *testing.T) {
	p, err := ParseMappingProfile(validLegacyTOML, "DualShock4.toml")
	if err != nil {
		t.Fatalf("ParseMappingProfile: %v", err)
	}

	if p.LegacySettings == nil {
		t.Fatal("LegacySettings is nil")
	}
	if p.LegacySettings["GRAB_DEVICE"] != "true" {
		t.Errorf("GRAB_DEVICE = %q, want %q", p.LegacySettings["GRAB_DEVICE"], "true")
	}
	if p.LegacySettings["LSTICK"] != "cursor" {
		t.Errorf("LSTICK = %q, want %q", p.LegacySettings["LSTICK"], "cursor")
	}
}

func TestParseMappingProfile_EmptyContent(t *testing.T) {
	p, err := ParseMappingProfile(emptyProfileTOML, "empty.toml")
	if err != nil {
		t.Fatalf("ParseMappingProfile: %v", err)
	}

	// Should not be unified or legacy.
	if p.IsUnifiedFormat() {
		t.Error("empty profile should not be unified")
	}
	if p.IsLegacyFormat() {
		t.Error("empty profile should not be legacy")
	}
}

func TestParseMappingProfile_MalformedTOML(t *testing.T) {
	_, err := ParseMappingProfile(malformedTOML, "bad.toml")
	if err == nil {
		t.Fatal("expected error for malformed TOML, got nil")
	}
}

// ---------------------------------------------------------------------------
// MappingProfile methods
// ---------------------------------------------------------------------------

func TestMappingProfile_MappingCount(t *testing.T) {
	tests := []struct {
		name  string
		toml  string
		count int
	}{
		{"unified_two_rules", validUnifiedTOML, 2},
		{"legacy_six_rules", validLegacyTOML, 6}, // 3 remap + 1 command + 2 movement
		{"empty", emptyProfileTOML, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := ParseMappingProfile(tt.toml, "test.toml")
			if err != nil {
				t.Fatalf("ParseMappingProfile: %v", err)
			}
			if got := p.MappingCount(); got != tt.count {
				t.Errorf("MappingCount() = %d, want %d", got, tt.count)
			}
		})
	}
}

func TestMappingProfile_DeviceName_Fallback(t *testing.T) {
	// No profile or device metadata -- falls back to SourceName.
	p := &MappingProfile{SourceName: "my-controller"}
	if got := p.DeviceName(); got != "my-controller" {
		t.Errorf("DeviceName() = %q, want %q", got, "my-controller")
	}

	// Device config present.
	p.Device = &DeviceConfig{Name: "MIDI Fighter"}
	if got := p.DeviceName(); got != "MIDI Fighter" {
		t.Errorf("DeviceName() = %q, want %q", got, "MIDI Fighter")
	}

	// Profile metadata overrides device config.
	p.Profile = &ProfileMeta{Device: "Profile Device"}
	if got := p.DeviceName(); got != "Profile Device" {
		t.Errorf("DeviceName() = %q, want %q", got, "Profile Device")
	}
}

// ---------------------------------------------------------------------------
// App override parsing
// ---------------------------------------------------------------------------

func TestParseMappingProfile_AppOverrides(t *testing.T) {
	toml := `
[profile]
schema_version = 1
device = "TestPad"

[[mapping]]
input = "BTN_SOUTH"
[mapping.output]
type = "key"
keys = ["KEY_ENTER"]

[[app_override]]
window_class = "firefox"
description = "Firefox overrides"

[[app_override.mapping]]
input = "BTN_SOUTH"
[app_override.mapping.output]
type = "command"
exec = ["xdotool", "key", "ctrl+t"]
`
	p, err := ParseMappingProfile(toml, "test.toml")
	if err != nil {
		t.Fatalf("ParseMappingProfile: %v", err)
	}

	if len(p.AppOverrides) != 1 {
		t.Fatalf("len(AppOverrides) = %d, want 1", len(p.AppOverrides))
	}
	ov := p.AppOverrides[0]
	if ov.WindowClass != "firefox" {
		t.Errorf("WindowClass = %q, want %q", ov.WindowClass, "firefox")
	}
	if len(ov.Mappings) != 1 {
		t.Fatalf("len(override Mappings) = %d, want 1", len(ov.Mappings))
	}
	if ov.Mappings[0].Output.Type != OutputCommand {
		t.Errorf("override output type = %q, want %q", ov.Mappings[0].Output.Type, OutputCommand)
	}

	// MappingCount should include override mappings.
	if got := p.MappingCount(); got != 2 {
		t.Errorf("MappingCount() = %d, want 2 (1 default + 1 override)", got)
	}
}
