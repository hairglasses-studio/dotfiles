#!/usr/bin/env bash
# hg-agent-docs.sh — Generate compatibility agent docs from a canonical AGENTS.md
# or a legacy CLAUDE.md source.
# Usage: hg-agent-docs.sh [--source auto|agents|claude] [REPO_DIR]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

SOURCE_MODE="auto"
REPO_DIR="."

usage() {
  cat <<'EOF'
Usage: hg-agent-docs.sh [--source auto|agents|claude] [REPO_DIR]

Generates AGENTS.md, CLAUDE.md, GEMINI.md, and .github/copilot-instructions.md
from the chosen canonical source.

- auto   : Prefer AGENTS.md when it is marked canonical; otherwise use CLAUDE.md.
- agents : Treat AGENTS.md as canonical and generate thin compatibility mirrors.
- claude : Treat CLAUDE.md as canonical and generate AGENTS.md/GEMINI.md mirrors.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --source)
      [[ $# -ge 2 ]] || hg_die "--source requires a value"
      SOURCE_MODE="$2"
      shift 2
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      REPO_DIR="$1"
      shift
      ;;
  esac
done

case "$SOURCE_MODE" in
  auto|agents|claude) ;;
  *)
    hg_die "Unsupported --source value: $SOURCE_MODE"
    ;;
esac

cd "$REPO_DIR"
REPO_NAME="$(basename "$PWD")"

extract_section() {
  local file="$1"
  shift
  local patterns=("$@")
  for pat in "${patterns[@]}"; do
    local result
    result=$(awk -v pat="$pat" '
      BEGIN { found=0 }
      /^## / {
        if (found) exit
        if (index($0, pat) > 0) { found=1; print; next }
      }
      found { print }
    ' "$file")
    if [[ -n "$result" ]]; then
      echo "$result"
      return
    fi
  done
}

extract_summary() {
  local file="$1"
  awk '
    NR==1 { next }
    /^$/ && !started { next }
    /^$/ && started { exit }
    /^#/ { exit }
    { started=1; print }
  ' "$file"
}

normalize_title() {
  local title="$1"
  title="${title#Claude Code Context for }"
  title="${title#Claude Code Context: }"
  title="${title#Claude Code Instructions for }"
  title="${title#Claude Code Instructions: }"
  title="${title#Agent Instructions for }"
  title="${title% — Claude Code Instructions}"
  title="${title% - Claude Code Instructions}"
  title="${title% — Agent Instructions}"
  title="${title% - Agent Instructions}"
  printf '%s\n' "$title"
}

is_agents_canonical() {
  [[ -f AGENTS.md ]] || return 1
  grep -qiE 'AGENTS\.md is canonical|Canonical instructions:\s*AGENTS\.md|Canonical source:\s*AGENTS\.md' AGENTS.md
}

detect_source_mode() {
  case "$SOURCE_MODE" in
    agents)
      [[ -f AGENTS.md ]] || hg_die "$REPO_NAME: AGENTS.md is required for --source agents"
      printf '%s\n' "agents"
      ;;
    claude)
      [[ -f CLAUDE.md ]] || hg_die "$REPO_NAME: CLAUDE.md is required for --source claude"
      printf '%s\n' "claude"
      ;;
    auto)
      if is_agents_canonical; then
        printf '%s\n' "agents"
      elif [[ -f CLAUDE.md ]]; then
        printf '%s\n' "claude"
      elif [[ -f AGENTS.md ]]; then
        printf '%s\n' "agents"
      else
        hg_die "$REPO_NAME: neither AGENTS.md nor CLAUDE.md exists"
      fi
      ;;
  esac
}

write_copilot_instructions() {
  mkdir -p .github
  cat > .github/copilot-instructions.md << 'EOF'
# Copilot Instructions

See AGENTS.md in the repository root for canonical project context, build commands, and architecture.
EOF
  hg_ok "Generated .github/copilot-instructions.md"
}

generate_from_claude() {
  local title summary commands arch patterns

  title=$(head -1 CLAUDE.md | sed 's/^# //')
  title="$(normalize_title "$title")"
  [[ -n "$title" ]] || title="$REPO_NAME"

  summary=$(extract_summary CLAUDE.md)
  commands=$(extract_section CLAUDE.md "Build & Test" "Build" "Commands")
  arch=$(extract_section CLAUDE.md "Architecture" "Package Map" "Project Structure")
  patterns=$(extract_section CLAUDE.md "Key Patterns" "Key Conventions" "Conventions" "Coding Conventions")

  {
    echo "# $title — Agent Instructions"
    echo ""
    echo "$summary"
    echo ""
    echo "> Canonical source: CLAUDE.md"
    echo ""

    if [[ -n "$commands" ]]; then
      echo "$commands" | sed '1s/## .*/## Build \& Test/'
      echo ""
    fi

    if [[ -n "$arch" ]]; then
      echo "$arch" | sed '1s/## .*/## Architecture/'
      echo ""
    fi

    if [[ -n "$patterns" ]]; then
      echo "$patterns" | sed '1s/## .*/## Key Conventions/'
      echo ""
    fi
  } > AGENTS.md

  hg_ok "Generated AGENTS.md ($(wc -l < AGENTS.md) lines)"

  {
    echo "# $title — Gemini CLI Instructions"
    echo ""
    echo "$summary" | head -1
    echo ""

    if [[ -n "$commands" ]]; then
      echo "## Build & Test"
      echo ""
      echo "$commands" | awk '/^```/{p=!p; print; next} p{print}'
      echo ""
    fi

    if [[ -n "$arch" ]]; then
      echo "## Architecture"
      echo ""
      echo "$arch" | grep -E '^\s*[-*0-9]' | head -10 || true
      echo ""
    fi

    if [[ -n "$patterns" ]]; then
      echo "## Key Conventions"
      echo ""
      echo "$patterns" | grep -E '^\s*[-*]' | head -8 || true
      echo ""
    fi
  } > GEMINI.md

  hg_ok "Generated GEMINI.md ($(wc -l < GEMINI.md) lines)"
  write_copilot_instructions
}

generate_from_agents() {
  local title summary

  title=$(head -1 AGENTS.md | sed 's/^# //')
  title="$(normalize_title "$title")"
  [[ -n "$title" ]] || title="$REPO_NAME"
  summary=$(extract_summary AGENTS.md)

  cat > CLAUDE.md <<EOF
# $title — Claude Code Instructions

This repo uses [AGENTS.md](AGENTS.md) as the canonical instruction file. Read it before making changes.

## Claude Notes

- Use [AGENTS.md](AGENTS.md) for build, test, architecture, and repo-specific conventions.
- Keep [CLAUDE.md](CLAUDE.md), [GEMINI.md](GEMINI.md), and \`.github/copilot-instructions.md\` as thin compatibility mirrors.
- Add Claude-specific memory or workflow notes here only when they cannot live in [AGENTS.md](AGENTS.md).

## Summary

$summary
EOF
  hg_ok "Generated CLAUDE.md ($(wc -l < CLAUDE.md) lines)"

  cat > GEMINI.md <<EOF
# $title — Gemini CLI Instructions

This repo uses [AGENTS.md](AGENTS.md) as the canonical instruction file.

- Read [AGENTS.md](AGENTS.md) first for build, test, architecture, and repo conventions.
- Treat [GEMINI.md](GEMINI.md) as a compatibility mirror, not the primary source of truth.

## Summary

$summary
EOF
  hg_ok "Generated GEMINI.md ($(wc -l < GEMINI.md) lines)"
  write_copilot_instructions
}

MODE="$(detect_source_mode)"
hg_info "Generating agent docs for $REPO_NAME from $MODE"

case "$MODE" in
  claude) generate_from_claude ;;
  agents) generate_from_agents ;;
esac
