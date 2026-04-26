#!/usr/bin/env bash
# check-shader-consistency.sh — Verify playlist ↔ registry ↔ .glsl three-way sync
#
# Ensures that ambient.txt, darkwindow-shaders.conf, and the actual .glsl files
# in kitty/shaders/darkwindow/ all agree on which shaders exist.
#
# Exit 0 if consistent, exit 1 with a diff of mismatches.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOTFILES="$(cd "$SCRIPT_DIR/.." && pwd)"

PLAYLIST="$DOTFILES/kitty/shaders/playlists/ambient.txt"
REGISTRY="$DOTFILES/hyprland/darkwindow-shaders.conf"
SHADER_DIR="$DOTFILES/kitty/shaders/darkwindow"

errors=0

# Extract sorted shader names from each source
playlist_names() {
  awk 'NF && $1 !~ /^#/ { sub(/\.glsl$/, ""); print }' "$PLAYLIST" | sort
}

registry_names() {
  grep -oP 'shader\[\K[^\]]+' "$REGISTRY" | sort
}

disk_names() {
  find "$SHADER_DIR" -maxdepth 1 -name '*.glsl' -printf '%f\n' | sed 's/\.glsl$//' | sort
}

# Check playlist entries have matching .glsl files
echo "Checking playlist → disk..."
while IFS= read -r name; do
  if [[ ! -f "$SHADER_DIR/${name}.glsl" ]]; then
    echo "  MISSING on disk: $name (in playlist but no .glsl file)"
    errors=$((errors + 1))
  fi
done < <(playlist_names)

# Check playlist entries have matching registry entries
echo "Checking playlist → registry..."
while IFS= read -r name; do
  echo "  MISSING in registry: $name (in playlist but not in darkwindow-shaders.conf)"
  errors=$((errors + 1))
done < <(comm -23 <(playlist_names) <(registry_names))

# Check registry entries have matching playlist entries
echo "Checking registry → playlist..."
while IFS= read -r name; do
  echo "  EXTRA in registry: $name (in darkwindow-shaders.conf but not in playlist)"
  errors=$((errors + 1))
done < <(comm -13 <(playlist_names) <(registry_names))

# Summary
playlist_count=$(playlist_names | wc -l)
registry_count=$(registry_names | wc -l)
disk_count=$(disk_names | wc -l)
echo "playlist=$playlist_count  registry=$registry_count  disk=$disk_count"

if [[ "$playlist_count" -ne "$registry_count" ]]; then
  echo "error: playlist count ($playlist_count) != registry count ($registry_count)" >&2
  errors=$((errors + 1))
fi

if [[ "$errors" -gt 0 ]]; then
  echo "error: $errors consistency issue(s) found" >&2
  exit 1
fi

echo "ok: all consistent"
exit 0
