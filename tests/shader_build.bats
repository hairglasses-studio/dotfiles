#!/usr/bin/env bats
# Tests for ghostty/shaders/bin/shader-build.sh — GLSL preprocessor

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    mock_notify_send

    # Build synthetic shader dir structure
    export TEST_SHADERS="${BATS_TEST_TMPDIR}/shaders"
    export TEST_LIB="${TEST_SHADERS}/lib"
    mkdir -p "$TEST_LIB" "${TEST_SHADERS}/bin"

    # Create shared library files
    cat > "${TEST_LIB}/hash.glsl" << 'GLSL'
float hash(vec2 p) {
    return fract(sin(dot(p, vec2(127.1, 311.7))) * 43758.5453);
}
GLSL

    cat > "${TEST_LIB}/noise.glsl" << 'GLSL'
float noise(vec2 p) {
    return 0.5;
}
GLSL

    # Create a patched copy of shader-build.sh that uses our test paths
    local real_script="${DOTFILES_DIR}/ghostty/shaders/bin/shader-build.sh"
    export PATCHED_BUILD="${BATS_TEST_TMPDIR}/shader-build.sh"
    {
        echo '#!/usr/bin/env bash'
        echo 'set -euo pipefail'
        echo "SCRIPT_DIR=\"${TEST_SHADERS}/bin\""
        echo "SHADERS_DIR=\"${TEST_SHADERS}\""
        echo "LIB_DIR=\"${TEST_LIB}\""
        # Stub out notify
        echo 'hg_notify() { true; }'
        echo 'hg_notify_low() { true; }'
        echo 'hg_notify_critical() { true; }'
        # Copy everything after the 'source' notify line
        sed -n '/^CHECK_ONLY=/,$p' "$real_script"
    } > "$PATCHED_BUILD"
    chmod +x "$PATCHED_BUILD"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

# Helper: create a shader file with the given content
make_shader() {
    local name="$1"
    shift
    printf '%s\n' "$@" > "${TEST_SHADERS}/${name}"
}

# --- Fresh include inlining ---

@test "shader-build: inlines library on fresh include directive" {
    make_shader "test.glsl" \
        "#version 420" \
        "" \
        '// #include "lib/hash.glsl"' \
        "" \
        "void main() { float h = hash(gl_FragCoord.xy); }"

    run bash "$PATCHED_BUILD" "${TEST_SHADERS}/test.glsl"
    assert_success
    assert_output --partial "Built: test.glsl"

    # Verify the inlined content
    run cat "${TEST_SHADERS}/test.glsl"
    assert_output --partial "BEGIN lib/hash.glsl"
    assert_output --partial "float hash(vec2 p)"
    assert_output --partial "END lib/hash.glsl"
}

@test "shader-build: preserves original include directive" {
    make_shader "test.glsl" \
        "#version 420" \
        '// #include "lib/hash.glsl"' \
        "void main() {}"

    run bash "$PATCHED_BUILD" "${TEST_SHADERS}/test.glsl"
    assert_success

    run cat "${TEST_SHADERS}/test.glsl"
    assert_output --partial '// #include "lib/hash.glsl"'
}

# --- Idempotent re-inlining ---

@test "shader-build: re-inlining is idempotent" {
    make_shader "test.glsl" \
        "#version 420" \
        '// #include "lib/hash.glsl"' \
        "void main() {}"

    # Build twice
    run bash "$PATCHED_BUILD" "${TEST_SHADERS}/test.glsl"
    assert_success
    local first_build
    first_build="$(cat "${TEST_SHADERS}/test.glsl")"

    run bash "$PATCHED_BUILD" "${TEST_SHADERS}/test.glsl"
    assert_success
    local second_build
    second_build="$(cat "${TEST_SHADERS}/test.glsl")"

    [[ "$first_build" == "$second_build" ]]
}

# --- Multiple includes ---

@test "shader-build: handles multiple include directives (separated by blank line)" {
    make_shader "test.glsl" \
        "#version 420" \
        '// #include "lib/hash.glsl"' \
        "" \
        '// #include "lib/noise.glsl"' \
        "" \
        "void main() {}"

    run bash "$PATCHED_BUILD" "${TEST_SHADERS}/test.glsl"
    assert_success

    run cat "${TEST_SHADERS}/test.glsl"
    assert_output --partial "BEGIN lib/hash.glsl"
    assert_output --partial "END lib/hash.glsl"
    assert_output --partial "BEGIN lib/noise.glsl"
    assert_output --partial "END lib/noise.glsl"
}

# --- Check mode ---

@test "shader-build: --check reports changes needed" {
    make_shader "test.glsl" \
        "#version 420" \
        '// #include "lib/hash.glsl"' \
        "void main() {}"

    run bash "$PATCHED_BUILD" --check "${TEST_SHADERS}/test.glsl"
    assert_failure
    assert_output --partial "CHANGED: test.glsl"
}

@test "shader-build: --check does not modify the file" {
    make_shader "test.glsl" \
        "#version 420" \
        '// #include "lib/hash.glsl"' \
        "void main() {}"
    local original
    original="$(cat "${TEST_SHADERS}/test.glsl")"

    run bash "$PATCHED_BUILD" --check "${TEST_SHADERS}/test.glsl"

    local after
    after="$(cat "${TEST_SHADERS}/test.glsl")"
    [[ "$original" == "$after" ]]
}

# --- No includes ---

@test "shader-build: shader without includes reports no change" {
    make_shader "plain.glsl" \
        "#version 420" \
        "void main() { gl_FragColor = vec4(1.0); }"

    run bash "$PATCHED_BUILD" "${TEST_SHADERS}/plain.glsl"
    assert_success
    assert_output --partial "No change: plain.glsl"
}

# --- Strip mode ---

@test "shader-build: --strip removes inlined blocks" {
    # First build to inline
    make_shader "test.glsl" \
        "#version 420" \
        '// #include "lib/hash.glsl"' \
        "void main() {}"

    run bash "$PATCHED_BUILD" "${TEST_SHADERS}/test.glsl"
    assert_success

    # Now strip
    run bash "$PATCHED_BUILD" --strip "${TEST_SHADERS}/test.glsl"
    assert_success

    # Inlined code should be gone
    run cat "${TEST_SHADERS}/test.glsl"
    refute_output --partial "float hash(vec2 p)"
    refute_output --partial "BEGIN lib/hash.glsl"
}

# --- Missing library warning ---

@test "shader-build: warns when included library does not exist" {
    make_shader "test.glsl" \
        "#version 420" \
        '// #include "lib/nonexistent.glsl"' \
        "void main() {}"

    run bash "$PATCHED_BUILD" "${TEST_SHADERS}/test.glsl" 2>&1
    assert_output --partial "WARNING"
    assert_output --partial "nonexistent.glsl"
}

# --- No files argument ---

@test "shader-build: exits cleanly with no files" {
    run bash "$PATCHED_BUILD"
    assert_success
    assert_output --partial "No files to process"
}
