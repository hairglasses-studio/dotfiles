#!/usr/bin/env bash
# shader-cycle — cycle through a curated shader list
# Bound to a global keybind (via Hyprland). Each press advances one step.
# Ghostty auto-reloads config via inotify.
#
# Usage:
#   shader-cycle           # advance to next
#   shader-cycle --prev    # go to previous
#   shader-cycle --show    # print current entry without advancing

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SHADERS_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
source "$SCRIPT_DIR/../../../scripts/lib/notify.sh"
STATE_DIR="$HOME/.local/state/ghostty"
CYCLE_FILE="$STATE_DIR/cycle-index"
BAR_STATE_DIR="$HOME/.local/state/shader-cycle"
BAR_STATE_FILE="$BAR_STATE_DIR/current"
GHOSTTY_CONFIG="$HOME/.config/ghostty/config"

# ── Cycle list ──────────────────────────────────
# Edit this array to change the cycle order.
# Will be repopulated after shader research (Phase 3).
CYCLE=(
  # placeholder — new shaders will be added here
)

CYCLE_LEN=${#CYCLE[@]}

mkdir -p "$STATE_DIR" "$BAR_STATE_DIR" 2>/dev/null

# ── Pre-flight checks ─────────────────────────
[[ -f "$GHOSTTY_CONFIG" ]] || { echo "error: ghostty config not found: $GHOSTTY_CONFIG" >&2; exit 1; }
(( CYCLE_LEN > 0 )) || { echo "error: cycle list is empty" >&2; exit 1; }

# ── Read current index ──────────────────────────
idx=0
if [[ -f "$CYCLE_FILE" ]]; then
  idx="$(< "$CYCLE_FILE")"
  [[ "$idx" =~ ^[0-9]+$ ]] || idx=0
fi

# ── Parse args ──────────────────────────────────
case "${1:-}" in
  --show)
    echo "${CYCLE[$idx]}"
    exit 0
    ;;
  --prev)
    idx=$(( (idx - 1 + CYCLE_LEN) % CYCLE_LEN ))
    ;;
  *)
    idx=$(( (idx + 1) % CYCLE_LEN ))
    ;;
esac

# ── Save new index ──────────────────────────────
printf '%s' "$idx" > "$CYCLE_FILE"

entry="${CYCLE[$idx]}"

# ── Check if shader needs animation ─────────────
needs_animation() {
  grep -qE '(ghostty_time|iTime|u_time)' "$1" 2>/dev/null
}

# ── Apply: Ghostty shader ───────────────────────
apply_ghostty_shader() {
  local shader_path="$SHADERS_DIR/$1"
  if [[ ! -f "$shader_path" ]]; then
    echo "Not found: $shader_path" >&2
    return 1
  fi

  local anim="false"
  needs_animation "$shader_path" && anim="true"

  local tmp
  tmp="$(mktemp "${GHOSTTY_CONFIG}.XXXXXX")"
  local relative_path="shaders/$(basename "$shader_path")"
  awk -v new="custom-shader = ${relative_path}" \
    '/^#* *custom-shader = / && !done { print new; done=1; next } 1' \
    "$GHOSTTY_CONFIG" \
    | sed "s|^custom-shader-animation = .*|custom-shader-animation = $anim|" \
    > "$tmp"
  mv -f "$tmp" "$GHOSTTY_CONFIG"
}

# ── Dispatch ────────────────────────────────────
apply_ghostty_shader "$entry"
printf '%s' "$SHADERS_DIR/$entry" > "$BAR_STATE_FILE"
echo "→ $entry"
hg_notify_low "Shader" "→ ${entry%.glsl}"
