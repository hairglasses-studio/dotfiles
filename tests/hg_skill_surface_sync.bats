#!/usr/bin/env bats

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export TEST_STUDIO_ROOT="${BATS_TEST_TMPDIR}/studio"
    export TEST_REPO_PATH="${BATS_TEST_TMPDIR}/repo"
    mkdir -p "${TEST_STUDIO_ROOT}/dotfiles" "${TEST_REPO_PATH}"
    ln -s "${DOTFILES_DIR}/scripts" "${TEST_STUDIO_ROOT}/dotfiles/scripts"
    ln -s "${DOTFILES_DIR}/AGENTS.md" "${TEST_STUDIO_ROOT}/dotfiles/AGENTS.md"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

make_fake_codexkit_wrapper() {
    local exit_code="${1:-0}"
    mkdir -p "${TEST_STUDIO_ROOT}/codexkit/scripts"
    cat > "${TEST_STUDIO_ROOT}/codexkit/scripts/skill-surface-sync.sh" <<EOF
#!/usr/bin/env bash
printf '%s\n' "\$@" > "${BATS_TEST_TMPDIR}/forwarded.args"
exit ${exit_code}
EOF
    chmod +x "${TEST_STUDIO_ROOT}/codexkit/scripts/skill-surface-sync.sh"
}

@test "hg-skill-surface-sync forwards repo path and flags to codexkit" {
    make_fake_codexkit_wrapper 0

    run env HG_STUDIO_ROOT="${TEST_STUDIO_ROOT}" \
        bash "${SCRIPTS_DIR}/hg-skill-surface-sync.sh" "${TEST_REPO_PATH}" --check
    assert_success

    run cat "${BATS_TEST_TMPDIR}/forwarded.args"
    assert_success
    assert_line --index 0 "${TEST_REPO_PATH}"
    assert_line --index 1 "--check"
}

@test "hg-skill-surface-sync preserves codexkit exit status" {
    make_fake_codexkit_wrapper 17

    run env HG_STUDIO_ROOT="${TEST_STUDIO_ROOT}" \
        bash "${SCRIPTS_DIR}/hg-skill-surface-sync.sh" "${TEST_REPO_PATH}"
    assert_failure
    [ "$status" -eq 17 ]
}

@test "hg-skill-surface-sync fails when codexkit wrapper is missing" {
    run env HG_STUDIO_ROOT="${TEST_STUDIO_ROOT}" \
        bash "${SCRIPTS_DIR}/hg-skill-surface-sync.sh" "${TEST_REPO_PATH}"
    assert_failure
    assert_output --partial "No managed skill sync entrypoint found"
}
