// Package setlist provides MCP setlist management tools.
package setlist

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for setlist management
type Module struct{}

var getClient = tools.LazyClient(clients.NewSetlistClient)

func (m *Module) Name() string {
	return "setlist"
}

func (m *Module) Description() string {
	return "DJ setlist planning and management"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_setlist_create",
				mcp.WithDescription("Create a new setlist for planning a DJ set or performance."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Setlist name")),
				mcp.WithString("description", mcp.Description("Setlist description")),
				mcp.WithString("venue", mcp.Description("Venue name")),
				mcp.WithString("date", mcp.Description("Performance date (YYYY-MM-DD)")),
			),
			Handler:             handleSetlistCreate,
			Category:            "setlist",
			Subcategory:         "management",
			Tags:                []string{"setlist", "create", "plan", "dj"},
			UseCases:            []string{"Create new DJ set plan", "Plan performance tracklist"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "setlist",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_setlist_list",
				mcp.WithDescription("List all setlists with summary information."),
				mcp.WithNumber("limit", mcp.Description("Maximum setlists to show (default: 10)")),
			),
			Handler:             handleSetlistList,
			Category:            "setlist",
			Subcategory:         "management",
			Tags:                []string{"setlist", "list", "browse"},
			UseCases:            []string{"View all setlists", "Find past setlists"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "setlist",
		},
		{
			Tool: mcp.NewTool("aftrs_setlist_view",
				mcp.WithDescription("View a setlist with all tracks and details."),
				mcp.WithString("setlist_id", mcp.Required(), mcp.Description("Setlist ID")),
			),
			Handler:             handleSetlistView,
			Category:            "setlist",
			Subcategory:         "management",
			Tags:                []string{"setlist", "view", "details"},
			UseCases:            []string{"View setlist tracks", "Check set plan"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "setlist",
		},
		{
			Tool: mcp.NewTool("aftrs_setlist_add_track",
				mcp.WithDescription("Add a track to a setlist."),
				mcp.WithString("setlist_id", mcp.Required(), mcp.Description("Setlist ID")),
				mcp.WithString("title", mcp.Required(), mcp.Description("Track title")),
				mcp.WithString("artist", mcp.Required(), mcp.Description("Track artist")),
				mcp.WithNumber("bpm", mcp.Description("Track BPM")),
				mcp.WithString("key", mcp.Description("Track key (e.g., '5A', 'Dm')")),
				mcp.WithNumber("energy", mcp.Description("Energy level 1-10")),
				mcp.WithString("notes", mcp.Description("Transition notes")),
			),
			Handler:             handleSetlistAddTrack,
			Category:            "setlist",
			Subcategory:         "tracks",
			Tags:                []string{"setlist", "track", "add"},
			UseCases:            []string{"Add track to set plan", "Build setlist"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "setlist",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_setlist_remove_track",
				mcp.WithDescription("Remove a track from a setlist by position."),
				mcp.WithString("setlist_id", mcp.Required(), mcp.Description("Setlist ID")),
				mcp.WithNumber("position", mcp.Required(), mcp.Description("Track position to remove (1-indexed)")),
			),
			Handler:             handleSetlistRemoveTrack,
			Category:            "setlist",
			Subcategory:         "tracks",
			Tags:                []string{"setlist", "track", "remove", "delete"},
			UseCases:            []string{"Remove track from set", "Edit setlist"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "setlist",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_setlist_reorder",
				mcp.WithDescription("Move a track to a new position in the setlist."),
				mcp.WithString("setlist_id", mcp.Required(), mcp.Description("Setlist ID")),
				mcp.WithNumber("from_position", mcp.Required(), mcp.Description("Current track position")),
				mcp.WithNumber("to_position", mcp.Required(), mcp.Description("New track position")),
			),
			Handler:             handleSetlistReorder,
			Category:            "setlist",
			Subcategory:         "tracks",
			Tags:                []string{"setlist", "reorder", "move"},
			UseCases:            []string{"Reorder set tracks", "Adjust flow"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "setlist",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_setlist_analyze",
				mcp.WithDescription("Analyze a setlist for BPM flow, key compatibility, and energy progression."),
				mcp.WithString("setlist_id", mcp.Required(), mcp.Description("Setlist ID")),
			),
			Handler:             handleSetlistAnalyze,
			Category:            "setlist",
			Subcategory:         "analysis",
			Tags:                []string{"setlist", "analyze", "flow", "bpm", "key"},
			UseCases:            []string{"Check set flow", "Find transition issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "setlist",
		},
		{
			Tool: mcp.NewTool("aftrs_setlist_export",
				mcp.WithDescription("Export a setlist to various formats."),
				mcp.WithString("setlist_id", mcp.Required(), mcp.Description("Setlist ID")),
				mcp.WithString("format", mcp.Description("Export format: text, m3u, json (default: text)")),
			),
			Handler:             handleSetlistExport,
			Category:            "setlist",
			Subcategory:         "export",
			Tags:                []string{"setlist", "export", "m3u", "playlist"},
			UseCases:            []string{"Export to playlist", "Share setlist"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "setlist",
		},
	}
}

func handleSetlistCreate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := tools.GetStringParam(req, "name")
	description := tools.GetStringParam(req, "description")
	venue := tools.GetStringParam(req, "venue")
	dateStr := tools.GetStringParam(req, "date")

	var date time.Time
	if dateStr != "" {
		var err error
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("invalid date format: %s (use YYYY-MM-DD)", dateStr)), nil
		}
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	setlist, err := client.CreateSetlist(ctx, name, description, venue, date)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Setlist Created\n\n")
	sb.WriteString(fmt.Sprintf("**ID:** `%s`\n", setlist.ID))
	sb.WriteString(fmt.Sprintf("**Name:** %s\n", setlist.Name))
	if description != "" {
		sb.WriteString(fmt.Sprintf("**Description:** %s\n", description))
	}
	if venue != "" {
		sb.WriteString(fmt.Sprintf("**Venue:** %s\n", venue))
	}
	if !date.IsZero() {
		sb.WriteString(fmt.Sprintf("**Date:** %s\n", date.Format("2006-01-02")))
	}
	sb.WriteString("\nUse `aftrs_setlist_add_track` to add tracks to this setlist.\n")

	return tools.TextResult(sb.String()), nil
}

func handleSetlistList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := tools.GetIntParam(req, "limit", 10)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	setlists := client.ListSetlists(ctx)

	var sb strings.Builder
	sb.WriteString("# Setlists\n\n")

	if len(setlists) == 0 {
		sb.WriteString("No setlists found.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("**Total:** %d setlists\n\n", len(setlists)))
	sb.WriteString("| Name | Venue | Date | Tracks | Updated |\n")
	sb.WriteString("|------|-------|------|--------|----------|\n")

	count := 0
	for _, s := range setlists {
		if count >= limit {
			break
		}

		venue := "-"
		if s.Venue != "" {
			venue = s.Venue
		}

		date := "-"
		if !s.Date.IsZero() {
			date = s.Date.Format("01-02")
		}

		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %d | %s |\n",
			s.Name,
			venue,
			date,
			len(s.Tracks),
			s.UpdatedAt.Format("01-02 15:04")))
		count++
	}

	return tools.TextResult(sb.String()), nil
}

func handleSetlistView(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	setlistID := tools.GetStringParam(req, "setlist_id")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	setlist, err := client.GetSetlist(ctx, setlistID)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# %s\n\n", setlist.Name))

	if setlist.Description != "" {
		sb.WriteString(fmt.Sprintf("*%s*\n\n", setlist.Description))
	}

	if setlist.Venue != "" {
		sb.WriteString(fmt.Sprintf("**Venue:** %s\n", setlist.Venue))
	}
	if !setlist.Date.IsZero() {
		sb.WriteString(fmt.Sprintf("**Date:** %s\n", setlist.Date.Format("2006-01-02")))
	}
	sb.WriteString(fmt.Sprintf("**ID:** `%s`\n\n", setlist.ID))

	if len(setlist.Tracks) == 0 {
		sb.WriteString("*No tracks yet*\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("## Tracks\n\n")
	for _, track := range setlist.Tracks {
		status := ""
		if track.Played {
			status = " ✅"
		}

		info := ""
		if track.BPM > 0 {
			info += fmt.Sprintf(" %.1f BPM", track.BPM)
		}
		if track.Key != "" {
			info += fmt.Sprintf(" [%s]", track.Key)
		}
		if track.Energy > 0 {
			info += fmt.Sprintf(" E%d", track.Energy)
		}

		sb.WriteString(fmt.Sprintf("%d. **%s** - %s%s%s\n",
			track.Position, track.Title, track.Artist, info, status))

		if track.Notes != "" {
			sb.WriteString(fmt.Sprintf("   → *%s*\n", track.Notes))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleSetlistAddTrack(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	setlistID := tools.GetStringParam(req, "setlist_id")
	title := tools.GetStringParam(req, "title")
	artist := tools.GetStringParam(req, "artist")
	bpm := tools.GetFloatParam(req, "bpm", 0)
	key := tools.GetStringParam(req, "key")
	energy := tools.GetIntParam(req, "energy", 0)
	notes := tools.GetStringParam(req, "notes")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	track := &clients.SetlistTrack{
		Title:  title,
		Artist: artist,
		BPM:    bpm,
		Key:    key,
		Energy: energy,
		Notes:  notes,
	}

	track, err = client.AddTrack(ctx, setlistID, track)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Track Added\n\n")
	sb.WriteString(fmt.Sprintf("**#%d:** %s - %s\n", track.Position, track.Title, track.Artist))
	if bpm > 0 {
		sb.WriteString(fmt.Sprintf("**BPM:** %.1f\n", bpm))
	}
	if key != "" {
		sb.WriteString(fmt.Sprintf("**Key:** %s\n", key))
	}
	if energy > 0 {
		sb.WriteString(fmt.Sprintf("**Energy:** %d\n", energy))
	}
	if notes != "" {
		sb.WriteString(fmt.Sprintf("**Notes:** %s\n", notes))
	}

	return tools.TextResult(sb.String()), nil
}

func handleSetlistRemoveTrack(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	setlistID := tools.GetStringParam(req, "setlist_id")
	position := tools.GetIntParam(req, "position", 0)

	if position < 1 {
		return tools.ErrorResult(fmt.Errorf("position must be >= 1")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.RemoveTrack(ctx, setlistID, position); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Track at position %d removed.", position)), nil
}

func handleSetlistReorder(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	setlistID := tools.GetStringParam(req, "setlist_id")
	fromPos := tools.GetIntParam(req, "from_position", 0)
	toPos := tools.GetIntParam(req, "to_position", 0)

	if fromPos < 1 || toPos < 1 {
		return tools.ErrorResult(fmt.Errorf("positions must be >= 1")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.ReorderTrack(ctx, setlistID, fromPos, toPos); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Track moved from position %d to %d.", fromPos, toPos)), nil
}

func handleSetlistAnalyze(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	setlistID := tools.GetStringParam(req, "setlist_id")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	analysis, err := client.AnalyzeSetlist(ctx, setlistID)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Setlist Analysis\n\n")

	sb.WriteString("## Overview\n\n")
	sb.WriteString(fmt.Sprintf("- **Total Tracks:** %d\n", analysis.TotalTracks))
	sb.WriteString(fmt.Sprintf("- **Total Duration:** %s\n", analysis.TotalDuration.Round(time.Minute)))
	sb.WriteString(fmt.Sprintf("- **Unique Artists:** %d\n", analysis.UniqueArtists))
	sb.WriteString("\n")

	sb.WriteString("## BPM Analysis\n\n")
	sb.WriteString(fmt.Sprintf("- **Average BPM:** %.1f\n", analysis.AverageBPM))
	sb.WriteString(fmt.Sprintf("- **BPM Range:** %s\n", analysis.BPMRange))
	sb.WriteString("\n")

	if len(analysis.EnergyFlow) > 0 {
		sb.WriteString("## Energy Flow\n\n")
		sb.WriteString("```\n")
		for i, energy := range analysis.EnergyFlow {
			bar := strings.Repeat("█", energy)
			sb.WriteString(fmt.Sprintf("%2d. %s %d\n", i+1, bar, energy))
		}
		sb.WriteString("```\n\n")
	}

	if len(analysis.LargeBPMJumps) > 0 {
		sb.WriteString("## Transition Warnings\n\n")
		sb.WriteString("### Large BPM Jumps\n")
		for _, jump := range analysis.LargeBPMJumps {
			sb.WriteString(fmt.Sprintf("- ⚠️ %s\n", jump))
		}
		sb.WriteString("\n")
	}

	if len(analysis.KeyClashes) > 0 {
		sb.WriteString("### Key Changes\n")
		for _, clash := range analysis.KeyClashes {
			sb.WriteString(fmt.Sprintf("- 🎹 %s\n", clash))
		}
	}

	if len(analysis.LargeBPMJumps) == 0 && len(analysis.KeyClashes) == 0 {
		sb.WriteString("✅ No major transition issues detected.\n")
	}

	return tools.TextResult(sb.String()), nil
}

func handleSetlistExport(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	setlistID := tools.GetStringParam(req, "setlist_id")
	format := tools.OptionalStringParam(req, "format", "text")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	exported, err := client.ExportSetlist(ctx, setlistID, format)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Setlist Export (%s)\n\n", format))
	sb.WriteString("```\n")
	sb.WriteString(exported)
	sb.WriteString("```\n")

	return tools.TextResult(sb.String()), nil
}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
