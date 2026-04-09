#!/usr/bin/env bats

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export TEST_HOME="${BATS_TEST_TMPDIR}/home"
    export TEST_ROOT="${TEST_HOME}/hairglasses-studio"
    export HOME="${TEST_HOME}"
    export HG_STUDIO_ROOT="${TEST_ROOT}"

    mkdir -p "${HOME}" "${HG_STUDIO_ROOT}"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

@test "install.sh prints link specs for core runtime and launcher shims" {
    run bash "${DOTFILES_DIR}/install.sh" --print-link-specs
    assert_success
    assert_output --partial "${DOTFILES_DIR}/ghostty|${HOME}/.config/ghostty"
    assert_output --partial "${DOTFILES_DIR}/kitty|${HOME}/.config/kitty"
    assert_output --partial "${DOTFILES_DIR}/scripts/app-launcher.sh|${HOME}/.local/bin/app-launcher"
    assert_output --partial "${DOTFILES_DIR}/scripts/app-switcher.sh|${HOME}/.local/bin/app-switcher"
}

@test "install.sh rejects unknown flags with exit code 2" {
    run bash "${DOTFILES_DIR}/install.sh" --definitely-not-a-real-flag
    assert_failure
    [[ "$status" -eq 2 ]]
    assert_output --partial "Unknown option"
}

@test "hg help renders the module surface" {
    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_DIR}/hg" --help
    assert_success
    assert_output --partial "MODULES"
    assert_output --partial "doctor"
    assert_output --partial "config"
}

@test "hg workflow sync help stays informational" {
    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_DIR}/hg-workflow-sync.sh" --help
    assert_success
    assert_output --partial "Hosted GitHub workflow sync is retired"
    assert_output --partial "informational only"
}

@test "hg workflow sync dry-run exits cleanly without mutating workflows" {
    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_DIR}/hg-workflow-sync.sh" --dry-run
    assert_success
    assert_output --partial "Hosted workflow sync is retired"
    assert_output --partial "Dry-run mode is informational only"
    assert_output --partial "No hosted workflows are managed"
}

@test "hg mcp mirror parity list exposes only mirror-managed bundled modules" {
    run bash "${SCRIPTS_DIR}/hg-mcp-mirror-parity.sh" --list
    assert_success
    assert_output --partial "dotfiles-mcp"
    assert_output --partial "process-mcp"
    assert_output --partial "systemd-mcp"
    assert_output --partial "tmux-mcp"
    refute_output --partial "mapitall"
    refute_output --partial "mapping"
}

@test "hg mcp mirror parity check passes for the tracked mirror set" {
    run bash "${SCRIPTS_DIR}/hg-mcp-mirror-parity.sh" --check
    assert_success
    assert_output --partial "PASS  dotfiles-mcp"
    assert_output --partial "PASS  process-mcp"
    assert_output --partial "PASS  systemd-mcp"
    assert_output --partial "PASS  tmux-mcp"
}
