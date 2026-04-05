---
description: "Generate changelog from conventional commits since last tag. $ARGUMENTS: (empty)=preview, 'write'=append to CHANGELOG.md"
user_invocable: true
---

Parse `$ARGUMENTS`:
- **(empty)**: Call `mcp__dotfiles__ops_changelog_generate` — preview changelog markdown
- **"write"**: Call `mcp__dotfiles__ops_changelog_generate` with `write=true` — append to CHANGELOG.md

Display the generated markdown grouped by: Features, Bug Fixes, Breaking Changes, Other.
