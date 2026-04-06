#!/usr/bin/env bats
# Tests for scripts/font-playlist.sh — Font management and cycling
# Tests: list parsing, cycling logic, state persistence, entry parsing
# Skips: actual font application, Ghostty config, desktop notifications

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"

    # Set up fake state and playlist dirs
    export TEST_STATE="${BATS_TEST_TMPDIR}/state"
    export TEST_PLAYLIST_DIR="${BATS_TEST_TMPDIR}/fonts"
    export TEST_DOTFILES="${BATS_TEST_TMPDIR}/dotfiles"
    export TEST_CONFIG="${TEST_DOTFILES}/ghostty/config"
    mkdir -p "$TEST_STATE" "$TEST_PLAYLIST_DIR" "$(dirname "$TEST_CONFIG")"

    # Create synthetic playlist
    cat > "${TEST_PLAYLIST_DIR}/test-playlist.txt" << 'EOF'
# Test font playlist — family|size
Hack Nerd Font|12
Iosevka Nerd Font Mono|12
Fira Code|14
Hack Nerd Font Mono|13
EOF

    # Create a minimal Ghostty config
    cat > "$TEST_CONFIG" << 'CFG'
font-family = Hack Nerd Font
font-size = 12
background = #000000
CFG

    # Create patched script
    local real_script="${SCRIPTS_DIR}/font-playlist.sh"
    export PATCHED_SCRIPT="${BATS_TEST_TMPDIR}/font-playlist.sh"
    {
        echo '#!/usr/bin/env bash'
        echo 'set -euo pipefail'
        echo ""
        # Provide hg-core stubs
        echo 'HG_CYAN=$'"'"'\033[36m'"'"''
        echo 'HG_GREEN=$'"'"'\033[32m'"'"''
        echo 'HG_RED=$'"'"'\033[31m'"'"''
        echo 'HG_YELLOW=$'"'"'\033[33m'"'"''
        echo 'HG_MAGENTA=$'"'"'\033[35m'"'"''
        echo 'HG_DIM=$'"'"'\033[2m'"'"''
        echo 'HG_BOLD=$'"'"'\033[1m'"'"''
        echo 'HG_RESET=$'"'"'\033[0m'"'"''
        echo "HG_DOTFILES=\"${TEST_DOTFILES}\""
        echo "HG_STATE_DIR=\"${TEST_STATE}\""
        echo 'hg_info()  { printf "[info]  %s\n" "$1"; }'
        echo 'hg_ok()    { printf "[ok]    %s\n" "$1"; }'
        echo 'hg_error() { printf "[err]   %s\n" "$1" >&2; }'
        echo 'hg_die()   { hg_error "$1"; exit "${2:-1}"; }'
        echo 'hg_notify_low() { true; }'
        echo ""
        echo "SCRIPT_DIR=\"${SCRIPTS_DIR}\""
        echo ""
        # Override the paths
        echo "_ghostty_config=\"${TEST_CONFIG}\""
        echo "_playlist_dir=\"${TEST_PLAYLIST_DIR}\""
        echo "_state_dir=\"${TEST_STATE}\""
        echo '_default_playlist="test-playlist"'
        echo ""
        echo 'mkdir -p "$_state_dir"'
        echo ""
        # Copy the function definitions from the real script
        sed -n '/^_get_idx()/,$p' "$real_script"
    } > "$PATCHED_SCRIPT"
    chmod +x "$PATCHED_SCRIPT"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

# --- Playlist loading ---

@test "font-playlist: list shows playlist entries" {
    run bash "$PATCHED_SCRIPT" list
    assert_success
    assert_output --partial "Hack Nerd Font"
    assert_output --partial "Iosevka Nerd Font Mono"
    assert_output --partial "Fira Code"
    assert_output --partial "Hack Nerd Font Mono"
}

@test "font-playlist: list shows entry count" {
    run bash "$PATCHED_SCRIPT" list
    assert_success
    assert_output --partial "4 fonts"
}

@test "font-playlist: list with explicit playlist name" {
    run bash "$PATCHED_SCRIPT" list test-playlist
    assert_success
    assert_output --partial "4 fonts"
}

# --- Cycling forward ---

@test "font-playlist: next advances to second entry" {
    run bash "$PATCHED_SCRIPT" next
    assert_success
    assert_output --partial "Iosevka Nerd Font Mono"
    assert_output --partial "12pt"
    assert_output --partial "[2/4]"
}

@test "font-playlist: next updates Ghostty config font-family" {
    bash "$PATCHED_SCRIPT" next
    run grep '^font-family = ' "$TEST_CONFIG"
    assert_success
    assert_output "font-family = Iosevka Nerd Font Mono"
}

@test "font-playlist: next updates Ghostty config font-size" {
    # Advance twice to get to Fira Code|14
    bash "$PATCHED_SCRIPT" next
    bash "$PATCHED_SCRIPT" next
    run grep '^font-size = ' "$TEST_CONFIG"
    assert_success
    assert_output "font-size = 14"
}

@test "font-playlist: next wraps around at end of playlist" {
    # Advance 4 times (back to index 0)
    bash "$PATCHED_SCRIPT" next
    bash "$PATCHED_SCRIPT" next
    bash "$PATCHED_SCRIPT" next
    run bash "$PATCHED_SCRIPT" next
    assert_success
    assert_output --partial "Hack Nerd Font"
    assert_output --partial "[1/4]"
}

# --- Cycling backward ---

@test "font-playlist: prev wraps to last entry from start" {
    run bash "$PATCHED_SCRIPT" prev
    assert_success
    assert_output --partial "Hack Nerd Font Mono"
    assert_output --partial "[4/4]"
}

# --- State persistence ---

@test "font-playlist: index persists across invocations" {
    bash "$PATCHED_SCRIPT" next
    bash "$PATCHED_SCRIPT" next
    # Index should be 2 now
    local idx
    idx=$(cat "${TEST_STATE}/test-playlist.idx")
    [[ "$idx" -eq 2 ]]
}

@test "font-playlist: reset sets index back to 0" {
    bash "$PATCHED_SCRIPT" next
    bash "$PATCHED_SCRIPT" next
    run bash "$PATCHED_SCRIPT" reset
    assert_success
    assert_output --partial "Reset"
    local idx
    idx=$(cat "${TEST_STATE}/test-playlist.idx")
    [[ "$idx" -eq 0 ]]
}

# --- Current display ---

@test "font-playlist: current shows active font and size" {
    run bash "$PATCHED_SCRIPT" current
    assert_success
    assert_output --partial "Hack Nerd Font"
    assert_output --partial "12pt"
}

# --- Missing playlist ---

@test "font-playlist: next with nonexistent playlist exits with error" {
    run bash "$PATCHED_SCRIPT" next nonexistent
    assert_failure
    assert_output --partial "Playlist not found"
}

# --- Usage output ---

@test "font-playlist: no arguments shows usage" {
    run bash "$PATCHED_SCRIPT"
    assert_success
    assert_output --partial "Usage: font-playlist.sh"
    assert_output --partial "next"
    assert_output --partial "prev"
    assert_output --partial "current"
    assert_output --partial "list"
    assert_output --partial "reset"
}
