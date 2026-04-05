package calendar

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	calendarAPITotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "calendar_api_total",
			Help: "Total number of Calendar API calls",
		},
		[]string{"endpoint", "status"},
	)

	calendarAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "calendar_api_duration_seconds",
			Help:    "Duration of Calendar API calls in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"endpoint"},
	)

	calendarErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "calendar_errors_total",
			Help: "Total number of Calendar errors",
		},
		[]string{"operation", "error_type"},
	)
)

// RecordCalendarAPI records a Calendar API call metric.
func RecordCalendarAPI(endpoint string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	calendarAPITotal.WithLabelValues(endpoint, status).Inc()
	calendarAPIDuration.WithLabelValues(endpoint).Observe(duration)
}

// RecordCalendarError records a Calendar error.
func RecordCalendarError(operation, errorType string) {
	calendarErrorsTotal.WithLabelValues(operation, errorType).Inc()
}
