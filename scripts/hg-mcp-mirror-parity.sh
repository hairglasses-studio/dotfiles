#!/usr/bin/env bash
set -euo pipefail

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
DOTFILES_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
MANIFEST_PATH="${HG_MCP_MIRROR_MANIFEST:-$DOTFILES_DIR/mcp/mirror-parity.json}"
DOC_PATH="$DOTFILES_DIR/docs/MCP-MIRROR-PARITY.md"

usage() {
  cat <<'EOF'
Usage: hg-mcp-mirror-parity.sh [--list|--check]

Track and validate the MCP modules that ship from dotfiles and also publish
from standalone mirror repositories.

Options:
  --list   Print the current mirror matrix.
  --check  Validate the manifest, docs, and README mirror banners.
  --help   Show this help text.
EOF
}

require_jq() {
  command -v jq >/dev/null 2>&1 || {
    printf 'jq is required\n' >&2
    exit 1
  }
}

validate_manifest() {
  jq -e '
    .version == 1 and
    (.mirrors | type == "array") and
    (.mirrors | length > 0) and
    ([.mirrors[] | (.sync_strategy // "tree_sync")] | all(. == "tree_sync" or . == "manual_projection")) and
    ([.mirrors[].module] | length == ([.[]] | unique | length)) and
    ([.mirrors[].standalone_repo] | length == ([.[]] | unique | length)) and
    ([.mirrors[].canonical_path] | length == ([.[]] | unique | length))
  ' "$MANIFEST_PATH" >/dev/null
}

list_matrix() {
  require_jq
  validate_manifest

  printf '%-14s %-16s %-22s %-18s %s\n' "module" "standalone-repo" "canonical-path" "sync-strategy" "purpose"
  printf '%-14s %-16s %-22s %-18s %s\n' "------" "---------------" "--------------" "-------------" "-------"
  jq -r '
    .mirrors[] |
    [.module, .standalone_repo, .canonical_path, (.sync_strategy // "tree_sync"), .purpose] | @tsv
  ' "$MANIFEST_PATH" | while IFS=$'\t' read -r module repo path strategy purpose; do
    printf '%-14s %-16s %-22s %-18s %s\n' "$module" "$repo" "$path" "$strategy" "$purpose"
  done
}

check_matrix() {
  require_jq
  validate_manifest

  [[ -f "$DOC_PATH" ]] || {
    printf 'FAIL  docs  missing %s\n' "$DOC_PATH"
    exit 1
  }

  local failures=0

  while IFS=$'\t' read -r module repo canonical_path strategy purpose; do
    local module_dir="$DOTFILES_DIR/$canonical_path"
    local readme_path="$module_dir/README.md"
    local canonical_url="https://github.com/hairglasses-studio/dotfiles"
    local standalone_url="https://github.com/hairglasses-studio/$repo"

    if [[ ! -d "$module_dir" ]]; then
      printf 'FAIL  %s  missing module dir %s\n' "$module" "$module_dir"
      failures=$((failures + 1))
      continue
    fi

    if [[ ! -f "$module_dir/go.mod" ]]; then
      printf 'FAIL  %s  missing go.mod under %s\n' "$module" "$canonical_path"
      failures=$((failures + 1))
      continue
    fi

    if [[ ! -f "$readme_path" ]]; then
      printf 'FAIL  %s  missing README.md\n' "$module"
      failures=$((failures + 1))
      continue
    fi

    if ! grep -q "$canonical_url" "$readme_path"; then
      printf 'FAIL  %s  README missing canonical dotfiles URL\n' "$module"
      failures=$((failures + 1))
      continue
    fi

    if ! grep -q "$canonical_path" "$readme_path"; then
      printf 'FAIL  %s  README missing canonical path %s\n' "$module" "$canonical_path"
      failures=$((failures + 1))
      continue
    fi

    if ! grep -q "$standalone_url" "$readme_path"; then
      printf 'FAIL  %s  README missing standalone repo URL %s\n' "$module" "$standalone_url"
      failures=$((failures + 1))
      continue
    fi

    if ! grep -qi 'publish mirror kept in parity' "$readme_path"; then
      printf 'FAIL  %s  README missing publish mirror parity banner\n' "$module"
      failures=$((failures + 1))
      continue
    fi

    if ! grep -q "\`$module\`" "$DOC_PATH"; then
      printf 'FAIL  %s  docs missing module row in MCP-MIRROR-PARITY.md\n' "$module"
      failures=$((failures + 1))
      continue
    fi

    printf 'PASS  %s  [%s] %s\n' "$module" "$strategy" "$purpose"
  done < <(
    jq -r '.mirrors[] | [.module, .standalone_repo, .canonical_path, (.sync_strategy // "tree_sync"), .purpose] | @tsv' "$MANIFEST_PATH"
  )

  if [[ "$failures" -gt 0 ]]; then
    exit 1
  fi
}

main() {
  case "${1:---list}" in
    --list)
      list_matrix
      ;;
    --check)
      check_matrix
      ;;
    --help|-h)
      usage
      ;;
    *)
      usage >&2
      exit 1
      ;;
  esac
}

main "$@"
