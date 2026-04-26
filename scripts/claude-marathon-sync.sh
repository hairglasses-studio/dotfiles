#!/usr/bin/env bash
# claude-marathon-sync.sh - PostToolUse hook: sync marathon completions to
# ROADMAP.md and docs-mcp roadmap state.
#
# Fires from PostToolUse with matcher "(^Bash$|(^|__)marathon_advance$)".
# - marathon_advance MCP calls append an idempotent bullet under
#   "## Completed Marathon Phases" in ROADMAP.md.
# - If docs-mcp's SQLite DB exists, a roadmap_events row is inserted so the
#   docs-mcp roadmap timeline sees the sprint completion immediately.
# - Bash git commits keep the legacy "feat(scope): Phase N" fallback.
#
# The hook always returns {"decision":"allow"}; it is instrumentation only.

set -uo pipefail

INPUT="$(cat 2>/dev/null || echo '{}')"
TOOL="$(printf '%s' "$INPUT" | jq -r '.tool_name // empty' 2>/dev/null || true)"

allow() {
  printf '{"decision":"allow"}\n'
}

json_message() {
  local message="$1"
  jq -cn --arg message "$message" '{decision:"allow",systemMessage:$message}'
}

safe_json_string() {
  printf '%s' "${1:-}" | jq -Rs .
}

repo_root_or_cwd() {
  local cwd="$1"

  if [[ -n "$cwd" && -d "$cwd" ]]; then
    cd "$cwd" 2>/dev/null || true
  fi

  git rev-parse --show-toplevel 2>/dev/null || pwd
}

insert_after_section() {
  local file="$1"
  local section="$2"
  local bullet="$3"

  if grep -qF "$section" "$file"; then
    awk -v section="$section" -v bullet="$bullet" '
      $0 == section && !inserted {
        print
        print ""
        print bullet
        inserted = 1
        next
      }
      { print }
    ' "$file" >"$file.tmp" && mv "$file.tmp" "$file"
  else
    awk -v section="$section" -v bullet="$bullet" '
      !done && /^## / {
        print section
        print ""
        print bullet
        print ""
        done = 1
      }
      { print }
      END {
        if (!done) {
          print ""
          print section
          print ""
          print bullet
        }
      }
    ' "$file" >"$file.tmp" && mv "$file.tmp" "$file"
  fi
}

record_docs_event() {
  local repo="$1"
  local summary="$2"
  local docs_db="${DOCS_MCP_DB:-${HOME:-}/hairglasses-studio/docs/.docs.sqlite}"

  [[ -n "$docs_db" && -f "$docs_db" ]] || return 1
  command -v sqlite3 >/dev/null 2>&1 || return 1

  local created repo_json summary_json created_json
  created="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
  repo_json="$(safe_json_string "$repo")"
  summary_json="$(safe_json_string "$summary")"
  created_json="$(safe_json_string "$created")"

  sqlite3 "$docs_db" "
    CREATE TABLE IF NOT EXISTS roadmap_events (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      repo TEXT NOT NULL,
      event_type TEXT NOT NULL,
      summary TEXT,
      created_at TEXT NOT NULL
    );
    INSERT INTO roadmap_events (repo, event_type, summary, created_at)
    VALUES ($repo_json, 'marathon_advance', $summary_json, $created_json);
  " >/dev/null 2>&1
}

handle_marathon_advance() {
  local cwd root roadmap repo
  cwd="$(printf '%s' "$INPUT" | jq -r '.cwd // empty' 2>/dev/null || true)"
  root="$(repo_root_or_cwd "$cwd")"
  roadmap="$root/ROADMAP.md"
  repo="$(basename "$root")"

  local marathon_id sprint_index sprint_name summary tests_passed marathon_done message
  marathon_id="$(printf '%s' "$INPUT" | jq -r '.tool_input.id // empty' 2>/dev/null || true)"
  sprint_index="$(printf '%s' "$INPUT" | jq -r '.tool_response.completed_sprint.index // empty' 2>/dev/null || true)"
  sprint_name="$(printf '%s' "$INPUT" | jq -r '.tool_response.completed_sprint.name // empty' 2>/dev/null || true)"
  summary="$(printf '%s' "$INPUT" | jq -r '.tool_input.summary // .tool_response.message // empty' 2>/dev/null || true)"
  tests_passed="$(printf '%s' "$INPUT" | jq -r '.tool_input.tests_passed // empty' 2>/dev/null || true)"
  marathon_done="$(printf '%s' "$INPUT" | jq -r '.tool_response.marathon_done // false' 2>/dev/null || true)"
  message="$(printf '%s' "$INPUT" | jq -r '.tool_response.message // empty' 2>/dev/null || true)"

  if [[ -z "$sprint_index" || "$sprint_index" == "null" ]]; then
    allow
    exit 0
  fi

  [[ -n "$marathon_id" && "$marathon_id" != "null" ]] || marathon_id="$repo"
  [[ -n "$sprint_name" && "$sprint_name" != "null" ]] || sprint_name="Sprint $sprint_index"
  [[ -n "$summary" && "$summary" != "null" ]] || summary="$message"

  local short_sha
  short_sha="$(git -C "$root" rev-parse --short HEAD 2>/dev/null || true)"

  local date key bullet status suffix
  date="$(date -u +%Y-%m-%d)"
  key="\`${marathon_id}\`: Sprint ${sprint_index}"
  status="tests=${tests_passed:-unknown}"
  suffix=""
  [[ "$marathon_done" == "true" ]] && suffix="; marathon complete"

  bullet="- ${date} ${key} - ${sprint_name}"
  [[ -n "$summary" && "$summary" != "null" ]] && bullet+=" - ${summary}"
  bullet+=" (${status}${suffix}"
  [[ -n "$short_sha" ]] && bullet+="; \`${short_sha}\`"
  bullet+=")"

  local wrote=0
  if [[ -f "$roadmap" ]] && ! grep -qF "$key" "$roadmap"; then
    insert_after_section "$roadmap" "## Completed Marathon Phases" "$bullet"
    wrote=1
  fi

  local event_summary recorded
  event_summary="${key} - ${sprint_name}"
  [[ -n "$summary" && "$summary" != "null" ]] && event_summary+=" - ${summary}"
  if record_docs_event "$repo" "$event_summary"; then
    recorded=1
  else
    recorded=0
  fi

  if (( wrote || recorded )); then
    json_message "marathon-sync: ${event_summary} (roadmap=${wrote}, docs_event=${recorded})"
  else
    allow
  fi
}

handle_git_commit_phase() {
  local cmd cwd subject scope phase short_sha date roadmap line_key bullet

  cmd="$(printf '%s' "$INPUT" | jq -r '.tool_input.command // empty' 2>/dev/null || true)"
  if [[ "$cmd" != *"git commit"* ]]; then
    allow
    exit 0
  fi

  cwd="$(printf '%s' "$INPUT" | jq -r '.cwd // empty' 2>/dev/null || true)"
  [[ -n "$cwd" && -d "$cwd" ]] && cd "$cwd" 2>/dev/null || true

  git rev-parse --git-dir >/dev/null 2>&1 || {
    allow
    exit 0
  }

  subject="$(git log -1 --pretty=format:'%s' 2>/dev/null || true)"
  if [[ ! "$subject" =~ ^[a-z]+\(([a-z0-9_-]+)\)[[:space:]]*:[[:space:]]*Phase[[:space:]]*([0-9]+) ]]; then
    allow
    exit 0
  fi

  scope="${BASH_REMATCH[1]}"
  phase="${BASH_REMATCH[2]}"
  short_sha="$(git log -1 --pretty=format:'%h' 2>/dev/null || echo '')"
  date="$(date -u +%Y-%m-%d)"
  roadmap="$(pwd)/ROADMAP.md"

  [[ -f "$roadmap" ]] || {
    allow
    exit 0
  }

  line_key="\`${scope}\`: Phase ${phase}"
  if grep -qF "$line_key" "$roadmap"; then
    allow
    exit 0
  fi

  bullet="- ${date} \`${scope}\`: Phase ${phase} - ${subject#*: } (\`${short_sha}\`)"
  insert_after_section "$roadmap" "## Completed Marathon Phases" "$bullet"

  json_message "marathon-sync: appended ${scope}:Phase ${phase} to ROADMAP.md"
}

case "$TOOL" in
  marathon_advance|*__marathon_advance)
    handle_marathon_advance
    ;;
  Bash)
    handle_git_commit_phase
    ;;
  *)
    allow
    ;;
esac
