package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Health check timeout
const HealthCheckTimeout = 10 * time.Second

// ServiceHealth represents the health status of a service
type ServiceHealth struct {
	Name      string `json:"name"`
	Status    string `json:"status"` // "healthy", "unhealthy", "degraded"
	Message   string `json:"message,omitempty"`
	Latency   string `json:"latency,omitempty"`
	CheckedAt string `json:"checked_at"`
}

// HealthCheckResult contains all service health statuses
type HealthCheckResult struct {
	Overall         string            `json:"overall"` // "healthy", "unhealthy", "degraded"
	Services        []ServiceHealth   `json:"services"`
	CircuitBreakers map[string]string `json:"circuit_breakers,omitempty"`
}

// Prometheus metrics for health checks
var (
	HealthCheckStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sync_health_check_status",
			Help: "Health check status (1=healthy, 0=unhealthy)",
		},
		[]string{"service"},
	)

	HealthCheckLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sync_health_check_latency_seconds",
			Help:    "Health check latency in seconds",
			Buckets: []float64{0.1, 0.5, 1, 2, 5, 10},
		},
		[]string{"service"},
	)
)

// HealthChecker performs health checks on sync dependencies
type HealthChecker struct {
	config *Config
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(config *Config) *HealthChecker {
	return &HealthChecker{config: config}
}

// Check performs health checks on all services
func (h *HealthChecker) Check(ctx context.Context) HealthCheckResult {
	ctx, cancel := context.WithTimeout(ctx, HealthCheckTimeout)
	defer cancel()

	services := []ServiceHealth{
		h.checkAWSCLI(ctx),
		h.checkS3Access(ctx),
		h.checkDynamoDB(ctx),
		h.checkFFmpeg(ctx),
	}

	// Determine overall status
	overall := "healthy"
	for _, svc := range services {
		if svc.Status == "unhealthy" {
			overall = "unhealthy"
			break
		} else if svc.Status == "degraded" && overall == "healthy" {
			overall = "degraded"
		}
	}

	return HealthCheckResult{
		Overall:         overall,
		Services:        services,
		CircuitBreakers: GlobalCircuitBreakers.Status(),
	}
}

// checkAWSCLI verifies AWS CLI is installed and configured
func (h *HealthChecker) checkAWSCLI(ctx context.Context) ServiceHealth {
	start := time.Now()
	result := ServiceHealth{
		Name:      "aws_cli",
		CheckedAt: time.Now().Format(time.RFC3339),
	}

	cmd := exec.CommandContext(ctx, "aws", "--version")
	output, err := cmd.Output()
	latency := time.Since(start)
	result.Latency = latency.Round(time.Millisecond).String()

	HealthCheckLatency.WithLabelValues("aws_cli").Observe(latency.Seconds())

	if err != nil {
		result.Status = "unhealthy"
		result.Message = fmt.Sprintf("AWS CLI not available: %v", err)
		HealthCheckStatus.WithLabelValues("aws_cli").Set(0)
	} else {
		result.Status = "healthy"
		result.Message = string(output[:min(len(output), 50)])
		HealthCheckStatus.WithLabelValues("aws_cli").Set(1)
	}

	return result
}

// checkS3Access verifies S3 bucket access
func (h *HealthChecker) checkS3Access(ctx context.Context) ServiceHealth {
	start := time.Now()
	result := ServiceHealth{
		Name:      "s3_bucket",
		CheckedAt: time.Now().Format(time.RFC3339),
	}

	args := []string{"s3", "ls", fmt.Sprintf("s3://%s/", h.config.S3Bucket), "--profile", h.config.AWSProfile}
	cmd := exec.CommandContext(ctx, "aws", args...)
	_, err := cmd.Output()
	latency := time.Since(start)
	result.Latency = latency.Round(time.Millisecond).String()

	HealthCheckLatency.WithLabelValues("s3_bucket").Observe(latency.Seconds())

	if err != nil {
		result.Status = "unhealthy"
		result.Message = fmt.Sprintf("S3 bucket inaccessible: %v", err)
		HealthCheckStatus.WithLabelValues("s3_bucket").Set(0)
	} else {
		result.Status = "healthy"
		result.Message = fmt.Sprintf("Bucket %s accessible", h.config.S3Bucket)
		HealthCheckStatus.WithLabelValues("s3_bucket").Set(1)
	}

	return result
}

// checkDynamoDB verifies DynamoDB table access
func (h *HealthChecker) checkDynamoDB(ctx context.Context) ServiceHealth {
	start := time.Now()
	result := ServiceHealth{
		Name:      "dynamodb",
		CheckedAt: time.Now().Format(time.RFC3339),
	}

	// Check if cr8_tracks table exists
	args := []string{
		"dynamodb", "describe-table",
		"--table-name", "cr8_tracks",
		"--profile", h.config.AWSProfile,
		"--output", "json",
	}
	cmd := exec.CommandContext(ctx, "aws", args...)
	output, err := cmd.Output()
	latency := time.Since(start)
	result.Latency = latency.Round(time.Millisecond).String()

	HealthCheckLatency.WithLabelValues("dynamodb").Observe(latency.Seconds())

	if err != nil {
		result.Status = "degraded"
		result.Message = "DynamoDB table not accessible (optional)"
		HealthCheckStatus.WithLabelValues("dynamodb").Set(0)
	} else {
		// Parse table status
		var tableInfo struct {
			Table struct {
				TableStatus string `json:"TableStatus"`
				ItemCount   int64  `json:"ItemCount"`
			} `json:"Table"`
		}
		if json.Unmarshal(output, &tableInfo) == nil {
			result.Status = "healthy"
			result.Message = fmt.Sprintf("Table ACTIVE, %d items", tableInfo.Table.ItemCount)
			HealthCheckStatus.WithLabelValues("dynamodb").Set(1)
		} else {
			result.Status = "healthy"
			result.Message = "Table accessible"
			HealthCheckStatus.WithLabelValues("dynamodb").Set(1)
		}
	}

	return result
}

// checkFFmpeg verifies ffmpeg is installed
func (h *HealthChecker) checkFFmpeg(ctx context.Context) ServiceHealth {
	start := time.Now()
	result := ServiceHealth{
		Name:      "ffmpeg",
		CheckedAt: time.Now().Format(time.RFC3339),
	}

	cmd := exec.CommandContext(ctx, "ffmpeg", "-version")
	output, err := cmd.Output()
	latency := time.Since(start)
	result.Latency = latency.Round(time.Millisecond).String()

	HealthCheckLatency.WithLabelValues("ffmpeg").Observe(latency.Seconds())

	if err != nil {
		result.Status = "degraded"
		result.Message = "ffmpeg not available (optional for conversion)"
		HealthCheckStatus.WithLabelValues("ffmpeg").Set(0)
	} else {
		result.Status = "healthy"
		// Extract version from first line
		lines := string(output)
		if idx := len(lines); idx > 60 {
			lines = lines[:60]
		}
		result.Message = lines
		HealthCheckStatus.WithLabelValues("ffmpeg").Set(1)
	}

	return result
}

// min helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// CheckRekordbox checks if Rekordbox database is accessible
func (h *HealthChecker) CheckRekordbox(ctx context.Context) ServiceHealth {
	start := time.Now()
	result := ServiceHealth{
		Name:      "rekordbox",
		CheckedAt: time.Now().Format(time.RFC3339),
	}

	// Check if Rekordbox database exists
	dbPath := fmt.Sprintf("%s/Library/Pioneer/rekordbox/master.db", os.Getenv("HOME"))
	if _, err := os.Stat(dbPath); err != nil {
		result.Status = "degraded"
		result.Message = "Rekordbox database not found"
		HealthCheckStatus.WithLabelValues("rekordbox").Set(0)
	} else {
		result.Status = "healthy"
		result.Message = "Rekordbox database accessible"
		HealthCheckStatus.WithLabelValues("rekordbox").Set(1)
	}

	result.Latency = time.Since(start).Round(time.Millisecond).String()
	return result
}
