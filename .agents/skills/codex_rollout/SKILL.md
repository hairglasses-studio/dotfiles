---
name: codex_rollout
description: Run the hairglasses Codex migration loop across repos with baseline, skill-surface, workflow, and audit discipline.
allowed-tools:
  - Bash
  - Read
  - Write
  - Grep
  - Glob
---

# Codex Rollout

Use this skill when migrating a repo or a small repo set from Claude-first state to the current Codex baseline.

## Default loop

1. Classify the repo first: operator, active Tier B, or exempt.
2. Establish the shared baseline before deeper enhancements: `AGENTS.md`, compatibility docs, `.codex/config.toml`, Copilot instructions, and live Codex workflows.
3. If the repo has a skill surface, move it to `.agents/skills/` and generate compatibility outputs with `codexkit skills sync <repo_path>`. In dotfiles-managed repos, `scripts/hg-skill-surface-sync.sh <repo_path>` is the thin compatibility wrapper around that command.
4. Prefer shared scripts such as `hg-workflow-sync.sh`, `hg-codex-audit.sh`, and `hg-codex-mcp-sync.sh` over one-off edits when the change is org-wide.
5. End with verification and an audit refresh so the remaining work is explicit and evidence-backed.

Read `references/workflows.md` for the rollout checklist and blocker handling rules.
