---
paths:
  - "mcp/mapping/**"
---

mapping conventions:
- Keep mapping workflows deterministic where possible; avoid hidden side effects in read-oriented operations.
- Prefer shared mapping or geometry helpers over copying parsing and transformation logic across commands.
- Keep module-local documentation aligned with any new capabilities or wire-format expectations.
