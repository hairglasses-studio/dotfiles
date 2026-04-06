package mapping

import (
	"os"
	"path/filepath"
	"testing"
)

// ---------------------------------------------------------------------------
// ListMappingProfiles — basic scan
// ---------------------------------------------------------------------------

func TestListMappingProfiles_BasicScan(t *testing.T) {
	dir := t.TempDir()

	// Write a unified profile.
	unified := `
[profile]
schema_version = 1
device = "DualSense"
tags = ["gamepad"]
description = "test profile"

[[mapping]]
input = "BTN_SOUTH"
[mapping.output]
type = "key"
keys = ["KEY_A"]
`
	if err := os.WriteFile(filepath.Join(dir, "dualsense.toml"), []byte(unified), 0644); err != nil {
		t.Fatal(err)
	}

	// Write a legacy profile.
	legacy := `
[remap]
BTN_SOUTH = ["KEY_Z"]
BTN_EAST = ["KEY_X"]
`
	if err := os.WriteFile(filepath.Join(dir, "dualshock4.toml"), []byte(legacy), 0644); err != nil {
		t.Fatal(err)
	}

	summaries, err := ListMappingProfiles(dir)
	if err != nil {
		t.Fatalf("ListMappingProfiles: %v", err)
	}

	if len(summaries) != 2 {
		t.Fatalf("len(summaries) = %d, want 2", len(summaries))
	}

	// Build a map for easier lookup.
	byName := make(map[string]MappingProfileSummary)
	for _, s := range summaries {
		byName[s.Name] = s
	}

	// Check unified profile.
	ds, ok := byName["dualsense"]
	if !ok {
		t.Fatal("missing summary for dualsense")
	}
	if ds.Format != "unified" {
		t.Errorf("dualsense format = %q, want %q", ds.Format, "unified")
	}
	if ds.DeviceName != "DualSense" {
		t.Errorf("dualsense DeviceName = %q, want %q", ds.DeviceName, "DualSense")
	}
	if ds.MappingCount != 1 {
		t.Errorf("dualsense MappingCount = %d, want 1", ds.MappingCount)
	}
	if ds.Description != "test profile" {
		t.Errorf("dualsense Description = %q, want %q", ds.Description, "test profile")
	}
	if len(ds.Tags) != 1 || ds.Tags[0] != "gamepad" {
		t.Errorf("dualsense Tags = %v, want [gamepad]", ds.Tags)
	}

	// Check legacy profile.
	ds4, ok := byName["dualshock4"]
	if !ok {
		t.Fatal("missing summary for dualshock4")
	}
	if ds4.Format != "legacy" {
		t.Errorf("dualshock4 format = %q, want %q", ds4.Format, "legacy")
	}
	if ds4.MappingCount != 2 {
		t.Errorf("dualshock4 MappingCount = %d, want 2", ds4.MappingCount)
	}
}

// ---------------------------------------------------------------------------
// ListMappingProfiles — empty directory
// ---------------------------------------------------------------------------

func TestListMappingProfiles_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	summaries, err := ListMappingProfiles(dir)
	if err != nil {
		t.Fatalf("ListMappingProfiles: %v", err)
	}
	if len(summaries) != 0 {
		t.Errorf("len(summaries) = %d, want 0", len(summaries))
	}
}

// ---------------------------------------------------------------------------
// ListMappingProfiles — non-existent directory
// ---------------------------------------------------------------------------

func TestListMappingProfiles_NonExistentDir(t *testing.T) {
	summaries, err := ListMappingProfiles("/nonexistent/path/to/profiles")
	if err != nil {
		t.Fatalf("expected no error for non-existent dir (IsNotExist handled), got: %v", err)
	}
	if len(summaries) != 0 {
		t.Errorf("len(summaries) = %d, want 0 for non-existent dir", len(summaries))
	}
}

// ---------------------------------------------------------------------------
// ListMappingProfiles — ignores non-TOML files and hidden files
// ---------------------------------------------------------------------------

func TestListMappingProfiles_IgnoresNonTOML(t *testing.T) {
	dir := t.TempDir()

	// Non-TOML files that should be skipped.
	for _, name := range []string{"readme.md", "notes.txt", "config.json", ".hidden.toml"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// One valid TOML.
	toml := `
[remap]
BTN_SOUTH = ["KEY_A"]
`
	if err := os.WriteFile(filepath.Join(dir, "valid.toml"), []byte(toml), 0644); err != nil {
		t.Fatal(err)
	}

	summaries, err := ListMappingProfiles(dir)
	if err != nil {
		t.Fatalf("ListMappingProfiles: %v", err)
	}
	if len(summaries) != 1 {
		t.Errorf("len(summaries) = %d, want 1 (only valid.toml)", len(summaries))
	}
}

// ---------------------------------------------------------------------------
// ListMappingProfiles — ignores subdirectories
// ---------------------------------------------------------------------------

func TestListMappingProfiles_IgnoresSubdirs(t *testing.T) {
	dir := t.TempDir()

	// Create a subdirectory (should be ignored).
	if err := os.Mkdir(filepath.Join(dir, "subdir"), 0755); err != nil {
		t.Fatal(err)
	}

	summaries, err := ListMappingProfiles(dir)
	if err != nil {
		t.Fatalf("ListMappingProfiles: %v", err)
	}
	if len(summaries) != 0 {
		t.Errorf("len(summaries) = %d, want 0 (no TOML files)", len(summaries))
	}
}

// ---------------------------------------------------------------------------
// ListMappingProfiles — malformed TOML files get error summaries
// ---------------------------------------------------------------------------

func TestListMappingProfiles_MalformedTOML(t *testing.T) {
	dir := t.TempDir()

	// Malformed TOML.
	if err := os.WriteFile(filepath.Join(dir, "broken.toml"), []byte("[bad\nnot valid"), 0644); err != nil {
		t.Fatal(err)
	}

	// Valid TOML.
	valid := `
[remap]
BTN_SOUTH = ["KEY_A"]
`
	if err := os.WriteFile(filepath.Join(dir, "good.toml"), []byte(valid), 0644); err != nil {
		t.Fatal(err)
	}

	summaries, err := ListMappingProfiles(dir)
	if err != nil {
		t.Fatalf("ListMappingProfiles: %v", err)
	}
	if len(summaries) != 2 {
		t.Fatalf("len(summaries) = %d, want 2", len(summaries))
	}

	byName := make(map[string]MappingProfileSummary)
	for _, s := range summaries {
		byName[s.Name] = s
	}

	broken, ok := byName["broken"]
	if !ok {
		t.Fatal("missing summary for broken")
	}
	if broken.Format != "error" {
		t.Errorf("broken format = %q, want %q", broken.Format, "error")
	}
	if broken.Error == "" {
		t.Error("expected non-empty Error for broken profile")
	}

	good, ok := byName["good"]
	if !ok {
		t.Fatal("missing summary for good")
	}
	if good.Format == "error" {
		t.Errorf("good profile should not have error format, got error: %s", good.Error)
	}
}

// ---------------------------------------------------------------------------
// ProfileToSummary
// ---------------------------------------------------------------------------

func TestProfileToSummary_Unified(t *testing.T) {
	p := &MappingProfile{
		SourceName: "test-profile",
		SourcePath: "/path/to/test-profile.toml",
		Profile: &ProfileMeta{
			SchemaVersion: 1,
			Device:        "MyDevice",
			Tags:          []string{"midi", "ableton"},
			Description:   "A test profile",
			AppClass:      "ableton",
		},
		Mappings: []MappingRule{
			{Input: "CC1", Output: OutputAction{Type: OutputOSC}},
			{Input: "CC2", Output: OutputAction{Type: OutputOSC}},
		},
	}

	s := ProfileToSummary(p)

	if s.Name != "test-profile" {
		t.Errorf("Name = %q, want %q", s.Name, "test-profile")
	}
	if s.Format != "unified" {
		t.Errorf("Format = %q, want %q", s.Format, "unified")
	}
	if s.DeviceName != "MyDevice" {
		t.Errorf("DeviceName = %q, want %q", s.DeviceName, "MyDevice")
	}
	if s.MappingCount != 2 {
		t.Errorf("MappingCount = %d, want 2", s.MappingCount)
	}
	if s.AppClass != "ableton" {
		t.Errorf("AppClass = %q, want %q", s.AppClass, "ableton")
	}
	if s.Description != "A test profile" {
		t.Errorf("Description = %q, want %q", s.Description, "A test profile")
	}
	if len(s.Tags) != 2 {
		t.Errorf("len(Tags) = %d, want 2", len(s.Tags))
	}
}

func TestProfileToSummary_LegacyWithAppClass(t *testing.T) {
	p := &MappingProfile{
		SourceName: "DualShock4::firefox",
		SourcePath: "/path/to/DualShock4::firefox.toml",
		Remap: map[string][]string{
			"BTN_SOUTH": {"KEY_ENTER"},
		},
	}

	s := ProfileToSummary(p)

	if s.Format != "legacy" {
		t.Errorf("Format = %q, want %q", s.Format, "legacy")
	}
	if s.AppClass != "firefox" {
		t.Errorf("AppClass = %q, want %q", s.AppClass, "firefox")
	}
}

func TestProfileToSummary_LegacyWithLayer(t *testing.T) {
	p := &MappingProfile{
		SourceName: "DualShock4::2",
		SourcePath: "/path/to/DualShock4::2.toml",
		Remap: map[string][]string{
			"BTN_SOUTH": {"KEY_A"},
		},
	}

	s := ProfileToSummary(p)

	// Numeric suffix is a layer, not an app class.
	if s.AppClass != "" {
		t.Errorf("AppClass = %q, want empty (numeric = layer, not app class)", s.AppClass)
	}
}
