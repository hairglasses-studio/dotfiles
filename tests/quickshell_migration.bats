#!/usr/bin/env bats

load 'test_helper'

@test "ticker bridge lists streams and emits plain text payloads" {
    run python3 "${SCRIPTS_DIR}/ticker-bridge.py" --list
    assert_success
    assert_output --partial '"keybinds"'
    assert_output --partial '"notifications"'

    run python3 -c '
import json, subprocess, sys
path = sys.argv[1]
payload = json.loads(subprocess.check_output(["python3", path, "--stream", "weather", "--once"]))
assert payload["ok"] is True
assert payload["stream"] == "weather"
assert "markup" in payload
assert "text" in payload
assert "<span" not in payload["text"]
print("stream=%s text_len=%d" % (payload["stream"], len(payload["text"])))
' "${SCRIPTS_DIR}/ticker-bridge.py"
    assert_success
    assert_output --partial "stream=weather"
}

@test "fleet telemetry bridge emits stateful numeric samples" {
    export XDG_STATE_HOME="${BATS_TEST_TMPDIR}/state"

    run python3 "${SCRIPTS_DIR}/fleet-telemetry-bridge.py" --once
    assert_success

    run python3 -c '
import json, sys
payload = json.loads(sys.argv[1])
assert payload["ok"] is True
for key in ("cpu_pct", "mem_pct", "net_kbps", "gpu_temp_c"):
    assert isinstance(payload[key], (int, float)), key
for key in ("cpu_text", "mem_text", "net_text", "gpu_text"):
    assert isinstance(payload[key], str) and payload[key], key
print("telemetry=%s/%s/%s/%s" % (payload["cpu_text"], payload["mem_text"], payload["net_text"], payload["gpu_text"]))
' "$output"
    assert_success
    assert_output --partial "telemetry=CPU"
}

@test "lyrics bridge emits MPRIS fallback without network fetch" {
    export PLAYERCTL_BIN="${BATS_TEST_TMPDIR}/playerctl"
    export LYRICS_BRIDGE_NO_FETCH=1
    cat > "$PLAYERCTL_BIN" <<'MOCK'
#!/usr/bin/env bash
case "${3:-}" in
  "{{status}}") printf 'Playing\n' ;;
  "{{title}}") printf 'Test Song\n' ;;
  "{{artist}}") printf 'Test Artist\n' ;;
  "{{album}}") printf 'Test Album\n' ;;
  "{{mpris:length}}") printf '180000000\n' ;;
  "{{position}}") printf '30000000\n' ;;
esac
MOCK
    chmod +x "$PLAYERCTL_BIN"

    run python3 "${SCRIPTS_DIR}/lyrics-bridge.py" --once
    assert_success
    run python3 -c '
import json, sys
payload = json.loads(sys.argv[1])
assert payload["ok"] is True
assert payload["status"] == "Playing"
assert payload["line"] == "Test Song -- Test Artist"
assert payload["synced"] is False
print(payload["line"])
' "$output"
    assert_success
    assert_output "Test Song -- Test Artist"
}

@test "notification bridge summarizes local history without owning dbus" {
    export XDG_STATE_HOME="${BATS_TEST_TMPDIR}/state"
    local history="${XDG_STATE_HOME}/dotfiles/desktop-control/notifications/history.jsonl"
    mkdir -p "$(dirname "$history")"
    cat > "$history" <<'JSONL'
{"id":"1","app":"low-app","summary":"Low","body":"","urgency":"low","visible":true,"dismissed":false}
{"id":"2","app":"crit-app","summary":"Critical","body":"Body","urgency":"critical","visible":true,"dismissed":false}
{"id":"3","app":"old-app","summary":"Old","body":"","urgency":"normal","visible":true,"dismissed":true}
JSONL

    run python3 "${SCRIPTS_DIR}/notification-bridge.py" --limit 5
    assert_success
    run python3 -c '
import json, sys
payload = json.loads(sys.argv[1])
assert payload["ok"] is True
assert payload["visible"] == 2
assert payload["critical"] == 1
assert payload["latest"]["app"] == "crit-app"
assert payload["latest"]["text"] == "Critical: Body"
print("visible=%d critical=%d" % (payload["visible"], payload["critical"]))
' "$output"
    assert_success
    assert_output "visible=2 critical=1"
}

@test "notification bridge can append and clear local history" {
    export XDG_STATE_HOME="${BATS_TEST_TMPDIR}/state"
    local history="${XDG_STATE_HOME}/dotfiles/desktop-control/notifications/history.jsonl"

    run python3 "${SCRIPTS_DIR}/notification-bridge.py" --append-entry-json '{"app":"qs","summary":"Hello","body":"World","urgency":"normal"}'
    assert_success
    assert_output --partial '"action":"append-entry"'
    assert_output --partial '"ok":true'
    [[ -s "$history" ]]

    run python3 "${SCRIPTS_DIR}/notification-bridge.py" --limit 5
    assert_success
    assert_output --partial '"app":"qs"'
    assert_output --partial '"text":"Hello: World"'

    run python3 "${SCRIPTS_DIR}/notification-bridge.py" --clear-history
    assert_success
    assert_output --partial '"action":"clear-history"'
    [[ ! -s "$history" ]]
}

@test "notification bridge persists dnd state without swaync" {
    export XDG_STATE_HOME="${BATS_TEST_TMPDIR}/state"
    export SWAYNC_CLIENT_BIN="${BATS_TEST_TMPDIR}/missing-swaync-client"

    run python3 "${SCRIPTS_DIR}/notification-bridge.py" --dnd true
    assert_success
    assert_output --partial '"action":"dnd"'
    assert_output --partial '"dnd":true'
    assert_output --partial '"swaync_ok":false'

    run python3 "${SCRIPTS_DIR}/notification-bridge.py" --limit 1
    assert_success
    assert_output --partial '"dnd":true'
}

@test "notification control falls back to swaync when quickshell ipc is unavailable" {
    export QUICKSHELL_BIN="${BATS_TEST_TMPDIR}/quickshell"
    export SWAYNC_CLIENT_BIN="${BATS_TEST_TMPDIR}/swaync-client"
    cat > "$QUICKSHELL_BIN" <<'MOCK'
#!/usr/bin/env bash
exit 1
MOCK
    cat > "$SWAYNC_CLIENT_BIN" <<'MOCK'
#!/usr/bin/env bash
printf 'swaync-client %s\n' "$*" >> "${BATS_TEST_TMPDIR}/notification-control.log"
MOCK
    chmod +x "$QUICKSHELL_BIN" "$SWAYNC_CLIENT_BIN"

    run bash "${SCRIPTS_DIR}/notification-control.sh" toggle-center
    assert_success
    run cat "${BATS_TEST_TMPDIR}/notification-control.log"
    assert_success
    assert_output "swaync-client -t -sw"
}

@test "notification control persists explicit dnd when quickshell ipc is unavailable" {
    export XDG_STATE_HOME="${BATS_TEST_TMPDIR}/state"
    export QUICKSHELL_BIN="${BATS_TEST_TMPDIR}/quickshell"
    export SWAYNC_CLIENT_BIN="${BATS_TEST_TMPDIR}/missing-swaync-client"
    cat > "$QUICKSHELL_BIN" <<'MOCK'
#!/usr/bin/env bash
exit 1
MOCK
    chmod +x "$QUICKSHELL_BIN"

    run bash "${SCRIPTS_DIR}/notification-control.sh" dnd-on
    assert_success
    run python3 "${SCRIPTS_DIR}/notification-bridge.py" --limit 1
    assert_success
    assert_output --partial '"dnd":true'

    run bash "${SCRIPTS_DIR}/notification-control.sh" dnd-off
    assert_success
    run python3 "${SCRIPTS_DIR}/notification-bridge.py" --limit 1
    assert_success
    assert_output --partial '"dnd":false'
}

@test "notification keybinds route through notification control wrapper" {
    run grep -E 'notification-control\.sh (toggle-center|dismiss-all|toggle-dnd)' \
        "${DOTFILES_DIR}/hyprland/hyprland.conf"
    assert_success
    refute_output --partial "swaync-client"
}

@test "notification daemon launcher falls back to swaync when Quickshell does not claim the bus" {
    # Environment: mock systemctl as a no-op and mock busctl to never list
    # the Notifications name. The launcher should hit the wait timeout
    # then exec swaync.
    export NOTIFICATION_OWNER_WAIT_ATTEMPTS=1
    export NOTIFICATION_OWNER_WAIT_SLEEP=0
    export SYSTEMCTL_BIN="${BATS_TEST_TMPDIR}/systemctl"
    export BUSCTL_BIN="${BATS_TEST_TMPDIR}/busctl"
    export GDBUS_BIN="${BATS_TEST_TMPDIR}/missing-gdbus"
    export SWAYNC_BIN="${BATS_TEST_TMPDIR}/swaync"
    cat > "$SYSTEMCTL_BIN" <<'MOCK'
#!/usr/bin/env bash
exit 0
MOCK
    cat > "$BUSCTL_BIN" <<'MOCK'
#!/usr/bin/env bash
# Empty bus list — no notification daemon present
exit 0
MOCK
    cat > "$SWAYNC_BIN" <<'MOCK'
#!/usr/bin/env bash
printf 'swaync fallback\n'
MOCK
    chmod +x "$SYSTEMCTL_BIN" "$BUSCTL_BIN" "$SWAYNC_BIN"

    run bash "${SCRIPTS_DIR}/notification-daemon-launch.sh"
    assert_success
    assert_output "swaync fallback"
}

@test "notification daemon launcher prefers Quickshell when it claims the bus" {
    export SYSTEMCTL_BIN="${BATS_TEST_TMPDIR}/systemctl"
    export BUSCTL_BIN="${BATS_TEST_TMPDIR}/busctl"
    export GDBUS_BIN="${BATS_TEST_TMPDIR}/missing-gdbus"
    export SWAYNC_BIN="${BATS_TEST_TMPDIR}/swaync"
    cat > "$SYSTEMCTL_BIN" <<'MOCK'
#!/usr/bin/env bash
printf 'systemctl %s\n' "$*" >> "${BATS_TEST_TMPDIR}/notification-launch.log"
MOCK
    cat > "$BUSCTL_BIN" <<'MOCK'
#!/usr/bin/env bash
cat <<'EOF'
NAME                              PID PROCESS USER CONNECTION UNIT SESSION DESCRIPTION
org.freedesktop.Notifications   1000 quickshell hg :1.42      -    -       -
EOF
MOCK
    cat > "$SWAYNC_BIN" <<'MOCK'
#!/usr/bin/env bash
printf 'swaync should not start\n' >> "${BATS_TEST_TMPDIR}/notification-launch.log"
MOCK
    chmod +x "$SYSTEMCTL_BIN" "$BUSCTL_BIN" "$SWAYNC_BIN"

    run bash "${SCRIPTS_DIR}/notification-daemon-launch.sh"
    assert_success
    [[ -f "${BATS_TEST_TMPDIR}/notification-launch.log" ]]
    run cat "${BATS_TEST_TMPDIR}/notification-launch.log"
    assert_success
    assert_output "systemctl --user start dotfiles-quickshell.service"
}

@test "notification dbus activation routes through migration launcher" {
    run grep -E '^Exec=/home/hg/hairglasses-studio/dotfiles/scripts/notification-daemon-launch\.sh$' \
        "${DOTFILES_DIR}/etc/dbus-1/services/org.freedesktop.Notifications.service"
    assert_success

    run grep -E '^SystemdService=swaync\.service$' \
        "${DOTFILES_DIR}/etc/dbus-1/services/org.freedesktop.Notifications.service"
    assert_failure
}

@test "shell stack mode rejects retired modes" {
    run bash "${SCRIPTS_DIR}/shell-stack-mode.sh" pilot
    [[ "$status" -eq 2 ]]
    assert_output --partial "Unknown argument: pilot"
}

@test "shell stack mode status emits machine-readable json" {
    run bash "${SCRIPTS_DIR}/shell-stack-mode.sh" --json status
    assert_success
    local status_json="$output"
    run jq -e '.mode == "status" and (.services | length > 0)' <<<"${status_json}"
    assert_success
    run jq -e '.services[] | select(.unit == "dotfiles-quickshell.service")' <<<"${status_json}"
    assert_success
    run jq -e '.shell_mode | type == "string"' <<<"${status_json}"
    assert_success
}

@test "shell stack helpers default to full-cutover and respect persisted mode" {
    export XDG_STATE_HOME="${BATS_TEST_TMPDIR}/state"

    # Empty state — defaults to full-cutover and signals quickshell wanted.
    run bash -lc '
source "$1"
shell_stack_load
shell_stack_quickshell_wanted
[[ "${SHELL_STACK_MODE}" == "full-cutover" ]]
' _ "${SCRIPTS_DIR}/lib/shell-stack.sh"
    assert_success

    # Persisted rollback — still loads, but quickshell is not wanted.
    mkdir -p "${XDG_STATE_HOME}/dotfiles/shell-stack"
    printf 'SHELL_STACK_MODE=rollback\n' > "${XDG_STATE_HOME}/dotfiles/shell-stack/env"
    run bash -lc '
source "$1"
shell_stack_load
[[ "${SHELL_STACK_MODE}" == "rollback" ]] && ! shell_stack_quickshell_wanted
' _ "${SCRIPTS_DIR}/lib/shell-stack.sh"
    assert_success
}

@test "shell stack mode applies full-cutover and persists env" {
    export XDG_STATE_HOME="${BATS_TEST_TMPDIR}/state"
    export PATH="${BATS_TEST_TMPDIR}:${PATH}"
    cat > "${BATS_TEST_TMPDIR}/systemctl" <<'MOCK'
#!/usr/bin/env bash
echo "systemctl $*" >> "${BATS_TEST_TMPDIR}/systemctl.log"
case "$*" in
  "--user is-active "*) echo inactive; exit 3 ;;
esac
exit 0
MOCK
    chmod +x "${BATS_TEST_TMPDIR}/systemctl"

    run bash "${SCRIPTS_DIR}/shell-stack-mode.sh" full-cutover
    assert_success
    assert_output --partial "+ systemctl --user restart dotfiles-quickshell.service"

    local env_file="${XDG_STATE_HOME}/dotfiles/shell-stack/env"
    [[ -f "$env_file" ]]
    run cat "$env_file"
    assert_success
    assert_output --partial "SHELL_STACK_MODE=full-cutover"
    refute_output --partial "QS_BAR_CUTOVER"
    refute_output --partial "QS_TICKER_CUTOVER"
    refute_output --partial "QS_MENU_CUTOVER"
    refute_output --partial "QS_DOCK_CUTOVER"
    refute_output --partial "QS_COMPANION_CUTOVER"
    refute_output --partial "QUICKSHELL_NOTIFICATION_OWNER"
}

@test "shell stack mode rollback stops quickshell" {
    export XDG_STATE_HOME="${BATS_TEST_TMPDIR}/state"
    export PATH="${BATS_TEST_TMPDIR}:${PATH}"
    cat > "${BATS_TEST_TMPDIR}/systemctl" <<'MOCK'
#!/usr/bin/env bash
exit 0
MOCK
    chmod +x "${BATS_TEST_TMPDIR}/systemctl"

    run bash "${SCRIPTS_DIR}/shell-stack-mode.sh" rollback
    assert_success
    assert_output --partial "+ systemctl --user stop dotfiles-quickshell.service"
}

@test "shell stack boot defaults to full-cutover when no mode is persisted" {
    export XDG_STATE_HOME="${BATS_TEST_TMPDIR}/state"
    export PATH="${BATS_TEST_TMPDIR}:${PATH}"
    cat > "${BATS_TEST_TMPDIR}/systemctl" <<'MOCK'
#!/usr/bin/env bash
exit 0
MOCK
    chmod +x "${BATS_TEST_TMPDIR}/systemctl"

    run bash "${SCRIPTS_DIR}/shell-stack-boot.sh"
    assert_success
    assert_output --partial "+ systemctl --user restart dotfiles-quickshell.service"
}

@test "hyprland startup routes shell owners through shell stack boot" {
    run grep -F 'shell-stack-boot.sh' "${DOTFILES_DIR}/hyprland/hyprland.conf"
    assert_success

    run grep -E 'exec-once = (swaync|systemctl --user start --no-block (ironbar\.service|dotfiles-quickshell\.service|dotfiles-keybind-ticker\.service))$' \
        "${DOTFILES_DIR}/hyprland/hyprland.conf"
    assert_failure
}

@test "hg shell module routes to the simplified mode set" {
    run env DOTFILES_DIR="${DOTFILES_DIR}" HG_STUDIO_ROOT= bash "${SCRIPTS_DIR}/hg" shell status
    assert_success
    assert_output --partial "shell-stack-mode"

    run env DOTFILES_DIR="${DOTFILES_DIR}" HG_STUDIO_ROOT= bash "${SCRIPTS_DIR}/hg" shell pilot
    [[ "$status" -ne 0 ]]
}
