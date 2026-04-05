package web

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/web/api"
)

// handleTools handles GET /api/v1/tools
func (s *Server) handleTools(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	allTools := s.registry.GetAllToolDefinitions()
	var toolResponses []api.ToolResponse
	for _, td := range allTools {
		toolResponses = append(toolResponses, api.ConvertToolToResponse(td))
	}

	writeJSON(w, http.StatusOK, toolResponses)
}

// handleTool handles /api/v1/tools/{name} and /api/v1/tools/{name}/execute
func (s *Server) handleTool(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/tools/")
	parts := strings.Split(path, "/")
	toolName := parts[0]

	if toolName == "" {
		writeError(w, http.StatusBadRequest, "tool name required")
		return
	}

	// Check for /execute suffix
	if len(parts) > 1 && parts[1] == "execute" {
		s.handleToolExecute(w, r, toolName)
		return
	}

	// Check for /schema suffix
	if len(parts) > 1 && parts[1] == "schema" {
		s.handleToolSchema(w, r, toolName)
		return
	}

	// GET single tool
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	td, ok := s.registry.GetTool(toolName)
	if !ok {
		writeError(w, http.StatusNotFound, "tool not found: "+toolName)
		return
	}

	writeJSON(w, http.StatusOK, api.ConvertToolToResponse(td))
}

// handleToolExecute handles POST /api/v1/tools/{name}/execute
func (s *Server) handleToolExecute(w http.ResponseWriter, r *http.Request, toolName string) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req api.ExecuteRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result := api.ExecuteTool(ctx, s.registry, toolName, req.Arguments)

	status := http.StatusOK
	if !result.Success {
		status = http.StatusInternalServerError
	}

	writeJSON(w, status, result)

	// Broadcast execution event
	s.BroadcastEvent(SSEEvent{
		Event: "execution",
		Data: map[string]interface{}{
			"tool":    toolName,
			"success": result.Success,
		},
	})
}

// handleToolSchema handles GET /api/v1/tools/{name}/schema
func (s *Server) handleToolSchema(w http.ResponseWriter, r *http.Request, toolName string) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	td, ok := s.registry.GetTool(toolName)
	if !ok {
		writeError(w, http.StatusNotFound, "tool not found: "+toolName)
		return
	}

	writeJSON(w, http.StatusOK, td.Tool.InputSchema)
}

// handleCategories handles GET /api/v1/categories
func (s *Server) handleCategories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	categories := api.GetCategories(s.registry)
	writeJSON(w, http.StatusOK, categories)
}

// handleStats handles GET /api/v1/stats
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	stats := s.registry.GetToolStats()
	writeJSON(w, http.StatusOK, stats)
}

// handleSearch handles GET /api/v1/search?q=query
func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, "query parameter 'q' required")
		return
	}

	results := api.SearchTools(s.registry, query)
	writeJSON(w, http.StatusOK, results)
}

// handleDashboardStatus handles GET /api/v1/dashboard/status
func (s *Server) handleDashboardStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	status := s.dashboard.GetFullStatus(ctx)
	writeJSON(w, http.StatusOK, status)
}

// handleDashboardQuick handles GET /api/v1/dashboard/quick
func (s *Server) handleDashboardQuick(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	quick := s.dashboard.GetQuickStatus(ctx)
	writeJSON(w, http.StatusOK, map[string]string{"status": quick})
}

// handleMonitorStart handles POST /api/v1/dashboard/monitor/start
func (s *Server) handleMonitorStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Interval int `json:"interval"` // seconds
	}
	if err := readJSON(r, &req); err != nil {
		req.Interval = 30 // default
	}

	interval := time.Duration(req.Interval) * time.Second
	if interval < 5*time.Second {
		interval = 5 * time.Second
	}

	s.dashboard.StartMonitoring(interval)
	writeJSON(w, http.StatusOK, map[string]bool{"monitoring": true})
}

// handleMonitorStop handles POST /api/v1/dashboard/monitor/stop
func (s *Server) handleMonitorStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	s.dashboard.StopMonitoring()
	writeJSON(w, http.StatusOK, map[string]bool{"monitoring": false})
}

// handleAlerts handles GET /api/v1/dashboard/alerts
func (s *Server) handleAlerts(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	includeResolved := r.URL.Query().Get("resolved") == "true"
	alerts := s.dashboard.GetAlerts(includeResolved)
	writeJSON(w, http.StatusOK, alerts)
}

// handleWorkflows handles /api/v1/workflows
func (s *Server) handleWorkflows(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		workflows := api.GetWorkflows()
		writeJSON(w, http.StatusOK, workflows)

	case http.MethodPost:
		var req api.WorkflowResponse
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
			return
		}
		if req.Name == "" {
			writeError(w, http.StatusBadRequest, "workflow name is required")
			return
		}
		if err := api.CreateWorkflow(r.Context(), req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusCreated, map[string]string{"status": "created", "name": req.Name})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleWorkflow handles /api/v1/workflows/{name}
func (s *Server) handleWorkflow(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v1/workflows/")
	parts := strings.Split(path, "/")
	workflowName := parts[0]

	if workflowName == "" {
		writeError(w, http.StatusBadRequest, "workflow name required")
		return
	}

	// Check for /run suffix
	if len(parts) > 1 && parts[1] == "run" {
		if r.Method != http.MethodPost {
			writeError(w, http.StatusMethodNotAllowed, "method not allowed")
			return
		}

		var params map[string]interface{}
		if err := readJSON(r, &params); err != nil {
			params = make(map[string]interface{})
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Minute)
		defer cancel()

		result := api.RunWorkflow(ctx, s.registry, workflowName, params)

		status := http.StatusOK
		if !result.Success {
			status = http.StatusInternalServerError
		}

		writeJSON(w, status, result)
		return
	}

	// Handle CRUD on individual workflow
	switch r.Method {
	case http.MethodGet:
		workflow, err := api.GetWorkflow(workflowName)
		if err != nil {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, workflow)

	case http.MethodPut:
		var req api.WorkflowResponse
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
			return
		}
		if err := api.UpdateWorkflow(r.Context(), workflowName, req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "updated", "name": req.Name})

	case http.MethodDelete:
		if err := api.DeleteWorkflow(r.Context(), workflowName); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted", "name": workflowName})

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleFavorites handles /api/v1/favorites
func (s *Server) handleFavorites(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		prefs := api.GetPreferences()
		writeJSON(w, http.StatusOK, prefs.Favorites)

	case http.MethodPost:
		var req struct {
			ToolName string `json:"toolName"`
		}
		if err := readJSON(r, &req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		prefs := api.GetPreferences()
		// Add to favorites if not already present
		for _, f := range prefs.Favorites {
			if f == req.ToolName {
				writeJSON(w, http.StatusOK, prefs.Favorites)
				return
			}
		}
		prefs.Favorites = append(prefs.Favorites, req.ToolName)
		api.SetFavorites(prefs.Favorites)
		writeJSON(w, http.StatusOK, prefs.Favorites)

	case http.MethodPut:
		var favorites []string
		if err := json.NewDecoder(r.Body).Decode(&favorites); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		api.SetFavorites(favorites)
		writeJSON(w, http.StatusOK, favorites)

	case http.MethodDelete:
		toolName := r.URL.Query().Get("toolName")
		if toolName == "" {
			writeError(w, http.StatusBadRequest, "toolName query parameter required")
			return
		}
		prefs := api.GetPreferences()
		var newFavorites []string
		for _, f := range prefs.Favorites {
			if f != toolName {
				newFavorites = append(newFavorites, f)
			}
		}
		api.SetFavorites(newFavorites)
		writeJSON(w, http.StatusOK, newFavorites)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

// handleAliases handles /api/v1/aliases
func (s *Server) handleAliases(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		prefs := api.GetPreferences()
		writeJSON(w, http.StatusOK, prefs.Aliases)

	case http.MethodPut:
		var aliases map[string]string
		if err := json.NewDecoder(r.Body).Decode(&aliases); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		api.SetAliases(aliases)
		writeJSON(w, http.StatusOK, aliases)

	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}
