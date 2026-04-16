#!/usr/bin/env bash
# claude-post-compact.sh — PostCompact hook: re-inject AGENTS.md essentials after
# context compaction so the agent retains orientation in the dotfiles repo.
#
# Claude Code calls this after compacting context. stdout is injected as
# additionalContext. Keep output under 100 lines — the whole point is brevity.
#
# Install: .claude/settings.json hooks.PostCompact

set -euo pipefail

DOTFILES_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

# ── Repo identity ────────────────────────────────────────────────────────────
repo_name="dotfiles"
mcp_server="dotfiles-mcp"

# ── Theme name (live, falls back to static) ──────────────────────────────────
theme_name="Hairglasses Neon"
palette_file="$DOTFILES_ROOT/theme/palette.env"
if [[ -f "$palette_file" ]]; then
  raw="$(grep -m1 '^THEME_NAME=' "$palette_file" 2>/dev/null | cut -d= -f2- | tr -d '"' || true)"
  [[ -n "$raw" ]] && theme_name="$raw"
fi

# ── Git context (branch + recent commit) ────────────────────────────────────
git_line=""
if git -C "$DOTFILES_ROOT" rev-parse --is-inside-work-tree &>/dev/null 2>&1; then
  branch="$(git -C "$DOTFILES_ROOT" branch --show-current 2>/dev/null || echo "detached")"
  last="$(git -C "$DOTFILES_ROOT" log --oneline -1 2>/dev/null || true)"
  git_line="Branch: $branch  |  Last commit: $last"
fi

cat <<ANCHOR
╔══ CONTEXT RE-ANCHOR (post-compaction) ══════════════════════════════════════╗
║ Repo: $repo_name  •  MCP: $mcp_server  •  Theme: $theme_name
ANCHOR

[[ -n "$git_line" ]] && printf '║ %s\n' "$git_line"

cat <<'ANCHOR'
╚═════════════════════════════════════════════════════════════════════════════╝

## Repo: dotfiles (hairglasses-studio/dotfiles)
Purpose: Desktop automation, workstation config, MCP server scaffolding,
         and org-wide repo tooling for hairglasses-studio (64 repos).

## Build & Verify
  bash -n scripts/*.sh scripts/lib/*.sh          # syntax-check all scripts
  bash ./scripts/hg-agent-home-sync.sh --check   # provider home parity
  bash ./scripts/hg-workflow-sync.sh --dry-run   # workflow propagation
  bash ./scripts/hg-agent-docs.sh --source auto . # doc generation
  bash ./scripts/hg-codex-audit.sh --write-workspace-cache --write-wiki-docs --write-json

## Key Variables (hyprland.conf)
  $mod = SUPER   (primary modifier key — never change to something else)

## Active Theme
  Palette: Hairglasses Neon  (theme/palette.env)
  Primary: #29f0ff  Secondary: #ff47d1  Tertiary: #3dffb5
  UI font: Maple Mono NF CN  Code font: Monaspace Neon

## Architecture Surfaces
  installer / bootstrap  →  install.sh, dotfiles.toml, scripts/
  workstation runtime    →  hyprland/, ironbar/, kitty/, systemd/, zsh/
  MCP / automation       →  mcp/, scripts/lib/, .agents/skills/

## Canonical File Hierarchy
  AGENTS.md            — canonical instructions (source of truth)
  CLAUDE.md / GEMINI.md — thin compatibility mirrors, do not edit directly
  .agents/skills/      — workflow skill surface (source of truth)
  .claude/skills/      — generated mirror via codexkit skills sync

## Working Rules
  - Edit scripts/ and scripts/lib/ instead of ad-hoc shell fragments.
  - Agent entrypoints: scripts/hg-codex-launch.sh / hg-claude-launch.sh / hg-gemini-launch.sh
  - Hook changes → document repo-local vs user-scope boundary.
  - Keyboard layout stays US QWERTY in all WM configs.
  - Clipboard: wl-copy / wl-paste  (Wayland session).
  - Terminal: kitty  (--gtk-single-instance=false in compositor keybinds).

## MCP Server (dotfiles-mcp)
  Run:  ./scripts/run-dotfiles-mcp.sh
  409 tools — use dotfiles_tool_search / dotfiles_tool_catalog for discovery.

## Verification Ladder
  1. bash -n scripts/*.sh scripts/lib/*.sh
  2. bash ./scripts/hg-agent-docs.sh --source auto .
  3. bash ./scripts/hg-workflow-sync.sh --dry-run
  4. bash ./scripts/hg-codex-audit.sh --write-workspace-cache --write-wiki-docs --write-json
ANCHOR
