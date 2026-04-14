#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/hg-agent-parity.sh"

export HG_STUDIO_ROOT HG_DOTFILES
TARGET="$HG_STUDIO_ROOT/codexkit/scripts/provider-settings-sync.sh"

[[ -f "$TARGET" ]] || {
  echo "codexkit provider sync entrypoint missing: $TARGET" >&2
  exit 1
}

forward_args=()
repo_path=""
repo_name=""
expect_repo_name_value=false
for arg in "$@"; do
  if [[ "$arg" == "--include-codex-config" ]]; then
    continue
  fi
  forward_args+=("$arg")
  if $expect_repo_name_value; then
    repo_name="$arg"
    expect_repo_name_value=false
    continue
  fi
  if [[ "$arg" == "--repo-name" ]]; then
    expect_repo_name_value=true
    continue
  fi
  if [[ "$arg" != -* && -z "$repo_path" ]]; then
    repo_path="$arg"
  fi
done

if [[ -z "$repo_name" && -n "$repo_path" ]]; then
  repo_name="$(basename "$repo_path")"
fi

exec bash "$TARGET" "${forward_args[@]}"
