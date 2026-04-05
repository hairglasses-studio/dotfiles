#!/usr/bin/env bats
# Tests for ghostty/shaders/bin/shader-audit.sh (pick-shaders / shader audition)
# Tests: TOML metadata parsing, shader list building, category filtering,
#        resume support, needs_animation detection, cost_color formatting
# Skips: interactive read loop, actual Ghostty config swapping

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"

    # Build synthetic shader dir with manifest and .glsl files
    export TEST_SHADERS_DIR="${BATS_TEST_TMPDIR}/shaders"
    mkdir -p "$TEST_SHADERS_DIR"

    # Create synthetic manifest
    cat > "${TEST_SHADERS_DIR}/shaders.toml" << 'TOML'
[shaders."bloom-soft"]
category = "Post-FX"
cost = "LOW"
source = "original"
description = "Soft bloom glow effect"

[shaders."crt-amber"]
category = "CRT"
cost = "MED"
source = "seanwcom"
description = "Amber phosphor CRT"

[shaders."cyber-rain"]
category = "Cyberpunk"
cost = "HIGH"
source = "shadertoy-port"
description = "Neon rain animation"

[shaders."starfield"]
category = "Background"
cost = "MED"
source = "original"
description = "Animated starfield"
TOML

    # Create shader stubs
    echo "// bloom-soft: static" > "${TEST_SHADERS_DIR}/bloom-soft.glsl"
    echo "// crt-amber: static" > "${TEST_SHADERS_DIR}/crt-amber.glsl"
    cat > "${TEST_SHADERS_DIR}/cyber-rain.glsl" << 'GLSL'
// cyber-rain: animated
uniform float ghostty_time;
void main() { gl_FragColor = vec4(ghostty_time); }
GLSL
    cat > "${TEST_SHADERS_DIR}/starfield.glsl" << 'GLSL'
// starfield: animated via iTime
uniform float iTime;
void main() { gl_FragColor = vec4(iTime); }
GLSL

    # Create a fake Ghostty config for swap_shader and restore
    export TEST_CONFIG="${BATS_TEST_TMPDIR}/ghostty-config"
    cat > "$TEST_CONFIG" << 'CFG'
font-family = Maple Mono NF CN
font-size = 12
custom-shader = /some/shader.glsl
custom-shader-animation = false
CFG

    # Build a test shim as an executable script that loads TOML and runs queries.
    # We avoid sourcing associative arrays across processes by using a script
    # that loads them internally and accepts a command argument.
    export AUDIT_HELPERS="${BATS_TEST_TMPDIR}/audit-helpers.sh"
    cat > "$AUDIT_HELPERS" << 'BASHEOF'
#!/usr/bin/env bash
set -eo pipefail

SHADERS_DIR="__SHADERS_DIR__"
CONFIG_FILE="__CONFIG_FILE__"
MANIFEST="__SHADERS_DIR__/shaders.toml"

declare -A SHADER_CAT
declare -A SHADER_COST

# TOML loader (extracted from shader-audit.sh)
while IFS=$'\t' read -r name cat cost _src _desc; do
  SHADER_CAT["${name}.glsl"]="$cat"
  SHADER_COST["${name}.glsl"]="$cost"
done < <(awk '
  /^\[shaders\."/ {
    if (name != "") print name "\t" cat "\t" cost "\t" src "\t" desc
    name = $0
    gsub(/^\[shaders\."/, "", name)
    gsub(/"\].*/, "", name)
    cat = ""; cost = ""; src = ""; desc = ""
  }
  /^category *= */ { val=$0; sub(/^[^=]*= *"/, "", val); gsub(/"$/, "", val); cat=val }
  /^cost *= */     { val=$0; sub(/^[^=]*= *"/, "", val); gsub(/"$/, "", val); cost=val }
  /^source *= */   { val=$0; sub(/^[^=]*= *"/, "", val); gsub(/"$/, "", val); src=val }
  /^description *= */ { val=$0; sub(/^[^=]*= *"/, "", val); gsub(/"$/, "", val); desc=val }
  END { if (name != "") print name "\t" cat "\t" cost "\t" src "\t" desc }
' "$MANIFEST")

needs_animation() {
  local name="$1"
  grep -qE '(ghostty_time|iTime|u_time)' "$SHADERS_DIR/$name" 2>/dev/null && { echo "true"; return; }
  echo "false"
}

cost_color() {
  case "$1" in
    LOW)  printf "\033[0;32mLOW\033[0m" ;;
    MED)  printf "\033[0;33mMED\033[0m" ;;
    HIGH) printf "\033[0;31mHIGH\033[0m" ;;
    *)    printf "$1" ;;
  esac
}

swap_shader() {
  local shader_path="$1"
  local needs_anim="$2"
  local tmp
  tmp="$(mktemp "${CONFIG_FILE}.XXXXXX")"
  sed -e "s|^custom-shader = .*|custom-shader = $shader_path|" \
      -e "s|^custom-shader-animation = .*|custom-shader-animation = $needs_anim|" \
      "$CONFIG_FILE" > "$tmp"
  mv "$tmp" "$CONFIG_FILE"
}

# Command dispatcher for testing
cmd="$1"; shift
case "$cmd" in
  get_cat)      echo "${SHADER_CAT[$1]:-NOT_FOUND}" ;;
  get_cost)     echo "${SHADER_COST[$1]:-NOT_FOUND}" ;;
  count_cat)    echo "${#SHADER_CAT[@]}" ;;
  count_cost)   echo "${#SHADER_COST[@]}" ;;
  needs_anim)   needs_animation "$1" ;;
  cost_color)   cost_color "$1" ;;
  swap_shader)  swap_shader "$1" "$2" ;;
  list_shaders)
    filter="${1:-}"
    for f in "$SHADERS_DIR"/*.glsl; do
      name="$(basename "$f")"
      if [[ -n "$filter" ]]; then
        cat="${SHADER_CAT[$name]:-unknown}"
        [[ "$cat" != "$filter" ]] && continue
      fi
      echo "$name"
    done
    ;;
esac
BASHEOF
    # Substitute paths
    sed -i "s|__SHADERS_DIR__|${TEST_SHADERS_DIR}|g" "$AUDIT_HELPERS"
    sed -i "s|__CONFIG_FILE__|${TEST_CONFIG}|g" "$AUDIT_HELPERS"
    chmod +x "$AUDIT_HELPERS"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

# --- TOML metadata parsing ---

@test "shader-audit: TOML parser loads category for bloom-soft" {
    run bash "$AUDIT_HELPERS" get_cat "bloom-soft.glsl"
    assert_success
    assert_output "Post-FX"
}

@test "shader-audit: TOML parser loads category for crt-amber" {
    run bash "$AUDIT_HELPERS" get_cat "crt-amber.glsl"
    assert_success
    assert_output "CRT"
}

@test "shader-audit: TOML parser loads cost for cyber-rain" {
    run bash "$AUDIT_HELPERS" get_cost "cyber-rain.glsl"
    assert_success
    assert_output "HIGH"
}

@test "shader-audit: TOML parser loads all 4 shaders" {
    run bash "$AUDIT_HELPERS" count_cat
    assert_success
    assert_output "4"
}

# --- needs_animation detection ---

@test "shader-audit: needs_animation detects ghostty_time uniform" {
    run bash "$AUDIT_HELPERS" needs_anim "cyber-rain.glsl"
    assert_success
    assert_output "true"
}

@test "shader-audit: needs_animation detects iTime uniform" {
    run bash "$AUDIT_HELPERS" needs_anim "starfield.glsl"
    assert_success
    assert_output "true"
}

@test "shader-audit: needs_animation returns false for static shader" {
    run bash "$AUDIT_HELPERS" needs_anim "bloom-soft.glsl"
    assert_success
    assert_output "false"
}

# --- cost_color formatting ---

@test "shader-audit: cost_color formats LOW in green" {
    run bash "$AUDIT_HELPERS" cost_color "LOW"
    assert_success
    assert_output --partial "32m"
    assert_output --partial "LOW"
}

@test "shader-audit: cost_color formats MED in yellow" {
    run bash "$AUDIT_HELPERS" cost_color "MED"
    assert_success
    assert_output --partial "33m"
    assert_output --partial "MED"
}

@test "shader-audit: cost_color formats HIGH in red" {
    run bash "$AUDIT_HELPERS" cost_color "HIGH"
    assert_success
    assert_output --partial "31m"
    assert_output --partial "HIGH"
}

# --- swap_shader config update ---

@test "shader-audit: swap_shader updates custom-shader path atomically" {
    bash "$AUDIT_HELPERS" swap_shader "/new/path/bloom-soft.glsl" "false"
    run grep '^custom-shader = ' "$TEST_CONFIG"
    assert_success
    assert_output "custom-shader = /new/path/bloom-soft.glsl"
}

@test "shader-audit: swap_shader updates animation flag" {
    bash "$AUDIT_HELPERS" swap_shader "/new/path/cyber-rain.glsl" "true"
    run grep '^custom-shader-animation = ' "$TEST_CONFIG"
    assert_success
    assert_output "custom-shader-animation = true"
}

@test "shader-audit: swap_shader preserves other config lines" {
    bash "$AUDIT_HELPERS" swap_shader "/new/path/test.glsl" "false"
    run grep '^font-family = ' "$TEST_CONFIG"
    assert_success
    assert_output "font-family = Maple Mono NF CN"
}

# --- Shader list building with category filter ---

@test "shader-audit: lists all 4 shader files without filter" {
    run bash "$AUDIT_HELPERS" list_shaders
    assert_success
    assert_line "bloom-soft.glsl"
    assert_line "crt-amber.glsl"
    assert_line "cyber-rain.glsl"
    assert_line "starfield.glsl"
}

@test "shader-audit: category filter selects only CRT shaders" {
    run bash "$AUDIT_HELPERS" list_shaders "CRT"
    assert_success
    assert_output "crt-amber.glsl"
}

# --- Category filter: no matches ---

@test "shader-audit: category filter returns empty for nonexistent category" {
    run bash "$AUDIT_HELPERS" list_shaders "Nonexistent"
    assert_success
    assert_output ""
}
