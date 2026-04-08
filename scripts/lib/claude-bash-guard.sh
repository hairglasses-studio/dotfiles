#!/usr/bin/env bash
# PreToolUse Bash hook: block known dangerous command patterns
# Exit 2 = block with feedback; Exit 0 = allow
set -euo pipefail

input="${CLAUDE_TOOL_INPUT:-}"

# Extract command from JSON input
cmd=$(echo "$input" | grep -oP '"command"\s*:\s*"([^"]*)"' | head -1 | sed 's/.*: *"//;s/"$//' || true)
[[ -z "$cmd" ]] && exit 0

# Block force-push to main/master
if echo "$cmd" | grep -qP 'git\s+push\s+.*--force.*\b(main|master)\b'; then
  echo "BLOCKED: force-push to main/master. Use a feature branch or --force-with-lease." >&2
  exit 2
fi
if echo "$cmd" | grep -qP 'git\s+push\s+--force\s+origin\s+(main|master)'; then
  echo "BLOCKED: force-push to main/master." >&2
  exit 2
fi

# Block destructive rm on root or home
if echo "$cmd" | grep -qP 'rm\s+-rf?\s+/\s' || echo "$cmd" | grep -qP 'rm\s+-rf?\s+/$'; then
  echo "BLOCKED: rm -rf / is never allowed." >&2
  exit 2
fi

# Block piped remote execution
if echo "$cmd" | grep -qP 'curl\s.*\|\s*(ba)?sh' || echo "$cmd" | grep -qP 'wget\s.*\|\s*(ba)?sh'; then
  echo "BLOCKED: piping remote content to shell. Download first, review, then execute." >&2
  exit 2
fi

exit 0
