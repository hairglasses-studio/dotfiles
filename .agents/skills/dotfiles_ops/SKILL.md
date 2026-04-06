---
name: dotfiles_ops
description: Workstation operations, repo-tooling, onboarding, and fleet-maintenance workflow for the hairglasses dotfiles repo.
allowed-tools:
  - Bash
  - Read
  - Write
  - Grep
  - Glob
  - mcp__dotfiles__dotfiles_list_configs
  - mcp__dotfiles__dotfiles_validate_config
  - mcp__dotfiles__dotfiles_reload_service
  - mcp__dotfiles__system_health_check
  - mcp__dotfiles__systemd_failed
---

# Dotfiles Ops

Use this skill for workstation health, repo scaffolding, workflow sync, CI triage, and general configuration operations in the dotfiles repo.

## Default loop

1. Prefer the shared scripts in `scripts/` and `scripts/lib/` over ad hoc replacements.
2. Validate the narrowest affected surface first, then run broader repo-wide checks when the change is shared.
3. Keep provider-neutral repo tooling provider-neutral; do not add Claude-only defaults where a shared Codex/Gemini path exists.
4. When a change affects fleet tooling, state whether it is repo-local, workspace-local, or user-global.

Read `references/workflows.md` for the compressed workflow catalog that replaces the previous ops-oriented Claude skill set.
