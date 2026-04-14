#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")" && pwd)"
source "$SCRIPT_DIR/lib/juhradial.sh"

overlay_script="$(juhradial_overlay_script)"
if [[ ! -f "$overlay_script" ]]; then
  printf 'juhradial overlay not installed at %s\n' "$overlay_script" >&2
  exit 1
fi

juhradial_export_graphical_env
export GIO_LAUNCHED_DESKTOP_FILE
GIO_LAUNCHED_DESKTOP_FILE="$(juhradial_desktop_file)"
export GIO_LAUNCHED_DESKTOP_FILE_PID
GIO_LAUNCHED_DESKTOP_FILE_PID="$$"

exec python3 "$overlay_script"
