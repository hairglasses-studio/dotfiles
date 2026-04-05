package soundcloud

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	soundcloudAPITotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "soundcloud_api_total",
			Help: "Total number of SoundCloud API calls",
		},
		[]string{"endpoint", "status"},
	)

	soundcloudAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "soundcloud_api_duration_seconds",
			Help:    "Duration of SoundCloud API calls in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"endpoint"},
	)

	soundcloudErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "soundcloud_errors_total",
			Help: "Total number of SoundCloud errors",
		},
		[]string{"operation", "error_type"},
	)
)

// RecordSoundCloudAPI records a SoundCloud API call metric.
func RecordSoundCloudAPI(endpoint string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	soundcloudAPITotal.WithLabelValues(endpoint, status).Inc()
	soundcloudAPIDuration.WithLabelValues(endpoint).Observe(duration)
}

// RecordSoundCloudError records a SoundCloud error.
func RecordSoundCloudError(operation, errorType string) {
	soundcloudErrorsTotal.WithLabelValues(operation, errorType).Inc()
}
