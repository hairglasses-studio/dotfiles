#!/usr/bin/env bash
set -euo pipefail

if [[ $# -lt 2 ]]; then
  printf 'Usage: %s <remote> <command> [args...]\n' "${0##*/}" >&2
  exit 2
fi

remote="$1"
shift

log_dir="${XDG_STATE_HOME:-$HOME/.local/state}/jellyfin-stack"
log_file="$log_dir/rclone-${remote}.log"
mkdir -p "$log_dir"

exec "$@" >>"$log_file" 2>&1
