package notion

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	notionAPITotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notion_api_total",
			Help: "Total number of Notion API calls",
		},
		[]string{"endpoint", "status"},
	)

	notionAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "notion_api_duration_seconds",
			Help:    "Duration of Notion API calls in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"endpoint"},
	)

	notionErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notion_errors_total",
			Help: "Total number of Notion errors",
		},
		[]string{"operation", "error_type"},
	)
)

// RecordNotionAPI records a Notion API call metric.
func RecordNotionAPI(endpoint string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	notionAPITotal.WithLabelValues(endpoint, status).Inc()
	notionAPIDuration.WithLabelValues(endpoint).Observe(duration)
}

// RecordNotionError records a Notion error.
func RecordNotionError(operation, errorType string) {
	notionErrorsTotal.WithLabelValues(operation, errorType).Inc()
}
