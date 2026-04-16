#!/usr/bin/env bash
set -euo pipefail

# kitty-shader-playlist.sh — Kitty theme rotation (legacy name retained)
#
# Historically this script wrapped kitty launches in `crtty -s <shader>` and
# later drove the Hypr-DarkWindow plugin to apply CRT shaders per window.
# Shaders are no longer applied to kitty at all — every new window instead
# gets a unique dark theme pulled from a ~320-entry community playlist. The
# script name is preserved only because hyprland, ironbar, install.sh, and
# mod-shader.sh already reference it by path.
#
# All commands operate on `kitty/themes/playlists/<name>.txt`, which is
# generated from `kitty/themes/themes.json` via
# `scripts/gen-kitty-theme-playlist.sh`.

SCRIPT_PATH="$(readlink -f "${BASH_SOURCE[0]:-$0}")"
SCRIPT_DIR="$(cd "$(dirname "$SCRIPT_PATH")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/notify.sh"

_dotfiles="$HG_DOTFILES"
_theme_playlist_dir="$_dotfiles/kitty/themes/playlists"
_state_dir="${XDG_STATE_HOME:-$HOME/.local/state}/kitty-shaders"
_theme_cache_dir="$_state_dir/theme-cache"
_default_playlist="ambient"
_current_theme="$_state_dir/current-theme"
_current_label="$_state_dir/current-label"
_current_theme_conf="$_state_dir/current-theme.conf"
_auto_rotate_playlist="$_state_dir/auto-rotate-playlist"
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

_playlist_file() {
  local playlist="$1"
  if [[ -f "$_theme_playlist_dir/${playlist}.txt" ]]; then
    printf '%s/%s.txt\n' "$_theme_playlist_dir" "$playlist"
    return 0
  fi
  if [[ -f "$_theme_playlist_dir/${_default_playlist}.txt" ]]; then
    printf '%s/%s.txt\n' "$_theme_playlist_dir" "$_default_playlist"
    return 0
  fi
  return 1
}

_queue_file() { printf '%s/theme-%s.queue\n' "$_state_dir" "$1"; }
_idx_file()   { printf '%s/theme-%s.idx\n'   "$_state_dir" "$1"; }
_hash_file()  { printf '%s/theme-%s.hash\n'  "$_state_dir" "$1"; }

_shuffle_stdin() {
  local -a lines
  mapfile -t lines
  local n=${#lines[@]}
  local i j tmp
  for (( i = n - 1; i > 0; i-- )); do
    j=$(( RANDOM % (i + 1) ))
    tmp="${lines[i]}"
    lines[i]="${lines[j]}"
    lines[j]="$tmp"
  done
  printf '%s\n' "${lines[@]}"
}

_playlist_hash() {
  md5sum "$1" | cut -d' ' -f1
}

_read_playlist() {
  awk 'NF && $1 !~ /^#/' "$1"
}

_ensure_queue() {
  local playlist="$1"
  local source_file queue_file idx_file hash_file
  source_file="$(_playlist_file "$playlist")" || return 1
  queue_file="$(_queue_file "$playlist")"
  idx_file="$(_idx_file "$playlist")"
  hash_file="$(_hash_file "$playlist")"

  local current_hash stored_hash="" idx=0 queue_len=0
  current_hash="$(_playlist_hash "$source_file")"
  [[ -f "$hash_file" ]] && stored_hash="$(< "$hash_file")"
  [[ -f "$idx_file" ]] && idx="$(< "$idx_file")"
  [[ "$idx" =~ ^[0-9]+$ ]] || idx=0
  [[ -f "$queue_file" ]] && queue_len="$(wc -l < "$queue_file" | tr -d ' ')"

  if [[ ! -f "$queue_file" ]] || [[ "$queue_len" -eq 0 ]] || [[ "$current_hash" != "$stored_hash" ]] || [[ "$idx" -ge "$queue_len" ]]; then
    _read_playlist "$source_file" | _shuffle_stdin > "$queue_file"
    printf '0' > "$idx_file"
    printf '%s' "$current_hash" > "$hash_file"
  fi
}

_entry_valid() {
  _dump_theme "$1" >/dev/null
}

_pick_next() {
  local playlist="$1"
  _ensure_queue "$playlist" || return 1

  local queue_file idx_file
  queue_file="$(_queue_file "$playlist")"
  idx_file="$(_idx_file "$playlist")"

  local idx queue_len entry attempts=0
  idx="$(< "$idx_file")"
  queue_len="$(wc -l < "$queue_file" | tr -d ' ')"

  while (( attempts < queue_len + 4 )); do
    if (( idx >= queue_len )); then
      _read_playlist "$(_playlist_file "$playlist")" | _shuffle_stdin > "$queue_file"
      idx=0
      queue_len="$(wc -l < "$queue_file" | tr -d ' ')"
    fi
    entry="$(sed -n "$((idx + 1))p" "$queue_file")"
    idx=$((idx + 1))
    printf '%s' "$idx" > "$idx_file"
    if _entry_valid "$entry"; then
      printf '%s\n' "$entry"
      return 0
    fi
    attempts=$((attempts + 1))
  done
  return 1
}

_pick_prev() {
  local playlist="$1"
  _ensure_queue "$playlist" || return 1

  local queue_file idx_file
  queue_file="$(_queue_file "$playlist")"
  idx_file="$(_idx_file "$playlist")"

  local idx queue_len entry attempts=0
  idx="$(< "$idx_file")"
  queue_len="$(wc -l < "$queue_file" | tr -d ' ')"

  while (( attempts < queue_len + 4 )); do
    idx=$(( (idx - 2 + queue_len) % queue_len ))
    entry="$(sed -n "$((idx + 1))p" "$queue_file")"
    printf '%s' "$((idx + 1))" > "$idx_file"
    if _entry_valid "$entry"; then
      printf '%s\n' "$entry"
      return 0
    fi
    attempts=$((attempts + 1))
  done
  return 1
}

_pick_random() {
  local playlist="$1"
  local source_file
  source_file="$(_playlist_file "$playlist")" || return 1

  local -a entries
  mapfile -t entries < <(_read_playlist "$source_file")
  local n=${#entries[@]}
  (( n > 0 )) || return 1

  local current="" entry attempts=0
  [[ -f "$_current_theme" ]] && current="$(< "$_current_theme")"

  while (( attempts < n + 4 )); do
    entry="${entries[$(( RANDOM % n ))]}"
    if [[ "$entry" == "$current" ]]; then
      attempts=$((attempts + 1))
      continue
    fi
    if _entry_valid "$entry"; then
      printf '%s\n' "$entry"
      return 0
    fi
    attempts=$((attempts + 1))
  done
  return 1
}

_pick_direction() {
  local direction="$1" playlist="$2"
  case "$direction" in
    next) _pick_next "$playlist" ;;
    prev) _pick_prev "$playlist" ;;
    random) _pick_random "$playlist" ;;
    *) return 1 ;;
  esac
}

_playlist_ordinal() {
  local playlist="$1" entry="$2"
  local source_file idx=0 total=0 line
  source_file="$(_playlist_file "$playlist")" || return 1
  while IFS= read -r line; do
    line="${line%%#*}"
    line="${line#"${line%%[![:space:]]*}"}"
    line="${line%"${line##*[![:space:]]}"}"
    [[ -n "$line" ]] || continue
    total=$((total + 1))
    [[ "$line" == "$entry" ]] && idx=$total
  done < "$source_file"
  (( total > 0 )) || return 1
  printf '%s\t%s\n' "$idx" "$total"
}

_queue_progress() {
  local playlist="$1"
  local queue_file idx_file idx total
  queue_file="$(_queue_file "$playlist")"
  idx_file="$(_idx_file "$playlist")"
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

_write_theme_state() {
  local theme="$1" playlist="$2" theme_conf="$3" mode="${4:-queue}"
  printf '%s' "$theme" > "$_current_theme"
  printf '%s\n' "$theme" > "$_current_label"
  cp "$theme_conf" "$_current_theme_conf"
  printf '%s' "$playlist" > "$_auto_rotate_playlist"
  _write_selection_state "$mode" "$playlist"
}

_try_apply_theme_to_active() {
  local theme_conf="$1"
  if [[ -n "${KITTY_WINDOW_ID:-}" ]]; then
    kitten @ set-colors "$theme_conf" >/dev/null 2>&1 || true
    return
  fi
  local listen="${KITTY_LISTEN_ON:-unix:@mykitty}"
  kitten @ --to "$listen" set-colors --match state:focused "$theme_conf" >/dev/null 2>&1 || true
}

_choose_theme() {
  local direction="$1" playlist="$2"
  local theme theme_conf
  theme="$(_pick_direction "$direction" "$playlist")" || hg_die "No valid themes in playlist: $playlist"
  theme_conf="$(_dump_theme "$theme")" || hg_die "Failed to resolve kitty theme: $theme"
  printf '%s\t%s\n' "$theme" "$theme_conf"
}

_cycle() {
  local direction="$1" playlist mode theme theme_conf
  playlist="$(_active_playlist)"
  [[ "$direction" == "random" ]] && mode="random" || mode="queue"
  IFS=$'\t' read -r theme theme_conf < <(_choose_theme "$direction" "$playlist")
  _write_theme_state "$theme" "$playlist" "$theme_conf" "$mode"
  _try_apply_theme_to_active "$theme_conf"
  hg_notify_low "Kitty Theme" "$theme"
  hg_ok "$theme"
}

cmd_next()   { _cycle next;   }
cmd_prev()   { _cycle prev;   }
cmd_random() { _cycle random; }

cmd_current() {
  [[ -f "$_current_theme" ]] || return 0
  cat "$_current_theme"
}

cmd_theme_current() { cmd_current; }

cmd_status() {
  local theme playlist selection_mode
  playlist="$(_active_playlist)"
  theme=""
  [[ -f "$_current_theme" ]] && theme="$(< "$_current_theme")"
  selection_mode=""
  [[ -f "$_current_selection_mode" ]] && selection_mode="$(< "$_current_selection_mode")"

  hg_info "playlist: ${playlist:-$_default_playlist}"
  hg_info "theme:    ${theme:-<none>}"
  case "$selection_mode" in
    queue)
      if [[ -f "$_current_position" && -f "$_current_total" ]]; then
        hg_info "position: $(< "$_current_position")/$(< "$_current_total")"
      fi
      ;;
    random) hg_info "position: random" ;;
    manual) hg_info "position: custom" ;;
    *)
      local position total
      if [[ -n "$theme" ]] && IFS=$'\t' read -r position total < <(_playlist_ordinal "$playlist" "$theme" 2>/dev/null); then
        hg_info "position: ${position}/${total}"
      fi
      ;;
  esac
}

cmd_set() {
  local theme="${1:-}"
  [[ -n "$theme" ]] || hg_die "Usage: kitty-shader-playlist set <theme>"
  local theme_conf
  theme_conf="$(_dump_theme "$theme")" || hg_die "Unknown kitty theme: $theme"
  local playlist
  playlist="$(_active_playlist)"
  _write_theme_state "$theme" "$playlist" "$theme_conf" manual
  _try_apply_theme_to_active "$theme_conf"
  hg_notify_low "Kitty Theme" "$theme"
  hg_ok "$theme"
}

cmd_list() {
  local playlist="${1:-$(_active_playlist)}"
  local source_file
  source_file="$(_playlist_file "$playlist")" || hg_die "Theme playlist not found: $playlist"
  printf 'theme playlist: %s\n' "$playlist"
  _read_playlist "$source_file"
}

cmd_reset() {
  local playlist="${1:-$(_active_playlist)}"
  rm -f "$(_queue_file "$playlist")" "$(_idx_file "$playlist")" "$(_hash_file "$playlist")"
  if [[ "$playlist" == "$(_active_playlist)" ]]; then
    rm -f "$_current_position" "$_current_total" "$_current_selection_mode"
  fi
  hg_ok "Reset kitty theme playlist state: $playlist"
}

cmd_spawn() {
  local playlist="${1:-$(_active_playlist)}"
  shift || true
  if [[ "${1:-}" == "--" ]]; then
    shift || true
  fi
  _ensure_queue "$playlist" >/dev/null 2>&1 || true
  exec kitty "$@"
}

cmd_theme_for_window() {
  local _window_id="${1:-}"
  local playlist="${2:-$(_active_playlist)}"
  local theme theme_conf
  theme="$(_pick_next "$playlist")" || hg_die "No valid kitty themes in playlist: $playlist"
  theme_conf="$(_dump_theme "$theme")" || hg_die "Failed to resolve kitty theme: $theme"
  _write_theme_state "$theme" "$playlist" "$theme_conf" queue
  printf '%s\t%s\t%s\n' "$theme" "$theme_conf" "$theme"
}

cmd_help() {
  cat <<'EOF'
Usage: kitty-shader-playlist <command> [args]

Commands:
  spawn [playlist] [-- kitty args...]
      Launch a fresh kitty OS window. No shader is applied; a unique theme
      is picked per newly created kitten Window by the random_theme watcher.

  next [playlist]
  prev [playlist]
  random [playlist]
      Advance / rewind / randomize the theme rotation on the currently
      focused kitty window.

  set <theme>
      Apply a specific theme (by display name) to the focused kitty window.

  current
  theme-current
      Print the active theme name.

  theme-for-window <window-id> [playlist]
      Internal helper for kitty/watchers/random_theme.py. Picks the next
      theme from the queue and returns: <theme>\t<conf-path>\t<label>.

  list [playlist]
      Print the theme entries in a playlist.

  status
      Print the current playlist, theme, and queue position.

  reset [playlist]
      Clear the shuffled queue state for a playlist.
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
    help|-h|--help) cmd_help ;;
    *) hg_die "Unknown command: $cmd" ;;
  esac
}

main "$@"
