#!/usr/bin/env bash
# juhradial-settings.sh — launch the installed juhradial settings dashboard
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")" && pwd)"
source "$SCRIPT_DIR/lib/juhradial.sh"

settings_script="$(juhradial_settings_script)"
if [[ ! -f "$settings_script" ]]; then
  printf 'juhradial settings not installed at %s\n' "$settings_script" >&2
  printf 'Run %s/scripts/juhradial-install.sh first.\n' "$(juhradial_dotfiles_dir)" >&2
  exit 1
fi

juhradial_export_graphical_env
exec python3 "$settings_script" "$@"
