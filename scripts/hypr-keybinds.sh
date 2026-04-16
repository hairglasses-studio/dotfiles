#!/usr/bin/env bash
set -euo pipefail
# hypr-keybinds.sh — Generate keybind reference markdown from live Hyprland config
# Parses hyprland.conf for section headers, queries hyprctl binds -j for
# described binds, and generates a Nerd Font-adorned markdown cheatsheet.
#
# Usage: hypr-keybinds.sh [--md] [--glow]
#   --md    Write to hyprland/keybinds.md only (no display)
#   --glow  Generate and display with glow (default)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOTFILES="${DOTFILES_DIR:-$HOME/hairglasses-studio/dotfiles}"
OUT="$DOTFILES/hyprland/keybinds.md"
HYPR_CONF="${XDG_CONFIG_HOME:-$HOME/.config}/hypr/hyprland.conf"

MODE="glow"
[[ "${1:-}" == "--md" ]] && MODE="md"

# ── Config section extraction ─────────────────────
# Walk the config to extract section headers and which bind descriptions
# belong to each section. This preserves the human-authored grouping.

resolve_config() {
  local file="$1"
  while IFS= read -r line; do
    if [[ "$line" =~ ^source[[:space:]]*=[[:space:]]*(.*) ]]; then
      local sourced="${BASH_REMATCH[1]}"
      sourced="${sourced/\~/$HOME}"
      for f in $sourced; do
        [[ -f "$f" ]] && resolve_config "$f"
      done
    else
      printf '%s\n' "$line"
    fi
  done < "$file"
}

# Section icon mapping (Nerd Font)
section_icon() {
  case "$1" in
    *Launch*)       echo "" ;;
    *Navigation*)   echo "" ;;
    *Rotary*|*encoder*) echo "󰸪" ;;
    *Move*window*)  echo "󰆾" ;;
    *Master*|*Layout*|*layout*) echo "" ;;
    *Workspace*)    echo "" ;;
    *Window*manage*) echo "" ;;
    *Window*group*) echo "󰆧" ;;
    *Monitor*focus*) echo "󰍹" ;;
    *Minimize*|*Special*work*) echo "󰘸" ;;
    *Notification*) echo "󰂚" ;;
    *Resize*)       echo "󰩨" ;;
    *Screenshot*)   echo "" ;;
    *Clipboard*|*Color*) echo "" ;;
    *Media*)        echo "󰎆" ;;
    *Night*|*light*) echo "" ;;
    *Shader*|*Kitty*) echo "" ;;
    *Claude*)       echo "󰚩" ;;
    *System*|*Lock*|*wallpaper*|*dashboard*) echo "" ;;
    *Scratchpad*)   echo "󰎟" ;;
    *Mouse*)        echo "󰍽" ;;
    *Special*)      echo "" ;;
    *)              echo "" ;;
  esac
}

# ── Key name translation ─────────────────────────
translate_key() {
  case "$1" in
    grave)        echo "\`" ;;
    bracketright) echo "]" ;;
    bracketleft)  echo "[" ;;
    semicolon)    echo ";" ;;
    backslash)    echo "\\\\" ;;
    minus)        echo "-" ;;
    comma)        echo "," ;;
    period)       echo "." ;;
    Return)       echo "Enter" ;;
    Escape)       echo "Esc" ;;
    space)        echo "Space" ;;
    left)         echo "←" ;;
    right)        echo "→" ;;
    up)           echo "↑" ;;
    down)         echo "↓" ;;
    mouse:272)    echo "LMB" ;;
    mouse:273)    echo "RMB" ;;
    Tab)          echo "Tab" ;;
    *)            echo "$1" ;;
  esac
}

# Modmask bitmask → human-readable
# 64=Super 1=Shift 4=Ctrl 8=Alt
format_mods() {
  local mask="$1"
  local parts=()
  ((mask & 64)) && parts+=("Super")
  ((mask & 1))  && parts+=("Shift")
  ((mask & 4))  && parts+=("Ctrl")
  ((mask & 8))  && parts+=("Alt")
  local IFS="+"
  echo "${parts[*]}"
}

# ── Build section→descriptions map from config ───
declare -a SECTIONS=()
declare -A SECTION_DESCS  # section → newline-separated descriptions

build_section_map() {
  local config_lines section="" submap=""

  config_lines="$(resolve_config "$HYPR_CONF")"

  while IFS= read -r line; do
    # Section headers
    if [[ "$line" =~ ^#[[:space:]]*──[[:space:]]*Keybinds:[[:space:]]*(.*[^─[:space:]])[[:space:]]*─ ]]; then
      section="${BASH_REMATCH[1]}"
      SECTIONS+=("$section")
      SECTION_DESCS["$section"]=""
      continue
    fi

    # Submap tracking
    if [[ "$line" =~ ^submap[[:space:]]*=[[:space:]]*(.*) ]]; then
      local name
      name="$(echo "${BASH_REMATCH[1]}" | xargs)"
      if [[ "$name" == "reset" ]]; then
        submap=""
      else
        submap="$name"
      fi
      continue
    fi

    # bindd lines — extract the description (3rd comma field)
    if [[ "$line" =~ ^bindd[elmnr]*[[:space:]]*=[[:space:]]*(.*) ]]; then
      local fields="${BASH_REMATCH[1]}"
      # Split: mods, key, description, dispatcher[, arg]
      IFS=',' read -r _ _ desc _ <<< "$fields"
      desc="$(echo "$desc" | xargs)"
      if [[ -n "$section" && -n "$desc" ]]; then
        if [[ -n "$submap" ]]; then
          SECTION_DESCS["$section"]+="[${submap}] ${desc}"$'\n'
        else
          SECTION_DESCS["$section"]+="${desc}"$'\n'
        fi
      fi
    fi
  done <<< "$config_lines"
}

# ── Query runtime binds ──────────────────────────
declare -A DESC_TO_BIND  # "description" → "Super+K"

query_runtime_binds() {
  local json
  json="$(hyprctl binds -j 2>/dev/null)" || return 1

  local count
  count="$(jq length <<< "$json")"

  for ((i=0; i<count; i++)); do
    local has_desc key modmask mouse submap desc
    has_desc="$(jq -r ".[$i].has_description" <<< "$json")"
    [[ "$has_desc" != "true" ]] && continue

    key="$(jq -r ".[$i].key" <<< "$json")"
    modmask="$(jq -r ".[$i].modmask" <<< "$json")"
    mouse="$(jq -r ".[$i].mouse" <<< "$json")"
    submap="$(jq -r ".[$i].submap" <<< "$json")"
    desc="$(jq -r ".[$i].description" <<< "$json")"

    local translated_key mods keybind_str
    translated_key="$(translate_key "$key")"
    mods="$(format_mods "$modmask")"

    if [[ -n "$mods" ]]; then
      keybind_str="${mods}+${translated_key}"
    else
      keybind_str="${translated_key}"
    fi

    # Prefix with submap if inside one
    local lookup_key="$desc"
    if [[ -n "$submap" ]]; then
      lookup_key="[${submap}] ${desc}"
    fi

    DESC_TO_BIND["$lookup_key"]="$keybind_str"
  done
}

# ── Generate markdown ─────────────────────────────
generate_md() {
  echo "# 󰌌 Keybinds"
  echo ""

  for section in "${SECTIONS[@]}"; do
    local icon
    icon="$(section_icon "$section")"
    local descs="${SECTION_DESCS[$section]:-}"
    [[ -z "$descs" ]] && continue

    echo "## ${icon} ${section}"
    echo ""
    echo "| Key | Action |"
    echo "|-----|--------|"

    while IFS= read -r desc; do
      [[ -z "$desc" ]] && continue

      local keybind="${DESC_TO_BIND[$desc]:-?}"

      # Strip submap prefix for display, show as badge
      local display_desc="$desc"
      local badge=""
      if [[ "$desc" =~ ^\[([^\]]+)\][[:space:]]*(.*) ]]; then
        badge="${BASH_REMATCH[1]}"
        display_desc="${BASH_REMATCH[2]}"
      fi

      # Format keybind with backtick code spans
      local formatted_key="\`${keybind}\`"
      if [[ -n "$badge" ]]; then
        display_desc="*${badge}:* ${display_desc}"
      fi

      echo "| ${formatted_key} | ${display_desc} |"
    done <<< "$descs"

    echo ""
  done
}

# ── Main ──────────────────────────────────────────
build_section_map
query_runtime_binds
generate_md > "$OUT"

if [[ "$MODE" == "glow" ]]; then
  exec glow -p "$OUT"
fi
