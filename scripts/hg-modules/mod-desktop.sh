#!/usr/bin/env bash
# mod-desktop.sh — hg desktop scene helpers

desktop_description() {
  echo "Desktop scenes — Hyprland presets, layouts, and project bootstrap"
}

desktop_commands() {
  cat <<'CMDS'
preset	Save, list, or restore Hyprland monitor presets
layout	Save, list, or restore Hyprland window layouts
project	Open a repo scene with optional preset/layout/tmux bootstrap
status	Show saved scene inventory and managed state paths
CMDS
}

_desktop_require_hypr() {
  hg_require jq hyprctl
}

_desktop_state_root() {
  printf '%s\n' "${XDG_STATE_HOME:-$HOME/.local/state}/dotfiles/desktop-control"
}

_desktop_ensure_state_dir() {
  local dir="$(_desktop_state_root)/$1"
  mkdir -p "$dir"
  printf '%s\n' "$dir"
}

_desktop_safe_name() {
  local name
  name="$(printf '%s' "$1" | tr '[:upper:]' '[:lower:]' | sed -E 's/[^a-z0-9]+/-/g; s/^-+//; s/-+$//')"
  printf '%s\n' "${name:-unnamed}"
}

_desktop_monitor_preset_path() {
  local dir
  dir="$(_desktop_ensure_state_dir "hypr/monitor-presets")"
  printf '%s/%s.json\n' "$dir" "$(_desktop_safe_name "$1")"
}

_desktop_layout_path() {
  local dir
  dir="$(_desktop_ensure_state_dir "hypr/layouts")"
  printf '%s/%s.json\n' "$dir" "$(_desktop_safe_name "$1")"
}

_desktop_now() {
  date -u +%Y-%m-%dT%H:%M:%SZ
}

_desktop_print_table() {
  if command -v column >/dev/null 2>&1; then
    column -t -s $'\t'
  else
    cat
  fi
}

_desktop_hypr_monitors_json() {
  hyprctl monitors all -j 2>/dev/null || hyprctl monitors -j
}

_desktop_monitor_commands_from_file() {
  local path="$1"
  jq -r '
    def fmt:
      tostring
      | if index(".") == null then . else sub("0+$"; "") | sub("\\.$"; "") end;
    if (.monitors | length) == 0 then
      "ERROR:preset does not contain any monitors"
    else
      .monitors[] |
      if ((.name // "") | length) == 0 then
        "ERROR:monitor entry is missing name"
      elif ((.enabled // false) | not) then
        "\(.name),disable"
      elif ((.width // 0) <= 0 or (.height // 0) <= 0) then
        "ERROR:monitor \(.name) is missing width/height"
      elif ((.scale // 0) <= 0) then
        "ERROR:monitor \(.name) has invalid scale"
      else
        "\(.name),\((.width // 0) | floor)x\((.height // 0) | floor)@\((.refresh_rate_hz // 0) | fmt),\((.x // 0) | floor)x\((.y // 0) | floor),\((.scale // 1) | fmt),\((.transform // 0) | floor)"
      end
    end
  ' "$path"
}

_desktop_window_score() {
  local saved_json="$1" current_json="$2"
  jq -nr --argjson saved "$saved_json" --argjson current "$current_json" '
    def text(v): if v == null then "" else (v | tostring) end;
    def contains_ci(haystack; needle): (text(haystack) | ascii_downcase | contains(text(needle) | ascii_downcase));
    if (($current.mapped // true) | not) then
      -1
    else
      (if text($saved.class) != "" and (text($saved.class) | ascii_downcase) == (text($current.class) | ascii_downcase) then
        50
      elif text($saved.initial_class) != "" and (text($saved.initial_class) | ascii_downcase) == (text($current.initialClass) | ascii_downcase) then
        45
      else
        -1000
      end) as $base
      | if $base < 0 then
          -1
        else
          $base
          + (if text($saved.title) != "" then
               if text($current.title) == text($saved.title) then 35
               elif contains_ci($current.title; $saved.title) then 20
               else 0 end
             else 0 end)
          + (if text($saved.initial_title) != "" then
               if text($current.initialTitle) == text($saved.initial_title) then 15
               elif contains_ci($current.initialTitle; $saved.initial_title) then 10
               else 0 end
             else 0 end)
          + (if (($saved.workspace // 0) > 0 and (($saved.workspace // 0) == (($current.workspace.id // 0) | floor))) then 5 else 0 end)
        end
    end
  '
}

_desktop_apply_saved_window() {
  local saved_json="$1" current_json="$2" dry_run="$3"
  local address selector current_ws current_floating current_pinned current_fullscreen
  local saved_ws saved_floating saved_pinned saved_fullscreen saved_fullscreen_mode pos_x pos_y size_w size_h

  address="$(jq -r '.address // ""' <<<"$current_json")"
  selector="address:${address}"
  current_ws="$(jq -r '(.workspace.id // 0) | floor' <<<"$current_json")"
  current_floating="$(jq -r 'if (.floating // false) then "true" else "false" end' <<<"$current_json")"
  current_pinned="$(jq -r 'if (.pinned // false) then "true" else "false" end' <<<"$current_json")"
  current_fullscreen="$(jq -r 'if (.fullscreen // .fullscreenClient // false) then "true" else "false" end' <<<"$current_json")"

  saved_ws="$(jq -r '(.workspace // 0) | floor' <<<"$saved_json")"
  saved_floating="$(jq -r 'if (.floating // false) then "true" else "false" end' <<<"$saved_json")"
  saved_pinned="$(jq -r 'if (.pinned // false) then "true" else "false" end' <<<"$saved_json")"
  saved_fullscreen="$(jq -r 'if (.fullscreen // false) then "true" else "false" end' <<<"$saved_json")"
  saved_fullscreen_mode="$(jq -r '(.fullscreen_mode // 0) | floor' <<<"$saved_json")"
  pos_x="$(jq -r '(.position[0] // 0) | floor' <<<"$saved_json")"
  pos_y="$(jq -r '(.position[1] // 0) | floor' <<<"$saved_json")"
  size_w="$(jq -r '(.size[0] // 0) | floor' <<<"$saved_json")"
  size_h="$(jq -r '(.size[1] // 0) | floor' <<<"$saved_json")"

  _desktop_run_action() {
    local label="$1"
    shift
    if [[ "$dry_run" == "true" ]]; then
      hg_info "dry-run ${label}"
      return 0
    fi
    if hyprctl "$@" >/dev/null 2>&1; then
      hg_info "$label"
      return 0
    fi
    hg_warn "failed ${label}"
    return 1
  }

  if [[ "$saved_ws" -gt 0 && "$saved_ws" != "$current_ws" ]]; then
    _desktop_run_action "movetoworkspacesilent ${saved_ws} ${selector}" dispatch movetoworkspacesilent "${saved_ws},${selector}" || true
  fi

  if [[ "$saved_floating" != "$current_floating" ]]; then
    _desktop_run_action "togglefloating ${selector}" dispatch togglefloating "$selector" || true
  fi

  if [[ "$saved_floating" == "true" ]]; then
    _desktop_run_action "movewindowpixel exact ${pos_x} ${pos_y} ${selector}" dispatch movewindowpixel "exact ${pos_x} ${pos_y},${selector}" || true
    _desktop_run_action "resizewindowpixel exact ${size_w} ${size_h} ${selector}" dispatch resizewindowpixel "exact ${size_w} ${size_h},${selector}" || true
  fi

  if [[ "$saved_pinned" != "$current_pinned" ]]; then
    _desktop_run_action "pin ${selector}" dispatch pin "$selector" || true
  fi

  if [[ "$saved_fullscreen" != "$current_fullscreen" ]]; then
    _desktop_run_action "focuswindow ${selector}" dispatch focuswindow "$selector" || true
    _desktop_run_action "fullscreen ${saved_fullscreen_mode}" dispatch fullscreen "$saved_fullscreen_mode" || true
  fi
}

_desktop_status() {
  local preset_dir layout_dir preset_count layout_count
  preset_dir="$(_desktop_ensure_state_dir "hypr/monitor-presets")"
  layout_dir="$(_desktop_ensure_state_dir "hypr/layouts")"
  preset_count="$(find "$preset_dir" -maxdepth 1 -name '*.json' 2>/dev/null | wc -l | tr -d ' ')"
  layout_count="$(find "$layout_dir" -maxdepth 1 -name '*.json' 2>/dev/null | wc -l | tr -d ' ')"

  printf "\n %s%sdesktop status%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"
  printf " %s%-16s%s %s\n" "$HG_DIM" "state root" "$HG_RESET" "$(_desktop_state_root)"
  printf " %s%-16s%s %s (%s preset%s)\n" "$HG_DIM" "monitor presets" "$HG_RESET" "$preset_dir" "$preset_count" "$([[ "$preset_count" == "1" ]] && printf '' || printf 's')"
  printf " %s%-16s%s %s (%s layout%s)\n" "$HG_DIM" "layouts" "$HG_RESET" "$layout_dir" "$layout_count" "$([[ "$layout_count" == "1" ]] && printf '' || printf 's')"
  printf "\n"
}

_desktop_preset_save() {
  local name="${1:-}"
  [[ -n "$name" ]] || hg_die "Usage: hg desktop preset save <name>"
  _desktop_require_hypr

  local path monitors_json count
  path="$(_desktop_monitor_preset_path "$name")"
  monitors_json="$(_desktop_hypr_monitors_json)" || hg_die "Could not read Hyprland monitors"

  jq -n \
    --arg name "$name" \
    --arg saved_at "$(_desktop_now)" \
    --argjson monitors "$monitors_json" '
      {
        name: $name,
        saved_at: $saved_at,
        monitors: (
          $monitors
          | map({
              name: (.name // ""),
              enabled: (if (.disabled // false) then false else true end),
              width: ((.width // 0) | floor),
              height: ((.height // 0) | floor),
              refresh_rate_hz: (.refreshRate // 0),
              x: ((.x // 0) | floor),
              y: ((.y // 0) | floor),
              scale: (.scale // 1),
              transform: ((.transform // 0) | floor),
              description: (.description // "")
            })
          | sort_by((if .enabled then 0 else 1 end), .name)
        )
      }
    ' >"$path"
  printf '\n' >>"$path"

  count="$(jq -r '.monitors | length' "$path")"
  hg_ok "Saved monitor preset ${name} (${count} monitor(s))"
  printf '%s\n' "$path"
}

_desktop_preset_list() {
  local dir
  dir="$(_desktop_ensure_state_dir "hypr/monitor-presets")"
  if ! compgen -G "$dir/*.json" >/dev/null; then
    hg_info "No saved monitor presets"
    return 0
  fi

  printf "\n %s%smonitor presets%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"
  while IFS= read -r path; do
    jq -r --arg path "$path" '[.name // ($path | split("/")[-1] | sub("\\.json$"; "")), (.saved_at // ""), ((.monitors | length) | tostring), $path] | @tsv' "$path"
  done < <(find "$dir" -maxdepth 1 -name '*.json' | sort) | _desktop_print_table
  printf "\n"
}

_desktop_preset_restore() {
  local name="${1:-}" dry_run="${2:-false}"
  [[ -n "$name" ]] || hg_die "Usage: hg desktop preset restore <name> [--dry-run]"
  _desktop_require_hypr

  local path
  path="$(_desktop_monitor_preset_path "$name")"
  [[ -f "$path" ]] || hg_die "Unknown monitor preset: $name"

  local -a commands=()
  local -a errors=()
  local line
  while IFS= read -r line; do
    [[ -n "$line" ]] || continue
    if [[ "$line" == ERROR:* ]]; then
      errors+=("${line#ERROR:}")
    else
      commands+=("$line")
    fi
  done < <(_desktop_monitor_commands_from_file "$path")

  if [[ "${#errors[@]}" -gt 0 ]]; then
    printf "\n %s%smonitor preset validation failed%s\n\n" "$HG_BOLD" "$HG_RED" "$HG_RESET" >&2
    printf '  - %s\n' "${errors[@]}" >&2
    return 1
  fi

  printf "\n %s%srestore monitor preset%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"
  local command
  for command in "${commands[@]}"; do
    if [[ "$dry_run" == "true" ]]; then
      hg_info "dry-run hyprctl keyword monitor ${command}"
      continue
    fi
    if hyprctl keyword monitor "$command" >/dev/null 2>&1; then
      hg_ok "Applied ${command}"
    else
      hg_warn "Failed ${command}"
    fi
  done
  printf "\n"
}

_desktop_layout_save() {
  local name="$1"
  shift
  [[ -n "$name" ]] || hg_die "Usage: hg desktop layout save <name> [--workspace N] [--launch key=command]"
  _desktop_require_hypr

  local -a workspace_ids=()
  local -a launch_specs=()
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --workspace)
        [[ $# -ge 2 ]] || hg_die "--workspace requires a value"
        workspace_ids+=("$2")
        shift 2
        ;;
      --launch)
        [[ $# -ge 2 ]] || hg_die "--launch requires key=command"
        launch_specs+=("$2")
        shift 2
        ;;
      *)
        hg_die "Unknown option for layout save: $1"
        ;;
    esac
  done

  local path clients_json workspaces_json monitors_json workspace_json launch_json
  path="$(_desktop_layout_path "$name")"
  clients_json="$(hyprctl clients -j)" || hg_die "Could not read Hyprland clients"
  workspaces_json="$(hyprctl workspaces -j 2>/dev/null || printf '[]\n')"
  monitors_json="$(_desktop_hypr_monitors_json)" || hg_die "Could not read Hyprland monitors"
  workspace_json="$(printf '%s\n' "${workspace_ids[@]}" | jq -Rcs 'split("\n") | map(select(length > 0) | tonumber)')"
  launch_json="$(printf '%s\n' "${launch_specs[@]}" | jq -Rcs '
    split("\n")
    | map(select(length > 0))
    | reduce .[] as $line ({};
        ($line | capture("^(?<key>[^=]+)=(?<value>.*)$")) as $pair
        | .[$pair.key] = $pair.value
      )
  ')"

  jq -n \
    --arg name "$name" \
    --arg saved_at "$(_desktop_now)" \
    --argjson clients "$clients_json" \
    --argjson workspaces "$workspaces_json" \
    --argjson monitors "$monitors_json" \
    --argjson workspace_ids "$workspace_json" \
    --argjson launches "$launch_json" '
      def monitor_name($id): first($monitors[]? | select(((.id // -1) | floor) == $id) | .name) // "";
      {
        name: $name,
        saved_at: $saved_at,
        windows: (
          $clients
          | map(
              select(
                ($workspace_ids | length) == 0
                or (($workspace_ids | index((.workspace.id // 0) | floor)) != null)
              )
            )
          | map({
              class: (.class // ""),
              title: (.title // ""),
              initial_class: (.initialClass // ""),
              initial_title: (.initialTitle // ""),
              workspace: ((.workspace.id // 0) | floor),
              workspace_name: (.workspace.name // ""),
              monitor: ((.monitor // 0) | floor),
              monitor_name: monitor_name((.monitor // 0) | floor),
              floating: (.floating // false),
              mapped: (if .mapped == null then true else .mapped end),
              pinned: (.pinned // false),
              fullscreen: (
                if (.fullscreen | type) == "boolean" then .fullscreen
                elif (.fullscreenClient | type) == "boolean" then .fullscreenClient
                else ((.fullscreen // 0) != 0)
                end
              ),
              fullscreen_mode: ((.fullscreenMode // 0) | floor),
              position: [((.at[0] // 0) | floor), ((.at[1] // 0) | floor)],
              size: [((.size[0] // 0) | floor), ((.size[1] // 0) | floor)],
              launch_command: (
                $launches[.class]
                // $launches[.initialClass]
                // $launches[.title]
                // ""
              )
            })
          | sort_by(.workspace, .class, .title)
        ),
        workspaces: (
          $workspaces
          | map({
              id: ((.id // 0) | floor),
              name: (.name // ""),
              monitor: (.monitor // "")
            })
          | sort_by(.id)
        )
      }
    ' >"$path"
  printf '\n' >>"$path"

  hg_ok "Saved layout ${name} ($(jq -r '.windows | length' "$path") window(s))"
  printf '%s\n' "$path"
}

_desktop_layout_list() {
  local dir
  dir="$(_desktop_ensure_state_dir "hypr/layouts")"
  if ! compgen -G "$dir/*.json" >/dev/null; then
    hg_info "No saved layouts"
    return 0
  fi

  printf "\n %s%slayouts%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"
  while IFS= read -r path; do
    jq -r --arg path "$path" '[.name // ($path | split("/")[-1] | sub("\\.json$"; "")), (.saved_at // ""), ((.windows | length) | tostring), $path] | @tsv' "$path"
  done < <(find "$dir" -maxdepth 1 -name '*.json' | sort) | _desktop_print_table
  printf "\n"
}

_desktop_layout_restore() {
  local name="$1"
  shift
  [[ -n "$name" ]] || hg_die "Usage: hg desktop layout restore <name> [--dry-run] [--no-launch]"
  _desktop_require_hypr

  local dry_run=false launch_missing=true
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --dry-run) dry_run=true ;;
      --no-launch) launch_missing=false ;;
      *) hg_die "Unknown option for layout restore: $1" ;;
    esac
    shift
  done

  local path
  path="$(_desktop_layout_path "$name")"
  [[ -f "$path" ]] || hg_die "Unknown layout: $name"

  local clients_json
  clients_json="$(hyprctl clients -j)" || hg_die "Could not read Hyprland clients"
  mapfile -t current_clients < <(jq -c '.[]' <<<"$clients_json")
  mapfile -t saved_windows < <(jq -c '.windows[]?' "$path")

  local -A used_current=()
  local -a pending_launches=()
  local -a unresolved=()
  local saved_json current_json
  local idx best_idx best_score score launch_command label

  printf "\n %s%srestore layout%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"

  for saved_json in "${saved_windows[@]}"; do
    best_idx=-1
    best_score=-1
    for idx in "${!current_clients[@]}"; do
      [[ -n "${used_current[$idx]:-}" ]] && continue
      score="$(_desktop_window_score "$saved_json" "${current_clients[$idx]}")"
      if [[ "$score" -gt "$best_score" ]]; then
        best_idx="$idx"
        best_score="$score"
      fi
    done

    if [[ "$best_idx" -ge 0 && "$best_score" -ge 50 ]]; then
      used_current["$best_idx"]=1
      label="$(jq -r '[.class // "", .title // ""] | map(select(length > 0)) | join(" — ")' <<<"$saved_json")"
      hg_ok "Matched ${label:-window}"
      _desktop_apply_saved_window "$saved_json" "${current_clients[$best_idx]}" "$dry_run"
      continue
    fi

    launch_command="$(jq -r '.launch_command // ""' <<<"$saved_json")"
    label="$(jq -r '[.class // "", .title // ""] | map(select(length > 0)) | join(" — ")' <<<"$saved_json")"
    if [[ "$launch_missing" == "true" && -n "$launch_command" ]]; then
      pending_launches+=("$saved_json")
      if [[ "$dry_run" == "true" ]]; then
        hg_info "dry-run launch ${label:-window}: ${launch_command}"
      else
        sh -lc "$launch_command" >/dev/null 2>&1 &
        hg_info "Launched ${label:-window}: ${launch_command}"
      fi
      continue
    fi

    unresolved+=("${label:-window}")
    hg_warn "Unresolved ${label:-window}"
  done

  if [[ "$dry_run" == "false" && "${#pending_launches[@]}" -gt 0 ]]; then
    sleep 1
    clients_json="$(hyprctl clients -j)" || true
    mapfile -t current_clients < <(jq -c '.[]' <<<"${clients_json:-[]}")
    for saved_json in "${pending_launches[@]}"; do
      best_idx=-1
      best_score=-1
      for idx in "${!current_clients[@]}"; do
        [[ -n "${used_current[$idx]:-}" ]] && continue
        score="$(_desktop_window_score "$saved_json" "${current_clients[$idx]}")"
        if [[ "$score" -gt "$best_score" ]]; then
          best_idx="$idx"
          best_score="$score"
        fi
      done
      label="$(jq -r '[.class // "", .title // ""] | map(select(length > 0)) | join(" — ")' <<<"$saved_json")"
      if [[ "$best_idx" -ge 0 && "$best_score" -ge 50 ]]; then
        used_current["$best_idx"]=1
        hg_ok "Matched launched ${label:-window}"
        _desktop_apply_saved_window "$saved_json" "${current_clients[$best_idx]}" false
      else
        unresolved+=("${label:-window}")
        hg_warn "Launch did not produce a match yet: ${label:-window}"
      fi
    done
  fi

  if [[ "${#unresolved[@]}" -gt 0 ]]; then
    printf "\n %sUNRESOLVED%s\n" "$HG_BOLD" "$HG_RESET"
    printf '  %s\n' "${unresolved[@]}"
    printf "\n"
  fi
}

_desktop_project_open() {
  local repo_path="${1:-}"
  shift || true
  [[ -n "$repo_path" ]] || hg_die "Usage: hg desktop project open <repo_path> [--monitor name] [--layout name] [--tmux-session name] [--tmux-command cmd] [--no-launch] [--dry-run]"
  [[ -d "$repo_path" ]] || hg_die "repo_path must be a directory"

  local monitor_preset="" layout="" tmux_session="" tmux_command="" dry_run=false launch_missing=true
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --monitor)
        [[ $# -ge 2 ]] || hg_die "--monitor requires a preset name"
        monitor_preset="$2"
        shift 2
        ;;
      --layout)
        [[ $# -ge 2 ]] || hg_die "--layout requires a layout name"
        layout="$2"
        shift 2
        ;;
      --tmux-session)
        [[ $# -ge 2 ]] || hg_die "--tmux-session requires a name"
        tmux_session="$2"
        shift 2
        ;;
      --tmux-command)
        [[ $# -ge 2 ]] || hg_die "--tmux-command requires a command"
        tmux_command="$2"
        shift 2
        ;;
      --no-launch)
        launch_missing=false
        shift
        ;;
      --dry-run)
        dry_run=true
        shift
        ;;
      *)
        hg_die "Unknown option for project open: $1"
        ;;
    esac
  done

  [[ -n "$tmux_session" ]] || tmux_session="$(basename "$repo_path")"
  printf "\n %s%sproject open%s\n\n" "$HG_BOLD" "$HG_CYAN" "$HG_RESET"
  hg_info "repo ${repo_path}"

  if [[ -n "$monitor_preset" ]]; then
    _desktop_preset_restore "$monitor_preset" "$dry_run"
  fi
  if [[ -n "$layout" ]]; then
    local -a layout_args=("$layout")
    [[ "$dry_run" == "true" ]] && layout_args+=(--dry-run)
    [[ "$launch_missing" == "false" ]] && layout_args+=(--no-launch)
    _desktop_layout_restore "${layout_args[@]}"
  fi

  if [[ "$dry_run" == "true" ]]; then
    hg_info "dry-run tmux session ${tmux_session} in ${repo_path}"
    [[ -n "$tmux_command" ]] && hg_info "dry-run tmux command ${tmux_command}"
    printf "\n"
    return 0
  fi

  hg_require tmux
  if tmux has-session -t "$tmux_session" >/dev/null 2>&1; then
    hg_ok "tmux session exists ${tmux_session}"
    printf "\n"
    return 0
  fi

  local -a tmux_args=(new-session -d -s "$tmux_session" -c "$repo_path")
  [[ -n "$tmux_command" ]] && tmux_args+=("$tmux_command")
  if tmux "${tmux_args[@]}" >/dev/null 2>&1; then
    hg_ok "Created tmux session ${tmux_session}"
  else
    hg_warn "Failed to create tmux session ${tmux_session}"
  fi
  printf "\n"
}

_desktop_preset_cmd() {
  local subcmd="${1:-}"
  shift || true
  case "$subcmd" in
    save) _desktop_preset_save "$@" ;;
    list) _desktop_preset_list "$@" ;;
    restore)
      local name="${1:-}"
      shift || true
      local dry_run=false
      while [[ $# -gt 0 ]]; do
        case "$1" in
          --dry-run) dry_run=true ;;
          *) hg_die "Unknown option for preset restore: $1" ;;
        esac
        shift
      done
      _desktop_preset_restore "$name" "$dry_run"
      ;;
    *) hg_die "Unknown desktop preset command: ${subcmd:-<empty>}" ;;
  esac
}

_desktop_layout_cmd() {
  local subcmd="${1:-}"
  shift || true
  case "$subcmd" in
    save) _desktop_layout_save "$@" ;;
    list) _desktop_layout_list "$@" ;;
    restore) _desktop_layout_restore "$@" ;;
    *) hg_die "Unknown desktop layout command: ${subcmd:-<empty>}" ;;
  esac
}

_desktop_project_cmd() {
  local subcmd="${1:-}"
  shift || true
  case "$subcmd" in
    open) _desktop_project_open "$@" ;;
    *) hg_die "Unknown desktop project command: ${subcmd:-<empty>}" ;;
  esac
}

desktop_run() {
  local cmd="${1:-}"
  shift || true

  case "$cmd" in
    preset) _desktop_preset_cmd "$@" ;;
    layout) _desktop_layout_cmd "$@" ;;
    project) _desktop_project_cmd "$@" ;;
    status) _desktop_status ;;
    *) hg_die "Unknown desktop command: ${cmd:-<empty>}. Run 'hg desktop --help'." ;;
  esac
}
