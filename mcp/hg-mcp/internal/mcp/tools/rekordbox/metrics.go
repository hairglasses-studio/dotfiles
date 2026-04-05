package rekordbox

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	rekordboxAPITotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rekordbox_api_total",
			Help: "Total number of Rekordbox API calls",
		},
		[]string{"endpoint", "status"},
	)

	rekordboxAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "rekordbox_api_duration_seconds",
			Help:    "Duration of Rekordbox API calls in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"endpoint"},
	)

	rekordboxErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rekordbox_errors_total",
			Help: "Total number of Rekordbox errors",
		},
		[]string{"operation", "error_type"},
	)
)

// RecordRekordboxAPI records a Rekordbox API call metric.
func RecordRekordboxAPI(endpoint string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	rekordboxAPITotal.WithLabelValues(endpoint, status).Inc()
	rekordboxAPIDuration.WithLabelValues(endpoint).Observe(duration)
}

// RecordRekordboxError records a Rekordbox error.
func RecordRekordboxError(operation, errorType string) {
	rekordboxErrorsTotal.WithLabelValues(operation, errorType).Inc()
}
