// Package puredata provides Pure Data (Pd) control tools for hg-mcp.
package puredata

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for Pure Data integration.
type Module struct{}

func (m *Module) Name() string {
	return "puredata"
}

func (m *Module) Description() string {
	return "Pure Data (Pd) visual programming audio control via FUDI/OSC"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_puredata_status",
				mcp.WithDescription("Get Pure Data connection status."),
			),
			Handler:             handleStatus,
			Category:            "audio",
			Subcategory:         "puredata",
			Tags:                []string{"puredata", "pd", "status", "fudi"},
			UseCases:            []string{"Check Pd connectivity"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "puredata",
		},
		{
			Tool: mcp.NewTool("aftrs_puredata_health",
				mcp.WithDescription("Check Pure Data health and get troubleshooting recommendations."),
			),
			Handler:             handleHealth,
			Category:            "audio",
			Subcategory:         "puredata",
			Tags:                []string{"puredata", "pd", "health", "diagnostics"},
			UseCases:            []string{"Diagnose Pd issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "puredata",
		},
		{
			Tool: mcp.NewTool("aftrs_puredata_send",
				mcp.WithDescription("Send a FUDI message to Pure Data."),
				mcp.WithString("message", mcp.Required(), mcp.Description("FUDI message (e.g., 'receiver value1 value2')")),
			),
			Handler:             handleSend,
			Category:            "audio",
			Subcategory:         "puredata",
			Tags:                []string{"puredata", "pd", "fudi", "send", "message"},
			UseCases:            []string{"Send messages to Pd patches", "Control patch parameters"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "puredata",
		},
		{
			Tool: mcp.NewTool("aftrs_puredata_dsp",
				mcp.WithDescription("Turn Pure Data DSP processing on or off."),
				mcp.WithBoolean("on", mcp.Required(), mcp.Description("True to enable DSP, false to disable")),
			),
			Handler:             handleDSP,
			Category:            "audio",
			Subcategory:         "puredata",
			Tags:                []string{"puredata", "pd", "dsp", "audio", "toggle"},
			UseCases:            []string{"Enable audio processing", "Disable DSP"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "puredata",
		},
	}
}

var getClient = tools.LazyClient(clients.GetPureDataClient)

// handleStatus returns Pure Data connection status.
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Pure Data client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Pure Data Status\n\n")

	if !status.Connected {
		sb.WriteString("**Status:** Not Connected\n\n")
		sb.WriteString(fmt.Sprintf("**FUDI:** %s:%d\n\n", status.Host, status.Port))
		sb.WriteString("## Setup Required\n\n")
		sb.WriteString("1. Open Pure Data\n")
		sb.WriteString("2. Create a [netreceive 3000] object in your patch\n")
		sb.WriteString("3. Connect it to receive FUDI messages\n\n")
		sb.WriteString("**Environment Variables:**\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export PUREDATA_HOST=localhost\n")
		sb.WriteString("export PUREDATA_PORT=3000\n")
		sb.WriteString("```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("**Status:** Connected\n")
	sb.WriteString(fmt.Sprintf("**FUDI:** %s:%d\n", status.Host, status.Port))

	return tools.TextResult(sb.String()), nil
}

// handleHealth returns Pure Data health and recommendations.
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
		issues = append(issues, "Not connected to Pure Data")
		recommendations = append(recommendations,
			"Start Pure Data with a patch containing [netreceive]",
			fmt.Sprintf("Verify Pd is listening on %s:%d", status.Host, status.Port),
			"Check PUREDATA_HOST and PUREDATA_PORT env vars",
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

// handleSend sends a FUDI message to Pure Data.
func handleSend(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	message, errResult := tools.RequireStringParam(req, "message")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SendFUDI(ctx, message)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Sent FUDI: %s;", message)), nil
}

// handleDSP controls DSP processing.
func handleDSP(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	on := tools.GetBoolParam(req, "on", false)

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	err = client.SetDSP(ctx, on)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	state := "off"
	if on {
		state = "on"
	}
	return tools.TextResult(fmt.Sprintf("DSP turned %s", state)), nil
}

// init registers this module with the global registry.
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
