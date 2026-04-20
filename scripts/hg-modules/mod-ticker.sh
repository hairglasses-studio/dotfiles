#!/usr/bin/env bash
# mod-ticker.sh — hg ticker module
# Pin/unpin streams, switch playlists, pause/resume, and inspect the ticker.

_TICKER_STATE="$HOME/.local/state/keybind-ticker"
_TICKER_SERVICE="dotfiles-keybind-ticker.service"
_TICKER_HEADLESS="$HG_DOTFILES/scripts/ticker-headless.py"
_TICKER_SHOT="$HG_DOTFILES/scripts/ticker-shot.sh"
_TICKER_RECORD="$HG_DOTFILES/scripts/ticker-record.sh"
_TICKER_SMOKE="$HG_DOTFILES/scripts/ticker-smoke-test.py"
_TICKER_GOLDEN="$HG_DOTFILES/scripts/ticker-golden.sh"

# Signal the running keybind-ticker to re-read state files (SIGUSR1). Falls
# back to a full service restart if no PID is found (e.g. during boot before
# the ticker has started). Restart happens in the user session manager so the
# service's Environment= lines are re-applied; a raw kill+exec would miss
# LD_PRELOAD for gtk4-layer-shell.
_ticker_reload() {
  local pids rc
  pids="$(pgrep -f 'keybind-ticker.py --layer' 2>/dev/null || true)"
  if [[ -n "$pids" ]]; then
    # shellcheck disable=SC2086
    if kill -USR1 $pids 2>/dev/null; then
      return 0
    fi
  fi
  systemctl --user restart "$_TICKER_SERVICE" >/dev/null 2>&1 || true
  return 0
}

ticker_description() {
  echo "Keybind ticker control (pin, playlist, pause, status)"
}

ticker_commands() {
  cat <<'CMDS'
status	Print current stream, playlist, and pause state
health	Per-stream success / error counters and last-ok age
pin	Pin a specific stream (pass stream name)
unpin	Release pin, resume rotation
playlist	Switch playlist (pass name: main/coding/focus/lock/recording)
pause	Toggle pause
shuffle	Shuffle-mode toggle for current playlist (on|off|toggle)
list-streams	List all available streams
list-playlists	List playlist files
show	Render a single stream plain-text once (pass stream name)
restart	Restart the ticker systemd service
logs	Tail the ticker systemd logs
shot	Capture a 28px-tall ticker-only PNG (safe for Claude ingestion)
record	Record a ticker-only MP4 via wf-recorder (default 60s)
smoke-test	Load every plugin + TOML stream and call build() — PASS/FAIL report
recover-monitor	Restore DP-2 to native mode after DSC fallback (ticker clipping)
next	Advance one stream (via DBus)
prev	Rewind one stream (via DBus)
pin-toggle	Pin current stream if unpinned, else unpin
reload	Hot-reload plugin modules + TOML catalogue (via DBus)
banner	Flash a toast banner via DBus — `hg ticker banner <text> [color]`
snooze-urgent	Dismiss an active urgent-mode escalation early
golden	Capture / diff per-stream reference PNGs (save | diff | list | clean)
input-test	Assert the DBus control-surface round-trips (pin, shuffle, next, urgent)
CMDS
}

_ticker_write_state() {
  local key="$1" value="$2"
  mkdir -p "$_TICKER_STATE"
  if [[ -z "$value" ]]; then
    rm -f "$_TICKER_STATE/$key"
  else
    printf '%s' "$value" > "$_TICKER_STATE/$key"
  fi
}

_ticker_read_state() {
  local key="$1" default="${2:-}"
  if [[ -f "$_TICKER_STATE/$key" ]]; then
    cat "$_TICKER_STATE/$key"
  else
    printf '%s' "$default"
  fi
}

ticker_run() {
  local cmd="${1:-status}"
  shift || true
  case "$cmd" in
    status)
      printf 'service   : %s\n' "$(systemctl --user is-active "$_TICKER_SERVICE")"
      printf 'playlist  : %s\n' "$(_ticker_read_state active-playlist main)"
      printf 'current   : %s\n' "$(_ticker_read_state current-stream '(rotating)')"
      printf 'pinned    : %s\n' "$(_ticker_read_state pinned-stream '(none)')"
      if [[ -f "$_TICKER_STATE/paused" ]]; then
        printf 'paused    : yes\n'
      else
        printf 'paused    : no\n'
      fi
      if [[ -f "$_TICKER_STATE/shuffle" ]]; then
        printf 'shuffle   : on\n'
      else
        printf 'shuffle   : off\n'
      fi
      ;;
    pin)
      local stream="${1:?usage: hg ticker pin <stream>}"
      _ticker_write_state pinned-stream "$stream"
      _ticker_reload
      printf 'pinned to %s\n' "$stream"
      ;;
    unpin)
      _ticker_write_state pinned-stream ""
      _ticker_reload
      printf 'unpinned\n'
      ;;
    playlist)
      local name="${1:?usage: hg ticker playlist <name>}"
      _ticker_write_state active-playlist "$name"
      _ticker_reload
      printf 'playlist switched to %s\n' "$name"
      ;;
    pause)
      if [[ -f "$_TICKER_STATE/paused" ]]; then
        rm -f "$_TICKER_STATE/paused"
        printf 'resumed\n'
      else
        mkdir -p "$_TICKER_STATE"
        : > "$_TICKER_STATE/paused"
        printf 'paused\n'
      fi
      _ticker_reload
      ;;
    shuffle)
      # Usage: hg ticker shuffle [on|off|toggle]   (default: toggle)
      local arg="${1:-toggle}"
      local flag="$_TICKER_STATE/shuffle"
      mkdir -p "$_TICKER_STATE"
      case "$arg" in
        on)      : > "$flag"; printf 'shuffle on\n' ;;
        off)     rm -f "$flag"; printf 'shuffle off\n' ;;
        toggle|"")
          if [[ -f "$flag" ]]; then
            rm -f "$flag"; printf 'shuffle off\n'
          else
            : > "$flag"; printf 'shuffle on\n'
          fi
          ;;
        *) printf 'usage: hg ticker shuffle [on|off|toggle]\n' >&2; return 2 ;;
      esac
      _ticker_reload
      ;;
    list-streams)
      python3 "$_TICKER_HEADLESS" --list
      ;;
    list-playlists)
      for f in "$HG_DOTFILES/ticker/content-playlists"/*.txt; do
        printf '%s\n' "$(basename "$f" .txt)"
      done
      ;;
    show)
      local stream="${1:?usage: hg ticker show <stream>}"
      python3 "$_TICKER_HEADLESS" --stream "$stream"
      printf '\n'
      ;;
    restart)
      systemctl --user restart "$_TICKER_SERVICE"
      systemctl --user is-active "$_TICKER_SERVICE"
      ;;
    logs)
      journalctl --user -u "$_TICKER_SERVICE" --since "2 min ago" --no-pager
      ;;
    shot)
      # Usage:
      #   hg ticker shot                 # current content → /tmp/ticker-shot.png
      #   hg ticker shot <stream>        # pin, wait past wipe, shoot, unpin
      #   hg ticker shot <stream> <path> # …and write to an explicit path
      # Any extra flags (e.g. --monitor DP-3 --scale 0.5) pass straight through.
      local -a args=()
      if [[ $# -ge 1 && "$1" != --* ]]; then
        args+=(--pin "$1"); shift
      fi
      if [[ $# -ge 1 && "$1" != --* ]]; then
        args+=(--output "$1"); shift
      fi
      "$_TICKER_SHOT" "${args[@]}" "$@"
      ;;
    smoke-test)
      python3 "$_TICKER_SMOKE" "$@"
      ;;
    record)
      # Usage:
      #   hg ticker record                  60s → ~/Videos/recordings/ticker-<ts>.mp4
      #   hg ticker record 30               30s recording
      #   hg ticker record 90 calendar      90s pinned on `calendar`
      #   hg ticker record <stream>         60s pinned on <stream>
      # Any extra flags (--audio, --codec, --monitor, --output, --no-switch)
      # pass straight through.
      local -a args=()
      if [[ $# -ge 1 && "$1" =~ ^[0-9]+$ ]]; then
        args+=("$1"); shift
      fi
      if [[ $# -ge 1 && "$1" != --* ]]; then
        args+=(--pin "$1"); shift
      fi
      "$_TICKER_RECORD" "${args[@]}" "$@"
      ;;
    recover-monitor)
      # Diagnose + restore the ultrawide from DSC fallback to native mode.
      # Passes positional args through: [monitor] [native-mode] [pos] [scale]
      "$HG_DOTFILES/scripts/hg-dsc-recover.sh" "$@"
      ;;
    health)
      # Usage: hg ticker health [--watch]
      if [[ "${1:-}" == "--watch" ]]; then
        exec watch -n 5 -t hg ticker health
      fi
      local snap="/tmp/ticker-health.json"
      if [[ ! -f "$snap" ]]; then
        printf 'no snapshot yet; ticker writes one every 30s\n' >&2
        return 1
      fi
      python3 - "$snap" <<'PY'
import json, os, sys, time
path = sys.argv[1]
with open(path) as f:
    data = json.load(f)
now = int(time.time())
age_snap = now - data.get("ts", now)
perf = data.get("perf", {})
print(f"playlist : {data.get('playlist','?')}   pid={data.get('pid','?')}   snapshot age={age_snap}s")
# New perf block — adaptive-quality tier + runtime flags
cur       = perf.get("current_stream") or "(unknown)"
pinned    = perf.get("pinned") or "(none)"
dwell_rem = perf.get("dwell_remaining")
dwell_s   = f"{dwell_rem}s" if dwell_rem is not None else "-"
tier      = perf.get("tier", 0)
ema       = perf.get("ema_frame_ms", 0.0)
bg        = perf.get("bg_inflight", 0)
urgent    = "YES" if perf.get("urgent_mode") else "no"
paused    = "YES" if perf.get("paused") else "no"
shuffle   = "YES" if perf.get("shuffle") else "no"
print(f"current  : {cur}   pinned={pinned}   dwell={dwell_s}")
print(f"perf     : tier={tier}   ema={ema:.2f}ms   bg_inflight={bg}   "
      f"urgent={urgent}   paused={paused}   shuffle={shuffle}")
print()
print(f"{'stream':<20} {'last ok':>9} {'last err':>9} {'ok':>6} {'err':>6} {'streak':>6}")
print("-" * 60)
for name in sorted(data.get("streams", {}).keys()):
    s = data["streams"][name]
    last_ok = (now - s["last_ok"]) if s["last_ok"] else -1
    last_err = (now - s["last_err"]) if s["last_err"] else -1
    ok_s = f"{last_ok}s" if last_ok >= 0 else "-"
    err_s = f"{last_err}s" if last_err >= 0 else "-"
    print(f"{name:<20} {ok_s:>9} {err_s:>9} {s['total_ok']:>6} {s['total_err']:>6} {s['consecutive_err']:>6}")
PY
      ;;
    next|prev|pin-toggle|reload|snooze-urgent)
      # DBus-driven verbs mapped to io.hairglasses.Ticker methods.
      # Requires the primary ticker instance to be running.
      local method
      case "$cmd" in
        next)           method="NextStream" ;;
        prev)           method="PrevStream" ;;
        pin-toggle)     method="PinToggle"  ;;
        reload)         method="ReloadPlugins" ;;
        snooze-urgent)  method="SnoozeUrgent"  ;;
      esac
      gdbus call --session \
        -d io.hairglasses.keybind_ticker \
        -o /io/hairglasses/Ticker \
        -m "io.hairglasses.Ticker.$method" >/dev/null
      ;;
    banner)
      # Usage: hg ticker banner <text> [color]
      local text="${1:?usage: hg ticker banner <text> [color]}"
      local color="${2:-#29f0ff}"
      gdbus call --session \
        -d io.hairglasses.keybind_ticker \
        -o /io/hairglasses/Ticker \
        -m io.hairglasses.Ticker.ShowBanner \
        "$text" "$color" >/dev/null
      ;;
    golden)
      "$_TICKER_GOLDEN" "$@"
      ;;
    input-test)
      "$HG_DOTFILES/scripts/ticker-input-test.sh" "$@"
      ;;
    *)
      printf 'unknown command: %s\n' "$cmd" >&2
      printf 'try: hg ticker --help\n' >&2
      return 2
      ;;
  esac
}
