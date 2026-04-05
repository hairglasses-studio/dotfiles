// Package lighting provides DMX and lighting control tools for hg-mcp.
package lighting

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewLightingClient)

// Module implements the ToolModule interface for lighting control
type Module struct{}

func (m *Module) Name() string {
	return "lighting"
}

func (m *Module) Description() string {
	return "DMX/ArtNet lighting control and fixture management"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_dmx_status",
				mcp.WithDescription("Get DMX universe status and connection info."),
			),
			Handler:             handleDMXStatus,
			Category:            "lighting",
			Subcategory:         "dmx",
			Tags:                []string{"dmx", "artnet", "status", "universe"},
			UseCases:            []string{"Check DMX connection", "Verify universe status"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "lighting",
		},
		{
			Tool: mcp.NewTool("aftrs_dmx_channels",
				mcp.WithDescription("Get or set DMX channel values."),
				mcp.WithNumber("start_channel",
					mcp.Description("Starting channel (1-512, default: 1)"),
				),
				mcp.WithNumber("count",
					mcp.Description("Number of channels to read (default: 16)"),
				),
				mcp.WithString("values",
					mcp.Description("Comma-separated values to set (e.g., '255,128,0')"),
				),
			),
			Handler:             handleDMXChannels,
			Category:            "lighting",
			Subcategory:         "dmx",
			Tags:                []string{"dmx", "channels", "control"},
			UseCases:            []string{"Read channel values", "Set channel values"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "lighting",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_fixture_list",
				mcp.WithDescription("List configured lighting fixtures."),
			),
			Handler:             handleFixtureList,
			Category:            "lighting",
			Subcategory:         "fixtures",
			Tags:                []string{"fixtures", "lights", "list"},
			UseCases:            []string{"Browse fixtures", "Check fixture config"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "lighting",
		},
		{
			Tool: mcp.NewTool("aftrs_fixture_control",
				mcp.WithDescription("Control a lighting fixture by name."),
				mcp.WithString("fixture",
					mcp.Required(),
					mcp.Description("Fixture name"),
				),
				mcp.WithNumber("dimmer",
					mcp.Description("Dimmer value (0-255)"),
				),
				mcp.WithString("color",
					mcp.Description("Color as hex (e.g., 'FF0000' for red) or name ('red', 'blue', etc.)"),
				),
			),
			Handler:             handleFixtureControl,
			Category:            "lighting",
			Subcategory:         "fixtures",
			Tags:                []string{"fixtures", "control", "color", "dimmer"},
			UseCases:            []string{"Control fixtures", "Set colors"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "lighting",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_scene_list",
				mcp.WithDescription("List saved lighting scenes."),
			),
			Handler:             handleSceneList,
			Category:            "lighting",
			Subcategory:         "scenes",
			Tags:                []string{"scenes", "presets", "list"},
			UseCases:            []string{"Browse scenes", "Find saved looks"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "lighting",
		},
		{
			Tool: mcp.NewTool("aftrs_scene_recall",
				mcp.WithDescription("Recall a saved lighting scene."),
				mcp.WithString("scene",
					mcp.Required(),
					mcp.Description("Scene name to recall"),
				),
			),
			Handler:             handleSceneRecall,
			Category:            "lighting",
			Subcategory:         "scenes",
			Tags:                []string{"scenes", "recall", "presets"},
			UseCases:            []string{"Load saved scene", "Quick lighting change"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "lighting",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_lighting_health",
				mcp.WithDescription("Get lighting system health score."),
			),
			Handler:             handleLightingHealth,
			Category:            "lighting",
			Subcategory:         "health",
			Tags:                []string{"health", "status", "diagnostics"},
			UseCases:            []string{"Check system health", "Pre-show verification"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "lighting",
		},
		{
			Tool: mcp.NewTool("aftrs_artnet_nodes",
				mcp.WithDescription("Discover ArtNet nodes on the network."),
			),
			Handler:             handleArtNetNodes,
			Category:            "lighting",
			Subcategory:         "artnet",
			Tags:                []string{"artnet", "nodes", "discovery", "network"},
			UseCases:            []string{"Find ArtNet devices", "Network diagnostics"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "lighting",
		},
		// Additional tools for expanded coverage
		{
			Tool: mcp.NewTool("aftrs_dmx_blackout",
				mcp.WithDescription("Blackout all lights (all channels to zero)."),
			),
			Handler:             handleBlackout,
			Category:            "lighting",
			Subcategory:         "dmx",
			Tags:                []string{"dmx", "blackout", "all", "off"},
			UseCases:            []string{"Emergency blackout", "End of show"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "lighting",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_dmx_full",
				mcp.WithDescription("Full up all lights (all channels to max)."),
			),
			Handler:             handleFullUp,
			Category:            "lighting",
			Subcategory:         "dmx",
			Tags:                []string{"dmx", "full", "all", "on"},
			UseCases:            []string{"Work lights", "Testing"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "lighting",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_fixture_dimmer",
				mcp.WithDescription("Quick dimmer control for a fixture."),
				mcp.WithString("fixture", mcp.Required(), mcp.Description("Fixture name")),
				mcp.WithNumber("level", mcp.Required(), mcp.Description("Dimmer level (0-100%)")),
			),
			Handler:             handleFixtureDimmer,
			Category:            "lighting",
			Subcategory:         "fixtures",
			Tags:                []string{"fixture", "dimmer", "level", "brightness"},
			UseCases:            []string{"Adjust brightness", "Fade fixtures"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "lighting",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_fixture_color",
				mcp.WithDescription("Quick color control for a fixture."),
				mcp.WithString("fixture", mcp.Required(), mcp.Description("Fixture name")),
				mcp.WithString("color", mcp.Required(), mcp.Description("Color (hex like 'FF0000' or name like 'red')")),
			),
			Handler:             handleFixtureColor,
			Category:            "lighting",
			Subcategory:         "fixtures",
			Tags:                []string{"fixture", "color", "rgb"},
			UseCases:            []string{"Set fixture color", "Quick color changes"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "lighting",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_scene_save",
				mcp.WithDescription("Save current lighting state as a scene."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Scene name")),
				mcp.WithString("description", mcp.Description("Scene description")),
			),
			Handler:             handleSceneSave,
			Category:            "lighting",
			Subcategory:         "scenes",
			Tags:                []string{"scene", "save", "preset", "store"},
			UseCases:            []string{"Save current look", "Create presets"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "lighting",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_scene_fade",
				mcp.WithDescription("Crossfade to a scene over time."),
				mcp.WithString("scene", mcp.Required(), mcp.Description("Scene name")),
				mcp.WithNumber("duration", mcp.Description("Fade duration in seconds (default 2)")),
			),
			Handler:             handleSceneFade,
			Category:            "lighting",
			Subcategory:         "scenes",
			Tags:                []string{"scene", "fade", "crossfade", "transition"},
			UseCases:            []string{"Smooth transitions", "Timed scene changes"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "lighting",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_group_list",
				mcp.WithDescription("List fixture groups."),
			),
			Handler:             handleGroupList,
			Category:            "lighting",
			Subcategory:         "groups",
			Tags:                []string{"groups", "fixtures", "list"},
			UseCases:            []string{"View fixture groups", "Check groupings"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "lighting",
		},
		{
			Tool: mcp.NewTool("aftrs_group_control",
				mcp.WithDescription("Control all fixtures in a group."),
				mcp.WithString("group", mcp.Required(), mcp.Description("Group name")),
				mcp.WithNumber("dimmer", mcp.Description("Dimmer level (0-100%)")),
				mcp.WithString("color", mcp.Description("Color (hex or name)")),
			),
			Handler:             handleGroupControl,
			Category:            "lighting",
			Subcategory:         "groups",
			Tags:                []string{"groups", "control", "fixtures"},
			UseCases:            []string{"Control multiple fixtures", "Group dimming"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "lighting",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_chase_list",
				mcp.WithDescription("List available chases/sequences."),
			),
			Handler:             handleChaseList,
			Category:            "lighting",
			Subcategory:         "chases",
			Tags:                []string{"chase", "sequence", "list"},
			UseCases:            []string{"View chases", "Check sequences"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "lighting",
		},
		{
			Tool: mcp.NewTool("aftrs_chase_control",
				mcp.WithDescription("Control chase playback."),
				mcp.WithString("chase", mcp.Required(), mcp.Description("Chase name")),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: start, stop, or tap")),
				mcp.WithNumber("bpm", mcp.Description("BPM for chase speed")),
			),
			Handler:             handleChaseControl,
			Category:            "lighting",
			Subcategory:         "chases",
			Tags:                []string{"chase", "sequence", "control", "playback"},
			UseCases:            []string{"Start/stop chases", "Adjust tempo"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "lighting",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_universe_select",
				mcp.WithDescription("Select active DMX universe."),
				mcp.WithNumber("universe", mcp.Required(), mcp.Description("Universe number (0-32767)")),
			),
			Handler:             handleUniverseSelect,
			Category:            "lighting",
			Subcategory:         "dmx",
			Tags:                []string{"universe", "dmx", "select"},
			UseCases:            []string{"Switch universes", "Multi-universe control"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "lighting",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_artnet_status",
				mcp.WithDescription("Get detailed ArtNet protocol status."),
			),
			Handler:             handleArtNetStatus,
			Category:            "lighting",
			Subcategory:         "artnet",
			Tags:                []string{"artnet", "status", "network", "protocol"},
			UseCases:            []string{"Check ArtNet connection", "Network diagnostics"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "lighting",
		},
		// Patch management tools
		{
			Tool: mcp.NewTool("aftrs_patch_list",
				mcp.WithDescription("List fixture patch assignments (fixture-to-channel mappings)."),
			),
			Handler:             handlePatchList,
			Category:            "lighting",
			Subcategory:         "patch",
			Tags:                []string{"patch", "fixtures", "channels", "mapping"},
			UseCases:            []string{"View patch assignments", "Check channel allocation"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "lighting",
		},
		{
			Tool: mcp.NewTool("aftrs_patch_fixture",
				mcp.WithDescription("Patch a fixture to a DMX address."),
				mcp.WithString("fixture", mcp.Required(), mcp.Description("Fixture type or name")),
				mcp.WithNumber("universe", mcp.Required(), mcp.Description("DMX universe (0-32767)")),
				mcp.WithNumber("address", mcp.Required(), mcp.Description("Start DMX address (1-512)")),
				mcp.WithString("name", mcp.Description("Custom fixture name")),
			),
			Handler:             handlePatchFixture,
			Category:            "lighting",
			Subcategory:         "patch",
			Tags:                []string{"patch", "fixture", "assign", "address"},
			UseCases:            []string{"Add fixture to patch", "Assign DMX address"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "lighting",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_patch_unpatch",
				mcp.WithDescription("Remove a fixture from the patch."),
				mcp.WithString("fixture", mcp.Required(), mcp.Description("Fixture name to unpatch")),
			),
			Handler:             handlePatchUnpatch,
			Category:            "lighting",
			Subcategory:         "patch",
			Tags:                []string{"patch", "unpatch", "remove", "delete"},
			UseCases:            []string{"Remove fixture from patch", "Clear address"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "lighting",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_patch_export",
				mcp.WithDescription("Export patch to JSON or CSV format."),
				mcp.WithString("format", mcp.Description("Export format: json (default) or csv")),
			),
			Handler:             handlePatchExport,
			Category:            "lighting",
			Subcategory:         "patch",
			Tags:                []string{"patch", "export", "backup", "json", "csv"},
			UseCases:            []string{"Backup patch sheet", "Export for documentation"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "lighting",
		},
		{
			Tool: mcp.NewTool("aftrs_dmx_batch_set",
				mcp.WithDescription("Set multiple DMX channel ranges in a single call. Accepts a JSON array of operations."),
				mcp.WithString("operations",
					mcp.Required(),
					mcp.Description("JSON array of operations: [{\"start_channel\":1,\"values\":\"255,128,0\"},{\"start_channel\":10,\"values\":\"200,100\"}]"),
				),
			),
			Handler:             handleDMXBatchSet,
			Category:            "lighting",
			Subcategory:         "dmx",
			Tags:                []string{"dmx", "batch", "channels", "control"},
			UseCases:            []string{"Set multiple channel ranges at once", "Batch fixture updates"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "lighting",
			IsWrite:             true,
		},
	}
}

// handleDMXStatus handles the aftrs_dmx_status tool
func handleDMXStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	status, err := client.GetDMXStatus(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# DMX Universe Status\n\n")

	statusEmoji := "🔴"
	if status.Active {
		statusEmoji = "🟢"
	}

	sb.WriteString(fmt.Sprintf("**Status:** %s %s\n", statusEmoji, map[bool]string{true: "Active", false: "Inactive"}[status.Active]))
	sb.WriteString(fmt.Sprintf("**Universe:** %d\n", status.Universe))
	sb.WriteString(fmt.Sprintf("**Channels:** %d\n", status.Channels))

	if status.Source != "" {
		sb.WriteString(fmt.Sprintf("**Source:** %s\n", status.Source))
	}

	sb.WriteString(fmt.Sprintf("\n**ArtNet Host:** %s\n", client.ArtNetHost()))

	if !status.Active {
		sb.WriteString("\n## Setup Required\n\n")
		sb.WriteString("Configure ArtNet connection:\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export ARTNET_HOST=192.168.1.100\n")
		sb.WriteString("```\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleDMXChannels handles the aftrs_dmx_channels tool
func handleDMXChannels(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startChannel := tools.GetIntParam(req, "start_channel", 1)
	count := tools.GetIntParam(req, "count", 16)
	valuesStr := tools.GetStringParam(req, "values")

	if startChannel < 1 {
		startChannel = 1
	}
	if startChannel > 512 {
		startChannel = 512
	}
	if count < 1 {
		count = 1
	}
	if count > 512-startChannel+1 {
		count = 512 - startChannel + 1
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder

	// If values provided, set them
	if valuesStr != "" {
		var values []int
		for _, v := range strings.Split(valuesStr, ",") {
			var val int
			fmt.Sscanf(strings.TrimSpace(v), "%d", &val)
			if val < 0 {
				val = 0
			}
			if val > 255 {
				val = 255
			}
			values = append(values, val)
		}

		if err := client.SetDMXChannels(ctx, startChannel, values); err != nil {
			return tools.ErrorResult(err), nil
		}

		sb.WriteString("# DMX Channels Set\n\n")
		sb.WriteString(fmt.Sprintf("Set %d channels starting at %d:\n\n", len(values), startChannel))
		for i, v := range values {
			sb.WriteString(fmt.Sprintf("- Channel %d: %d\n", startChannel+i, v))
		}
		return tools.TextResult(sb.String()), nil
	}

	// Otherwise, read channels
	channels, err := client.GetDMXChannels(ctx, startChannel, count)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	sb.WriteString("# DMX Channel Values\n\n")
	sb.WriteString(fmt.Sprintf("Channels %d-%d:\n\n", startChannel, startChannel+count-1))
	sb.WriteString("| Channel | Value |\n")
	sb.WriteString("|---------|-------|\n")

	for i, v := range channels {
		sb.WriteString(fmt.Sprintf("| %d | %d |\n", startChannel+i, v))
	}

	return tools.TextResult(sb.String()), nil
}

// handleFixtureList handles the aftrs_fixture_list tool
func handleFixtureList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	fixtures, err := client.ListFixtures(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Lighting Fixtures\n\n")

	if len(fixtures) == 0 {
		sb.WriteString("No fixtures configured.\n\n")
		sb.WriteString("## Setup\n\n")
		sb.WriteString("Configure fixtures via environment variable:\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export LIGHTING_CONFIG=/path/to/fixtures.json\n")
		sb.WriteString("```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** fixtures:\n\n", len(fixtures)))
	sb.WriteString("| Name | Type | Channel | Universe |\n")
	sb.WriteString("|------|------|---------|----------|\n")

	for _, f := range fixtures {
		sb.WriteString(fmt.Sprintf("| %s | %s | %d-%d | %d |\n",
			f.Name, f.Type, f.StartChannel, f.StartChannel+f.NumChannels-1, f.Universe))
	}

	return tools.TextResult(sb.String()), nil
}

// handleFixtureControl handles the aftrs_fixture_control tool
func handleFixtureControl(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fixture, errResult := tools.RequireStringParam(req, "fixture")
	if errResult != nil {
		return errResult, nil
	}

	dimmer := tools.GetIntParam(req, "dimmer", -1)
	color := tools.GetStringParam(req, "color")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	values := make(map[string]int)
	if dimmer >= 0 {
		if dimmer > 255 {
			dimmer = 255
		}
		values["dimmer"] = dimmer
	}
	if color != "" {
		// Parse color
		r, g, b := parseColor(color)
		values["red"] = r
		values["green"] = g
		values["blue"] = b
	}

	if err := client.ControlFixture(ctx, fixture, values); err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Fixture Control: %s\n\n", fixture))
	sb.WriteString("**Updated values:**\n\n")
	for k, v := range values {
		sb.WriteString(fmt.Sprintf("- %s: %d\n", k, v))
	}

	return tools.TextResult(sb.String()), nil
}

// parseColor parses a color string to RGB values
func parseColor(color string) (r, g, b int) {
	color = strings.ToLower(color)

	// Named colors
	namedColors := map[string][3]int{
		"red":     {255, 0, 0},
		"green":   {0, 255, 0},
		"blue":    {0, 0, 255},
		"white":   {255, 255, 255},
		"yellow":  {255, 255, 0},
		"cyan":    {0, 255, 255},
		"magenta": {255, 0, 255},
		"orange":  {255, 165, 0},
		"purple":  {128, 0, 128},
		"pink":    {255, 192, 203},
	}

	if rgb, ok := namedColors[color]; ok {
		return rgb[0], rgb[1], rgb[2]
	}

	// Try hex
	color = strings.TrimPrefix(color, "#")
	if len(color) == 6 {
		fmt.Sscanf(color, "%02x%02x%02x", &r, &g, &b)
	}

	return r, g, b
}

// handleSceneList handles the aftrs_scene_list tool
func handleSceneList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	scenes, err := client.ListScenes(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Lighting Scenes\n\n")

	if len(scenes) == 0 {
		sb.WriteString("No scenes saved.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** scenes:\n\n", len(scenes)))
	sb.WriteString("| Name | Description |\n")
	sb.WriteString("|------|-------------|\n")

	for _, s := range scenes {
		sb.WriteString(fmt.Sprintf("| %s | %s |\n", s.Name, s.Description))
	}

	return tools.TextResult(sb.String()), nil
}

// handleSceneRecall handles the aftrs_scene_recall tool
func handleSceneRecall(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	scene, errResult := tools.RequireStringParam(req, "scene")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.RecallScene(ctx, scene); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Scene '%s' recalled successfully.", scene)), nil
}

// handleLightingHealth handles the aftrs_lighting_health tool
func handleLightingHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Lighting System Health\n\n")

	statusEmoji := "✅"
	if health.Status == "degraded" {
		statusEmoji = "⚠️"
	} else if health.Status == "critical" {
		statusEmoji = "❌"
	}

	sb.WriteString(fmt.Sprintf("**Health Score:** %d/100 %s\n", health.Score, statusEmoji))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n\n", health.Status))

	sb.WriteString("## System Info\n\n")
	sb.WriteString(fmt.Sprintf("| Metric | Value |\n"))
	sb.WriteString(fmt.Sprintf("|--------|-------|\n"))
	sb.WriteString(fmt.Sprintf("| Universes | %d |\n", health.UniverseCount))
	sb.WriteString(fmt.Sprintf("| Fixtures | %d |\n", health.FixtureCount))
	sb.WriteString(fmt.Sprintf("| ArtNet Nodes | %d |\n", health.NodesOnline))

	if len(health.Issues) > 0 {
		sb.WriteString("\n## Issues\n\n")
		for _, issue := range health.Issues {
			sb.WriteString(fmt.Sprintf("- ⚠️ %s\n", issue))
		}
	}

	if len(health.Recommendations) > 0 {
		sb.WriteString("\n## Recommendations\n\n")
		for _, rec := range health.Recommendations {
			sb.WriteString(fmt.Sprintf("- 💡 %s\n", rec))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleArtNetNodes handles the aftrs_artnet_nodes tool
func handleArtNetNodes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	nodes, err := client.DiscoverArtNetNodes(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# ArtNet Node Discovery\n\n")

	if len(nodes) == 0 {
		sb.WriteString("No ArtNet nodes found on the network.\n\n")
		sb.WriteString("## Troubleshooting\n\n")
		sb.WriteString("- Verify ArtNet nodes are powered on\n")
		sb.WriteString("- Check network connectivity\n")
		sb.WriteString("- Ensure nodes are on the same subnet\n")
		return tools.TextResult(sb.String()), nil
	}

	online := 0
	for _, n := range nodes {
		if n.Online {
			online++
		}
	}

	sb.WriteString(fmt.Sprintf("Found **%d** nodes (%d online):\n\n", len(nodes), online))
	sb.WriteString("| Name | IP | Universe | Status |\n")
	sb.WriteString("|------|----|-----------|---------|\n")

	for _, node := range nodes {
		status := "🔴 Offline"
		if node.Online {
			status = "🟢 Online"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %d | %s |\n", node.Name, node.IP, node.Universe, status))
	}

	return tools.TextResult(sb.String()), nil
}

// handleBlackout handles the aftrs_dmx_blackout tool
func handleBlackout(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.Blackout(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult("Blackout - all channels set to 0"), nil
}

// handleFullUp handles the aftrs_dmx_full tool
func handleFullUp(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.FullUp(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult("Full up - all channels set to 255"), nil
}

// handleFixtureDimmer handles the aftrs_fixture_dimmer tool
func handleFixtureDimmer(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fixture := tools.GetStringParam(req, "fixture")
	level := tools.GetIntParam(req, "level", -1)

	if fixture == "" || level < 0 {
		return tools.ErrorResult(fmt.Errorf("fixture and level are required")), nil
	}

	// Convert 0-100% to 0-255
	dmxLevel := level * 255 / 100
	if dmxLevel > 255 {
		dmxLevel = 255
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SetFixtureDimmer(ctx, fixture, dmxLevel)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Set %s dimmer to %d%%", fixture, level)), nil
}

// handleFixtureColor handles the aftrs_fixture_color tool
func handleFixtureColor(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fixture := tools.GetStringParam(req, "fixture")
	color := tools.GetStringParam(req, "color")

	if fixture == "" || color == "" {
		return tools.ErrorResult(fmt.Errorf("fixture and color are required")), nil
	}

	r, g, b := parseColor(color)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SetFixtureColor(ctx, fixture, r, g, b)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Set %s color to RGB(%d, %d, %d)", fixture, r, g, b)), nil
}

// handleSceneSave handles the aftrs_scene_save tool
func handleSceneSave(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := tools.GetStringParam(req, "name")
	description := tools.GetStringParam(req, "description")

	if name == "" {
		return tools.ErrorResult(fmt.Errorf("scene name is required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SaveScene(ctx, name, description)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Scene '%s' saved", name)), nil
}

// handleSceneFade handles the aftrs_scene_fade tool
func handleSceneFade(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	scene, errResult := tools.RequireStringParam(req, "scene")
	if errResult != nil {
		return errResult, nil
	}
	duration := tools.GetIntParam(req, "duration", 2)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.FadeToScene(ctx, scene, duration*1000)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Fading to '%s' over %d seconds", scene, duration)), nil
}

// handleGroupList handles the aftrs_group_list tool
func handleGroupList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	groups, err := client.ListGroups(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Fixture Groups\n\n")

	if len(groups) == 0 {
		sb.WriteString("No fixture groups configured.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** groups:\n\n", len(groups)))
	sb.WriteString("| Name | Fixtures |\n")
	sb.WriteString("|------|----------|\n")

	for _, g := range groups {
		sb.WriteString(fmt.Sprintf("| %s | %d |\n", g.Name, len(g.Fixtures)))
	}

	return tools.TextResult(sb.String()), nil
}

// handleGroupControl handles the aftrs_group_control tool
func handleGroupControl(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	group, errResult := tools.RequireStringParam(req, "group")
	if errResult != nil {
		return errResult, nil
	}
	dimmer := tools.GetIntParam(req, "dimmer", -1)
	color := tools.GetStringParam(req, "color")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	values := make(map[string]int)
	if dimmer >= 0 {
		values["dimmer"] = dimmer * 255 / 100
	}
	if color != "" {
		r, g, b := parseColor(color)
		values["red"] = r
		values["green"] = g
		values["blue"] = b
	}

	err = client.ControlGroup(ctx, group, values)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Controlled group '%s':\n", group))
	for k, v := range values {
		sb.WriteString(fmt.Sprintf("- %s: %d\n", k, v))
	}

	return tools.TextResult(sb.String()), nil
}

// handleChaseList handles the aftrs_chase_list tool
func handleChaseList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	chases, err := client.ListChases(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Chases\n\n")

	if len(chases) == 0 {
		sb.WriteString("No chases configured.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** chases:\n\n", len(chases)))
	sb.WriteString("| Name | Steps | BPM | Status |\n")
	sb.WriteString("|------|-------|-----|--------|\n")

	for _, c := range chases {
		status := "Stopped"
		if c.Running {
			status = "Running"
		}
		sb.WriteString(fmt.Sprintf("| %s | %d | %.1f | %s |\n", c.Name, c.Steps, c.BPM, status))
	}

	return tools.TextResult(sb.String()), nil
}

// handleChaseControl handles the aftrs_chase_control tool
func handleChaseControl(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	chase := tools.GetStringParam(req, "chase")
	action := tools.GetStringParam(req, "action")
	bpm := float64(tools.GetIntParam(req, "bpm", 0))

	if chase == "" || action == "" {
		return tools.ErrorResult(fmt.Errorf("chase and action are required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "start":
		err = client.StartChase(ctx, chase)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Started chase '%s'", chase)), nil

	case "stop":
		err = client.StopChase(ctx, chase)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Stopped chase '%s'", chase)), nil

	case "tap":
		if bpm > 0 {
			err = client.SetChaseBPM(ctx, chase, bpm)
			if err != nil {
				return tools.ErrorResult(err), nil
			}
			return tools.TextResult(fmt.Sprintf("Set chase '%s' to %.1f BPM", chase, bpm)), nil
		}
		return tools.TextResult("Tap tempo registered"), nil

	default:
		return tools.ErrorResult(fmt.Errorf("invalid action: %s (use start, stop, or tap)", action)), nil
	}
}

// handleUniverseSelect handles the aftrs_universe_select tool
func handleUniverseSelect(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	universe := tools.GetIntParam(req, "universe", -1)

	if universe < 0 {
		return tools.ErrorResult(fmt.Errorf("universe is required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SelectUniverse(ctx, universe)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Selected universe %d", universe)), nil
}

// handleArtNetStatus handles the aftrs_artnet_status tool
func handleArtNetStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	status, err := client.GetArtNetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# ArtNet Status\n\n")

	connStatus := "Disconnected"
	if status.Connected {
		connStatus = "Connected"
	}
	sb.WriteString(fmt.Sprintf("**Status:** %s\n", connStatus))
	sb.WriteString(fmt.Sprintf("**IP:** %s\n", status.IP))
	sb.WriteString(fmt.Sprintf("**Port:** %d\n", status.Port))
	sb.WriteString(fmt.Sprintf("**Nodes Found:** %d\n", status.NodesFound))

	if len(status.Universes) > 0 {
		sb.WriteString(fmt.Sprintf("**Universes:** %v\n", status.Universes))
	}

	if status.PacketsSent > 0 || status.PacketsRecv > 0 {
		sb.WriteString("\n## Statistics\n\n")
		sb.WriteString(fmt.Sprintf("- Packets Sent: %d\n", status.PacketsSent))
		sb.WriteString(fmt.Sprintf("- Packets Received: %d\n", status.PacketsRecv))
	}

	if status.LastError != "" {
		sb.WriteString(fmt.Sprintf("\n**Last Error:** %s\n", status.LastError))
	}

	return tools.TextResult(sb.String()), nil
}

// handlePatchList handles the aftrs_patch_list tool
func handlePatchList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	patches, err := client.GetPatchList(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Fixture Patch\n\n")

	if len(patches) == 0 {
		sb.WriteString("No fixtures patched.\n\n")
		sb.WriteString("Use `aftrs_patch_fixture` to add fixtures to the patch.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("**%d** fixtures patched:\n\n", len(patches)))
	sb.WriteString("| Name | Type | Universe | Address | Channels |\n")
	sb.WriteString("|------|------|----------|---------|----------|\n")

	for _, p := range patches {
		sb.WriteString(fmt.Sprintf("| %s | %s | %d | %d | %d |\n",
			p.Name, p.FixtureType, p.Universe, p.Address, p.Channels))
	}

	return tools.TextResult(sb.String()), nil
}

// handlePatchFixture handles the aftrs_patch_fixture tool
func handlePatchFixture(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fixtureType := tools.GetStringParam(req, "fixture")
	universe := tools.GetIntParam(req, "universe", 0)
	address := tools.GetIntParam(req, "address", 1)
	name := tools.GetStringParam(req, "name")

	if fixtureType == "" {
		return tools.ErrorResult(fmt.Errorf("fixture type is required")), nil
	}
	if address < 1 || address > 512 {
		return tools.ErrorResult(fmt.Errorf("address must be between 1 and 512")), nil
	}

	if name == "" {
		name = fmt.Sprintf("%s_%d.%d", fixtureType, universe, address)
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.PatchFixture(ctx, fixtureType, name, universe, address)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Patched **%s** (%s) at universe %d, address %d",
		name, fixtureType, universe, address)), nil
}

// handlePatchUnpatch handles the aftrs_patch_unpatch tool
func handlePatchUnpatch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fixture, errResult := tools.RequireStringParam(req, "fixture")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.UnpatchFixture(ctx, fixture)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Unpatched fixture: %s", fixture)), nil
}

// handlePatchExport handles the aftrs_patch_export tool
func handlePatchExport(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	format := tools.OptionalStringParam(req, "format", "json")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	data, err := client.ExportPatch(ctx, format)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Patch Export (%s)\n\n", format))
	sb.WriteString("```" + format + "\n")
	sb.WriteString(data)
	sb.WriteString("\n```\n")

	return tools.TextResult(sb.String()), nil
}

// handleDMXBatchSet sets multiple DMX channel ranges in one call
func handleDMXBatchSet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	opsStr, errResult := tools.RequireStringParam(req, "operations")
	if errResult != nil {
		return errResult, nil
	}

	var ops []struct {
		StartChannel int    `json:"start_channel"`
		Values       string `json:"values"`
	}
	if err := json.Unmarshal([]byte(opsStr), &ops); err != nil {
		return tools.ErrorResult(fmt.Errorf("invalid JSON: %w", err)), nil
	}
	if len(ops) == 0 {
		return tools.ErrorResult(fmt.Errorf("operations array is empty")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# DMX Batch Set\n\n")

	var errors []string
	for i, op := range ops {
		if op.StartChannel < 1 || op.StartChannel > 512 {
			errors = append(errors, fmt.Sprintf("op %d: start_channel %d out of range (1-512)", i, op.StartChannel))
			continue
		}
		if op.Values == "" {
			errors = append(errors, fmt.Sprintf("op %d: values is empty", i))
			continue
		}

		var values []int
		for _, v := range strings.Split(op.Values, ",") {
			var val int
			fmt.Sscanf(strings.TrimSpace(v), "%d", &val)
			if val < 0 {
				val = 0
			}
			if val > 255 {
				val = 255
			}
			values = append(values, val)
		}

		if err := client.SetDMXChannels(ctx, op.StartChannel, values); err != nil {
			errors = append(errors, fmt.Sprintf("op %d: %v", i, err))
			continue
		}
		sb.WriteString(fmt.Sprintf("- Set %d channels starting at ch%d\n", len(values), op.StartChannel))
	}

	if len(errors) > 0 {
		sb.WriteString(fmt.Sprintf("\n**Errors:** %d\n", len(errors)))
		for _, e := range errors {
			sb.WriteString(fmt.Sprintf("- %s\n", e))
		}
	}

	sb.WriteString(fmt.Sprintf("\n**Total operations:** %d\n", len(ops)))
	return tools.TextResult(sb.String()), nil
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
