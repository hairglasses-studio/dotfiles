#!/usr/bin/env bash
# bar-hn-cache.sh — Hacker News top stories cache for the ticker hn-top stream
#
# Fetches the first 10 story IDs from the HN Firebase API, then fetches each
# item's title + score. Writes 1 line per story to /tmp/bar-hn.txt:
#   #<id>\t<score>\t<title>
# Paired with bar-hn.timer (10-minute interval).

set -euo pipefail

CACHE_FILE="/tmp/bar-hn.txt"
TMPFILE="$(mktemp /tmp/bar-hn.XXXXXX)"
trap 'rm -f "$TMPFILE"' EXIT

if ! command -v curl >/dev/null || ! command -v jq >/dev/null; then
  printf 'curl/jq missing\n' > "$TMPFILE"
  mv "$TMPFILE" "$CACHE_FILE"
  exit 0
fi

API="https://hacker-news.firebaseio.com/v0"
top_ids="$(curl -fsSL --max-time 5 "$API/topstories.json" 2>/dev/null \
  | jq -r '.[:10][]' 2>/dev/null)" || exit 0

[[ -z "$top_ids" ]] && exit 0

{
  while IFS= read -r id; do
    [[ -z "$id" ]] && continue
    item="$(curl -fsSL --max-time 3 "$API/item/$id.json" 2>/dev/null)" || continue
    title="$(printf '%s' "$item" | jq -r '.title // empty' 2>/dev/null)"
    score="$(printf '%s' "$item" | jq -r '.score // 0'     2>/dev/null)"
    [[ -z "$title" ]] && continue
    # Keep titles under 100 chars for the ticker
    title="${title:0:100}"
    # Collapse tabs to spaces — ticker parser splits on tab
    title="${title//$'\t'/ }"
    printf '#%s\t%s\t%s\n' "$id" "$score" "$title"
  done <<< "$top_ids"
} > "$TMPFILE"

mv "$TMPFILE" "$CACHE_FILE"
