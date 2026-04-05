// Package healing provides self-healing remediation MCP tools
package healing

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Module implements the healing tools module
type Module struct{}

// Name returns the module name
func (m *Module) Name() string { return "healing" }

// Description returns the module description
func (m *Module) Description() string {
	return "Self-healing remediation for AV systems"
}

// Tools returns all healing tools
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_healing_playbooks",
				mcp.WithDescription("List available remediation playbooks"),
				mcp.WithString("service",
					mcp.Description("Filter by service (obs, resolume, touchdesigner, ndi, etc.)"),
				),
			),
			Handler:  handleListPlaybooks,
			Category: "healing",
		},
		{
			Tool: mcp.NewTool("aftrs_healing_playbook_get",
				mcp.WithDescription("Get details of a specific playbook"),
				mcp.WithString("playbook_id",
					mcp.Required(),
					mcp.Description("Playbook ID"),
				),
			),
			Handler:  handleGetPlaybook,
			Category: "healing",
		},
		{
			Tool: mcp.NewTool("aftrs_healing_execute",
				mcp.WithDescription("Execute a remediation playbook. High-risk playbooks require approval."),
				mcp.WithString("playbook_id",
					mcp.Required(),
					mcp.Description("Playbook ID to execute"),
				),
				mcp.WithBoolean("auto_approve",
					mcp.Description("Auto-approve high-risk playbooks (default: false)"),
				),
			),
			Handler:  handleExecute,
			Category: "healing",
		},
		{
			Tool: mcp.NewTool("aftrs_healing_status",
				mcp.WithDescription("Get status of a remediation execution"),
				mcp.WithString("execution_id",
					mcp.Required(),
					mcp.Description("Execution ID"),
				),
			),
			Handler:  handleStatus,
			Category: "healing",
		},
		{
			Tool: mcp.NewTool("aftrs_healing_approve",
				mcp.WithDescription("Approve a pending remediation execution"),
				mcp.WithString("execution_id",
					mcp.Required(),
					mcp.Description("Execution ID to approve"),
				),
				mcp.WithString("approved_by",
					mcp.Description("Who is approving"),
				),
			),
			Handler:  handleApprove,
			Category: "healing",
		},
		{
			Tool: mcp.NewTool("aftrs_healing_reject",
				mcp.WithDescription("Reject a pending remediation execution"),
				mcp.WithString("execution_id",
					mcp.Required(),
					mcp.Description("Execution ID to reject"),
				),
				mcp.WithString("reason",
					mcp.Description("Reason for rejection"),
				),
			),
			Handler:  handleReject,
			Category: "healing",
		},
		{
			Tool: mcp.NewTool("aftrs_healing_pending",
				mcp.WithDescription("List pending approval requests"),
			),
			Handler:  handlePending,
			Category: "healing",
		},
		{
			Tool: mcp.NewTool("aftrs_healing_history",
				mcp.WithDescription("Get recent remediation execution history"),
				mcp.WithNumber("limit",
					mcp.Description("Max results (default: 20)"),
				),
			),
			Handler:  handleHistory,
			Category: "healing",
		},
		{
			Tool: mcp.NewTool("aftrs_healing_config",
				mcp.WithDescription("Get or update self-healing configuration"),
				mcp.WithBoolean("enabled",
					mcp.Description("Enable/disable self-healing"),
				),
				mcp.WithNumber("max_auto_risk",
					mcp.Description("Max risk score for auto-execution (0-100)"),
				),
				mcp.WithNumber("require_approval_above",
					mcp.Description("Risk score requiring approval (0-100)"),
				),
			),
			Handler:  handleConfig,
			Category: "healing",
		},
		{
			Tool: mcp.NewTool("aftrs_healing_stats",
				mcp.WithDescription("Get self-healing statistics"),
			),
			Handler:  handleStats,
			Category: "healing",
		},
		// === Adaptive Healing Tools (v2.17) ===
		{
			Tool: mcp.NewTool("aftrs_healing_learn",
				mcp.WithDescription("Record a successful manual fix as a new playbook. Captures the fix sequence for future automated remediation."),
				mcp.WithString("name",
					mcp.Required(),
					mcp.Description("Name for the new playbook"),
				),
				mcp.WithString("issue",
					mcp.Required(),
					mcp.Description("Description of the issue this fixes"),
				),
				mcp.WithString("service",
					mcp.Description("Service/system this playbook applies to (resolume, touchdesigner, obs, etc.)"),
				),
				mcp.WithString("fix_steps",
					mcp.Required(),
					mcp.Description("Steps to fix (comma-separated or newline-separated)"),
				),
				mcp.WithNumber("risk_score",
					mcp.Description("Risk score 0-100 (default: 50)"),
				),
				mcp.WithBoolean("auto_approve",
					mcp.Description("Allow auto-execution without approval (default: false)"),
				),
			),
			Handler:             handleHealingLearn,
			Category:            "healing",
			Subcategory:         "adaptive",
			Tags:                []string{"healing", "learn", "playbook", "capture"},
			UseCases:            []string{"Document manual fixes", "Create automation from experience"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "healing",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_healing_suggest",
				mcp.WithDescription("AI-powered fix suggestions based on symptoms and learned patterns. Combines playbook knowledge with pattern matching."),
				mcp.WithString("symptoms",
					mcp.Required(),
					mcp.Description("Comma-separated list of current symptoms"),
				),
				mcp.WithString("service",
					mcp.Description("Filter suggestions by service (optional)"),
				),
				mcp.WithBoolean("include_playbooks",
					mcp.Description("Include matching playbooks (default: true)"),
				),
				mcp.WithBoolean("include_patterns",
					mcp.Description("Include pattern-based suggestions (default: true)"),
				),
			),
			Handler:             handleHealingSuggest,
			Category:            "healing",
			Subcategory:         "adaptive",
			Tags:                []string{"healing", "suggest", "ai", "fix"},
			UseCases:            []string{"Get fix recommendations", "Intelligent troubleshooting"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "healing",
		},
		{
			Tool: mcp.NewTool("aftrs_healing_pattern",
				mcp.WithDescription("Detect recurring issue patterns from execution history. Identifies issues that happen repeatedly."),
				mcp.WithNumber("days",
					mcp.Description("Days of history to analyze (default: 30)"),
				),
				mcp.WithNumber("min_occurrences",
					mcp.Description("Minimum occurrences to be considered a pattern (default: 3)"),
				),
				mcp.WithString("service",
					mcp.Description("Filter by service (optional)"),
				),
			),
			Handler:             handleHealingPattern,
			Category:            "healing",
			Subcategory:         "adaptive",
			Tags:                []string{"healing", "pattern", "recurring", "analysis"},
			UseCases:            []string{"Find recurring issues", "Identify systemic problems"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "healing",
		},
		{
			Tool: mcp.NewTool("aftrs_healing_auto_enable",
				mcp.WithDescription("Enable or disable auto-fix for specific playbooks. Auto-enabled playbooks run without approval."),
				mcp.WithString("playbook_id",
					mcp.Required(),
					mcp.Description("Playbook ID to configure"),
				),
				mcp.WithBoolean("enabled",
					mcp.Required(),
					mcp.Description("Enable (true) or disable (false) auto-fix"),
				),
				mcp.WithNumber("max_daily",
					mcp.Description("Maximum auto-executions per day (default: 10)"),
				),
			),
			Handler:             handleHealingAutoEnable,
			Category:            "healing",
			Subcategory:         "adaptive",
			Tags:                []string{"healing", "auto", "enable", "configure"},
			UseCases:            []string{"Enable auto-remediation", "Configure automatic fixes"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "healing",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_healing_rollback",
				mcp.WithDescription("Rollback a recent auto-fix execution. Attempts to undo changes made by a playbook."),
				mcp.WithString("execution_id",
					mcp.Required(),
					mcp.Description("Execution ID to rollback"),
				),
				mcp.WithString("reason",
					mcp.Description("Reason for rollback"),
				),
			),
			Handler:             handleHealingRollback,
			Category:            "healing",
			Subcategory:         "adaptive",
			Tags:                []string{"healing", "rollback", "undo", "revert"},
			UseCases:            []string{"Undo failed fix", "Revert auto-remediation"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "healing",
			IsWrite:             true,
		},
	}
}

func handleListPlaybooks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	service := tools.GetStringParam(request, "service")

	client := clients.GetSelfHealingClient()
	playbooks := client.ListPlaybooks(service)

	result := map[string]interface{}{
		"count":     len(playbooks),
		"playbooks": playbooks,
	}

	return tools.JSONResult(result), nil
}

func handleGetPlaybook(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	playbookID := tools.GetStringParam(request, "playbook_id")

	client := clients.GetSelfHealingClient()
	playbook, err := client.GetPlaybook(playbookID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return tools.JSONResult(playbook), nil
}

func handleExecute(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	playbookID := tools.GetStringParam(request, "playbook_id")
	autoApprove := tools.GetBoolParam(request, "auto_approve", false)

	client := clients.GetSelfHealingClient()
	exec, err := client.ExecutePlaybook(playbookID, autoApprove)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]interface{}{
		"execution_id": exec.ID,
		"playbook":     exec.PlaybookName,
		"status":       exec.Status,
		"risk_score":   exec.RiskScore,
	}

	if exec.Status == "approval_required" {
		result["message"] = fmt.Sprintf("Risk score %d requires approval. Use aftrs_healing_approve to proceed.", exec.RiskScore)
	}

	return tools.JSONResult(result), nil
}

func handleStatus(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	executionID := tools.GetStringParam(request, "execution_id")

	client := clients.GetSelfHealingClient()
	exec, err := client.GetExecution(executionID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return tools.JSONResult(exec), nil
}

func handleApprove(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	executionID := tools.GetStringParam(request, "execution_id")
	approvedBy := tools.GetStringParam(request, "approved_by")
	if approvedBy == "" {
		approvedBy = "operator"
	}

	client := clients.GetSelfHealingClient()
	exec, err := client.ApproveExecution(executionID, approvedBy)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	result := map[string]interface{}{
		"status":       "approved",
		"execution_id": exec.ID,
		"playbook":     exec.PlaybookName,
		"approved_by":  exec.ApprovedBy,
	}

	return tools.JSONResult(result), nil
}

func handleReject(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	executionID := tools.GetStringParam(request, "execution_id")
	reason := tools.GetStringParam(request, "reason")
	if reason == "" {
		reason = "rejected by operator"
	}

	client := clients.GetSelfHealingClient()
	if err := client.RejectExecution(executionID, reason); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText(fmt.Sprintf("Execution %s rejected: %s", executionID, reason)), nil
}

func handlePending(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetSelfHealingClient()
	pending := client.ListPendingApprovals()

	result := map[string]interface{}{
		"count":   len(pending),
		"pending": pending,
	}

	return tools.JSONResult(result), nil
}

func handleHistory(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit := tools.GetIntParam(request, "limit", 20)

	client := clients.GetSelfHealingClient()
	executions := client.ListExecutions(limit)

	result := map[string]interface{}{
		"count":      len(executions),
		"executions": executions,
	}

	return tools.JSONResult(result), nil
}

func handleConfig(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetSelfHealingClient()
	config := client.GetConfig()

	// Check if updating
	args, ok := request.Params.Arguments.(map[string]interface{})
	if ok {
		updated := false
		if v, ok := args["enabled"].(bool); ok {
			config.Enabled = v
			updated = true
		}
		if v, ok := args["max_auto_risk"].(float64); ok {
			config.MaxAutoRiskScore = int(v)
			updated = true
		}
		if v, ok := args["require_approval_above"].(float64); ok {
			config.RequireApprovalAbove = int(v)
			updated = true
		}

		if updated {
			client.UpdateConfig(config)
		}
	}

	return tools.JSONResult(config), nil
}

func handleStats(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client := clients.GetSelfHealingClient()
	stats := client.GetStats()

	return tools.JSONResult(stats), nil
}

// === Adaptive Healing Handlers (v2.17) ===

func handleHealingLearn(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, errResult := tools.RequireStringParam(request, "name")
	if errResult != nil {
		return errResult, nil
	}
	issue, errResult := tools.RequireStringParam(request, "issue")
	if errResult != nil {
		return errResult, nil
	}
	fixStepsStr, errResult := tools.RequireStringParam(request, "fix_steps")
	if errResult != nil {
		return errResult, nil
	}
	service := tools.GetStringParam(request, "service")
	riskScore := tools.GetIntParam(request, "risk_score", 50)
	autoApprove := tools.GetBoolParam(request, "auto_approve", false)

	// Parse fix steps (comma or newline separated)
	var fixSteps []string
	if strings.Contains(fixStepsStr, "\n") {
		fixSteps = strings.Split(fixStepsStr, "\n")
	} else {
		fixSteps = strings.Split(fixStepsStr, ",")
	}
	for i := range fixSteps {
		fixSteps[i] = strings.TrimSpace(fixSteps[i])
	}

	// Filter empty steps
	var cleanSteps []string
	for _, step := range fixSteps {
		if step != "" {
			cleanSteps = append(cleanSteps, step)
		}
	}

	if len(cleanSteps) == 0 {
		return mcp.NewToolResultError("at least one fix step is required"), nil
	}

	client := clients.GetSelfHealingClient()

	// Create new playbook from learned fix
	playbook := &clients.RemediationPlaybook{
		ID:          fmt.Sprintf("learned-%d", time.Now().UnixNano()),
		Name:        name,
		Description: issue,
		Service:     service,
		TriggerType: "manual",
		RiskScore:   riskScore,
		Steps:       make([]clients.RemediationStep, len(cleanSteps)),
		Cooldown:    5 * time.Minute,
		Tags:        []string{"learned"},
	}

	// Add auto-enabled tag if requested
	if autoApprove {
		playbook.Tags = append(playbook.Tags, "auto-enabled")
	}

	for i, step := range cleanSteps {
		playbook.Steps[i] = clients.RemediationStep{
			ID:      fmt.Sprintf("%d", i+1),
			Name:    fmt.Sprintf("Step %d", i+1),
			Type:    "command",
			Command: step,
			OnError: "continue",
		}
	}

	if err := client.AddPlaybook(playbook); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to add playbook: %v", err)), nil
	}

	result := map[string]interface{}{
		"status":       "learned",
		"playbook_id":  playbook.ID,
		"name":         playbook.Name,
		"service":      playbook.Service,
		"steps":        len(cleanSteps),
		"risk_score":   playbook.RiskScore,
		"auto_approve": autoApprove,
		"message":      "Playbook created from learned fix. Use aftrs_healing_execute to run it.",
	}

	return tools.JSONResult(result), nil
}

func handleHealingSuggest(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	symptomsStr, errResult := tools.RequireStringParam(request, "symptoms")
	if errResult != nil {
		return errResult, nil
	}
	service := tools.GetStringParam(request, "service")
	includePlaybooks := tools.GetBoolParam(request, "include_playbooks", true)
	includePatterns := tools.GetBoolParam(request, "include_patterns", true)

	symptoms := strings.Split(symptomsStr, ",")
	for i := range symptoms {
		symptoms[i] = strings.TrimSpace(strings.ToLower(symptoms[i]))
	}

	var sb strings.Builder
	sb.WriteString("# Fix Suggestions\n\n")
	sb.WriteString(fmt.Sprintf("**Symptoms:** %s\n", symptomsStr))
	if service != "" {
		sb.WriteString(fmt.Sprintf("**Service Filter:** %s\n", service))
	}
	sb.WriteString("\n")

	suggestionsFound := 0

	// Get playbook-based suggestions
	if includePlaybooks {
		client := clients.GetSelfHealingClient()
		playbooks := client.ListPlaybooks(service)

		sb.WriteString("## Matching Playbooks\n\n")
		matchingPlaybooks := 0
		for _, pb := range playbooks {
			// Simple symptom matching in description
			descLower := strings.ToLower(pb.Description)
			nameLower := strings.ToLower(pb.Name)
			matched := false
			for _, symptom := range symptoms {
				if strings.Contains(descLower, symptom) || strings.Contains(nameLower, symptom) {
					matched = true
					break
				}
			}
			if matched {
				matchingPlaybooks++
				suggestionsFound++
				riskEmoji := "🟢"
				if pb.RiskScore > 70 {
					riskEmoji = "🔴"
				} else if pb.RiskScore > 40 {
					riskEmoji = "🟡"
				}
				sb.WriteString(fmt.Sprintf("### %s %s\n", pb.Name, riskEmoji))
				sb.WriteString(fmt.Sprintf("**ID:** `%s`\n", pb.ID))
				sb.WriteString(fmt.Sprintf("**Description:** %s\n", pb.Description))
				sb.WriteString(fmt.Sprintf("**Risk Score:** %d\n", pb.RiskScore))
				// Check if auto-enabled via tags
				for _, tag := range pb.Tags {
					if tag == "auto-enabled" {
						sb.WriteString("**Auto-Approve:** Yes ✅\n")
						break
					}
				}
				sb.WriteString(fmt.Sprintf("\n*Run with:* `aftrs_healing_execute playbook_id=%s`\n\n", pb.ID))
			}
		}
		if matchingPlaybooks == 0 {
			sb.WriteString("No matching playbooks found.\n\n")
		}
	}

	// Get pattern-based suggestions from learning module
	if includePatterns {
		sb.WriteString("## Pattern-Based Suggestions\n\n")

		learningClient, err := clients.NewLearningClient()
		if err == nil {
			fixes, err := learningClient.SuggestFixes(ctx, symptoms)
			if err == nil && len(fixes) > 0 {
				for i, fix := range fixes {
					suggestionsFound++
					sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, fix))
				}
			} else {
				sb.WriteString("No pattern-based suggestions available.\n")
			}
		} else {
			sb.WriteString("Learning client unavailable.\n")
		}
		sb.WriteString("\n")
	}

	// Summary
	sb.WriteString("## Summary\n\n")
	if suggestionsFound > 0 {
		sb.WriteString(fmt.Sprintf("Found **%d** suggestions for the given symptoms.\n", suggestionsFound))
	} else {
		sb.WriteString("No suggestions found. Consider:\n")
		sb.WriteString("- Using `aftrs_pattern_learn` to document this issue when resolved\n")
		sb.WriteString("- Using `aftrs_healing_learn` to create a playbook from the fix\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleHealingPattern(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	days := tools.GetIntParam(request, "days", 30)
	minOccurrences := tools.GetIntParam(request, "min_occurrences", 3)
	service := tools.GetStringParam(request, "service")

	client := clients.GetSelfHealingClient()

	// Get recent executions
	executions := client.ListExecutions(1000) // Get many to analyze

	// Filter by time
	cutoff := time.Now().AddDate(0, 0, -days)
	var recentExecs []*clients.RemediationExecution
	for _, exec := range executions {
		if exec.StartedAt.After(cutoff) {
			if service == "" || exec.Service == service {
				recentExecs = append(recentExecs, exec)
			}
		}
	}

	// Count playbook usage patterns
	playbookCounts := make(map[string]int)
	playbookNames := make(map[string]string)
	playbookServices := make(map[string]string)
	for _, exec := range recentExecs {
		playbookCounts[exec.PlaybookID]++
		playbookNames[exec.PlaybookID] = exec.PlaybookName
		playbookServices[exec.PlaybookID] = exec.Service
	}

	// Find recurring patterns
	type patternInfo struct {
		PlaybookID   string
		PlaybookName string
		Service      string
		Count        int
	}
	var patterns []patternInfo
	for id, count := range playbookCounts {
		if count >= minOccurrences {
			patterns = append(patterns, patternInfo{
				PlaybookID:   id,
				PlaybookName: playbookNames[id],
				Service:      playbookServices[id],
				Count:        count,
			})
		}
	}

	// Sort by count descending
	for i := 0; i < len(patterns)-1; i++ {
		for j := i + 1; j < len(patterns); j++ {
			if patterns[j].Count > patterns[i].Count {
				patterns[i], patterns[j] = patterns[j], patterns[i]
			}
		}
	}

	var sb strings.Builder
	sb.WriteString("# Recurring Issue Patterns\n\n")
	sb.WriteString(fmt.Sprintf("**Analysis Period:** Last %d days\n", days))
	sb.WriteString(fmt.Sprintf("**Minimum Occurrences:** %d\n", minOccurrences))
	if service != "" {
		sb.WriteString(fmt.Sprintf("**Service Filter:** %s\n", service))
	}
	sb.WriteString(fmt.Sprintf("**Total Executions Analyzed:** %d\n\n", len(recentExecs)))

	if len(patterns) == 0 {
		sb.WriteString("No recurring patterns found matching criteria.\n\n")
		sb.WriteString("This could mean:\n")
		sb.WriteString("- Issues are being resolved effectively\n")
		sb.WriteString("- Not enough history to detect patterns\n")
		sb.WriteString("- Try lowering min_occurrences or increasing days\n")
	} else {
		sb.WriteString("## Detected Patterns\n\n")
		sb.WriteString("| Playbook | Service | Occurrences | Recommendation |\n")
		sb.WriteString("|----------|---------|-------------|----------------|\n")

		for _, p := range patterns {
			rec := "Monitor"
			if p.Count >= 10 {
				rec = "🔴 Investigate root cause"
			} else if p.Count >= 5 {
				rec = "🟡 Consider auto-enable"
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %d | %s |\n",
				p.PlaybookName, p.Service, p.Count, rec))
		}

		sb.WriteString("\n## Recommendations\n\n")
		if len(patterns) > 0 && patterns[0].Count >= 5 {
			sb.WriteString(fmt.Sprintf("- High frequency issue: **%s** (%d times)\n",
				patterns[0].PlaybookName, patterns[0].Count))
			sb.WriteString("  Consider enabling auto-fix: `aftrs_healing_auto_enable`\n")
		}
		sb.WriteString("- Use `aftrs_healing_playbook_get` to review playbook details\n")
		sb.WriteString("- Investigate root causes for frequently recurring issues\n")
	}

	return mcp.NewToolResultText(sb.String()), nil
}

func handleHealingAutoEnable(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	playbookID, errResult := tools.RequireStringParam(request, "playbook_id")
	if errResult != nil {
		return errResult, nil
	}
	enabled := tools.GetBoolParam(request, "enabled", false)
	maxDaily := tools.GetIntParam(request, "max_daily", 10)

	client := clients.GetSelfHealingClient()

	// Get the playbook to verify it exists
	playbook, err := client.GetPlaybook(playbookID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("playbook not found: %v", err)), nil
	}

	// Update auto-approve setting
	if err := client.SetPlaybookAutoApprove(playbookID, enabled, maxDaily); err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("failed to update playbook: %v", err)), nil
	}

	status := "disabled"
	emoji := "🔴"
	if enabled {
		status = "enabled"
		emoji = "✅"
	}

	result := map[string]interface{}{
		"status":      "updated",
		"playbook_id": playbookID,
		"name":        playbook.Name,
		"auto_fix":    status,
		"max_daily":   maxDaily,
		"message":     fmt.Sprintf("%s Auto-fix %s for playbook '%s'", emoji, status, playbook.Name),
	}

	if enabled {
		result["warning"] = "This playbook will now execute automatically without approval when triggered."
	}

	return tools.JSONResult(result), nil
}

func handleHealingRollback(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	executionID, errResult := tools.RequireStringParam(request, "execution_id")
	if errResult != nil {
		return errResult, nil
	}
	reason := tools.GetStringParam(request, "reason")
	if reason == "" {
		reason = "manual rollback requested"
	}

	client := clients.GetSelfHealingClient()

	// Get the execution
	exec, err := client.GetExecution(executionID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("execution not found: %v", err)), nil
	}

	// Check if rollback is possible
	if exec.Status != "success" && exec.Status != "failed" {
		return mcp.NewToolResultError(fmt.Sprintf("cannot rollback execution in '%s' state", exec.Status)), nil
	}

	// Check if execution is recent enough (within 24 hours)
	if exec.CompletedAt != nil && time.Since(*exec.CompletedAt) > 24*time.Hour {
		return mcp.NewToolResultError("execution is too old to rollback (>24 hours)"), nil
	}

	// Attempt rollback
	rollbackResult, err := client.RollbackExecution(executionID, reason)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("rollback failed: %v", err)), nil
	}

	result := map[string]interface{}{
		"status":       "rolled_back",
		"execution_id": executionID,
		"playbook":     exec.PlaybookName,
		"reason":       reason,
		"rolled_back":  rollbackResult.StepsRolledBack,
		"message":      fmt.Sprintf("Rolled back %d steps from execution %s", rollbackResult.StepsRolledBack, executionID[:8]),
	}

	if len(rollbackResult.Warnings) > 0 {
		result["warnings"] = rollbackResult.Warnings
	}

	return tools.JSONResult(result), nil
}
