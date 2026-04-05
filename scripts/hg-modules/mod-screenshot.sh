#!/usr/bin/env bash
# mod-screenshot.sh — hg screenshot module
# Screenshots — crop, full, window, list, open

_SS_DIR="$HOME/Pictures/screenshots"
_SS_CROP="$HG_DOTFILES/scripts/screenshot-crop.sh"

screenshot_description() {
  echo "Screenshots — crop, full, window, list, open"
}

screenshot_commands() {
  cat <<'CMDS'
crop	Crop-select region screenshot
full	Full screen or monitor screenshot [monitor]
window	Active window screenshot
list	List recent screenshots
open	Open a recent screenshot [n]
CMDS
}

_screenshot_save_notify() {
  local filepath="$1"
  echo -n "$filepath" | wl-copy 2>/dev/null
  notify-send -a "Screenshot" -i "$filepath" "Screenshot saved" "$filepath" 2>/dev/null
  hg_ok "$filepath"
}

_screenshot_full() {
  hg_require grim wl-copy
  mkdir -p "$_SS_DIR"
  local monitor="${1:-}" filename filepath
  filename="$(date +%Y%m%d_%H%M%S).png"
  filepath="$_SS_DIR/$filename"
  if [[ -n "$monitor" ]]; then
    grim -o "$monitor" "$filepath" || hg_die "grim failed"
  else
    grim "$filepath" || hg_die "grim failed"
  fi
  _screenshot_save_notify "$filepath"
}

_screenshot_window() {
  hg_require grim jq wl-copy
  source "$HG_DOTFILES/scripts/lib/compositor.sh"
  mkdir -p "$_SS_DIR"
  local filename filepath json region
  filename="$(date +%Y%m%d_%H%M%S).png"
  filepath="$_SS_DIR/$filename"
  json="$(compositor_query activewindow 2>/dev/null)"
  [[ -n "$json" ]] || hg_die "No active window detected"

  local comp_type
  comp_type="$(compositor_type)"
  if [[ "$comp_type" == "hyprland" ]]; then
    local ax ay sx sy
    ax="$(echo "$json" | jq -r '.at[0]')"
    ay="$(echo "$json" | jq -r '.at[1]')"
    sx="$(echo "$json" | jq -r '.size[0]')"
    sy="$(echo "$json" | jq -r '.size[1]')"
    region="${ax},${ay} ${sx}x${sy}"
  else
    hg_die "Window screenshot not supported on $(compositor_type)"
  fi

  grim -g "$region" "$filepath" || hg_die "grim failed"
  _screenshot_save_notify "$filepath"
}

_screenshot_list() {
  [[ -d "$_SS_DIR" ]] || { hg_info "No screenshots yet"; return 0; }
  local files
  files="$(find "$_SS_DIR" -maxdepth 1 -name '*.png' -printf '%T@\t%p\n' 2>/dev/null | sort -rn | head -10 | cut -f2)"
  [[ -n "$files" ]] || { hg_info "No screenshots yet"; return 0; }
  printf "\n %s%srecent screenshots%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"
  local i=1
  echo "$files" | while read -r f; do
    local size
    size="$(du -h "$f" 2>/dev/null | cut -f1)"
    printf "  %s%2d%s  %s%-30s%s %s%s%s\n" "$HG_CYAN" "$i" "$HG_RESET" "$HG_BOLD" "$(basename "$f")" "$HG_RESET" "$HG_DIM" "$size" "$HG_RESET"
    i=$((i + 1))
  done
  printf "\n"
}

_screenshot_open() {
  local n="${1:-1}"
  local file
  file="$(ls -t "$_SS_DIR"/*.png 2>/dev/null | sed -n "${n}p")"
  [[ -n "$file" ]] || hg_die "No screenshot at position $n"
  hg_info "Opening: $(basename "$file")"
  xdg-open "$file" 2>/dev/null &
  disown
}

screenshot_run() {
  local cmd="${1:-}"
  shift || true

  case "$cmd" in
    crop)   exec bash "$_SS_CROP" ;;
    full)   _screenshot_full "$@" ;;
    window) _screenshot_window ;;
    list)   _screenshot_list ;;
    open)   _screenshot_open "$@" ;;
    *)      hg_die "Unknown screenshot command: $cmd. Run 'hg screenshot --help'." ;;
  esac
}
