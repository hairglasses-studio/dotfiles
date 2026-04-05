---
description: "Show test coverage report with optional threshold gate. $ARGUMENTS: (empty)=report, a number like '80'=gate at 80% minimum"
user_invocable: true
---

Parse `$ARGUMENTS`:
- **(empty)**: Call `mcp__dotfiles__ops_coverage_report` — show per-package coverage breakdown
- **Number (e.g. "80")**: Call `mcp__dotfiles__ops_coverage_report` with `threshold=$ARGUMENTS` — report + fail if below threshold

Display: package | coverage % | functions covered | gate status (PASS/FAIL)
Highlight packages below threshold in red.
