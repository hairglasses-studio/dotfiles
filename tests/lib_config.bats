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

# --- Reload / restart tests ---

@test "config_reload_service: hyprshell touches watched files instead of restarting the service" {
    export PATH="${BATS_TEST_TMPDIR}:${PATH}"
    mkdir -p "${HOME}/.config/hyprshell"
    printf 'version = 3\n' > "${HOME}/.config/hyprshell/config.toml"
    printf 'window {}\n' > "${HOME}/.config/hyprshell/styles.css"

    cat > "${BATS_TEST_TMPDIR}/pgrep" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
    chmod +x "${BATS_TEST_TMPDIR}/pgrep"

    cat > "${BATS_TEST_TMPDIR}/touch" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "${BATS_TEST_TMPDIR}/touch.log"
/usr/bin/touch "$@"
EOF
    chmod +x "${BATS_TEST_TMPDIR}/touch"

    cat > "${BATS_TEST_TMPDIR}/systemctl" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "${BATS_TEST_TMPDIR}/systemctl.log"
EOF
    chmod +x "${BATS_TEST_TMPDIR}/systemctl"

    config_reload_service hyprshell --quiet

    run cat "${BATS_TEST_TMPDIR}/touch.log"
    assert_success
    assert_output --partial "${HOME}/.config/hyprshell/config.toml"
    assert_output --partial "${HOME}/.config/hyprshell/styles.css"

    run test -f "${BATS_TEST_TMPDIR}/systemctl.log"
    assert_failure
}

@test "config_reload_service: hyprshell watched-file reload works through repo symlinks" {
    export PATH="${BATS_TEST_TMPDIR}:${PATH}"
    local repo_root="${BATS_TEST_TMPDIR}/repo"
    local live_dir="${HOME}/.config/hyprshell"
    mkdir -p "${repo_root}/hyprshell" "${live_dir}"
    printf 'version = 3\n' > "${repo_root}/hyprshell/config.toml"
    printf 'window {}\n' > "${repo_root}/hyprshell/styles.css"
    ln -s "${repo_root}/hyprshell/config.toml" "${live_dir}/config.toml"
    ln -s "${repo_root}/hyprshell/styles.css" "${live_dir}/styles.css"

    local config_before styles_before
    config_before="$(stat -c %Y "${repo_root}/hyprshell/config.toml")"
    styles_before="$(stat -c %Y "${repo_root}/hyprshell/styles.css")"
    sleep 1

    cat > "${BATS_TEST_TMPDIR}/pgrep" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
    chmod +x "${BATS_TEST_TMPDIR}/pgrep"

    cat > "${BATS_TEST_TMPDIR}/systemctl" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "${BATS_TEST_TMPDIR}/systemctl.log"
EOF
    chmod +x "${BATS_TEST_TMPDIR}/systemctl"

    run config_reload_service hyprshell --quiet
    assert_success

    local config_after styles_after
    config_after="$(stat -c %Y "${repo_root}/hyprshell/config.toml")"
    styles_after="$(stat -c %Y "${repo_root}/hyprshell/styles.css")"
    [[ "${config_after}" -gt "${config_before}" ]]
    [[ "${styles_after}" -gt "${styles_before}" ]]

    run test -f "${BATS_TEST_TMPDIR}/systemctl.log"
    assert_failure
}

@test "config_reload_service: hyprshell fails cleanly when the process is not running" {
    export PATH="${BATS_TEST_TMPDIR}:${PATH}"
    cat > "${BATS_TEST_TMPDIR}/pgrep" <<'EOF'
#!/usr/bin/env bash
exit 1
EOF
    chmod +x "${BATS_TEST_TMPDIR}/pgrep"

    run config_reload_service hyprshell --quiet
    assert_failure
}

@test "config_restart_service: hyprshell still has an explicit systemctl restart path" {
    export PATH="${BATS_TEST_TMPDIR}:${PATH}"
    cat > "${BATS_TEST_TMPDIR}/systemctl" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "${BATS_TEST_TMPDIR}/systemctl.log"
EOF
    chmod +x "${BATS_TEST_TMPDIR}/systemctl"

    run config_restart_service hyprshell --quiet
    assert_success

    run cat "${BATS_TEST_TMPDIR}/systemctl.log"
    assert_success
    assert_output --partial "--user restart dotfiles-hyprshell.service"
}
