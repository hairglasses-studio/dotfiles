#!/usr/bin/env bash
# PreToolUse Bash hook: block known dangerous command patterns
# Exit 2 = block with feedback; Exit 0 = allow
set -euo pipefail

read_input() {
  local input=""

  if [[ ! -t 0 ]]; then
    input="$(cat)"
    if [[ -n "$input" ]]; then
      printf '%s' "$input"
      return
    fi
  fi

  printf '%s' "${CLAUDE_TOOL_INPUT:-}"
}

extract_command() {
  local input="$1"

  if [[ -z "$input" ]]; then
    return 0
  fi

  if command -v jq >/dev/null 2>&1; then
    jq -r '.command // empty' <<<"$input" 2>/dev/null
    return 0
  fi

  printf '%s' "$input" |
    tr '\n' ' ' |
    grep -oE '"command"[[:space:]]*:[[:space:]]*"([^"\\]|\\.)*"' |
    head -1 |
    sed -E 's/^"command"[[:space:]]*:[[:space:]]*"//; s/"$//'
}

input="$(read_input)"
cmd="$(extract_command "$input" || true)"
[[ -z "$cmd" ]] && exit 0

# Block force-push to main/master
if printf '%s\n' "$cmd" | grep -qE 'git[[:space:]]+push([[:space:]].*)?([[:space:]]+--force|[[:space:]]+--force-with-lease)([[:space:]].*)?([[:space:]]|^)(main|master)$'; then
  echo "BLOCK: force-push to main/master is disabled. Use a feature branch or --force-with-lease on a non-protected branch." >&2
  exit 2
fi
if printf '%s\n' "$cmd" | grep -qE 'git[[:space:]]+push([[:space:]].*)?(origin[[:space:]]+)?(main|master)([[:space:]].*)?([[:space:]]+--force|[[:space:]]+--force-with-lease)'; then
  echo "BLOCK: force-push to main/master is disabled. Use a feature branch or --force-with-lease on a non-protected branch." >&2
  exit 2
fi

# Block destructive rm on root
if printf '%s\n' "$cmd" | grep -qE '(^|[[:space:]])rm[[:space:]]+-r[fF]?[[:space:]]+/([[:space:]]|$)'; then
  echo "BLOCK: rm -rf / is never allowed." >&2
  exit 2
fi

# Block piped remote execution
if printf '%s\n' "$cmd" | grep -qE 'curl[^|]*\|[[:space:]]*(ba)?sh' || printf '%s\n' "$cmd" | grep -qE 'wget[^|]*\|[[:space:]]*(ba)?sh'; then
  echo "BLOCK: piping remote content to shell is disabled. Download first, review it, then execute." >&2
  exit 2
fi

exit 0
