#!/usr/bin/env bash
# hg-fleet-health.sh — Fleet health dashboard for hairglasses-studio repos.
# Reports: last commit, CI status, test file count per repo.
# Usage: hg-fleet-health.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

hg_require git gh

STUDIO="$HOME/hairglasses-studio"

hg_info "Fleet Health Check"
printf "%s%s%s\n" "$HG_DIM" "$(date '+%Y-%m-%d %H:%M')" "$HG_RESET"
echo ""

TOTAL=0
PASSING=0
FAILING=0

printf "%s%-30s %-14s %-12s %-8s%s\n" "$HG_BOLD" "REPO" "LAST COMMIT" "CI" "TESTS" "$HG_RESET"
printf "%s%s%s\n" "$HG_DIM" "$(printf '─%.0s' {1..70})" "$HG_RESET"

for d in "$STUDIO"/*/; do
  [[ -d "$d/.git" ]] || continue
  name=$(basename "$d")
  [[ "$name" == "dotfiles" ]] && continue

  cd "$d"
  TOTAL=$((TOTAL + 1))

  # Last commit date
  LAST_COMMIT=$(git log -1 --format='%cd' --date=short 2>/dev/null || echo "n/a")

  # Check staleness (>7 days)
  COMMIT_EPOCH=$(git log -1 --format='%ct' 2>/dev/null || echo "0")
  NOW_EPOCH=$(date +%s)
  DAYS_AGO=$(( (NOW_EPOCH - COMMIT_EPOCH) / 86400 ))
  if [[ $DAYS_AGO -gt 7 ]]; then
    LAST_COMMIT_FMT="${HG_YELLOW}${LAST_COMMIT}${HG_RESET}"
  else
    LAST_COMMIT_FMT="${HG_GREEN}${LAST_COMMIT}${HG_RESET}"
  fi

  # CI status via gh
  CI_STATUS="-"
  CI_JSON=$(gh run list --limit 1 --json conclusion,status 2>/dev/null || echo "[]")
  if [[ "$CI_JSON" != "[]" ]] && [[ "$CI_JSON" != "" ]]; then
    CONCLUSION=$(echo "$CI_JSON" | grep -oP '"conclusion"\s*:\s*"\K[^"]*' | head -1 || true)
    STATUS=$(echo "$CI_JSON" | grep -oP '"status"\s*:\s*"\K[^"]*' | head -1 || true)
    if [[ "$CONCLUSION" == "success" ]]; then
      CI_STATUS="${HG_GREEN}pass${HG_RESET}"
      PASSING=$((PASSING + 1))
    elif [[ "$CONCLUSION" == "failure" ]]; then
      CI_STATUS="${HG_RED}fail${HG_RESET}"
      FAILING=$((FAILING + 1))
    elif [[ "$STATUS" == "in_progress" ]] || [[ "$STATUS" == "queued" ]]; then
      CI_STATUS="${HG_YELLOW}running${HG_RESET}"
    elif [[ -n "$CONCLUSION" ]]; then
      CI_STATUS="${HG_DIM}${CONCLUSION}${HG_RESET}"
    else
      CI_STATUS="${HG_DIM}none${HG_RESET}"
    fi
  else
    CI_STATUS="${HG_DIM}none${HG_RESET}"
  fi

  # Count test files
  TEST_COUNT=0
  if [[ -f go.mod ]]; then
    TEST_COUNT=$(find . -name '*_test.go' -not -path './vendor/*' 2>/dev/null | wc -l)
  elif [[ -f package.json ]]; then
    TEST_COUNT=$(find . -name '*.test.*' -o -name '*.spec.*' -not -path './node_modules/*' 2>/dev/null | wc -l)
  elif [[ -f pyproject.toml ]] || [[ -f requirements.txt ]]; then
    TEST_COUNT=$(find . -name 'test_*.py' -o -name '*_test.py' 2>/dev/null | wc -l)
  fi

  if [[ $TEST_COUNT -eq 0 ]]; then
    TEST_FMT="${HG_DIM}0${HG_RESET}"
  else
    TEST_FMT="${HG_GREEN}${TEST_COUNT}${HG_RESET}"
  fi

  printf "%-30s %-14b %-12b %-8b\n" "$name" "$LAST_COMMIT_FMT" "$CI_STATUS" "$TEST_FMT"
done

echo ""
printf "%s%s%s\n" "$HG_DIM" "$(printf '─%.0s' {1..70})" "$HG_RESET"
printf "Total: %d repos | " "$TOTAL"
printf "%sPassing: %d%s | " "$HG_GREEN" "$PASSING" "$HG_RESET"
printf "%sFailing: %d%s\n" "$HG_RED" "$FAILING" "$HG_RESET"
