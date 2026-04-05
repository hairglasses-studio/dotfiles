// Package sacn provides sACN/E1.31 lighting protocol tools for hg-mcp.
// sACN (Streaming Architecture for Control Networks) is an ANSI standard
// for transmitting DMX512 data over IP networks using multicast UDP.
// It complements ArtNet as an alternative DMX-over-IP protocol.
package sacn

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for sACN/E1.31.
type Module struct{}

// getClient returns the singleton sACN client (thread-safe via LazyClient).
var getClient = tools.LazyClient(clients.NewSACNClient)

func (m *Module) Name() string {
	return "sacn"
}

func (m *Module) Description() string {
	return "sACN/E1.31 DMX-over-IP protocol for lighting control"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_sacn_status",
				mcp.WithDescription("Get sACN/E1.31 system status including bind address, source name, and priority."),
			),
			Handler:             handleStatus,
			Category:            "lighting",
			Subcategory:         "sacn",
			Tags:                []string{"sacn", "e131", "dmx", "lighting", "status"},
			UseCases:            []string{"Check sACN configuration", "Verify DMX-over-IP status"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "sacn",
		},
		{
			Tool: mcp.NewTool("aftrs_sacn_health",
				mcp.WithDescription("Check sACN system health including network binding and multicast capability."),
			),
			Handler:             handleHealth,
			Category:            "lighting",
			Subcategory:         "sacn",
			Tags:                []string{"sacn", "health", "diagnostics", "network"},
			UseCases:            []string{"Diagnose sACN issues", "Check multicast capability"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "sacn",
		},
		{
			Tool: mcp.NewTool("aftrs_sacn_send",
				mcp.WithDescription("Send DMX channel values to a sACN universe via multicast UDP."),
				mcp.WithNumber("universe", mcp.Required(), mcp.Description("sACN universe number (1-63999)")),
				mcp.WithString("values", mcp.Required(), mcp.Description("Comma-separated channel:value pairs (e.g., '1:255,2:128,3:0')")),
			),
			Handler:             handleSend,
			Category:            "lighting",
			Subcategory:         "sacn",
			Tags:                []string{"sacn", "dmx", "send", "universe", "channels"},
			UseCases:            []string{"Send DMX data via sACN", "Set channel values"},
			Complexity:          tools.ComplexitySimple,
			IsWrite:             true,
			CircuitBreakerGroup: "sacn",
		},
		{
			Tool: mcp.NewTool("aftrs_sacn_discover",
				mcp.WithDescription("Discover sACN sources on the network by listening for E1.31 discovery packets."),
			),
			Handler:             handleDiscover,
			Category:            "lighting",
			Subcategory:         "sacn",
			Tags:                []string{"sacn", "discover", "sources", "network"},
			UseCases:            []string{"Find sACN sources", "Discover lighting controllers"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "sacn",
		},
	}
}

// handleStatus returns sACN system status.
func handleStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, err), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# sACN/E1.31 Status\n\n")
	if status.Active {
		sb.WriteString("**Status:** Active\n")
	} else {
		sb.WriteString("**Status:** Inactive\n")
	}
	sb.WriteString(fmt.Sprintf("**Bind Address:** %s\n", status.BindAddr))
	sb.WriteString(fmt.Sprintf("**Source Name:** %s\n", status.SourceName))
	sb.WriteString(fmt.Sprintf("**Priority:** %d\n", status.Priority))

	return tools.TextResult(sb.String()), nil
}

// handleHealth returns sACN system health.
func handleHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, err), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# sACN/E1.31 Health\n\n")
	sb.WriteString(fmt.Sprintf("**Score:** %d/100\n", health.Score))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n", health.Status))

	if len(health.Issues) > 0 {
		sb.WriteString("\n## Issues\n\n")
		for _, issue := range health.Issues {
			sb.WriteString(fmt.Sprintf("- %s\n", issue))
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

// handleSend sends DMX data to a sACN universe.
func handleSend(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	universe := tools.GetIntParam(req, "universe", 0)
	if universe < 1 || universe > 63999 {
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("universe must be 1-63999, got %d", universe)), nil
	}

	valuesStr, errResult := tools.RequireStringParam(req, "values")
	if errResult != nil {
		return errResult, nil
	}

	channelValues, err := parseChannelValuePairs(valuesStr)
	if err != nil {
		return tools.CodedErrorResult(tools.ErrInvalidParam, err), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, err), nil
	}

	// Build 512-byte DMX frame and set channel values
	data := make([]byte, 512)
	for ch, val := range channelValues {
		data[ch-1] = byte(val)
	}

	if err := client.SendUniverse(ctx, universe, data); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"universe": universe,
		"channels": channelValues,
		"success":  true,
	}), nil
}

// handleDiscover discovers sACN sources on the network.
func handleDiscover(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, err), nil
	}

	sources, err := client.DiscoverSources(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	if len(sources) == 0 {
		return tools.TextResult("# sACN Source Discovery\n\nNo sACN sources found on the network."), nil
	}

	var sb strings.Builder
	sb.WriteString("# sACN Source Discovery\n\n")
	sb.WriteString(fmt.Sprintf("**Found:** %d sources\n\n", len(sources)))
	sb.WriteString("| Name | CID | Universe | Priority | IP |\n")
	sb.WriteString("|---|---|---|---|---|\n")
	for _, src := range sources {
		sb.WriteString(fmt.Sprintf("| %s | %s | %d | %d | %s |\n",
			src.Name, src.CID, src.Universe, src.Priority, src.IP))
	}

	return tools.TextResult(sb.String()), nil
}

// parseChannelValuePairs parses "1:255,2:128,3:0" format.
func parseChannelValuePairs(input string) (map[int]int, error) {
	result := make(map[int]int)
	pairs := strings.Split(input, ",")
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid channel:value pair '%s', expected format '1:255'", pair)
		}
		ch := 0
		val := 0
		if _, err := fmt.Sscanf(parts[0], "%d", &ch); err != nil {
			return nil, fmt.Errorf("invalid channel number '%s': %w", parts[0], err)
		}
		if _, err := fmt.Sscanf(parts[1], "%d", &val); err != nil {
			return nil, fmt.Errorf("invalid value '%s': %w", parts[1], err)
		}
		if ch < 1 || ch > 512 {
			return nil, fmt.Errorf("channel %d out of range (1-512)", ch)
		}
		if val < 0 || val > 255 {
			return nil, fmt.Errorf("value %d out of range (0-255)", val)
		}
		result[ch] = val
	}
	return result, nil
}

// init registers this module with the global registry.
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
