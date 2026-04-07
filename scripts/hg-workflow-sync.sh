#!/usr/bin/env bash
# hg-workflow-sync.sh — Sync shared workflow files across managed repos.
# Usage: hg-workflow-sync.sh [--dry-run] [--commit] [--push] [--ensure-missing] [--repos=a,b,c] [--include-compatibility] [--include-deprecated]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/hg-workspace.sh"

DOTFILES="$HG_DOTFILES"
ORG_GITHUB="$HG_STUDIO_ROOT/.github"

DRY_RUN=false
COMMIT=false
PUSH=false
ENSURE_MISSING=false
INCLUDE_COMPATIBILITY=false
INCLUDE_DEPRECATED=false
REPO_FILTER=""

for arg in "$@"; do
  case "$arg" in
    --dry-run) DRY_RUN=true ;;
    --commit)  COMMIT=true ;;
    --push)    PUSH=true; COMMIT=true ;;
    --ensure-missing) ENSURE_MISSING=true ;;
    --include-compatibility) INCLUDE_COMPATIBILITY=true ;;
    --include-deprecated) INCLUDE_DEPRECATED=true ;;
    --repos=*) REPO_FILTER="${arg#--repos=}" ;;
  esac
done

workflow_source() {
  local wf="$1"
  if [[ -f "$ORG_GITHUB/workflow-templates/$wf" ]]; then
    printf '%s\n' "$ORG_GITHUB/workflow-templates/$wf"
  else
    printf '%s\n' "$ORG_GITHUB/.github/workflows/$wf"
  fi
}

repo_names() {
  local selected=()
  hg_workspace_parse_repo_filter "$REPO_FILTER" selected
  if [[ "${#selected[@]}" -gt 0 ]]; then
    printf '%s\n' "${selected[@]}"
    return
  fi

  local jq_filter='.repos[]'
  if ! $INCLUDE_COMPATIBILITY; then
    jq_filter+=' | select((.lifecycle // "canonical") != "compatibility")'
  fi
  if ! $INCLUDE_DEPRECATED; then
    jq_filter+=' | select((.lifecycle // "canonical") != "deprecated")'
  fi
  jq_filter+=' | select((.ci_profile // "none") != "none" or (.review_profile // "none") != "none")'
  hg_workspace_repo_names "$jq_filter"
}

ci_source_for_profile() {
  local profile="$1"
  case "$profile" in
    go_standard) printf '%s\n' "$ORG_GITHUB/workflow-templates/ci-go.yml" ;;
    node_standard) printf '%s\n' "$ORG_GITHUB/workflow-templates/ci-node.yml" ;;
    python_standard) printf '%s\n' "$ORG_GITHUB/workflow-templates/ci-python.yml" ;;
    *) printf '%s\n' "" ;;
  esac
}

should_manage_workflow() {
  local workflow="$1"
  local review_profile="$2"
  local ci_profile="$3"

  case "$workflow" in
    ci.yml)
      [[ "$ci_profile" != "none" && "$ci_profile" != "go_custom" ]]
      ;;
    *)
      [[ "$review_profile" != "none" ]]
      ;;
  esac
}

source_for_workflow() {
  local workflow="$1"
  local ci_profile="$2"
  case "$workflow" in
    ci.yml) ci_source_for_profile "$ci_profile" ;;
    dependabot-auto-merge.yml) printf '%s\n' "$HG_STUDIO_ROOT/mcpkit/.github/workflows/$workflow" ;;
    *) workflow_source "$workflow" ;;
  esac
}

COMMON_WORKFLOWS=(
  claude-review.yml
  claude-security.yml
  codex-review.yml
  codex-security.yml
  codex-structured-audit.yml
  codex-baseline-guard.yml
  ai-dispatch.yml
  dependabot-auto-merge.yml
)

hg_info "Workflow sync — comparing against manifest-backed profiles"
$DRY_RUN && hg_warn "Dry-run mode — no files will be changed"
echo ""

UPDATED=0
CURRENT=0

while IFS= read -r name; do
  [[ -n "$name" ]] || continue
  repo_path="$(hg_workspace_repo_path "$name")"
  [[ -d "$repo_path/.git" ]] || continue

  review_profile="$(hg_workspace_repo_field "$name" "review_profile" "none")"
  ci_profile="$(hg_workspace_repo_field "$name" "ci_profile" "none")"

  REPO_CHANGES=false

  for wf in "${COMMON_WORKFLOWS[@]}" ci.yml; do
    should_manage_workflow "$wf" "$review_profile" "$ci_profile" || continue

    source_file="$(source_for_workflow "$wf" "$ci_profile")"
    [[ -f "$source_file" ]] || continue

    target="$repo_path/.github/workflows/$wf"
    if [[ ! -f "$target" ]]; then
      if ! $ENSURE_MISSING; then
        continue
      fi

      if $DRY_RUN; then
        printf "%s%-25s %s (would create)%s\n" "$HG_YELLOW" "$name" "$wf" "$HG_RESET"
      else
        mkdir -p "$repo_path/.github/workflows"
        command cp -f "$source_file" "$target"
        printf "%s%-25s %s (created)%s\n" "$HG_GREEN" "$name" "$wf" "$HG_RESET"
        REPO_CHANGES=true
      fi
      UPDATED=$((UPDATED + 1))
      continue
    fi

    if ! diff -q "$source_file" "$target" &>/dev/null; then
      if $DRY_RUN; then
        printf "%s%-25s %s (would update)%s\n" "$HG_YELLOW" "$name" "$wf" "$HG_RESET"
      else
        command cp -f "$source_file" "$target"
        printf "%s%-25s %s (updated)%s\n" "$HG_GREEN" "$name" "$wf" "$HG_RESET"
        REPO_CHANGES=true
      fi
      UPDATED=$((UPDATED + 1))
    else
      CURRENT=$((CURRENT + 1))
    fi
  done

  if $REPO_CHANGES && $COMMIT; then
    (
      cd "$repo_path"
      git add .github/workflows/ 2>/dev/null
      if ! git diff --cached --quiet 2>/dev/null; then
        git commit -q -m "ci: sync workflows via hg-workflow-sync.sh"
        if $PUSH; then
          git pull --rebase -q 2>/dev/null
          git push -q 2>&1 && printf "%s  → pushed%s\n" "$HG_DIM" "$HG_RESET" || printf "%s  → push failed%s\n" "$HG_RED" "$HG_RESET"
        fi
      fi
    )
  fi
done < <(repo_names)

echo ""
if $DRY_RUN; then
  hg_info "$UPDATED workflows would be updated, $CURRENT already current"
else
  hg_ok "$UPDATED workflows updated, $CURRENT already current"
fi
