// Package repo_analysis provides MCP tools for scanning, comparing, and
// analyzing Go repositories across the hairglasses-studio fleet.
package repo_analysis

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/hairglasses-studio/hg-mcp/internal/mcp/tools"
)

// Module implements the ToolModule interface for repo analysis tools.
type Module struct{}

func init() {
	tools.GetRegistry().RegisterModule(&Module{})
}

func (m *Module) Name() string {
	return "repo_analysis"
}

func (m *Module) Description() string {
	return "Repository scanning, comparison, and fleet-wide analysis tools"
}

func (m *Module) Tools() []tools.ToolDefinition {
	return []tools.ToolDefinition{
		// === Repo Scan (1) ===
		{
			Tool: mcp.NewTool("aftrs_repo_scan",
				mcp.WithDescription("Scan a single repository and return structured metadata including languages, frameworks, dependencies, and classification."),
				mcp.WithString("repo_path", mcp.Required(), mcp.Description("Absolute path or repo name (resolved to ~/hairglasses-studio/<name>)")),
				mcp.WithString("mode", mcp.Description("Scan mode: 'fast' (file-based) or 'deep' (uses go commands). Default: fast")),
				mcp.WithBoolean("include_loc", mcp.Description("Include line-of-code counts per language. Default: false")),
			),
			Handler:             handleRepoScan,
			Category:            "repo_analysis",
			Subcategory:         "scan",
			Tags:                []string{"repo", "scan", "analysis", "go", "dependencies"},
			UseCases:            []string{"Analyze repo structure", "Detect frameworks and dependencies", "Classify repo type"},
			Complexity:          tools.ComplexityModerate,
			CircuitBreakerGroup: "repo_analysis",
		},
		// === Repo Compare (1) ===
		{
			Tool: mcp.NewTool("aftrs_repo_compare",
				mcp.WithDescription("Compare two or more repositories side-by-side: languages, frameworks, dependencies, and version drift."),
				mcp.WithString("repos", mcp.Required(), mcp.Description("Comma-separated repo names or absolute paths")),
				mcp.WithString("focus", mcp.Description("Focus area: 'languages', 'frameworks', 'deps', or 'all'. Default: all")),
			),
			Handler:             handleRepoCompare,
			Category:            "repo_analysis",
			Subcategory:         "compare",
			Tags:                []string{"repo", "compare", "diff", "dependencies", "drift"},
			UseCases:            []string{"Compare repo tech stacks", "Detect dependency version drift", "Find shared dependencies"},
			Complexity:          tools.ComplexityComplex,
			CircuitBreakerGroup: "repo_analysis",
		},
		// === Fleet Scan (1) ===
		{
			Tool: mcp.NewTool("aftrs_fleet_scan",
				mcp.WithDescription("Scan all repositories in ~/hairglasses-studio/ and return a fleet-wide summary with tag counts, language distribution, and framework matrix."),
				mcp.WithString("filter", mcp.Description("Filter results by tag or language name")),
				mcp.WithBoolean("include_non_go", mcp.Description("Include non-Go repositories in the scan. Default: true")),
			),
			Handler:             handleFleetScan,
			Category:            "repo_analysis",
			Subcategory:         "fleet",
			Tags:                []string{"fleet", "scan", "overview", "repos", "analysis"},
			UseCases:            []string{"Fleet-wide tech inventory", "Tag distribution analysis", "Find repos by technology"},
			Complexity:          tools.ComplexityComplex,
			CircuitBreakerGroup: "repo_analysis",
		},
	}
}

// === Tool 1: aftrs_repo_scan ===

func handleRepoScan(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	repoPath, errResult := tools.RequireStringParam(req, "repo_path")
	if errResult != nil {
		return errResult, nil
	}

	mode := tools.OptionalStringParam(req, "mode", "fast")
	includeLOC := tools.GetBoolParam(req, "include_loc", false)

	repoPath = resolveRepoPath(repoPath)

	if mode != "fast" && mode != "deep" {
		return tools.ErrorResult(fmt.Errorf("invalid mode %q: must be 'fast' or 'deep'", mode)), nil
	}

	var result *ScanResult
	var err error

	if mode == "deep" {
		result, err = scanRepoDeep(repoPath, includeLOC)
	} else {
		result, err = scanRepo(repoPath, includeLOC)
	}

	if err != nil {
		return tools.ErrorResult(fmt.Errorf("scan failed: %w", err)), nil
	}

	return tools.JSONResult(result), nil
}

// scanRepoDeep uses `go mod edit -json` for more accurate dependency parsing.
func scanRepoDeep(repoPath string, includeLOC bool) (*ScanResult, error) {
	// Start with the fast scan as a base.
	result, err := scanRepo(repoPath, includeLOC)
	if err != nil {
		return nil, err
	}

	// Override with go mod edit -json if go.mod exists.
	goModPath := filepath.Join(repoPath, "go.mod")
	if fileExists(goModPath) {
		modInfo, parseErr := parseGoModJSON(repoPath)
		if parseErr == nil {
			result.ModulePath = modInfo.ModulePath
			result.GoVersion = modInfo.GoVersion
			result.DirectDeps = len(modInfo.DirectDeps)
			result.ReplaceCount = modInfo.ReplaceCount

			// Re-apply rules with the (potentially more accurate) deps.
			matches := applyRules(modInfo.DirectDeps)
			result.Frameworks = matches["frameworks"]
			result.Protocols = matches["protocols"]
			result.Datastores = matches["datastores"]
			result.Cloud = matches["cloud"]
			result.AI = matches["ai"]
			result.Observability = matches["observability"]

			// Re-classify with updated deps.
			result.Type = classifyRepo(repoPath, modInfo.DirectDeps, result.Languages)
			result.Relevance = scoreRelevance(modInfo.DirectDeps, result.Languages)
			result.Tags = buildTags(result)
		}
	}

	return result, nil
}

// === Tool 2: aftrs_repo_compare ===

// CompareResult holds the comparison output for multiple repos.
type CompareResult struct {
	Repos          []ScanResult        `json:"repos"`
	VersionDrift   []VersionDriftEntry `json:"version_drift,omitempty"`
	SharedDeps     []string            `json:"shared_deps,omitempty"`
	DepOverlap     float64             `json:"dep_overlap_pct"`
	LanguageMatrix map[string][]string `json:"language_matrix,omitempty"`
}

// VersionDriftEntry represents a dependency at different versions across repos.
type VersionDriftEntry struct {
	Dependency string            `json:"dependency"`
	Versions   map[string]string `json:"versions"` // repo name -> version
}

func handleRepoCompare(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	reposParam, errResult := tools.RequireStringParam(req, "repos")
	if errResult != nil {
		return errResult, nil
	}

	focus := tools.OptionalStringParam(req, "focus", "all")
	if focus != "all" && focus != "languages" && focus != "frameworks" && focus != "deps" {
		return tools.ErrorResult(fmt.Errorf("invalid focus %q: must be 'languages', 'frameworks', 'deps', or 'all'", focus)), nil
	}

	repoNames := strings.Split(reposParam, ",")
	if len(repoNames) < 2 {
		return tools.ErrorResult(fmt.Errorf("at least 2 repos required for comparison, got %d", len(repoNames))), nil
	}

	// Scan each repo.
	scans := make([]ScanResult, 0, len(repoNames))
	for _, name := range repoNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		path := resolveRepoPath(name)
		result, err := scanRepo(path, false)
		if err != nil {
			return tools.ErrorResult(fmt.Errorf("failed to scan %s: %w", name, err)), nil
		}
		scans = append(scans, *result)
	}

	if len(scans) < 2 {
		return tools.ErrorResult(fmt.Errorf("at least 2 valid repos required for comparison")), nil
	}

	compare := CompareResult{
		Repos: scans,
	}

	// Detect version drift (requires reading go.mod for versions).
	if focus == "all" || focus == "deps" {
		compare.VersionDrift = detectVersionDrift(scans)
		compare.SharedDeps, compare.DepOverlap = computeDepOverlap(scans)
	}

	// Build language matrix.
	if focus == "all" || focus == "languages" {
		compare.LanguageMatrix = buildLanguageMatrix(scans)
	}

	return tools.JSONResult(compare), nil
}

// detectVersionDrift finds dependencies used by multiple repos at different versions.
func detectVersionDrift(scans []ScanResult) []VersionDriftEntry {
	type depVersion struct {
		repo    string
		version string
	}

	depVersions := make(map[string][]depVersion)

	for _, scan := range scans {
		goModPath := filepath.Join(scan.Path, "go.mod")
		if !fileExists(goModPath) {
			continue
		}
		versions := parseGoModVersions(goModPath)
		for dep, ver := range versions {
			depVersions[dep] = append(depVersions[dep], depVersion{
				repo:    scan.Name,
				version: ver,
			})
		}
	}

	var drift []VersionDriftEntry
	for dep, versions := range depVersions {
		if len(versions) < 2 {
			continue
		}
		// Check if any versions differ.
		hasDrift := false
		first := versions[0].version
		for _, v := range versions[1:] {
			if v.version != first {
				hasDrift = true
				break
			}
		}
		if !hasDrift {
			continue
		}
		entry := VersionDriftEntry{
			Dependency: dep,
			Versions:   make(map[string]string),
		}
		for _, v := range versions {
			entry.Versions[v.repo] = v.version
		}
		drift = append(drift, entry)
	}

	// Sort by dependency name.
	sort.Slice(drift, func(i, j int) bool {
		return drift[i].Dependency < drift[j].Dependency
	})

	return drift
}

// parseGoModVersions reads a go.mod file and returns a map of dependency -> version
// for direct dependencies only (non-indirect).
func parseGoModVersions(goModPath string) map[string]string {
	f, err := os.Open(goModPath)
	if err != nil {
		return nil
	}
	defer f.Close()

	versions := make(map[string]string)
	scanner := bufio.NewScanner(f)
	inRequire := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "require (" {
			inRequire = true
			continue
		}
		if line == ")" {
			inRequire = false
			continue
		}

		if inRequire {
			parts := strings.Fields(line)
			if len(parts) >= 2 && !strings.HasPrefix(parts[0], "//") {
				versions[parts[0]] = parts[1]
			}
		}

		if strings.HasPrefix(line, "require ") {
			parts := strings.Fields(line)
			if len(parts) >= 3 {
				versions[parts[1]] = parts[2]
			}
		}
	}

	return versions
}

// computeDepOverlap finds shared dependencies and overlap percentage.
func computeDepOverlap(scans []ScanResult) ([]string, float64) {
	if len(scans) < 2 {
		return nil, 0
	}

	// Collect dependency sets per repo.
	depSets := make([]map[string]bool, len(scans))
	allDeps := make(map[string]bool)

	for i, scan := range scans {
		depSets[i] = make(map[string]bool)
		goModPath := filepath.Join(scan.Path, "go.mod")
		if !fileExists(goModPath) {
			continue
		}
		versions := parseGoModVersions(goModPath)
		for dep := range versions {
			depSets[i][dep] = true
			allDeps[dep] = true
		}
	}

	// Find intersection (deps present in ALL repos).
	var shared []string
	for dep := range allDeps {
		inAll := true
		for _, set := range depSets {
			if !set[dep] {
				inAll = false
				break
			}
		}
		if inAll {
			shared = append(shared, dep)
		}
	}
	sort.Strings(shared)

	// Overlap = |intersection| / |union| * 100.
	overlap := 0.0
	if len(allDeps) > 0 {
		overlap = float64(len(shared)) / float64(len(allDeps)) * 100
		overlap = float64(int(overlap*10)) / 10 // Round to 1 decimal
	}

	return shared, overlap
}

// buildLanguageMatrix maps each language to the repos that use it.
func buildLanguageMatrix(scans []ScanResult) map[string][]string {
	matrix := make(map[string][]string)
	for _, scan := range scans {
		for _, lang := range scan.Languages {
			matrix[lang.Name] = append(matrix[lang.Name], scan.Name)
		}
	}
	return matrix
}

// === Tool 3: aftrs_fleet_scan ===

// FleetScanResult holds the aggregated fleet scan output.
type FleetScanResult struct {
	TotalRepos       int                `json:"total_repos"`
	ScannedRepos     int                `json:"scanned_repos"`
	Repos            []ScanResult       `json:"repos"`
	TagCounts        map[string]int     `json:"tag_counts"`
	LanguageDistrib  map[string]int     `json:"language_distribution"`
	TypeDistrib      map[string]int     `json:"type_distribution"`
	RelevanceDistrib map[string]int     `json:"relevance_distribution"`
	FrameworkMatrix  map[string][]string `json:"framework_matrix"`
	Errors           []FleetScanError   `json:"errors,omitempty"`
}

// FleetScanError records a per-repo scan failure.
type FleetScanError struct {
	Repo  string `json:"repo"`
	Error string `json:"error"`
}

func handleFleetScan(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filter := tools.GetStringParam(req, "filter")
	includeNonGo := tools.GetBoolParam(req, "include_non_go", true)

	home, err := os.UserHomeDir()
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot determine home directory: %w", err)), nil
	}
	studioDir := filepath.Join(home, "hairglasses-studio")

	entries, err := os.ReadDir(studioDir)
	if err != nil {
		return tools.ErrorResult(fmt.Errorf("cannot read studio directory: %w", err)), nil
	}

	// Collect directory paths to scan.
	var repoPaths []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		repoPaths = append(repoPaths, filepath.Join(studioDir, name))
	}

	// Scan repos concurrently, limited to 8 goroutines.
	type scanResultEntry struct {
		result *ScanResult
		err    error
		path   string
	}

	results := make([]scanResultEntry, len(repoPaths))
	sem := make(chan struct{}, 8)
	var wg sync.WaitGroup

	for i, path := range repoPaths {
		wg.Add(1)
		go func(idx int, repoPath string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			r, scanErr := scanRepo(repoPath, false)
			results[idx] = scanResultEntry{result: r, err: scanErr, path: repoPath}
		}(i, path)
	}
	wg.Wait()

	// Aggregate results.
	fleet := FleetScanResult{
		TotalRepos:       len(repoPaths),
		TagCounts:        make(map[string]int),
		LanguageDistrib:  make(map[string]int),
		TypeDistrib:      make(map[string]int),
		RelevanceDistrib: make(map[string]int),
		FrameworkMatrix:  make(map[string][]string),
	}

	for _, sr := range results {
		if sr.err != nil {
			fleet.Errors = append(fleet.Errors, FleetScanError{
				Repo:  filepath.Base(sr.path),
				Error: sr.err.Error(),
			})
			continue
		}

		scan := sr.result

		// Filter: skip non-Go repos if requested.
		if !includeNonGo {
			hasGo := false
			for _, lang := range scan.Languages {
				if lang.Name == "Go" {
					hasGo = true
					break
				}
			}
			if !hasGo {
				continue
			}
		}

		// Filter: by tag or language if specified.
		if filter != "" {
			filterLower := strings.ToLower(filter)
			matched := false
			for _, tag := range scan.Tags {
				if strings.EqualFold(tag, filterLower) {
					matched = true
					break
				}
			}
			if !matched {
				for _, lang := range scan.Languages {
					if strings.EqualFold(lang.Name, filterLower) {
						matched = true
						break
					}
				}
			}
			if !matched {
				continue
			}
		}

		fleet.Repos = append(fleet.Repos, *scan)
		fleet.ScannedRepos++

		// Aggregate tags.
		for _, tag := range scan.Tags {
			fleet.TagCounts[tag]++
		}

		// Aggregate languages.
		for _, lang := range scan.Languages {
			fleet.LanguageDistrib[lang.Name] += lang.Files
		}

		// Aggregate type distribution.
		fleet.TypeDistrib[scan.Type]++

		// Aggregate relevance.
		fleet.RelevanceDistrib[scan.Relevance]++

		// Framework matrix: framework -> list of repos using it.
		for _, fw := range scan.Frameworks {
			fleet.FrameworkMatrix[fw.Name] = append(fleet.FrameworkMatrix[fw.Name], scan.Name)
		}
	}

	// Sort repos by name.
	sort.Slice(fleet.Repos, func(i, j int) bool {
		return fleet.Repos[i].Name < fleet.Repos[j].Name
	})

	return tools.JSONResult(fleet), nil
}
