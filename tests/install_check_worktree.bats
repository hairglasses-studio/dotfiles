#!/usr/bin/env bats

load 'test_helper'

setup() {
    BATS_TEST_TMPDIR="$(mktemp -d)"
    export BATS_TEST_TMPDIR
    export HOME="${BATS_TEST_TMPDIR}/home"
    export PATH="${PATH}"
    mkdir -p "${HOME}/.local/state/hypr" "${HOME}/.local/state/kitty/sessions"
    touch "${HOME}/.local/state/hypr/monitors.dynamic.conf"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

@test "install.sh --check reports generated runtime state without retired writable roots" {
    run env HOME="${HOME}" bash "${DOTFILES_DIR}/install.sh" --check
    assert_failure
    assert_output --partial "Checking writable config roots..."
    assert_output --partial "OK (generated Hyprland monitor include): ${HOME}/.local/state/hypr/monitors.dynamic.conf"
    assert_output --partial "OK (kitty session state directory): ${HOME}/.local/state/kitty/sessions"
    refute_output --partial "hyprshell writable config dir"
}

@test "install.sh --check warns when required generated runtime file is missing" {
    rm -f "${HOME}/.local/state/hypr/monitors.dynamic.conf"

    run env HOME="${HOME}" bash "${DOTFILES_DIR}/install.sh" --check
    assert_failure
    assert_output --partial "Missing generated runtime file: ${HOME}/.local/state/hypr/monitors.dynamic.conf"
}

@test "install.sh --check warns when required generated runtime dir is missing" {
    rm -rf "${HOME}/.local/state/kitty/sessions"

    run env HOME="${HOME}" bash "${DOTFILES_DIR}/install.sh" --check
    assert_failure
    assert_output --partial "Missing generated runtime dir: ${HOME}/.local/state/kitty/sessions"
}
