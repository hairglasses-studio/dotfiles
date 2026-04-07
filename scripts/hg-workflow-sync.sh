#!/usr/bin/env bash
# hg-workflow-sync.sh — Sync shared workflow files across hairglasses-studio repos.
# Compares each repo's workflows against canonical sources and updates stale copies.
# Usage: hg-workflow-sync.sh [--dry-run] [--commit] [--push] [--ensure-missing] [--repos=a,b,c]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

STUDIO="$HOME/hairglasses-studio"
DOTFILES="$STUDIO/dotfiles"
ORG_GITHUB="$STUDIO/.github"

DRY_RUN=false
COMMIT=false
PUSH=false
ENSURE_MISSING=false
REPO_FILTER=""

for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=true ;;
    --commit)  COMMIT=true ;;
    --push)    PUSH=true; COMMIT=true ;;
    --ensure-missing) ENSURE_MISSING=true ;;
    --repos=*) REPO_FILTER="${arg#--repos=}" ;;
  esac
done

repo_is_selected() {
  local name="$1"
  [[ -z "$REPO_FILTER" ]] && return 0

  local entry
  IFS=',' read -r -a entries <<<"$REPO_FILTER"
  for entry in "${entries[@]}"; do
    [[ "$name" == "$entry" ]] && return 0
  done
  return 1
}

workflow_source() {
  local wf="$1"
  if [[ -f "$ORG_GITHUB/workflow-templates/$wf" ]]; then
    printf '%s\n' "$ORG_GITHUB/workflow-templates/$wf"
  else
    printf '%s\n' "$ORG_GITHUB/.github/workflows/$wf"
  fi
}

# ── Canonical workflow sources ───────────────
declare -A CANONICAL
CANONICAL[ci.yml]="$DOTFILES/make/ci-go.yml"
CANONICAL[claude-review.yml]="$(workflow_source claude-review.yml)"
CANONICAL[claude-security.yml]="$(workflow_source claude-security.yml)"
CANONICAL[codex-review.yml]="$(workflow_source codex-review.yml)"
CANONICAL[codex-security.yml]="$(workflow_source codex-security.yml)"
CANONICAL[codex-structured-audit.yml]="$(workflow_source codex-structured-audit.yml)"
CANONICAL[codex-baseline-guard.yml]="$ORG_GITHUB/workflow-templates/codex-baseline-guard.yml"
CANONICAL[ai-dispatch.yml]="$(workflow_source ai-dispatch.yml)"
CANONICAL[dependabot-auto-merge.yml]="$STUDIO/mcpkit/.github/workflows/dependabot-auto-merge.yml"

hg_info "Workflow sync — comparing against canonical sources"
$DRY_RUN && hg_warn "Dry-run mode — no files will be changed"
echo ""

UPDATED=0
SKIPPED=0
CURRENT=0

while IFS= read -r d; do
  [[ -d "$d/.git" ]] || continue
  [[ -d "$d/.github/workflows" ]] || continue
  name=$(basename "$d")
  repo_is_selected "$name" || continue

  # Skip repos with custom CI (ralphglasses has 260L custom CI)
  [[ "$name" == "ralphglasses" ]] && continue

  REPO_CHANGES=false

  for wf in "${!CANONICAL[@]}"; do
    target="$d/.github/workflows/$wf"
    source="${CANONICAL[$wf]}"

    [[ -f "$source" ]] || continue

    if [[ ! -f "$target" ]]; then
      if ! $ENSURE_MISSING; then
        continue
      fi

      if [[ "$wf" == "ci.yml" || "$wf" == "dependabot-auto-merge.yml" ]]; then
        continue
      fi

      if $DRY_RUN; then
        printf "%s%-25s %s (would create)%s\n" "$HG_YELLOW" "$name" "$wf" "$HG_RESET"
      else
        mkdir -p "$d/.github/workflows"
        command cp -f "$source" "$target"
        printf "%s%-25s %s (created)%s\n" "$HG_GREEN" "$name" "$wf" "$HG_RESET"
        REPO_CHANGES=true
      fi
      UPDATED=$((UPDATED + 1))
      continue
    fi

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
done < <(find "$STUDIO" -mindepth 1 -maxdepth 1 -type d | sort)

echo ""
if $DRY_RUN; then
  hg_info "$UPDATED workflows would be updated, $CURRENT already current"
else
  hg_ok "$UPDATED workflows updated, $CURRENT already current"
fi
