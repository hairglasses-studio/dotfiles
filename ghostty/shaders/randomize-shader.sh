#!/usr/bin/env bash
# randomize-shader.sh — pick a random shader (different from current) and write it to Ghostty config.
# Ghostty auto-detects config changes via FSEvents, so no explicit reload signal is needed.
# Called from AeroSpace keybind (alt-shift-s) or manually.

set -euo pipefail

SHADER_DIR="${HOME}/.config/ghostty/shaders"
CONFIG="${HOME}/.config/ghostty/config"

# Collect all .glsl files
shaders=()
for f in "${SHADER_DIR}"/*.glsl; do
  [[ -f "$f" ]] && shaders+=("$f")
done

# Nothing to do if no shaders
(( ${#shaders[@]} > 0 )) || exit 0

# Read current shader from config to avoid picking the same one
current=""
if [[ -f "$CONFIG" ]]; then
  current="$(grep -m1 '^custom-shader = ' "$CONFIG" 2>/dev/null | sed 's/^custom-shader = //' || true)"
fi

# Pick a random shader, retrying up to 5 times to avoid the current one
pick="${shaders[0]}"
for _ in 1 2 3 4 5; do
  pick="${shaders[RANDOM % ${#shaders[@]}]}"
  [[ "$pick" != "$current" ]] && break
done

# Determine if the shader needs animation (references time uniforms)
anim="false"
grep -qE '(ghostty_time|iTime|u_time)' "$pick" 2>/dev/null && anim="true"

# Atomic config update: write to temp file, then replace
tmp="$(mktemp "${CONFIG}.XXXXXX")"
sed "s|^custom-shader = .*|custom-shader = ${pick}|" "$CONFIG" \
  | sed "s|^custom-shader-animation = .*|custom-shader-animation = ${anim}|" \
  > "$tmp"
mv "$tmp" "$CONFIG"
