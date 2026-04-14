#!/usr/bin/env bats

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export HOME="${BATS_TEST_TMPDIR}/home"
    export PATH="${PATH}"
    mkdir -p "${HOME}/.config/hyprshell" "${HOME}/.local/state/hypr" "${HOME}/.local/state/kitty/sessions"
    touch "${HOME}/.local/state/hypr/monitors.dynamic.conf"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

@test "install.sh --check classifies writable and generated runtime state separately" {
    run env HOME="${HOME}" bash "${DOTFILES_DIR}/install.sh" --check
    assert_failure
    assert_output --partial "OK (hyprshell writable config dir): ${HOME}/.config/hyprshell"
    assert_output --partial "OK (generated Hyprland monitor include): ${HOME}/.local/state/hypr/monitors.dynamic.conf"
    assert_output --partial "OK (kitty session state directory): ${HOME}/.local/state/kitty/sessions"
}
