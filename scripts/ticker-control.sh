#!/usr/bin/env bash
# ticker-control.sh - Quickshell ticker IPC with legacy DBus/state fallback.

set -euo pipefail

script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
dotfiles_dir="$(cd "$script_dir/.." && pwd)"
config_path="${QUICKSHELL_CONFIG_PATH:-$dotfiles_dir/quickshell/shell.qml}"
quickshell_bin="${QUICKSHELL_BIN:-quickshell}"
ipc_timeout="${QUICKSHELL_IPC_TIMEOUT:-0.8}"
state_dir="${XDG_STATE_HOME:-$HOME/.local/state}/keybind-ticker"
legacy_bus="io.hairglasses.keybind_ticker"
legacy_obj="/io/hairglasses/Ticker"
legacy_iface="io.hairglasses.Ticker"

usage() {
  cat <<'EOF'
Usage: ticker-control.sh <status|next|prev|pin|unpin|pin-toggle|pause|shuffle|playlist|preset|reload|banner|urgent|snooze-urgent|list-streams|list-playlists|show>

Quickshell IPC is preferred. The legacy keybind-ticker DBus/state files are
used only when Quickshell is unavailable.
EOF
}

qs_call() {
  command -v "$quickshell_bin" >/dev/null 2>&1 || return 1
  timeout "$ipc_timeout" "$quickshell_bin" ipc --path "$config_path" --newest call ticker "$@" >/dev/null 2>&1
}

qs_call_output() {
  command -v "$quickshell_bin" >/dev/null 2>&1 || return 1
  timeout "$ipc_timeout" "$quickshell_bin" ipc --path "$config_path" --newest call ticker "$@"
}

legacy_call() {
  local method="$1"; shift || true
  gdbus call --session -d "$legacy_bus" -o "$legacy_obj" -m "$legacy_iface.$method" "$@" >/dev/null
}

write_state() {
  local key="$1" value="${2:-}"
  mkdir -p "$state_dir"
  if [[ -z "$value" ]]; then
    rm -f "$state_dir/$key"
  else
    printf '%s' "$value" > "$state_dir/$key"
  fi
}

touch_flag() {
  mkdir -p "$state_dir"
  : > "$state_dir/$1"
}

legacy_reload() {
  local pids
  pids="$(pgrep -f 'keybind-ticker.py --layer' 2>/dev/null || true)"
  if [[ -n "$pids" ]]; then
    # shellcheck disable=SC2086
    kill -USR1 $pids 2>/dev/null || true
  else
    systemctl --user restart dotfiles-keybind-ticker.service >/dev/null 2>&1 || true
  fi
}

state_value() {
  local key="$1" default="${2:-}"
  if [[ -f "$state_dir/$key" ]]; then
    cat "$state_dir/$key"
  else
    printf '%s' "$default"
  fi
}

legacy_status_json() {
  local playlist current pinned paused=false shuffle=false urgent=false
  playlist="$(state_value active-playlist main)"
  current="$(state_value current-stream "")"
  pinned="$(state_value pinned-stream "")"
  [[ -f "$state_dir/paused" ]] && paused=true
  [[ -f "$state_dir/shuffle" ]] && shuffle=true
  [[ -f "$state_dir/urgent-until" ]] && urgent=true
  jq -cn \
    --arg service "legacy" \
    --arg playlist "$playlist" \
    --arg current "$current" \
    --arg pinned "$pinned" \
    --argjson paused "$paused" \
    --argjson shuffle "$shuffle" \
    --argjson urgent "$urgent" \
    '{service:$service,playlist:$playlist,current_stream:$current,pinned:$pinned,paused:$paused,shuffle:$shuffle,urgent:$urgent}'
}

print_human_status() {
  local json="$1"
  jq -r '
    "service   : \(.service // "quickshell")",
    "playlist  : \(.playlist // "main")",
    "current   : \(.current_stream // .stream // "")",
    "pinned    : \((.pinned // "") | if . == "" then "(none)" else . end)",
    "paused    : \(if .paused then "yes" else "no" end)",
    "shuffle   : \(if .shuffle then "on" else "off" end)",
    "urgent    : \(if .urgent then "yes" else "no" end)"
  ' <<<"$json"
}

cmd="${1:-status}"
shift || true

case "$cmd" in
  status)
    if [[ "${1:-}" == "--json" ]]; then
      if json="$(qs_call_output status 2>/dev/null)"; then
        printf '%s\n' "$json"
      else
        legacy_status_json
      fi
    else
      json="$(qs_call_output status 2>/dev/null || legacy_status_json)"
      print_human_status "$json"
    fi
    ;;
  next)
    qs_call next || legacy_call NextStream
    ;;
  prev)
    qs_call prev || legacy_call PrevStream
    ;;
  pin)
    stream="${1:?usage: ticker-control.sh pin <stream>}"
    qs_call pin "$stream" || { write_state pinned-stream "$stream"; write_state current-stream "$stream"; legacy_call Pin "$stream" 2>/dev/null || legacy_reload; }
    printf 'pinned to %s\n' "$stream"
    ;;
  unpin)
    qs_call unpin || { write_state pinned-stream ""; legacy_call Unpin 2>/dev/null || legacy_reload; }
    printf 'unpinned\n'
    ;;
  pin-toggle)
    qs_call pinToggle || legacy_call PinToggle
    ;;
  pause)
    if qs_call pauseToggle; then
      :
    elif [[ -f "$state_dir/paused" ]]; then
      rm -f "$state_dir/paused"; legacy_reload
    else
      touch_flag paused; legacy_reload
    fi
    ;;
  shuffle)
    mode="${1:-toggle}"
    if qs_call shuffle "$mode"; then
      :
    else
      case "$mode" in
        on) touch_flag shuffle ;;
        off) rm -f "$state_dir/shuffle" ;;
        toggle|"") [[ -f "$state_dir/shuffle" ]] && rm -f "$state_dir/shuffle" || touch_flag shuffle ;;
        *) printf 'usage: ticker-control.sh shuffle [on|off|toggle]\n' >&2; exit 2 ;;
      esac
      legacy_call Shuffle "$mode" 2>/dev/null || legacy_reload
    fi
    ;;
  playlist)
    name="${1:?usage: ticker-control.sh playlist <name>}"
    qs_call playlist "$name" || { write_state active-playlist "$name"; legacy_call SetPlaylist "$name" 2>/dev/null || legacy_reload; }
    printf 'playlist switched to %s\n' "$name"
    ;;
  preset)
    name="${1:?usage: ticker-control.sh preset <name>}"
    qs_call preset "$name" || { write_state preset "$name"; legacy_call SetPreset "$name" 2>/dev/null || legacy_reload; }
    printf 'preset switched to %s\n' "$name"
    ;;
  reload)
    qs_call reload || legacy_call ReloadPlugins 2>/dev/null || legacy_reload
    ;;
  banner)
    text="${1:?usage: ticker-control.sh banner <text> [color]}"
    color="${2:-#29f0ff}"
    qs_call banner "$text" "$color" || legacy_call ShowBanner "$text" "$color"
    ;;
  urgent)
    enabled="${1:-true}"
    qs_call urgent "$enabled" || legacy_call SetUrgent "$enabled"
    ;;
  snooze-urgent)
    qs_call snoozeUrgent || legacy_call SnoozeUrgent
    ;;
  list-streams)
    if streams="$(qs_call_output listStreams 2>/dev/null)"; then
      printf '%s\n' "$streams"
    else
      python3 "$script_dir/ticker-headless.py" --list
    fi
    ;;
  list-playlists)
    for f in "$dotfiles_dir/ticker/content-playlists"/*.txt; do
      printf '%s\n' "$(basename "$f" .txt)"
    done
    ;;
  show)
    stream="${1:?usage: ticker-control.sh show <stream>}"
    python3 "$script_dir/ticker-headless.py" --stream "$stream"
    printf '\n'
    ;;
  -h|--help|help)
    usage
    ;;
  *)
    usage >&2
    exit 2
    ;;
esac
