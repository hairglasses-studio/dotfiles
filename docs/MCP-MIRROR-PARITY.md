# MCP Mirror Parity

This note documents which bundled `dotfiles/mcp/*` modules are canonical source trees and which standalone repositories are publish mirrors of those trees.

## Canonical Rule

For the mirror-managed modules listed below, the code under `dotfiles/mcp/*` is the canonical source of truth. The standalone repositories exist as publish mirrors and should not become the place where canonical implementation changes start.

## Mirror Map

The workspace manifest declares these mirror relationships.

For standalone CI and repo-local smoke coverage, this repo also vendors the same mirror set in [mcp/mirror-parity.json](../mcp/mirror-parity.json). Use [scripts/hg-mcp-mirror-parity.sh](../scripts/hg-mcp-mirror-parity.sh) to validate that the manifest, this note, and the bundled README mirror banners stay aligned.

| Canonical bundled path | Standalone mirror repo | Visibility | Workflow policy |
| --- | --- | --- | --- |
| `dotfiles/mcp/dotfiles-mcp` | `hairglasses-studio/dotfiles-mcp` | public | `repo_owned` |
| `dotfiles/mcp/systemd-mcp` | `hairglasses-studio/systemd-mcp` | public | `repo_owned` |
| `dotfiles/mcp/tmux-mcp` | `hairglasses-studio/tmux-mcp` | public | `repo_owned` |
| `dotfiles/mcp/process-mcp` | `hairglasses-studio/process-mcp` | public | `repo_owned` |

## Bundled Modules That Are Not Mirror-Managed

These bundled modules are part of the workspace but are not declared as standalone publish mirrors in `workspace/manifest.json`.

- `dotfiles/mcp/hg-mcp`
- `dotfiles/mcp/mapitall`
- `dotfiles/mcp/mapping`

`mcp/go.work` also treats `mapping` as a local replacement dependency:

```text
go 1.26.1

use (
    ./dotfiles-mcp
    ./hg-mcp
    ./mapitall
    ./process-mcp
    ./systemd-mcp
    ./tmux-mcp
)

replace github.com/hairglasses-studio/mapping => ./mapping
```

## Sync Mechanism

Two scripts define how mirror parity works.

- `scripts/sync-standalone-mcp-repos.sh`: selects manifest entries with `lifecycle=mirror` and a non-empty `mirror_of`, then runs the canonical sync or check flow.
- `scripts/mcp-mirror.sh`: performs the actual `rsync`-based sync or diff while preserving mirror-only repository metadata.

## Mirror-Only Metadata Preserved During Sync

The mirror sync intentionally excludes repository-local metadata and release scaffolding so standalone repos can keep publish-specific packaging.

Excluded paths and artifacts include:

- `.git/`
- `.github/`
- `.claude/`
- `.codex/`
- `.gemini/`
- `.ralph/`
- `.well-known/`
- coverage artifacts, test binaries, logs
- module-specific release files such as `.goreleaser.yaml`

Implication:

- Code and module content should originate in the canonical bundled tree.
- Release wiring, hosted workflows, and publish metadata can stay repo-local in the standalone mirror.

## Verification Commands

Start with the manifest-backed mirror check rather than comparing directories manually.

### Repo-local parity check

```bash
bash ./scripts/hg-mcp-mirror-parity.sh --check
```

### Check all declared mirrors

```bash
bash ./scripts/sync-standalone-mcp-repos.sh check
```

### Check a specific mirror

```bash
bash ./scripts/sync-standalone-mcp-repos.sh check --repos=dotfiles-mcp
bash ./scripts/sync-standalone-mcp-repos.sh check --repos=systemd-mcp
bash ./scripts/sync-standalone-mcp-repos.sh check --repos=tmux-mcp
bash ./scripts/sync-standalone-mcp-repos.sh check --repos=process-mcp
```

### Sync canonical changes into standalone mirrors

```bash
bash ./scripts/sync-standalone-mcp-repos.sh sync
```

### Bootstrap a canonical tree from an existing mirror

```bash
bash ./scripts/sync-standalone-mcp-repos.sh bootstrap --repos=<mirror-repo>
```

## Dirty-Tree Safety

`scripts/sync-standalone-mcp-repos.sh` refuses to sync dirty canonical paths or dirty mirror repos unless `--allow-dirty` is set.

That guard exists to prevent accidental parity writes from stomping on in-progress local work.

The repo-local checker is read-only. It validates metadata and bundled docs without touching the standalone mirror repos.

## Change Policy

When you touch one of the mirror-managed MCP surfaces:

1. Change the bundled canonical tree in `dotfiles/mcp/*`.
2. Verify the module locally in the dotfiles workspace.
3. Run the mirror parity check or sync flow.
4. Treat the standalone repo as the publish target, not the implementation origin.

## Why this note exists

The repo bundles multiple MCP modules while also feeding standalone public repositories. Without an explicit parity note, future automation has to rediscover which tree is canonical and which files are intentionally repo-local in the mirrors.
