#!/usr/bin/env bash
# juhradial-patch-guard.sh — verify the repo patch set still applies to the pinned upstream commit
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")" && pwd)"
source "$SCRIPT_DIR/lib/juhradial.sh"

quiet=false
local_source=false

for arg in "$@"; do
  case "$arg" in
    --quiet) quiet=true ;;
    --local-source) local_source=true ;;
    *)
      printf 'Unknown option: %s\n' "$arg" >&2
      exit 2
      ;;
  esac
done

log() {
  $quiet || printf '[juhradial-patch-guard] %s\n' "$*"
}

require_cmd() {
  local cmd="$1"
  command -v "$cmd" >/dev/null 2>&1 || {
    printf 'Missing required command: %s\n' "$cmd" >&2
    exit 1
  }
}

cleanup() {
  [[ -n "${tmp_dir:-}" && -d "${tmp_dir:-}" ]] && rm -rf "$tmp_dir"
}

require_cmd git
require_cmd cargo

patch_dir="$(juhradial_patch_dir)"
[[ -d "$patch_dir" ]] || {
  log "No juhradial patch directory found; nothing to verify"
  exit 0
}

tmp_dir="$(mktemp -d)"
trap cleanup EXIT
src_dir="$tmp_dir/src"

if $local_source; then
  local_checkout="$(juhradial_source_dir)"
  [[ -d "$local_checkout/.git" ]] || {
    printf 'Local juhradial source checkout not found at %s\n' "$local_checkout" >&2
    exit 1
  }
  log "Cloning local juhradial source checkout"
  git clone --quiet "$local_checkout" "$src_dir"
else
  log "Cloning pinned juhradial upstream"
  git clone --quiet "$(juhradial_repo_url)" "$src_dir"
fi

git -C "$src_dir" checkout --detach "$(juhradial_pinned_commit)" >/dev/null

while IFS= read -r patch; do
  [[ -n "$patch" ]] || continue
  log "Checking $(basename "$patch")"
  git -C "$src_dir" apply --check "$patch"
  git -C "$src_dir" apply "$patch"
done < <(find "$patch_dir" -maxdepth 1 -type f -name '*.patch' | sort)

log "Running daemon tests"
cargo test --manifest-path "$src_dir/daemon/Cargo.toml" --release >/dev/null

log "Patch guard passed for $(juhradial_pinned_commit)"
