package security

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// AuditEventType represents the type of audit event.
type AuditEventType string

const (
	AuditToolCall    AuditEventType = "tool_call"
	AuditToolSuccess AuditEventType = "tool_success"
	AuditToolError   AuditEventType = "tool_error"
	AuditAccessDeny  AuditEventType = "access_denied"
	AuditLogin       AuditEventType = "login"
	AuditSecretRead  AuditEventType = "secret_read"
)

// AuditEvent represents a single audit log entry.
type AuditEvent struct {
	ID        string         `json:"id"`
	Timestamp time.Time      `json:"timestamp"`
	Type      AuditEventType `json:"type"`
	User      string         `json:"user"`
	Tool      string         `json:"tool,omitempty"`
	Action    string         `json:"action,omitempty"`
	Params    map[string]any `json:"params,omitempty"`
	Result    string         `json:"result,omitempty"`
	Error     string         `json:"error,omitempty"`
	Duration  time.Duration  `json:"duration_ms,omitempty"`
	IP        string         `json:"ip,omitempty"`
	SessionID string         `json:"session_id,omitempty"`
}

// AuditLogger handles audit event logging.
type AuditLogger struct {
	mu         sync.Mutex
	events     []AuditEvent
	maxEvents  int
	logFile    string
	writeQueue chan AuditEvent
	stopCh     chan struct{}
}

// NewAuditLogger creates a new audit logger.
func NewAuditLogger(logFile string, maxEvents int) *AuditLogger {
	logger := &AuditLogger{
		events:     make([]AuditEvent, 0, maxEvents),
		maxEvents:  maxEvents,
		logFile:    logFile,
		writeQueue: make(chan AuditEvent, 100),
		stopCh:     make(chan struct{}),
	}

	// Start background writer if log file is specified
	if logFile != "" {
		go logger.backgroundWriter()
	}

	return logger
}

// Global audit logger instance
var GlobalAuditLogger = NewAuditLogger(
	filepath.Join(os.Getenv("HOME"), ".hg-mcp", "audit.log"),
	1000,
)

// Log records an audit event.
func (l *AuditLogger) Log(event AuditEvent) {
	event.Timestamp = time.Now()
	if event.ID == "" {
		event.ID = fmt.Sprintf("%d-%d", event.Timestamp.UnixNano(), time.Now().Nanosecond())
	}

	l.mu.Lock()
	// Add to in-memory buffer (circular)
	if len(l.events) >= l.maxEvents {
		l.events = l.events[1:]
	}
	l.events = append(l.events, event)
	l.mu.Unlock()

	// Queue for file write if enabled
	if l.logFile != "" {
		select {
		case l.writeQueue <- event:
		default:
			// Queue full, drop event
		}
	}
}

// LogToolCall logs a tool invocation.
func (l *AuditLogger) LogToolCall(user, tool string, params map[string]any) {
	l.Log(AuditEvent{
		Type:   AuditToolCall,
		User:   user,
		Tool:   tool,
		Params: sanitizeParams(params),
	})
}

// LogToolResult logs a tool result.
func (l *AuditLogger) LogToolResult(user, tool string, duration time.Duration, err error) {
	event := AuditEvent{
		User:     user,
		Tool:     tool,
		Duration: duration,
	}

	if err != nil {
		event.Type = AuditToolError
		event.Error = err.Error()
	} else {
		event.Type = AuditToolSuccess
		event.Result = "success"
	}

	l.Log(event)
}

// LogAccessDenied logs an access denied event.
func (l *AuditLogger) LogAccessDenied(user, tool, reason string) {
	l.Log(AuditEvent{
		Type:  AuditAccessDeny,
		User:  user,
		Tool:  tool,
		Error: reason,
	})
}

// GetRecentEvents returns recent audit events.
func (l *AuditLogger) GetRecentEvents(limit int) []AuditEvent {
	l.mu.Lock()
	defer l.mu.Unlock()

	if limit <= 0 || limit > len(l.events) {
		limit = len(l.events)
	}

	// Return most recent events
	start := len(l.events) - limit
	result := make([]AuditEvent, limit)
	copy(result, l.events[start:])
	return result
}

// GetEventsByUser returns events for a specific user.
func (l *AuditLogger) GetEventsByUser(user string, limit int) []AuditEvent {
	l.mu.Lock()
	defer l.mu.Unlock()

	var result []AuditEvent
	for i := len(l.events) - 1; i >= 0 && len(result) < limit; i-- {
		if l.events[i].User == user {
			result = append(result, l.events[i])
		}
	}
	return result
}

// GetEventsByTool returns events for a specific tool.
func (l *AuditLogger) GetEventsByTool(tool string, limit int) []AuditEvent {
	l.mu.Lock()
	defer l.mu.Unlock()

	var result []AuditEvent
	for i := len(l.events) - 1; i >= 0 && len(result) < limit; i-- {
		if l.events[i].Tool == tool {
			result = append(result, l.events[i])
		}
	}
	return result
}

// GetErrorEvents returns recent error events.
func (l *AuditLogger) GetErrorEvents(limit int) []AuditEvent {
	l.mu.Lock()
	defer l.mu.Unlock()

	var result []AuditEvent
	for i := len(l.events) - 1; i >= 0 && len(result) < limit; i-- {
		if l.events[i].Type == AuditToolError || l.events[i].Type == AuditAccessDeny {
			result = append(result, l.events[i])
		}
	}
	return result
}

// AuditStats contains summary statistics.
type AuditStats struct {
	TotalEvents    int            `json:"total_events"`
	EventsByType   map[string]int `json:"events_by_type"`
	EventsByUser   map[string]int `json:"events_by_user"`
	ErrorCount     int            `json:"error_count"`
	AccessDenied   int            `json:"access_denied"`
	TopTools       map[string]int `json:"top_tools"`
	AverageLatency time.Duration  `json:"average_latency_ms"`
}

// GetStats returns summary statistics.
func (l *AuditLogger) GetStats() AuditStats {
	l.mu.Lock()
	defer l.mu.Unlock()

	stats := AuditStats{
		TotalEvents:  len(l.events),
		EventsByType: make(map[string]int),
		EventsByUser: make(map[string]int),
		TopTools:     make(map[string]int),
	}

	var totalLatency time.Duration
	var latencyCount int

	for _, e := range l.events {
		stats.EventsByType[string(e.Type)]++
		stats.EventsByUser[e.User]++
		if e.Tool != "" {
			stats.TopTools[e.Tool]++
		}
		if e.Type == AuditToolError {
			stats.ErrorCount++
		}
		if e.Type == AuditAccessDeny {
			stats.AccessDenied++
		}
		if e.Duration > 0 {
			totalLatency += e.Duration
			latencyCount++
		}
	}

	if latencyCount > 0 {
		stats.AverageLatency = totalLatency / time.Duration(latencyCount)
	}

	return stats
}

// backgroundWriter writes events to file asynchronously.
func (l *AuditLogger) backgroundWriter() {
	// Ensure directory exists
	dir := filepath.Dir(l.logFile)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return
	}

	for {
		select {
		case event := <-l.writeQueue:
			l.writeEvent(event)
		case <-l.stopCh:
			return
		}
	}
}

// writeEvent writes a single event to the log file.
func (l *AuditLogger) writeEvent(event AuditEvent) {
	f, err := os.OpenFile(l.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
	if err != nil {
		return
	}
	defer f.Close()

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	f.Write(data)
	f.WriteString("\n")
}

// Close stops the background writer.
func (l *AuditLogger) Close() {
	close(l.stopCh)
}

// sanitizeParams removes sensitive values from parameters.
func sanitizeParams(params map[string]any) map[string]any {
	if params == nil {
		return nil
	}

	sensitiveKeys := []string{
		"password", "secret", "token", "key", "credential",
		"auth", "bearer", "api_key", "private", "oauth",
	}

	result := make(map[string]any)
	for k, v := range params {
		isSensitive := false
		for _, sensitive := range sensitiveKeys {
			if k == sensitive || contains(k, sensitive) {
				isSensitive = true
				break
			}
		}
		if isSensitive {
			result[k] = "[REDACTED]"
		} else {
			result[k] = v
		}
	}
	return result
}

// contains checks if a string contains a substring (case-insensitive).
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		match := true
		for j := 0; j < len(substr); j++ {
			sc := s[i+j]
			uc := substr[j]
			// Simple case-insensitive comparison
			if sc >= 'A' && sc <= 'Z' {
				sc += 32
			}
			if uc >= 'A' && uc <= 'Z' {
				uc += 32
			}
			if sc != uc {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// Convenience functions using global logger

// LogToolInvocation logs a tool call with the global logger.
func LogToolInvocation(ctx context.Context, user, tool string, params map[string]any) {
	GlobalAuditLogger.LogToolCall(user, tool, params)
}

// LogToolCompletion logs a tool result with the global logger.
func LogToolCompletion(ctx context.Context, user, tool string, duration time.Duration, err error) {
	GlobalAuditLogger.LogToolResult(user, tool, duration, err)
}
