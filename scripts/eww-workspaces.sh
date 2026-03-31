#!/usr/bin/env bash
# eww-workspaces.sh — Compositor-aware workspace listener for eww bar
# Outputs JSON: [{num, focused, name, urgent}] on each workspace change

_sway() {
  while true; do
    swaymsg -t get_workspaces | jq -c '[.[] | {num, focused, name, urgent}]'
    swaymsg -t subscribe '["workspace"]' > /dev/null 2>&1
  done
}

_hyprland() {
  _hypr_workspaces() {
    local active
    active=$(hyprctl activeworkspace -j 2>/dev/null | jq -r '.id // 0')
    hyprctl workspaces -j 2>/dev/null | jq -c --argjson active "$active" \
      '[.[] | {num: .id, focused: (.id == $active), name: (.name // "\(.id)"), urgent: false}] | sort_by(.num)'
  }
  _hypr_workspaces
  socat -u "UNIX-CONNECT:/tmp/hypr/$HYPRLAND_INSTANCE_SIGNATURE/.socket2.sock" - 2>/dev/null | while read -r line; do
    case "$line" in
      workspace*|createworkspace*|destroyworkspace*|focusedmon*|moveworkspace*)
        _hypr_workspaces
        ;;
    esac
  done
}

if [[ -n "$HYPRLAND_INSTANCE_SIGNATURE" ]]; then
  _hyprland
elif [[ -n "$SWAYSOCK" ]]; then
  _sway
else
  echo "[]"
fi
