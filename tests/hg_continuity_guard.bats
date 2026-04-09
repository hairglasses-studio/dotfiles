#!/usr/bin/env bats

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export TEST_BIN="${BATS_TEST_TMPDIR}/bin"
    export HOME="${BATS_TEST_TMPDIR}/home"
    export PATH="${TEST_BIN}:${PATH}"
    export SYSTEMCTL_LOG="${BATS_TEST_TMPDIR}/systemctl.log"
    mkdir -p "${TEST_BIN}" "${HOME}"

    cat > "${HOME}/.tmux.conf" <<'EOF'
set -g @plugin 'tmux-plugins/tpm'
set -g @plugin 'tmux-plugins/tmux-resurrect'
set -g @plugin 'tmux-plugins/tmux-continuum'
EOF

    cat > "${TEST_BIN}/systemctl" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "${SYSTEMCTL_LOG}"
exit 0
EOF
    chmod +x "${TEST_BIN}/systemctl"

    cat > "${TEST_BIN}/tmux" <<'EOF'
#!/usr/bin/env bash
if [[ "${1:-}" == "has-session" ]]; then
    exit "${TMUX_MAIN_SESSION_EXIT:-1}"
fi
exit 0
EOF
    chmod +x "${TEST_BIN}/tmux"

    cat > "${TEST_BIN}/notify-send" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
    chmod +x "${TEST_BIN}/notify-send"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

_run_hg_module() {
    local module="$1"
    shift
    env \
        HOME="${HOME}" \
        PATH="${PATH}" \
        DOTFILES_DIR="${DOTFILES_DIR}" \
        HG_DOTFILES="${DOTFILES_DIR}" \
        bash -lc '
            source "'"${DOTFILES_DIR}"'/scripts/lib/hg-core.sh"
            source "'"${DOTFILES_DIR}"'/scripts/hg-modules/mod-'"${module}"'.sh"
            '"${module}"'_run "$@"
        ' bash "$@"
}

@test "hg config lane classifies hyprshell reload and restart lanes" {
    run _run_hg_module config lane hyprshell
    assert_success
    assert_line --index 0 "hyprshell	reload	safe_reload"
    assert_line --index 1 "hyprshell	restart	explicit_restart"
}

@test "hg config restart blocks explicit restart when tmux persistence health fails" {
    run _run_hg_module config restart hyprshell
    assert_failure
    assert_output --partial "Blocked explicit restart for hyprshell"

    run test -f "${SYSTEMCTL_LOG}"
    assert_failure
}

@test "hg config restart --force bypasses failed tmux persistence health" {
    run _run_hg_module config restart --force hyprshell
    assert_success
    assert_output --partial "Restarted hyprshell"

    run cat "${SYSTEMCTL_LOG}"
    assert_success
    assert_output --partial "--user restart dotfiles-hyprshell.service"
}

@test "hg rice restart-ui blocks when tmux persistence health fails" {
    run _run_hg_module rice restart-ui
    assert_failure
    assert_output --partial "Blocked UI restart"
}

@test "hg rice restart-ui --force restarts service-backed companions" {
    run _run_hg_module rice restart-ui --force
    assert_success
    assert_output --partial "UI companion services restarted"

    run cat "${SYSTEMCTL_LOG}"
    assert_success
    assert_output --partial "--user restart dotfiles-hyprshell.service"
    assert_output --partial "--user restart dotfiles-hypr-dock.service"
    assert_output --partial "--user restart dotfiles-hyprdynamicmonitors.service"
    assert_output --partial "--user restart dotfiles-hyprland-autoname-workspaces.service"
}
