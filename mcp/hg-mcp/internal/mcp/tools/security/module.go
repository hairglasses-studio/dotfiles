package security

import (
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Module implements the ToolModule interface for security tools
type Module struct{}

func (m *Module) Name() string {
	return "security"
}

func (m *Module) Description() string {
	return "Security and access control tools for identity, audit logging, and RBAC"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("security_whoami",
				mcp.WithDescription("Get current user identity including AWS, GitHub, and Kubernetes access info"),
			),
			Handler:             handleWhoami,
			Category:            "security",
			Subcategory:         "identity",
			Tags:                []string{"security", "identity", "whoami", "user", "access"},
			UseCases:            []string{"Check current user identity", "Verify credentials", "Debug access issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "security",
		},
		{
			Tool: mcp.NewTool("security_audit_log",
				mcp.WithDescription("View recent audit log entries for tool invocations"),
				mcp.WithNumber("limit",
					mcp.Description("Maximum number of entries to return (default: 50)"),
				),
				mcp.WithString("user",
					mcp.Description("Filter by username"),
				),
				mcp.WithString("tool",
					mcp.Description("Filter by tool name"),
				),
				mcp.WithBoolean("errors_only",
					mcp.Description("Show only error events"),
				),
			),
			Handler:             handleAuditLog,
			Category:            "security",
			Subcategory:         "audit",
			Tags:                []string{"security", "audit", "log", "history", "monitoring"},
			UseCases:            []string{"Review tool usage", "Debug issues", "Security audit"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "security",
		},
		{
			Tool: mcp.NewTool("security_audit_stats",
				mcp.WithDescription("Get summary statistics for audit events"),
			),
			Handler:             handleAuditStats,
			Category:            "security",
			Subcategory:         "audit",
			Tags:                []string{"security", "audit", "statistics", "metrics"},
			UseCases:            []string{"View tool usage metrics", "Monitor system activity"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "security",
		},
		{
			Tool: mcp.NewTool("security_roles",
				mcp.WithDescription("List all available roles and their permissions"),
			),
			Handler:             handleListRoles,
			Category:            "security",
			Subcategory:         "rbac",
			Tags:                []string{"security", "rbac", "roles", "permissions"},
			UseCases:            []string{"View available roles", "Understand permissions"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "security",
		},
		{
			Tool: mcp.NewTool("security_access_check",
				mcp.WithDescription("Check if a user has access to a specific tool"),
				mcp.WithString("user",
					mcp.Description("Username to check (defaults to current user)"),
				),
				mcp.WithString("tool",
					mcp.Description("Tool name to check access for"),
					mcp.Required(),
				),
			),
			Handler:             handleAccessCheck,
			Category:            "security",
			Subcategory:         "rbac",
			Tags:                []string{"security", "rbac", "access", "check"},
			UseCases:            []string{"Verify tool access", "Debug permission issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "security",
		},
		{
			Tool: mcp.NewTool("security_user_access",
				mcp.WithDescription("Get access information for a user including roles and permissions"),
				mcp.WithString("user",
					mcp.Description("Username to get access info for (defaults to current user)"),
				),
			),
			Handler:             handleUserAccess,
			Category:            "security",
			Subcategory:         "rbac",
			Tags:                []string{"security", "rbac", "user", "access", "permissions"},
			UseCases:            []string{"View user permissions", "Audit user access"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "security",
		},
	}
}
