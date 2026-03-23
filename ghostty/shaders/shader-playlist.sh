#!/usr/bin/env zsh
# shader-playlist.sh — shuffled shader playlist engine for Ghostty
# Source this from aliases.zsh; call shader-playlist-next <playlist-name>

_shader_playlist_dir="${0:A:h}/playlists"
_shader_state_dir="$HOME/.local/state/ghostty"
_shader_base_dir="$HOME/.config/ghostty/shaders"
_ghostty_config="$HOME/.config/ghostty/config"

# Fisher-Yates shuffle: reads lines from file, prints shuffled
_shader_shuffle() {
  local -a lines=("${(@f)$(< "$1")}")
  local i j tmp n=${#lines}
  for (( i = n; i > 1; i-- )); do
    j=$(( RANDOM % i + 1 ))
    tmp="${lines[$i]}"
    lines[$i]="${lines[$j]}"
    lines[$j]="$tmp"
  done
  printf '%s\n' "${lines[@]}"
}

# Check if a shader needs custom-shader-animation = true
_shader_needs_animation() {
  grep -qE '(ghostty_time|iTime|u_time)' "$1" 2>/dev/null
}

# Main entry point: advance playlist and apply shader
# Usage: shader-playlist-next <playlist-name>
shader-playlist-next() {
  local name="$1"
  [[ -z "$name" ]] && { echo "Usage: shader-playlist-next <playlist-name>" >&2; return 1; }

  local playlist_file="$_shader_playlist_dir/${name}.txt"
  [[ -f "$playlist_file" ]] || { echo "Playlist not found: $playlist_file" >&2; return 1; }

  mkdir -p "$_shader_state_dir" 2>/dev/null

  local queue_file="$_shader_state_dir/${name}.queue"
  local idx_file="$_shader_state_dir/${name}.idx"
  local hash_file="$_shader_state_dir/${name}.hash"

  # Hash current playlist to detect edits
  local current_hash
  current_hash="$(md5 -q "$playlist_file" 2>/dev/null || md5sum "$playlist_file" | cut -d' ' -f1)"
  local stored_hash=""
  [[ -f "$hash_file" ]] && stored_hash="$(< "$hash_file")"

  # Read current index
  local idx=0
  if [[ -f "$idx_file" ]]; then
    idx="$(< "$idx_file")"
    [[ "$idx" =~ ^[0-9]+$ ]] || idx=0
  fi

  # Count queue length
  local queue_len=0
  [[ -f "$queue_file" ]] && queue_len="$(wc -l < "$queue_file" | tr -d ' ')"

  # Reshuffle if needed
  if [[ ! -f "$queue_file" ]] || [[ "$queue_len" -eq 0 ]] || \
     [[ "$current_hash" != "$stored_hash" ]] || [[ "$idx" -ge "$queue_len" ]]; then
    _shader_shuffle "$playlist_file" > "$queue_file"
    idx=0
    printf '%s' "$current_hash" > "$hash_file"
    queue_len="$(wc -l < "$queue_file" | tr -d ' ')"
  fi

  # Pick shader, skipping missing files (max 5 retries)
  local shader_name shader_path attempts=0
  while (( attempts < 5 )); do
    if (( idx >= queue_len )); then
      # Exhausted mid-skip — reshuffle
      _shader_shuffle "$playlist_file" > "$queue_file"
      idx=0
      queue_len="$(wc -l < "$queue_file" | tr -d ' ')"
    fi

    shader_name="$(sed -n "$((idx + 1))p" "$queue_file")"
    shader_path="$_shader_base_dir/$shader_name"
    idx=$((idx + 1))

    if [[ -f "$shader_path" ]]; then
      break
    fi
    attempts=$((attempts + 1))
  done

  # Save index
  printf '%s' "$idx" > "$idx_file"

  [[ -f "$shader_path" ]] || return 1

  # Determine animation setting
  local anim="false"
  _shader_needs_animation "$shader_path" && anim="true"

  # Atomic config update
  local tmp
  tmp="$(mktemp "${_ghostty_config}.XXXXXX")"
  sed -e "s|^custom-shader = .*|custom-shader = $shader_path|" \
      -e "s|^# custom-shader.*|custom-shader = $shader_path|" \
      -e "s|^custom-shader-animation = .*|custom-shader-animation = $anim|" \
      "$_ghostty_config" > "$tmp"
  command mv -f "$tmp" "$_ghostty_config"
}
