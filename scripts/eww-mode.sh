#!/usr/bin/env bash
# eww-mode.sh — Compositor-aware mode/submap listener for eww bar
# Outputs the current keybind mode name on each change

_sway() {
  echo "default"
  swaymsg -t subscribe '["mode"]' | jq -r --unbuffered '.change'
}

_hyprland() {
  echo "default"
  socat -u "UNIX-CONNECT:/tmp/hypr/$HYPRLAND_INSTANCE_SIGNATURE/.socket2.sock" - 2>/dev/null | while read -r line; do
    case "$line" in
      submap\>\>)
        echo "default"
        ;;
      submap\>\>*)
        echo "${line#submap>>}"
        ;;
    esac
  done
}

if [[ -n "$HYPRLAND_INSTANCE_SIGNATURE" ]]; then
  _hyprland
elif [[ -n "$SWAYSOCK" ]]; then
  _sway
else
  echo "default"
fi
