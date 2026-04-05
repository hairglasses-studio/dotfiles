package tools

import (
	"strings"
	"testing"

	"github.com/hairglasses-studio/mcpkit/registry"
	"github.com/mark3labs/mcp-go/mcp"
)

// testRegistry creates an isolated registry with sample tools for testing.
func testRegistry() *ToolRegistry {
	r := NewToolRegistry()

	r.RegisterModule(&testMod{tools: []ToolDefinition{
		{Tool: mcp.Tool{Name: "aftrs_wled_status", Description: "WLED status"}, Category: "wled"},
		{Tool: mcp.Tool{Name: "aftrs_ledfx_status", Description: "LedFX status"}, Category: "ledfx"},
		{Tool: mcp.Tool{Name: "aftrs_grandma3_status", Description: "grandMA3 status"}, Category: "grandma3"},
		{Tool: mcp.Tool{Name: "aftrs_nanoleaf_status", Description: "Nanoleaf status"}, Category: "nanoleaf"},
		{Tool: mcp.Tool{Name: "aftrs_hue_status", Description: "Hue status"}, Category: "hue"},
		{Tool: mcp.Tool{Name: "aftrs_resolume_status", Description: "Resolume status"}, Category: "resolume"},
		{Tool: mcp.Tool{Name: "aftrs_ableton_status", Description: "Ableton status"}, Category: "ableton"},
		{Tool: mcp.Tool{Name: "aftrs_discord_send", Description: "Discord send"}, Category: "discord"},
		{Tool: mcp.Tool{Name: "aftrs_inventory_list", Description: "Inventory list"}, Category: "inventory"},
		{Tool: mcp.Tool{Name: "aftrs_twitch_status", Description: "Twitch status"}, Category: "twitch"},
		{Tool: mcp.Tool{Name: "aftrs_tool_discover", Description: "Tool discovery"}, Category: "discovery"},
		{Tool: mcp.Tool{Name: "aftrs_custom_explicit", Description: "Explicit group"}, Category: "unknown_cat", RuntimeGroup: "custom_group"},
	}})

	return r
}

// testMod implements ToolModule for testing.
type testMod struct {
	tools []ToolDefinition
}

func (m *testMod) Name() string            { return "testmod" }
func (m *testMod) Description() string     { return "test" }
func (m *testMod) Tools() []ToolDefinition { return m.tools }

func TestRuntimeGroupAutoAssignment(t *testing.T) {
	r := testRegistry()

	tests := []struct {
		toolName string
		expected string
	}{
		{"aftrs_wled_status", RuntimeGroupLighting},
		{"aftrs_ledfx_status", RuntimeGroupLighting},
		{"aftrs_grandma3_status", RuntimeGroupLighting},
		{"aftrs_nanoleaf_status", RuntimeGroupLighting},
		{"aftrs_hue_status", RuntimeGroupLighting},
		{"aftrs_resolume_status", RuntimeGroupVJVideo},
		{"aftrs_ableton_status", RuntimeGroupAudioProduction},
		{"aftrs_discord_send", RuntimeGroupMessaging},
		{"aftrs_inventory_list", RuntimeGroupInventory},
		{"aftrs_twitch_status", RuntimeGroupStreaming},
		{"aftrs_tool_discover", RuntimeGroupPlatform},
	}

	for _, tc := range tests {
		td, ok := r.GetTool(tc.toolName)
		if !ok {
			t.Fatalf("tool %s not found", tc.toolName)
		}
		if td.RuntimeGroup != tc.expected {
			t.Errorf("tool %s: got RuntimeGroup=%q, want %q", tc.toolName, td.RuntimeGroup, tc.expected)
		}
	}
}

func TestRuntimeGroupExplicitOverride(t *testing.T) {
	r := testRegistry()

	td, ok := r.GetTool("aftrs_custom_explicit")
	if !ok {
		t.Fatal("tool aftrs_custom_explicit not found")
	}
	if td.RuntimeGroup != "custom_group" {
		t.Errorf("explicit RuntimeGroup not preserved: got %q, want %q", td.RuntimeGroup, "custom_group")
	}
}

func TestListToolsByRuntimeGroup(t *testing.T) {
	r := testRegistry()

	lighting := r.ListToolsByRuntimeGroup(RuntimeGroupLighting)
	if len(lighting) != 5 {
		t.Errorf("expected 5 lighting tools, got %d: %v", len(lighting), lighting)
	}

	for i := 1; i < len(lighting); i++ {
		if lighting[i] < lighting[i-1] {
			t.Errorf("tools not sorted: %v", lighting)
			break
		}
	}
}

func TestGetRuntimeGroupStats(t *testing.T) {
	r := testRegistry()

	stats := r.GetRuntimeGroupStats()
	if stats[RuntimeGroupLighting] != 5 {
		t.Errorf("expected 5 lighting tools, got %d", stats[RuntimeGroupLighting])
	}
	if stats[RuntimeGroupVJVideo] != 1 {
		t.Errorf("expected 1 vj_video tool, got %d", stats[RuntimeGroupVJVideo])
	}
	if stats["custom_group"] != 1 {
		t.Errorf("expected 1 custom_group tool, got %d", stats["custom_group"])
	}
}

func TestToolStatsIncludesRuntimeGroup(t *testing.T) {
	r := testRegistry()

	stats := r.GetToolStats()
	if stats.ByRuntimeGroup == nil {
		t.Fatal("ByRuntimeGroup is nil")
	}
	if stats.ByRuntimeGroup[RuntimeGroupLighting] != 5 {
		t.Errorf("expected 5 lighting in ByRuntimeGroup, got %d", stats.ByRuntimeGroup[RuntimeGroupLighting])
	}
}

func TestSearchToolsByRuntimeGroup(t *testing.T) {
	r := testRegistry()

	results := r.SearchTools("lighting")
	if len(results) == 0 {
		t.Fatal("expected search results for 'lighting'")
	}

	foundRuntimeMatch := false
	for _, res := range results {
		if res.MatchType == "runtime_group" {
			foundRuntimeMatch = true
			break
		}
	}
	if !foundRuntimeMatch {
		t.Error("expected at least one runtime_group match type")
	}
}

func TestInferIsWrite(t *testing.T) {
	writes := []string{
		"aftrs_inventory_create", "aftrs_inventory_delete", "aftrs_discord_send",
		"aftrs_inventory_update", "aftrs_inventory_add", "aftrs_gmail_post",
		"aftrs_inventory_remove", "aftrs_system_reset", "aftrs_rclone_sync",
		"aftrs_inventory_import", "aftrs_ebay_publish", "aftrs_service_start",
		"aftrs_service_stop", "aftrs_workflow_trigger", "aftrs_task_execute",
	}
	for _, name := range writes {
		if !registry.InferIsWrite(name) {
			t.Errorf("InferIsWrite(%q) should be true (write)", name)
		}
	}

	reads := []string{
		"aftrs_inventory_list", "aftrs_inventory_get", "aftrs_system_status",
		"aftrs_lighting_health", "aftrs_tool_discover", "aftrs_ebay_search",
		"aftrs_inventory_stats", "aftrs_system_info", "aftrs_rclone_check",
	}
	for _, name := range reads {
		if registry.InferIsWrite(name) {
			t.Errorf("InferIsWrite(%q) should be false (read-only)", name)
		}
	}
}

func TestApplyMCPAnnotations(t *testing.T) {
	readTool := ToolDefinition{
		Tool:    mcp.Tool{Name: "aftrs_inventory_list"},
		IsWrite: false,
	}
	annotated := registry.ApplyMCPAnnotations(readTool, "aftrs_")

	if annotated.Tool.Annotations.Title != "Inventory List" {
		t.Errorf("title: got %q, want %q", annotated.Tool.Annotations.Title, "Inventory List")
	}
	if annotated.Tool.Annotations.ReadOnlyHint == nil || !*annotated.Tool.Annotations.ReadOnlyHint {
		t.Error("read-only tool should have ReadOnlyHint=true")
	}
	if annotated.Tool.Annotations.DestructiveHint == nil || *annotated.Tool.Annotations.DestructiveHint {
		t.Error("read-only tool should have DestructiveHint=false")
	}

	deleteTool := ToolDefinition{
		Tool:    mcp.Tool{Name: "aftrs_inventory_delete"},
		IsWrite: true,
	}
	annotated = registry.ApplyMCPAnnotations(deleteTool, "aftrs_")

	if annotated.Tool.Annotations.ReadOnlyHint == nil || *annotated.Tool.Annotations.ReadOnlyHint {
		t.Error("write tool should have ReadOnlyHint=false")
	}
	if annotated.Tool.Annotations.DestructiveHint == nil || !*annotated.Tool.Annotations.DestructiveHint {
		t.Error("delete tool should have DestructiveHint=true")
	}

	createTool := ToolDefinition{
		Tool:    mcp.Tool{Name: "aftrs_inventory_create"},
		IsWrite: true,
	}
	annotated = registry.ApplyMCPAnnotations(createTool, "aftrs_")

	if annotated.Tool.Annotations.DestructiveHint == nil || *annotated.Tool.Annotations.DestructiveHint {
		t.Error("create tool should have DestructiveHint=false")
	}

	updateTool := ToolDefinition{
		Tool:    mcp.Tool{Name: "aftrs_config_set"},
		IsWrite: true,
	}
	annotated = registry.ApplyMCPAnnotations(updateTool, "aftrs_")

	if annotated.Tool.Annotations.IdempotentHint == nil || !*annotated.Tool.Annotations.IdempotentHint {
		t.Error("set tool should have IdempotentHint=true")
	}
}

func TestTruncateResponse(t *testing.T) {
	small := mcp.NewToolResultText("hello")
	result := truncateResponse(small)
	if tc, ok := result.Content[0].(mcp.TextContent); ok {
		if tc.Text != "hello" {
			t.Errorf("small response should not be truncated, got %q", tc.Text)
		}
	}

	big := make([]byte, MaxResponseSize+1000)
	for i := range big {
		big[i] = 'x'
	}
	large := mcp.NewToolResultText(string(big))
	result = truncateResponse(large)
	if tc, ok := result.Content[0].(mcp.TextContent); ok {
		if len(tc.Text) > MaxResponseSize+100 {
			t.Errorf("response should be truncated, got length %d", len(tc.Text))
		}
		if !strings.Contains(tc.Text, "[TRUNCATED") {
			t.Error("truncated response should contain truncation marker")
		}
	}

	if truncateResponse(nil) != nil {
		t.Error("nil response should return nil")
	}
}

func TestIsWriteAutoInferInRegisterModule(t *testing.T) {
	r := NewToolRegistry()

	r.RegisterModule(&testMod{tools: []ToolDefinition{
		{Tool: mcp.Tool{Name: "aftrs_test_delete"}, Category: "test"},
		{Tool: mcp.Tool{Name: "aftrs_test_list"}, Category: "test"},
		{Tool: mcp.Tool{Name: "aftrs_test_create"}, Category: "test"},
		{Tool: mcp.Tool{Name: "aftrs_test_send"}, Category: "test"},
	}})

	if td, ok := r.GetTool("aftrs_test_delete"); ok && !td.IsWrite {
		t.Error("_delete tool should be auto-inferred as IsWrite=true")
	}
	if td, ok := r.GetTool("aftrs_test_list"); ok && td.IsWrite {
		t.Error("_list tool should remain IsWrite=false")
	}
	if td, ok := r.GetTool("aftrs_test_create"); ok && !td.IsWrite {
		t.Error("_create tool should be auto-inferred as IsWrite=true")
	}
	if td, ok := r.GetTool("aftrs_test_send"); ok && !td.IsWrite {
		t.Error("_send tool should be auto-inferred as IsWrite=true")
	}
}

func TestCategoryToRuntimeGroupCoverage(t *testing.T) {
	validGroups := map[string]bool{
		RuntimeGroupDJMusic:         true,
		RuntimeGroupVJVideo:         true,
		RuntimeGroupLighting:        true,
		RuntimeGroupAudioProduction: true,
		RuntimeGroupShowControl:     true,
		RuntimeGroupInfrastructure:  true,
		RuntimeGroupMessaging:       true,
		RuntimeGroupInventory:       true,
		RuntimeGroupStreaming:       true,
		RuntimeGroupPlatform:        true,
	}

	for cat, group := range categoryToRuntimeGroup {
		if !validGroups[group] {
			t.Errorf("category %q maps to unknown group %q", cat, group)
		}
	}
}

func TestShouldDeferToolByProfile(t *testing.T) {
	defaultTool := ToolDefinition{Category: "wled", RuntimeGroup: RuntimeGroupLighting}
	if !shouldDeferTool("default", defaultTool) {
		t.Fatal("expected lighting tools to defer in default profile")
	}

	platformTool := ToolDefinition{Category: "discovery", RuntimeGroup: RuntimeGroupPlatform}
	if shouldDeferTool("default", platformTool) {
		t.Fatal("expected platform discovery tools to stay eager in default profile")
	}

	opsTool := ToolDefinition{Category: "system", RuntimeGroup: RuntimeGroupInfrastructure}
	if shouldDeferTool("ops", opsTool) {
		t.Fatal("expected infrastructure tools to stay eager in ops profile")
	}

	fullTool := ToolDefinition{Category: "spotify", RuntimeGroup: RuntimeGroupDJMusic}
	if shouldDeferTool("full", fullTool) {
		t.Fatal("expected full profile to disable deferral")
	}
}
