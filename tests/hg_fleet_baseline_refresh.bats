#!/usr/bin/env bats

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export TEST_HOME="${BATS_TEST_TMPDIR}/home"
    export TEST_ROOT="${TEST_HOME}/hairglasses-studio"
    export SCRIPTS_REAL="${DOTFILES_DIR}/scripts"

    mkdir -p "${TEST_HOME}" "${TEST_ROOT}/workspace" "${TEST_ROOT}/docs/agent-parity"

    export HOME="${TEST_HOME}"
    export HG_STUDIO_ROOT="${TEST_ROOT}"
    export GIT_CONFIG_GLOBAL="${TEST_HOME}/.gitconfig"

    git config --global user.name "Codex Test"
    git config --global user.email "codex@example.com"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

write_manifest() {
    cat > "${TEST_ROOT}/workspace/manifest.json" <<EOF
{
  "version": 1,
  "repos": [
    {
      "name": "alpha",
      "baseline_profile": "make_check",
      "workflow_policy": "managed",
      "workflow_family": "go_mcp"
    },
    {
      "name": "beta",
      "baseline_profile": "make_check",
      "workflow_policy": "managed",
      "workflow_family": "go_mcp"
    }
  ]
}
EOF
}

init_repo() {
    local repo_name="$1"
    mkdir -p "${TEST_ROOT}/${repo_name}"
    git -C "${TEST_ROOT}/${repo_name}" init -q
    git -C "${TEST_ROOT}/${repo_name}" branch -M main
}

seed_make_check_repo() {
    local repo_name="$1"
    init_repo "${repo_name}"
    cat > "${TEST_ROOT}/${repo_name}/Makefile" <<'EOF'
check:
	@echo "ok"
EOF
    mkdir -p "${TEST_ROOT}/${repo_name}/.github/workflows"
    cat > "${TEST_ROOT}/${repo_name}/.github/workflows/ci.yml" <<'EOF'
name: ci
on: push
EOF
    git -C "${TEST_ROOT}/${repo_name}" add Makefile .github/workflows/ci.yml
    git -C "${TEST_ROOT}/${repo_name}" commit -q -m "seed repo"
}

matrix_path() {
    printf '%s\n' "${TEST_ROOT}/docs/agent-parity/workspace-health-matrix.json"
}

json_query() {
    local python_expr="$1"
    python3 - "$python_expr" "$(matrix_path)" <<'PY'
import json
import sys

expr = sys.argv[1]
path = sys.argv[2]
data = json.load(open(path, encoding="utf-8"))
print(eval(expr, {"data": data}))
PY
}

@test "hg-fleet-baseline-refresh: pass rows record command and workflow status" {
    write_manifest
    seed_make_check_repo "alpha"
    seed_make_check_repo "beta"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" \
        bash "${SCRIPTS_REAL}/hg-fleet-baseline-refresh.sh" "${TEST_ROOT}"
    assert_success
    assert_output --partial "workspace-health-matrix.json"

    run json_query 'next(row["local_baseline_status"] for row in data["repos"] if row["repo"] == "alpha")'
    assert_success
    assert_output "pass"

    run json_query 'next(row["baseline_command"] for row in data["repos"] if row["repo"] == "alpha")'
    assert_success
    assert_output "GOWORK=off GOFLAGS=-buildvcs=false make check"

    run json_query 'next(row["workflow_status"] for row in data["repos"] if row["repo"] == "alpha")'
    assert_success
    assert_output "present"
}

@test "hg-fleet-baseline-refresh: dirty repo is classified before running baseline" {
    write_manifest
    seed_make_check_repo "alpha"
    seed_make_check_repo "beta"
    printf '\n# dirty\n' >> "${TEST_ROOT}/alpha/Makefile"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" \
        bash "${SCRIPTS_REAL}/hg-fleet-baseline-refresh.sh" "${TEST_ROOT}" --repo alpha
    assert_success

    run json_query 'next(row["local_baseline_status"] for row in data["repos"] if row["repo"] == "alpha")'
    assert_success
    assert_output "dirty"

    run json_query 'next(row["baseline_command"] for row in data["repos"] if row["repo"] == "alpha")'
    assert_success
    assert_output ""
}

@test "hg-fleet-baseline-refresh: non-main repos are classified as not_main" {
    write_manifest
    seed_make_check_repo "alpha"
    seed_make_check_repo "beta"
    git -C "${TEST_ROOT}/alpha" checkout -q -b feature/fleet-refresh

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" \
        bash "${SCRIPTS_REAL}/hg-fleet-baseline-refresh.sh" "${TEST_ROOT}" --repo alpha
    assert_success

    run json_query 'next(row["local_baseline_status"] for row in data["repos"] if row["repo"] == "alpha")'
    assert_success
    assert_output "not_main"

    run json_query 'next(row["current_branch"] for row in data["repos"] if row["repo"] == "alpha")'
    assert_success
    assert_output "feature/fleet-refresh"
}

@test "hg-fleet-baseline-refresh: repo filter narrows output rows" {
    write_manifest
    seed_make_check_repo "alpha"
    seed_make_check_repo "beta"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" \
        bash "${SCRIPTS_REAL}/hg-fleet-baseline-refresh.sh" "${TEST_ROOT}" --repo beta
    assert_success

    run json_query 'len(data["repos"])'
    assert_success
    assert_output "1"

    run json_query 'data["repos"][0]["repo"]'
    assert_success
    assert_output "beta"
}
