package ableton

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// AbletonAPITotal tracks Ableton OSC API calls
	AbletonAPITotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ableton_api_total",
			Help: "Total number of Ableton OSC API calls",
		},
		[]string{"endpoint", "status"},
	)

	// AbletonAPIDuration tracks Ableton API call duration
	AbletonAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "ableton_api_duration_seconds",
			Help:    "Duration of Ableton API calls in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"endpoint"},
	)

	// AbletonPlayingActive tracks whether Ableton is currently playing
	AbletonPlayingActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "ableton_playing_active",
			Help: "Whether Ableton is currently playing (1=yes, 0=no)",
		},
	)

	// AbletonBPM tracks current BPM
	AbletonBPM = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "ableton_bpm",
			Help: "Current Ableton BPM",
		},
	)

	// AbletonTransportOpsTotal tracks transport operations
	AbletonTransportOpsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ableton_transport_ops_total",
			Help: "Total number of Ableton transport operations",
		},
		[]string{"action", "status"},
	)

	// AbletonErrorsTotal tracks Ableton errors
	AbletonErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "ableton_errors_total",
			Help: "Total number of Ableton errors",
		},
		[]string{"operation", "error_type"},
	)
)

// RecordAbletonAPI records an Ableton API call metric
func RecordAbletonAPI(endpoint string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	AbletonAPITotal.WithLabelValues(endpoint, status).Inc()
	AbletonAPIDuration.WithLabelValues(endpoint).Observe(duration)
}

// RecordAbletonError records an Ableton error metric
func RecordAbletonError(operation, errorType string) {
	AbletonErrorsTotal.WithLabelValues(operation, errorType).Inc()
}
