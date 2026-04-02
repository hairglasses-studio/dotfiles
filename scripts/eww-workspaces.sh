#!/usr/bin/env bash
# eww-workspaces.sh — Compositor-aware workspace listener for eww bar
# Outputs JSON: [{num, focused, name, urgent}] on each workspace change

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib/compositor.sh"

_sway() {
  local monitor="${1:-}"

  _sway_workspaces() {
    local mon_filter="$1"
    if [[ -n "$mon_filter" ]]; then
      swaymsg -t get_workspaces | jq -c \
        --arg mon "$mon_filter" \
        '[.[] | select(.output == $mon) | {num, focused, name, urgent}]'
    else
      swaymsg -t get_workspaces | jq -c '[.[] | {num, focused, name, urgent}]'
    fi
  }

  while true; do
    _sway_workspaces "$monitor"
    swaymsg -t subscribe '["workspace"]' > /dev/null 2>&1
  done
}

_hyprland() {
  local monitor="${1:-}"

  # Static workspace-to-monitor mapping (matches monitors.conf)
  declare -A WS_MAP=(
    [DP-1]="1 2 3"
    [DP-2]="4 5 6 7 8 9"
  )

  _hypr_workspaces() {
    local mon_filter="$1"
    local monitors_json active_id

    monitors_json=$(hyprctl monitors -j 2>/dev/null)

    if [[ -n "$mon_filter" ]]; then
      active_id=$(printf '%s' "$monitors_json" | jq -r \
        --arg mon "$mon_filter" \
        '[.[] | select(.name == $mon)] | .[0].activeWorkspace.id // 0')

      local expected="${WS_MAP[$mon_filter]:-}"
      if [[ -n "$expected" ]]; then
        local expected_json actual_json
        expected_json=$(echo "$expected" | tr ' ' '\n' | jq -R 'tonumber' | jq -s '.')
        actual_json=$(hyprctl workspaces -j 2>/dev/null)

        printf '%s' "$actual_json" | jq -c \
          --argjson active "$active_id" \
          --argjson expected "$expected_json" \
          --arg mon "$mon_filter" \
          '. as $workspaces |
          [$expected[] | . as $id |
            ([$workspaces[] | select(.id == $id and .monitor == $mon)] | first // null) as $ws |
            {num: $id, focused: ($id == $active), name: ($ws.name // "\($id)"), urgent: false,
             occupied: ($ws != null)}
          ] | sort_by(.num)'
      else
        hyprctl workspaces -j 2>/dev/null | jq -c \
          --argjson active "$active_id" \
          --arg mon "$mon_filter" \
          '[.[] | select(.monitor == $mon) | {num: .id, focused: (.id == $active), name: (.name // "\(.id)"), urgent: false, occupied: true}] | sort_by(.num)'
      fi
    else
      active_id=$(hyprctl activeworkspace -j 2>/dev/null | jq -r '.id // 0')

      hyprctl workspaces -j 2>/dev/null | jq -c \
        --argjson active "$active_id" \
        '[.[] | {num: .id, focused: (.id == $active), name: (.name // "\(.id)"), urgent: false, occupied: true}] | sort_by(.num)'
    fi
  }

  _hypr_workspaces "$monitor"
  socat -u "UNIX-CONNECT:/tmp/hypr/$HYPRLAND_INSTANCE_SIGNATURE/.socket2.sock" - 2>/dev/null | while read -r line; do
    case "$line" in
      workspace*|createworkspace*|destroyworkspace*|focusedmon*|moveworkspace*|activespecial*)
        _hypr_workspaces "$monitor"
        ;;
    esac
  done
}

case "$(compositor_type)" in
  hyprland) _hyprland "$1" ;;
  sway)     _sway "$1" ;;
  *)        echo "[]" ;;
esac
