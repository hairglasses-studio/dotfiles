#!/usr/bin/env bash
# hg-screenshot.sh — Unified screenshot capture for all entry points
# Single abstraction layer: keybinds, MCP tools, hg CLI, gamepad all call this.
# Backend: wayshot (zwlr_screencopy_v1). Swap _capture() to change backends.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"
source "$SCRIPT_DIR/lib/compositor.sh"
source "$SCRIPT_DIR/lib/notify.sh"

SS_DIR="$HOME/Pictures/screenshots"
REC_PIDFILE="/tmp/hg-recording.pid"
REC_DIR="$HOME/Videos/recordings"

# ── Backend abstraction ─────────────────────────
# All captures go through here. Change this one function to swap backends.
_capture() {
  local outpath="$1" region="${2:-}" monitor="${3:-}"
  local -a args=()
  if [[ -n "$region" ]]; then
    args+=(-s "$region")
  fi
  if [[ -n "$monitor" ]]; then
    args+=(-o "$monitor")
  fi
  args+=(-f "$outpath")
  wayshot "${args[@]}" 2>/dev/null
}

_capture_stdout() {
  local region="${1:-}" monitor="${2:-}"
  local -a args=(--stdout)
  if [[ -n "$region" ]]; then
    args+=(-s "$region")
  fi
  if [[ -n "$monitor" ]]; then
    args+=(-o "$monitor")
  fi
  wayshot "${args[@]}" 2>/dev/null
}

_grim_supports_window_capture() {
  command -v grim >/dev/null 2>&1 || return 1
  grim --help 2>&1 | grep -q -- '-w'
}

# ── Post-processing ─────────────────────────────
_timestamp() { date +%Y%m%d_%H%M%S; }

_save_file() {
  local tmpfile="$1"
  mkdir -p "$SS_DIR"
  local dest="$SS_DIR/$(_timestamp).png"
  mv -f "$tmpfile" "$dest"
  echo "$dest"
}

_to_clipboard() {
  local file="$1"
  wl-copy < "$file" 2>/dev/null
}

_post_process() {
  local tmpfile="$1" do_save="$2" do_clip="$3" do_notify="$4" explicit_output="$5"
  local final="$tmpfile"

  if [[ -n "$explicit_output" ]]; then
    mkdir -p "$(dirname "$explicit_output")"
    mv -f "$tmpfile" "$explicit_output"
    final="$explicit_output"
  elif [[ "$do_save" == "1" ]]; then
    final="$(_save_file "$tmpfile")"
  fi

  if [[ "$do_clip" == "1" ]]; then
    _to_clipboard "$final"
  fi

  if [[ "$do_notify" == "1" ]]; then
    if [[ "$do_save" == "1" || -n "$explicit_output" ]]; then
      hg_notify "Screenshot" "Saved: $(basename "$final")"
    else
      hg_notify_low "Screenshot" "Copied to clipboard"
    fi
  fi

  # Clean up temp file if it still exists and wasn't moved
  [[ "$final" != "$tmpfile" ]] && rm -f "$tmpfile" 2>/dev/null || true

  echo "$final"
}

# ── Modes ────────────────────────────────────────
_mode_full() {
  local tmp
  tmp="$(mktemp /tmp/hg-ss-XXXXXX.png)"
  _capture "$tmp" "" "" || hg_die "capture failed"
  _post_process "$tmp" "$DO_SAVE" "$DO_CLIP" "$DO_NOTIFY" "$OUTPUT_PATH"
}

_mode_monitor() {
  local mon="$1"
  local tmp
  tmp="$(mktemp /tmp/hg-ss-XXXXXX.png)"
  _capture "$tmp" "" "$mon" || hg_die "capture failed for monitor $mon"
  _post_process "$tmp" "$DO_SAVE" "$DO_CLIP" "$DO_NOTIFY" "$OUTPUT_PATH"
}

_mode_region() {
  hg_require slurp
  local region
  region="$(slurp 2>/dev/null)" || exit 1
  local tmp
  tmp="$(mktemp /tmp/hg-ss-XXXXXX.png)"
  _capture "$tmp" "$region" "" || hg_die "capture failed"
  _post_process "$tmp" "$DO_SAVE" "$DO_CLIP" "$DO_NOTIFY" "$OUTPUT_PATH"
}

_mode_window() {
  hg_require jq
  local json region address
  json="$(compositor_query activewindow 2>/dev/null)"
  [[ -n "$json" ]] || hg_die "No active window detected"
  address="$(echo "$json" | jq -r '.address // empty')"

  if [[ -n "$address" ]] && _grim_supports_window_capture; then
    local tmp
    tmp="$(mktemp /tmp/hg-ss-XXXXXX.png)"
    grim -w "$address" "$tmp" 2>/dev/null || hg_die "grim window capture failed"
    _post_process "$tmp" "$DO_SAVE" "$DO_CLIP" "$DO_NOTIFY" "$OUTPUT_PATH"
    return 0
  fi

  local ax ay sx sy
  ax="$(echo "$json" | jq -r '.at[0]')"
  ay="$(echo "$json" | jq -r '.at[1]')"
  sx="$(echo "$json" | jq -r '.size[0]')"
  sy="$(echo "$json" | jq -r '.size[1]')"
  region="${ax},${ay} ${sx}x${sy}"
  local tmp
  tmp="$(mktemp /tmp/hg-ss-XXXXXX.png)"
  _capture "$tmp" "$region" "" || hg_die "capture failed"
  _post_process "$tmp" "$DO_SAVE" "$DO_CLIP" "$DO_NOTIFY" "$OUTPUT_PATH"
}

_mode_ocr() {
  hg_require slurp tesseract
  local region
  region="$(slurp 2>/dev/null)" || exit 1
  local tmp
  tmp="$(mktemp /tmp/hg-ss-XXXXXX.png)"
  _capture "$tmp" "$region" "" || hg_die "capture failed"
  local text
  text="$(tesseract "$tmp" - 2>/dev/null)" || hg_die "OCR failed"
  rm -f "$tmp"
  echo -n "$text" | wl-copy 2>/dev/null
  if [[ "$DO_NOTIFY" == "1" ]]; then
    local preview="${text:0:80}"
    hg_notify "OCR" "$preview"
  fi
  echo "$text"
}

_mode_annotate() {
  hg_require slurp satty
  local region
  region="$(slurp 2>/dev/null)" || exit 1
  _capture_stdout "$region" "" | satty --filename - &
  disown
}

_mode_delay() {
  local secs="${1:-3}"
  if [[ "$DO_NOTIFY" == "1" ]]; then
    hg_notify "Screenshot" "Capturing in ${secs}s..."
  fi
  sleep "$secs"
  _mode_full
}

_mode_record() {
  hg_require wf-recorder
  if [[ -f "$REC_PIDFILE" ]]; then
    local pid
    pid="$(cat "$REC_PIDFILE")"
    if kill -0 "$pid" 2>/dev/null; then
      kill -INT "$pid" 2>/dev/null
      rm -f "$REC_PIDFILE"
      if [[ "$DO_NOTIFY" == "1" ]]; then
        hg_notify "Recording" "Stopped"
      fi
      return 0
    fi
    rm -f "$REC_PIDFILE"
  fi

  mkdir -p "$REC_DIR"
  local outfile="$REC_DIR/$(_timestamp).mp4"
  # Auto-select focused monitor to avoid interactive prompt
  local focused_output
  focused_output="$(hyprctl monitors -j 2>/dev/null | jq -r '.[] | select(.focused) | .name' 2>/dev/null)" || true
  local -a rec_args=(-f "$outfile")
  if [[ -n "$focused_output" ]]; then
    rec_args+=(-o "$focused_output")
  fi
  wf-recorder "${rec_args[@]}" &
  echo $! > "$REC_PIDFILE"
  disown
  if [[ "$DO_NOTIFY" == "1" ]]; then
    hg_notify "Recording" "Started: $(basename "$outfile")"
  fi
}

# ── Usage ────────────────────────────────────────
_usage() {
  cat <<'EOF'
hg-screenshot.sh — Unified screenshot capture

USAGE:
  hg-screenshot.sh <mode> [options]

MODES:
  full              Capture all monitors (default)
  monitor <NAME>    Capture specific monitor (e.g., DP-1, DP-2)
  region            Interactive region select via slurp
  window            Active window capture
  ocr               Region select -> OCR -> text to clipboard
  annotate          Region select -> satty annotation
  delay [N]         Delay N seconds (default 3), then full capture
  record            Toggle screen recording (wf-recorder)

OPTIONS:
  --save            Save to ~/Pictures/screenshots/ with timestamp
  --clipboard       Copy image to clipboard via wl-copy (default)
  --both            Save AND clipboard
  --output PATH     Explicit output path
  --notify          Send desktop notification (default)
  --quiet           Suppress notifications
  --stdout          Write PNG to stdout (for piping)
EOF
}

# ── Parse args ───────────────────────────────────
main() {
  local mode="" monitor_name="" delay_secs="3"
  DO_SAVE=0
  DO_CLIP=1
  DO_NOTIFY=1
  OUTPUT_PATH=""
  local do_stdout=0

  while [[ $# -gt 0 ]]; do
    case "$1" in
      full)      mode="full"; shift ;;
      monitor)   mode="monitor"; monitor_name="${2:-}"; shift 2 || hg_die "monitor requires a name" ;;
      region)    mode="region"; shift ;;
      window)    mode="window"; shift ;;
      ocr)       mode="ocr"; shift ;;
      annotate)  mode="annotate"; shift ;;
      delay)     mode="delay"; delay_secs="${2:-3}"; shift; shift 2>/dev/null || true ;;
      record)    mode="record"; shift ;;
      --save)    DO_SAVE=1; shift ;;
      --clipboard) DO_CLIP=1; shift ;;
      --both)    DO_SAVE=1; DO_CLIP=1; shift ;;
      --output)  OUTPUT_PATH="${2:-}"; DO_SAVE=0; shift 2 || hg_die "--output requires a path" ;;
      --notify)  DO_NOTIFY=1; shift ;;
      --quiet)   DO_NOTIFY=0; shift ;;
      --stdout)  do_stdout=1; shift ;;
      -h|--help) _usage; exit 0 ;;
      *)         hg_die "Unknown argument: $1. Run with --help." ;;
    esac
  done

  mode="${mode:-full}"
  hg_require wayshot wl-copy

  # stdout mode: bypass post-processing, pipe directly
  if [[ "$do_stdout" == "1" ]]; then
    case "$mode" in
      full)    _capture_stdout "" "" ;;
      monitor) _capture_stdout "" "$monitor_name" ;;
      region)  hg_require slurp; local r; r="$(slurp 2>/dev/null)" || exit 1; _capture_stdout "$r" "" ;;
      *)       hg_die "--stdout only works with full, monitor, or region modes" ;;
    esac
    return
  fi

  case "$mode" in
    full)     _mode_full ;;
    monitor)  _mode_monitor "$monitor_name" ;;
    region)   _mode_region ;;
    window)   _mode_window ;;
    ocr)      _mode_ocr ;;
    annotate) _mode_annotate ;;
    delay)    _mode_delay "$delay_secs" ;;
    record)   _mode_record ;;
    *)        _usage; exit 1 ;;
  esac
}

main "$@"
