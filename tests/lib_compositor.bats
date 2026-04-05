#!/usr/bin/env bats
# Tests for scripts/lib/compositor.sh

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    mock_compositor
    # Clear compositor env vars for clean detection tests
    unset HYPRLAND_INSTANCE_SIGNATURE
    source "${LIB_DIR}/compositor.sh"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

# --- Detection tests ---

@test "compositor_type: returns hyprland when HYPRLAND_INSTANCE_SIGNATURE is set" {
    export HYPRLAND_INSTANCE_SIGNATURE="test_instance"
    run compositor_type
    assert_success
    assert_output "hyprland"
}

@test "compositor_type: returns unknown when no compositor env vars set (Linux)" {
    unset HYPRLAND_INSTANCE_SIGNATURE
    run compositor_type
    assert_success
    assert_output "unknown"
}

# --- IPC tests (with mocked compositor) ---

@test "compositor_msg: calls hyprctl dispatch when Hyprland detected" {
    export HYPRLAND_INSTANCE_SIGNATURE="test_instance"
    run compositor_msg "workspace 3"
    assert_success
    # Verify the mock was called
    run cat "${BATS_TEST_TMPDIR}/compositor.log"
    assert_output --partial "MOCK_hyprctl"
}

@test "compositor_reload: calls hyprctl reload when Hyprland detected" {
    export HYPRLAND_INSTANCE_SIGNATURE="test_instance"
    run compositor_reload
    assert_success
    run cat "${BATS_TEST_TMPDIR}/compositor.log"
    assert_output --partial "MOCK_hyprctl"
}

@test "compositor_workspace: dispatches workspace switch" {
    export HYPRLAND_INSTANCE_SIGNATURE="test_instance"
    run compositor_workspace 5
    assert_success
    run cat "${BATS_TEST_TMPDIR}/compositor.log"
    assert_output --partial "MOCK_hyprctl"
}

@test "compositor_query: returns output for valid query type" {
    export HYPRLAND_INSTANCE_SIGNATURE="test_instance"
    run compositor_query workspaces
    assert_success
}

@test "compositor_msg: produces no output for unknown compositor" {
    unset HYPRLAND_INSTANCE_SIGNATURE
    run compositor_msg "test"
    assert_success
    assert_output ""
}

# --- hypr_socket2 tests ---

@test "hypr_socket2: returns path containing instance signature" {
    export HYPRLAND_INSTANCE_SIGNATURE="test_sig_123"
    run hypr_socket2
    assert_success
    assert_output --partial "test_sig_123"
    assert_output --partial ".socket2.sock"
}
