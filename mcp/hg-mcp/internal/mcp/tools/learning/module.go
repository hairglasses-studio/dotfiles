// Package learning provides pattern learning and troubleshooting tools for hg-mcp.
package learning

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

var getClient = tools.LazyClient(clients.NewLearningClient)

// Module implements the ToolModule interface for learning tools
type Module struct{}

func (m *Module) Name() string {
	return "learning"
}

func (m *Module) Description() string {
	return "Pattern learning and troubleshooting tools that improve over time"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_pattern_learn",
				mcp.WithDescription("Learn a pattern from a resolved issue. Captures symptoms, root cause, and resolution for future matching."),
				mcp.WithString("symptoms", mcp.Description("Comma-separated list of symptoms"), mcp.Required()),
				mcp.WithString("root_cause", mcp.Description("Root cause of the issue"), mcp.Required()),
				mcp.WithString("resolution", mcp.Description("How it was resolved"), mcp.Required()),
				mcp.WithString("equipment", mcp.Description("Comma-separated list of affected equipment")),
				mcp.WithString("venue", mcp.Description("Venue where issue occurred")),
			),
			Handler:             handlePatternLearn,
			Category:            "learning",
			Subcategory:         "patterns",
			Tags:                []string{"learning", "pattern", "resolution", "capture"},
			UseCases:            []string{"Document resolved issues", "Build pattern library"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "learning",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_pattern_match",
				mcp.WithDescription("Match current symptoms to learned patterns. Returns best matches with confidence scores."),
				mcp.WithString("symptoms", mcp.Description("Comma-separated list of current symptoms"), mcp.Required()),
				mcp.WithString("equipment", mcp.Description("Equipment involved (optional)")),
				mcp.WithString("venue", mcp.Description("Current venue (optional)")),
			),
			Handler:             handlePatternMatch,
			Category:            "learning",
			Subcategory:         "patterns",
			Tags:                []string{"learning", "pattern", "match", "diagnosis"},
			UseCases:            []string{"Find similar past issues", "Get fix suggestions"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "learning",
		},
		{
			Tool: mcp.NewTool("aftrs_rca_generate",
				mcp.WithDescription("Generate a root cause analysis based on symptoms and learned patterns."),
				mcp.WithString("issue", mcp.Description("Description of the issue"), mcp.Required()),
				mcp.WithString("symptoms", mcp.Description("Comma-separated list of symptoms"), mcp.Required()),
			),
			Handler:             handleRCAGenerate,
			Category:            "learning",
			Subcategory:         "analysis",
			Tags:                []string{"learning", "rca", "analysis", "root cause"},
			UseCases:            []string{"Diagnose complex issues", "Generate post-mortem"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "learning",
		},
		{
			Tool: mcp.NewTool("aftrs_fix_suggest",
				mcp.WithDescription("Suggest fixes based on current symptoms and learned patterns."),
				mcp.WithString("symptoms", mcp.Description("Comma-separated list of symptoms"), mcp.Required()),
			),
			Handler:             handleFixSuggest,
			Category:            "learning",
			Subcategory:         "suggestions",
			Tags:                []string{"learning", "fix", "suggest", "resolution"},
			UseCases:            []string{"Get quick fix suggestions", "Troubleshoot issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "learning",
		},
		{
			Tool: mcp.NewTool("aftrs_troubleshoot",
				mcp.WithDescription("Get a guided troubleshooting sequence based on the issue and learned patterns."),
				mcp.WithString("issue", mcp.Description("Description of the issue"), mcp.Required()),
			),
			Handler:             handleTroubleshoot,
			Category:            "learning",
			Subcategory:         "troubleshooting",
			Tags:                []string{"learning", "troubleshoot", "guide", "steps"},
			UseCases:            []string{"Step-by-step troubleshooting", "Guided diagnosis"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "learning",
		},
		{
			Tool: mcp.NewTool("aftrs_symptom_correlate",
				mcp.WithDescription("Find symptoms that commonly occur together with the given symptom."),
				mcp.WithString("symptom", mcp.Description("The symptom to find correlations for"), mcp.Required()),
			),
			Handler:             handleSymptomCorrelate,
			Category:            "learning",
			Subcategory:         "analysis",
			Tags:                []string{"learning", "symptom", "correlate", "analysis"},
			UseCases:            []string{"Find related symptoms", "Predict cascading issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "learning",
		},
		{
			Tool: mcp.NewTool("aftrs_equipment_history",
				mcp.WithDescription("Get issue history and reliability score for specific equipment."),
				mcp.WithString("equipment", mcp.Description("Equipment name"), mcp.Required()),
			),
			Handler:             handleEquipmentHistory,
			Category:            "learning",
			Subcategory:         "history",
			Tags:                []string{"learning", "equipment", "history", "reliability"},
			UseCases:            []string{"Check equipment reliability", "Review past issues"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "learning",
		},
		{
			Tool: mcp.NewTool("aftrs_venue_patterns",
				mcp.WithDescription("Get venue-specific patterns, common issues, and recommendations."),
				mcp.WithString("venue", mcp.Description("Venue name"), mcp.Required()),
			),
			Handler:             handleVenuePatterns,
			Category:            "learning",
			Subcategory:         "venues",
			Tags:                []string{"learning", "venue", "patterns", "history"},
			UseCases:            []string{"Prepare for venue", "Review venue quirks"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "learning",
		},
	}
}

// handlePatternLearn handles the aftrs_pattern_learn tool
func handlePatternLearn(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	symptomsStr, errResult := tools.RequireStringParam(req, "symptoms")
	if errResult != nil {
		return errResult, nil
	}
	rootCause, errResult := tools.RequireStringParam(req, "root_cause")
	if errResult != nil {
		return errResult, nil
	}
	resolution, errResult := tools.RequireStringParam(req, "resolution")
	if errResult != nil {
		return errResult, nil
	}
	equipmentStr := tools.GetStringParam(req, "equipment")
	venue := tools.GetStringParam(req, "venue")

	symptoms := strings.Split(symptomsStr, ",")
	for i := range symptoms {
		symptoms[i] = strings.TrimSpace(symptoms[i])
	}

	var equipment []string
	if equipmentStr != "" {
		equipment = strings.Split(equipmentStr, ",")
		for i := range equipment {
			equipment[i] = strings.TrimSpace(equipment[i])
		}
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create learning client: %w", err)), nil
	}

	pattern, err := client.LearnFromResolution(ctx, symptoms, rootCause, resolution, equipment, venue)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to learn pattern: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Pattern Learned\n\n")
	sb.WriteString(fmt.Sprintf("**ID:** %s\n", pattern.ID))
	sb.WriteString(fmt.Sprintf("**Category:** %s\n", pattern.Category))
	sb.WriteString(fmt.Sprintf("**Occurrences:** %d\n", pattern.Occurrences))
	sb.WriteString(fmt.Sprintf("**Weight:** %.2f\n\n", pattern.Weight))

	sb.WriteString("## Symptoms\n")
	for _, s := range pattern.Symptoms {
		sb.WriteString(fmt.Sprintf("- %s\n", s))
	}

	sb.WriteString(fmt.Sprintf("\n## Root Cause\n%s\n", pattern.RootCause))
	sb.WriteString(fmt.Sprintf("\n## Resolution\n%s\n", pattern.Resolution))

	if len(pattern.Equipment) > 0 {
		sb.WriteString("\n## Equipment\n")
		sb.WriteString(strings.Join(pattern.Equipment, ", "))
		sb.WriteString("\n")
	}

	if pattern.Venue != "" {
		sb.WriteString(fmt.Sprintf("\n## Venue\n%s\n", pattern.Venue))
	}

	return tools.TextResult(sb.String()), nil
}

// handlePatternMatch handles the aftrs_pattern_match tool
func handlePatternMatch(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	symptomsStr, errResult := tools.RequireStringParam(req, "symptoms")
	if errResult != nil {
		return errResult, nil
	}
	equipment := tools.GetStringParam(req, "equipment")
	venue := tools.GetStringParam(req, "venue")

	symptoms := strings.Split(symptomsStr, ",")
	for i := range symptoms {
		symptoms[i] = strings.TrimSpace(symptoms[i])
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create learning client: %w", err)), nil
	}

	matches, err := client.MatchPatterns(ctx, symptoms, equipment, venue)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to match patterns: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Pattern Matches\n\n")
	sb.WriteString(fmt.Sprintf("**Symptoms:** %s\n\n", symptomsStr))
	sb.WriteString(fmt.Sprintf("Found **%d** matching patterns:\n\n", len(matches)))

	if len(matches) == 0 {
		sb.WriteString("No matching patterns found. Consider using `aftrs_pattern_learn` to document this issue.\n")
	} else {
		for i, m := range matches {
			if i >= 5 {
				sb.WriteString(fmt.Sprintf("\n... and %d more matches\n", len(matches)-5))
				break
			}

			emoji := getConfidenceEmoji(m.Confidence)
			sb.WriteString(fmt.Sprintf("### %d. %s (%.0f%% confidence) %s\n\n", i+1, m.Pattern.ID, m.Confidence*100, emoji))
			sb.WriteString(fmt.Sprintf("**Root Cause:** %s\n", m.Pattern.RootCause))
			sb.WriteString(fmt.Sprintf("**Resolution:** %s\n", m.Pattern.Resolution))
			sb.WriteString(fmt.Sprintf("**Matched Symptoms:** %s\n", strings.Join(m.MatchedSymptoms, ", ")))
			sb.WriteString(fmt.Sprintf("**Seen:** %d times\n\n", m.Pattern.Occurrences))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleRCAGenerate handles the aftrs_rca_generate tool
func handleRCAGenerate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	issue, errResult := tools.RequireStringParam(req, "issue")
	if errResult != nil {
		return errResult, nil
	}
	symptomsStr, errResult := tools.RequireStringParam(req, "symptoms")
	if errResult != nil {
		return errResult, nil
	}

	symptoms := strings.Split(symptomsStr, ",")
	for i := range symptoms {
		symptoms[i] = strings.TrimSpace(symptoms[i])
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create learning client: %w", err)), nil
	}

	rca, err := client.GenerateRCA(ctx, issue, symptoms)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to generate RCA: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Root Cause Analysis\n\n")
	sb.WriteString(fmt.Sprintf("**Issue:** %s\n\n", rca.Issue))

	emoji := getConfidenceEmoji(rca.Confidence)
	sb.WriteString(fmt.Sprintf("## Probable Root Cause %s\n", emoji))
	sb.WriteString(fmt.Sprintf("%s\n", rca.ProbableRootCause))
	sb.WriteString(fmt.Sprintf("\n**Confidence:** %.0f%%\n\n", rca.Confidence*100))

	if len(rca.ContributingFactors) > 0 {
		sb.WriteString("## Contributing Factors\n")
		for _, f := range rca.ContributingFactors {
			sb.WriteString(fmt.Sprintf("- %s\n", f))
		}
		sb.WriteString("\n")
	}

	if len(rca.SuggestedFixes) > 0 {
		sb.WriteString("## Suggested Fixes\n")
		for i, f := range rca.SuggestedFixes {
			sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, f))
		}
		sb.WriteString("\n")
	}

	if len(rca.SimilarPastIssues) > 0 {
		sb.WriteString("## Similar Past Issues\n")
		for _, i := range rca.SimilarPastIssues {
			sb.WriteString(fmt.Sprintf("- %s\n", i))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleFixSuggest handles the aftrs_fix_suggest tool
func handleFixSuggest(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	symptomsStr, errResult := tools.RequireStringParam(req, "symptoms")
	if errResult != nil {
		return errResult, nil
	}

	symptoms := strings.Split(symptomsStr, ",")
	for i := range symptoms {
		symptoms[i] = strings.TrimSpace(symptoms[i])
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create learning client: %w", err)), nil
	}

	fixes, err := client.SuggestFixes(ctx, symptoms)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to suggest fixes: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Suggested Fixes\n\n")
	sb.WriteString(fmt.Sprintf("**Symptoms:** %s\n\n", symptomsStr))
	sb.WriteString("## Recommended Actions\n\n")

	for i, fix := range fixes {
		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, fix))
	}

	return tools.TextResult(sb.String()), nil
}

// handleTroubleshoot handles the aftrs_troubleshoot tool
func handleTroubleshoot(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	issue, errResult := tools.RequireStringParam(req, "issue")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create learning client: %w", err)), nil
	}

	guide, err := client.GetTroubleshootGuide(ctx, issue)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get troubleshoot guide: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Troubleshooting Guide\n\n")
	sb.WriteString(fmt.Sprintf("**Issue:** %s\n", guide.Issue))
	sb.WriteString(fmt.Sprintf("**Category:** %s\n", guide.Category))
	sb.WriteString(fmt.Sprintf("**Estimated Time:** %d minutes\n\n", guide.EstTime))

	sb.WriteString("## Steps\n\n")

	for _, step := range guide.Steps {
		sb.WriteString(fmt.Sprintf("### Step %d: %s\n\n", step.StepNum, step.Action))
		if step.Expected != "" {
			sb.WriteString(fmt.Sprintf("**Expected:** %s\n", step.Expected))
		}
		if step.IfFails != "" {
			sb.WriteString(fmt.Sprintf("**If fails:** %s\n", step.IfFails))
		}
		if len(step.Tools) > 0 {
			sb.WriteString(fmt.Sprintf("**Tools:** %s\n", strings.Join(step.Tools, ", ")))
		}
		sb.WriteString("\n")
	}

	if len(guide.BasedOn) > 0 {
		sb.WriteString("## Based On\n")
		sb.WriteString(fmt.Sprintf("Patterns: %s\n", strings.Join(guide.BasedOn, ", ")))
	}

	return tools.TextResult(sb.String()), nil
}

// handleSymptomCorrelate handles the aftrs_symptom_correlate tool
func handleSymptomCorrelate(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	symptom, errResult := tools.RequireStringParam(req, "symptom")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create learning client: %w", err)), nil
	}

	correlations, err := client.CorrelateSymptoms(ctx, symptom)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to correlate symptoms: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Symptom Correlations\n\n")
	sb.WriteString(fmt.Sprintf("**Primary Symptom:** %s\n\n", symptom))
	sb.WriteString(fmt.Sprintf("Found **%d** correlated symptoms:\n\n", len(correlations)))

	if len(correlations) == 0 {
		sb.WriteString("No correlations found. More patterns need to be learned.\n")
	} else {
		sb.WriteString("| Related Symptom | Frequency | Common Cause |\n")
		sb.WriteString("|-----------------|-----------|-------------|\n")

		for _, c := range correlations {
			for _, r := range c.Related {
				sb.WriteString(fmt.Sprintf("| %s | %d | %s |\n", r, c.Frequency, c.CommonCause))
			}
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleEquipmentHistory handles the aftrs_equipment_history tool
func handleEquipmentHistory(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	equipment, errResult := tools.RequireStringParam(req, "equipment")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create learning client: %w", err)), nil
	}

	history, err := client.GetEquipmentHistory(ctx, equipment)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get equipment history: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Equipment History\n\n")
	sb.WriteString(fmt.Sprintf("**Equipment:** %s\n\n", history.Equipment))

	// Reliability score with emoji
	emoji := "🟢"
	if history.ReliabilityScore < 50 {
		emoji = "🔴"
	} else if history.ReliabilityScore < 80 {
		emoji = "🟡"
	}

	sb.WriteString(fmt.Sprintf("## Reliability Score: %.0f%% %s\n\n", history.ReliabilityScore, emoji))
	sb.WriteString(fmt.Sprintf("**Total Issues:** %d\n", history.TotalIssues))

	if !history.LastIncident.IsZero() {
		sb.WriteString(fmt.Sprintf("**Last Incident:** %s\n", history.LastIncident.Format("2006-01-02")))
	}

	if len(history.CommonPatterns) > 0 {
		sb.WriteString("\n## Common Issues\n")
		for _, p := range history.CommonPatterns {
			sb.WriteString(fmt.Sprintf("- %s\n", p))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// handleVenuePatterns handles the aftrs_venue_patterns tool
func handleVenuePatterns(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	venue, errResult := tools.RequireStringParam(req, "venue")
	if errResult != nil {
		return errResult, nil
	}

	client, err := getClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create learning client: %w", err)), nil
	}

	pattern, err := client.GetVenuePatterns(ctx, venue)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get venue patterns: %w", err)), nil
	}

	var sb strings.Builder
	sb.WriteString("# Venue Patterns\n\n")
	sb.WriteString(fmt.Sprintf("**Venue:** %s\n\n", pattern.Venue))

	// Success rate with emoji
	emoji := "🟢"
	if pattern.SuccessRate < 50 {
		emoji = "🔴"
	} else if pattern.SuccessRate < 80 {
		emoji = "🟡"
	}

	sb.WriteString(fmt.Sprintf("## Success Rate: %.0f%% %s\n\n", pattern.SuccessRate, emoji))
	sb.WriteString(fmt.Sprintf("**Total Shows:** %d\n", pattern.TotalShows))

	if len(pattern.CommonIssues) > 0 {
		sb.WriteString("\n## Common Issues\n")
		for _, i := range pattern.CommonIssues {
			sb.WriteString(fmt.Sprintf("- %s\n", i))
		}
	}

	if len(pattern.Recommendations) > 0 {
		sb.WriteString("\n## Recommendations\n")
		for _, r := range pattern.Recommendations {
			sb.WriteString(fmt.Sprintf("- %s\n", r))
		}
	}

	if len(pattern.Quirks) > 0 {
		sb.WriteString("\n## Venue Quirks\n")
		for _, q := range pattern.Quirks {
			sb.WriteString(fmt.Sprintf("- %s\n", q))
		}
	}

	return tools.TextResult(sb.String()), nil
}

// Helper functions

func getConfidenceEmoji(confidence float64) string {
	if confidence >= 0.7 {
		return "🟢"
	} else if confidence >= 0.4 {
		return "🟡"
	}
	return "🔴"
}

// init registers this module with the global registry
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
