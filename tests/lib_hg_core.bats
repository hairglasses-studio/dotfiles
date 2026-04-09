#!/usr/bin/env bats
# Tests for scripts/lib/hg-core.sh

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    # Prevent hg-core from creating state dir in real home
    export HG_STATE_DIR="${BATS_TEST_TMPDIR}/state"
    source "${LIB_DIR}/hg-core.sh"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

# --- Color variable tests ---

@test "hg-core: Snazzy color variables are defined" {
    [[ -n "${HG_RED}" ]]
    [[ -n "${HG_GREEN}" ]]
    [[ -n "${HG_CYAN}" ]]
    [[ -n "${HG_YELLOW}" ]]
    [[ -n "${HG_MAGENTA}" ]]
    [[ -n "${HG_RESET}" ]]
}

@test "hg-core: HG_RED contains 24-bit ANSI code for #ff5c57" {
    [[ "${HG_RED}" == *"38;2;255;92;87"* ]]
}

@test "hg-core: HG_CYAN contains 24-bit ANSI code for #57c7ff" {
    [[ "${HG_CYAN}" == *"38;2;87;199;255"* ]]
}

@test "hg-core: HG_GREEN contains 24-bit ANSI code for #5af78e" {
    [[ "${HG_GREEN}" == *"38;2;90;247;142"* ]]
}

@test "hg-core: HG_RESET is ANSI reset sequence" {
    [[ "${HG_RESET}" == *"[0m"* ]]
}

@test "hg-core: HG_DIM is defined" {
    [[ -n "${HG_DIM}" ]]
}

@test "hg-core: HG_BOLD is defined" {
    [[ -n "${HG_BOLD}" ]]
}

# --- Output function tests ---

@test "hg_info: prints message with [info] tag to stdout" {
    run hg_info "test message"
    assert_success
    assert_output --partial "[info]"
    assert_output --partial "test message"
}

@test "hg_ok: prints message with [ok] tag to stdout" {
    run hg_ok "operation complete"
    assert_success
    assert_output --partial "[ok]"
    assert_output --partial "operation complete"
}

@test "hg_warn: prints message with [warn] tag" {
    run hg_warn "be careful"
    assert_success
    assert_output --partial "[warn]"
    assert_output --partial "be careful"
}

@test "hg_error: prints message with [err] tag to stderr" {
    run bash -c "source '${LIB_DIR}/hg-core.sh'; hg_error 'something failed' 2>&1"
    assert_output --partial "[err]"
    assert_output --partial "something failed"
}

@test "hg_die: exits with status 1 by default" {
    run bash -c "source '${LIB_DIR}/hg-core.sh'; hg_die 'fatal error' 2>&1"
    assert_failure 1
    assert_output --partial "fatal error"
}

@test "hg_die: exits with custom exit code when provided" {
    run bash -c "source '${LIB_DIR}/hg-core.sh'; hg_die 'custom code' 42 2>&1"
    assert_failure 42
}

# --- Dependency check tests ---

@test "hg_require: succeeds for existing command (bash)" {
    run hg_require bash
    assert_success
}

@test "hg_require: succeeds for multiple existing commands" {
    run hg_require bash cat ls
    assert_success
}

@test "hg_require: fails for nonexistent command" {
    run bash -c "source '${LIB_DIR}/hg-core.sh'; hg_require nonexistent_command_xyz_12345 2>&1"
    assert_failure
}

@test "hg_require: error message includes missing command name" {
    run bash -c "source '${LIB_DIR}/hg-core.sh'; hg_require nonexistent_command_xyz_12345 2>&1"
    assert_output --partial "nonexistent_command_xyz_12345"
}

@test "hg_require: fails on second missing command in a list" {
    run bash -c "source '${LIB_DIR}/hg-core.sh'; hg_require bash nonexistent_xyz_99 2>&1"
    assert_failure
    assert_output --partial "nonexistent_xyz_99"
}

# --- Path variable tests ---

@test "hg-core: HG_DOTFILES is set" {
    [[ -n "${HG_DOTFILES}" ]]
}

@test "hg-core ignores an invalid HG_STUDIO_ROOT and falls back to the sourced repo path" {
    local expected_dotfiles expected_studio
    expected_dotfiles="$(cd "${DOTFILES_DIR}" && pwd)"
    expected_studio="$(cd "${DOTFILES_DIR}/.." && pwd)"

    run env \
        HG_STUDIO_ROOT="${BATS_TEST_TMPDIR}/missing-studio" \
        DOTFILES_DIR="" \
        HG_STATE_DIR="${BATS_TEST_TMPDIR}/state" \
        bash -lc '
            source "'"${LIB_DIR}"'/hg-core.sh"
            printf "%s\n%s\n" "$HG_STUDIO_ROOT" "$HG_DOTFILES"
        '
    assert_success
    assert_line --index 0 "${expected_studio}"
    assert_line --index 1 "${expected_dotfiles}"
}

@test "hg-core ignores an invalid DOTFILES_DIR and falls back to the sourced repo path" {
    local expected_dotfiles expected_studio
    expected_dotfiles="$(cd "${DOTFILES_DIR}" && pwd)"
    expected_studio="$(cd "${DOTFILES_DIR}/.." && pwd)"

    run env \
        DOTFILES_DIR="${BATS_TEST_TMPDIR}/missing-dotfiles" \
        HG_STUDIO_ROOT="" \
        HG_STATE_DIR="${BATS_TEST_TMPDIR}/state" \
        bash -lc '
            source "'"${LIB_DIR}"'/hg-core.sh"
            printf "%s\n%s\n" "$HG_STUDIO_ROOT" "$HG_DOTFILES"
        '
    assert_success
    assert_line --index 0 "${expected_studio}"
    assert_line --index 1 "${expected_dotfiles}"
}

@test "hg-core: HG_STATE_DIR is set" {
    [[ -n "${HG_STATE_DIR}" ]]
}

@test "hg-core: state directory is created on source" {
    [[ -d "${HG_STATE_DIR}" ]]
}
