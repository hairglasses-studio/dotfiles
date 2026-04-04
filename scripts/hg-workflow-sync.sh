#!/usr/bin/env bash
# hg-workflow-sync.sh — Sync CI workflow files across all hairglasses-studio repos.
# Compares each repo's workflows against canonical sources and updates stale copies.
# Usage: hg-workflow-sync.sh [--dry-run] [--commit] [--push]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

STUDIO="$HOME/hairglasses-studio"
DOTFILES="$STUDIO/dotfiles"
ORG_GITHUB="$STUDIO/.github"

DRY_RUN=false
COMMIT=false
PUSH=false

for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=true ;;
    --commit)  COMMIT=true ;;
    --push)    PUSH=true; COMMIT=true ;;
  esac
done

# ── Canonical workflow sources ───────────────
declare -A CANONICAL
CANONICAL[ci.yml]="$DOTFILES/make/ci-go.yml"
CANONICAL[claude-review.yml]="$ORG_GITHUB/workflow-templates/claude-review.yml"
CANONICAL[claude-security.yml]="$ORG_GITHUB/workflow-templates/claude-security.yml"
CANONICAL[codex-review.yml]="$ORG_GITHUB/workflow-templates/codex-review.yml"
CANONICAL[codex-security.yml]="$ORG_GITHUB/workflow-templates/codex-security.yml"
CANONICAL[dependabot-auto-merge.yml]="$STUDIO/mcpkit/.github/workflows/dependabot-auto-merge.yml"

hg_info "Workflow sync — comparing against canonical sources"
$DRY_RUN && hg_warn "Dry-run mode — no files will be changed"
echo ""

UPDATED=0
SKIPPED=0
CURRENT=0

for d in "$STUDIO"/*/; do
  [[ -d "$d/.git" ]] || continue
  [[ -d "$d/.github/workflows" ]] || continue
  name=$(basename "$d")

  # Skip repos with custom CI (ralphglasses has 260L custom CI)
  [[ "$name" == "ralphglasses" ]] && continue
  [[ "$name" == "dotfiles" ]] && continue
  [[ "$name" == ".github" ]] && continue

  REPO_CHANGES=false

  for wf in "${!CANONICAL[@]}"; do
    target="$d/.github/workflows/$wf"
    source="${CANONICAL[$wf]}"

    [[ -f "$target" ]] || continue
    [[ -f "$source" ]] || continue

    # Only sync ci.yml for Go repos
    if [[ "$wf" == "ci.yml" ]] && [[ ! -f "$d/go.mod" ]]; then
      continue
    fi

    if ! diff -q "$source" "$target" &>/dev/null; then
      if $DRY_RUN; then
        printf "%s%-25s %s (would update)%s\n" "$HG_YELLOW" "$name" "$wf" "$HG_RESET"
      else
        command cp -f "$source" "$target"
        printf "%s%-25s %s (updated)%s\n" "$HG_GREEN" "$name" "$wf" "$HG_RESET"
        REPO_CHANGES=true
      fi
      UPDATED=$((UPDATED + 1))
    else
      CURRENT=$((CURRENT + 1))
    fi
  done

  if $REPO_CHANGES && $COMMIT; then
    cd "$d"
      git add .github/workflows/ 2>/dev/null
    if ! git diff --cached --quiet 2>/dev/null; then
      git commit -q -m "ci: sync workflows via hg-workflow-sync.sh"
      if $PUSH; then
        git pull --rebase -q 2>/dev/null
        git push -q 2>&1 && printf "%s  → pushed%s\n" "$HG_DIM" "$HG_RESET" || printf "%s  → push failed%s\n" "$HG_RED" "$HG_RESET"
      fi
    fi
  fi
done

echo ""
if $DRY_RUN; then
  hg_info "$UPDATED workflows would be updated, $CURRENT already current"
else
  hg_ok "$UPDATED workflows updated, $CURRENT already current"
fi
