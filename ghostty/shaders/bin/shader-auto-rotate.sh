#!/usr/bin/env zsh
# shader-auto-rotate.sh — Called by systemd timer to rotate shaders.
# Reuses shader-playlist.sh engine to advance Ghostty shader.

set -euo pipefail

# Source the playlist engine (defines shader-playlist-next)
source "${0:A:h}/shader-playlist.sh"

# Determine active Ghostty playlist
ghostty_playlist="ambient"
playlist_cfg="$HOME/.local/state/ghostty/auto-rotate-playlist"
if [[ -f "$playlist_cfg" ]]; then
  ghostty_playlist="$(< "$playlist_cfg")"
  [[ -z "$ghostty_playlist" ]] && ghostty_playlist="ambient"
fi

# Advance Ghostty shader
shader-playlist-next "$ghostty_playlist"
