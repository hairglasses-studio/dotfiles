#!/usr/bin/env bash
set -euo pipefail

# kitty-shader-playlist.sh — Kitty shader + theme rotation via Hypr-DarkWindow
# Canonical shader assets live in kitty/shaders/darkwindow/ (pre-transpiled
# from ghostty sources for the Hypr-DarkWindow plugin API) and are paired
# with Kitty theme playlists for per-spawn visual variation. The active
# shader is written to $_state_dir/darkwindow-active.glsl, which hyprland.conf
# registers as the `kitty-shader` named shader. Cycling triggers an
# `hyprctl reload` so DarkWindow re-compiles the updated file content.

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/notify.sh"

_dotfiles="$HG_DOTFILES"
_shader_dir="$_dotfiles/kitty/shaders/darkwindow"
_shader_playlist_dir="$_dotfiles/kitty/shaders/playlists"
_theme_playlist_dir="$_dotfiles/kitty/themes/playlists"
_state_dir="${XDG_STATE_HOME:-$HOME/.local/state}/kitty-shaders"
_theme_cache_dir="$_state_dir/theme-cache"
_default_playlist="ambient"
_current_shader="$_state_dir/current"
_current_theme="$_state_dir/current-theme"
_current_label="$_state_dir/current-label"
_current_theme_conf="$_state_dir/current-theme.conf"
_pending_theme="$_state_dir/pending-theme"
_auto_rotate_playlist="$_state_dir/auto-rotate-playlist"
_crtty_active="$_state_dir/darkwindow-active.glsl"
_current_position="$_state_dir/current-position"
_current_total="$_state_dir/current-total"
_current_selection_mode="$_state_dir/current-selection-mode"

mkdir -p "$_state_dir" "$_theme_cache_dir"

_active_playlist() {
  if [[ -f "$_auto_rotate_playlist" ]]; then
    tr -d '\n' < "$_auto_rotate_playlist"
  else
    printf '%s' "$_default_playlist"
  fi
}

_sanitize_theme_name() {
  printf '%s' "$1" | tr '[:upper:]' '[:lower:]' | tr -cs 'a-z0-9' '_'
}

_theme_cache_path() {
  printf '%s/%s.conf' "$_theme_cache_dir" "$(_sanitize_theme_name "$1")"
}

_dump_theme() {
  local theme="$1"
  local cache
  cache="$(_theme_cache_path "$theme")"
  if [[ ! -s "$cache" ]]; then
    local tmp="${cache}.tmp"
    if ! kitty +kitten themes --dump-theme "$theme" > "$tmp" 2>/dev/null; then
      rm -f "$tmp"
      return 1
    fi
    mv "$tmp" "$cache"
  fi
  printf '%s\n' "$cache"
}

_shader_name() {
  local name="$1"
  [[ "$name" == *.glsl ]] || name="${name}.glsl"
  printf '%s\n' "$name"
}

_shader_path() {
  local name
  name="$(_shader_name "$1")"
  local path="$_shader_dir/$name"
  [[ -f "$path" ]] || return 1
  printf '%s\n' "$path"
}

_playlist_file() {
  local kind="$1" playlist="$2"
  local base
  case "$kind" in
    shader) base="$_shader_playlist_dir" ;;
    theme) base="$_theme_playlist_dir" ;;
    *) return 1 ;;
  esac
  if [[ -f "$base/${playlist}.txt" ]]; then
    printf '%s/%s.txt\n' "$base" "$playlist"
    return 0
  fi
  if [[ "$kind" == "theme" && -f "$base/${_default_playlist}.txt" ]]; then
    printf '%s/%s.txt\n' "$base" "$_default_playlist"
    return 0
  fi
  return 1
}

_queue_file() {
  printf '%s/%s-%s.queue\n' "$_state_dir" "$1" "$2"
}

_idx_file() {
  printf '%s/%s-%s.idx\n' "$_state_dir" "$1" "$2"
}

_hash_file() {
  printf '%s/%s-%s.hash\n' "$_state_dir" "$1" "$2"
}

_shuffle_stdin() {
  local -a lines
  mapfile -t lines
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

_playlist_hash() {
  md5sum "$1" | cut -d' ' -f1
}

_ensure_queue() {
  local kind="$1" playlist="$2"
  local source_file queue_file idx_file hash_file
  source_file="$(_playlist_file "$kind" "$playlist")" || return 1
  queue_file="$(_queue_file "$kind" "$playlist")"
  idx_file="$(_idx_file "$kind" "$playlist")"
  hash_file="$(_hash_file "$kind" "$playlist")"

  local current_hash stored_hash="" idx=0 queue_len=0
  current_hash="$(_playlist_hash "$source_file")"
  [[ -f "$hash_file" ]] && stored_hash="$(< "$hash_file")"
  [[ -f "$idx_file" ]] && idx="$(< "$idx_file")"
  [[ "$idx" =~ ^[0-9]+$ ]] || idx=0
  [[ -f "$queue_file" ]] && queue_len="$(wc -l < "$queue_file" | tr -d ' ')"

  if [[ ! -f "$queue_file" ]] || [[ "$queue_len" -eq 0 ]] || [[ "$current_hash" != "$stored_hash" ]] || [[ "$idx" -ge "$queue_len" ]]; then
    awk 'NF && $1 !~ /^#/' "$source_file" | _shuffle_stdin > "$queue_file"
    printf '0' > "$idx_file"
    printf '%s' "$current_hash" > "$hash_file"
  fi
}

_entry_valid() {
  local kind="$1" entry="$2"
  case "$kind" in
    shader) _shader_path "$entry" >/dev/null ;;
    theme) _dump_theme "$entry" >/dev/null ;;
    *) return 1 ;;
  esac
}

_pick_next() {
  local kind="$1" playlist="$2"
  _ensure_queue "$kind" "$playlist" || return 1

  local queue_file idx_file
  queue_file="$(_queue_file "$kind" "$playlist")"
  idx_file="$(_idx_file "$kind" "$playlist")"

  local idx queue_len entry attempts=0
  idx="$(< "$idx_file")"
  queue_len="$(wc -l < "$queue_file" | tr -d ' ')"

  while (( attempts < queue_len + 4 )); do
    if (( idx >= queue_len )); then
      awk 'NF && $1 !~ /^#/' "$(_playlist_file "$kind" "$playlist")" | _shuffle_stdin > "$queue_file"
      idx=0
      queue_len="$(wc -l < "$queue_file" | tr -d ' ')"
    fi
    entry="$(sed -n "$((idx + 1))p" "$queue_file")"
    idx=$((idx + 1))
    printf '%s' "$idx" > "$idx_file"
    if _entry_valid "$kind" "$entry"; then
      printf '%s\n' "$entry"
      return 0
    fi
    attempts=$((attempts + 1))
  done
  return 1
}

_pick_prev() {
  local kind="$1" playlist="$2"
  _ensure_queue "$kind" "$playlist" || return 1

  local queue_file idx_file
  queue_file="$(_queue_file "$kind" "$playlist")"
  idx_file="$(_idx_file "$kind" "$playlist")"

  local idx queue_len entry attempts=0
  idx="$(< "$idx_file")"
  queue_len="$(wc -l < "$queue_file" | tr -d ' ')"

  while (( attempts < queue_len + 4 )); do
    idx=$(( (idx - 2 + queue_len) % queue_len ))
    entry="$(sed -n "$((idx + 1))p" "$queue_file")"
    printf '%s' "$((idx + 1))" > "$idx_file"
    if _entry_valid "$kind" "$entry"; then
      printf '%s\n' "$entry"
      return 0
    fi
    attempts=$((attempts + 1))
  done
  return 1
}

_pick_random() {
  local kind="$1" playlist="$2"
  local source_file
  source_file="$(_playlist_file "$kind" "$playlist")" || return 1

  local -a entries
  mapfile -t entries < <(awk 'NF && $1 !~ /^#/' "$source_file")
  local n=${#entries[@]}
  (( n > 0 )) || return 1

  local current="" entry attempts=0
  if [[ "$kind" == "shader" && -f "$_current_shader" ]]; then
    current="$(< "$_current_shader")"
  fi
  if [[ "$kind" == "theme" && -f "$_current_theme" ]]; then
    current="$(< "$_current_theme")"
  fi

  while (( attempts < n + 4 )); do
    entry="${entries[$(( RANDOM % n ))]}"
    if [[ "$entry" == "$current" ]]; then
      attempts=$((attempts + 1))
      continue
    fi
    if _entry_valid "$kind" "$entry"; then
      printf '%s\n' "$entry"
      return 0
    fi
    attempts=$((attempts + 1))
  done
  return 1
}

_pick_shader() {
  local direction="$1" playlist="$2"
  case "$direction" in
    next) _pick_next shader "$playlist" ;;
    prev) _pick_prev shader "$playlist" ;;
    random) _pick_random shader "$playlist" ;;
    *) return 1 ;;
  esac
}

_pick_theme() {
  local direction="$1" playlist="$2"
  case "$direction" in
    next) _pick_next theme "$playlist" ;;
    prev) _pick_prev theme "$playlist" ;;
    random) _pick_random theme "$playlist" ;;
    *) return 1 ;;
  esac
}

_state_label() {
  local shader="$1" theme="$2"
  if [[ -n "$shader" ]]; then
    printf '%s · %s\n' "$theme" "${shader%.glsl}"
  else
    printf '%s\n' "$theme"
  fi
}

_playlist_ordinal() {
  local kind="$1" playlist="$2" entry="$3"
  local source_file normalized idx=0 total=0 line
  source_file="$(_playlist_file "$kind" "$playlist")" || return 1
  normalized="${entry%.glsl}"
  while IFS= read -r line; do
    line="${line%%#*}"
    line="${line#"${line%%[![:space:]]*}"}"
    line="${line%"${line##*[![:space:]]}"}"
    [[ -n "$line" ]] || continue
    total=$((total + 1))
    if [[ "${line%.glsl}" == "$normalized" ]]; then
      idx=$total
    fi
  done < "$source_file"
  (( total > 0 )) || return 1
  printf '%s\t%s\n' "$idx" "$total"
}

_queue_progress() {
  local playlist="$1"
  local queue_file idx_file idx total
  queue_file="$(_queue_file shader "$playlist")"
  idx_file="$(_idx_file shader "$playlist")"
  [[ -f "$queue_file" && -f "$idx_file" ]] || return 1
  idx="$(< "$idx_file")"
  [[ "$idx" =~ ^[0-9]+$ ]] || return 1
  total="$(wc -l < "$queue_file" | tr -d ' ')"
  (( total > 0 )) || return 1
  printf '%s\t%s\n' "$idx" "$total"
}

_write_selection_state() {
  local mode="$1" playlist="$2"
  case "$mode" in
    queue)
      if IFS=$'\t' read -r position total < <(_queue_progress "$playlist"); then
        printf '%s' "$position" > "$_current_position"
        printf '%s' "$total" > "$_current_total"
      else
        rm -f "$_current_position" "$_current_total"
      fi
      ;;
    *)
      rm -f "$_current_position" "$_current_total"
      ;;
  esac
  printf '%s' "$mode" > "$_current_selection_mode"
}

_write_visual_state() {
  local shader="$1" theme="$2" playlist="$3" theme_conf="$4" mode="${5:-queue}"
  printf '%s' "$shader" > "$_current_shader"
  printf '%s' "$theme" > "$_current_theme"
  _state_label "$shader" "$theme" > "$_current_label"
  cp "$theme_conf" "$_current_theme_conf"
  cp "$(_shader_path "$shader")" "$_crtty_active"
  printf '%s' "$theme" > "$_pending_theme"
  printf '%s' "$playlist" > "$_auto_rotate_playlist"
  _write_selection_state "$mode" "$playlist"
  # Re-compile the darkwindow named shader from the updated file. Skip during
  # initial state-dir warm-up (no HYPRLAND_INSTANCE_SIGNATURE) so install and
  # headless paths remain quiet.
  if [[ -n "${HYPRLAND_INSTANCE_SIGNATURE:-}" ]] && command -v hyprctl >/dev/null 2>&1; then
    hyprctl reload >/dev/null 2>&1 || true
  fi
}

_theme_only_state() {
  local shader="${1:-}" theme="$2" theme_conf="$3"
  [[ -n "$shader" ]] && printf '%s' "$shader" > "$_current_shader"
  printf '%s' "$theme" > "$_current_theme"
  _state_label "$shader" "$theme" > "$_current_label"
  cp "$theme_conf" "$_current_theme_conf"
}

_try_apply_theme_to_active() {
  local theme_conf="$1"
  if [[ -n "${KITTY_WINDOW_ID:-}" ]]; then
    kitten @ set-colors "$theme_conf" >/dev/null 2>&1 || true
    return
  fi
  if [[ -n "${KITTY_LISTEN_ON:-}" ]]; then
    kitten @ --to "$KITTY_LISTEN_ON" set-colors --match state:focused_os_window "$theme_conf" >/dev/null 2>&1 || true
    return
  fi
  kitten @ --to unix:@mykitty set-colors --match state:focused_os_window "$theme_conf" >/dev/null 2>&1 || true
}

_choose_visual() {
  local direction="$1" playlist="$2"
  local shader theme theme_conf
  shader="$(_pick_shader "$direction" "$playlist")" || hg_die "No valid shaders in playlist: $playlist"
  theme="$(_pick_theme "$direction" "$playlist")" || hg_die "No valid themes in playlist: $playlist"
  theme_conf="$(_dump_theme "$theme")" || hg_die "Failed to resolve Kitty theme: $theme"
  printf '%s\t%s\t%s\n' "$shader" "$theme" "$theme_conf"
}

cmd_next() {
  local playlist="${1:-$(_active_playlist)}"
  local shader theme theme_conf
  IFS=$'\t' read -r shader theme theme_conf < <(_choose_visual next "$playlist")
  _write_visual_state "$shader" "$theme" "$playlist" "$theme_conf" queue
  _try_apply_theme_to_active "$theme_conf"
  hg_notify_low "Kitty Visual" "$(_state_label "$shader" "$theme")"
  hg_ok "$(_state_label "$shader" "$theme")"
}

cmd_prev() {
  local playlist="${1:-$(_active_playlist)}"
  local shader theme theme_conf
  IFS=$'\t' read -r shader theme theme_conf < <(_choose_visual prev "$playlist")
  _write_visual_state "$shader" "$theme" "$playlist" "$theme_conf" queue
  _try_apply_theme_to_active "$theme_conf"
  hg_notify_low "Kitty Visual" "$(_state_label "$shader" "$theme")"
  hg_ok "$(_state_label "$shader" "$theme")"
}

cmd_random() {
  local playlist="${1:-$(_active_playlist)}"
  local shader theme theme_conf
  IFS=$'\t' read -r shader theme theme_conf < <(_choose_visual random "$playlist")
  _write_visual_state "$shader" "$theme" "$playlist" "$theme_conf" random
  _try_apply_theme_to_active "$theme_conf"
  hg_notify_low "Kitty Visual" "$(_state_label "$shader" "$theme")"
  hg_ok "$(_state_label "$shader" "$theme")"
}

cmd_current() {
  [[ -f "$_current_shader" ]] || return 0
  local shader
  shader="$(< "$_current_shader")"
  printf '%s\n' "${shader%.glsl}"
}

cmd_theme_current() {
  [[ -f "$_current_theme" ]] || return 0
  cat "$_current_theme"
}

cmd_status() {
  local shader theme playlist label position total selection_mode
  if [[ -f "$_current_shader" ]]; then
    shader="$(< "$_current_shader")"
  else
    shader=""
  fi
  if [[ -f "$_current_theme" ]]; then
    theme="$(< "$_current_theme")"
  else
    theme=""
  fi
  playlist="$(_active_playlist)"
  label="$(_state_label "$shader" "$theme")"
  if [[ -f "$_current_selection_mode" ]]; then
    selection_mode="$(< "$_current_selection_mode")"
  else
    selection_mode=""
  fi
  hg_info "playlist: ${playlist:-$_default_playlist}"
  hg_info "shader:   ${shader%.glsl}"
  hg_info "theme:    $theme"
  hg_info "label:    $label"
  case "$selection_mode" in
    queue)
      if [[ -f "$_current_position" && -f "$_current_total" ]]; then
        hg_info "position: $(< "$_current_position")/$(< "$_current_total")"
      fi
      ;;
    random)
      hg_info "position: random"
      ;;
    manual)
      hg_info "position: custom"
      ;;
    *)
      if [[ -n "$shader" ]] && IFS=$'\t' read -r position total < <(_playlist_ordinal shader "$playlist" "$shader" 2>/dev/null); then
        hg_info "position: ${position}/${total}"
      fi
      ;;
  esac
}

cmd_set() {
  local shader_input="${1:-}"
  local theme_input="${2:-}"
  [[ -n "$shader_input" ]] || hg_die "Usage: kitty-shader-playlist set <shader> [theme]"

  local playlist="$(_active_playlist)"
  local shader theme theme_conf
  shader="$(_shader_name "$shader_input")"
  _shader_path "$shader" >/dev/null || hg_die "Shader not found: $shader"

  if [[ -n "$theme_input" ]]; then
    theme="$theme_input"
  elif [[ -f "$_current_theme" ]]; then
    theme="$(< "$_current_theme")"
  else
    theme="$(_pick_theme next "$playlist")"
  fi
  theme_conf="$(_dump_theme "$theme")" || hg_die "Failed to resolve Kitty theme: $theme"

  _write_visual_state "$shader" "$theme" "$playlist" "$theme_conf" manual
  _try_apply_theme_to_active "$theme_conf"
  hg_notify_low "Kitty Visual" "$(_state_label "$shader" "$theme")"
  hg_ok "$(_state_label "$shader" "$theme")"
}

cmd_list() {
  local playlist="${1:-$(_active_playlist)}"
  local shader_file theme_file
  shader_file="$(_playlist_file shader "$playlist")" || hg_die "Shader playlist not found: $playlist"
  theme_file="$(_playlist_file theme "$playlist")" || theme_file="$(_playlist_file theme "$_default_playlist")"

  printf '%s\n' "shader playlist: $playlist"
  awk 'NF && $1 !~ /^#/' "$shader_file"
  printf '\n%s\n' "theme playlist: ${playlist}"
  awk 'NF && $1 !~ /^#/' "$theme_file"
}

cmd_reset() {
  local playlist="${1:-$(_active_playlist)}"
  rm -f \
    "$(_queue_file shader "$playlist")" "$(_idx_file shader "$playlist")" "$(_hash_file shader "$playlist")" \
    "$(_queue_file theme "$playlist")" "$(_idx_file theme "$playlist")" "$(_hash_file theme "$playlist")"
  if [[ "$playlist" == "$(_active_playlist)" ]]; then
    rm -f "$_current_position" "$_current_total" "$_current_selection_mode"
  fi
  hg_ok "Reset Kitty visual playlist state: $playlist"
}

cmd_engine() {
  if command -v crtty >/dev/null 2>&1; then
    printf 'crtty\n'
  else
    printf 'kitty-only\n'
  fi
}

cmd_build() {
  "$_dotfiles/kitty/shaders/bin/shader-build.sh" "${1:-build}"
}

cmd_spawn() {
  local playlist="${1:-$(_active_playlist)}"
  shift || true
  if [[ "${1:-}" == "--" ]]; then
    shift || true
  fi

  local shader theme theme_conf
  IFS=$'\t' read -r shader theme theme_conf < <(_choose_visual next "$playlist")
  _write_visual_state "$shader" "$theme" "$playlist" "$theme_conf" queue
  # Shader application is handled by the Hypr-DarkWindow windowrule in
  # hyprland.conf — the new kitty window picks it up automatically once
  # _write_visual_state has reloaded the plugin with the updated file.
  exec kitty "$@"
}

cmd_theme_for_window() {
  local _window_id="${1:-}"
  local playlist="${2:-$(_active_playlist)}"
  local theme shader theme_conf
  if [[ -f "$_pending_theme" ]]; then
    theme="$(< "$_pending_theme")"
    rm -f "$_pending_theme"
  else
    theme="$(_pick_theme next "$playlist")" || hg_die "No valid Kitty themes in playlist: $playlist"
  fi
  theme_conf="$(_dump_theme "$theme")" || hg_die "Failed to resolve Kitty theme: $theme"
  if [[ -f "$_current_shader" ]]; then
    shader="$(< "$_current_shader")"
  else
    shader=""
  fi
  _theme_only_state "$shader" "$theme" "$theme_conf"
  printf '%s\t%s\t%s\n' "$theme" "$theme_conf" "$(_state_label "$shader" "$theme")"
}

cmd_help() {
  cat <<'EOF'
Usage: kitty-shader-playlist <command> [args]

Commands:
  spawn [playlist] [-- kitty args...]
      Launch kitty with the next shader in the playlist and queue the matching
      theme for the first new window.

  next [playlist]
  prev [playlist]
  random [playlist]
      Advance the shader + theme rotation and apply the theme to the active
      kitty window when remote control is available.

  set <shader> [theme]
      Force a shader and optionally a Kitty theme.

  current
      Print the active shader name without the .glsl suffix.

  theme-current
      Print the active Kitty theme name.

  theme-for-window <window-id> [playlist]
      Internal helper for the Kitty watcher.

  list [playlist]
      Print the shader and theme entries for a playlist.

  status
      Print the current playlist, shader, and theme.

  reset [playlist]
      Clear cached queue/index state for the playlist.

  engine
      Print either "crtty" or "kitty-only".

  build [mode]
      Run the Kitty shader build helper.
EOF
}

main() {
  local cmd="${1:-help}"
  shift || true
  case "$cmd" in
    spawn) cmd_spawn "$@" ;;
    next) cmd_next "$@" ;;
    prev) cmd_prev "$@" ;;
    random) cmd_random "$@" ;;
    set) cmd_set "$@" ;;
    current) cmd_current "$@" ;;
    theme-current) cmd_theme_current "$@" ;;
    theme-for-window) cmd_theme_for_window "$@" ;;
    list) cmd_list "$@" ;;
    status) cmd_status "$@" ;;
    reset) cmd_reset "$@" ;;
    engine) cmd_engine "$@" ;;
    build) cmd_build "$@" ;;
    help|-h|--help) cmd_help ;;
    *) hg_die "Unknown command: $cmd" ;;
  esac
}

main "$@"
