#!/usr/bin/env bash
set -euo pipefail

# kitty-shader-playlist.sh — Dual-engine shader playlist for kitty terminal
# Drives both CRTty (LD_PRELOAD) and Hypr-DarkWindow (compositor) shader backends.
# Source shaders live in ghostty/shaders/ (Ghostty format) and are transpiled
# to kitty/shaders/crtty/ and kitty/shaders/darkwindow/ by shader-build.sh.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/notify.sh"

_dotfiles="$HG_DOTFILES"
_playlist_dir="$_dotfiles/ghostty/shaders/playlists"
_crtty_dir="$_dotfiles/kitty/shaders/crtty"
_darkwindow_dir="$_dotfiles/kitty/shaders/darkwindow"
_state_dir="$HOME/.local/state/kitty-shaders"
_crtty_active="$_state_dir/crtty-active.glsl"
_darkwindow_active="$_state_dir/darkwindow-active.glsl"
_default_playlist="ambient"

mkdir -p "$_state_dir"

# --- Engine management ---

_get_engine() {
  local engine_file="$_state_dir/engine"
  if [[ -f "$engine_file" ]]; then
    cat "$engine_file"
  else
    echo "both"
  fi
}

_set_engine() {
  printf '%s' "$1" > "$_state_dir/engine"
}

# --- Fisher-Yates shuffle ---

_shuffle() {
  local file="$1"
  local -a lines
  mapfile -t lines < "$file"
  local n=${#lines[@]}
  local i j tmp
  for (( i = n - 1; i > 0; i-- )); do
    j=$(( RANDOM % (i + 1) ))
    tmp="${lines[$i]}"
    lines[$i]="${lines[$j]}"
    lines[$j]="$tmp"
  done
  printf '%s\n' "${lines[@]}"
}

# --- Playlist pick (returns shader name on stdout) ---

_playlist_pick() {
  local name="$1"
  [[ -z "$name" ]] && return 1

  local playlist_file="$_playlist_dir/${name}.txt"
  [[ -f "$playlist_file" ]] || { hg_error "Playlist not found: $name"; return 1; }

  local queue_file="$_state_dir/${name}.queue"
  local idx_file="$_state_dir/${name}.idx"
  local hash_file="$_state_dir/${name}.hash"

  # Hash playlist to detect edits
  local current_hash
  current_hash="$(md5sum "$playlist_file" | cut -d' ' -f1)"
  local stored_hash=""
  [[ -f "$hash_file" ]] && stored_hash="$(< "$hash_file")"

  # Read index
  local idx=0
  if [[ -f "$idx_file" ]]; then
    idx="$(< "$idx_file")"
    [[ "$idx" =~ ^[0-9]+$ ]] || idx=0
  fi

  # Count queue
  local queue_len=0
  [[ -f "$queue_file" ]] && queue_len="$(wc -l < "$queue_file" | tr -d ' ')"

  # Reshuffle if needed
  if [[ ! -f "$queue_file" ]] || [[ "$queue_len" -eq 0 ]] || \
     [[ "$current_hash" != "$stored_hash" ]] || [[ "$idx" -ge "$queue_len" ]]; then
    _shuffle "$playlist_file" > "$queue_file"
    idx=0
    printf '%s' "$current_hash" > "$hash_file"
    queue_len="$(wc -l < "$queue_file" | tr -d ' ')"
  fi

  # Pick shader, skip missing (max 5 retries)
  local shader_name attempts=0
  while (( attempts < 5 )); do
    if (( idx >= queue_len )); then
      _shuffle "$playlist_file" > "$queue_file"
      idx=0
      queue_len="$(wc -l < "$queue_file" | tr -d ' ')"
    fi
    shader_name="$(sed -n "$((idx + 1))p" "$queue_file")"
    idx=$((idx + 1))

    # Check shader exists in at least one backend
    if [[ -f "$_crtty_dir/$shader_name" ]] || [[ -f "$_darkwindow_dir/$shader_name" ]]; then
      break
    fi
    attempts=$((attempts + 1))
  done

  printf '%s' "$idx" > "$idx_file"
  printf '%s' "$shader_name"
}

# --- Apply shader to active backends ---

_apply_shader() {
  local name="$1"
  local engine
  engine="$(_get_engine)"

  case "$engine" in
    crtty)
      if [[ -f "$_crtty_dir/$name" ]]; then
        cp "$_crtty_dir/$name" "$_crtty_active"
      fi
      ;;
    darkwindow)
      if [[ -f "$_darkwindow_dir/$name" ]]; then
        cp "$_darkwindow_dir/$name" "$_darkwindow_active"
      fi
      ;;
    both)
      [[ -f "$_crtty_dir/$name" ]] && cp "$_crtty_dir/$name" "$_crtty_active"
      [[ -f "$_darkwindow_dir/$name" ]] && cp "$_darkwindow_dir/$name" "$_darkwindow_active"
      ;;
  esac

  # State for eww bar
  printf '%s' "$name" > "$_state_dir/current"
  hg_notify_low "Shader" "${name%.glsl}"
}

# --- Commands ---

cmd_next() {
  local playlist="${1:-$_default_playlist}"
  local shader_name
  shader_name="$(_playlist_pick "$playlist")" || return 1
  _apply_shader "$shader_name"
  hg_ok "${shader_name%.glsl}  [engine: $(_get_engine)]"
}

cmd_prev() {
  local playlist="${1:-$_default_playlist}"
  local idx_file="$_state_dir/${playlist}.idx"
  local queue_file="$_state_dir/${playlist}.queue"

  [[ -f "$idx_file" && -f "$queue_file" ]] || { cmd_next "$playlist"; return; }

  local idx queue_len shader_name
  idx="$(< "$idx_file")"
  queue_len="$(wc -l < "$queue_file" | tr -d ' ')"
  idx=$(( (idx - 2 + queue_len) % queue_len ))
  shader_name="$(sed -n "$((idx + 1))p" "$queue_file")"
  printf '%s' "$((idx + 1))" > "$idx_file"

  _apply_shader "$shader_name"
  hg_ok "${shader_name%.glsl}  [engine: $(_get_engine)]"
}

cmd_current() {
  if [[ -f "$_state_dir/current" ]]; then
    local name
    name="$(< "$_state_dir/current")"
    hg_info "${name%.glsl}  [engine: $(_get_engine)]"
  else
    hg_info "No shader active"
  fi
}

cmd_set() {
  local name="$1"
  [[ -z "$name" ]] && { hg_error "Usage: kitty-shader-playlist set <name.glsl>"; return 1; }
  # Add .glsl extension if missing
  [[ "$name" == *.glsl ]] || name="${name}.glsl"

  if [[ ! -f "$_crtty_dir/$name" && ! -f "$_darkwindow_dir/$name" ]]; then
    hg_error "Shader not found: $name"
    return 1
  fi

  _apply_shader "$name"
  hg_ok "Set: ${name%.glsl}  [engine: $(_get_engine)]"
}

cmd_random() {
  local engine
  engine="$(_get_engine)"
  local shader_dir="$_crtty_dir"
  [[ "$engine" == "darkwindow" ]] && shader_dir="$_darkwindow_dir"

  local -a shaders
  mapfile -t shaders < <(ls "$shader_dir"/*.glsl 2>/dev/null | xargs -I{} basename {})
  local n=${#shaders[@]}
  [[ "$n" -eq 0 ]] && { hg_error "No shaders found"; return 1; }

  # Avoid repeat
  local current=""
  [[ -f "$_state_dir/current" ]] && current="$(< "$_state_dir/current")"

  local pick attempts=0
  while (( attempts < 5 )); do
    pick="${shaders[$(( RANDOM % n ))]}"
    [[ "$pick" != "$current" ]] && break
    attempts=$((attempts + 1))
  done

  _apply_shader "$pick"
  hg_ok "${pick%.glsl}  [engine: $engine]"
}

cmd_list() {
  local playlist="${1:-$_default_playlist}"
  local playlist_file="$_playlist_dir/${playlist}.txt"
  [[ -f "$playlist_file" ]] || { hg_error "Playlist not found: $playlist"; return 1; }

  local current=""
  [[ -f "$_state_dir/current" ]] && current="$(< "$_state_dir/current")"

  local count idx_file idx=0
  count="$(grep -cv '^\s*$' "$playlist_file" || true)"
  idx_file="$_state_dir/${playlist}.idx"
  [[ -f "$idx_file" ]] && idx="$(< "$idx_file")"

  hg_info "Playlist: $playlist ($count shaders) [engine: $(_get_engine)]"
  local i=0
  while IFS= read -r name; do
    [[ -z "$name" ]] && continue
    if [[ "$name" == "$current" ]]; then
      printf "  ${HG_CYAN}>${HG_RESET} ${HG_BOLD}%s${HG_RESET}\n" "${name%.glsl}"
    elif [[ "$i" -eq "$((idx - 1))" ]]; then
      printf "  ${HG_DIM}*${HG_RESET} %s\n" "${name%.glsl}"
    else
      printf "    %s\n" "${name%.glsl}"
    fi
    i=$((i + 1))
  done < "$playlist_file"
}

cmd_reset() {
  local playlist="${1:-$_default_playlist}"
  printf '0' > "$_state_dir/${playlist}.idx"
  rm -f "$_state_dir/${playlist}.queue" "$_state_dir/${playlist}.hash"
  hg_ok "Reset playlist: $playlist"
}

cmd_engine() {
  local arg="${1:-}"
  case "$arg" in
    crtty|darkwindow|both)
      _set_engine "$arg"
      hg_ok "Engine: $arg"
      ;;
    cycle)
      local current
      current="$(_get_engine)"
      case "$current" in
        both)       _set_engine "crtty"; hg_ok "Engine: crtty" ;;
        crtty)      _set_engine "darkwindow"; hg_ok "Engine: darkwindow" ;;
        darkwindow) _set_engine "both"; hg_ok "Engine: both" ;;
      esac
      ;;
    "")
      hg_info "Engine: $(_get_engine)"
      ;;
    *)
      hg_error "Unknown engine: $arg (use crtty, darkwindow, both, or cycle)"
      return 1
      ;;
  esac
}

cmd_build() {
  "$_dotfiles/kitty/shaders/bin/shader-build.sh" "${1:-build}"
}

# --- Dispatch ---

case "${1:-}" in
  next)    cmd_next "${2:-}" ;;
  prev)    cmd_prev "${2:-}" ;;
  current) cmd_current ;;
  set)     cmd_set "${2:-}" ;;
  list)    cmd_list "${2:-}" ;;
  random)  cmd_random ;;
  reset)   cmd_reset "${2:-}" ;;
  engine)  cmd_engine "${2:-}" ;;
  build)   cmd_build "${2:-}" ;;
  *)
    cat <<EOF
Usage: kitty-shader-playlist <command> [args]

Commands:
  next  [playlist]           Advance to next shader (default: $_default_playlist)
  prev  [playlist]           Go back one shader
  current                    Show active shader + engine
  set   <name>               Apply a specific shader
  list  [playlist]           Show playlist contents
  random                     Pick a random shader
  reset [playlist]           Reset playlist state
  engine [crtty|darkwindow|both|cycle]  Get/set shader backend
  build  [build|check|clean] Transpile shaders
EOF
    ;;
esac
