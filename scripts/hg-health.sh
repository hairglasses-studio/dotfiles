#!/usr/bin/env bash
# hg-health.sh — Org-wide repo health dashboard for hairglasses-studio.
# Reports: build status, Go version, workflow freshness, pipeline inclusion.
# Usage: hg-health.sh [--json]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

STUDIO="$HOME/hairglasses-studio"
DOTFILES="$STUDIO/dotfiles"
TARGET_GO=$(cat "$DOTFILES/make/go-version" 2>/dev/null | tr -d '[:space:]')
JSON_OUTPUT=false
[[ "${1:-}" == "--json" ]] && JSON_OUTPUT=true

hg_info "hairglasses-studio Health Dashboard"
printf "%s%s%s\n" "$HG_DIM" "$(date '+%Y-%m-%d %H:%M')" "$HG_RESET"
echo ""

TOTAL=0
HEALTHY=0
WARN=0
# shellcheck disable=SC2034
FAIL=0

printf "%s%-25s %-8s %-10s %-10s %-10s %-8s%s\n" "$HG_BOLD" "REPO" "LANG" "BUILD" "GO VER" "PIPELINE" "CI" "$HG_RESET"
printf "%s%s%s\n" "$HG_DIM" "$(printf '─%.0s' {1..80})" "$HG_RESET"

for d in "$STUDIO"/*/; do
  [[ -d "$d/.git" ]] || continue
  name=$(basename "$d")
  [[ "$name" == ".github" ]] && continue
  [[ "$name" == "dotfiles" ]] && continue

  cd "$d"
  TOTAL=$((TOTAL + 1))
  ISSUES=0

  # Detect language
  LANG="-"
  [[ -f go.mod ]] && LANG="Go"
  [[ -f package.json ]] && [[ "$LANG" == "-" ]] && LANG="Node"
  [[ -f pyproject.toml ]] && [[ "$LANG" == "-" ]] && LANG="Py"

  # Build check (Go only, quick)
  BUILD="-"
  if [[ "$LANG" == "Go" ]]; then
    if go build ./... 2>/dev/null; then
      BUILD="${HG_GREEN}OK${HG_RESET}"
    else
      BUILD="${HG_RED}FAIL${HG_RESET}"
      ISSUES=$((ISSUES + 1))
    fi
  fi

  # Go version check
  GOVER="-"
  if [[ -f go.mod ]]; then
    ver=$(grep -m1 '^go ' go.mod | awk '{print $2}')
    if [[ "$ver" == "$TARGET_GO" ]]; then
      GOVER="${HG_GREEN}${ver}${HG_RESET}"
    else
      GOVER="${HG_YELLOW}${ver}${HG_RESET}"
      ISSUES=$((ISSUES + 1))
    fi
  fi

  # Pipeline.mk inclusion
  PIPELINE="-"
  if [[ -f Makefile ]]; then
    if grep -q 'pipeline.mk' Makefile 2>/dev/null; then
      PIPELINE="${HG_GREEN}yes${HG_RESET}"
    else
      PIPELINE="${HG_YELLOW}no${HG_RESET}"
      ISSUES=$((ISSUES + 1))
    fi
  fi

  # CI workflow present
  CI="-"
  if [[ -f .github/workflows/ci.yml ]]; then
    CI="${HG_GREEN}yes${HG_RESET}"
  elif [[ -d .github/workflows ]]; then
    CI="${HG_YELLOW}other${HG_RESET}"
  else
    CI="${HG_DIM}none${HG_RESET}"
  fi

  # Tally
  if [[ $ISSUES -eq 0 ]] && [[ "$LANG" != "-" ]]; then
    HEALTHY=$((HEALTHY + 1))
  elif [[ $ISSUES -gt 0 ]]; then
    WARN=$((WARN + 1))
  fi

  printf "%-25s %-8s %-10b %-10b %-10b %-8b\n" "$name" "$LANG" "$BUILD" "$GOVER" "$PIPELINE" "$CI"
done

echo ""
printf "%s%s%s\n" "$HG_DIM" "$(printf '─%.0s' {1..80})" "$HG_RESET"
printf "Total: %d | " "$TOTAL"
printf "%sHealthy: %d%s | " "$HG_GREEN" "$HEALTHY" "$HG_RESET"
printf "%sWarnings: %d%s | " "$HG_YELLOW" "$WARN" "$HG_RESET"
printf "Target Go: %s\n" "$TARGET_GO"
