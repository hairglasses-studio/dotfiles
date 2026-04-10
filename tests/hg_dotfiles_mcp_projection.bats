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
    assert_output --partial "alpha.go"
    assert_output --partial "beta.go"
    assert_output --partial "cmd"
    assert_output --partial "internal/githubstars"
}

@test "hg-dotfiles-mcp-projection emits machine-readable JSON" {
    run bash -lc "bash '${SCRIPTS_DIR}/hg-dotfiles-mcp-projection.sh' plan --canonical '${TEST_CANONICAL}' --standalone '${TEST_STANDALONE}' --json | jq -r '.status, .go_projection.canonical_only[0], .standalone_owned.root_entries[0]'"
    assert_success
    assert_line --index 0 "projection_needed"
    assert_line --index 1 "alpha.go"
    assert_line --index 2 "cmd"
}
