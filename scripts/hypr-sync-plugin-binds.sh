#!/usr/bin/env bash
# hypr-sync-plugin-binds.sh — reload plugin-backed keybinds after plugins load.
# Called from Hyprland `exec =` at startup so plugin dispatchers are registered
# without parse-time "Invalid dispatcher" errors. Note: `exec =` does NOT
# fire on `hyprctl reload` in Hyprland 0.54.2 (verified empirically). Runtime
# re-application of monitor transforms is handled by
# dotfiles-hypr-monitor-watch.service.

set -euo pipefail

BIND_FILE="${HOME}/.config/hypr/plugin-binds.conf"
REQUIRED_PLUGINS=(
  split-monitor-workspaces
  hyprexpo
  hy3
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
    # Re-apply general:layout now that the hy3 plugin is loaded. At config
    # parse time hy3 isn't registered yet so Hyprland silently falls back
    # to dwindle and stays there until a runtime override fires.
    hyprctl keyword general:layout hy3 >/dev/null || true
    exit 0
  fi

  sleep 0.25
done

echo "hypr-sync-plugin-binds: timed out waiting for plugin dispatchers" >&2
exit 0
