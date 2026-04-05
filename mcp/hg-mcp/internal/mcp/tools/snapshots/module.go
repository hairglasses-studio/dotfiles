// Package snapshots provides MCP show state snapshot tools.
package snapshots

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for snapshots
type Module struct{}

var getClient = tools.LazyClient(clients.NewSnapshotsClient)

func (m *Module) Name() string {
	return "snapshots"
}

func (m *Module) Description() string {
	return "Show state snapshots for capturing and recalling system states"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_snapshot_capture",
				mcp.WithDescription("Capture current state of all systems as a snapshot. Saves Ableton, Resolume, grandMA3, OBS, and Showkontrol states."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Name for this snapshot")),
				mcp.WithString("description", mcp.Description("Optional description")),
				mcp.WithString("systems", mcp.Description("Comma-separated systems to capture (default: all). Options: ableton, resolume, grandma3, obs, showkontrol")),
				mcp.WithString("tags", mcp.Description("Comma-separated tags for organization")),
			),
			Handler:             handleSnapshotCapture,
			Category:            "snapshots",
			Subcategory:         "management",
			Tags:                []string{"snapshot", "capture", "state", "backup"},
			UseCases:            []string{"Save show state before changes", "Create restore point", "Backup current settings"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "snapshots",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_snapshot_recall",
				mcp.WithDescription("Restore system states from a previously captured snapshot."),
				mcp.WithString("snapshot_id", mcp.Required(), mcp.Description("Snapshot ID to recall")),
				mcp.WithString("systems", mcp.Description("Comma-separated systems to restore (default: all in snapshot)")),
			),
			Handler:             handleSnapshotRecall,
			Category:            "snapshots",
			Subcategory:         "management",
			Tags:                []string{"snapshot", "recall", "restore", "state"},
			UseCases:            []string{"Restore show state", "Rollback changes", "Quick scene reset"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "snapshots",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_snapshot_list",
				mcp.WithDescription("List all available snapshots with details."),
				mcp.WithString("snapshot_id", mcp.Description("Get details of a specific snapshot")),
			),
			Handler:             handleSnapshotList,
			Category:            "snapshots",
			Subcategory:         "management",
			Tags:                []string{"snapshot", "list", "view"},
			UseCases:            []string{"View available snapshots", "Find snapshot to restore"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "snapshots",
		},
		{
			Tool: mcp.NewTool("aftrs_snapshot_diff",
				mcp.WithDescription("Compare two snapshots and show differences."),
				mcp.WithString("snapshot1", mcp.Required(), mcp.Description("First snapshot ID")),
				mcp.WithString("snapshot2", mcp.Required(), mcp.Description("Second snapshot ID")),
			),
			Handler:             handleSnapshotDiff,
			Category:            "snapshots",
			Subcategory:         "management",
			Tags:                []string{"snapshot", "diff", "compare"},
			UseCases:            []string{"Compare show states", "Track changes", "Audit modifications"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "snapshots",
		},
		{
			Tool: mcp.NewTool("aftrs_snapshot_auto_configure",
				mcp.WithDescription("Configure automatic snapshot triggers (pre-show, post-show, timed intervals)."),
				mcp.WithString("action", mcp.Description("Action: status (default), enable, or disable")),
				mcp.WithString("triggers", mcp.Description("Comma-separated triggers: pre_show, post_show, interval")),
				mcp.WithNumber("interval_ms", mcp.Description("Interval in milliseconds (for interval trigger)")),
				mcp.WithString("systems", mcp.Description("Comma-separated systems to auto-capture (default: all)")),
				mcp.WithNumber("max_keep", mcp.Description("Max auto-snapshots to retain (default: 10)")),
				mcp.WithString("name_prefix", mcp.Description("Prefix for auto-snapshot names (default: auto)")),
			),
			Handler:             handleAutoSnapshotConfigure,
			Category:            "snapshots",
			Subcategory:         "automation",
			Tags:                []string{"snapshot", "auto", "schedule", "trigger"},
			UseCases:            []string{"Enable pre-show auto-snapshots", "Configure timed snapshots", "Set snapshot retention"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "snapshots",
		},
	}
}

func handleSnapshotCapture(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := tools.GetStringParam(req, "name")
	description := tools.GetStringParam(req, "description")
	systemsStr := tools.GetStringParam(req, "systems")
	tagsStr := tools.GetStringParam(req, "tags")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Parse systems
	var systems []string
	if systemsStr != "" {
		for _, s := range strings.Split(systemsStr, ",") {
			systems = append(systems, strings.TrimSpace(s))
		}
	}

	// Parse tags
	var tags []string
	if tagsStr != "" {
		for _, t := range strings.Split(tagsStr, ",") {
			tags = append(tags, strings.TrimSpace(t))
		}
	}

	snapshot, err := client.CaptureSnapshot(ctx, name, description, systems, tags)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Snapshot Captured\n\n")
	sb.WriteString(fmt.Sprintf("**ID:** `%s`\n", snapshot.ID))
	sb.WriteString(fmt.Sprintf("**Name:** %s\n", snapshot.Name))
	if snapshot.Description != "" {
		sb.WriteString(fmt.Sprintf("**Description:** %s\n", snapshot.Description))
	}
	sb.WriteString(fmt.Sprintf("**Created:** %s\n", snapshot.CreatedAt.Format("2006-01-02 15:04:05")))
	sb.WriteString("\n## Systems Captured\n\n")
	sb.WriteString("| System | Connected | Status |\n")
	sb.WriteString("|--------|-----------|--------|\n")

	for sys, state := range snapshot.Systems {
		connected := "No"
		if state.Connected {
			connected = "Yes"
		}
		status := "OK"
		if state.Error != "" {
			status = "Error: " + state.Error
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", sys, connected, status))
	}

	return tools.TextResult(sb.String()), nil
}

func handleSnapshotRecall(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	snapshotID := tools.GetStringParam(req, "snapshot_id")
	systemsStr := tools.GetStringParam(req, "systems")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Parse systems
	var systems []string
	if systemsStr != "" {
		for _, s := range strings.Split(systemsStr, ",") {
			systems = append(systems, strings.TrimSpace(s))
		}
	}

	// Get snapshot info first
	snapshot, err := client.GetSnapshot(ctx, snapshotID)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.RecallSnapshot(ctx, snapshotID, systems); err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Snapshot Recalled\n\n")
	sb.WriteString(fmt.Sprintf("**Snapshot:** %s (`%s`)\n", snapshot.Name, snapshot.ID[:12]))
	sb.WriteString(fmt.Sprintf("**Originally captured:** %s\n\n", snapshot.CreatedAt.Format("2006-01-02 15:04:05")))

	sb.WriteString("## Systems Restored\n\n")
	if len(systems) == 0 {
		for sys := range snapshot.Systems {
			sb.WriteString(fmt.Sprintf("- %s\n", sys))
		}
	} else {
		for _, sys := range systems {
			sb.WriteString(fmt.Sprintf("- %s\n", sys))
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleSnapshotList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	snapshotID := tools.GetStringParam(req, "snapshot_id")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// If specific snapshot requested
	if snapshotID != "" {
		snapshot, err := client.GetSnapshot(ctx, snapshotID)
		if err != nil {
			return tools.ErrorResult(err), nil
		}

		var sb strings.Builder
		sb.WriteString("# Snapshot Details\n\n")
		sb.WriteString(fmt.Sprintf("**ID:** `%s`\n", snapshot.ID))
		sb.WriteString(fmt.Sprintf("**Name:** %s\n", snapshot.Name))
		if snapshot.Description != "" {
			sb.WriteString(fmt.Sprintf("**Description:** %s\n", snapshot.Description))
		}
		sb.WriteString(fmt.Sprintf("**Created:** %s\n", snapshot.CreatedAt.Format("2006-01-02 15:04:05")))
		if len(snapshot.Tags) > 0 {
			sb.WriteString(fmt.Sprintf("**Tags:** %s\n", strings.Join(snapshot.Tags, ", ")))
		}

		sb.WriteString("\n## Systems\n\n")
		for sys, state := range snapshot.Systems {
			sb.WriteString(fmt.Sprintf("### %s\n", sys))
			sb.WriteString(fmt.Sprintf("- Connected: %t\n", state.Connected))
			sb.WriteString(fmt.Sprintf("- Captured: %s\n", state.CapturedAt.Format("15:04:05")))
			if state.Error != "" {
				sb.WriteString(fmt.Sprintf("- Error: %s\n", state.Error))
			}
			// Show key state values
			for key, val := range state.State {
				if key != "tracks" && key != "layers" {
					sb.WriteString(fmt.Sprintf("- %s: %v\n", key, val))
				}
			}
			sb.WriteString("\n")
		}

		return tools.TextResult(sb.String()), nil
	}

	// List all snapshots
	snapshots := client.ListSnapshots(ctx)

	var sb strings.Builder
	sb.WriteString("# Available Snapshots\n\n")

	if len(snapshots) == 0 {
		sb.WriteString("No snapshots found. Use `aftrs_snapshot_capture` to create one.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** snapshots:\n\n", len(snapshots)))
	sb.WriteString("| ID | Name | Created | Systems | Tags |\n")
	sb.WriteString("|----|------|---------|---------|------|\n")

	for _, snap := range snapshots {
		systems := make([]string, 0, len(snap.Systems))
		for sys := range snap.Systems {
			systems = append(systems, sys)
		}
		tags := "-"
		if len(snap.Tags) > 0 {
			tags = strings.Join(snap.Tags, ", ")
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
			snap.ID[:12], snap.Name, snap.CreatedAt.Format("01-02 15:04"),
			strings.Join(systems, ", "), tags))
	}

	return tools.TextResult(sb.String()), nil
}

func handleSnapshotDiff(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	snapshot1 := tools.GetStringParam(req, "snapshot1")
	snapshot2 := tools.GetStringParam(req, "snapshot2")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	diff, err := client.DiffSnapshots(ctx, snapshot1, snapshot2)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Snapshot Comparison\n\n")
	sb.WriteString(fmt.Sprintf("**Snapshot 1:** `%s`\n", snapshot1[:min(12, len(snapshot1))]))
	sb.WriteString(fmt.Sprintf("**Snapshot 2:** `%s`\n", snapshot2[:min(12, len(snapshot2))]))
	sb.WriteString(fmt.Sprintf("**Summary:** %s\n\n", diff.Summary))

	if len(diff.Differences) == 0 {
		sb.WriteString("No differences found.\n")
		return tools.TextResult(sb.String()), nil
	}

	for sys, sysDiff := range diff.Differences {
		sb.WriteString(fmt.Sprintf("## %s\n\n", sys))

		if len(sysDiff.Added) > 0 {
			sb.WriteString("**Added:**\n")
			for key, val := range sysDiff.Added {
				sb.WriteString(fmt.Sprintf("- %s: %v\n", key, val))
			}
			sb.WriteString("\n")
		}

		if len(sysDiff.Removed) > 0 {
			sb.WriteString("**Removed:**\n")
			for key, val := range sysDiff.Removed {
				sb.WriteString(fmt.Sprintf("- %s: %v\n", key, val))
			}
			sb.WriteString("\n")
		}

		if len(sysDiff.Changed) > 0 {
			sb.WriteString("**Changed:**\n")
			for _, change := range sysDiff.Changed {
				sb.WriteString(fmt.Sprintf("- %s: `%v` → `%v`\n", change.Field, change.OldValue, change.NewValue))
			}
			sb.WriteString("\n")
		}
	}

	return tools.TextResult(sb.String()), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
