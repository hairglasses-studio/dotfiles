package chataigne

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	chataigneAPITotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chataigne_api_total",
			Help: "Total number of Chataigne API calls",
		},
		[]string{"endpoint", "status"},
	)

	chataigneAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "chataigne_api_duration_seconds",
			Help:    "Duration of Chataigne API calls in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"endpoint"},
	)

	chataigneErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chataigne_errors_total",
			Help: "Total number of Chataigne errors",
		},
		[]string{"operation", "error_type"},
	)
)

// RecordChataigneAPI records a Chataigne API call metric.
func RecordChataigneAPI(endpoint string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	chataigneAPITotal.WithLabelValues(endpoint, status).Inc()
	chataigneAPIDuration.WithLabelValues(endpoint).Observe(duration)
}

// RecordChataigneError records a Chataigne error.
func RecordChataigneError(operation, errorType string) {
	chataigneErrorsTotal.WithLabelValues(operation, errorType).Inc()
}
