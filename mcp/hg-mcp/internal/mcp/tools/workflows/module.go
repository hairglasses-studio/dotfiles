// Package workflows provides automation workflow tools for hg-mcp.
package workflows

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewWorkflowsClient)

// Module implements the ToolModule interface for workflow tools
type Module struct{}

func (m *Module) Name() string {
	return "workflows"
}

func (m *Module) Description() string {
	return "Automation workflow tools for show management and studio operations"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_workflow_list",
				mcp.WithDescription("List all available workflows. Optionally filter by category (show, backup, maintenance, test)."),
				mcp.WithString("category", mcp.Description("Filter by category: show, backup, maintenance, test")),
			),
			Handler:             handleWorkflowList,
			Category:            "workflows",
			Subcategory:         "management",
			Tags:                []string{"workflow", "list", "automation"},
			UseCases:            []string{"See available workflows", "Browse automation options"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "workflows",
		},
		{
			Tool: mcp.NewTool("aftrs_workflow_run",
				mcp.WithDescription("Execute a workflow by ID. Returns execution status and can be monitored with aftrs_workflow_status."),
				mcp.WithString("workflow_id", mcp.Description("ID of the workflow to run"), mcp.Required()),
				mcp.WithBoolean("dry_run", mcp.Description("If true, simulate without executing")),
			),
			Handler:             handleWorkflowRun,
			Category:            "workflows",
			Subcategory:         "execution",
			Tags:                []string{"workflow", "run", "execute", "automation"},
			UseCases:            []string{"Run automation sequences", "Execute predefined workflows"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "workflows",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_workflow_status",
				mcp.WithDescription("Get the status of a workflow execution."),
				mcp.WithString("execution_id", mcp.Description("ID of the execution to check"), mcp.Required()),
			),
			Handler:             handleWorkflowStatus,
			Category:            "workflows",
			Subcategory:         "monitoring",
			Tags:                []string{"workflow", "status", "monitor"},
			UseCases:            []string{"Check workflow progress", "Monitor execution"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "workflows",
		},
		{
			Tool: mcp.NewTool("aftrs_show_startup",
				mcp.WithDescription("Execute the automated show startup sequence. Brings all systems online in the correct order."),
			),
			Handler:             handleShowStartup,
			Category:            "workflows",
			Subcategory:         "shows",
			Tags:                []string{"workflow", "show", "startup", "automation"},
			UseCases:            []string{"Start a show", "Initialize all systems"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "workflows",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_show_shutdown",
				mcp.WithDescription("Execute the graceful show shutdown sequence. Safely powers down all systems."),
			),
			Handler:             handleShowShutdown,
			Category:            "workflows",
			Subcategory:         "shows",
			Tags:                []string{"workflow", "show", "shutdown", "automation"},
			UseCases:            []string{"End a show", "Safe shutdown"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "workflows",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_backup_all",
				mcp.WithDescription("Backup all project files to NAS. Includes TouchDesigner, Resolume, and vault."),
			),
			Handler:             handleBackupAll,
			Category:            "workflows",
			Subcategory:         "backup",
			Tags:                []string{"workflow", "backup", "automation"},
			UseCases:            []string{"Full backup", "Archive projects"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "workflows",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_sync_assets",
				mcp.WithDescription("Sync project assets to NAS storage."),
			),
			Handler:             handleSyncAssets,
			Category:            "workflows",
			Subcategory:         "backup",
			Tags:                []string{"workflow", "sync", "assets", "automation"},
			UseCases:            []string{"Sync media", "Update NAS"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "workflows",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_cue_sequence",
				mcp.WithDescription("Execute a sequence of cues. Provide cues as comma-separated 'number:name:action' items."),
				mcp.WithString("cues", mcp.Description("Comma-separated cues in format 'number:name:action' (e.g., '1:Intro:play,2:Main:fade')"), mcp.Required()),
			),
			Handler:             handleCueSequence,
			Category:            "workflows",
			Subcategory:         "shows",
			Tags:                []string{"workflow", "cue", "sequence", "automation"},
			UseCases:            []string{"Run cue list", "Automate show sections"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "workflows",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_panic_mode",
				mcp.WithDescription("EMERGENCY: Immediately stops all outputs - lighting blackout, stops playback, mutes audio, sends alert."),
			),
			Handler:             handlePanicMode,
			Category:            "workflows",
			Subcategory:         "emergency",
			Tags:                []string{"workflow", "panic", "emergency", "stop"},
			UseCases:            []string{"Emergency stop", "Crisis response"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "workflows",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_test_sequence",
				mcp.WithDescription("Test all systems in sequence. Validates network, software, lighting, and streaming."),
			),
			Handler:             handleTestSequence,
			Category:            "workflows",
			Subcategory:         "test",
			Tags:                []string{"workflow", "test", "verify", "automation"},
			UseCases:            []string{"System verification", "Pre-show test"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "workflows",
		},
		// Phase 16C: Enhanced Workflow Tools
		{
			Tool: mcp.NewTool("aftrs_workflow_compose",
				mcp.WithDescription("Create a workflow from a tool sequence. Use -> for sequential steps, comma for parallel. Example: 'tool1 -> tool2, tool3 -> tool4'"),
				mcp.WithString("name", mcp.Required(), mcp.Description("Workflow name")),
				mcp.WithString("description", mcp.Description("Workflow description")),
				mcp.WithString("sequence", mcp.Required(), mcp.Description("Tool sequence: 'tool1 -> tool2' (sequential) or 'tool1, tool2' (parallel)")),
			),
			Handler:             handleWorkflowCompose,
			Category:            "workflows",
			Subcategory:         "composition",
			Tags:                []string{"workflow", "compose", "create", "automation"},
			UseCases:            []string{"Create custom workflows", "Build automation sequences"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "workflows",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_workflow_validate",
				mcp.WithDescription("Validate a workflow for correctness before execution. Checks dependencies, cycles, and required fields."),
				mcp.WithString("workflow_id", mcp.Required(), mcp.Description("ID of the workflow to validate")),
			),
			Handler:             handleWorkflowValidate,
			Category:            "workflows",
			Subcategory:         "management",
			Tags:                []string{"workflow", "validate", "check", "verify"},
			UseCases:            []string{"Verify workflow correctness", "Debug workflow issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "workflows",
		},
		{
			Tool: mcp.NewTool("aftrs_workflow_condition",
				mcp.WithDescription("Add a conditional execution rule to a workflow step. Step only runs if condition is met."),
				mcp.WithString("workflow_id", mcp.Required(), mcp.Description("ID of the workflow")),
				mcp.WithString("step_id", mcp.Required(), mcp.Description("ID of the step to add condition to")),
				mcp.WithString("variable", mcp.Required(), mcp.Description("Variable to check (e.g., 'step1.success', 'env.SHOW_TYPE')")),
				mcp.WithString("operator", mcp.Required(), mcp.Description("Comparison: eq, ne, gt, lt, gte, lte, contains, exists")),
				mcp.WithString("value", mcp.Description("Value to compare against (not needed for 'exists')")),
			),
			Handler:             handleWorkflowCondition,
			Category:            "workflows",
			Subcategory:         "composition",
			Tags:                []string{"workflow", "condition", "if", "branch"},
			UseCases:            []string{"Add conditional logic", "Create branching workflows"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "workflows",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_workflow_delete",
				mcp.WithDescription("Delete a custom workflow. Built-in workflows cannot be deleted."),
				mcp.WithString("workflow_id", mcp.Required(), mcp.Description("ID of the workflow to delete")),
			),
			Handler:             handleWorkflowDelete,
			Category:            "workflows",
			Subcategory:         "management",
			Tags:                []string{"workflow", "delete", "remove"},
			UseCases:            []string{"Remove custom workflows", "Clean up old workflows"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "workflows",
			IsWrite:             true,
		},
	}
}

// handleWorkflowList handles the aftrs_workflow_list tool
func handleWorkflowList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	category := tools.GetStringParam(req, "category")

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create workflows client: %w", err)), nil
	}

	workflows, err := client.ListWorkflows(ctx, category)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to list workflows: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Available Workflows\n\n")

	if category != "" {
		sb.WriteString(fmt.Sprintf("**Category:** %s\n\n", category))
	}

	sb.WriteString(fmt.Sprintf("Found **%d** workflows:\n\n", len(workflows)))

	if len(workflows) == 0 {
		sb.WriteString("No workflows found.\n")
	} else {
		sb.WriteString("| ID | Name | Category | Steps | Timeout |\n")
		sb.WriteString("|----|------|----------|-------|----------|\n")

		for _, w := range workflows {
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %d | %ds |\n",
				w.ID, w.Name, w.Category, len(w.Steps), w.Timeout))
		}

		sb.WriteString("\n## Workflow Details\n\n")

		for _, w := range workflows {
			sb.WriteString(fmt.Sprintf("### %s\n", w.Name))
			sb.WriteString(fmt.Sprintf("%s\n\n", w.Description))
			sb.WriteString("**Steps:**\n")
			for i, step := range w.Steps {
				deps := ""
				if len(step.DependsOn) > 0 {
					deps = fmt.Sprintf(" (after: %s)", strings.Join(step.DependsOn, ", "))
				}
				sb.WriteString(fmt.Sprintf("%d. %s%s\n", i+1, step.Name, deps))
			}
			sb.WriteString("\n")
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleWorkflowRun handles the aftrs_workflow_run tool
func handleWorkflowRun(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workflowID := tools.GetStringParam(req, "workflow_id")
	dryRun := tools.GetBoolParam(req, "dry_run", false)

	if workflowID == "" {
		return tools.ErrorResult(errors.New("workflow_id parameter is required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create workflows client: %w", err)), nil
	}

	exec, err := client.RunWorkflow(ctx, workflowID, dryRun)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to run workflow: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Workflow Started\n\n")

	if dryRun {
		sb.WriteString("**Mode:** Dry Run (simulation)\n\n")
	}

	sb.WriteString(fmt.Sprintf("**Execution ID:** `%s`\n", exec.ID))
	sb.WriteString(fmt.Sprintf("**Workflow:** %s\n", exec.WorkflowID))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n", exec.Status))
	sb.WriteString(fmt.Sprintf("**Started:** %s\n\n", exec.StartedAt.Format("2006-01-02 15:04:05")))

	sb.WriteString("Use `aftrs_workflow_status` with the execution ID to monitor progress.\n")

	return tools.TextResult(sb.String()), nil
}

// handleWorkflowStatus handles the aftrs_workflow_status tool
func handleWorkflowStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	executionID := tools.GetStringParam(req, "execution_id")

	if executionID == "" {
		return tools.ErrorResult(errors.New("execution_id parameter is required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create workflows client: %w", err)), nil
	}

	exec, err := client.GetExecutionStatus(ctx, executionID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Workflow Execution Status\n\n")

	emoji := getStatusEmoji(exec.Status)
	sb.WriteString(fmt.Sprintf("**Execution ID:** `%s`\n", exec.ID))
	sb.WriteString(fmt.Sprintf("**Workflow:** %s\n", exec.WorkflowID))
	sb.WriteString(fmt.Sprintf("**Status:** %s %s\n", emoji, exec.Status))
	sb.WriteString(fmt.Sprintf("**Progress:** %d%%\n", exec.Progress))
	sb.WriteString(fmt.Sprintf("**Started:** %s\n", exec.StartedAt.Format("2006-01-02 15:04:05")))

	if !exec.CompletedAt.IsZero() {
		duration := exec.CompletedAt.Sub(exec.StartedAt)
		sb.WriteString(fmt.Sprintf("**Completed:** %s\n", exec.CompletedAt.Format("2006-01-02 15:04:05")))
		sb.WriteString(fmt.Sprintf("**Duration:** %s\n", duration.Round(time.Millisecond)))
	}

	if exec.Error != "" {
		sb.WriteString(fmt.Sprintf("\n**Error:** %s\n", exec.Error))
	}

	sb.WriteString("\n## Step Results\n\n")
	sb.WriteString("| Step | Status | Duration |\n")
	sb.WriteString("|------|--------|----------|\n")

	for _, step := range exec.StepResults {
		stepEmoji := getStatusEmoji(step.Status)
		duration := "-"
		if !step.CompletedAt.IsZero() {
			duration = step.CompletedAt.Sub(step.StartedAt).Round(time.Millisecond).String()
		}
		sb.WriteString(fmt.Sprintf("| %s | %s %s | %s |\n", step.StepID, stepEmoji, step.Status, duration))
	}

	return tools.TextResult(sb.String()), nil
}

// handleShowStartup handles the aftrs_show_startup tool
func handleShowStartup(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create workflows client: %w", err)), nil
	}

	exec, err := client.ExecuteShowStartup(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to start show startup: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Show Startup Initiated\n\n")
	sb.WriteString("Starting automated show startup sequence...\n\n")
	sb.WriteString(fmt.Sprintf("**Execution ID:** `%s`\n", exec.ID))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n\n", exec.Status))

	sb.WriteString("## Sequence\n")
	sb.WriteString("1. Check Network\n")
	sb.WriteString("2. Verify UNRAID\n")
	sb.WriteString("3. Check TouchDesigner\n")
	sb.WriteString("4. Check Resolume\n")
	sb.WriteString("5. Initialize Lighting\n")
	sb.WriteString("6. Verify NDI Sources\n")
	sb.WriteString("7. Run Preflight Check\n")

	return tools.TextResult(sb.String()), nil
}

// handleShowShutdown handles the aftrs_show_shutdown tool
func handleShowShutdown(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create workflows client: %w", err)), nil
	}

	exec, err := client.ExecuteShowShutdown(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to start show shutdown: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Show Shutdown Initiated\n\n")
	sb.WriteString("Starting graceful show shutdown sequence...\n\n")
	sb.WriteString(fmt.Sprintf("**Execution ID:** `%s`\n", exec.ID))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n\n", exec.Status))

	sb.WriteString("## Sequence\n")
	sb.WriteString("1. Fade Lighting\n")
	sb.WriteString("2. Stop Resolume Playback\n")
	sb.WriteString("3. Backup TouchDesigner Project\n")
	sb.WriteString("4. Log Show End\n")

	return tools.TextResult(sb.String()), nil
}

// handleBackupAll handles the aftrs_backup_all tool
func handleBackupAll(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create workflows client: %w", err)), nil
	}

	exec, err := client.ExecuteBackupAll(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to start backup: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Backup All Initiated\n\n")
	sb.WriteString("Starting full backup to NAS...\n\n")
	sb.WriteString(fmt.Sprintf("**Execution ID:** `%s`\n", exec.ID))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n\n", exec.Status))

	sb.WriteString("## Backing Up\n")
	sb.WriteString("1. TouchDesigner projects\n")
	sb.WriteString("2. Resolume compositions\n")
	sb.WriteString("3. Vault documents\n")
	sb.WriteString("4. Verify backups\n")

	return tools.TextResult(sb.String()), nil
}

// handleSyncAssets handles the aftrs_sync_assets tool
func handleSyncAssets(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create workflows client: %w", err)), nil
	}

	exec, err := client.ExecuteSyncAssets(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to start sync: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Asset Sync Initiated\n\n")
	sb.WriteString("Starting asset sync to NAS...\n\n")
	sb.WriteString(fmt.Sprintf("**Execution ID:** `%s`\n", exec.ID))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n\n", exec.Status))

	sb.WriteString("## Syncing\n")
	sb.WriteString("1. Check NAS availability\n")
	sb.WriteString("2. Sync media files\n")
	sb.WriteString("3. Verify sync complete\n")

	return tools.TextResult(sb.String()), nil
}

// handleCueSequence handles the aftrs_cue_sequence tool
func handleCueSequence(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	cuesStr := tools.GetStringParam(req, "cues")

	if cuesStr == "" {
		return tools.ErrorResult(errors.New("cues parameter is required")), nil
	}

	// Parse cues from string format "number:name:action,number:name:action"
	cueStrings := strings.Split(cuesStr, ",")
	var cues []clients.Cue

	for _, cs := range cueStrings {
		parts := strings.Split(strings.TrimSpace(cs), ":")
		if len(parts) < 3 {
			continue
		}
		cues = append(cues, clients.Cue{
			Number: parts[0],
			Name:   parts[1],
			Action: parts[2],
		})
	}

	if len(cues) == 0 {
		return tools.ErrorResult(errors.New("no valid cues provided")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create workflows client: %w", err)), nil
	}

	exec, err := client.ExecuteCueSequence(ctx, cues)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to start cue sequence: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Cue Sequence Initiated\n\n")
	sb.WriteString(fmt.Sprintf("**Execution ID:** `%s`\n", exec.ID))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n\n", exec.Status))

	sb.WriteString("## Cues\n")
	for i, cue := range cues {
		sb.WriteString(fmt.Sprintf("%d. **%s** (%s): %s\n", i+1, cue.Number, cue.Name, cue.Action))
	}

	return tools.TextResult(sb.String()), nil
}

// handlePanicMode handles the aftrs_panic_mode tool
func handlePanicMode(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create workflows client: %w", err)), nil
	}

	exec, err := client.ExecutePanicMode(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to execute panic mode: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# PANIC MODE ACTIVATED\n\n")
	sb.WriteString("**EMERGENCY STOP INITIATED**\n\n")
	sb.WriteString(fmt.Sprintf("**Execution ID:** `%s`\n", exec.ID))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n\n", exec.Status))

	sb.WriteString("## Actions Taken\n")
	sb.WriteString("1. Lighting Blackout\n")
	sb.WriteString("2. All Playback Stopped\n")
	sb.WriteString("3. Audio Muted\n")
	sb.WriteString("4. Alert Sent\n")

	return tools.TextResult(sb.String()), nil
}

// handleTestSequence handles the aftrs_test_sequence tool
func handleTestSequence(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create workflows client: %w", err)), nil
	}

	exec, err := client.ExecuteTestSequence(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to start test sequence: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Test Sequence Initiated\n\n")
	sb.WriteString("Starting full system test...\n\n")
	sb.WriteString(fmt.Sprintf("**Execution ID:** `%s`\n", exec.ID))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n\n", exec.Status))

	sb.WriteString("## Tests\n")
	sb.WriteString("1. Network connectivity\n")
	sb.WriteString("2. TouchDesigner status\n")
	sb.WriteString("3. Resolume status\n")
	sb.WriteString("4. Lighting status\n")
	sb.WriteString("5. NDI sources\n")
	sb.WriteString("6. UNRAID status\n")
	sb.WriteString("7. Full health check\n")

	return tools.TextResult(sb.String()), nil
}

// Helper function
func getStatusEmoji(status string) string {
	switch status {
	case "completed":
		return "✅"
	case "running":
		return "🔄"
	case "failed":
		return "❌"
	case "cancelled":
		return "⚪"
	default:
		return "⏳"
	}
}

// handleWorkflowCompose handles the aftrs_workflow_compose tool
func handleWorkflowCompose(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name := tools.GetStringParam(req, "name")
	description := tools.GetStringParam(req, "description")
	sequence := tools.GetStringParam(req, "sequence")

	if name == "" || sequence == "" {
		return tools.ErrorResult(errors.New("name and sequence parameters are required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create workflows client: %w", err)), nil
	}

	workflow, err := client.ComposeWorkflow(ctx, name, description, sequence)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to compose workflow: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Workflow Created\n\n")
	sb.WriteString(fmt.Sprintf("**ID:** `%s`\n", workflow.ID))
	sb.WriteString(fmt.Sprintf("**Name:** %s\n", workflow.Name))
	if workflow.Description != "" {
		sb.WriteString(fmt.Sprintf("**Description:** %s\n", workflow.Description))
	}
	sb.WriteString(fmt.Sprintf("**Steps:** %d\n\n", len(workflow.Steps)))

	sb.WriteString("## Steps\n\n")
	for i, step := range workflow.Steps {
		parallel := ""
		if step.ParallelGroup != "" {
			parallel = fmt.Sprintf(" [%s]", step.ParallelGroup)
		}
		deps := ""
		if len(step.DependsOn) > 0 {
			deps = fmt.Sprintf(" (after: %s)", strings.Join(step.DependsOn, ", "))
		}
		sb.WriteString(fmt.Sprintf("%d. **%s**%s: `%s`%s\n", i+1, step.ID, parallel, step.Tool, deps))
	}

	sb.WriteString("\n## Usage\n")
	sb.WriteString(fmt.Sprintf("Run with: `aftrs_workflow_run workflow_id:\"%s\"`\n", workflow.ID))
	sb.WriteString(fmt.Sprintf("Validate with: `aftrs_workflow_validate workflow_id:\"%s\"`\n", workflow.ID))

	return tools.TextResult(sb.String()), nil
}

// handleWorkflowValidate handles the aftrs_workflow_validate tool
func handleWorkflowValidate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workflowID := tools.GetStringParam(req, "workflow_id")

	if workflowID == "" {
		return tools.ErrorResult(errors.New("workflow_id parameter is required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create workflows client: %w", err)), nil
	}

	validation, err := client.ValidateWorkflow(ctx, workflowID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to validate workflow: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Workflow Validation\n\n")
	sb.WriteString(fmt.Sprintf("**Workflow ID:** `%s`\n", validation.WorkflowID))

	if validation.IsValid {
		sb.WriteString("**Status:** ✅ Valid\n\n")
	} else {
		sb.WriteString("**Status:** ❌ Invalid\n\n")
	}

	sb.WriteString("## Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Steps:** %d\n", validation.StepCount))
	sb.WriteString(fmt.Sprintf("- **Estimated Duration:** %ds\n", validation.EstimatedDuration))

	if len(validation.Issues) > 0 {
		sb.WriteString("\n## Issues\n\n")
		for _, issue := range validation.Issues {
			sb.WriteString(fmt.Sprintf("- ❌ %s\n", issue))
		}
	}

	if len(validation.Warnings) > 0 {
		sb.WriteString("\n## Warnings\n\n")
		for _, warning := range validation.Warnings {
			sb.WriteString(fmt.Sprintf("- ⚠️ %s\n", warning))
		}
	}

	if validation.IsValid && len(validation.Warnings) == 0 {
		sb.WriteString("\n✅ Workflow is ready to run.\n")
	}

	return tools.TextResult(sb.String()), nil
}

// handleWorkflowCondition handles the aftrs_workflow_condition tool
func handleWorkflowCondition(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workflowID := tools.GetStringParam(req, "workflow_id")
	stepID := tools.GetStringParam(req, "step_id")
	variable := tools.GetStringParam(req, "variable")
	operator := tools.GetStringParam(req, "operator")
	value := tools.GetStringParam(req, "value")

	if workflowID == "" || stepID == "" || variable == "" || operator == "" {
		return tools.ErrorResult(errors.New("workflow_id, step_id, variable, and operator are required")), nil
	}

	validOps := map[string]bool{"eq": true, "ne": true, "gt": true, "lt": true, "gte": true, "lte": true, "contains": true, "exists": true}
	if !validOps[operator] {
		return tools.ErrorResult(fmt.Errorf("invalid operator: %s (valid: eq, ne, gt, lt, gte, lte, contains, exists)", operator)), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create workflows client: %w", err)), nil
	}

	if err := client.AddConditionToStep(ctx, workflowID, stepID, variable, operator, value); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to add condition: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Condition Added\n\n")
	sb.WriteString(fmt.Sprintf("**Workflow:** `%s`\n", workflowID))
	sb.WriteString(fmt.Sprintf("**Step:** `%s`\n\n", stepID))
	sb.WriteString("## Condition\n")
	sb.WriteString(fmt.Sprintf("- **Variable:** `%s`\n", variable))
	sb.WriteString(fmt.Sprintf("- **Operator:** `%s`\n", operator))
	if value != "" {
		sb.WriteString(fmt.Sprintf("- **Value:** `%s`\n", value))
	}
	sb.WriteString(fmt.Sprintf("\nStep will only run if: `%s %s %s`\n", variable, operator, value))

	return tools.TextResult(sb.String()), nil
}

// handleWorkflowDelete handles the aftrs_workflow_delete tool
func handleWorkflowDelete(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	workflowID := tools.GetStringParam(req, "workflow_id")

	if workflowID == "" {
		return tools.ErrorResult(errors.New("workflow_id parameter is required")), nil
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create workflows client: %w", err)), nil
	}

	if err := client.DeleteWorkflow(ctx, workflowID); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to delete workflow: %w", err)), nil
	}

	return tools.TextResult(fmt.Sprintf("Workflow `%s` deleted successfully.", workflowID)), nil
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
