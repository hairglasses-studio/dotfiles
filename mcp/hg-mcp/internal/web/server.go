// Package web provides the HTTP server for the hg-mcp web UI.
package web

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Server is the web UI HTTP server.
type Server struct {
	mux          *http.ServeMux
	registry     *tools.ToolRegistry
	dashboard    *clients.DashboardClient
	sseClients   map[string]chan SSEEvent
	sseClientsMu sync.RWMutex
	staticFS     fs.FS
	devMode      bool
}

// Config holds server configuration.
type Config struct {
	Port      string
	DevMode   bool
	StaticDir string
}

// NewServer creates a new web server.
func NewServer(cfg Config) *Server {
	s := &Server{
		mux:        http.NewServeMux(),
		registry:   tools.GetRegistry(),
		dashboard:  clients.GetDashboardClient(),
		sseClients: make(map[string]chan SSEEvent),
		devMode:    cfg.DevMode,
	}

	// Set up static file serving
	if cfg.DevMode {
		// In dev mode, proxy to Vite dev server
		log.Println("Web UI running in dev mode - proxy to Vite at http://localhost:5173")
	} else {
		// In production, serve embedded static files
		staticFS, err := GetStaticFS()
		if err != nil {
			log.Printf("Warning: could not load embedded static files: %v", err)
		} else {
			s.staticFS = staticFS
		}
	}

	s.setupRoutes()
	return s
}

// setupRoutes configures all HTTP routes.
func (s *Server) setupRoutes() {
	// API routes
	s.mux.HandleFunc("/api/v1/tools", s.handleTools)
	s.mux.HandleFunc("/api/v1/tools/", s.handleTool)
	s.mux.HandleFunc("/api/v1/categories", s.handleCategories)
	s.mux.HandleFunc("/api/v1/stats", s.handleStats)
	s.mux.HandleFunc("/api/v1/search", s.handleSearch)

	// Dashboard API routes
	s.mux.HandleFunc("/api/v1/dashboard/status", s.handleDashboardStatus)
	s.mux.HandleFunc("/api/v1/dashboard/quick", s.handleDashboardQuick)
	s.mux.HandleFunc("/api/v1/dashboard/monitor/start", s.handleMonitorStart)
	s.mux.HandleFunc("/api/v1/dashboard/monitor/stop", s.handleMonitorStop)
	s.mux.HandleFunc("/api/v1/dashboard/alerts", s.handleAlerts)

	// Workflow routes
	s.mux.HandleFunc("/api/v1/workflows", s.handleWorkflows)
	s.mux.HandleFunc("/api/v1/workflows/", s.handleWorkflow)

	// Preferences routes
	s.mux.HandleFunc("/api/v1/favorites", s.handleFavorites)
	s.mux.HandleFunc("/api/v1/aliases", s.handleAliases)

	// Inventory API routes
	s.mux.HandleFunc("/api/v1/inventory", s.handleInventoryItems)
	s.mux.HandleFunc("/api/v1/inventory/", s.handleInventoryItem)

	// SSE for real-time updates
	s.mux.HandleFunc("/api/v1/sse", s.handleSSE)

	// Health check — structured JSON with registry stats
	s.mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		registry := tools.GetRegistry()
		unconfigured := registry.GetUnconfiguredCategories()
		unconfiguredCount := 0
		for _, count := range unconfigured {
			unconfiguredCount += count
		}

		health := map[string]interface{}{
			"status":               "ok",
			"tools":                registry.ToolCount(),
			"modules":              registry.ModuleCount(),
			"runtime_groups":       registry.GetRuntimeGroupStats(),
			"unconfigured_tools":   unconfiguredCount,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(health)
	})

	// Static files (SPA)
	s.mux.HandleFunc("/", s.handleStatic)
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Add CORS headers for development
	if s.devMode {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	s.mux.ServeHTTP(w, r)
}

// Start starts the HTTP server.
func (s *Server) Start(port string) error {
	log.Printf("Starting web UI server on :%s", port)
	return http.ListenAndServe(":"+port, s)
}

// handleStatic serves static files for the SPA.
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	if s.staticFS == nil {
		http.Error(w, "Static files not available", http.StatusNotFound)
		return
	}

	path := r.URL.Path
	if path == "/" {
		path = "index.html"
	} else {
		path = strings.TrimPrefix(path, "/")
	}

	// Try to read the file
	content, err := fs.ReadFile(s.staticFS, path)
	if err != nil {
		// For SPA, serve index.html for any non-existent path
		content, err = fs.ReadFile(s.staticFS, "index.html")
		if err != nil {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		path = "index.html"
	}

	// Set content type based on extension
	ext := filepath.Ext(path)
	switch ext {
	case ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	case ".json":
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".ico":
		w.Header().Set("Content-Type", "image/x-icon")
	case ".woff", ".woff2":
		w.Header().Set("Content-Type", "font/woff2")
	}

	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

// writeJSON writes a JSON response.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// readJSON reads a JSON request body.
func readJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// apiError represents an API error response.
type apiError struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// writeError writes an error response.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, apiError{Error: message})
}

// SSEEvent represents a server-sent event.
type SSEEvent struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

// handleSSE handles Server-Sent Events for real-time updates.
func (s *Server) handleSSE(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create client channel
	clientID := fmt.Sprintf("%d", time.Now().UnixNano())
	events := make(chan SSEEvent, 10)

	s.sseClientsMu.Lock()
	s.sseClients[clientID] = events
	s.sseClientsMu.Unlock()

	defer func() {
		s.sseClientsMu.Lock()
		delete(s.sseClients, clientID)
		s.sseClientsMu.Unlock()
		close(events)
	}()

	// Send initial status
	ctx := r.Context()
	status := s.dashboard.GetFullStatus(ctx)
	fmt.Fprintf(w, "event: status\ndata: %s\n\n", mustJSON(status))
	flusher.Flush()

	// Keep-alive ticker
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Status update ticker
	statusTicker := time.NewTicker(30 * time.Second)
	defer statusTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		case <-statusTicker.C:
			status := s.dashboard.GetFullStatus(ctx)
			fmt.Fprintf(w, "event: status\ndata: %s\n\n", mustJSON(status))
			flusher.Flush()
		case event := <-events:
			fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Event, mustJSON(event.Data))
			flusher.Flush()
		}
	}
}

// BroadcastEvent sends an event to all SSE clients.
func (s *Server) BroadcastEvent(event SSEEvent) {
	s.sseClientsMu.RLock()
	defer s.sseClientsMu.RUnlock()

	for _, ch := range s.sseClients {
		select {
		case ch <- event:
		default:
			// Skip if client buffer is full
		}
	}
}

func mustJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}
