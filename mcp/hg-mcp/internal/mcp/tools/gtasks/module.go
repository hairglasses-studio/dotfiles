// Package gtasks provides MCP tools for Google Tasks integration.
package gtasks

import (
	"context"
	"fmt"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Module implements the Google Tasks tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "gtasks"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Google Tasks integration for task and to-do list management"
}

// Tools returns the Google Tasks tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	allTools := []tools.ToolDefinition{
		// Task Lists
		{
			Tool: mcp.NewTool("aftrs_tasks_lists",
				mcp.WithDescription("List all Google Tasks lists"),
			),
			Handler:     handleListTaskLists,
			Category:    "tasks",
			Subcategory: "lists",
			Tags:        []string{"tasks", "google", "lists"},
			UseCases:    []string{"View all task lists", "Find list IDs", "Organize tasks"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_tasks_list_create",
				mcp.WithDescription("Create a new task list"),
				mcp.WithString("title", mcp.Required(), mcp.Description("Name for the new task list")),
			),
			Handler:     handleCreateTaskList,
			Category:    "tasks",
			Subcategory: "lists",
			Tags:        []string{"tasks", "create", "lists"},
			UseCases:    []string{"Create project list", "Organize by category", "New task collection"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_tasks_list_delete",
				mcp.WithDescription("Delete a task list"),
				mcp.WithString("list_id", mcp.Required(), mcp.Description("ID of the task list to delete")),
			),
			Handler:     handleDeleteTaskList,
			Category:    "tasks",
			Subcategory: "lists",
			Tags:        []string{"tasks", "delete", "lists"},
			UseCases:    []string{"Remove completed project", "Clean up old lists"},
			Complexity:  tools.ComplexityModerate,
		},
		// Tasks
		{
			Tool: mcp.NewTool("aftrs_tasks_get",
				mcp.WithDescription("Get tasks from a task list"),
				mcp.WithString("list_id", mcp.Description("Task list ID (default: @default for primary list)")),
				mcp.WithBoolean("show_completed", mcp.Description("Include completed tasks (default: false)")),
				mcp.WithNumber("max_results", mcp.Description("Maximum tasks to return (default: 100)")),
			),
			Handler:     handleGetTasks,
			Category:    "tasks",
			Subcategory: "tasks",
			Tags:        []string{"tasks", "list", "todos"},
			UseCases:    []string{"View pending tasks", "Check task list", "Review todos"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_tasks_today",
				mcp.WithDescription("Get tasks due today"),
				mcp.WithString("list_id", mcp.Description("Task list ID (default: @default)")),
			),
			Handler:     handleTodayTasks,
			Category:    "tasks",
			Subcategory: "tasks",
			Tags:        []string{"tasks", "today", "daily"},
			UseCases:    []string{"Morning review", "Today's priorities", "Daily planning"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_tasks_overdue",
				mcp.WithDescription("Get overdue tasks"),
				mcp.WithString("list_id", mcp.Description("Task list ID (default: @default)")),
			),
			Handler:     handleOverdueTasks,
			Category:    "tasks",
			Subcategory: "tasks",
			Tags:        []string{"tasks", "overdue", "urgent"},
			UseCases:    []string{"Find late tasks", "Priority review", "Catch up planning"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_tasks_upcoming",
				mcp.WithDescription("Get tasks due in the next N days"),
				mcp.WithString("list_id", mcp.Description("Task list ID (default: @default)")),
				mcp.WithNumber("days", mcp.Description("Number of days to look ahead (default: 7)")),
			),
			Handler:     handleUpcomingTasks,
			Category:    "tasks",
			Subcategory: "tasks",
			Tags:        []string{"tasks", "upcoming", "planning"},
			UseCases:    []string{"Week planning", "Upcoming deadlines", "Schedule review"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_task_create",
				mcp.WithDescription("Create a new task"),
				mcp.WithString("title", mcp.Required(), mcp.Description("Task title")),
				mcp.WithString("notes", mcp.Description("Task notes/description")),
				mcp.WithString("list_id", mcp.Description("Task list ID (default: @default)")),
				mcp.WithString("due", mcp.Description("Due date (YYYY-MM-DD or ISO 8601)")),
				mcp.WithString("parent_id", mcp.Description("Parent task ID for subtasks")),
			),
			Handler:     handleCreateTask,
			Category:    "tasks",
			Subcategory: "tasks",
			Tags:        []string{"tasks", "create", "todo"},
			UseCases:    []string{"Add new task", "Create reminder", "Schedule work"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_task_update",
				mcp.WithDescription("Update an existing task"),
				mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID to update")),
				mcp.WithString("list_id", mcp.Description("Task list ID (default: @default)")),
				mcp.WithString("title", mcp.Description("New task title")),
				mcp.WithString("notes", mcp.Description("New task notes")),
				mcp.WithString("due", mcp.Description("New due date (YYYY-MM-DD or ISO 8601)")),
			),
			Handler:     handleUpdateTask,
			Category:    "tasks",
			Subcategory: "tasks",
			Tags:        []string{"tasks", "update", "edit"},
			UseCases:    []string{"Change due date", "Update notes", "Rename task"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_task_complete",
				mcp.WithDescription("Mark a task as completed"),
				mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID to complete")),
				mcp.WithString("list_id", mcp.Description("Task list ID (default: @default)")),
			),
			Handler:     handleCompleteTask,
			Category:    "tasks",
			Subcategory: "tasks",
			Tags:        []string{"tasks", "complete", "done"},
			UseCases:    []string{"Finish task", "Mark done", "Complete item"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_task_uncomplete",
				mcp.WithDescription("Mark a completed task as needing action"),
				mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID to uncomplete")),
				mcp.WithString("list_id", mcp.Description("Task list ID (default: @default)")),
			),
			Handler:     handleUncompleteTask,
			Category:    "tasks",
			Subcategory: "tasks",
			Tags:        []string{"tasks", "reopen", "undo"},
			UseCases:    []string{"Reopen task", "Undo completion", "Needs more work"},
			Complexity:  tools.ComplexitySimple,
		},
		{
			Tool: mcp.NewTool("aftrs_task_delete",
				mcp.WithDescription("Delete a task"),
				mcp.WithString("task_id", mcp.Required(), mcp.Description("Task ID to delete")),
				mcp.WithString("list_id", mcp.Description("Task list ID (default: @default)")),
			),
			Handler:     handleDeleteTask,
			Category:    "tasks",
			Subcategory: "tasks",
			Tags:        []string{"tasks", "delete", "remove"},
			UseCases:    []string{"Remove task", "Delete item", "Clean up"},
			Complexity:  tools.ComplexityModerate,
		},
		{
			Tool: mcp.NewTool("aftrs_tasks_clear_completed",
				mcp.WithDescription("Clear all completed tasks from a list"),
				mcp.WithString("list_id", mcp.Description("Task list ID (default: @default)")),
			),
			Handler:     handleClearCompleted,
			Category:    "tasks",
			Subcategory: "tasks",
			Tags:        []string{"tasks", "clear", "cleanup"},
			UseCases:    []string{"Clean up list", "Remove done tasks", "Fresh start"},
			Complexity:  tools.ComplexityModerate,
		},
	}
	for i := range allTools {
		allTools[i].CircuitBreakerGroup = "google"
	}
	return allTools
}

var getTasksClient = tools.LazyClient(clients.GetTasksClient)

func handleListTaskLists(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTasksClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Tasks client: %w", err)), nil
	}

	lists, err := clients.GetTaskListsCached(ctx, client)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to list task lists: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"lists": lists,
		"count": len(lists),
	}), nil
}

func handleCreateTaskList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTasksClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Tasks client: %w", err)), nil
	}

	title, errResult := tools.RequireStringParam(req, "title")
	if errResult != nil {
		return errResult, nil
	}

	list, err := client.CreateTaskList(ctx, title)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to create task list: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"list":    list,
		"message": fmt.Sprintf("Created task list: %s", list.Title),
	}), nil
}

func handleDeleteTaskList(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTasksClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Tasks client: %w", err)), nil
	}

	listID, errResult := tools.RequireStringParam(req, "list_id")
	if errResult != nil {
		return errResult, nil
	}

	if err := client.DeleteTaskList(ctx, listID); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to delete task list: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"message": "Task list deleted successfully",
	}), nil
}

func handleGetTasks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTasksClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Tasks client: %w", err)), nil
	}

	listID := tools.GetStringParam(req, "list_id")
	showCompleted := tools.GetBoolParam(req, "show_completed", false)
	maxResults := tools.GetIntParam(req, "max_results", 100)

	tasks, err := client.ListTasks(ctx, listID, showCompleted, false, time.Time{}, time.Time{}, int64(maxResults))
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get tasks: %w", err)), nil
	}

	listIDVal := listID
	if listIDVal == "" {
		listIDVal = "@default"
	}
	return tools.JSONResult(map[string]interface{}{
		"tasks":   tasks,
		"count":   len(tasks),
		"list_id": listIDVal,
	}), nil
}

func handleTodayTasks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTasksClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Tasks client: %w", err)), nil
	}

	listID := tools.GetStringParam(req, "list_id")

	tasks, err := client.GetTodayTasks(ctx, listID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get today's tasks: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"tasks": tasks,
		"count": len(tasks),
		"date":  time.Now().Format("2006-01-02"),
	}), nil
}

func handleOverdueTasks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTasksClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Tasks client: %w", err)), nil
	}

	listID := tools.GetStringParam(req, "list_id")

	tasks, err := client.GetOverdueTasks(ctx, listID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get overdue tasks: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"tasks": tasks,
		"count": len(tasks),
	}), nil
}

func handleUpcomingTasks(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTasksClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Tasks client: %w", err)), nil
	}

	listID := tools.GetStringParam(req, "list_id")
	days := tools.GetIntParam(req, "days", 7)

	tasks, err := client.GetUpcomingTasks(ctx, listID, days)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get upcoming tasks: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"tasks":      tasks,
		"count":      len(tasks),
		"days_ahead": days,
	}), nil
}

func handleCreateTask(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTasksClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Tasks client: %w", err)), nil
	}

	title, errResult := tools.RequireStringParam(req, "title")
	if errResult != nil {
		return errResult, nil
	}

	notes := tools.GetStringParam(req, "notes")
	listID := tools.GetStringParam(req, "list_id")
	parentID := tools.GetStringParam(req, "parent_id")
	dueStr := tools.GetStringParam(req, "due")

	var due time.Time
	if dueStr != "" {
		var err error
		due, err = time.Parse(time.RFC3339, dueStr)
		if err != nil {
			due, err = time.Parse("2006-01-02", dueStr)
			if err != nil {
				return tools.ErrorResult(fmt.Errorf("invalid due date format: %w", err)), nil
			}
		}
	}

	task, err := client.CreateTask(ctx, listID, title, notes, parentID, "", due)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to create task: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"task":    task,
		"message": fmt.Sprintf("Created task: %s", task.Title),
	}), nil
}

func handleUpdateTask(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTasksClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Tasks client: %w", err)), nil
	}

	taskID, errResult := tools.RequireStringParam(req, "task_id")
	if errResult != nil {
		return errResult, nil
	}

	listID := tools.GetStringParam(req, "list_id")
	title := tools.GetStringParam(req, "title")
	notes := tools.GetStringParam(req, "notes")
	dueStr := tools.GetStringParam(req, "due")

	var due time.Time
	if dueStr != "" {
		var err error
		due, err = time.Parse(time.RFC3339, dueStr)
		if err != nil {
			due, err = time.Parse("2006-01-02", dueStr)
			if err != nil {
				return tools.ErrorResult(fmt.Errorf("invalid due date format: %w", err)), nil
			}
		}
	}

	task, err := client.UpdateTask(ctx, listID, taskID, title, notes, "", due)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to update task: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"task":    task,
		"message": "Task updated successfully",
	}), nil
}

func handleCompleteTask(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTasksClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Tasks client: %w", err)), nil
	}

	taskID, errResult := tools.RequireStringParam(req, "task_id")
	if errResult != nil {
		return errResult, nil
	}

	listID := tools.GetStringParam(req, "list_id")

	task, err := client.CompleteTask(ctx, listID, taskID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to complete task: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"task":    task,
		"message": fmt.Sprintf("Completed task: %s", task.Title),
	}), nil
}

func handleUncompleteTask(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTasksClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Tasks client: %w", err)), nil
	}

	taskID, errResult := tools.RequireStringParam(req, "task_id")
	if errResult != nil {
		return errResult, nil
	}

	listID := tools.GetStringParam(req, "list_id")

	task, err := client.UncompleteTask(ctx, listID, taskID)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to uncomplete task: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"task":    task,
		"message": fmt.Sprintf("Reopened task: %s", task.Title),
	}), nil
}

func handleDeleteTask(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTasksClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Tasks client: %w", err)), nil
	}

	taskID, errResult := tools.RequireStringParam(req, "task_id")
	if errResult != nil {
		return errResult, nil
	}

	listID := tools.GetStringParam(req, "list_id")

	if err := client.DeleteTask(ctx, listID, taskID); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to delete task: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"message": "Task deleted successfully",
	}), nil
}

func handleClearCompleted(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTasksClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create Tasks client: %w", err)), nil
	}

	listID := tools.GetStringParam(req, "list_id")

	if err := client.ClearCompletedTasks(ctx, listID); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to clear completed tasks: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"message": "Completed tasks cleared successfully",
	}), nil
}
