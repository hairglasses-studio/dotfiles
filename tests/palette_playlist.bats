#!/usr/bin/env bats

load 'test_helper'

# palette-playlist.sh is the palette rotation CLI. These cases exercise
# the read-only subcommands (list / current / status) against a temp
# worktree, so they can't touch the user's live theme/palette.env. The
# mutating subcommands (set / next / prev / random / reset) delegate to
# palette-propagate.sh which has its own chezmoi-driven integration
# path; testing them here would require a full chezmoi workspace and
# is out of scope for bats smoke.

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export HOME="${BATS_TEST_TMPDIR}/home"
    export XDG_STATE_HOME="${HOME}/.local/state"
    mkdir -p "${HOME}" "${XDG_STATE_HOME}"

    # Copy just the palette + scripts surface the tests need.
    mkdir -p "${BATS_TEST_TMPDIR}/dotfiles/theme/palettes"
    mkdir -p "${BATS_TEST_TMPDIR}/dotfiles/scripts/lib"
    cp "${DOTFILES_DIR}/scripts/palette-playlist.sh" "${BATS_TEST_TMPDIR}/dotfiles/scripts/"
    cp "${DOTFILES_DIR}/scripts/lib/hg-core.sh" "${BATS_TEST_TMPDIR}/dotfiles/scripts/lib/"
    cp "${DOTFILES_DIR}/scripts/lib/notify.sh" "${BATS_TEST_TMPDIR}/dotfiles/scripts/lib/"
    cp "${DOTFILES_DIR}"/theme/palettes/*.env "${BATS_TEST_TMPDIR}/dotfiles/theme/palettes/"

    # Active palette symlink.
    ln -s "palettes/hairglasses-neon.env" "${BATS_TEST_TMPDIR}/dotfiles/theme/palette.env"

    # HG_DOTFILES is what lib/hg-core.sh uses to resolve the repo root.
    export HG_DOTFILES="${BATS_TEST_TMPDIR}/dotfiles"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

@test "palette-playlist list marks the active palette with a star" {
    run bash "${BATS_TEST_TMPDIR}/dotfiles/scripts/palette-playlist.sh" list
    assert_success
    assert_output --partial "* hairglasses-neon"
    assert_output --partial "amber"
    assert_output --partial "synthwave"
}

@test "palette-playlist current prints the active palette name" {
    run bash "${BATS_TEST_TMPDIR}/dotfiles/scripts/palette-playlist.sh" current
    assert_success
    assert_output "hairglasses-neon"
}

@test "palette-playlist status reports the symlink target" {
    run bash "${BATS_TEST_TMPDIR}/dotfiles/scripts/palette-playlist.sh" status
    assert_success
    assert_output --partial "palette: hairglasses-neon"
    assert_output --partial "target:  palettes/hairglasses-neon.env"
    assert_output --partial "total:   9 palette(s)"
}

@test "palette-playlist help prints the command surface" {
    run bash "${BATS_TEST_TMPDIR}/dotfiles/scripts/palette-playlist.sh" help
    assert_success
    assert_output --partial "list"
    assert_output --partial "set <name>"
    assert_output --partial "next"
    assert_output --partial "random"
    assert_output --partial "theme/palette.env"
}

@test "palette-playlist rejects unknown commands" {
    run bash "${BATS_TEST_TMPDIR}/dotfiles/scripts/palette-playlist.sh" frobnicate
    assert_failure
    assert_output --partial "Unknown command: frobnicate"
}
