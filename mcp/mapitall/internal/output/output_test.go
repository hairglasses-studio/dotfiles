package output

import (
	"testing"

	"github.com/hairglasses-studio/mapping"
)

// mockTarget is a test double for Target.
type mockTarget struct {
	typ      mapping.OutputType
	calls    []mockCall
	closeErr error
}

type mockCall struct {
	action mapping.OutputAction
	value  float64
}

func (m *mockTarget) Type() mapping.OutputType { return m.typ }
func (m *mockTarget) Execute(action mapping.OutputAction, value float64) error {
	m.calls = append(m.calls, mockCall{action: action, value: value})
	return nil
}
func (m *mockTarget) Close() error { return m.closeErr }

func TestRegistryRegisterAndGet(t *testing.T) {
	r := NewRegistry()
	mt := &mockTarget{typ: mapping.OutputCommand}
	r.Register(mt)

	got := r.Get(mapping.OutputCommand)
	if got != mt {
		t.Fatal("expected registered target back")
	}

	if r.Get(mapping.OutputOSC) != nil {
		t.Fatal("expected nil for unregistered type")
	}
}

func TestRegistryClose(t *testing.T) {
	r := NewRegistry()
	mt := &mockTarget{typ: mapping.OutputCommand}
	r.Register(mt)

	if err := r.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
}

func TestResolveKeyCode(t *testing.T) {
	tests := []struct {
		name string
		code uint16
		ok   bool
	}{
		{"KEY_ENTER", 28, true},
		{"KEY_A", 30, true},
		{"KEY_SPACE", 57, true},
		{"BTN_SOUTH", 0x130, true},
		{"BTN_MODE", 0x13c, true},
		{"BOGUS_KEY", 0, false},
	}
	for _, tt := range tests {
		code, ok := ResolveKeyCode(tt.name)
		if ok != tt.ok {
			t.Errorf("ResolveKeyCode(%q): ok=%v, want %v", tt.name, ok, tt.ok)
		}
		if code != tt.code {
			t.Errorf("ResolveKeyCode(%q): code=%d, want %d", tt.name, code, tt.code)
		}
	}
}
