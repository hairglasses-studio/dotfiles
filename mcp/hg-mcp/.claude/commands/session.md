# Session Summary

Create a summary for handoff to a new Claude Code session.

## Usage
```
/session
```

## Steps

1. Review recent git commits to understand what was worked on
2. Check current branch and any uncommitted changes
3. Review Roadmap.md for current phase and priorities
4. Summarize:
   - What was accomplished this session
   - Current state of the project
   - Next steps / priorities
   - Any blockers or open questions

## Output Format

```markdown
# hg-mcp Session Summary

## Session Date
[Date]

## Accomplished
- Item 1
- Item 2

## Current State
- Version: v0.x
- Branch: main
- Uncommitted changes: yes/no

## Priority Next Steps
1. Step 1
2. Step 2

## Open Questions
- Question 1

## Quick Start for Next Session
\`\`\`bash
cd ~/aftrs-studio/hg-mcp
git status
\`\`\`
```

## Notes

- Save summaries to `~/aftrs-vault/sessions/` if vault exists
- Include any context Luke needs to continue work
- Keep it copy/paste friendly
