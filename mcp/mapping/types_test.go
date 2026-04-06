package mapping

import (
	"math"
	"testing"
)

// ---------------------------------------------------------------------------
// ValueTransform.Transform — all curve types
// ---------------------------------------------------------------------------

func TestTransform_Linear(t *testing.T) {
	vt := &ValueTransform{
		InputRange:  [2]float64{0, 100},
		OutputRange: [2]float64{0, 1},
		Curve:       CurveLinear,
	}

	tests := []struct {
		raw  float64
		want float64
	}{
		{0, 0},
		{50, 0.5},
		{100, 1.0},
		{-10, 0},   // clamped
		{200, 1.0}, // clamped
	}
	for _, tt := range tests {
		got := vt.Transform(tt.raw)
		if math.Abs(got-tt.want) > 0.001 {
			t.Errorf("Transform(%v) = %v, want %v", tt.raw, got, tt.want)
		}
	}
}

func TestTransform_Logarithmic(t *testing.T) {
	vt := &ValueTransform{
		InputRange:  [2]float64{0, 127},
		OutputRange: [2]float64{0, 1},
		Curve:       CurveLogarithmic,
	}

	// Logarithmic curve: midpoint should be > 0.5 (log compresses high values).
	mid := vt.Transform(63.5)
	if mid <= 0.5 {
		t.Errorf("log Transform(63.5) = %v, expected > 0.5", mid)
	}

	// Endpoints should still be 0 and 1.
	if got := vt.Transform(0); math.Abs(got) > 0.001 {
		t.Errorf("log Transform(0) = %v, want ~0", got)
	}
	if got := vt.Transform(127); math.Abs(got-1.0) > 0.001 {
		t.Errorf("log Transform(127) = %v, want ~1.0", got)
	}
}

func TestTransform_Exponential(t *testing.T) {
	vt := &ValueTransform{
		InputRange:  [2]float64{0, 127},
		OutputRange: [2]float64{0, 1},
		Curve:       CurveExponential,
	}

	// Exponential curve: midpoint should be < 0.5 (exp compresses low values).
	mid := vt.Transform(63.5)
	if mid >= 0.5 {
		t.Errorf("exp Transform(63.5) = %v, expected < 0.5", mid)
	}

	// Endpoints.
	if got := vt.Transform(0); math.Abs(got) > 0.001 {
		t.Errorf("exp Transform(0) = %v, want ~0", got)
	}
	if got := vt.Transform(127); math.Abs(got-1.0) > 0.001 {
		t.Errorf("exp Transform(127) = %v, want ~1.0", got)
	}
}

func TestTransform_SCurve(t *testing.T) {
	vt := &ValueTransform{
		InputRange:  [2]float64{0, 127},
		OutputRange: [2]float64{0, 1},
		Curve:       CurveSCurve,
	}

	// S-curve midpoint should be ~0.5.
	mid := vt.Transform(63.5)
	if math.Abs(mid-0.5) > 0.05 {
		t.Errorf("scurve Transform(63.5) = %v, want ~0.5", mid)
	}

	// Endpoints.
	if got := vt.Transform(0); math.Abs(got) > 0.001 {
		t.Errorf("scurve Transform(0) = %v, want ~0", got)
	}
	if got := vt.Transform(127); math.Abs(got-1.0) > 0.001 {
		t.Errorf("scurve Transform(127) = %v, want ~1.0", got)
	}
}

func TestTransform_SCurve_CustomParam(t *testing.T) {
	vt := &ValueTransform{
		InputRange:  [2]float64{0, 100},
		OutputRange: [2]float64{0, 1},
		Curve:       CurveSCurve,
		CurveParam:  5.0, // not the default 2.0
	}

	// With high param, the S-curve is more pronounced.
	// Low values should be compressed (closer to 0).
	low := vt.Transform(20)
	if low >= 0.2 {
		t.Errorf("scurve(param=5) Transform(20) = %v, expected < 0.2", low)
	}

	// Midpoint should still be ~0.5 for symmetric S-curve.
	mid := vt.Transform(50)
	if math.Abs(mid-0.5) > 0.05 {
		t.Errorf("scurve(param=5) Transform(50) = %v, want ~0.5", mid)
	}
}

// ---------------------------------------------------------------------------
// ValueTransform.Transform — custom ranges
// ---------------------------------------------------------------------------

func TestTransform_CustomOutputRange(t *testing.T) {
	vt := &ValueTransform{
		InputRange:  [2]float64{0, 255},
		OutputRange: [2]float64{-1, 1}, // bipolar output
		Curve:       CurveLinear,
	}

	tests := []struct {
		raw  float64
		want float64
	}{
		{0, -1},
		{127.5, 0},
		{255, 1},
	}
	for _, tt := range tests {
		got := vt.Transform(tt.raw)
		if math.Abs(got-tt.want) > 0.01 {
			t.Errorf("Transform(%v) = %v, want %v", tt.raw, got, tt.want)
		}
	}
}

func TestTransform_NonZeroInputStart(t *testing.T) {
	vt := &ValueTransform{
		InputRange:  [2]float64{100, 200},
		OutputRange: [2]float64{0, 10},
		Curve:       CurveLinear,
	}

	if got := vt.Transform(100); math.Abs(got) > 0.001 {
		t.Errorf("Transform(100) = %v, want 0", got)
	}
	if got := vt.Transform(150); math.Abs(got-5) > 0.001 {
		t.Errorf("Transform(150) = %v, want 5", got)
	}
	if got := vt.Transform(200); math.Abs(got-10) > 0.001 {
		t.Errorf("Transform(200) = %v, want 10", got)
	}
}

func TestTransform_EqualInputRange(t *testing.T) {
	vt := &ValueTransform{
		InputRange:  [2]float64{50, 50}, // degenerate
		OutputRange: [2]float64{0, 1},
		Curve:       CurveLinear,
	}

	// Should return output min when input range is zero-width.
	got := vt.Transform(50)
	if got != 0 {
		t.Errorf("Transform(50) with equal input range = %v, want 0 (output min)", got)
	}
}

// ---------------------------------------------------------------------------
// ValueTransform.Transform — threshold behavior
// ---------------------------------------------------------------------------

func TestTransform_Threshold(t *testing.T) {
	vt := &ValueTransform{
		InputRange:  [2]float64{0, 255},
		OutputRange: [2]float64{0, 1},
		Curve:       CurveLinear,
		Threshold:   0.5,
	}

	// Threshold is stored but not automatically applied by Transform.
	// Transform returns the continuous value; threshold is checked by the caller.
	// Just verify Transform still works correctly with threshold set.
	got := vt.Transform(127.5)
	if math.Abs(got-0.5) > 0.01 {
		t.Errorf("Transform(127.5) = %v, want ~0.5", got)
	}
}

// ---------------------------------------------------------------------------
// ApplyCurve — edge cases
// ---------------------------------------------------------------------------

func TestApplyCurve_LogZero(t *testing.T) {
	vt := &ValueTransform{Curve: CurveLogarithmic}
	got := vt.ApplyCurve(0)
	if got != 0 {
		t.Errorf("ApplyCurve(0) for log = %v, want 0", got)
	}
}

func TestApplyCurve_LogNegative(t *testing.T) {
	vt := &ValueTransform{Curve: CurveLogarithmic}
	got := vt.ApplyCurve(-1)
	if got != 0 {
		t.Errorf("ApplyCurve(-1) for log = %v, want 0", got)
	}
}

func TestApplyCurve_DefaultParam(t *testing.T) {
	// CurveParam=0 should use default of 2.0.
	vt := &ValueTransform{Curve: CurveExponential, CurveParam: 0}
	got := vt.ApplyCurve(0.5)
	// Should be the same as CurveParam=2.0.
	vt2 := &ValueTransform{Curve: CurveExponential, CurveParam: 2.0}
	want := vt2.ApplyCurve(0.5)
	if math.Abs(got-want) > 0.001 {
		t.Errorf("ApplyCurve with CurveParam=0 gave %v, with CurveParam=2 gave %v", got, want)
	}
}

// ---------------------------------------------------------------------------
// Condition.Evaluate
// ---------------------------------------------------------------------------

func TestCondition_Equals(t *testing.T) {
	tests := []struct {
		name   string
		cond   Condition
		vars   map[string]any
		expect bool
	}{
		{
			name:   "string match",
			cond:   Condition{Variable: "mode", Equals: "combat"},
			vars:   map[string]any{"mode": "combat"},
			expect: true,
		},
		{
			name:   "string mismatch",
			cond:   Condition{Variable: "mode", Equals: "combat"},
			vars:   map[string]any{"mode": "explore"},
			expect: false,
		},
		{
			name:   "int match via sprintf",
			cond:   Condition{Variable: "count", Equals: 5},
			vars:   map[string]any{"count": 5},
			expect: true,
		},
		{
			name:   "missing variable",
			cond:   Condition{Variable: "mode", Equals: "combat"},
			vars:   map[string]any{},
			expect: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cond.Evaluate(tt.vars); got != tt.expect {
				t.Errorf("Evaluate() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestCondition_NotEqual(t *testing.T) {
	tests := []struct {
		name   string
		cond   Condition
		vars   map[string]any
		expect bool
	}{
		{
			name:   "different value",
			cond:   Condition{Variable: "mode", NotEqual: "combat"},
			vars:   map[string]any{"mode": "explore"},
			expect: true,
		},
		{
			name:   "same value",
			cond:   Condition{Variable: "mode", NotEqual: "combat"},
			vars:   map[string]any{"mode": "combat"},
			expect: false,
		},
		{
			name:   "missing variable",
			cond:   Condition{Variable: "mode", NotEqual: "combat"},
			vars:   map[string]any{},
			expect: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cond.Evaluate(tt.vars); got != tt.expect {
				t.Errorf("Evaluate() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestCondition_GreaterThan(t *testing.T) {
	tests := []struct {
		name   string
		cond   Condition
		vars   map[string]any
		expect bool
	}{
		{
			name:   "above threshold",
			cond:   Condition{Variable: "score", GreaterThan: 50},
			vars:   map[string]any{"score": 75.0},
			expect: true,
		},
		{
			name:   "equal to threshold",
			cond:   Condition{Variable: "score", GreaterThan: 50},
			vars:   map[string]any{"score": 50.0},
			expect: false, // not strictly greater
		},
		{
			name:   "below threshold",
			cond:   Condition{Variable: "score", GreaterThan: 50},
			vars:   map[string]any{"score": 25.0},
			expect: false,
		},
		{
			name:   "int value",
			cond:   Condition{Variable: "score", GreaterThan: 50},
			vars:   map[string]any{"score": 75},
			expect: true,
		},
		{
			name:   "string value cannot compare",
			cond:   Condition{Variable: "score", GreaterThan: 50},
			vars:   map[string]any{"score": "high"},
			expect: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cond.Evaluate(tt.vars); got != tt.expect {
				t.Errorf("Evaluate() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestCondition_LessThan(t *testing.T) {
	tests := []struct {
		name   string
		cond   Condition
		vars   map[string]any
		expect bool
	}{
		{
			name:   "below threshold",
			cond:   Condition{Variable: "health", LessThan: 20},
			vars:   map[string]any{"health": 10.0},
			expect: true,
		},
		{
			name:   "equal to threshold",
			cond:   Condition{Variable: "health", LessThan: 20},
			vars:   map[string]any{"health": 20.0},
			expect: false, // not strictly less
		},
		{
			name:   "above threshold",
			cond:   Condition{Variable: "health", LessThan: 20},
			vars:   map[string]any{"health": 50.0},
			expect: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cond.Evaluate(tt.vars); got != tt.expect {
				t.Errorf("Evaluate() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestCondition_CombinedGreaterAndLess(t *testing.T) {
	// Both GreaterThan and LessThan set: acts as a range check.
	cond := Condition{Variable: "val", GreaterThan: 10, LessThan: 100}

	tests := []struct {
		name   string
		val    any
		expect bool
	}{
		{"in range", 50.0, true},
		{"at lower bound", 10.0, false},
		{"at upper bound", 100.0, false},
		{"below range", 5.0, false},
		{"above range", 150.0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vars := map[string]any{"val": tt.val}
			if got := cond.Evaluate(vars); got != tt.expect {
				t.Errorf("Evaluate() = %v, want %v (val=%v)", got, tt.expect, tt.val)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ToFloat64
// ---------------------------------------------------------------------------

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name   string
		input  any
		want   float64
		wantOK bool
	}{
		{"float64", 3.14, 3.14, true},
		{"float32", float32(2.5), 2.5, true},
		{"int", 42, 42.0, true},
		{"int64", int64(100), 100.0, true},
		{"string", "hello", 0, false},
		{"nil", nil, 0, false},
		{"bool", true, 0, false},
		{"slice", []int{1, 2}, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := ToFloat64(tt.input)
			if ok != tt.wantOK {
				t.Errorf("ToFloat64(%v) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if ok && math.Abs(got-tt.want) > 0.001 {
				t.Errorf("ToFloat64(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// MappingProfile format detection
// ---------------------------------------------------------------------------

func TestMappingProfile_IsUnifiedFormat(t *testing.T) {
	tests := []struct {
		name    string
		profile *MappingProfile
		want    bool
	}{
		{
			"with profile and schema",
			&MappingProfile{Profile: &ProfileMeta{SchemaVersion: 1}},
			true,
		},
		{
			"with profile but zero schema",
			&MappingProfile{Profile: &ProfileMeta{SchemaVersion: 0}},
			false,
		},
		{
			"nil profile",
			&MappingProfile{},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.profile.IsUnifiedFormat(); got != tt.want {
				t.Errorf("IsUnifiedFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMappingProfile_IsLegacyFormat(t *testing.T) {
	tests := []struct {
		name    string
		profile *MappingProfile
		want    bool
	}{
		{
			"has remap",
			&MappingProfile{Remap: map[string][]string{"BTN_SOUTH": {"KEY_A"}}},
			true,
		},
		{
			"has commands",
			&MappingProfile{Commands: map[string][]string{"BTN_NORTH": {"echo"}}},
			true,
		},
		{
			"has movements",
			&MappingProfile{Movements: map[string]string{"ABS_X": "cursor"}},
			true,
		},
		{
			"empty profile",
			&MappingProfile{},
			false,
		},
		{
			"unified profile with remap should be unified",
			&MappingProfile{
				Profile: &ProfileMeta{SchemaVersion: 1},
				Remap:   map[string][]string{"BTN_SOUTH": {"KEY_A"}},
			},
			false, // unified takes precedence
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.profile.IsLegacyFormat(); got != tt.want {
				t.Errorf("IsLegacyFormat() = %v, want %v", got, tt.want)
			}
		})
	}
}
