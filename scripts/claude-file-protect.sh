#!/usr/bin/env bash
# claude-file-protect.sh — PreToolUse hook: block writes to protected files
# Protected files are critical infrastructure that should only change via
# deliberate commits, not as side-effects of agent operations.
#
# Input: JSON on stdin with tool_name, tool_input fields
# Output: JSON to stdout — {"decision":"block","reason":"..."} or {"decision":"allow"}

set -euo pipefail

PROTECTED_FILES=(
  "go.mod"
  "go.sum"
  "go.work"
  "go.work.sum"
  "pipeline.mk"
  "Makefile"
  ".well-known/mcp.json"
  "snapshots/contract/overview.json"
  "snapshots/contract/tools.json"
)

# Read tool invocation from stdin
INPUT="$(cat)"

TOOL_NAME="$(printf '%s' "$INPUT" | jq -r '.tool_name // empty' 2>/dev/null || true)"

# Only check file-writing tools
case "$TOOL_NAME" in
  Write|Edit|NotebookEdit) ;;
  *) printf '{"decision":"allow"}\n'; exit 0 ;;
esac

# Extract the file path from tool input
FILE_PATH="$(printf '%s' "$INPUT" | jq -r '.tool_input.file_path // .tool_input.path // empty' 2>/dev/null || true)"

if [[ -z "$FILE_PATH" ]]; then
  printf '{"decision":"allow"}\n'
  exit 0
fi

# Normalize to relative path
FILE_BASE="$(basename "$FILE_PATH")"
FILE_REL="${FILE_PATH##*/hairglasses-studio/dotfiles/}"

for protected in "${PROTECTED_FILES[@]}"; do
  if [[ "$FILE_REL" == "$protected" || "$FILE_BASE" == "$protected" || "$FILE_PATH" == *"/$protected" ]]; then
    printf '{"decision":"block","reason":"Protected file: %s is infrastructure-critical. Edit it in a dedicated commit, not as a side-effect."}\n' "$protected"
    exit 0
  fi
done

printf '{"decision":"allow"}\n'
