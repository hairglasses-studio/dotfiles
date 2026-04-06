package repo_analysis

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// extensionMap maps file extensions to language names.
var extensionMap = map[string]string{
	".go":    "Go",
	".py":    "Python",
	".ts":    "TypeScript",
	".tsx":   "TypeScript",
	".js":    "JavaScript",
	".jsx":   "JavaScript",
	".rs":    "Rust",
	".md":    "Markdown",
	".yaml":  "YAML",
	".yml":   "YAML",
	".json":  "JSON",
	".toml":  "TOML",
	".proto": "Protobuf",
	".sh":    "Shell",
	".bash":  "Shell",
	".zsh":   "Shell",
	".css":   "CSS",
	".scss":  "CSS",
	".html":  "HTML",
	".htm":   "HTML",
	".sql":   "SQL",
	".c":     "C",
	".h":     "C",
	".cpp":   "C++",
	".java":  "Java",
	".rb":    "Ruby",
	".lua":   "Lua",
	".nix":   "Nix",
	".tf":    "Terraform",
	".svelte": "Svelte",
	".vue":   "Vue",
}

// LanguageInfo holds file count and percentage for a language.
type LanguageInfo struct {
	Name       string  `json:"name"`
	Percentage float64 `json:"percentage"`
	Files      int     `json:"files"`
}

// ScanResult is the structured output of a repo scan.
type ScanResult struct {
	Name          string       `json:"name"`
	Path          string       `json:"path"`
	ModulePath    string       `json:"module_path,omitempty"`
	GoVersion     string       `json:"go_version,omitempty"`
	Languages     []LanguageInfo `json:"languages"`
	Frameworks    []TagMatch   `json:"frameworks"`
	Protocols     []TagMatch   `json:"protocols"`
	Datastores    []TagMatch   `json:"datastores"`
	Cloud         []TagMatch   `json:"cloud"`
	AI            []TagMatch   `json:"ai"`
	Observability []TagMatch   `json:"observability"`
	BuildSystems  []string     `json:"build_systems"`
	DirectDeps    int          `json:"direct_deps"`
	ReplaceCount  int          `json:"replace_count"`
	HasGoWork     bool         `json:"has_go_work"`
	Type          string       `json:"type"`
	Relevance     string       `json:"relevance"`
	Tags          []string     `json:"tags"`
	ScannedAt     string       `json:"scanned_at"`
	LOC           *LOCInfo     `json:"loc,omitempty"`
}

// LOCInfo holds line-of-code counts by language.
type LOCInfo struct {
	Total      int            `json:"total"`
	ByLanguage map[string]int `json:"by_language"`
}

// GoModInfo holds parsed go.mod metadata.
type GoModInfo struct {
	ModulePath   string
	GoVersion    string
	DirectDeps   []string
	ReplaceCount int
}

// scanRepo performs a scan of a single repository and returns structured metadata.
func scanRepo(repoPath string, includeLOC bool) (*ScanResult, error) {
	repoPath, err := filepath.Abs(repoPath)
	if err != nil {
		return nil, fmt.Errorf("resolving path: %w", err)
	}

	info, err := os.Stat(repoPath)
	if err != nil {
		return nil, fmt.Errorf("path does not exist: %s", repoPath)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: %s", repoPath)
	}

	name := filepath.Base(repoPath)
	result := &ScanResult{
		Name:      name,
		Path:      repoPath,
		ScannedAt: time.Now().UTC().Format(time.RFC3339),
	}

	// Detect languages by walking the file tree.
	langs, totalFiles := detectLanguages(repoPath)
	result.Languages = buildLanguageInfo(langs, totalFiles)

	// Parse go.mod if it exists.
	goModPath := filepath.Join(repoPath, "go.mod")
	var depPaths []string
	if fileExists(goModPath) {
		modInfo, parseErr := parseGoMod(goModPath)
		if parseErr == nil {
			result.ModulePath = modInfo.ModulePath
			result.GoVersion = modInfo.GoVersion
			result.DirectDeps = len(modInfo.DirectDeps)
			result.ReplaceCount = modInfo.ReplaceCount
			depPaths = modInfo.DirectDeps
		}
	}

	// Check for go.work.
	result.HasGoWork = fileExists(filepath.Join(repoPath, "go.work"))

	// Apply import-to-tag rules.
	matches := applyRules(depPaths)
	result.Frameworks = matches["frameworks"]
	result.Protocols = matches["protocols"]
	result.Datastores = matches["datastores"]
	result.Cloud = matches["cloud"]
	result.AI = matches["ai"]
	result.Observability = matches["observability"]

	// Detect build systems.
	result.BuildSystems = detectBuildSystems(repoPath)

	// Classify repo type.
	result.Type = classifyRepo(repoPath, depPaths, result.Languages)

	// Score relevance.
	result.Relevance = scoreRelevance(depPaths, result.Languages)

	// Collect tags.
	result.Tags = buildTags(result)

	// LOC counting (optional).
	if includeLOC {
		result.LOC = countLOC(repoPath, langs)
	}

	return result, nil
}

// resolveRepoPath converts a name or absolute path to an absolute path.
// If the input contains no "/" it is treated as a repo name under ~/hairglasses-studio/.
func resolveRepoPath(input string) string {
	if strings.Contains(input, "/") {
		return input
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("/home", os.Getenv("USER"), "hairglasses-studio", input)
	}
	return filepath.Join(home, "hairglasses-studio", input)
}

// detectLanguages walks the repo directory and counts files per language.
// It skips hidden directories, vendor/, node_modules/, and common build output.
func detectLanguages(repoPath string) (map[string]int, int) {
	counts := make(map[string]int)
	total := 0

	_ = filepath.WalkDir(repoPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if d.IsDir() {
			name := d.Name()
			// Skip hidden, vendor, node_modules, and build directories.
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" ||
				name == "dist" || name == "__pycache__" || name == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(d.Name()))
		if lang, ok := extensionMap[ext]; ok {
			counts[lang]++
			total++
		}

		return nil
	})

	return counts, total
}

// buildLanguageInfo converts raw counts to sorted LanguageInfo with percentages.
func buildLanguageInfo(counts map[string]int, total int) []LanguageInfo {
	if total == 0 {
		return nil
	}

	langs := make([]LanguageInfo, 0, len(counts))
	for name, count := range counts {
		pct := float64(count) / float64(total) * 100
		langs = append(langs, LanguageInfo{
			Name:       name,
			Percentage: float64(int(pct*10)) / 10, // Round to 1 decimal
			Files:      count,
		})
	}

	// Sort by file count descending.
	sort.Slice(langs, func(i, j int) bool {
		return langs[i].Files > langs[j].Files
	})

	return langs
}

// parseGoMod parses a go.mod file to extract module path, Go version,
// direct dependencies, and replace directive count.
func parseGoMod(goModPath string) (*GoModInfo, error) {
	f, err := os.Open(goModPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	info := &GoModInfo{}
	scanner := bufio.NewScanner(f)
	inRequire := false
	inReplace := false

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip comments and empty lines.
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Module path.
		if strings.HasPrefix(line, "module ") {
			info.ModulePath = strings.TrimPrefix(line, "module ")
			continue
		}

		// Go version.
		if strings.HasPrefix(line, "go ") && info.GoVersion == "" {
			info.GoVersion = strings.TrimPrefix(line, "go ")
			continue
		}

		// Require block.
		if line == "require (" {
			inRequire = true
			inReplace = false
			continue
		}

		// Replace block.
		if line == "replace (" {
			inReplace = true
			inRequire = false
			continue
		}

		// End of block.
		if line == ")" {
			inRequire = false
			inReplace = false
			continue
		}

		// Inside require block: extract dependency path.
		if inRequire {
			parts := strings.Fields(line)
			if len(parts) >= 1 && !strings.HasPrefix(parts[0], "//") {
				info.DirectDeps = append(info.DirectDeps, parts[0])
			}
			continue
		}

		// Inside replace block: count replacements.
		if inReplace {
			if !strings.HasPrefix(line, "//") {
				info.ReplaceCount++
			}
			continue
		}

		// Inline require.
		if strings.HasPrefix(line, "require ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				info.DirectDeps = append(info.DirectDeps, parts[1])
			}
			continue
		}

		// Inline replace.
		if strings.HasPrefix(line, "replace ") {
			info.ReplaceCount++
			continue
		}
	}

	return info, scanner.Err()
}

// detectBuildSystems checks for the presence of common build system files.
func detectBuildSystems(repoPath string) []string {
	var systems []string

	checks := []struct {
		file string
		name string
	}{
		{"Makefile", "make"},
		{"pipeline.mk", "pipeline.mk"},
		{".github/workflows", "github-actions"},
		{".gitlab-ci.yml", "gitlab-ci"},
		{"Dockerfile", "docker"},
		{"docker-compose.yml", "docker-compose"},
		{"docker-compose.yaml", "docker-compose"},
		{"Taskfile.yml", "taskfile"},
		{"Taskfile.yaml", "taskfile"},
		{"justfile", "just"},
		{"package.json", "npm"},
		{"pyproject.toml", "pyproject"},
		{"Cargo.toml", "cargo"},
	}

	seen := make(map[string]bool)
	for _, check := range checks {
		if fileExists(filepath.Join(repoPath, check.file)) && !seen[check.name] {
			seen[check.name] = true
			systems = append(systems, check.name)
		}
	}

	return systems
}

// scoreRelevance assigns a relevance score based on dependencies and languages.
func scoreRelevance(depPaths []string, langs []LanguageInfo) string {
	for _, dep := range depPaths {
		if strings.HasPrefix(dep, "github.com/hairglasses-studio/mcpkit") {
			return "high"
		}
	}
	for _, lang := range langs {
		if lang.Name == "Go" {
			return "medium"
		}
	}
	return "low"
}

// buildTags builds a deduplicated tag list from scan results.
func buildTags(result *ScanResult) []string {
	seen := make(map[string]bool)
	var tags []string

	addTag := func(tag string) {
		if !seen[tag] {
			seen[tag] = true
			tags = append(tags, tag)
		}
	}

	// Add primary language tag.
	if len(result.Languages) > 0 {
		addTag(strings.ToLower(result.Languages[0].Name))
	}

	// Add tags from rule matches.
	allMatches := map[string][]TagMatch{
		"frameworks":    result.Frameworks,
		"protocols":     result.Protocols,
		"datastores":    result.Datastores,
		"cloud":         result.Cloud,
		"ai":            result.AI,
		"observability": result.Observability,
	}
	for _, matches := range allMatches {
		for _, m := range matches {
			addTag(m.Name)
		}
	}

	// Add repo type tag.
	if result.Type != "" && result.Type != "unknown" {
		addTag(result.Type)
	}

	sort.Strings(tags)
	return tags
}

// countLOC counts lines of code per language in the repository.
func countLOC(repoPath string, langCounts map[string]int) *LOCInfo {
	loc := &LOCInfo{
		ByLanguage: make(map[string]int),
	}

	_ = filepath.WalkDir(repoPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if strings.HasPrefix(name, ".") || name == "vendor" || name == "node_modules" ||
				name == "dist" || name == "__pycache__" || name == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		ext := strings.ToLower(filepath.Ext(d.Name()))
		lang, ok := extensionMap[ext]
		if !ok {
			return nil
		}

		lines := countFileLines(path)
		loc.ByLanguage[lang] += lines
		loc.Total += lines

		return nil
	})

	return loc
}

// countFileLines counts the number of non-empty lines in a file.
func countFileLines(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		count++
	}
	return count
}

// goModJSON holds the parsed output of `go mod edit -json`.
type goModJSON struct {
	Module struct {
		Path string `json:"Path"`
	} `json:"Module"`
	Go      string `json:"Go"`
	Require []struct {
		Path     string `json:"Path"`
		Version  string `json:"Version"`
		Indirect bool   `json:"Indirect"`
	} `json:"Require"`
	Replace []struct {
		Old struct {
			Path string `json:"Path"`
		} `json:"Old"`
		New struct {
			Path string `json:"Path"`
		} `json:"New"`
	} `json:"Replace"`
}

// parseGoModJSON uses `go mod edit -json` for more accurate parsing.
// Falls back to manual parsing if the command fails.
func parseGoModJSON(repoPath string) (*GoModInfo, error) {
	cmd := exec.Command("go", "mod", "edit", "-json")
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(), "GOWORK=off")

	out, err := cmd.Output()
	if err != nil {
		// Fall back to manual parsing.
		return parseGoMod(filepath.Join(repoPath, "go.mod"))
	}

	var mod goModJSON
	if err := json.Unmarshal(out, &mod); err != nil {
		return parseGoMod(filepath.Join(repoPath, "go.mod"))
	}

	info := &GoModInfo{
		ModulePath: mod.Module.Path,
		GoVersion:  mod.Go,
	}

	for _, req := range mod.Require {
		if !req.Indirect {
			info.DirectDeps = append(info.DirectDeps, req.Path)
		}
	}
	info.ReplaceCount = len(mod.Replace)

	return info, nil
}
