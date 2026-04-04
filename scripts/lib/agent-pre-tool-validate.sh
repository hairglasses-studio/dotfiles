#!/usr/bin/env bash
# agent-pre-tool-validate.sh — Pre-edit validation hook for agent-driven file edits.
# Reads JSON from stdin, extracts a candidate file path, and blocks obviously
# broken writes for the formats we can cheaply validate.

set -euo pipefail

extract_file_path() {
  local input="$1"
  local match

  match="$(
    printf '%s' "$input" |
      tr '\n' ' ' |
      grep -oE '"(file_path|path)"[[:space:]]*:[[:space:]]*"[^"]*"' |
      head -1 || true
  )"

  [[ -n "$match" ]] || return 1
  printf '%s\n' "$match" | sed -E 's/^"(file_path|path)"[[:space:]]*:[[:space:]]*"//; s/"$//'
}

input="$(cat)"
file_path="$(extract_file_path "$input" || true)"
[[ -n "$file_path" ]] || exit 0

case "$file_path" in
  *.yuck)
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
    if command -v sassc >/dev/null 2>&1 && [[ -f "$file_path" ]]; then
      if ! sassc --style compressed "$file_path" /dev/null 2>/dev/null; then
        echo "[validate] SCSS syntax error in $file_path" >&2
        exit 1
      fi
    fi
    ;;
esac

exit 0
