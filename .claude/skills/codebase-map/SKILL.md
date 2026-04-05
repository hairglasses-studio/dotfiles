---
name: codebase-map
description: Analyze all repos in ~/hairglasses-studio — tag with languages/protocols/frameworks, generate dependency graph
allowed-tools: mcp__dotfiles__ops_repo_analyze, mcp__dotfiles__ops_dep_graph, mcp__dotfiles__ops_fleet_iterate, Bash, Read
---

Map the entire hairglasses-studio codebase. `$ARGUMENTS` can be:
- Empty: full analysis (all repos + dependency graph)
- `graph`: dependency graph only (Mermaid output)
- `tags`: repo analysis/tagging only
- A repo name: analyze that single repo

## Workflow

### Full analysis (default)

1. **Dependency graph** — run `ops_dep_graph(filter="internal", format="mermaid")` to get org-wide module dependency map

2. **Repo analysis** — for each repo in ~/hairglasses-studio, run `ops_repo_analyze(repo=<path>)` via Bash loop:
   ```bash
   for dir in ~/hairglasses-studio/*/; do
     # ops_repo_analyze called per repo
   done
   ```
   Or call ops_repo_analyze on 5-10 key repos directly.

3. **Display results:**

```
## Codebase Map: hairglasses-studio

### Dependency Graph
```mermaid
{mermaid output from ops_dep_graph}
```

### Repo Profiles
| Repo | Language | Protocols | Frameworks | Tests | MCP? | CI? | Tags |
|------|----------|-----------|------------|-------|------|-----|------|
| mcpkit | Go | MCP, gRPC | Observability, Testing | 700+ | Yes | Yes | mcp-server, tested, ci |
| dotfiles-mcp | Go | MCP, REST | LLM, CLI | 500+ | Yes | Yes | mcp-server, tested, ci |
| ... | | | | | | | |

### Summary
- {N} repos analyzed
- Languages: Go ({N}), Node ({N}), Python ({N}), Shell ({N})
- Protocols detected: MCP ({N}), REST ({N}), gRPC ({N})
- {N} repos with CI, {N} with tests
```

### Graph only

1. Run `ops_dep_graph(filter="internal", format="mermaid")`
2. Display the Mermaid graph inline
3. List org modules found

### Single repo

1. Run `ops_repo_analyze(repo="~/hairglasses-studio/$ARGUMENTS")`
2. Display full profile with all detected metadata
