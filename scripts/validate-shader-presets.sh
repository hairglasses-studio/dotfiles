#!/usr/bin/env bash
set -euo pipefail

# validate-shader-presets.sh — Verify every [shaders.<name>.presets.*]
# key in kitty/shaders/presets.toml resolves to a .glsl file under
# kitty/shaders/darkwindow/.
#
# The MCP tool `shader_preset_apply` writes param substitutions into a
# temporary working copy of the shader before reloading kitty. A preset
# block for a non-existent shader name silently fails when the tool
# tries to open the source file, so catch the drift at edit-time.
#
# Usage:
#   scripts/validate-shader-presets.sh
#
# Exit codes:
#   0  every preset shader resolves
#   1  at least one preset references a missing shader
#   2  presets.toml or shader dir missing

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOTFILES="$(cd "$SCRIPT_DIR/.." && pwd)"
PRESETS="$DOTFILES/kitty/shaders/presets.toml"
SHADER_DIR="$DOTFILES/kitty/shaders/darkwindow"

if [[ ! -f "$PRESETS" ]]; then
    printf 'presets.toml not found: %s\n' "$PRESETS" >&2
    exit 2
fi
if [[ ! -d "$SHADER_DIR" ]]; then
    printf 'shader directory not found: %s\n' "$SHADER_DIR" >&2
    exit 2
fi

# Extract unique shader names from [shaders.<name>.presets.*] headers.
mapfile -t names < <(
    grep -oE '^\[shaders\.[a-zA-Z0-9_-]+\.' "$PRESETS" \
        | sed -E 's/^\[shaders\.([a-zA-Z0-9_-]+)\.$/\1/' \
        | sort -u
)

errors=0
for name in "${names[@]}"; do
    [[ -z "$name" ]] && continue
    if [[ ! -f "$SHADER_DIR/${name}.glsl" ]]; then
        printf 'MISSING: presets.toml -> %s.glsl\n' "$name"
        errors=$((errors + 1))
    fi
done

printf 'presets=%d shaders=%d errors=%d\n' \
    "${#names[@]}" "${#names[@]}" "$errors"

[[ $errors -eq 0 ]]
