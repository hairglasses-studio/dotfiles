package mapping

import (
	"testing"
)

// ---------------------------------------------------------------------------
// ValidateProfile — unified format
// ---------------------------------------------------------------------------

func TestValidateProfile_ValidUnified(t *testing.T) {
	p := &MappingProfile{
		Profile: &ProfileMeta{SchemaVersion: 1, Device: "TestPad"},
		Device:  &DeviceConfig{Name: "TestPad", Type: InputClassGamepad},
		Mappings: []MappingRule{
			{
				Input:  "BTN_SOUTH",
				Output: OutputAction{Type: OutputKey, Keys: []string{"KEY_A"}},
			},
		},
	}

	issues := ValidateProfile(p)
	for _, issue := range issues {
		if issue.Severity == "error" {
			t.Errorf("unexpected error: %s — %s", issue.Field, issue.Message)
		}
	}
}

func TestValidateProfile_MissingInput(t *testing.T) {
	p := &MappingProfile{
		Profile: &ProfileMeta{SchemaVersion: 1},
		Mappings: []MappingRule{
			{
				Input:  "", // missing
				Output: OutputAction{Type: OutputKey, Keys: []string{"KEY_A"}},
			},
		},
	}

	issues := ValidateProfile(p)
	found := false
	for _, issue := range issues {
		if issue.Severity == "error" && issue.Field == "mapping[0].input" {
			found = true
		}
	}
	if !found {
		t.Error("expected error for missing input field")
	}
}

func TestValidateProfile_MissingOutputType(t *testing.T) {
	p := &MappingProfile{
		Profile: &ProfileMeta{SchemaVersion: 1},
		Mappings: []MappingRule{
			{
				Input:  "BTN_SOUTH",
				Output: OutputAction{Type: ""}, // missing
			},
		},
	}

	issues := ValidateProfile(p)
	found := false
	for _, issue := range issues {
		if issue.Severity == "error" && issue.Field == "mapping[0].output.type" {
			found = true
		}
	}
	if !found {
		t.Error("expected error for missing output type")
	}
}

func TestValidateProfile_SchemaVersionZero_NotUnified(t *testing.T) {
	// SchemaVersion 0 means IsUnifiedFormat() returns false, so it falls
	// through to the empty/unrecognized branch.
	p := &MappingProfile{
		Profile: &ProfileMeta{SchemaVersion: 0},
		Mappings: []MappingRule{
			{
				Input:  "BTN_SOUTH",
				Output: OutputAction{Type: OutputKey, Keys: []string{"KEY_A"}},
			},
		},
	}

	issues := ValidateProfile(p)
	found := false
	for _, issue := range issues {
		if issue.Severity == "warning" && issue.Field == "format" {
			found = true
		}
	}
	if !found {
		t.Error("expected warning about unrecognized format for schema_version=0")
	}
}

func TestValidateProfile_InvalidCurveType(t *testing.T) {
	p := &MappingProfile{
		Profile: &ProfileMeta{SchemaVersion: 1},
		Mappings: []MappingRule{
			{
				Input:  "ABS_Z",
				Output: OutputAction{Type: OutputOSC, Address: "/fader"},
				Value: &ValueTransform{
					Curve: CurveType("bezier"), // invalid
				},
			},
		},
	}

	issues := ValidateProfile(p)
	found := false
	for _, issue := range issues {
		if issue.Severity == "error" && issue.Field == "mapping[0].value.curve" {
			found = true
		}
	}
	if !found {
		t.Error("expected error for unknown curve type")
	}
}

func TestValidateProfile_ValidCurveTypes(t *testing.T) {
	curves := []CurveType{
		"",
		CurveLinear,
		CurveLogarithmic,
		CurveExponential,
		CurveSCurve,
	}
	for _, curve := range curves {
		t.Run(string(curve), func(t *testing.T) {
			p := &MappingProfile{
				Profile: &ProfileMeta{SchemaVersion: 1},
				Mappings: []MappingRule{
					{
						Input:  "ABS_Z",
						Output: OutputAction{Type: OutputOSC},
						Value: &ValueTransform{
							Curve: curve,
						},
					},
				},
			}
			issues := ValidateProfile(p)
			for _, issue := range issues {
				if issue.Severity == "error" && issue.Field == "mapping[0].value.curve" {
					t.Errorf("curve %q should be valid, got error: %s", curve, issue.Message)
				}
			}
		})
	}
}

func TestValidateProfile_EqualInputRange_Warning(t *testing.T) {
	p := &MappingProfile{
		Profile: &ProfileMeta{SchemaVersion: 1},
		Mappings: []MappingRule{
			{
				Input:  "ABS_Z",
				Output: OutputAction{Type: OutputOSC},
				Value: &ValueTransform{
					InputRange: [2]float64{50, 50}, // min == max, nonzero
				},
			},
		},
	}

	issues := ValidateProfile(p)
	found := false
	for _, issue := range issues {
		if issue.Severity == "warning" && issue.Field == "mapping[0].value.input_range" {
			found = true
		}
	}
	if !found {
		t.Error("expected warning for input_range with min == max")
	}
}

func TestValidateProfile_ZeroInputRange_NoWarning(t *testing.T) {
	p := &MappingProfile{
		Profile: &ProfileMeta{SchemaVersion: 1},
		Mappings: []MappingRule{
			{
				Input:  "ABS_Z",
				Output: OutputAction{Type: OutputOSC},
				Value: &ValueTransform{
					InputRange: [2]float64{0, 0}, // zero defaults are fine
				},
			},
		},
	}

	issues := ValidateProfile(p)
	for _, issue := range issues {
		if issue.Severity == "warning" && issue.Field == "mapping[0].value.input_range" {
			t.Error("should not warn for zero input_range (default)")
		}
	}
}

func TestValidateProfile_EmptyMappings(t *testing.T) {
	p := &MappingProfile{
		Profile:  &ProfileMeta{SchemaVersion: 1},
		Mappings: []MappingRule{},
	}

	issues := ValidateProfile(p)
	// No errors expected for empty but valid unified profile.
	for _, issue := range issues {
		if issue.Severity == "error" {
			t.Errorf("unexpected error for empty mappings: %s — %s", issue.Field, issue.Message)
		}
	}
}

// ---------------------------------------------------------------------------
// ValidateProfile — legacy format
// ---------------------------------------------------------------------------

func TestValidateProfile_LegacyFormat(t *testing.T) {
	p := &MappingProfile{
		Remap: map[string][]string{
			"BTN_SOUTH": {"KEY_A"},
		},
	}

	issues := ValidateProfile(p)
	found := false
	for _, issue := range issues {
		if issue.Severity == "info" && issue.Field == "format" {
			found = true
		}
	}
	if !found {
		t.Error("expected info about legacy format detection")
	}
}

// ---------------------------------------------------------------------------
// ValidateProfile — empty/unrecognized format
// ---------------------------------------------------------------------------

func TestValidateProfile_EmptyProfile(t *testing.T) {
	p := &MappingProfile{}

	issues := ValidateProfile(p)
	found := false
	for _, issue := range issues {
		if issue.Severity == "warning" && issue.Field == "format" {
			found = true
		}
	}
	if !found {
		t.Error("expected warning about empty/unrecognized format")
	}
}

// ---------------------------------------------------------------------------
// ValidateProfile — app override issues
// ---------------------------------------------------------------------------

func TestValidateProfile_AppOverrideMissingWindowClass(t *testing.T) {
	p := &MappingProfile{
		Profile: &ProfileMeta{SchemaVersion: 1},
		AppOverrides: []AppOverride{
			{
				WindowClass: "", // missing
				Mappings: []MappingRule{
					{Input: "BTN_SOUTH", Output: OutputAction{Type: OutputKey}},
				},
			},
		},
	}

	issues := ValidateProfile(p)
	found := false
	for _, issue := range issues {
		if issue.Severity == "error" && issue.Field == "app_override[0].window_class" {
			found = true
		}
	}
	if !found {
		t.Error("expected error for missing window_class in app_override")
	}
}

func TestValidateProfile_AppOverrideMissingInput(t *testing.T) {
	p := &MappingProfile{
		Profile: &ProfileMeta{SchemaVersion: 1},
		AppOverrides: []AppOverride{
			{
				WindowClass: "firefox",
				Mappings: []MappingRule{
					{Input: "", Output: OutputAction{Type: OutputKey}}, // missing input
				},
			},
		},
	}

	issues := ValidateProfile(p)
	found := false
	for _, issue := range issues {
		if issue.Severity == "error" && issue.Field == "app_override[0].mapping[0].input" {
			found = true
		}
	}
	if !found {
		t.Error("expected error for missing input in app_override mapping")
	}
}

// ---------------------------------------------------------------------------
// ValidateValueTransform — nil
// ---------------------------------------------------------------------------

func TestValidateValueTransform_Nil(t *testing.T) {
	issues := ValidateValueTransform(nil, "test")
	if len(issues) != 0 {
		t.Errorf("expected no issues for nil ValueTransform, got %d", len(issues))
	}
}

// ---------------------------------------------------------------------------
// Multiple issues at once
// ---------------------------------------------------------------------------

func TestValidateProfile_MultipleIssues(t *testing.T) {
	p := &MappingProfile{
		Profile: &ProfileMeta{SchemaVersion: 1},
		Mappings: []MappingRule{
			{Input: "", Output: OutputAction{Type: ""}},  // two errors: missing input + missing output.type
			{Input: "BTN_SOUTH", Output: OutputAction{Type: OutputKey}}, // valid
			{Input: "ABS_Z", Output: OutputAction{Type: OutputOSC}, Value: &ValueTransform{Curve: "invalid"}}, // error: invalid curve
		},
	}

	issues := ValidateProfile(p)
	errorCount := 0
	for _, issue := range issues {
		if issue.Severity == "error" {
			errorCount++
		}
	}
	// missing input + missing output.type + invalid curve = 3
	if errorCount != 3 {
		t.Errorf("expected 3 errors, got %d", errorCount)
		for _, issue := range issues {
			t.Logf("  %s: %s — %s", issue.Severity, issue.Field, issue.Message)
		}
	}
}
