#!/usr/bin/env bats
# Tests for rg-status-bar.sh — Fleet status bar cache (ralphglasses distro)
# Tests: model abbreviation, segment formatting, cache freshness, help/error
# Skips: i3blocks/waybar integration, actual MCP calls

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"

    # The real script lives at the symlink target in ralphglasses
    local real_script
    real_script="$(readlink -f "${SCRIPTS_DIR}/rg-status-bar.sh" 2>/dev/null || true)"
    if [[ -z "$real_script" ]]; then
        real_script="${HG_STUDIO_ROOT:-$HOME/hairglasses-studio}/ralphglasses/distro/scripts/rg-status-bar.sh"
    fi
    [[ -f "$real_script" ]] || skip "rg-status-bar source script unavailable"

    # Create a patched copy that uses our test paths and mocked commands
    export PATCHED_SCRIPT="${BATS_TEST_TMPDIR}/rg-status-bar.sh"
    {
        echo '#!/usr/bin/env bash'
        echo 'set -euo pipefail'
        echo ""
        echo "RG_BIN=\"${BATS_TEST_TMPDIR}/mock-rg\""
        echo "CACHE_FILE=\"${BATS_TEST_TMPDIR}/rg-status.json\""
        echo "CACHE_MAX_AGE=90"
        # Copy everything from COLOR_GREEN onward (skip the original constants block)
        sed -n '/^COLOR_GREEN=/,$p' "$real_script"
    } > "$PATCHED_SCRIPT"
    chmod +x "$PATCHED_SCRIPT"

    # Create mock ralphglasses binary that returns nothing
    cat > "${BATS_TEST_TMPDIR}/mock-rg" << 'MOCK'
#!/usr/bin/env bash
exit 1
MOCK
    chmod +x "${BATS_TEST_TMPDIR}/mock-rg"

    # Build a synthetic fleet cache for segment reader tests
    export TEST_CACHE="${BATS_TEST_TMPDIR}/rg-status.json"
    cat > "$TEST_CACHE" << 'JSON'
{
  "fleet": { "running": 3, "completed": 5, "failed": 1, "pending": 2, "total": 11 },
  "loops": { "total_runs": 11, "completed": 5, "converge_pct": 45 },
  "cost": { "total_spend_usd": 12.34 },
  "models": [
    { "model": "claude-opus-4-6", "count": 4 },
    { "model": "claude-sonnet-4-6", "count": 3 },
    { "model": "o4-mini", "count": 2 }
  ],
  "repos": { "scanned": 76, "targeted": 12 },
  "iters": { "total": 87, "avg_per_run": 7.9 }
}
JSON
    # Touch with current time to make cache fresh
    touch "$TEST_CACHE"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

# --- Model abbreviation (source the declare block for testing) ---

@test "rg-status-bar: abbrev_model maps claude-opus-4-6 to opus" {
    eval "$(sed -n '/^declare -A MODEL_ABBREV/,/^)/p' "$PATCHED_SCRIPT")"
    eval "$(sed -n '/^abbrev_model/,/^}/p' "$PATCHED_SCRIPT")"
    run abbrev_model "claude-opus-4-6"
    assert_success
    assert_output "opus"
}

@test "rg-status-bar: abbrev_model maps claude-sonnet-4-6 to son" {
    eval "$(sed -n '/^declare -A MODEL_ABBREV/,/^)/p' "$PATCHED_SCRIPT")"
    eval "$(sed -n '/^abbrev_model/,/^}/p' "$PATCHED_SCRIPT")"
    run abbrev_model "claude-sonnet-4-6"
    assert_success
    assert_output "son"
}

@test "rg-status-bar: abbrev_model maps o4-mini to o4m" {
    eval "$(sed -n '/^declare -A MODEL_ABBREV/,/^)/p' "$PATCHED_SCRIPT")"
    eval "$(sed -n '/^abbrev_model/,/^}/p' "$PATCHED_SCRIPT")"
    run abbrev_model "o4-mini"
    assert_success
    assert_output "o4m"
}

@test "rg-status-bar: abbrev_model truncates unknown models to 3 chars" {
    eval "$(sed -n '/^declare -A MODEL_ABBREV/,/^)/p' "$PATCHED_SCRIPT")"
    eval "$(sed -n '/^abbrev_model/,/^}/p' "$PATCHED_SCRIPT")"
    run abbrev_model "llama-3.1-70b"
    assert_success
    assert_output "lla"
}

# --- Segment reader: fleet ---

@test "rg-status-bar: segment fleet outputs running/completed/failed counts" {
    run bash "$PATCHED_SCRIPT" --segment fleet
    assert_success
    assert_output --partial "3"
    assert_output --partial "5"
    assert_output --partial "1"
}

# --- Segment reader: cost ---

@test "rg-status-bar: segment cost outputs dollar amount" {
    run bash "$PATCHED_SCRIPT" --segment cost
    assert_success
    assert_output --partial "12.34"
}

# --- Segment reader: repos ---

@test "rg-status-bar: segment repos outputs scanned and targeted counts" {
    run bash "$PATCHED_SCRIPT" --segment repos
    assert_success
    assert_output --partial "76"
    assert_output --partial "12"
}

# --- Segment reader: iters ---

@test "rg-status-bar: segment iters outputs total and average" {
    run bash "$PATCHED_SCRIPT" --segment iters
    assert_success
    assert_output --partial "87"
    assert_output --partial "7.9"
}

# --- Segment reader: loops ---

@test "rg-status-bar: segment loops outputs run count and converge percent" {
    run bash "$PATCHED_SCRIPT" --segment loops
    assert_success
    assert_output --partial "11"
    assert_output --partial "45%"
}

# --- Help ---

@test "rg-status-bar: --help shows usage" {
    run bash "$PATCHED_SCRIPT" --help
    assert_success
    assert_output --partial "Usage"
}

# --- Unknown argument ---

@test "rg-status-bar: unknown argument exits with error" {
    run bash "$PATCHED_SCRIPT" --bogus
    assert_failure
    assert_output --partial "Unknown argument"
}

# --- Unknown segment ---

@test "rg-status-bar: unknown segment name exits with error" {
    run bash "$PATCHED_SCRIPT" --segment bogus
    assert_failure
    assert_output --partial "Unknown segment"
}

# --- No arguments ---

@test "rg-status-bar: no arguments shows usage and exits non-zero" {
    run bash "$PATCHED_SCRIPT"
    assert_failure
    assert_output --partial "Usage"
}
