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
    assert_output --partial "${DOTFILES_DIR}/scripts/kitty-shell-launch.sh|${HOME}/.local/bin/kitty-shell-launch"
    assert_output --partial "${DOTFILES_DIR}/scripts/app-launcher.sh|${HOME}/.local/bin/app-launcher"
    assert_output --partial "${DOTFILES_DIR}/scripts/app-switcher.sh|${HOME}/.local/bin/app-switcher"
}

@test "installers keep kitty save-session helpers opt-in" {
    run bash -lc "! grep -Eq 'dotfiles-kitty-save-session\\.(service|timer)' '${DOTFILES_DIR}/install.sh' '${DOTFILES_DIR}/manjaro/install.sh'"
    assert_success
}

@test "repo config pins the shell-first kitty launch policy" {
    run bash -lc "grep -F 'default_terminal = \"\$HOME/.local/bin/kitty-shell-launch\"' '${DOTFILES_DIR}/hyprshell/config.toml' && grep -Eq '^startup_session[[:space:]]+none$' '${DOTFILES_DIR}/kitty/kitty.conf'"
    assert_success
}

@test "launcher consumers stay pinned to the managed kitty wrappers" {
    run bash -lc "grep -F '\$HOME/.local/bin/kitty-dev-launch' '${DOTFILES_DIR}/hyprland/hyprland.conf' && grep -F '\$HOME/.local/bin/kitty-visual-launch' '${DOTFILES_DIR}/hyprland/hyprland.conf' && grep -F '\$HOME/.local/bin/kitty-visual-launch --class=scratchpad' '${DOTFILES_DIR}/pypr/config.toml' && grep -F '\$HOME/.local/bin/kitty-visual-launch --class=scratchpad' '${DOTFILES_DIR}/hyprland/pyprland.toml' && grep -F 'kitty-visual-launch -e /home/hg/.local/bin/hg-codex-launch.sh' '${DOTFILES_DIR}/juhradial/config.json' '${DOTFILES_DIR}/juhradial/profiles.json' && grep -F 'kitty-visual-launch' '${DOTFILES_DIR}/ironbar/config.toml' '${DOTFILES_DIR}/makima/Microsoft Xbox Series S|X Controller.toml'"
    assert_success
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

@test "hg input help exposes the verify workflow" {
    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_DIR}/hg" input --help
    assert_success
    assert_output --partial "verify"
    assert_output --partial "wheel-fix"
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

@test "hg mcp mirror parity list exposes mirrored bundled modules" {
    run bash "${SCRIPTS_DIR}/hg-mcp-mirror-parity.sh" --list
    assert_success
    assert_output --partial "dotfiles-mcp"
    assert_output --partial "manual_projection"
    assert_output --partial "mapitall"
    assert_output --partial "mapping"
    assert_output --partial "tmux-mcp"
}

@test "hg mcp mirror parity check passes for the tracked mirror set" {
    run bash "${SCRIPTS_DIR}/hg-mcp-mirror-parity.sh" --check
    assert_success
    assert_output --partial "PASS  dotfiles-mcp"
    assert_output --partial "PASS  mapping"
    assert_output --partial "PASS  tmux-mcp"
}
