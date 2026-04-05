// Package clients provides API clients for external services.
// session_memory.go implements session memory for team knowledge and insights
package clients

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SessionMemoryClient manages team knowledge and session insights
type SessionMemoryClient struct {
	mu           sync.RWMutex
	vaultPath    string
	userMemories []*UserMemory
	insights     []*SessionInsight
	memoryIndex  map[string]int // memory_id -> slice index
	insightIndex map[string]int // insight_id -> slice index
}

// UserMemory represents team knowledge (#remember pattern)
type UserMemory struct {
	ID       string    `json:"id"`
	Content  string    `json:"content"`  // "Resolume needs restart after 4 hours"
	Category string    `json:"category"` // equipment, workflow, venue, show
	Tags     []string  `json:"tags,omitempty"`
	AddedBy  string    `json:"added_by,omitempty"`
	AddedAt  time.Time `json:"added_at"`
}

// SessionInsight captures learnings from a troubleshooting session
type SessionInsight struct {
	ID        string `json:"id"`
	SessionID string `json:"session_id,omitempty"`

	// What was observed
	Symptoms []string `json:"symptoms"`

	// Resolution path
	StepsWorked []InsightStep `json:"steps_worked"`

	// Context
	RootCause string            `json:"root_cause,omitempty"`
	Equipment []string          `json:"equipment,omitempty"`
	Venue     string            `json:"venue,omitempty"`
	Show      string            `json:"show,omitempty"`
	Context   map[string]string `json:"context,omitempty"`

	// Quality
	QualityScore int      `json:"quality_score"` // 1-5
	Pitfalls     []string `json:"pitfalls,omitempty"`

	// Metadata
	GeneratedAt time.Time `json:"generated_at"`
	Summary     string    `json:"summary,omitempty"`
}

// InsightStep represents a resolution step
type InsightStep struct {
	Tool        string `json:"tool,omitempty"`
	Action      string `json:"action"`
	Description string `json:"description"`
	Outcome     string `json:"outcome"` // success, failed, partial
	Order       int    `json:"order"`
}

// MemorySearchResult contains search results
type MemorySearchResult struct {
	Memories     []*ScoredMemory  `json:"memories,omitempty"`
	Insights     []*ScoredInsight `json:"insights,omitempty"`
	TotalResults int              `json:"total_results"`
}

// ScoredMemory is a memory with relevance score
type ScoredMemory struct {
	Memory *UserMemory `json:"memory"`
	Score  float64     `json:"score"`
}

// ScoredInsight is an insight with relevance score
type ScoredInsight struct {
	Insight *SessionInsight `json:"insight"`
	Score   float64         `json:"score"`
}

// RelevantContext for loading memories related to a topic
type RelevantContext struct {
	Memories       []*UserMemory     `json:"memories"`
	PastInsights   []*SessionInsight `json:"past_insights"`
	SuggestedSteps []string          `json:"suggested_steps"`
	Warnings       []string          `json:"warnings"`
}

var (
	sessionMemoryOnce     sync.Once
	sessionMemoryInstance *SessionMemoryClient
)

// GetSessionMemoryClient returns the singleton session memory client
func GetSessionMemoryClient() *SessionMemoryClient {
	sessionMemoryOnce.Do(func() {
		sessionMemoryInstance, _ = NewSessionMemoryClient()
	})
	return sessionMemoryInstance
}

// NewSessionMemoryClient creates a new session memory client
func NewSessionMemoryClient() (*SessionMemoryClient, error) {
	vaultPath := os.Getenv("OBSIDIAN_VAULT_PATH")
	if vaultPath == "" {
		vaultPath = filepath.Join(os.Getenv("HOME"), "Documents", "obsidian-vault")
	}

	client := &SessionMemoryClient{
		vaultPath:    vaultPath,
		userMemories: make([]*UserMemory, 0),
		insights:     make([]*SessionInsight, 0),
		memoryIndex:  make(map[string]int),
		insightIndex: make(map[string]int),
	}

	// Load from vault
	client.loadFromVault()

	return client, nil
}

// Remember saves a new user memory
func (c *SessionMemoryClient) Remember(content, category string, tags []string) (*UserMemory, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	memory := &UserMemory{
		ID:       uuid.New().String()[:8],
		Content:  content,
		Category: category,
		Tags:     tags,
		AddedAt:  time.Now(),
	}

	c.userMemories = append(c.userMemories, memory)
	c.memoryIndex[memory.ID] = len(c.userMemories) - 1

	// Persist to vault
	if err := c.saveMemoriesToVault(); err != nil {
		return memory, fmt.Errorf("saved to memory but vault persist failed: %w", err)
	}

	return memory, nil
}

// Forget removes a memory by ID
func (c *SessionMemoryClient) Forget(memoryID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	idx, exists := c.memoryIndex[memoryID]
	if !exists {
		return fmt.Errorf("memory not found: %s", memoryID)
	}

	// Remove from slice
	c.userMemories = append(c.userMemories[:idx], c.userMemories[idx+1:]...)

	// Rebuild index
	c.memoryIndex = make(map[string]int)
	for i, m := range c.userMemories {
		c.memoryIndex[m.ID] = i
	}

	return c.saveMemoriesToVault()
}

// ListMemories returns all memories, optionally filtered by category
func (c *SessionMemoryClient) ListMemories(category string) []*UserMemory {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if category == "" {
		result := make([]*UserMemory, len(c.userMemories))
		copy(result, c.userMemories)
		return result
	}

	var result []*UserMemory
	for _, m := range c.userMemories {
		if strings.EqualFold(m.Category, category) {
			result = append(result, m)
		}
	}
	return result
}

// Search searches across memories and insights
func (c *SessionMemoryClient) Search(query string, limit int) *MemorySearchResult {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if limit <= 0 {
		limit = 10
	}

	queryLower := strings.ToLower(query)
	queryTerms := strings.Fields(queryLower)

	// Score memories
	var scoredMemories []*ScoredMemory
	for _, m := range c.userMemories {
		score := c.scoreMemory(m, queryTerms)
		if score > 0 {
			scoredMemories = append(scoredMemories, &ScoredMemory{
				Memory: m,
				Score:  score,
			})
		}
	}

	// Score insights
	var scoredInsights []*ScoredInsight
	for _, i := range c.insights {
		score := c.scoreInsight(i, queryTerms)
		if score > 0 {
			scoredInsights = append(scoredInsights, &ScoredInsight{
				Insight: i,
				Score:   score,
			})
		}
	}

	// Sort by score descending
	sort.Slice(scoredMemories, func(i, j int) bool {
		return scoredMemories[i].Score > scoredMemories[j].Score
	})
	sort.Slice(scoredInsights, func(i, j int) bool {
		return scoredInsights[i].Score > scoredInsights[j].Score
	})

	// Apply limit
	if len(scoredMemories) > limit {
		scoredMemories = scoredMemories[:limit]
	}
	if len(scoredInsights) > limit {
		scoredInsights = scoredInsights[:limit]
	}

	return &MemorySearchResult{
		Memories:     scoredMemories,
		Insights:     scoredInsights,
		TotalResults: len(scoredMemories) + len(scoredInsights),
	}
}

// scoreMemory calculates relevance score for a memory
func (c *SessionMemoryClient) scoreMemory(m *UserMemory, queryTerms []string) float64 {
	var score float64
	contentLower := strings.ToLower(m.Content)
	categoryLower := strings.ToLower(m.Category)

	for _, term := range queryTerms {
		if strings.Contains(contentLower, term) {
			score += 1.0
		}
		if strings.Contains(categoryLower, term) {
			score += 0.5
		}
		for _, tag := range m.Tags {
			if strings.Contains(strings.ToLower(tag), term) {
				score += 0.3
			}
		}
	}

	return score
}

// scoreInsight calculates relevance score for an insight
func (c *SessionMemoryClient) scoreInsight(i *SessionInsight, queryTerms []string) float64 {
	var score float64

	for _, term := range queryTerms {
		// Check symptoms
		for _, s := range i.Symptoms {
			if strings.Contains(strings.ToLower(s), term) {
				score += 1.0
			}
		}
		// Check root cause
		if strings.Contains(strings.ToLower(i.RootCause), term) {
			score += 1.5
		}
		// Check equipment
		for _, e := range i.Equipment {
			if strings.Contains(strings.ToLower(e), term) {
				score += 0.5
			}
		}
		// Check summary
		if strings.Contains(strings.ToLower(i.Summary), term) {
			score += 0.8
		}
	}

	return score
}

// SaveInsight saves a new session insight
func (c *SessionMemoryClient) SaveInsight(insight *SessionInsight) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if insight.ID == "" {
		insight.ID = uuid.New().String()[:8]
	}
	if insight.GeneratedAt.IsZero() {
		insight.GeneratedAt = time.Now()
	}

	c.insights = append(c.insights, insight)
	c.insightIndex[insight.ID] = len(c.insights) - 1

	return c.saveInsightsToVault()
}

// ListInsights returns insights, optionally filtered
func (c *SessionMemoryClient) ListInsights(equipment, venue string, limit int) []*SessionInsight {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if limit <= 0 {
		limit = 20
	}

	var result []*SessionInsight
	for _, i := range c.insights {
		if equipment != "" {
			found := false
			for _, e := range i.Equipment {
				if strings.EqualFold(e, equipment) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		if venue != "" && !strings.EqualFold(i.Venue, venue) {
			continue
		}
		result = append(result, i)
	}

	// Sort by date descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].GeneratedAt.After(result[j].GeneratedAt)
	})

	if len(result) > limit {
		result = result[:limit]
	}

	return result
}

// LoadContext returns relevant memories and insights for a topic
func (c *SessionMemoryClient) LoadContext(topic string, equipment []string) *RelevantContext {
	searchResult := c.Search(topic, 5)

	ctx := &RelevantContext{
		Memories:       make([]*UserMemory, 0),
		PastInsights:   make([]*SessionInsight, 0),
		SuggestedSteps: make([]string, 0),
		Warnings:       make([]string, 0),
	}

	// Extract memories
	for _, sm := range searchResult.Memories {
		ctx.Memories = append(ctx.Memories, sm.Memory)
	}

	// Extract insights
	for _, si := range searchResult.Insights {
		ctx.PastInsights = append(ctx.PastInsights, si.Insight)

		// Extract suggested steps from past successful resolutions
		for _, step := range si.Insight.StepsWorked {
			if step.Outcome == "success" {
				ctx.SuggestedSteps = append(ctx.SuggestedSteps, step.Description)
			}
		}

		// Extract warnings from pitfalls
		ctx.Warnings = append(ctx.Warnings, si.Insight.Pitfalls...)
	}

	// Search by equipment if provided
	for _, eq := range equipment {
		eqResults := c.Search(eq, 3)
		for _, si := range eqResults.Insights {
			// Add unique pitfalls as warnings
			for _, p := range si.Insight.Pitfalls {
				if !containsString(ctx.Warnings, p) {
					ctx.Warnings = append(ctx.Warnings, p)
				}
			}
		}
	}

	return ctx
}

// GetStats returns memory statistics
func (c *SessionMemoryClient) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Count by category
	categoryCounts := make(map[string]int)
	for _, m := range c.userMemories {
		categoryCounts[m.Category]++
	}

	return map[string]interface{}{
		"total_memories":    len(c.userMemories),
		"total_insights":    len(c.insights),
		"memory_categories": categoryCounts,
		"vault_path":        c.vaultPath,
	}
}

// loadFromVault loads memories and insights from vault
func (c *SessionMemoryClient) loadFromVault() {
	memoriesPath := filepath.Join(c.vaultPath, "memory", "user-memories.json")
	insightsPath := filepath.Join(c.vaultPath, "memory", "session-insights.json")

	// Load memories
	if data, err := os.ReadFile(memoriesPath); err == nil {
		var memories []*UserMemory
		if json.Unmarshal(data, &memories) == nil {
			c.userMemories = memories
			for i, m := range memories {
				c.memoryIndex[m.ID] = i
			}
		}
	}

	// Load insights
	if data, err := os.ReadFile(insightsPath); err == nil {
		var insights []*SessionInsight
		if json.Unmarshal(data, &insights) == nil {
			c.insights = insights
			for i, ins := range insights {
				c.insightIndex[ins.ID] = i
			}
		}
	}
}

// saveMemoriesToVault persists memories to vault
func (c *SessionMemoryClient) saveMemoriesToVault() error {
	memoryDir := filepath.Join(c.vaultPath, "memory")
	if err := os.MkdirAll(memoryDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c.userMemories, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(memoryDir, "user-memories.json"), data, 0644)
}

// saveInsightsToVault persists insights to vault
func (c *SessionMemoryClient) saveInsightsToVault() error {
	memoryDir := filepath.Join(c.vaultPath, "memory")
	if err := os.MkdirAll(memoryDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c.insights, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(memoryDir, "session-insights.json"), data, 0644)
}

// containsString checks if a string slice contains a value
func containsString(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
