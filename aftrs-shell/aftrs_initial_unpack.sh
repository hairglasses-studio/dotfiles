#!/usr/bin/env bash
# Copies config seeds from aftrs_init/config into the user's config dir
# Convention: everything under aftrs_init/config/<subpath> -> ~/.config/<subpath>
set -euo pipefail

HOST="$(hostname -s || echo unknown)"
USE_GUM=false
if command -v gum >/dev/null 2>&1 && [ "$HOST" != "wizardstower" ]; then USE_GUM=true; fi
log()   { if $USE_GUM; then gum log --level info "$@"; else echo "[INFO] $*"; fi; }
warn()  { if $USE_GUM; then gum log --level warn "$@"; else echo "[WARN] $*" >&2; fi; }

dry_run=false
DEST_CONFIG_BASE="${DEST_CONFIG_BASE:-$HOME/.config}"
LOCAL_ROOT="${AFTRS_LOCAL_ROOT:-/home/hg/Docs/aftrs-shell}"
SEEDS_BASE="${SEEDS_BASE:-$LOCAL_ROOT/aftrs-shell/aftrs_init/config}"

usage() {
  cat <<'USAGE'
Usage: aftrs_initial_unpack.sh [--dry-run] [--dest-config-dir DIR]

Copies all files from aftrs_init/config to the target config dir (default: ~/.config),
preserving subdirectories (e.g., oh-my-posh/aftrs.omp.json -> ~/.config/oh-my-posh/aftrs.omp.json).
USAGE
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --dry-run) dry_run=true ;;
    --dest-config-dir) shift; DEST_CONFIG_BASE="$1" ;;
    -h|--help) usage; exit 0 ;;
    --) shift; break ;;
    *) warn "Unknown option: $1"; usage >&2; exit 1 ;;
  esac
  shift
done

if [ ! -d "$SEEDS_BASE" ]; then
  warn "Seed directory not found: $SEEDS_BASE"
  exit 0
fi

log "Seeding configs from $SEEDS_BASE to $DEST_CONFIG_BASE"
while IFS= read -r -d '' f; do
  rel="${f#$SEEDS_BASE/}"
  dest_dir="$DEST_CONFIG_BASE/$(dirname "$rel")"
  dest="$DEST_CONFIG_BASE/$rel"
  log " -> $dest"
  if ! $dry_run; then
    mkdir -p "$dest_dir"
    cp -f "$f" "$dest"
  fi
done < <(find "$SEEDS_BASE" -type f -print0)
log "Config seeding complete"


