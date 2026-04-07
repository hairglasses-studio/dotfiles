// Package errbudget provides per-tool consecutive error tracking.
//
// It monitors tool invocations and tracks consecutive failures per tool name.
// After a configurable threshold of consecutive failures (default 3), subsequent
// calls return a degraded response indicating the tool is temporarily unhealthy,
// without executing the underlying handler. The counter resets on a successful
// invocation.
//
// Usage as middleware in wrapHandler or standalone:
//
//	tracker := errbudget.NewTracker(3)
//	wrapped := tracker.Wrap(toolName, originalHandler)
//
// Or query status directly:
//
//	status := tracker.Status(toolName)
//	allStatus := tracker.AllStatus()
package errbudget

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// DefaultThreshold is the number of consecutive failures before degraded mode.
const DefaultThreshold = 3

// HandlerFunc matches the MCP tool handler signature.
type HandlerFunc func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error)

// ToolStatus reports the error budget state for a single tool.
type ToolStatus struct {
	ToolName           string    `json:"tool_name"`
	ConsecutiveErrors  int       `json:"consecutive_errors"`
	Threshold          int       `json:"threshold"`
	Degraded           bool      `json:"degraded"`
	LastError          string    `json:"last_error,omitempty"`
	LastErrorTime      time.Time `json:"last_error_time,omitempty"`
	LastSuccessTime    time.Time `json:"last_success_time,omitempty"`
	TotalInvocations   int64     `json:"total_invocations"`
	TotalErrors        int64     `json:"total_errors"`
	DegradedSinceCount int64     `json:"degraded_rejections"`
}

// toolState is the internal mutable state for a tool.
type toolState struct {
	consecutiveErrors int
	lastError         string
	lastErrorTime     time.Time
	lastSuccessTime   time.Time
	totalInvocations  int64
	totalErrors       int64
	degradedCount     int64
}

// Tracker tracks consecutive errors per tool and gates execution.
type Tracker struct {
	mu        sync.RWMutex
	threshold int
	tools     map[string]*toolState
}

// NewTracker creates a Tracker with the given consecutive-error threshold.
// A threshold <= 0 defaults to DefaultThreshold.
func NewTracker(threshold int) *Tracker {
	if threshold <= 0 {
		threshold = DefaultThreshold
	}
	return &Tracker{
		threshold: threshold,
		tools:     make(map[string]*toolState, 64),
	}
}

// getOrCreate returns the state for a tool, creating it if needed.
// Caller must hold t.mu (write lock).
func (t *Tracker) getOrCreate(name string) *toolState {
	s, ok := t.tools[name]
	if !ok {
		s = &toolState{}
		t.tools[name] = s
	}
	return s
}

// RecordSuccess records a successful invocation, resetting the consecutive error count.
func (t *Tracker) RecordSuccess(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	s := t.getOrCreate(name)
	s.consecutiveErrors = 0
	s.lastSuccessTime = time.Now()
	s.totalInvocations++
}

// RecordError records a failed invocation.
func (t *Tracker) RecordError(name string, errMsg string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	s := t.getOrCreate(name)
	s.consecutiveErrors++
	s.lastError = errMsg
	s.lastErrorTime = time.Now()
	s.totalInvocations++
	s.totalErrors++
}

// IsDegraded returns true if the tool has exceeded the consecutive error threshold.
func (t *Tracker) IsDegraded(name string) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	s, ok := t.tools[name]
	if !ok {
		return false
	}
	return s.consecutiveErrors >= t.threshold
}

// Reset clears the error state for a specific tool.
func (t *Tracker) Reset(name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if s, ok := t.tools[name]; ok {
		s.consecutiveErrors = 0
		s.lastError = ""
	}
}

// ResetAll clears error state for all tools.
func (t *Tracker) ResetAll() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for _, s := range t.tools {
		s.consecutiveErrors = 0
		s.lastError = ""
	}
}

// Status returns the current error budget status for a tool.
func (t *Tracker) Status(name string) ToolStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()
	s, ok := t.tools[name]
	if !ok {
		return ToolStatus{
			ToolName:  name,
			Threshold: t.threshold,
		}
	}
	return ToolStatus{
		ToolName:           name,
		ConsecutiveErrors:  s.consecutiveErrors,
		Threshold:          t.threshold,
		Degraded:           s.consecutiveErrors >= t.threshold,
		LastError:          s.lastError,
		LastErrorTime:      s.lastErrorTime,
		LastSuccessTime:    s.lastSuccessTime,
		TotalInvocations:   s.totalInvocations,
		TotalErrors:        s.totalErrors,
		DegradedSinceCount: s.degradedCount,
	}
}

// AllStatus returns the error budget status for all tracked tools.
func (t *Tracker) AllStatus() []ToolStatus {
	t.mu.RLock()
	defer t.mu.RUnlock()

	statuses := make([]ToolStatus, 0, len(t.tools))
	for name, s := range t.tools {
		statuses = append(statuses, ToolStatus{
			ToolName:           name,
			ConsecutiveErrors:  s.consecutiveErrors,
			Threshold:          t.threshold,
			Degraded:           s.consecutiveErrors >= t.threshold,
			LastError:          s.lastError,
			LastErrorTime:      s.lastErrorTime,
			LastSuccessTime:    s.lastSuccessTime,
			TotalInvocations:   s.totalInvocations,
			TotalErrors:        s.totalErrors,
			DegradedSinceCount: s.degradedCount,
		})
	}
	return statuses
}

// DegradedTools returns the names of tools currently in degraded state.
func (t *Tracker) DegradedTools() []string {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var degraded []string
	for name, s := range t.tools {
		if s.consecutiveErrors >= t.threshold {
			degraded = append(degraded, name)
		}
	}
	return degraded
}

// Wrap returns a handler that checks the error budget before executing.
// If the tool is degraded, it returns a degraded response without calling
// the underlying handler. Otherwise, it invokes the handler and records
// success or failure.
func (t *Tracker) Wrap(toolName string, handler HandlerFunc) HandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Check if tool is degraded
		if t.IsDegraded(toolName) {
			t.mu.Lock()
			s := t.getOrCreate(toolName)
			s.degradedCount++
			s.totalInvocations++
			lastErr := s.lastError
			consec := s.consecutiveErrors
			t.mu.Unlock()

			return &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: fmt.Sprintf(
							"[DEGRADED] %s is temporarily unavailable after %d consecutive errors (threshold: %d). Last error: %s. Use aftrs_errbudget_reset to recover.",
							toolName, consec, t.threshold, lastErr,
						),
					},
				},
				IsError: true,
			}, nil
		}

		// Execute the handler
		result, err := handler(ctx, req)

		// Track the outcome
		if err != nil {
			t.RecordError(toolName, err.Error())
		} else if result != nil && result.IsError {
			// Extract error text from the result
			errText := "unknown error"
			for _, content := range result.Content {
				if tc, ok := content.(mcp.TextContent); ok {
					errText = tc.Text
					break
				}
			}
			t.RecordError(toolName, errText)
		} else {
			t.RecordSuccess(toolName)
		}

		return result, err
	}
}
