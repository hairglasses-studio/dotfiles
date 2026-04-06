---
name: handoff
description: Generate an Agent Handoff Protocol document from the current session and git state. Captures branch, iterations, dirty files, and pending work.
allowed-tools: mcp__dotfiles__ops_session_handoff, mcp__dotfiles__ops_session_status
---

Generate a structured handoff document for the current work session. `$ARGUMENTS` can be:
- (empty) — generate handoff for most recent session
- `--write` — generate and write HANDOFF.md to the repo
- `<session_id>` — generate for a specific ops session

## Workflow

1. Parse `$ARGUMENTS`:
   - If contains `--write`: set `write=true`
   - If contains a session ID (hex string): set `session_id`

2. Call `ops_session_handoff(session_id=<parsed>, write=<parsed>)`

3. Display the generated handoff document directly (it's already formatted markdown)

4. If `write=true`, also show:
```
Handoff written to: {written_path}
```

5. If no session exists:
```
## Handoff

No active ops session found. Generating from git state only.

{handoff content with git state but no session section}
```
