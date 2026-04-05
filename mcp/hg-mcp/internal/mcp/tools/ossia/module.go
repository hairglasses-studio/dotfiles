// Package ossia provides ossia score interactive show control tools for hg-mcp.
package ossia

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for ossia score integration.
type Module struct{}

func (m *Module) Name() string {
	return "ossia"
}

func (m *Module) Description() string {
	return "ossia score interactive show control via OSCQuery"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_ossia_status",
				mcp.WithDescription("Get ossia score connection status and device info."),
			),
			Handler:             handleStatus,
			Category:            "automation",
			Subcategory:         "ossia",
			Tags:                []string{"ossia", "status", "oscquery", "show-control"},
			UseCases:            []string{"Check ossia connectivity", "View OSCQuery endpoint"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ossia",
		},
		{
			Tool: mcp.NewTool("aftrs_ossia_health",
				mcp.WithDescription("Check ossia score health and get troubleshooting recommendations."),
			),
			Handler:             handleHealth,
			Category:            "automation",
			Subcategory:         "ossia",
			Tags:                []string{"ossia", "health", "diagnostics"},
			UseCases:            []string{"Diagnose ossia issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ossia",
		},
		{
			Tool: mcp.NewTool("aftrs_ossia_devices",
				mcp.WithDescription("List discovered devices and parameters from the OSCQuery tree."),
			),
			Handler:             handleDevices,
			Category:            "automation",
			Subcategory:         "ossia",
			Tags:                []string{"ossia", "devices", "parameters", "oscquery"},
			UseCases:            []string{"Browse parameter tree", "Discover controllable devices"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ossia",
		},
		{
			Tool: mcp.NewTool("aftrs_ossia_transport",
				mcp.WithDescription("Control ossia score transport (play, stop, set position)."),
				mcp.WithString("action", mcp.Required(), mcp.Description("Action: play, stop, or position")),
				mcp.WithNumber("position", mcp.Description("Position in seconds (required for position action)")),
			),
			Handler:             handleTransport,
			Category:            "automation",
			Subcategory:         "ossia",
			Tags:                []string{"ossia", "transport", "play", "stop", "timeline"},
			UseCases:            []string{"Control timeline playback", "Seek to position"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "ossia",
		},
	}
}

var getClient = tools.LazyClient(clients.GetOssiaClient)

// handleStatus returns ossia connection status.
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create ossia client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# ossia score Status\n\n")

	if !status.Connected {
		sb.WriteString("**Status:** Not Connected\n\n")
		sb.WriteString(fmt.Sprintf("**OSCQuery:** %s\n\n", status.URL))
		sb.WriteString("## Setup Required\n\n")
		sb.WriteString("1. Open ossia score\n")
		sb.WriteString("2. Enable OSCQuery device in the Device Explorer\n")
		sb.WriteString("3. Default port is 5678\n\n")
		sb.WriteString("**Environment Variables:**\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export OSSIA_HOST=localhost\n")
		sb.WriteString("export OSSIA_PORT=5678\n")
		sb.WriteString("```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("**Status:** Connected\n")
	sb.WriteString(fmt.Sprintf("**OSCQuery:** %s\n", status.URL))

	return tools.TextResult(sb.String()), nil
}

// handleHealth returns ossia health and recommendations.
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
		issues = append(issues, "Not connected to ossia score")
		recommendations = append(recommendations,
			"Start ossia score and enable OSCQuery",
			fmt.Sprintf("Verify ossia is listening on %s", status.URL),
			"Check OSSIA_HOST and OSSIA_PORT env vars",
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

// handleDevices lists discovered devices from the OSCQuery tree.
func handleDevices(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	devices, err := client.GetDevices(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# ossia Devices\n\n")

	if len(devices) == 0 {
		sb.WriteString("No devices discovered.\n\n")
		sb.WriteString("*Note: Requires active OSCQuery connection to discover devices.*\n")
		return tools.TextResult(sb.String()), nil
	}

	for _, d := range devices {
		sb.WriteString(fmt.Sprintf("## %s\n\n", d.Name))
		if len(d.Parameters) > 0 {
			sb.WriteString("| Path | Type | Value |\n")
			sb.WriteString("|------|------|-------|\n")
			for _, p := range d.Parameters {
				sb.WriteString(fmt.Sprintf("| %s | %s | %v |\n", p.FullPath, p.Type, p.Value))
			}
		} else {
			sb.WriteString("No parameters.\n")
		}
		sb.WriteString("\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleTransport controls ossia transport.
func handleTransport(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	action, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	switch action {
	case "play":
		err = client.TransportPlay(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Transport: play"), nil

	case "stop":
		err = client.TransportStop(ctx)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult("Transport: stop"), nil

	case "position":
		position := tools.GetFloatParam(req, "position", -1)
		if position < 0 {
			return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("position is required and must be non-negative")), nil
		}
		err = client.TransportSetPosition(ctx, position)
		if err != nil {
			return tools.ErrorResult(err), nil
		}
		return tools.TextResult(fmt.Sprintf("Transport: position set to %.2fs", position)), nil

	default:
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid action: %s (use play, stop, or position)", action)), nil
	}
}

// init registers this module with the global registry.
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
