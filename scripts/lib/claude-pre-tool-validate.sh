#!/usr/bin/env bash
# claude-pre-tool-validate.sh — PreToolUse hook for Claude Code
# Validates config file syntax before Write/Edit tools apply changes.
# Reads JSON from stdin, extracts file_path, runs format-specific checks.
# Return 0 to allow, non-zero to block the edit.

set -euo pipefail

# Read tool input from stdin (Claude pipes JSON)
input="$(cat)"

# Extract file_path from JSON (lightweight — no jq dependency)
file_path="$(echo "$input" | grep -oP '"file_path"\s*:\s*"[^"]*"' | head -1 | sed 's/.*"file_path"\s*:\s*"//;s/"$//')"
[[ -n "$file_path" ]] || exit 0

case "$file_path" in
  *.yuck)
    # Basic lisp paren matching — count open vs close parens
    if [[ -f "$file_path" ]]; then
      open=$(grep -o '(' "$file_path" 2>/dev/null | wc -l)
      close=$(grep -o ')' "$file_path" 2>/dev/null | wc -l)
      if (( open != close )); then
        echo "[validate] Unbalanced parentheses in $file_path: open=$open close=$close" >&2
        exit 1
      fi
    fi
    ;;
  *.scss)
    # Check sassc availability and run syntax check
    if command -v sassc &>/dev/null; then
      if [[ -f "$file_path" ]]; then
        if ! sassc --style compressed "$file_path" /dev/null 2>/dev/null; then
          echo "[validate] SCSS syntax error in $file_path" >&2
          exit 1
        fi
      fi
    fi
    ;;
esac

exit 0
