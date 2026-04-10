#!/usr/bin/env bats

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export TEST_CANONICAL="${BATS_TEST_TMPDIR}/canonical"
    export TEST_STANDALONE="${BATS_TEST_TMPDIR}/standalone"
    mkdir -p "${TEST_CANONICAL}/scripts" "${TEST_STANDALONE}/internal/dotfiles" "${TEST_STANDALONE}/cmd/tool" "${TEST_STANDALONE}/internal/githubstars"

    cat > "${TEST_CANONICAL}/go.mod" <<'EOF'
module example.com/canonical

go 1.24
EOF
    cat > "${TEST_CANONICAL}/README.md" <<'EOF'
# canonical
EOF
    cat > "${TEST_CANONICAL}/scripts/host-smoke.sh" <<'EOF'
#!/usr/bin/env bash
echo host-smoke
EOF
    cat > "${TEST_STANDALONE}/README.md" <<'EOF'
# canonical
EOF
    mkdir -p "${TEST_STANDALONE}/scripts"
    cat > "${TEST_STANDALONE}/scripts/host-smoke.sh" <<'EOF'
#!/usr/bin/env bash
echo standalone-host-smoke
EOF
    cat > "${TEST_CANONICAL}/alpha.go" <<'EOF'
package main

func alpha() string { return "alpha" }
EOF
    cat > "${TEST_CANONICAL}/beta.go" <<'EOF'
package main

func beta() string { return "beta" }
EOF

    cat > "${TEST_STANDALONE}/internal/dotfiles/beta.go" <<'EOF'
package dotfiles

func beta() string { return "stale beta" }
EOF
    cat > "${TEST_STANDALONE}/internal/dotfiles/main.go" <<'EOF'
package dotfiles

func Setup() {}
EOF
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

@test "hg-dotfiles-mcp-projection reports projection drift and standalone-owned surfaces" {
    run bash "${SCRIPTS_DIR}/hg-dotfiles-mcp-projection.sh" check --canonical "${TEST_CANONICAL}" --standalone "${TEST_STANDALONE}"
    assert_success
    assert_output --partial "status              projection_needed"
    assert_output --partial "drifted           1"
    assert_output --partial "scripts/host-smoke.sh"
    assert_output --partial "alpha.go"
    assert_output --partial "beta.go"
    assert_output --partial "cmd"
    assert_output --partial "internal/githubstars"
}

@test "hg-dotfiles-mcp-projection emits machine-readable JSON" {
    run bash -lc "bash '${SCRIPTS_DIR}/hg-dotfiles-mcp-projection.sh' plan --canonical '${TEST_CANONICAL}' --standalone '${TEST_STANDALONE}' --json | jq -r '.status, .direct_copy.drifted[0], .go_projection.projection_required[0], .standalone_owned.root_entries[0]'"
    assert_success
    assert_line --index 0 "projection_needed"
    assert_line --index 1 "scripts/host-smoke.sh"
    assert_line --index 2 "alpha.go"
    assert_line --index 3 "cmd"
}

@test "hg-dotfiles-mcp-projection emits diff previews when requested" {
    run bash -lc "bash '${SCRIPTS_DIR}/hg-dotfiles-mcp-projection.sh' plan --canonical '${TEST_CANONICAL}' --standalone '${TEST_STANDALONE}' --json --diff-preview --diff-lines 6 | jq -r '.direct_copy.drift_preview_enabled, .direct_copy.drift_preview_lines, .direct_copy.drift_previews[0].path, (.direct_copy.drift_previews[0].preview | contains(\"standalone-host-smoke\")), .go_projection.drift_previews[0].path, (.go_projection.drift_previews[0].preview | contains(\"stale beta\"))'"
    assert_success
    assert_line --index 0 "true"
    assert_line --index 1 "6"
    assert_line --index 2 "scripts/host-smoke.sh"
    assert_line --index 3 "true"
    assert_line --index 4 "beta.go"
    assert_line --index 5 "true"
}

@test "hg-dotfiles-mcp-projection classifies intentional direct-copy drift separately" {
    cat > "${TEST_CANONICAL}/.goreleaser.yml" <<'EOF'
version: 2
EOF
    cat > "${TEST_STANDALONE}/README.md" <<'EOF'
# standalone mirror
EOF

    run bash -lc "bash '${SCRIPTS_DIR}/hg-dotfiles-mcp-projection.sh' plan --canonical '${TEST_CANONICAL}' --standalone '${TEST_STANDALONE}' --json | jq -r '.direct_copy.required_drift_count, .direct_copy.intentional_drift_count, .direct_copy.required_missing_count, .direct_copy.intentional_missing_count, .direct_copy.intentional_drift[0], .direct_copy.intentional_missing[0]'"
    assert_success
    assert_line --index 0 "0"
    assert_line --index 1 "2"
    assert_line --index 2 "0"
    assert_line --index 3 "1"
    assert_line --index 4 "README.md"
    assert_line --index 5 ".goreleaser.yml"
}

@test "hg-dotfiles-mcp-projection prints diff previews in text mode when requested" {
    run bash "${SCRIPTS_DIR}/hg-dotfiles-mcp-projection.sh" check --canonical "${TEST_CANONICAL}" --standalone "${TEST_STANDALONE}" --diff-preview --diff-lines 5
    assert_success
    assert_output --partial "direct-copy diff previews (first 5 lines per file)"
    assert_output --partial "overlapping drift previews (first 5 lines per file)"
    assert_output --partial "standalone-host-smoke"
    assert_output --partial "stale beta"
}

@test "hg-dotfiles-mcp-projection classifies intentional canonical-only differences separately" {
    cat > "${TEST_CANONICAL}/contract_snapshot_cli.go" <<'EOF'
package main

func contractOnly() {}
EOF
    cat > "${TEST_CANONICAL}/workflow_surface_test.go" <<'EOF'
package main

func workflowSurfaceOnly() {}
EOF

    run bash -lc "bash '${SCRIPTS_DIR}/hg-dotfiles-mcp-projection.sh' plan --canonical '${TEST_CANONICAL}' --standalone '${TEST_STANDALONE}' --json | jq -r '.go_projection.canonical_only_count, .go_projection.projection_required_count, .go_projection.intentional_canonical_only_count, .go_projection.intentional_canonical_only[0], .go_projection.intentional_canonical_only[1]'"
    assert_success
    assert_line --index 0 "3"
    assert_line --index 1 "1"
    assert_line --index 2 "2"
    assert_line --index 3 "contract_snapshot_cli.go"
    assert_line --index 4 "workflow_surface_test.go"
}

@test "hg-dotfiles-mcp-projection tracks workflow drift as required direct-copy projection" {
    mkdir -p "${TEST_CANONICAL}/.github/workflows" "${TEST_STANDALONE}/.github/workflows"
    cat > "${TEST_CANONICAL}/.github/workflows/release.yml" <<'EOF'
name: Release
EOF
    cat > "${TEST_STANDALONE}/.github/workflows/release.yml" <<'EOF'
name: Mirror Release
EOF

    run bash -lc "bash '${SCRIPTS_DIR}/hg-dotfiles-mcp-projection.sh' plan --canonical '${TEST_CANONICAL}' --standalone '${TEST_STANDALONE}' --json | jq -r '.direct_copy.required_drift_count, .direct_copy.required_drift[0]'"
    assert_success
    assert_line --index 0 "1"
    assert_line --index 1 ".github/workflows/release.yml"
}
