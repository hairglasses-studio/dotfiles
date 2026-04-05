// Package ola provides Open Lighting Architecture DMX control tools for hg-mcp.
package ola

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for OLA integration.
type Module struct{}

func (m *Module) Name() string {
	return "ola"
}

func (m *Module) Description() string {
	return "Open Lighting Architecture DMX universe management via HTTP REST"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_ola_status",
				mcp.WithDescription("Get OLA connection status and plugin list."),
			),
			Handler:             handleStatus,
			Category:            "lighting",
			Subcategory:         "ola",
			Tags:                []string{"ola", "status", "dmx", "lighting"},
			UseCases:            []string{"Check OLA connectivity", "List installed plugins"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ola",
		},
		{
			Tool: mcp.NewTool("aftrs_ola_health",
				mcp.WithDescription("Check OLA health and get troubleshooting recommendations."),
			),
			Handler:             handleHealth,
			Category:            "lighting",
			Subcategory:         "ola",
			Tags:                []string{"ola", "health", "diagnostics"},
			UseCases:            []string{"Diagnose OLA issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ola",
		},
		{
			Tool: mcp.NewTool("aftrs_ola_universe_info",
				mcp.WithDescription("Get universe metadata."),
				mcp.WithNumber("universe", mcp.Required(), mcp.Description("Universe ID")),
			),
			Handler:             handleUniverseInfo,
			Category:            "lighting",
			Subcategory:         "ola",
			Tags:                []string{"ola", "universe", "dmx", "info"},
			UseCases:            []string{"View universe configuration", "Check port assignments"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ola",
		},
		{
			Tool: mcp.NewTool("aftrs_ola_dmx_get",
				mcp.WithDescription("Read DMX channel values from a universe."),
				mcp.WithNumber("universe", mcp.Required(), mcp.Description("Universe ID")),
				mcp.WithNumber("start_channel", mcp.Description("Starting channel (1-512, default: 1)")),
				mcp.WithNumber("count", mcp.Description("Number of channels to read (default: 16)")),
			),
			Handler:             handleDmxGet,
			Category:            "lighting",
			Subcategory:         "ola",
			Tags:                []string{"ola", "dmx", "channels", "read"},
			UseCases:            []string{"Read DMX values", "Monitor channel output"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "ola",
		},
		{
			Tool: mcp.NewTool("aftrs_ola_dmx_set",
				mcp.WithDescription("Write DMX channel values to a universe."),
				mcp.WithNumber("universe", mcp.Required(), mcp.Description("Universe ID")),
				mcp.WithNumber("start_channel", mcp.Description("Starting channel (1-512, default: 1)")),
				mcp.WithString("values", mcp.Required(), mcp.Description("Comma-separated channel values (0-255)")),
			),
			Handler:             handleDmxSet,
			Category:            "lighting",
			Subcategory:         "ola",
			Tags:                []string{"ola", "dmx", "channels", "write"},
			UseCases:            []string{"Set DMX values", "Control lighting fixtures"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "ola",
		},
	}
}

var getClient = tools.LazyClient(clients.GetOLAClient)

// handleStatus returns OLA connection status and plugin list.
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create OLA client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# OLA Status\n\n")

	if !status.Connected {
		sb.WriteString("**Status:** Not Connected\n\n")
		sb.WriteString(fmt.Sprintf("**Target:** %s\n\n", status.URL))
		sb.WriteString("## Setup Required\n\n")
		sb.WriteString("1. Install OLA: `sudo apt install ola`\n")
		sb.WriteString("2. Start the daemon: `olad`\n")
		sb.WriteString("3. Access web UI at http://localhost:9090\n\n")
		sb.WriteString("**Environment Variables:**\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export OLA_HOST=localhost\n")
		sb.WriteString("export OLA_PORT=9090\n")
		sb.WriteString("```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("**Status:** Connected\n")
	sb.WriteString(fmt.Sprintf("**URL:** %s\n\n", status.URL))

	plugins, _ := client.GetPlugins(ctx)
	if len(plugins) > 0 {
		sb.WriteString("## Plugins\n\n")
		sb.WriteString("| ID | Name | Active | Enabled |\n")
		sb.WriteString("|----|------|--------|---------|\n")
		for _, p := range plugins {
			active := "No"
			if p.Active {
				active = "Yes"
			}
			enabled := "No"
			if p.Enabled {
				enabled = "Yes"
			}
			sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s |\n", p.ID, p.Name, active, enabled))
		}
	} else {
		sb.WriteString("No plugins found.\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleHealth returns OLA health and recommendations.
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
		issues = append(issues, "Not connected to OLA daemon")
		recommendations = append(recommendations,
			"Start olad: sudo systemctl start olad",
			fmt.Sprintf("Verify OLA is listening on %s", status.URL),
			"Check OLA_HOST and OLA_PORT env vars",
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

// handleUniverseInfo returns universe metadata.
func handleUniverseInfo(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	universeID := tools.GetIntParam(req, "universe", -1)
	if universeID < 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("universe ID is required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	info, err := client.GetUniverseInfo(ctx, universeID)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(info), nil
}

// handleDmxGet reads DMX channel values.
func handleDmxGet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	universeID := tools.GetIntParam(req, "universe", -1)
	if universeID < 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("universe ID is required")), nil
	}

	startChannel := tools.GetIntParam(req, "start_channel", 1)
	count := tools.GetIntParam(req, "count", 16)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	channels, err := client.GetDMX(ctx, universeID)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Slice to requested range
	start := startChannel - 1 // Convert to 0-based
	if start < 0 {
		start = 0
	}
	end := start + count
	if end > len(channels) {
		end = len(channels)
	}
	if start >= len(channels) {
		start = len(channels) - 1
	}

	slice := channels[start:end]

	result := map[string]interface{}{
		"universe":      universeID,
		"start_channel": startChannel,
		"count":         len(slice),
		"values":        slice,
	}
	return tools.JSONResult(result), nil
}

// handleDmxSet writes DMX channel values.
func handleDmxSet(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	universeID := tools.GetIntParam(req, "universe", -1)
	if universeID < 0 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("universe ID is required")), nil
	}

	valuesStr, errResult := tools.RequireStringParam(req, "values")
	if errResult != nil {
		return errResult, nil
	}

	startChannel := tools.GetIntParam(req, "start_channel", 1)

	// Parse comma-separated values
	parts := strings.Split(valuesStr, ",")
	values := make([]int, 0, len(parts))
	for _, p := range parts {
		v, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("invalid value: %s", p)), nil
		}
		if v < 0 || v > 255 {
			return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("value out of range (0-255): %d", v)), nil
		}
		values = append(values, v)
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SetDMX(ctx, universeID, startChannel, values)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Set %d channels starting at %d on universe %d", len(values), startChannel, universeID)), nil
}

// init registers this module with the global registry.
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
