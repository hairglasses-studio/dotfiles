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
    assert_output --partial "${DOTFILES_DIR}/kitty|${HOME}/.config/kitty"
    assert_output --partial "${DOTFILES_DIR}/ironbar|${HOME}/.config/ironbar"
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
    # hyprland/pyprland.toml was removed as a stale duplicate of pypr/config.toml
    # in the April 2026 cleanup; the makima Xbox controller TOML was dropped with
    # the gamepad-remapper retirement. The live pinned consumers are just the
    # three below.
    run bash -lc "grep -F '\$HOME/.local/bin/kitty-dev-launch' '${DOTFILES_DIR}/hyprland/hyprland.conf' && grep -F '\$HOME/.local/bin/kitty-visual-launch' '${DOTFILES_DIR}/hyprland/hyprland.conf' && grep -F '\$HOME/.local/bin/kitty-visual-launch --class=scratchpad' '${DOTFILES_DIR}/pypr/config.toml' && grep -F 'kitty-visual-launch' '${DOTFILES_DIR}/ironbar/config.toml'"
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

@test "hg gamepad help exposes the controller workflow" {
    # Replaces the old `hg input` test. The `input` module was retired when
    # makima/juhradial were removed; the surviving controller surface is
    # `hg gamepad` (Xbox + makima service control).
    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_DIR}/hg" gamepad --help
    assert_success
    assert_output --partial "status"
    assert_output --partial "profiles"
    assert_output --partial "restart"
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
}

@test "hg mcp mirror parity check passes for the tracked mirror set" {
    run bash "${SCRIPTS_DIR}/hg-mcp-mirror-parity.sh" --check
    assert_success
    assert_output --partial "PASS  dotfiles-mcp"
    assert_output --partial "PASS  mapitall"
    assert_output --partial "PASS  mapping"
    # tmux-mcp, systemd-mcp, process-mcp were consolidated into dotfiles-mcp
    # on 2026-04-16; they now live in the `consolidated` array of
    # mcp/mirror-parity.json and are no longer tracked by the parity checker.
    refute_output --partial "PASS  tmux-mcp"
    refute_output --partial "PASS  systemd-mcp"
    refute_output --partial "PASS  process-mcp"
}

@test "kitty theme playlists all resolve against the bundled catalog" {
    run bash "${SCRIPTS_DIR}/kitty-playlist-validate.sh"
    assert_success
    assert_output --partial "all resolve"
    refute_output --partial "MISSING"
}

@test "scripts/ has no unreferenced orphans outside the allowlist" {
    # A new script that neither lands in the allowlist nor gets wired into a
    # consumer (install.sh, systemd, a config, another script, etc.) is a
    # common class of cleanup debt. Keep the audit green; if a genuinely
    # new hook-installable or manual-invoke utility lands, extend the
    # allowlist inside audit-orphan-scripts.sh with a category comment.
    run bash "${SCRIPTS_DIR}/audit-orphan-scripts.sh" --strict
    assert_success
    assert_output --partial "orphans=0"
}

@test "install.sh retroarch entries resolve to executable scripts" {
    # install.sh maps \$DOTFILES_DIR/scripts/retroarch-*.{py,sh} to
    # \$HOME/.local/bin/retroarch-*. A rename that misses one side silently
    # breaks the installer; this test catches that by verifying every
    # retroarch link-spec on the left side exists and is executable.
    run bash "${DOTFILES_DIR}/install.sh" --print-link-specs
    assert_success

    local missing=()
    local non_executable=()
    while IFS='|' read -r src dest; do
        [[ "${src}" == *retroarch-* ]] || continue
        local resolved="${src/\$DOTFILES_DIR/${DOTFILES_DIR}}"
        if [[ ! -f "${resolved}" ]]; then
            missing+=("${resolved}")
        elif [[ ! -x "${resolved}" ]]; then
            non_executable+=("${resolved}")
        fi
    done <<< "${output}"

    [[ ${#missing[@]} -eq 0 ]] || fail "missing retroarch sources: ${missing[*]}"
    [[ ${#non_executable[@]} -eq 0 ]] || fail "retroarch scripts missing +x bit: ${non_executable[*]}"
}
