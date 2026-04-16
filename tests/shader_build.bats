#!/usr/bin/env bats
# Tests for kitty/shaders/bin/shader-build.sh — Kitty DarkWindow catalog validator

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export TEST_DOTFILES="${BATS_TEST_TMPDIR}/dotfiles"
    export TEST_SCRIPT="${TEST_DOTFILES}/kitty/shaders/bin/shader-build.sh"
    export TEST_SHADER_DIR="${TEST_DOTFILES}/kitty/shaders/darkwindow"
    export TEST_LIB_DIR="${TEST_DOTFILES}/scripts/lib"

    mkdir -p "${TEST_DOTFILES}/kitty/shaders/bin" "${TEST_SHADER_DIR}" "${TEST_LIB_DIR}"
    cp "${DOTFILES_DIR}/kitty/shaders/bin/shader-build.sh" "${TEST_SCRIPT}"
    cp "${DOTFILES_DIR}/scripts/lib/hg-core.sh" "${TEST_LIB_DIR}/hg-core.sh"
    chmod +x "${TEST_SCRIPT}"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

@test "shader-build: build reports direct Kitty DarkWindow catalog usage" {
    printf 'void main() {}\n' > "${TEST_SHADER_DIR}/digital-mist.glsl"
    printf 'void main() {}\n' > "${TEST_SHADER_DIR}/neon-glow.glsl"

    run bash "${TEST_SCRIPT}"
    assert_success
    assert_output --partial "No transpilation required; using Kitty DarkWindow shaders directly (2 files)"
}

@test "shader-build: check validates the shader catalog" {
    printf 'void main() {}\n' > "${TEST_SHADER_DIR}/digital-mist.glsl"

    run bash "${TEST_SCRIPT}" check
    assert_success
    assert_output --partial "Kitty shader catalog ready (1 DarkWindow shaders)"
}

@test "shader-build: clean is a no-op for canonical DarkWindow assets" {
    run bash "${TEST_SCRIPT}" clean
    assert_success
    assert_output --partial "Nothing to clean; DarkWindow shaders are stored canonically"
}

@test "shader-build: check fails when the shader directory is missing" {
    rm -rf "${TEST_SHADER_DIR}"

    run bash "${TEST_SCRIPT}" check
    assert_failure
    assert_output --partial "Shader directory not found"
}
