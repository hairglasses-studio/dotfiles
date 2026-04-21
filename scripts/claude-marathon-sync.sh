#!/usr/bin/env bash
# claude-marathon-sync.sh — PostToolUse hook: append marathon-phase
# completions to ROADMAP.md when a `feat(<scope>): Phase N ...` commit
# lands. Closes the Tier-2 hook roadmap item
# (".claude/rules/autonomy gap — marathon completion events").
#
# How it fires:
#   PostToolUse → matcher "Bash" → we inspect the tool_input for
#   `git commit` with a message matching `feat(X): Phase N`. On match,
#   parses the scope + phase number from the subject and appends a
#   bullet under `## Completed Marathon Phases` in ROADMAP.md (creates
#   the section if it's missing). Idempotent — skips if the exact
#   `scope:phase` bullet is already present.
#
# The hook always returns {"decision":"allow"} — it's pure
# instrumentation + post-facto docs sync, never blocks a tool call.
#
# Wire in settings.json:
#   "PostToolUse": [{
#     "matcher": "Bash",
#     "hooks":[{"type":"command","command":"<abs>/claude-marathon-sync.sh"}]
#   }]

set -uo pipefail

INPUT="$(cat 2>/dev/null || echo '{}')"
TOOL="$(printf '%s' "$INPUT" | jq -r '.tool_name // empty' 2>/dev/null || true)"

# Only interested in Bash → git commit calls.
if [[ "$TOOL" != "Bash" ]]; then
  printf '{"decision":"allow"}\n'; exit 0
fi

CMD="$(printf '%s' "$INPUT" | jq -r '.tool_input.command // empty' 2>/dev/null || true)"
# Only fire for git commit commands; be lenient about pipes / heredocs.
if [[ "$CMD" != *"git commit"* ]]; then
  printf '{"decision":"allow"}\n'; exit 0
fi

# Where did this commit happen? Use CWD from the tool input if present;
# fall back to process cwd.
CWD="$(printf '%s' "$INPUT" | jq -r '.cwd // empty' 2>/dev/null || true)"
[[ -n "$CWD" && -d "$CWD" ]] && cd "$CWD" 2>/dev/null || true

# Extract the subject of the newest commit (works whether the hook
# fires during or immediately after `git commit`; the tool_input's
# -m "..." text could be a heredoc, so re-reading from git log is
# more reliable). If we're not inside a repo, exit cleanly.
git rev-parse --git-dir >/dev/null 2>&1 || {
  printf '{"decision":"allow"}\n'; exit 0
}

SUBJECT="$(git log -1 --pretty=format:'%s' 2>/dev/null || true)"
# Match: feat(<scope>): Phase <N> ...  (also perf/fix/refactor)
if [[ ! "$SUBJECT" =~ ^[a-z]+\(([a-z0-9_-]+)\)[[:space:]]*:[[:space:]]*Phase[[:space:]]*([0-9]+) ]]; then
  printf '{"decision":"allow"}\n'; exit 0
fi
SCOPE="${BASH_REMATCH[1]}"
PHASE="${BASH_REMATCH[2]}"
SHORT_SHA="$(git log -1 --pretty=format:'%h' 2>/dev/null || echo '')"
DATE="$(date -u +%Y-%m-%d)"

ROADMAP="$CWD/ROADMAP.md"
[[ -f "$ROADMAP" ]] || {
  printf '{"decision":"allow"}\n'; exit 0
}

# Idempotency: bail if an entry with the same scope:phase already exists.
LINE_KEY="\`${SCOPE}\`: Phase ${PHASE}"
if grep -qF "$LINE_KEY" "$ROADMAP"; then
  printf '{"decision":"allow"}\n'; exit 0
fi

# Build the bullet line.
BULLET="- ${DATE} \`${SCOPE}\`: Phase ${PHASE} — ${SUBJECT#*: } (\`${SHORT_SHA}\`)"

SECTION="## Completed Marathon Phases"
# Append under the existing section if present; otherwise prepend it
# just before the first "## Planned" or "## Future" section (or EOF).
if grep -qF "$SECTION" "$ROADMAP"; then
  # Insert bullet after the section heading (preserving order under it).
  awk -v section="$SECTION" -v bullet="$BULLET" '
    BEGIN { printed = 0 }
    $0 == section { print; getline nextline; if (nextline == "") { print ""; print bullet; next } else { print nextline; print bullet; next } }
    { print }
  ' "$ROADMAP" > "$ROADMAP.tmp" && mv "$ROADMAP.tmp" "$ROADMAP"
else
  # Prepend a new section before the first "## " after the intro. If
  # there is none, append to EOF.
  awk -v section="$SECTION" -v bullet="$BULLET" '
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
  ' "$ROADMAP" > "$ROADMAP.tmp" && mv "$ROADMAP.tmp" "$ROADMAP"
fi

printf '{"decision":"allow","systemMessage":"marathon-sync: appended %s:Phase %s to ROADMAP.md"}\n' \
  "$SCOPE" "$PHASE"
