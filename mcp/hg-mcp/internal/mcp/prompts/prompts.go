// Package prompts provides MCP prompt handlers for guided AI interactions.
package prompts

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// RegisterPrompts registers all MCP prompts with the server
func RegisterPrompts(s *server.MCPServer) {
	// Show preflight checklist prompt
	s.AddPrompt(
		mcp.NewPrompt("show_preflight",
			mcp.WithPromptDescription("Pre-show checklist with system health checks and connectivity verification"),
			mcp.WithArgument("venue",
				mcp.ArgumentDescription("Venue name or identifier"),
				mcp.RequiredArgument(),
			),
			mcp.WithArgument("show_type",
				mcp.ArgumentDescription("Type of show: dj, live, hybrid, or corporate"),
			),
		),
		handleShowPreflight,
	)

	// Troubleshoot system prompt
	s.AddPrompt(
		mcp.NewPrompt("troubleshoot_system",
			mcp.WithPromptDescription("Guided diagnostics for a specific system with pattern matching"),
			mcp.WithArgument("system_name",
				mcp.ArgumentDescription("System to troubleshoot: ableton, resolume, grandma3, obs, touchdesigner, atem"),
				mcp.RequiredArgument(),
			),
		),
		handleTroubleshootSystem,
	)

	// Setup BPM sync prompt
	s.AddPrompt(
		mcp.NewPrompt("setup_bpm_sync",
			mcp.WithPromptDescription("Step-by-step BPM synchronization configuration across systems"),
			mcp.WithArgument("master_source",
				mcp.ArgumentDescription("Master tempo source: ableton, resolume, or manual"),
				mcp.RequiredArgument(),
			),
		),
		handleSetupBPMSync,
	)

	// Create workflow prompt
	s.AddPrompt(
		mcp.NewPrompt("create_workflow",
			mcp.WithPromptDescription("Guide user through creating a custom workflow from available tools"),
			mcp.WithArgument("goal",
				mcp.ArgumentDescription("What the workflow should accomplish"),
				mcp.RequiredArgument(),
			),
		),
		handleCreateWorkflow,
	)

	// Investigate issue prompt
	s.AddPrompt(
		mcp.NewPrompt("investigate_issue",
			mcp.WithPromptDescription("Pattern-matched troubleshooting based on symptoms"),
			mcp.WithArgument("symptoms",
				mcp.ArgumentDescription("Description of the issue or symptoms observed"),
				mcp.RequiredArgument(),
			),
		),
		handleInvestigateIssue,
	)
}

// handleShowPreflight generates a pre-show checklist prompt
func handleShowPreflight(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	venue := req.Params.Arguments["venue"]
	showType := req.Params.Arguments["show_type"]
	if showType == "" {
		showType = "live"
	}

	systemChecks := getSystemChecksForShowType(showType)

	promptText := fmt.Sprintf(`# Pre-Show Checklist for %s

## Show Type: %s

Please run through this checklist to ensure all systems are ready:

### 1. System Health Checks
%s

### 2. Connectivity Verification
- Run: aftrs_trigger_status
- Run: aftrs_sync_bpm_status
- Verify all systems show "connected": true

### 3. Audio/Visual Checks
- Run: aftrs_resolume_status
- Run: aftrs_ableton_status
- Run: aftrs_obs_status

### 4. Backup Verification
- Run: aftrs_backup_all (with dry_run: true)
- Confirm backup paths are accessible

### 5. BPM Sync Setup
- Run: aftrs_sync_bpm_master with source based on show type
- Run: aftrs_sync_bpm_link for each system

### Recommended Actions
After completing the checklist:
1. Use aftrs_workflow_run with "show_startup" workflow
2. Run aftrs_trigger_test to verify cross-system triggers
3. Monitor with aftrs_trigger_health

Please start by running the health check tools listed above and report any issues.`, venue, showType, systemChecks)

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Pre-show checklist for %s (%s)", venue, showType),
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: promptText,
				},
			},
		},
	}, nil
}

// handleTroubleshootSystem generates a troubleshooting prompt for a specific system
func handleTroubleshootSystem(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	systemName := strings.ToLower(req.Params.Arguments["system_name"])

	diagnosticSteps := getDiagnosticSteps(systemName)

	promptText := fmt.Sprintf(`# Troubleshooting Guide: %s

## Step 1: Check Connection Status
Run: aftrs_%s_status

## Step 2: Verify Health
Run: aftrs_%s_health (if available)

## Step 3: Common Issues for %s
%s

## Step 4: Pattern Matching
Run: aftrs_pattern_match with the error message or symptoms

## Step 5: Knowledge Graph Context
Run: aftrs_context_from_graph with entity: "%s"

## Recovery Actions
If issues persist:
1. Check network connectivity to the %s host
2. Verify the application is running
3. Check for port conflicts
4. Review recent configuration changes

Please start with Step 1 and report the status output.`,
		systemName, systemName, systemName, systemName, diagnosticSteps, systemName, systemName)

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Troubleshooting guide for %s", systemName),
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: promptText,
				},
			},
		},
	}, nil
}

// handleSetupBPMSync generates a BPM sync configuration prompt
func handleSetupBPMSync(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	masterSource := strings.ToLower(req.Params.Arguments["master_source"])

	promptText := fmt.Sprintf(`# BPM Sync Configuration

## Master Source: %s

### Step 1: Set Master Source
Run: aftrs_sync_bpm_master with source: "%s"

### Step 2: Link Systems
For each system that should follow the master:

**Resolume (visuals):**
Run: aftrs_sync_bpm_link with system: "resolume", linked: true

**grandMA3 (lighting):**
Run: aftrs_sync_bpm_link with system: "grandma3", linked: true

### Step 3: Initial Sync
Run: aftrs_sync_bpm_push (without bpm parameter to sync from master)

### Step 4: Verify Sync
Run: aftrs_sync_bpm_status
- Check that all systems show the same BPM
- Verify "in_sync": true

### Step 5: Test Tap Tempo
Run: aftrs_sync_tap_tempo multiple times to test live BPM adjustment

### Troubleshooting
If systems drift:
- Run: aftrs_sync_bpm_health for diagnostics
- Check network latency between systems
- Verify OSC ports are not blocked

Please proceed with Step 1 to set the master source.`, masterSource, masterSource)

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("BPM sync setup with %s as master", masterSource),
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: promptText,
				},
			},
		},
	}, nil
}

// handleCreateWorkflow generates a workflow creation guide
func handleCreateWorkflow(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	goal := req.Params.Arguments["goal"]

	promptText := fmt.Sprintf(`# Workflow Creation Guide

## Goal: %s

### Step 1: Identify Required Tools
First, let's find the tools needed for this workflow:
Run: aftrs_tool_search with query matching your goal

### Step 2: Plan the Steps
Based on the available tools, we'll create a sequence of steps.

Common workflow patterns:
- **Startup workflows**: health_check → load_config → start_services → verify
- **Backup workflows**: pause → capture_state → save → verify → resume
- **Sync workflows**: get_state → compare → push_changes → verify

### Step 3: Define Dependencies
Each step can depend on previous steps:
- Use DependsOn to specify which steps must complete first
- Steps without dependencies can run in parallel

### Step 4: Add Error Handling
For each step, define:
- RetryCount: number of retries on failure
- OnFailure: "continue", "stop", or "rollback"

### Step 5: Create the Workflow
Run: aftrs_workflow_compose with:
- name: descriptive name
- description: what it does
- steps: array of step definitions

### Step 6: Validate
Run: aftrs_workflow_validate with the workflow ID

### Step 7: Test
Run: aftrs_workflow_run with dry_run: true

Let's start by searching for relevant tools. What specific systems or operations does this workflow need to interact with?`, goal)

	return &mcp.GetPromptResult{
		Description: fmt.Sprintf("Workflow creation guide for: %s", goal),
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: promptText,
				},
			},
		},
	}, nil
}

// handleInvestigateIssue generates an issue investigation prompt
func handleInvestigateIssue(ctx context.Context, req mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	symptoms := req.Params.Arguments["symptoms"]

	promptText := fmt.Sprintf(`# Issue Investigation

## Reported Symptoms
%s

### Step 1: Pattern Matching
Run: aftrs_pattern_match with symptoms: "%s"

This will search the knowledge base for similar past issues.

### Step 2: System Health Overview
Run these health checks to identify affected systems:
- aftrs_trigger_health
- aftrs_sync_bpm_health
- aftrs_resolume_health
- aftrs_ableton_health
- aftrs_grandma3_health

### Step 3: Knowledge Graph Context
Run: aftrs_context_from_graph with relevant entity names from symptoms

### Step 4: Recent Activity
Check if this is related to recent changes:
- Review workflow execution history
- Check for recent configuration changes

### Step 5: Correlation Analysis
Based on findings, identify:
- Which system(s) are affected
- When the issue started
- What changed recently

### Step 6: Resolution
Based on pattern matches:
1. Apply suggested fix from knowledge base
2. Or escalate with detailed findings

### Step 7: Learning
After resolution:
Run: aftrs_learn_from_resolution to update the knowledge base

Please start with Step 1 to search for similar past issues.`, symptoms, symptoms)

	return &mcp.GetPromptResult{
		Description: "Issue investigation based on symptoms",
		Messages: []mcp.PromptMessage{
			{
				Role: mcp.RoleUser,
				Content: mcp.TextContent{
					Type: "text",
					Text: promptText,
				},
			},
		},
	}, nil
}

// getSystemChecksForShowType returns system checks based on show type
func getSystemChecksForShowType(showType string) string {
	switch strings.ToLower(showType) {
	case "dj":
		return `- Run: aftrs_traktor_status or aftrs_serato_status
- Run: aftrs_midi_status (for controllers)
- Run: aftrs_resolume_status (for visuals)
- Run: aftrs_grandma3_status (for lighting)`
	case "live":
		return `- Run: aftrs_ableton_status
- Run: aftrs_resolume_status
- Run: aftrs_grandma3_status
- Run: aftrs_showkontrol_status (for timecode)`
	case "hybrid":
		return `- Run: aftrs_ableton_status
- Run: aftrs_traktor_status or aftrs_serato_status
- Run: aftrs_resolume_status
- Run: aftrs_grandma3_status
- Run: aftrs_midi_status`
	case "corporate":
		return `- Run: aftrs_obs_status (for streaming)
- Run: aftrs_atem_status (for switching)
- Run: aftrs_ptz_status (for cameras)
- Run: aftrs_grandma3_status (for lighting)`
	default:
		return `- Run: aftrs_consolidated_health
- Run: aftrs_trigger_status
- Run: aftrs_sync_bpm_status`
	}
}

// getDiagnosticSteps returns diagnostic steps for a specific system
func getDiagnosticSteps(systemName string) string {
	switch systemName {
	case "ableton":
		return `**Common Ableton Issues:**
- OSC not receiving: Check M4L device is loaded and OSC port matches
- Scene triggers not working: Verify track/scene indices are correct
- BPM sync drift: Check Link is enabled in Ableton preferences`
	case "resolume":
		return `**Common Resolume Issues:**
- API not responding: Enable API in Resolume settings
- Clips not triggering: Check layer/column indices (1-based)
- BPM not syncing: Verify tempo source is set correctly`
	case "grandma3":
		return `**Common grandMA3 Issues:**
- OSC connection failed: Check IP address and OSC port in console
- Cues not firing: Verify executor and sequence numbers
- Speed master not syncing: Check BPM fader assignment`
	case "obs":
		return `**Common OBS Issues:**
- WebSocket connection failed: Enable WebSocket server in OBS
- Scene switch failed: Check scene name matches exactly
- Stream not starting: Verify stream settings are configured`
	case "touchdesigner":
		return `**Common TouchDesigner Issues:**
- HTTP API not responding: Check web server is enabled
- Parameters not updating: Verify parameter paths
- NDI output not visible: Check NDI tools are installed`
	case "atem":
		return `**Common ATEM Issues:**
- Connection timeout: Verify IP address and network access
- Input switch failed: Check input numbers match physical connections
- Macro not running: Verify macro is recorded and index is correct`
	default:
		return `**General Troubleshooting:**
- Check network connectivity
- Verify application is running
- Check port numbers and IP addresses
- Review recent configuration changes`
	}
}
