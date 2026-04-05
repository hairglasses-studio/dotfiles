// Package dante provides MCP tools for Dante audio network management.
package dante

import (
	"context"
	"fmt"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Module implements the Dante tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "dante"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Dante audio network discovery, routing, and monitoring"
}

// Tools returns the Dante tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_dante_status",
				mcp.WithDescription("Get Dante network status including device count and master clock"),
			),
			Handler:             handleDanteStatus,
			Category:            "dante",
			Subcategory:         "status",
			Tags:                []string{"dante", "audio", "network", "status"},
			UseCases:            []string{"Check network status", "Verify Dante connectivity", "View master clock"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "dante",
		},
		{
			Tool: mcp.NewTool("aftrs_dante_discover",
				mcp.WithDescription("Discover Dante devices on the network using mDNS"),
				mcp.WithNumber("timeout",
					mcp.Description("Discovery timeout in seconds (default: 3)"),
				),
			),
			Handler:             handleDanteDiscover,
			Category:            "dante",
			Subcategory:         "discovery",
			Tags:                []string{"dante", "discover", "mdns", "devices"},
			UseCases:            []string{"Find Dante devices", "Refresh device list", "Network scan"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "dante",
		},
		{
			Tool: mcp.NewTool("aftrs_dante_devices",
				mcp.WithDescription("List all known Dante devices with channels and status"),
			),
			Handler:             handleDanteDevices,
			Category:            "dante",
			Subcategory:         "devices",
			Tags:                []string{"dante", "devices", "list"},
			UseCases:            []string{"List all devices", "View device channels", "Check device status"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "dante",
		},
		{
			Tool: mcp.NewTool("aftrs_dante_routes",
				mcp.WithDescription("List all audio routes between Dante devices"),
			),
			Handler:             handleDanteRoutes,
			Category:            "dante",
			Subcategory:         "routing",
			Tags:                []string{"dante", "routes", "routing", "audio"},
			UseCases:            []string{"View audio routing", "Check channel connections", "Audit routing"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "dante",
		},
		{
			Tool: mcp.NewTool("aftrs_dante_route",
				mcp.WithDescription("Create or delete an audio route between Dante devices"),
				mcp.WithString("action",
					mcp.Required(),
					mcp.Description("Action to perform: create or delete"),
					mcp.Enum("create", "delete"),
				),
				mcp.WithString("tx_device",
					mcp.Description("Transmitting device name"),
				),
				mcp.WithNumber("tx_channel",
					mcp.Description("Transmitting channel number"),
				),
				mcp.WithString("rx_device",
					mcp.Required(),
					mcp.Description("Receiving device name"),
				),
				mcp.WithNumber("rx_channel",
					mcp.Required(),
					mcp.Description("Receiving channel number"),
				),
			),
			Handler:             handleDanteRoute,
			Category:            "dante",
			Subcategory:         "routing",
			Tags:                []string{"dante", "route", "routing", "patch"},
			UseCases:            []string{"Create audio route", "Delete audio route", "Patch channels"},
			Complexity:          tools.ComplexityComplex,
			CircuitBreakerGroup: "dante",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_dante_latency",
				mcp.WithDescription("Check Dante network latency and performance metrics"),
			),
			Handler:             handleDanteLatency,
			Category:            "dante",
			Subcategory:         "monitoring",
			Tags:                []string{"dante", "latency", "performance", "monitoring"},
			UseCases:            []string{"Check network latency", "Monitor performance", "Diagnose issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "dante",
		},
		{
			Tool: mcp.NewTool("aftrs_dante_health",
				mcp.WithDescription("Check Dante network health and get troubleshooting recommendations"),
			),
			Handler:             handleDanteHealth,
			Category:            "dante",
			Subcategory:         "status",
			Tags:                []string{"dante", "health", "diagnostics", "troubleshooting"},
			UseCases:            []string{"Diagnose network issues", "Check device health", "Get troubleshooting tips"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "dante",
		},
	}
}

var getDanteClient = tools.LazyClient(clients.NewDanteClient)

func handleDanteStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getDanteClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Dante client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	return tools.JSONResult(status), nil
}

func handleDanteDiscover(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getDanteClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Dante client: %w", err)), nil
	}

	timeoutSec := tools.GetIntParam(req, "timeout", 3)
	timeout := time.Duration(timeoutSec) * time.Second

	devices, err := client.DiscoverDevices(ctx, timeout)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to discover devices: %w", err)), nil
	}

	result := map[string]interface{}{
		"devices": devices,
		"count":   len(devices),
		"timeout": timeoutSec,
		"message": fmt.Sprintf("Discovered %d Dante device(s)", len(devices)),
	}
	return tools.JSONResult(result), nil
}

func handleDanteDevices(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getDanteClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Dante client: %w", err)), nil
	}

	devices, err := client.GetDevices(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get devices: %w", err)), nil
	}

	result := map[string]interface{}{
		"devices": devices,
		"count":   len(devices),
	}
	return tools.JSONResult(result), nil
}

func handleDanteRoutes(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getDanteClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Dante client: %w", err)), nil
	}

	routes, err := client.GetRoutes(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get routes: %w", err)), nil
	}

	result := map[string]interface{}{
		"routes": routes,
		"count":  len(routes),
	}
	return tools.JSONResult(result), nil
}

func handleDanteRoute(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getDanteClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Dante client: %w", err)), nil
	}

	action, errResult := tools.RequireStringParam(req, "action")
	if errResult != nil {
		return errResult, nil
	}

	rxDevice, errResult := tools.RequireStringParam(req, "rx_device")
	if errResult != nil {
		return errResult, nil
	}

	rxChannel, errResult := tools.RequireIntParam(req, "rx_channel")
	if errResult != nil {
		return errResult, nil
	}

	switch action {
	case "create":
		txDevice, errResult := tools.RequireStringParam(req, "tx_device")
		if errResult != nil {
			return errResult, nil
		}

		txChannel, errResult := tools.RequireIntParam(req, "tx_channel")
		if errResult != nil {
			return errResult, nil
		}

		route, err := client.CreateRoute(ctx, txDevice, txChannel, rxDevice, rxChannel)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to create route: %w", err)), nil
		}

		result := map[string]interface{}{
			"success": true,
			"action":  "create",
			"route":   route,
			"message": fmt.Sprintf("Route created: %s", route.ID),
		}
		return tools.JSONResult(result), nil

	case "delete":
		if err := client.DeleteRoute(ctx, rxDevice, rxChannel); err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to delete route: %w", err)), nil
		}

		result := map[string]interface{}{
			"success":    true,
			"action":     "delete",
			"rx_device":  rxDevice,
			"rx_channel": rxChannel,
			"message":    fmt.Sprintf("Route deleted for %s channel %d", rxDevice, rxChannel),
		}
		return tools.JSONResult(result), nil

	default:
		return tools.ErrorResult(fmt.Errorf("invalid action: %s", action)), nil
	}
}

func handleDanteLatency(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getDanteClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Dante client: %w", err)), nil
	}

	latency, err := client.GetLatency(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get latency: %w", err)), nil
	}

	return tools.JSONResult(latency), nil
}

func handleDanteHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getDanteClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Dante client: %w", err)), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get health: %w", err)), nil
	}

	return tools.JSONResult(health), nil
}
