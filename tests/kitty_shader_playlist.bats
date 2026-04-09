#!/usr/bin/env bats

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export HOME="${BATS_TEST_TMPDIR}/home"
    export DOTFILES_DIR="${BATS_TEST_TMPDIR}/dotfiles"
    export PATH="${BATS_TEST_TMPDIR}/bin:${PATH}"
    mkdir -p \
        "${HOME}" \
        "${DOTFILES_DIR}/kitty/shaders/crtty" \
        "${DOTFILES_DIR}/kitty/shaders/playlists" \
        "${DOTFILES_DIR}/kitty/themes/playlists" \
        "${BATS_TEST_TMPDIR}/bin"

    cat > "${DOTFILES_DIR}/kitty/shaders/crtty/digital-mist.glsl" <<'EOF'
void main() {}
EOF
    cat > "${DOTFILES_DIR}/kitty/shaders/crtty/neon-glow.glsl" <<'EOF'
void main() {}
EOF
    cat > "${DOTFILES_DIR}/kitty/shaders/playlists/ambient.txt" <<'EOF'
digital-mist
EOF
    cat > "${DOTFILES_DIR}/kitty/shaders/playlists/best-of.txt" <<'EOF'
neon-glow
digital-mist
EOF
    cat > "${DOTFILES_DIR}/kitty/themes/playlists/ambient.txt" <<'EOF'
Dracula
EOF
    cat > "${DOTFILES_DIR}/kitty/themes/playlists/best-of.txt" <<'EOF'
Gruvbox Dark
Dracula
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
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

@test "kitty-shader-playlist set writes paired state" {
    run bash "${SCRIPTS_DIR}/kitty-shader-playlist.sh" set digital-mist Dracula
    assert_success
    assert_output --partial "Dracula · digital-mist"

    run bash "${SCRIPTS_DIR}/kitty-shader-playlist.sh" current
    assert_success
    assert_output "digital-mist"

    run bash "${SCRIPTS_DIR}/kitty-shader-playlist.sh" theme-current
    assert_success
    assert_output "Dracula"

    run cat "${HOME}/.local/state/kitty-shaders/current-label"
    assert_success
    assert_output "Dracula · digital-mist"
}

@test "kitty-shader-playlist theme-for-window consumes pending theme and prints label" {
    run bash "${SCRIPTS_DIR}/kitty-shader-playlist.sh" set neon-glow "Gruvbox Dark"
    assert_success

    run bash "${SCRIPTS_DIR}/kitty-shader-playlist.sh" theme-for-window 4242 best-of
    assert_success
    assert_output --partial $'Gruvbox Dark\t'
    assert_output --partial "Gruvbox Dark · neon-glow"

    run test ! -e "${HOME}/.local/state/kitty-shaders/pending-theme"
    assert_success
}

@test "kitty-shader-playlist status reports label and playlist position" {
    run bash "${SCRIPTS_DIR}/kitty-shader-playlist.sh" reset ambient
    assert_success

    run bash "${SCRIPTS_DIR}/kitty-shader-playlist.sh" next ambient
    assert_success
    assert_output --partial "Dracula · digital-mist"

    run bash "${SCRIPTS_DIR}/kitty-shader-playlist.sh" status
    assert_success
    assert_output --partial "playlist: ambient"
    assert_output --partial "shader:   digital-mist"
    assert_output --partial "theme:    Dracula"
    assert_output --partial "label:    Dracula · digital-mist"
    assert_output --partial "position: 1/1"
}

@test "kitty-shader-playlist set marks status as custom" {
    run bash "${SCRIPTS_DIR}/kitty-shader-playlist.sh" set neon-glow "Gruvbox Dark"
    assert_success

    run bash "${SCRIPTS_DIR}/kitty-shader-playlist.sh" status
    assert_success
    assert_output --partial "label:    Gruvbox Dark · neon-glow"
    assert_output --partial "position: custom"
}

@test "kitty-shader-playlist derives dotfiles root from script path when HOME differs" {
    run env -u DOTFILES_DIR -u HG_STUDIO_ROOT HOME="${BATS_TEST_TMPDIR}/alt-home" \
        bash "${SCRIPTS_DIR}/kitty-shader-playlist.sh" list ambient
    assert_success
    assert_output --partial "shader playlist: ambient"
}
