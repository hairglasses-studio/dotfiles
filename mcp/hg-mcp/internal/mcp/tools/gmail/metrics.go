package gmail

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Google service metrics — shared across Gmail, Calendar, Sheets, Drive modules.
// Prefixed by specific service name to distinguish API usage.
var (
	// GmailAPITotal tracks Gmail API calls
	GmailAPITotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "google_gmail_api_total",
			Help: "Total number of Gmail API calls",
		},
		[]string{"operation", "status"},
	)

	// GmailAPIDuration tracks Gmail API call duration
	GmailAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "google_gmail_api_duration_seconds",
			Help:    "Duration of Gmail API calls in seconds",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2, 5, 10},
		},
		[]string{"operation"},
	)

	// GmailMessagesTotal tracks message operations
	GmailMessagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "google_gmail_messages_total",
			Help: "Total number of Gmail message operations",
		},
		[]string{"operation", "status"},
	)

	// GmailErrorsTotal tracks Gmail errors
	GmailErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "google_gmail_errors_total",
			Help: "Total number of Gmail errors",
		},
		[]string{"operation", "error_type"},
	)
)

// RecordGmailAPI records a Gmail API call metric
func RecordGmailAPI(operation string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	GmailAPITotal.WithLabelValues(operation, status).Inc()
	GmailAPIDuration.WithLabelValues(operation).Observe(duration)
}

// RecordGmailMessage records a message operation
func RecordGmailMessage(operation string, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	GmailMessagesTotal.WithLabelValues(operation, status).Inc()
}

// RecordGmailError records a Gmail error
func RecordGmailError(operation, errorType string) {
	GmailErrorsTotal.WithLabelValues(operation, errorType).Inc()
}
