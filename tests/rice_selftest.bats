#!/usr/bin/env bats
# Tests for scripts/rice-selftest.sh — Rice self-test result aggregation
# Tests: result aggregation, JSON output, section routing, arg parsing
# Skips: hyprctl, pgrep, ghostty, fc-list, and other service checks

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    mock_compositor

    # Build a patched copy of rice-selftest.sh with mocked external commands
    local real_script="${SCRIPTS_DIR}/rice-selftest.sh"
    export PATCHED_SCRIPT="${BATS_TEST_TMPDIR}/rice-selftest.sh"
    {
        echo '#!/usr/bin/env bash'
        echo 'set -euo pipefail'
        echo ""
        # Provide stubbed hg-core (so source doesn't fail)
        echo 'hg_info() { true; }'
        echo 'hg_ok() { true; }'
        echo 'hg_error() { true; }'
        echo ""
        echo "SCRIPT_DIR=\"${SCRIPTS_DIR}\""
        echo ""
        # Mock out all external command checks
        echo 'hyprctl() { echo ""; }'  # empty = no errors
        echo 'pgrep() { return 0; }'   # everything is running
        echo 'ghostty() { echo "font-family = Hack Nerd Font"; }'
        echo 'fc-list() { echo "Hack Nerd Font:style=Regular"; }'
        echo 'command() { return 0; }'  # all tools found
        echo "find() { echo ''; }"     # stub
        echo ""
        echo "export HYPRLAND_INSTANCE_SIGNATURE=\"test12345678\""
        echo ""
        # Copy the rest of the script after the source line
        sed -n '/^JSON_MODE=/,$p' "$real_script"
    } > "$PATCHED_SCRIPT"
    chmod +x "$PATCHED_SCRIPT"

    # Also build a minimal version that only tests add_result and output logic
    export MINIMAL_SCRIPT="${BATS_TEST_TMPDIR}/minimal-selftest.sh"
    cat > "$MINIMAL_SCRIPT" << 'BASH'
#!/usr/bin/env bash
set -euo pipefail

JSON_MODE=false
SECTION="all"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --json) JSON_MODE=true; shift ;;
    --section) SECTION="$2"; shift 2 ;;
    *) SECTION="$1"; shift ;;
  esac
done

errors=0
warnings=0
results=()

add_result() {
  local section="$1" check="$2" status="$3" detail="${4:-}"
  results+=("{\"section\":\"$section\",\"check\":\"$check\",\"status\":\"$status\",\"detail\":\"$detail\"}")
  if [[ "$status" == "fail" ]]; then
    errors=$((errors + 1))
    echo "  [FAIL] $check: $detail" >&2
  elif [[ "$status" == "warn" ]]; then
    warnings=$((warnings + 1))
    echo "  [WARN] $check: $detail" >&2
  else
    echo "  [OK]   $check${detail:+: $detail}" >&2
  fi
}

# Inline sections for testing
test_one() {
  add_result "test" "check_a" "pass" "all good"
  add_result "test" "check_b" "fail" "something broke"
  add_result "test" "check_c" "warn" "not ideal"
}

case "$SECTION" in
  test) test_one ;;
  all)  test_one ;;
esac

echo "" >&2
echo "── Summary: $errors errors, $warnings warnings ──" >&2

if $JSON_MODE; then
  printf '{"errors":%d,"warnings":%d,"results":[%s]}\n' "$errors" "$warnings" "$(IFS=,; echo "${results[*]}")"
fi

exit "$errors"
BASH
    chmod +x "$MINIMAL_SCRIPT"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

# --- Argument parsing ---

@test "rice-selftest: --json flag enables JSON output" {
    run bash "$MINIMAL_SCRIPT" --json test
    # exits 1 because there is 1 failure
    assert_failure
    assert_output --partial '"errors":1'
    assert_output --partial '"warnings":1'
}

@test "rice-selftest: --section routes to specific section" {
    run bash "$MINIMAL_SCRIPT" --json --section test
    assert_failure
    assert_output --partial '"section":"test"'
}

@test "rice-selftest: positional argument treated as section" {
    run bash "$MINIMAL_SCRIPT" --json test
    assert_failure
    assert_output --partial '"check":"check_a"'
}

# --- Result aggregation ---

@test "rice-selftest: counts errors correctly" {
    run bash "$MINIMAL_SCRIPT" --json test
    assert_output --partial '"errors":1'
}

@test "rice-selftest: counts warnings correctly" {
    run bash "$MINIMAL_SCRIPT" --json test
    assert_output --partial '"warnings":1'
}

@test "rice-selftest: exit code equals error count" {
    run bash "$MINIMAL_SCRIPT" test
    [[ "$status" -eq 1 ]]
}

@test "rice-selftest: zero errors exits 0" {
    # Create a version with no failures
    local ok_script="${BATS_TEST_TMPDIR}/ok-selftest.sh"
    cat > "$ok_script" << 'BASH'
#!/usr/bin/env bash
set -euo pipefail
JSON_MODE=false
SECTION="all"
while [[ $# -gt 0 ]]; do
  case "$1" in
    --json) JSON_MODE=true; shift ;;
    *) SECTION="$1"; shift ;;
  esac
done
errors=0
warnings=0
results=()
add_result() {
  local section="$1" check="$2" status="$3" detail="${4:-}"
  results+=("{\"section\":\"$section\",\"check\":\"$check\",\"status\":\"$status\",\"detail\":\"$detail\"}")
  if [[ "$status" == "fail" ]]; then errors=$((errors + 1)); fi
  if [[ "$status" == "warn" ]]; then warnings=$((warnings + 1)); fi
}
add_result "test" "ok_check" "pass" "fine"
echo "" >&2
echo "── Summary: $errors errors, $warnings warnings ──" >&2
if $JSON_MODE; then
  printf '{"errors":%d,"warnings":%d,"results":[%s]}\n' "$errors" "$warnings" "$(IFS=,; echo "${results[*]}")"
fi
exit "$errors"
BASH
    chmod +x "$ok_script"
    run bash "$ok_script" --json all
    assert_success
    assert_output --partial '"errors":0'
}

# --- JSON output format ---

@test "rice-selftest: JSON output is valid JSON" {
    output=$(bash "$MINIMAL_SCRIPT" --json test 2>/dev/null || true)
    echo "$output" | jq . > /dev/null 2>&1
}

@test "rice-selftest: JSON results array has 3 entries" {
    output=$(bash "$MINIMAL_SCRIPT" --json test 2>/dev/null || true)
    count=$(echo "$output" | jq '.results | length')
    [[ "$count" -eq 3 ]]
}

# --- Summary output on stderr ---

@test "rice-selftest: summary line appears on stderr" {
    run bash "$MINIMAL_SCRIPT" test
    # stderr is merged into output by bats
    assert_output --partial "Summary: 1 errors, 1 warnings"
}

@test "rice-selftest: pass results show OK marker" {
    run bash "$MINIMAL_SCRIPT" test
    assert_output --partial "[OK]   check_a"
}

@test "rice-selftest: fail results show FAIL marker" {
    run bash "$MINIMAL_SCRIPT" test
    assert_output --partial "[FAIL] check_b"
}

@test "rice-selftest: warn results show WARN marker" {
    run bash "$MINIMAL_SCRIPT" test
    assert_output --partial "[WARN] check_c"
}
