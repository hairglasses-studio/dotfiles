package dep_graph

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for dependency graph tools.
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

func (m *Module) Name() string        { return "dep_graph" }
func (m *Module) Description() string { return "Cross-repo Go dependency graph generation and drift detection" }

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		{
			Tool: mcp.NewTool("aftrs_dep_graph",
				mcp.WithDescription("Generate a cross-repo dependency graph for ~/hairglasses-studio. Shows how Go modules depend on each other across the org."),
				mcp.WithString("format", mcp.Description("Output format: json, mermaid, or dot (default: json)")),
				mcp.WithBoolean("include_replaces", mcp.Description("Include replace directives in the graph (default: true)")),
				mcp.WithBoolean("include_external", mcp.Description("Include external (non-org) dependencies (default: false)")),
				mcp.WithString("filter_repo", mcp.Description("Only show edges involving this repo name")),
			),
			Handler:             handleDepGraph,
			Category:            "dep_graph",
			Subcategory:         "graph",
			Tags:                []string{"dependency", "graph", "go.mod", "cross-repo", "modules"},
			UseCases:            []string{"Visualize cross-repo deps", "Audit module graph", "Find dependency paths"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "dep_graph",
		},
		{
			Tool: mcp.NewTool("aftrs_dep_drift",
				mcp.WithDescription("Detect version drift across repos: where repos use different versions of the same Go dependency."),
				mcp.WithString("module_filter", mcp.Description("Only check drift for this module (e.g. 'mcpkit')")),
				mcp.WithString("severity", mcp.Description("Filter by severity: all, major, minor (default: all)")),
			),
			Handler:             handleDepDrift,
			Category:            "dep_graph",
			Subcategory:         "drift",
			Tags:                []string{"dependency", "drift", "version", "skew", "audit"},
			UseCases:            []string{"Find version skew", "Audit dependency freshness", "Pre-release checks"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "dep_graph",
		},
	}
}

func handleDepGraph(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	format := tools.OptionalStringParam(req, "format", "json")
	includeReplaces := tools.GetBoolParam(req, "include_replaces", true)
	includeExternal := tools.GetBoolParam(req, "include_external", false)
	filterRepo := tools.GetStringParam(req, "filter_repo")

	// Validate format.
	switch format {
	case "json", "mermaid", "dot":
	default:
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("unsupported format %q: must be json, mermaid, or dot", format)), nil
	}

	root, err := studioRoot()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	// Collect edges.
	allEdges, err := CollectEdges(root)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to collect dependency edges: %w", err)), nil
	}

	// Filter to internal-only if not including external.
	edges := allEdges
	if !includeExternal {
		edges = FilterEdgesInternal(allEdges)
	}

	// Filter by repo if specified.
	if filterRepo != "" {
		edges = FilterEdgesByRepo(edges, filterRepo)
	}

	// Collect replaces if requested.
	var replaces []Replace
	if includeReplaces {
		replaces, _ = CollectReplaces(root) // best-effort
		if filterRepo != "" {
			var filtered []Replace
			for _, r := range replaces {
				if r.Repo == filterRepo {
					filtered = append(filtered, r)
				}
			}
			replaces = filtered
		}
	}

	// Collect go.work info.
	workspaces, _ := CollectGoWorkModules(root) // best-effort

	output := GraphOutput{
		Generated:    time.Now().UTC().Format(time.RFC3339),
		Format:       format,
		Edges:        edges,
		Replaces:     replaces,
		GoWorkspaces: workspaces,
		RepoCount:    CountUniqueRepos(edges),
		EdgeCount:    len(edges),
	}

	// Render to text format if requested.
	switch format {
	case "mermaid":
		output.Rendered = FormatMermaid(edges, replaces)
	case "dot":
		output.Rendered = FormatDOT(edges, replaces)
	}

	return tools.JSONResult(output), nil
}

func handleDepDrift(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	moduleFilter := tools.GetStringParam(req, "module_filter")
	severity := tools.OptionalStringParam(req, "severity", "all")

	switch severity {
	case "all", "major", "minor":
	default:
		return tools.CodedErrorResult(tools.ErrInvalidParam, fmt.Errorf("unsupported severity %q: must be all, major, or minor", severity)), nil
	}

	root, err := studioRoot()
	if err != nil {
		return tools.ErrorResult(err), nil
	}

	edges, err := CollectEdges(root)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("failed to collect dependency edges: %w", err)), nil
	}

	// Group by "To" module: module -> (sourceRepo -> version).
	type versionInfo struct {
		repo    string
		version string
	}
	moduleVersions := make(map[string][]versionInfo)
	for _, e := range edges {
		if e.ToVersion == "" {
			continue
		}
		// If module_filter is set, only match modules containing the filter string.
		if moduleFilter != "" && !strings.Contains(e.To, moduleFilter) {
			continue
		}
		moduleVersions[e.To] = append(moduleVersions[e.To], versionInfo{
			repo:    e.SourceRepo,
			version: e.ToVersion,
		})
	}

	// Detect drift: modules with >1 unique version across consumers.
	var driftEntries []DriftEntry
	for mod, infos := range moduleVersions {
		// Deduplicate: keep latest version per repo.
		repoVersions := make(map[string]string)
		for _, vi := range infos {
			existing, ok := repoVersions[vi.repo]
			if !ok || compareVersions(vi.version, existing) > 0 {
				repoVersions[vi.repo] = vi.version
			}
		}

		// Check for drift.
		uniqueVersions := make(map[string]bool)
		for _, v := range repoVersions {
			uniqueVersions[v] = true
		}
		if len(uniqueVersions) <= 1 {
			continue
		}

		// Determine latest and severity.
		latest := findLatestVersion(repoVersions)
		sev := classifyDrift(repoVersions)

		// Apply severity filter.
		if severity != "all" && sev != severity {
			continue
		}

		behindCount := 0
		for _, v := range repoVersions {
			if v != latest {
				behindCount++
			}
		}

		driftEntries = append(driftEntries, DriftEntry{
			Module:      mod,
			Versions:    repoVersions,
			Severity:    sev,
			Latest:      latest,
			BehindCount: behindCount,
		})
	}

	// Sort drift entries by severity (major first), then module name.
	sort.Slice(driftEntries, func(i, j int) bool {
		si := severityRank(driftEntries[i].Severity)
		sj := severityRank(driftEntries[j].Severity)
		if si != sj {
			return si > sj
		}
		return driftEntries[i].Module < driftEntries[j].Module
	})

	report := DriftReport{
		TotalModulesChecked: len(moduleVersions),
		ModulesWithDrift:    len(driftEntries),
		DriftEntries:        driftEntries,
	}

	return tools.JSONResult(report), nil
}

// severityRank returns a numeric rank for sorting (higher = more severe).
func severityRank(s string) int {
	switch s {
	case "major":
		return 3
	case "minor":
		return 2
	case "patch":
		return 1
	default:
		return 0
	}
}

// findLatestVersion returns the highest version string from a map of repo->version.
func findLatestVersion(repoVersions map[string]string) string {
	latest := ""
	for _, v := range repoVersions {
		if latest == "" || compareVersions(v, latest) > 0 {
			latest = v
		}
	}
	return latest
}

// compareVersions does a simple lexicographic comparison of semver strings.
// Returns >0 if a > b, 0 if equal, <0 if a < b.
// This handles the common case of "v0.2.0" vs "v0.4.1" etc.
func compareVersions(a, b string) int {
	aParts := parseVersion(a)
	bParts := parseVersion(b)
	for i := 0; i < 3; i++ {
		if aParts[i] != bParts[i] {
			return aParts[i] - bParts[i]
		}
	}
	return 0
}

// parseVersion extracts [major, minor, patch] from a version string like "v1.2.3".
func parseVersion(v string) [3]int {
	v = strings.TrimPrefix(v, "v")
	// Strip pre-release and build metadata.
	if idx := strings.Index(v, "-"); idx >= 0 {
		v = v[:idx]
	}
	if idx := strings.Index(v, "+"); idx >= 0 {
		v = v[:idx]
	}
	parts := strings.SplitN(v, ".", 3)
	var result [3]int
	for i := 0; i < len(parts) && i < 3; i++ {
		n := 0
		for _, c := range parts[i] {
			if c >= '0' && c <= '9' {
				n = n*10 + int(c-'0')
			}
		}
		result[i] = n
	}
	return result
}

// classifyDrift determines drift severity based on version differences.
func classifyDrift(repoVersions map[string]string) string {
	versions := make([][3]int, 0, len(repoVersions))
	for _, v := range repoVersions {
		versions = append(versions, parseVersion(v))
	}
	if len(versions) < 2 {
		return "patch"
	}

	hasMajorDiff := false
	hasMinorDiff := false
	for i := 1; i < len(versions); i++ {
		if versions[i][0] != versions[0][0] {
			hasMajorDiff = true
		}
		if versions[i][1] != versions[0][1] {
			hasMinorDiff = true
		}
	}

	if hasMajorDiff {
		return "major"
	}
	if hasMinorDiff {
		return "minor"
	}
	return "patch"
}
