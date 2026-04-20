#!/usr/bin/env bash
# ticker-golden.sh — capture + diff golden-frame baselines per stream.
#
# Wraps scripts/ticker-shot.sh to build a regression-test harness for
# visual changes. `save` captures a deterministic shot per stream into
# ~/.local/state/keybind-ticker/golden/<stream>.png; `diff` re-captures
# and pixel-diffs against the baseline via `magick compare -metric AE`.
# Exits non-zero if any stream delta exceeds --threshold (default 50
# pixel differences — tolerates minor rendering noise from the scroll
# offset changing between save and diff).
#
# Designed for manual "before big change / after big change" smoke,
# not a CI gate — frames inherently depend on wall-clock data (time,
# live streams). See Tier C.5 in the marathon plan for that.
#
# Usage:
#   ticker-golden.sh save [streams...]           save all or named streams
#   ticker-golden.sh diff [streams...]           diff all or named
#   ticker-golden.sh list                        list saved baselines
#   ticker-golden.sh clean                       delete all baselines
#
# Flags (before subcommand args):
#   --dir <path>      override baseline dir (default ~/.local/state/keybind-ticker/golden)
#   --threshold <n>   max allowed AE pixel delta per stream (default 50)
#   --fast            skip slow-threaded streams
#
# Exit codes:
#   0   all diffs within threshold (or save succeeded)
#   1   one or more streams exceeded threshold
#   2   missing baseline for a requested stream in diff mode

set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "$(readlink -f "$0")")" && pwd)"
TICKER_SHOT="$SCRIPT_DIR/ticker-shot.sh"
DEFAULT_DIR="$HOME/.local/state/keybind-ticker/golden"

BASELINE_DIR="$DEFAULT_DIR"
THRESHOLD=50
FAST=0

_die() { printf 'ticker-golden: %s\n' "$*" >&2; exit 1; }

# Parse global flags first; remaining args go to subcommand
ARGS=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --dir)       BASELINE_DIR="${2:-}"; shift 2 ;;
    --threshold) THRESHOLD="${2:-50}";  shift 2 ;;
    --fast)      FAST=1;                shift ;;
    -h|--help)
      sed -n '2,25p' "$0" | sed 's/^# \{0,1\}//'
      exit 0 ;;
    *) ARGS+=("$1"); shift ;;
  esac
done
set -- "${ARGS[@]}"

SUBCMD="${1:-}"
[[ -z "$SUBCMD" ]] && { sed -n '2,25p' "$0" | sed 's/^# \{0,1\}//'; exit 0; }
shift

command -v magick >/dev/null 2>&1 || _die "magick (ImageMagick) not installed"
[[ -x "$TICKER_SHOT" ]] || _die "$TICKER_SHOT not executable"

mkdir -p "$BASELINE_DIR"

_all_streams() {
  # Ask hg for the stream list, strip any `plugin:...` user drop-ins
  # (those aren't included in the main playlist by default and would
  # force an unnecessary playlist switch).
  hg ticker list-streams 2>/dev/null | grep -v '^plugin:' | sort
}

_maybe_skip_slow() {
  # If --fast, filter out streams whose META declares slow=True. Probes
  # each plugin file for `"slow": True`.
  local name="$1"
  [[ "$FAST" -eq 0 ]] && return 1
  local pyfile="$SCRIPT_DIR/lib/ticker_streams/${name//-/_}.py"
  [[ -r "$pyfile" ]] || return 1
  grep -q '"slow": True' "$pyfile" 2>/dev/null
}

_streams_for() {
  if [[ $# -gt 0 ]]; then
    printf '%s\n' "$@"
  else
    _all_streams
  fi
}

case "$SUBCMD" in
  list)
    if [[ -d "$BASELINE_DIR" ]]; then
      cd "$BASELINE_DIR"
      for f in *.png; do
        [[ -f "$f" ]] || continue
        printf '%-25s  %s  %s\n' \
          "${f%.png}" \
          "$(magick identify -format '%wx%h' "$f" 2>/dev/null)" \
          "$(stat -c %y "$f" | cut -d. -f1)"
      done
    fi
    ;;

  clean)
    if [[ -d "$BASELINE_DIR" ]]; then
      rm -f "$BASELINE_DIR"/*.png
      printf 'cleaned baselines in %s\n' "$BASELINE_DIR"
    fi
    ;;

  save)
    failed=0
    while IFS= read -r stream; do
      [[ -z "$stream" ]] && continue
      if _maybe_skip_slow "$stream"; then
        printf 'skip  %s  (slow)\n' "$stream"
        continue
      fi
      out="$BASELINE_DIR/$stream.png"
      printf 'save  %-25s  ' "$stream"
      if "$TICKER_SHOT" --pin "$stream" --output "$out" >/dev/null 2>&1; then
        printf 'ok  (%s bytes)\n' "$(stat -c %s "$out" 2>/dev/null)"
      else
        printf 'FAIL\n'
        failed=1
      fi
    done < <(_streams_for "$@")
    exit $failed
    ;;

  diff)
    exit_rc=0
    tmp="$(mktemp -d /tmp/ticker-golden-XXXXXX)"
    trap 'rm -rf "$tmp"' EXIT
    while IFS= read -r stream; do
      [[ -z "$stream" ]] && continue
      if _maybe_skip_slow "$stream"; then
        printf 'skip  %s  (slow)\n' "$stream"
        continue
      fi
      ref="$BASELINE_DIR/$stream.png"
      if [[ ! -f "$ref" ]]; then
        printf 'miss  %-25s  no baseline (run `save` first)\n' "$stream"
        exit_rc=2
        continue
      fi
      fresh="$tmp/$stream.png"
      if ! "$TICKER_SHOT" --pin "$stream" --output "$fresh" >/dev/null 2>&1; then
        printf 'err   %-25s  shot failed\n' "$stream"
        exit_rc=1
        continue
      fi
      # magick compare -metric AE prints the pixel count to stderr;
      # exit code is 0 for identical, 1 for differences, 2 for error.
      delta="$(magick compare -metric AE "$ref" "$fresh" null: 2>&1 || true)"
      delta_int="${delta%%[^0-9]*}"
      delta_int="${delta_int:-0}"
      if (( delta_int > THRESHOLD )); then
        printf 'DIFF  %-25s  %d px over threshold (%d)\n' \
          "$stream" "$delta_int" "$THRESHOLD"
        exit_rc=1
      else
        printf 'ok    %-25s  %d px within threshold\n' \
          "$stream" "$delta_int"
      fi
    done < <(_streams_for "$@")
    exit $exit_rc
    ;;

  *)
    _die "unknown subcommand: $SUBCMD (try --help)"
    ;;
esac
