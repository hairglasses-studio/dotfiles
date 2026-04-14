#!/usr/bin/env bash
# hypr-sync-plugin-binds.sh — reload plugin-backed keybinds after plugins load.
# Called from Hyprland `exec =` so plugin dispatchers are registered on startup
# and after config reloads, without parse-time "Invalid dispatcher" errors.

set -euo pipefail

BIND_FILE="${HOME}/.config/hypr/plugin-binds.conf"
REQUIRED_PLUGINS=(
  split-monitor-workspaces
  hyprexpo
)

if [[ ! -f "$BIND_FILE" ]]; then
  echo "hypr-sync-plugin-binds: missing bind file: $BIND_FILE" >&2
  exit 0
fi

for _ in {1..40}; do
  plugins="$(hyprctl plugins list 2>/dev/null || true)"
  ready=true

  for plugin in "${REQUIRED_PLUGINS[@]}"; do
    if ! grep -q "^Plugin ${plugin} by " <<<"$plugins"; then
      ready=false
      break
    fi
  done

  if [[ "$ready" == true ]]; then
    if ! hyprctl keyword source "$BIND_FILE" >/dev/null; then
      echo "hypr-sync-plugin-binds: failed to source bind file" >&2
      exit 1
    fi
    exit 0
  fi

  sleep 0.25
done

echo "hypr-sync-plugin-binds: timed out waiting for plugin dispatchers" >&2
exit 0
