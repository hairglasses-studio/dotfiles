// Package federation provides MCP server federation tools
package federation

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Module implements the federation tools module
type Module struct{}

// Name returns the module name
func (m *Module) Name() string { return "federation" }

// Description returns the module description
func (m *Module) Description() string {
	return "MCP server federation for connecting to remote MCP servers"
}

// Tools returns all federation tools
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_federation_servers",
				mcp.WithDescription("List known MCP servers that can be connected (TouchDesigner, Home Assistant, OBS, etc.)"),
				mcp.WithString("category", mcp.Description("Filter by category (creative, infrastructure, streaming, audio)")),
			),
			Handler:             handleListKnownServers,
			Category:            "federation",
			CircuitBreakerGroup: "federation",
		},
		{
			Tool: mcp.NewTool("aftrs_federation_connect",
				mcp.WithDescription("Connect to a remote MCP server to federate its tools"),
				mcp.WithString("name", mcp.Required(), mcp.Description("Server name (e.g., touchdesigner, home-assistant)")),
				mcp.WithString("url", mcp.Required(), mcp.Description("Server URL (e.g., http://localhost:9988)")),
				mcp.WithString("transport", mcp.Description("Transport type: http, sse (default: http)")),
			),
			Handler:             handleConnect,
			Category:            "federation",
			CircuitBreakerGroup: "federation",
		},
		{
			Tool: mcp.NewTool("aftrs_federation_disconnect",
				mcp.WithDescription("Disconnect from a federated MCP server"),
				mcp.WithString("name", mcp.Required(), mcp.Description("Server name to disconnect")),
			),
			Handler:             handleDisconnect,
			Category:            "federation",
			CircuitBreakerGroup: "federation",
		},
		{
			Tool: mcp.NewTool("aftrs_federation_connections",
				mcp.WithDescription("List all federated server connections and their status"),
			),
			Handler:             handleListConnections,
			Category:            "federation",
			CircuitBreakerGroup: "federation",
		},
		{
			Tool: mcp.NewTool("aftrs_federation_tools",
				mcp.WithDescription("List all tools from federated servers"),
				mcp.WithString("server", mcp.Description("Filter by server name (optional)")),
				mcp.WithString("search", mcp.Description("Search tools by name or description")),
			),
			Handler:             handleListTools,
			Category:            "federation",
			CircuitBreakerGroup: "federation",
		},
		{
			Tool: mcp.NewTool("aftrs_federation_call",
				mcp.WithDescription("Call a tool on a federated server"),
				mcp.WithString("server", mcp.Required(), mcp.Description("Server name")),
				mcp.WithString("tool", mcp.Required(), mcp.Description("Tool name to call")),
				mcp.WithObject("args", mcp.Description("Tool arguments as JSON object")),
			),
			Handler:             handleCallTool,
			Category:            "federation",
			CircuitBreakerGroup: "federation",
		},
		{
			Tool: mcp.NewTool("aftrs_federation_ping",
				mcp.WithDescription("Ping a federated server to check connectivity"),
				mcp.WithString("name", mcp.Required(), mcp.Description("Server name to ping")),
			),
			Handler:             handlePing,
			Category:            "federation",
			CircuitBreakerGroup: "federation",
		},
		{
			Tool: mcp.NewTool("aftrs_federation_refresh",
				mcp.WithDescription("Refresh tool lists from all connected servers"),
			),
			Handler:             handleRefresh,
			Category:            "federation",
			CircuitBreakerGroup: "federation",
		},
		{
			Tool: mcp.NewTool("aftrs_federation_stats",
				mcp.WithDescription("Get federation statistics"),
			),
			Handler:             handleStats,
			Category:            "federation",
			CircuitBreakerGroup: "federation",
		},
	}
}

func handleListKnownServers(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	category := tools.GetStringParam(request, "category")

	client := clients.GetMCPFederationClient()
	servers := client.KnownServers()

	var filtered []clients.KnownMCPServer
	for _, s := range servers {
		if category == "" || strings.EqualFold(s.Category, category) {
			filtered = append(filtered, s)
		}
	}

	result := map[string]interface{}{
		"count":   len(filtered),
		"servers": filtered,
	}

	return tools.JSONResult(result), nil
}

func handleConnect(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(request, "name")
	if errResult != nil {
		return errResult, nil
	}
	url, errResult := tools.RequireStringParam(request, "url")
	if errResult != nil {
		return errResult, nil
	}
	transport := tools.GetStringParam(request, "transport")
	if transport == "" {
		transport = "http"
	}

	client := clients.GetMCPFederationClient()
	server, err := client.Connect(ctx, name, url, transport)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	result := map[string]interface{}{
		"status":      server.Status,
		"name":        server.Name,
		"url":         server.URL,
		"tools_count": len(server.Tools),
	}

	if server.Status == "error" {
		result["error"] = server.Error
	}

	return tools.JSONResult(result), nil
}

func handleDisconnect(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(request, "name")
	if errResult != nil {
		return errResult, nil
	}

	client := clients.GetMCPFederationClient()
	if err := client.Disconnect(name); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Disconnected from %s", name)), nil
}

func handleListConnections(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetMCPFederationClient()
	connections := client.ListConnections()

	var servers []map[string]interface{}
	for _, c := range connections {
		servers = append(servers, map[string]interface{}{
			"name":        c.Name,
			"url":         c.URL,
			"status":      c.Status,
			"tools_count": len(c.Tools),
			"last_ping":   c.LastPing.Format("2006-01-02 15:04:05"),
		})
	}

	result := map[string]interface{}{
		"count":       len(servers),
		"connections": servers,
	}

	return tools.JSONResult(result), nil
}

func handleListTools(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	server := tools.GetStringParam(request, "server")
	search := tools.GetStringParam(request, "search")

	client := clients.GetMCPFederationClient()

	var allTools []clients.FederatedTool
	if search != "" {
		allTools = client.SearchTools(search)
	} else {
		allTools = client.ListAllTools()
	}

	// Filter by server if specified
	var filtered []clients.FederatedTool
	for _, t := range allTools {
		if server == "" || strings.EqualFold(t.Server, server) {
			filtered = append(filtered, t)
		}
	}

	result := map[string]interface{}{
		"count": len(filtered),
		"tools": filtered,
	}

	return tools.JSONResult(result), nil
}

func handleCallTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	serverName, errResult := tools.RequireStringParam(request, "server")
	if errResult != nil {
		return errResult, nil
	}
	toolName, errResult := tools.RequireStringParam(request, "tool")
	if errResult != nil {
		return errResult, nil
	}

	var args map[string]interface{}
	if rawArgs, ok := request.Params.Arguments.(map[string]interface{}); ok {
		if a, ok := rawArgs["args"].(map[string]interface{}); ok {
			args = a
		}
	}

	client := clients.GetMCPFederationClient()
	result, err := client.CallTool(ctx, serverName, toolName, args)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	output := map[string]interface{}{
		"success":    result.Success,
		"latency_ms": result.Latency,
	}

	if result.Success {
		output["result"] = result.Result
	} else {
		output["error"] = result.Error
	}

	return tools.JSONResult(output), nil
}

func handlePing(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(request, "name")
	if errResult != nil {
		return errResult, nil
	}

	client := clients.GetMCPFederationClient()
	ok, latency, err := client.Ping(ctx, name)

	result := map[string]interface{}{
		"server":     name,
		"reachable":  ok,
		"latency_ms": latency.Milliseconds(),
	}

	if err != nil {
		result["error"] = err.Error()
	}

	return tools.JSONResult(result), nil
}

func handleRefresh(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetMCPFederationClient()
	if err := client.RefreshAll(ctx); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	stats := client.GetStats()
	result := map[string]interface{}{
		"status":            "refreshed",
		"connected_servers": stats.ConnectedServers,
		"total_tools":       stats.TotalTools,
	}

	return tools.JSONResult(result), nil
}

func handleStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetMCPFederationClient()
	stats := client.GetStats()

	return tools.JSONResult(stats), nil
}
