#!/usr/bin/env bats

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export HOME="${BATS_TEST_TMPDIR}/home"
    mkdir -p "${HOME}"

    export REMOTE_REPO="${BATS_TEST_TMPDIR}/remote.git"
    export WORK_REPO="${BATS_TEST_TMPDIR}/work"
    export PEER_REPO="${BATS_TEST_TMPDIR}/peer"

    git init --bare "${REMOTE_REPO}" >/dev/null 2>&1
    git clone "${REMOTE_REPO}" "${WORK_REPO}" >/dev/null 2>&1
    git -C "${WORK_REPO}" config user.name "Test User"
    git -C "${WORK_REPO}" config user.email "test@example.com"
    git -C "${WORK_REPO}" branch -M main
    printf 'initial\n' > "${WORK_REPO}/README.md"
    git -C "${WORK_REPO}" add README.md
    git -C "${WORK_REPO}" commit -m "initial" >/dev/null
    git -C "${WORK_REPO}" push -u origin main >/dev/null 2>&1
    git -C "${REMOTE_REPO}" symbolic-ref HEAD refs/heads/main
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

@test "publish preflight reports in_place for a clean repo with origin/main" {
    run bash "${SCRIPTS_DIR}/hg-publish-preflight.sh" --json "${WORK_REPO}"
    assert_success
    local json_output="${output}"
    run jq -r '.lane' <<<"${json_output}"
    assert_success
    assert_output "in_place"
    run jq -r '.compare_ref' <<<"${json_output}"
    assert_success
    assert_output "origin/main"
    run jq -r '.dirty_tracked,.dirty_untracked' <<<"${json_output}"
    assert_success
    assert_line --index 0 "0"
    assert_line --index 1 "0"
}

@test "publish preflight recommends a clean worktree when the repo is dirty" {
    printf 'dirty\n' >> "${WORK_REPO}/README.md"

    run bash "${SCRIPTS_DIR}/hg-publish-preflight.sh" --json "${WORK_REPO}"
    assert_success
    local json_output="${output}"
    run jq -r '.lane,.reason,.dirty_tracked' <<<"${json_output}"
    assert_success
    assert_line --index 0 "clean_worktree"
    assert_line --index 1 "worktree has uncommitted changes"
    assert_line --index 2 "1"
}

@test "publish preflight recommends a clean worktree when the branch is behind origin/main" {
    git clone "${REMOTE_REPO}" "${PEER_REPO}" >/dev/null 2>&1
    git -C "${PEER_REPO}" config user.name "Peer User"
    git -C "${PEER_REPO}" config user.email "peer@example.com"
    git -C "${PEER_REPO}" branch -M main
    printf 'peer change\n' >> "${PEER_REPO}/README.md"
    git -C "${PEER_REPO}" add README.md
    git -C "${PEER_REPO}" commit -m "peer change" >/dev/null
    git -C "${PEER_REPO}" push origin main >/dev/null 2>&1
    git -C "${WORK_REPO}" fetch origin main >/dev/null 2>&1

    run bash "${SCRIPTS_DIR}/hg-publish-preflight.sh" --json "${WORK_REPO}"
    assert_success
    local json_output="${output}"
    run jq -r '.lane,.reason,.behind' <<<"${json_output}"
    assert_success
    assert_line --index 0 "clean_worktree"
    assert_line --index 1 "local branch is behind origin/main"
    assert_line --index 2 "1"
}

@test "publish preflight blocks non-repositories" {
    run bash "${SCRIPTS_DIR}/hg-publish-preflight.sh" --json "${BATS_TEST_TMPDIR}"
    assert_failure
    local json_output="${output}"
    run jq -r '.lane,.reason' <<<"${json_output}"
    assert_success
    assert_line --index 0 "blocked"
    assert_line --index 1 "not a git repository"
}
