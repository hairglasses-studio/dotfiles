# Dotfiles Recovery Workflow Reference

This skill replaces the previous narrow catalog around crash recovery, session replay, and handoff generation.

## Recovery Queue

- Start with crash detection and the recovery report to rank dead sessions.
- Prioritize sessions with unpushed commits or uncommitted code.
- Show the recommended next action for each candidate instead of raw tool output.

## Session Replay

- Use session search when the operator has only a topic, branch, or repo name.
- Use session detail and replay before issuing resume guidance.
- Pair the replay with current repo status so stale guidance does not get reused blindly.

## Handoff And Autopsy

- Produce a handoff when work should continue soon on the same branch.
- Produce an autopsy when the important output is the failure mode, missing work, or root cause.
- Include dirty files, branch, last verified command, and open risks.

## Legacy Compression

The previous Claude skills now collapse into this one canonical surface:

- `recover`, `recover-session`, `recovery-dashboard`, `recovery-verify`
- `find-session`, `session-autopsy`, `session-diff`, `session-status`, `handoff`
