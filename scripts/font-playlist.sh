#!/usr/bin/env bash
set -euo pipefail

# font-playlist.sh — Sequential font auditioning for Kitty
# Cycles through font playlists, updating Kitty config atomically.
# Kitty auto-reloads on config change — no restart needed.

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/notify.sh"

_kitty_config="$HG_DOTFILES/kitty/kitty.conf"
_playlist_dir="$HG_DOTFILES/kitty/fonts"
_state_dir="$HOME/.local/state/fonts"
_default_playlist="dense-tryout"

mkdir -p "$_state_dir"

_get_idx() {
  local playlist="$1"
  local idx_file="$_state_dir/${playlist}.idx"
  if [[ -f "$idx_file" ]]; then
    cat "$idx_file"
  else
    echo 0
  fi
}

_set_idx() {
  local playlist="$1" idx="$2"
  printf '%s' "$idx" > "$_state_dir/${playlist}.idx"
}

_load_playlist() {
  local playlist="$1"
  local file="$_playlist_dir/${playlist}.txt"
  [[ -f "$file" ]] || hg_die "Playlist not found: $file"
  # Strip comments and blank lines
  grep -v '^\s*#' "$file" | grep -v '^\s*$'
}

_count_entries() {
  local playlist="$1"
  _load_playlist "$playlist" | wc -l
}

_get_entry() {
  local playlist="$1" idx="$2"
  _load_playlist "$playlist" | sed -n "$((idx + 1))p"
}

_parse_family() { echo "$1" | cut -d'|' -f1; }
_parse_size()   { echo "$1" | cut -d'|' -f2; }

_apply_font() {
  local family="$1" size="$2"
  local bold italic bold_italic
  local tmp
  bold="${family} Bold"
  italic="${family} Italic"
  bold_italic="${family} Bold Italic"
  tmp="$(mktemp "${_kitty_config}.XXXXXX")"
  sed -e "s|^font_family .*|font_family      $family|" \
      -e "s|^bold_font .*|bold_font        $bold|" \
      -e "s|^italic_font .*|italic_font      $italic|" \
      -e "s|^bold_italic_font .*|bold_italic_font $bold_italic|" \
      -e "s|^font_size .*|font_size $size|" \
      "$_kitty_config" > "$tmp"
  mv -f "$tmp" "$_kitty_config"
  # Reload kitty via SIGUSR1
  kill -USR1 "$(pidof kitty)" 2>/dev/null || true
}

_current_font() {
  grep '^font_family ' "$_kitty_config" | head -1 | sed 's/^font_family  *//'
}

_current_size() {
  grep '^font_size ' "$_kitty_config" | head -1 | awk '{print $2}'
}

cmd_next() {
  local playlist="${1:-$_default_playlist}"
  local count idx entry family size
  count=$(_count_entries "$playlist")
  idx=$(_get_idx "$playlist")
  idx=$(( (idx + 1) % count ))
  _set_idx "$playlist" "$idx"
  entry=$(_get_entry "$playlist" "$idx")
  family=$(_parse_family "$entry")
  size=$(_parse_size "$entry")
  _apply_font "$family" "$size"
  hg_ok "$family @ ${size}pt  [$((idx + 1))/$count]"
  hg_notify_low "Font" "$family @ ${size}pt"
}

cmd_prev() {
  local playlist="${1:-$_default_playlist}"
  local count idx entry family size
  count=$(_count_entries "$playlist")
  idx=$(_get_idx "$playlist")
  idx=$(( (idx - 1 + count) % count ))
  _set_idx "$playlist" "$idx"
  entry=$(_get_entry "$playlist" "$idx")
  family=$(_parse_family "$entry")
  size=$(_parse_size "$entry")
  _apply_font "$family" "$size"
  hg_ok "$family @ ${size}pt  [$((idx + 1))/$count]"
  hg_notify_low "Font" "$family @ ${size}pt"
}

cmd_current() {
  local family size
  family=$(_current_font)
  size=$(_current_size)
  hg_info "$family @ ${size}pt"
}

cmd_set() {
  local family size
  family=$(_current_font)
  size=$(_current_size)
  hg_ok "Locked: $family @ ${size}pt"
  hg_info "Bold/italic faces now track the selected family automatically."
}

cmd_list() {
  local playlist="${1:-$_default_playlist}"
  local count idx i entry family size current_family
  count=$(_count_entries "$playlist")
  idx=$(_get_idx "$playlist")
  current_family=$(_current_font)

  hg_info "Playlist: $playlist ($count fonts)"
  for (( i=0; i<count; i++ )); do
    entry=$(_get_entry "$playlist" "$i")
    family=$(_parse_family "$entry")
    size=$(_parse_size "$entry")
    if [[ "$family" == "$current_family" ]]; then
      printf "  %s>%s %s%s%s @ %spt\n" "$HG_CYAN" "$HG_RESET" "$HG_BOLD" "$family" "$HG_RESET" "$size"
    elif [[ "$i" -eq "$idx" ]]; then
      printf "  %s*%s %s @ %spt\n" "$HG_DIM" "$HG_RESET" "$family" "$size"
    else
      printf "    %s @ %spt\n" "$family" "$size"
    fi
  done
}

cmd_reset() {
  local playlist="${1:-$_default_playlist}"
  _set_idx "$playlist" 0
  hg_ok "Reset $playlist index to 0"
}

case "${1:-}" in
  next)    cmd_next "${2:-}" ;;
  prev)    cmd_prev "${2:-}" ;;
  current) cmd_current ;;
  set)     cmd_set ;;
  list)    cmd_list "${2:-}" ;;
  reset)   cmd_reset "${2:-}" ;;
  *)
    cat <<EOF
Usage: font-playlist.sh <command> [playlist]

Commands:
  next  [playlist]  Advance to next font (default: $_default_playlist)
  prev  [playlist]  Go back one font
  current           Show current font + size
  set               Lock current font as your default
  list  [playlist]  Show all entries (> = active, * = index)
  reset [playlist]  Reset playlist index to 0
EOF
    ;;
esac
