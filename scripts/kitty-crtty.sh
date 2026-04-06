#!/usr/bin/env bash
# kitty-crtty.sh — Launch kitty with CRTty shader injection
# Falls back to plain kitty if CRTty is not installed.

ACTIVE="$HOME/.local/state/kitty-shaders/crtty-active.glsl"

if command -v crtty &>/dev/null; then
  if [[ -f "$ACTIVE" ]]; then
    exec crtty -s "$ACTIVE" --app kitty "$@"
  else
    exec crtty --app kitty "$@"
  fi
else
  exec kitty "$@"
fi
