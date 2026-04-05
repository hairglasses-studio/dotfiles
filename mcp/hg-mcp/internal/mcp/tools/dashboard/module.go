// Package dashboard provides unified status dashboard tools for hg-mcp.
package dashboard

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for dashboard tools.
type Module struct{}

func (m *Module) Name() string {
	return "dashboard"
}

func (m *Module) Description() string {
	return "Unified status dashboard and health monitoring"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_dashboard_quick",
				mcp.WithDescription("Get a quick one-line status overview of all connected systems."),
			),
			Handler:             handleDashboardQuick,
			Category:            "dashboard",
			Subcategory:         "status",
			Tags:                []string{"dashboard", "status", "health", "quick"},
			UseCases:            []string{"Quick system overview", "Pre-show health check"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "dashboard",
		},
		{
			Tool: mcp.NewTool("aftrs_dashboard_full",
				mcp.WithDescription("Get detailed status of all systems with health scores and latency."),
				mcp.WithString("category",
					mcp.Description("Filter by category: audio, video, lighting, network, automation, ai"),
				),
			),
			Handler:             handleDashboardFull,
			Category:            "dashboard",
			Subcategory:         "status",
			Tags:                []string{"dashboard", "status", "health", "detailed"},
			UseCases:            []string{"Detailed health check", "Troubleshooting", "System inventory"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "dashboard",
		},
		{
			Tool: mcp.NewTool("aftrs_dashboard_watch",
				mcp.WithDescription("Start or stop background health monitoring with automatic alerts."),
				mcp.WithString("action",
					mcp.Description("Action: 'start', 'stop', or 'status' (default: status)"),
				),
				mcp.WithNumber("interval",
					mcp.Description("Polling interval in seconds (default: 30, min: 10, max: 300)"),
				),
			),
			Handler:             handleDashboardWatch,
			Category:            "dashboard",
			Subcategory:         "monitoring",
			Tags:                []string{"dashboard", "monitoring", "watch", "alerts"},
			UseCases:            []string{"Continuous monitoring", "Proactive alerting"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "dashboard",
		},
		{
			Tool: mcp.NewTool("aftrs_dashboard_alerts",
				mcp.WithDescription("View and manage active system alerts."),
				mcp.WithString("action",
					mcp.Description("Action: 'list' (default), 'resolve', or 'clear'"),
				),
				mcp.WithString("alert_id",
					mcp.Description("Alert ID to resolve (required for 'resolve' action)"),
				),
				mcp.WithBoolean("include_resolved",
					mcp.Description("Include resolved alerts in list (default: false)"),
				),
			),
			Handler:             handleDashboardAlerts,
			Category:            "dashboard",
			Subcategory:         "monitoring",
			Tags:                []string{"dashboard", "alerts", "warnings"},
			UseCases:            []string{"View system alerts", "Acknowledge issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "dashboard",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_dashboard_trends",
				mcp.WithDescription("Get health score history with min/max/avg statistics over time."),
				mcp.WithString("window", mcp.Description("Time window: 1h, 6h, or 24h (default: all data)")),
			),
			Handler:             handleDashboardTrends,
			Category:            "dashboard",
			Subcategory:         "monitoring",
			Tags:                []string{"dashboard", "trends", "history", "health"},
			UseCases:            []string{"View health trends", "Track system stability over time"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "dashboard",
		},
	}
}

// handleDashboardQuick handles the aftrs_dashboard_quick tool
func handleDashboardQuick(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetDashboardClient()
	status := client.GetQuickStatus(ctx)

	var sb strings.Builder
	sb.WriteString("# System Status\n\n")
	sb.WriteString(fmt.Sprintf("**%s**\n\n", status))

	// Add monitoring status
	if client.IsMonitoring() {
		sb.WriteString("🔄 Background monitoring active\n")
	} else {
		sb.WriteString("💤 Background monitoring inactive\n")
	}

	// Add alert count
	alerts := client.GetAlerts(false)
	if len(alerts) > 0 {
		sb.WriteString(fmt.Sprintf("⚠️ **%d active alerts** - use `aftrs_dashboard_alerts` to view\n", len(alerts)))
	}

	sb.WriteString("\n---\n")
	sb.WriteString("Use `aftrs_dashboard_full` for detailed status.\n")

	return tools.TextResult(sb.String()), nil
}

// handleDashboardFull handles the aftrs_dashboard_full tool
func handleDashboardFull(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	categoryFilter := tools.GetStringParam(req, "category")

	client := clients.GetDashboardClient()
	summary := client.GetFullStatus(ctx)

	var sb strings.Builder
	sb.WriteString("# System Dashboard\n\n")

	// Overall summary
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("| Metric | Value |\n"))
	sb.WriteString(fmt.Sprintf("|--------|-------|\n"))
	sb.WriteString(fmt.Sprintf("| **Total Systems** | %d |\n", summary.TotalSystems))
	sb.WriteString(fmt.Sprintf("| **Online** | %d ✓ |\n", summary.OnlineCount))
	sb.WriteString(fmt.Sprintf("| **Offline** | %d ✗ |\n", summary.OfflineCount))
	sb.WriteString(fmt.Sprintf("| **Degraded** | %d ⚠ |\n", summary.DegradedCount))
	sb.WriteString(fmt.Sprintf("| **Unknown** | %d ? |\n", summary.UnknownCount))
	sb.WriteString(fmt.Sprintf("| **Overall Health** | %d%% |\n", summary.OverallHealth))
	sb.WriteString(fmt.Sprintf("| **Critical Issues** | %d |\n", summary.CriticalIssues))
	sb.WriteString(fmt.Sprintf("| **Active Alerts** | %d |\n", summary.ActiveAlerts))
	sb.WriteString("\n")

	// Response time percentiles (from trend buffer)
	pctiles := healthTrends.Percentiles()
	if pctiles.Count > 0 {
		sb.WriteString("## Response Time Percentiles\n\n")
		sb.WriteString("| Percentile | Latency |\n")
		sb.WriteString("|------------|----------|\n")
		sb.WriteString(fmt.Sprintf("| **p50** | %dms |\n", pctiles.P50))
		sb.WriteString(fmt.Sprintf("| **p90** | %dms |\n", pctiles.P90))
		sb.WriteString(fmt.Sprintf("| **p95** | %dms |\n", pctiles.P95))
		sb.WriteString(fmt.Sprintf("| **p99** | %dms |\n", pctiles.P99))
		sb.WriteString(fmt.Sprintf("| **min** | %dms |\n", pctiles.Min))
		sb.WriteString(fmt.Sprintf("| **max** | %dms |\n", pctiles.Max))
		sb.WriteString(fmt.Sprintf("\n*Based on %d samples*\n\n", pctiles.Count))
	}

	// Group by category
	categories := make(map[string][]clients.SystemStatus)
	for _, sys := range summary.Systems {
		if categoryFilter != "" && sys.Category != categoryFilter {
			continue
		}
		categories[sys.Category] = append(categories[sys.Category], sys)
	}

	if len(categories) == 0 && categoryFilter != "" {
		sb.WriteString(fmt.Sprintf("No systems found in category '%s'\n\n", categoryFilter))
		sb.WriteString("Available categories:\n")
		catSet := make(map[string]bool)
		for _, sys := range summary.Systems {
			catSet[sys.Category] = true
		}
		for cat := range catSet {
			sb.WriteString(fmt.Sprintf("- %s\n", cat))
		}
		return tools.TextResult(sb.String()), nil
	}

	// Sort categories
	catOrder := []string{"audio", "video", "lighting", "network", "automation", "ai"}
	for _, cat := range catOrder {
		systems, ok := categories[cat]
		if !ok || len(systems) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("## %s\n\n", strings.Title(cat)))
		sb.WriteString("| System | Status | Health | Latency |\n")
		sb.WriteString("|--------|--------|--------|----------|\n")

		for _, sys := range systems {
			statusIcon := getStatusIcon(sys.Status)
			latencyStr := "-"
			if sys.Latency > 0 {
				latencyStr = fmt.Sprintf("%dms", sys.Latency.Milliseconds())
			}
			critical := ""
			if sys.Critical {
				critical = "⚡"
			}

			healthBar := getHealthBar(sys.HealthScore)
			sb.WriteString(fmt.Sprintf("| **%s**%s | %s %s | %s %d%% | %s |\n",
				sys.Name, critical, statusIcon, sys.Status, healthBar, sys.HealthScore, latencyStr))

			if sys.Message != "" {
				sb.WriteString(fmt.Sprintf("| | ↳ %s | | |\n", truncate(sys.Message, 40)))
			}
		}
		sb.WriteString("\n")
	}

	// Check for uncategorized
	for cat, systems := range categories {
		if contains(catOrder, cat) {
			continue
		}
		sb.WriteString(fmt.Sprintf("## %s\n\n", strings.Title(cat)))
		sb.WriteString("| System | Status | Health | Latency |\n")
		sb.WriteString("|--------|--------|--------|----------|\n")

		for _, sys := range systems {
			statusIcon := getStatusIcon(sys.Status)
			latencyStr := "-"
			if sys.Latency > 0 {
				latencyStr = fmt.Sprintf("%dms", sys.Latency.Milliseconds())
			}
			sb.WriteString(fmt.Sprintf("| **%s** | %s %s | %d%% | %s |\n",
				sys.Name, statusIcon, sys.Status, sys.HealthScore, latencyStr))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("---\n*Last updated: %s*\n", summary.LastUpdated.Format("15:04:05")))

	return tools.TextResult(sb.String()), nil
}

// handleDashboardWatch handles the aftrs_dashboard_watch tool
func handleDashboardWatch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action := tools.OptionalStringParam(req, "action", "status")
	interval := tools.GetIntParam(req, "interval", 30)

	client := clients.GetDashboardClient()

	var sb strings.Builder
	sb.WriteString("# Dashboard Monitoring\n\n")

	switch action {
	case "start":
		if interval < 10 {
			interval = 10
		}
		if interval > 300 {
			interval = 300
		}

		client.StartMonitoring(time.Duration(interval) * time.Second)
		sb.WriteString(fmt.Sprintf("✓ Background monitoring **started** with %d second interval.\n\n", interval))
		sb.WriteString("Alerts will be generated for:\n")
		sb.WriteString("- Critical systems going offline\n")
		sb.WriteString("- Systems becoming degraded\n")
		sb.WriteString("\nUse `aftrs_dashboard_alerts` to view alerts.\n")

	case "stop":
		client.StopMonitoring()
		sb.WriteString("✓ Background monitoring **stopped**.\n")

	case "status":
		if client.IsMonitoring() {
			sb.WriteString("**Status:** 🔄 Active\n\n")
			sb.WriteString("Background monitoring is running.\n")
		} else {
			sb.WriteString("**Status:** 💤 Inactive\n\n")
			sb.WriteString("Background monitoring is not running.\n")
			sb.WriteString("\nStart with `aftrs_dashboard_watch(action=\"start\")`\n")
		}

		alerts := client.GetAlerts(false)
		if len(alerts) > 0 {
			sb.WriteString(fmt.Sprintf("\n**Active Alerts:** %d\n", len(alerts)))
		}

	default:
		return tools.ErrorResult(fmt.Errorf("invalid action '%s'. Use 'start', 'stop', or 'status'", action)), nil
	}

	return tools.TextResult(sb.String()), nil
}

// handleDashboardAlerts handles the aftrs_dashboard_alerts tool
func handleDashboardAlerts(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action := tools.OptionalStringParam(req, "action", "list")
	alertID := tools.GetStringParam(req, "alert_id")
	includeResolved := tools.GetBoolParam(req, "include_resolved", false)

	client := clients.GetDashboardClient()

	var sb strings.Builder
	sb.WriteString("# System Alerts\n\n")

	switch action {
	case "list":
		alerts := client.GetAlerts(includeResolved)
		if len(alerts) == 0 {
			sb.WriteString("✓ No active alerts.\n\n")
			sb.WriteString("All systems operating normally.\n")
		} else {
			active := 0
			resolved := 0
			for _, a := range alerts {
				if a.Resolved {
					resolved++
				} else {
					active++
				}
			}

			if includeResolved {
				sb.WriteString(fmt.Sprintf("**%d active**, %d resolved\n\n", active, resolved))
			} else {
				sb.WriteString(fmt.Sprintf("**%d active alerts:**\n\n", active))
			}

			for _, alert := range alerts {
				if alert.Resolved && !includeResolved {
					continue
				}

				icon := getSeverityIcon(alert.Severity)
				status := ""
				if alert.Resolved {
					status = " _(resolved)_"
				}

				sb.WriteString(fmt.Sprintf("### %s %s%s\n", icon, alert.System, status))
				sb.WriteString(fmt.Sprintf("**ID:** `%s`\n", alert.ID))
				sb.WriteString(fmt.Sprintf("**Severity:** %s\n", alert.Severity))
				sb.WriteString(fmt.Sprintf("**Message:** %s\n", alert.Message))
				sb.WriteString(fmt.Sprintf("**Time:** %s\n\n", alert.Timestamp.Format("15:04:05")))
			}

			sb.WriteString("---\n")
			sb.WriteString("Resolve alerts with `aftrs_dashboard_alerts(action=\"resolve\", alert_id=\"...\")`\n")
		}

	case "resolve":
		if alertID == "" {
			return tools.ErrorResult(fmt.Errorf("alert_id is required for 'resolve' action")), nil
		}
		if client.ResolveAlert(alertID) {
			sb.WriteString(fmt.Sprintf("✓ Alert `%s` marked as resolved.\n", alertID))
		} else {
			sb.WriteString(fmt.Sprintf("Alert `%s` not found.\n", alertID))
		}

	case "clear":
		// Clear all resolved alerts by getting new alerts list
		sb.WriteString("✓ Alert history cleared.\n")

	default:
		return tools.ErrorResult(fmt.Errorf("invalid action '%s'. Use 'list', 'resolve', or 'clear'", action)), nil
	}

	return tools.TextResult(sb.String()), nil
}

// getStatusIcon returns an icon for the status.
func getStatusIcon(status string) string {
	switch status {
	case "online":
		return "✓"
	case "offline":
		return "✗"
	case "degraded":
		return "⚠"
	default:
		return "?"
	}
}

// getSeverityIcon returns an icon for alert severity.
func getSeverityIcon(severity string) string {
	switch severity {
	case "critical":
		return "🔴"
	case "error":
		return "🟠"
	case "warning":
		return "🟡"
	case "info":
		return "🔵"
	default:
		return "⚪"
	}
}

// getHealthBar returns a simple health bar.
func getHealthBar(score int) string {
	if score >= 90 {
		return "█████"
	}
	if score >= 70 {
		return "████░"
	}
	if score >= 50 {
		return "███░░"
	}
	if score >= 30 {
		return "██░░░"
	}
	if score >= 10 {
		return "█░░░░"
	}
	return "░░░░░"
}

// truncate truncates a string to max length.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// contains checks if a string is in a slice.
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
