package security

import (
	"context"
	"os/user"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/hairglasses-studio/hg-mcp/pkg/secrets"
	"github.com/hairglasses-studio/hg-mcp/pkg/security"
)

func handleWhoami(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	identity, err := secrets.Whoami(ctx)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Add roles from RBAC
	identity.Roles = make([]string, 0)
	roles := security.GlobalRBAC.GetUserRoles(identity.User)
	for _, r := range roles {
		identity.Roles = append(identity.Roles, string(r))
	}

	// Return as structured map
	result := identity.ToMap()
	result["access"] = map[string]bool{
		"has_aws":        identity.HasAWSAccess(),
		"has_github":     identity.HasGitHubAccess(),
		"has_kubernetes": identity.HasKubernetesAccess(),
		"has_1password":  identity.Has1PasswordAccess(),
	}

	return tools.JSONResult(result), nil
}

func handleAuditLog(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := tools.GetIntParam(request, "limit", 50)
	filterUser := tools.GetStringParam(request, "user")
	filterTool := tools.GetStringParam(request, "tool")
	errorsOnly := tools.GetBoolParam(request, "errors_only", false)

	var events []security.AuditEvent

	if errorsOnly {
		events = security.GlobalAuditLogger.GetErrorEvents(limit)
	} else if filterUser != "" {
		events = security.GlobalAuditLogger.GetEventsByUser(filterUser, limit)
	} else if filterTool != "" {
		events = security.GlobalAuditLogger.GetEventsByTool(filterTool, limit)
	} else {
		events = security.GlobalAuditLogger.GetRecentEvents(limit)
	}

	return tools.JSONResult(map[string]interface{}{
		"count":  len(events),
		"events": events,
	}), nil
}

func handleAuditStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	stats := security.GlobalAuditLogger.GetStats()
	return tools.JSONResult(stats), nil
}

func handleListRoles(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	roles := security.GetAllRoles()
	return tools.JSONResult(map[string]interface{}{
		"roles": roles,
	}), nil
}

func handleAccessCheck(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	username := tools.GetStringParam(request, "user")
	if username == "" {
		if u, err := user.Current(); err == nil {
			username = u.Username
		}
	}

	toolName, errResult := tools.RequireStringParam(request, "tool")
	if errResult != nil {
		return errResult, nil
	}

	hasAccess := security.GlobalRBAC.CanAccessTool(username, toolName)
	requiredPerm := security.GlobalRBAC.GetToolPermission(toolName)

	result := map[string]interface{}{
		"user":                username,
		"tool":                toolName,
		"has_access":          hasAccess,
		"required_permission": requiredPerm,
		"user_roles":          security.GlobalRBAC.GetUserRoles(username),
	}

	return tools.JSONResult(result), nil
}

func handleUserAccess(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	username := tools.GetStringParam(request, "user")
	if username == "" {
		if u, err := user.Current(); err == nil {
			username = u.Username
		}
	}

	accessInfo := security.GlobalRBAC.GetUserAccess(username)
	return tools.JSONResult(accessInfo), nil
}
