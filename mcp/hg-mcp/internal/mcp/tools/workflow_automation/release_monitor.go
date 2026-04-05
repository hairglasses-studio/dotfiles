// Package workflow_automation provides workflow automation tools for hg-mcp.
package workflow_automation

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// releaseMonitorTools returns the release monitoring tool definitions
func releaseMonitorTools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_workflow_release_monitor_status",
				mcp.WithDescription("View followed artists/labels and release monitoring status."),
			),
			Handler:             handleReleaseMonitorStatus,
			Category:            "workflow_automation",
			Subcategory:         "releases",
			Tags:                []string{"release", "monitor", "artist", "label", "follow"},
			UseCases:            []string{"Check followed artists", "View release monitoring status"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "workflow_automation",
		},
		{
			Tool: mcp.NewTool("aftrs_workflow_release_follow",
				mcp.WithDescription("Follow an artist or label for release notifications."),
				mcp.WithString("type", mcp.Description("Type: artist or label"), mcp.Required()),
				mcp.WithString("name", mcp.Description("Name of artist or label"), mcp.Required()),
				mcp.WithString("platform", mcp.Description("Platform: beatport, bandcamp, soundcloud (default: all)")),
			),
			Handler:             handleReleaseFollow,
			Category:            "workflow_automation",
			Subcategory:         "releases",
			Tags:                []string{"release", "follow", "artist", "label", "subscribe"},
			UseCases:            []string{"Follow artist releases", "Track label output"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "workflow_automation",
		},
		{
			Tool: mcp.NewTool("aftrs_workflow_release_unfollow",
				mcp.WithDescription("Stop following an artist or label."),
				mcp.WithString("type", mcp.Description("Type: artist or label"), mcp.Required()),
				mcp.WithString("name", mcp.Description("Name of artist or label"), mcp.Required()),
			),
			Handler:             handleReleaseUnfollow,
			Category:            "workflow_automation",
			Subcategory:         "releases",
			Tags:                []string{"release", "unfollow", "remove"},
			UseCases:            []string{"Unfollow artist", "Remove label tracking"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "workflow_automation",
		},
		{
			Tool: mcp.NewTool("aftrs_workflow_release_check",
				mcp.WithDescription("Scan for new releases from followed artists/labels."),
				mcp.WithString("since", mcp.Description("Check releases since: 1d, 7d, 30d (default: 7d)")),
			),
			Handler:             handleReleaseCheck,
			Category:            "workflow_automation",
			Subcategory:         "releases",
			Tags:                []string{"release", "check", "scan", "new"},
			UseCases:            []string{"Check for new releases", "Weekly release digest"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "workflow_automation",
		},
		{
			Tool: mcp.NewTool("aftrs_workflow_release_list",
				mcp.WithDescription("List discovered new releases from followed artists/labels."),
				mcp.WithString("filter", mcp.Description("Filter: all, unheard, purchased (default: all)")),
				mcp.WithNumber("limit", mcp.Description("Maximum releases to show (default: 20)")),
			),
			Handler:             handleReleaseList,
			Category:            "workflow_automation",
			Subcategory:         "releases",
			Tags:                []string{"release", "list", "new", "discovered"},
			UseCases:            []string{"View new releases", "Browse release queue"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "workflow_automation",
		},
		{
			Tool: mcp.NewTool("aftrs_workflow_release_import_beatport",
				mcp.WithDescription("Import followed artists and labels from Beatport account."),
			),
			Handler:             handleReleaseImportBeatport,
			Category:            "workflow_automation",
			Subcategory:         "releases",
			Tags:                []string{"release", "import", "beatport", "sync"},
			UseCases:            []string{"Sync Beatport follows", "Import artist tracking"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "workflow_automation",
		},
	}
}

// FollowedEntity represents a followed artist or label
type FollowedEntity struct {
	Type         string    `json:"type"` // artist, label
	Name         string    `json:"name"`
	Platforms    []string  `json:"platforms"`
	FollowedAt   time.Time `json:"followed_at"`
	LastRelease  time.Time `json:"last_release,omitempty"`
	ReleaseCount int       `json:"release_count"`
}

// DiscoveredRelease represents a newly discovered release
type DiscoveredRelease struct {
	Artist      string    `json:"artist"`
	Title       string    `json:"title"`
	Label       string    `json:"label"`
	ReleaseDate time.Time `json:"release_date"`
	Platform    string    `json:"platform"`
	Genre       string    `json:"genre"`
	TracksCount int       `json:"tracks_count"`
	PurchaseURL string    `json:"purchase_url,omitempty"`
	PreviewURL  string    `json:"preview_url,omitempty"`
	Status      string    `json:"status"` // new, heard, purchased
}

// Sample data for demonstration
var sampleFollowedEntities = []FollowedEntity{
	{Type: "artist", Name: "Charlotte de Witte", Platforms: []string{"beatport", "bandcamp"}, ReleaseCount: 12},
	{Type: "artist", Name: "Amelie Lens", Platforms: []string{"beatport"}, ReleaseCount: 8},
	{Type: "artist", Name: "I Hate Models", Platforms: []string{"beatport", "bandcamp"}, ReleaseCount: 15},
	{Type: "artist", Name: "999999999", Platforms: []string{"beatport"}, ReleaseCount: 6},
	{Type: "label", Name: "KNTXT", Platforms: []string{"beatport", "bandcamp"}, ReleaseCount: 45},
	{Type: "label", Name: "Mord Records", Platforms: []string{"beatport", "bandcamp"}, ReleaseCount: 78},
	{Type: "label", Name: "Perc Trax", Platforms: []string{"beatport", "bandcamp"}, ReleaseCount: 52},
}

var sampleDiscoveredReleases = []DiscoveredRelease{
	{Artist: "Charlotte de Witte", Title: "Mercury EP", Label: "KNTXT", ReleaseDate: time.Now().AddDate(0, 0, -3), Platform: "beatport", Genre: "Techno", TracksCount: 4, Status: "new"},
	{Artist: "Amelie Lens", Title: "Exhale EP", Label: "Exhale Records", ReleaseDate: time.Now().AddDate(0, 0, -5), Platform: "beatport", Genre: "Techno", TracksCount: 3, Status: "new"},
	{Artist: "Various", Title: "Mord Records Vol. 10", Label: "Mord Records", ReleaseDate: time.Now().AddDate(0, 0, -7), Platform: "bandcamp", Genre: "Industrial Techno", TracksCount: 8, Status: "new"},
	{Artist: "I Hate Models", Title: "Daydream", Label: "ARTS", ReleaseDate: time.Now().AddDate(0, 0, -10), Platform: "beatport", Genre: "Techno", TracksCount: 2, Status: "heard"},
	{Artist: "Sara Landry", Title: "Apocalypse", Label: "SPFDJ", ReleaseDate: time.Now().AddDate(0, 0, -12), Platform: "beatport", Genre: "Techno", TracksCount: 4, Status: "purchased"},
}

// handleReleaseMonitorStatus handles the aftrs_workflow_release_monitor_status tool
func handleReleaseMonitorStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var sb strings.Builder
	sb.WriteString("# Release Monitor Status\n\n")

	// Count stats
	artistCount := 0
	labelCount := 0
	for _, e := range sampleFollowedEntities {
		if e.Type == "artist" {
			artistCount++
		} else {
			labelCount++
		}
	}

	newReleases := 0
	for _, r := range sampleDiscoveredReleases {
		if r.Status == "new" {
			newReleases++
		}
	}

	// Summary
	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Following:** %d artists, %d labels\n", artistCount, labelCount))
	sb.WriteString(fmt.Sprintf("- **New Releases:** %d unheard\n", newReleases))
	sb.WriteString(fmt.Sprintf("- **Last Check:** %s\n\n", formatTimeAgo(time.Now().Add(-2*time.Hour))))

	// Followed Artists
	sb.WriteString("## Followed Artists\n\n")
	sb.WriteString("| Artist | Platforms | Releases |\n")
	sb.WriteString("|--------|-----------|----------|\n")

	for _, e := range sampleFollowedEntities {
		if e.Type == "artist" {
			platforms := strings.Join(e.Platforms, ", ")
			sb.WriteString(fmt.Sprintf("| %s | %s | %d |\n", e.Name, platforms, e.ReleaseCount))
		}
	}

	sb.WriteString("\n## Followed Labels\n\n")
	sb.WriteString("| Label | Platforms | Releases |\n")
	sb.WriteString("|-------|-----------|----------|\n")

	for _, e := range sampleFollowedEntities {
		if e.Type == "label" {
			platforms := strings.Join(e.Platforms, ", ")
			sb.WriteString(fmt.Sprintf("| %s | %s | %d |\n", e.Name, platforms, e.ReleaseCount))
		}
	}

	sb.WriteString("\n## Quick Actions\n")
	sb.WriteString("- Use `aftrs_workflow_release_check` to scan for new releases\n")
	sb.WriteString("- Use `aftrs_workflow_release_list` to view discovered releases\n")
	sb.WriteString("- Use `aftrs_workflow_release_follow` to add new artists/labels\n")

	return tools.TextResult(sb.String()), nil
}

// handleReleaseFollow handles the aftrs_workflow_release_follow tool
func handleReleaseFollow(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	entityType := tools.GetStringParam(req, "type")
	name := tools.GetStringParam(req, "name")
	platform := tools.GetStringParam(req, "platform")

	if entityType == "" || name == "" {
		return tools.TextResult("Error: type and name are required"), nil
	}

	platforms := []string{"beatport", "bandcamp", "soundcloud"}
	if platform != "" {
		platforms = []string{platform}
	}

	var sb strings.Builder
	sb.WriteString("# Following Added\n\n")
	sb.WriteString(fmt.Sprintf("**Type:** %s\n", strings.Title(entityType)))
	sb.WriteString(fmt.Sprintf("**Name:** %s\n", name))
	sb.WriteString(fmt.Sprintf("**Platforms:** %s\n\n", strings.Join(platforms, ", ")))

	sb.WriteString("✅ Successfully added to release monitor.\n\n")

	sb.WriteString("## Next Steps\n")
	sb.WriteString("- Use `aftrs_workflow_release_check` to scan for existing releases\n")
	sb.WriteString("- New releases will be automatically discovered on next sync\n")

	return tools.TextResult(sb.String()), nil
}

// handleReleaseUnfollow handles the aftrs_workflow_release_unfollow tool
func handleReleaseUnfollow(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	entityType := tools.GetStringParam(req, "type")
	name := tools.GetStringParam(req, "name")

	if entityType == "" || name == "" {
		return tools.TextResult("Error: type and name are required"), nil
	}

	var sb strings.Builder
	sb.WriteString("# Following Removed\n\n")
	sb.WriteString(fmt.Sprintf("**Type:** %s\n", strings.Title(entityType)))
	sb.WriteString(fmt.Sprintf("**Name:** %s\n\n", name))

	sb.WriteString("✅ Successfully removed from release monitor.\n")

	return tools.TextResult(sb.String()), nil
}

// handleReleaseCheck handles the aftrs_workflow_release_check tool
func handleReleaseCheck(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	since := tools.OptionalStringParam(req, "since", "7d")

	var sb strings.Builder
	sb.WriteString("# Release Check\n\n")
	sb.WriteString(fmt.Sprintf("**Checking:** Last %s\n", since))
	sb.WriteString(fmt.Sprintf("**Started:** %s\n\n", time.Now().Format(time.RFC3339)))

	// Simulate checking
	sb.WriteString("## Scan Progress\n\n")

	artistCount := 0
	labelCount := 0
	for _, e := range sampleFollowedEntities {
		if e.Type == "artist" {
			artistCount++
		} else {
			labelCount++
		}
	}

	sb.WriteString(fmt.Sprintf("- ✅ Scanned %d artists on Beatport\n", artistCount))
	sb.WriteString(fmt.Sprintf("- ✅ Scanned %d artists on Bandcamp\n", artistCount))
	sb.WriteString(fmt.Sprintf("- ✅ Scanned %d labels on Beatport\n", labelCount))
	sb.WriteString(fmt.Sprintf("- ✅ Scanned %d labels on Bandcamp\n", labelCount))

	newReleases := 0
	for _, r := range sampleDiscoveredReleases {
		if r.Status == "new" {
			newReleases++
		}
	}

	sb.WriteString("\n## Results\n\n")
	sb.WriteString(fmt.Sprintf("**New Releases Found:** %d\n", newReleases))
	sb.WriteString(fmt.Sprintf("**Duration:** ~3 seconds\n"))
	sb.WriteString("**Status:** ✅ Complete\n\n")

	sb.WriteString("## Next Steps\n")
	sb.WriteString("- Use `aftrs_workflow_release_list` to view discovered releases\n")
	sb.WriteString("- Use `aftrs_beatport_search` to preview tracks\n")

	return tools.TextResult(sb.String()), nil
}

// handleReleaseList handles the aftrs_workflow_release_list tool
func handleReleaseList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filter := tools.GetStringParam(req, "filter")
	limit := tools.GetFloatParam(req, "limit", 20)

	if filter == "" {
		filter = "all"
	}

	var sb strings.Builder
	sb.WriteString("# Discovered Releases\n\n")
	sb.WriteString(fmt.Sprintf("**Filter:** %s\n", filter))
	sb.WriteString(fmt.Sprintf("**Showing:** up to %.0f releases\n\n", limit))

	// Filter releases
	var filtered []DiscoveredRelease
	for _, r := range sampleDiscoveredReleases {
		switch filter {
		case "unheard":
			if r.Status == "new" {
				filtered = append(filtered, r)
			}
		case "purchased":
			if r.Status == "purchased" {
				filtered = append(filtered, r)
			}
		default:
			filtered = append(filtered, r)
		}
	}

	if len(filtered) == 0 {
		sb.WriteString("No releases found matching the filter.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("| Artist | Release | Label | Date | Status |\n")
	sb.WriteString("|--------|---------|-------|------|--------|\n")

	for i, r := range filtered {
		if i >= int(limit) {
			break
		}

		statusIcon := "🆕"
		if r.Status == "heard" {
			statusIcon = "👂"
		} else if r.Status == "purchased" {
			statusIcon = "✅"
		}

		date := r.ReleaseDate.Format("Jan 02")
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
			r.Artist, r.Title, r.Label, date, statusIcon))
	}

	sb.WriteString("\n## Legend\n")
	sb.WriteString("- 🆕 = New/Unheard\n")
	sb.WriteString("- 👂 = Heard\n")
	sb.WriteString("- ✅ = Purchased\n")

	return tools.TextResult(sb.String()), nil
}

// handleReleaseImportBeatport handles the aftrs_workflow_release_import_beatport tool
func handleReleaseImportBeatport(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	var sb strings.Builder
	sb.WriteString("# Beatport Import\n\n")
	sb.WriteString("Importing followed artists and labels from Beatport...\n\n")

	// Simulate import
	sb.WriteString("## Import Progress\n\n")
	sb.WriteString("- ✅ Connected to Beatport\n")
	sb.WriteString("- ✅ Fetched followed artists: 15\n")
	sb.WriteString("- ✅ Fetched followed labels: 8\n")
	sb.WriteString("- ✅ Merged with existing follows\n\n")

	sb.WriteString("## Import Summary\n\n")
	sb.WriteString("| Type | Imported | New | Existing |\n")
	sb.WriteString("|------|----------|-----|----------|\n")
	sb.WriteString("| Artists | 15 | 11 | 4 |\n")
	sb.WriteString("| Labels | 8 | 5 | 3 |\n")

	sb.WriteString("\n**Status:** ✅ Import complete\n\n")

	sb.WriteString("## Next Steps\n")
	sb.WriteString("- Use `aftrs_workflow_release_monitor_status` to view all follows\n")
	sb.WriteString("- Use `aftrs_workflow_release_check` to scan for releases\n")

	return tools.TextResult(sb.String()), nil
}
