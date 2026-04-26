#!/usr/bin/env bats
# Tests for the Claude Code hook trilogy:
#   - scripts/claude-tdd-reminder.sh     (PreToolUse advisory)
#   - scripts/claude-verify-gate.sh      (Stop advisory)
#   - scripts/claude-verify-track.sh     (PreToolUse instrumentation)
#   - scripts/claude-session-ledger.sh   (Stop write + SessionStart read)
#   - scripts/claude-marathon-sync.sh    (PostToolUse roadmap sync)
#   - scripts/claude-phase-gate.sh       (dev-loop phase enforcement)
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

@test "marathon-sync: ignores unrelated tools" {
    run bash -c "echo '{\"tool_name\":\"Read\",\"tool_input\":{\"file_path\":\"README.md\"}}' | $MARATHON_SYNC"
    [ "$status" -eq 0 ]
    [[ "$output" == '{"decision":"allow"}' ]]
}

@test "marathon-sync: marathon_advance appends ROADMAP entry and records docs event" {
    cd "$BATS_TEST_TMPDIR"
    mkdir -p repo fakebin
    cd repo
    git init -q 2>/dev/null
    git config user.email "test@test.test" 2>/dev/null
    git config user.name "test" 2>/dev/null
    cat > ROADMAP.md <<'EOF'
# Roadmap

## Planned

- [ ] Hook wiring
EOF
    git add ROADMAP.md && git commit -qm "init" 2>/dev/null

    cat > "$BATS_TEST_TMPDIR/fakebin/sqlite3" <<'EOF'
#!/usr/bin/env bash
printf 'db=%s\n' "$1" >> "$SQLITE_CAPTURE"
shift
printf '%s\n' "$@" >> "$SQLITE_CAPTURE"
cat >> "$SQLITE_CAPTURE"
exit 0
EOF
    chmod +x "$BATS_TEST_TMPDIR/fakebin/sqlite3"
    touch "$BATS_TEST_TMPDIR/docs.sqlite"

    payload="$(jq -nc --arg cwd "$PWD" '{
      hook_event_name: "PostToolUse",
      tool_name: "mcp__docs-mcp__marathon_advance",
      tool_use_id: "tool-1",
      cwd: $cwd,
      tool_input: {
        id: "demo-marathon",
        summary: "wired roadmap hook",
        tests_passed: true
      },
      tool_response: {
        completed_sprint: {
          index: 2,
          name: "Roadmap hook",
          state: "completed"
        },
        marathon_done: false,
        message: "Sprint 2 complete. Next: Sprint 3"
      }
    }')"

    run env PATH="$BATS_TEST_TMPDIR/fakebin:$PATH" \
      DOCS_MCP_DB="$BATS_TEST_TMPDIR/docs.sqlite" \
      SQLITE_CAPTURE="$BATS_TEST_TMPDIR/sqlite.log" \
      bash -c 'printf "%s" "$1" | "$2"' _ "$payload" "$MARATHON_SYNC"

    [ "$status" -eq 0 ]
    [[ "$output" == *"systemMessage"* ]]
    grep -qF "\`demo-marathon\`: Sprint 2" ROADMAP.md
    grep -qF 'Roadmap hook - wired roadmap hook' ROADMAP.md
    grep -qF 'roadmap_events' "$BATS_TEST_TMPDIR/sqlite.log"
    grep -qF 'marathon_advance' "$BATS_TEST_TMPDIR/sqlite.log"
}

@test "marathon-sync: Bash Phase commit fallback still appends ROADMAP entry" {
    cd "$BATS_TEST_TMPDIR"
    mkdir -p phaserepo && cd phaserepo
    git init -q 2>/dev/null
    git config user.email "test@test.test" 2>/dev/null
    git config user.name "test" 2>/dev/null
    cat > ROADMAP.md <<'EOF'
# Roadmap

## Completed Marathon Phases

## Planned
EOF
    git add ROADMAP.md && git commit -qm "init" 2>/dev/null
    touch phase.txt
    git add phase.txt && git commit -qm "feat(ticker): Phase 3 add hook sync" 2>/dev/null

    payload="$(jq -nc --arg cwd "$PWD" '{
      hook_event_name: "PostToolUse",
      tool_name: "Bash",
      tool_use_id: "tool-2",
      cwd: $cwd,
      tool_input: {command: "git commit -m \"feat(ticker): Phase 3 add hook sync\""},
      tool_response: {success: true}
    }')"

    run bash -c 'printf "%s" "$1" | "$2"' _ "$payload" "$MARATHON_SYNC"
    [ "$status" -eq 0 ]
    [[ "$output" == *"ticker:Phase 3"* ]]
    grep -qF "\`ticker\`: Phase 3" ROADMAP.md
}

# --- claude-phase-gate.sh ---

@test "phase-gate: non dev-loop prompt stays silent" {
    run bash -c "echo '{\"hook_event_name\":\"UserPromptSubmit\",\"prompt\":\"check status\",\"session_id\":\"p0\"}' | $PHASE_GATE"
    [ "$status" -eq 0 ]
    [[ "$output" == '{"decision":"allow"}' ]]
}

@test "phase-gate: dev-loop start blocks writes until approval" {
    run bash -c "echo '{\"hook_event_name\":\"UserPromptSubmit\",\"prompt\":\"/dev-loop\",\"session_id\":\"p1\"}' | $PHASE_GATE"
    [ "$status" -eq 0 ]
    [[ "$output" == *"phase gate active"* ]]

    run bash -c "echo '{\"hook_event_name\":\"PreToolUse\",\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"/tmp/foo.go\"},\"session_id\":\"p1\"}' | $PHASE_GATE"
    [ "$status" -eq 0 ]
    [[ "$output" == *"\"decision\":\"block\""* ]]
    [[ "$output" == *"plan is approved"* ]]
}

@test "phase-gate: approval unlocks writes but ship waits for verify" {
    run bash -c "echo '{\"hook_event_name\":\"UserPromptSubmit\",\"prompt\":\"/dev-loop\",\"session_id\":\"p2\"}' | $PHASE_GATE"
    [ "$status" -eq 0 ]

    run bash -c "echo '{\"hook_event_name\":\"UserPromptSubmit\",\"prompt\":\"approve dev-loop plan\",\"session_id\":\"p2\"}' | $PHASE_GATE"
    [ "$status" -eq 0 ]
    [[ "$output" == *"approved"* ]]

    run bash -c "echo '{\"hook_event_name\":\"PreToolUse\",\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"/tmp/foo.go\"},\"session_id\":\"p2\"}' | $PHASE_GATE"
    [ "$status" -eq 0 ]
    [[ "$output" == '{"decision":"allow"}' ]]

    run bash -c "echo '{\"hook_event_name\":\"PreToolUse\",\"tool_name\":\"Bash\",\"tool_input\":{\"command\":\"git commit -m test\"},\"session_id\":\"p2\"}' | $PHASE_GATE"
    [ "$status" -eq 0 ]
    [[ "$output" == *"\"decision\":\"block\""* ]]
    [[ "$output" == *"verification"* ]]
}

@test "phase-gate: ship command waits for verify even before writes" {
    run bash -c "echo '{\"hook_event_name\":\"UserPromptSubmit\",\"prompt\":\"/dev-loop\",\"session_id\":\"p2b\"}' | $PHASE_GATE"
    run bash -c "echo '{\"hook_event_name\":\"UserPromptSubmit\",\"prompt\":\"approve dev-loop plan\",\"session_id\":\"p2b\"}' | $PHASE_GATE"

    run bash -c "echo '{\"hook_event_name\":\"PreToolUse\",\"tool_name\":\"Bash\",\"tool_input\":{\"command\":\"gh pr create --fill\"},\"session_id\":\"p2b\"}' | $PHASE_GATE"
    [ "$status" -eq 0 ]
    [[ "$output" == *"\"decision\":\"block\""* ]]
    [[ "$output" == *"verification"* ]]
}

@test "phase-gate: successful verify unlocks ship command" {
    run bash -c "echo '{\"hook_event_name\":\"UserPromptSubmit\",\"prompt\":\"/dev-loop\",\"session_id\":\"p3\"}' | $PHASE_GATE"
    run bash -c "echo '{\"hook_event_name\":\"UserPromptSubmit\",\"prompt\":\"approve dev-loop plan\",\"session_id\":\"p3\"}' | $PHASE_GATE"
    run bash -c "echo '{\"hook_event_name\":\"PreToolUse\",\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"/tmp/foo.go\"},\"session_id\":\"p3\"}' | $PHASE_GATE"

    run bash -c "echo '{\"hook_event_name\":\"PostToolUse\",\"tool_name\":\"Bash\",\"tool_input\":{\"command\":\"make test\"},\"tool_response\":{\"success\":true},\"session_id\":\"p3\"}' | $PHASE_GATE"
    [ "$status" -eq 0 ]
    [[ "$output" == *"verification recorded"* ]]

    run bash -c "echo '{\"hook_event_name\":\"PreToolUse\",\"tool_name\":\"Bash\",\"tool_input\":{\"command\":\"git commit -m test\"},\"session_id\":\"p3\"}' | $PHASE_GATE"
    [ "$status" -eq 0 ]
    [[ "$output" == '{"decision":"allow"}' ]]
}

@test "phase-gate: writes after verify require another verify" {
    run bash -c "echo '{\"hook_event_name\":\"UserPromptSubmit\",\"prompt\":\"/dev-loop\",\"session_id\":\"p4\"}' | $PHASE_GATE"
    run bash -c "echo '{\"hook_event_name\":\"UserPromptSubmit\",\"prompt\":\"approve dev-loop plan\",\"session_id\":\"p4\"}' | $PHASE_GATE"
    run bash -c "echo '{\"hook_event_name\":\"PreToolUse\",\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"/tmp/foo.go\"},\"session_id\":\"p4\"}' | $PHASE_GATE"
    run bash -c "echo '{\"hook_event_name\":\"PostToolUse\",\"tool_name\":\"Bash\",\"tool_input\":{\"command\":\"make test\"},\"tool_response\":{\"success\":true},\"session_id\":\"p4\"}' | $PHASE_GATE"
    run bash -c "echo '{\"hook_event_name\":\"PreToolUse\",\"tool_name\":\"Edit\",\"tool_input\":{\"file_path\":\"/tmp/foo.go\"},\"session_id\":\"p4\"}' | $PHASE_GATE"

    run bash -c "echo '{\"hook_event_name\":\"PreToolUse\",\"tool_name\":\"Bash\",\"tool_input\":{\"command\":\"gh pr create --fill\"},\"session_id\":\"p4\"}' | $PHASE_GATE"
    [ "$status" -eq 0 ]
    [[ "$output" == *"\"decision\":\"block\""* ]]
}

@test "phase-gate: stop blocks implementation with unverified writes" {
    run bash -c "echo '{\"hook_event_name\":\"UserPromptSubmit\",\"prompt\":\"/dev-loop\",\"session_id\":\"p5\"}' | $PHASE_GATE"
    run bash -c "echo '{\"hook_event_name\":\"UserPromptSubmit\",\"prompt\":\"approve dev-loop plan\",\"session_id\":\"p5\"}' | $PHASE_GATE"
    run bash -c "echo '{\"hook_event_name\":\"PreToolUse\",\"tool_name\":\"Write\",\"tool_input\":{\"file_path\":\"/tmp/foo.go\"},\"session_id\":\"p5\"}' | $PHASE_GATE"

    run bash -c "echo '{\"hook_event_name\":\"Stop\",\"session_id\":\"p5\"}' | $PHASE_GATE"
    [ "$status" -eq 2 ]
    [[ "$output" == *"\"decision\":\"block\""* ]]
}
