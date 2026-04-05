package dashboard

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// HealthDataPoint represents a single health check measurement.
type HealthDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Score     int       `json:"score"`
	Status    string    `json:"status"`
	LatencyMs int64     `json:"latency_ms,omitempty"`
}

// TrendRingBuffer is a fixed-size ring buffer for health data points.
type TrendRingBuffer struct {
	mu     sync.RWMutex
	data   []HealthDataPoint
	size   int
	cursor int
	count  int
}

// NewTrendRingBuffer creates a ring buffer with the given capacity.
func NewTrendRingBuffer(size int) *TrendRingBuffer {
	return &TrendRingBuffer{
		data: make([]HealthDataPoint, size),
		size: size,
	}
}

// Add appends a data point to the ring buffer.
func (rb *TrendRingBuffer) Add(point HealthDataPoint) {
	rb.mu.Lock()
	defer rb.mu.Unlock()
	rb.data[rb.cursor] = point
	rb.cursor = (rb.cursor + 1) % rb.size
	if rb.count < rb.size {
		rb.count++
	}
}

// Points returns all data points in chronological order.
func (rb *TrendRingBuffer) Points() []HealthDataPoint {
	rb.mu.RLock()
	defer rb.mu.RUnlock()

	if rb.count == 0 {
		return nil
	}

	result := make([]HealthDataPoint, rb.count)
	if rb.count < rb.size {
		copy(result, rb.data[:rb.count])
	} else {
		start := rb.cursor
		copy(result, rb.data[start:])
		copy(result[rb.size-start:], rb.data[:start])
	}
	return result
}

// Stats returns min, max, and average scores from the buffer.
func (rb *TrendRingBuffer) Stats() (min, max int, avg float64, count int) {
	points := rb.Points()
	if len(points) == 0 {
		return 0, 0, 0, 0
	}

	min = points[0].Score
	max = points[0].Score
	sum := 0

	for _, p := range points {
		if p.Score < min {
			min = p.Score
		}
		if p.Score > max {
			max = p.Score
		}
		sum += p.Score
	}

	return min, max, float64(sum) / float64(len(points)), len(points)
}

// LatencyPercentiles represents response time percentile statistics.
type LatencyPercentiles struct {
	P50   int64 `json:"p50_ms"`
	P90   int64 `json:"p90_ms"`
	P95   int64 `json:"p95_ms"`
	P99   int64 `json:"p99_ms"`
	Min   int64 `json:"min_ms"`
	Max   int64 `json:"max_ms"`
	Count int   `json:"sample_count"`
}

// Percentiles calculates response time percentiles from stored latency data.
func (rb *TrendRingBuffer) Percentiles() LatencyPercentiles {
	points := rb.Points()
	if len(points) == 0 {
		return LatencyPercentiles{}
	}

	// Collect non-zero latencies
	latencies := make([]int64, 0, len(points))
	for _, p := range points {
		if p.LatencyMs > 0 {
			latencies = append(latencies, p.LatencyMs)
		}
	}
	if len(latencies) == 0 {
		return LatencyPercentiles{Count: len(points)}
	}

	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

	percentile := func(p float64) int64 {
		idx := int(float64(len(latencies)-1) * p)
		return latencies[idx]
	}

	return LatencyPercentiles{
		P50:   percentile(0.50),
		P90:   percentile(0.90),
		P95:   percentile(0.95),
		P99:   percentile(0.99),
		Min:   latencies[0],
		Max:   latencies[len(latencies)-1],
		Count: len(latencies),
	}
}

// Global ring buffer for health trend data (100 most recent checks).
var healthTrends = NewTrendRingBuffer(100)

// RecordHealthCheck adds a health check result to the trend buffer.
// Called by the dashboard watcher or manually.
func RecordHealthCheck(score int, status string, latencyMs ...int64) {
	point := HealthDataPoint{
		Timestamp: time.Now(),
		Score:     score,
		Status:    status,
	}
	if len(latencyMs) > 0 {
		point.LatencyMs = latencyMs[0]
	}
	healthTrends.Add(point)
}

// handleDashboardTrends handles the aftrs_dashboard_trends tool.
func handleDashboardTrends(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	window := tools.GetStringParam(req, "window")

	points := healthTrends.Points()
	minScore, maxScore, avgScore, count := healthTrends.Stats()

	// Filter by time window if specified
	if window != "" && len(points) > 0 {
		var dur time.Duration
		switch window {
		case "1h":
			dur = time.Hour
		case "6h":
			dur = 6 * time.Hour
		case "24h":
			dur = 24 * time.Hour
		default:
			return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid window: %s (use 1h, 6h, or 24h)", window)), nil
		}

		cutoff := time.Now().Add(-dur)
		filtered := make([]HealthDataPoint, 0, len(points))
		for _, p := range points {
			if p.Timestamp.After(cutoff) {
				filtered = append(filtered, p)
			}
		}
		points = filtered

		// Recalculate stats for filtered data
		if len(points) > 0 {
			minScore = points[0].Score
			maxScore = points[0].Score
			sum := 0
			for _, p := range points {
				if p.Score < minScore {
					minScore = p.Score
				}
				if p.Score > maxScore {
					maxScore = p.Score
				}
				sum += p.Score
			}
			avgScore = float64(sum) / float64(len(points))
			count = len(points)
		} else {
			minScore, maxScore, avgScore, count = 0, 0, 0, 0
		}
	}

	pctiles := healthTrends.Percentiles()

	result := map[string]interface{}{
		"data_points":         count,
		"min_score":           minScore,
		"max_score":           maxScore,
		"avg_score":           fmt.Sprintf("%.1f", avgScore),
		"latency_percentiles": pctiles,
		"history":             points,
	}

	if window != "" {
		result["window"] = window
	}

	return tools.JSONResult(result), nil
}
