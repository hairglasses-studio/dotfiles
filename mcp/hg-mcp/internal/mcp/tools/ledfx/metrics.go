package ledfx

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ledfxAPITotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ledfx_api_total",
			Help: "Total number of LedFX API calls",
		},
		[]string{"endpoint", "status"},
	)

	ledfxAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ledfx_api_duration_seconds",
			Help:    "Duration of LedFX API calls in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"endpoint"},
	)

	ledfxErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ledfx_errors_total",
			Help: "Total number of LedFX errors",
		},
		[]string{"operation", "error_type"},
	)
)

// RecordLedFXAPI records a LedFX API call metric.
func RecordLedFXAPI(endpoint string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	ledfxAPITotal.WithLabelValues(endpoint, status).Inc()
	ledfxAPIDuration.WithLabelValues(endpoint).Observe(duration)
}

// RecordLedFXError records a LedFX error.
func RecordLedFXError(operation, errorType string) {
	ledfxErrorsTotal.WithLabelValues(operation, errorType).Inc()
}
