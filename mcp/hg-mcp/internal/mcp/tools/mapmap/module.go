// Package mapmap provides MapMap video mapping tools for hg-mcp.
package mapmap

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for MapMap integration.
type Module struct{}

func (m *Module) Name() string {
	return "mapmap"
}

func (m *Module) Description() string {
	return "MapMap video mapping control via OSC"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_mapmap_status",
				mcp.WithDescription("Get MapMap connection status."),
			),
			Handler:             handleStatus,
			Category:            "video",
			Subcategory:         "mapmap",
			Tags:                []string{"mapmap", "status", "video", "mapping"},
			UseCases:            []string{"Check MapMap connectivity"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "mapmap",
		},
		{
			Tool: mcp.NewTool("aftrs_mapmap_health",
				mcp.WithDescription("Check MapMap health and get troubleshooting recommendations."),
			),
			Handler:             handleHealth,
			Category:            "video",
			Subcategory:         "mapmap",
			Tags:                []string{"mapmap", "health", "diagnostics"},
			UseCases:            []string{"Diagnose MapMap issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "mapmap",
		},
		{
			Tool: mcp.NewTool("aftrs_mapmap_surface",
				mcp.WithDescription("Control a mapping surface (set opacity, show/hide)."),
				mcp.WithNumber("surface_id", mcp.Required(), mcp.Description("Surface index (0-based)")),
				mcp.WithNumber("opacity", mcp.Description("Surface opacity (0.0-1.0)")),
				mcp.WithBoolean("visible", mcp.Description("Show or hide the surface")),
			),
			Handler:             handleSurface,
			Category:            "video",
			Subcategory:         "mapmap",
			Tags:                []string{"mapmap", "surface", "opacity", "projection"},
			UseCases:            []string{"Adjust surface opacity", "Toggle surface visibility"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "mapmap",
		},
	}
}

var getClient = tools.LazyClient(clients.GetMapMapClient)

// handleStatus returns MapMap connection status.
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create MapMap client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# MapMap Status\n\n")

	if !status.Connected {
		sb.WriteString("**Status:** Not Connected\n\n")
		sb.WriteString(fmt.Sprintf("**OSC:** %s:%d\n\n", status.Host, status.Port))
		sb.WriteString("## Setup Required\n\n")
		sb.WriteString("1. Open MapMap\n")
		sb.WriteString("2. Enable OSC in Preferences\n\n")
		sb.WriteString("**Environment Variables:**\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export MAPMAP_HOST=localhost\n")
		sb.WriteString("export MAPMAP_PORT=12345\n")
		sb.WriteString("```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("**Status:** Connected\n")
	sb.WriteString(fmt.Sprintf("**OSC:** %s:%d\n", status.Host, status.Port))

	return tools.TextResult(sb.String()), nil
}

// handleHealth returns MapMap health and recommendations.
func handleHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("health check failed: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("health check failed: %w", err)), nil
	}

	score := 100
	var issues []string
	var recommendations []string

	if !status.Connected {
		score -= 50
		issues = append(issues, "Not connected to MapMap")
		recommendations = append(recommendations,
			"Start MapMap and enable OSC",
			fmt.Sprintf("Verify MapMap is listening on %s:%d", status.Host, status.Port),
			"Check MAPMAP_HOST and MAPMAP_PORT env vars",
		)
	}

	healthStatus := "healthy"
	if score < 80 {
		healthStatus = "degraded"
	}
	if score < 50 {
		healthStatus = "critical"
	}

	health := map[string]interface{}{
		"score":           score,
		"status":          healthStatus,
		"connected":       status.Connected,
		"issues":          issues,
		"recommendations": recommendations,
	}
	return tools.JSONResult(health), nil
}

// handleSurface controls a mapping surface.
func handleSurface(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	surfaceID := tools.GetIntParam(req, "surface_id", -1)
	if surfaceID < 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("surface_id is required and must be non-negative")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var actions []string

	if v := tools.GetFloatParam(req, "opacity", -1); v >= 0 {
		err = client.SetSurfaceOpacity(ctx, surfaceID, v)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		actions = append(actions, fmt.Sprintf("opacity=%.2f", v))
	}

	if req.GetArguments()["visible"] != nil {
		visible := tools.GetBoolParam(req, "visible", false)
		err = client.SetSurfaceVisible(ctx, surfaceID, visible)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		actions = append(actions, fmt.Sprintf("visible=%v", visible))
	}

	if len(actions) == 0 {
		return tools.TextResult(fmt.Sprintf("Surface %d: no changes (provide opacity or visible)", surfaceID)), nil
	}

	return tools.TextResult(fmt.Sprintf("Surface %d: set %s", surfaceID, strings.Join(actions, ", "))), nil
}

// init registers this module with the global registry.
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
