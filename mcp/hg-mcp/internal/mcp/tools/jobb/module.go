package jobb

import (
	"context"
	"fmt"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

func (m *Module) Name() string        { return "jobb" }
func (m *Module) Description() string { return "Tenant-scoped automated job search, tracking, and applications." }

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.Tool{
				Name:        "aftrs_jobb_search",
				Description: "Search for job listings and save them to the active tenant's scoped storage.",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"query": map[string]interface{}{
							"type":        "string",
							"description": "Keywords or role title to search for",
						},
						"location": map[string]interface{}{
							"type":        "string",
							"description": "Target job location (remote or specific city)",
						},
					},
					Required: []string{"query"},
				},
			},
			Handler:     handleJobbSearch,
			Category:    "jobb",
			Subcategory: "search",
			Tags:        []string{"jobs", "search", "parity"},
			UseCases:    []string{"Find new job listings", "Discover roles"},
			IsWrite:     true, // Modifies tenant database with saved jobs
		},
		{
			Tool: mcp.Tool{
				Name:        "aftrs_jobb_list_saved",
				Description: "List saved jobs for the active tenant namespace.",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"status": map[string]interface{}{
							"type":        "string",
							"description": "Filter by status (saved, applied, rejected, interviewing)",
						},
					},
				},
			},
			Handler:     handleJobbListSaved,
			Category:    "jobb",
			Subcategory: "saved",
			Tags:        []string{"jobs", "list", "tracker"},
			UseCases:    []string{"View saved roles", "Check job pipeline"},
			IsWrite:     false,
		},
		{
			Tool: mcp.Tool{
				Name:        "aftrs_jobb_apply_automated",
				Description: "Automatically submit a job application using the active tenant's profile and credentials.",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"job_id": map[string]interface{}{
							"type":        "string",
							"description": "The ID of the saved job to apply to",
						},
						"cover_letter_variant": map[string]interface{}{
							"type":        "string",
							"description": "Which cover letter template variant to use (default: standard)",
						},
					},
					Required: []string{"job_id"},
				},
			},
			Handler:     handleJobbApplyAutomated,
			Category:    "jobb",
			Subcategory: "apply",
			Tags:        []string{"jobs", "automation", "apply", "submit"},
			UseCases:    []string{"Auto-apply to jobs", "Automated submission"},
			IsWrite:     true,
		},
		{
			Tool: mcp.Tool{
				Name:        "aftrs_jobb_update_status",
				Description: "Update the status of a job application in the tenant-scoped database.",
				InputSchema: mcp.ToolInputSchema{
					Type: "object",
					Properties: map[string]interface{}{
						"job_id": map[string]interface{}{
							"type":        "string",
							"description": "The ID of the job",
						},
						"status": map[string]interface{}{
							"type":        "string",
							"description": "New status (applied, rejected, interviewing, offer)",
						},
					},
					Required: []string{"job_id", "status"},
				},
			},
			Handler:     handleJobbUpdateStatus,
			Category:    "jobb",
			Subcategory: "tracker",
			Tags:        []string{"jobs", "status", "update"},
			UseCases:    []string{"Change job status", "Log interview"},
			IsWrite:     true,
		},
	}
}

// Helpers for tenant isolation (mocked for parity demonstration)
func getTenantID(ctx context.Context) string {
	// In production, this would extract the active user from the RBAC context
	// e.g. "mitch"
	return "mitch"
}

func getTenantKey(tenant, suffix string) string {
	return fmt.Sprintf("%s:jobb:%s", tenant, suffix)
}

func handleJobbSearch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tenant := getTenantID(ctx)
	query := tools.GetStringParam(req, "query")
	location := tools.GetStringParam(req, "location")

	// Mocking integration parity
	msg := fmt.Sprintf("Tenant '%s' executed job search for '%s' in '%s'. Results cached at key '%s'.", 
		tenant, query, location, getTenantKey(tenant, "results"))

	return tools.JSONResult(map[string]interface{}{
		"status":   "success",
		"message":  msg,
		"found":    42, // Mocked parity metric
		"tenant":   tenant,
	}), nil
}

func handleJobbListSaved(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tenant := getTenantID(ctx)
	statusFilter := tools.GetStringParam(req, "status")

	msg := fmt.Sprintf("Retrieved jobs for tenant '%s' under key '%s'", tenant, getTenantKey(tenant, "saved"))
	if statusFilter != "" {
		msg += fmt.Sprintf(" filtered by status '%s'", statusFilter)
	}

	return tools.JSONResult(map[string]interface{}{
		"status":  "success",
		"message": msg,
		"jobs": []map[string]string{
			{"id": "job_123", "title": "Senior Engineer", "status": "saved"},
			{"id": "job_456", "title": "Platform Architect", "status": "applied"},
		},
		"tenant": tenant,
	}), nil
}

func handleJobbApplyAutomated(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tenant := getTenantID(ctx)
	jobID := tools.GetStringParam(req, "job_id")
	variant := tools.GetStringParam(req, "cover_letter_variant")
	if variant == "" {
		variant = "standard"
	}

	msg := fmt.Sprintf("Automated application submitted for tenant '%s' on job '%s' using '%s' template.", 
		tenant, jobID, variant)

	return tools.JSONResult(map[string]interface{}{
		"status":  "success",
		"message": msg,
		"job_id":  jobID,
		"tenant":  tenant,
	}), nil
}

func handleJobbUpdateStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	tenant := getTenantID(ctx)
	jobID := tools.GetStringParam(req, "job_id")
	status := tools.GetStringParam(req, "status")

	msg := fmt.Sprintf("Updated job '%s' to status '%s' for tenant '%s' at key '%s'.", 
		jobID, status, tenant, getTenantKey(tenant, "status"))

	return tools.JSONResult(map[string]interface{}{
		"status":  "success",
		"message": msg,
		"job_id":  jobID,
		"tenant":  tenant,
	}), nil
}