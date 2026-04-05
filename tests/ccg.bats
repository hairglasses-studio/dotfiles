#!/usr/bin/env bats
# Tests for scripts/ccg.sh — Global Claude Code session browser
# Tests pure helper functions; skips FZF interaction and actual session launching.

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"

    # ccg.sh is zsh; we source individual functions via a bash-compatible shim.
    # Extract the pure helper functions into a sourceable bash file.
    export CCG_HELPERS="${BATS_TEST_TMPDIR}/ccg-helpers.sh"
    cat > "$CCG_HELPERS" << 'BASH'
#!/usr/bin/env bash

# Snazzy palette ANSI codes (copied from ccg.sh)
C_RESET=$'\033[0m'
C_GREEN=$'\033[38;2;90;247;142m'
C_CYAN=$'\033[38;2;87;199;255m'
C_MAGENTA=$'\033[38;2;255;106;193m'
C_YELLOW=$'\033[38;2;243;249;157m'
C_RED=$'\033[38;2;255;92;87m'
C_DIM=$'\033[90m'
C_BOLD=$'\033[1m'
C_WHITE=$'\033[38;2;241;241;240m'

_ccg_relative_age() {
    local ts_epoch="$1"
    local now_epoch
    now_epoch=$(date +%s)
    local diff=$(( now_epoch - ts_epoch ))
    if (( diff < 60 )); then
        echo "${diff}s"
    elif (( diff < 3600 )); then
        echo "$(( diff / 60 ))m"
    elif (( diff < 86400 )); then
        echo "$(( diff / 3600 ))h"
    elif (( diff < 604800 )); then
        echo "$(( diff / 86400 ))d"
    elif (( diff < 2592000 )); then
        echo "$(( diff / 604800 ))w"
    else
        echo "$(( diff / 2592000 ))mo"
    fi
}

_ccg_age_color() {
    local ts_epoch="$1"
    local now_epoch
    now_epoch=$(date +%s)
    local diff=$(( now_epoch - ts_epoch ))
    if (( diff < 3600 )); then
        printf '%s' "$C_YELLOW"
    elif (( diff < 86400 )); then
        printf '%s' "$C_WHITE"
    else
        printf '%s' "$C_DIM"
    fi
}

_ccg_model_short() {
    case "$1" in
        *opus*)   echo "opus" ;;
        *sonnet*) echo "snnt" ;;
        *haiku*)  echo "hiku" ;;
        *)        echo "${1:0:8}" ;;
    esac
}

_ccg_model_color() {
    case "$1" in
        *opus*)   printf '%s' "$C_MAGENTA" ;;
        *sonnet*) printf '%s' "$C_CYAN" ;;
        *haiku*)  printf '%s' "$C_YELLOW" ;;
        *)        printf '%s' "$C_DIM" ;;
    esac
}

_ccg_truncate() {
    local str="$1" max="$2"
    if (( ${#str} > max )); then
        echo "${str:0:$((max-1))}~"
    else
        echo "$str"
    fi
}

_ccg_repo_from_cwd() {
    local cwd="$1"
    local studio="$HOME/hairglasses-studio/"
    if [[ "$cwd" == "$studio"* ]]; then
        local rel="${cwd#$studio}"
        echo "${rel%%/*}"
    elif [[ "$cwd" == "$HOME" ]]; then
        echo "~"
    else
        echo "$(basename "$cwd")"
    fi
}

_ccg_is_alive() {
    local pid="$1"
    [[ -n "$pid" ]] && [[ "$pid" != "0" ]] && kill -0 "$pid" 2>/dev/null
}
BASH
    chmod +x "$CCG_HELPERS"
    source "$CCG_HELPERS"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

# --- _ccg_relative_age ---

@test "ccg: relative_age returns seconds for <60s difference" {
    local now
    now=$(date +%s)
    local ts=$(( now - 30 ))
    run _ccg_relative_age "$ts"
    assert_success
    assert_output "30s"
}

@test "ccg: relative_age returns minutes for <1h difference" {
    local now
    now=$(date +%s)
    local ts=$(( now - 300 ))
    run _ccg_relative_age "$ts"
    assert_success
    assert_output "5m"
}

@test "ccg: relative_age returns hours for <1d difference" {
    local now
    now=$(date +%s)
    local ts=$(( now - 7200 ))
    run _ccg_relative_age "$ts"
    assert_success
    assert_output "2h"
}

@test "ccg: relative_age returns days for <1w difference" {
    local now
    now=$(date +%s)
    local ts=$(( now - 259200 ))
    run _ccg_relative_age "$ts"
    assert_success
    assert_output "3d"
}

@test "ccg: relative_age returns weeks for <30d difference" {
    local now
    now=$(date +%s)
    local ts=$(( now - 1209600 ))
    run _ccg_relative_age "$ts"
    assert_success
    assert_output "2w"
}

@test "ccg: relative_age returns months for >=30d difference" {
    local now
    now=$(date +%s)
    local ts=$(( now - 5184000 ))
    run _ccg_relative_age "$ts"
    assert_success
    assert_output "2mo"
}

# --- _ccg_model_short ---

@test "ccg: model_short abbreviates opus correctly" {
    run _ccg_model_short "claude-opus-4-6"
    assert_success
    assert_output "opus"
}

@test "ccg: model_short abbreviates sonnet correctly" {
    run _ccg_model_short "claude-sonnet-4-6"
    assert_success
    assert_output "snnt"
}

@test "ccg: model_short abbreviates haiku correctly" {
    run _ccg_model_short "claude-3-haiku"
    assert_success
    assert_output "hiku"
}

@test "ccg: model_short truncates unknown models to 8 chars" {
    run _ccg_model_short "gpt-4o-2025-01"
    assert_success
    assert_output "gpt-4o-2"
}

# --- _ccg_model_color ---

@test "ccg: model_color returns magenta for opus" {
    run _ccg_model_color "claude-opus-4-6"
    assert_success
    assert_output --partial "255;106;193"
}

@test "ccg: model_color returns cyan for sonnet" {
    run _ccg_model_color "claude-sonnet-4-6"
    assert_success
    assert_output --partial "87;199;255"
}

@test "ccg: model_color returns yellow for haiku" {
    run _ccg_model_color "claude-3-haiku"
    assert_success
    assert_output --partial "249;157"
}

# --- _ccg_truncate ---

@test "ccg: truncate returns string as-is when under limit" {
    run _ccg_truncate "hello" 10
    assert_success
    assert_output "hello"
}

@test "ccg: truncate shortens string at limit with tilde" {
    run _ccg_truncate "hello-world-foo" 10
    assert_success
    assert_output "hello-wor~"
}

@test "ccg: truncate handles exact-length string" {
    run _ccg_truncate "12345" 5
    assert_success
    assert_output "12345"
}

# --- _ccg_repo_from_cwd ---

@test "ccg: repo_from_cwd extracts repo name from studio path" {
    run _ccg_repo_from_cwd "$HOME/hairglasses-studio/mcpkit"
    assert_success
    assert_output "mcpkit"
}

@test "ccg: repo_from_cwd extracts repo from nested path" {
    run _ccg_repo_from_cwd "$HOME/hairglasses-studio/dotfiles/scripts"
    assert_success
    assert_output "dotfiles"
}

@test "ccg: repo_from_cwd returns tilde for home directory" {
    run _ccg_repo_from_cwd "$HOME"
    assert_success
    assert_output "~"
}

@test "ccg: repo_from_cwd returns basename for non-studio path" {
    run _ccg_repo_from_cwd "/tmp/some-project"
    assert_success
    assert_output "some-project"
}

# --- _ccg_is_alive ---

@test "ccg: is_alive returns false for pid 0" {
    run _ccg_is_alive "0"
    assert_failure
}

@test "ccg: is_alive returns false for empty pid" {
    run _ccg_is_alive ""
    assert_failure
}

@test "ccg: is_alive returns true for own process" {
    run _ccg_is_alive "$$"
    assert_success
}

# --- _ccg_age_color ---

@test "ccg: age_color returns yellow for recent (<1h)" {
    local now
    now=$(date +%s)
    local ts=$(( now - 600 ))
    run _ccg_age_color "$ts"
    assert_success
    assert_output --partial "249;157"
}

@test "ccg: age_color returns white for today (<24h)" {
    local now
    now=$(date +%s)
    local ts=$(( now - 43200 ))
    run _ccg_age_color "$ts"
    assert_success
    assert_output --partial "241;241;240"
}

@test "ccg: age_color returns dim for old (>24h)" {
    local now
    now=$(date +%s)
    local ts=$(( now - 172800 ))
    run _ccg_age_color "$ts"
    assert_success
    assert_output --partial "90m"
}
