package companion

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	companionAPITotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "companion_api_total",
			Help: "Total number of Companion API calls",
		},
		[]string{"endpoint", "status"},
	)

	companionAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "companion_api_duration_seconds",
			Help:    "Duration of Companion API calls in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"endpoint"},
	)

	companionErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "companion_errors_total",
			Help: "Total number of Companion errors",
		},
		[]string{"operation", "error_type"},
	)
)

// RecordCompanionAPI records a Companion API call metric.
func RecordCompanionAPI(endpoint string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	companionAPITotal.WithLabelValues(endpoint, status).Inc()
	companionAPIDuration.WithLabelValues(endpoint).Observe(duration)
}

// RecordCompanionError records a Companion error.
func RecordCompanionError(operation, errorType string) {
	companionErrorsTotal.WithLabelValues(operation, errorType).Inc()
}
