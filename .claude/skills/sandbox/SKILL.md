---
description: "Manage GPU-accelerated Docker sandboxes for isolated dotfile testing. $ARGUMENTS: (empty)=list, 'create'=create+start, 'test'=full validate cycle, 'screenshot <id>'=capture, 'exec <id> <cmd>'=run command, 'destroy <id>'=teardown, 'diff <id>'=compare symlinks"
user_invocable: true
allowed-tools: mcp__dotfiles__sandbox_create, mcp__dotfiles__sandbox_start, mcp__dotfiles__sandbox_stop, mcp__dotfiles__sandbox_destroy, mcp__dotfiles__sandbox_list, mcp__dotfiles__sandbox_status, mcp__dotfiles__sandbox_sync, mcp__dotfiles__sandbox_test, mcp__dotfiles__sandbox_exec, mcp__dotfiles__sandbox_diff, mcp__dotfiles__sandbox_screenshot, mcp__dotfiles__sandbox_visual_diff, mcp__dotfiles__sandbox_validate
---

Parse `$ARGUMENTS`:

- **(empty)** or **"list"**: Call `mcp__dotfiles__sandbox_list` — show all sandboxes with status
- **"create"**: Create and start a new sandbox:
  1. `sandbox_create(profile="full")` — create container with GPU
  2. `sandbox_start(id=<id>)` — wait for Hyprland ready
  3. Report ID and status
- **"test"** or **"validate"**: Full validation cycle (same as /sandbox-test default):
  1. `sandbox_create` → `sandbox_start` → `sandbox_sync` → `sandbox_test(suite="all")` → `sandbox_screenshot` → `sandbox_destroy`
  2. Report test results + screenshot
- **"screenshot <id>"**: Call `sandbox_screenshot(id=<id>, wait_sec=3)` — capture and display
- **"visual-diff <id>"**: Call `sandbox_visual_diff(id=<id>)` — compare against reference screenshot
- **"exec <id> <cmd>"**: Call `sandbox_exec(id=<id>, command=<cmd>)` — run command inside sandbox
- **"sync <id>"**: Call `sandbox_sync(id=<id>)` — deploy dotfiles + reload Hyprland
- **"diff <id>"**: Call `sandbox_diff(id=<id>)` — compare symlink health inside vs outside
- **"status <id>"**: Call `sandbox_status(id=<id>)` — GPU utilization, memory, CPU
- **"stop <id>"**: Call `sandbox_stop(id=<id>)`
- **"destroy <id>"**: Call `sandbox_destroy(id=<id>, force=true)`

Present results in dashboard format. For test results, show per-suite pass/fail table.
