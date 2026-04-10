# Codex Rollout Workflow

## Baseline

- Confirm `AGENTS.md` is canonical and compatibility docs point back to it.
- Add a real `.codex/config.toml` with the shared profile pack.
- Ensure `.github/copilot-instructions.md` exists.
- Propagate `codex-review`, `codex-security`, `codex-structured-audit`, `codex-baseline-guard`, and `ai-dispatch` from the shared templates.

## Skill Surface

- When `.claude/skills` exists, add `.agents/skills/surface.yaml` and a compressed canonical skill.
- Generate `.claude/skills/*` compatibility outputs with `codexkit skills sync <repo_path>`. Use `hg-skill-surface-sync.sh <repo_path>` only when you intentionally want the repo-local compatibility wrapper.
- Do not keep hand-authored `.claude/skills` files in managed repos.

## MCP And Agents

- Preserve generated MCP blocks in `.codex/config.toml` and add the shared profiles around them.
- Add starter subagents only where the repo is active enough to benefit from them.
- Avoid repo-local hooks or plugin surfaces unless the repo already has a real use for them.

## Blockers

- For repos with merge conflicts or unmerged paths, avoid touching the conflicted surface and report the blocked work explicitly.
- For real-machine validation tasks, add the local automation or fixture coverage you can, then document the environment still required for final acceptance.

## Close-Out

- Run targeted syntax and drift checks.
- Refresh the Codex audit output when the changed repos affect inventory or rollout status.
- Summarize what was completed, what remains blocked, and which repos were intentionally left baseline-only.
