// Package paramsync provides MCP parameter mapping tools.
package paramsync

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for parameter sync
type Module struct{}

var getClient = tools.LazyClient(clients.NewParamSyncClient)

func (m *Module) Name() string {
	return "paramsync"
}

func (m *Module) Description() string {
	return "Cross-system parameter mapping and modulation"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_param_map",
				mcp.WithDescription("Create a parameter mapping between two systems. Maps source parameter to target with optional value transform."),
				mcp.WithString("name", mcp.Required(), mcp.Description("Name for this mapping")),
				mcp.WithString("source_system", mcp.Required(), mcp.Description("Source system: ableton, resolume, grandma3, touchdesigner")),
				mcp.WithString("source_path", mcp.Required(), mcp.Description("Source param path (e.g., 'track/0/device/1/param/3', 'layer/1/opacity', 'master/level')")),
				mcp.WithString("target_system", mcp.Required(), mcp.Description("Target system: ableton, resolume, grandma3, touchdesigner")),
				mcp.WithString("target_path", mcp.Required(), mcp.Description("Target param path")),
				mcp.WithNumber("input_min", mcp.Description("Input minimum value (default: 0)")),
				mcp.WithNumber("input_max", mcp.Description("Input maximum value (default: 1)")),
				mcp.WithNumber("output_min", mcp.Description("Output minimum value (default: 0)")),
				mcp.WithNumber("output_max", mcp.Description("Output maximum value (default: 1)")),
				mcp.WithString("curve", mcp.Description("Transform curve: linear, exponential, logarithmic")),
				mcp.WithBoolean("invert", mcp.Description("Invert the output value")),
			),
			Handler:             handleParamMap,
			Category:            "paramsync",
			Subcategory:         "mapping",
			Tags:                []string{"parameter", "mapping", "modulation", "sync"},
			UseCases:            []string{"Map Ableton device to Resolume effect", "Control lighting from audio"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "paramsync",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_param_maps",
				mcp.WithDescription("List all parameter mappings."),
				mcp.WithString("mapping_id", mcp.Description("Get details of a specific mapping")),
			),
			Handler:             handleParamMaps,
			Category:            "paramsync",
			Subcategory:         "mapping",
			Tags:                []string{"parameter", "mapping", "list"},
			UseCases:            []string{"View active mappings", "Check mapping configuration"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "paramsync",
		},
		{
			Tool: mcp.NewTool("aftrs_param_push",
				mcp.WithDescription("Manually push a value through a parameter mapping."),
				mcp.WithString("mapping_id", mcp.Required(), mcp.Description("Mapping ID to push through")),
				mcp.WithNumber("value", mcp.Required(), mcp.Description("Value to push (will be transformed)")),
			),
			Handler:             handleParamPush,
			Category:            "paramsync",
			Subcategory:         "mapping",
			Tags:                []string{"parameter", "push", "value"},
			UseCases:            []string{"Test mapping", "Manual parameter control"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "paramsync",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_param_sync",
				mcp.WithDescription("Sync a mapping by reading source and pushing to target."),
				mcp.WithString("mapping_id", mcp.Required(), mcp.Description("Mapping ID to sync")),
			),
			Handler:             handleParamSync,
			Category:            "paramsync",
			Subcategory:         "mapping",
			Tags:                []string{"parameter", "sync", "read", "write"},
			UseCases:            []string{"Sync parameter from source to target", "One-shot sync"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "paramsync",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_param_sync_start",
				mcp.WithDescription("Start continuous parameter sync. Polls source at given interval and pushes to target."),
				mcp.WithString("mapping_id", mcp.Required(), mcp.Description("Mapping ID to continuously sync")),
				mcp.WithNumber("interval_ms", mcp.Description("Sync interval in milliseconds (default: 100, min: 10)")),
			),
			Handler:             handleParamSyncStart,
			Category:            "paramsync",
			Subcategory:         "continuous",
			Tags:                []string{"parameter", "sync", "continuous", "start"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "paramsync",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_param_sync_stop",
				mcp.WithDescription("Stop continuous parameter sync for a mapping."),
				mcp.WithString("mapping_id", mcp.Required(), mcp.Description("Mapping ID to stop syncing")),
			),
			Handler:             handleParamSyncStop,
			Category:            "paramsync",
			Subcategory:         "continuous",
			Tags:                []string{"parameter", "sync", "continuous", "stop"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "paramsync",
			IsWrite:             true,
		},
	}
}

func handleParamSyncStart(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	mappingID := tools.GetStringParam(req, "mapping_id")
	intervalMs := tools.GetIntParam(req, "interval_ms", 100)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.StartContinuousSync(ctx, mappingID, intervalMs); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Started continuous sync for mapping `%s` at %dms interval", mappingID, intervalMs)), nil
}

func handleParamSyncStop(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	mappingID := tools.GetStringParam(req, "mapping_id")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.StopContinuousSync(mappingID); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Stopped continuous sync for mapping `%s`", mappingID)), nil
}

func handleParamMap(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := tools.GetStringParam(req, "name")
	sourceSystem := tools.GetStringParam(req, "source_system")
	sourcePath := tools.GetStringParam(req, "source_path")
	targetSystem := tools.GetStringParam(req, "target_system")
	targetPath := tools.GetStringParam(req, "target_path")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	source := clients.ParamEndpoint{
		System: sourceSystem,
		Path:   sourcePath,
	}

	target := clients.ParamEndpoint{
		System: targetSystem,
		Path:   targetPath,
	}

	// Build transform if any values provided
	var transform *clients.ParamTransform
	inputMin := tools.GetFloatParam(req, "input_min", -1)
	inputMax := tools.GetFloatParam(req, "input_max", -1)
	outputMin := tools.GetFloatParam(req, "output_min", -1)
	outputMax := tools.GetFloatParam(req, "output_max", -1)
	curve := tools.GetStringParam(req, "curve")
	invert := tools.GetBoolParam(req, "invert", false)

	if inputMin >= 0 || inputMax >= 0 || outputMin >= 0 || outputMax >= 0 || curve != "" || invert {
		transform = &clients.ParamTransform{
			InputMin:  inputMin,
			InputMax:  inputMax,
			OutputMin: outputMin,
			OutputMax: outputMax,
			Curve:     curve,
			Invert:    invert,
		}
		// Set defaults for unspecified values
		if transform.InputMin < 0 {
			transform.InputMin = 0
		}
		if transform.InputMax < 0 {
			transform.InputMax = 1
		}
		if transform.OutputMin < 0 {
			transform.OutputMin = 0
		}
		if transform.OutputMax < 0 {
			transform.OutputMax = 1
		}
	}

	mapping, err := client.CreateMapping(ctx, name, source, target, transform)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Parameter Mapping Created\n\n")
	sb.WriteString(fmt.Sprintf("**ID:** `%s`\n", mapping.ID))
	sb.WriteString(fmt.Sprintf("**Name:** %s\n", mapping.Name))
	sb.WriteString(fmt.Sprintf("**Source:** %s → `%s`\n", mapping.Source.System, mapping.Source.Path))
	sb.WriteString(fmt.Sprintf("**Target:** %s → `%s`\n", mapping.Target.System, mapping.Target.Path))

	if mapping.Transform != nil {
		sb.WriteString("\n## Transform\n")
		sb.WriteString(fmt.Sprintf("- Input range: %.2f - %.2f\n", mapping.Transform.InputMin, mapping.Transform.InputMax))
		sb.WriteString(fmt.Sprintf("- Output range: %.2f - %.2f\n", mapping.Transform.OutputMin, mapping.Transform.OutputMax))
		if mapping.Transform.Curve != "" {
			sb.WriteString(fmt.Sprintf("- Curve: %s\n", mapping.Transform.Curve))
		}
		if mapping.Transform.Invert {
			sb.WriteString("- Inverted: Yes\n")
		}
	}

	return tools.TextResult(sb.String()), nil
}

func handleParamMaps(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	mappingID := tools.GetStringParam(req, "mapping_id")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Single mapping details
	if mappingID != "" {
		mapping, err := client.GetMapping(ctx, mappingID)
		if err != nil {
			return tools.ErrorResult(err), nil
		}

		var sb strings.Builder
		sb.WriteString("# Mapping Details\n\n")
		sb.WriteString(fmt.Sprintf("**ID:** `%s`\n", mapping.ID))
		sb.WriteString(fmt.Sprintf("**Name:** %s\n", mapping.Name))
		sb.WriteString(fmt.Sprintf("**Enabled:** %t\n", mapping.Enabled))
		sb.WriteString(fmt.Sprintf("**Created:** %s\n", mapping.CreatedAt.Format("2006-01-02 15:04:05")))
		sb.WriteString(fmt.Sprintf("**Sync Count:** %d\n", mapping.SyncCount))
		if !mapping.LastSync.IsZero() {
			sb.WriteString(fmt.Sprintf("**Last Sync:** %s\n", mapping.LastSync.Format("15:04:05")))
		}

		sb.WriteString("\n## Source\n")
		sb.WriteString(fmt.Sprintf("- System: %s\n", mapping.Source.System))
		sb.WriteString(fmt.Sprintf("- Path: `%s`\n", mapping.Source.Path))

		sb.WriteString("\n## Target\n")
		sb.WriteString(fmt.Sprintf("- System: %s\n", mapping.Target.System))
		sb.WriteString(fmt.Sprintf("- Path: `%s`\n", mapping.Target.Path))

		if mapping.Transform != nil {
			sb.WriteString("\n## Transform\n")
			sb.WriteString(fmt.Sprintf("- Input: %.2f - %.2f\n", mapping.Transform.InputMin, mapping.Transform.InputMax))
			sb.WriteString(fmt.Sprintf("- Output: %.2f - %.2f\n", mapping.Transform.OutputMin, mapping.Transform.OutputMax))
			sb.WriteString(fmt.Sprintf("- Curve: %s\n", mapping.Transform.Curve))
			sb.WriteString(fmt.Sprintf("- Inverted: %t\n", mapping.Transform.Invert))
		}

		return tools.TextResult(sb.String()), nil
	}

	// List all mappings
	mappings := client.ListMappings(ctx)

	var sb strings.Builder
	sb.WriteString("# Parameter Mappings\n\n")

	if len(mappings) == 0 {
		sb.WriteString("No mappings configured. Use `aftrs_param_map` to create one.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** mappings:\n\n", len(mappings)))
	sb.WriteString("| ID | Name | Source | Target | Enabled | Syncs |\n")
	sb.WriteString("|----|------|--------|--------|---------|-------|\n")

	for _, m := range mappings {
		enabled := "Yes"
		if !m.Enabled {
			enabled = "No"
		}
		source := fmt.Sprintf("%s:%s", m.Source.System, truncate(m.Source.Path, 15))
		target := fmt.Sprintf("%s:%s", m.Target.System, truncate(m.Target.Path, 15))
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %d |\n",
			m.ID[:12], m.Name, source, target, enabled, m.SyncCount))
	}

	return tools.TextResult(sb.String()), nil
}

func handleParamPush(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	mappingID := tools.GetStringParam(req, "mapping_id")
	value := tools.GetFloatParam(req, "value", 0)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.PushValue(ctx, mappingID, value); err != nil {
		return tools.ErrorResult(err), nil
	}

	mapping, _ := client.GetMapping(ctx, mappingID)
	targetInfo := "target"
	if mapping != nil {
		targetInfo = fmt.Sprintf("%s:%s", mapping.Target.System, mapping.Target.Path)
	}

	return tools.TextResult(fmt.Sprintf("✅ Pushed value **%.3f** through mapping `%s` to %s", value, mappingID[:12], targetInfo)), nil
}

func handleParamSync(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	mappingID := tools.GetStringParam(req, "mapping_id")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if err := client.SyncMapping(ctx, mappingID); err != nil {
		return tools.ErrorResult(err), nil
	}

	mapping, _ := client.GetMapping(ctx, mappingID)
	if mapping != nil {
		return tools.TextResult(fmt.Sprintf("✅ Synced mapping `%s` (%s)\n\n- Source: %s:%s → Target: %s:%s",
			mapping.Name, mappingID[:12],
			mapping.Source.System, mapping.Source.Path,
			mapping.Target.System, mapping.Target.Path)), nil
	}

	return tools.TextResult(fmt.Sprintf("✅ Synced mapping `%s`", mappingID[:12])), nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
