# MCP Mirror Parity

This note is the source of truth for the bundled `mcp/` modules that also publish from standalone repositories.

The machine-readable manifest lives at [`mcp/mirror-parity.json`](../mcp/mirror-parity.json). The verification entrypoint is [`scripts/hg-mcp-mirror-parity.sh`](../scripts/hg-mcp-mirror-parity.sh).

## Mirrored Modules

| Module | Standalone repo | Canonical path | Sync strategy | Purpose |
|--------|------------------|----------------|---------------|---------|
| `dotfiles-mcp` | `hairglasses-studio/dotfiles-mcp` | `mcp/dotfiles-mcp` | `manual_projection` | Desktop and workstation control-plane MCP server |
| `mapitall` | `hairglasses-studio/mapitall` | `mcp/mapitall` | `tree_sync` | Controller and input mapping MCP runtime |
| `mapping` | `hairglasses-studio/mapping` | `mcp/mapping` | `tree_sync` | Shared Go package for mapping and profile resolution |
| `process-mcp` | `hairglasses-studio/process-mcp` | `mcp/process-mcp` | `tree_sync` | Linux process inspection and debugging MCP server |
| `systemd-mcp` | `hairglasses-studio/systemd-mcp` | `mcp/systemd-mcp` | `tree_sync` | Systemd service and timer management MCP server |
| `tmux-mcp` | `hairglasses-studio/tmux-mcp` | `mcp/tmux-mcp` | `tree_sync` | Tmux session and workspace orchestration MCP server |

## Verification

Run the parity checker after touching bundled MCP module READMEs, the manifest, or public docs:

```bash
bash ./scripts/hg-mcp-mirror-parity.sh --check
```

For `dotfiles-mcp`, also refresh the dedicated projection plan:

```bash
bash ./scripts/hg-dotfiles-mcp-projection.sh check
bash ./scripts/hg-dotfiles-mcp-projection.sh check --diff-preview --diff-lines 12
```

When an editable standalone checkout is available, the projection helper can now
apply the carry-forward directly:

```bash
bash ./scripts/hg-dotfiles-mcp-projection.sh apply --standalone /tmp/dotfiles-mcp-real
```

If the manifest mirror path is bare, set `HG_DOTFILES_MCP_APPLY_WORKTREE` to an
editable checkout or create `/tmp/dotfiles-mcp-real`, then the wrapper can use
the repo-specific apply path automatically:

```bash
HG_DOTFILES_MCP_APPLY_WORKTREE=/tmp/dotfiles-mcp-real \
  bash ./scripts/sync-standalone-mcp-repos.sh sync --repos=dotfiles-mcp
```

If a local standalone mirror is managed through a bare repo and its local
`refs/heads/main` may have drifted behind `refs/remotes/origin/main`, run the
bare-mirror hygiene path:

```bash
bash ./scripts/sync-standalone-mcp-repos.sh hygiene --refresh-origin --repos=dotfiles-mcp
bash ./scripts/sync-standalone-mcp-repos.sh hygiene --repos=dotfiles-mcp
bash ./scripts/sync-standalone-mcp-repos.sh hygiene --repair-bare-main --repos=dotfiles-mcp
```

Only mirrors with `sync_strategy: tree_sync` are safe to mutate through the generic
`mcp-mirror.sh` rsync path. Mirrors marked `manual_projection` need a dedicated
repo-local projection workflow. `dotfiles-mcp` is in that category because the
standalone repo has its own root-level package layout, generated surfaces, and
publish metadata that are not tree-isomorphic to `dotfiles/mcp/dotfiles-mcp`.
The dedicated planner/apply helper reports:

- root assets that still move 1:1 into the standalone repo
- bundled root Go files that map into `internal/dotfiles/*.go`
- canonical-only additions that still require projection, plus intentional
  canonical-only differences such as `contract_snapshot_cli.go` and
  `workflow_surface_test.go`
- overlapping files that already drift
- standalone-owned surfaces such as `cmd/*`, `internal/githubstars`, and contract snapshots

The checker validates:

1. The manifest is well-formed and has unique module, repo, and path entries.
2. Every mirrored module has a `go.mod` and `README.md` under the canonical `mcp/` path.
3. Each mirrored README states the canonical dotfiles path, the standalone repo URL, and the publish-mirror parity banner.
4. This document includes a row for each mirrored module.

## CI Hook

The repo smoke workflow is expected to run the mirror parity checker alongside the installer and CLI entrypoint smoke tests. Treat parity drift as a repo-health failure, not a docs-only issue.

For mirrored modules with live host dependencies, the canonical dotfiles workflow is still the source of truth for test partitioning and runner requirements even when standalone README badges point at the publish mirror.
