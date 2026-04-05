// Package opc provides Open Pixel Control LED tools for hg-mcp.
package opc

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for OPC integration.
type Module struct{}

func (m *Module) Name() string {
	return "opc"
}

func (m *Module) Description() string {
	return "Open Pixel Control LED pixel control via TCP binary protocol"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_opc_status",
				mcp.WithDescription("Get Open Pixel Control server connection status."),
			),
			Handler:             handleStatus,
			Category:            "lighting",
			Subcategory:         "opc",
			Tags:                []string{"opc", "status", "led", "pixels"},
			UseCases:            []string{"Check OPC server connectivity"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "opc",
		},
		{
			Tool: mcp.NewTool("aftrs_opc_health",
				mcp.WithDescription("Check OPC server health and get troubleshooting recommendations."),
			),
			Handler:             handleHealth,
			Category:            "lighting",
			Subcategory:         "opc",
			Tags:                []string{"opc", "health", "diagnostics"},
			UseCases:            []string{"Diagnose OPC issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "opc",
		},
		{
			Tool: mcp.NewTool("aftrs_opc_set_pixels",
				mcp.WithDescription("Send RGB pixel data to LED strip."),
				mcp.WithNumber("channel", mcp.Description("OPC channel (0-255, default: 0)")),
				mcp.WithString("pixels", mcp.Required(), mcp.Description("Comma-separated RGB values (e.g., '255,0,0,0,255,0,0,0,255' for red,green,blue pixels)")),
			),
			Handler:             handleSetPixels,
			Category:            "lighting",
			Subcategory:         "opc",
			Tags:                []string{"opc", "pixels", "led", "rgb", "color"},
			UseCases:            []string{"Set LED colors", "Control pixel strips"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "opc",
		},
	}
}

var getClient = tools.LazyClient(clients.GetOPCClient)

// handleStatus returns OPC server connection status.
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create OPC client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Open Pixel Control Status\n\n")

	if !status.Connected {
		sb.WriteString("**Status:** Not Connected\n\n")
		sb.WriteString(fmt.Sprintf("**Server:** %s:%d\n\n", status.Host, status.Port))
		sb.WriteString("## Setup Required\n\n")
		sb.WriteString("1. Start an OPC server (e.g., Fadecandy, gl_server)\n")
		sb.WriteString("2. Default port is 7890\n\n")
		sb.WriteString("**Environment Variables:**\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export OPC_HOST=localhost\n")
		sb.WriteString("export OPC_PORT=7890\n")
		sb.WriteString("```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("**Status:** Connected\n")
	sb.WriteString(fmt.Sprintf("**Server:** %s:%d\n", status.Host, status.Port))

	return tools.TextResult(sb.String()), nil
}

// handleHealth returns OPC health and recommendations.
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
		issues = append(issues, "Not connected to OPC server")
		recommendations = append(recommendations,
			"Start an OPC server (Fadecandy, gl_server, etc.)",
			fmt.Sprintf("Verify server is listening on %s:%d", status.Host, status.Port),
			"Check OPC_HOST and OPC_PORT env vars",
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

// handleSetPixels sends RGB pixel data.
func handleSetPixels(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	pixelsStr, errResult := tools.RequireStringParam(req, "pixels")
	if errResult != nil {
		return errResult, nil
	}

	channel := tools.GetIntParam(req, "channel", 0)

	// Parse comma-separated RGB values
	parts := strings.Split(pixelsStr, ",")
	pixels := make([]byte, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid pixel value: %s", p)), nil
		}
		if v < 0 || v > 255 {
			return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("pixel value out of range (0-255): %d", v)), nil
		}
		pixels = append(pixels, byte(v))
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SetPixels(ctx, channel, pixels)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	numPixels := len(pixels) / 3
	return tools.TextResult(fmt.Sprintf("Set %d pixels on channel %d", numPixels, channel)), nil
}

// init registers this module with the global registry.
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
