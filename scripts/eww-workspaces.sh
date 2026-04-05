#!/usr/bin/env bash
# eww-workspaces.sh — Compositor-aware workspace listener for eww bar
# Outputs JSON: [{num, focused, name, urgent}] on each workspace change

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
source "$SCRIPT_DIR/lib/compositor.sh"

_hyprland() {
  local monitor="${1:-}"

  # Build workspace-to-monitor mapping dynamically from Hyprland config.
  # workspacerules returns desc: strings; resolve to connector names via monitors.
  declare -A WS_MAP
  local _monitors_init
  _monitors_init=$(hyprctl monitors -j 2>/dev/null)

  while IFS= read -r line; do
    local ws_id rule_mon resolved
    ws_id=$(echo "$line" | jq -r '.id')
    rule_mon=$(echo "$line" | jq -r '.monitor')

    # Resolve desc:... to connector name (e.g. "desc:XEC ES-G32C1Q" -> "DP-1")
    if [[ "$rule_mon" == desc:* ]]; then
      local desc_substr="${rule_mon#desc:}"
      resolved=$(printf '%s' "$_monitors_init" | jq -r \
        --arg d "$desc_substr" \
        '[.[] | select(.description | startswith($d))] | .[0].name // empty')
    else
      resolved="$rule_mon"
    fi

    if [[ -n "$resolved" ]]; then
      WS_MAP[$resolved]="${WS_MAP[$resolved]:+${WS_MAP[$resolved]} }$ws_id"
    fi
  done < <(hyprctl workspacerules -j 2>/dev/null | jq -c '.[] | select(.monitor != "") | {id: (.workspaceString | tonumber? // empty), monitor}' 2>/dev/null)

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
          ] | sort_by(.num)
          | . as $arr | [range(length) | . as $i | $arr[$i] + {
            prev_empty: (if $i == 0 then true elif ($arr[$i-1].occupied | not) then true else false end),
            next_empty: (if $i == (($arr|length)-1) then true elif ($arr[$i+1].occupied | not) then true else false end)
          }]'
      else
        hyprctl workspaces -j 2>/dev/null | jq -c \
          --argjson active "$active_id" \
          --arg mon "$mon_filter" \
          '[.[] | select(.monitor == $mon) | {num: .id, focused: (.id == $active), name: (.name // "\(.id)"), urgent: false, occupied: true}] | sort_by(.num)
          | . as $arr | [range(length) | . as $i | $arr[$i] + {
            prev_empty: (if $i == 0 then true elif ($arr[$i-1].occupied | not) then true else false end),
            next_empty: (if $i == (($arr|length)-1) then true elif ($arr[$i+1].occupied | not) then true else false end)
          }]'
      fi
    else
      active_id=$(hyprctl activeworkspace -j 2>/dev/null | jq -r '.id // 0')

      hyprctl workspaces -j 2>/dev/null | jq -c \
        --argjson active "$active_id" \
        '[.[] | {num: .id, focused: (.id == $active), name: (.name // "\(.id)"), urgent: false, occupied: true}] | sort_by(.num)
        | . as $arr | [range(length) | . as $i | $arr[$i] + {
          prev_empty: (if $i == 0 then true elif ($arr[$i-1].occupied | not) then true else false end),
          next_empty: (if $i == (($arr|length)-1) then true elif ($arr[$i+1].occupied | not) then true else false end)
        }]'
    fi
  }

  _hypr_listen() {
    _hypr_workspaces "$monitor"
    socat -u "UNIX-CONNECT:$(hypr_socket2)" - 2>/dev/null | while read -r line; do
      case "$line" in
        workspace*|createworkspace*|destroyworkspace*|focusedmon*|moveworkspace*|activespecial*)
          _hypr_workspaces "$monitor"
          ;;
      esac
    done
  }

  resilient_listen _hypr_listen
}

case "$(compositor_type)" in
  hyprland) _hyprland "$1" ;;
  *)        echo "[]" ;;
esac
