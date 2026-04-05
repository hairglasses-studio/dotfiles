// Package clients provides API clients for external services.
package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/hairglasses-studio/hg-mcp/internal/config"
)

// LearningClient provides pattern learning and troubleshooting capabilities
type LearningClient struct {
	vaultPath    string
	learningRate float64
	minSamples   int
	weightBounds [2]float64
	patterns     []LearnedPattern
	symptoms     []SymptomRecord
}

// LearnedPattern represents a learned issue/resolution pattern
type LearnedPattern struct {
	ID          string    `json:"id"`
	Category    string    `json:"category"` // equipment, software, network, etc.
	Symptoms    []string  `json:"symptoms"`
	RootCause   string    `json:"root_cause"`
	Resolution  string    `json:"resolution"`
	Steps       []string  `json:"steps,omitempty"`
	Equipment   []string  `json:"equipment,omitempty"`
	Venue       string    `json:"venue,omitempty"`
	Occurrences int       `json:"occurrences"`
	LastSeen    time.Time `json:"last_seen"`
	AvgResTime  float64   `json:"avg_resolution_time_mins"`
	Weight      float64   `json:"weight"`
	Source      string    `json:"source"` // vault path of learning document
}

// SymptomRecord represents a recorded symptom
type SymptomRecord struct {
	ID           string    `json:"id"`
	Description  string    `json:"description"`
	Equipment    string    `json:"equipment,omitempty"`
	Venue        string    `json:"venue,omitempty"`
	Category     string    `json:"category"`
	Severity     string    `json:"severity"` // low, medium, high, critical
	Timestamp    time.Time `json:"timestamp"`
	Resolved     bool      `json:"resolved"`
	ResolutionID string    `json:"resolution_id,omitempty"`
}

// PatternMatch represents a matched pattern
type PatternMatch struct {
	Pattern         *LearnedPattern `json:"pattern"`
	Confidence      float64         `json:"confidence"`
	MatchedSymptoms []string        `json:"matched_symptoms"`
}

// RootCauseAnalysis represents a generated RCA
type RootCauseAnalysis struct {
	Issue               string   `json:"issue"`
	ProbableRootCause   string   `json:"probable_root_cause"`
	ContributingFactors []string `json:"contributing_factors"`
	SuggestedFixes      []string `json:"suggested_fixes"`
	SimilarPastIssues   []string `json:"similar_past_issues"`
	Confidence          float64  `json:"confidence"`
}

// TroubleshootStep represents a troubleshooting step
type TroubleshootStep struct {
	StepNum     int      `json:"step_num"`
	Action      string   `json:"action"`
	Expected    string   `json:"expected"`
	IfFails     string   `json:"if_fails,omitempty"`
	Tools       []string `json:"tools,omitempty"`
	Checkpoints []string `json:"checkpoints,omitempty"`
}

// TroubleshootGuide represents a troubleshooting guide
type TroubleshootGuide struct {
	Issue    string             `json:"issue"`
	Category string             `json:"category"`
	Steps    []TroubleshootStep `json:"steps"`
	EstTime  int                `json:"estimated_time_mins"`
	BasedOn  []string           `json:"based_on_patterns"`
}

// SymptomCorrelation represents correlated symptoms
type SymptomCorrelation struct {
	PrimarySymptom string   `json:"primary_symptom"`
	Related        []string `json:"related_symptoms"`
	CommonCause    string   `json:"common_cause,omitempty"`
	Frequency      int      `json:"frequency"`
}

// EquipmentHistory represents equipment issue history
type EquipmentHistory struct {
	Equipment        string    `json:"equipment"`
	TotalIssues      int       `json:"total_issues"`
	RecentIssues     []string  `json:"recent_issues"`
	CommonPatterns   []string  `json:"common_patterns"`
	ReliabilityScore float64   `json:"reliability_score"`
	LastIncident     time.Time `json:"last_incident,omitempty"`
}

// VenuePattern represents venue-specific patterns
type VenuePattern struct {
	Venue           string   `json:"venue"`
	TotalShows      int      `json:"total_shows"`
	CommonIssues    []string `json:"common_issues"`
	Quirks          []string `json:"quirks"`
	Recommendations []string `json:"recommendations"`
	SuccessRate     float64  `json:"success_rate"`
}

// NewLearningClient creates a new learning client
func NewLearningClient() (*LearningClient, error) {
	return &LearningClient{
		vaultPath:    config.Get().AftrsVaultPath,
		learningRate: 0.2, // From webb
		minSamples:   3,   // From webb
		weightBounds: [2]float64{0.1, 5.0},
		patterns:     []LearnedPattern{},
		symptoms:     []SymptomRecord{},
	}, nil
}

// LearnFromResolution learns a pattern from a resolved issue
func (c *LearningClient) LearnFromResolution(ctx context.Context, symptoms []string, rootCause, resolution string, equipment []string, venue string) (*LearnedPattern, error) {
	// Check if similar pattern exists
	for i, p := range c.patterns {
		similarity := c.calculateSimilarity(p.Symptoms, symptoms)
		if similarity > 0.7 {
			// Update existing pattern
			c.patterns[i].Occurrences++
			c.patterns[i].LastSeen = time.Now()
			c.patterns[i].Weight = c.adjustWeight(c.patterns[i].Weight, true)
			return &c.patterns[i], nil
		}
	}

	// Create new pattern
	pattern := LearnedPattern{
		ID:          fmt.Sprintf("pattern_%d", len(c.patterns)+1),
		Category:    c.categorizeSymptoms(symptoms),
		Symptoms:    symptoms,
		RootCause:   rootCause,
		Resolution:  resolution,
		Equipment:   equipment,
		Venue:       venue,
		Occurrences: 1,
		LastSeen:    time.Now(),
		Weight:      1.0,
	}

	c.patterns = append(c.patterns, pattern)

	// Persist to vault
	c.persistPattern(&pattern)

	return &pattern, nil
}

// MatchPatterns matches symptoms to learned patterns
func (c *LearningClient) MatchPatterns(ctx context.Context, symptoms []string, equipment, venue string) ([]PatternMatch, error) {
	c.loadPatterns()

	var matches []PatternMatch

	for _, p := range c.patterns {
		matched := []string{}
		for _, s := range symptoms {
			for _, ps := range p.Symptoms {
				if strings.Contains(strings.ToLower(ps), strings.ToLower(s)) ||
					strings.Contains(strings.ToLower(s), strings.ToLower(ps)) {
					matched = append(matched, s)
					break
				}
			}
		}

		if len(matched) == 0 {
			continue
		}

		confidence := float64(len(matched)) / float64(len(symptoms)) * p.Weight

		// Boost for matching equipment or venue
		if equipment != "" && containsAny(p.Equipment, equipment) {
			confidence *= 1.2
		}
		if venue != "" && strings.EqualFold(p.Venue, venue) {
			confidence *= 1.1
		}

		// Cap confidence at 1.0
		if confidence > 1.0 {
			confidence = 1.0
		}

		matches = append(matches, PatternMatch{
			Pattern:         &p,
			Confidence:      confidence,
			MatchedSymptoms: matched,
		})
	}

	// Sort by confidence
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Confidence > matches[j].Confidence
	})

	return matches, nil
}

// GenerateRCA generates a root cause analysis
func (c *LearningClient) GenerateRCA(ctx context.Context, issue string, symptoms []string) (*RootCauseAnalysis, error) {
	matches, _ := c.MatchPatterns(ctx, symptoms, "", "")

	rca := &RootCauseAnalysis{
		Issue:               issue,
		ContributingFactors: []string{},
		SuggestedFixes:      []string{},
		SimilarPastIssues:   []string{},
		Confidence:          0.0,
	}

	if len(matches) == 0 {
		rca.ProbableRootCause = "Unable to determine - no matching patterns"
		rca.SuggestedFixes = []string{"Document this issue for future learning"}
		return rca, nil
	}

	// Use top match as probable root cause
	topMatch := matches[0]
	rca.ProbableRootCause = topMatch.Pattern.RootCause
	rca.Confidence = topMatch.Confidence

	// Gather contributing factors from all matches
	seen := make(map[string]bool)
	for _, m := range matches {
		if !seen[m.Pattern.RootCause] {
			if m.Pattern.RootCause != topMatch.Pattern.RootCause {
				rca.ContributingFactors = append(rca.ContributingFactors, m.Pattern.RootCause)
			}
			seen[m.Pattern.RootCause] = true
		}
	}

	// Gather fixes
	seenFixes := make(map[string]bool)
	for _, m := range matches {
		if !seenFixes[m.Pattern.Resolution] {
			rca.SuggestedFixes = append(rca.SuggestedFixes, m.Pattern.Resolution)
			seenFixes[m.Pattern.Resolution] = true
		}
	}

	// Gather similar past issues
	for _, m := range matches[:min(3, len(matches))] {
		rca.SimilarPastIssues = append(rca.SimilarPastIssues,
			fmt.Sprintf("%s (seen %d times)", m.Pattern.ID, m.Pattern.Occurrences))
	}

	return rca, nil
}

// SuggestFixes suggests fixes based on symptoms
func (c *LearningClient) SuggestFixes(ctx context.Context, symptoms []string) ([]string, error) {
	matches, _ := c.MatchPatterns(ctx, symptoms, "", "")

	fixes := []string{}
	seen := make(map[string]bool)

	for _, m := range matches {
		if m.Confidence < 0.3 {
			continue
		}
		if !seen[m.Pattern.Resolution] {
			fixes = append(fixes, m.Pattern.Resolution)
			seen[m.Pattern.Resolution] = true
		}
		for _, step := range m.Pattern.Steps {
			if !seen[step] {
				fixes = append(fixes, step)
				seen[step] = true
			}
		}
	}

	if len(fixes) == 0 {
		fixes = append(fixes, "No matching fixes found. Consider documenting this issue.")
	}

	return fixes, nil
}

// GetTroubleshootGuide generates a troubleshooting guide
func (c *LearningClient) GetTroubleshootGuide(ctx context.Context, issue string) (*TroubleshootGuide, error) {
	// Extract symptoms from issue description
	symptoms := c.extractSymptoms(issue)
	category := c.categorizeSymptoms(symptoms)

	matches, _ := c.MatchPatterns(ctx, symptoms, "", "")

	guide := &TroubleshootGuide{
		Issue:    issue,
		Category: category,
		Steps:    []TroubleshootStep{},
		EstTime:  15, // Default
		BasedOn:  []string{},
	}

	// Generate steps based on matched patterns
	stepNum := 1

	// Always start with basic checks
	guide.Steps = append(guide.Steps, TroubleshootStep{
		StepNum:  stepNum,
		Action:   "Verify all equipment is powered on and connected",
		Expected: "All status lights should be green/active",
		IfFails:  "Check power cables and network connections",
	})
	stepNum++

	// Add pattern-based steps
	for _, m := range matches {
		if m.Confidence < 0.4 {
			continue
		}
		guide.BasedOn = append(guide.BasedOn, m.Pattern.ID)

		if len(m.Pattern.Steps) > 0 {
			for _, step := range m.Pattern.Steps {
				guide.Steps = append(guide.Steps, TroubleshootStep{
					StepNum: stepNum,
					Action:  step,
				})
				stepNum++
			}
		} else {
			guide.Steps = append(guide.Steps, TroubleshootStep{
				StepNum:  stepNum,
				Action:   fmt.Sprintf("Check for: %s", m.Pattern.RootCause),
				Expected: "Issue should be resolved",
				IfFails:  "Move to next step",
			})
			stepNum++
		}
	}

	// Add final step
	guide.Steps = append(guide.Steps, TroubleshootStep{
		StepNum: stepNum,
		Action:  "If issue persists, document symptoms and escalate",
		Tools:   []string{"aftrs_pattern_learn", "aftrs_discord_notify"},
	})

	return guide, nil
}

// CorrelateSymptoms finds correlated symptoms
func (c *LearningClient) CorrelateSymptoms(ctx context.Context, symptom string) ([]SymptomCorrelation, error) {
	c.loadPatterns()

	correlations := make(map[string]*SymptomCorrelation)

	for _, p := range c.patterns {
		hasSymptom := false
		for _, s := range p.Symptoms {
			if strings.Contains(strings.ToLower(s), strings.ToLower(symptom)) {
				hasSymptom = true
				break
			}
		}

		if !hasSymptom {
			continue
		}

		// Find correlated symptoms from the same pattern
		for _, s := range p.Symptoms {
			if strings.Contains(strings.ToLower(s), strings.ToLower(symptom)) {
				continue
			}

			if _, exists := correlations[s]; !exists {
				correlations[s] = &SymptomCorrelation{
					PrimarySymptom: symptom,
					Related:        []string{},
					Frequency:      0,
				}
			}
			correlations[s].Related = append(correlations[s].Related, s)
			correlations[s].Frequency++
			if correlations[s].CommonCause == "" {
				correlations[s].CommonCause = p.RootCause
			}
		}
	}

	var result []SymptomCorrelation
	for _, c := range correlations {
		result = append(result, *c)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Frequency > result[j].Frequency
	})

	return result, nil
}

// GetEquipmentHistory gets issue history for equipment
func (c *LearningClient) GetEquipmentHistory(ctx context.Context, equipment string) (*EquipmentHistory, error) {
	c.loadPatterns()

	history := &EquipmentHistory{
		Equipment:        equipment,
		TotalIssues:      0,
		RecentIssues:     []string{},
		CommonPatterns:   []string{},
		ReliabilityScore: 100.0,
	}

	patternCounts := make(map[string]int)

	for _, p := range c.patterns {
		if !containsAny(p.Equipment, equipment) {
			continue
		}

		history.TotalIssues += p.Occurrences
		patternCounts[p.RootCause]++

		if history.LastIncident.Before(p.LastSeen) {
			history.LastIncident = p.LastSeen
		}
	}

	// Calculate reliability score (100 - issues per month normalized)
	if history.TotalIssues > 0 {
		history.ReliabilityScore = 100.0 - float64(history.TotalIssues*5)
		if history.ReliabilityScore < 0 {
			history.ReliabilityScore = 0
		}
	}

	// Get common patterns
	for cause, count := range patternCounts {
		if count >= c.minSamples {
			history.CommonPatterns = append(history.CommonPatterns, cause)
		}
	}

	return history, nil
}

// GetVenuePatterns gets patterns for a venue
func (c *LearningClient) GetVenuePatterns(ctx context.Context, venue string) (*VenuePattern, error) {
	c.loadPatterns()

	pattern := &VenuePattern{
		Venue:           venue,
		TotalShows:      0,
		CommonIssues:    []string{},
		Quirks:          []string{},
		Recommendations: []string{},
		SuccessRate:     100.0,
	}

	issueCounts := make(map[string]int)
	resolved := 0
	total := 0

	for _, p := range c.patterns {
		if !strings.EqualFold(p.Venue, venue) {
			continue
		}

		total++
		issueCounts[p.RootCause]++

		// Assuming all patterns represent resolved issues
		resolved++
	}

	if total > 0 {
		pattern.SuccessRate = float64(resolved) / float64(total) * 100
	}

	// Extract common issues
	for issue, count := range issueCounts {
		if count >= 2 {
			pattern.CommonIssues = append(pattern.CommonIssues, issue)
			pattern.Recommendations = append(pattern.Recommendations,
				fmt.Sprintf("Watch for: %s (seen %d times)", issue, count))
		}
	}

	return pattern, nil
}

// Helper methods

func (c *LearningClient) calculateSimilarity(a, b []string) float64 {
	if len(a) == 0 || len(b) == 0 {
		return 0
	}
	matches := 0
	for _, s1 := range a {
		for _, s2 := range b {
			if strings.EqualFold(s1, s2) {
				matches++
				break
			}
		}
	}
	return float64(matches) / float64(max(len(a), len(b)))
}

func (c *LearningClient) adjustWeight(current float64, success bool) float64 {
	if success {
		current += c.learningRate
	} else {
		current -= c.learningRate
	}
	if current < c.weightBounds[0] {
		return c.weightBounds[0]
	}
	if current > c.weightBounds[1] {
		return c.weightBounds[1]
	}
	return current
}

func (c *LearningClient) categorizeSymptoms(symptoms []string) string {
	for _, s := range symptoms {
		lower := strings.ToLower(s)
		switch {
		case strings.Contains(lower, "network") || strings.Contains(lower, "connection"):
			return "network"
		case strings.Contains(lower, "display") || strings.Contains(lower, "video") || strings.Contains(lower, "image"):
			return "video"
		case strings.Contains(lower, "audio") || strings.Contains(lower, "sound"):
			return "audio"
		case strings.Contains(lower, "light") || strings.Contains(lower, "dmx"):
			return "lighting"
		case strings.Contains(lower, "crash") || strings.Contains(lower, "freeze"):
			return "software"
		case strings.Contains(lower, "performance") || strings.Contains(lower, "slow"):
			return "performance"
		}
	}
	return "general"
}

func (c *LearningClient) extractSymptoms(text string) []string {
	// Simple extraction - split by common delimiters
	text = strings.ToLower(text)
	symptoms := []string{}

	// Look for common symptom keywords
	keywords := []string{"not working", "error", "failure", "crash", "slow", "freeze", "disconnect", "no signal", "black screen", "no audio"}
	for _, kw := range keywords {
		if strings.Contains(text, kw) {
			symptoms = append(symptoms, kw)
		}
	}

	if len(symptoms) == 0 {
		// Use the whole text as a symptom
		symptoms = append(symptoms, text)
	}

	return symptoms
}

func (c *LearningClient) loadPatterns() {
	// Load from vault if patterns are empty
	if len(c.patterns) > 0 {
		return
	}

	learningsPath := filepath.Join(c.vaultPath, "learnings")
	if _, err := os.Stat(learningsPath); os.IsNotExist(err) {
		return
	}

	filepath.Walk(learningsPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		var pattern LearnedPattern
		if err := json.Unmarshal(data, &pattern); err == nil {
			c.patterns = append(c.patterns, pattern)
		}

		return nil
	})
}

func (c *LearningClient) persistPattern(pattern *LearnedPattern) error {
	learningsPath := filepath.Join(c.vaultPath, "learnings")
	os.MkdirAll(learningsPath, 0755)

	data, err := json.MarshalIndent(pattern, "", "  ")
	if err != nil {
		return err
	}

	filename := filepath.Join(learningsPath, pattern.ID+".json")
	return os.WriteFile(filename, data, 0644)
}

func containsAny(slice []string, item string) bool {
	for _, s := range slice {
		if strings.EqualFold(s, item) || strings.Contains(strings.ToLower(s), strings.ToLower(item)) {
			return true
		}
	}
	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
