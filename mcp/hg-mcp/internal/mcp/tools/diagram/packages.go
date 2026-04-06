package diagram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// goListPackage mirrors the subset of `go list -json` output we need.
type goListPackage struct {
	Dir        string   `json:"Dir"`
	ImportPath string   `json:"ImportPath"`
	Name       string   `json:"Name"`
	GoFiles    []string `json:"GoFiles"`
	Imports    []string `json:"Imports"`
	Module     *struct {
		Path string `json:"Path"`
	} `json:"Module"`
}

// handleDiagramPackages implements the aftrs_diagram_packages tool.
func handleDiagramPackages(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoPath, errResult := tools.RequireStringParam(req, "repo_path")
	if errResult != nil {
		return errResult, nil
	}

	pkgFilter := tools.OptionalStringParam(req, "package_filter", "./...")
	format := tools.OptionalStringParam(req, "format", "mermaid")
	maxDepth := tools.GetIntParam(req, "max_depth", 0)
	internalOnly := tools.GetBoolParam(req, "internal_only", true)

	packages, modulePath, err := listPackages(ctx, repoPath, pkgFilter)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("go list failed: %w", err)), nil
	}

	if len(packages) == 0 {
		return tools.TextResult("No packages found matching the filter."), nil
	}

	// Build the set of internal import paths for filtering.
	internalSet := make(map[string]bool, len(packages))
	for _, pkg := range packages {
		internalSet[pkg.ImportPath] = true
	}

	// Convert to PackageInfo and build edges.
	var pkgInfos []PackageInfo
	var edges []ImportEdge

	for _, pkg := range packages {
		shortPath := pkg.ImportPath
		if modulePath != "" && strings.HasPrefix(shortPath, modulePath+"/") {
			shortPath = strings.TrimPrefix(shortPath, modulePath+"/")
		} else if shortPath == modulePath {
			shortPath = "."
		}

		info := PackageInfo{
			ImportPath: pkg.ImportPath,
			Name:       pkg.Name,
			Dir:        pkg.Dir,
			GoFiles:    pkg.GoFiles,
			Imports:    pkg.Imports,
			ShortPath:  shortPath,
		}
		pkgInfos = append(pkgInfos, info)

		for _, imp := range pkg.Imports {
			if internalOnly && !internalSet[imp] {
				continue
			}
			// Derive short paths for edges.
			fromShort := shortPath
			toShort := imp
			if modulePath != "" && strings.HasPrefix(toShort, modulePath+"/") {
				toShort = strings.TrimPrefix(toShort, modulePath+"/")
			} else if toShort == modulePath {
				toShort = "."
			}
			edges = append(edges, ImportEdge{From: fromShort, To: toShort})
		}
	}

	// Apply max_depth filter: only keep packages whose shortPath has at most
	// maxDepth slash-separated segments.
	if maxDepth > 0 {
		pkgInfos, edges = filterByDepth(pkgInfos, edges, maxDepth)
	}

	// Emit diagram.
	var source string
	switch format {
	case "d2":
		source = EmitD2Packages(pkgInfos, edges, modulePath)
	default:
		source = EmitMermaidPackages(pkgInfos, edges, modulePath)
	}

	output := DiagramOutput{
		Source:       source,
		Format:       format,
		PackageCount: len(pkgInfos),
		EdgeCount:    len(edges),
	}

	return tools.JSONResult(output), nil
}

// listPackages runs `go list -json <pattern>` in the given directory.
// It returns the parsed packages and the module path.
func listPackages(ctx context.Context, dir, pattern string) ([]goListPackage, string, error) {
	cmd := exec.CommandContext(ctx, "go", "list", "-json", pattern)
	cmd.Dir = dir
	cmd.Env = append(cmd.Environ(), "GOWORK=off")

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, "", fmt.Errorf("%s: %s", err, stderr.String())
	}

	// go list -json emits concatenated JSON objects, not an array.
	dec := json.NewDecoder(&stdout)
	var packages []goListPackage
	var modulePath string

	for dec.More() {
		var pkg goListPackage
		if err := dec.Decode(&pkg); err != nil {
			return nil, "", fmt.Errorf("failed to decode go list output: %w", err)
		}
		if modulePath == "" && pkg.Module != nil {
			modulePath = pkg.Module.Path
		}
		packages = append(packages, pkg)
	}

	return packages, modulePath, nil
}

// filterByDepth keeps only packages whose ShortPath has at most maxDepth
// segments, and edges whose both endpoints survive filtering.
func filterByDepth(pkgs []PackageInfo, edges []ImportEdge, maxDepth int) ([]PackageInfo, []ImportEdge) {
	kept := make(map[string]bool)
	var filtered []PackageInfo

	for _, pkg := range pkgs {
		depth := strings.Count(pkg.ShortPath, "/") + 1
		if pkg.ShortPath == "." {
			depth = 0
		}
		if depth <= maxDepth {
			kept[pkg.ShortPath] = true
			filtered = append(filtered, pkg)
		}
	}

	var filteredEdges []ImportEdge
	for _, e := range edges {
		if kept[e.From] && kept[e.To] {
			filteredEdges = append(filteredEdges, e)
		}
	}

	return filtered, filteredEdges
}
