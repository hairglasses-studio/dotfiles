package output

import (
	"encoding/binary"
	"math"
	"testing"
)

func TestBuildOSCMessage(t *testing.T) {
	msg := buildOSCMessage("/test/value", 0.75)

	// Verify address is at the start and null-padded to 4-byte boundary.
	addr := "/test/value"
	padded := len(addr) + 1 // null terminator
	for padded%4 != 0 {
		padded++
	}

	if len(msg) < padded {
		t.Fatalf("message too short: %d bytes", len(msg))
	}

	// Check the address pattern.
	gotAddr := string(msg[:len(addr)])
	if gotAddr != addr {
		t.Errorf("address: got %q, want %q", gotAddr, addr)
	}

	// Check null terminator after address.
	if msg[len(addr)] != 0 {
		t.Error("missing null terminator after address")
	}

	// Find the type tag string after address padding.
	typeTagStart := padded
	if string(msg[typeTagStart:typeTagStart+2]) != ",f" {
		t.Errorf("type tag: got %q, want \",f\"", string(msg[typeTagStart:typeTagStart+2]))
	}

	// Type tag padded to 4 bytes: ",f\0\0"
	floatStart := typeTagStart + 4

	// Read the float32 value (big-endian).
	bits := binary.BigEndian.Uint32(msg[floatStart : floatStart+4])
	got := math.Float32frombits(bits)
	if got != 0.75 {
		t.Errorf("float value: got %f, want 0.75", got)
	}
}

func TestOscString(t *testing.T) {
	tests := []struct {
		in  string
		len int // expected padded length
	}{
		{"", 4},          // 1 null byte + 3 pad = 4
		{"a", 4},         // 2 bytes + 2 pad = 4
		{"ab", 4},        // 3 bytes + 1 pad = 4
		{"abc", 4},       // 4 bytes + 0 pad = 4
		{"abcd", 8},      // 5 bytes + 3 pad = 8
		{"/test", 8},     // 6 bytes + 2 pad = 8
	}
	for _, tt := range tests {
		got := oscString(tt.in)
		if len(got) != tt.len {
			t.Errorf("oscString(%q): len=%d, want %d", tt.in, len(got), tt.len)
		}
		if len(got)%4 != 0 {
			t.Errorf("oscString(%q): not 4-byte aligned (%d)", tt.in, len(got))
		}
	}
}
