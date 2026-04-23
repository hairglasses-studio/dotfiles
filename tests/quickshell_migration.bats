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

@test "shell stack mode defaults to dry-run for cutover commands" {
    run bash "${SCRIPTS_DIR}/shell-stack-mode.sh" full-pilot
    assert_success
    assert_output --partial "[dry] systemctl --user start dotfiles-quickshell.service"
    assert_output --partial "[dry] systemctl --user stop ironbar.service"
    assert_output --partial "[dry] systemctl --user stop dotfiles-keybind-ticker.service"
}

@test "shell stack mode status has machine-readable json" {
    run bash "${SCRIPTS_DIR}/shell-stack-mode.sh" --json status
    assert_success
    run jq -e '.mode == "status" and (.services[] | select(.unit == "dotfiles-quickshell.service"))' <<<"${output}"
    assert_success
}

@test "hg shell module routes to shell stack dry-run controls" {
    run env DOTFILES_DIR="${DOTFILES_DIR}" HG_STUDIO_ROOT= bash "${SCRIPTS_DIR}/hg" shell full-pilot
    assert_success
    assert_output --partial "[dry] systemctl --user start dotfiles-quickshell.service"
    assert_output --partial "[dry] systemctl --user stop ironbar.service"
}
