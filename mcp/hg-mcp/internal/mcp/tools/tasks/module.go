// Package tasks provides MCP task management tools for hg-mcp.
package tasks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	tasksmgr "github.com/hairglasses-studio/hg-mcp/internal/mcp/tasks"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for task management
type Module struct{}

func (m *Module) Name() string {
	return "tasks"
}

func (m *Module) Description() string {
	return "Async task management for long-running operations"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_task_list",
				mcp.WithDescription("List all async tasks. Optionally filter by status or type."),
				mcp.WithString("status", mcp.Description("Filter by status: pending, working, completed, failed, cancelled")),
				mcp.WithString("type", mcp.Description("Filter by task type (e.g., stems, inference, backup)")),
			),
			Handler:             handleTaskList,
			Category:            "tasks",
			Subcategory:         "management",
			Tags:                []string{"task", "async", "list", "status"},
			UseCases:            []string{"View running tasks", "Check task queue"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tasks",
		},
		{
			Tool: mcp.NewTool("aftrs_task_status",
				mcp.WithDescription("Get detailed status of a specific task."),
				mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID to check")),
			),
			Handler:             handleTaskStatus,
			Category:            "tasks",
			Subcategory:         "management",
			Tags:                []string{"task", "async", "status", "progress"},
			UseCases:            []string{"Check task progress", "Get task result"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tasks",
		},
		{
			Tool: mcp.NewTool("aftrs_task_cancel",
				mcp.WithDescription("Cancel a pending or running task."),
				mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID to cancel")),
			),
			Handler:             handleTaskCancel,
			Category:            "tasks",
			Subcategory:         "management",
			Tags:                []string{"task", "async", "cancel", "stop"},
			UseCases:            []string{"Cancel long-running task", "Stop processing"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tasks",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_task_cleanup",
				mcp.WithDescription("Remove old completed/failed tasks from memory."),
				mcp.WithNumber("older_than_hours", mcp.Description("Remove tasks older than N hours (default: 24)")),
			),
			Handler:             handleTaskCleanup,
			Category:            "tasks",
			Subcategory:         "management",
			Tags:                []string{"task", "cleanup", "maintenance"},
			UseCases:            []string{"Clean up old tasks", "Free memory"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tasks",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_task_summary",
				mcp.WithDescription("Get a summary of all tasks by status."),
			),
			Handler:             handleTaskSummary,
			Category:            "tasks",
			Subcategory:         "management",
			Tags:                []string{"task", "summary", "overview"},
			UseCases:            []string{"View task queue health", "Monitor async operations"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "tasks",
		},
	}
}

// handleTaskList handles the aftrs_task_list tool
func handleTaskList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	statusFilter := tools.GetStringParam(req, "status")
	typeFilter := tools.GetStringParam(req, "type")

	mgr := tasksmgr.GetTaskManager()
	tasks := mgr.ListTasks(ctx, tasksmgr.TaskStatus(statusFilter), typeFilter)

	var sb strings.Builder
	sb.WriteString("# Async Tasks\n\n")

	if len(tasks) == 0 {
		sb.WriteString("No tasks found.\n")
		return tools.TextResult(sb.String()), nil
	}

	sb.WriteString(fmt.Sprintf("Found **%d** tasks:\n\n", len(tasks)))
	sb.WriteString("| ID | Type | Status | Progress | Description | Age |\n")
	sb.WriteString("|----|------|--------|----------|-------------|-----|\n")

	for _, task := range tasks {
		age := time.Since(task.CreatedAt).Round(time.Second)
		progress := fmt.Sprintf("%.0f%%", task.Progress*100)
		status := getTaskStatusEmoji(task.Status) + " " + string(task.Status)
		desc := task.Description
		if len(desc) > 30 {
			desc = desc[:30] + "..."
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %s | %s |\n",
			task.ID[:12], task.Type, status, progress, desc, age))
	}

	return tools.TextResult(sb.String()), nil
}

// handleTaskStatus handles the aftrs_task_status tool
func handleTaskStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID, errResult := tools.RequireStringParam(req, "task_id")
	if errResult != nil {
		return errResult, nil
	}

	mgr := tasksmgr.GetTaskManager()
	task, err := mgr.GetTask(ctx, taskID)
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	var sb strings.Builder
	sb.WriteString("# Task Status\n\n")
	sb.WriteString(fmt.Sprintf("**ID:** `%s`\n", task.ID))
	sb.WriteString(fmt.Sprintf("**Type:** %s\n", task.Type))
	sb.WriteString(fmt.Sprintf("**Description:** %s\n", task.Description))
	sb.WriteString(fmt.Sprintf("**Status:** %s %s\n", getTaskStatusEmoji(task.Status), task.Status))
	sb.WriteString(fmt.Sprintf("**Progress:** %.0f%%\n", task.Progress*100))
	sb.WriteString(fmt.Sprintf("**Message:** %s\n", task.Message))
	sb.WriteString(fmt.Sprintf("**Created:** %s\n", task.CreatedAt.Format("2006-01-02 15:04:05")))

	if !task.StartedAt.IsZero() {
		sb.WriteString(fmt.Sprintf("**Started:** %s\n", task.StartedAt.Format("2006-01-02 15:04:05")))
	}

	if !task.CompletedAt.IsZero() {
		sb.WriteString(fmt.Sprintf("**Completed:** %s\n", task.CompletedAt.Format("2006-01-02 15:04:05")))
		duration := task.CompletedAt.Sub(task.StartedAt)
		sb.WriteString(fmt.Sprintf("**Duration:** %s\n", duration.Round(time.Millisecond)))
	}

	if task.Error != "" {
		sb.WriteString(fmt.Sprintf("\n## Error\n\n```\n%s\n```\n", task.Error))
	}

	if task.Result != nil {
		sb.WriteString(fmt.Sprintf("\n## Result\n\n```\n%v\n```\n", task.Result))
	}

	return tools.TextResult(sb.String()), nil
}

// handleTaskCancel handles the aftrs_task_cancel tool
func handleTaskCancel(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	taskID, errResult := tools.RequireStringParam(req, "task_id")
	if errResult != nil {
		return errResult, nil
	}

	mgr := tasksmgr.GetTaskManager()
	if err := mgr.CancelTask(ctx, taskID); err != nil {
		return tools.ErrorResult(err), nil
	}

	return tools.TextResult(fmt.Sprintf("Task `%s` cancelled.", taskID)), nil
}

// handleTaskCleanup handles the aftrs_task_cleanup tool
func handleTaskCleanup(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	hours := tools.GetIntParam(req, "older_than_hours", 24)
	if hours < 1 {
		hours = 1
	}

	mgr := tasksmgr.GetTaskManager()
	count := mgr.CleanupOldTasks(ctx, time.Duration(hours)*time.Hour)

	return tools.TextResult(fmt.Sprintf("Cleaned up %d old tasks (older than %d hours).", count, hours)), nil
}

// handleTaskSummary handles the aftrs_task_summary tool
func handleTaskSummary(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	mgr := tasksmgr.GetTaskManager()
	summary := mgr.GetSummary(ctx)

	var sb strings.Builder
	sb.WriteString("# Task Summary\n\n")
	sb.WriteString(fmt.Sprintf("**Total Tasks:** %d\n\n", summary.TotalTasks))

	sb.WriteString("| Status | Count |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| ⏳ Pending | %d |\n", summary.PendingCount))
	sb.WriteString(fmt.Sprintf("| 🔄 Working | %d |\n", summary.WorkingCount))
	sb.WriteString(fmt.Sprintf("| ✅ Completed | %d |\n", summary.CompletedCount))
	sb.WriteString(fmt.Sprintf("| ❌ Failed | %d |\n", summary.FailedCount))
	sb.WriteString(fmt.Sprintf("| ⚪ Cancelled | %d |\n", summary.CancelledCount))

	return tools.TextResult(sb.String()), nil
}

func getTaskStatusEmoji(status tasksmgr.TaskStatus) string {
	switch status {
	case tasksmgr.TaskStatusPending:
		return "⏳"
	case tasksmgr.TaskStatusWorking:
		return "🔄"
	case tasksmgr.TaskStatusCompleted:
		return "✅"
	case tasksmgr.TaskStatusFailed:
		return "❌"
	case tasksmgr.TaskStatusCancelled:
		return "⚪"
	default:
		return "❓"
	}
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
