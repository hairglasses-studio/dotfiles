#!/usr/bin/env bash
# bar-archnews-cache.sh — Arch Linux news cache for the ticker arch-news stream
#
# Fetches the Arch Linux news RSS feed and extracts the 5 most recent titles.
# Writes summary + titles to /tmp/bar-archnews.txt.
# Paired with bar-archnews.timer (hourly).
#
# Output:
#   Line 1: <N> recent
#   Lines 2+: <title>

set -euo pipefail

CACHE_FILE="/tmp/bar-archnews.txt"
TMPFILE="$(mktemp /tmp/bar-archnews.XXXXXX)"
trap 'rm -f "$TMPFILE"' EXIT

FEED="https://archlinux.org/feeds/news/"

if ! command -v curl >/dev/null || ! command -v xmllint >/dev/null; then
  printf 'curl/xmllint missing\n' > "$TMPFILE"
  mv "$TMPFILE" "$CACHE_FILE"
  exit 0
fi

# Fetch feed (5s timeout). On failure, leave stale cache if it exists.
xml="$(curl -fsSL --max-time 5 "$FEED" 2>/dev/null)" || exit 0

titles="$(printf '%s' "$xml" | xmllint --xpath '//item/title/text()' - 2>/dev/null \
  || true)"

# Split titles: each <title> ends before the next starts; xmllint concatenates
# without separators, so we extract via a fresh pass with sed
titles_list="$(printf '%s' "$xml" \
  | tr '\n' ' ' \
  | grep -oP '<item>.*?</item>' \
  | grep -oP '<title>\K[^<]+' \
  | head -5)"

decode_entities() {
  sed -e 's/&gt;/>/g' -e 's/&lt;/</g' -e 's/&amp;/\&/g' -e 's/&quot;/"/g' -e "s/&apos;/'/g" -e 's/&#x27;/'\''/g' -e 's/&#39;/'\''/g'
}

{
  count="$(printf '%s\n' "$titles_list" | grep -c . || true)"
  printf '%d recent\n' "${count:-0}"
  printf '%s\n' "$titles_list" | decode_entities
} > "$TMPFILE"

mv "$TMPFILE" "$CACHE_FILE"
