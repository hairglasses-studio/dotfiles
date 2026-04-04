#!/usr/bin/env bash
# hg-agent-docs.sh — Generate AGENTS.md, GEMINI.md, and copilot-instructions.md from CLAUDE.md.
# Usage: hg-agent-docs.sh [REPO_DIR]
# Reads CLAUDE.md from the repo root and generates the three derivative files.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

REPO_DIR="${1:-.}"
cd "$REPO_DIR"
REPO_NAME="$(basename "$PWD")"

if [[ ! -f CLAUDE.md ]]; then
  hg_warn "$REPO_NAME: no CLAUDE.md found, skipping"
  exit 0
fi

hg_info "Generating agent docs for $REPO_NAME"

# ── Extract sections from CLAUDE.md ──────────
# Extracts content between ## headers. Returns everything from the matching
# header to the next ## header (exclusive).
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

# Get the first line (title)
TITLE=$(head -1 CLAUDE.md | sed 's/^# //')

normalize_title() {
  local title="$1"
  title="${title% — Claude Code Instructions}"
  title="${title% - Claude Code Instructions}"
  printf '%s\n' "$title"
}

TITLE="$(normalize_title "$TITLE")"
if [[ -z "$TITLE" ]]; then
  TITLE="$REPO_NAME"
fi

# Get the first paragraph after the title (summary)
SUMMARY=$(awk '
  NR==1 { next }
  /^$/ && !started { next }
  /^$/ && started { exit }
  /^#/ { exit }
  { started=1; print }
' CLAUDE.md)

# Extract key sections — try multiple header names
COMMANDS=$(extract_section CLAUDE.md "Build & Test" "Build" "Commands")
ARCH=$(extract_section CLAUDE.md "Architecture" "Package Map" "Project Structure")
PATTERNS=$(extract_section CLAUDE.md "Key Patterns" "Key Conventions" "Conventions" "Coding Conventions")

# ── Generate AGENTS.md ───────────────────────
{
  echo "# $TITLE — Agent Instructions"
  echo ""
  echo "$SUMMARY"
  echo ""

  if [[ -n "$COMMANDS" ]]; then
    # Rename header to standardized form
    echo "$COMMANDS" | sed '1s/## .*/## Build \& Test/'
    echo ""
  fi

  if [[ -n "$ARCH" ]]; then
    echo "$ARCH" | sed '1s/## .*/## Architecture/'
    echo ""
  fi

  if [[ -n "$PATTERNS" ]]; then
    echo "$PATTERNS" | sed '1s/## .*/## Key Conventions/'
    echo ""
  fi
} > AGENTS.md

hg_ok "Generated AGENTS.md ($(wc -l < AGENTS.md) lines)"

# ── Generate GEMINI.md ───────────────────────
{
  echo "# $TITLE — Gemini CLI Instructions"
  echo ""
  # Compress summary to first sentence
  echo "$SUMMARY" | head -1
  echo ""

  if [[ -n "$COMMANDS" ]]; then
    echo "## Build & Test"
    echo ""
    # Keep only code blocks from commands section
    echo "$COMMANDS" | awk '/^```/{p=!p; print; next} p{print}'
    echo ""
  fi

  if [[ -n "$ARCH" ]]; then
    echo "## Architecture"
    echo ""
    # Keep only bullet points and numbered lists
    echo "$ARCH" | grep -E '^\s*[-*0-9]' | head -10 || true
    echo ""
  fi

  if [[ -n "$PATTERNS" ]]; then
    echo "## Key Conventions"
    echo ""
    # Keep only bullet points, max 8
    echo "$PATTERNS" | grep -E '^\s*[-*]' | head -8 || true
    echo ""
  fi
} > GEMINI.md

hg_ok "Generated GEMINI.md ($(wc -l < GEMINI.md) lines)"

# ── Generate copilot-instructions.md ─────────
mkdir -p .github
cat > .github/copilot-instructions.md << 'COPEOF'
# Copilot Instructions

See AGENTS.md in the repository root for full project context, build commands, and architecture.
COPEOF

hg_ok "Generated .github/copilot-instructions.md"
