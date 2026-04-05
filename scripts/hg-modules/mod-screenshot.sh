#!/usr/bin/env bash
# mod-screenshot.sh — hg screenshot module
# Delegates to hg-screenshot.sh for all capture operations.

_SS_DIR="$HOME/Pictures/screenshots"
_SS_SCRIPT="$HG_DOTFILES/scripts/hg-screenshot.sh"

screenshot_description() {
  echo "Screenshots — crop, full, window, ocr, annotate, delay, record, list, open"
}

screenshot_commands() {
  cat <<'CMDS'
crop	Crop-select region screenshot → save + clipboard
full	Full screen or monitor screenshot [monitor]
window	Active window screenshot
ocr	Region → OCR → text to clipboard
annotate	Region → satty annotation
delay	Delayed capture [seconds, default 3]
record	Toggle screen recording
list	List recent screenshots
open	Open a recent screenshot [n]
CMDS
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
    crop)     exec bash "$_SS_SCRIPT" region --both ;;
    full)     exec bash "$_SS_SCRIPT" ${1:+monitor "$1"} ${1:---save} ${1:+--save} ;;
    window)   exec bash "$_SS_SCRIPT" window --save ;;
    ocr)      exec bash "$_SS_SCRIPT" ocr ;;
    annotate) exec bash "$_SS_SCRIPT" annotate ;;
    delay)    exec bash "$_SS_SCRIPT" delay "${1:-3}" --save ;;
    record)   exec bash "$_SS_SCRIPT" record ;;
    list)     _screenshot_list ;;
    open)     _screenshot_open "$@" ;;
    *)        hg_die "Unknown screenshot command: $cmd. Run 'hg screenshot --help'." ;;
  esac
}
