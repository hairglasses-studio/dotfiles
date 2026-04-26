#!/usr/bin/env bats

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export TEST_STUDIO_ROOT="${BATS_TEST_TMPDIR}/studio"
    export TEST_HOME="${BATS_TEST_TMPDIR}/home"
    export TEST_REPO_PATH="${TEST_STUDIO_ROOT}/demo"
    export TEST_ANTIGRAVITY_SYNC="${BATS_TEST_TMPDIR}/antigravity-sync.sh"

    mkdir -p \
        "${TEST_STUDIO_ROOT}/workspace" \
        "${TEST_STUDIO_ROOT}/dotfiles" \
        "${TEST_REPO_PATH}" \
        "${TEST_HOME}/.claude/commands" \
        "${TEST_HOME}/.claude/skills" \
        "${TEST_HOME}/.agents/skills" \
        "${TEST_HOME}/.codex/skills" \
        "${TEST_HOME}/.gemini"

    ln -s "${DOTFILES_DIR}/scripts" "${TEST_STUDIO_ROOT}/dotfiles/scripts"
    ln -s "${DOTFILES_DIR}/AGENTS.md" "${TEST_STUDIO_ROOT}/dotfiles/AGENTS.md"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

write_manifest() {
    cat > "${TEST_STUDIO_ROOT}/workspace/manifest.json" <<'EOF'
{
  "version": 2,
  "repos": [
    {
      "name": "demo",
      "baseline_target": true,
      "scope": "active_first_party",
      "workflow_policy": "local_only"
    }
  ]
}
EOF
}

make_fake_antigravity_sync() {
    cat > "${TEST_ANTIGRAVITY_SYNC}" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

mode="write"
for arg in "$@"; do
    case "$arg" in
        --check) mode="check" ;;
        --dry-run) mode="dry-run" ;;
    esac
done

metadata_path="${HOME}/.gemini/antigravity/.hg-antigravity-sync.json"
mkdir -p "$(dirname "${metadata_path}")"

python3 - "${metadata_path}" "${mode}" <<'PY'
import json
import pathlib
import sys

metadata_path = pathlib.Path(sys.argv[1])
mode = sys.argv[2]
home = pathlib.Path.home()
state = {
    "claude_home_present": (home / ".claude" / "CLAUDE.md").exists(),
    "gemini_home_present": (home / ".gemini" / "GEMINI.md").exists(),
}

if mode == "check":
    if not metadata_path.exists() or json.loads(metadata_path.read_text()) != state:
        print("stub antigravity drift", file=sys.stderr)
        raise SystemExit(1)
else:
    metadata_path.write_text(json.dumps(state, sort_keys=True) + "\n")
PY
EOF
    chmod +x "${TEST_ANTIGRAVITY_SYNC}"
}

make_fake_skill_surface_sync() {
    mkdir -p "${TEST_STUDIO_ROOT}/codexkit/scripts"
    cat > "${TEST_STUDIO_ROOT}/codexkit/scripts/skill-surface-sync.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
exit 0
EOF
    chmod +x "${TEST_STUDIO_ROOT}/codexkit/scripts/skill-surface-sync.sh"
}

@test "hg-workspace-global-sync skips retired shared root MCP entries and anchors relative Codex commands" {
    write_manifest
    make_fake_antigravity_sync

    mkdir -p "${TEST_REPO_PATH}/.codex"

    cat > "${TEST_STUDIO_ROOT}/.mcp.json" <<'EOF'
{
  "mcpServers": {}
}
EOF

    cat > "${TEST_REPO_PATH}/.mcp.json" <<'EOF'
{
  "mcpServers": {
    "demo": {
      "command": "./scripts/run-demo-mcp.sh"
    }
  }
}
EOF

    cat > "${TEST_REPO_PATH}/.codex/mcp-profile-policy.json" <<'EOF'
{
  "version": 1,
  "profiles": [
    {
      "name": "demo_desktop",
      "from": "demo",
      "mode": "desktop",
      "global_name": "demo",
      "global_codex": true,
      "enabled_tools": ["demo_status"]
    }
  ]
}
EOF

    run env \
        HOME="${TEST_HOME}" \
        HG_STUDIO_ROOT="${TEST_STUDIO_ROOT}" \
        HG_WORKSPACE_HOME="${TEST_HOME}" \
        HG_WORKSPACE_OWNER="tester" \
        HG_ANTIGRAVITY_SYNC_SCRIPT="${TEST_ANTIGRAVITY_SYNC}" \
        bash "${SCRIPTS_DIR}/hg-workspace-global-sync.sh" --root "${TEST_STUDIO_ROOT}" --tools-only
    assert_success
    assert_output --partial "Skipping shared root MCP server absent"

    run grep -F "[mcp_servers.studio_demo]" "${TEST_HOME}/.codex/config.toml"
    assert_success

    run grep -F 'command = "./scripts/run-demo-mcp.sh"' "${TEST_HOME}/.codex/config.toml"
    assert_success

    run grep -F "cwd = \"${TEST_REPO_PATH}\"" "${TEST_HOME}/.codex/config.toml"
    assert_success

    run grep -F "[mcp_servers.studio_systemd]" "${TEST_HOME}/.codex/config.toml"
    assert_failure
}

@test "hg-workspace-global-sync refreshes home context before antigravity metadata check" {
    write_manifest
    make_fake_antigravity_sync

    run env \
        HOME="${TEST_HOME}" \
        HG_STUDIO_ROOT="${TEST_STUDIO_ROOT}" \
        HG_WORKSPACE_HOME="${TEST_HOME}" \
        HG_WORKSPACE_OWNER="tester" \
        HG_ANTIGRAVITY_SYNC_SCRIPT="${TEST_ANTIGRAVITY_SYNC}" \
        bash "${SCRIPTS_DIR}/hg-workspace-global-sync.sh" --root "${TEST_STUDIO_ROOT}" --skills-only
    assert_success

    run env \
        HOME="${TEST_HOME}" \
        HG_STUDIO_ROOT="${TEST_STUDIO_ROOT}" \
        HG_WORKSPACE_HOME="${TEST_HOME}" \
        HG_WORKSPACE_OWNER="tester" \
        HG_ANTIGRAVITY_SYNC_SCRIPT="${TEST_ANTIGRAVITY_SYNC}" \
        bash "${SCRIPTS_DIR}/hg-workspace-global-sync.sh" --root "${TEST_STUDIO_ROOT}" --skills-only --check
    assert_success
    assert_output --partial "Workspace global sync up to date"

    run grep -F '"claude_home_present": true' "${TEST_HOME}/.gemini/antigravity/.hg-antigravity-sync.json"
    assert_success

    run grep -F '"gemini_home_present": true' "${TEST_HOME}/.gemini/antigravity/.hg-antigravity-sync.json"
    assert_success
}

@test "hg-workspace-global-sync keeps workspace-owned repo skills out of portable codex and agents dirs" {
    write_manifest
    make_fake_antigravity_sync
    make_fake_skill_surface_sync

    mkdir -p \
        "${TEST_REPO_PATH}/.agents/skills/demo_skill" \
        "${TEST_HOME}/.codex/skills/demo-demo_skill" \
        "${TEST_HOME}/.agents/skills/demo_demo_skill"

    cat > "${TEST_REPO_PATH}/.agents/skills/surface.yaml" <<'EOF'
{"version":1,"skills":[{"name":"demo_skill","claude_include_canonical":true}]}
EOF

    cat > "${TEST_REPO_PATH}/.agents/skills/demo_skill/SKILL.md" <<'EOF'
---
name: demo_skill
description: Demo repo workflow.
allowed-tools:
  - Bash
---

# Demo
EOF

    cat > "${TEST_HOME}/.codex/skills/demo-demo_skill/SKILL.md" <<'EOF'
<!-- GENERATED BY hg-global-skill-sync.sh FROM ~/.claude/skills/demo_skill/SKILL.md; DO NOT EDIT -->
EOF

    cat > "${TEST_HOME}/.agents/skills/demo_demo_skill/SKILL.md" <<'EOF'
<!-- GENERATED BY hg-global-skill-sync.sh FROM ~/.claude/skills/demo_skill/SKILL.md; DO NOT EDIT -->
EOF

    run env \
        HOME="${TEST_HOME}" \
        HG_STUDIO_ROOT="${TEST_STUDIO_ROOT}" \
        HG_WORKSPACE_HOME="${TEST_HOME}" \
        HG_WORKSPACE_OWNER="tester" \
        HG_ANTIGRAVITY_SYNC_SCRIPT="${TEST_ANTIGRAVITY_SYNC}" \
        bash "${SCRIPTS_DIR}/hg-workspace-global-sync.sh" --root "${TEST_STUDIO_ROOT}" --skills-only
    assert_success

    assert [ -f "${TEST_HOME}/.claude/skills/demo_skill/SKILL.md" ]
    assert [ -f "${TEST_HOME}/.claude/skills/demo_skill/.hg-workspace-global-sync.json" ]
    refute [ -e "${TEST_HOME}/.codex/skills/demo-demo_skill" ]
    refute [ -e "${TEST_HOME}/.agents/skills/demo_demo_skill" ]

    run env \
        HOME="${TEST_HOME}" \
        HG_STUDIO_ROOT="${TEST_STUDIO_ROOT}" \
        HG_WORKSPACE_HOME="${TEST_HOME}" \
        HG_WORKSPACE_OWNER="tester" \
        HG_ANTIGRAVITY_SYNC_SCRIPT="${TEST_ANTIGRAVITY_SYNC}" \
        bash "${SCRIPTS_DIR}/hg-workspace-global-sync.sh" --root "${TEST_STUDIO_ROOT}" --skills-only --check
    assert_success
    assert_output --partial "Workspace global sync up to date"
}
