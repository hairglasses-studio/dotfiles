#!/usr/bin/env bats

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export TEST_HOME="${BATS_TEST_TMPDIR}/home"
    export TEST_ROOT="${TEST_HOME}/hairglasses-studio"
    export SCRIPTS_REAL="${DOTFILES_DIR}/scripts"

    mkdir -p "${TEST_HOME}" "${TEST_ROOT}/workspace" "${TEST_ROOT}/.github/workflow-templates" "${TEST_ROOT}/mcpkit/.github/workflows"

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
    local repo_name="$1"
    local lifecycle="${2:-canonical}"
    local mirror_of="${3:-null}"
    local ci_profile="${4:-none}"
    local review_profile="${5:-none}"

    cat > "${TEST_ROOT}/workspace/manifest.json" <<EOF
{
  "version": 2,
  "repos": [
    {
      "name": "${repo_name}",
      "baseline_target": true,
      "lifecycle": "${lifecycle}",
      "mirror_of": ${mirror_of},
      "ci_profile": "${ci_profile}",
      "review_profile": "${review_profile}"
    }
  ]
}
EOF
}

init_repo() {
    local repo_name="$1"
    mkdir -p "${TEST_ROOT}/${repo_name}"
    git -C "${TEST_ROOT}/${repo_name}" init -q
}

write_agents() {
    local repo_name="$1"
    cat > "${TEST_ROOT}/${repo_name}/AGENTS.md" <<'EOF'
# Demo — Agent Instructions

> Canonical instructions: AGENTS.md

## Build & Test

```bash
make test
```
EOF
}

@test "hg-repo-profile-sync: sync creates Gemini project settings baseline" {
    write_manifest "demo"
    init_repo "demo"
    write_agents "demo"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-repo-profile-sync.sh" sync --repos=demo
    assert_success

    run test -f "${TEST_ROOT}/demo/.gemini/settings.json"
    assert_success

    run grep -F 'AGENTS.md' "${TEST_ROOT}/demo/.gemini/settings.json"
    assert_success
}

@test "hg-repo-profile-sync: sync skips dirty generated agent docs until allow-dirty is set" {
    write_manifest "demo"
    init_repo "demo"
    write_agents "demo"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-agent-docs.sh" "${TEST_ROOT}/demo"
    assert_success

    git -C "${TEST_ROOT}/demo" add AGENTS.md CLAUDE.md GEMINI.md .github/copilot-instructions.md
    git -C "${TEST_ROOT}/demo" commit -q -m "seed generated docs"

    printf '\ndirty manual edit\n' >> "${TEST_ROOT}/demo/CLAUDE.md"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-repo-profile-sync.sh" sync --repos=demo
    assert_success
    assert_output --partial "skipping dirty generated agent docs"

    run grep -F "dirty manual edit" "${TEST_ROOT}/demo/CLAUDE.md"
    assert_success

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-repo-profile-sync.sh" sync --allow-dirty --repos=demo
    assert_success
    refute_output --partial "skipping dirty generated agent docs"

    run grep -F "dirty manual edit" "${TEST_ROOT}/demo/CLAUDE.md"
    assert_failure
}

@test "hg-workflow-sync: allow-dirty updates a dirty managed workflow" {
    write_manifest "demo" "canonical" "null" "none" "public_ai_review"
    init_repo "demo"
    mkdir -p "${TEST_ROOT}/demo/.github/workflows"

    cat > "${TEST_ROOT}/.github/workflow-templates/claude-review.yml" <<'EOF'
name: claude-review
on: pull_request
EOF

    cat > "${TEST_ROOT}/demo/.github/workflows/claude-review.yml" <<'EOF'
name: old-review
EOF

    git -C "${TEST_ROOT}/demo" add .github/workflows/claude-review.yml
    git -C "${TEST_ROOT}/demo" commit -q -m "seed workflow"

    printf '\n# dirty\n' >> "${TEST_ROOT}/demo/.github/workflows/claude-review.yml"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-workflow-sync.sh" --repos=demo
    assert_success
    assert_output --partial "skipping dirty workflow claude-review.yml"

    run grep -F "# dirty" "${TEST_ROOT}/demo/.github/workflows/claude-review.yml"
    assert_success

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-workflow-sync.sh" --allow-dirty --repos=demo
    assert_success

    run grep -F "# dirty" "${TEST_ROOT}/demo/.github/workflows/claude-review.yml"
    assert_failure
    run grep -F "name: claude-review" "${TEST_ROOT}/demo/.github/workflows/claude-review.yml"
    assert_success
}

@test "sync-standalone-mcp-repos: check works through bash wrapper when mcp-mirror helper is non-executable" {
    write_manifest "mirror-repo" "mirror" "\"canonical-src\""
    mkdir -p "${TEST_ROOT}/canonical-src"
    printf 'hello\n' > "${TEST_ROOT}/canonical-src/README.md"

    init_repo "mirror-repo"
    printf 'hello\n' > "${TEST_ROOT}/mirror-repo/README.md"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/sync-standalone-mcp-repos.sh" check --repos=mirror-repo
    assert_success
}
