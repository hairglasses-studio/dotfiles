#!/usr/bin/env bash
# hg-instruction-check.sh — Verify canonical instruction file alignment across repos
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

ROOT="${1:-$HOME/hairglasses-studio}"
[[ -d "$ROOT" ]] || hg_die "Root directory not found: $ROOT"

pass=0
warn=0
fail=0

for repo in "$ROOT"/*/; do
  [[ -d "$repo/.git" ]] || continue
  name="$(basename "$repo")"

  has_agents=0
  has_claude=0
  has_gemini=0
  has_copilot=0
  agents_canonical=0
  claude_points=0

  [[ -f "$repo/AGENTS.md" ]] && has_agents=1
  [[ -f "$repo/CLAUDE.md" ]] && has_claude=1
  [[ -f "$repo/GEMINI.md" ]] && has_gemini=1
  [[ -f "$repo/.github/copilot-instructions.md" ]] && has_copilot=1

  # Skip repos with neither instruction file
  if [[ "$has_agents" -eq 0 && "$has_claude" -eq 0 ]]; then
    continue
  fi

  # Check canonical marking in AGENTS.md
  if [[ "$has_agents" -eq 1 ]]; then
    if grep -qiE 'Canonical instructions:\s*AGENTS\.md' "$repo/AGENTS.md" 2>/dev/null; then
      agents_canonical=1
    fi
  fi

  # Check CLAUDE.md points to AGENTS.md
  if [[ "$has_claude" -eq 1 ]]; then
    if grep -qi 'canonical instruction file\|canonical instructions:\s*AGENTS' "$repo/CLAUDE.md" 2>/dev/null; then
      claude_points=1
    fi
  fi

  # Evaluate
  if [[ "$agents_canonical" -eq 1 && "$claude_points" -eq 1 ]]; then
    pass=$((pass + 1))
  elif [[ "$agents_canonical" -eq 1 && "$claude_points" -eq 0 && "$has_claude" -eq 1 ]]; then
    hg_warn "$name: AGENTS.md is canonical but CLAUDE.md doesn't point to it"
    warn=$((warn + 1))
  elif [[ "$agents_canonical" -eq 0 && "$has_agents" -eq 1 ]]; then
    hg_warn "$name: AGENTS.md exists but has no canonical marking"
    warn=$((warn + 1))
  else
    pass=$((pass + 1))
  fi

  # Check for missing companion files
  if [[ "$has_agents" -eq 1 && "$has_gemini" -eq 0 ]]; then
    hg_warn "$name: has AGENTS.md but missing GEMINI.md"
    warn=$((warn + 1))
  fi
  if [[ "$has_agents" -eq 1 && "$has_copilot" -eq 0 ]]; then
    hg_warn "$name: has AGENTS.md but missing .github/copilot-instructions.md"
    warn=$((warn + 1))
  fi

  # Check staleness: AGENTS.md newer than mirrors
  if [[ "$agents_canonical" -eq 1 && "$has_claude" -eq 1 ]]; then
    if [[ "$repo/AGENTS.md" -nt "$repo/CLAUDE.md" ]]; then
      hg_warn "$name: AGENTS.md is newer than CLAUDE.md (mirror may be stale)"
      warn=$((warn + 1))
    fi
  fi
done

echo ""
hg_info "Instruction check: $pass pass, $warn warnings, $fail failures"

if [[ "$fail" -gt 0 ]]; then
  exit 1
fi
