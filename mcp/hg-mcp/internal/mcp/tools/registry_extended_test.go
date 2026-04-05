package tools

import (
	"context"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

// ---- toolNameToTitle ----

func TestToolNameToTitle(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"aftrs_gmail_send", "Gmail Send"},
		{"aftrs_inventory_list", "Inventory List"},
		{"aftrs_hue_set_color", "Hue Set Color"},
		{"aftrs_a", "A"},
		{"aftrs_", ""},
		{"plain_name", "Plain Name"},
		{"aftrs_single", "Single"},
		{"aftrs_UPPER_case", "UPPER Case"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := toolNameToTitle(tc.name)
			if got != tc.want {
				t.Errorf("toolNameToTitle(%q) = %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

// ---- ObjectOutputSchema ----

func TestObjectOutputSchema(t *testing.T) {
	props := map[string]interface{}{
		"status": map[string]string{"type": "string"},
		"count":  map[string]string{"type": "integer"},
	}
	required := []string{"status"}

	schema := ObjectOutputSchema(props, required)
	if schema == nil {
		t.Fatal("ObjectOutputSchema returned nil")
	}
	if schema.Type != "object" {
		t.Errorf("Type = %q, want %q", schema.Type, "object")
	}
	if len(schema.Properties) != 2 {
		t.Errorf("Properties count = %d, want 2", len(schema.Properties))
	}
	if len(schema.Required) != 1 || schema.Required[0] != "status" {
		t.Errorf("Required = %v, want [status]", schema.Required)
	}
}

func TestObjectOutputSchema_Empty(t *testing.T) {
	schema := ObjectOutputSchema(nil, nil)
	if schema == nil {
		t.Fatal("ObjectOutputSchema returned nil")
	}
	if schema.Type != "object" {
		t.Errorf("Type = %q, want %q", schema.Type, "object")
	}
}

// ---- Registry: ListModules / ListTools ----

func TestNewToolRegistry_Empty(t *testing.T) {
	r := NewToolRegistry()
	if r.ToolCount() != 0 {
		t.Errorf("ToolCount = %d, want 0", r.ToolCount())
	}
	if r.ModuleCount() != 0 {
		t.Errorf("ModuleCount = %d, want 0", r.ModuleCount())
	}
	if modules := r.ListModules(); len(modules) != 0 {
		t.Errorf("ListModules = %v, want empty", modules)
	}
	if tools := r.ListTools(); len(tools) != 0 {
		t.Errorf("ListTools = %v, want empty", tools)
	}
}

func TestRegistryListModules(t *testing.T) {
	r := NewToolRegistry()
	r.RegisterModule(&testMod{tools: []ToolDefinition{
		{Tool: mcp.Tool{Name: "aftrs_t1"}, Category: "test"},
	}})
	modules := r.ListModules()
	if len(modules) != 1 || modules[0] != "testmod" {
		t.Errorf("ListModules = %v, want [testmod]", modules)
	}
}

func TestRegistryListToolsSorted(t *testing.T) {
	r := NewToolRegistry()
	r.RegisterModule(&testMod{tools: []ToolDefinition{
		{Tool: mcp.Tool{Name: "aftrs_zebra"}, Category: "test"},
		{Tool: mcp.Tool{Name: "aftrs_alpha"}, Category: "test"},
		{Tool: mcp.Tool{Name: "aftrs_middle"}, Category: "test"},
	}})
	tools := r.ListTools()
	if len(tools) != 3 {
		t.Fatalf("ListTools count = %d, want 3", len(tools))
	}
	if tools[0] != "aftrs_alpha" || tools[1] != "aftrs_middle" || tools[2] != "aftrs_zebra" {
		t.Errorf("ListTools not sorted: %v", tools)
	}
}

func TestRegistryListToolsByCategory(t *testing.T) {
	r := NewToolRegistry()
	r.RegisterModule(&testMod{tools: []ToolDefinition{
		{Tool: mcp.Tool{Name: "aftrs_a"}, Category: "catA"},
		{Tool: mcp.Tool{Name: "aftrs_b"}, Category: "catB"},
		{Tool: mcp.Tool{Name: "aftrs_c"}, Category: "catA"},
	}})
	catA := r.ListToolsByCategory("catA")
	if len(catA) != 2 {
		t.Errorf("ListToolsByCategory(catA) count = %d, want 2", len(catA))
	}
	catC := r.ListToolsByCategory("catC")
	if len(catC) != 0 {
		t.Errorf("ListToolsByCategory(catC) count = %d, want 0", len(catC))
	}
}

func TestRegistryGetTool_NotFound(t *testing.T) {
	r := NewToolRegistry()
	_, ok := r.GetTool("nonexistent")
	if ok {
		t.Error("GetTool should return false for nonexistent tool")
	}
}

func TestRegistryGetModule_NotFound(t *testing.T) {
	r := NewToolRegistry()
	_, ok := r.GetModule("nonexistent")
	if ok {
		t.Error("GetModule should return false for nonexistent module")
	}
}

func TestRegistryGetAllToolDefinitions(t *testing.T) {
	r := NewToolRegistry()
	r.RegisterModule(&testMod{tools: []ToolDefinition{
		{Tool: mcp.Tool{Name: "aftrs_x"}, Category: "test"},
		{Tool: mcp.Tool{Name: "aftrs_y"}, Category: "test"},
	}})
	all := r.GetAllToolDefinitions()
	if len(all) != 2 {
		t.Errorf("GetAllToolDefinitions count = %d, want 2", len(all))
	}
}

func TestRegistryToolAndModuleCount(t *testing.T) {
	r := NewToolRegistry()
	r.RegisterModule(&testMod{tools: []ToolDefinition{
		{Tool: mcp.Tool{Name: "aftrs_x"}, Category: "test"},
		{Tool: mcp.Tool{Name: "aftrs_y"}, Category: "test"},
	}})
	if r.ToolCount() != 2 {
		t.Errorf("ToolCount = %d, want 2", r.ToolCount())
	}
	if r.ModuleCount() != 1 {
		t.Errorf("ModuleCount = %d, want 1", r.ModuleCount())
	}
}

// ---- GetToolCatalog ----

func TestGetToolCatalog(t *testing.T) {
	r := NewToolRegistry()
	r.RegisterModule(&testMod{tools: []ToolDefinition{
		{Tool: mcp.Tool{Name: "aftrs_a"}, Category: "lighting", Subcategory: "wled"},
		{Tool: mcp.Tool{Name: "aftrs_b"}, Category: "lighting", Subcategory: "hue"},
		{Tool: mcp.Tool{Name: "aftrs_c"}, Category: "lighting"},
		{Tool: mcp.Tool{Name: "aftrs_d"}, Category: "audio"},
	}})

	catalog := r.GetToolCatalog()
	if len(catalog) != 2 {
		t.Errorf("catalog has %d categories, want 2", len(catalog))
	}
	if lighting, ok := catalog["lighting"]; ok {
		if len(lighting) != 3 {
			t.Errorf("lighting has %d subcategories, want 3 (wled, hue, general)", len(lighting))
		}
		if _, ok := lighting["general"]; !ok {
			t.Error("expected 'general' subcategory for tools without explicit subcategory")
		}
	} else {
		t.Error("expected 'lighting' category in catalog")
	}
}

// ---- GetToolStats ----

func TestGetToolStats_Comprehensive(t *testing.T) {
	r := NewToolRegistry()
	r.RegisterModule(&testMod{tools: []ToolDefinition{
		{Tool: mcp.Tool{Name: "aftrs_test_delete"}, Category: "test", Complexity: ComplexitySimple},
		{Tool: mcp.Tool{Name: "aftrs_test_list"}, Category: "test", Complexity: ComplexityModerate},
		{Tool: mcp.Tool{Name: "aftrs_other_get"}, Category: "other", Complexity: ComplexityComplex, Deprecated: true},
	}})

	stats := r.GetToolStats()
	if stats.TotalTools != 3 {
		t.Errorf("TotalTools = %d, want 3", stats.TotalTools)
	}
	if stats.ModuleCount != 1 {
		t.Errorf("ModuleCount = %d, want 1", stats.ModuleCount)
	}
	if stats.ByCategory["test"] != 2 {
		t.Errorf("ByCategory[test] = %d, want 2", stats.ByCategory["test"])
	}
	if stats.ByComplexity["simple"] != 1 {
		t.Errorf("ByComplexity[simple] = %d, want 1", stats.ByComplexity["simple"])
	}
	if stats.WriteToolsCount != 1 { // aftrs_test_delete inferred as write
		t.Errorf("WriteToolsCount = %d, want 1", stats.WriteToolsCount)
	}
	if stats.ReadOnlyCount != 2 {
		t.Errorf("ReadOnlyCount = %d, want 2", stats.ReadOnlyCount)
	}
	if stats.DeprecatedCount != 1 {
		t.Errorf("DeprecatedCount = %d, want 1", stats.DeprecatedCount)
	}
}

// ---- SearchTools ----

func TestSearchTools_ByName(t *testing.T) {
	r := NewToolRegistry()
	r.RegisterModule(&testMod{tools: []ToolDefinition{
		{Tool: mcp.Tool{Name: "aftrs_nanoleaf_status", Description: "Check nanoleaf"}, Category: "nanoleaf", Tags: []string{"lights"}},
		{Tool: mcp.Tool{Name: "aftrs_hue_status", Description: "Check hue"}, Category: "hue"},
	}})

	results := r.SearchTools("nanoleaf")
	if len(results) == 0 {
		t.Fatal("expected results for 'nanoleaf'")
	}
	// First result should be the nanoleaf tool (highest score due to name match)
	if results[0].Tool.Tool.Name != "aftrs_nanoleaf_status" {
		t.Errorf("first result = %q, want aftrs_nanoleaf_status", results[0].Tool.Tool.Name)
	}
	if results[0].MatchType != "name" {
		t.Errorf("match type = %q, want name", results[0].MatchType)
	}
}

func TestSearchTools_ByTag(t *testing.T) {
	r := NewToolRegistry()
	r.RegisterModule(&testMod{tools: []ToolDefinition{
		{Tool: mcp.Tool{Name: "aftrs_x", Description: "Does x"}, Category: "test", Tags: []string{"special_tag"}},
		{Tool: mcp.Tool{Name: "aftrs_y", Description: "Does y"}, Category: "test"},
	}})

	results := r.SearchTools("special_tag")
	if len(results) == 0 {
		t.Fatal("expected results for tag search")
	}
	foundTag := false
	for _, res := range results {
		if res.MatchType == "tag" {
			foundTag = true
			break
		}
	}
	if !foundTag {
		t.Error("expected tag match type in results")
	}
}

func TestSearchTools_NoResults(t *testing.T) {
	r := NewToolRegistry()
	r.RegisterModule(&testMod{tools: []ToolDefinition{
		{Tool: mcp.Tool{Name: "aftrs_x", Description: "Does x"}, Category: "test"},
	}})

	results := r.SearchTools("zzzznonexistent")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestSearchTools_ByDescription(t *testing.T) {
	r := NewToolRegistry()
	r.RegisterModule(&testMod{tools: []ToolDefinition{
		{Tool: mcp.Tool{Name: "aftrs_x", Description: "Manages the lighting controller system"}, Category: "test"},
	}})

	results := r.SearchTools("controller")
	if len(results) == 0 {
		t.Fatal("expected results for description search")
	}
	foundDesc := false
	for _, res := range results {
		if res.MatchType == "description" {
			foundDesc = true
		}
	}
	if !foundDesc {
		t.Error("expected description match type")
	}
}

func TestSearchTools_ShortWordsSkippedInDescription(t *testing.T) {
	r := NewToolRegistry()
	r.RegisterModule(&testMod{tools: []ToolDefinition{
		{Tool: mcp.Tool{Name: "aftrs_z", Description: "An example tool"}, Category: "other"},
	}})

	// Single-word query "an" is <= 2 chars, so should not match on description
	results := r.SearchTools("an")
	for _, res := range results {
		if res.MatchType == "description" {
			t.Error("short words (<=2 chars) should not match on description alone")
		}
	}
}

// ---- GetUnconfiguredCategories ----

func TestGetUnconfiguredCategories(t *testing.T) {
	r := NewToolRegistry()
	r.RegisterModule(&testMod{tools: []ToolDefinition{
		{Tool: mcp.Tool{Name: "aftrs_with_cb"}, Category: "lighting", CircuitBreakerGroup: "wled"},
		{Tool: mcp.Tool{Name: "aftrs_no_cb"}, Category: "orphan"},
	}})

	unconf := r.GetUnconfiguredCategories()
	if _, ok := unconf["orphan"]; !ok {
		t.Error("expected orphan category to be unconfigured (no circuit breaker group)")
	}
	if _, ok := unconf["lighting"]; ok {
		t.Error("expected lighting to NOT be in unconfigured (has circuit breaker group)")
	}
}

// ---- hgToolProfile ----

func TestHgToolProfile_EnvValues(t *testing.T) {
	tests := []struct {
		envVal string
		want   string
	}{
		{"", "default"},
		{"default", "default"},
		{"DEFAULT", "default"},
		{" Default ", "default"},
		{"ops", "ops"},
		{"OPS", "ops"},
		{" ops ", "ops"},
		{"full", "full"},
		{"FULL", "full"},
		{"unknown_value", "default"},
		{"  ", "default"},
	}

	for _, tc := range tests {
		t.Run("env="+tc.envVal, func(t *testing.T) {
			t.Setenv("HG_MCP_PROFILE", tc.envVal)
			got := hgToolProfile()
			if got != tc.want {
				t.Errorf("hgToolProfile() with env=%q = %q, want %q", tc.envVal, got, tc.want)
			}
		})
	}
}

// ---- shouldDeferTool additional cases ----

func TestShouldDeferTool_DefaultProfileCategories(t *testing.T) {
	// Categories that should NOT be deferred in default profile
	eagerCategories := []string{"consolidated", "workflows", "workflow_automation", "studio", "dashboard", "gateway"}
	for _, cat := range eagerCategories {
		td := ToolDefinition{Category: cat, RuntimeGroup: RuntimeGroupShowControl}
		if shouldDeferTool("default", td) {
			t.Errorf("shouldDeferTool(default, %q) should be false", cat)
		}
	}

	// A regular category that should be deferred in default profile
	td := ToolDefinition{Category: "ledfx", RuntimeGroup: RuntimeGroupLighting}
	if !shouldDeferTool("default", td) {
		t.Error("shouldDeferTool(default, ledfx/lighting) should be true")
	}
}

func TestShouldDeferTool_OpsProfile(t *testing.T) {
	// Platform, Infrastructure, ShowControl should NOT defer in ops
	tests := []struct {
		group string
		defer_ bool
	}{
		{RuntimeGroupPlatform, false},
		{RuntimeGroupInfrastructure, false},
		{RuntimeGroupShowControl, false},
		{RuntimeGroupDJMusic, true},
		{RuntimeGroupVJVideo, true},
		{RuntimeGroupLighting, true},
		{RuntimeGroupMessaging, true},
	}
	for _, tc := range tests {
		td := ToolDefinition{Category: "test", RuntimeGroup: tc.group}
		got := shouldDeferTool("ops", td)
		if got != tc.defer_ {
			t.Errorf("shouldDeferTool(ops, group=%q) = %v, want %v", tc.group, got, tc.defer_)
		}
	}
}

// ---- applyMCPAnnotations ----

func TestApplyMCPAnnotations_Local(t *testing.T) {
	td := ToolDefinition{
		Tool:    mcp.Tool{Name: "aftrs_test_list"},
		IsWrite: false,
	}
	annotated := applyMCPAnnotations(td)

	if annotated.Tool.Annotations.Title != "Test List" {
		t.Errorf("title = %q, want %q", annotated.Tool.Annotations.Title, "Test List")
	}
	if annotated.Tool.Annotations.ReadOnlyHint == nil || !*annotated.Tool.Annotations.ReadOnlyHint {
		t.Error("read-only tool should have ReadOnlyHint=true")
	}
	if annotated.Tool.Annotations.DestructiveHint == nil || *annotated.Tool.Annotations.DestructiveHint {
		t.Error("read-only tool should have DestructiveHint=false")
	}
	if annotated.Tool.Annotations.IdempotentHint == nil || !*annotated.Tool.Annotations.IdempotentHint {
		t.Error("read-only tool should have IdempotentHint=true")
	}
	if annotated.Tool.Annotations.OpenWorldHint == nil || !*annotated.Tool.Annotations.OpenWorldHint {
		t.Error("tool should have OpenWorldHint=true")
	}
}

func TestApplyMCPAnnotations_WriteTool(t *testing.T) {
	td := ToolDefinition{
		Tool:    mcp.Tool{Name: "aftrs_test_create"},
		IsWrite: true,
	}
	annotated := applyMCPAnnotations(td)

	if annotated.Tool.Annotations.ReadOnlyHint == nil || *annotated.Tool.Annotations.ReadOnlyHint {
		t.Error("write tool should have ReadOnlyHint=false")
	}
	if annotated.Tool.Annotations.DestructiveHint == nil || !*annotated.Tool.Annotations.DestructiveHint {
		t.Error("write tool should have DestructiveHint=true")
	}
	if annotated.Tool.Annotations.IdempotentHint == nil || *annotated.Tool.Annotations.IdempotentHint {
		t.Error("write tool should have IdempotentHint=false")
	}
}

// ---- RegisterModule overwrites on duplicate name ----

func TestRegisterModuleOverwrites(t *testing.T) {
	r := NewToolRegistry()

	r.RegisterModule(&testMod{tools: []ToolDefinition{
		{Tool: mcp.Tool{Name: "aftrs_x"}, Category: "test"},
	}})

	// Register same module name again with a different tool
	r.RegisterModule(&testMod{tools: []ToolDefinition{
		{Tool: mcp.Tool{Name: "aftrs_y"}, Category: "test"},
	}})

	// Module should be overwritten
	if r.ModuleCount() != 1 {
		t.Errorf("ModuleCount = %d, want 1 (overwrite)", r.ModuleCount())
	}
	// Both tools should exist (tools are accumulated)
	if r.ToolCount() != 2 {
		t.Errorf("ToolCount = %d, want 2", r.ToolCount())
	}
}

// ---- LazyClient ----

func TestLazyClient_Success(t *testing.T) {
	callCount := 0
	get := LazyClient(func() (string, error) {
		callCount++
		return "hello", nil
	})

	val1, err := get()
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	if val1 != "hello" {
		t.Errorf("first call = %q, want hello", val1)
	}

	val2, err := get()
	if err != nil {
		t.Fatalf("second call: %v", err)
	}
	if val2 != "hello" {
		t.Errorf("second call = %q, want hello", val2)
	}

	if callCount != 1 {
		t.Errorf("constructor called %d times, want 1", callCount)
	}
}

// ---- handler with context (wrapHandler basic coverage via direct tool call) ----

func TestRegistryGetModule(t *testing.T) {
	r := testRegistry()

	mod, ok := r.GetModule("testmod")
	if !ok {
		t.Fatal("expected testmod to be found")
	}
	if mod.Name() != "testmod" {
		t.Errorf("Name = %q, want testmod", mod.Name())
	}
	if mod.Description() != "test" {
		t.Errorf("Description = %q, want test", mod.Description())
	}
}

// ---- Multiple module registration ----

type anotherTestMod struct {
	name  string
	tools []ToolDefinition
}

func (m *anotherTestMod) Name() string            { return m.name }
func (m *anotherTestMod) Description() string     { return "another test mod" }
func (m *anotherTestMod) Tools() []ToolDefinition { return m.tools }

func TestRegistryMultipleModules(t *testing.T) {
	r := NewToolRegistry()
	r.RegisterModule(&anotherTestMod{
		name: "mod_a",
		tools: []ToolDefinition{
			{Tool: mcp.Tool{Name: "aftrs_a_list"}, Category: "test", Handler: func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) { return nil, nil }},
		},
	})
	r.RegisterModule(&anotherTestMod{
		name: "mod_b",
		tools: []ToolDefinition{
			{Tool: mcp.Tool{Name: "aftrs_b_list"}, Category: "test", Handler: func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) { return nil, nil }},
			{Tool: mcp.Tool{Name: "aftrs_b_create"}, Category: "test", Handler: func(_ context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) { return nil, nil }},
		},
	})

	if r.ModuleCount() != 2 {
		t.Errorf("ModuleCount = %d, want 2", r.ModuleCount())
	}
	if r.ToolCount() != 3 {
		t.Errorf("ToolCount = %d, want 3", r.ToolCount())
	}

	modules := r.ListModules()
	if len(modules) != 2 {
		t.Fatalf("ListModules count = %d, want 2", len(modules))
	}
	// Should be sorted
	if modules[0] != "mod_a" || modules[1] != "mod_b" {
		t.Errorf("ListModules = %v, want [mod_a mod_b]", modules)
	}
}
