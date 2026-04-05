package bot

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/hairglasses-studio/hg-mcp/internal/config"
)

// Session represents an active studio session
type Session struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Operator   string         `json:"operator"`
	StartTime  time.Time      `json:"start_time"`
	EndTime    *time.Time     `json:"end_time,omitempty"`
	Events     []SessionEvent `json:"events"`
	Highlights []string       `json:"highlights"`
	MessageID  string         `json:"message_id,omitempty"` // Discord message ID for updates
}

// SessionEvent represents an event during a session
type SessionEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"` // start, end, stream, recording, error, note, milestone
	Description string    `json:"description"`
	Details     string    `json:"details,omitempty"`
}

// SessionLogger manages studio session logging
type SessionLogger struct {
	mu            sync.RWMutex
	vaultPath     string
	session       *discordgo.Session
	notifier      *Notifier
	activeSession *Session
}

// NewSessionLogger creates a new session logger
func NewSessionLogger(discordSession *discordgo.Session, notifier *Notifier) *SessionLogger {
	return &SessionLogger{
		vaultPath: config.Get().AftrsVaultPath,
		session:   discordSession,
		notifier:  notifier,
	}
}

// StartSession begins a new studio session
func (l *SessionLogger) StartSession(ctx context.Context, name, operator string) (*Session, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.activeSession != nil {
		return nil, fmt.Errorf("session already active: %s", l.activeSession.Name)
	}

	now := time.Now()
	session := &Session{
		ID:         now.Format("20060102-150405"),
		Name:       name,
		Operator:   operator,
		StartTime:  now,
		Events:     []SessionEvent{},
		Highlights: []string{},
	}

	// Add start event
	session.Events = append(session.Events, SessionEvent{
		Timestamp:   now,
		Type:        "start",
		Description: fmt.Sprintf("Session started by %s", operator),
	})

	l.activeSession = session

	// Notify Discord
	if l.notifier != nil {
		l.notifier.NotifySessionStart(name, operator)
	}

	// Save to vault
	if err := l.saveSessionToVault(session); err != nil {
		// Log but don't fail
		fmt.Printf("Warning: failed to save session to vault: %v\n", err)
	}

	return session, nil
}

// EndSession ends the current session
func (l *SessionLogger) EndSession(ctx context.Context) (*Session, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.activeSession == nil {
		return nil, fmt.Errorf("no active session")
	}

	now := time.Now()
	l.activeSession.EndTime = &now

	// Add end event
	l.activeSession.Events = append(l.activeSession.Events, SessionEvent{
		Timestamp:   now,
		Type:        "end",
		Description: "Session ended",
	})

	// Calculate duration
	duration := now.Sub(l.activeSession.StartTime)
	durationStr := formatDuration(duration)

	// Notify Discord
	if l.notifier != nil {
		l.notifier.NotifySessionEnd(
			l.activeSession.Name,
			l.activeSession.Operator,
			durationStr,
			l.activeSession.Highlights,
		)
	}

	// Save final state to vault
	if err := l.saveSessionToVault(l.activeSession); err != nil {
		fmt.Printf("Warning: failed to save session to vault: %v\n", err)
	}

	session := l.activeSession
	l.activeSession = nil

	return session, nil
}

// LogEvent logs an event to the current session
func (l *SessionLogger) LogEvent(eventType, description, details string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.activeSession == nil {
		return fmt.Errorf("no active session")
	}

	event := SessionEvent{
		Timestamp:   time.Now(),
		Type:        eventType,
		Description: description,
		Details:     details,
	}

	l.activeSession.Events = append(l.activeSession.Events, event)

	// Auto-highlight certain event types
	if eventType == "stream" || eventType == "milestone" || eventType == "error" {
		l.activeSession.Highlights = append(l.activeSession.Highlights, description)
	}

	// Save updated session
	return l.saveSessionToVault(l.activeSession)
}

// AddHighlight adds a highlight to the current session
func (l *SessionLogger) AddHighlight(highlight string) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.activeSession == nil {
		return fmt.Errorf("no active session")
	}

	l.activeSession.Highlights = append(l.activeSession.Highlights, highlight)
	return l.saveSessionToVault(l.activeSession)
}

// GetActiveSession returns the current active session
func (l *SessionLogger) GetActiveSession() *Session {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.activeSession
}

// GetSessionSummary generates a summary for a session
func (l *SessionLogger) GetSessionSummary(session *Session) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Session: %s\n\n", session.Name))
	sb.WriteString(fmt.Sprintf("**Operator:** %s\n", session.Operator))
	sb.WriteString(fmt.Sprintf("**Started:** %s\n", session.StartTime.Format("Mon Jan 2, 3:04 PM")))

	if session.EndTime != nil {
		duration := session.EndTime.Sub(session.StartTime)
		sb.WriteString(fmt.Sprintf("**Ended:** %s\n", session.EndTime.Format("Mon Jan 2, 3:04 PM")))
		sb.WriteString(fmt.Sprintf("**Duration:** %s\n", formatDuration(duration)))
	} else {
		duration := time.Since(session.StartTime)
		sb.WriteString(fmt.Sprintf("**Status:** Active (%s)\n", formatDuration(duration)))
	}

	if len(session.Highlights) > 0 {
		sb.WriteString("\n## Highlights\n")
		for _, h := range session.Highlights {
			sb.WriteString(fmt.Sprintf("- %s\n", h))
		}
	}

	if len(session.Events) > 0 {
		sb.WriteString("\n## Timeline\n")
		for _, e := range session.Events {
			sb.WriteString(fmt.Sprintf("- **%s** [%s] %s\n",
				e.Timestamp.Format("15:04"),
				e.Type,
				e.Description,
			))
		}
	}

	return sb.String()
}

// GetDailySummary generates a summary of sessions for a specific date
func (l *SessionLogger) GetDailySummary(ctx context.Context, date time.Time) (string, error) {
	dateStr := date.Format("2006-01-02")
	sessionsPath := filepath.Join(l.vaultPath, "sessions", dateStr)

	if _, err := os.Stat(sessionsPath); os.IsNotExist(err) {
		return fmt.Sprintf("No sessions found for %s", date.Format("Mon Jan 2, 2006")), nil
	}

	files, err := os.ReadDir(sessionsPath)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# Daily Summary: %s\n\n", date.Format("Monday, January 2, 2006")))

	sessionCount := 0
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}
		sessionCount++

		content, err := os.ReadFile(filepath.Join(sessionsPath, file.Name()))
		if err != nil {
			continue
		}

		sb.WriteString(string(content))
		sb.WriteString("\n---\n\n")
	}

	if sessionCount == 0 {
		return fmt.Sprintf("No sessions found for %s", date.Format("Mon Jan 2, 2006")), nil
	}

	sb.WriteString(fmt.Sprintf("\n**Total Sessions:** %d\n", sessionCount))

	return sb.String(), nil
}

// GetWeeklySummary generates a summary of sessions for the past week
func (l *SessionLogger) GetWeeklySummary(ctx context.Context) (string, error) {
	var sb strings.Builder
	sb.WriteString("# Weekly Session Summary\n\n")

	now := time.Now()
	totalSessions := 0

	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		sessionsPath := filepath.Join(l.vaultPath, "sessions", dateStr)

		if _, err := os.Stat(sessionsPath); os.IsNotExist(err) {
			continue
		}

		files, err := os.ReadDir(sessionsPath)
		if err != nil {
			continue
		}

		sessionCount := 0
		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
				sessionCount++
			}
		}

		if sessionCount > 0 {
			sb.WriteString(fmt.Sprintf("- **%s:** %d session(s)\n",
				date.Format("Mon Jan 2"),
				sessionCount,
			))
			totalSessions += sessionCount
		}
	}

	if totalSessions == 0 {
		return "No sessions recorded in the past week.", nil
	}

	sb.WriteString(fmt.Sprintf("\n**Total:** %d sessions this week\n", totalSessions))

	return sb.String(), nil
}

// saveSessionToVault saves the session to the vault
func (l *SessionLogger) saveSessionToVault(session *Session) error {
	// Ensure sessions directory exists
	dateStr := session.StartTime.Format("2006-01-02")
	sessionsPath := filepath.Join(l.vaultPath, "sessions", dateStr)
	if err := os.MkdirAll(sessionsPath, 0755); err != nil {
		return err
	}

	// Generate markdown content
	content := l.GetSessionSummary(session)

	// Save to file
	filename := fmt.Sprintf("%s-%s.md", session.ID, sanitizeFilename(session.Name))
	filePath := filepath.Join(sessionsPath, filename)

	return os.WriteFile(filePath, []byte(content), 0644)
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60

	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}

// sanitizeFilename removes or replaces invalid filename characters
func sanitizeFilename(name string) string {
	// Replace spaces and special chars
	replacer := strings.NewReplacer(
		" ", "-",
		"/", "-",
		"\\", "-",
		":", "-",
		"*", "",
		"?", "",
		"\"", "",
		"<", "",
		">", "",
		"|", "",
	)
	return strings.ToLower(replacer.Replace(name))
}
