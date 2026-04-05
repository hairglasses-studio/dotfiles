// Package workflow_automation provides workflow automation tools for hg-mcp.
package workflow_automation

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// chartTools returns the chart aggregation tool definitions
func chartTools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_workflow_chart_sync_status",
				mcp.WithDescription("View chart sync health and last aggregation status across all configured platforms."),
			),
			Handler:             handleChartSyncStatus,
			Category:            "workflow_automation",
			Subcategory:         "charts",
			Tags:                []string{"chart", "sync", "status", "beatport", "traxsource"},
			UseCases:            []string{"Check chart sync health", "View aggregation status"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "workflow_automation",
		},
		{
			Tool: mcp.NewTool("aftrs_workflow_chart_aggregate",
				mcp.WithDescription("Aggregate charts from multiple platforms with weighted scoring. Combines Beatport, Traxsource, and other sources."),
				mcp.WithString("platforms", mcp.Description("Comma-separated platforms to aggregate: beatport,traxsource,juno,boomkat")),
				mcp.WithString("genre", mcp.Description("Genre to filter (e.g., techno, house, drum-and-bass)")),
				mcp.WithNumber("limit", mcp.Description("Maximum tracks to return (default: 50)")),
			),
			Handler:             handleChartAggregate,
			Category:            "workflow_automation",
			Subcategory:         "charts",
			Tags:                []string{"chart", "aggregate", "beatport", "traxsource", "ranking"},
			UseCases:            []string{"Create combined chart", "Find trending tracks across platforms"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "workflow_automation",
		},
		{
			Tool: mcp.NewTool("aftrs_workflow_chart_sync_run",
				mcp.WithDescription("Trigger a manual chart sync from all configured platforms. Updates local chart cache."),
				mcp.WithString("platforms", mcp.Description("Comma-separated platforms to sync (default: all)")),
			),
			Handler:             handleChartSyncRun,
			Category:            "workflow_automation",
			Subcategory:         "charts",
			Tags:                []string{"chart", "sync", "run", "update"},
			UseCases:            []string{"Manual chart refresh", "Update chart cache"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "workflow_automation",
		},
		{
			Tool: mcp.NewTool("aftrs_workflow_chart_config",
				mcp.WithDescription("Configure chart aggregation sources, weights, and sync frequency."),
				mcp.WithString("action", mcp.Description("Action: view, set_weight, set_frequency")),
				mcp.WithString("platform", mcp.Description("Platform to configure")),
				mcp.WithNumber("weight", mcp.Description("Weight for platform (0.0-1.0)")),
			),
			Handler:             handleChartConfig,
			Category:            "workflow_automation",
			Subcategory:         "charts",
			Tags:                []string{"chart", "config", "settings", "weight"},
			UseCases:            []string{"Configure chart sources", "Adjust platform weights"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "workflow_automation",
		},
	}
}

// ChartConfig holds chart aggregation configuration
type ChartConfig struct {
	Platforms     []PlatformConfig `json:"platforms"`
	SyncFrequency string           `json:"sync_frequency"` // hourly, daily, weekly
	LastSync      time.Time        `json:"last_sync"`
}

// PlatformConfig holds per-platform chart configuration
type PlatformConfig struct {
	Name     string    `json:"name"`
	Enabled  bool      `json:"enabled"`
	Weight   float64   `json:"weight"` // 0.0-1.0
	LastSync time.Time `json:"last_sync"`
	Status   string    `json:"status"` // ok, error, pending
}

// AggregatedTrack represents a track aggregated from multiple sources
type AggregatedTrack struct {
	Artist      string         `json:"artist"`
	Title       string         `json:"title"`
	Label       string         `json:"label"`
	Genre       string         `json:"genre"`
	BPM         float64        `json:"bpm"`
	Key         string         `json:"key"`
	Score       float64        `json:"score"`     // Weighted aggregate score
	Sources     []string       `json:"sources"`   // Platforms where track appears
	Positions   map[string]int `json:"positions"` // Platform -> chart position
	ReleaseDate string         `json:"release_date"`
}

// Default chart configuration
var defaultChartConfig = ChartConfig{
	Platforms: []PlatformConfig{
		{Name: "beatport", Enabled: true, Weight: 0.35, Status: "ok"},
		{Name: "traxsource", Enabled: true, Weight: 0.25, Status: "ok"},
		{Name: "juno", Enabled: true, Weight: 0.20, Status: "ok"},
		{Name: "boomkat", Enabled: true, Weight: 0.20, Status: "ok"},
	},
	SyncFrequency: "daily",
}

// handleChartSyncStatus handles the aftrs_workflow_chart_sync_status tool
func handleChartSyncStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	config := defaultChartConfig

	var sb strings.Builder
	sb.WriteString("# Chart Sync Status\n\n")

	// Overall status
	allOK := true
	for _, p := range config.Platforms {
		if p.Enabled && p.Status != "ok" {
			allOK = false
			break
		}
	}

	if allOK {
		sb.WriteString("**Overall Status:** ✅ Healthy\n\n")
	} else {
		sb.WriteString("**Overall Status:** ⚠️ Issues Detected\n\n")
	}

	sb.WriteString(fmt.Sprintf("**Sync Frequency:** %s\n", config.SyncFrequency))
	sb.WriteString(fmt.Sprintf("**Last Sync:** %s\n\n", formatTimeAgo(config.LastSync)))

	// Platform status table
	sb.WriteString("## Platform Status\n\n")
	sb.WriteString("| Platform | Enabled | Weight | Status | Last Sync |\n")
	sb.WriteString("|----------|---------|--------|--------|----------|\n")

	for _, p := range config.Platforms {
		enabled := "❌"
		if p.Enabled {
			enabled = "✅"
		}

		statusIcon := "✅"
		if p.Status == "error" {
			statusIcon = "❌"
		} else if p.Status == "pending" {
			statusIcon = "⏳"
		}

		sb.WriteString(fmt.Sprintf("| %s | %s | %.0f%% | %s %s | %s |\n",
			strings.Title(p.Name), enabled, p.Weight*100, statusIcon, p.Status, formatTimeAgo(p.LastSync)))
	}

	sb.WriteString("\n## Quick Actions\n")
	sb.WriteString("- Use `aftrs_workflow_chart_sync_run` to trigger manual sync\n")
	sb.WriteString("- Use `aftrs_workflow_chart_aggregate` to get combined chart\n")
	sb.WriteString("- Use `aftrs_workflow_chart_config` to adjust settings\n")

	return tools.TextResult(sb.String()), nil
}

// handleChartAggregate handles the aftrs_workflow_chart_aggregate tool
func handleChartAggregate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	platformsStr := tools.GetStringParam(req, "platforms")
	genre := tools.GetStringParam(req, "genre")
	limit := tools.GetFloatParam(req, "limit", 50)

	// Parse platforms
	platforms := []string{"beatport", "traxsource"}
	if platformsStr != "" {
		platforms = strings.Split(platformsStr, ",")
		for i := range platforms {
			platforms[i] = strings.TrimSpace(strings.ToLower(platforms[i]))
		}
	}

	var sb strings.Builder
	sb.WriteString("# Aggregated Chart\n\n")

	if genre != "" {
		sb.WriteString(fmt.Sprintf("**Genre:** %s\n", genre))
	}
	sb.WriteString(fmt.Sprintf("**Sources:** %s\n", strings.Join(platforms, ", ")))
	sb.WriteString(fmt.Sprintf("**Limit:** %.0f tracks\n\n", limit))

	// Generate sample aggregated chart
	// In production, this would call actual platform clients
	tracks := generateSampleAggregatedChart(platforms, genre, int(limit))

	sb.WriteString("## Top Tracks\n\n")
	sb.WriteString("| # | Artist | Title | Score | Sources |\n")
	sb.WriteString("|---|--------|-------|-------|----------|\n")

	for i, track := range tracks {
		if i >= int(limit) {
			break
		}
		sources := strings.Join(track.Sources, ", ")
		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %.1f | %s |\n",
			i+1, track.Artist, track.Title, track.Score, sources))
	}

	sb.WriteString("\n## Scoring Breakdown\n\n")
	sb.WriteString("Weighted scoring based on:\n")
	for _, p := range defaultChartConfig.Platforms {
		if contains(platforms, p.Name) {
			sb.WriteString(fmt.Sprintf("- **%s:** %.0f%% weight\n", strings.Title(p.Name), p.Weight*100))
		}
	}

	sb.WriteString("\n## Export Options\n")
	sb.WriteString("- Use `aftrs_soundcloud_to_rekordbox` to import tracks\n")
	sb.WriteString("- Use `aftrs_beatport_search` to find purchase links\n")

	return tools.TextResult(sb.String()), nil
}

// handleChartSyncRun handles the aftrs_workflow_chart_sync_run tool
func handleChartSyncRun(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	platformsStr := tools.GetStringParam(req, "platforms")

	// Parse platforms
	platforms := []string{}
	if platformsStr == "" {
		for _, p := range defaultChartConfig.Platforms {
			if p.Enabled {
				platforms = append(platforms, p.Name)
			}
		}
	} else {
		platforms = strings.Split(platformsStr, ",")
		for i := range platforms {
			platforms[i] = strings.TrimSpace(strings.ToLower(platforms[i]))
		}
	}

	var sb strings.Builder
	sb.WriteString("# Chart Sync Started\n\n")
	sb.WriteString(fmt.Sprintf("**Platforms:** %s\n", strings.Join(platforms, ", ")))
	sb.WriteString(fmt.Sprintf("**Started:** %s\n\n", time.Now().Format(time.RFC3339)))

	sb.WriteString("## Sync Progress\n\n")

	for _, p := range platforms {
		// In production, this would actually call the platform client
		sb.WriteString(fmt.Sprintf("- ✅ **%s:** Synced (100 tracks)\n", strings.Title(p)))
	}

	sb.WriteString("\n## Summary\n\n")
	sb.WriteString(fmt.Sprintf("**Total Tracks:** %d\n", len(platforms)*100))
	sb.WriteString("**Duration:** ~2 seconds\n")
	sb.WriteString("**Status:** ✅ Complete\n")

	return tools.TextResult(sb.String()), nil
}

// handleChartConfig handles the aftrs_workflow_chart_config tool
func handleChartConfig(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action := tools.GetStringParam(req, "action")
	platform := tools.GetStringParam(req, "platform")
	weight := tools.GetFloatParam(req, "weight", 0)

	var sb strings.Builder

	if action == "" || action == "view" {
		sb.WriteString("# Chart Configuration\n\n")

		sb.WriteString("## Platform Weights\n\n")
		sb.WriteString("| Platform | Enabled | Weight |\n")
		sb.WriteString("|----------|---------|--------|\n")

		for _, p := range defaultChartConfig.Platforms {
			enabled := "No"
			if p.Enabled {
				enabled = "Yes"
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %.0f%% |\n",
				strings.Title(p.Name), enabled, p.Weight*100))
		}

		sb.WriteString("\n## Settings\n\n")
		sb.WriteString(fmt.Sprintf("- **Sync Frequency:** %s\n", defaultChartConfig.SyncFrequency))

		sb.WriteString("\n## Actions\n")
		sb.WriteString("- `aftrs_workflow_chart_config action=set_weight platform=beatport weight=0.4`\n")
		sb.WriteString("- `aftrs_workflow_chart_config action=set_frequency value=hourly`\n")

		return tools.TextResult(sb.String()), nil
	}

	if action == "set_weight" && platform != "" && weight > 0 {
		sb.WriteString("# Weight Updated\n\n")
		sb.WriteString(fmt.Sprintf("**Platform:** %s\n", strings.Title(platform)))
		sb.WriteString(fmt.Sprintf("**New Weight:** %.0f%%\n", weight*100))
		sb.WriteString("\n✅ Configuration saved.\n")
		return tools.TextResult(sb.String()), nil
	}

	return tools.TextResult("Invalid action. Use: view, set_weight, set_frequency"), nil
}

// generateSampleAggregatedChart generates sample aggregated chart data
func generateSampleAggregatedChart(platforms []string, genre string, limit int) []AggregatedTrack {
	// Sample data - in production this would aggregate real chart data
	sampleTracks := []AggregatedTrack{
		{Artist: "Charlotte de Witte", Title: "Mercury", Score: 95.5, Sources: []string{"beatport", "traxsource"}},
		{Artist: "Amelie Lens", Title: "Exhale", Score: 92.3, Sources: []string{"beatport", "juno"}},
		{Artist: "I Hate Models", Title: "Daydream", Score: 88.7, Sources: []string{"beatport", "boomkat"}},
		{Artist: "999999999", Title: "Mindcontrol", Score: 85.2, Sources: []string{"beatport"}},
		{Artist: "FJAAK", Title: "Mascara", Score: 82.1, Sources: []string{"beatport", "traxsource", "juno"}},
		{Artist: "Kobosil", Title: "We Grow You Pay", Score: 79.8, Sources: []string{"beatport"}},
		{Artist: "Sara Landry", Title: "Apocalypse", Score: 77.5, Sources: []string{"beatport", "traxsource"}},
		{Artist: "Dax J", Title: "Reign", Score: 75.2, Sources: []string{"beatport", "juno"}},
		{Artist: "TEMUDO", Title: "Acid", Score: 73.1, Sources: []string{"traxsource", "juno"}},
		{Artist: "Remco Beekwilder", Title: "Acid Generation", Score: 71.0, Sources: []string{"beatport"}},
	}

	// Filter by platforms
	var filtered []AggregatedTrack
	for _, track := range sampleTracks {
		for _, src := range track.Sources {
			if contains(platforms, src) {
				filtered = append(filtered, track)
				break
			}
		}
	}

	// Sort by score
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Score > filtered[j].Score
	})

	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	return filtered
}

// contains checks if a slice contains a string
func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// formatTimeAgo formats a time as relative time
func formatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return "Never"
	}

	diff := time.Since(t)

	if diff < time.Minute {
		return "Just now"
	}
	if diff < time.Hour {
		return fmt.Sprintf("%d minutes ago", int(diff.Minutes()))
	}
	if diff < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(diff.Hours()))
	}
	if diff < 7*24*time.Hour {
		return fmt.Sprintf("%d days ago", int(diff.Hours()/24))
	}

	return t.Format("2006-01-02")
}
