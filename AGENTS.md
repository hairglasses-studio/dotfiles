# dotfiles — Agent Instructions

> Canonical instructions: AGENTS.md

Desktop automation, local workstation config, MCP server scaffolding, and org-wide repo tooling for `hairglasses-studio`.

## Start Here

- Read [CLAUDE.md](CLAUDE.md) for the detailed architecture map, script catalog, and dotfiles-specific operational caveats.
- Treat [AGENTS.md](AGENTS.md) as the canonical cross-provider instruction file.
- Keep [CLAUDE.md](CLAUDE.md), `GEMINI.md`, and `.github/copilot-instructions.md` as compatibility mirrors of the core guidance here.

## Working Rules

- Use the shared scripts in `scripts/` and `scripts/lib/` instead of ad hoc replacements when touching onboarding, workflow sync, hook, or audit flows.
- Preserve provider-neutral behavior in repo scaffolding. New repo setup should not reintroduce placeholder Codex configs or Claude-only defaults.
- Canonical reusable workflows live under `.agents/skills/`; `.claude/skills/` is a generated compatibility layer, not the source of truth.
- When editing hook-related scripts, document what is repo-local versus what is installed into user scope by tools such as `hgmux`.

## Build And Verify

- Script checks: `bash -n scripts/*.sh scripts/lib/*.sh`
- Migration inventory: `bash ./scripts/hg-codex-audit.sh --write-workspace-cache --write-wiki-docs --write-json`
- Workflow propagation: `bash ./scripts/hg-workflow-sync.sh --dry-run`
- Repo doc generation: `bash ./scripts/hg-agent-docs.sh --source auto .`

## Shared Research Repository

Cross-project research lives at `~/hairglasses-studio/docs/` (git: hairglasses-studio/docs). When launching research agents, check existing docs first and write reusable research outputs back to the shared repo rather than local docs/.

## Explicit Skill Surface

- `.agents/skills/` is the canonical workflow-skill surface for this repo.
- Generated compatibility mirrors under `.claude/skills/` must come from `scripts/hg-skill-surface-sync.sh`.
- `.codex/agents/*.toml` remains the Codex delegation-role surface, separate from workflow skills.
