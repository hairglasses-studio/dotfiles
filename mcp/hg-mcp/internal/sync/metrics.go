package sync

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// SyncFilesTotal tracks total files synced per service/user/playlist
	SyncFilesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sync_files_total",
			Help: "Total number of files synced",
		},
		[]string{"service", "user", "playlist"},
	)

	// SyncFilesFailed tracks failed file syncs
	SyncFilesFailed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sync_files_failed",
			Help: "Total number of failed file syncs",
		},
		[]string{"service", "user"},
	)

	// SyncDuration tracks sync operation duration
	SyncDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "sync_duration_seconds",
			Help:    "Duration of sync operations in seconds",
			Buckets: []float64{1, 5, 10, 30, 60, 120, 300, 600},
		},
		[]string{"service", "user"},
	)

	// SyncLastSuccess tracks the timestamp of last successful sync
	SyncLastSuccess = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sync_last_success_timestamp",
			Help: "Unix timestamp of last successful sync",
		},
		[]string{"service", "user", "playlist"},
	)

	// SyncFilesGauge tracks current file count per playlist
	SyncFilesGauge = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sync_files_current",
			Help: "Current number of files in playlist",
		},
		[]string{"service", "user", "playlist"},
	)

	// SyncPendingFiles tracks files pending Rekordbox import
	SyncPendingFiles = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "sync_pending_files",
			Help: "Number of files pending Rekordbox import",
		},
	)

	// SyncActiveOps tracks currently running sync operations
	SyncActiveOps = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sync_active_operations",
			Help: "Number of currently active sync operations",
		},
		[]string{"service"},
	)

	// SyncErrorsTotal tracks sync errors by type
	SyncErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sync_errors_total",
			Help: "Total number of sync errors by type",
		},
		[]string{"service", "error_type"},
	)

	// SyncBytesTotal tracks total bytes synced
	SyncBytesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sync_bytes_total",
			Help: "Total bytes synced",
		},
		[]string{"service", "user"},
	)
)

// RecordSyncResult records metrics from a SyncResult
func RecordSyncResult(result SyncResult) {
	duration := result.EndTime.Sub(result.StartTime).Seconds()

	// Record duration
	SyncDuration.WithLabelValues(result.Service, result.User).Observe(duration)

	// Record files synced
	if result.Synced > 0 {
		SyncFilesTotal.WithLabelValues(result.Service, result.User, result.Playlist).Add(float64(result.Synced))
	}

	// Record failures
	if result.Failed > 0 {
		SyncFilesFailed.WithLabelValues(result.Service, result.User).Add(float64(result.Failed))
	}

	// Record errors
	for range result.Errors {
		SyncErrorsTotal.WithLabelValues(result.Service, "general").Inc()
	}

	// Update last success timestamp
	if result.Failed == 0 && len(result.Errors) == 0 {
		SyncLastSuccess.WithLabelValues(result.Service, result.User, result.Playlist).SetToCurrentTime()
	}

	// Update current file count
	if result.Total > 0 {
		SyncFilesGauge.WithLabelValues(result.Service, result.User, result.Playlist).Set(float64(result.Total))
	}
}

// StartSyncOp marks the start of a sync operation
func StartSyncOp(service string) {
	SyncActiveOps.WithLabelValues(service).Inc()
}

// EndSyncOp marks the end of a sync operation
func EndSyncOp(service string) {
	SyncActiveOps.WithLabelValues(service).Dec()
}
