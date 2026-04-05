// Package chains provides MCP tools for workflow chain execution.
package chains

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/chains"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the chains tool module
type Module struct{}

// NewModule creates a new chains module
func NewModule() *Module {
	return &Module{}
}

// Name returns the module name
func (m *Module) Name() string {
	return "chains"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Workflow chain execution for multi-step automated operations"
}

// Tools returns all tool definitions in this module
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		m.chainListTool(),
		m.chainGetTool(),
		m.chainExecuteTool(),
		m.chainStatusTool(),
		m.chainApproveTool(),
		m.chainCancelTool(),
		m.chainPendingTool(),
		m.chainHistoryTool(),
	}
}

func (m *Module) chainListTool() tools.ToolDefinition {
	return tools.ToolDefinition{
		Tool: mcp.NewTool("aftrs_chain_list",
			mcp.WithDescription("List available workflow chains. Filter by category: show, stream, backup, dj, lighting"),
			mcp.WithString("category",
				mcp.Description("Filter by category (optional)"),
				mcp.Enum("show", "stream", "backup", "dj", "lighting", ""),
			),
		),
		Handler:             m.handleChainList,
		Category:            "chains",
		Subcategory:         "management",
		Tags:                []string{"chain", "workflow", "automation", "list"},
		UseCases:            []string{"List available automation workflows", "Find chains by category"},
		Complexity:          tools.ComplexitySimple,
		CircuitBreakerGroup: "chains",
		IsWrite:             false,
	}
}

func (m *Module) handleChainList(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	category := tools.GetStringParam(request, "category")

	executor := chains.GetExecutor()
	chainList := executor.ListChains(category)

	var sb strings.Builder
	sb.WriteString("# Available Workflow Chains\n\n")

	if category != "" {
		sb.WriteString(fmt.Sprintf("**Category:** %s\n\n", category))
	}

	sb.WriteString(fmt.Sprintf("**Total:** %d chains\n\n", len(chainList)))

	// Group by category
	byCategory := make(map[string][]*chains.Chain)
	for _, c := range chainList {
		byCategory[c.Category] = append(byCategory[c.Category], c)
	}

	for cat, chainGroup := range byCategory {
		sb.WriteString(fmt.Sprintf("## %s\n\n", strings.Title(cat)))
		for _, c := range chainGroup {
			sb.WriteString(fmt.Sprintf("### `%s` - %s\n", c.ID, c.Name))
			sb.WriteString(fmt.Sprintf("%s\n\n", c.Description))
			sb.WriteString(fmt.Sprintf("- **Steps:** %d\n", len(c.Steps)))
			if len(c.Parameters) > 0 {
				sb.WriteString("- **Parameters:** ")
				params := make([]string, 0, len(c.Parameters))
				for _, p := range c.Parameters {
					req := ""
					if p.Required {
						req = " (required)"
					}
					params = append(params, fmt.Sprintf("`%s`%s", p.Name, req))
				}
				sb.WriteString(strings.Join(params, ", "))
				sb.WriteString("\n")
			}
			sb.WriteString("\n")
		}
	}

	return tools.TextResult(sb.String()), nil
}

func (m *Module) chainGetTool() tools.ToolDefinition {
	return tools.ToolDefinition{
		Tool: mcp.NewTool("aftrs_chain_get",
			mcp.WithDescription("Get detailed information about a specific workflow chain including all steps"),
			mcp.WithString("chain_id",
				mcp.Required(),
				mcp.Description("Chain ID (e.g., show_startup, stream_start)"),
			),
		),
		Handler:             m.handleChainGet,
		Category:            "chains",
		Subcategory:         "management",
		Tags:                []string{"chain", "workflow", "details"},
		UseCases:            []string{"View chain steps", "Understand workflow structure"},
		Complexity:          tools.ComplexitySimple,
		CircuitBreakerGroup: "chains",
		IsWrite:             false,
	}
}

func (m *Module) handleChainGet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	chainID, errResult := tools.RequireStringParam(request, "chain_id")
	if errResult != nil {
		return errResult, nil
	}

	executor := chains.GetExecutor()
	chain, ok := executor.GetChain(chainID)
	if !ok {
		return tools.ErrorResult(fmt.Errorf("chain not found: %s", chainID)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Chain: %s\n\n", chain.Name))
	sb.WriteString(fmt.Sprintf("**ID:** `%s`\n", chain.ID))
	sb.WriteString(fmt.Sprintf("**Category:** %s\n", chain.Category))
	sb.WriteString(fmt.Sprintf("**Description:** %s\n", chain.Description))
	if chain.Timeout > 0 {
		sb.WriteString(fmt.Sprintf("**Timeout:** %s\n", chain.Timeout))
	}
	sb.WriteString("\n")

	// Parameters
	if len(chain.Parameters) > 0 {
		sb.WriteString("## Parameters\n\n")
		sb.WriteString("| Name | Type | Required | Default | Description |\n")
		sb.WriteString("|------|------|----------|---------|-------------|\n")
		for _, p := range chain.Parameters {
			req := "No"
			if p.Required {
				req = "Yes"
			}
			def := "-"
			if p.Default != nil {
				def = fmt.Sprintf("`%v`", p.Default)
			}
			sb.WriteString(fmt.Sprintf("| `%s` | %s | %s | %s | %s |\n", p.Name, p.Type, req, def, p.Description))
		}
		sb.WriteString("\n")
	}

	// Steps
	sb.WriteString("## Steps\n\n")
	for i, step := range chain.Steps {
		sb.WriteString(fmt.Sprintf("### Step %d: %s\n", i+1, step.Name))
		sb.WriteString(fmt.Sprintf("- **Type:** %s\n", step.Type))
		if step.Tool != "" {
			sb.WriteString(fmt.Sprintf("- **Tool:** `%s`\n", step.Tool))
		}
		if len(step.Inputs) > 0 {
			sb.WriteString("- **Inputs:** ")
			inputs := make([]string, 0, len(step.Inputs))
			for k, v := range step.Inputs {
				inputs = append(inputs, fmt.Sprintf("%s=%v", k, v))
			}
			sb.WriteString(strings.Join(inputs, ", "))
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("- **On Error:** %s\n", step.OnError))
		if step.GateMessage != "" {
			sb.WriteString(fmt.Sprintf("- **Gate Message:** %s\n", step.GateMessage))
		}
		if step.DelayAfter > 0 {
			sb.WriteString(fmt.Sprintf("- **Delay After:** %s\n", step.DelayAfter))
		}
		sb.WriteString("\n")
	}

	// Triggers
	if len(chain.Triggers) > 0 {
		sb.WriteString("## Triggers\n\n")
		for _, t := range chain.Triggers {
			status := "disabled"
			if t.Enabled {
				status = "enabled"
			}
			sb.WriteString(fmt.Sprintf("- **%s** (%s)", t.Type, status))
			if t.Schedule != "" {
				sb.WriteString(fmt.Sprintf(" - Schedule: `%s`", t.Schedule))
			}
			if t.Event != "" {
				sb.WriteString(fmt.Sprintf(" - Event: `%s`", t.Event))
			}
			sb.WriteString("\n")
		}
	}

	return tools.TextResult(sb.String()), nil
}

func (m *Module) chainExecuteTool() tools.ToolDefinition {
	return tools.ToolDefinition{
		Tool: mcp.NewTool("aftrs_chain_execute",
			mcp.WithDescription("Execute a workflow chain. Returns execution ID to track progress with aftrs_chain_status"),
			mcp.WithString("chain_id",
				mcp.Required(),
				mcp.Description("Chain ID to execute"),
			),
			mcp.WithObject("params",
				mcp.Description("Parameters for the chain (JSON object)"),
			),
		),
		Handler:             m.handleChainExecute,
		Category:            "chains",
		Subcategory:         "execution",
		Tags:                []string{"chain", "workflow", "execute", "run"},
		UseCases:            []string{"Start automated workflow", "Run show startup sequence"},
		Complexity:          tools.ComplexityModerate,
		CircuitBreakerGroup: "chains",
		IsWrite:             true,
	}
}

func (m *Module) handleChainExecute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	chainID, errResult := tools.RequireStringParam(request, "chain_id")
	if errResult != nil {
		return errResult, nil
	}

	params := make(map[string]interface{})
	if args, ok := request.Params.Arguments.(map[string]interface{}); ok {
		if p, ok := args["params"].(map[string]interface{}); ok {
			params = p
		}
	}

	executor := chains.GetExecutor()
	exec, err := executor.Execute(ctx, chainID, params, "user")
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Chain Execution Started\n\n")
	sb.WriteString(fmt.Sprintf("**Execution ID:** `%s`\n", exec.ID))
	sb.WriteString(fmt.Sprintf("**Chain:** %s\n", exec.ChainName))
	sb.WriteString(fmt.Sprintf("**Status:** %s\n", exec.Status))
	sb.WriteString(fmt.Sprintf("**Started:** %s\n", exec.StartedAt.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("**Progress:** %d/%d steps\n\n", exec.CurrentStep, exec.TotalSteps))

	if len(params) > 0 {
		sb.WriteString("**Parameters:**\n")
		for k, v := range params {
			sb.WriteString(fmt.Sprintf("- %s: %v\n", k, v))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("Use `aftrs_chain_status` with this execution ID to check progress.\n")

	return tools.TextResult(sb.String()), nil
}

func (m *Module) chainStatusTool() tools.ToolDefinition {
	return tools.ToolDefinition{
		Tool: mcp.NewTool("aftrs_chain_status",
			mcp.WithDescription("Get status of a chain execution including step results"),
			mcp.WithString("execution_id",
				mcp.Required(),
				mcp.Description("Execution ID from aftrs_chain_execute"),
			),
		),
		Handler:             m.handleChainStatus,
		Category:            "chains",
		Subcategory:         "execution",
		Tags:                []string{"chain", "workflow", "status", "progress"},
		UseCases:            []string{"Check workflow progress", "View step results"},
		Complexity:          tools.ComplexitySimple,
		CircuitBreakerGroup: "chains",
		IsWrite:             false,
	}
}

func (m *Module) handleChainStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	execID, errResult := tools.RequireStringParam(request, "execution_id")
	if errResult != nil {
		return errResult, nil
	}

	executor := chains.GetExecutor()
	exec, ok := executor.GetExecution(execID)
	if !ok {
		return tools.ErrorResult(fmt.Errorf("execution not found: %s", execID)), nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Chain Execution: %s\n\n", exec.ChainName))

	// Status with emoji
	statusEmoji := map[chains.ChainStatus]string{
		chains.ChainStatusPending:   "⏳",
		chains.ChainStatusRunning:   "🔄",
		chains.ChainStatusPaused:    "⏸️",
		chains.ChainStatusCompleted: "✅",
		chains.ChainStatusFailed:    "❌",
		chains.ChainStatusCancelled: "🚫",
	}
	emoji := statusEmoji[exec.Status]

	sb.WriteString(fmt.Sprintf("**Status:** %s %s\n", emoji, exec.Status))
	sb.WriteString(fmt.Sprintf("**Progress:** %d/%d steps\n", exec.CurrentStep, exec.TotalSteps))
	sb.WriteString(fmt.Sprintf("**Started:** %s\n", exec.StartedAt.Format(time.RFC3339)))
	if exec.CompletedAt != nil {
		sb.WriteString(fmt.Sprintf("**Completed:** %s\n", exec.CompletedAt.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("**Duration:** %s\n", exec.CompletedAt.Sub(exec.StartedAt).Round(time.Second)))
	}
	if exec.Error != "" {
		sb.WriteString(fmt.Sprintf("**Error:** %s\n", exec.Error))
	}
	sb.WriteString("\n")

	// Step results
	if len(exec.StepResults) > 0 {
		sb.WriteString("## Step Results\n\n")
		for i, result := range exec.StepResults {
			stepEmoji := statusEmoji[result.Status]
			sb.WriteString(fmt.Sprintf("### Step %d: %s %s\n", i+1, stepEmoji, result.StepName))
			sb.WriteString(fmt.Sprintf("- **Status:** %s\n", result.Status))
			if result.CompletedAt != nil {
				duration := result.CompletedAt.Sub(result.StartedAt).Round(time.Millisecond)
				sb.WriteString(fmt.Sprintf("- **Duration:** %s\n", duration))
			}
			if result.Error != "" {
				sb.WriteString(fmt.Sprintf("- **Error:** %s\n", result.Error))
			}
			if result.Retries > 0 {
				sb.WriteString(fmt.Sprintf("- **Retries:** %d\n", result.Retries))
			}
			sb.WriteString("\n")
		}
	}

	// Pending gate
	if exec.Status == chains.ChainStatusPaused {
		sb.WriteString("## ⏸️ Waiting for Approval\n\n")
		sb.WriteString("This chain is paused at a gate step. Use `aftrs_chain_approve` to continue.\n")
	}

	return tools.TextResult(sb.String()), nil
}

func (m *Module) chainApproveTool() tools.ToolDefinition {
	return tools.ToolDefinition{
		Tool: mcp.NewTool("aftrs_chain_approve",
			mcp.WithDescription("Approve or reject a pending gate in a chain execution"),
			mcp.WithString("execution_id",
				mcp.Required(),
				mcp.Description("Execution ID with pending gate"),
			),
			mcp.WithBoolean("approved",
				mcp.Required(),
				mcp.Description("True to approve, false to reject"),
			),
			mcp.WithString("comment",
				mcp.Description("Optional comment"),
			),
		),
		Handler:             m.handleChainApprove,
		Category:            "chains",
		Subcategory:         "execution",
		Tags:                []string{"chain", "workflow", "approve", "gate"},
		UseCases:            []string{"Approve workflow continuation", "Reject risky operation"},
		Complexity:          tools.ComplexitySimple,
		CircuitBreakerGroup: "chains",
		IsWrite:             true,
	}
}

func (m *Module) handleChainApprove(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	execID, errResult := tools.RequireStringParam(request, "execution_id")
	if errResult != nil {
		return errResult, nil
	}

	approved := tools.GetBoolParam(request, "approved", false)
	comment := tools.GetStringParam(request, "comment")

	executor := chains.GetExecutor()
	err := executor.ApproveGate(ctx, execID, approved, "user", comment)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	action := "approved"
	if !approved {
		action = "rejected"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Gate %s\n\n", strings.Title(action)))
	sb.WriteString(fmt.Sprintf("**Execution ID:** `%s`\n", execID))
	sb.WriteString(fmt.Sprintf("**Action:** %s\n", action))
	if comment != "" {
		sb.WriteString(fmt.Sprintf("**Comment:** %s\n", comment))
	}
	sb.WriteString("\n")

	if approved {
		sb.WriteString("Chain execution will continue.\n")
	} else {
		sb.WriteString("Chain execution has been cancelled.\n")
	}

	return tools.TextResult(sb.String()), nil
}

func (m *Module) chainCancelTool() tools.ToolDefinition {
	return tools.ToolDefinition{
		Tool: mcp.NewTool("aftrs_chain_cancel",
			mcp.WithDescription("Cancel a running or paused chain execution"),
			mcp.WithString("execution_id",
				mcp.Required(),
				mcp.Description("Execution ID to cancel"),
			),
		),
		Handler:             m.handleChainCancel,
		Category:            "chains",
		Subcategory:         "execution",
		Tags:                []string{"chain", "workflow", "cancel", "stop"},
		UseCases:            []string{"Stop runaway workflow", "Cancel mistaken execution"},
		Complexity:          tools.ComplexitySimple,
		CircuitBreakerGroup: "chains",
		IsWrite:             true,
	}
}

func (m *Module) handleChainCancel(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	execID, errResult := tools.RequireStringParam(request, "execution_id")
	if errResult != nil {
		return errResult, nil
	}

	executor := chains.GetExecutor()
	err := executor.CancelExecution(execID)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Execution Cancelled\n\n")
	sb.WriteString(fmt.Sprintf("**Execution ID:** `%s`\n", execID))
	sb.WriteString("\nThe chain execution has been cancelled.\n")

	return tools.TextResult(sb.String()), nil
}

func (m *Module) chainPendingTool() tools.ToolDefinition {
	return tools.ToolDefinition{
		Tool: mcp.NewTool("aftrs_chain_pending",
			mcp.WithDescription("List chain executions waiting for gate approval"),
		),
		Handler:             m.handleChainPending,
		Category:            "chains",
		Subcategory:         "execution",
		Tags:                []string{"chain", "workflow", "pending", "gates"},
		UseCases:            []string{"Check for pending approvals", "Review gate requests"},
		Complexity:          tools.ComplexitySimple,
		CircuitBreakerGroup: "chains",
		IsWrite:             false,
	}
}

func (m *Module) handleChainPending(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	executor := chains.GetExecutor()
	pending := executor.ListPendingGates()

	var sb strings.Builder
	sb.WriteString("# Pending Gate Approvals\n\n")

	if len(pending) == 0 {
		sb.WriteString("No pending approvals.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("**Total:** %d pending\n\n", len(pending)))

	for _, gate := range pending {
		sb.WriteString(fmt.Sprintf("## %s\n\n", gate.ChainName))
		sb.WriteString(fmt.Sprintf("**Execution ID:** `%s`\n", gate.ExecutionID))
		sb.WriteString(fmt.Sprintf("**Step:** %s\n", gate.StepName))
		sb.WriteString(fmt.Sprintf("**Message:** %s\n", gate.Message))
		sb.WriteString(fmt.Sprintf("**Waiting Since:** %s\n\n", gate.CreatedAt.Format(time.RFC3339)))
	}

	sb.WriteString("Use `aftrs_chain_approve` with an execution ID to approve or reject.\n")

	return tools.TextResult(sb.String()), nil
}

func (m *Module) chainHistoryTool() tools.ToolDefinition {
	return tools.ToolDefinition{
		Tool: mcp.NewTool("aftrs_chain_history",
			mcp.WithDescription("View recent chain execution history"),
			mcp.WithNumber("limit",
				mcp.Description("Number of executions to return (default: 10)"),
			),
		),
		Handler:             m.handleChainHistory,
		Category:            "chains",
		Subcategory:         "execution",
		Tags:                []string{"chain", "workflow", "history", "executions"},
		UseCases:            []string{"Review past executions", "Check workflow success rate"},
		Complexity:          tools.ComplexitySimple,
		CircuitBreakerGroup: "chains",
		IsWrite:             false,
	}
}

func (m *Module) handleChainHistory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := tools.GetIntParam(request, "limit", 10)

	executor := chains.GetExecutor()
	executions := executor.ListExecutions(limit)

	var sb strings.Builder
	sb.WriteString("# Chain Execution History\n\n")

	if len(executions) == 0 {
		sb.WriteString("No executions found.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("**Showing:** %d executions\n\n", len(executions)))

	statusEmoji := map[chains.ChainStatus]string{
		chains.ChainStatusPending:   "⏳",
		chains.ChainStatusRunning:   "🔄",
		chains.ChainStatusPaused:    "⏸️",
		chains.ChainStatusCompleted: "✅",
		chains.ChainStatusFailed:    "❌",
		chains.ChainStatusCancelled: "🚫",
	}

	sb.WriteString("| Status | Chain | Started | Duration | Steps |\n")
	sb.WriteString("|--------|-------|---------|----------|-------|\n")

	for _, exec := range executions {
		emoji := statusEmoji[exec.Status]
		duration := "-"
		if exec.CompletedAt != nil {
			duration = exec.CompletedAt.Sub(exec.StartedAt).Round(time.Second).String()
		}
		sb.WriteString(fmt.Sprintf("| %s %s | %s | %s | %s | %d/%d |\n",
			emoji, exec.Status,
			exec.ChainName,
			exec.StartedAt.Format("Jan 2 15:04"),
			duration,
			exec.CurrentStep, exec.TotalSteps,
		))
	}

	return tools.TextResult(sb.String()), nil
}

// Register the module
func init() {
	tools.GetRegistry().RegisterModule(NewModule())
}
