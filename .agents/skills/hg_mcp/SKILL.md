---
name: hg_mcp
description: Studio system health investigation, session handoff, and MCP tool discovery for hg-mcp.
allowed-tools:
  - Bash
  - Read
  - Grep
  - Glob
---

# hg-mcp

Use this skill for studio system investigation, session summaries, and MCP tool discovery in the hg-mcp server (dotfiles/mcp/hg-mcp/).

## Default loop

1. Determine the requested action from context:
   - **investigate** (default): Check studio system health (TouchDesigner, streaming, Unraid, or all)
   - **session**: Create a session handoff summary (recent commits, branch state, priorities)
   - **tools**: Search available MCP tools by keyword

2. For `investigate`:
   - Check the requested system(s): `td` (TouchDesigner), `stream` (streaming), `unraid`, `all`
   - Use consolidated tools when available (e.g., `aftrs_studio_health_full`)
   - Report status as Healthy/Degraded/Critical with metrics table
   - Provide recommendations and quick-fix commands

3. For `session`:
   - Review recent git commits and current branch state
   - Check Roadmap.md for current phase
   - Summarize accomplishments, current state, next steps, blockers

4. For `tools`:
   - Search for MCP tools matching the keyword
   - List matching tools with descriptions and example usage
