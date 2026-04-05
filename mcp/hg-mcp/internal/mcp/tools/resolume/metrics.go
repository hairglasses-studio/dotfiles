package resolume

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// ResolumeAPITotal tracks Resolume OSC/HTTP API calls
	ResolumeAPITotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "resolume_api_total",
			Help: "Total number of Resolume API calls",
		},
		[]string{"endpoint", "status"},
	)

	// ResolumeAPIDuration tracks Resolume API call duration
	ResolumeAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "resolume_api_duration_seconds",
			Help:    "Duration of Resolume API calls in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"endpoint"},
	)

	// ResolumeClipTriggersTotal tracks clip trigger events
	ResolumeClipTriggersTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "resolume_clip_triggers_total",
			Help: "Total number of clip trigger operations",
		},
		[]string{"type", "status"},
	)

	// ResolumeBPMChangesTotal tracks BPM changes
	ResolumeBPMChangesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "resolume_bpm_changes_total",
			Help: "Total number of BPM change operations",
		},
	)

	// ResolumeEffectChangesTotal tracks effect toggles and parameter changes
	ResolumeEffectChangesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "resolume_effect_changes_total",
			Help: "Total number of effect change operations",
		},
		[]string{"operation"},
	)

	// ResolumeErrorsTotal tracks Resolume errors
	ResolumeErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "resolume_errors_total",
			Help: "Total number of Resolume errors",
		},
		[]string{"operation", "error_type"},
	)

	// ResolumeCacheHits tracks cache hit/miss for cached methods
	ResolumeCacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "resolume_cache_hits_total",
			Help: "Total number of Resolume cache hits and misses",
		},
		[]string{"method", "result"},
	)

	// ResolumeLayerOpsTotal tracks layer operations (opacity, bypass, solo, clear)
	ResolumeLayerOpsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "resolume_layer_ops_total",
			Help: "Total number of layer operations",
		},
		[]string{"operation", "status"},
	)
)

// RecordResolumeAPI records a Resolume API call metric
func RecordResolumeAPI(endpoint string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	ResolumeAPITotal.WithLabelValues(endpoint, status).Inc()
	ResolumeAPIDuration.WithLabelValues(endpoint).Observe(duration)
}

// RecordResolumeError records a Resolume error
func RecordResolumeError(operation, errorType string) {
	ResolumeErrorsTotal.WithLabelValues(operation, errorType).Inc()
}
