package sacn

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	sacnAPITotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sacn_api_total",
			Help: "Total number of SACN API calls",
		},
		[]string{"endpoint", "status"},
	)

	sacnAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sacn_api_duration_seconds",
			Help:    "Duration of SACN API calls in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"endpoint"},
	)

	sacnErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sacn_errors_total",
			Help: "Total number of SACN errors",
		},
		[]string{"operation", "error_type"},
	)
)

// RecordSACNAPI records a SACN API call metric.
func RecordSACNAPI(endpoint string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	sacnAPITotal.WithLabelValues(endpoint, status).Inc()
	sacnAPIDuration.WithLabelValues(endpoint).Observe(duration)
}

// RecordSACNError records a SACN error.
func RecordSACNError(operation, errorType string) {
	sacnErrorsTotal.WithLabelValues(operation, errorType).Inc()
}
