package wled

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	wledAPITotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wled_api_total",
			Help: "Total number of WLED API calls",
		},
		[]string{"endpoint", "status"},
	)

	wledAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "wled_api_duration_seconds",
			Help:    "Duration of WLED API calls in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"endpoint"},
	)

	wledErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "wled_errors_total",
			Help: "Total number of WLED errors",
		},
		[]string{"operation", "error_type"},
	)
)

// RecordWLEDAPI records a WLED API call metric.
func RecordWLEDAPI(endpoint string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	wledAPITotal.WithLabelValues(endpoint, status).Inc()
	wledAPIDuration.WithLabelValues(endpoint).Observe(duration)
}

// RecordWLEDError records a WLED error.
func RecordWLEDError(operation, errorType string) {
	wledErrorsTotal.WithLabelValues(operation, errorType).Inc()
}
