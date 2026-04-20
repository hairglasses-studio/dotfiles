#!/usr/bin/env bash
# bar-ci-cache.sh — Cache writer for the Ironbar / ticker CI status widget
#
# Queries `gh run list -L 1` for each active CI-bearing repo in
# workspace/manifest.json (lifecycle=canonical, workflow_family != none/null)
# and writes a compact aggregate to /tmp/bar-ci.txt.
#
# Run by bar-ci.timer (every 5 minutes).
#
# Cache format:
#   Line 1: PASS=<n> FAIL=<n> RUN=<n>
#   Lines 2+: <repo>\t<conclusion-or-status>  (only for non-success runs)

set -euo pipefail

CACHE_FILE="/tmp/bar-ci.txt"
TMPFILE="$(mktemp /tmp/bar-ci.XXXXXX)"
trap 'rm -f "$TMPFILE"' EXIT

MANIFEST="$HOME/hairglasses-studio/workspace/manifest.json"
ORG="hairglasses-studio"

if ! command -v gh >/dev/null || ! command -v jq >/dev/null; then
  printf 'CI=unavailable\n' > "$TMPFILE"
  mv "$TMPFILE" "$CACHE_FILE"
  exit 0
fi

repos=()
if [[ -f "$MANIFEST" ]]; then
  while IFS= read -r r; do
    repos+=("$r")
  done < <(jq -r '.repos[]
                  | select(.lifecycle == "canonical"
                           and .workflow_family != "none"
                           and .workflow_family != null)
                  | .name' "$MANIFEST")
fi

pass=0; fail=0; run=0
fails=()

for repo in "${repos[@]}"; do
  json="$(gh run list -L 1 -R "${ORG}/${repo}" \
            --json status,conclusion,name 2>/dev/null || true)"
  [[ -z "$json" || "$json" == "[]" ]] && continue
  status="$(jq -r '.[0].status // "unknown"' <<<"$json")"
  conclusion="$(jq -r '.[0].conclusion // "unknown"' <<<"$json")"
  if [[ "$status" != "completed" ]]; then
    run=$((run + 1))
    fails+=("${repo}	running:${status}")
  elif [[ "$conclusion" == "success" ]]; then
    pass=$((pass + 1))
  else
    fail=$((fail + 1))
    fails+=("${repo}	${conclusion}")
  fi
done

{
  printf 'PASS=%d FAIL=%d RUN=%d\n' "$pass" "$fail" "$run"
  for line in "${fails[@]}"; do
    printf '%s\n' "$line"
  done
} > "$TMPFILE"

mv "$TMPFILE" "$CACHE_FILE"
