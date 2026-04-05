#!/usr/bin/env bats
# Tests for scripts/lib/notify.sh

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    mock_notify_send
    source "${LIB_DIR}/notify.sh"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

# --- Normal notification ---

@test "hg_notify: sends notification via notify-send" {
    run hg_notify "Shader" "Switched to bloom-soft"
    assert_success
    run cat "${BATS_TEST_TMPDIR}/notifications.log"
    assert_output --partial "Shader"
    assert_output --partial "Switched to bloom-soft"
}

# --- Low urgency ---

@test "hg_notify_low: sends notification with -u low flag" {
    run hg_notify_low "Wallpaper" "cyber-rain"
    assert_success
    run cat "${BATS_TEST_TMPDIR}/notifications.log"
    assert_output --partial "-u low"
    assert_output --partial "Wallpaper"
    assert_output --partial "cyber-rain"
}

# --- Critical urgency ---

@test "hg_notify_critical: sends notification with -u critical flag" {
    run hg_notify_critical "Battery" "MX Master 4 low: 15%"
    assert_success
    run cat "${BATS_TEST_TMPDIR}/notifications.log"
    assert_output --partial "-u critical"
    assert_output --partial "Battery"
}

# --- App name passed via -a flag ---

@test "hg_notify: passes app name via -a flag" {
    run hg_notify "TestApp" "test body"
    assert_success
    run cat "${BATS_TEST_TMPDIR}/notifications.log"
    assert_output --partial "-a TestApp"
}

# --- Graceful degradation ---

@test "hg_notify: returns 0 when notify-send is not available" {
    # Remove mock notify-send from PATH
    rm -f "${BATS_TEST_TMPDIR}/notify-send"
    # Use an isolated PATH without notify-send
    export PATH="/usr/bin:/bin"
    # Re-source to pick up the new PATH
    source "${LIB_DIR}/notify.sh"
    run hg_notify "Test" "Should not crash"
    assert_success
}

@test "hg_notify_low: returns 0 when notify-send is not available" {
    rm -f "${BATS_TEST_TMPDIR}/notify-send"
    export PATH="/usr/bin:/bin"
    source "${LIB_DIR}/notify.sh"
    run hg_notify_low "Test" "Should not crash"
    assert_success
}

@test "hg_notify_critical: returns 0 when notify-send is not available" {
    rm -f "${BATS_TEST_TMPDIR}/notify-send"
    export PATH="/usr/bin:/bin"
    source "${LIB_DIR}/notify.sh"
    run hg_notify_critical "Test" "Should not crash"
    assert_success
}
