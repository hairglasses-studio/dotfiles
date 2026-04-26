#!/usr/bin/env bats
# Tests for the stable launcher/switcher entrypoints. Both scripts are
# now thin delegators to scripts/menu-control.sh; the previous
# wofi/hyprshell-overview test surface is gone.

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export TEST_BIN="${BATS_TEST_TMPDIR}/bin"
    export PATH="${TEST_BIN}:${PATH}"
    mkdir -p "${TEST_BIN}"
    export MENU_LOG="${BATS_TEST_TMPDIR}/menu-control.log"

    # Stub out menu-control.sh so the entrypoints can be exercised
    # without spinning up Quickshell IPC or wofi.
    cat > "${BATS_TEST_TMPDIR}/menu-control.sh" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "${MENU_LOG}"
exit 0
EOF
    chmod +x "${BATS_TEST_TMPDIR}/menu-control.sh"

    # Drop a sibling app-launcher.sh / app-switcher.sh into the tmp dir
    # that points at the stub instead of the real menu-control.sh.
    cp "${SCRIPTS_DIR}/app-launcher.sh" "${BATS_TEST_TMPDIR}/app-launcher.sh"
    cp "${SCRIPTS_DIR}/app-switcher.sh" "${BATS_TEST_TMPDIR}/app-switcher.sh"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

@test "app-launcher delegates to menu-control apps" {
    run bash "${BATS_TEST_TMPDIR}/app-launcher.sh"
    assert_success
    run cat "${MENU_LOG}"
    assert_success
    assert_output "apps"
}

@test "app-switcher delegates to menu-control windows by default" {
    run bash "${BATS_TEST_TMPDIR}/app-switcher.sh"
    assert_success
    run cat "${MENU_LOG}"
    assert_success
    assert_output "windows"
}

@test "app-switcher reverse arg still goes to windows" {
    run bash "${BATS_TEST_TMPDIR}/app-switcher.sh" reverse
    assert_success
    run cat "${MENU_LOG}"
    assert_success
    assert_output "windows"
}

@test "app-switcher close arg dispatches menu-control close" {
    run bash "${BATS_TEST_TMPDIR}/app-switcher.sh" close
    assert_success
    run cat "${MENU_LOG}"
    assert_success
    assert_output "close"
}

@test "app-switcher rejects unknown args with exit code 2" {
    run bash "${BATS_TEST_TMPDIR}/app-switcher.sh" definitely-not-a-real-arg
    [[ "$status" -eq 2 ]]
    assert_output --partial "Usage:"
}

@test "install.sh print-link-specs includes managed launcher shims" {
    run bash "${DOTFILES_DIR}/install.sh" --print-link-specs
    assert_success
    assert_output --partial "${DOTFILES_DIR}/scripts/app-launcher.sh|${HOME}/.local/bin/app-launcher"
    assert_output --partial "${DOTFILES_DIR}/scripts/app-switcher.sh|${HOME}/.local/bin/app-switcher"
}
