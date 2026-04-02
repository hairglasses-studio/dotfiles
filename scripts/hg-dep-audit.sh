#!/usr/bin/env bash
# hg-dep-audit.sh — Audit Go dependency versions across all hairglasses-studio repos.
# Reports version skew for key shared dependencies.
# Usage: hg-dep-audit.sh [--json]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

STUDIO="$HOME/hairglasses-studio"
JSON_OUTPUT=false
[[ "${1:-}" == "--json" ]] && JSON_OUTPUT=true

# Key dependencies to track
DEPS=(
  "github.com/mark3labs/mcp-go"
  "github.com/spf13/cobra"
  "github.com/charmbracelet/bubbletea"
  "github.com/charmbracelet/lipgloss"
  "google.golang.org/api"
  "github.com/google/uuid"
  "go.opentelemetry.io/otel"
)

hg_info "Dependency version audit across hairglasses-studio Go repos"
echo ""

# Collect versions
declare -A DEP_VERSIONS  # dep:repo → version

for d in "$STUDIO"/*/; do
  [[ -f "$d/go.mod" ]] || continue
  name=$(basename "$d")

  for dep in "${DEPS[@]}"; do
    ver=$(grep -m1 "^\s*${dep} " "$d/go.mod" 2>/dev/null | awk '{print $2}' || true)
    if [[ -n "$ver" ]]; then
      DEP_VERSIONS["$dep:$name"]="$ver"
    fi
  done
done

# Report per dependency
for dep in "${DEPS[@]}"; do
  # Collect all versions for this dep
  declare -A VERSIONS
  VERSIONS=()
  for key in "${!DEP_VERSIONS[@]}"; do
    if [[ "$key" == "$dep:"* ]]; then
      repo="${key#*:}"
      ver="${DEP_VERSIONS[$key]}"
      VERSIONS[$ver]+="$repo "
    fi
  done

  if [[ ${#VERSIONS[@]} -eq 0 ]]; then
    continue
  fi

  # Check if all versions are the same
  if [[ ${#VERSIONS[@]} -eq 1 ]]; then
    ver="${!VERSIONS[@]}"
    count=$(echo "${VERSIONS[$ver]}" | wc -w)
    printf "%s%-45s %s (%d repos)%s\n" "$HG_GREEN" "$dep" "$ver" "$count" "$HG_RESET"
  else
    printf "%s%-45s %d versions (SKEW)%s\n" "$HG_YELLOW" "$dep" "${#VERSIONS[@]}" "$HG_RESET"
    for ver in $(echo "${!VERSIONS[@]}" | tr ' ' '\n' | sort -V); do
      repos="${VERSIONS[$ver]}"
      count=$(echo "$repos" | wc -w)
      printf "  %s%-20s %s(%d: %s)%s\n" "$HG_DIM" "$ver" "$HG_DIM" "$count" "$(echo $repos | tr ' ' ', ' | sed 's/,$//')" "$HG_RESET"
    done
  fi

  unset VERSIONS
done

echo ""
hg_ok "Audit complete"
