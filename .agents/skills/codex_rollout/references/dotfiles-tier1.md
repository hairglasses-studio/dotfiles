# Dotfiles Tier-1 Rollout Overlay

Use this overlay when `codex_rollout` applies to `dotfiles` or another public workstation and control-plane repo with bundled MCP modules.

## Expectations

- Tier: `tier-1`
- Lifecycle: `active`
- Language profile: `Go/Shell/Config`
- Canonical instructions live in `AGENTS.md`, with `CLAUDE.md`, `GEMINI.md`, and `.github/copilot-instructions.md` as mirrors.
- Searchable repo-local docs under `docs/` are part of the shipped operator surface, not private scratch notes.

## Required Additions Beyond Baseline

- Searchable architecture and provenance note for installer flow, workstation runtime, and `mcp/` packaging.
- At least one bounded operator resource for repeated publish, parity, or recovery flows.
- Smoke coverage for `install.sh`, `scripts/hg`, workflow wrappers, and other operator entrypoints.
- Mirror parity note plus repo-local verification for any bundled module with a standalone publish mirror.
- Ralph note for every autonomous tranche under `ralphglasses/docs/ralph-roadmap/`.

## Close-Out Checklist

- Verify the changed surface directly, not just generic shell syntax.
- State whether fallout is repo-local, workspace-local, or user-global.
- If local machine validation is blocked, prefer additive docs or tests that can still land remotely and record any checkout drift or missing environment gates.
- Keep new notes focused and searchable instead of dumping long free-form status into one file.

## Dotfiles-Specific Anchors

- Installer and bootstrap: `install.sh`, `dotfiles.toml`
- Operator tooling: `scripts/hg*`, `scripts/lib/hg-*`, `.agents/skills/`
- Public docs: `docs/ARCHITECTURE-PROVENANCE.md`, `docs/MCP-MIRROR-PARITY.md`, `docs/INSTALL-AND-OPERATIONS.md`
- MCP packaging: `mcp/`, `.mcp.json`, `scripts/hg-codex-mcp-sync.sh`, `scripts/hg-mcp-mirror-parity.sh`
