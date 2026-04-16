#!/usr/bin/env bats
# Tests for the Ghostty→DarkWindow shader transpiler and validation pipeline

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export TRANSPILER="${DOTFILES_DIR}/scripts/lib/shader-transpile-darkwindow.sh"
    export VALIDATOR="${DOTFILES_DIR}/scripts/validate-darkwindow-shaders.sh"
    export CONSISTENCY="${DOTFILES_DIR}/scripts/check-shader-consistency.sh"

    mkdir -p "${BATS_TEST_TMPDIR}/src" "${BATS_TEST_TMPDIR}/out"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

# Helper: transpile a source shader and validate the output
transpile_and_validate() {
    local src="$1" name="$2"
    local out="${BATS_TEST_TMPDIR}/out/${name}.glsl"
    bash "$TRANSPILER" "$src" "$out"
    bash "$VALIDATOR" "${BATS_TEST_TMPDIR}/out" --baseline 1 --out "${BATS_TEST_TMPDIR}/val"
}

@test "transpiler produces valid output for a minimal shader" {
    cat > "${BATS_TEST_TMPDIR}/src/minimal.glsl" <<'SHADER'
void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    fragColor = texture(iChannel0, fragCoord / iResolution.xy);
}
SHADER
    run transpile_and_validate "${BATS_TEST_TMPDIR}/src/minimal.glsl" "minimal"
    assert_success
}

@test "transpiler handles cursor uniforms (vec4 synthesis)" {
    cat > "${BATS_TEST_TMPDIR}/src/cursor.glsl" <<'SHADER'
void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 pos = iCurrentCursor.xy;
    vec2 size = iCurrentCursor.zw;
    vec2 prev = iPreviousCursor.xy;
    fragColor = texture(iChannel0, fragCoord / iResolution.xy);
}
SHADER
    run transpile_and_validate "${BATS_TEST_TMPDIR}/src/cursor.glsl" "cursor"
    assert_success
}

@test "transpiler handles normalize collision" {
    cat > "${BATS_TEST_TMPDIR}/src/normcol.glsl" <<'SHADER'
vec2 normalize(vec2 value, float isPosition) {
    return value * 2.0 / iResolution.xy - isPosition;
}
void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec2 uv = normalize(fragCoord, 1.0);
    fragColor = texture(iChannel0, uv);
}
SHADER
    run transpile_and_validate "${BATS_TEST_TMPDIR}/src/normcol.glsl" "normcol"
    assert_success
}

@test "transpiler handles iBackgroundColor as vec3" {
    cat > "${BATS_TEST_TMPDIR}/src/bgcolor.glsl" <<'SHADER'
void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    vec4 orig = texture(iChannel0, fragCoord / iResolution.xy);
    float d = distance(orig.rgb, iBackgroundColor);
    fragColor = orig * d;
}
SHADER
    run transpile_and_validate "${BATS_TEST_TMPDIR}/src/bgcolor.glsl" "bgcolor"
    assert_success
}

@test "transpiler handles brace-init rewrite for ES 300" {
    cat > "${BATS_TEST_TMPDIR}/src/braceinit.glsl" <<'SHADER'
void mainImage(out vec4 fragColor, in vec2 fragCoord) {
    float[] weights = {0.1, 0.3, 0.6};
    fragColor = texture(iChannel0, fragCoord / iResolution.xy) * weights[1];
}
SHADER
    run transpile_and_validate "${BATS_TEST_TMPDIR}/src/braceinit.glsl" "braceinit"
    assert_success
}

@test "transpiler handles mainImage without in qualifier" {
    cat > "${BATS_TEST_TMPDIR}/src/noin.glsl" <<'SHADER'
void mainImage(out vec4 fragColor, vec2 fragCoord) {
    fragColor = texture(iChannel0, fragCoord / iResolution.xy);
}
SHADER
    run transpile_and_validate "${BATS_TEST_TMPDIR}/src/noin.glsl" "noin"
    assert_success
}

@test "baseline regression: at least 120 shaders pass" {
    run bash "$VALIDATOR" --baseline 120
    assert_success
}

@test "playlist and registry are consistent" {
    run bash "$CONSISTENCY"
    assert_success
    assert_output --partial "ok: all consistent"
}
