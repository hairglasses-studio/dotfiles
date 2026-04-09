#!/usr/bin/env bash
# juhradial-sync.sh — copy repo-owned juhradial seed config into ~/.config/juhradial
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/juhradial.sh"

quiet=false
for arg in "$@"; do
  case "$arg" in
    --quiet) quiet=true ;;
    *)
      printf 'Unknown option: %s\n' "$arg" >&2
      exit 2
      ;;
  esac
done

log() {
  $quiet || printf '[juhradial-sync] %s\n' "$*"
}

seed_dir="$(juhradial_seed_dir)"
config_dir="$(juhradial_config_dir)"
mkdir -p "$config_dir"

sync_file() {
  local name="$1"
  local src="$seed_dir/$name"
  local dst="$config_dir/$name"
  local backup

  if [[ ! -f "$src" ]]; then
    printf 'Missing seed file: %s\n' "$src" >&2
    exit 1
  fi

  if [[ -f "$dst" ]] && ! cmp -s "$src" "$dst"; then
    backup="$dst.bak.$(date +%Y%m%d-%H%M%S)"
    cp "$dst" "$backup"
    log "Backed up $(basename "$dst") to $(basename "$backup")"
  fi

  install -Dm644 "$src" "$dst"
  log "Copied $name"
}

sync_file config.json
sync_file profiles.json

log "Seed config synced to $config_dir"
