// Package videorouting provides MCP video routing tools.
package videorouting

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for video routing
type Module struct{}

var getClient = tools.LazyClient(clients.NewVideoRoutingClient)

func (m *Module) Name() string {
	return "videorouting"
}

func (m *Module) Description() string {
	return "Video routing and NDI source management"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_video_sources",
				mcp.WithDescription("List all available video sources from Resolume, OBS, ATEM, and NDI. Discovers sources across systems."),
			),
			Handler:             handleVideoSources,
			Category:            "video",
			Subcategory:         "routing",
			Tags:                []string{"video", "ndi", "sources", "discover"},
			UseCases:            []string{"Find available video sources", "Check NDI streams", "View video inputs"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "videorouting",
		},
		{
			Tool: mcp.NewTool("aftrs_video_route",
				mcp.WithDescription("Create or manage video routes between systems. Routes NDI/video sources to destinations."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: create, delete, enable, disable, list")),
				mcp.WithString("name", mcp.Description("Route name (for create)")),
				mcp.WithString("source_id", mcp.Description("Source ID (for create)")),
				mcp.WithString("destination", mcp.Description("Destination system: resolume, obs, atem, touchdesigner")),
				mcp.WithString("dest_input", mcp.Description("Destination input (e.g., 'layer/1/column/1' for Resolume, 'me/1/input/1' for ATEM)")),
				mcp.WithString("route_id", mcp.Description("Route ID (for delete/enable/disable)")),
			),
			Handler:             handleVideoRoute,
			Category:            "video",
			Subcategory:         "routing",
			Tags:                []string{"video", "ndi", "route", "routing"},
			UseCases:            []string{"Route NDI to Resolume", "Send video to ATEM", "Configure video matrix"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "videorouting",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_video_matrix",
				mcp.WithDescription("View the complete video routing matrix showing all sources, destinations, and active routes."),
			),
			Handler:             handleVideoMatrix,
			Category:            "video",
			Subcategory:         "routing",
			Tags:                []string{"video", "matrix", "routing", "overview"},
			UseCases:            []string{"View video routing overview", "Check active routes", "Audit video flow"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "videorouting",
		},
	}
}

func handleVideoSources(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	sources, err := client.DiscoverSources(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Video Sources\n\n")

	if len(sources) == 0 {
		sb.WriteString("No video sources discovered.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** sources:\n\n", len(sources)))

	// Group by system
	bySystem := make(map[string][]*clients.VideoSource)
	for _, src := range sources {
		bySystem[src.System] = append(bySystem[src.System], src)
	}

	for system, systemSources := range bySystem {
		sb.WriteString(fmt.Sprintf("## %s\n\n", strings.Title(system)))
		sb.WriteString("| ID | Name | Type | Resolution | Status |\n")
		sb.WriteString("|----|------|------|------------|--------|\n")

		for _, src := range systemSources {
			status := "Offline"
			if src.Connected {
				status = "Online"
			}
			resolution := src.Resolution
			if resolution == "" {
				resolution = "-"
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
				truncateID(src.ID), src.Name, src.Type, resolution, status))
		}
		sb.WriteString("\n")
	}

	return tools.TextResult(sb.String()), nil
}

func handleVideoRoute(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action := tools.GetStringParam(req, "action")

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "create":
		name := tools.GetStringParam(req, "name")
		sourceID := tools.GetStringParam(req, "source_id")
		destination := tools.GetStringParam(req, "destination")
		destInput := tools.GetStringParam(req, "dest_input")

		if name == "" || sourceID == "" || destination == "" {
			return tools.ErrorResult(fmt.Errorf("name, source_id, and destination are required for create")), nil
		}

		route, err := client.CreateRoute(ctx, name, sourceID, destination, destInput)
		if err != nil {
			return tools.ErrorResult(err), nil
		}

		return tools.TextResult(fmt.Sprintf("# Route Created\n\n**ID:** `%s`\n**Name:** %s\n**Source:** %s\n**Destination:** %s → %s\n**Active:** Yes",
			route.ID, route.Name, route.Source.Name, route.Destination, route.DestInput)), nil

	case "delete":
		routeID, errResult := tools.RequireStringParam(req, "route_id")
		if errResult != nil {
			return errResult, nil
		}

		if err := client.DeleteRoute(ctx, routeID); err != nil {
			return tools.ErrorResult(err), nil
		}

		return tools.TextResult(fmt.Sprintf("✅ Route `%s` deleted.", truncateID(routeID))), nil

	case "enable":
		routeID, errResult := tools.RequireStringParam(req, "route_id")
		if errResult != nil {
			return errResult, nil
		}

		if err := client.SetRouteActive(ctx, routeID, true); err != nil {
			return tools.ErrorResult(err), nil
		}

		return tools.TextResult(fmt.Sprintf("✅ Route `%s` enabled.", truncateID(routeID))), nil

	case "disable":
		routeID, errResult := tools.RequireStringParam(req, "route_id")
		if errResult != nil {
			return errResult, nil
		}

		if err := client.SetRouteActive(ctx, routeID, false); err != nil {
			return tools.ErrorResult(err), nil
		}

		return tools.TextResult(fmt.Sprintf("✅ Route `%s` disabled.", truncateID(routeID))), nil

	case "list":
		matrix, err := client.GetMatrix(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}

		var sb strings.Builder
		sb.WriteString("# Video Routes\n\n")

		if len(matrix.Routes) == 0 {
			sb.WriteString("No routes configured.\n")
			return tools.TextResult(sb.String()), nil
		}

		sb.WriteString("| ID | Name | Source | Destination | Active |\n")
		sb.WriteString("|----|------|--------|-------------|--------|\n")

		for _, route := range matrix.Routes {
			active := "No"
			if route.Active {
				active = "Yes"
			}
			dest := fmt.Sprintf("%s:%s", route.Destination, route.DestInput)
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
				truncateID(route.ID), route.Name, route.Source.Name, dest, active))
		}

		return tools.TextResult(sb.String()), nil

	default:
		return tools.ErrorResult(fmt.Errorf("unknown action: %s (valid: create, delete, enable, disable, list)", action)), nil
	}
}

func handleVideoMatrix(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	matrix, err := client.GetMatrix(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Video Routing Matrix\n\n")
	sb.WriteString(fmt.Sprintf("*Updated: %s*\n\n", matrix.UpdatedAt.Format("15:04:05")))

	// Sources summary
	sb.WriteString("## Sources\n\n")
	sb.WriteString(fmt.Sprintf("**%d** sources available:\n\n", len(matrix.Sources)))

	onlineCount := 0
	for _, src := range matrix.Sources {
		if src.Connected {
			onlineCount++
		}
	}
	sb.WriteString(fmt.Sprintf("- Online: %d\n", onlineCount))
	sb.WriteString(fmt.Sprintf("- Offline: %d\n\n", len(matrix.Sources)-onlineCount))

	// Destinations
	sb.WriteString("## Destinations\n\n")
	for _, dest := range matrix.Destinations {
		sb.WriteString(fmt.Sprintf("- %s\n", dest))
	}
	sb.WriteString("\n")

	// Active routes
	sb.WriteString("## Active Routes\n\n")
	if len(matrix.Routes) == 0 {
		sb.WriteString("No routes configured.\n")
	} else {
		activeCount := 0
		for _, route := range matrix.Routes {
			if route.Active {
				activeCount++
				sb.WriteString(fmt.Sprintf("- **%s**: %s → %s:%s\n",
					route.Name, route.Source.Name, route.Destination, route.DestInput))
			}
		}
		if activeCount == 0 {
			sb.WriteString("No active routes.\n")
		}
	}

	return tools.TextResult(sb.String()), nil
}

func truncateID(id string) string {
	if len(id) > 12 {
		return id[:12]
	}
	return id
}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
