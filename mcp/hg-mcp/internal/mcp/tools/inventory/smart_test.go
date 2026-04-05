package inventory

import (
	"context"
	"testing"
)

func TestCollectSMARTData_InvalidPath(t *testing.T) {
	_, err := CollectSMARTData(context.Background(), "/tmp/not-a-device")
	if err == nil {
		t.Error("expected error for invalid device path")
	}
}

func TestCollectSMARTData_EmptyPath(t *testing.T) {
	// Empty path should default to /dev/sda
	// Will fail because smartctl likely not installed or /dev/sda doesn't exist,
	// but should not panic
	_, err := CollectSMARTData(context.Background(), "")
	if err == nil {
		t.Skip("smartctl available and /dev/sda exists — skipping error test")
	}
	// Just verify we got an error and didn't panic
	t.Logf("expected error: %v", err)
}
