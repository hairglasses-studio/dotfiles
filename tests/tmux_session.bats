#!/usr/bin/env bats

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export TEST_BIN="${BATS_TEST_TMPDIR}/bin"
    export PATH="${TEST_BIN}:${PATH}"
    export TMUX_LOG="${BATS_TEST_TMPDIR}/tmux.log"
    mkdir -p "${TEST_BIN}"

    cat > "${TEST_BIN}/tmux" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "${TMUX_LOG}"
if [[ "${1:-}" == "has-session" ]]; then
    exit "${TMUX_HAS_SESSION_EXIT:-1}"
fi
exit 0
EOF
    chmod +x "${TEST_BIN}/tmux"

    export REAL_SCRIPTS_DIR="$(cd "${SCRIPTS_DIR}" && pwd)"
    export REAL_DOTFILES_DIR="$(cd "${DOTFILES_DIR}" && pwd)"
    export REAL_STUDIO_DIR="$(cd "${DOTFILES_DIR}/.." && pwd)"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

@test "kitty-dev-launch routes terminals through the persistent tmux main session" {
    cat > "${TEST_BIN}/kitty-launcher" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "${BATS_TEST_TMPDIR}/launcher.log"
EOF
    chmod +x "${TEST_BIN}/kitty-launcher"

    run env KITTY_DEV_LAUNCHER="${TEST_BIN}/kitty-launcher" bash "${SCRIPTS_DIR}/kitty-dev-launch.sh" --class=dev-terminal
    assert_success

    run cat "${BATS_TEST_TMPDIR}/launcher.log"
    assert_success
    assert_output --partial "--class=dev-terminal"
    assert_output --partial "-e ${REAL_SCRIPTS_DIR}/tmux-main-session.sh"
}

@test "tmux-main-session attaches to the persistent main session when it already exists" {
    export TMUX_HAS_SESSION_EXIT=0

    run bash "${SCRIPTS_DIR}/tmux-main-session.sh"
    assert_success

    run cat "${TMUX_LOG}"
    assert_success
    assert_line --index 0 "has-session -t main"
    assert_line --index 1 "attach-session -t main"
}

@test "tmux-main-session creates the persistent main session in the studio root when missing" {
    export TMUX_HAS_SESSION_EXIT=1

    run bash "${SCRIPTS_DIR}/tmux-main-session.sh"
    assert_success

    run cat "${TMUX_LOG}"
    assert_success
    assert_line --index 0 "has-session -t main"
    assert_output --partial "new-session -s main -c ${REAL_STUDIO_DIR}"
}

@test "dropdown-session bootstraps the session once and reattaches without killing it" {
    export TMUX_HAS_SESSION_EXIT=1

    run bash "${SCRIPTS_DIR}/dropdown-session.sh"
    assert_success

    run cat "${TMUX_LOG}"
    assert_success
    assert_line --index 0 "has-session -t dropdown"
    assert_output --partial "new-session -d -s dropdown -c ${REAL_STUDIO_DIR}"
    assert_output --partial "HG_AGENT_SESSION_QUIET=1 ralphglasses --scan-path ${REAL_STUDIO_DIR}"
    assert_output --partial "split-window -t dropdown -h -c ${DOTFILES_DIR} claude"
    assert_output --partial "select-pane -t dropdown:0.0"
    assert_line --index 4 "attach-session -t dropdown"
    refute_output --partial "kill-session"
}

@test "dropdown-session reattaches to the existing session without re-bootstrap" {
    export TMUX_HAS_SESSION_EXIT=0

    run bash "${SCRIPTS_DIR}/dropdown-session.sh"
    assert_success

    run cat "${TMUX_LOG}"
    assert_success
    assert_line --index 0 "has-session -t dropdown"
    assert_line --index 1 "attach-session -t dropdown"
}

@test "dropdown-terminal launches the persistent dropdown session helper without killing tmux state" {
    export HYPRLAND_INSTANCE_SIGNATURE="test-session"

    cat > "${TEST_BIN}/kitty-launcher" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "${BATS_TEST_TMPDIR}/launcher.log"
EOF
    chmod +x "${TEST_BIN}/kitty-launcher"

    cat > "${TEST_BIN}/hyprctl" <<'EOF'
#!/usr/bin/env bash
case "${1:-} ${2:-}" in
    "clients -j")
        printf '[]\n'
        ;;
    *)
        printf '%s\n' "$*" >> "${BATS_TEST_TMPDIR}/hyprctl.log"
        ;;
esac
EOF
    chmod +x "${TEST_BIN}/hyprctl"

    run env KITTY_DROPDOWN_LAUNCHER="${TEST_BIN}/kitty-launcher" bash "${SCRIPTS_DIR}/dropdown-terminal.sh"
    assert_success

    run cat "${BATS_TEST_TMPDIR}/launcher.log"
    assert_success
    assert_output --partial "--class=dropdown-cyber"
    assert_output --partial "-e ${REAL_SCRIPTS_DIR}/dropdown-session.sh"
    refute_output --partial "kill-session"

    run cat "${BATS_TEST_TMPDIR}/hyprctl.log"
    assert_success
    assert_output --partial "dispatch movetoworkspacesilent special:dropdown,class:^(dropdown-cyber)$"
    assert_output --partial "dispatch togglespecialworkspace dropdown"
}
