#!/usr/bin/env bash
# mod-media.sh — hg media module
# Terminal video & ASCII art tools — wraps tplay, chafa, timg, mpv, ascii-image-converter, yt-dlp

_CYBER_CHARMAP=" ░▒▓█"

media_description() {
  echo "Terminal video & ASCII art — play, convert, stream"
}

media_commands() {
  cat <<'CMDS'
play	Play video URL or file [--mode ascii|blocks|sixel|kitty|tct]
image	Display image in terminal [--mode ascii|braille|sixel|kitty|blocks]
ascii	ASCII art shortcut (auto-detects image vs video)
webcam	Stream webcam to terminal [--mode ascii|blocks]
yt	Search YouTube and play via fzf
tools	Show installed media tools and versions
CMDS
}

# ── Helpers ────────────────────────────────────

_is_url() {
  [[ "$1" =~ ^https?:// ]]
}

_is_youtube() {
  [[ "$1" =~ (youtube\.com|youtu\.be|youtube\.com/shorts) ]]
}

_is_image() {
  local ext="${1##*.}"
  ext="${ext,,}"
  [[ "$ext" =~ ^(png|jpg|jpeg|gif|bmp|webp|svg|tiff|ico|qoi)$ ]]
}

_parse_mode() {
  local mode=""
  local args=()
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --mode) mode="$2"; shift 2 ;;
      *)      args+=("$1"); shift ;;
    esac
  done
  echo "$mode"
  printf '%s\n' "${args[@]}"
}

# ── Subcommand: play ───────────────────────────

_media_play() {
  local mode="" target=""
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --mode) mode="$2"; shift 2 ;;
      *)      [[ -z "$target" ]] && target="$1"; shift ;;
    esac
  done

  [[ -n "$target" ]] || hg_die "Usage: hg media play <url|file> [--mode ascii|blocks|sixel|kitty|tct]"

  if _is_youtube "$target" || _is_url "$target"; then
    mode="${mode:-blocks}"
    case "$mode" in
      blocks) hg_require tplay;  exec tplay "$target" --char-map "$_CYBER_CHARMAP" ;;
      ascii)  hg_require tplay;  exec tplay "$target" ;;
      sixel)  hg_require yt-dlp chafa; yt-dlp -q -o - "$target" | exec chafa --format sixel - ;;
      kitty)  hg_require yt-dlp timg;  yt-dlp -q -o - "$target" | exec timg --pixelation=kitty - ;;
      tct)    hg_require mpv;    exec mpv --vo=tct --vo-tct-algo=half-blocks "$target" ;;
      *)      hg_die "Unknown mode: $mode (ascii|blocks|sixel|kitty|tct)" ;;
    esac
  else
    [[ -f "$target" ]] || hg_die "File not found: $target"
    mode="${mode:-tct}"
    case "$mode" in
      tct)    hg_require mpv;    exec mpv --vo=tct --vo-tct-algo=half-blocks "$target" ;;
      blocks) hg_require tplay;  exec tplay "$target" --char-map "$_CYBER_CHARMAP" ;;
      ascii)  hg_require tplay;  exec tplay "$target" ;;
      sixel)  hg_require chafa;  exec chafa --format sixel "$target" ;;
      kitty)  hg_require timg;   exec timg --pixelation=kitty "$target" ;;
      *)      hg_die "Unknown mode: $mode (ascii|blocks|sixel|kitty|tct)" ;;
    esac
  fi
}

# ── Subcommand: image ──────────────────────────

_media_image() {
  local mode="" target=""
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --mode) mode="$2"; shift 2 ;;
      *)      [[ -z "$target" ]] && target="$1"; shift ;;
    esac
  done

  [[ -n "$target" ]] || hg_die "Usage: hg media image <file> [--mode ascii|braille|sixel|kitty|blocks]"
  [[ -f "$target" ]] || hg_die "File not found: $target"

  mode="${mode:-ascii}"
  case "$mode" in
    ascii)   hg_require ascii-image-converter; exec ascii-image-converter -C "$target" ;;
    braille) hg_require ascii-image-converter; exec ascii-image-converter -C -b "$target" ;;
    sixel)   hg_require chafa;  exec chafa --format sixel "$target" ;;
    kitty)   hg_require timg;   exec timg --pixelation=kitty "$target" ;;
    blocks)  hg_require chafa;  exec chafa --format symbols "$target" ;;
    *)       hg_die "Unknown mode: $mode (ascii|braille|sixel|kitty|blocks)" ;;
  esac
}

# ── Subcommand: ascii ──────────────────────────

_media_ascii() {
  local target="${1:-}"
  [[ -n "$target" ]] || hg_die "Usage: hg media ascii <file|url>"

  if _is_url "$target"; then
    _media_play "$target" --mode blocks
  elif _is_image "$target"; then
    _media_image "$target" --mode ascii
  else
    _media_play "$target" --mode blocks
  fi
}

# ── Subcommand: webcam ─────────────────────────

_media_webcam() {
  hg_require tplay
  local mode="${1:-}"
  local device="/dev/video0"
  [[ -e "$device" ]] || hg_die "No webcam found at $device"

  case "$mode" in
    --mode)
      shift
      case "${1:-blocks}" in
        blocks) exec tplay "$device" --char-map "$_CYBER_CHARMAP" ;;
        ascii)  exec tplay "$device" ;;
        *)      hg_die "Unknown mode: $1 (ascii|blocks)" ;;
      esac
      ;;
    *)
      exec tplay "$device" --char-map "$_CYBER_CHARMAP"
      ;;
  esac
}

# ── Subcommand: yt ─────────────────────────────

_media_yt() {
  local query="$*"
  [[ -n "$query" ]] || hg_die "Usage: hg media yt <search query>"

  hg_require yt-dlp fzf tplay

  hg_info "Searching YouTube: $query"
  local results
  results="$(yt-dlp "ytsearch10:$query" --flat-playlist --print "%(id)s\t%(title)s\t%(duration_string)s" 2>/dev/null)"

  [[ -n "$results" ]] || hg_die "No results found"

  local pick
  pick="$(echo "$results" | fzf \
    --ansi --reverse --height=60% \
    --delimiter=$'\t' --with-nth=2.. \
    --header="Select a video (enter to play)" \
    --preview-window=hidden)"

  [[ -n "$pick" ]] || return 0

  local video_id
  video_id="$(echo "$pick" | cut -f1)"
  local url="https://youtube.com/watch?v=$video_id"
  hg_info "Playing: $(echo "$pick" | cut -f2)"
  exec tplay "$url" --char-map "$_CYBER_CHARMAP"
}

# ── Subcommand: tools ──────────────────────────

_media_tools() {
  printf "\n %s%sterminal media tools%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"

  local -a tools=(
    "tplay:ASCII video player — YouTube, local, webcam, streams"
    "chafa:Image/video → terminal (sixel, kitty, braille, blocks)"
    "timg:Terminal image/video viewer (sixel, kitty, halfblock)"
    "ascii-image-converter:Image → ASCII/braille art with color"
    "mpv:Video player (--vo=tct for terminal block art)"
    "yt-dlp:YouTube downloader (tplay backend)"
  )

  for entry in "${tools[@]}"; do
    local cmd="${entry%%:*}"
    local desc="${entry#*:}"
    local ver
    if command -v "$cmd" &>/dev/null; then
      ver="$("$cmd" --version 2>&1 | head -1 | grep -oE '[0-9]+\.[0-9]+[^ ]*' | head -1)"
      [[ -z "$ver" ]] && ver="installed"
      printf "  %s%-24s%s %s%-10s%s %s\n" "$HG_GREEN" "$cmd" "$HG_RESET" "$HG_DIM" "$ver" "$HG_RESET" "$desc"
    else
      printf "  %s%-24s%s %s%-10s%s %s\n" "$HG_RED" "$cmd" "$HG_RESET" "$HG_DIM" "missing" "$HG_RESET" "$desc"
    fi
  done
  printf "\n"
}

# ── Dispatch ───────────────────────────────────

media_run() {
  local cmd="${1:-}"
  shift || true

  case "$cmd" in
    play)    _media_play "$@" ;;
    image)   _media_image "$@" ;;
    ascii)   _media_ascii "$@" ;;
    webcam)  _media_webcam "$@" ;;
    yt)      _media_yt "$@" ;;
    tools)   _media_tools ;;
    *)       hg_die "Unknown media command: $cmd. Run 'hg media --help'." ;;
  esac
}
