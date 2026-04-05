#!/usr/bin/env bats
# Tests for scripts/lib/config.sh

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    mock_compositor
    mock_notify_send
    # Override HOME so config_backup doesn't write to real home
    export HOME="${BATS_TEST_TMPDIR}/fakehome"
    mkdir -p "${HOME}"
    source "${LIB_DIR}/config.sh"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

# --- Atomic write tests ---

@test "config_atomic_write: creates file with correct content" {
    local target="${BATS_TEST_TMPDIR}/test-config"
    config_atomic_write "${target}" "line1
line2
line3"
    assert [ -f "${target}" ]
    run cat "${target}"
    assert_output "line1
line2
line3"
}

@test "config_atomic_write: overwrites existing file" {
    local target="${BATS_TEST_TMPDIR}/test-config"
    echo "original content" > "${target}"
    config_atomic_write "${target}" "new content"
    run cat "${target}"
    assert_output "new content"
}

@test "config_atomic_write: no leftover temp files on success" {
    local target="${BATS_TEST_TMPDIR}/test-config"
    config_atomic_write "${target}" "clean write"
    # Count temp files matching the pattern
    local temps
    temps=$(ls "${BATS_TEST_TMPDIR}"/test-config.?????? 2>/dev/null | wc -l)
    [[ "${temps}" -eq 0 ]]
}

@test "config_atomic_write: writes via printf (no trailing newline)" {
    local target="${BATS_TEST_TMPDIR}/test-config"
    config_atomic_write "${target}" "no newline"
    # printf '%s' does not append newline
    local content
    content=$(cat "${target}")
    [[ "${content}" == "no newline" ]]
}

# --- Sed replace tests ---

@test "config_sed_replace: replaces pattern in file" {
    local target="${BATS_TEST_TMPDIR}/test-config"
    echo "color=red" > "${target}"
    config_sed_replace "${target}" "s/red/blue/"
    run cat "${target}"
    assert_output "color=blue"
}

@test "config_sed_replace: handles multiple lines" {
    local target="${BATS_TEST_TMPDIR}/test-config"
    printf "line1=old\nline2=old\n" > "${target}"
    config_sed_replace "${target}" "s/old/new/g"
    run cat "${target}"
    assert_line --index 0 "line1=new"
    assert_line --index 1 "line2=new"
}

@test "config_sed_replace: no leftover temp files on success" {
    local target="${BATS_TEST_TMPDIR}/test-config"
    echo "value=1" > "${target}"
    config_sed_replace "${target}" "s/1/2/"
    local temps
    temps=$(ls "${BATS_TEST_TMPDIR}"/test-config.?????? 2>/dev/null | wc -l)
    [[ "${temps}" -eq 0 ]]
}

# --- Backup tests ---

@test "config_backup: creates backup file" {
    local target="${BATS_TEST_TMPDIR}/test-config"
    echo "backup me" > "${target}"
    config_backup "${target}"
    local backup_dir="${HOME}/.dotfiles-backup/$(date +%Y%m%d)"
    local backups
    backups=$(ls "${backup_dir}"/test-config.* 2>/dev/null | wc -l)
    [[ "${backups}" -ge 1 ]]
}

@test "config_backup: backup content matches original" {
    local target="${BATS_TEST_TMPDIR}/test-config"
    echo "important data" > "${target}"
    config_backup "${target}"
    local backup_dir="${HOME}/.dotfiles-backup/$(date +%Y%m%d)"
    local backup_file
    backup_file=$(ls "${backup_dir}"/test-config.* 2>/dev/null | head -1)
    run cat "${backup_file}"
    assert_output "important data"
}

@test "config_backup: creates backup directory if it does not exist" {
    local target="${BATS_TEST_TMPDIR}/test-config"
    echo "data" > "${target}"
    local backup_dir="${HOME}/.dotfiles-backup/$(date +%Y%m%d)"
    [[ ! -d "${backup_dir}" ]]  # should not exist yet
    config_backup "${target}"
    [[ -d "${backup_dir}" ]]  # should exist now
}
