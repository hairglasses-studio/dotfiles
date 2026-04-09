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

Inspect or validate the repo-local MCP mirror contract used by public docs and CI.

Options:
  --list   Print the tracked mirror matrix.
  --check  Validate the manifest, docs, and bundled module README banners.
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
    ([.mirrors[].module] | length == (. | unique | length)) and
    ([.mirrors[].standalone_repo] | length == (. | unique | length)) and
    ([.mirrors[].repo_subpath] | length == (. | unique | length)) and
    ([.mirrors[].workspace_path] | length == (. | unique | length))
  ' "$MANIFEST_PATH" >/dev/null
}

list_matrix() {
  require_jq
  validate_manifest

  printf '%-13s %-16s %-28s %s\n' "module" "standalone-repo" "workspace-path" "purpose"
  printf '%-13s %-16s %-28s %s\n' "------" "---------------" "--------------" "-------"
  jq -r '
    .mirrors[] |
    [.module, .standalone_repo, .workspace_path, .purpose] | @tsv
  ' "$MANIFEST_PATH" | while IFS=$'\t' read -r module repo path purpose; do
    printf '%-13s %-16s %-28s %s\n' "$module" "$repo" "$path" "$purpose"
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

  while IFS=$'\t' read -r module repo repo_subpath workspace_path purpose; do
    local module_dir="$DOTFILES_DIR/$repo_subpath"
    local readme_path="$module_dir/README.md"
    local canonical_url="https://github.com/hairglasses-studio/dotfiles"
    local standalone_url="https://github.com/hairglasses-studio/$repo"

    if [[ ! -d "$module_dir" ]]; then
      printf 'FAIL  %s  missing module dir %s\n' "$module" "$module_dir"
      failures=$((failures + 1))
      continue
    fi

    if [[ ! -f "$module_dir/go.mod" ]]; then
      printf 'FAIL  %s  missing go.mod under %s\n' "$module" "$repo_subpath"
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

    if ! grep -q "$repo_subpath" "$readme_path"; then
      printf 'FAIL  %s  README missing bundled repo path %s\n' "$module" "$repo_subpath"
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

    if ! grep -q "\`$workspace_path\`" "$DOC_PATH"; then
      printf 'FAIL  %s  docs missing workspace path row %s\n' "$module" "$workspace_path"
      failures=$((failures + 1))
      continue
    fi

    if ! grep -q "\`hairglasses-studio/$repo\`" "$DOC_PATH"; then
      printf 'FAIL  %s  docs missing standalone repo row %s\n' "$module" "$repo"
      failures=$((failures + 1))
      continue
    fi

    printf 'PASS  %s  %s\n' "$module" "$purpose"
  done < <(
    jq -r '
      .mirrors[] |
      [.module, .standalone_repo, .repo_subpath, .workspace_path, .purpose] | @tsv
    ' "$MANIFEST_PATH"
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
