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

@test "install.sh --check fails when writable config dir is a symlink" {
    rm -rf "${HOME}/.config/hyprshell"
    mkdir -p "${HOME}/.config"
    mkdir -p "${HOME}/.config/hyprshell-target"
    ln -s "${HOME}/.config/hyprshell-target" "${HOME}/.config/hyprshell"

    run env HOME="${HOME}" bash "${DOTFILES_DIR}/install.sh" --check
    assert_failure
    assert_output --partial "Expected writable dir, found symlink: ${HOME}/.config/hyprshell"
}

@test "install.sh --check fails when writable config dir is not writable by current user" {
    if [[ "$(id -u)" -eq 0 ]]; then
        skip "root bypasses directory write-bit checks"
    fi

    chmod 0555 "${HOME}/.config/hyprshell"

    run env HOME="${HOME}" bash "${DOTFILES_DIR}/install.sh" --check
    assert_failure
    assert_output --partial "Writable dir not writable by current user (hyprshell writable config dir): ${HOME}/.config/hyprshell"
    assert_output --partial "owner="
    assert_output --partial "mode="
}
