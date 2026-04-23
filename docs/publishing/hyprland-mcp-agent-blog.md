---
title: "Controlling Hyprland with an AI Agent via MCP"
published: false
tags: [hyprland, mcp, linux, devtools]
---

# Controlling Hyprland with an AI Agent via MCP

Most AI coding agents are still trapped at the command line. They can edit files, run tests, and call APIs, but the desktop itself is usually invisible to them. That is a missed opportunity on Linux, where the compositor, shell, terminal, notification daemon, and systemd user services are already scriptable.

This dotfiles repo treats the workstation as an MCP-addressable control plane. The agent does not just say "edit the Hyprland config." It can inspect the current monitor layout, capture a screenshot, reload the compositor, read config errors, and decide whether the change actually worked.

The current checked-in `dotfiles-mcp` contract snapshot exposes:

- 434 tools
- 41 registered modules
- 25 resources
- 13 prompts

Only the discovery surface loads eagerly by default. Everything else is deferred, so the agent starts with a small searchable catalog instead of dragging hundreds of workstation tools into context on every prompt.

## Why MCP Fits a Wayland Workstation

Hyprland is already built around IPC. The problem is not that the desktop lacks automation hooks. The problem is that raw shell snippets are brittle:

- `hyprctl` output needs parsing.
- Monitor edits can leave you with a broken display layout.
- UI state often requires screenshots or accessibility-tree inspection.
- Service reload order matters when Hyprland, bars, and notification daemons depend on each other.

MCP gives those operations names, schemas, and guardrails. Instead of asking an agent to invent commands from memory, the server exposes focused tools like:

```text
dotfiles_tool_search
dotfiles_desktop_status
dotfiles_workspace_scene
dotfiles_cascade_reload
desktop_form_fields
desktop_fill_form
session_wait_ready
session_status
```

The agent can search the catalog, inspect the exact schema, call the tool, and react to structured output.

## A Practical Loop

A typical Hyprland config iteration looks like this:

1. Read the current monitor and workspace state.
2. Snapshot the risky config before editing.
3. Apply the smallest persistent config change.
4. Reload Hyprland.
5. Check `configerrors`.
6. Inspect the live monitor state.
7. Keep the change only if the runtime state matches the intent.

That loop is especially important for monitor, VRR, and NVIDIA-sensitive settings. A bad edit is not just a failed unit test. It can make the workstation uncomfortable or unusable until the compositor is repaired.

The same pattern works for visual rice changes:

1. Capture a screenshot.
2. Adjust the bar, shader, gap, or palette config.
3. Reload only the affected service.
4. Capture again.
5. Compare the result.

The interesting part is not that an agent can run `hyprctl reload`. The interesting part is that it can treat the desktop like any other tested subsystem: observe, mutate, verify, and roll back when needed.

## Discovery First

The server uses a discovery-first shape. The initial tool surface is intentionally small:

```text
dotfiles_tool_search
dotfiles_tool_schema
dotfiles_tool_catalog
dotfiles_tool_stats
dotfiles_server_health
dotfiles_desktop_status
dotfiles_workstation_diagnostics
dotfiles_workspace_scene
```

That lets the agent ask for "monitor", "screenshot", "kitty", "systemd", or "GitHub stars" and load only the relevant tools. It also makes the server usable in clients where loading a 400-tool catalog at startup would be wasteful.

## What This Enables

This setup changes the kinds of tasks that are reasonable to delegate:

- "Move the HDMI monitor back to landscape and verify Hyprland accepted it."
- "Open a project layout with an editor, terminal, and docs browser."
- "Screenshot the active window, OCR the error, and draft an issue."
- "Rotate to a lighter terminal shader and confirm the active shader changed."
- "Audit the desktop control plane and tell me which service is stale."

Those are not pure coding tasks, but they are still engineering tasks. They involve state, constraints, verification, and rollback.

## Design Lessons

Three patterns mattered most.

First, make read tools cheap and obvious. Agents should inspect before they mutate. `dotfiles_workspace_scene` and `dotfiles_desktop_status` are front doors, not convenience wrappers.

Second, keep dangerous actions explicit. Batch workflows default to dry-run where possible, and risky compositor changes are paired with snapshot and reload checks.

Third, do not rely on memory. The repo carries skills for desktop UI work, risky Hyprland config changes, screenshot workflows, and repo audits. The agent can use those workflow docs alongside MCP schemas instead of guessing the correct sequence.

## The Bigger Point

The workstation is becoming another programmable production surface. MCP is useful here because it gives the local machine a typed operational API: tools for state, resources for context, prompts for workflows, and contract snapshots for drift detection.

That does not make the agent magic. It makes the boundary clearer. The agent still has to make good engineering decisions, but now it can see and verify the desktop it is changing.

For Linux users who already live in Hyprland, systemd user services, terminal multiplexers, and shell scripts, that is the difference between "an agent that edits dotfiles" and "an agent that can operate the workstation."
