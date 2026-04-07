// Package clients provides API clients for external services.
package clients

import (
	"context"
	"strings"
)

// RouterClient routes natural language queries to appropriate tools
type RouterClient struct{}

// RouteResult represents the result of routing a query
type RouteResult struct {
	Tool         string            `json:"tool"`
	Confidence   float64           `json:"confidence"`
	Parameters   map[string]string `json:"parameters,omitempty"`
	Explanation  string            `json:"explanation"`
	Alternatives []string          `json:"alternatives,omitempty"`
}

// ToolRoute defines a tool and its matching patterns
type ToolRoute struct {
	Tool        string
	Patterns    []string
	Keywords    []string
	Description string
}

// NewRouterClient creates a new router client
func NewRouterClient() (*RouterClient, error) {
	return &RouterClient{}, nil
}

// Route routes a natural language query to the best matching tool
func (c *RouterClient) Route(ctx context.Context, query string) (*RouteResult, error) {
	query = strings.ToLower(query)

	routes := c.getRoutes()
	var bestMatch *RouteResult
	highestScore := 0.0

	for _, route := range routes {
		score := c.calculateMatchScore(query, route)
		if score > highestScore {
			highestScore = score
			bestMatch = &RouteResult{
				Tool:        route.Tool,
				Confidence:  score,
				Explanation: route.Description,
				Parameters:  c.extractParameters(query, route),
			}
		}
	}

	if bestMatch == nil || bestMatch.Confidence < 0.2 {
		return &RouteResult{
			Tool:        "aftrs_tool_search",
			Confidence:  0.5,
			Explanation: "No direct match found. Searching for relevant tools.",
			Parameters:  map[string]string{"query": query},
		}, nil
	}

	// Find alternatives
	alternatives := []string{}
	for _, route := range routes {
		score := c.calculateMatchScore(query, route)
		if score > 0.3 && route.Tool != bestMatch.Tool {
			alternatives = append(alternatives, route.Tool)
			if len(alternatives) >= 3 {
				break
			}
		}
	}
	bestMatch.Alternatives = alternatives

	return bestMatch, nil
}

// calculateMatchScore calculates how well a query matches a route
func (c *RouterClient) calculateMatchScore(query string, route ToolRoute) float64 {
	score := 0.0

	// Pattern matching (highest weight)
	for _, pattern := range route.Patterns {
		if strings.Contains(query, pattern) {
			score += 0.4
			break
		}
	}

	// Keyword matching
	keywordsMatched := 0
	for _, keyword := range route.Keywords {
		if strings.Contains(query, keyword) {
			keywordsMatched++
		}
	}
	if len(route.Keywords) > 0 {
		score += float64(keywordsMatched) / float64(len(route.Keywords)) * 0.6
	}

	return score
}

// extractParameters extracts parameters from the query for the matched tool
func (c *RouterClient) extractParameters(query string, route ToolRoute) map[string]string {
	params := make(map[string]string)

	// Extract common parameter patterns
	switch route.Tool {
	case "aftrs_discord_send":
		// Try to extract message content
		if idx := strings.Index(query, "say "); idx != -1 {
			params["message"] = query[idx+4:]
		} else if idx := strings.Index(query, "send "); idx != -1 {
			params["message"] = query[idx+5:]
		}

	case "aftrs_graph_search", "aftrs_pattern_match":
		// Use the whole query as search
		params["query"] = query

	case "aftrs_troubleshoot":
		params["issue"] = query

	case "aftrs_equipment_history":
		// Try to extract equipment name
		for _, word := range []string{"touchdesigner", "resolume", "dmx", "ndi", "lighting"} {
			if strings.Contains(query, word) {
				params["equipment"] = word
				break
			}
		}
	}

	return params
}

// getRoutes returns all available routes
func (c *RouterClient) getRoutes() []ToolRoute {
	return []ToolRoute{
		// TouchDesigner
		{
			Tool:        "aftrs_td_status",
			Patterns:    []string{"touchdesigner fps", "td fps", "touchdesigner status", "td status"},
			Keywords:    []string{"touchdesigner", "td", "fps", "cook", "frame"},
			Description: "Get TouchDesigner status including FPS, cook time, and errors",
		},
		{
			Tool:        "aftrs_td_performance",
			Patterns:    []string{"touchdesigner performance", "td performance", "td slow"},
			Keywords:    []string{"touchdesigner", "td", "performance", "slow", "gpu", "memory"},
			Description: "Get detailed TouchDesigner performance metrics",
		},

		// Resolume
		{
			Tool:        "aftrs_resolume_status",
			Patterns:    []string{"resolume status", "resolume bpm"},
			Keywords:    []string{"resolume", "vj", "bpm", "layer"},
			Description: "Get Resolume Arena status",
		},

		// NDI/Streaming
		{
			Tool:        "aftrs_ndi_sources",
			Patterns:    []string{"ndi sources", "ndi available", "list ndi"},
			Keywords:    []string{"ndi", "sources", "stream", "video"},
			Description: "List available NDI sources",
		},
		{
			Tool:        "aftrs_stream_dashboard",
			Patterns:    []string{"streaming status", "stream dashboard"},
			Keywords:    []string{"stream", "streaming", "broadcast", "capture"},
			Description: "Get streaming dashboard with NDI and capture status",
		},

		// Lighting
		{
			Tool:        "aftrs_lighting_status",
			Patterns:    []string{"lighting status", "dmx status", "lights"},
			Keywords:    []string{"lighting", "dmx", "artnet", "fixtures"},
			Description: "Get lighting and DMX status",
		},

		// Studio Health
		{
			Tool:        "hairglasses_studio_health_full",
			Patterns:    []string{"studio health", "health check", "system status", "everything ok"},
			Keywords:    []string{"health", "status", "check", "systems", "studio"},
			Description: "Get comprehensive studio health: TD + Resolume + DMX + NDI + UNRAID",
		},
		{
			Tool:        "aftrs_show_preflight",
			Patterns:    []string{"preflight", "ready for show", "pre-show check"},
			Keywords:    []string{"preflight", "ready", "show", "check"},
			Description: "Run pre-show checklist to verify all systems",
		},

		// Shows
		{
			Tool:        "aftrs_show_startup",
			Patterns:    []string{"start show", "startup", "begin show", "start the show"},
			Keywords:    []string{"start", "show", "startup", "begin"},
			Description: "Execute automated show startup sequence",
		},
		{
			Tool:        "aftrs_show_shutdown",
			Patterns:    []string{"end show", "shutdown", "stop show"},
			Keywords:    []string{"shutdown", "stop", "end", "close"},
			Description: "Execute graceful show shutdown sequence",
		},
		{
			Tool:        "aftrs_panic_mode",
			Patterns:    []string{"panic", "emergency stop", "kill everything"},
			Keywords:    []string{"panic", "emergency", "stop", "kill"},
			Description: "EMERGENCY: Immediately stop all outputs",
		},

		// UNRAID
		{
			Tool:        "aftrs_unraid_status",
			Patterns:    []string{"unraid status", "nas status", "server status"},
			Keywords:    []string{"unraid", "nas", "server", "array", "disk"},
			Description: "Get UNRAID array health and status",
		},

		// Learning/Troubleshooting
		{
			Tool:        "aftrs_troubleshoot",
			Patterns:    []string{"troubleshoot", "diagnose", "what's wrong with", "fix"},
			Keywords:    []string{"troubleshoot", "problem", "issue", "fix", "diagnose", "wrong"},
			Description: "Get guided troubleshooting based on learned patterns",
		},
		{
			Tool:        "aftrs_fix_suggest",
			Patterns:    []string{"how to fix", "suggest fix", "solution for"},
			Keywords:    []string{"fix", "suggest", "solution", "repair"},
			Description: "Get fix suggestions based on symptoms",
		},
		{
			Tool:        "aftrs_pattern_match",
			Patterns:    []string{"seen this before", "similar issue", "match pattern"},
			Keywords:    []string{"pattern", "similar", "before", "history"},
			Description: "Match symptoms to learned patterns",
		},

		// Knowledge
		{
			Tool:        "aftrs_graph_search",
			Patterns:    []string{"search vault", "find in vault", "search documents"},
			Keywords:    []string{"search", "find", "vault", "document", "knowledge"},
			Description: "Graph-enhanced search across vault documents",
		},
		{
			Tool:        "aftrs_similar_shows",
			Patterns:    []string{"similar shows", "past shows like", "shows like this"},
			Keywords:    []string{"similar", "shows", "past", "like"},
			Description: "Find similar past shows",
		},
		{
			Tool:        "aftrs_resolution_path",
			Patterns:    []string{"how was this fixed", "past resolution"},
			Keywords:    []string{"resolution", "fixed", "solved", "before"},
			Description: "Find how similar issues were resolved",
		},

		// Equipment/Venue
		{
			Tool:        "aftrs_equipment_history",
			Patterns:    []string{"equipment history", "reliability of", "issues with"},
			Keywords:    []string{"equipment", "history", "reliability", "issues"},
			Description: "Get equipment issue history and reliability score",
		},
		{
			Tool:        "aftrs_venue_patterns",
			Patterns:    []string{"venue patterns", "venue quirks", "venue issues"},
			Keywords:    []string{"venue", "location", "patterns", "quirks"},
			Description: "Get venue-specific patterns and recommendations",
		},

		// Backup
		{
			Tool:        "aftrs_backup_all",
			Patterns:    []string{"backup everything", "backup all", "full backup"},
			Keywords:    []string{"backup", "save", "archive"},
			Description: "Backup all project files to NAS",
		},

		// Discord
		{
			Tool:        "aftrs_discord_send",
			Patterns:    []string{"send discord", "message discord", "notify team"},
			Keywords:    []string{"discord", "send", "message", "notify", "team"},
			Description: "Send a message to Discord",
		},

		// Retro Gaming
		{
			Tool:        "aftrs_ps2_status",
			Patterns:    []string{"ps2 status", "pcsx2 status", "emulator status"},
			Keywords:    []string{"ps2", "pcsx2", "emulator", "playstation"},
			Description: "Get PS2 emulator status",
		},

		// Test
		{
			Tool:        "aftrs_test_sequence",
			Patterns:    []string{"test everything", "test all systems", "run tests"},
			Keywords:    []string{"test", "verify", "check"},
			Description: "Test all systems in sequence",
		},
	}
}
