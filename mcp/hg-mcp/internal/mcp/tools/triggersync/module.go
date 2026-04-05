// Package triggersync provides MCP tools for cross-system scene/clip/cue triggering.
package triggersync

import (
	"context"
	"fmt"

	"github.com/hairglasses-studio/hg-mcp/internal/clients"
	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Module implements the Trigger Sync tools module
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

// Name returns the module name
func (m *Module) Name() string {
	return "triggersync"
}

// Description returns the module description
func (m *Module) Description() string {
	return "Cross-system scene, clip, and cue triggering for Ableton, Resolume, OBS, and grandMA3"
}

// Tools returns the Trigger Sync tool definitions
func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_trigger_scene",
				mcp.WithDescription("Fire a scene across all connected systems (Ableton scene, Resolume column, OBS scene, grandMA3 cue)"),
				mcp.WithNumber("scene_index", mcp.Required(), mcp.Description("Scene index (0-based)")),
				mcp.WithBoolean("dry_run", mcp.Description("If true, test without actually triggering")),
			),
			Handler:             handleTriggerScene,
			Category:            "triggersync",
			Subcategory:         "control",
			Tags:                []string{"trigger", "scene", "sync", "cross-system"},
			UseCases:            []string{"Fire scene 3 across all systems", "Sync scene change", "Unified cue triggering"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "triggersync",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_trigger_map",
				mcp.WithDescription("View or configure cross-system trigger mappings"),
				mcp.WithString("action", mcp.Description("Action: list, get, delete"), mcp.Enum("list", "get", "delete")),
				mcp.WithString("mapping_id", mcp.Description("Mapping ID (for get/delete actions)")),
			),
			Handler:             handleTriggerMap,
			Category:            "triggersync",
			Subcategory:         "config",
			Tags:                []string{"trigger", "mapping", "config"},
			UseCases:            []string{"List all mappings", "View mapping details", "Delete a mapping"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "triggersync",
		},
		{
			Tool: mcp.NewTool("aftrs_trigger_link",
				mcp.WithDescription("Create a trigger mapping that links actions across systems"),
				mcp.WithString("name", mcp.Required(), mcp.Description("Name for the mapping")),
				mcp.WithString("description", mcp.Description("Description of what this mapping does")),
				mcp.WithArray("triggers", mcp.Required(), mcp.Description("List of trigger targets"), func(schema map[string]any) {
					schema["items"] = map[string]any{
						"type": "object",
						"properties": map[string]any{
							"system":     map[string]any{"type": "string", "enum": []string{"ableton", "resolume", "grandma3", "obs"}},
							"action":     map[string]any{"type": "string"},
							"identifier": map[string]any{"type": "string"},
						},
						"required": []string{"system", "action", "identifier"},
					}
				}),
			),
			Handler:             handleTriggerLink,
			Category:            "triggersync",
			Subcategory:         "config",
			Tags:                []string{"trigger", "link", "mapping", "create"},
			UseCases:            []string{"Link Ableton scene to Resolume column", "Create cue sync mapping", "Define trigger chain"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "triggersync",
			IsWrite:             true,
		},
		{
			Tool: mcp.NewTool("aftrs_trigger_test",
				mcp.WithDescription("Test a trigger chain without actually executing it"),
				mcp.WithString("mapping_id", mcp.Description("Mapping ID to test")),
				mcp.WithNumber("scene_index", mcp.Description("Alternatively, test unified scene trigger")),
			),
			Handler:             handleTriggerTest,
			Category:            "triggersync",
			Subcategory:         "control",
			Tags:                []string{"trigger", "test", "dry-run"},
			UseCases:            []string{"Test trigger chain", "Validate mapping", "Check connectivity"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "triggersync",
		},
		{
			Tool: mcp.NewTool("aftrs_trigger_status",
				mcp.WithDescription("Get trigger sync status and connected systems"),
			),
			Handler:             handleTriggerStatus,
			Category:            "triggersync",
			Subcategory:         "status",
			Tags:                []string{"trigger", "status", "systems"},
			UseCases:            []string{"Check connected systems", "View trigger capabilities", "Verify setup"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "triggersync",
		},
		{
			Tool: mcp.NewTool("aftrs_trigger_health",
				mcp.WithDescription("Check trigger sync health and get troubleshooting recommendations"),
			),
			Handler:             handleTriggerHealth,
			Category:            "triggersync",
			Subcategory:         "status",
			Tags:                []string{"trigger", "health", "diagnostics"},
			UseCases:            []string{"Check trigger health", "Diagnose issues", "Verify connectivity"},
			Complexity:          tools.ComplexitySimple,
			CircuitBreakerGroup: "triggersync",
		},
	}
}

var getTriggerSyncClient = tools.LazyClient(clients.NewTriggerSyncClient)

func handleTriggerScene(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTriggerSyncClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create trigger sync client: %w", err)), nil
	}

	sceneIndex := tools.GetIntParam(req, "scene_index", -1)
	if sceneIndex < 0 {
		return tools.ErrorResult(fmt.Errorf("scene_index is required and must be >= 0")), nil
	}

	dryRun := tools.GetBoolParam(req, "dry_run", false)

	results, err := client.TriggerScene(ctx, sceneIndex, dryRun)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to trigger scene: %w", err)), nil
	}

	// Count successes and failures
	successCount := 0
	failCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		} else {
			failCount++
		}
	}

	return tools.JSONResult(map[string]interface{}{
		"scene_index":   sceneIndex,
		"dry_run":       dryRun,
		"results":       results,
		"success_count": successCount,
		"fail_count":    failCount,
		"message":       fmt.Sprintf("Scene %d triggered on %d/%d systems", sceneIndex, successCount, len(results)),
	}), nil
}

func handleTriggerMap(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTriggerSyncClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create trigger sync client: %w", err)), nil
	}

	action := tools.OptionalStringParam(req, "action", "list")

	mappingID := tools.GetStringParam(req, "mapping_id")

	switch action {
	case "list":
		mappings, err := client.GetMappings(ctx)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to get mappings: %w", err)), nil
		}

		return tools.JSONResult(map[string]interface{}{
			"mappings": mappings,
			"count":    len(mappings),
		}), nil

	case "get":
		if mappingID == "" {
			return tools.ErrorResult(fmt.Errorf("mapping_id is required for get action")), nil
		}
		mappings, err := client.GetMappings(ctx)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to get mappings: %w", err)), nil
		}
		for _, m := range mappings {
			if m.ID == mappingID {
				return tools.JSONResult(m), nil
			}
		}
		return tools.ErrorResult(fmt.Errorf("mapping not found: %s", mappingID)), nil

	case "delete":
		if mappingID == "" {
			return tools.ErrorResult(fmt.Errorf("mapping_id is required for delete action")), nil
		}
		if err := client.DeleteMapping(ctx, mappingID); err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to delete mapping: %w", err)), nil
		}
		return tools.JSONResult(map[string]interface{}{
			"success":    true,
			"mapping_id": mappingID,
			"message":    fmt.Sprintf("Mapping '%s' deleted", mappingID),
		}), nil

	default:
		return tools.ErrorResult(fmt.Errorf("invalid action: %s", action)), nil
	}
}

func handleTriggerLink(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTriggerSyncClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create trigger sync client: %w", err)), nil
	}

	name, errResult := tools.RequireStringParam(req, "name")
	if errResult != nil {
		return errResult, nil
	}

	description := tools.GetStringParam(req, "description")

	// Parse triggers array
	var triggers []clients.TriggerTarget
	if argsMap, ok := req.Params.Arguments.(map[string]interface{}); ok {
		if triggersRaw, exists := argsMap["triggers"]; exists {
			if triggersArray, ok := triggersRaw.([]interface{}); ok {
				for _, t := range triggersArray {
					if triggerMap, ok := t.(map[string]interface{}); ok {
						target := clients.TriggerTarget{
							System:     fmt.Sprintf("%v", triggerMap["system"]),
							Action:     fmt.Sprintf("%v", triggerMap["action"]),
							Identifier: fmt.Sprintf("%v", triggerMap["identifier"]),
						}
						triggers = append(triggers, target)
					}
				}
			}
		}
	}

	if len(triggers) == 0 {
		return tools.ErrorResult(fmt.Errorf("at least one trigger is required")), nil
	}

	mapping := clients.TriggerMapping{
		Name:        name,
		Description: description,
		Triggers:    triggers,
	}

	if err := client.CreateMapping(ctx, mapping); err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to create mapping: %w", err)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"success":       true,
		"name":          name,
		"trigger_count": len(triggers),
		"message":       fmt.Sprintf("Mapping '%s' created with %d triggers", name, len(triggers)),
	}), nil
}

func handleTriggerTest(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTriggerSyncClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create trigger sync client: %w", err)), nil
	}

	mappingID := tools.GetStringParam(req, "mapping_id")
	sceneIndex := tools.GetIntParam(req, "scene_index", -1)

	var results []clients.TriggerResult
	var testErr error

	if mappingID != "" {
		results, testErr = client.TriggerMappingByID(ctx, mappingID, true)
	} else if sceneIndex >= 0 {
		results, testErr = client.TriggerScene(ctx, sceneIndex, true)
	} else {
		return tools.ErrorResult(fmt.Errorf("either mapping_id or scene_index is required")), nil
	}

	if testErr != nil {
		return tools.ErrorResult(fmt.Errorf("test failed: %w", testErr)), nil
	}

	return tools.JSONResult(map[string]interface{}{
		"dry_run": true,
		"results": results,
		"message": "Test completed (no actions executed)",
	}), nil
}

func handleTriggerStatus(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTriggerSyncClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create trigger sync client: %w", err)), nil
	}

	status, err := client.GetStatus(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get status: %w", err)), nil
	}

	return tools.JSONResult(status), nil
}

func handleTriggerHealth(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	client, err := getTriggerSyncClient()
	if err != nil {
		return tools.CodedErrorResult(tools.ErrClientInit, fmt.Errorf("failed to create trigger sync client: %w", err)), nil
	}

	health, err := client.GetHealth(ctx)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to get health: %w", err)), nil
	}

	return tools.JSONResult(health), nil
}
