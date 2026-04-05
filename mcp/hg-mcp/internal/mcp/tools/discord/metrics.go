package discord

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// DiscordMessagesTotal tracks message operations (send, edit, delete)
	DiscordMessagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "discord_messages_total",
			Help: "Total number of Discord message operations",
		},
		[]string{"operation", "status"},
	)

	// DiscordAPITotal tracks Discord API calls by endpoint
	DiscordAPITotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "discord_api_total",
			Help: "Total number of Discord API calls",
		},
		[]string{"endpoint", "status"},
	)

	// DiscordAPIDuration tracks Discord API call duration
	DiscordAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "discord_api_duration_seconds",
			Help:    "Duration of Discord API calls in seconds",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5},
		},
		[]string{"endpoint"},
	)

	// DiscordWebhooksTotal tracks webhook invocations
	DiscordWebhooksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "discord_webhooks_total",
			Help: "Total number of Discord webhook invocations",
		},
		[]string{"status"},
	)

	// DiscordErrorsTotal tracks Discord errors by type
	DiscordErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "discord_errors_total",
			Help: "Total number of Discord errors",
		},
		[]string{"operation", "error_type"},
	)

	// DiscordCacheHits tracks cache hit/miss for cached methods
	DiscordCacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "discord_cache_hits_total",
			Help: "Total number of Discord cache hits and misses",
		},
		[]string{"method", "result"},
	)
)

// RecordDiscordAPI records a Discord API call metric
func RecordDiscordAPI(endpoint string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	DiscordAPITotal.WithLabelValues(endpoint, status).Inc()
	DiscordAPIDuration.WithLabelValues(endpoint).Observe(duration)
}

// RecordDiscordMessage records a message operation
func RecordDiscordMessage(operation string, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	DiscordMessagesTotal.WithLabelValues(operation, status).Inc()
}

// RecordDiscordError records a Discord error
func RecordDiscordError(operation, errorType string) {
	DiscordErrorsTotal.WithLabelValues(operation, errorType).Inc()
}
