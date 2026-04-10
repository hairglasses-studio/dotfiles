#!/usr/bin/env bash
set -euo pipefail

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

MODE="plan"
JSON_MODE=false
CANONICAL_PATH="${HG_DOTFILES}/mcp/dotfiles-mcp"
STANDALONE_PATH="${HG_STUDIO_ROOT}/dotfiles-mcp"
TEMP_WORKTREE=""
TEMP_BASE=""
STANDALONE_INSPECT_ROOT=""

usage() {
  cat <<'EOF'
Usage: hg-dotfiles-mcp-projection.sh [plan|check] [--canonical PATH] [--standalone PATH] [--json]

Inspect the bundled dotfiles MCP module and the standalone dotfiles-mcp publish
repo, then emit the current manual-projection plan and drift summary.

Modes:
  plan    Print the projection plan (default)
  check   Validate that the projection plan can be resolved and inspected

Options:
  --canonical PATH   Override the bundled canonical module path
  --standalone PATH  Override the standalone repo path (bare or non-bare)
  --json             Emit machine-readable JSON
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
    plan|check)
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
    -h|--help)
      usage
      exit 0
      ;;
    *)
      hg_die "Unknown argument: $1"
      ;;
  esac
done

hg_require git jq perl

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

resolve_standalone_inspect_root() {
  local root="$1"
  if git -C "$root" rev-parse --is-bare-repository >/dev/null 2>&1; then
    if [[ "$(git -C "$root" rev-parse --is-bare-repository)" == "true" ]]; then
      TEMP_BASE="$(mktemp -d "${TMPDIR:-/tmp}/dotfiles-mcp-projection.XXXXXX")"
      TEMP_WORKTREE="${TEMP_BASE}/worktree"
      git -C "$root" worktree add --quiet --detach "$TEMP_WORKTREE" >/dev/null
      STANDALONE_INSPECT_ROOT="$TEMP_WORKTREE"
      return 0
    fi
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

resolve_standalone_inspect_root "$STANDALONE_PATH"
TARGET_PACKAGE_DIR="${STANDALONE_INSPECT_ROOT}/internal/dotfiles"

CANONICAL_GO_LIST="$(temp_file)"
TARGET_GO_LIST="$(temp_file)"
GO_IDENTICAL_LIST="$(temp_file)"
GO_DRIFTED_LIST="$(temp_file)"
GO_CANONICAL_ONLY_LIST="$(temp_file)"
GO_TARGET_ONLY_LIST="$(temp_file)"
COPY_ALLOWLIST_PRESENT_LIST="$(temp_file)"
COPY_READY_LIST="$(temp_file)"
COPY_MISSING_LIST="$(temp_file)"
STANDALONE_ROOT_ONLY_LIST="$(temp_file)"
STANDALONE_INTERNAL_ONLY_LIST="$(temp_file)"

find "$CANONICAL_PATH" -maxdepth 1 -name '*.go' -printf '%f\n' | sort >"$CANONICAL_GO_LIST"
find "$TARGET_PACKAGE_DIR" -maxdepth 1 -name '*.go' -printf '%f\n' | sort >"$TARGET_GO_LIST"

comm -23 "$CANONICAL_GO_LIST" "$TARGET_GO_LIST" >"$GO_CANONICAL_ONLY_LIST"
comm -13 "$CANONICAL_GO_LIST" "$TARGET_GO_LIST" >"$GO_TARGET_ONLY_LIST"

while IFS= read -r file; do
  [[ -n "$file" ]] || continue
  tmp_rewritten="$(temp_file)"
  perl -0pe 's/^package main$/package dotfiles/m' "$CANONICAL_PATH/$file" >"$tmp_rewritten"
  if cmp -s "$tmp_rewritten" "$TARGET_PACKAGE_DIR/$file"; then
    printf '%s\n' "$file" >>"$GO_IDENTICAL_LIST"
  else
    printf '%s\n' "$file" >>"$GO_DRIFTED_LIST"
  fi
  rm -f "$tmp_rewritten"
done < <(comm -12 "$CANONICAL_GO_LIST" "$TARGET_GO_LIST")

copy_allowlist=(
  ".github/PULL_REQUEST_TEMPLATE.md"
  ".github/copilot-instructions.md"
  ".github/pull_request_template.md"
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

for rel in "${copy_allowlist[@]}"; do
  if [[ -f "$CANONICAL_PATH/$rel" ]]; then
    printf '%s\n' "$rel" >>"$COPY_ALLOWLIST_PRESENT_LIST"
    if [[ -f "$STANDALONE_INSPECT_ROOT/$rel" ]]; then
      printf '%s\n' "$rel" >>"$COPY_READY_LIST"
    else
      printf '%s\n' "$rel" >>"$COPY_MISSING_LIST"
    fi
  fi
done

find "$STANDALONE_INSPECT_ROOT" -mindepth 1 -maxdepth 1 -printf '%f\n' | sort | while IFS= read -r entry; do
  [[ -e "$CANONICAL_PATH/$entry" ]] || printf '%s\n' "$entry"
done >"$STANDALONE_ROOT_ONLY_LIST"

if [[ -d "$STANDALONE_INSPECT_ROOT/internal" ]]; then
  find "$STANDALONE_INSPECT_ROOT/internal" -mindepth 1 -maxdepth 1 -type d -printf 'internal/%f\n' | sort | while IFS= read -r entry; do
    [[ "$entry" == "internal/dotfiles" ]] || printf '%s\n' "$entry"
  done >"$STANDALONE_INTERNAL_ONLY_LIST"
fi

plan_status="in_sync"
if [[ "$(count_lines "$GO_CANONICAL_ONLY_LIST")" -gt 0 || "$(count_lines "$GO_DRIFTED_LIST")" -gt 0 || "$(count_lines "$COPY_MISSING_LIST")" -gt 0 ]]; then
  plan_status="projection_needed"
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
    --argjson canonical_go_count "$(count_lines "$CANONICAL_GO_LIST")" \
    --argjson target_go_count "$(count_lines "$TARGET_GO_LIST")" \
    --argjson identical_count "$(count_lines "$GO_IDENTICAL_LIST")" \
    --argjson drifted_count "$(count_lines "$GO_DRIFTED_LIST")" \
    --argjson canonical_only_count "$(count_lines "$GO_CANONICAL_ONLY_LIST")" \
    --argjson target_only_count "$(count_lines "$GO_TARGET_ONLY_LIST")" \
    --argjson copy_candidate_count "$(count_lines "$COPY_ALLOWLIST_PRESENT_LIST")" \
    --argjson copy_ready_count "$(count_lines "$COPY_READY_LIST")" \
    --argjson copy_missing_count "$(count_lines "$COPY_MISSING_LIST")" \
    --argjson copy_candidates "$(json_array_file "$COPY_ALLOWLIST_PRESENT_LIST")" \
    --argjson copy_ready "$(json_array_file "$COPY_READY_LIST")" \
    --argjson copy_missing "$(json_array_file "$COPY_MISSING_LIST")" \
    --argjson go_identical "$(json_array_file "$GO_IDENTICAL_LIST")" \
    --argjson go_drifted "$(json_array_file "$GO_DRIFTED_LIST")" \
    --argjson go_canonical_only "$(json_array_file "$GO_CANONICAL_ONLY_LIST")" \
    --argjson go_target_only "$(json_array_file "$GO_TARGET_ONLY_LIST")" \
    --argjson standalone_root_only "$(json_array_file "$STANDALONE_ROOT_ONLY_LIST")" \
    --argjson standalone_internal_only "$(json_array_file "$STANDALONE_INTERNAL_ONLY_LIST")" \
    '{
      mode: $mode,
      status: $status,
      canonical_path: $canonical_path,
      standalone_repo: $standalone_repo,
      inspect_root: $inspect_root,
      inspect_mode: $inspect_mode,
      target_package_dir: $target_package_dir,
      direct_copy: {
        candidate_count: $copy_candidate_count,
        ready_count: $copy_ready_count,
        missing_count: $copy_missing_count,
        candidates: $copy_candidates,
        ready: $copy_ready,
        missing: $copy_missing
      },
      go_projection: {
        canonical_count: $canonical_go_count,
        target_count: $target_go_count,
        identical_count: $identical_count,
        drifted_count: $drifted_count,
        canonical_only_count: $canonical_only_count,
        target_only_count: $target_only_count,
        identical: $go_identical,
        drifted: $go_drifted,
        canonical_only: $go_canonical_only,
        target_only: $go_target_only
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
printf 'target package dir  %s\n' "$TARGET_PACKAGE_DIR"
printf '\n'

printf 'direct-copy candidates\n'
printf '  candidates        %s\n' "$(count_lines "$COPY_ALLOWLIST_PRESENT_LIST")"
printf '  ready             %s\n' "$(count_lines "$COPY_READY_LIST")"
printf '  missing           %s\n' "$(count_lines "$COPY_MISSING_LIST")"
if [[ -s "$COPY_MISSING_LIST" ]]; then
  sed 's/^/    - /' "$COPY_MISSING_LIST"
fi
printf '\n'

printf 'go projection\n'
printf '  canonical root    %s\n' "$(count_lines "$CANONICAL_GO_LIST")"
printf '  target package    %s\n' "$(count_lines "$TARGET_GO_LIST")"
printf '  identical         %s\n' "$(count_lines "$GO_IDENTICAL_LIST")"
printf '  drifted overlap   %s\n' "$(count_lines "$GO_DRIFTED_LIST")"
printf '  canonical only    %s\n' "$(count_lines "$GO_CANONICAL_ONLY_LIST")"
printf '  target only       %s\n' "$(count_lines "$GO_TARGET_ONLY_LIST")"
if [[ -s "$GO_CANONICAL_ONLY_LIST" ]]; then
  printf '  canonical-only additions requiring projection\n'
  sed 's/^/    - /' "$GO_CANONICAL_ONLY_LIST"
fi
if [[ -s "$GO_DRIFTED_LIST" ]]; then
  printf '  overlapping files that already drift from the bundled source\n'
  sed 's/^/    - /' "$GO_DRIFTED_LIST"
fi
if [[ -s "$GO_TARGET_ONLY_LIST" ]]; then
  printf '  standalone-only package files\n'
  sed 's/^/    - /' "$GO_TARGET_ONLY_LIST"
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
