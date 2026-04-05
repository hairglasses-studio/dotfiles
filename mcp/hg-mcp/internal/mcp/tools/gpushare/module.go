// Package gpushare provides Syphon/Spout GPU texture sharing detection tools for hg-mcp.
package gpushare

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for GPU texture sharing detection.
type Module struct{}

func (m *Module) Name() string {
	return "gpushare"
}

func (m *Module) Description() string {
	return "Syphon/Spout GPU texture sharing source detection"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_gpushare_status",
				mcp.WithDescription("Get Syphon/Spout GPU texture sharing detection status."),
			),
			Handler:             handleStatus,
			Category:            "video",
			Subcategory:         "gpushare",
			Tags:                []string{"gpushare", "syphon", "spout", "status", "gpu"},
			UseCases:            []string{"Check GPU texture sharing helper connectivity"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "gpushare",
		},
		{
			Tool: mcp.NewTool("aftrs_gpushare_sources",
				mcp.WithDescription("List active Syphon/Spout sources sharing GPU textures."),
			),
			Handler:             handleSources,
			Category:            "video",
			Subcategory:         "gpushare",
			Tags:                []string{"gpushare", "syphon", "spout", "sources", "texture"},
			UseCases:            []string{"Discover shared GPU textures", "List video sources"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "gpushare",
		},
	}
}

var getClient = tools.LazyClient(clients.GetGPUShareClient)

// handleStatus returns GPU texture sharing detection status.
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create gpushare client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# GPU Texture Sharing Status\n\n")

	if !status.Connected {
		sb.WriteString("**Status:** Not Connected\n\n")
		sb.WriteString(fmt.Sprintf("**Helper:** %s\n\n", status.HelperURL))
		sb.WriteString("## Setup Required\n\n")
		sb.WriteString("A local helper process is needed to detect Syphon (macOS) or Spout (Windows) sources.\n\n")
		sb.WriteString("**Environment Variables:**\n")
		sb.WriteString("```bash\n")
		sb.WriteString("export GPUSHARE_HELPER_URL=http://localhost:9876\n")
		sb.WriteString("```\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString("**Status:** Connected\n")
	sb.WriteString(fmt.Sprintf("**Helper:** %s\n", status.HelperURL))
	sb.WriteString(fmt.Sprintf("**Platform:** %s\n", status.Platform))

	return tools.TextResult(sb.String()), nil
}

// handleSources lists active Syphon/Spout sources.
func handleSources(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	sources, err := client.GetSources(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# GPU Texture Sources\n\n")

	if len(sources) == 0 {
		sb.WriteString("No active sources detected.\n\n")
		sb.WriteString("*Note: Requires the GPU share helper process and active Syphon/Spout applications.*\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** sources:\n\n", len(sources)))
	sb.WriteString("| Name | Application | Type | Resolution |\n")
	sb.WriteString("|------|-------------|------|------------|\n")
	for _, s := range sources {
		res := "N/A"
		if s.Width > 0 && s.Height > 0 {
			res = fmt.Sprintf("%dx%d", s.Width, s.Height)
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", s.Name, s.Application, s.Type, res))
	}

	return tools.TextResult(sb.String()), nil
}

// init registers this module with the global registry.
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
