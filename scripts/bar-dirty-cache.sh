#!/usr/bin/env bash
# bar-dirty-cache.sh — Dirty-repo cache for the ticker dirty-repos stream
#
# Walks ~/hairglasses-studio/*/ and counts uncommitted changes per repo via
# `git status --porcelain`. Writes summary + per-repo lines to
# /tmp/bar-dirty.txt. Paired with bar-dirty.timer (5-minute interval).
#
# Output:
#   Line 1: <total-dirty-repos> repos  <total-files> changes
#   Lines 2+: <count>  <repo>  (top N by count descending)

set -euo pipefail

CACHE_FILE="/tmp/bar-dirty.txt"
TMPFILE="$(mktemp /tmp/bar-dirty.XXXXXX)"
trap 'rm -f "$TMPFILE"' EXIT

WORKSPACE="$HOME/hairglasses-studio"
[[ -d "$WORKSPACE" ]] || { : > "$TMPFILE"; mv "$TMPFILE" "$CACHE_FILE"; exit 0; }

total_repos=0
total_files=0
declare -a rows

for repo in "$WORKSPACE"/*/; do
  [[ -d "$repo/.git" ]] || continue
  count=$(git -C "$repo" status --porcelain 2>/dev/null | wc -l)
  (( count > 0 )) || continue
  total_repos=$((total_repos + 1))
  total_files=$((total_files + count))
  name=$(basename "$repo")
  rows+=("$(printf '%d\t%s' "$count" "$name")")
done

{
  printf '%d repos %d changes\n' "$total_repos" "$total_files"
  printf '%s\n' "${rows[@]}" | sort -rn | head -10 \
    | awk -F '\t' '{printf "%s  %s\n", $1, $2}'
} > "$TMPFILE"

mv "$TMPFILE" "$CACHE_FILE"
