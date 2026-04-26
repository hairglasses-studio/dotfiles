# hairglasses-studio Shared Infrastructure

Centralized development tooling and governance for all 44 repos in the org. Two hubs manage shared components — edits here propagate to all repos automatically (symlinks) or via sync scripts (copies).

## Architecture

```
dotfiles/                          .github/ (org repo)
├── make/                          ├── CONTRIBUTING.md
│   ├── pipeline.mk               ├── CODEOWNERS
│   ├── golangci.yml               ├── PULL_REQUEST_TEMPLATE.md
│   ├── go-version                 ├── CODE_OF_CONDUCT.md
│   ├── ci-go.yml                  ├── SECURITY.md
│   └── ralphrc-base               ├── FUNDING.yml
├── git/                           ├── ISSUE_TEMPLATE/
│   ├── gitignore-{base,go,node,python}  │   ├── bug_report.yml
├── editorconfig                   │   ├── feature_request.yml
├── scripts/                       │   └── ai-task.yml
│   ├── hg-pipeline.sh             ├── workflow-templates/
│   ├── hg-new-repo.sh             │   ├── ci-go.yml
│   ├── hg-agent-docs.sh           │   ├── claude-review.yml
│   ├── hg-go-sync.sh              │   └── claude-security.yml
│   ├── hg-gitignore.sh            └── profile/
│   ├── hg-install-hooks.sh            └── README.md
│   └── lib/
│       ├── hg-core.sh
│       ├── config.sh
│       └── compositor.sh
```

**dotfiles/** — Local dev tooling consumed via symlinks, Makefile includes, and script calls. Push here and all repos pick up changes immediately (symlinked files) or after running a sync script (copied files).

**.github/** — GitHub platform governance. Provides org-wide defaults for CONTRIBUTING, CODEOWNERS, issue/PR templates, and workflow templates. Individual repos inherit automatically — no local copies needed. Per-repo overrides still work (local files take priority).

## How Repos Consume Shared Components

| Method | Example | Auto-updates? |
|--------|---------|---------------|
| **Symlink** | `.editorconfig -> ../../dotfiles/editorconfig` | Yes (instant) |
| **Symlink** | `.golangci.yml -> ../../dotfiles/make/golangci.yml` | Yes (instant) |
| **Makefile include** | `-include $(HOME)/.../dotfiles/make/pipeline.mk` | Yes (instant) |
| **Script call** | `make install-hooks` (calls hg-install-hooks.sh) | Yes (instant) |
| **GitHub inheritance** | `.github/` org repo → all repos | Yes (instant) |
| **Copy** | CI workflows (ci.yml, claude-review.yml, codex-review.yml, etc.) | No (use hg-workflow-sync.sh) |
| **Copy** | dependabot.yml | No (manual or hg-new-repo.sh) |

## Shared Components

### Build Pipeline (`make/`)

| File | Purpose | Consumed via |
|------|---------|-------------|
| `pipeline.mk` | Targets: build, test, vet, lint, check, install-hooks, pipeline-info | `-include` in Makefile |
| `golangci.yml` | Go linter config (errcheck, govet, staticcheck, gocritic, misspell) | Symlink as `.golangci.yml` |
| `go-version` | Pinned Go version (currently 1.26.1) | `hg-go-sync.sh` reads this |
| `ci-go.yml` | Standalone CI workflow template (lint → test → build) | Copied to `.github/workflows/ci.yml` |
| `ralphrc-base` | Shared Ralph agent defaults (timeouts, quality gates, tools) | Source from per-repo `.ralphrc` |

### Scripts (`scripts/hg-*.sh`)

| Script | Purpose | Usage |
|--------|---------|-------|
| `hg-pipeline.sh` | Universal build+test for Go/Node/Python | `make pipeline-check` or `/pipeline` skill |
| `hg-new-repo.sh` | Scaffold new repo with standard governance, workflows, and agent docs | `hg-new-repo.sh <name> [go\|node\|python]` |
| `hg-agent-docs.sh` | Generate compatibility docs from canonical AGENTS.md or legacy CLAUDE.md | Run after editing the repo's canonical instruction file |
| `hg-go-sync.sh` | Sync Go version across all repos | `hg-go-sync.sh [--dry-run] [--tidy]` |
| `hg-gitignore.sh` | Assemble .gitignore from templates | `hg-gitignore.sh go > .gitignore` |
| `hg-install-hooks.sh` | Install language-appropriate pre-commit hooks | `make install-hooks` |

### Libraries (`scripts/lib/`)

| Library | Purpose |
|---------|---------|
| `hg-core.sh` | Snazzy palette colors, logging (hg_info/ok/warn/error/die), paths |
| `config.sh` | Atomic config writes (mktemp+mv), backups, service reload |
| `compositor.sh` | Hyprland/AeroSpace IPC abstraction |

See `scripts/lib/README.md` for usage examples.

### Git Templates (`git/`)

| File | Purpose |
|------|---------|
| `gitignore-base` | Universal patterns (secrets, OS, editors, AI agents) |
| `gitignore-go` | Go binaries, coverage, vendor, .ralph/ |
| `gitignore-node` | node_modules, dist, build |
| `gitignore-python` | __pycache__, .venv, .uv, .coverage |

### Editor Config

`editorconfig` — Symlinked as `.editorconfig`. Go=tabs, Python=4-space, YAML/JSON/Shell=2-space, Makefiles=tabs.

## Adding a New Repo

```bash
~/hairglasses-studio/dotfiles/scripts/hg-new-repo.sh my-new-project go
```

Creates: go.mod, Makefile (with pipeline.mk include), .editorconfig (symlink), .golangci.yml (symlink), .gitignore, LICENSE, CONTRIBUTING.md, CI workflows, dependabot.yml, `.codex/config.toml`, `CLAUDE.md`, `AGENTS.md`, `GEMINI.md`, `.github/copilot-instructions.md`, and pre-commit hooks. One command, fully standard.

## Updating Shared Components

| Change | Action |
|--------|--------|
| Edit linter rules | Edit `dotfiles/make/golangci.yml` → push dotfiles → all repos pick up via symlink |
| Bump Go version | Edit `dotfiles/make/go-version` → run `hg-go-sync.sh --tidy` → commit each repo |
| Update CI template | Edit `dotfiles/make/ci-go.yml` → manually update repos or run future `hg-workflow-sync.sh` |
| Update governance | Edit `.github/` org repo → all repos inherit automatically |
| Add build target | Edit `dotfiles/make/pipeline.mk` → push dotfiles → all repos pick up via include |

## Claude Code Integration

- **`/pipeline [check|ship|loop]`** — 6-step pipeline skill: build → test → reconnect MCP → verify → push → loop/propose
- **`/go-check`** — Quick health check: build + vet + test + lint in parallel
- **PostToolUse hook** — Auto-reloads Hyprland/Quickshell/swaync when config files are written

## Provider Notes

- Repo scaffolding now emits provider-neutral instruction files alongside `CLAUDE.md`.
- Workflow sync treats `claude-*` and `codex-*` review templates as canonical artifacts.
- Local hook automation in `dotfiles/.claude/` is still deeper than Codex parity; track those gaps in `docs/codex-migration/`.
