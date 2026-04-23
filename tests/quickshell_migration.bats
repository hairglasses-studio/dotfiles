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

@test "notification daemon launcher falls back to swaync by default" {
    export XDG_STATE_HOME="${BATS_TEST_TMPDIR}/state"
    export SWAYNC_BIN="${BATS_TEST_TMPDIR}/swaync"
    cat > "$SWAYNC_BIN" <<'MOCK'
#!/usr/bin/env bash
printf 'swaync fallback\n'
MOCK
    chmod +x "$SWAYNC_BIN"

    run bash "${SCRIPTS_DIR}/notification-daemon-launch.sh"
    assert_success
    assert_output "swaync fallback"
}

@test "notification daemon launcher starts quickshell owner when cutover is persisted" {
    export XDG_STATE_HOME="${BATS_TEST_TMPDIR}/state"
    export SYSTEMCTL_BIN="${BATS_TEST_TMPDIR}/systemctl"
    export BUSCTL_BIN="${BATS_TEST_TMPDIR}/busctl"
    export GDBUS_BIN="${BATS_TEST_TMPDIR}/missing-gdbus"
    export SWAYNC_BIN="${BATS_TEST_TMPDIR}/swaync"
    mkdir -p "${XDG_STATE_HOME}/dotfiles/shell-stack"
    printf 'QUICKSHELL_NOTIFICATION_OWNER=1\n' > "${XDG_STATE_HOME}/dotfiles/shell-stack/env"
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

@test "shell stack mode defaults to dry-run for cutover commands" {
    run bash "${SCRIPTS_DIR}/shell-stack-mode.sh" full-pilot
    assert_success
    assert_output --partial "[dry] write"
    assert_output --partial "[dry] systemctl --user restart dotfiles-quickshell.service"
    assert_output --partial "[dry] systemctl --user stop ironbar.service"
    assert_output --partial "[dry] systemctl --user stop dotfiles-keybind-ticker.service"
}

@test "shell stack mode status has machine-readable json" {
    run bash "${SCRIPTS_DIR}/shell-stack-mode.sh" --json status
    assert_success
    local status_json="$output"
    run jq -e '.mode == "status" and (.services[] | select(.unit == "dotfiles-quickshell.service"))' <<<"${status_json}"
    assert_success
    run jq -e '.shell_mode and (.notification_owner | type == "boolean")' <<<"${status_json}"
    assert_success
}

@test "shell stack helpers read persisted cutover env" {
    export XDG_STATE_HOME="${BATS_TEST_TMPDIR}/state"
    mkdir -p "${XDG_STATE_HOME}/dotfiles/shell-stack"
    cat > "${XDG_STATE_HOME}/dotfiles/shell-stack/env" <<'ENV'
SHELL_STACK_MODE=full-cutover
QS_BAR_CUTOVER=1
QS_TICKER_CUTOVER=1
QUICKSHELL_NOTIFICATION_OWNER=1
ENV

    run bash -lc '
source "$1"
shell_stack_load
shell_stack_bar_cutover
shell_stack_ticker_cutover
shell_stack_notification_cutover
shell_stack_quickshell_wanted
' _ "${SCRIPTS_DIR}/lib/shell-stack.sh"
    assert_success
}

@test "shell stack boot applies selected mode in dry-run" {
    run bash "${SCRIPTS_DIR}/shell-stack-boot.sh" --dry-run --mode full-cutover
    assert_success
    assert_output --partial "[dry] systemctl --user stop swaync.service"
    assert_output --partial "[dry] systemctl --user restart dotfiles-quickshell.service"
    assert_output --partial "[dry] systemctl --user stop ironbar.service"
    assert_output --partial "[dry] systemctl --user stop dotfiles-keybind-ticker.service"
}

@test "hyprland startup routes shell owners through shell stack boot" {
    run grep -F 'shell-stack-boot.sh' "${DOTFILES_DIR}/hyprland/hyprland.conf"
    assert_success

    run grep -E 'exec-once = (swaync|systemctl --user start --no-block (ironbar\.service|dotfiles-quickshell\.service|dotfiles-keybind-ticker\.service))$' \
        "${DOTFILES_DIR}/hyprland/hyprland.conf"
    assert_failure
}

@test "shell stack apply persists full cutover environment" {
    export XDG_STATE_HOME="${BATS_TEST_TMPDIR}/state"
    export PATH="${BATS_TEST_TMPDIR}:${PATH}"
    cat > "${BATS_TEST_TMPDIR}/systemctl" <<'MOCK'
#!/usr/bin/env bash
echo "systemctl $*" >> "${BATS_TEST_TMPDIR}/systemctl.log"
case "$*" in
  "--user list-unit-files swaync.service") exit 0 ;;
  "--user is-active "*) echo inactive; exit 3 ;;
esac
exit 0
MOCK
    chmod +x "${BATS_TEST_TMPDIR}/systemctl"

    run bash "${SCRIPTS_DIR}/shell-stack-mode.sh" --apply full-cutover
    assert_success
    assert_output --partial "+ systemctl --user stop swaync.service"
    assert_output --partial "+ systemctl --user restart dotfiles-quickshell.service"
    assert_output --partial "+ systemctl --user stop ironbar.service"
    assert_output --partial "+ systemctl --user stop dotfiles-keybind-ticker.service"

    local env_file="${XDG_STATE_HOME}/dotfiles/shell-stack/env"
    [[ -f "$env_file" ]]
    run grep -E '^(SHELL_STACK_MODE=full-cutover|QS_BAR_CUTOVER=1|QS_TICKER_CUTOVER=1|QUICKSHELL_NOTIFICATION_OWNER=1)$' "$env_file"
    assert_success
}

@test "hg shell module routes to shell stack dry-run controls" {
    run env DOTFILES_DIR="${DOTFILES_DIR}" HG_STUDIO_ROOT= bash "${SCRIPTS_DIR}/hg" shell full-pilot
    assert_success
    assert_output --partial "[dry] systemctl --user restart dotfiles-quickshell.service"
    assert_output --partial "[dry] systemctl --user stop ironbar.service"
}

@test "hg shell module exposes notification and full cutover modes" {
    run env DOTFILES_DIR="${DOTFILES_DIR}" HG_STUDIO_ROOT= bash "${SCRIPTS_DIR}/hg" shell notification-cutover
    assert_success
    assert_output --partial "[dry] systemctl --user stop swaync.service"
    assert_output --partial "[dry] systemctl --user restart dotfiles-quickshell.service"

    run env DOTFILES_DIR="${DOTFILES_DIR}" HG_STUDIO_ROOT= bash "${SCRIPTS_DIR}/hg" shell full-cutover
    assert_success
    assert_output --partial "[dry] systemctl --user restart dotfiles-quickshell.service"
    assert_output --partial "[dry] systemctl --user stop ironbar.service"
    assert_output --partial "[dry] systemctl --user stop dotfiles-keybind-ticker.service"
    assert_output --partial "[dry] systemctl --user stop swaync.service"
}
