package tools

import (
	"context"
	"strings"
	"testing"

	"github.com/hairglasses-studio/mcpkit/registry"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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

// deferredTestRegistry creates a registry with tools across multiple runtime groups.
func deferredTestRegistry() *ToolRegistry {
	r := NewToolRegistry()

	r.RegisterModule(&testMod{tools: []ToolDefinition{
		// Platform (always eager)
		{Tool: mcp.Tool{Name: "aftrs_tool_discover"}, Category: "discovery", Handler: func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText("ok"), nil
		}},
		{Tool: mcp.Tool{Name: "aftrs_router_ask"}, Category: "router", Handler: func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText("ok"), nil
		}},
		// Lighting (deferred in default)
		{Tool: mcp.Tool{Name: "aftrs_wled_status"}, Category: "wled", Handler: func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText("ok"), nil
		}},
		{Tool: mcp.Tool{Name: "aftrs_hue_set"}, Category: "hue", Handler: func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText("ok"), nil
		}},
		// DJ Music (deferred in default)
		{Tool: mcp.Tool{Name: "aftrs_rekordbox_list"}, Category: "rekordbox", Handler: func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText("ok"), nil
		}},
		// Infrastructure (eager in ops, deferred in default)
		{Tool: mcp.Tool{Name: "aftrs_unraid_status"}, Category: "unraid", Handler: func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			return mcp.NewToolResultText("ok"), nil
		}},
	}})

	return r
}

func TestRegisterWithServer_DefaultProfile(t *testing.T) {
	r := deferredTestRegistry()
	r.SetProfile("default")

	s := server.NewMCPServer("test", "0.1.0", server.WithToolCapabilities(true))
	r.RegisterWithServer(s)

	// Platform tools should be eager
	if r.DeferredToolCount() == 0 {
		t.Fatal("expected some deferred tools in default profile")
	}
	if r.DeferredToolCount() == r.ToolCount() {
		t.Fatal("expected some eager tools in default profile")
	}

	// Check deferred groups
	counts := r.DeferredGroupCounts()
	if _, ok := counts[RuntimeGroupLighting]; !ok {
		t.Error("expected lighting to be deferred in default profile")
	}
	if _, ok := counts[RuntimeGroupDJMusic]; !ok {
		t.Error("expected dj_music to be deferred in default profile")
	}
	if _, ok := counts[RuntimeGroupPlatform]; ok {
		t.Error("platform should NOT be deferred in default profile")
	}
}

func TestRegisterWithServer_FullProfile(t *testing.T) {
	r := deferredTestRegistry()
	r.SetProfile("full")

	s := server.NewMCPServer("test", "0.1.0", server.WithToolCapabilities(true))
	r.RegisterWithServer(s)

	if r.DeferredToolCount() != 0 {
		t.Errorf("full profile should have 0 deferred tools, got %d", r.DeferredToolCount())
	}
}

func TestLoadDomain(t *testing.T) {
	r := deferredTestRegistry()
	r.SetProfile("default")

	s := server.NewMCPServer("test", "0.1.0", server.WithToolCapabilities(true))
	r.RegisterWithServer(s)

	initialDeferred := r.DeferredToolCount()

	// Load lighting domain
	loaded, err := r.LoadDomain(RuntimeGroupLighting)
	if err != nil {
		t.Fatalf("LoadDomain failed: %v", err)
	}
	if loaded == 0 {
		t.Fatal("expected to load at least one lighting tool")
	}

	// Deferred count should decrease
	if r.DeferredToolCount() >= initialDeferred {
		t.Error("deferred count should have decreased after LoadDomain")
	}

	// Loading same domain again should be a no-op
	loaded2, err := r.LoadDomain(RuntimeGroupLighting)
	if err != nil {
		t.Fatalf("second LoadDomain failed: %v", err)
	}
	if loaded2 != 0 {
		t.Errorf("second load of same domain should return 0, got %d", loaded2)
	}
}

func TestLoadDomain_BeforeServer(t *testing.T) {
	r := NewToolRegistry()

	_, err := r.LoadDomain(RuntimeGroupLighting)
	if err == nil {
		t.Fatal("expected error when calling LoadDomain before RegisterWithServer")
	}
}

func TestLoadAllDeferred(t *testing.T) {
	r := deferredTestRegistry()
	r.SetProfile("default")

	s := server.NewMCPServer("test", "0.1.0", server.WithToolCapabilities(true))
	r.RegisterWithServer(s)

	if r.DeferredToolCount() == 0 {
		t.Fatal("expected deferred tools before LoadAllDeferred")
	}

	total := r.LoadAllDeferred()
	if total == 0 {
		t.Fatal("expected LoadAllDeferred to load at least one tool")
	}
	if r.DeferredToolCount() != 0 {
		t.Errorf("expected 0 deferred tools after LoadAllDeferred, got %d", r.DeferredToolCount())
	}

	// Second call should be no-op
	total2 := r.LoadAllDeferred()
	if total2 != 0 {
		t.Errorf("second LoadAllDeferred should return 0, got %d", total2)
	}
}

func TestGetProfile_Default(t *testing.T) {
	r := NewToolRegistry()
	t.Setenv("HG_MCP_PROFILE", "")

	if got := r.GetProfile(); got != "default" {
		t.Errorf("GetProfile() = %q, want %q", got, "default")
	}
}

func TestSetProfile(t *testing.T) {
	r := NewToolRegistry()
	r.SetProfile("ops")

	if got := r.GetProfile(); got != "ops" {
		t.Errorf("GetProfile() = %q, want %q", got, "ops")
	}
}

func TestDeferredGroupCounts_Empty(t *testing.T) {
	r := NewToolRegistry()
	counts := r.DeferredGroupCounts()
	if len(counts) != 0 {
		t.Errorf("expected empty deferred group counts, got %v", counts)
	}
}

func TestAllRuntimeGroups(t *testing.T) {
	groups := AllRuntimeGroups()
	if len(groups) == 0 {
		t.Fatal("AllRuntimeGroups returned empty")
	}
	// Verify sorted
	for i := 1; i < len(groups); i++ {
		if groups[i] < groups[i-1] {
			t.Errorf("AllRuntimeGroups not sorted: %v", groups)
			break
		}
	}
}

func TestEagerGroups(t *testing.T) {
	// Default profile should have platform only
	def := EagerGroups("default")
	if len(def) != 1 || def[0] != RuntimeGroupPlatform {
		t.Errorf("EagerGroups(default) = %v, want [platform]", def)
	}

	// Ops should have platform + infrastructure + show_control
	ops := EagerGroups("ops")
	if len(ops) != 3 {
		t.Errorf("EagerGroups(ops) = %v, want 3 groups", ops)
	}

	// Full should have all
	full := EagerGroups("full")
	if len(full) != len(AllRuntimeGroups()) {
		t.Errorf("EagerGroups(full) = %d groups, want %d", len(full), len(AllRuntimeGroups()))
	}
}

func TestDeferredGroups(t *testing.T) {
	deferred := DeferredGroups("default")
	if len(deferred) == 0 {
		t.Fatal("default profile should have deferred groups")
	}
	// Platform should NOT be in deferred
	for _, g := range deferred {
		if g == RuntimeGroupPlatform {
			t.Error("platform should not be in deferred groups for default profile")
		}
	}

	// Full should have no deferred groups
	if deferred := DeferredGroups("full"); deferred != nil {
		t.Errorf("full profile should have nil deferred groups, got %v", deferred)
	}
}

func TestRuntimeGroupLabel(t *testing.T) {
	if label := RuntimeGroupLabel(RuntimeGroupDJMusic); label != "DJ & Music" {
		t.Errorf("label for dj_music = %q, want %q", label, "DJ & Music")
	}
	if label := RuntimeGroupLabel("unknown"); label != "unknown" {
		t.Errorf("label for unknown = %q, want %q", label, "unknown")
	}
}
