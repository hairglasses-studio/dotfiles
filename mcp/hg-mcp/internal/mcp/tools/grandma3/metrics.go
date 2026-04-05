package grandma3

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// GMA3APITotal tracks grandMA3 API calls
	GMA3APITotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gma3_api_total",
			Help: "Total number of grandMA3 API calls",
		},
		[]string{"endpoint", "status"},
	)

	// GMA3APIDuration tracks grandMA3 API call duration
	GMA3APIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gma3_api_duration_seconds",
			Help:    "Duration of grandMA3 API calls in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"endpoint"},
	)

	// GMA3CommandsTotal tracks grandMA3 command executions
	GMA3CommandsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gma3_commands_total",
			Help: "Total number of grandMA3 commands executed",
		},
		[]string{"command_type", "status"},
	)

	// GMA3BlackoutActive tracks blackout state
	GMA3BlackoutActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "gma3_blackout_active",
			Help: "Whether grandMA3 blackout is active (1=yes, 0=no)",
		},
	)

	// GMA3ErrorsTotal tracks grandMA3 errors
	GMA3ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gma3_errors_total",
			Help: "Total number of grandMA3 errors",
		},
		[]string{"operation", "error_type"},
	)
)

// RecordGMA3API records a grandMA3 API call metric
func RecordGMA3API(endpoint string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	GMA3APITotal.WithLabelValues(endpoint, status).Inc()
	GMA3APIDuration.WithLabelValues(endpoint).Observe(duration)
}

// RecordGMA3Error records a grandMA3 error metric
func RecordGMA3Error(operation, errorType string) {
	GMA3ErrorsTotal.WithLabelValues(operation, errorType).Inc()
}
