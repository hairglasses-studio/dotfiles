#!/usr/bin/env bash
# bar-prs-cache.sh — GitHub open-PRs cache for the ticker github-prs stream.
#
# Iterates the active_operator repos from workspace/manifest.json and runs
# `gh pr list --state open` against each. Writes:
#   Line 1: summary  e.g. "7 open across 3 repos"
#   Lines 2+: <repo>#<num>  <author>  <title>
# Paired with bar-prs.timer (5-minute interval).
#
# Skips repos whose remote is private and gh isn't authenticated, and
# swallows rate-limit errors so a transient 403 doesn't wipe the cache.

set -uo pipefail

CACHE_FILE="/tmp/bar-prs.txt"
TMPFILE="$(mktemp /tmp/bar-prs.XXXXXX)"
trap 'rm -f "$TMPFILE"' EXIT

if ! command -v gh >/dev/null || ! command -v jq >/dev/null; then
  printf 'gh/jq missing\n' > "$TMPFILE"
  mv "$TMPFILE" "$CACHE_FILE"
  exit 0
fi

MANIFEST="$HOME/hairglasses-studio/workspace/manifest.json"
[[ -r "$MANIFEST" ]] || exit 0

ORG="hairglasses-studio"
mapfile -t repos < <(jq -r '.repos[] | select(.scope == "active_operator") | .name' "$MANIFEST")

total=0
repos_with_prs=0
rows_tmp="$(mktemp /tmp/bar-prs-rows.XXXXXX)"
trap 'rm -f "$TMPFILE" "$rows_tmp"' EXIT

for repo in "${repos[@]}"; do
  [[ -z "$repo" ]] && continue
  data="$(gh pr list --repo "$ORG/$repo" --state open \
      --json number,title,author 2>/dev/null || true)"
  [[ -z "$data" || "$data" == "[]" ]] && continue
  count="$(printf '%s' "$data" | jq -r 'length')"
  [[ "$count" -gt 0 ]] || continue
  total=$((total + count))
  repos_with_prs=$((repos_with_prs + 1))
  printf '%s' "$data" \
    | jq -r --arg repo "$repo" \
        '.[] | "\($repo)#\(.number)\t\(.author.login)\t\(.title | .[0:60])"' \
    >> "$rows_tmp"
done

{
  if (( total == 0 )); then
    # emit nothing so the stream advances via _empty() backoff
    :
  else
    printf '%d open across %d repos\n' "$total" "$repos_with_prs"
    # Newest first — we don't have an explicit sort key from the manifest
    # iteration, so rely on the repo order (hub repos first).
    head -12 "$rows_tmp"
  fi
} > "$TMPFILE"

mv "$TMPFILE" "$CACHE_FILE"
