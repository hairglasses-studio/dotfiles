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
