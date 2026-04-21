#!/usr/bin/env bash
set -euo pipefail

# validate-shader-playlists.sh — Verify every entry in
# kitty/shaders/playlists/*.txt resolves to a real .glsl file under
# kitty/shaders/darkwindow/.
#
# Runtime shader rotators ignore playlist entries that fail to load, so
# a rename or removal silently drops the shader from rotation. This gate
# catches the drift at edit-time.
#
# Usage:
#   scripts/validate-shader-playlists.sh          # scan every *.txt
#   scripts/validate-shader-playlists.sh NAME     # scan one playlist
#
# Exit codes:
#   0  every entry resolves
#   1  at least one entry missing
#   2  shader dir or playlist dir missing

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOTFILES="$(cd "$SCRIPT_DIR/.." && pwd)"
SHADER_DIR="$DOTFILES/kitty/shaders/darkwindow"
PLAYLIST_DIR="$DOTFILES/kitty/shaders/playlists"

if [[ ! -d "$SHADER_DIR" ]]; then
    printf 'shader directory not found: %s\n' "$SHADER_DIR" >&2
    exit 2
fi
if [[ ! -d "$PLAYLIST_DIR" ]]; then
    printf 'playlist directory not found: %s\n' "$PLAYLIST_DIR" >&2
    exit 2
fi

playlists=()
if [[ $# -gt 0 ]]; then
    for arg in "$@"; do
        if [[ "$arg" == *.txt && -f "$PLAYLIST_DIR/$arg" ]]; then
            playlists+=("$PLAYLIST_DIR/$arg")
        elif [[ -f "$PLAYLIST_DIR/${arg}.txt" ]]; then
            playlists+=("$PLAYLIST_DIR/${arg}.txt")
        else
            printf 'playlist not found: %s\n' "$arg" >&2
            exit 2
        fi
    done
else
    while IFS= read -r -d '' f; do
        playlists+=("$f")
    done < <(find "$PLAYLIST_DIR" -maxdepth 1 -type f -name '*.txt' -print0)
fi

total_entries=0
errors=0
for playlist in "${playlists[@]}"; do
    while IFS= read -r line; do
        trimmed="${line#"${line%%[![:space:]]*}"}"
        trimmed="${trimmed%"${trimmed##*[![:space:]]}"}"
        [[ -z "$trimmed" || "${trimmed:0:1}" == "#" ]] && continue
        total_entries=$((total_entries + 1))
        if [[ ! -f "$SHADER_DIR/$trimmed" ]]; then
            printf 'MISSING: %s -> %s\n' "$(basename "$playlist")" "$trimmed"
            errors=$((errors + 1))
        fi
    done < "$playlist"
done

printf 'playlists=%d entries=%d errors=%d\n' \
    "${#playlists[@]}" "$total_entries" "$errors"

[[ $errors -eq 0 ]]
