#!/usr/bin/env bash
set -euo pipefail
# font-mix.sh — Monaspace multi-font profile manager for Ghostty
# Ghostty supports different font families per text style (regular/bold/italic/bold-italic).
# Monaspace's 5 variants are metrics-compatible, enabling seamless mixing.
#
# Usage:
#   font-mix cyberpunk       # Apply cyberpunk preset
#   font-mix list            # List available presets
#   font-mix show            # Show current font mapping
#   font-mix next            # Cycle to next preset
#   font-mix set Ne Xe Ar Kr # Custom: shortcodes for each style

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/config.sh"

GHOSTTY_CONFIG="$HOME/.config/ghostty/config"
STATE_DIR="$HOME/.local/state/font-mix"
STATE_FILE="$STATE_DIR/current"

# ── Monaspace variant map ──────────────────────────
declare -A VARIANTS=(
  [Ne]="MonaspiceNe Nerd Font Mono"
  [Ar]="MonaspiceAr Nerd Font Mono"
  [Rn]="MonaspiceRn Nerd Font Mono"
  [Kr]="MonaspiceKr Nerd Font Mono"
  [Xe]="MonaspiceXe Nerd Font Mono"
)

declare -A VARIANT_NAMES=(
  [Ne]="Neon (neo-grotesque)"
  [Ar]="Argon (humanist)"
  [Rn]="Radon (handwriting)"
  [Kr]="Krypton (mechanical)"
  [Xe]="Xenon (slab serif)"
)

# ── Presets: regular bold italic bold-italic ───────
declare -A PRESETS=(
  [cyberpunk]="Ne Xe Ar Kr"
  [hacker]="Kr Xe Rn Ne"
  [writer]="Ar Xe Rn Kr"
  [mono]="Ne Ne Ne Ne"
  [radon]="Rn Xe Ar Kr"
  [xenon]="Xe Kr Rn Ne"
)

PRESET_ORDER=(cyberpunk hacker writer radon xenon mono)

# ── Helpers ────────────────────────────────────────

resolve_font() {
  local code="$1"
  if [[ -n "${VARIANTS[$code]+x}" ]]; then
    echo "${VARIANTS[$code]}"
  else
    echo "Unknown variant: $code (valid: Ne Ar Rn Kr Xe)" >&2
    return 1
  fi
}

apply_fonts() {
  local reg="$1" bold="$2" italic="$3" bolditalic="$4"

  local reg_font bold_font italic_font bi_font
  reg_font="$(resolve_font "$reg")" || return 1
  bold_font="$(resolve_font "$bold")" || return 1
  italic_font="$(resolve_font "$italic")" || return 1
  bi_font="$(resolve_font "$bolditalic")" || return 1

  # Atomic config update
  local tmp
  tmp="$(mktemp "${GHOSTTY_CONFIG}.XXXXXX")"
  sed \
    -e "s|^font-family = .*|font-family = $reg_font|" \
    -e "s|^font-family-bold = .*|font-family-bold = $bold_font|" \
    -e "s|^font-family-italic = .*|font-family-italic = $italic_font|" \
    -e "s|^font-family-bold-italic = .*|font-family-bold-italic = $bi_font|" \
    "$GHOSTTY_CONFIG" > "$tmp"
  mv -f "$tmp" "$GHOSTTY_CONFIG"

  # Save state
  mkdir -p "$STATE_DIR"
  printf '%s %s %s %s' "$reg" "$bold" "$italic" "$bolditalic" > "$STATE_FILE"
}

show_current() {
  local reg bold italic bi
  reg="$(grep -m1 '^font-family = ' "$GHOSTTY_CONFIG" | sed 's/^font-family = //')"
  bold="$(grep -m1 '^font-family-bold = ' "$GHOSTTY_CONFIG" | sed 's/^font-family-bold = //')"
  italic="$(grep -m1 '^font-family-italic = ' "$GHOSTTY_CONFIG" | sed 's/^font-family-italic = //')"
  bi="$(grep -m1 '^font-family-bold-italic = ' "$GHOSTTY_CONFIG" | sed 's/^font-family-bold-italic = //')"

  # Detect preset name
  local preset_name="custom"
  if [[ -f "$STATE_FILE" ]]; then
    local codes
    codes="$(cat "$STATE_FILE")"
    for name in "${PRESET_ORDER[@]}"; do
      if [[ "${PRESETS[$name]}" == "$codes" ]]; then
        preset_name="$name"
        break
      fi
    done
  fi

  printf "Profile: %s\n" "$preset_name"
  printf "  Regular:     %s\n" "$reg"
  printf "  Bold:        %s\n" "$bold"
  printf "  Italic:      %s\n" "$italic"
  printf "  Bold-Italic: %s\n" "$bi"
}

list_presets() {
  printf "Available font-mix presets:\n\n"
  for name in "${PRESET_ORDER[@]}"; do
    local codes="${PRESETS[$name]}"
    read -r r b i bi <<< "$codes"
    printf "  %-12s  %s / %s / %s / %s\n" "$name" \
      "${VARIANT_NAMES[$r]}" "${VARIANT_NAMES[$b]}" \
      "${VARIANT_NAMES[$i]}" "${VARIANT_NAMES[$bi]}"
  done
  printf "\nVariant codes: Ne=Neon Ar=Argon Rn=Radon Kr=Krypton Xe=Xenon\n"
}

next_preset() {
  local current_codes=""
  [[ -f "$STATE_FILE" ]] && current_codes="$(cat "$STATE_FILE")"

  local current_idx=0
  for i in "${!PRESET_ORDER[@]}"; do
    if [[ "${PRESETS[${PRESET_ORDER[$i]}]}" == "$current_codes" ]]; then
      current_idx=$i
      break
    fi
  done

  local next_idx=$(( (current_idx + 1) % ${#PRESET_ORDER[@]} ))
  local next_name="${PRESET_ORDER[$next_idx]}"
  local codes="${PRESETS[$next_name]}"
  read -r r b i bi <<< "$codes"
  apply_fonts "$r" "$b" "$i" "$bi"
  hg_notify_low "Font Mix" "→ $next_name"
  printf "Applied: %s\n" "$next_name"
}

# ── Main ───────────────────────────────────────────

cmd="${1:-show}"
shift || true

case "$cmd" in
  list)
    list_presets
    ;;
  show)
    show_current
    ;;
  next)
    next_preset
    ;;
  set)
    if [[ $# -ne 4 ]]; then
      echo "Usage: font-mix set <regular> <bold> <italic> <bold-italic>" >&2
      echo "  Codes: Ne Ar Rn Kr Xe" >&2
      exit 1
    fi
    apply_fonts "$1" "$2" "$3" "$4"
    printf '%s %s %s %s' "$1" "$2" "$3" "$4" > "$STATE_FILE"
    hg_notify_low "Font Mix" "→ custom ($1/$2/$3/$4)"
    printf "Applied custom: %s / %s / %s / %s\n" "$1" "$2" "$3" "$4"
    ;;
  *)
    # Treat as preset name
    if [[ -n "${PRESETS[$cmd]+x}" ]]; then
      codes="${PRESETS[$cmd]}"
      read -r r b i bi <<< "$codes"
      apply_fonts "$r" "$b" "$i" "$bi"
      printf '%s' "$r $b $i $bi" > "$STATE_FILE"
      hg_notify_low "Font Mix" "→ $cmd"
      printf "Applied: %s\n" "$cmd"
    else
      echo "Unknown command or preset: $cmd" >&2
      echo "Usage: font-mix [cyberpunk|hacker|writer|radon|xenon|mono|list|show|next|set ...]" >&2
      exit 1
    fi
    ;;
esac
