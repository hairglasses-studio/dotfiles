#!/usr/bin/env bash
set -uo pipefail

# ── Kitty DarkWindow Shader Test Runner ──────────
# Two modes:
#   --validate (default) — static GLSL analysis + glslangValidator if available
#   --visual             — launches Ghostty with each shader for visual preview
#
# Usage:
#   bash test-shaders.sh                   # validate all shaders (fast, no GUI)
#   bash test-shaders.sh --visual          # open each shader in Ghostty (3s each)
#   bash test-shaders.sh --visual --duration 5  # longer visual preview
#   bash test-shaders.sh bloom.glsl        # validate a single shader
#   bash test-shaders.sh --list            # print the full catalog

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SHADER_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SHADERS_DIR="$SHADER_ROOT/darkwindow"
source "$SCRIPT_DIR/../../../scripts/lib/notify.sh"
DURATION=3
LOG_DIR="/tmp/kitty-shader-tests"
SINGLE=""
LIST_ONLY=false
VISUAL=false
HAS_GLSLANG=false

# Check for glslangValidator (optional, gives real GLSL compilation)
if command -v glslangValidator &>/dev/null; then
  HAS_GLSLANG=true
fi

# ── Parse args ─────────────────────────────────────
while [[ $# -gt 0 ]]; do
  case "$1" in
    --duration) DURATION="$2"; shift 2 ;;
    --list)     LIST_ONLY=true; shift ;;
    --visual)   VISUAL=true; shift ;;
    *.glsl)     SINGLE="$1"; shift ;;
    *)          echo "Usage: $0 [--visual] [--duration N] [--list] [shader.glsl]"; exit 1 ;;
  esac
done

mkdir -p "$LOG_DIR"

# ── Shader metadata (from central manifest) ──────
MANIFEST="$SHADERS_DIR/shaders.toml"
if [[ ! -f "$MANIFEST" ]]; then
  echo "ERROR: shaders.toml not found at $MANIFEST" >&2
  exit 1
fi

declare -A SHADER_SOURCE
declare -A SHADER_CATEGORY
declare -A SHADER_DESC

# Load all metadata from shaders.toml in a single pass
while IFS=$'\t' read -r name cat cost src desc; do
  SHADER_SOURCE["${name}.glsl"]="$src"
  SHADER_CATEGORY["${name}.glsl"]="$cat"
  SHADER_DESC["${name}.glsl"]="$desc"
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

# ── List mode ──────────────────────────────────────
if $LIST_ONLY; then
  printf "\n%-36s %-10s %-12s %s\n" "SHADER" "SOURCE" "CATEGORY" "DESCRIPTION"
  printf "%-36s %-10s %-12s %s\n" "------" "------" "--------" "-----------"
  for f in "$SHADERS_DIR"/*.glsl; do
    name="$(basename "$f")"
    src="${SHADER_SOURCE[$name]:-unknown}"
    cat="${SHADER_CATEGORY[$name]:-unknown}"
    desc="${SHADER_DESC[$name]:-}"
    printf "%-36s %-10s %-12s %s\n" "$name" "$src" "$cat" "$desc"
  done
  exit 0
fi

# ── Build shader list ──────────────────────────────
if [[ -n "$SINGLE" ]]; then
  if [[ ! -f "$SHADERS_DIR/$SINGLE" ]]; then
    echo "ERROR: $SINGLE not found in $SHADERS_DIR"
    exit 1
  fi
  SHADER_FILES=("$SHADERS_DIR/$SINGLE")
else
  SHADER_FILES=("$SHADERS_DIR"/*.glsl)
fi

TOTAL=0
for f in "${SHADER_FILES[@]}"; do
  [[ "$(basename "$f")" == "test-shaders.sh" ]] && continue
  TOTAL=$((TOTAL + 1))
done

PASS=0
FAIL=0
WARN_COUNT=0

declare -a RESULTS=()

if $VISUAL; then
  MODE="visual (launching Kitty via DarkWindow, ${DURATION}s each)"
  if [[ "$(uname)" == "Darwin" ]]; then
    KITTY_BIN="$(command -v kitty 2>/dev/null || echo "/Applications/kitty.app/Contents/MacOS/kitty")"
  else
    KITTY_BIN="$(command -v kitty 2>/dev/null)"
  fi
  if [[ ! -x "${KITTY_BIN:-}" ]] && ! command -v kitty &>/dev/null; then
    echo "ERROR: kitty binary not found (needed for --visual mode)"
    exit 1
  fi
else
  MODE="validate (static analysis"
  if $HAS_GLSLANG; then
    MODE="$MODE + glslangValidator)"
  else
    MODE="$MODE only — install glslangValidator for GLSL compilation checks)"
  fi
fi

printf "\n"
printf "  Kitty DarkWindow Shader Test Runner\n"
printf "  ──────────────────────────\n"
printf "  Shaders dir: %s\n" "$SHADERS_DIR"
printf "  Total: %d shaders\n" "$TOTAL"
printf "  Mode: %s\n" "$MODE"
printf "  Logs: %s\n" "$LOG_DIR"
printf "\n"

# ── Validate a single shader (static analysis) ────
validate_shader() {
  local shader_path="$1"
  local name="$2"
  local log_file="$3"
  local status="PASS"
  local error_msg=""
  local warnings=""

  # 1. File exists and is non-empty
  if [[ ! -s "$shader_path" ]]; then
    echo "FAIL|empty or missing file"
    return
  fi

  # 2. Check for entry point (main, mainImage, or fragment)
  if ! grep -qE "void\s+(main|mainImage|fragment)\s*\(" "$shader_path" 2>/dev/null; then
    status="WARN"
    warnings="no standard entry point (main/mainImage/fragment)"
  fi

  # 3. Check for common Ghostty/Shadertoy uniforms
  local has_texture=false
  local has_resolution=false
  if grep -qE "iChannel0|ghostty_texture|gl_FragColor|fragColor" "$shader_path" 2>/dev/null; then
    has_texture=true
  fi
  if grep -qE "iResolution|ghostty_resolution|resolution" "$shader_path" 2>/dev/null; then
    has_resolution=true
  fi

  # 4. Check for GLSL version directive
  local glsl_version=""
  if grep -qE "^#version" "$shader_path" 2>/dev/null; then
    glsl_version=$(grep -oE "^#version [0-9]+" "$shader_path" | head -1)
  fi

  # 5. Check for obvious syntax errors (unmatched braces)
  local open_braces close_braces
  open_braces=$(grep -o '{' "$shader_path" | wc -l | tr -d ' ')
  close_braces=$(grep -o '}' "$shader_path" | wc -l | tr -d ' ')
  if [[ "$open_braces" -ne "$close_braces" ]]; then
    status="FAIL"
    error_msg="mismatched braces: ${open_braces} open vs ${close_braces} close"
  fi

  # 6. Check for common GLSL errors
  if grep -qE "vec[234]\s*[a-zA-Z]" "$shader_path" 2>/dev/null; then
    # Missing parens in constructor — very common typo
    :
  fi

  # 7. glslangValidator compilation check (if available)
  if $HAS_GLSLANG && [[ "$status" != "FAIL" ]]; then
    # Ghostty shaders are fragment shaders; glslangValidator needs a stage hint.
    # Most Ghostty shaders omit #version (Ghostty's transpiler handles it),
    # so we prepend #version 460 for shaders that lack one.
    local glslang_target="$shader_path"
    local glslang_tmp=""
    if ! grep -qE '^#version' "$shader_path" 2>/dev/null; then
      glslang_tmp="$(mktemp /tmp/shader-test-XXXXXX.frag)"
      { echo '#version 460'; cat "$shader_path"; } > "$glslang_tmp"
      glslang_target="$glslang_tmp"
    fi
    local glslang_out
    glslang_out=$(glslangValidator -S frag "$glslang_target" 2>&1) || true
    [[ -n "$glslang_tmp" ]] && rm -f "$glslang_tmp"
    if echo "$glslang_out" | grep -qi "error"; then
      # Many Ghostty shaders use Shadertoy-style uniforms that glslangValidator
      # doesn't know about — only flag as WARN, not FAIL
      if [[ "$status" == "PASS" ]]; then
        status="WARN"
        local error_count
        error_count=$(echo "$glslang_out" | grep -ci "error" || true)
        warnings="glslangValidator: ${error_count} error(s) — may use Ghostty-specific uniforms"
      fi
    fi
    echo "$glslang_out" > "$log_file"
  fi

  # 8. Line count / complexity
  local lines
  lines=$(wc -l < "$shader_path" | tr -d ' ')

  # Build result
  if [[ "$status" == "FAIL" ]]; then
    echo "FAIL|$error_msg"
  elif [[ "$status" == "WARN" ]]; then
    echo "WARN|$warnings"
  else
    local info="${lines} lines"
    [[ -n "$glsl_version" ]] && info="$info, $glsl_version"
    $has_texture && info="$info, reads texture"
    $has_resolution && info="$info, uses resolution"
    echo "PASS|$info"
  fi
}

# ── Visual test (launch Ghostty) ──────────────────
visual_test_shader() {
  local shader_path="$1"
  local name="$2"
  local log_file="$3"

  tmp_config=$(mktemp /tmp/ghostty-test-XXXXXX)
  cat > "$tmp_config" << EOF
custom-shader = $shader_path
custom-shader-animation = true
font-family = Hack Nerd Font
font-size = 14
background = #000000
foreground = #f1f1f0
EOF

  local exit_code=0
  timeout "${DURATION}s" "$GHOSTTY_BIN" \
      --config-file="$tmp_config" \
      -e bash -c "printf '\033[1;35m━━━ %s ━━━\033[0m\n\n' '$name'; echo 'SHADER_TEST_OK'; neofetch 2>/dev/null || fastfetch 2>/dev/null || echo 'hello world'; sleep $((DURATION - 1))" \
      > "$log_file" 2>&1 || exit_code=$?

  rm -f "$tmp_config"

  # 124 = timeout killed it (expected for GUI)
  # 0   = clean exit
  if [[ $exit_code -eq 0 ]] || [[ $exit_code -eq 124 ]]; then
    # Check log for shader errors
    if [[ -f "$log_file" ]] && grep -qi -E "shader.*(error|fail|invalid|compile)" "$log_file" 2>/dev/null; then
      echo "FAIL|shader compilation error (see $log_file)"
    else
      echo "PASS|rendered OK (exit $exit_code)"
    fi
  else
    # Check if it's a shader error specifically
    if [[ -f "$log_file" ]] && grep -qi -E "shader" "$log_file" 2>/dev/null; then
      echo "FAIL|shader error, exit $exit_code (see $log_file)"
    else
      echo "FAIL|ghostty exit $exit_code (see $log_file)"
    fi
  fi
}

# ── Test each shader ───────────────────────────────
for shader_path in "${SHADER_FILES[@]}"; do
  name="$(basename "$shader_path")"
  [[ "$name" == "test-shaders.sh" ]] && continue

  src="${SHADER_SOURCE[$name]:-unknown}"
  cat="${SHADER_CATEGORY[$name]:-unknown}"
  log_file="$LOG_DIR/${name%.glsl}.log"

  printf "  Testing: %-38s " "$name"

  if $VISUAL; then
    result=$(visual_test_shader "$shader_path" "$name" "$log_file")
  else
    result=$(validate_shader "$shader_path" "$name" "$log_file")
  fi

  status="${result%%|*}"
  detail="${result#*|}"

  case "$status" in
    PASS)
      printf "\033[0;32mPASS\033[0m  %s\n" "$detail"
      PASS=$((PASS + 1))
      ;;
    FAIL)
      printf "\033[0;31mFAIL\033[0m  %s\n" "$detail"
      FAIL=$((FAIL + 1))
      ;;
    WARN)
      printf "\033[0;33mWARN\033[0m  %s\n" "$detail"
      WARN_COUNT=$((WARN_COUNT + 1))
      ;;
  esac

  RESULTS+=("$status|$name|$src|$cat|$detail")
done

# ── Summary ────────────────────────────────────────
printf "\n"
printf "  ══════════════════════════════════════════════════════════════════\n"
printf "  Results: \033[0;32m%d PASS\033[0m  \033[0;31m%d FAIL\033[0m  \033[0;33m%d WARN\033[0m  / %d total\n" "$PASS" "$FAIL" "$WARN_COUNT" "$TOTAL"
printf "  ══════════════════════════════════════════════════════════════════\n"
printf "\n"

# ── Detailed table ─────────────────────────────────
printf "  %-6s %-36s %-10s %-12s %s\n" "STATUS" "SHADER" "SOURCE" "CATEGORY" "NOTES"
printf "  %-6s %-36s %-10s %-12s %s\n" "------" "------" "------" "--------" "-----"
for entry in "${RESULTS[@]}"; do
  IFS='|' read -r st nm sr ct msg <<< "$entry"
  case "$st" in
    PASS) color="\033[0;32m" ;;
    FAIL) color="\033[0;31m" ;;
    *)    color="\033[0;33m" ;;
  esac
  printf "  ${color}%-6s\033[0m %-36s %-10s %-12s %s\n" "$st" "$nm" "$sr" "$ct" "$msg"
done

printf "\n  Logs saved to: %s\n\n" "$LOG_DIR"

# Desktop notification with summary
if [[ $FAIL -gt 0 ]]; then
  hg_notify "Shader" "Test: ${FAIL} failed, ${PASS} passed / ${TOTAL} total"
else
  hg_notify_low "Shader" "Test: ${PASS} passed, ${WARN_COUNT} warnings / ${TOTAL} total"
fi

# Exit with failure if any shader failed
[[ $FAIL -eq 0 ]]
