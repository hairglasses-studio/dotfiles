package obs

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// OBSAPITotal tracks OBS WebSocket API calls
	OBSAPITotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "obs_api_total",
			Help: "Total number of OBS API calls",
		},
		[]string{"endpoint", "status"},
	)

	// OBSAPIDuration tracks OBS API call duration
	OBSAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "obs_api_duration_seconds",
			Help:    "Duration of OBS API calls in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"endpoint"},
	)

	// OBSStreamingActive tracks whether OBS is currently streaming
	OBSStreamingActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "obs_streaming_active",
			Help: "Whether OBS is currently streaming (1=yes, 0=no)",
		},
	)

	// OBSRecordingActive tracks whether OBS is currently recording
	OBSRecordingActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "obs_recording_active",
			Help: "Whether OBS is currently recording (1=yes, 0=no)",
		},
	)

	// OBSSceneSwitchesTotal tracks scene switch events
	OBSSceneSwitchesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "obs_scene_switches_total",
			Help: "Total number of OBS scene switches",
		},
	)

	// OBSErrorsTotal tracks OBS errors
	OBSErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "obs_errors_total",
			Help: "Total number of OBS errors",
		},
		[]string{"operation", "error_type"},
	)
)

// RecordOBSAPI records an OBS API call metric
func RecordOBSAPI(endpoint string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	OBSAPITotal.WithLabelValues(endpoint, status).Inc()
	OBSAPIDuration.WithLabelValues(endpoint).Observe(duration)
}

// RecordOBSError records an OBS error metric
func RecordOBSError(operation, errorType string) {
	OBSErrorsTotal.WithLabelValues(operation, errorType).Inc()
}
