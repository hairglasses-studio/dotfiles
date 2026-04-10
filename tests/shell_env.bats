#!/usr/bin/env bats

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export HOME="${BATS_TEST_TMPDIR}/home"
    mkdir -p "${HOME}"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

@test "zsh quiet mode defines managed provider wrappers without loading the full shell stack" {
    run env HOME="${HOME}" DOTFILES_DIR="${DOTFILES_DIR}" HG_AGENT_SESSION_QUIET=1 zsh -fc '
        source "$DOTFILES_DIR/zsh/zshrc"
        functions codex claude gemini
    '
    assert_success
    assert_output --partial '"$DOTFILES_DIR/scripts/hg-codex-launch.sh"'
    assert_output --partial '"$DOTFILES_DIR/scripts/hg-claude-launch.sh"'
    assert_output --partial '"$DOTFILES_DIR/scripts/hg-gemini-launch.sh"'
    assert_output --partial "HG_AGENT_SESSION_QUIET=1"
}
