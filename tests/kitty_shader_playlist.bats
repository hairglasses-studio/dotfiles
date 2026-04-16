#!/usr/bin/env bats

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export HOME="${BATS_TEST_TMPDIR}/home"
    export DOTFILES_DIR="${BATS_TEST_TMPDIR}/dotfiles"
    export PATH="${BATS_TEST_TMPDIR}/bin:${PATH}"
    mkdir -p \
        "${HOME}" \
        "${DOTFILES_DIR}/kitty/shaders/darkwindow" \
        "${DOTFILES_DIR}/kitty/shaders/playlists" \
        "${DOTFILES_DIR}/kitty/themes/playlists" \
        "${BATS_TEST_TMPDIR}/bin"

    cat > "${DOTFILES_DIR}/kitty/shaders/playlists/ambient.txt" <<'EOF'
digital-mist.glsl
neon-glow.glsl
EOF
    cat > "${DOTFILES_DIR}/kitty/themes/playlists/ambient.txt" <<'EOF'
Dracula
Gruvbox Dark
EOF

    cat > "${BATS_TEST_TMPDIR}/bin/kitty" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
if [[ "${1:-}" == "+kitten" && "${2:-}" == "themes" && "${3:-}" == "--dump-theme" ]]; then
    theme="${4:-unknown}"
    printf '# theme: %s\nbackground #111111\nforeground #eeeeee\n' "$theme"
    exit 0
fi
echo "unexpected kitty invocation: $*" >&2
exit 1
EOF
    chmod +x "${BATS_TEST_TMPDIR}/bin/kitty"

    cat > "${BATS_TEST_TMPDIR}/bin/kitten" <<'EOF'
#!/usr/bin/env bash
echo "$*" >> "${BATS_TEST_TMPDIR}/kitten.log"
exit 0
EOF
    chmod +x "${BATS_TEST_TMPDIR}/bin/kitten"

    cat > "${BATS_TEST_TMPDIR}/bin/notify-send" <<'EOF'
#!/usr/bin/env bash
echo "$*" >> "${BATS_TEST_TMPDIR}/notify.log"
exit 0
EOF
    chmod +x "${BATS_TEST_TMPDIR}/bin/notify-send"

    cat > "${BATS_TEST_TMPDIR}/bin/hyprctl" <<'EOF'
#!/usr/bin/env bash
echo "$*" >> "${BATS_TEST_TMPDIR}/hyprctl.log"
echo "ok"
EOF
    chmod +x "${BATS_TEST_TMPDIR}/bin/hyprctl"

    cat > "${BATS_TEST_TMPDIR}/bin/shuf" <<'EOF'
#!/usr/bin/env bash
# Deterministic shuf: always pick the first line
if [[ "${1:-}" == "-n" ]]; then
    head -1
else
    cat
fi
EOF
    chmod +x "${BATS_TEST_TMPDIR}/bin/shuf"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

@test "kitty-shader-playlist set applies a theme by name" {
    run bash "${SCRIPTS_DIR}/kitty-shader-playlist.sh" set Dracula
    assert_success
    assert_output --partial "Dracula"

    run bash "${SCRIPTS_DIR}/kitty-shader-playlist.sh" current
    assert_success
    assert_output "Dracula"
}

@test "kitty-shader-playlist next cycles the theme" {
    run bash "${SCRIPTS_DIR}/kitty-shader-playlist.sh" reset ambient
    assert_success

    run bash "${SCRIPTS_DIR}/kitty-shader-playlist.sh" next
    assert_success
    # Should output a theme name from the playlist
    assert_output --partial "[ok]"
}

@test "kitty-shader-playlist status reports playlist and theme" {
    run bash "${SCRIPTS_DIR}/kitty-shader-playlist.sh" next
    assert_success

    run bash "${SCRIPTS_DIR}/kitty-shader-playlist.sh" status
    assert_success
    assert_output --partial "playlist: ambient"
    assert_output --partial "theme:"
}

@test "kitty-shader-playlist shader-next dispatches via hyprctl" {
    run bash "${SCRIPTS_DIR}/kitty-shader-playlist.sh" shader-next
    assert_success

    run cat "${BATS_TEST_TMPDIR}/hyprctl.log"
    assert_output --partial "dispatch darkwindow:shadeactive"
}

@test "kitty-shader-playlist shader-toggle re-dispatches current shader" {
    # First apply a shader
    run bash "${SCRIPTS_DIR}/kitty-shader-playlist.sh" shader-next
    assert_success

    # Clear the log
    : > "${BATS_TEST_TMPDIR}/hyprctl.log"

    # Toggle should re-dispatch the same shader
    run bash "${SCRIPTS_DIR}/kitty-shader-playlist.sh" shader-toggle
    assert_success
    assert_output --partial "toggled"

    run cat "${BATS_TEST_TMPDIR}/hyprctl.log"
    assert_output --partial "dispatch darkwindow:shadeactive"
}

@test "kitty-shader-playlist derives dotfiles root from script path when HOME differs" {
    run env -u DOTFILES_DIR -u HG_STUDIO_ROOT HOME="${BATS_TEST_TMPDIR}/alt-home" \
        bash "${SCRIPTS_DIR}/kitty-shader-playlist.sh" list ambient
    assert_success
    assert_output --partial "theme playlist: ambient"
}
