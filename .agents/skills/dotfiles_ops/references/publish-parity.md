# Publish And Parity Workflow Reference

Use this reference when a change touches public docs, bundled MCP modules, skill and agent surfaces, or repo metadata that feeds standalone publish mirrors.

## Change Classes

- Workstation-only: runtime config under `hyprland/`, `eww/`, `kitty/`, `ghostty/`, or `systemd/`; no public mirror action unless the operator contract changed.
- Public operator surface: `install.sh`, `scripts/hg*`, `AGENTS.md`, compatibility docs, workflows, and public-facing docs; keep examples and verification current.
- Mirror-managed MCP: `mcp/dotfiles-mcp`, `mcp/process-mcp`, `mcp/systemd-mcp`, and `mcp/tmux-mcp`; verify repo-local parity docs before touching standalone mirrors.
- Bundled-only MCP: `mcp/hg-mcp`, `mcp/mapitall`, and `mcp/mapping`; do not invent standalone mirror obligations.

## Default Loop

1. Place the change with `docs/ARCHITECTURE-PROVENANCE.md` so installer, workstation, and `mcp/` ownership stay explicit.
2. If the operator or public surface changed, align `docs/INSTALL-AND-OPERATIONS.md` or another searchable repo-local note with the real command path.
3. If a mirror-managed MCP module changed, run `bash ./scripts/hg-mcp-mirror-parity.sh --check` first. Only use `bash ./scripts/sync-standalone-mcp-repos.sh check` or a live sync for mirrors whose `sync_strategy` is `tree_sync`; capture a deliberate projection plan for mirrors marked `manual_projection`.
4. If agent or skill instructions changed, treat `.agents/skills/` and `AGENTS.md` as canonical, then regenerate compatibility surfaces with `bash ./scripts/hg-skill-surface-sync.sh` and `bash ./scripts/hg-agent-docs.sh --source auto .`.
5. End with the narrowest syntax or smoke check that proves the changed entrypoint or public contract.

## Evidence To Capture

- Surface classification: workstation-only, public operator, mirror-managed MCP, or bundled-only MCP.
- Exact verification commands used.
- Whether any follow-up lives in a standalone mirror repo, workspace-local automation, or user-global config.

## Avoid

- Do not hand-edit `.claude/skills/` or treat compatibility docs as canonical.
- Do not update standalone mirror repos before the bundled path, docs, and parity checks agree.
- Do not bury repo-specific operator rules in free-form model prompts when they can live here.
