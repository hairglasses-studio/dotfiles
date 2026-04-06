package diagram

import (
	"strings"
	"testing"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools/testutil"
)

// ---------------------------------------------------------------------------
// Module registration
// ---------------------------------------------------------------------------

func TestModuleInfo(t *testing.T) {
	testutil.AssertModuleValid(t, &Module{})
}

// ---------------------------------------------------------------------------
// D2 emitter — packages
// ---------------------------------------------------------------------------

func TestEmitD2Packages(t *testing.T) {
	packages := []PackageInfo{
		{ImportPath: "example.com/mymod/registry", Name: "registry", ShortPath: "registry"},
		{ImportPath: "example.com/mymod/handler", Name: "handler", ShortPath: "handler"},
		{ImportPath: "example.com/mymod/internal/core", Name: "core", ShortPath: "internal/core"},
	}
	edges := []ImportEdge{
		{From: "handler", To: "registry"},
		{From: "internal/core", To: "registry"},
	}

	out := EmitD2Packages(packages, edges, "example.com/mymod")

	if !strings.Contains(out, "direction: down") {
		t.Error("expected direction: down")
	}
	if !strings.Contains(out, "shape: package") {
		t.Error("expected shape: package for D2 packages")
	}
	if !strings.Contains(out, "-> ") {
		t.Error("expected edge arrows in D2 output")
	}
	if !strings.Contains(out, "imports") {
		t.Error("expected 'imports' label on edges")
	}
	if !strings.Contains(out, "registry") {
		t.Error("expected registry package in output")
	}
	if !strings.Contains(out, "handler") {
		t.Error("expected handler package in output")
	}

	t.Logf("D2 packages output:\n%s", out)
}

func TestEmitD2PackagesEmpty(t *testing.T) {
	out := EmitD2Packages(nil, nil, "example.com/empty")
	if !strings.Contains(out, "direction: down") {
		t.Error("expected direction header even for empty diagram")
	}
}

// ---------------------------------------------------------------------------
// D2 emitter — interfaces
// ---------------------------------------------------------------------------

func TestEmitD2Interfaces(t *testing.T) {
	interfaces := []InterfaceInfo{
		{
			Package: "registry",
			Name:    "ToolModule",
			Methods: []string{"Name() string", "Description() string", "Tools() []ToolDefinition"},
		},
	}
	structs := []StructInfo{
		{Package: "diagram", Name: "Module", Methods: []string{"Name() string", "Description() string", "Tools() []ToolDefinition"}},
	}
	impls := []ImplEdge{
		{Struct: "Module", Interface: "ToolModule", Package: "diagram"},
	}

	out := EmitD2Interfaces(interfaces, structs, impls)

	if !strings.Contains(out, "direction: down") {
		t.Error("expected direction: down")
	}
	if !strings.Contains(out, "shape: class") {
		t.Error("expected shape: class for types")
	}
	if !strings.Contains(out, "implements") {
		t.Error("expected implements label on edges")
	}
	if !strings.Contains(out, "stroke-dash: 3") {
		t.Error("expected dashed stroke for implements edges")
	}
	if !strings.Contains(out, "ToolModule") {
		t.Error("expected ToolModule interface in output")
	}
	if !strings.Contains(out, "#57c7ff") {
		t.Error("expected Snazzy cyan color for interfaces")
	}

	t.Logf("D2 interfaces output:\n%s", out)
}

// ---------------------------------------------------------------------------
// Mermaid emitter — packages
// ---------------------------------------------------------------------------

func TestEmitMermaidPackages(t *testing.T) {
	packages := []PackageInfo{
		{ImportPath: "example.com/mymod/registry", Name: "registry", ShortPath: "registry"},
		{ImportPath: "example.com/mymod/handler", Name: "handler", ShortPath: "handler"},
		{ImportPath: "example.com/mymod/internal/core", Name: "core", ShortPath: "internal/core"},
	}
	edges := []ImportEdge{
		{From: "handler", To: "registry"},
		{From: "internal/core", To: "registry"},
	}

	out := EmitMermaidPackages(packages, edges, "example.com/mymod")

	if !strings.Contains(out, "graph TD") {
		t.Error("expected graph TD header")
	}
	if !strings.Contains(out, "subgraph internal") {
		t.Error("expected subgraph for internal group")
	}
	if !strings.Contains(out, "-->") {
		t.Error("expected --> edges in Mermaid output")
	}
	if !strings.Contains(out, "registry") {
		t.Error("expected registry in output")
	}

	t.Logf("Mermaid packages output:\n%s", out)
}

func TestEmitMermaidPackagesEmpty(t *testing.T) {
	out := EmitMermaidPackages(nil, nil, "example.com/empty")
	if !strings.Contains(out, "graph TD") {
		t.Error("expected graph TD header even for empty diagram")
	}
}

// ---------------------------------------------------------------------------
// Mermaid emitter — interfaces
// ---------------------------------------------------------------------------

func TestEmitMermaidInterfaces(t *testing.T) {
	interfaces := []InterfaceInfo{
		{
			Package: "registry",
			Name:    "ToolModule",
			Methods: []string{"Name() string", "Description() string", "Tools() []ToolDefinition"},
		},
	}
	structs := []StructInfo{
		{Package: "diagram", Name: "Module", Methods: []string{"Name() string", "Description() string"}},
	}
	impls := []ImplEdge{
		{Struct: "Module", Interface: "ToolModule", Package: "diagram"},
	}

	out := EmitMermaidInterfaces(interfaces, structs, impls)

	if !strings.Contains(out, "classDiagram") {
		t.Error("expected classDiagram header")
	}
	if !strings.Contains(out, "<<interface>>") {
		t.Error("expected <<interface>> annotation")
	}
	if !strings.Contains(out, "..|>") {
		t.Error("expected ..|> implementation arrow")
	}
	if !strings.Contains(out, "ToolModule") {
		t.Error("expected ToolModule in output")
	}
	if !strings.Contains(out, "+Name() string") {
		t.Error("expected method signatures with + prefix")
	}

	t.Logf("Mermaid interfaces output:\n%s", out)
}

// ---------------------------------------------------------------------------
// Package parsing helpers
// ---------------------------------------------------------------------------

func TestFilterByDepth(t *testing.T) {
	pkgs := []PackageInfo{
		{ShortPath: "registry"},
		{ShortPath: "internal/core"},
		{ShortPath: "internal/core/deep"},
	}
	edges := []ImportEdge{
		{From: "registry", To: "internal/core"},
		{From: "internal/core", To: "internal/core/deep"},
	}

	filtered, filteredEdges := filterByDepth(pkgs, edges, 2)

	if len(filtered) != 2 {
		t.Errorf("expected 2 packages at depth <= 2, got %d", len(filtered))
	}
	if len(filteredEdges) != 1 {
		t.Errorf("expected 1 edge between depth-filtered packages, got %d", len(filteredEdges))
	}
}

func TestFilterByDepthRootPackage(t *testing.T) {
	pkgs := []PackageInfo{
		{ShortPath: "."},
		{ShortPath: "sub"},
	}
	edges := []ImportEdge{
		{From: ".", To: "sub"},
	}

	filtered, filteredEdges := filterByDepth(pkgs, edges, 1)
	if len(filtered) != 2 {
		t.Errorf("expected 2 packages, got %d", len(filtered))
	}
	if len(filteredEdges) != 1 {
		t.Errorf("expected 1 edge, got %d", len(filteredEdges))
	}
}

// ---------------------------------------------------------------------------
// Interface scanning helpers
// ---------------------------------------------------------------------------

func TestNormalizeParams(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", ""},
		{"ctx context.Context", "context.Context"},
		{"ctx context.Context, name string", "context.Context, string"},
		{"a, b int", "a, int"}, // Imperfect but acceptable for heuristic
		{"req mcp.CallToolRequest", "mcp.CallToolRequest"},
	}

	for _, tc := range tests {
		got := normalizeParams(tc.input)
		if got != tc.expected {
			t.Errorf("normalizeParams(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestMethodNameFromSig(t *testing.T) {
	tests := []struct {
		sig  string
		name string
	}{
		{"Name() string", "Name"},
		{"Tools() []ToolDefinition", "Tools"},
		{"HandleAction(context.Context, mcp.CallToolRequest) *mcp.CallToolResult", "HandleAction"},
	}

	for _, tc := range tests {
		got := methodNameFromSig(tc.sig)
		if got != tc.name {
			t.Errorf("methodNameFromSig(%q) = %q, want %q", tc.sig, got, tc.name)
		}
	}
}

func TestFindImplementations(t *testing.T) {
	interfaces := []InterfaceInfo{
		{
			Package: "tools",
			Name:    "ToolModule",
			Methods: []string{"Name() string", "Description() string", "Tools() []ToolDefinition"},
		},
	}
	structs := []StructInfo{
		{
			Package: "diagram",
			Name:    "Module",
			Methods: []string{"Name() string", "Description() string", "Tools() []ToolDefinition"},
		},
		{
			Package: "other",
			Name:    "Incomplete",
			Methods: []string{"Name() string"},
		},
	}

	impls := findImplementations(interfaces, structs)

	if len(impls) != 1 {
		t.Fatalf("expected 1 implementation, got %d", len(impls))
	}
	if impls[0].Struct != "Module" {
		t.Errorf("expected Module to implement ToolModule, got %s", impls[0].Struct)
	}
	if impls[0].Interface != "ToolModule" {
		t.Errorf("expected interface ToolModule, got %s", impls[0].Interface)
	}
}

func TestFindImplementationsEmptyInterface(t *testing.T) {
	interfaces := []InterfaceInfo{
		{Package: "io", Name: "Reader", Methods: nil},
	}
	structs := []StructInfo{
		{Package: "buf", Name: "Buffer", Methods: []string{"Read([]byte) int, error"}},
	}

	impls := findImplementations(interfaces, structs)
	if len(impls) != 0 {
		t.Errorf("expected 0 implementations for empty interface, got %d", len(impls))
	}
}

// ---------------------------------------------------------------------------
// D2 helper functions
// ---------------------------------------------------------------------------

func TestD2ID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"registry", "registry"},
		{"internal/core", "internal_core"},
		{"my-package", "my_package"},
		{".", "root"},
	}

	for _, tc := range tests {
		got := d2ID(tc.input)
		if got != tc.expected {
			t.Errorf("d2ID(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestMermaidID(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"registry", "registry"},
		{"internal/core", "internal_core"},
		{".", "root"},
	}

	for _, tc := range tests {
		got := mermaidID(tc.input)
		if got != tc.expected {
			t.Errorf("mermaidID(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}
