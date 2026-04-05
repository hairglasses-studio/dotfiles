package pages

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/config"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// handlePagesHealth checks the pages module configuration, API connectivity, and page statistics.
func handlePagesHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	score := 100
	var issues []string

	// Check 1: Notion API key configured
	apiKey := config.GetEnv("NOTION_API_KEY", "")
	if apiKey == "" {
		apiKey = config.GetEnv("NOTION_TOKEN", "")
	}
	apiKeyOK := apiKey != ""
	if !apiKeyOK {
		score -= 100
		issues = append(issues, "NOTION_API_KEY / NOTION_TOKEN not set")
	}

	// Check 2: Database ID configured
	dbID := config.GetEnv("HG_PAGES_DATABASE_ID", "")
	dbIDOK := dbID != ""
	if !dbIDOK {
		score -= 50
		issues = append(issues, "HG_PAGES_DATABASE_ID not set")
	}

	// Check 3: API connectivity + page stats (only if both keys are present)
	apiOK := false
	var statusCounts map[string]int
	var categoryCounts map[string]int
	templateCount := 0
	totalPages := 0

	if apiKeyOK && dbIDOK {
		client, err := getClient()
		if err != nil {
			score -= 40
			issues = append(issues, fmt.Sprintf("Client init failed: %v", err))
		} else {
			// Test connectivity with a minimal query
			testQuery := &clients.NotionDatabaseQuery{PageSize: 1}
			_, _, _, err := client.QueryDatabase(ctx, dbID, testQuery)
			if err != nil {
				score -= 40
				issues = append(issues, fmt.Sprintf("API unreachable: %v", err))
			} else {
				apiOK = true

				// Gather page statistics
				statusCounts, categoryCounts, templateCount, totalPages = gatherPageStats(ctx, client, dbID)
			}
		}
	}

	// Clamp score
	if score < 0 {
		score = 0
	}

	// Determine status level
	statusLevel := "healthy"
	if score < 100 {
		statusLevel = "degraded"
	}
	if score <= 10 {
		statusLevel = "critical"
	}

	// Build markdown output
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Pages Module Health\n\n"))
	sb.WriteString(fmt.Sprintf("**Score:** %d/100 — **%s**\n\n", score, statusLevel))

	// Configuration table
	sb.WriteString("## Configuration\n\n")
	sb.WriteString("| Setting | Status |\n")
	sb.WriteString("|---------|--------|\n")
	sb.WriteString(fmt.Sprintf("| NOTION_API_KEY | %s |\n", boolIcon(apiKeyOK)))
	sb.WriteString(fmt.Sprintf("| HG_PAGES_DATABASE_ID | %s |\n", boolIcon(dbIDOK)))
	sb.WriteString(fmt.Sprintf("| API Connectivity | %s |\n", boolIcon(apiOK)))

	if len(issues) > 0 {
		sb.WriteString("\n## Issues\n\n")
		for _, issue := range issues {
			sb.WriteString(fmt.Sprintf("- %s\n", issue))
		}
	}

	if apiOK {
		sb.WriteString(fmt.Sprintf("\n## Statistics\n\n"))
		sb.WriteString(fmt.Sprintf("**Total pages:** %d\n", totalPages))
		sb.WriteString(fmt.Sprintf("**Templates:** %d\n\n", templateCount))

		if len(statusCounts) > 0 {
			sb.WriteString("### By Status\n\n")
			sb.WriteString("| Status | Count |\n")
			sb.WriteString("|--------|-------|\n")
			for _, s := range []string{"Draft", "Active", "Archived"} {
				if c, ok := statusCounts[s]; ok {
					sb.WriteString(fmt.Sprintf("| %s | %d |\n", s, c))
				}
			}
		}

		if len(categoryCounts) > 0 {
			sb.WriteString("\n### By Category\n\n")
			sb.WriteString("| Category | Count |\n")
			sb.WriteString("|----------|-------|\n")
			for cat, count := range categoryCounts {
				label := cat
				if label == "" {
					label = "(none)"
				}
				sb.WriteString(fmt.Sprintf("| %s | %d |\n", label, count))
			}
		}
	}

	return tools.TextResult(sb.String()), nil
}

// gatherPageStats queries pages and templates to build status/category breakdowns.
func gatherPageStats(ctx context.Context, client *clients.NotionClient, dbID string) (statusCounts, categoryCounts map[string]int, templateCount, totalPages int) {
	statusCounts = make(map[string]int)
	categoryCounts = make(map[string]int)

	// Query all non-template pages (up to 100 for stats)
	pageQuery := &clients.NotionDatabaseQuery{
		Filter: map[string]interface{}{
			"property": "Template",
			"checkbox": map[string]interface{}{"equals": false},
		},
		PageSize: 100,
	}
	pages, _, _, err := client.QueryDatabase(ctx, dbID, pageQuery)
	if err == nil {
		totalPages = len(pages)
		for _, p := range pages {
			status := extractSelect(p.Properties, "Status")
			if status != "" {
				statusCounts[status]++
			}
			category := extractSelect(p.Properties, "Category")
			categoryCounts[category]++
		}
	}

	// Query templates
	templateQuery := &clients.NotionDatabaseQuery{
		Filter: map[string]interface{}{
			"property": "Template",
			"checkbox": map[string]interface{}{"equals": true},
		},
		PageSize: 100,
	}
	templates, _, _, err := client.QueryDatabase(ctx, dbID, templateQuery)
	if err == nil {
		templateCount = len(templates)
	}

	return
}

func boolIcon(ok bool) string {
	if ok {
		return "OK"
	}
	return "MISSING"
}
