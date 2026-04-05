// Package clients provides API clients for external services.
package clients

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/config"
)

// VaultClient provides access to the Obsidian vault for knowledge management
type VaultClient struct {
	vaultPath string
}

// VaultDocument represents a document in the vault
type VaultDocument struct {
	Path     string    `json:"path"`
	Name     string    `json:"name"`
	Content  string    `json:"content,omitempty"`
	Tags     []string  `json:"tags,omitempty"`
	Modified time.Time `json:"modified"`
	Size     int64     `json:"size"`
}

// VaultSearchResult represents a search result
type VaultSearchResult struct {
	Document VaultDocument `json:"document"`
	Score    float64       `json:"score"`
	Snippet  string        `json:"snippet"`
}

// ProjectNotes represents notes for a project
type ProjectNotes struct {
	Project     string          `json:"project"`
	Description string          `json:"description,omitempty"`
	Documents   []VaultDocument `json:"documents"`
	LastUpdated time.Time       `json:"last_updated"`
}

// Setlist represents a performance setlist
type Setlist struct {
	Name  string        `json:"name"`
	Date  time.Time     `json:"date"`
	Venue string        `json:"venue,omitempty"`
	Items []SetlistItem `json:"items"`
	Notes string        `json:"notes,omitempty"`
}

// SetlistItem represents an item in a setlist
type SetlistItem struct {
	Order    int    `json:"order"`
	Name     string `json:"name"`
	Duration string `json:"duration,omitempty"`
	Notes    string `json:"notes,omitempty"`
	Cue      string `json:"cue,omitempty"`
}

// Runbook represents an operational runbook
type Runbook struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Category    string   `json:"category"`
	Description string   `json:"description,omitempty"`
	Steps       []string `json:"steps,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

// ShowEvent represents a logged show event
type ShowEvent struct {
	Timestamp   time.Time `json:"timestamp"`
	EventType   string    `json:"event_type"` // start, end, cue, issue, note
	Description string    `json:"description"`
	Details     string    `json:"details,omitempty"`
}

// ShowHistory represents historical show data
type ShowHistory struct {
	Name     string      `json:"name"`
	Date     time.Time   `json:"date"`
	Venue    string      `json:"venue,omitempty"`
	Duration string      `json:"duration,omitempty"`
	Events   []ShowEvent `json:"events,omitempty"`
	Notes    string      `json:"notes,omitempty"`
}

// StudioSession represents a studio work session
type StudioSession struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Operator   string     `json:"operator"`
	StartTime  time.Time  `json:"start_time"`
	EndTime    *time.Time `json:"end_time,omitempty"`
	Duration   string     `json:"duration,omitempty"`
	Highlights []string   `json:"highlights,omitempty"`
	Path       string     `json:"path"`
}

// SessionSummary represents aggregated session data
type SessionSummary struct {
	Period        string          `json:"period"` // daily, weekly, monthly
	StartDate     time.Time       `json:"start_date"`
	EndDate       time.Time       `json:"end_date"`
	TotalSessions int             `json:"total_sessions"`
	TotalDuration string          `json:"total_duration,omitempty"`
	Sessions      []StudioSession `json:"sessions"`
}

// NewVaultClient creates a new vault client
func NewVaultClient() (*VaultClient, error) {
	vaultPath := config.Get().AftrsVaultPath

	return &VaultClient{
		vaultPath: vaultPath,
	}, nil
}

// VaultPath returns the configured vault path
func (c *VaultClient) VaultPath() string {
	return c.vaultPath
}

// VaultExists checks if the vault directory exists
func (c *VaultClient) VaultExists() bool {
	_, err := os.Stat(c.vaultPath)
	return err == nil
}

// EnsureVaultStructure creates the vault directory structure if it doesn't exist
func (c *VaultClient) EnsureVaultStructure(ctx context.Context) error {
	dirs := []string{
		"projects",
		"sessions",
		"assets",
		"learnings",
		"gaming",
		"runbooks",
		"daily",
		"shows",
		"setlists",
	}

	for _, dir := range dirs {
		path := filepath.Join(c.vaultPath, dir)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("failed to create %s: %w", dir, err)
		}
	}

	return nil
}

// Search searches the vault for documents matching the query
func (c *VaultClient) Search(ctx context.Context, query string) ([]VaultSearchResult, error) {
	results := []VaultSearchResult{}

	if !c.VaultExists() {
		return results, nil
	}

	query = strings.ToLower(query)

	// Walk through vault and search markdown files
	err := filepath.Walk(c.vaultPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Read file content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentStr := string(content)
		contentLower := strings.ToLower(contentStr)

		// Check if query matches
		if strings.Contains(contentLower, query) || strings.Contains(strings.ToLower(info.Name()), query) {
			// Extract snippet
			snippet := extractSnippet(contentStr, query, 100)

			relPath, _ := filepath.Rel(c.vaultPath, path)
			results = append(results, VaultSearchResult{
				Document: VaultDocument{
					Path:     relPath,
					Name:     strings.TrimSuffix(info.Name(), ".md"),
					Modified: info.ModTime(),
					Size:     info.Size(),
				},
				Score:   1.0,
				Snippet: snippet,
			})
		}

		return nil
	})

	if err != nil {
		return results, err
	}

	return results, nil
}

// extractSnippet extracts a snippet around the query match
func extractSnippet(content, query string, length int) string {
	contentLower := strings.ToLower(content)
	queryLower := strings.ToLower(query)

	idx := strings.Index(contentLower, queryLower)
	if idx == -1 {
		// Return first part of content
		if len(content) > length {
			return content[:length] + "..."
		}
		return content
	}

	start := idx - length/2
	if start < 0 {
		start = 0
	}

	end := idx + len(query) + length/2
	if end > len(content) {
		end = len(content)
	}

	snippet := content[start:end]
	if start > 0 {
		snippet = "..." + snippet
	}
	if end < len(content) {
		snippet = snippet + "..."
	}

	return snippet
}

// SaveDocument saves a document to the vault
func (c *VaultClient) SaveDocument(ctx context.Context, path, content string) error {
	if !c.VaultExists() {
		if err := c.EnsureVaultStructure(ctx); err != nil {
			return err
		}
	}

	fullPath := filepath.Join(c.vaultPath, path)

	// Ensure parent directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Ensure .md extension
	if !strings.HasSuffix(fullPath, ".md") {
		fullPath += ".md"
	}

	return os.WriteFile(fullPath, []byte(content), 0644)
}

// GetProjectNotes returns notes for a specific project
func (c *VaultClient) GetProjectNotes(ctx context.Context, project string) (*ProjectNotes, error) {
	notes := &ProjectNotes{
		Project:   project,
		Documents: []VaultDocument{},
	}

	projectPath := filepath.Join(c.vaultPath, "projects", project)
	if _, err := os.Stat(projectPath); os.IsNotExist(err) {
		return notes, nil
	}

	// List all markdown files in the project directory
	files, err := os.ReadDir(projectPath)
	if err != nil {
		return notes, err
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		doc := VaultDocument{
			Path:     filepath.Join("projects", project, file.Name()),
			Name:     strings.TrimSuffix(file.Name(), ".md"),
			Modified: info.ModTime(),
			Size:     info.Size(),
		}

		notes.Documents = append(notes.Documents, doc)

		if info.ModTime().After(notes.LastUpdated) {
			notes.LastUpdated = info.ModTime()
		}
	}

	return notes, nil
}

// ListSetlists returns all setlists
func (c *VaultClient) ListSetlists(ctx context.Context) ([]Setlist, error) {
	setlists := []Setlist{}

	setlistPath := filepath.Join(c.vaultPath, "setlists")
	if _, err := os.Stat(setlistPath); os.IsNotExist(err) {
		return setlists, nil
	}

	files, err := os.ReadDir(setlistPath)
	if err != nil {
		return setlists, err
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		setlists = append(setlists, Setlist{
			Name: strings.TrimSuffix(file.Name(), ".md"),
			Date: info.ModTime(),
		})
	}

	return setlists, nil
}

// GetSetlist returns a specific setlist
func (c *VaultClient) GetSetlist(ctx context.Context, name string) (*Setlist, error) {
	setlistPath := filepath.Join(c.vaultPath, "setlists", name+".md")

	content, err := os.ReadFile(setlistPath)
	if err != nil {
		return nil, fmt.Errorf("setlist not found: %s", name)
	}

	info, _ := os.Stat(setlistPath)

	setlist := &Setlist{
		Name:  name,
		Date:  info.ModTime(),
		Notes: string(content),
	}

	// Parse items from content (simple line-based parsing)
	lines := strings.Split(string(content), "\n")
	order := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			order++
			setlist.Items = append(setlist.Items, SetlistItem{
				Order: order,
				Name:  strings.TrimPrefix(strings.TrimPrefix(line, "- "), "* "),
			})
		}
	}

	return setlist, nil
}

// SearchRunbooks searches for runbooks
func (c *VaultClient) SearchRunbooks(ctx context.Context, query string) ([]Runbook, error) {
	runbooks := []Runbook{}

	runbookPath := filepath.Join(c.vaultPath, "runbooks")
	if _, err := os.Stat(runbookPath); os.IsNotExist(err) {
		return runbooks, nil
	}

	query = strings.ToLower(query)

	err := filepath.Walk(runbookPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		name := strings.TrimSuffix(info.Name(), ".md")
		if strings.Contains(strings.ToLower(name), query) {
			relPath, _ := filepath.Rel(c.vaultPath, path)
			category := filepath.Dir(relPath)
			if category == "runbooks" {
				category = "general"
			} else {
				category = strings.TrimPrefix(category, "runbooks/")
			}

			runbooks = append(runbooks, Runbook{
				Name:     name,
				Path:     relPath,
				Category: category,
			})
		}

		return nil
	})

	if err != nil {
		return runbooks, err
	}

	return runbooks, nil
}

// LogShowEvent logs an event for the current show
func (c *VaultClient) LogShowEvent(ctx context.Context, eventType, description, details string) error {
	if !c.VaultExists() {
		if err := c.EnsureVaultStructure(ctx); err != nil {
			return err
		}
	}

	today := time.Now().Format("2006-01-02")
	showPath := filepath.Join(c.vaultPath, "shows", today+".md")

	// Read existing content or create new
	var content string
	if data, err := os.ReadFile(showPath); err == nil {
		content = string(data)
	} else {
		content = fmt.Sprintf("# Show Log: %s\n\n", today)
	}

	// Append new event
	timestamp := time.Now().Format("15:04:05")
	event := fmt.Sprintf("## %s - %s\n\n%s\n\n", timestamp, eventType, description)
	if details != "" {
		event += fmt.Sprintf("**Details:** %s\n\n", details)
	}

	content += event

	return os.WriteFile(showPath, []byte(content), 0644)
}

// GetShowHistory returns historical show data
func (c *VaultClient) GetShowHistory(ctx context.Context, limit int) ([]ShowHistory, error) {
	history := []ShowHistory{}

	showsPath := filepath.Join(c.vaultPath, "shows")
	if _, err := os.Stat(showsPath); os.IsNotExist(err) {
		return history, nil
	}

	files, err := os.ReadDir(showsPath)
	if err != nil {
		return history, err
	}

	// Process files in reverse order (newest first)
	count := 0
	for i := len(files) - 1; i >= 0 && count < limit; i-- {
		file := files[i]
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		name := strings.TrimSuffix(file.Name(), ".md")
		date, err := time.Parse("2006-01-02", name)
		if err != nil {
			continue
		}

		info, _ := file.Info()

		history = append(history, ShowHistory{
			Name:  name,
			Date:  date,
			Notes: fmt.Sprintf("Show log for %s (%d bytes)", name, info.Size()),
		})
		count++
	}

	return history, nil
}

// ListSessions returns sessions for a specific date
func (c *VaultClient) ListSessions(ctx context.Context, date time.Time) ([]StudioSession, error) {
	sessions := []StudioSession{}

	dateStr := date.Format("2006-01-02")
	sessionsPath := filepath.Join(c.vaultPath, "sessions", dateStr)

	if _, err := os.Stat(sessionsPath); os.IsNotExist(err) {
		return sessions, nil
	}

	files, err := os.ReadDir(sessionsPath)
	if err != nil {
		return sessions, err
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
			continue
		}

		info, err := file.Info()
		if err != nil {
			continue
		}

		// Parse session info from filename (ID-name.md)
		name := strings.TrimSuffix(file.Name(), ".md")
		parts := strings.SplitN(name, "-", 2)
		sessionID := name
		sessionName := name
		if len(parts) == 2 {
			sessionID = parts[0]
			sessionName = parts[1]
		}

		sessions = append(sessions, StudioSession{
			ID:        sessionID,
			Name:      sessionName,
			StartTime: info.ModTime(),
			Path:      filepath.Join("sessions", dateStr, file.Name()),
		})
	}

	return sessions, nil
}

// GetRecentSessions returns the most recent sessions
func (c *VaultClient) GetRecentSessions(ctx context.Context, limit int) ([]StudioSession, error) {
	sessions := []StudioSession{}

	sessionsPath := filepath.Join(c.vaultPath, "sessions")
	if _, err := os.Stat(sessionsPath); os.IsNotExist(err) {
		return sessions, nil
	}

	// List date directories
	dateDirs, err := os.ReadDir(sessionsPath)
	if err != nil {
		return sessions, err
	}

	// Process in reverse order (newest first)
	for i := len(dateDirs) - 1; i >= 0 && len(sessions) < limit; i-- {
		dir := dateDirs[i]
		if !dir.IsDir() {
			continue
		}

		// Parse date from directory name
		_, err := time.Parse("2006-01-02", dir.Name())
		if err != nil {
			continue
		}

		dayPath := filepath.Join(sessionsPath, dir.Name())
		files, err := os.ReadDir(dayPath)
		if err != nil {
			continue
		}

		for _, file := range files {
			if len(sessions) >= limit {
				break
			}
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
				continue
			}

			info, err := file.Info()
			if err != nil {
				continue
			}

			name := strings.TrimSuffix(file.Name(), ".md")
			sessions = append(sessions, StudioSession{
				ID:        name,
				Name:      name,
				StartTime: info.ModTime(),
				Path:      filepath.Join("sessions", dir.Name(), file.Name()),
			})
		}
	}

	return sessions, nil
}

// GetSessionContent reads the content of a session file
func (c *VaultClient) GetSessionContent(ctx context.Context, path string) (string, error) {
	fullPath := filepath.Join(c.vaultPath, path)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("session not found: %s", path)
	}
	return string(content), nil
}

// GetSessionSummary returns aggregated session data for a period
func (c *VaultClient) GetSessionSummary(ctx context.Context, period string) (*SessionSummary, error) {
	summary := &SessionSummary{
		Period:   period,
		Sessions: []StudioSession{},
	}

	now := time.Now()
	var startDate time.Time

	switch period {
	case "daily":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		summary.EndDate = now
	case "weekly":
		startDate = now.AddDate(0, 0, -7)
		summary.EndDate = now
	case "monthly":
		startDate = now.AddDate(0, -1, 0)
		summary.EndDate = now
	default:
		startDate = now.AddDate(0, 0, -7) // Default to weekly
		summary.Period = "weekly"
		summary.EndDate = now
	}
	summary.StartDate = startDate

	sessionsPath := filepath.Join(c.vaultPath, "sessions")
	if _, err := os.Stat(sessionsPath); os.IsNotExist(err) {
		return summary, nil
	}

	dateDirs, err := os.ReadDir(sessionsPath)
	if err != nil {
		return summary, err
	}

	for _, dir := range dateDirs {
		if !dir.IsDir() {
			continue
		}

		date, err := time.Parse("2006-01-02", dir.Name())
		if err != nil {
			continue
		}

		if date.Before(startDate) || date.After(summary.EndDate) {
			continue
		}

		dayPath := filepath.Join(sessionsPath, dir.Name())
		files, err := os.ReadDir(dayPath)
		if err != nil {
			continue
		}

		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") {
				continue
			}

			info, err := file.Info()
			if err != nil {
				continue
			}

			name := strings.TrimSuffix(file.Name(), ".md")
			summary.Sessions = append(summary.Sessions, StudioSession{
				ID:        name,
				Name:      name,
				StartTime: info.ModTime(),
				Path:      filepath.Join("sessions", dir.Name(), file.Name()),
			})
			summary.TotalSessions++
		}
	}

	return summary, nil
}
