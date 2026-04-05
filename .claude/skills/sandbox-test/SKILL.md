---
name: sandbox-test
description: Test dotfile changes in a GPU-accelerated Docker sandbox with visual preview screenshot
allowed-tools: mcp__dotfiles__sandbox_create, mcp__dotfiles__sandbox_start, mcp__dotfiles__sandbox_stop, mcp__dotfiles__sandbox_destroy, mcp__dotfiles__sandbox_screenshot, mcp__dotfiles__sandbox_test, mcp__dotfiles__sandbox_sync, mcp__dotfiles__sandbox_status, mcp__dotfiles__sandbox_list, mcp__dotfiles__sandbox_diff, mcp__dotfiles__sandbox_exec, mcp__dotfiles__sandbox_validate, mcp__dotfiles__sandbox_visual_diff
---

Test dotfile changes in an isolated Docker sandbox with GPU-accelerated Hyprland rendering. `$ARGUMENTS` can be:
- Empty: full cycle (create, start, sync, test, screenshot, destroy)
- `screenshot`: create sandbox, take screenshot, destroy
- `keep`: full cycle but keep the sandbox running for inspection
- A sandbox ID: take a screenshot of an existing sandbox

## Workflow

### Full cycle (default)

1. **Create** — `sandbox_create(profile="full")`
2. **Start** — `sandbox_start(id=<id>)` — waits for Hyprland to be ready
3. **Sync** — `sandbox_sync(id=<id>)` — deploys dotfile symlinks + reloads config
4. **Test** — `sandbox_test(id=<id>, suite="all")` — run BATS + config validation + symlink checks
5. **Screenshot** — `sandbox_screenshot(id=<id>, wait_sec=5)` — capture the rendered desktop
6. **Report**:

```
## Sandbox Test Results

### Tests
| Suite | Status | Details |
|-------|--------|---------|
| bats | pass/fail | {output summary} |
| symlinks | pass/fail | {details} |
| config | pass/fail | {details} |

### Visual Preview
{screenshot displayed inline}

### Status
GPU: {gpu info} | Uptime: {uptime}
```

7. **Destroy** — `sandbox_destroy(id=<id>, force=true)` (unless `keep` mode)

### Screenshot only

1. Create + start sandbox
2. Wait 5 seconds for visual settle
3. Take screenshot and display inline
4. Destroy sandbox

### Existing sandbox (ID argument)

1. `sandbox_screenshot(id="$ARGUMENTS", wait_sec=3)`
2. Display screenshot inline with sandbox status
