#!/usr/bin/env bash
set -euo pipefail

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

MODE="plan"
JSON_MODE=false
DIFF_PREVIEW=false
DIFF_PREVIEW_LINES=20
REFRESH_BARE_ORIGIN=false
REFRESHED_BARE_ORIGIN=false
CANONICAL_PATH="${HG_DOTFILES}/mcp/dotfiles-mcp"
STANDALONE_PATH="${HG_STUDIO_ROOT}/dotfiles-mcp"
TEMP_WORKTREE=""
TEMP_BASE=""
STANDALONE_INSPECT_ROOT=""
APPLY_WORKTREE_OVERRIDE="${HG_DOTFILES_MCP_APPLY_WORKTREE:-}"

usage() {
  cat <<'EOF'
Usage: hg-dotfiles-mcp-projection.sh [plan|check|apply] [--canonical PATH] [--standalone PATH] [--json] [--diff-preview] [--diff-lines N] [--refresh-bare-origin]

Inspect the bundled dotfiles MCP module and the standalone dotfiles-mcp publish
repo, then emit or apply the current projection plan and drift summary.

Modes:
  plan    Print the projection plan (default)
  check   Validate that the projection plan can be resolved and inspected
  apply   Apply required projection changes into an editable standalone checkout,
          then re-run plan mode to confirm required drift is cleared

Options:
  --canonical PATH   Override the bundled canonical module path
  --standalone PATH  Override the standalone repo path (bare or non-bare)
  --json             Emit machine-readable JSON
  --diff-preview     Include drift previews for overlap and direct-copy drift
  --diff-lines N     Limit each drift preview to N lines (default: 20)
  --refresh-bare-origin
                     Refresh refs/remotes/origin/main before inspecting a bare
                     standalone mirror
  -h, --help         Show this help text
EOF
}

cleanup() {
  if [[ -n "$TEMP_WORKTREE" && -n "$TEMP_BASE" && -d "$TEMP_BASE" ]]; then
    git -C "$STANDALONE_PATH" worktree remove --force "$TEMP_WORKTREE" >/dev/null 2>&1 || true
    rm -rf "$TEMP_BASE"
  fi
}

trap cleanup EXIT

while [[ $# -gt 0 ]]; do
  case "$1" in
    plan|check|apply)
      MODE="$1"
      shift
      ;;
    --canonical)
      CANONICAL_PATH="$2"
      shift 2
      ;;
    --standalone)
      STANDALONE_PATH="$2"
      shift 2
      ;;
    --json)
      JSON_MODE=true
      shift
      ;;
    --diff-preview)
      DIFF_PREVIEW=true
      shift
      ;;
    --diff-lines)
      DIFF_PREVIEW=true
      DIFF_PREVIEW_LINES="$2"
      shift 2
      ;;
    --refresh-bare-origin)
      REFRESH_BARE_ORIGIN=true
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      hg_die "Unknown argument: $1"
      ;;
  esac
done

hg_require diff git jq perl
if [[ "$MODE" == "apply" ]]; then
  hg_require gofmt
fi

resolve_path_candidate() {
  local candidate="${1:-}"
  [[ -n "$candidate" ]] || return 1
  [[ -e "$candidate" ]] || return 1
  (cd "$candidate" && pwd)
}

resolve_standalone_repo_path() {
  local requested="$1"
  local canonical="$2"
  local candidate
  local -a candidates=("$requested")

  if [[ "$canonical" =~ ^(/home/[^/]+)/hairglasses-studio/dotfiles/ ]]; then
    candidates+=("${BASH_REMATCH[1]}/hairglasses-studio/dotfiles-mcp")
  fi
  candidates+=("/home/hg/hairglasses-studio/dotfiles-mcp")

  for candidate in "${candidates[@]}"; do
    if resolved="$(resolve_path_candidate "$candidate")"; then
      printf '%s\n' "$resolved"
      return 0
    fi
  done

  hg_die "Standalone path not found: $requested"
}

CANONICAL_PATH="$(resolve_path_candidate "$CANONICAL_PATH")" || hg_die "Canonical path not found: $CANONICAL_PATH"
STANDALONE_PATH="$(resolve_standalone_repo_path "$STANDALONE_PATH" "$CANONICAL_PATH")"

[[ -f "$CANONICAL_PATH/go.mod" ]] || hg_die "Canonical module missing go.mod: $CANONICAL_PATH"
[[ "$DIFF_PREVIEW_LINES" =~ ^[1-9][0-9]*$ ]] || hg_die "--diff-lines must be a positive integer"

resolve_standalone_inspect_root() {
  local root="$1"
  local worktree_path
  local repo_name
  local is_bare="false"

  if git -C "$root" rev-parse --is-bare-repository >/dev/null 2>&1; then
    is_bare="$(git -C "$root" rev-parse --is-bare-repository)"
  fi

  if [[ "$MODE" == "apply" ]]; then
    if [[ "$is_bare" != "true" && -d "$root/internal/dotfiles" ]]; then
      STANDALONE_INSPECT_ROOT="$root"
      return 0
    fi

    if [[ -n "$APPLY_WORKTREE_OVERRIDE" && -d "$APPLY_WORKTREE_OVERRIDE/internal/dotfiles" ]]; then
      STANDALONE_INSPECT_ROOT="$(cd "$APPLY_WORKTREE_OVERRIDE" && pwd)"
      return 0
    fi

    if [[ "$is_bare" == "true" ]]; then
      while IFS= read -r worktree_path; do
        [[ -n "$worktree_path" ]] || continue
        if [[ -d "$worktree_path/internal/dotfiles" ]]; then
          STANDALONE_INSPECT_ROOT="$(cd "$worktree_path" && pwd)"
          return 0
        fi
      done < <(git -C "$root" worktree list --porcelain 2>/dev/null | awk '/^worktree /{sub(/^worktree /, "", $0); print $0}')
    fi

    repo_name="$(basename "$root")"
    if [[ -d "/tmp/${repo_name}-real/internal/dotfiles" ]]; then
      STANDALONE_INSPECT_ROOT="$(cd "/tmp/${repo_name}-real" && pwd)"
      return 0
    fi

    hg_die "Apply mode requires an editable standalone checkout with internal/dotfiles; pass --standalone CHECKOUT, set HG_DOTFILES_MCP_APPLY_WORKTREE, or create /tmp/${repo_name}-real"
  fi

  if [[ "$is_bare" == "true" ]]; then
      if $REFRESH_BARE_ORIGIN; then
        if git -C "$root" remote get-url origin >/dev/null 2>&1; then
          if git -C "$root" fetch --prune origin '+refs/heads/main:refs/remotes/origin/main' >/dev/null 2>&1; then
            REFRESHED_BARE_ORIGIN=true
          else
            hg_warn "Failed to refresh refs/remotes/origin/main for bare standalone repo: $root"
          fi
        else
          hg_warn "Bare standalone repo has no origin remote; cannot refresh origin/main: $root"
        fi
      fi
      local inspect_ref="HEAD"
      if git -C "$root" rev-parse --verify refs/remotes/origin/main >/dev/null 2>&1; then
        inspect_ref="refs/remotes/origin/main"
      elif git -C "$root" rev-parse --verify refs/heads/main >/dev/null 2>&1; then
        inspect_ref="refs/heads/main"
      fi
      TEMP_BASE="$(mktemp -d "${TMPDIR:-/tmp}/dotfiles-mcp-projection.XXXXXX")"
      TEMP_WORKTREE="${TEMP_BASE}/worktree"
      git -C "$root" worktree add --quiet --detach "$TEMP_WORKTREE" "$inspect_ref" >/dev/null
      STANDALONE_INSPECT_ROOT="$TEMP_WORKTREE"
      return 0
  fi

  if [[ -d "$root/internal/dotfiles" ]]; then
    STANDALONE_INSPECT_ROOT="$root"
    return 0
  fi

  local top
  top="$(git -C "$root" rev-parse --show-toplevel 2>/dev/null || true)"
  if [[ -n "$top" && -d "$top/internal/dotfiles" ]]; then
    STANDALONE_INSPECT_ROOT="$top"
    return 0
  fi

  hg_die "Standalone repo does not expose internal/dotfiles: $root"
}

temp_file() {
  mktemp "${TMPDIR:-/tmp}/dotfiles-mcp-projection-data.XXXXXX"
}

count_lines() {
  if [[ -s "$1" ]]; then
    wc -l <"$1" | tr -d ' '
  else
    printf '0\n'
  fi
}

json_array_file() {
  jq -Rsc 'split("\n") | map(select(length > 0))' "$1"
}

json_objects_file() {
  if [[ -s "$1" ]]; then
    jq -cs '.' "$1"
  else
    printf '[]\n'
  fi
}

append_diff_preview() {
  local rel="$1"
  local left="$2"
  local right="$3"
  local kind="$4"
  local outfile="$5"
  local preview
  local numstat
  local added="0"
  local deleted="0"

  $DIFF_PREVIEW || return 0

  preview="$(
    diff -u \
      --label "canonical/$rel" \
      --label "standalone/$rel" \
      "$left" \
      "$right" 2>/dev/null | awk -v limit="$DIFF_PREVIEW_LINES" '
        NR <= 2 { print; next }
        /^@@ / { in_hunk = 1 }
        {
          if (!in_hunk) {
            next
          }
          if (body_lines < limit) {
            print
            body_lines++
          }
        }
      ' || true
  )"
  numstat="$(git --no-pager diff --no-index --numstat -- "$left" "$right" 2>/dev/null || true)"

  if [[ -n "$numstat" ]]; then
    added="$(awk 'NR==1 {print ($1 == "-" ? 0 : $1)}' <<<"$numstat")"
    deleted="$(awk 'NR==1 {print ($2 == "-" ? 0 : $2)}' <<<"$numstat")"
  fi

  jq -cn \
    --arg path "$rel" \
    --arg kind "$kind" \
    --arg preview "$preview" \
    --argjson added "${added:-0}" \
    --argjson deleted "${deleted:-0}" \
    '{path: $path, kind: $kind, added: $added, deleted: $deleted, preview: $preview}' >>"$outfile"
  printf '\n' >>"$outfile"
}

apply_direct_copy_file() {
  local rel="$1"
  local target="$STANDALONE_INSPECT_ROOT/$rel"
  mkdir -p "$(dirname "$target")"
  cp "$CANONICAL_PATH/$rel" "$target"
}

apply_projected_go_file() {
  local file="$1"
  local target="$TARGET_PACKAGE_DIR/$file"
  mkdir -p "$(dirname "$target")"
  perl -0pe 's/^package main$/package dotfiles/m' "$CANONICAL_PATH/$file" >"$target"
  gofmt -w "$target"
}

apply_internal_package_dir() {
  local pkg="$1"
  local source="$CANONICAL_PATH/internal/$pkg"
  local target="$STANDALONE_INSPECT_ROOT/internal/$pkg"
  rm -rf "$target"
  mkdir -p "$(dirname "$target")"
  cp -a "$source" "$target"
}

resolve_standalone_inspect_root "$STANDALONE_PATH"
TARGET_PACKAGE_DIR="${STANDALONE_INSPECT_ROOT}/internal/dotfiles"

CANONICAL_GO_LIST="$(temp_file)"
TARGET_GO_LIST="$(temp_file)"
GO_IDENTICAL_LIST="$(temp_file)"
GO_DRIFTED_LIST="$(temp_file)"
GO_DRIFTED_REQUIRED_LIST="$(temp_file)"
GO_DRIFTED_INTENTIONAL_LIST="$(temp_file)"
GO_DRIFTED_REVIEW_LIST="$(temp_file)"
GO_CANONICAL_ONLY_LIST="$(temp_file)"
GO_CANONICAL_ONLY_REQUIRED_LIST="$(temp_file)"
GO_CANONICAL_ONLY_INTENTIONAL_LIST="$(temp_file)"
GO_TARGET_ONLY_LIST="$(temp_file)"
INTERNAL_IMPORT_LIST="$(temp_file)"
INTERNAL_PRESENT_LIST="$(temp_file)"
INTERNAL_IDENTICAL_LIST="$(temp_file)"
INTERNAL_DRIFTED_LIST="$(temp_file)"
INTERNAL_DRIFTED_REQUIRED_LIST="$(temp_file)"
INTERNAL_DRIFTED_INTENTIONAL_LIST="$(temp_file)"
INTERNAL_DRIFTED_REVIEW_LIST="$(temp_file)"
INTERNAL_MISSING_LIST="$(temp_file)"
INTERNAL_MISSING_REQUIRED_LIST="$(temp_file)"
INTERNAL_MISSING_INTENTIONAL_LIST="$(temp_file)"
COPY_ALLOWLIST_PRESENT_LIST="$(temp_file)"
COPY_PRESENT_LIST="$(temp_file)"
COPY_IDENTICAL_LIST="$(temp_file)"
COPY_DRIFTED_LIST="$(temp_file)"
COPY_DRIFTED_REQUIRED_LIST="$(temp_file)"
COPY_DRIFTED_INTENTIONAL_LIST="$(temp_file)"
COPY_DRIFTED_REVIEW_LIST="$(temp_file)"
COPY_MISSING_LIST="$(temp_file)"
COPY_MISSING_REQUIRED_LIST="$(temp_file)"
COPY_MISSING_INTENTIONAL_LIST="$(temp_file)"
COPY_DRIFT_PREVIEW_JSON="$(temp_file)"
STANDALONE_ROOT_ONLY_LIST="$(temp_file)"
STANDALONE_INTERNAL_ONLY_LIST="$(temp_file)"
GO_DRIFT_PREVIEW_JSON="$(temp_file)"

find "$CANONICAL_PATH" -maxdepth 1 -name '*.go' -printf '%f\n' | sort >"$CANONICAL_GO_LIST"
find "$TARGET_PACKAGE_DIR" -maxdepth 1 -name '*.go' -printf '%f\n' | sort >"$TARGET_GO_LIST"
shopt -s nullglob
canonical_source_files=("$CANONICAL_PATH"/*.go "$CANONICAL_PATH"/*_test.go)
shopt -u nullglob
if [[ "${#canonical_source_files[@]}" -gt 0 ]]; then
  { grep -rhoE 'github.com/hairglasses-studio/dotfiles-mcp/internal/[A-Za-z0-9_]+' "${canonical_source_files[@]}" || true; } \
    | sed 's#.*internal/##' \
    | sort -u >"$INTERNAL_IMPORT_LIST"
else
  : >"$INTERNAL_IMPORT_LIST"
fi

comm -23 "$CANONICAL_GO_LIST" "$TARGET_GO_LIST" >"$GO_CANONICAL_ONLY_LIST"
comm -13 "$CANONICAL_GO_LIST" "$TARGET_GO_LIST" >"$GO_TARGET_ONLY_LIST"

intentional_canonical_only=(
  "contract_snapshot_cli.go"
  "workflow_surface_test.go"
)

for file in "${intentional_canonical_only[@]}"; do
  if grep -Fxq "$file" "$GO_CANONICAL_ONLY_LIST"; then
    printf '%s\n' "$file" >>"$GO_CANONICAL_ONLY_INTENTIONAL_LIST"
  fi
done

while IFS= read -r file; do
  [[ -n "$file" ]] || continue
  if ! grep -Fxq "$file" "$GO_CANONICAL_ONLY_INTENTIONAL_LIST"; then
    printf '%s\n' "$file" >>"$GO_CANONICAL_ONLY_REQUIRED_LIST"
  fi
done <"$GO_CANONICAL_ONLY_LIST"

while IFS= read -r file; do
  [[ -n "$file" ]] || continue
  tmp_rewritten="$(temp_file)"
  perl -0pe 's/^package main$/package dotfiles/m' "$CANONICAL_PATH/$file" >"$tmp_rewritten"
  if cmp -s "$tmp_rewritten" "$TARGET_PACKAGE_DIR/$file"; then
    printf '%s\n' "$file" >>"$GO_IDENTICAL_LIST"
  else
    printf '%s\n' "$file" >>"$GO_DRIFTED_LIST"
    append_diff_preview "$file" "$tmp_rewritten" "$TARGET_PACKAGE_DIR/$file" "go_overlap" "$GO_DRIFT_PREVIEW_JSON"
  fi
  rm -f "$tmp_rewritten"
done < <(comm -12 "$CANONICAL_GO_LIST" "$TARGET_GO_LIST")

intentional_go_drift=(
  "contract_snapshot.go"
  "contract_snapshot_test.go"
  "main.go"
  "mod_claude_session.go"
  "mod_claude_session_helpers_test.go"
  "mod_clipboard.go"
  "mod_desktop_interact.go"
  "mod_github_test.go"
  "mod_hyprland_helpers_test.go"
  "mod_input_simulate.go"
  "mod_learn.go"
  "mod_mapping_daemon.go"
  "mod_ops.go"
  "mod_ops_test.go"
  "mod_process.go"
  "mod_screen.go"
  "mod_system.go"
  "mod_tmux.go"
  "oss.go"
  "oss_test.go"
)

review_go_drift=(
  "context.go"
  "contract_profile_surface_test.go"
  "discovery_test.go"
  "discovery_workstation_diagnostics.go"
  "discovery_workstation_diagnostics_test.go"
  "mod_desktop_semantic.go"
  "mod_desktop_semantic_session_test.go"
  "mod_desktop_session.go"
  "mod_expansion_test.go"
  "workflow_catalog.go"
)

for file in "${intentional_go_drift[@]}"; do
  if grep -Fxq "$file" "$GO_DRIFTED_LIST"; then
    printf '%s\n' "$file" >>"$GO_DRIFTED_INTENTIONAL_LIST"
  fi
done

for file in "${review_go_drift[@]}"; do
  if grep -Fxq "$file" "$GO_DRIFTED_LIST"; then
    printf '%s\n' "$file" >>"$GO_DRIFTED_REVIEW_LIST"
  fi
done

while IFS= read -r file; do
  [[ -n "$file" ]] || continue
  if ! grep -Fxq "$file" "$GO_DRIFTED_INTENTIONAL_LIST" && ! grep -Fxq "$file" "$GO_DRIFTED_REVIEW_LIST"; then
    printf '%s\n' "$file" >>"$GO_DRIFTED_REQUIRED_LIST"
  fi
done <"$GO_DRIFTED_LIST"

intentional_internal_packages=(
  "githubstars"
  "mapping"
  "tracing"
)

review_internal_packages=()

while IFS= read -r pkg; do
  [[ -n "$pkg" ]] || continue
  canonical_dir="$CANONICAL_PATH/internal/$pkg"
  target_dir="$STANDALONE_INSPECT_ROOT/internal/$pkg"
  [[ -d "$canonical_dir" ]] || continue

  if [[ -d "$target_dir" ]]; then
    printf '%s\n' "$pkg" >>"$INTERNAL_PRESENT_LIST"
    if diff -qr "$canonical_dir" "$target_dir" >/dev/null 2>&1; then
      printf '%s\n' "$pkg" >>"$INTERNAL_IDENTICAL_LIST"
    else
      printf '%s\n' "$pkg" >>"$INTERNAL_DRIFTED_LIST"
    fi
  else
    printf '%s\n' "$pkg" >>"$INTERNAL_MISSING_LIST"
  fi
done <"$INTERNAL_IMPORT_LIST"

for pkg in "${intentional_internal_packages[@]}"; do
  if grep -Fxq "$pkg" "$INTERNAL_DRIFTED_LIST"; then
    printf '%s\n' "$pkg" >>"$INTERNAL_DRIFTED_INTENTIONAL_LIST"
  fi
  if grep -Fxq "$pkg" "$INTERNAL_MISSING_LIST"; then
    printf '%s\n' "$pkg" >>"$INTERNAL_MISSING_INTENTIONAL_LIST"
  fi
done

for pkg in "${review_internal_packages[@]}"; do
  if grep -Fxq "$pkg" "$INTERNAL_DRIFTED_LIST"; then
    printf '%s\n' "$pkg" >>"$INTERNAL_DRIFTED_REVIEW_LIST"
  fi
done

while IFS= read -r pkg; do
  [[ -n "$pkg" ]] || continue
  if ! grep -Fxq "$pkg" "$INTERNAL_DRIFTED_INTENTIONAL_LIST" && ! grep -Fxq "$pkg" "$INTERNAL_DRIFTED_REVIEW_LIST"; then
    printf '%s\n' "$pkg" >>"$INTERNAL_DRIFTED_REQUIRED_LIST"
  fi
done <"$INTERNAL_DRIFTED_LIST"

while IFS= read -r pkg; do
  [[ -n "$pkg" ]] || continue
  if ! grep -Fxq "$pkg" "$INTERNAL_MISSING_INTENTIONAL_LIST"; then
    printf '%s\n' "$pkg" >>"$INTERNAL_MISSING_REQUIRED_LIST"
  fi
done <"$INTERNAL_MISSING_LIST"

copy_allowlist=(
  ".github/PULL_REQUEST_TEMPLATE.md"
  ".github/copilot-instructions.md"
  ".github/pull_request_template.md"
  ".github/workflows/ci.yml"
  ".github/workflows/publish-guard.yml"
  ".github/workflows/release.yml"
  ".github/workflows/server-card-validate.yml"
  ".gitignore"
  ".golangci-lint.yml"
  ".goreleaser.yaml"
  ".goreleaser.yml"
  ".pre-commit-config.yaml"
  ".prompt-improver.yaml"
  ".ralphrc"
  ".well-known/mcp.json"
  "AGENTS.md"
  "CHANGELOG.md"
  "CLAUDE.md"
  "CODE_OF_CONDUCT.md"
  "CONTRIBUTING.md"
  "GEMINI.md"
  "LICENSE"
  "Makefile"
  "README.md"
  "REVIEW.md"
  "ROADMAP.md"
  "SECURITY.md"
  "scripts/host-smoke.sh"
  "scripts/release-parity.sh"
)

intentional_copy_drift=(
  ".github/copilot-instructions.md"
  ".gitignore"
  ".goreleaser.yaml"
  ".well-known/mcp.json"
  "AGENTS.md"
  "CHANGELOG.md"
  "CLAUDE.md"
  "CONTRIBUTING.md"
  "GEMINI.md"
  "Makefile"
  "README.md"
  "REVIEW.md"
  "ROADMAP.md"
  "scripts/host-smoke.sh"
  "scripts/release-parity.sh"
)

intentional_copy_missing=(
  ".goreleaser.yml"
)

review_copy_drift=()

for rel in "${copy_allowlist[@]}"; do
  if [[ -f "$CANONICAL_PATH/$rel" ]]; then
    printf '%s\n' "$rel" >>"$COPY_ALLOWLIST_PRESENT_LIST"
    if [[ -f "$STANDALONE_INSPECT_ROOT/$rel" ]]; then
      printf '%s\n' "$rel" >>"$COPY_PRESENT_LIST"
      if cmp -s "$CANONICAL_PATH/$rel" "$STANDALONE_INSPECT_ROOT/$rel"; then
        printf '%s\n' "$rel" >>"$COPY_IDENTICAL_LIST"
      else
        printf '%s\n' "$rel" >>"$COPY_DRIFTED_LIST"
        append_diff_preview "$rel" "$CANONICAL_PATH/$rel" "$STANDALONE_INSPECT_ROOT/$rel" "direct_copy" "$COPY_DRIFT_PREVIEW_JSON"
      fi
    else
      printf '%s\n' "$rel" >>"$COPY_MISSING_LIST"
    fi
  fi
done

for rel in "${intentional_copy_drift[@]}"; do
  if grep -Fxq "$rel" "$COPY_DRIFTED_LIST"; then
    printf '%s\n' "$rel" >>"$COPY_DRIFTED_INTENTIONAL_LIST"
  fi
done

for rel in "${review_copy_drift[@]}"; do
  if grep -Fxq "$rel" "$COPY_DRIFTED_LIST"; then
    printf '%s\n' "$rel" >>"$COPY_DRIFTED_REVIEW_LIST"
  fi
done

while IFS= read -r rel; do
  [[ -n "$rel" ]] || continue
  if ! grep -Fxq "$rel" "$COPY_DRIFTED_INTENTIONAL_LIST" && ! grep -Fxq "$rel" "$COPY_DRIFTED_REVIEW_LIST"; then
    printf '%s\n' "$rel" >>"$COPY_DRIFTED_REQUIRED_LIST"
  fi
done <"$COPY_DRIFTED_LIST"

for rel in "${intentional_copy_missing[@]}"; do
  if grep -Fxq "$rel" "$COPY_MISSING_LIST"; then
    printf '%s\n' "$rel" >>"$COPY_MISSING_INTENTIONAL_LIST"
  fi
done

while IFS= read -r rel; do
  [[ -n "$rel" ]] || continue
  if ! grep -Fxq "$rel" "$COPY_MISSING_INTENTIONAL_LIST"; then
    printf '%s\n' "$rel" >>"$COPY_MISSING_REQUIRED_LIST"
  fi
done <"$COPY_MISSING_LIST"

find "$STANDALONE_INSPECT_ROOT" -mindepth 1 -maxdepth 1 -printf '%f\n' | sort | while IFS= read -r entry; do
  [[ -e "$CANONICAL_PATH/$entry" ]] || printf '%s\n' "$entry"
done >"$STANDALONE_ROOT_ONLY_LIST"

if [[ -d "$STANDALONE_INSPECT_ROOT/internal" ]]; then
  find "$STANDALONE_INSPECT_ROOT/internal" -mindepth 1 -maxdepth 1 -type d -printf 'internal/%f\n' | sort | while IFS= read -r entry; do
    [[ "$entry" == "internal/dotfiles" ]] || printf '%s\n' "$entry"
  done >"$STANDALONE_INTERNAL_ONLY_LIST"
fi

plan_status="in_sync"
if [[ "$(count_lines "$GO_CANONICAL_ONLY_REQUIRED_LIST")" -gt 0 || "$(count_lines "$GO_DRIFTED_REQUIRED_LIST")" -gt 0 || "$(count_lines "$COPY_DRIFTED_REQUIRED_LIST")" -gt 0 || "$(count_lines "$COPY_MISSING_REQUIRED_LIST")" -gt 0 || "$(count_lines "$INTERNAL_DRIFTED_REQUIRED_LIST")" -gt 0 || "$(count_lines "$INTERNAL_MISSING_REQUIRED_LIST")" -gt 0 ]]; then
  plan_status="projection_needed"
elif [[ "$(count_lines "$GO_DRIFTED_REVIEW_LIST")" -gt 0 || "$(count_lines "$COPY_DRIFTED_REVIEW_LIST")" -gt 0 || "$(count_lines "$INTERNAL_DRIFTED_REVIEW_LIST")" -gt 0 ]]; then
  plan_status="parity_review_needed"
fi

if [[ "$MODE" == "apply" ]]; then
  APPLIED_LOG="$(temp_file)"

  while IFS= read -r rel; do
    [[ -n "$rel" ]] || continue
    apply_direct_copy_file "$rel"
    printf 'copy %s\n' "$rel" >>"$APPLIED_LOG"
  done < <(
    cat "$COPY_DRIFTED_REQUIRED_LIST" "$COPY_MISSING_REQUIRED_LIST" 2>/dev/null | awk 'NF {print}' | sort -u
  )

  while IFS= read -r file; do
    [[ -n "$file" ]] || continue
    apply_projected_go_file "$file"
    printf 'project %s\n' "$file" >>"$APPLIED_LOG"
  done < <(
    cat "$GO_DRIFTED_REQUIRED_LIST" "$GO_CANONICAL_ONLY_REQUIRED_LIST" 2>/dev/null | awk 'NF {print}' | sort -u
  )

  while IFS= read -r pkg; do
    [[ -n "$pkg" ]] || continue
    apply_internal_package_dir "$pkg"
    printf 'package %s\n' "$pkg" >>"$APPLIED_LOG"
  done < <(
    cat "$INTERNAL_DRIFTED_REQUIRED_LIST" "$INTERNAL_MISSING_REQUIRED_LIST" 2>/dev/null | awk 'NF {print}' | sort -u
  )

  if ! $JSON_MODE; then
    printf 'dotfiles-mcp projection apply\n'
    printf 'standalone root     %s\n' "$STANDALONE_INSPECT_ROOT"
    printf 'applied changes\n'
    if [[ -s "$APPLIED_LOG" ]]; then
      sed 's/^/  - /' "$APPLIED_LOG"
    else
      printf '  - none (already in sync)\n'
    fi
    printf '\n'
  fi

  reexec_args=(
    plan
    --canonical "$CANONICAL_PATH"
    --standalone "$STANDALONE_INSPECT_ROOT"
  )
  if $JSON_MODE; then
    reexec_args+=(--json)
  fi
  if $DIFF_PREVIEW; then
    reexec_args+=(--diff-preview --diff-lines "$DIFF_PREVIEW_LINES")
  fi
  exec bash "$SCRIPT_PATH" "${reexec_args[@]}"
fi

inspect_mode="worktree"
if [[ -n "$TEMP_WORKTREE" ]]; then
  inspect_mode="bare_repo_ephemeral_worktree"
fi

if $JSON_MODE; then
  jq -n \
    --arg mode "$MODE" \
    --arg status "$plan_status" \
    --arg canonical_path "$CANONICAL_PATH" \
    --arg standalone_repo "$STANDALONE_PATH" \
    --arg inspect_root "$STANDALONE_INSPECT_ROOT" \
    --arg inspect_mode "$inspect_mode" \
    --arg target_package_dir "$TARGET_PACKAGE_DIR" \
    --arg main_projection_note "Project bundled root Go files into internal/dotfiles with package main rewritten to package dotfiles; preserve standalone cmd/* entrypoints and the Setup-based bootstrap in internal/dotfiles/main.go." \
    --argjson refreshed_bare_origin "$REFRESHED_BARE_ORIGIN" \
    --argjson diff_preview_enabled "$DIFF_PREVIEW" \
    --argjson diff_preview_lines "$DIFF_PREVIEW_LINES" \
    --argjson canonical_go_count "$(count_lines "$CANONICAL_GO_LIST")" \
    --argjson target_go_count "$(count_lines "$TARGET_GO_LIST")" \
    --argjson identical_count "$(count_lines "$GO_IDENTICAL_LIST")" \
    --argjson drifted_count "$(count_lines "$GO_DRIFTED_LIST")" \
    --argjson drifted_required_count "$(count_lines "$GO_DRIFTED_REQUIRED_LIST")" \
    --argjson drifted_intentional_count "$(count_lines "$GO_DRIFTED_INTENTIONAL_LIST")" \
    --argjson drifted_review_count "$(count_lines "$GO_DRIFTED_REVIEW_LIST")" \
    --argjson canonical_only_count "$(count_lines "$GO_CANONICAL_ONLY_LIST")" \
    --argjson projection_required_count "$(count_lines "$GO_CANONICAL_ONLY_REQUIRED_LIST")" \
    --argjson intentional_canonical_only_count "$(count_lines "$GO_CANONICAL_ONLY_INTENTIONAL_LIST")" \
    --argjson target_only_count "$(count_lines "$GO_TARGET_ONLY_LIST")" \
    --argjson internal_import_count "$(count_lines "$INTERNAL_IMPORT_LIST")" \
    --argjson internal_present_count "$(count_lines "$INTERNAL_PRESENT_LIST")" \
    --argjson internal_identical_count "$(count_lines "$INTERNAL_IDENTICAL_LIST")" \
    --argjson internal_drifted_count "$(count_lines "$INTERNAL_DRIFTED_LIST")" \
    --argjson internal_required_drift_count "$(count_lines "$INTERNAL_DRIFTED_REQUIRED_LIST")" \
    --argjson internal_intentional_drift_count "$(count_lines "$INTERNAL_DRIFTED_INTENTIONAL_LIST")" \
    --argjson internal_review_drift_count "$(count_lines "$INTERNAL_DRIFTED_REVIEW_LIST")" \
    --argjson internal_missing_count "$(count_lines "$INTERNAL_MISSING_LIST")" \
    --argjson internal_required_missing_count "$(count_lines "$INTERNAL_MISSING_REQUIRED_LIST")" \
    --argjson internal_intentional_missing_count "$(count_lines "$INTERNAL_MISSING_INTENTIONAL_LIST")" \
    --argjson copy_candidate_count "$(count_lines "$COPY_ALLOWLIST_PRESENT_LIST")" \
    --argjson copy_present_count "$(count_lines "$COPY_PRESENT_LIST")" \
    --argjson copy_identical_count "$(count_lines "$COPY_IDENTICAL_LIST")" \
    --argjson copy_drifted_count "$(count_lines "$COPY_DRIFTED_LIST")" \
    --argjson copy_drifted_required_count "$(count_lines "$COPY_DRIFTED_REQUIRED_LIST")" \
    --argjson copy_drifted_intentional_count "$(count_lines "$COPY_DRIFTED_INTENTIONAL_LIST")" \
    --argjson copy_drifted_review_count "$(count_lines "$COPY_DRIFTED_REVIEW_LIST")" \
    --argjson copy_missing_count "$(count_lines "$COPY_MISSING_LIST")" \
    --argjson copy_missing_required_count "$(count_lines "$COPY_MISSING_REQUIRED_LIST")" \
    --argjson copy_missing_intentional_count "$(count_lines "$COPY_MISSING_INTENTIONAL_LIST")" \
    --argjson copy_candidates "$(json_array_file "$COPY_ALLOWLIST_PRESENT_LIST")" \
    --argjson copy_present "$(json_array_file "$COPY_PRESENT_LIST")" \
    --argjson copy_identical "$(json_array_file "$COPY_IDENTICAL_LIST")" \
    --argjson copy_drifted "$(json_array_file "$COPY_DRIFTED_LIST")" \
    --argjson copy_drifted_required "$(json_array_file "$COPY_DRIFTED_REQUIRED_LIST")" \
    --argjson copy_drifted_intentional "$(json_array_file "$COPY_DRIFTED_INTENTIONAL_LIST")" \
    --argjson copy_drifted_review "$(json_array_file "$COPY_DRIFTED_REVIEW_LIST")" \
    --argjson copy_missing "$(json_array_file "$COPY_MISSING_LIST")" \
    --argjson copy_missing_required "$(json_array_file "$COPY_MISSING_REQUIRED_LIST")" \
    --argjson copy_missing_intentional "$(json_array_file "$COPY_MISSING_INTENTIONAL_LIST")" \
    --argjson copy_drift_previews "$(json_objects_file "$COPY_DRIFT_PREVIEW_JSON")" \
    --argjson go_identical "$(json_array_file "$GO_IDENTICAL_LIST")" \
    --argjson go_drifted "$(json_array_file "$GO_DRIFTED_LIST")" \
    --argjson go_drifted_required "$(json_array_file "$GO_DRIFTED_REQUIRED_LIST")" \
    --argjson go_drifted_intentional "$(json_array_file "$GO_DRIFTED_INTENTIONAL_LIST")" \
    --argjson go_drifted_review "$(json_array_file "$GO_DRIFTED_REVIEW_LIST")" \
    --argjson go_canonical_only "$(json_array_file "$GO_CANONICAL_ONLY_LIST")" \
    --argjson go_projection_required "$(json_array_file "$GO_CANONICAL_ONLY_REQUIRED_LIST")" \
    --argjson go_intentional_canonical_only "$(json_array_file "$GO_CANONICAL_ONLY_INTENTIONAL_LIST")" \
    --argjson go_target_only "$(json_array_file "$GO_TARGET_ONLY_LIST")" \
    --argjson go_drift_previews "$(json_objects_file "$GO_DRIFT_PREVIEW_JSON")" \
    --argjson internal_imports "$(json_array_file "$INTERNAL_IMPORT_LIST")" \
    --argjson internal_present "$(json_array_file "$INTERNAL_PRESENT_LIST")" \
    --argjson internal_identical "$(json_array_file "$INTERNAL_IDENTICAL_LIST")" \
    --argjson internal_drifted "$(json_array_file "$INTERNAL_DRIFTED_LIST")" \
    --argjson internal_required_drift "$(json_array_file "$INTERNAL_DRIFTED_REQUIRED_LIST")" \
    --argjson internal_intentional_drift "$(json_array_file "$INTERNAL_DRIFTED_INTENTIONAL_LIST")" \
    --argjson internal_review_drift "$(json_array_file "$INTERNAL_DRIFTED_REVIEW_LIST")" \
    --argjson internal_missing "$(json_array_file "$INTERNAL_MISSING_LIST")" \
    --argjson internal_required_missing "$(json_array_file "$INTERNAL_MISSING_REQUIRED_LIST")" \
    --argjson internal_intentional_missing "$(json_array_file "$INTERNAL_MISSING_INTENTIONAL_LIST")" \
    --argjson standalone_root_only "$(json_array_file "$STANDALONE_ROOT_ONLY_LIST")" \
    --argjson standalone_internal_only "$(json_array_file "$STANDALONE_INTERNAL_ONLY_LIST")" \
    '{
      mode: $mode,
      status: $status,
      canonical_path: $canonical_path,
      standalone_repo: $standalone_repo,
      inspect_root: $inspect_root,
      inspect_mode: $inspect_mode,
      refreshed_bare_origin: $refreshed_bare_origin,
      target_package_dir: $target_package_dir,
      direct_copy: {
        candidate_count: $copy_candidate_count,
        drift_preview_enabled: $diff_preview_enabled,
        drift_preview_lines: $diff_preview_lines,
        present_count: $copy_present_count,
        identical_count: $copy_identical_count,
        drifted_count: $copy_drifted_count,
        required_drift_count: $copy_drifted_required_count,
        intentional_drift_count: $copy_drifted_intentional_count,
        review_drift_count: $copy_drifted_review_count,
        missing_count: $copy_missing_count,
        required_missing_count: $copy_missing_required_count,
        intentional_missing_count: $copy_missing_intentional_count,
        candidates: $copy_candidates,
        present: $copy_present,
        identical: $copy_identical,
        drifted: $copy_drifted,
        missing: $copy_missing,
        required_drift: $copy_drifted_required,
        intentional_drift: $copy_drifted_intentional,
        review_drift: $copy_drifted_review,
        required_missing: $copy_missing_required,
        intentional_missing: $copy_missing_intentional,
        drift_previews: $copy_drift_previews
      },
      go_projection: {
        canonical_count: $canonical_go_count,
        target_count: $target_go_count,
        drift_preview_enabled: $diff_preview_enabled,
        drift_preview_lines: $diff_preview_lines,
        identical_count: $identical_count,
        drifted_count: $drifted_count,
        required_drift_count: $drifted_required_count,
        intentional_drift_count: $drifted_intentional_count,
        review_drift_count: $drifted_review_count,
        canonical_only_count: $canonical_only_count,
        projection_required_count: $projection_required_count,
        intentional_canonical_only_count: $intentional_canonical_only_count,
        target_only_count: $target_only_count,
        identical: $go_identical,
        drifted: $go_drifted,
        required_drift: $go_drifted_required,
        intentional_drift: $go_drifted_intentional,
        review_drift: $go_drifted_review,
        canonical_only: $go_canonical_only,
        projection_required: $go_projection_required,
        intentional_canonical_only: $go_intentional_canonical_only,
        target_only: $go_target_only,
        drift_previews: $go_drift_previews
      },
      internal_dependencies: {
        import_count: $internal_import_count,
        present_count: $internal_present_count,
        identical_count: $internal_identical_count,
        drifted_count: $internal_drifted_count,
        required_drift_count: $internal_required_drift_count,
        intentional_drift_count: $internal_intentional_drift_count,
        review_drift_count: $internal_review_drift_count,
        missing_count: $internal_missing_count,
        required_missing_count: $internal_required_missing_count,
        intentional_missing_count: $internal_intentional_missing_count,
        imports: $internal_imports,
        present: $internal_present,
        identical: $internal_identical,
        drifted: $internal_drifted,
        required_drift: $internal_required_drift,
        intentional_drift: $internal_intentional_drift,
        review_drift: $internal_review_drift,
        missing: $internal_missing,
        required_missing: $internal_required_missing,
        intentional_missing: $internal_intentional_missing
      },
      standalone_owned: {
        root_entries: $standalone_root_only,
        internal_packages: $standalone_internal_only
      },
      rewrite_rules: [
        "Map bundled root *.go files to internal/dotfiles/*.go",
        "Rewrite package main to package dotfiles for projected Go files",
        $main_projection_note
      ]
    }'
  exit 0
fi

printf 'dotfiles-mcp manual projection %s\n' "$MODE"
printf 'status              %s\n' "$plan_status"
printf 'canonical           %s\n' "$CANONICAL_PATH"
printf 'standalone repo     %s\n' "$STANDALONE_PATH"
printf 'inspect root        %s\n' "$STANDALONE_INSPECT_ROOT"
printf 'inspect mode        %s\n' "$inspect_mode"
printf 'refreshed origin    %s\n' "$REFRESHED_BARE_ORIGIN"
printf 'target package dir  %s\n' "$TARGET_PACKAGE_DIR"
printf '\n'

printf 'direct-copy candidates\n'
printf '  candidates        %s\n' "$(count_lines "$COPY_ALLOWLIST_PRESENT_LIST")"
printf '  present           %s\n' "$(count_lines "$COPY_PRESENT_LIST")"
printf '  identical         %s\n' "$(count_lines "$COPY_IDENTICAL_LIST")"
printf '  drifted           %s\n' "$(count_lines "$COPY_DRIFTED_LIST")"
printf '  drifted required  %s\n' "$(count_lines "$COPY_DRIFTED_REQUIRED_LIST")"
printf '  drifted intent.   %s\n' "$(count_lines "$COPY_DRIFTED_INTENTIONAL_LIST")"
printf '  drifted review    %s\n' "$(count_lines "$COPY_DRIFTED_REVIEW_LIST")"
printf '  missing           %s\n' "$(count_lines "$COPY_MISSING_LIST")"
printf '  missing required  %s\n' "$(count_lines "$COPY_MISSING_REQUIRED_LIST")"
printf '  missing intent.   %s\n' "$(count_lines "$COPY_MISSING_INTENTIONAL_LIST")"
if [[ -s "$COPY_DRIFTED_REQUIRED_LIST" ]]; then
  printf '  direct-copy candidates requiring projection\n'
  sed 's/^/    - /' "$COPY_DRIFTED_REQUIRED_LIST"
fi
if [[ -s "$COPY_DRIFTED_INTENTIONAL_LIST" ]]; then
  printf '  intentional mirror-owned direct-copy drift\n'
  sed 's/^/    - /' "$COPY_DRIFTED_INTENTIONAL_LIST"
fi
if [[ -s "$COPY_DRIFTED_REVIEW_LIST" ]]; then
  printf '  direct-copy drift requiring parity review\n'
  sed 's/^/    - /' "$COPY_DRIFTED_REVIEW_LIST"
fi
if $DIFF_PREVIEW && [[ -s "$COPY_DRIFT_PREVIEW_JSON" ]]; then
  printf '  direct-copy diff previews (first %s lines per file)\n' "$DIFF_PREVIEW_LINES"
  jq -sr '.[] | "    - " + .path + " (+" + (.added|tostring) + " / -" + (.deleted|tostring) + ")\n" + (.preview | split("\n") | map(select(length > 0) | "      " + .) | join("\n"))' "$COPY_DRIFT_PREVIEW_JSON"
fi
if [[ -s "$COPY_MISSING_REQUIRED_LIST" ]]; then
  printf '  missing direct-copy files requiring projection\n'
  sed 's/^/    - /' "$COPY_MISSING_REQUIRED_LIST"
fi
if [[ -s "$COPY_MISSING_INTENTIONAL_LIST" ]]; then
  printf '  intentional mirror-owned missing files\n'
  sed 's/^/    - /' "$COPY_MISSING_INTENTIONAL_LIST"
fi
printf '\n'

printf 'go projection\n'
printf '  canonical root    %s\n' "$(count_lines "$CANONICAL_GO_LIST")"
printf '  target package    %s\n' "$(count_lines "$TARGET_GO_LIST")"
printf '  identical         %s\n' "$(count_lines "$GO_IDENTICAL_LIST")"
printf '  drifted overlap   %s\n' "$(count_lines "$GO_DRIFTED_LIST")"
printf '  drifted required  %s\n' "$(count_lines "$GO_DRIFTED_REQUIRED_LIST")"
printf '  drifted intent.   %s\n' "$(count_lines "$GO_DRIFTED_INTENTIONAL_LIST")"
printf '  drifted review    %s\n' "$(count_lines "$GO_DRIFTED_REVIEW_LIST")"
printf '  canonical only    %s\n' "$(count_lines "$GO_CANONICAL_ONLY_LIST")"
printf '  projection needed %s\n' "$(count_lines "$GO_CANONICAL_ONLY_REQUIRED_LIST")"
printf '  intentional diff  %s\n' "$(count_lines "$GO_CANONICAL_ONLY_INTENTIONAL_LIST")"
printf '  target only       %s\n' "$(count_lines "$GO_TARGET_ONLY_LIST")"
if [[ -s "$GO_DRIFTED_REQUIRED_LIST" ]]; then
  printf '  overlapping files requiring projection alignment\n'
  sed 's/^/    - /' "$GO_DRIFTED_REQUIRED_LIST"
fi
if [[ -s "$GO_DRIFTED_INTENTIONAL_LIST" ]]; then
  printf '  intentional standalone-owned overlap drift\n'
  sed 's/^/    - /' "$GO_DRIFTED_INTENTIONAL_LIST"
fi
if [[ -s "$GO_DRIFTED_REVIEW_LIST" ]]; then
  printf '  overlapping files requiring parity review\n'
  sed 's/^/    - /' "$GO_DRIFTED_REVIEW_LIST"
fi
if [[ -s "$GO_CANONICAL_ONLY_REQUIRED_LIST" ]]; then
  printf '  canonical-only additions requiring projection\n'
  sed 's/^/    - /' "$GO_CANONICAL_ONLY_REQUIRED_LIST"
fi
if [[ -s "$GO_CANONICAL_ONLY_INTENTIONAL_LIST" ]]; then
  printf '  intentional canonical-only differences\n'
  sed 's/^/    - /' "$GO_CANONICAL_ONLY_INTENTIONAL_LIST"
fi
if $DIFF_PREVIEW && [[ -s "$GO_DRIFT_PREVIEW_JSON" ]]; then
  printf '  overlapping drift previews (first %s lines per file)\n' "$DIFF_PREVIEW_LINES"
  jq -sr '.[] | "    - " + .path + " (+" + (.added|tostring) + " / -" + (.deleted|tostring) + ")\n" + (.preview | split("\n") | map(select(length > 0) | "      " + .) | join("\n"))' "$GO_DRIFT_PREVIEW_JSON"
fi
if [[ -s "$GO_TARGET_ONLY_LIST" ]]; then
  printf '  standalone-only package files\n'
  sed 's/^/    - /' "$GO_TARGET_ONLY_LIST"
fi
printf '\n'

printf 'internal dependencies\n'
printf '  imported         %s\n' "$(count_lines "$INTERNAL_IMPORT_LIST")"
printf '  present          %s\n' "$(count_lines "$INTERNAL_PRESENT_LIST")"
printf '  identical        %s\n' "$(count_lines "$INTERNAL_IDENTICAL_LIST")"
printf '  drifted          %s\n' "$(count_lines "$INTERNAL_DRIFTED_LIST")"
printf '  drifted required %s\n' "$(count_lines "$INTERNAL_DRIFTED_REQUIRED_LIST")"
printf '  drifted intent.  %s\n' "$(count_lines "$INTERNAL_DRIFTED_INTENTIONAL_LIST")"
printf '  drifted review   %s\n' "$(count_lines "$INTERNAL_DRIFTED_REVIEW_LIST")"
printf '  missing          %s\n' "$(count_lines "$INTERNAL_MISSING_LIST")"
printf '  missing required %s\n' "$(count_lines "$INTERNAL_MISSING_REQUIRED_LIST")"
printf '  missing intent.  %s\n' "$(count_lines "$INTERNAL_MISSING_INTENTIONAL_LIST")"
if [[ -s "$INTERNAL_DRIFTED_REQUIRED_LIST" ]]; then
  printf '  internal packages requiring projection alignment\n'
  sed 's/^/    - /' "$INTERNAL_DRIFTED_REQUIRED_LIST"
fi
if [[ -s "$INTERNAL_DRIFTED_INTENTIONAL_LIST" ]]; then
  printf '  intentional standalone-owned internal package drift\n'
  sed 's/^/    - /' "$INTERNAL_DRIFTED_INTENTIONAL_LIST"
fi
if [[ -s "$INTERNAL_DRIFTED_REVIEW_LIST" ]]; then
  printf '  internal packages requiring parity review\n'
  sed 's/^/    - /' "$INTERNAL_DRIFTED_REVIEW_LIST"
fi
if [[ -s "$INTERNAL_MISSING_REQUIRED_LIST" ]]; then
  printf '  missing internal packages requiring projection\n'
  sed 's/^/    - /' "$INTERNAL_MISSING_REQUIRED_LIST"
fi
if [[ -s "$INTERNAL_MISSING_INTENTIONAL_LIST" ]]; then
  printf '  intentional standalone-owned missing internal packages\n'
  sed 's/^/    - /' "$INTERNAL_MISSING_INTENTIONAL_LIST"
fi
printf '\n'

printf 'standalone-owned surfaces\n'
if [[ -s "$STANDALONE_ROOT_ONLY_LIST" ]]; then
  sed 's/^/  - /' "$STANDALONE_ROOT_ONLY_LIST"
fi
if [[ -s "$STANDALONE_INTERNAL_ONLY_LIST" ]]; then
  sed 's/^/  - /' "$STANDALONE_INTERNAL_ONLY_LIST"
fi
printf '\n'

printf 'rewrite rules\n'
printf '  - Project bundled root Go files into internal/dotfiles/*.go\n'
printf '  - Rewrite package main -> package dotfiles for projected Go files\n'
printf '  - Preserve standalone cmd/* entrypoints and the Setup-based bootstrap in internal/dotfiles/main.go\n'
