#!/usr/bin/env bash
# quickshell-dock-data.sh - data and actions for the Quickshell dock.

set -euo pipefail

action="${1:-list}"
item_id="${2:-}"

spawn() {
  setsid -f "$@" >/dev/null 2>&1 || nohup "$@" >/dev/null 2>&1 &
}

hypr_clients() {
  hyprctl -j clients 2>/dev/null || printf '[]\n'
}

active_address() {
  hyprctl -j activewindow 2>/dev/null | jq -r '.address // ""' 2>/dev/null || true
}

pin_ids() {
  printf '%s\n' terminal browser files editor music chat
}

pin_title() {
  case "$1" in
    terminal) printf 'Terminal' ;;
    browser) printf 'Browser' ;;
    files) printf 'Files' ;;
    editor) printf 'Editor' ;;
    music) printf 'Music' ;;
    chat) printf 'Chat' ;;
    *) printf '%s' "$1" ;;
  esac
}

pin_badge() {
  case "$1" in
    terminal) printf 'TERM' ;;
    browser) printf 'WEB' ;;
    files) printf 'FILE' ;;
    editor) printf 'CODE' ;;
    music) printf 'AUD' ;;
    chat) printf 'MSG' ;;
    *) printf 'APP' ;;
  esac
}

pin_regex() {
  case "$1" in
    terminal) printf 'kitty|Alacritty|foot|wezterm|org.wezfurlong.wezterm' ;;
    browser) printf 'zen|firefox|brave|chromium|chrome|librewolf' ;;
    files) printf 'org.gnome.Nautilus|Nautilus|thunar|dolphin|pcmanfm' ;;
    editor) printf 'Code|code|codium|cursor|nvim|neovide|Zed' ;;
    music) printf 'Spotify|spotify|spotify-launcher|org.gnome.Lollypop|Tauon' ;;
    chat) printf 'vesktop|discord|Slack|slack|signal|Element' ;;
    *) printf '^$' ;;
  esac
}

pin_launchable() {
  case "$1" in
    terminal) command -v kitty >/dev/null || command -v alacritty >/dev/null || command -v foot >/dev/null ;;
    browser) command -v zen-browser >/dev/null || command -v firefox >/dev/null || command -v brave-browser >/dev/null || command -v chromium >/dev/null || command -v xdg-open >/dev/null ;;
    files) command -v nautilus >/dev/null || command -v thunar >/dev/null || command -v dolphin >/dev/null || command -v xdg-open >/dev/null ;;
    editor) command -v code >/dev/null || command -v cursor >/dev/null || command -v codium >/dev/null || command -v zed >/dev/null || command -v kitty >/dev/null ;;
    music) command -v spotify >/dev/null || command -v spotify-launcher >/dev/null || command -v tauon >/dev/null ;;
    chat) command -v vesktop >/dev/null || command -v discord >/dev/null || command -v slack >/dev/null || command -v signal-desktop >/dev/null || command -v element-desktop >/dev/null ;;
    *) return 1 ;;
  esac
}

launch_pin() {
  case "$1" in
    terminal)
      if command -v kitty >/dev/null; then spawn kitty
      elif command -v alacritty >/dev/null; then spawn alacritty
      elif command -v foot >/dev/null; then spawn foot
      else return 1
      fi
      ;;
    browser)
      if command -v zen-browser >/dev/null; then spawn zen-browser
      elif command -v firefox >/dev/null; then spawn firefox
      elif command -v brave-browser >/dev/null; then spawn brave-browser
      elif command -v chromium >/dev/null; then spawn chromium
      else spawn xdg-open about:blank
      fi
      ;;
    files)
      if command -v nautilus >/dev/null; then spawn nautilus
      elif command -v thunar >/dev/null; then spawn thunar
      elif command -v dolphin >/dev/null; then spawn dolphin
      else spawn xdg-open "$HOME"
      fi
      ;;
    editor)
      if command -v code >/dev/null; then spawn code
      elif command -v cursor >/dev/null; then spawn cursor
      elif command -v codium >/dev/null; then spawn codium
      elif command -v zed >/dev/null; then spawn zed
      else spawn kitty nvim
      fi
      ;;
    music)
      if command -v spotify >/dev/null; then spawn spotify
      elif command -v spotify-launcher >/dev/null; then spawn spotify-launcher
      else spawn tauon
      fi
      ;;
    chat)
      if command -v vesktop >/dev/null; then spawn vesktop
      elif command -v discord >/dev/null; then spawn discord
      elif command -v slack >/dev/null; then spawn slack
      elif command -v signal-desktop >/dev/null; then spawn signal-desktop
      else spawn element-desktop
      fi
      ;;
    *)
      return 1
      ;;
  esac
}

focus_address() {
  local address="$1"
  [[ -n "$address" ]] || return 1
  hyprctl dispatch focuswindow "address:$address" >/dev/null 2>&1
}

first_window_for_pin() {
  local id="$1"
  local clients regex
  clients="$(hypr_clients)"
  regex="$(pin_regex "$id")"
  jq -r --arg re "$regex" '[.[] | select((.class // "") | test($re; "i"))][0].address // ""' <<<"$clients"
}

activate_pin() {
  local id="$1" address
  address="$(first_window_for_pin "$id")"
  if [[ -n "$address" ]]; then
    focus_address "$address"
  else
    launch_pin "$id"
  fi
}

combined_pin_regex() {
  local first=true id
  while IFS= read -r id; do
    if $first; then
      first=false
    else
      printf '|'
    fi
    printf '(%s)' "$(pin_regex "$id")"
  done < <(pin_ids)
}

list_entries() {
  local clients active pins_json windows_json combined id regex running first_address is_active title badge
  clients="$(hypr_clients)"
  active="$(active_address)"
  pins_json="[]"

  while IFS= read -r id; do
    pin_launchable "$id" || continue
    regex="$(pin_regex "$id")"
    title="$(pin_title "$id")"
    badge="$(pin_badge "$id")"
    running="$(jq --arg re "$regex" '[.[] | select((.class // "") | test($re; "i"))] | length' <<<"$clients")"
    first_address="$(jq -r --arg re "$regex" '[.[] | select((.class // "") | test($re; "i"))][0].address // ""' <<<"$clients")"
    is_active="$(jq -r --arg re "$regex" --arg active "$active" '[.[] | select((.address // "") == $active and ((.class // "") | test($re; "i")))] | length > 0' <<<"$clients")"
    pins_json="$(jq \
      --arg id "$id" \
      --arg title "$title" \
      --arg badge "$badge" \
      --arg address "$first_address" \
      --argjson running "$running" \
      --argjson active "$is_active" \
      '. + [{id:$id,title:$title,subtitle:($running|tostring)+" window(s)",badge:$badge,address:$address,running:$running,active:$active,pinned:true}]' \
      <<<"$pins_json")"
  done < <(pin_ids)

  combined="$(combined_pin_regex)"
  windows_json="$(jq -c --arg active "$active" --arg re "$combined" '
    [.[] |
      select(.mapped != false) |
      select(((.class // "") | test($re; "i")) | not) |
      {
        id: ("window:" + (.address // "")),
        title: (.class // "Window"),
        subtitle: (.title // ""),
        badge: "WIN",
        address: (.address // ""),
        running: 1,
        active: ((.address // "") == $active),
        pinned: false
      }
    ][0:8]' <<<"$clients")"

  jq -n -c --argjson pins "$pins_json" --argjson windows "$windows_json" '{entries: ($pins + $windows), updated: now}'
}

case "$action" in
  list)
    list_entries
    ;;
  activate)
    if [[ "$item_id" == window:* ]]; then
      focus_address "${item_id#window:}"
    else
      activate_pin "$item_id"
    fi
    ;;
  launch)
    if [[ "$item_id" == window:* ]]; then
      focus_address "${item_id#window:}"
    else
      launch_pin "$item_id"
    fi
    ;;
  *)
    printf 'Usage: %s <list|activate|launch> [id]\n' "${0##*/}" >&2
    exit 2
    ;;
esac
