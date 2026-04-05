// Package vault provides Obsidian vault knowledge management tools for hg-mcp.
package vault

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewVaultClient)

// Module implements the ToolModule interface for vault/knowledge management
type Module struct{}

func (m *Module) Name() string {
	return "vault"
}

func (m *Module) Description() string {
	return "Obsidian vault knowledge management and documentation"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_vault_search",
				mcp.WithDescription("Search the vault for documents matching a query."),
				mcp.WithString("query",
					mcp.Required(),
					mcp.Description("Search query"),
				),
			),
			Handler:             handleVaultSearch,
			Category:            "vault",
			Subcategory:         "search",
			Tags:                []string{"vault", "search", "documentation", "notes"},
			UseCases:            []string{"Find project notes", "Search runbooks"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "vault",
		},
		{
			Tool: mcp.NewTool("aftrs_vault_save",
				mcp.WithDescription("Save a document to the vault."),
				mcp.WithString("path",
					mcp.Required(),
					mcp.Description("Document path within vault (e.g., 'sessions/2024-01-15')"),
				),
				mcp.WithString("content",
					mcp.Required(),
					mcp.Description("Document content (markdown)"),
				),
			),
			Handler:             handleVaultSave,
			Category:            "vault",
			Subcategory:         "documents",
			Tags:                []string{"vault", "save", "write", "notes"},
			UseCases:            []string{"Save session notes", "Create documentation"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "vault",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_project_notes",
				mcp.WithDescription("Get notes for a specific project."),
				mcp.WithString("project",
					mcp.Required(),
					mcp.Description("Project name"),
				),
			),
			Handler:             handleProjectNotes,
			Category:            "vault",
			Subcategory:         "projects",
			Tags:                []string{"vault", "project", "notes", "documentation"},
			UseCases:            []string{"View project documentation", "Get project context"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "vault",
		},
		{
			Tool: mcp.NewTool("aftrs_setlist_list",
				mcp.WithDescription("List all performance setlists."),
			),
			Handler:             handleSetlistList,
			Category:            "vault",
			Subcategory:         "setlists",
			Tags:                []string{"vault", "setlist", "performance", "show"},
			UseCases:            []string{"Browse setlists", "Find past performances"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "vault",
		},
		{
			Tool: mcp.NewTool("aftrs_setlist_get",
				mcp.WithDescription("Get details for a specific setlist."),
				mcp.WithString("name",
					mcp.Required(),
					mcp.Description("Setlist name"),
				),
			),
			Handler:             handleSetlistGet,
			Category:            "vault",
			Subcategory:         "setlists",
			Tags:                []string{"vault", "setlist", "performance", "cues"},
			UseCases:            []string{"Load setlist for show", "Review past performance"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "vault",
		},
		{
			Tool: mcp.NewTool("aftrs_runbook_search",
				mcp.WithDescription("Search for operational runbooks."),
				mcp.WithString("query",
					mcp.Required(),
					mcp.Description("Search query"),
				),
			),
			Handler:             handleRunbookSearch,
			Category:            "vault",
			Subcategory:         "runbooks",
			Tags:                []string{"vault", "runbook", "operations", "howto"},
			UseCases:            []string{"Find troubleshooting guides", "Get setup instructions"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "vault",
		},
		{
			Tool: mcp.NewTool("aftrs_show_log",
				mcp.WithDescription("Log an event for the current show."),
				mcp.WithString("event_type",
					mcp.Required(),
					mcp.Description("Event type: 'start', 'end', 'cue', 'issue', 'note'"),
				),
				mcp.WithString("description",
					mcp.Required(),
					mcp.Description("Event description"),
				),
				mcp.WithString("details",
					mcp.Description("Additional details (optional)"),
				),
			),
			Handler:             handleShowLog,
			Category:            "vault",
			Subcategory:         "shows",
			Tags:                []string{"vault", "show", "log", "events"},
			UseCases:            []string{"Log show events", "Track issues during performance"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "vault",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_show_history",
				mcp.WithDescription("Get past show history."),
				mcp.WithNumber("limit",
					mcp.Description("Maximum number of shows to return (default: 10)"),
				),
			),
			Handler:             handleShowHistory,
			Category:            "vault",
			Subcategory:         "shows",
			Tags:                []string{"vault", "show", "history", "archive"},
			UseCases:            []string{"Review past shows", "Find historical data"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "vault",
		},
		// Session tools
		{
			Tool: mcp.NewTool("aftrs_session_list",
				mcp.WithDescription("List studio sessions for a date or recent sessions."),
				mcp.WithString("date",
					mcp.Description("Date in YYYY-MM-DD format (default: today)"),
				),
				mcp.WithNumber("limit",
					mcp.Description("Number of recent sessions to return if no date specified"),
				),
			),
			Handler:             handleSessionList,
			Category:            "vault",
			Subcategory:         "sessions",
			Tags:                []string{"vault", "session", "list", "history"},
			UseCases:            []string{"View today's sessions", "Find past sessions"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "vault",
		},
		{
			Tool: mcp.NewTool("aftrs_session_get",
				mcp.WithDescription("Get the content of a specific session."),
				mcp.WithString("path",
					mcp.Required(),
					mcp.Description("Session path (from session_list)"),
				),
			),
			Handler:             handleSessionGet,
			Category:            "vault",
			Subcategory:         "sessions",
			Tags:                []string{"vault", "session", "details", "view"},
			UseCases:            []string{"View session details", "Review past session"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "vault",
		},
		{
			Tool: mcp.NewTool("aftrs_session_summary",
				mcp.WithDescription("Get aggregated session summary for a period."),
				mcp.WithString("period",
					mcp.Description("Period: 'daily', 'weekly', or 'monthly' (default: weekly)"),
				),
			),
			Handler:             handleSessionSummary,
			Category:            "vault",
			Subcategory:         "sessions",
			Tags:                []string{"vault", "session", "summary", "analytics"},
			UseCases:            []string{"Weekly activity report", "Session statistics"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "vault",
		},
	}
}

// getClient creates a new vault client

// handleVaultSearch handles the aftrs_vault_search tool
func handleVaultSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Vault Search: '%s'\n\n", query))

	if !client.VaultExists() {
		sb.WriteString("**Vault Not Found**\n\n")
		sb.WriteString(fmt.Sprintf("Expected path: `%s`\n\n", client.VaultPath()))
		sb.WriteString("## Setup\n\n")
		sb.WriteString("Create the vault directory or set the environment variable:\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export AFTRS_VAULT_PATH=~/aftrs-vault\n")
		sb.WriteString("mkdir -p ~/aftrs-vault\n")
		sb.WriteString("```\n")
		return tools.TextResult(sb.String()), nil
	}

	results, err := client.Search(ctx, query)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(results) == 0 {
		sb.WriteString("No documents found.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** documents:\n\n", len(results)))

	for _, result := range results {
		sb.WriteString(fmt.Sprintf("### %s\n", result.Document.Name))
		sb.WriteString(fmt.Sprintf("**Path:** `%s`\n", result.Document.Path))
		sb.WriteString(fmt.Sprintf("**Modified:** %s\n\n", result.Document.Modified.Format("2006-01-02 15:04")))
		if result.Snippet != "" {
			sb.WriteString(fmt.Sprintf("> %s\n\n", result.Snippet))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleVaultSave handles the aftrs_vault_save tool
func handleVaultSave(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, errResult := tools.RequireStringParam(req, "path")
	if errResult != nil {
		return errResult, nil
	}
	content, errResult := tools.RequireStringParam(req, "content")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SaveDocument(ctx, path, content)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Document saved to: %s", path)), nil
}

// handleProjectNotes handles the aftrs_project_notes tool
func handleProjectNotes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	project, errResult := tools.RequireStringParam(req, "project")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	notes, err := client.GetProjectNotes(ctx, project)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Project: %s\n\n", project))

	if len(notes.Documents) == 0 {
		sb.WriteString("No documentation found for this project.\n\n")
		sb.WriteString("Create project notes in:\n")
		sb.WriteString(fmt.Sprintf("`%s/projects/%s/`\n", client.VaultPath(), project))
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** documents:\n\n", len(notes.Documents)))
	sb.WriteString("| Document | Modified | Size |\n")
	sb.WriteString("|----------|----------|------|\n")

	for _, doc := range notes.Documents {
		sb.WriteString(fmt.Sprintf("| %s | %s | %d bytes |\n",
			doc.Name, doc.Modified.Format("2006-01-02"), doc.Size))
	}

	if !notes.LastUpdated.IsZero() {
		sb.WriteString(fmt.Sprintf("\n**Last Updated:** %s\n", notes.LastUpdated.Format("2006-01-02 15:04")))
	}

	return tools.TextResult(sb.String()), nil
}

// handleSetlistList handles the aftrs_setlist_list tool
func handleSetlistList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	setlists, err := client.ListSetlists(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Setlists\n\n")

	if len(setlists) == 0 {
		sb.WriteString("No setlists found.\n\n")
		sb.WriteString("Create setlists in:\n")
		sb.WriteString(fmt.Sprintf("`%s/setlists/`\n", client.VaultPath()))
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** setlists:\n\n", len(setlists)))
	sb.WriteString("| Name | Date | Venue |\n")
	sb.WriteString("|------|------|-------|\n")

	for _, setlist := range setlists {
		venue := setlist.Venue
		if venue == "" {
			venue = "-"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
			setlist.Name, setlist.Date.Format("2006-01-02"), venue))
	}

	return tools.TextResult(sb.String()), nil
}

// handleSetlistGet handles the aftrs_setlist_get tool
func handleSetlistGet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	setlist, err := client.GetSetlist(ctx, name)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Setlist: %s\n\n", setlist.Name))
	sb.WriteString(fmt.Sprintf("**Date:** %s\n", setlist.Date.Format("2006-01-02")))
	if setlist.Venue != "" {
		sb.WriteString(fmt.Sprintf("**Venue:** %s\n", setlist.Venue))
	}

	if len(setlist.Items) > 0 {
		sb.WriteString("\n## Items\n\n")
		sb.WriteString("| # | Name | Duration | Cue |\n")
		sb.WriteString("|---|------|----------|-----|\n")

		for _, item := range setlist.Items {
			duration := item.Duration
			if duration == "" {
				duration = "-"
			}
			cue := item.Cue
			if cue == "" {
				cue = "-"
			}
			sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n",
				item.Order, item.Name, duration, cue))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleRunbookSearch handles the aftrs_runbook_search tool
func handleRunbookSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, errResult := tools.RequireStringParam(req, "query")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	runbooks, err := client.SearchRunbooks(ctx, query)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Runbook Search: '%s'\n\n", query))

	if len(runbooks) == 0 {
		sb.WriteString("No runbooks found.\n\n")
		sb.WriteString("Create runbooks in:\n")
		sb.WriteString(fmt.Sprintf("`%s/runbooks/`\n", client.VaultPath()))
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** runbooks:\n\n", len(runbooks)))
	sb.WriteString("| Name | Category | Path |\n")
	sb.WriteString("|------|----------|------|\n")

	for _, rb := range runbooks {
		sb.WriteString(fmt.Sprintf("| %s | %s | `%s` |\n",
			rb.Name, rb.Category, rb.Path))
	}

	return tools.TextResult(sb.String()), nil
}

// handleShowLog handles the aftrs_show_log tool
func handleShowLog(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	eventType, errResult := tools.RequireStringParam(req, "event_type")
	if errResult != nil {
		return errResult, nil
	}
	description, errResult := tools.RequireStringParam(req, "description")
	if errResult != nil {
		return errResult, nil
	}
	details := tools.GetStringParam(req, "details")

	// Validate event type
	validTypes := map[string]bool{
		"start": true, "end": true, "cue": true, "issue": true, "note": true,
	}
	if !validTypes[eventType] {
		return tools.ErrorResult(fmt.Errorf("invalid event_type: %s (use start, end, cue, issue, or note)", eventType)), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.LogShowEvent(ctx, eventType, description, details)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Logged %s event: %s", eventType, description)), nil
}

// handleShowHistory handles the aftrs_show_history tool
func handleShowHistory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := tools.GetIntParam(req, "limit", 10)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	history, err := client.GetShowHistory(ctx, limit)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Show History\n\n")

	if len(history) == 0 {
		sb.WriteString("No shows logged yet.\n\n")
		sb.WriteString("Use `aftrs_show_log` to log show events.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Last **%d** shows:\n\n", len(history)))
	sb.WriteString("| Date | Venue | Duration |\n")
	sb.WriteString("|------|-------|----------|\n")

	for _, show := range history {
		venue := show.Venue
		if venue == "" {
			venue = "-"
		}
		duration := show.Duration
		if duration == "" {
			duration = "-"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n",
			show.Date.Format("2006-01-02"), venue, duration))
	}

	return tools.TextResult(sb.String()), nil
}

// handleSessionList handles the aftrs_session_list tool
func handleSessionList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	dateStr := tools.GetStringParam(req, "date")
	limit := tools.GetIntParam(req, "limit", 10)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder

	if dateStr != "" {
		// List sessions for specific date
		date, err := parseDate(dateStr)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("invalid date format: %s (use YYYY-MM-DD)", dateStr)), nil
		}

		sessions, err := client.ListSessions(ctx, date)
		if err != nil {
			return tools.ErrorResult(err), nil
		}

		sb.WriteString(fmt.Sprintf("# Sessions: %s\n\n", date.Format("Monday, January 2, 2006")))

		if len(sessions) == 0 {
			sb.WriteString("No sessions found for this date.\n")
			return tools.TextResult(sb.String()), nil
		}

		sb.WriteString(fmt.Sprintf("Found **%d** session(s):\n\n", len(sessions)))
		sb.WriteString("| Name | Started | Path |\n")
		sb.WriteString("|------|---------|------|\n")

		for _, s := range sessions {
			sb.WriteString(fmt.Sprintf("| %s | %s | `%s` |\n",
				s.Name, s.StartTime.Format("3:04 PM"), s.Path))
		}
	} else {
		// List recent sessions
		sessions, err := client.GetRecentSessions(ctx, limit)
		if err != nil {
			return tools.ErrorResult(err), nil
		}

		sb.WriteString("# Recent Sessions\n\n")

		if len(sessions) == 0 {
			sb.WriteString("No sessions found.\n\n")
			sb.WriteString("Sessions are stored in:\n")
			sb.WriteString(fmt.Sprintf("`%s/sessions/`\n", client.VaultPath()))
			return tools.TextResult(sb.String()), nil
		}

		sb.WriteString(fmt.Sprintf("Last **%d** session(s):\n\n", len(sessions)))
		sb.WriteString("| Name | Date | Path |\n")
		sb.WriteString("|------|------|------|\n")

		for _, s := range sessions {
			sb.WriteString(fmt.Sprintf("| %s | %s | `%s` |\n",
				s.Name, s.StartTime.Format("2006-01-02 15:04"), s.Path))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleSessionGet handles the aftrs_session_get tool
func handleSessionGet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	path, errResult := tools.RequireStringParam(req, "path")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	content, err := client.GetSessionContent(ctx, path)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(content), nil
}

// handleSessionSummary handles the aftrs_session_summary tool
func handleSessionSummary(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	period := tools.OptionalStringParam(req, "period", "weekly")

	// Validate period
	validPeriods := map[string]bool{"daily": true, "weekly": true, "monthly": true}
	if !validPeriods[period] {
		return tools.ErrorResult(fmt.Errorf("invalid period: %s (use daily, weekly, or monthly)", period)), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	summary, err := client.GetSessionSummary(ctx, period)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	caser := cases.Title(language.English)
	sb.WriteString(fmt.Sprintf("# Session Summary: %s\n\n", caser.String(period)))
	sb.WriteString(fmt.Sprintf("**Period:** %s to %s\n",
		summary.StartDate.Format("Jan 2"),
		summary.EndDate.Format("Jan 2, 2006")))
	sb.WriteString(fmt.Sprintf("**Total Sessions:** %d\n\n", summary.TotalSessions))

	if len(summary.Sessions) == 0 {
		sb.WriteString("No sessions recorded during this period.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("## Sessions\n\n")
	sb.WriteString("| Name | Date | Path |\n")
	sb.WriteString("|------|------|------|\n")

	for _, s := range summary.Sessions {
		sb.WriteString(fmt.Sprintf("| %s | %s | `%s` |\n",
			s.Name, s.StartTime.Format("2006-01-02"), s.Path))
	}

	return tools.TextResult(sb.String()), nil
}

// parseDate parses a date string in YYYY-MM-DD format
func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
