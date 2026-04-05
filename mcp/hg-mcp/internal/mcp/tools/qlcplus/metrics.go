package qlcplus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	qlcplusAPITotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "qlcplus_api_total",
			Help: "Total number of QLCPlus API calls",
		},
		[]string{"endpoint", "status"},
	)

	qlcplusAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "qlcplus_api_duration_seconds",
			Help:    "Duration of QLCPlus API calls in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1},
		},
		[]string{"endpoint"},
	)

	qlcplusErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "qlcplus_errors_total",
			Help: "Total number of QLCPlus errors",
		},
		[]string{"operation", "error_type"},
	)
)

// RecordQLCPlusAPI records a QLCPlus API call metric.
func RecordQLCPlusAPI(endpoint string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	qlcplusAPITotal.WithLabelValues(endpoint, status).Inc()
	qlcplusAPIDuration.WithLabelValues(endpoint).Observe(duration)
}

// RecordQLCPlusError records a QLCPlus error.
func RecordQLCPlusError(operation, errorType string) {
	qlcplusErrorsTotal.WithLabelValues(operation, errorType).Inc()
}
