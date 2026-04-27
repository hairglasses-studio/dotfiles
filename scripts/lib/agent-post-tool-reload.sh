#!/usr/bin/env bash
# PostToolUse reload hook shared by Claude-compatible agent configs.
#
# Reads a hook JSON envelope from stdin, detects the written file path, and
# fires narrow reload/format actions for known workstation config surfaces.
# All actions are best-effort and the hook always allows the original tool.

set -euo pipefail

INPUT="$(cat 2>/dev/null || echo '{}')"

jq_field() {
  local query="$1"
  printf '%s' "$INPUT" | jq -r "$query // empty" 2>/dev/null || true
}

json_allow() {
  printf '{"decision":"allow"}\n'
}

TOOL_NAME="$(jq_field '.tool_name')"

case "$TOOL_NAME" in
  Write|Edit|NotebookEdit) ;;
  *) json_allow; exit 0 ;;
esac

FILE_PATH="$(jq_field '.tool_input.file_path // .tool_input.path')"
FILE_PATH="${FILE_PATH:-${CLAUDE_FILE:-}}"

if [[ -z "$FILE_PATH" ]]; then
  json_allow
  exit 0
fi

format_go() {
  local file="$1"
  command -v goimports >/dev/null 2>&1 && goimports -w "$file" >/dev/null 2>&1 || true
  command -v gofumpt >/dev/null 2>&1 && gofumpt -w "$file" >/dev/null 2>&1 || true
}

reload_hyprland() {
  command -v hyprctl >/dev/null 2>&1 && hyprctl reload >/dev/null 2>&1 || true
}

reload_swaync() {
  command -v swaync-client >/dev/null 2>&1 && swaync-client --reload-config >/dev/null 2>&1 || true
}

reload_eww() {
  command -v eww >/dev/null 2>&1 && eww reload >/dev/null 2>&1 || true
}

reload_sway() {
  command -v swaymsg >/dev/null 2>&1 && swaymsg reload >/dev/null 2>&1 || true
}

restart_makima() {
  command -v sudo >/dev/null 2>&1 && command -v systemctl >/dev/null 2>&1 \
    && sudo -n systemctl restart makima >/dev/null 2>&1 || true
}

restart_logid() {
  command -v sudo >/dev/null 2>&1 && command -v systemctl >/dev/null 2>&1 \
    && sudo -n cp "$FILE_PATH" /etc/logid.cfg >/dev/null 2>&1 \
    && sudo -n systemctl restart logid >/dev/null 2>&1 || true
}

rebuild_bat_cache() {
  command -v bat >/dev/null 2>&1 && bat cache --build >/dev/null 2>&1 || true
}

case "$FILE_PATH" in
  *.go)
    format_go "$FILE_PATH"
    ;;
  */hypr/*|*/hyprland/*)
    reload_hyprland
    ;;
  */swaync/*)
    reload_swaync
    ;;
  */eww/*.yuck|*/eww/*.scss)
    reload_eww
    ;;
  */sway/*)
    reload_sway
    ;;
  */makima/*.toml)
    restart_makima
    ;;
  */logiops/logid.cfg)
    restart_logid
    ;;
  */bat/config)
    rebuild_bat_cache
    ;;
esac

json_allow
