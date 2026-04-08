// Package clients provides client implementations for external services.
package clients

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/tasks/v1"

	"github.com/hairglasses-studio/hg-mcp/internal/config"
)

// TasksClient provides Google Tasks operations
type TasksClient struct {
	service *tasks.Service
	mu      sync.RWMutex
}

// TaskList represents a Google Tasks list
type TaskList struct {
	ID      string    `json:"id"`
	Title   string    `json:"title"`
	Updated time.Time `json:"updated,omitempty"`
}

// Task represents a Google Tasks task
type Task struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Notes     string     `json:"notes,omitempty"`
	Status    string     `json:"status"` // needsAction or completed
	Due       time.Time  `json:"due,omitempty"`
	Completed time.Time  `json:"completed,omitempty"`
	Parent    string     `json:"parent,omitempty"`
	Position  string     `json:"position,omitempty"`
	Links     []TaskLink `json:"links,omitempty"`
	ListID    string     `json:"list_id"`
}

// TaskLink represents a link attached to a task
type TaskLink struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
	Link        string `json:"link"`
}

var (
	tasksClient     *TasksClient
	tasksClientOnce sync.Once
	tasksClientErr  error
)

// GetTasksClient returns the singleton Tasks client
func GetTasksClient() (*TasksClient, error) {
	tasksClientOnce.Do(func() {
		tasksClient, tasksClientErr = NewTasksClient()
	})
	return tasksClient, tasksClientErr
}

// NewTasksClient creates a new Google Tasks client
func NewTasksClient() (*TasksClient, error) {
	ctx := context.Background()

	var opts []option.ClientOption

	cfg := config.GetOrLoad()
	if cfg.GoogleApplicationCredentials != "" {
		opts = append(opts, option.WithCredentialsFile(cfg.GoogleApplicationCredentials))
	} else if cfg.GoogleAPIKey != "" {
		opts = append(opts, option.WithAPIKey(cfg.GoogleAPIKey))
	} else {
		return nil, fmt.Errorf("no Google credentials configured (set GOOGLE_APPLICATION_CREDENTIALS or GOOGLE_API_KEY)")
	}

	service, err := tasks.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Tasks service: %w", err)
	}

	return &TasksClient{service: service}, nil
}

// IsConfigured returns true if the client is properly configured
func (c *TasksClient) IsConfigured() bool {
	return c != nil && c.service != nil
}

// ListTaskLists returns all task lists
func (c *TasksClient) ListTaskLists(ctx context.Context) ([]TaskList, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	lists, err := c.service.Tasklists.List().Context(ctx).MaxResults(100).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list task lists: %w", err)
	}

	result := make([]TaskList, 0, len(lists.Items))
	for _, l := range lists.Items {
		tl := TaskList{
			ID:    l.Id,
			Title: l.Title,
		}
		if l.Updated != "" {
			if t, err := time.Parse(time.RFC3339, l.Updated); err == nil {
				tl.Updated = t
			}
		}
		result = append(result, tl)
	}

	return result, nil
}

// GetTaskList gets a specific task list by ID
func (c *TasksClient) GetTaskList(ctx context.Context, listID string) (*TaskList, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	list, err := c.service.Tasklists.Get(listID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get task list: %w", err)
	}

	tl := &TaskList{
		ID:    list.Id,
		Title: list.Title,
	}
	if list.Updated != "" {
		if t, err := time.Parse(time.RFC3339, list.Updated); err == nil {
			tl.Updated = t
		}
	}

	return tl, nil
}

// CreateTaskList creates a new task list
func (c *TasksClient) CreateTaskList(ctx context.Context, title string) (*TaskList, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	list, err := c.service.Tasklists.Insert(&tasks.TaskList{
		Title: title,
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create task list: %w", err)
	}

	return &TaskList{
		ID:    list.Id,
		Title: list.Title,
	}, nil
}

// DeleteTaskList deletes a task list
func (c *TasksClient) DeleteTaskList(ctx context.Context, listID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	err := c.service.Tasklists.Delete(listID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to delete task list: %w", err)
	}
	return nil
}

// UpdateTaskList updates a task list's title
func (c *TasksClient) UpdateTaskList(ctx context.Context, listID, title string) (*TaskList, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	list, err := c.service.Tasklists.Update(listID, &tasks.TaskList{
		Title: title,
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update task list: %w", err)
	}

	return &TaskList{
		ID:    list.Id,
		Title: list.Title,
	}, nil
}

// ListTasks returns tasks from a task list
func (c *TasksClient) ListTasks(ctx context.Context, listID string, showCompleted, showHidden bool, dueMax, dueMin time.Time, maxResults int64) ([]Task, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if listID == "" {
		listID = "@default"
	}
	if maxResults <= 0 {
		maxResults = 100
	}

	call := c.service.Tasks.List(listID).
		Context(ctx).
		MaxResults(maxResults).
		ShowCompleted(showCompleted).
		ShowHidden(showHidden)

	if !dueMax.IsZero() {
		call = call.DueMax(dueMax.Format(time.RFC3339))
	}
	if !dueMin.IsZero() {
		call = call.DueMin(dueMin.Format(time.RFC3339))
	}

	taskList, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list tasks: %w", err)
	}

	result := make([]Task, 0, len(taskList.Items))
	for _, t := range taskList.Items {
		result = append(result, convertTask(t, listID))
	}

	return result, nil
}

// GetTask gets a specific task by ID
func (c *TasksClient) GetTask(ctx context.Context, listID, taskID string) (*Task, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if listID == "" {
		listID = "@default"
	}

	task, err := c.service.Tasks.Get(listID, taskID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	result := convertTask(task, listID)
	return &result, nil
}

// CreateTask creates a new task
func (c *TasksClient) CreateTask(ctx context.Context, listID, title, notes, parent, previous string, due time.Time) (*Task, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if listID == "" {
		listID = "@default"
	}

	newTask := &tasks.Task{
		Title: title,
		Notes: notes,
	}

	if !due.IsZero() {
		newTask.Due = due.Format(time.RFC3339)
	}

	call := c.service.Tasks.Insert(listID, newTask).Context(ctx)
	if parent != "" {
		call = call.Parent(parent)
	}
	if previous != "" {
		call = call.Previous(previous)
	}

	task, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	result := convertTask(task, listID)
	return &result, nil
}

// UpdateTask updates an existing task
func (c *TasksClient) UpdateTask(ctx context.Context, listID, taskID, title, notes, status string, due time.Time) (*Task, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if listID == "" {
		listID = "@default"
	}

	// First get the existing task
	existing, err := c.service.Tasks.Get(listID, taskID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Update fields
	if title != "" {
		existing.Title = title
	}
	if notes != "" {
		existing.Notes = notes
	}
	if status != "" {
		existing.Status = status
	}
	if !due.IsZero() {
		existing.Due = due.Format(time.RFC3339)
	}

	task, err := c.service.Tasks.Update(listID, taskID, existing).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	result := convertTask(task, listID)
	return &result, nil
}

// CompleteTask marks a task as completed
func (c *TasksClient) CompleteTask(ctx context.Context, listID, taskID string) (*Task, error) {
	return c.UpdateTask(ctx, listID, taskID, "", "", "completed", time.Time{})
}

// UncompleteTask marks a task as needing action
func (c *TasksClient) UncompleteTask(ctx context.Context, listID, taskID string) (*Task, error) {
	return c.UpdateTask(ctx, listID, taskID, "", "", "needsAction", time.Time{})
}

// DeleteTask deletes a task
func (c *TasksClient) DeleteTask(ctx context.Context, listID, taskID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if listID == "" {
		listID = "@default"
	}

	err := c.service.Tasks.Delete(listID, taskID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}
	return nil
}

// MoveTask moves a task to a new position
func (c *TasksClient) MoveTask(ctx context.Context, listID, taskID, parent, previous string) (*Task, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if listID == "" {
		listID = "@default"
	}

	call := c.service.Tasks.Move(listID, taskID).Context(ctx)
	if parent != "" {
		call = call.Parent(parent)
	}
	if previous != "" {
		call = call.Previous(previous)
	}

	task, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to move task: %w", err)
	}

	result := convertTask(task, listID)
	return &result, nil
}

// ClearCompletedTasks clears all completed tasks from a list
func (c *TasksClient) ClearCompletedTasks(ctx context.Context, listID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if listID == "" {
		listID = "@default"
	}

	err := c.service.Tasks.Clear(listID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("failed to clear completed tasks: %w", err)
	}
	return nil
}

// GetTodayTasks returns tasks due today
func (c *TasksClient) GetTodayTasks(ctx context.Context, listID string) ([]Task, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	endOfDay := startOfDay.AddDate(0, 0, 1)
	return c.ListTasks(ctx, listID, false, false, endOfDay, startOfDay, 100)
}

// GetOverdueTasks returns overdue tasks
func (c *TasksClient) GetOverdueTasks(ctx context.Context, listID string) ([]Task, error) {
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return c.ListTasks(ctx, listID, false, false, startOfDay, time.Time{}, 100)
}

// GetUpcomingTasks returns tasks due in the next N days
func (c *TasksClient) GetUpcomingTasks(ctx context.Context, listID string, days int) ([]Task, error) {
	if days <= 0 {
		days = 7
	}
	now := time.Now()
	dueMax := now.AddDate(0, 0, days)
	return c.ListTasks(ctx, listID, false, false, dueMax, now, 100)
}

func convertTask(t *tasks.Task, listID string) Task {
	task := Task{
		ID:       t.Id,
		Title:    t.Title,
		Notes:    t.Notes,
		Status:   t.Status,
		Parent:   t.Parent,
		Position: t.Position,
		ListID:   listID,
	}

	if t.Due != "" {
		if due, err := time.Parse(time.RFC3339, t.Due); err == nil {
			task.Due = due
		}
	}
	if t.Completed != nil && *t.Completed != "" {
		if completed, err := time.Parse(time.RFC3339, *t.Completed); err == nil {
			task.Completed = completed
		}
	}

	for _, link := range t.Links {
		task.Links = append(task.Links, TaskLink{
			Type:        link.Type,
			Description: link.Description,
			Link:        link.Link,
		})
	}

	return task
}
