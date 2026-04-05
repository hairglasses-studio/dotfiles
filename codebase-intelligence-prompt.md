/effort max

# Codebase Intelligence: Repo Tagging, Dependency Maps & Tool Research

You are building codebase intelligence infrastructure for the hairglasses-studio org (14 repos, primarily Go). This session produces 3 deliverables: a repo tagging analysis, a cross-repo dependency map, and a Go UML/architecture diagram — plus a researched tool comparison for generating these programmatically via MCP tools.

## Background

The ~/hairglasses-studio org has 14 repos (see ~/hairglasses-studio/CLAUDE.md for the full tier list). Primary language is Go 1.26.1 with mcpkit as the shared framework. We recently built a community dotfiles link index at dotfiles/links/ using a 10-round scraping pattern with YAML frontmatter entries, _registry.json dedup, and per-round git checkpoints. We want to apply similar structured-indexing principles to map our OWN codebase.

## Phase 1: Repo Analysis & Tagging (Agents 1-2, parallel)

**Agent 1 — Repo Scanner**: For each repo in ~/hairglasses-studio/*, read its go.mod (or package.json/pyproject.toml), CLAUDE.md, and README.md. Produce a tag profile:

    ~/hairglasses-studio/docs/inventory/repo-tags.json

Schema per repo:
    {
      "repo": "mcpkit",
      "languages": ["go"],
      "go_version": "1.26.1",
      "frameworks": ["mcpkit"],
      "protocols": ["mcp", "stdio", "json-rpc"],
      "key_packages": ["github.com/hairglasses-studio/mcpkit"],
      "external_deps": ["github.com/mark3labs/mcp-go"],
      "has_mcp_server": true,
      "tool_count": 35,
      "test_count": 700,
      "tier": 1,
      "tags": ["framework", "public", "go"]
    }

**Agent 2 — Dependency Mapper**: For each Go repo, parse go.mod to extract:
- Direct dependencies (require blocks)
- Internal cross-repo dependencies (github.com/hairglasses-studio/*)
- Shared external dependencies across repos

Output: `~/hairglasses-studio/docs/inventory/dep-graph.json` with nodes (repos) and edges (dependencies).

## Phase 2: Diagram Generation Research (Agents 3-4, parallel)

**Agent 3 — Dependency Visualization Tools**: Research and compare Go-ecosystem tools for generating cross-repo dependency maps. Search GitHub for each, check stars/activity/Go compatibility:

| Category | Candidates to Evaluate |
|----------|----------------------|
| Go dep graph | `loov/goda`, `adonovan/spaghetti`, `kisielk/godepgraph` |
| General dep viz | `graphviz/dot`, `mermaid-js/mermaid-cli`, `d2lang/d2` |
| Monorepo analysis | `ossf/scorecard`, `google/osv-scanner` |

For each tool, report: GitHub stars, last commit, Go support, output formats, CLI usage, whether it can handle a go.work workspace. Recommend the best fit for our 14-repo Go workspace.

**Agent 4 — Architecture Diagram Tools**: Research UML and modern alternatives for Go codebases:

| Category | Candidates to Evaluate |
|----------|----------------------|
| Go UML/arch | `jfeliu007/goplantuml`, `ofabry/go-callvis`, `rsc/grind` |
| Modern diagram-as-code | `d2lang/d2`, `mermaid-js/mermaid-cli`, `structurizr/cli` |
| Code intelligence | `sourcegraph/scip-go`, `go-architect` |

For each: stars, last commit, output formats (SVG/PNG/Mermaid/PlantUML), whether it handles interfaces/packages/call graphs, CLI invocation example. Recommend the best fit for mapping our Go module architecture.

## Phase 3: Generate Maps

Using the winning tools from Phase 2 research:

1. **Install the chosen tools** (via `go install` or downloading binaries)
2. **Generate dependency map**: Run against ~/hairglasses-studio repos, output to `~/hairglasses-studio/docs/inventory/dep-map.svg` (or .png/.mermaid)
3. **Generate architecture diagram**: Run against the Go repos, output to `~/hairglasses-studio/docs/inventory/architecture-diagram.svg`
4. If the tools require configuration, create config files in `~/hairglasses-studio/docs/inventory/`

## Phase 4: MCP Tool / Skill Design

Based on what worked in Phases 1-3, draft designs for:

1. **`repo_tag` MCP tool** — Analyzes a single repo and returns its tag profile JSON. Input: repo path. Output: the schema from Phase 1.
2. **`codebase_map` MCP tool** — Generates a dependency graph across repos. Input: workspace root. Output: graph JSON + optional diagram file.
3. **`/codebase-index` Claude skill** — Orchestrates both tools across the org, produces the full inventory.

Write the designs as markdown specs in `~/hairglasses-studio/docs/inventory/tool-designs.md` with: tool name, input schema, output schema, implementation approach, which Go packages to use.

## Checkpoints

After each phase, commit:
- Phase 1: `git -C ~/hairglasses-studio/docs add inventory/ && git -C ~/hairglasses-studio/docs commit -m "inventory: repo tags + dep graph"`
- Phase 2: `git -C ~/hairglasses-studio/docs add inventory/ && git -C ~/hairglasses-studio/docs commit -m "inventory: tool comparison research"`
- Phase 3: `git -C ~/hairglasses-studio/docs add inventory/ && git -C ~/hairglasses-studio/docs commit -m "inventory: generated dep map + architecture diagram"`
- Phase 4: `git -C ~/hairglasses-studio/docs add inventory/ && git -C ~/hairglasses-studio/docs commit -m "inventory: MCP tool + skill designs"`

## Completion

Print summary:
- Repos tagged (count + any that failed)
- Cross-repo dependency count
- Tool comparison winner + rationale
- Diagrams generated (paths + formats)
- Skill/tool designs written

## Important

- All output goes to ~/hairglasses-studio/docs/inventory/ (create if needed)
- Git: hairglasses / mitch@hairglasses.studio, no GPG
- Prefer Go tools that work with go.work workspaces
- If a tool candidate is archived or unmaintained (no commits in 2+ years), note it but deprioritize
- For diagram tools, SVG output is preferred (scalable), with Mermaid as fallback (renders in GitHub markdown)
