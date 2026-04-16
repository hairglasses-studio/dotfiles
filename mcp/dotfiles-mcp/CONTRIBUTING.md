# Contributing

This project supports development with **Claude Code**, **Gemini CLI**, and **OpenAI Codex CLI**. Any provider can lead development.

## Development Setup

### 1. Clone and build

```bash
git clone https://github.com/hairglasses-studio/dotfiles-mcp
cd dotfiles-mcp
GOWORK=off go build ./...
GOWORK=off go vet ./...
GOWORK=off go test ./... -count=1
```

### 2. Verify

```bash
make pipeline-check   # build + vet + test (via shared pipeline)
```

Or use the pipeline script directly:

```bash
~/hairglasses-studio/dotfiles/scripts/hg-pipeline.sh
```

For publish-surface work, also run the contract and parity checks:

```bash
make contract-snapshot   # regenerate .well-known/mcp.json + snapshots/contract/*
make contract-check      # verify committed artifacts match the live registry
make contract-diff       # summarize public surface deltas against a base ref
make publish-check       # vet + test + contract-check + release-parity
```

## Architecture

dotfiles-mcp is a single-binary MCP server built from many module files, with the exact public surface committed under `snapshots/contract/` and published through `.well-known/mcp.json`. Current counts should always come from the generated bundle rather than hand-maintained prose.

Core files:

- `main.go` -- Server setup, tool registration
- `discovery.go` -- Discovery-first search, schema, stats, and health entrypoints
- `contract_snapshot.go` -- Public contract bundle generation for the canonical module
- `mod_*.go` -- Feature modules (30+ total; run `ls mod_*.go` for the full list). Key examples: `mod_hyprland.go` (compositor tools), `mod_shader.go` (Kitty visual pipeline), `mod_input.go` (input device management)
- `oss.go` -- Open-source readiness scoring

All tools are built on [mcpkit](https://github.com/hairglasses-studio/mcpkit) using `handler.TypedHandler` generics and `registry.ToolDefinition`.

## Making Changes

1. Create a branch: `git checkout -b feat/my-change`
2. Make your changes
3. Run the pipeline: `GOWORK=off go build ./... && GOWORK=off go vet ./... && GOWORK=off go test ./... -count=1`
4. If you changed tools, resources, prompts, README/ROADMAP contract prose, or `.well-known/mcp.json`, run `make publish-check`
5. Regenerate and commit `snapshots/contract/*` plus `.well-known/mcp.json` when the public contract changes
6. Commit with a descriptive message
7. Push and open a PR

## Code Style

- **Go**: `gofmt` formatting, `go vet` clean
- Error handling: `handler.CodedErrorResult(handler.ErrInvalidParam, err)` -- never naked panics
- Thread safety: `sync.RWMutex` with `RLock` for reads, `Lock` for writes
- Param extraction: `handler.GetStringParam`, `handler.GetIntParam`, `handler.GetBoolParam`

Editor settings are in `.editorconfig` -- most editors pick this up automatically.

## Pre-commit Hooks

Install with:

```bash
make install-hooks
```

This runs vet + fast tests before each commit.

## CI

The canonical module tracks both code-health and publish-surface checks. Any change that affects the exposed contract should keep these artifacts in sync:

- `.well-known/mcp.json`
- `snapshots/contract/overview.json`
- `snapshots/contract/tools.json`
- `snapshots/contract/resources.json`
- `snapshots/contract/templates.json`
- `snapshots/contract/prompts.json`

`make publish-check` is the fastest local approximation of the release gate for this embedded module.

## Questions?

Open an issue or tag `@hairglasses` in your PR.
