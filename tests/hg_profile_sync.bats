#!/usr/bin/env bats

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export TEST_HOME="${BATS_TEST_TMPDIR}/home"
    export TEST_ROOT="${TEST_HOME}/hairglasses-studio"
    export SCRIPTS_REAL="${DOTFILES_DIR}/scripts"

    mkdir -p "${TEST_HOME}" "${TEST_ROOT}/workspace" "${TEST_ROOT}/.github/workflow-templates" "${TEST_ROOT}/mcpkit/.github/workflows"
    ln -s "${DOTFILES_DIR}" "${TEST_ROOT}/dotfiles"

    export HOME="${TEST_HOME}"
    export HG_STUDIO_ROOT="${TEST_ROOT}"
    export GIT_CONFIG_GLOBAL="${TEST_HOME}/.gitconfig"

    git config --global user.name "Codex Test"
    git config --global user.email "codex@example.com"

    install_codexkit_stubs
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
    local scope="${6:-active_first_party}"

    cat > "${TEST_ROOT}/workspace/manifest.json" <<EOF
{
  "version": 2,
  "repos": [
    {
      "name": "${repo_name}",
      "baseline_target": true,
      "scope": "${scope}",
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

write_parity_objectives() {
    local relpath="$1"
    local json="$2"

    mkdir -p "$(dirname "${TEST_ROOT}/docs/${relpath}")"
    printf '%s\n' "${json}" > "${TEST_ROOT}/docs/${relpath}"
}

install_codexkit_stubs() {
    mkdir -p "${TEST_ROOT}/codexkit/scripts"

    cat > "${TEST_ROOT}/codexkit/scripts/provider-settings-sync.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

REPO_PATH="${1:?repo path required}"
shift
REPO_NAME=""
CHECK=false
ALLOW_DIRTY=false
INCLUDE_CODEX_CONFIG=false

while [[ $# -gt 0 ]]; do
  case "$1" in
    --repo-name)
      REPO_NAME="${2:?repo name required}"
      shift 2
      ;;
    --check)
      CHECK=true
      shift
      ;;
    --allow-dirty)
      ALLOW_DIRTY=true
      shift
      ;;
    --include-codex-config)
      shift
      ;;
    *)
      shift
      ;;
  esac
done

REPO_NAME="${REPO_NAME:-$(basename "$REPO_PATH")}"

mkdir -p "${REPO_PATH}/.claude" "${REPO_PATH}/.gemini"

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

extension_required=false
objectives_path="${HG_STUDIO_ROOT}/docs/projects/agent-parity/parity-objectives.json"
if [[ ! -f "$objectives_path" ]]; then
  objectives_path="${HG_STUDIO_ROOT}/docs/projects/codex-migration/parity-objectives.json"
fi
if [[ -f "$objectives_path" ]] && [[ "$(jq -r --arg repo "$REPO_NAME" 'if ((.repo_overrides[$repo] // {}) | has("gemini_extension_scaffold")) then .repo_overrides[$repo].gemini_extension_scaffold elif ((.defaults // {}) | has("gemini_extension_scaffold")) then .defaults.gemini_extension_scaffold else false end' "$objectives_path")" == "true" ]]; then
  extension_required=true
fi

if $extension_required; then
  cat > "${tmpdir}/settings.json" <<'EOF2'
{
  "hooks": {},
  "instructions": ["AGENTS.md"],
  "extensions": [
    {
      "name": "gemini-extension",
      "required": true
    }
  ]
}
EOF2
else
  cat > "${tmpdir}/settings.json" <<'EOF2'
{
  "hooks": {},
  "instructions": ["AGENTS.md"]
}
EOF2
fi

cat > "${tmpdir}/claude-settings.json" <<'EOF2'
{
  "hooks": {}
}
EOF2

if $CHECK; then
  if ! cmp -s "${tmpdir}/settings.json" "${REPO_PATH}/.gemini/settings.json" || ! cmp -s "${tmpdir}/claude-settings.json" "${REPO_PATH}/.claude/settings.json"; then
    if $extension_required; then
      echo "gemini-extension (required)" >&2
    fi
    exit 1
  fi
  exit 0
fi

generated_drift=false
if git -C "${REPO_PATH}" ls-files --error-unmatch .gemini/settings.json >/dev/null 2>&1 && [[ -f "${REPO_PATH}/.gemini/settings.json" ]] && ! cmp -s "${tmpdir}/settings.json" "${REPO_PATH}/.gemini/settings.json"; then
  generated_drift=true
fi
if git -C "${REPO_PATH}" ls-files --error-unmatch .claude/settings.json >/dev/null 2>&1 && [[ -f "${REPO_PATH}/.claude/settings.json" ]] && ! cmp -s "${tmpdir}/claude-settings.json" "${REPO_PATH}/.claude/settings.json"; then
  generated_drift=true
fi
tracked_dirty=false
for rel in .gemini/settings.json .claude/settings.json; do
  if git -C "${REPO_PATH}" ls-files --error-unmatch "$rel" >/dev/null 2>&1 && [[ -n "$(git -C "${REPO_PATH}" status --porcelain -- "$rel" 2>/dev/null)" ]]; then
    tracked_dirty=true
  fi
done

if ! $ALLOW_DIRTY && { $generated_drift || $tracked_dirty; }; then
  echo "skipping dirty gemini-settings" >&2
  exit 1
fi

cp "${tmpdir}/settings.json" "${REPO_PATH}/.gemini/settings.json"
cp "${tmpdir}/claude-settings.json" "${REPO_PATH}/.claude/settings.json"
EOF
    chmod +x "${TEST_ROOT}/codexkit/scripts/provider-settings-sync.sh"

    cat > "${TEST_ROOT}/codexkit/scripts/codex-mcp-sync.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

REPO_PATH="${1:?repo path required}"
shift
DRY_RUN=false
if [[ "${1:-}" == "--dry-run" ]]; then
  DRY_RUN=true
fi

command_line="$(python3 - "${REPO_PATH}/.mcp.json" <<'PY'
import json
import sys

with open(sys.argv[1], "r", encoding="utf-8") as f:
    data = json.load(f)

server = next(iter(data["mcpServers"].values()))
print(server["command"])
print(json.dumps(server.get("args", [])))
PY
)"

command_value="$(printf '%s\n' "$command_line" | sed -n '1p')"
args_value="$(printf '%s\n' "$command_line" | sed -n '2p')"

generated_block=$'\n# BEGIN GENERATED MCP SERVERS: codex-mcp-sync\ncommand = "'"${command_value}"$'"\nargs = '"${args_value}"$'\n# END GENERATED MCP SERVERS: codex-mcp-sync\n'

config_path="${REPO_PATH}/.codex/config.toml"
current="$(cat "$config_path")"
if [[ "$current" == *"$generated_block"* ]]; then
  exit 0
fi

if $DRY_RUN; then
  echo "generated codex mcp block drift"
  exit 0
fi

python3 - "$config_path" "$generated_block" <<'PY'
from pathlib import Path
import sys

path = Path(sys.argv[1])
block = sys.argv[2]
text = path.read_text(encoding="utf-8")
begin = "# BEGIN GENERATED MCP SERVERS: codex-mcp-sync"
end = "# END GENERATED MCP SERVERS: codex-mcp-sync"

if begin in text and end in text:
    prefix = text.split(begin, 1)[0].rstrip()
    text = prefix + "\n" + block.lstrip("\n")
else:
    text = text.rstrip() + block

path.write_text(text if text.endswith("\n") else text + "\n", encoding="utf-8")
PY
EOF
    chmod +x "${TEST_ROOT}/codexkit/scripts/codex-mcp-sync.sh"

    cat > "${TEST_ROOT}/codexkit/scripts/skill-surface-sync.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

REPO_PATH="${1:?repo path required}"
shift || true
CHECK=false
if [[ "${1:-}" == "--check" ]]; then
  CHECK=true
fi

if $CHECK && [[ ! -f "${REPO_PATH}/.agents/skills/surface.yaml" ]]; then
  echo "missing surface manifest" >&2
  exit 1
fi

if [[ -f "${REPO_PATH}/.agents/skills/surface.yaml" ]]; then
  python3 - "${REPO_PATH}" <<'PY'
import json
import pathlib
import sys

repo = pathlib.Path(sys.argv[1])
manifest = json.loads((repo / ".agents/skills/surface.yaml").read_text(encoding="utf-8"))
for skill in manifest.get("skills", []):
    name = skill["name"]
    source = repo / ".agents" / "skills" / name / "SKILL.md"
    if not source.exists():
        continue
    target = repo / ".claude" / "skills" / name / "SKILL.md"
    target.parent.mkdir(parents=True, exist_ok=True)
    target.write_text(source.read_text(encoding="utf-8"), encoding="utf-8")
PY
fi
EOF
    chmod +x "${TEST_ROOT}/codexkit/scripts/skill-surface-sync.sh"

    cat > "${TEST_ROOT}/codexkit/scripts/agent-parity-audit.sh" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail

if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
  cat <<'USAGE'
Usage: hg-agent-parity-audit.sh [--help]
USAGE
  exit 0
fi

echo "stub audit" >&2
exit 0
EOF
    chmod +x "${TEST_ROOT}/codexkit/scripts/agent-parity-audit.sh"
}

prepare_sync_script_copy() {
    local dest_dir="$1"
    mkdir -p "${dest_dir}/lib"
    cp "${SCRIPTS_REAL}/sync-standalone-mcp-repos.sh" "${dest_dir}/sync-standalone-mcp-repos.sh"
    cp "${SCRIPTS_REAL}/mcp-mirror.sh" "${dest_dir}/mcp-mirror.sh"
    cp "${SCRIPTS_REAL}/lib/hg-core.sh" "${dest_dir}/lib/hg-core.sh"
    cp "${SCRIPTS_REAL}/lib/hg-workspace.sh" "${dest_dir}/lib/hg-workspace.sh"
    chmod +x "${dest_dir}/sync-standalone-mcp-repos.sh" "${dest_dir}/mcp-mirror.sh"
}

@test "hg-repo-profile-sync: sync creates Gemini project settings baseline" {
    write_manifest "demo"
    init_repo "demo"
    write_agents "demo"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-repo-profile-sync.sh" sync
    assert_success

    run test -f "${TEST_ROOT}/demo/.gemini/settings.json"
    assert_success

    run grep -F 'AGENTS.md' "${TEST_ROOT}/demo/.gemini/settings.json"
    assert_success
}

@test "hg-agent-parity library: canonical objectives override historical fallback" {
    write_parity_objectives "projects/agent-parity/parity-objectives.json" '{
  "version": 1,
  "defaults": {"gemini_extension_scaffold": false},
  "scope_defaults": {},
  "repo_overrides": {"demo": {"gemini_extension_scaffold": false}}
}'
    write_parity_objectives "projects/codex-migration/parity-objectives.json" '{
  "version": 1,
  "defaults": {"gemini_extension_scaffold": false},
  "scope_defaults": {},
  "repo_overrides": {"demo": {"gemini_extension_scaffold": true}}
}'

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash -lc '
        source "'"${SCRIPTS_REAL}/lib/hg-core.sh"'"
        source "'"${SCRIPTS_REAL}/lib/hg-agent-parity.sh"'"
        hg_parity_repo_objective_bool demo gemini_extension_scaffold false
    '
    assert_success
    assert_output "false"
}

@test "hg-agent-parity library: explicit objectives path can use historical objectives" {
    write_parity_objectives "projects/codex-migration/parity-objectives.json" '{
  "version": 1,
  "defaults": {"gemini_extension_scaffold": false},
  "scope_defaults": {},
  "repo_overrides": {"demo": {"gemini_extension_scaffold": true}}
}'

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" HG_PARITY_OBJECTIVES_PATH="${TEST_ROOT}/docs/projects/codex-migration/parity-objectives.json" bash -lc '
        source "'"${SCRIPTS_REAL}/lib/hg-core.sh"'"
        source "'"${SCRIPTS_REAL}/lib/hg-agent-parity.sh"'"
        hg_parity_repo_objective_bool demo gemini_extension_scaffold false
    '
    assert_success
    assert_output "true"
}

@test "hg-provider-settings-sync: sync skips dirty generated Gemini settings until allow-dirty is set" {
    write_manifest "demo"
    init_repo "demo"
    write_agents "demo"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-repo-profile-sync.sh" sync
    assert_success

    git -C "${TEST_ROOT}/demo" add -A
    git -C "${TEST_ROOT}/demo" commit -q -m "seed generated provider settings"

    printf '\nmanual drift\n' >> "${TEST_ROOT}/demo/.gemini/settings.json"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-provider-settings-sync.sh" "${TEST_ROOT}/demo" --repo-name demo
    assert_failure
    assert_output --partial "skipping dirty gemini-settings"

    run grep -F "manual drift" "${TEST_ROOT}/demo/.gemini/settings.json"
    assert_success

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-provider-settings-sync.sh" "${TEST_ROOT}/demo" --repo-name demo --allow-dirty
    assert_success

    run grep -F "manual drift" "${TEST_ROOT}/demo/.gemini/settings.json"
    assert_failure
}

@test "hg-provider-settings-sync: required Gemini extensions are reported as required" {
    write_manifest "demo"
    init_repo "demo"
    write_agents "demo"
    write_parity_objectives "projects/agent-parity/parity-objectives.json" '{
  "version": 1,
  "defaults": {"gemini_extension_scaffold": false},
  "scope_defaults": {},
  "repo_overrides": {"demo": {"gemini_extension_scaffold": true}}
}'

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-provider-settings-sync.sh" "${TEST_ROOT}/demo" --repo-name demo --check
    assert_failure
    assert_output --partial "gemini-extension (required)"
    refute_output --partial "gemini-extension (not managed)"
}

@test "hg-provider-settings-sync: wrapper ignores explicit Codex config injection" {
    write_manifest "demo" "canonical" "null" "none" "none" "active_operator"
    init_repo "demo"
    write_agents "demo"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-provider-settings-sync.sh" "${TEST_ROOT}/demo" --repo-name demo --include-codex-config
    assert_success

    run test -f "${TEST_ROOT}/demo/.codex/config.toml"
    assert_failure
}

@test "hg-provider-settings-sync: non-operator wrapper does not inject Codex config" {
    write_manifest "demo"
    init_repo "demo"
    write_agents "demo"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-provider-settings-sync.sh" "${TEST_ROOT}/demo" --repo-name demo
    assert_success

    run test -f "${TEST_ROOT}/demo/.codex/config.toml"
    assert_failure
}

@test "hg-codex-mcp-sync: repo-relative launcher commands are preserved" {
    write_manifest "demo"
    init_repo "demo"
    mkdir -p "${TEST_ROOT}/demo/.codex" "${TEST_ROOT}/demo/scripts/mcp"

    cat > "${TEST_ROOT}/demo/.mcp.json" <<'EOF'
{
  "mcpServers": {
    "demo": {
      "command": "./scripts/mcp/demo-mcp.sh",
      "args": ["alpha", "beta"]
    }
  }
}
EOF

    cat > "${TEST_ROOT}/demo/.codex/config.toml" <<'EOF'
approval_policy = "on-request"
sandbox_mode = "workspace-write"
EOF

    cat > "${TEST_ROOT}/demo/scripts/mcp/demo-mcp.sh" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
    chmod +x "${TEST_ROOT}/demo/scripts/mcp/demo-mcp.sh"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-codex-mcp-sync.sh" "${TEST_ROOT}/demo"
    assert_success

    run grep -F 'command = "./scripts/mcp/demo-mcp.sh"' "${TEST_ROOT}/demo/.codex/config.toml"
    assert_success

    run grep -F 'args = ["alpha", "beta"]' "${TEST_ROOT}/demo/.codex/config.toml"
    assert_success
}

@test "hg-codex-mcp-sync: relative repo path targets caller repo rather than codexkit root" {
    write_manifest "demo"
    init_repo "demo"
    mkdir -p "${TEST_ROOT}/demo/.codex" "${TEST_ROOT}/demo/scripts/mcp"

    cat > "${TEST_ROOT}/demo/.mcp.json" <<'EOF'
{
  "mcpServers": {
    "demo": {
      "command": "./scripts/mcp/demo-mcp.sh"
    }
  }
}
EOF

    cat > "${TEST_ROOT}/demo/.codex/config.toml" <<'EOF'
approval_policy = "on-request"
sandbox_mode = "workspace-write"
EOF

    cat > "${TEST_ROOT}/demo/scripts/mcp/demo-mcp.sh" <<'EOF'
#!/usr/bin/env bash
exit 0
EOF
    chmod +x "${TEST_ROOT}/demo/scripts/mcp/demo-mcp.sh"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash -lc "cd '${TEST_ROOT}/demo' && bash '${SCRIPTS_REAL}/hg-codex-mcp-sync.sh' ."
    assert_success

    run grep -F 'command = "./scripts/mcp/demo-mcp.sh"' "${TEST_ROOT}/demo/.codex/config.toml"
    assert_success
}

@test "hg-codex-bootstrap: initializes default skill surface without surfacekit" {
    write_manifest "demo"
    init_repo "demo"
    write_agents "demo"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-codex-bootstrap.sh" "${TEST_ROOT}/demo" --repo-name demo --allow-dirty --default-skill-name demo_ops --default-skill-description "Demo repo ops"
    assert_success

    run test -f "${TEST_ROOT}/demo/.agents/skills/surface.yaml"
    assert_success

    run test -f "${TEST_ROOT}/demo/.agents/skills/demo_ops/SKILL.md"
    assert_success

    run test -f "${TEST_ROOT}/demo/.codex/config.toml"
    assert_success

    run grep -F '[profiles.readonly_quiet]' "${TEST_ROOT}/demo/.codex/config.toml"
    assert_failure

    run test -f "${TEST_ROOT}/demo/.claude/settings.json"
    assert_success

    run test -f "${TEST_ROOT}/demo/.gemini/settings.json"
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

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-repo-profile-sync.sh" sync
    assert_success
    assert_output --partial "skipping dirty generated agent docs"

    run grep -F "dirty manual edit" "${TEST_ROOT}/demo/CLAUDE.md"
    assert_success

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-repo-profile-sync.sh" sync --allow-dirty
    assert_success
    refute_output --partial "skipping dirty generated agent docs"

    run grep -F "dirty manual edit" "${TEST_ROOT}/demo/CLAUDE.md"
    assert_failure
}

@test "hg-workflow-sync: retired workflow sync ignores legacy workflow arguments" {
    write_manifest "demo" "canonical" "null" "none" "public_ai_review"
    init_repo "demo"
    mkdir -p "${TEST_ROOT}/demo/.github/workflows"

    cat > "${TEST_ROOT}/demo/.github/workflows/claude-review.yml" <<'EOF'
name: old-review
EOF
    printf '\n# dirty\n' >> "${TEST_ROOT}/demo/.github/workflows/claude-review.yml"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-workflow-sync.sh" --repos=demo
    assert_success
    assert_output --partial "Ignoring retired argument: --repos=demo"
    assert_output --partial "Hosted workflow sync is retired under the local-only automation policy"

    run grep -F "# dirty" "${TEST_ROOT}/demo/.github/workflows/claude-review.yml"
    assert_success

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-workflow-sync.sh" --allow-dirty --repos=demo
    assert_success
    assert_output --partial "Ignoring retired argument: --allow-dirty"
    assert_output --partial "Ignoring retired argument: --repos=demo"

    run grep -F "# dirty" "${TEST_ROOT}/demo/.github/workflows/claude-review.yml"
    assert_success
}

@test "provider parity checks agree on drifted Gemini settings" {
    write_manifest "demo"
    init_repo "demo"
    write_agents "demo"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-repo-profile-sync.sh" sync
    assert_success

    git -C "${TEST_ROOT}/demo" add -A
    git -C "${TEST_ROOT}/demo" commit -q -m "seed profile"

    python - <<'PY' "${TEST_ROOT}/demo/.gemini/settings.json"
import json
import sys

path = sys.argv[1]
with open(path, "r", encoding="utf-8") as f:
    data = json.load(f)
data.setdefault("hooks", {})["SessionEnd"] = [{"hooks": [{"type": "command", "command": "echo drift"}]}]
with open(path, "w", encoding="utf-8") as f:
    json.dump(data, f, indent=2)
    f.write("\n")
PY

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-repo-profile-sync.sh" verify
    assert_failure
    assert_output --partial "provider settings out of sync"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-agent-parity-sync.sh" --check
    assert_failure
    assert_output --partial "provider settings sync failed"
}

@test "hg-repo-profile-sync: active first-party repos get lean Codex config" {
    write_manifest "demo"
    init_repo "demo"
    write_agents "demo"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-repo-profile-sync.sh" sync
    assert_success

    run grep -F 'approval_policy = "on-request"' "${TEST_ROOT}/demo/.codex/config.toml"
    assert_success

    run grep -F '[profiles.readonly_quiet]' "${TEST_ROOT}/demo/.codex/config.toml"
    assert_failure
}

@test "hg-repo-profile-sync: active operator repos keep full Codex profile pack" {
    write_manifest "demo" "canonical" "null" "none" "none" "active_operator"
    init_repo "demo"
    write_agents "demo"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-repo-profile-sync.sh" sync
    assert_success

    run grep -F '[profiles.readonly_quiet]' "${TEST_ROOT}/demo/.codex/config.toml"
    assert_success
}

@test "parity audit entrypoints provide help output" {
    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-agent-parity-audit.sh" --help
    assert_success
    assert_output --partial "Usage: hg-agent-parity-audit.sh"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/hg-codex-audit.sh" --help
    assert_success
    assert_output --partial "Usage: hg-agent-parity-audit.sh"
}

@test "codexkit wrappers derive and export studio root from script location" {
    local alt_root="${BATS_TEST_TMPDIR}/derived-root"
    mkdir -p "${alt_root}/dotfiles/scripts/lib" "${alt_root}/codexkit/scripts"
    cp "${DOTFILES_DIR}/AGENTS.md" "${alt_root}/dotfiles/AGENTS.md"
    cp "${SCRIPTS_REAL}/lib/hg-core.sh" "${alt_root}/dotfiles/scripts/lib/hg-core.sh"
    cp "${SCRIPTS_REAL}/hg-agent-parity-audit.sh" "${alt_root}/dotfiles/scripts/hg-agent-parity-audit.sh"

    cat > "${alt_root}/codexkit/scripts/agent-parity-audit.sh" <<EOF
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "\${HG_STUDIO_ROOT:-}" > "${alt_root}/exported-studio-root.txt"
EOF
    chmod +x "${alt_root}/codexkit/scripts/agent-parity-audit.sh" "${alt_root}/dotfiles/scripts/hg-agent-parity-audit.sh"

    run env -u HG_STUDIO_ROOT -u DOTFILES_DIR HOME="${HOME}" bash "${alt_root}/dotfiles/scripts/hg-agent-parity-audit.sh"
    assert_success

    run cat "${alt_root}/exported-studio-root.txt"
    assert_success
    assert_output "${alt_root}"
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

@test "sync-standalone-mcp-repos: skips mirrors marked manual_projection" {
    local sync_copy="${TEST_ROOT}/sync-copy-no-helper"
    prepare_sync_script_copy "${sync_copy}"

    write_manifest "mirror-repo" "mirror" "\"canonical-src\""
    mkdir -p "${TEST_ROOT}/canonical-src"
    printf 'hello\n' > "${TEST_ROOT}/canonical-src/README.md"

    init_repo "mirror-repo"
    printf 'hello\n' > "${TEST_ROOT}/mirror-repo/README.md"

    cat > "${TEST_ROOT}/mirror-parity.json" <<'EOF'
{
  "version": 1,
  "mirrors": [
    {
      "module": "mirror-repo",
      "standalone_repo": "mirror-repo",
      "canonical_path": "canonical-src",
      "purpose": "test mirror",
      "sync_strategy": "manual_projection"
    }
  ]
}
EOF

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" HG_MCP_MIRROR_MANIFEST="${TEST_ROOT}/mirror-parity.json" bash "${sync_copy}/sync-standalone-mcp-repos.sh" check --repos=mirror-repo
    assert_success
    assert_output --partial "sync_strategy=manual_projection"
}

@test "sync-standalone-mcp-repos: uses repo-specific manual projection helper when present" {
    local sync_copy="${TEST_ROOT}/sync-copy-with-helper"
    prepare_sync_script_copy "${sync_copy}"

    write_manifest "mirror-repo" "mirror" "\"canonical-src\""
    mkdir -p "${TEST_ROOT}/canonical-src"
    printf 'module example.com/canonical\n' > "${TEST_ROOT}/canonical-src/go.mod"

    init_repo "mirror-repo"
    mkdir -p "${TEST_ROOT}/mirror-repo/internal/dotfiles"

    cat > "${TEST_ROOT}/mirror-parity.json" <<'EOF'
{
  "version": 1,
  "mirrors": [
    {
      "module": "mirror-repo",
      "standalone_repo": "mirror-repo",
      "canonical_path": "canonical-src",
      "purpose": "test mirror",
      "sync_strategy": "manual_projection"
    }
  ]
}
EOF

    cat > "${sync_copy}/hg-mirror-repo-projection.sh" <<EOF
#!/usr/bin/env bash
printf '%s\n' "\$@" > "${TEST_ROOT}/projection.args"
EOF
    chmod +x "${sync_copy}/hg-mirror-repo-projection.sh"

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" HG_MCP_MIRROR_MANIFEST="${TEST_ROOT}/mirror-parity.json" bash "${sync_copy}/sync-standalone-mcp-repos.sh" check --repos=mirror-repo
    assert_success
    assert_output --partial "repo-specific helper hg-mirror-repo-projection.sh"

    run cat "${TEST_ROOT}/projection.args"
    assert_success
    assert_line --index 0 "check"
    assert_line --index 1 "--refresh-bare-origin"
    assert_line --index 2 "--canonical"
    assert_line --index 4 "--standalone"
}

@test "sync-standalone-mcp-repos: hygiene flags stale local main in bare mirror repos" {
    write_manifest "mirror-repo" "mirror" "\"canonical-src\""
    mkdir -p "${TEST_ROOT}/canonical-src"

    git init -q "${TEST_ROOT}/origin-repo"
    git -C "${TEST_ROOT}/origin-repo" config user.name "Codex Test"
    git -C "${TEST_ROOT}/origin-repo" config user.email "codex@example.com"
    printf 'one\n' > "${TEST_ROOT}/origin-repo/README.md"
    git -C "${TEST_ROOT}/origin-repo" add README.md
    git -C "${TEST_ROOT}/origin-repo" commit -qm "initial"
    git -C "${TEST_ROOT}/origin-repo" branch -M main

    git clone --bare "${TEST_ROOT}/origin-repo" "${TEST_ROOT}/mirror-repo" >/dev/null 2>&1

    printf 'two\n' >> "${TEST_ROOT}/origin-repo/README.md"
    git -C "${TEST_ROOT}/origin-repo" commit -am "advance" -q
    git -C "${TEST_ROOT}/mirror-repo" fetch origin main:refs/remotes/origin/main >/dev/null 2>&1

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/sync-standalone-mcp-repos.sh" hygiene --repos=mirror-repo
    assert_failure
    assert_output --partial "bare main stale"
}

@test "sync-standalone-mcp-repos: hygiene can refresh origin before checking bare mirrors" {
    write_manifest "mirror-repo" "mirror" "\"canonical-src\""
    mkdir -p "${TEST_ROOT}/canonical-src"

    git init -q "${TEST_ROOT}/origin-repo"
    git -C "${TEST_ROOT}/origin-repo" config user.name "Codex Test"
    git -C "${TEST_ROOT}/origin-repo" config user.email "codex@example.com"
    printf 'one\n' > "${TEST_ROOT}/origin-repo/README.md"
    git -C "${TEST_ROOT}/origin-repo" add README.md
    git -C "${TEST_ROOT}/origin-repo" commit -qm "initial"
    git -C "${TEST_ROOT}/origin-repo" branch -M main

    git clone --bare "${TEST_ROOT}/origin-repo" "${TEST_ROOT}/mirror-repo" >/dev/null 2>&1

    printf 'two\n' >> "${TEST_ROOT}/origin-repo/README.md"
    git -C "${TEST_ROOT}/origin-repo" commit -am "advance" -q

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/sync-standalone-mcp-repos.sh" hygiene --refresh-origin --repos=mirror-repo
    assert_failure
    assert_output --partial "bare main stale"
}

@test "sync-standalone-mcp-repos: hygiene can repair stale local main in bare mirror repos" {
    write_manifest "mirror-repo" "mirror" "\"canonical-src\""
    mkdir -p "${TEST_ROOT}/canonical-src"

    git init -q "${TEST_ROOT}/origin-repo"
    git -C "${TEST_ROOT}/origin-repo" config user.name "Codex Test"
    git -C "${TEST_ROOT}/origin-repo" config user.email "codex@example.com"
    printf 'one\n' > "${TEST_ROOT}/origin-repo/README.md"
    git -C "${TEST_ROOT}/origin-repo" add README.md
    git -C "${TEST_ROOT}/origin-repo" commit -qm "initial"
    git -C "${TEST_ROOT}/origin-repo" branch -M main

    git clone --bare "${TEST_ROOT}/origin-repo" "${TEST_ROOT}/mirror-repo" >/dev/null 2>&1

    printf 'two\n' >> "${TEST_ROOT}/origin-repo/README.md"
    git -C "${TEST_ROOT}/origin-repo" commit -am "advance" -q
    git -C "${TEST_ROOT}/mirror-repo" fetch origin main:refs/remotes/origin/main >/dev/null 2>&1

    run env HOME="${HOME}" HG_STUDIO_ROOT="${HG_STUDIO_ROOT}" bash "${SCRIPTS_REAL}/sync-standalone-mcp-repos.sh" hygiene --repair-bare-main --repos=mirror-repo
    assert_success
    assert_output --partial "bare main in sync"
    assert_output --partial "1 bare mirror repos repaired"

    run git -C "${TEST_ROOT}/mirror-repo" rev-parse refs/heads/main refs/remotes/origin/main
    assert_success
    assert_line --index 0 "${lines[1]}"
}
