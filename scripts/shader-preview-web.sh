#!/usr/bin/env bash
# shader-preview-web.sh — Launch hypr-shader-preview for a local shader
#
# Wraps h-banii/hypr-shader-preview so any shader in this repo (kitty
# DarkWindow, hyprshade, wallpaper, or a raw path) can be previewed in the
# browser with hot-reloading. DarkWindow shaders are transpiled to Hyprland
# (GLSL ES 300) format via scripts/lib/shader-transpile-hyprland.sh.
#
# Usage:
#   shader-preview-web.sh <shader>                 # auto-detect type
#   shader-preview-web.sh --type darkwindow amber-crt
#   shader-preview-web.sh --type raw my.frag
#   shader-preview-web.sh --image /path/to/bg.png crt
#   shader-preview-web.sh --no-browser sakura      # server+copy only
#   shader-preview-web.sh --stop                   # kill running dev server
#
# Flags:
#   --type <auto|darkwindow|hyprland|raw>   default: auto (detects by entrypoint)
#   --image <path>                          background image (copied to public/images/)
#   --size WxH                              preview resolution (default 1920x1080)
#   --port N                                default 5173
#   --no-browser                            skip xdg-open
#   --no-hide-buttons                       keep the UI buttons (record/etc.)
#   --stop                                  stop the background dev server
#   --status                                print running state

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOTFILES_DIR="${DOTFILES_DIR:-$(cd "$SCRIPT_DIR/.." && pwd)}"
TRANSPILER="$SCRIPT_DIR/lib/shader-transpile-hyprland.sh"
APP_DIR="${HYPR_SHADER_PREVIEW_DIR:-$HOME/.local/share/hypr-shader-preview}"
PID_FILE="${XDG_RUNTIME_DIR:-/tmp}/hypr-shader-preview.pid"
LOG_FILE="${XDG_RUNTIME_DIR:-/tmp}/hypr-shader-preview.log"

force_type="auto"
image=""
size="1920x1080"
port="5173"
open_browser=true
hide_buttons=true
action="preview"
shader=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    --type)             force_type="$2"; shift 2 ;;
    --image)            image="$2"; shift 2 ;;
    --size)             size="$2"; shift 2 ;;
    --port)             port="$2"; shift 2 ;;
    --no-browser)       open_browser=false; shift ;;
    --no-hide-buttons)  hide_buttons=false; shift ;;
    --stop)             action="stop"; shift ;;
    --status)           action="status"; shift ;;
    -h|--help)          sed -n '2,26p' "$0"; exit 0 ;;
    -*) echo "unknown flag: $1" >&2; exit 2 ;;
    *) shader="$1"; shift ;;
  esac
done

BASE_PATH="/hypr-shader-preview/"

server_up() {
  curl -sfL --max-time 1 "http://localhost:$port$BASE_PATH" >/dev/null 2>&1
}

stop_server() {
  if [[ -f "$PID_FILE" ]]; then
    pid="$(cat "$PID_FILE")"
    if kill -0 "$pid" 2>/dev/null; then
      pkill -P "$pid" 2>/dev/null || true
      kill "$pid" 2>/dev/null || true
      echo "stopped hypr-shader-preview (pid $pid)"
    fi
    rm -f "$PID_FILE"
  else
    echo "no pid file; not managed by this script"
  fi
}

ensure_app() {
  if [[ ! -d "$APP_DIR" ]]; then
    echo "hypr-shader-preview not found at $APP_DIR" >&2
    echo "clone with: git clone https://github.com/h-banii/hypr-shader-preview $APP_DIR" >&2
    exit 1
  fi
  if [[ ! -d "$APP_DIR/node_modules" ]]; then
    echo "installing npm deps in $APP_DIR ..." >&2
    (cd "$APP_DIR" && npm install) >&2
  fi
}

start_server() {
  if server_up; then
    echo "server already up at http://localhost:$port$BASE_PATH" >&2
    return 0
  fi
  echo "starting dev server (logs: $LOG_FILE) ..." >&2
  (cd "$APP_DIR" && nohup npm start -- --port "$port" >"$LOG_FILE" 2>&1 &
   echo "$!" >"$PID_FILE")
  for _ in $(seq 1 30); do
    if server_up; then return 0; fi
    sleep 0.5
  done
  echo "dev server failed to come up within 15s — see $LOG_FILE" >&2
  return 1
}

detect_type() {
  local f="$1"
  if   grep -qE '^\s*void\s+windowShader\s*\('  "$f" 2>/dev/null; then echo "darkwindow"
  elif grep -qE '^\s*void\s+mainImage\s*\('     "$f" 2>/dev/null; then echo "shadertoy"
  else echo "raw"
  fi
}

resolve_shader() {
  local name="$1"
  if [[ -f "$name" ]]; then echo "$name"; return; fi
  for cand in \
    "$DOTFILES_DIR/kitty/shaders/darkwindow/${name%.glsl}.glsl" \
    "$DOTFILES_DIR/hyprland/shaders/${name%.frag}.frag" \
    "$DOTFILES_DIR/hyprland/shaders/${name%.glsl}.glsl" \
    "$DOTFILES_DIR/wallpaper-shaders/${name%.frag}.frag"; do
    [[ -f "$cand" ]] && { echo "$cand"; return; }
  done
  return 1
}

case "$action" in
  stop)   stop_server;   exit 0 ;;
  status)
    if server_up; then
      echo "up: http://localhost:$port$BASE_PATH"
      [[ -f "$PID_FILE" ]] && echo "pid: $(cat "$PID_FILE")"
    else
      echo "down"
    fi
    exit 0
    ;;
esac

[[ -n "$shader" ]] || { echo "usage: shader-preview-web.sh [flags] <shader>" >&2; exit 2; }
src="$(resolve_shader "$shader")" || { echo "shader not found: $shader" >&2; exit 1; }

ensure_app

# Determine type + produce final frag
shader_type="$force_type"
[[ "$shader_type" == "auto" ]] && shader_type="$(detect_type "$src")"

target_name="_hg_preview.frag"
target="$APP_DIR/src/public/shaders/$target_name"

case "$shader_type" in
  darkwindow)
    echo "transpiling DarkWindow → Hyprland: $(basename "$src")" >&2
    "$TRANSPILER" --type darkwindow "$src" "$target"
    ;;
  shadertoy)
    echo "transpiling Shadertoy → Hyprland: $(basename "$src")" >&2
    "$TRANSPILER" --type shadertoy "$src" "$target"
    ;;
  hyprland|raw)
    cp "$src" "$target"
    ;;
  *)
    echo "unknown type: $shader_type" >&2; exit 2 ;;
esac

# Optional background image
img_param="default.png"
if [[ -n "$image" ]]; then
  if [[ -f "$image" ]]; then
    cp "$image" "$APP_DIR/src/public/images/_hg_bg.png"
    img_param="_hg_bg.png"
  else
    echo "warn: image not found: $image (using default.png)" >&2
  fi
fi

start_server

w="${size%x*}"
h="${size#*x}"
hb=""
$hide_buttons && hb="&hide_buttons=true"
url="http://localhost:$port$BASE_PATH?shader=$target_name&image=$img_param&width=$w&height=$h$hb"

echo "preview: $url"
if $open_browser; then
  (xdg-open "$url" >/dev/null 2>&1 &) || echo "(xdg-open failed — open $url manually)" >&2
fi
