// Package tasks provides MCP task tracking for long-running operations.
// This implements a basic version of MCP Tasks (SEP-1686) for async operations
// like stems separation, AI inference, and file processing.
package tasks

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/config"
)

// TaskManager manages async tasks and their status
type TaskManager struct {
	mu      sync.RWMutex
	tasks   map[string]*Task
	dataDir string // Optional directory for persistence (empty = in-memory only)
}

// Task represents an async operation with progress tracking
type Task struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`        // e.g., "stems", "inference", "backup"
	Description string                 `json:"description"` // Human-readable description
	Status      TaskStatus             `json:"status"`      // working, completed, failed, cancelled
	Progress    float64                `json:"progress"`    // 0.0 to 1.0
	Message     string                 `json:"message"`     // Current status message
	Result      interface{}            `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   time.Time              `json:"started_at,omitempty"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CompletedAt time.Time              `json:"completed_at,omitempty"`
}

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusWorking   TaskStatus = "working"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// Global task manager instance
var globalManager *TaskManager
var managerOnce sync.Once

// GetTaskManager returns the global task manager.
// It uses AFTRS_DATA_DIR (or ~/.hg-mcp) for persistence.
func GetTaskManager() *TaskManager {
	managerOnce.Do(func() {
		globalManager = NewTaskManager(config.Get().AftrsDataDir)
	})
	return globalManager
}

// NewTaskManager creates a new TaskManager with optional persistence.
// If dataDir is non-empty, tasks are loaded from and saved to tasks.json.
func NewTaskManager(dataDir string) *TaskManager {
	m := &TaskManager{
		tasks:   make(map[string]*Task),
		dataDir: dataDir,
	}
	m.loadFromDisk()
	return m
}

// CreateTask creates a new task and returns its ID
func (m *TaskManager) CreateTask(ctx context.Context, taskType, description string, metadata map[string]interface{}) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	id := fmt.Sprintf("task_%d", time.Now().UnixNano())
	now := time.Now()

	task := &Task{
		ID:          id,
		Type:        taskType,
		Description: description,
		Status:      TaskStatusPending,
		Progress:    0,
		Message:     "Task created",
		Metadata:    metadata,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	m.tasks[id] = task
	m.saveToDisk()
	return id
}

// StartTask marks a task as working
func (m *TaskManager) StartTask(ctx context.Context, taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	now := time.Now()
	task.Status = TaskStatusWorking
	task.StartedAt = now
	task.UpdatedAt = now
	task.Message = "Task started"

	m.saveToDisk()
	return nil
}

// UpdateProgress updates task progress and message
func (m *TaskManager) UpdateProgress(ctx context.Context, taskID string, progress float64, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	if progress < 0 {
		progress = 0
	}
	if progress > 1 {
		progress = 1
	}

	task.Progress = progress
	task.Message = message
	task.UpdatedAt = time.Now()

	m.saveToDisk()
	return nil
}

// CompleteTask marks a task as completed with a result
func (m *TaskManager) CompleteTask(ctx context.Context, taskID string, result interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	now := time.Now()
	task.Status = TaskStatusCompleted
	task.Progress = 1.0
	task.Result = result
	task.Message = "Task completed"
	task.UpdatedAt = now
	task.CompletedAt = now

	m.saveToDisk()
	return nil
}

// FailTask marks a task as failed with an error
func (m *TaskManager) FailTask(ctx context.Context, taskID string, err error) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	now := time.Now()
	task.Status = TaskStatusFailed
	task.Error = err.Error()
	task.Message = "Task failed"
	task.UpdatedAt = now
	task.CompletedAt = now

	m.saveToDisk()
	return nil
}

// CancelTask marks a task as cancelled
func (m *TaskManager) CancelTask(ctx context.Context, taskID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return fmt.Errorf("task not found: %s", taskID)
	}

	if task.Status == TaskStatusCompleted || task.Status == TaskStatusFailed {
		return fmt.Errorf("cannot cancel %s task", task.Status)
	}

	now := time.Now()
	task.Status = TaskStatusCancelled
	task.Message = "Task cancelled"
	task.UpdatedAt = now
	task.CompletedAt = now

	m.saveToDisk()
	return nil
}

// GetTask returns a task by ID
func (m *TaskManager) GetTask(ctx context.Context, taskID string) (*Task, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	task, ok := m.tasks[taskID]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}

	// Return a copy
	taskCopy := *task
	return &taskCopy, nil
}

// ListTasks returns all tasks, optionally filtered by status
func (m *TaskManager) ListTasks(ctx context.Context, status TaskStatus, taskType string) []*Task {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var tasks []*Task
	for _, task := range m.tasks {
		if status != "" && task.Status != status {
			continue
		}
		if taskType != "" && task.Type != taskType {
			continue
		}
		taskCopy := *task
		tasks = append(tasks, &taskCopy)
	}

	return tasks
}

// CleanupOldTasks removes completed/failed tasks older than the given duration
func (m *TaskManager) CleanupOldTasks(ctx context.Context, olderThan time.Duration) int {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-olderThan)
	count := 0

	for id, task := range m.tasks {
		if task.Status == TaskStatusCompleted || task.Status == TaskStatusFailed || task.Status == TaskStatusCancelled {
			if task.CompletedAt.Before(cutoff) {
				delete(m.tasks, id)
				count++
			}
		}
	}

	if count > 0 {
		m.saveToDisk()
	}
	return count
}

// TaskSummary provides a quick summary of tasks
type TaskSummary struct {
	TotalTasks     int `json:"total_tasks"`
	PendingCount   int `json:"pending_count"`
	WorkingCount   int `json:"working_count"`
	CompletedCount int `json:"completed_count"`
	FailedCount    int `json:"failed_count"`
	CancelledCount int `json:"cancelled_count"`
}

// GetSummary returns a summary of all tasks
func (m *TaskManager) GetSummary(ctx context.Context) *TaskSummary {
	m.mu.RLock()
	defer m.mu.RUnlock()

	summary := &TaskSummary{
		TotalTasks: len(m.tasks),
	}

	for _, task := range m.tasks {
		switch task.Status {
		case TaskStatusPending:
			summary.PendingCount++
		case TaskStatusWorking:
			summary.WorkingCount++
		case TaskStatusCompleted:
			summary.CompletedCount++
		case TaskStatusFailed:
			summary.FailedCount++
		case TaskStatusCancelled:
			summary.CancelledCount++
		}
	}

	return summary
}

// RunAsync runs a function as an async task and returns the task ID
func (m *TaskManager) RunAsync(ctx context.Context, taskType, description string, fn func(ctx context.Context, updateProgress func(progress float64, message string)) (interface{}, error)) string {
	taskID := m.CreateTask(ctx, taskType, description, nil)

	go func() {
		taskCtx := context.Background()
		if err := m.StartTask(taskCtx, taskID); err != nil {
			slog.Warn("task state update failed", "task_id", taskID, "op", "start", "error", err)
		}

		updateFn := func(progress float64, message string) {
			if err := m.UpdateProgress(taskCtx, taskID, progress, message); err != nil {
				slog.Warn("task state update failed", "task_id", taskID, "op", "progress", "error", err)
			}
		}

		result, err := fn(taskCtx, updateFn)
		if err != nil {
			if fErr := m.FailTask(taskCtx, taskID, err); fErr != nil {
				slog.Warn("task state update failed", "task_id", taskID, "op", "fail", "error", fErr)
			}
		} else {
			if cErr := m.CompleteTask(taskCtx, taskID, result); cErr != nil {
				slog.Warn("task state update failed", "task_id", taskID, "op", "complete", "error", cErr)
			}
		}
	}()

	return taskID
}

// tasksFilePath returns the path to the tasks persistence file.
func (m *TaskManager) tasksFilePath() string {
	if m.dataDir == "" {
		return ""
	}
	return filepath.Join(m.dataDir, "tasks.json")
}

// loadFromDisk loads persisted tasks from disk on init.
func (m *TaskManager) loadFromDisk() {
	path := m.tasksFilePath()
	if path == "" {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return // File doesn't exist yet, that's fine
	}
	var tasks []*Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		slog.Warn("failed to load persisted tasks", "path", path, "error", err)
		return
	}
	for _, t := range tasks {
		m.tasks[t.ID] = t
	}
}

// saveToDisk persists all tasks to disk. Must be called with mu held.
func (m *TaskManager) saveToDisk() {
	path := m.tasksFilePath()
	if path == "" {
		return
	}
	tasks := make([]*Task, 0, len(m.tasks))
	for _, t := range m.tasks {
		tasks = append(tasks, t)
	}
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return
	}
	_ = os.MkdirAll(m.dataDir, 0755)
	_ = os.WriteFile(path, data, 0644)
}
