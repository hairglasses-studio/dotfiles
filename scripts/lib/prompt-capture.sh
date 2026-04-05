#!/usr/bin/env bash
# prompt-capture.sh — UserPromptSubmit hook: captures multi-line prompts to the docs repo
# Called by Claude Code on every prompt submission. Reads JSON from stdin.
# Always exits 0 (non-blocking).
set -euo pipefail

DOCS_PROMPTS="$HOME/hairglasses-studio/docs/prompts"
INDEX_FILE="$DOCS_PROMPTS/.prompt-index.jsonl"
STUDIO_ROOT="$HOME/hairglasses-studio"

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

# Read the full hook JSON from stdin once
read_hook_input() {
  HOOK_JSON="$(cat)"
}

extract_field() {
  printf '%s' "$HOOK_JSON" | jq -r "$1 // empty" 2>/dev/null
}

# Determine repo name from CWD path.
# /home/hg/hairglasses-studio/dotfiles -> dotfiles
# /home/hg/other/project -> _other
repo_from_cwd() {
  local cwd="$1"
  if [[ "$cwd" == "$STUDIO_ROOT"/* ]]; then
    # Strip studio root, take first path component
    local rel="${cwd#"$STUDIO_ROOT"/}"
    printf '%s' "${rel%%/*}"
  else
    printf '_other'
  fi
}

# Filter: should we capture this prompt?
should_capture() {
  local prompt="$1"

  # Skip empty
  [[ -z "$prompt" ]] && return 1

  # Count words
  local wc
  wc=$(printf '%s' "$prompt" | wc -w)
  (( wc < 5 )) && return 1

  # Skip single-line prompts (fewer than 2 lines)
  local lc
  lc=$(printf '%s' "$prompt" | wc -l)
  (( lc < 2 )) && return 1

  # Skip slash commands
  [[ "$prompt" =~ ^/ ]] && return 1

  # Skip conversational one-worders
  local lower
  lower=$(printf '%s' "$prompt" | tr '[:upper:]' '[:lower:]' | xargs)
  case "$lower" in
    y|n|yes|no|ok|okay|sure|continue|approve|deny|lgtm|thanks|thx|done|cancel|stop|go|yep|nope|k)
      return 1 ;;
  esac

  return 0
}

compute_hash() {
  printf '%s' "$1" | sha256sum | cut -d' ' -f1
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

main() {
  read_hook_input

  local prompt cwd session_id
  prompt="$(extract_field '.prompt')"
  cwd="$(extract_field '.cwd')"
  session_id="$(extract_field '.session_id')"

  # Filter
  should_capture "$prompt" || exit 0

  # Compute metadata
  local hash short_hash repo timestamp word_count
  hash="$(compute_hash "$prompt")"
  short_hash="${hash:0:12}"
  repo="$(repo_from_cwd "${cwd:-$PWD}")"
  timestamp="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  word_count=$(printf '%s' "$prompt" | wc -w)

  # Ensure directories exist
  local unsorted_dir="$DOCS_PROMPTS/$repo/unsorted"
  mkdir -p "$unsorted_dir"

  local target="$unsorted_dir/${short_hash}.md"

  # Skip if already captured (dedup by hash)
  [[ -f "$target" ]] && exit 0

  # Build TOML frontmatter + prompt body
  local content
  content="$(cat <<FRONTMATTER
+++
hash = "$hash"
short_hash = "$short_hash"
repo = "$repo"
timestamp = "$timestamp"
session_id = "${session_id:-}"
word_count = $word_count
task_type = ""
score = 0
grade = ""
tags = []
status = "unsorted"
original_hash = ""
+++

$prompt
FRONTMATTER
)"

  # Atomic write the .md file
  local tmp
  tmp="$(mktemp "${target}.XXXXXX")"
  printf '%s\n' "$content" > "$tmp"
  mv -f "$tmp" "$target"

  # Append to JSONL index with proper JSON escaping and file locking
  local json_line
  json_line=$(jq -n \
    --arg hash "$hash" \
    --arg short_hash "$short_hash" \
    --arg repo "$repo" \
    --arg ts "$timestamp" \
    --arg sid "${session_id:-}" \
    --argjson wc "$word_count" \
    --arg prompt "${prompt:0:500}" \
    '{hash:$hash, short_hash:$short_hash, repo:$repo, timestamp:$ts, session_id:$sid, word_count:$wc, task_type:"", score:0, grade:"", tags:[], status:"unsorted", prompt:$prompt}')
  (flock -x 200; printf '%s\n' "$json_line" >> "$INDEX_FILE") 200>"${INDEX_FILE}.lock"
}

main "$@"
exit 0
