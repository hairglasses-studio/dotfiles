package errbudget

import (
	"context"
	"errors"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestNewTracker(t *testing.T) {
	tr := NewTracker(5)
	if tr.threshold != 5 {
		t.Errorf("expected threshold 5, got %d", tr.threshold)
	}

	// Default threshold
	tr = NewTracker(0)
	if tr.threshold != DefaultThreshold {
		t.Errorf("expected default threshold %d, got %d", DefaultThreshold, tr.threshold)
	}

	tr = NewTracker(-1)
	if tr.threshold != DefaultThreshold {
		t.Errorf("expected default threshold %d, got %d", DefaultThreshold, tr.threshold)
	}
}

func TestRecordSuccessResetsCounter(t *testing.T) {
	tr := NewTracker(3)
	tr.RecordError("tool_a", "err1")
	tr.RecordError("tool_a", "err2")

	if tr.IsDegraded("tool_a") {
		t.Fatal("tool should not be degraded after 2 errors (threshold 3)")
	}

	tr.RecordSuccess("tool_a")
	status := tr.Status("tool_a")
	if status.ConsecutiveErrors != 0 {
		t.Errorf("expected 0 consecutive errors after success, got %d", status.ConsecutiveErrors)
	}
	if status.TotalInvocations != 3 {
		t.Errorf("expected 3 total invocations, got %d", status.TotalInvocations)
	}
}

func TestDegradedAfterThreshold(t *testing.T) {
	tr := NewTracker(2)

	tr.RecordError("tool_b", "fail1")
	if tr.IsDegraded("tool_b") {
		t.Fatal("should not be degraded after 1 error (threshold 2)")
	}

	tr.RecordError("tool_b", "fail2")
	if !tr.IsDegraded("tool_b") {
		t.Fatal("should be degraded after 2 errors (threshold 2)")
	}

	status := tr.Status("tool_b")
	if !status.Degraded {
		t.Error("status should report degraded")
	}
	if status.LastError != "fail2" {
		t.Errorf("expected last error 'fail2', got %q", status.LastError)
	}
	if status.TotalErrors != 2 {
		t.Errorf("expected 2 total errors, got %d", status.TotalErrors)
	}
}

func TestResetClearsState(t *testing.T) {
	tr := NewTracker(1)
	tr.RecordError("tool_c", "oops")
	if !tr.IsDegraded("tool_c") {
		t.Fatal("should be degraded")
	}

	tr.Reset("tool_c")
	if tr.IsDegraded("tool_c") {
		t.Fatal("should not be degraded after reset")
	}
}

func TestResetAll(t *testing.T) {
	tr := NewTracker(1)
	tr.RecordError("tool_x", "err")
	tr.RecordError("tool_y", "err")

	tr.ResetAll()
	if tr.IsDegraded("tool_x") || tr.IsDegraded("tool_y") {
		t.Fatal("no tools should be degraded after ResetAll")
	}
}

func TestDegradedTools(t *testing.T) {
	tr := NewTracker(1)
	tr.RecordError("degraded1", "err")
	tr.RecordError("degraded2", "err")
	tr.RecordSuccess("healthy1")

	degraded := tr.DegradedTools()
	if len(degraded) != 2 {
		t.Fatalf("expected 2 degraded tools, got %d", len(degraded))
	}
}

func TestAllStatus(t *testing.T) {
	tr := NewTracker(3)
	tr.RecordSuccess("tool_a")
	tr.RecordError("tool_b", "fail")

	statuses := tr.AllStatus()
	if len(statuses) != 2 {
		t.Fatalf("expected 2 statuses, got %d", len(statuses))
	}
}

func TestStatusForUnknownTool(t *testing.T) {
	tr := NewTracker(3)
	status := tr.Status("nonexistent")
	if status.ConsecutiveErrors != 0 {
		t.Error("unknown tool should have 0 consecutive errors")
	}
	if status.Degraded {
		t.Error("unknown tool should not be degraded")
	}
	if status.Threshold != 3 {
		t.Errorf("expected threshold 3, got %d", status.Threshold)
	}
}

func TestIsDegradedForUnknownTool(t *testing.T) {
	tr := NewTracker(3)
	if tr.IsDegraded("unknown") {
		t.Error("unknown tool should not be degraded")
	}
}

func TestWrapSuccessPath(t *testing.T) {
	tr := NewTracker(3)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: "ok"},
			},
		}, nil
	}

	wrapped := tr.Wrap("my_tool", handler)
	result, err := wrapped(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("result should not be an error")
	}

	status := tr.Status("my_tool")
	if status.ConsecutiveErrors != 0 {
		t.Errorf("expected 0 consecutive errors, got %d", status.ConsecutiveErrors)
	}
}

func TestWrapErrorPath(t *testing.T) {
	tr := NewTracker(2)

	callCount := 0
	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		callCount++
		return nil, errors.New("connection refused")
	}

	wrapped := tr.Wrap("failing_tool", handler)

	// First call: error, but not yet degraded
	_, _ = wrapped(context.Background(), mcp.CallToolRequest{})
	if tr.IsDegraded("failing_tool") {
		t.Fatal("should not be degraded after 1 error")
	}

	// Second call: error, now degraded
	_, _ = wrapped(context.Background(), mcp.CallToolRequest{})
	if !tr.IsDegraded("failing_tool") {
		t.Fatal("should be degraded after 2 errors")
	}

	// Third call: should be blocked (handler not called)
	result, err := wrapped(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("degraded response should not return error: %v", err)
	}
	if !result.IsError {
		t.Error("degraded response should have IsError=true")
	}
	if callCount != 2 {
		t.Errorf("handler should have been called 2 times, got %d", callCount)
	}

	// Check degraded count
	status := tr.Status("failing_tool")
	if status.DegradedSinceCount != 1 {
		t.Errorf("expected 1 degraded rejection, got %d", status.DegradedSinceCount)
	}
}

func TestWrapErrorResultPath(t *testing.T) {
	tr := NewTracker(1)

	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: "something broke"},
			},
			IsError: true,
		}, nil
	}

	wrapped := tr.Wrap("err_result_tool", handler)
	_, _ = wrapped(context.Background(), mcp.CallToolRequest{})

	if !tr.IsDegraded("err_result_tool") {
		t.Fatal("should be degraded after error result")
	}
	status := tr.Status("err_result_tool")
	if status.LastError != "something broke" {
		t.Errorf("expected last error 'something broke', got %q", status.LastError)
	}
}

func TestWrapRecoveryAfterDegraded(t *testing.T) {
	tr := NewTracker(1)
	tr.RecordError("recover_tool", "initial failure")

	// Tool is degraded
	if !tr.IsDegraded("recover_tool") {
		t.Fatal("should be degraded")
	}

	// Reset it
	tr.Reset("recover_tool")

	// Now it should work
	handler := func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{Type: "text", Text: "recovered"},
			},
		}, nil
	}

	wrapped := tr.Wrap("recover_tool", handler)
	result, err := wrapped(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.IsError {
		t.Error("should not be error after reset")
	}
}
