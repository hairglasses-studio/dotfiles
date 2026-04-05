// Package streaming provides NDI and streaming tools for hg-mcp.
package streaming

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewNDIClient)

// Module implements the ToolModule interface for streaming
type Module struct{}

func (m *Module) Name() string {
	return "streaming"
}

func (m *Module) Description() string {
	return "NDI source discovery and streaming health monitoring"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_ndi_sources",
				mcp.WithDescription("List available NDI sources on the network."),
			),
			Handler:             handleNDISources,
			Category:            "streaming",
			Subcategory:         "ndi",
			Tags:                []string{"ndi", "sources", "discovery", "video"},
			UseCases:            []string{"Find NDI sources", "Check available video feeds"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "streaming",
		},
		{
			Tool: mcp.NewTool("aftrs_ndi_status",
				mcp.WithDescription("Get detailed status of an NDI source including frame rate and bandwidth."),
				mcp.WithString("source",
					mcp.Required(),
					mcp.Description("NDI source name or partial match"),
				),
			),
			Handler:             handleNDIStatus,
			Category:            "streaming",
			Subcategory:         "ndi",
			Tags:                []string{"ndi", "status", "video", "performance"},
			UseCases:            []string{"Check NDI source health", "Monitor video quality"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "streaming",
		},
		{
			Tool: mcp.NewTool("aftrs_stream_health",
				mcp.WithDescription("Get combined streaming health check including NDI, OBS, and capture devices."),
			),
			Handler:             handleStreamHealth,
			Category:            "streaming",
			Subcategory:         "health",
			Tags:                []string{"streaming", "health", "ndi", "obs", "video"},
			UseCases:            []string{"Check streaming infrastructure", "Pre-stream verification"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "streaming",
		},
		{
			Tool: mcp.NewTool("aftrs_stream_start",
				mcp.WithDescription("Start streaming to a destination (placeholder for future OBS integration)."),
				mcp.WithString("destination",
					mcp.Required(),
					mcp.Description("Streaming destination (e.g., 'twitch', 'youtube', 'custom')"),
				),
			),
			Handler:             handleStreamStart,
			Category:            "streaming",
			Subcategory:         "control",
			Tags:                []string{"streaming", "start", "obs", "broadcast"},
			UseCases:            []string{"Start live stream", "Begin broadcast"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "streaming",
			IsWrite:             true,
		},
	}
}

// handleNDISources handles the aftrs_ndi_sources tool
func handleNDISources(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	sources, err := client.DiscoverSources(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# NDI Sources\n\n")

	if len(sources) == 0 {
		sb.WriteString("No NDI sources discovered on the network.\n\n")
		sb.WriteString("## Troubleshooting\n\n")
		sb.WriteString("- Ensure NDI-enabled applications are running\n")
		sb.WriteString("- Check that NDI output is enabled in source applications\n")
		sb.WriteString("- Verify network connectivity between devices\n")
		sb.WriteString("- Install NDI Tools for `ndi-find` command support\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** NDI sources:\n\n", len(sources)))
	sb.WriteString("| Source | Host | Status |\n")
	sb.WriteString("|--------|------|--------|\n")

	for _, source := range sources {
		status := "🔴 Offline"
		if source.Connected {
			status = "🟢 Online"
		}
		host := source.Host
		if host == "" {
			host = "Unknown"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", source.Name, host, status))
	}

	sb.WriteString("\n---\n")
	sb.WriteString("Use `aftrs_ndi_status(source=\"name\")` for detailed source info.\n")

	return tools.TextResult(sb.String()), nil
}

// handleNDIStatus handles the aftrs_ndi_status tool
func handleNDIStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	sourceName, errResult := tools.RequireStringParam(req, "source")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	source, err := client.GetSourceStatus(ctx, sourceName)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# NDI Source: %s\n\n", source.Name))

	status := "🔴 Offline"
	if source.Connected {
		status = "🟢 Online"
	}
	sb.WriteString(fmt.Sprintf("**Status:** %s\n", status))
	sb.WriteString(fmt.Sprintf("**Host:** %s\n\n", source.Host))

	if source.Connected {
		sb.WriteString("## Stream Info\n\n")
		sb.WriteString(fmt.Sprintf("| Property | Value |\n"))
		sb.WriteString(fmt.Sprintf("|----------|-------|\n"))
		sb.WriteString(fmt.Sprintf("| Resolution | %dx%d |\n", source.Width, source.Height))
		sb.WriteString(fmt.Sprintf("| Frame Rate | %.1f fps |\n", source.FPS))
		sb.WriteString(fmt.Sprintf("| Bandwidth | %s |\n", source.Bandwidth))
	}

	return tools.TextResult(sb.String()), nil
}

// handleStreamHealth handles the aftrs_stream_health tool
func handleStreamHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	health, err := client.GetStreamHealth(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Streaming Health\n\n")

	// Status emoji
	statusEmoji := "✅"
	if health.Status == "degraded" {
		statusEmoji = "⚠️"
	} else if health.Status == "critical" {
		statusEmoji = "❌"
	}

	sb.WriteString(fmt.Sprintf("**Overall Score:** %d/100 %s\n", health.Score, statusEmoji))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n\n", health.Status))

	sb.WriteString("## Component Health\n\n")
	sb.WriteString("| Component | Score |\n")
	sb.WriteString("|-----------|-------|\n")
	sb.WriteString(fmt.Sprintf("| NDI | %d%% |\n", health.NDIHealth))
	sb.WriteString(fmt.Sprintf("| OBS | %d%% |\n", health.OBSHealth))
	sb.WriteString(fmt.Sprintf("| Capture | %d%% |\n", health.CaptureHealth))

	if len(health.Issues) > 0 {
		sb.WriteString("\n## Issues\n\n")
		for _, issue := range health.Issues {
			sb.WriteString(fmt.Sprintf("- ⚠️ %s\n", issue))
		}
	}

	if len(health.Recommendations) > 0 {
		sb.WriteString("\n## Recommendations\n\n")
		for _, rec := range health.Recommendations {
			sb.WriteString(fmt.Sprintf("- %s\n", rec))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleStreamStart handles the aftrs_stream_start tool
func handleStreamStart(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	destination, errResult := tools.RequireStringParam(req, "destination")
	if errResult != nil {
		return errResult, nil
	}

	var sb strings.Builder
	sb.WriteString("# Stream Control\n\n")
	sb.WriteString("**Status:** ⚠️ Not Implemented\n\n")
	sb.WriteString("Stream control requires OBS WebSocket integration.\n\n")
	sb.WriteString("## Setup Required\n\n")
	sb.WriteString("1. Enable OBS WebSocket server (Tools → WebSocket Server Settings)\n")
	sb.WriteString("2. Set environment variables:\n")
	sb.WriteString("```bash\n")
	sb.WriteString("export OBS_WEBSOCKET_URL=ws://localhost:4455\n")
	sb.WriteString("export OBS_WEBSOCKET_PASSWORD=your-password\n")
	sb.WriteString("```\n\n")
	sb.WriteString(fmt.Sprintf("**Requested Destination:** %s\n", destination))

	return tools.TextResult(sb.String()), nil
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
