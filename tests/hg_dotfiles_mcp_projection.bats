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

@test "hg-dotfiles-mcp-projection refreshes bare origin when requested" {
    local refresh_canonical="${BATS_TEST_TMPDIR}/refresh-canonical"
    local refresh_origin="${BATS_TEST_TMPDIR}/refresh-origin"
    local refresh_bare="${BATS_TEST_TMPDIR}/refresh-bare"

    mkdir -p "${refresh_canonical}/internal/dotfiles"
    cat > "${refresh_canonical}/go.mod" <<'EOF'
module example.com/refresh

go 1.24
EOF
    cat > "${refresh_canonical}/README.md" <<'EOF'
# refresh
EOF
    cat > "${refresh_canonical}/internal/dotfiles/main.go" <<'EOF'
package dotfiles

func Setup() {}
EOF

    mkdir -p "${refresh_origin}"
    git init -q "${refresh_origin}"
    git -C "${refresh_origin}" config user.name "Codex Test"
    git -C "${refresh_origin}" config user.email "codex@example.com"
    mkdir -p "${refresh_origin}/internal/dotfiles"
    cat > "${refresh_origin}/go.mod" <<'EOF'
module example.com/refresh

go 1.24
EOF
    cat > "${refresh_origin}/README.md" <<'EOF'
# refresh
EOF
    cat > "${refresh_origin}/internal/dotfiles/main.go" <<'EOF'
package dotfiles

func Setup() {}
EOF
    git -C "${refresh_origin}" add go.mod README.md internal/dotfiles/main.go
    git -C "${refresh_origin}" commit -qm "initial"
    git -C "${refresh_origin}" branch -M main

    git clone --bare "${refresh_origin}" "${refresh_bare}" >/dev/null 2>&1

    cat > "${refresh_origin}/README.md" <<'EOF'
# refresh standalone
EOF
    git -C "${refresh_origin}" add README.md
    git -C "${refresh_origin}" commit -qm "advance"

    run bash -lc "bash '${SCRIPTS_DIR}/hg-dotfiles-mcp-projection.sh' plan --canonical '${refresh_canonical}' --standalone '${refresh_bare}' --json | jq -r '.refreshed_bare_origin, .direct_copy.intentional_drift_count'"
    assert_success
    assert_line --index 0 "false"
    assert_line --index 1 "0"

    run bash -lc "bash '${SCRIPTS_DIR}/hg-dotfiles-mcp-projection.sh' plan --canonical '${refresh_canonical}' --standalone '${refresh_bare}' --refresh-bare-origin --json | jq -r '.refreshed_bare_origin, .direct_copy.intentional_drift_count, .direct_copy.intentional_drift[0]'"
    assert_success
    assert_line --index 0 "true"
    assert_line --index 1 "1"
    assert_line --index 2 "README.md"
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

@test "hg-dotfiles-mcp-projection marks review-only overlap drift as parity review" {
    cat > "${TEST_STANDALONE}/internal/dotfiles/alpha.go" <<'EOF'
package dotfiles

func alpha() string { return "alpha" }
EOF
    cat > "${TEST_STANDALONE}/internal/dotfiles/beta.go" <<'EOF'
package dotfiles

func beta() string { return "beta" }
EOF
    cat > "${TEST_CANONICAL}/discovery_workstation_diagnostics.go" <<'EOF'
package main

func diagnosticsValue() string { return "canonical" }
EOF
    cat > "${TEST_STANDALONE}/internal/dotfiles/discovery_workstation_diagnostics.go" <<'EOF'
package dotfiles

func diagnosticsValue() string { return "standalone" }
EOF

    run bash -lc "bash '${SCRIPTS_DIR}/hg-dotfiles-mcp-projection.sh' plan --canonical '${TEST_CANONICAL}' --standalone '${TEST_STANDALONE}' --json | jq -r '.status, .go_projection.required_drift_count, .go_projection.review_drift_count, .go_projection.review_drift[0]'"
    assert_success
    assert_line --index 0 "parity_review_needed"
    assert_line --index 1 "0"
    assert_line --index 2 "1"
    assert_line --index 3 "discovery_workstation_diagnostics.go"
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
