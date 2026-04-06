// Package diagram provides Go architecture diagram generation tools (D2 and Mermaid).
//
// Tools analyze Go module structure using `go list -json` and produce
// package dependency or interface implementation diagrams in D2 or Mermaid format.
// No external Go analysis libraries are used; all inspection is done via the
// Go toolchain CLI and simple file scanning.
package diagram

import (
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for diagram generation tools.
type Module struct{}

func (m *Module) Name() string {
	return "diagram"
}

func (m *Module) Description() string {
	return "Go architecture diagram generation (D2 and Mermaid)"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_diagram_packages",
				mcp.WithDescription("Generate a package-level dependency diagram for a Go module. Runs `go list -json` to discover packages and their imports, then emits a D2 or Mermaid diagram showing the internal dependency graph."),
				mcp.WithString("repo_path", mcp.Description("Path to Go module root"), mcp.Required()),
				mcp.WithString("package_filter", mcp.Description("Go package pattern (default: ./...)")),
				mcp.WithString("format", mcp.Description("Output format: d2 or mermaid (default: mermaid)"), mcp.Enum("d2", "mermaid")),
				mcp.WithNumber("max_depth", mcp.Description("Max recursion depth, 0=unlimited (default: 0)")),
				mcp.WithBoolean("internal_only", mcp.Description("Only show internal packages, skip stdlib and external (default: true)")),
			),
			Handler:             handleDiagramPackages,
			Category:            "diagram",
			CircuitBreakerGroup: "diagram",
			Tags:                []string{"diagram", "packages", "dependencies", "d2", "mermaid", "architecture"},
			UseCases:   []string{"Visualize Go module dependencies", "Generate architecture diagrams", "Understand package structure"},
			Complexity: tools.ComplexityModerate,
		},
		{
			Tool: mcp.NewTool("aftrs_diagram_interfaces",
				mcp.WithDescription("Generate an interface/type hierarchy diagram for a Go module. Scans Go files for interface and struct definitions, matches method signatures to find implementors, then emits a D2 or Mermaid class diagram. Uses name/signature heuristic matching, not full type checking."),
				mcp.WithString("repo_path", mcp.Description("Path to Go module root"), mcp.Required()),
				mcp.WithString("package_filter", mcp.Description("Go package pattern (default: ./...)")),
				mcp.WithString("format", mcp.Description("Output format: d2 or mermaid (default: mermaid)"), mcp.Enum("d2", "mermaid")),
				mcp.WithString("interface_filter", mcp.Description("Only show implementors of this interface name (optional)")),
			),
			Handler:             handleDiagramInterfaces,
			Category:            "diagram",
			CircuitBreakerGroup: "diagram",
			Tags:                []string{"diagram", "interfaces", "types", "hierarchy", "d2", "mermaid", "architecture"},
			UseCases:   []string{"Visualize interface implementations", "Understand type hierarchy", "Find interface implementors"},
			Complexity: tools.ComplexityModerate,
		},
	}
}

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// PackageInfo holds metadata about a Go package parsed from `go list -json`.
type PackageInfo struct {
	ImportPath string   `json:"import_path"`
	Name       string   `json:"name"`
	Dir        string   `json:"dir"`
	GoFiles    []string `json:"go_files"`
	Imports    []string `json:"imports"`
	ShortPath  string   `json:"short_path"` // path relative to module root
}

// ImportEdge represents a directed dependency between two packages.
type ImportEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// InterfaceInfo holds a discovered interface definition.
type InterfaceInfo struct {
	Package string   `json:"package"`
	Name    string   `json:"name"`
	Methods []string `json:"methods"` // method signatures like "Name() string"
}

// StructInfo holds a discovered struct definition.
type StructInfo struct {
	Package string   `json:"package"`
	Name    string   `json:"name"`
	Methods []string `json:"methods"` // method signatures
}

// ImplEdge records that a struct satisfies an interface (heuristic match).
type ImplEdge struct {
	Struct    string `json:"struct"`
	Interface string `json:"interface"`
	Package   string `json:"package"` // package of the struct
}

// DiagramOutput is the structured result returned by diagram tools.
type DiagramOutput struct {
	Source       string `json:"source"`        // D2 or Mermaid text
	Format       string `json:"format"`        // "d2" or "mermaid"
	PackageCount int    `json:"package_count"`
	TypeCount    int    `json:"type_count"`
	EdgeCount    int    `json:"edge_count"`
}

// init registers this module with the global registry.
func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}
