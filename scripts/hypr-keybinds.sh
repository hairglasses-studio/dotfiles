#!/usr/bin/env bash
# hypr-keybinds.sh — Live keybinding reference from Hyprland config
# Usage: hypr-keybinds.sh [search_term] [--all]
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

# ── Args ──────────────────────────────────────────
SEARCH=""
SHOW_ALL=false
while [[ $# -gt 0 ]]; do
  case "$1" in
    --all|-a) SHOW_ALL=true; shift ;;
    --help|-h)
      echo "Usage: hypr-keybinds.sh [search_term] [--all]"
      echo "  search_term  Jump to first match in output"
      echo "  --all        Include media keys and mouse binds"
      exit 0 ;;
    *) SEARCH="$1"; shift ;;
  esac
done

# ── Config resolution ─────────────────────────────
HYPR_CONF="${XDG_CONFIG_HOME:-$HOME/.config}/hypr/hyprland.conf"
[[ -f "$HYPR_CONF" ]] || hg_die "Hyprland config not found: $HYPR_CONF"

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

# ── Variable extraction ──────────────────────────
declare -A VARS
extract_vars() {
  while IFS= read -r line; do
    if [[ "$line" =~ ^\$([a-zA-Z_][a-zA-Z0-9_]*)[[:space:]]*=[[:space:]]*(.*) ]]; then
      VARS["${BASH_REMATCH[1]}"]="${BASH_REMATCH[2]}"
    fi
  done
}

expand_vars() {
  local s="$1"
  for var in "${!VARS[@]}"; do
    s="${s//\$$var/${VARS[$var]}}"
  done
  echo "$s"
}

# ── Key name translation ─────────────────────────
translate_key() {
  case "$1" in
    grave)        echo "\`" ;;
    bracketright) echo "]" ;;
    bracketleft)  echo "[" ;;
    semicolon)    echo ";" ;;
    backslash)    echo "\\" ;;
    minus)        echo "-" ;;
    comma)        echo "," ;;
    period)       echo "." ;;
    Return)       echo "Enter" ;;
    Escape)       echo "Esc" ;;
    space)        echo "Space" ;;
    left)         echo "Left" ;;
    right)        echo "Right" ;;
    up)           echo "Up" ;;
    down)         echo "Down" ;;
    mouse:272)    echo "LMB" ;;
    mouse:273)    echo "RMB" ;;
    *)            echo "$1" ;;
  esac
}

# ── Dispatcher-to-action mapping ─────────────────
dispatcher_action() {
  local disp="$1" arg="$2"
  case "$disp" in
    exec)
      exec_action "$arg" ;;
    movefocus)
      local dir; dir=$(dir_name "$arg")
      echo "Focus $dir" ;;
    movewindow)
      local dir; dir=$(dir_name "$arg")
      echo "Move window $dir" ;;
    split-workspace)
      echo "Workspace $arg" ;;
    split-movetoworkspace)
      echo "Move to workspace $arg" ;;
    split-changemonitor)
      echo "Move window to other monitor" ;;
    killactive)
      echo "Close window" ;;
    fullscreen)
      echo "Fullscreen" ;;
    togglefloating)
      echo "Toggle float" ;;
    cyclenext)
      echo "Cycle focus" ;;
    focusurgentorlast)
      echo "Focus urgent/last" ;;
    layoutmsg)
      layout_action "$arg" ;;
    resizeactive)
      echo "Resize $arg" ;;
    submap)
      if [[ "$arg" == "reset" ]]; then
        echo "Exit mode"
      else
        echo "Enter $arg mode"
      fi ;;
    hyprexpo:expo)
      echo "Workspace overview" ;;
    togglespecialworkspace)
      echo "Toggle $arg workspace" ;;
    workspace)
      echo "Workspace $arg" ;;
    movetoworkspace)
      echo "Move to workspace $arg" ;;
    exit)
      echo "Exit Hyprland" ;;
    pseudo)
      echo "Toggle pseudotile" ;;
    *)
      echo "$disp ${arg:+$arg}" ;;
  esac
}

dir_name() {
  case "$1" in
    l) echo "left" ;; r) echo "right" ;; u) echo "up" ;; d) echo "down" ;;
    +1) echo "next" ;; -1) echo "prev" ;;
    *) echo "$1" ;;
  esac
}

layout_action() {
  local arg="$1"
  case "$arg" in
    orientationcycle*)  echo "Cycle orientation" ;;
    swapwithmaster*)    echo "Swap with master" ;;
    addmaster*)         echo "Add master" ;;
    removemaster*)      echo "Remove master" ;;
    focusmaster*)       echo "Focus master" ;;
    rollnext*)          echo "Roll stack forward" ;;
    rollprev*)          echo "Roll stack backward" ;;
    swapnext*)          echo "Swap with next" ;;
    swapprev*)          echo "Swap with prev" ;;
    mfact*)
      if [[ "$arg" == *"+"* ]]; then
        echo "Increase master ratio"
      else
        echo "Decrease master ratio"
      fi ;;
    *) echo "Layout: $arg" ;;
  esac
}

exec_action() {
  local arg="$1"
  case "$arg" in
    *"ghostty --gtk-single-instance=false"*)
      echo "New terminal" ;;
    *"wofi --show drun"*|*'$menu'*)
      echo "App launcher" ;;
    "pypr toggle term")
      echo "Scratchpad: terminal" ;;
    "pypr toggle keybinds")
      echo "Keybind reference" ;;
    "pypr toggle volume")
      echo "Scratchpad: volume" ;;
    "pypr toggle files")
      echo "Scratchpad: files" ;;
    *"pypr toggle "*)
      local name="${arg##*pypr toggle }"
      echo "Scratchpad: $name" ;;
    *"hyprctl reload"*)
      echo "Reload config" ;;
    *"grim -g"*)
      echo "Screenshot region" ;;
    *"screenshot-crop"*)
      echo "Screenshot crop" ;;
    *"grim -"*)
      echo "Screenshot full" ;;
    *"playerctl play-pause"*)
      echo "Media play/pause" ;;
    *"playerctl next"*)
      echo "Media next" ;;
    *"playerctl previous"*)
      echo "Media previous" ;;
    *"wpctl set-mute"*)
      echo "Toggle mute" ;;
    *"wpctl set-volume"*"+")
      echo "Volume up" ;;
    *"wpctl set-volume"*"-")
      echo "Volume down" ;;
    *"brightnessctl"*"+")
      echo "Brightness up" ;;
    *"brightnessctl"*"-")
      echo "Brightness down" ;;
    *"hyprlock"*)
      echo "Lock screen" ;;
    *"eww open --toggle powermenu"*)
      echo "Toggle powermenu" ;;
    *"eww open --toggle dashboard"*)
      echo "Toggle dashboard" ;;
    *"eww open --toggle "*)
      local widget="${arg##*--toggle }"
      echo "Toggle $widget" ;;
    *"shader-cycle.sh next"*)
      echo "Shader next" ;;
    *"shader-cycle.sh prev"*)
      echo "Shader prev" ;;
    *"shader-random"*)
      echo "Shader random" ;;
    *"shader-wallpaper.sh next"*)
      echo "Wallpaper next" ;;
    *"shader-wallpaper.sh random"*)
      echo "Wallpaper random" ;;
    *"shader-wallpaper.sh static"*)
      echo "Wallpaper static" ;;
    *"touch "*"ghostty/config"*)
      echo "Reload Ghostty" ;;
    *"cliphist"*)
      echo "Clipboard history" ;;
    *"hyprpicker"*)
      echo "Color picker" ;;
    *"hyprsunset"*)
      echo "Toggle night light" ;;
    *"claude-session-picker"*)
      echo "Claude session picker" ;;
    *)
      # Fallback: basename of first word
      local cmd="${arg%% *}"
      cmd="${cmd##*/}"
      echo "$cmd" ;;
  esac
}

# ── Format keybind string ─────────────────────────
format_keybind() {
  local mods="$1" key="$2"
  local parts=()

  # Parse modifiers
  local mod_str
  mod_str="$(echo "$mods" | xargs)" # trim
  if [[ -n "$mod_str" ]]; then
    for m in $mod_str; do
      case "$m" in
        SUPER)  parts+=("${HG_MAGENTA}Super${HG_RESET}") ;;
        SHIFT)  parts+=("${HG_MAGENTA}Shift${HG_RESET}") ;;
        CTRL)   parts+=("${HG_MAGENTA}Ctrl${HG_RESET}") ;;
        ALT)    parts+=("${HG_MAGENTA}Alt${HG_RESET}") ;;
        *)      parts+=("${HG_MAGENTA}${m}${HG_RESET}") ;;
      esac
    done
  fi

  # Translate and add key
  local translated
  translated="$(translate_key "$key")"
  parts+=("${HG_GREEN}${translated}${HG_RESET}")

  # Join with dim +
  local sep="${HG_DIM}+${HG_RESET}"
  local result="${parts[0]}"
  for ((i=1; i<${#parts[@]}; i++)); do
    result+="${sep}${parts[$i]}"
  done
  echo "$result"
}

# Visible length (strip ANSI)
vis_len() {
  local stripped
  stripped="$(echo -e "$1" | sed 's/\x1b\[[0-9;]*m//g')"
  echo "${#stripped}"
}

# ── Main output generation ────────────────────────
generate() {
  local config_lines
  config_lines="$(resolve_config "$HYPR_CONF")"

  # Extract variables
  extract_vars <<< "$config_lines"

  local current_section="" current_submap="" first_section=true
  local COL_WIDTH=30

  # Title
  printf "\n  %s%s Hyprland Keybinds%s  %s(q=quit /=search)%s\n" \
    "$HG_CYAN" "$HG_BOLD" "$HG_RESET" "$HG_DIM" "$HG_RESET"

  while IFS= read -r line; do
    # Section headers
    if [[ "$line" =~ ^#[[:space:]]*──[[:space:]]*Keybinds:[[:space:]]*(.*[^─[:space:]])[[:space:]]*─ ]]; then
      current_section="${BASH_REMATCH[1]}"
      if $first_section; then
        first_section=false
      fi
      printf "\n  %s%s── %s ──%s\n\n" "$HG_CYAN" "$HG_BOLD" "$current_section" "$HG_RESET"
      continue
    fi

    # Submap tracking
    if [[ "$line" =~ ^submap[[:space:]]*=[[:space:]]*(.*) ]]; then
      local submap_name="${BASH_REMATCH[1]}"
      submap_name="$(echo "$submap_name" | xargs)"
      if [[ "$submap_name" == "reset" ]]; then
        current_submap=""
      else
        current_submap="$submap_name"
        printf "  %s[%s mode]%s\n\n" "$HG_YELLOW" "$submap_name" "$HG_RESET"
      fi
      continue
    fi

    # Bind lines
    if [[ "$line" =~ ^(bind[elmnr]*)[[:space:]]*=[[:space:]]*(.*) ]]; then
      local bind_type="${BASH_REMATCH[1]}"
      local bind_args="${BASH_REMATCH[2]}"

      # Skip mouse binds unless --all
      if [[ "$bind_type" == "bindm" ]] && ! $SHOW_ALL; then
        continue
      fi

      # Split on commas (max 4 fields: mods, key, dispatcher, arg)
      IFS=',' read -r raw_mods raw_key raw_disp raw_arg <<< "$bind_args"

      # Trim whitespace
      raw_mods="$(echo "$raw_mods" | xargs)"
      raw_key="$(echo "$raw_key" | xargs)"
      raw_disp="$(echo "$raw_disp" | xargs)"
      raw_arg="$(echo "$raw_arg" | xargs)"

      # Expand variables
      raw_mods="$(expand_vars "$raw_mods")"
      raw_arg="$(expand_vars "$raw_arg")"

      # Skip media keys unless --all
      if ! $SHOW_ALL && [[ "$raw_key" == XF86* ]]; then
        continue
      fi

      # Skip submap reset/escape binds (not useful as reference)
      if [[ "$raw_disp" == "submap" && "$raw_arg" == "reset" ]]; then
        continue
      fi

      # Format keybind and action
      local keybind action
      keybind="$(format_keybind "$raw_mods" "$raw_key")"
      action="$(dispatcher_action "$raw_disp" "$raw_arg")"

      # Pad keybind column
      local vlen
      vlen="$(vis_len "$keybind")"
      local padding=$((COL_WIDTH - vlen))
      ((padding < 2)) && padding=2
      local pad
      pad="$(printf '%*s' "$padding" '')"

      printf "  %s%s%s\n" "$keybind" "$pad" "$action"
    fi
  done <<< "$config_lines"

  printf "\n"
}

# ── Output ────────────────────────────────────────
if [[ -n "$SEARCH" ]]; then
  generate | less -R -i +/"$SEARCH"
else
  generate | less -R
fi
