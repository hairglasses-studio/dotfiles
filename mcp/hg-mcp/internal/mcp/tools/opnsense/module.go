// Package opnsense provides MCP tools for OPNsense firewall management
package opnsense

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewOPNsenseClient)

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Module implements the opnsense tools module
type Module struct{}

// Name returns the module name
func (m *Module) Name() string { return "opnsense" }

// Description returns the module description
func (m *Module) Description() string {
	return "OPNsense firewall management and network diagnostics"
}

// Tools returns all opnsense tools
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		// System tools
		{
			Tool:                mcp.NewTool("aftrs_opnsense_status", mcp.WithDescription("Get OPNsense firewall status including uptime, CPU, memory, and state table usage")),
			Handler:             handleStatus,
			Category:            "opnsense",
			Subcategory:         "system",
			Tags:                []string{"firewall", "status", "health", "opnsense"},
			UseCases:            []string{"Check firewall health", "Monitor system resources"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "opnsense",
		},
		{
			Tool:                mcp.NewTool("aftrs_opnsense_health", mcp.WithDescription("Get OPNsense health assessment with score and recommendations")),
			Handler:             handleHealth,
			Category:            "opnsense",
			Subcategory:         "system",
			Tags:                []string{"firewall", "health", "assessment"},
			UseCases:            []string{"Health check", "Get recommendations"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "opnsense",
		},

		// Firewall tools
		{
			Tool: mcp.NewTool("aftrs_opnsense_firewall_rules",
				mcp.WithDescription("List active firewall rules with filtering options"),
				mcp.WithString("interface", mcp.Description("Filter by interface (wan, lan, opt1, etc.)")),
				mcp.WithString("action", mcp.Description("Filter by action (pass, block, reject)")),
			),
			Handler:             handleFirewallRules,
			Category:            "opnsense",
			Subcategory:         "firewall",
			Tags:                []string{"firewall", "rules", "security"},
			UseCases:            []string{"Review firewall rules", "Security audit"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "opnsense",
		},
		{
			Tool: mcp.NewTool("aftrs_opnsense_firewall_states",
				mcp.WithDescription("View current connection state table"),
				mcp.WithNumber("limit", mcp.Description("Max number of states to return (default: 50)")),
			),
			Handler:             handleFirewallStates,
			Category:            "opnsense",
			Subcategory:         "firewall",
			Tags:                []string{"firewall", "states", "connections"},
			UseCases:            []string{"Monitor active connections", "Debug connectivity"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "opnsense",
		},

		// NAT tools
		{
			Tool:                mcp.NewTool("aftrs_opnsense_nat_rules", mcp.WithDescription("List NAT and port forward rules")),
			Handler:             handleNATRules,
			Category:            "opnsense",
			Subcategory:         "nat",
			Tags:                []string{"nat", "port-forward", "rules"},
			UseCases:            []string{"Review port forwards", "Check NAT configuration"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "opnsense",
		},

		// Network tools
		{
			Tool:                mcp.NewTool("aftrs_opnsense_interfaces", mcp.WithDescription("List network interfaces with status and traffic statistics")),
			Handler:             handleInterfaces,
			Category:            "opnsense",
			Subcategory:         "network",
			Tags:                []string{"interfaces", "network", "vlans"},
			UseCases:            []string{"Check interface status", "Monitor traffic"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "opnsense",
		},
		{
			Tool:                mcp.NewTool("aftrs_opnsense_routes", mcp.WithDescription("View routing table")),
			Handler:             handleRoutes,
			Category:            "opnsense",
			Subcategory:         "network",
			Tags:                []string{"routes", "routing", "network"},
			UseCases:            []string{"Check routing", "Debug connectivity"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "opnsense",
		},

		// Services tools
		{
			Tool:                mcp.NewTool("aftrs_opnsense_services", mcp.WithDescription("List services with running status")),
			Handler:             handleServices,
			Category:            "opnsense",
			Subcategory:         "services",
			Tags:                []string{"services", "status"},
			UseCases:            []string{"Check service status", "Monitor services"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "opnsense",
		},
		{
			Tool: mcp.NewTool("aftrs_opnsense_service_restart",
				mcp.WithDescription("Restart a whitelisted service (unbound, dhcpd, ntpd, openvpn, haproxy)"),
				mcp.WithString("service", mcp.Required(), mcp.Description("Service name to restart"), mcp.Enum("unbound", "dnsmasq", "dhcpd", "ntpd", "openvpn", "haproxy", "squid", "monit")),
			),
			Handler:             handleServiceRestart,
			Category:            "opnsense",
			Subcategory:         "services",
			Tags:                []string{"services", "restart"},
			UseCases:            []string{"Restart DNS", "Restart VPN"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "opnsense",
			IsWrite:             true,
		},

		// Logs tools
		{
			Tool: mcp.NewTool("aftrs_opnsense_logs",
				mcp.WithDescription("View recent firewall logs"),
				mcp.WithNumber("limit", mcp.Description("Number of log entries (default: 50)")),
			),
			Handler:             handleLogs,
			Category:            "opnsense",
			Subcategory:         "logs",
			Tags:                []string{"logs", "firewall", "audit"},
			UseCases:            []string{"Review blocked traffic", "Security monitoring"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "opnsense",
		},

		// Diagnostics tools
		{
			Tool: mcp.NewTool("aftrs_opnsense_ping",
				mcp.WithDescription("Ping a host from the firewall"),
				mcp.WithString("host", mcp.Required(), mcp.Description("Host to ping (IP or hostname)")),
				mcp.WithNumber("count", mcp.Description("Number of pings (default: 4)")),
			),
			Handler:             handlePing,
			Category:            "opnsense",
			Subcategory:         "diagnostics",
			Tags:                []string{"ping", "network", "diagnostics"},
			UseCases:            []string{"Test connectivity", "Check latency"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "opnsense",
		},
		{
			Tool: mcp.NewTool("aftrs_opnsense_traceroute",
				mcp.WithDescription("Traceroute to a host from the firewall"),
				mcp.WithString("host", mcp.Required(), mcp.Description("Host to trace (IP or hostname)")),
			),
			Handler:             handleTraceroute,
			Category:            "opnsense",
			Subcategory:         "diagnostics",
			Tags:                []string{"traceroute", "network", "diagnostics"},
			UseCases:            []string{"Trace network path", "Debug routing"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "opnsense",
		},

		// Backup tools
		{
			Tool:                mcp.NewTool("aftrs_opnsense_config_backup", mcp.WithDescription("Download configuration backup")),
			Handler:             handleConfigBackup,
			Category:            "opnsense",
			Subcategory:         "backup",
			Tags:                []string{"backup", "config", "disaster-recovery"},
			UseCases:            []string{"Backup configuration", "Disaster recovery"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "opnsense",
		},

		// Network overview (consolidated)
		{
			Tool:                mcp.NewTool("aftrs_opnsense_network_overview", mcp.WithDescription("Consolidated network overview: interfaces + routes + active connections")),
			Handler:             handleNetworkOverview,
			Category:            "opnsense",
			Subcategory:         "consolidated",
			Tags:                []string{"network", "overview", "consolidated"},
			UseCases:            []string{"Full network status", "Quick overview"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "opnsense",
		},
	}
}

// Handler implementations

func handleStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var sb strings.Builder
	sb.WriteString("# OPNsense Firewall Status\n\n")

	if status.Connected {
		sb.WriteString("**Status:** Connected\n\n")
		sb.WriteString(fmt.Sprintf("**Hostname:** %s\n", status.Hostname))
		sb.WriteString(fmt.Sprintf("**Version:** %s\n", status.Version))
		sb.WriteString(fmt.Sprintf("**Uptime:** %s\n", status.Uptime))
		sb.WriteString(fmt.Sprintf("**CPU Usage:** %.1f%%\n", status.CPUUsage))
		sb.WriteString(fmt.Sprintf("**Memory Usage:** %.1f%%\n", status.MemoryUsage))
		if status.PFStatesMax > 0 {
			sb.WriteString(fmt.Sprintf("**State Table:** %d / %d\n", status.PFStatesCount, status.PFStatesMax))
		}
	} else {
		sb.WriteString("**Status:** Disconnected\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleHealth(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	emoji := "?"
	switch health.Status {
	case "healthy":
		emoji = "?"
	case "degraded":
		emoji = "?"
	case "critical", "unavailable":
		emoji = "?"
	}

	var sb strings.Builder
	sb.WriteString("# OPNsense Health\n\n")
	sb.WriteString(fmt.Sprintf("**Score:** %d/100 %s\n", health.Score, emoji))
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

	return mcp.NewToolResultText(sb.String()), nil
}

func handleFirewallRules(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	iface := tools.GetStringParam(request, "interface")
	action := tools.GetStringParam(request, "action")

	client, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	rules, err := client.GetFirewallRules(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Filter rules
	var filtered []clients.OPNsenseFirewallRule
	for _, r := range rules {
		if iface != "" && !strings.EqualFold(r.Interface, iface) {
			continue
		}
		if action != "" && !strings.EqualFold(r.Action, action) {
			continue
		}
		filtered = append(filtered, r)
	}

	var sb strings.Builder
	sb.WriteString("# Firewall Rules\n\n")
	sb.WriteString(fmt.Sprintf("**Total:** %d rules\n\n", len(filtered)))

	sb.WriteString("| # | Enabled | Action | Interface | Protocol | Source | Destination | Port | Description |\n")
	sb.WriteString("|---|---------|--------|-----------|----------|--------|-------------|------|-------------|\n")

	for i, r := range filtered {
		enabled := "?"
		if r.Enabled {
			enabled = "?"
		}
		sb.WriteString(fmt.Sprintf("| %d | %s | %s | %s | %s | %s | %s | %s | %s |\n",
			i+1, enabled, r.Action, r.Interface, r.Protocol, r.Source, r.Destination, r.Port, r.Description))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleFirewallStates(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := tools.GetIntParam(request, "limit", 50)

	client, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	states, err := client.GetFirewallStates(ctx, limit)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"count":  len(states),
		"states": states,
	}), nil
}

func handleNATRules(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	rules, err := client.GetNATRules(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var sb strings.Builder
	sb.WriteString("# NAT / Port Forward Rules\n\n")
	sb.WriteString(fmt.Sprintf("**Total:** %d rules\n\n", len(rules)))

	sb.WriteString("| Enabled | Interface | Protocol | External | Internal | Description |\n")
	sb.WriteString("|---------|-----------|----------|----------|----------|-------------|\n")

	for _, r := range rules {
		enabled := "?"
		if r.Enabled {
			enabled = "?"
		}
		external := r.ExternalPort
		internal := fmt.Sprintf("%s:%s", r.InternalIP, r.InternalPort)
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
			enabled, r.Interface, r.Protocol, external, internal, r.Description))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleInterfaces(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	interfaces, err := client.GetInterfaces(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var sb strings.Builder
	sb.WriteString("# Network Interfaces\n\n")
	sb.WriteString(fmt.Sprintf("**Total:** %d interfaces\n\n", len(interfaces)))

	sb.WriteString("| Name | Device | Status | IP Address | Traffic In | Traffic Out |\n")
	sb.WriteString("|------|--------|--------|------------|------------|-------------|\n")

	for _, iface := range interfaces {
		status := "?"
		if iface.Status == "up" {
			status = "?"
		}
		bytesIn := formatBytes(iface.BytesIn)
		bytesOut := formatBytes(iface.BytesOut)
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
			iface.Name, iface.Device, status, iface.IPAddress, bytesIn, bytesOut))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleRoutes(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	routes, err := client.GetRoutes(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var sb strings.Builder
	sb.WriteString("# Routing Table\n\n")
	sb.WriteString(fmt.Sprintf("**Total:** %d routes\n\n", len(routes)))

	sb.WriteString("| Destination | Gateway | Flags | Interface |\n")
	sb.WriteString("|-------------|---------|-------|------------|\n")

	for _, r := range routes {
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			r.Destination, r.Gateway, r.Flags, r.Interface))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleServices(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	services, err := client.GetServices(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var sb strings.Builder
	sb.WriteString("# Services\n\n")

	running := 0
	for _, s := range services {
		if s.Running {
			running++
		}
	}
	sb.WriteString(fmt.Sprintf("**Running:** %d / %d services\n\n", running, len(services)))

	sb.WriteString("| Service | Status | Enabled | Description |\n")
	sb.WriteString("|---------|--------|---------|-------------|\n")

	for _, s := range services {
		status := "? Stopped"
		if s.Running {
			status = "? Running"
		}
		enabled := "No"
		if s.Enabled {
			enabled = "Yes"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
			s.Name, status, enabled, s.Description))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleServiceRestart(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	service, errResult := tools.RequireStringParam(request, "service")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	if err := client.RestartService(ctx, service); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Service '%s' restart initiated successfully", service)), nil
}

func handleLogs(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := tools.GetIntParam(request, "limit", 50)

	client, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	logs, err := client.GetLogs(ctx, limit)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var sb strings.Builder
	sb.WriteString("# Firewall Logs\n\n")
	sb.WriteString(fmt.Sprintf("**Showing:** %d entries\n\n", len(logs)))

	sb.WriteString("| Time | Action | Interface | Protocol | Source | Destination |\n")
	sb.WriteString("|------|--------|-----------|----------|--------|-------------|\n")

	for _, l := range logs {
		timestamp := l.Timestamp.Format("15:04:05")
		source := fmt.Sprintf("%s:%s", l.SourceIP, l.SourcePort)
		dest := fmt.Sprintf("%s:%s", l.DestIP, l.DestPort)
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
			timestamp, l.Action, l.Interface, l.Protocol, source, dest))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handlePing(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	host, errResult := tools.RequireStringParam(request, "host")
	if errResult != nil {
		return errResult, nil
	}
	count := tools.GetIntParam(request, "count", 4)

	client, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result, err := client.Ping(ctx, host, count)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return tools.JSONResult(result), nil
}

func handleTraceroute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	host, errResult := tools.RequireStringParam(request, "host")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	hops, err := client.Traceroute(ctx, host)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Traceroute to %s\n\n", host))

	for i, hop := range hops {
		sb.WriteString(fmt.Sprintf("%2d. %v\n", i+1, hop))
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleConfigBackup(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	data, err := client.BackupConfig(ctx)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Configuration backup created (%d bytes)\n\nNote: Full backup data available via API response", len(data))), nil
}

func handleNetworkOverview(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var sb strings.Builder
	sb.WriteString("# Network Overview\n\n")

	// Status
	status, _ := client.GetStatus(ctx)
	if status != nil && status.Connected {
		sb.WriteString("## System\n\n")
		sb.WriteString(fmt.Sprintf("**Hostname:** %s\n", status.Hostname))
		sb.WriteString(fmt.Sprintf("**Uptime:** %s\n", status.Uptime))
		sb.WriteString(fmt.Sprintf("**CPU:** %.1f%% | **Memory:** %.1f%%\n\n", status.CPUUsage, status.MemoryUsage))
	}

	// Interfaces
	interfaces, _ := client.GetInterfaces(ctx)
	if len(interfaces) > 0 {
		sb.WriteString("## Interfaces\n\n")
		sb.WriteString("| Name | IP Address | Status | Traffic |\n")
		sb.WriteString("|------|------------|--------|--------|\n")
		for _, iface := range interfaces {
			if !iface.Enabled {
				continue
			}
			status := "?"
			if iface.Status == "up" {
				status = "?"
			}
			traffic := fmt.Sprintf("? %s / ? %s", formatBytes(iface.BytesIn), formatBytes(iface.BytesOut))
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
				iface.Name, iface.IPAddress, status, traffic))
		}
		sb.WriteString("\n")
	}

	// Active states summary
	states, _ := client.GetFirewallStates(ctx, 1000)
	sb.WriteString("## Connection States\n\n")
	sb.WriteString(fmt.Sprintf("**Active connections:** %d\n\n", len(states)))

	return mcp.NewToolResultText(sb.String()), nil
}

// formatBytes formats bytes to human readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
