#!/usr/bin/env bats

load 'test_helper'

host_smoke_script() {
    printf '%s\n' "${DOTFILES_DIR}/mcp/dotfiles-mcp/scripts/host-smoke.sh"
}

write_stub() {
    local name="$1"
    local body="$2"
    cat > "${BATS_TEST_TMPDIR}/bin/${name}" <<EOF
#!/usr/bin/env bash
${body}
EOF
    chmod +x "${BATS_TEST_TMPDIR}/bin/${name}"
}

setup_host_smoke_stubs() {
    mkdir -p "${BATS_TEST_TMPDIR}/bin" "${BATS_TEST_TMPDIR}/dotfiles/juhradial"
    printf '{}\n' > "${BATS_TEST_TMPDIR}/dotfiles/juhradial/config.json"

    write_stub "python3" '
if [[ "${1:-}" == "--version" ]]; then
  echo "Python 3.12.0"
  exit 0
fi
if [[ "${1:-}" == "-c" && "${2:-}" == *"import pyatspi"* ]]; then
  echo "pyatspi import ok"
  exit 0
fi
echo "python3 stub"
'

    for cmd in hyprctl ydotool wtype bluetoothctl busctl gh git dbus-run-session wayland-info grim wl-copy wl-paste kwin_wayland; do
        write_stub "${cmd}" "echo '${cmd} stub'"
    done
}

@test "dotfiles-mcp host smoke reports semantic and session checks in json mode" {
    setup_host_smoke_stubs

    run env \
        PATH="${BATS_TEST_TMPDIR}/bin:/usr/bin:/bin" \
        HYPRLAND_INSTANCE_SIGNATURE="hypr-test" \
        WAYLAND_DISPLAY="wayland-test" \
        DBUS_SESSION_BUS_ADDRESS="unix:path=/tmp/test-bus" \
        DOTFILES_DIR="${BATS_TEST_TMPDIR}/dotfiles" \
        bash "$(host_smoke_script)" --json

    assert_success
    json_path="${BATS_TEST_TMPDIR}/host-smoke.json"
    printf '%s\n' "$output" > "$json_path"

    run jq -r '
      .strict_skip,
      (.checks[] | select(.group=="semantic" and .name=="pyatspi") | .status),
      (.checks[] | select(.group=="session" and .name=="wayland-info") | .status),
      (.checks[] | select(.group=="session" and .name=="wayland display") | .detail)
    ' "$json_path"
    assert_success
    assert_line --index 0 "false"
    assert_line --index 1 "ok"
    assert_line --index 2 "ok"
    assert_line --index 3 "WAYLAND_DISPLAY=wayland-test"
}

@test "dotfiles-mcp host smoke can fail on skips when strict-skip is enabled" {
    setup_host_smoke_stubs

    run env \
        PATH="${BATS_TEST_TMPDIR}/bin:/usr/bin:/bin" \
        bash "$(host_smoke_script)" --json --strict-skip

    assert_failure
    json_path="${BATS_TEST_TMPDIR}/host-smoke-strict.json"
    printf '%s\n' "$output" > "$json_path"

    run jq -r '.strict_skip, .status' "$json_path"
    assert_success
    assert_line --index 0 "true"
    assert_line --index 1 "fail"

    skip_count="$(jq '[.checks[] | select(.status=="skip")] | length' "$json_path")"
    [ "$skip_count" -gt 0 ]
}
