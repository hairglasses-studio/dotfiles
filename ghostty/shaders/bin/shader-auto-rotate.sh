#!/usr/bin/env zsh
# shader-auto-rotate.sh — Called by launchd on an interval to rotate shaders.
# Reuses shader-playlist.sh engine to advance Ghostty + both Tattoy layers.
# Ghostty auto-reloads config via FSEvents; Tattoy watches its config file.

set -euo pipefail

# Source the playlist engine (defines shader-playlist-next, tattoy-playlist-next)
source "${0:A:h}/shader-playlist.sh"

# Determine active Ghostty playlist
ghostty_playlist="low-intensity"
playlist_cfg="$HOME/.local/state/ghostty/auto-rotate-playlist"
if [[ -f "$playlist_cfg" ]]; then
  ghostty_playlist="$(< "$playlist_cfg")"
  [[ -z "$ghostty_playlist" ]] && ghostty_playlist="low-intensity"
fi

# Advance Ghostty shader
shader-playlist-next "$ghostty_playlist"

# Advance both Tattoy layers (non-fatal if Tattoy isn't configured)
tattoy-playlist-next 2>/dev/null || true
