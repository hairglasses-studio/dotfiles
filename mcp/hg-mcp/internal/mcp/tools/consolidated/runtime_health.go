package consolidated

import (
	"context"
	"fmt"
	"sort"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/config"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

func init() {
	tools.GetRegistry().RegisterModule(&runtimeHealthModule{})
}

type runtimeHealthModule struct{}

func (m *runtimeHealthModule) Name() string        { return "runtime_health" }
func (m *runtimeHealthModule) Description() string { return "Runtime group health monitoring" }

func (m *runtimeHealthModule) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_runtime_health",
				mcp.WithDescription("Check health/configuration status of runtime groups. Shows total tools, configured vs unconfigured counts, and lists unconfigured categories per group."),
				mcp.WithString("runtime_group", mcp.Description("Runtime group to check (e.g. dj_music, lighting, vj_video). Leave empty for all groups.")),
			),
			Handler:    handleRuntimeHealth,
			Category:   "consolidated",
			Subcategory: "health",
			Tags:       []string{"health", "runtime", "config", "status", "monitoring", "readiness"},
			UseCases:   []string{"Check which tools are configured", "Pre-show readiness audit", "Diagnose unconfigured services"},
			Complexity: tools.ComplexitySimple,
		},
	}
}

// runtimeGroupHealth represents the health status of a single runtime group.
type runtimeGroupHealth struct {
	Group                  string   `json:"group"`
	TotalTools             int      `json:"total_tools"`
	ConfiguredTools        int      `json:"configured_tools"`
	UnconfiguredTools      int      `json:"unconfigured_tools"`
	HealthPercent          int      `json:"health_percent"`
	UnconfiguredCategories []string `json:"unconfigured_categories,omitempty"`
}

func handleRuntimeHealth(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	group := tools.GetStringParam(req, "runtime_group")
	registry := tools.GetRegistry()

	// Get per-group tool counts and unconfigured categories (keyed by category).
	groupStats := registry.GetRuntimeGroupStats()
	unconfiguredCats := registry.GetUnconfiguredCategories() // map[category]toolCount

	// Build a map of runtime_group -> list of unconfigured categories with counts.
	// We need to walk all tool definitions to associate categories with their runtime group.
	allTools := registry.GetAllToolDefinitions()

	// categoryGroup maps each category to its runtime group.
	categoryGroup := make(map[string]string)
	for _, td := range allTools {
		rg := td.RuntimeGroup
		if rg == "" {
			rg = "unassigned"
		}
		categoryGroup[td.Category] = rg
	}

	// unconfiguredByGroup: runtime_group -> category -> tool count
	unconfiguredByGroup := make(map[string]map[string]int)
	for cat, count := range unconfiguredCats {
		rg := categoryGroup[cat]
		if rg == "" {
			rg = "unassigned"
		}
		if unconfiguredByGroup[rg] == nil {
			unconfiguredByGroup[rg] = make(map[string]int)
		}
		unconfiguredByGroup[rg][cat] = count
	}

	// Determine which groups to report.
	var groups []string
	if group != "" {
		if _, ok := groupStats[group]; !ok {
			return tools.ErrorResult(fmt.Errorf("unknown runtime group %q — valid groups: %v", group, sortedKeys(groupStats))), nil
		}
		groups = []string{group}
	} else {
		groups = sortedKeys(groupStats)
	}

	// Build health report.
	report := make([]runtimeGroupHealth, 0, len(groups))
	for _, g := range groups {
		total := groupStats[g]
		unconfiguredToolCount := 0
		var unconfiguredCatNames []string
		if cats, ok := unconfiguredByGroup[g]; ok {
			for cat, cnt := range cats {
				unconfiguredToolCount += cnt
				unconfiguredCatNames = append(unconfiguredCatNames, fmt.Sprintf("%s(%d)", cat, cnt))
			}
			sort.Strings(unconfiguredCatNames)
		}
		configured := total - unconfiguredToolCount
		if configured < 0 {
			configured = 0
		}
		healthPct := 0
		if total > 0 {
			healthPct = (configured * 100) / total
		}

		report = append(report, runtimeGroupHealth{
			Group:                  g,
			TotalTools:             total,
			ConfiguredTools:        configured,
			UnconfiguredTools:      unconfiguredToolCount,
			HealthPercent:          healthPct,
			UnconfiguredCategories: unconfiguredCatNames,
		})
	}

	// Summary stats.
	totalAll := 0
	configuredAll := 0
	for _, h := range report {
		totalAll += h.TotalTools
		configuredAll += h.ConfiguredTools
	}

	// Use config.AuditConfig for a service-level summary.
	cfgReport := config.AuditConfig()

	result := map[string]interface{}{
		"groups":               report,
		"total_tools":          totalAll,
		"configured_tools":     configuredAll,
		"unconfigured_tools":   totalAll - configuredAll,
		"services_configured":  len(cfgReport.Configured),
		"services_missing":     len(cfgReport.Missing),
		"missing_services":     cfgReport.Missing,
	}

	return tools.JSONResult(result), nil
}

// sortedKeys returns the sorted keys of a map[string]int.
func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
