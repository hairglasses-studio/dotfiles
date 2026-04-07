// Package dep_graph provides cross-repo Go dependency graph generation and drift detection.
package dep_graph

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// orgPatterns are the GitHub org prefixes used to identify internal modules.
var orgPatterns = []string{
	"github.com/hairglasses-studio/",
	"github.com/glasshairs/",
	"github.com/hairglasses-studio/",
}

// Edge represents a dependency edge between two Go modules.
type Edge struct {
	From        string `json:"from"`
	FromVersion string `json:"from_version"`
	To          string `json:"to"`
	ToVersion   string `json:"to_version"`
	SourceRepo  string `json:"source_repo"`
}

// Replace represents a replace directive in a go.mod file.
type Replace struct {
	Repo    string `json:"repo"`
	Old     string `json:"old"`
	New     string `json:"new"`
	IsLocal bool   `json:"is_local"`
}

// GraphOutput is the full output structure for the dep_graph tool.
type GraphOutput struct {
	Generated    string              `json:"generated"`
	Format       string              `json:"format"`
	Edges        []Edge              `json:"edges"`
	Replaces     []Replace           `json:"replaces,omitempty"`
	GoWorkspaces map[string][]string `json:"go_workspaces,omitempty"`
	RepoCount    int                 `json:"repo_count"`
	EdgeCount    int                 `json:"edge_count"`
	Rendered     string              `json:"rendered,omitempty"`
}

// DriftEntry represents a module with version drift across consumers.
type DriftEntry struct {
	Module      string            `json:"module"`
	Versions    map[string]string `json:"versions"`
	Severity    string            `json:"severity"`
	Latest      string            `json:"latest"`
	BehindCount int               `json:"behind_count"`
}

// DriftReport is the full output for the dep_drift tool.
type DriftReport struct {
	TotalModulesChecked int          `json:"total_modules_checked"`
	ModulesWithDrift    int          `json:"modules_with_drift"`
	DriftEntries        []DriftEntry `json:"drift_entries"`
}

// isOrgModule checks whether a module path belongs to one of the org patterns.
func isOrgModule(modPath string) bool {
	for _, p := range orgPatterns {
		if strings.HasPrefix(modPath, p) {
			return true
		}
	}
	return false
}

// studioRoot returns the default studio root directory.
func studioRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}
	return filepath.Join(home, "hairglasses-studio"), nil
}

// findGoModFiles walks studioRoot up to maxDepth levels looking for go.mod files.
func findGoModFiles(root string, maxDepth int) ([]string, error) {
	var mods []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable dirs
		}
		// Calculate depth relative to root.
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		depth := 0
		if rel != "." {
			depth = strings.Count(rel, string(os.PathSeparator)) + 1
		}
		// Skip hidden directories and common non-Go dirs.
		if d.IsDir() {
			base := d.Name()
			if strings.HasPrefix(base, ".") || base == "vendor" || base == "node_modules" || base == "testdata" {
				return filepath.SkipDir
			}
			if depth > maxDepth {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == "go.mod" {
			mods = append(mods, path)
		}
		return nil
	})
	return mods, err
}

// repoNameFromGoMod extracts the top-level repo directory name from a go.mod path.
func repoNameFromGoMod(goModPath, root string) string {
	rel, err := filepath.Rel(root, filepath.Dir(goModPath))
	if err != nil {
		return filepath.Base(filepath.Dir(goModPath))
	}
	// Take the first path component as the repo name.
	parts := strings.SplitN(rel, string(os.PathSeparator), 2)
	return parts[0]
}

// parseModGraphLine parses a single line of "go mod graph" output.
// Format: "from@version to@version" (version may be absent for the main module).
func parseModGraphLine(line string) (fromMod, fromVer, toMod, toVer string, ok bool) {
	parts := strings.Fields(line)
	if len(parts) != 2 {
		return "", "", "", "", false
	}
	fromMod, fromVer = splitModVersion(parts[0])
	toMod, toVer = splitModVersion(parts[1])
	return fromMod, fromVer, toMod, toVer, true
}

// splitModVersion splits "module@version" into module path and version.
func splitModVersion(s string) (string, string) {
	if idx := strings.LastIndex(s, "@"); idx > 0 {
		return s[:idx], s[idx+1:]
	}
	return s, ""
}

// CollectEdges runs "go mod graph" in each module directory and returns dependency edges.
func CollectEdges(root string) ([]Edge, error) {
	goModFiles, err := findGoModFiles(root, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to find go.mod files: %w", err)
	}

	seen := make(map[string]bool)
	var edges []Edge

	for _, goMod := range goModFiles {
		dir := filepath.Dir(goMod)
		repoName := repoNameFromGoMod(goMod, root)

		cmd := exec.Command("go", "mod", "graph")
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GOWORK=off")

		out, err := cmd.Output()
		if err != nil {
			// Skip repos that fail (e.g., missing deps).
			continue
		}

		scanner := bufio.NewScanner(strings.NewReader(string(out)))
		for scanner.Scan() {
			line := scanner.Text()
			fromMod, fromVer, toMod, toVer, ok := parseModGraphLine(line)
			if !ok {
				continue
			}

			key := fromMod + "@" + fromVer + "->" + toMod + "@" + toVer + ":" + repoName
			if seen[key] {
				continue
			}
			seen[key] = true

			edges = append(edges, Edge{
				From:        fromMod,
				FromVersion: fromVer,
				To:          toMod,
				ToVersion:   toVer,
				SourceRepo:  repoName,
			})
		}
	}

	return edges, nil
}

// goModEditJSON is the subset of "go mod edit -json" output we care about.
type goModEditJSON struct {
	Module struct {
		Path string `json:"Path"`
	} `json:"Module"`
	Replace []struct {
		Old struct {
			Path    string `json:"Path"`
			Version string `json:"Version"`
		} `json:"Old"`
		New struct {
			Path    string `json:"Path"`
			Version string `json:"Version"`
		} `json:"New"`
	} `json:"Replace"`
}

// CollectReplaces runs "go mod edit -json" in each module directory to find replace directives.
func CollectReplaces(root string) ([]Replace, error) {
	goModFiles, err := findGoModFiles(root, 3)
	if err != nil {
		return nil, fmt.Errorf("failed to find go.mod files: %w", err)
	}

	var replaces []Replace

	for _, goMod := range goModFiles {
		dir := filepath.Dir(goMod)
		repoName := repoNameFromGoMod(goMod, root)

		cmd := exec.Command("go", "mod", "edit", "-json")
		cmd.Dir = dir
		cmd.Env = append(os.Environ(), "GOWORK=off")

		out, err := cmd.Output()
		if err != nil {
			continue
		}

		var modJSON goModEditJSON
		if err := json.Unmarshal(out, &modJSON); err != nil {
			continue
		}

		for _, r := range modJSON.Replace {
			isLocal := !strings.Contains(r.New.Path, ".") || strings.HasPrefix(r.New.Path, ".") || strings.HasPrefix(r.New.Path, "/")
			replaces = append(replaces, Replace{
				Repo:    repoName,
				Old:     r.Old.Path,
				New:     r.New.Path,
				IsLocal: isLocal,
			})
		}
	}

	return replaces, nil
}

// CollectGoWorkModules finds all go.work files and extracts their use directives.
func CollectGoWorkModules(root string) (map[string][]string, error) {
	workspaces := make(map[string][]string)

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			base := d.Name()
			if strings.HasPrefix(base, ".") || base == "vendor" || base == "node_modules" {
				return filepath.SkipDir
			}
			// Max depth 2 for go.work files.
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return nil
			}
			if rel != "." && strings.Count(rel, string(os.PathSeparator)) >= 2 {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Name() == "go.work" {
			modules, parseErr := parseGoWorkUse(path)
			if parseErr != nil {
				return nil
			}
			rel, relErr := filepath.Rel(root, path)
			if relErr != nil {
				rel = path
			}
			workspaces[rel] = modules
		}
		return nil
	})

	return workspaces, err
}

// parseGoWorkUse reads a go.work file and extracts module directories from "use" directives.
func parseGoWorkUse(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var modules []string
	inUseBlock := false
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Single-line: use ./path
		if strings.HasPrefix(line, "use ") && !strings.HasPrefix(line, "use (") {
			mod := strings.TrimSpace(strings.TrimPrefix(line, "use"))
			if mod != "" {
				modules = append(modules, mod)
			}
			continue
		}

		// Block start: use (
		if strings.HasPrefix(line, "use (") || line == "use (" {
			inUseBlock = true
			continue
		}

		// Block end
		if inUseBlock && line == ")" {
			inUseBlock = false
			continue
		}

		// Inside use block
		if inUseBlock && line != "" && !strings.HasPrefix(line, "//") {
			modules = append(modules, line)
		}
	}

	return modules, scanner.Err()
}

// FilterEdgesInternal returns only edges where the "To" module matches an org pattern.
func FilterEdgesInternal(edges []Edge) []Edge {
	var filtered []Edge
	for _, e := range edges {
		if isOrgModule(e.To) {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// FilterEdgesByRepo returns only edges involving a specific repo name.
func FilterEdgesByRepo(edges []Edge, repo string) []Edge {
	var filtered []Edge
	for _, e := range edges {
		if e.SourceRepo == repo {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// CountUniqueRepos returns the number of unique source repos in the edge set.
func CountUniqueRepos(edges []Edge) int {
	repos := make(map[string]bool)
	for _, e := range edges {
		repos[e.SourceRepo] = true
	}
	return len(repos)
}
