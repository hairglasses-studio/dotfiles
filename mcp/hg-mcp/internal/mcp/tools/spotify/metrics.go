package spotify

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// SpotifyAPITotal tracks Spotify API calls by endpoint
	SpotifyAPITotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "spotify_api_total",
			Help: "Total number of Spotify API calls",
		},
		[]string{"endpoint", "status"},
	)

	// SpotifyAPIDuration tracks Spotify API call duration
	SpotifyAPIDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "spotify_api_duration_seconds",
			Help:    "Duration of Spotify API calls in seconds",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5},
		},
		[]string{"endpoint"},
	)

	// SpotifySearchTotal tracks search operations
	SpotifySearchTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "spotify_search_total",
			Help: "Total number of Spotify search operations",
		},
		[]string{"type", "status"},
	)

	// SpotifyTracksAnalyzed tracks audio feature lookups
	SpotifyTracksAnalyzed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "spotify_tracks_analyzed_total",
			Help: "Total number of tracks with audio features retrieved",
		},
	)

	// SpotifyErrorsTotal tracks Spotify errors by type
	SpotifyErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "spotify_errors_total",
			Help: "Total number of Spotify errors",
		},
		[]string{"operation", "error_type"},
	)

	// SpotifyCacheHits tracks cache hit/miss for cached methods
	SpotifyCacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "spotify_cache_hits_total",
			Help: "Total number of Spotify cache hits and misses",
		},
		[]string{"method", "result"},
	)

	// SpotifyAuthTotal tracks authentication attempts
	SpotifyAuthTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "spotify_auth_total",
			Help: "Total number of Spotify auth attempts",
		},
		[]string{"status"},
	)
)

// RecordSpotifyAPI records a Spotify API call metric
func RecordSpotifyAPI(endpoint string, duration float64, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	SpotifyAPITotal.WithLabelValues(endpoint, status).Inc()
	SpotifyAPIDuration.WithLabelValues(endpoint).Observe(duration)
}

// RecordSpotifySearch records a search operation
func RecordSpotifySearch(searchType string, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	SpotifySearchTotal.WithLabelValues(searchType, status).Inc()
}

// RecordSpotifyError records a Spotify error
func RecordSpotifyError(operation, errorType string) {
	SpotifyErrorsTotal.WithLabelValues(operation, errorType).Inc()
}
