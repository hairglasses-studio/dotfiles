package midi

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	midiAPITotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "midi_api_total",
			Help: "Total number of MIDI API calls",
		},
		[]string{"endpoint", "status"},
	)

	midiAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "midi_api_duration_seconds",
			Help:    "Duration of MIDI API calls in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"endpoint"},
	)

	midiErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "midi_errors_total",
			Help: "Total number of MIDI errors",
		},
		[]string{"operation", "error_type"},
	)
)

// RecordMIDIAPI records a MIDI API call metric.
func RecordMIDIAPI(endpoint string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	midiAPITotal.WithLabelValues(endpoint, status).Inc()
	midiAPIDuration.WithLabelValues(endpoint).Observe(duration)
}

// RecordMIDIError records a MIDI error.
func RecordMIDIError(operation, errorType string) {
	midiErrorsTotal.WithLabelValues(operation, errorType).Inc()
}
