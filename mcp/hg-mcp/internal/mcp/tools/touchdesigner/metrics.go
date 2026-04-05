package touchdesigner

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// TDAPITotal tracks TouchDesigner API calls
	TDAPITotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "touchdesigner_api_total",
			Help: "Total number of TouchDesigner API calls",
		},
		[]string{"endpoint", "status"},
	)

	// TDAPIDuration tracks TouchDesigner API call duration
	TDAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "touchdesigner_api_duration_seconds",
			Help:    "Duration of TouchDesigner API calls in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2},
		},
		[]string{"endpoint"},
	)

	// TDFPS tracks the current TouchDesigner FPS
	TDFPS = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "touchdesigner_fps",
			Help: "Current TouchDesigner frames per second",
		},
	)

	// TDCookTime tracks the current cook time
	TDCookTime = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "touchdesigner_cook_time_ms",
			Help: "Current TouchDesigner cook time in milliseconds",
		},
	)

	// TDGPUMemory tracks GPU memory usage
	TDGPUMemory = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "touchdesigner_gpu_memory_bytes",
			Help: "TouchDesigner GPU memory usage in bytes",
		},
	)

	// TDErrorsTotal tracks TouchDesigner errors
	TDErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "touchdesigner_errors_total",
			Help: "Total number of TouchDesigner errors",
		},
		[]string{"operation", "error_type"},
	)
)

// RecordTDAPI records a TouchDesigner API call metric
func RecordTDAPI(endpoint string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	TDAPITotal.WithLabelValues(endpoint, status).Inc()
	TDAPIDuration.WithLabelValues(endpoint).Observe(duration)
}

// RecordTDError records a TouchDesigner error metric
func RecordTDError(operation, errorType string) {
	TDErrorsTotal.WithLabelValues(operation, errorType).Inc()
}
