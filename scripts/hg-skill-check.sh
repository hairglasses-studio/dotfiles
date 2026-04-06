#!/usr/bin/env bash
# hg-skill-check.sh — Verify skill surface consistency across repos
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

ROOT="${1:-$HOME/hairglasses-studio}"
[[ -d "$ROOT" ]] || hg_die "Root directory not found: $ROOT"

SYNC_SCRIPT="$SCRIPT_DIR/hg-skill-surface-sync.sh"

pass=0
warn=0
fail=0

# Check repo-level skill surfaces
for repo in "$ROOT"/*/; do
  [[ -d "$repo/.git" ]] || continue
  name="$(basename "$repo")"

  has_surface=0
  has_legacy=0
  has_mcp_json=0
  has_mcp_policy=0

  [[ -f "$repo/.agents/skills/surface.yaml" ]] && has_surface=1
  [[ -d "$repo/.claude/commands" ]] && has_legacy=1
  [[ -f "$repo/.mcp.json" ]] && has_mcp_json=1
  [[ -f "$repo/.codex/mcp-profile-policy.json" ]] && has_mcp_policy=1

  # Check surface sync consistency
  if [[ "$has_surface" -eq 1 && -x "$SYNC_SCRIPT" ]]; then
    if "$SYNC_SCRIPT" "$repo" --check >/dev/null 2>&1; then
      pass=$((pass + 1))
    else
      hg_warn "$name: skill surface out of sync (run hg-skill-surface-sync.sh)"
      warn=$((warn + 1))
    fi
  fi

  # Check for orphan legacy commands alongside surface.yaml
  if [[ "$has_surface" -eq 1 && "$has_legacy" -eq 1 ]]; then
    legacy_count=$(find "$repo/.claude/commands" -name '*.md' 2>/dev/null | wc -l)
    if [[ "$legacy_count" -gt 0 ]]; then
      hg_warn "$name: has $legacy_count orphan .claude/commands/ alongside surface.yaml"
      warn=$((warn + 1))
    fi
  fi

  # Check MCP policy completeness (skip example-only .mcp.json)
  if [[ "$has_mcp_json" -eq 1 && "$has_mcp_policy" -eq 0 ]]; then
    # Check if .mcp.json has real servers (not just _example_ prefixed)
    real_servers=$(jq -r '.mcpServers // {} | keys[]' "$repo/.mcp.json" 2>/dev/null | grep -cv '^_' || true)
    if [[ "$real_servers" -gt 0 ]]; then
      hg_warn "$name: has .mcp.json ($real_servers servers) but no .codex/mcp-profile-policy.json"
      warn=$((warn + 1))
    fi
  fi
done

# Check global user skill parity
claude_count=$(find "$HOME/.claude/commands" -name '*.md' 2>/dev/null | wc -l)
agents_count=$(find "$HOME/.agents/skills" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l)

if [[ "$claude_count" -gt 0 ]]; then
  hg_info "Global user skills: $claude_count Claude / $agents_count Codex-compatible"
fi

# Check global skill sync
if [[ -x "$SCRIPT_DIR/hg-global-skill-sync.sh" ]]; then
  if "$SCRIPT_DIR/hg-global-skill-sync.sh" --check >/dev/null 2>&1; then
    pass=$((pass + 1))
    hg_ok "Global skill sync: up to date"
  else
    hg_warn "Global skill sync: out of date (run hg-global-skill-sync.sh)"
    warn=$((warn + 1))
  fi
fi

echo ""
hg_info "Skill check: $pass pass, $warn warnings, $fail failures"

if [[ "$fail" -gt 0 ]]; then
  exit 1
fi
