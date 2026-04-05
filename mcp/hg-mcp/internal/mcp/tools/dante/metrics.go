package dante

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	danteAPITotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dante_api_total",
			Help: "Total number of Dante API calls",
		},
		[]string{"endpoint", "status"},
	)

	danteAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "dante_api_duration_seconds",
			Help:    "Duration of Dante API calls in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"endpoint"},
	)

	danteErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "dante_errors_total",
			Help: "Total number of Dante errors",
		},
		[]string{"operation", "error_type"},
	)
)

// RecordDanteAPI records a Dante API call metric.
func RecordDanteAPI(endpoint string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	danteAPITotal.WithLabelValues(endpoint, status).Inc()
	danteAPIDuration.WithLabelValues(endpoint).Observe(duration)
}

// RecordDanteError records a Dante error.
func RecordDanteError(operation, errorType string) {
	danteErrorsTotal.WithLabelValues(operation, errorType).Inc()
}
