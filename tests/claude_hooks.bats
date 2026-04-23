#!/usr/bin/env bats
# Tests for the Claude Code hook trilogy:
#   - scripts/claude-tdd-reminder.sh     (PreToolUse advisory)
#   - scripts/claude-verify-gate.sh      (Stop advisory)
#   - scripts/claude-verify-track.sh     (PreToolUse instrumentation)
#   - scripts/claude-session-ledger.sh   (Stop write + SessionStart read)
#   - scripts/claude-marathon-sync.sh    (PostToolUse roadmap sync)
#   - scripts/claude-skill-activate.sh   (context skill suggestions)
#
# Each hook reads JSON from stdin and writes JSON to stdout. Tests feed
# synthetic tool-call envelopes and assert the output shape.

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    # Isolate cache dirs so tests don't pollute the user's real ~/.cache
    export HOME="$BATS_TEST_TMPDIR/home"
    mkdir -p "$HOME"

    export DOTFILES_DIR="$(cd "$BATS_TEST_DIRNAME/.." && pwd)"
    export TDD_HOOK="$DOTFILES_DIR/scripts/claude-tdd-reminder.sh"
    export VERIFY_GATE="$DOTFILES_DIR/scripts/claude-verify-gate.sh"
    export VERIFY_TRACK="$DOTFILES_DIR/scripts/claude-verify-track.sh"
    export LEDGER="$DOTFILES_DIR/scripts/claude-session-ledger.sh"
    export MARATHON_SYNC="$DOTFILES_DIR/scripts/claude-marathon-sync.sh"
    export PHASE_GATE="$DOTFILES_DIR/scripts/claude-phase-gate.sh"
    export CLAUDE_PHASE_GATE_DIR="$BATS_TEST_TMPDIR/phase-gate"
    export SKILL_ACTIVATE="$DOTFILES_DIR/scripts/claude-skill-activate.sh"
    export CLAUDE_SKILL_ACTIVATE_DIR="$BATS_TEST_TMPDIR/skill-activate"
}

teardown() {
    rm -rf "$BATS_TEST_TMPDIR"
}

# --- claude-tdd-reminder.sh ---

@test "tdd-reminder: Go source write with no prior tests → nudge" {
    run bash -c "echo '{\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"/tmp/foo.go\"},\"session_id\":\"t1\"}' | $TDD_HOOK"
    [ "$status" -eq 0 ]
    [[ "$output" == *"systemMessage"* ]]
    [[ "$output" == *"TDD reminder"* ]]
}

@test "tdd-reminder: test file write → silent (and records test state)" {
    run bash -c "echo '{\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"/tmp/foo_test.go\"},\"session_id\":\"t2\"}' | $TDD_HOOK"
    [ "$status" -eq 0 ]
    [[ "$output" == '{"decision":"allow"}' ]]
}

@test "tdd-reminder: source after test in same session → silent" {
    run bash -c "echo '{\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"/tmp/bar_test.go\"},\"session_id\":\"t3\"}' | $TDD_HOOK"
    [ "$status" -eq 0 ]
    run bash -c "echo '{\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"/tmp/bar.go\"},\"session_id\":\"t3\"}' | $TDD_HOOK"
    [ "$status" -eq 0 ]
    [[ "$output" == '{"decision":"allow"}' ]]
    [[ "$output" != *"systemMessage"* ]]
}

@test "tdd-reminder: generated file (_gen.go) → silent, no reminder" {
    run bash -c "echo '{\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"/tmp/types_gen.go\"},\"session_id\":\"t4\"}' | $TDD_HOOK"
    [ "$status" -eq 0 ]
    [[ "$output" == '{"decision":"allow"}' ]]
}

@test "tdd-reminder: non-Go file → silent" {
    run bash -c "echo '{\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"/tmp/README.md\"},\"session_id\":\"t5\"}' | $TDD_HOOK"
    [ "$status" -eq 0 ]
    [[ "$output" == '{"decision":"allow"}' ]]
}

@test "tdd-reminder: Read tool → silent" {
    run bash -c "echo '{\"tool_name\":\"Read\",\"tool_input\":{\"file_path\":\"/tmp/foo.go\"},\"session_id\":\"t6\"}' | $TDD_HOOK"
    [ "$status" -eq 0 ]
    [[ "$output" == '{"decision":"allow"}' ]]
}

# --- claude-verify-gate.sh + claude-verify-track.sh ---

@test "verify-gate: no session activity → stop silent" {
    run bash -c "echo '{\"session_id\":\"v1\"}' | $VERIFY_GATE"
    [ "$status" -eq 0 ]
    [[ "$output" == "{}" ]]
}

@test "verify-track+gate: source without tests → stop reminds" {
    run bash -c "echo '{\"session_id\":\"v2\",\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"/tmp/foo.go\"}}' | $VERIFY_TRACK"
    [ "$status" -eq 0 ]
    run bash -c "echo '{\"session_id\":\"v2\"}' | $VERIFY_GATE"
    [ "$status" -eq 0 ]
    [[ "$output" == *"systemMessage"* ]]
    [[ "$output" == *"Verify-before-complete"* ]]
}

@test "verify-gate: idempotent on repeat stop (reminds once)" {
    run bash -c "echo '{\"session_id\":\"v3\",\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"/tmp/foo.go\"}}' | $VERIFY_TRACK"
    run bash -c "echo '{\"session_id\":\"v3\"}' | $VERIFY_GATE"
    [[ "$output" == *"systemMessage"* ]]
    # Second stop — should be silent because sentinel was written
    run bash -c "echo '{\"session_id\":\"v3\"}' | $VERIFY_GATE"
    [ "$status" -eq 0 ]
    [[ "$output" == "{}" ]]
}

@test "verify-track+gate: source + test command → stop silent" {
    run bash -c "echo '{\"session_id\":\"v4\",\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"/tmp/foo.go\"}}' | $VERIFY_TRACK"
    run bash -c "echo '{\"session_id\":\"v4\",\"tool_name\":\"Bash\",\"tool_input\":{\"command\":\"go test ./...\"}}' | $VERIFY_TRACK"
    run bash -c "echo '{\"session_id\":\"v4\"}' | $VERIFY_GATE"
    [ "$status" -eq 0 ]
    [[ "$output" == "{}" ]]
}

@test "verify-track: pytest command is recognized" {
    run bash -c "echo '{\"session_id\":\"v5\",\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"/tmp/foo.py\"}}' | $VERIFY_TRACK"
    run bash -c "echo '{\"session_id\":\"v5\",\"tool_name\":\"Bash\",\"tool_input\":{\"command\":\"pytest tests/\"}}' | $VERIFY_TRACK"
    run bash -c "echo '{\"session_id\":\"v5\"}' | $VERIFY_GATE"
    [ "$status" -eq 0 ]
    [[ "$output" == "{}" ]]
}

@test "verify-track: test file (*_test.go) doesn't register as source" {
    run bash -c "echo '{\"session_id\":\"v6\",\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"/tmp/foo_test.go\"}}' | $VERIFY_TRACK"
    run bash -c "echo '{\"session_id\":\"v6\"}' | $VERIFY_GATE"
    [ "$status" -eq 0 ]
    [[ "$output" == "{}" ]]
}

# --- claude-session-ledger.sh ---

@test "session-ledger: read with no prior ledger → no-op" {
    cd "$BATS_TEST_TMPDIR"
    mkdir -p noledger && cd noledger
    run bash -c "echo '{\"session_id\":\"l1\"}' | $LEDGER read"
    [ "$status" -eq 0 ]
    [[ "$output" == "{}" ]]
}

@test "session-ledger: write creates YAML file with expected keys" {
    cd "$BATS_TEST_TMPDIR"
    mkdir -p fakerepo && cd fakerepo
    git init -q 2>/dev/null
    git config user.email "test@test.test" 2>/dev/null
    git config user.name "test" 2>/dev/null
    touch README && git add README && git commit -qm "init" 2>/dev/null

    run bash -c "echo '{\"session_id\":\"l2\"}' | $LEDGER write"
    [ "$status" -eq 0 ]

    local ledger_file="$HOME/.cache/claude-session-ledger/project-fakerepo/latest.yaml"
    [ -f "$ledger_file" ]
    grep -q "session_id: \"l2\"" "$ledger_file"
    grep -q "project: \"fakerepo\"" "$ledger_file"
    grep -q "branch:" "$ledger_file"
    grep -q "source_writes:" "$ledger_file"
}

@test "session-ledger: read after write emits additionalContext" {
    cd "$BATS_TEST_TMPDIR"
    mkdir -p fakerepo2 && cd fakerepo2
    git init -q 2>/dev/null
    git config user.email "test@test.test" 2>/dev/null
    git config user.name "test" 2>/dev/null
    touch README && git add README && git commit -qm "init" 2>/dev/null

    run bash -c "echo '{\"session_id\":\"l3\"}' | $LEDGER write"
    [ "$status" -eq 0 ]

    run bash -c "echo '{\"session_id\":\"l3\"}' | $LEDGER read"
    [ "$status" -eq 0 ]
    [[ "$output" == *"additionalContext"* ]]
    [[ "$output" == *"fakerepo2"* ]]
}

@test "session-ledger: invalid mode → exit 1" {
    run bash -c "echo '{\"session_id\":\"l4\"}' | $LEDGER invalid"
    [ "$status" -eq 1 ]
}

# --- claude-marathon-sync.sh ---

@test "settings: marathon sync is wired as a PostToolUse Bash hook" {
    run jq -e \
        '.hooks.PostToolUse[]?
         | select(.matcher == "Bash")
         | .hooks[]?
         | select(.type == "command" and .command == "./scripts/claude-marathon-sync.sh")' \
        "$DOTFILES_DIR/.claude/settings.json"
    [ "$status" -eq 0 ]
}

@test "marathon-sync: phase commit appends one roadmap completion" {
    cd "$BATS_TEST_TMPDIR"
    mkdir -p marathonrepo && cd marathonrepo
    git init -q 2>/dev/null
    git config user.email "test@test.test" 2>/dev/null
    git config user.name "test" 2>/dev/null
    printf '# Roadmap\n\n## Completed Marathon Phases\n\n## Planned\n' > ROADMAP.md
    git add ROADMAP.md && git commit -qm "init" 2>/dev/null

    printf 'change\n' > feature.txt
    git add feature.txt && git commit -qm "feat(ticker): Phase 8 polish ticker" 2>/dev/null

    jq -n --arg cwd "$PWD" '{
        "tool_name": "Bash",
        "tool_input": {"command": "git commit -m \"feat(ticker): Phase 8 polish ticker\""},
        "cwd": $cwd
    }' > hook-input.json

    run bash -c "$MARATHON_SYNC < hook-input.json"
    [ "$status" -eq 0 ]
    [[ "$output" == *"marathon-sync: appended ticker:Phase 8"* ]]
    grep -q '`ticker`: Phase 8' ROADMAP.md

    run bash -c "$MARATHON_SYNC < hook-input.json"
    [ "$status" -eq 0 ]
    [ "$(grep -c '`ticker`: Phase 8' ROADMAP.md)" -eq 1 ]
}

# --- claude-phase-gate.sh ---

@test "phase-gate: default plan phase blocks file writes" {
    run bash -c "echo '{\"session_id\":\"pg1\",\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"/tmp/foo.go\"}}' | $PHASE_GATE"
    [ "$status" -eq 0 ]
    [[ "$output" == *'"decision": "block"'* || "$output" == *'"decision":"block"'* ]]
    [[ "$output" == *"plan phase blocks implementation tools"* ]]
}

@test "phase-gate: implementation requires an explicit review label" {
    run "$PHASE_GATE" mark review --session pg2
    [ "$status" -eq 0 ]

    run "$PHASE_GATE" mark implement --session pg2
    [ "$status" -eq 1 ]
    [[ "$output" == *"requires --reviewed-by"* ]]

    run "$PHASE_GATE" mark implement --session pg2 --reviewed-by human-review
    [ "$status" -eq 0 ]
    [[ "$output" == *'"phase": "implement"'* || "$output" == *'"phase":"implement"'* ]]

    run bash -c "echo '{\"session_id\":\"pg2\",\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"/tmp/foo.go\"}}' | $PHASE_GATE"
    [ "$status" -eq 0 ]
    [[ "$output" == '{"decision":"allow"}' ]]
}

@test "phase-gate: verify phase blocks edits but allows test commands" {
    run "$PHASE_GATE" mark review --session pg3
    [ "$status" -eq 0 ]
    run "$PHASE_GATE" mark implement --session pg3 --reviewed-by human-review
    [ "$status" -eq 0 ]
    run "$PHASE_GATE" mark verify --session pg3
    [ "$status" -eq 0 ]

    run bash -c "echo '{\"session_id\":\"pg3\",\"tool_name\":\"Edit\",\"tool_input\":{\"file_path\":\"/tmp/foo.go\"}}' | $PHASE_GATE"
    [ "$status" -eq 0 ]
    [[ "$output" == *'"decision": "block"'* || "$output" == *'"decision":"block"'* ]]
    [[ "$output" == *"verify phase blocks further edits"* ]]

    run bash -c "echo '{\"session_id\":\"pg3\",\"tool_name\":\"Bash\",\"tool_input\":{\"command\":\"go test ./...\"}}' | $PHASE_GATE"
    [ "$status" -eq 0 ]
    [[ "$output" == '{"decision":"allow"}' ]]
}

@test "phase-gate: marker commands are allowed before implementation" {
    run bash -c "echo '{\"session_id\":\"pg4\",\"tool_name\":\"Bash\",\"tool_input\":{\"command\":\"./scripts/claude-phase-gate.sh mark review --session pg4\"}}' | $PHASE_GATE"
    [ "$status" -eq 0 ]
    [[ "$output" == '{"decision":"allow"}' ]]
}

# --- claude-skill-activate.sh ---

@test "skill-activate: dotfiles context suggests desktop and ops skills once" {
    cd "$BATS_TEST_TMPDIR"
    mkdir -p skillrepo/hyprland skillrepo/scripts skillrepo/mcp/dotfiles-mcp
    touch skillrepo/AGENTS.md skillrepo/install.sh skillrepo/go.mod

    jq -n --arg cwd "$PWD/skillrepo" '{
        "session_id": "sa1",
        "tool_name": "Read",
        "cwd": $cwd
    }' > skill-input.json

    run bash -c "$SKILL_ACTIVATE < skill-input.json"
    [ "$status" -eq 0 ]
    [[ "$output" == *"Skill auto-activation"* ]]
    [[ "$output" == *"dotfiles_ui"* ]]
    [[ "$output" == *"dotfiles_ops"* ]]
    [[ "$output" == *"hg_mcp"* ]]
    [[ "$output" == *"go-check"* ]]

    run bash -c "$SKILL_ACTIVATE < skill-input.json"
    [ "$status" -eq 0 ]
    [[ "$output" == '{"decision":"allow"}' ]]
}

@test "skill-activate: shader file target suggests shader_forge" {
    cd "$BATS_TEST_TMPDIR"
    mkdir -p shaderrepo
    touch shaderrepo/AGENTS.md

    jq -n --arg cwd "$PWD/shaderrepo" '{
        "session_id": "sa2",
        "tool_name": "Edit",
        "cwd": $cwd,
        "tool_input": {"file_path": "/tmp/hg-test.glsl"}
    }' > shader-input.json

    run bash -c "$SKILL_ACTIVATE < shader-input.json"
    [ "$status" -eq 0 ]
    [[ "$output" == *"shader_forge"* ]]
}
