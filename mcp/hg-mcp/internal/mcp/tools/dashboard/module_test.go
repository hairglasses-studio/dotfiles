package dashboard

import (
	"context"
	"testing"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
	"github.com/mark3labs/mcp-go/mcp"
)

func TestModuleInfo(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}

func TestRingBuffer(t *testing.T) {
	rb := NewTrendRingBuffer(5)

	// Empty buffer
	if pts := rb.Points(); pts != nil {
		t.Errorf("empty buffer should return nil, got %d points", len(pts))
	}
	min, max, avg, count := rb.Stats()
	if count != 0 || min != 0 || max != 0 || avg != 0 {
		t.Error("empty buffer stats should all be zero")
	}

	// Add 3 points
	rb.Add(HealthDataPoint{Score: 80, Status: "healthy"})
	rb.Add(HealthDataPoint{Score: 60, Status: "degraded"})
	rb.Add(HealthDataPoint{Score: 100, Status: "healthy"})

	min, max, avg, count = rb.Stats()
	if count != 3 {
		t.Errorf("count = %d, want 3", count)
	}
	if min != 60 {
		t.Errorf("min = %d, want 60", min)
	}
	if max != 100 {
		t.Errorf("max = %d, want 100", max)
	}
	if avg != 80.0 {
		t.Errorf("avg = %f, want 80.0", avg)
	}
}

func TestRingBufferOverflow(t *testing.T) {
	rb := NewTrendRingBuffer(3)

	// Fill beyond capacity
	rb.Add(HealthDataPoint{Score: 10})
	rb.Add(HealthDataPoint{Score: 20})
	rb.Add(HealthDataPoint{Score: 30})
	rb.Add(HealthDataPoint{Score: 40}) // Overwrites first
	rb.Add(HealthDataPoint{Score: 50}) // Overwrites second

	pts := rb.Points()
	if len(pts) != 3 {
		t.Fatalf("expected 3 points, got %d", len(pts))
	}

	// Should have 30, 40, 50 (oldest first)
	if pts[0].Score != 30 || pts[1].Score != 40 || pts[2].Score != 50 {
		t.Errorf("expected [30,40,50], got [%d,%d,%d]", pts[0].Score, pts[1].Score, pts[2].Score)
	}
}

func TestHandleDashboardTrends(t *testing.T) {
	// Seed some data
	RecordHealthCheck(90, "healthy")
	RecordHealthCheck(75, "degraded")
	RecordHealthCheck(95, "healthy")

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{}

	result, err := handleDashboardTrends(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if result.IsError {
		t.Error("expected success")
	}
	content := result.Content[0].(mcp.TextContent)
	t.Logf("Trends: %.300s", content.Text)
}

func TestHandleDashboardTrendsInvalidWindow(t *testing.T) {
	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]interface{}{
		"window": "invalid",
	}

	result, err := handleDashboardTrends(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsError {
		t.Error("expected error for invalid window")
	}
}
