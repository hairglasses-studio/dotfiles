// Package clients provides API clients for external services.
// research_swarm.go implements autonomous pattern discovery from tool usage
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

	"github.com/hairglasses-studio/hg-mcp/internal/config"
	"github.com/google/uuid"
)

// ResearchSwarmClient manages pattern discovery and improvement suggestions
type ResearchSwarmClient struct {
	mu           sync.RWMutex
	vaultPath    string
	usageHistory []*ToolUsage
	patterns     []*UsagePattern
	suggestions  []*ImprovementSuggestion
	workers      map[string]*SwarmWorker
	config       *SwarmConfig
	stats        *SwarmStats
}

// ToolUsage records a single tool invocation
type ToolUsage struct {
	ID        string                 `json:"id"`
	Tool      string                 `json:"tool"`
	Category  string                 `json:"category"`
	Args      map[string]interface{} `json:"args,omitempty"`
	Success   bool                   `json:"success"`
	Duration  time.Duration          `json:"duration_ms"`
	Timestamp time.Time              `json:"timestamp"`
	SessionID string                 `json:"session_id,omitempty"`
	Context   map[string]string      `json:"context,omitempty"`
	PrevTool  string                 `json:"prev_tool,omitempty"`
	NextTool  string                 `json:"next_tool,omitempty"`
}

// UsagePattern represents a discovered pattern
type UsagePattern struct {
	ID           string            `json:"id"`
	Type         string            `json:"type"` // sequential, co-occurring, error_recovery, optimization
	Tools        []string          `json:"tools"`
	Frequency    int               `json:"frequency"`
	Confidence   float64           `json:"confidence"`
	Description  string            `json:"description"`
	DiscoveredAt time.Time         `json:"discovered_at"`
	LastSeen     time.Time         `json:"last_seen"`
	Context      map[string]string `json:"context,omitempty"`
}

// ImprovementSuggestion represents a suggested improvement
type ImprovementSuggestion struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"` // consolidation, chain, alias, new_tool
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Priority    int       `json:"priority"` // 1-5, 5 being highest
	Impact      string    `json:"impact"`   // high, medium, low
	Effort      string    `json:"effort"`   // high, medium, low
	Status      string    `json:"status"`   // new, reviewed, accepted, rejected, implemented
	Pattern     string    `json:"pattern_id,omitempty"`
	Tools       []string  `json:"tools,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	ReviewedAt  time.Time `json:"reviewed_at,omitempty"`
	ReviewedBy  string    `json:"reviewed_by,omitempty"`
}

// SwarmWorker represents a background analysis worker
type SwarmWorker struct {
	ID       string    `json:"id"`
	Type     string    `json:"type"`   // pattern_miner, sequence_analyzer, error_analyzer, optimization_finder
	Status   string    `json:"status"` // idle, running, completed, error
	LastRun  time.Time `json:"last_run"`
	Findings int       `json:"findings"`
	Errors   int       `json:"errors"`
}

// SwarmConfig configures the research swarm
type SwarmConfig struct {
	Enabled             bool          `json:"enabled"`
	MinPatternFrequency int           `json:"min_pattern_frequency"` // Minimum occurrences to consider a pattern
	AnalysisInterval    time.Duration `json:"analysis_interval"`
	MaxHistorySize      int           `json:"max_history_size"`
	AutoSuggest         bool          `json:"auto_suggest"`
}

// SwarmStats contains statistics about the research swarm
type SwarmStats struct {
	TotalUsage        int            `json:"total_usage"`
	TotalPatterns     int            `json:"total_patterns"`
	TotalSuggestions  int            `json:"total_suggestions"`
	PatternsByType    map[string]int `json:"patterns_by_type"`
	SuggestionsByType map[string]int `json:"suggestions_by_type"`
	TopTools          []ToolStat     `json:"top_tools"`
	TopSequences      []SequenceStat `json:"top_sequences"`
	LastAnalysis      time.Time      `json:"last_analysis"`
}

// ToolStat represents tool usage statistics
type ToolStat struct {
	Tool        string  `json:"tool"`
	Count       int     `json:"count"`
	SuccessRate float64 `json:"success_rate"`
	AvgDuration int64   `json:"avg_duration_ms"`
}

// SequenceStat represents a frequently used tool sequence
type SequenceStat struct {
	Sequence []string `json:"sequence"`
	Count    int      `json:"count"`
}

var (
	researchSwarmOnce     sync.Once
	researchSwarmInstance *ResearchSwarmClient
)

// GetResearchSwarmClient returns the singleton research swarm client
func GetResearchSwarmClient() *ResearchSwarmClient {
	researchSwarmOnce.Do(func() {
		researchSwarmInstance, _ = NewResearchSwarmClient()
	})
	return researchSwarmInstance
}

// NewResearchSwarmClient creates a new research swarm client
func NewResearchSwarmClient() (*ResearchSwarmClient, error) {
	vaultPath := config.Get().AftrsVaultPath

	client := &ResearchSwarmClient{
		vaultPath:    vaultPath,
		usageHistory: make([]*ToolUsage, 0),
		patterns:     make([]*UsagePattern, 0),
		suggestions:  make([]*ImprovementSuggestion, 0),
		workers:      make(map[string]*SwarmWorker),
		config: &SwarmConfig{
			Enabled:             true,
			MinPatternFrequency: 3,
			AnalysisInterval:    time.Hour,
			MaxHistorySize:      10000,
			AutoSuggest:         true,
		},
		stats: &SwarmStats{
			PatternsByType:    make(map[string]int),
			SuggestionsByType: make(map[string]int),
		},
	}

	// Initialize workers
	workerTypes := []string{"pattern_miner", "sequence_analyzer", "error_analyzer", "optimization_finder"}
	for _, wt := range workerTypes {
		client.workers[wt] = &SwarmWorker{
			ID:     wt,
			Type:   wt,
			Status: "idle",
		}
	}

	// Load from vault
	client.loadFromVault()

	return client, nil
}

// RecordUsage records a tool usage event
func (c *ResearchSwarmClient) RecordUsage(tool, category string, args map[string]interface{}, success bool, duration time.Duration, sessionID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Get previous tool for sequence tracking
	prevTool := ""
	if len(c.usageHistory) > 0 {
		last := c.usageHistory[len(c.usageHistory)-1]
		if last.SessionID == sessionID && time.Since(last.Timestamp) < 5*time.Minute {
			prevTool = last.Tool
			// Update previous tool's next_tool
			last.NextTool = tool
		}
	}

	usage := &ToolUsage{
		ID:        uuid.New().String()[:8],
		Tool:      tool,
		Category:  category,
		Args:      args,
		Success:   success,
		Duration:  duration,
		Timestamp: time.Now(),
		SessionID: sessionID,
		PrevTool:  prevTool,
	}

	c.usageHistory = append(c.usageHistory, usage)

	// Trim history if too large
	if len(c.usageHistory) > c.config.MaxHistorySize {
		c.usageHistory = c.usageHistory[len(c.usageHistory)-c.config.MaxHistorySize:]
	}

	// Update stats
	c.stats.TotalUsage = len(c.usageHistory)

	return c.saveUsageToVault()
}

// AnalyzePatterns runs pattern analysis on usage history
func (c *ResearchSwarmClient) AnalyzePatterns() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Sequential patterns
	c.findSequentialPatterns()

	// Co-occurring patterns
	c.findCoOccurringPatterns()

	// Error recovery patterns
	c.findErrorRecoveryPatterns()

	// Update stats
	c.stats.TotalPatterns = len(c.patterns)
	c.stats.LastAnalysis = time.Now()

	// Generate suggestions if enabled
	if c.config.AutoSuggest {
		c.generateSuggestions()
	}

	return c.savePatternsToVault()
}

// findSequentialPatterns finds tools used in sequence
func (c *ResearchSwarmClient) findSequentialPatterns() {
	sequences := make(map[string]int)

	for _, usage := range c.usageHistory {
		if usage.PrevTool != "" {
			key := fmt.Sprintf("%s -> %s", usage.PrevTool, usage.Tool)
			sequences[key]++
		}
	}

	for seq, count := range sequences {
		if count >= c.config.MinPatternFrequency {
			parts := strings.Split(seq, " -> ")
			// Check if pattern already exists
			exists := false
			for _, p := range c.patterns {
				if p.Type == "sequential" && len(p.Tools) == 2 && p.Tools[0] == parts[0] && p.Tools[1] == parts[1] {
					p.Frequency = count
					p.LastSeen = time.Now()
					exists = true
					break
				}
			}
			if !exists {
				c.patterns = append(c.patterns, &UsagePattern{
					ID:           uuid.New().String()[:8],
					Type:         "sequential",
					Tools:        parts,
					Frequency:    count,
					Confidence:   float64(count) / float64(len(c.usageHistory)) * 100,
					Description:  fmt.Sprintf("%s is frequently followed by %s", parts[0], parts[1]),
					DiscoveredAt: time.Now(),
					LastSeen:     time.Now(),
				})
			}
		}
	}

	c.stats.PatternsByType["sequential"] = len(c.patterns)
}

// findCoOccurringPatterns finds tools used together in sessions
func (c *ResearchSwarmClient) findCoOccurringPatterns() {
	sessionTools := make(map[string]map[string]bool)

	for _, usage := range c.usageHistory {
		if usage.SessionID == "" {
			continue
		}
		if sessionTools[usage.SessionID] == nil {
			sessionTools[usage.SessionID] = make(map[string]bool)
		}
		sessionTools[usage.SessionID][usage.Tool] = true
	}

	// Count co-occurrences
	coOccur := make(map[string]int)
	for _, tools := range sessionTools {
		toolList := make([]string, 0, len(tools))
		for t := range tools {
			toolList = append(toolList, t)
		}
		sort.Strings(toolList)

		// Check pairs
		for i := 0; i < len(toolList); i++ {
			for j := i + 1; j < len(toolList); j++ {
				key := fmt.Sprintf("%s + %s", toolList[i], toolList[j])
				coOccur[key]++
			}
		}
	}

	for pair, count := range coOccur {
		if count >= c.config.MinPatternFrequency {
			parts := strings.Split(pair, " + ")
			// Check if pattern exists
			exists := false
			for _, p := range c.patterns {
				if p.Type == "co-occurring" && len(p.Tools) == 2 && p.Tools[0] == parts[0] && p.Tools[1] == parts[1] {
					p.Frequency = count
					p.LastSeen = time.Now()
					exists = true
					break
				}
			}
			if !exists {
				c.patterns = append(c.patterns, &UsagePattern{
					ID:           uuid.New().String()[:8],
					Type:         "co-occurring",
					Tools:        parts,
					Frequency:    count,
					Confidence:   float64(count) / float64(len(sessionTools)) * 100,
					Description:  fmt.Sprintf("%s and %s are frequently used together", parts[0], parts[1]),
					DiscoveredAt: time.Now(),
					LastSeen:     time.Now(),
				})
			}
		}
	}
}

// findErrorRecoveryPatterns finds tools used after failures
func (c *ResearchSwarmClient) findErrorRecoveryPatterns() {
	recoveries := make(map[string]int)

	for i := 1; i < len(c.usageHistory); i++ {
		prev := c.usageHistory[i-1]
		curr := c.usageHistory[i]

		if !prev.Success && curr.Success && prev.SessionID == curr.SessionID {
			key := fmt.Sprintf("%s (failed) -> %s (success)", prev.Tool, curr.Tool)
			recoveries[key]++
		}
	}

	for pattern, count := range recoveries {
		if count >= c.config.MinPatternFrequency {
			// Extract tool names
			parts := strings.Split(pattern, " -> ")
			failedTool := strings.TrimSuffix(parts[0], " (failed)")
			successTool := strings.TrimSuffix(parts[1], " (success)")

			exists := false
			for _, p := range c.patterns {
				if p.Type == "error_recovery" && len(p.Tools) >= 2 && p.Tools[0] == failedTool && p.Tools[1] == successTool {
					p.Frequency = count
					p.LastSeen = time.Now()
					exists = true
					break
				}
			}
			if !exists {
				c.patterns = append(c.patterns, &UsagePattern{
					ID:           uuid.New().String()[:8],
					Type:         "error_recovery",
					Tools:        []string{failedTool, successTool},
					Frequency:    count,
					Confidence:   float64(count) / float64(len(c.usageHistory)) * 100,
					Description:  fmt.Sprintf("When %s fails, %s is often used to recover", failedTool, successTool),
					DiscoveredAt: time.Now(),
					LastSeen:     time.Now(),
				})
			}
		}
	}
}

// generateSuggestions creates improvement suggestions from patterns
func (c *ResearchSwarmClient) generateSuggestions() {
	for _, pattern := range c.patterns {
		// Check if suggestion already exists
		exists := false
		for _, s := range c.suggestions {
			if s.Pattern == pattern.ID {
				exists = true
				break
			}
		}
		if exists {
			continue
		}

		var suggestion *ImprovementSuggestion

		switch pattern.Type {
		case "sequential":
			if pattern.Frequency >= 5 && pattern.Confidence >= 10 {
				suggestion = &ImprovementSuggestion{
					ID:          uuid.New().String()[:8],
					Type:        "chain",
					Title:       fmt.Sprintf("Create chain: %s → %s", pattern.Tools[0], pattern.Tools[1]),
					Description: fmt.Sprintf("These tools are used in sequence %d times. Consider creating a workflow chain.", pattern.Frequency),
					Priority:    min(5, pattern.Frequency/5),
					Impact:      "medium",
					Effort:      "low",
					Status:      "new",
					Pattern:     pattern.ID,
					Tools:       pattern.Tools,
					CreatedAt:   time.Now(),
				}
			}
		case "co-occurring":
			if pattern.Frequency >= 10 && pattern.Confidence >= 20 {
				suggestion = &ImprovementSuggestion{
					ID:          uuid.New().String()[:8],
					Type:        "consolidation",
					Title:       fmt.Sprintf("Consolidate: %s + %s", pattern.Tools[0], pattern.Tools[1]),
					Description: fmt.Sprintf("These tools are used together in %d sessions. Consider creating a consolidated tool.", pattern.Frequency),
					Priority:    min(5, pattern.Frequency/10),
					Impact:      "high",
					Effort:      "medium",
					Status:      "new",
					Pattern:     pattern.ID,
					Tools:       pattern.Tools,
					CreatedAt:   time.Now(),
				}
			}
		case "error_recovery":
			if pattern.Frequency >= 3 {
				suggestion = &ImprovementSuggestion{
					ID:          uuid.New().String()[:8],
					Type:        "new_tool",
					Title:       fmt.Sprintf("Auto-recovery: %s → %s", pattern.Tools[0], pattern.Tools[1]),
					Description: fmt.Sprintf("When %s fails, %s is used for recovery %d times. Consider adding automatic recovery.", pattern.Tools[0], pattern.Tools[1], pattern.Frequency),
					Priority:    4, // Error recovery is important
					Impact:      "high",
					Effort:      "medium",
					Status:      "new",
					Pattern:     pattern.ID,
					Tools:       pattern.Tools,
					CreatedAt:   time.Now(),
				}
			}
		}

		if suggestion != nil {
			c.suggestions = append(c.suggestions, suggestion)
			c.stats.SuggestionsByType[suggestion.Type]++
		}
	}

	c.stats.TotalSuggestions = len(c.suggestions)
}

// GetPatterns returns discovered patterns
func (c *ResearchSwarmClient) GetPatterns(patternType string, minFrequency int) []*UsagePattern {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []*UsagePattern
	for _, p := range c.patterns {
		if (patternType == "" || p.Type == patternType) && p.Frequency >= minFrequency {
			result = append(result, p)
		}
	}

	// Sort by frequency descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].Frequency > result[j].Frequency
	})

	return result
}

// GetSuggestions returns improvement suggestions
func (c *ResearchSwarmClient) GetSuggestions(suggestionType, status string) []*ImprovementSuggestion {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []*ImprovementSuggestion
	for _, s := range c.suggestions {
		if (suggestionType == "" || s.Type == suggestionType) && (status == "" || s.Status == status) {
			result = append(result, s)
		}
	}

	// Sort by priority descending
	sort.Slice(result, func(i, j int) bool {
		return result[i].Priority > result[j].Priority
	})

	return result
}

// UpdateSuggestionStatus updates a suggestion's status
func (c *ResearchSwarmClient) UpdateSuggestionStatus(id, status, reviewedBy string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, s := range c.suggestions {
		if s.ID == id {
			s.Status = status
			s.ReviewedAt = time.Now()
			s.ReviewedBy = reviewedBy
			return c.saveSuggestionsToVault()
		}
	}

	return fmt.Errorf("suggestion not found: %s", id)
}

// GetToolStats returns statistics for tools
func (c *ResearchSwarmClient) GetToolStats(limit int) []ToolStat {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if limit <= 0 {
		limit = 10
	}

	toolCounts := make(map[string]int)
	toolSuccess := make(map[string]int)
	toolDurations := make(map[string][]int64)

	for _, u := range c.usageHistory {
		toolCounts[u.Tool]++
		if u.Success {
			toolSuccess[u.Tool]++
		}
		toolDurations[u.Tool] = append(toolDurations[u.Tool], int64(u.Duration/time.Millisecond))
	}

	var stats []ToolStat
	for tool, count := range toolCounts {
		avgDuration := int64(0)
		if len(toolDurations[tool]) > 0 {
			sum := int64(0)
			for _, d := range toolDurations[tool] {
				sum += d
			}
			avgDuration = sum / int64(len(toolDurations[tool]))
		}

		stats = append(stats, ToolStat{
			Tool:        tool,
			Count:       count,
			SuccessRate: float64(toolSuccess[tool]) / float64(count) * 100,
			AvgDuration: avgDuration,
		})
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Count > stats[j].Count
	})

	if len(stats) > limit {
		stats = stats[:limit]
	}

	return stats
}

// GetWorkerStatus returns the status of swarm workers
func (c *ResearchSwarmClient) GetWorkerStatus() []*SwarmWorker {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var workers []*SwarmWorker
	for _, w := range c.workers {
		workers = append(workers, w)
	}
	return workers
}

// GetStats returns overall swarm statistics
func (c *ResearchSwarmClient) GetStats() *SwarmStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Update derived stats
	c.stats.TopTools = c.GetToolStats(10)

	return c.stats
}

// GetConfig returns the swarm configuration
func (c *ResearchSwarmClient) GetConfig() *SwarmConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

// UpdateConfig updates the swarm configuration
func (c *ResearchSwarmClient) UpdateConfig(config *SwarmConfig) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.config = config
}

// loadFromVault loads data from vault
func (c *ResearchSwarmClient) loadFromVault() {
	swarmDir := filepath.Join(c.vaultPath, "swarm")

	// Load usage
	if data, err := os.ReadFile(filepath.Join(swarmDir, "usage.json")); err == nil {
		json.Unmarshal(data, &c.usageHistory)
	}

	// Load patterns
	if data, err := os.ReadFile(filepath.Join(swarmDir, "patterns.json")); err == nil {
		json.Unmarshal(data, &c.patterns)
	}

	// Load suggestions
	if data, err := os.ReadFile(filepath.Join(swarmDir, "suggestions.json")); err == nil {
		json.Unmarshal(data, &c.suggestions)
	}

	// Update stats
	c.stats.TotalUsage = len(c.usageHistory)
	c.stats.TotalPatterns = len(c.patterns)
	c.stats.TotalSuggestions = len(c.suggestions)
}

// saveUsageToVault persists usage history
func (c *ResearchSwarmClient) saveUsageToVault() error {
	swarmDir := filepath.Join(c.vaultPath, "swarm")
	if err := os.MkdirAll(swarmDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c.usageHistory, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(swarmDir, "usage.json"), data, 0644)
}

// savePatternsToVault persists patterns
func (c *ResearchSwarmClient) savePatternsToVault() error {
	swarmDir := filepath.Join(c.vaultPath, "swarm")
	if err := os.MkdirAll(swarmDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c.patterns, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(swarmDir, "patterns.json"), data, 0644)
}

// saveSuggestionsToVault persists suggestions
func (c *ResearchSwarmClient) saveSuggestionsToVault() error {
	swarmDir := filepath.Join(c.vaultPath, "swarm")
	if err := os.MkdirAll(swarmDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c.suggestions, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(swarmDir, "suggestions.json"), data, 0644)
}
