#!/usr/bin/env bats
# Tests for kitty/shaders/bin/shader-meta.sh — TOML parser and metadata queries

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"

    # Build a synthetic shader dir with manifest and .glsl files
    export TEST_SHADERS_DIR="${BATS_TEST_TMPDIR}/shaders"
    export TEST_PLAYLISTS_DIR="${TEST_SHADERS_DIR}/playlists"
    mkdir -p "$TEST_SHADERS_DIR" "$TEST_PLAYLISTS_DIR"

    # Create synthetic manifest
    cat > "${TEST_SHADERS_DIR}/shaders.toml" << 'TOML'
[shaders."bloom-soft"]
category = "Post-FX"
cost = "LOW"
source = "original"
description = "Soft bloom glow effect"
playlists = ["low-intensity", "best-of"]

[shaders."crt-amber"]
category = "CRT"
cost = "MED"
source = "seanwcom"
description = "Amber phosphor CRT"
playlists = ["low-intensity"]

[shaders."cyber-rain"]
category = "Cyberpunk"
cost = "HIGH"
source = "shadertoy-port"
description = "Neon rain animation"
playlists = ["high-intensity", "best-of"]
TOML

    # Create matching .glsl stub files
    echo "// bloom-soft stub" > "${TEST_SHADERS_DIR}/bloom-soft.glsl"
    echo "// crt-amber stub" > "${TEST_SHADERS_DIR}/crt-amber.glsl"
    echo "// cyber-rain stub" > "${TEST_SHADERS_DIR}/cyber-rain.glsl"

    # Create a patched copy of shader-meta.sh that uses our test paths
    local real_script="${DOTFILES_DIR}/kitty/shaders/bin/shader-meta.sh"
    export PATCHED_META="${BATS_TEST_TMPDIR}/shader-meta.sh"
    {
        echo '#!/usr/bin/env bash'
        echo 'set -euo pipefail'
        echo "SCRIPT_DIR=\"${BATS_TEST_TMPDIR}\""
        echo "SHADERS_DIR=\"${TEST_SHADERS_DIR}\""
        echo "MANIFEST=\"${TEST_SHADERS_DIR}/shaders.toml\""
        echo "PLAYLISTS_DIR=\"${TEST_PLAYLISTS_DIR}\""
        # Copy everything after the first PLAYLISTS_DIR= line from the real script
        sed -n '/^PLAYLISTS_DIR=/,$p' "$real_script" | tail -n +2
    } > "$PATCHED_META"
    chmod +x "$PATCHED_META"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

# --- get command tests ---

@test "shader-meta: get category returns correct value" {
    run bash "$PATCHED_META" get bloom-soft category
    assert_success
    assert_output "Post-FX"
}

@test "shader-meta: get cost returns correct value" {
    run bash "$PATCHED_META" get crt-amber cost
    assert_success
    assert_output "MED"
}

@test "shader-meta: get description returns full text" {
    run bash "$PATCHED_META" get cyber-rain description
    assert_success
    assert_output "Neon rain animation"
}

@test "shader-meta: get source returns correct value" {
    run bash "$PATCHED_META" get bloom-soft source
    assert_success
    assert_output "original"
}

@test "shader-meta: get playlists returns array items" {
    run bash "$PATCHED_META" get bloom-soft playlists
    assert_success
    assert_line --index 0 "low-intensity"
    assert_line --index 1 "best-of"
}

@test "shader-meta: get strips .glsl extension from name" {
    run bash "$PATCHED_META" get bloom-soft.glsl category
    assert_success
    assert_output "Post-FX"
}

# --- list command tests ---

@test "shader-meta: list without filters returns all shaders" {
    run bash "$PATCHED_META" list
    assert_success
    assert_line "bloom-soft.glsl"
    assert_line "crt-amber.glsl"
    assert_line "cyber-rain.glsl"
}

@test "shader-meta: list --category filters correctly" {
    run bash "$PATCHED_META" list --category CRT
    assert_success
    assert_output "crt-amber.glsl"
}

@test "shader-meta: list --cost filters correctly" {
    run bash "$PATCHED_META" list --cost HIGH
    assert_success
    assert_output "cyber-rain.glsl"
}

@test "shader-meta: list --category with no matches produces empty output" {
    run bash "$PATCHED_META" list --category Nonexistent
    assert_success
    assert_output ""
}

# --- validate command tests ---

@test "shader-meta: validate passes when manifest matches files" {
    run bash "$PATCHED_META" validate
    assert_success
    assert_output --partial "OK"
}

@test "shader-meta: validate detects missing manifest entry" {
    echo "// orphan stub" > "${TEST_SHADERS_DIR}/orphan-shader.glsl"
    run bash "$PATCHED_META" validate
    assert_failure
    assert_output --partial "MISSING in manifest: orphan-shader"
}

@test "shader-meta: validate detects orphan manifest entry" {
    rm "${TEST_SHADERS_DIR}/cyber-rain.glsl"
    run bash "$PATCHED_META" validate
    assert_failure
    assert_output --partial "ORPHAN in manifest: cyber-rain"
}

# --- fzf-lines command tests ---

@test "shader-meta: fzf-lines outputs tab-separated lines" {
    run bash "$PATCHED_META" fzf-lines
    assert_success
    assert_output --partial "bloom-soft.glsl"
    assert_output --partial "crt-amber.glsl"
    assert_output --partial "cyber-rain.glsl"
}

@test "shader-meta: fzf-lines includes category in output" {
    run bash "$PATCHED_META" fzf-lines
    assert_success
    assert_output --partial "CRT"
    assert_output --partial "Post-FX"
    assert_output --partial "Cyberpunk"
}

# --- help and error ---

@test "shader-meta: help shows usage" {
    run bash "$PATCHED_META" help
    assert_success
    assert_output --partial "Usage: shader-meta"
}

@test "shader-meta: unknown command exits with error" {
    run bash "$PATCHED_META" nonexistent
    assert_failure
    assert_output --partial "Unknown command: nonexistent"
}
