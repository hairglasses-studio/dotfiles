# Dotfiles UI Workflow Reference

This skill replaces the previous narrow catalog around desktop visuals, widgets, shaders, and screenshot review.

## Ironbar And Desktop Services

- Inspect menubar state before restarting it.
- Use cascade reloads when a change affects multiple themed services.
- Report both service health and visible outcome after a reload.

## Hyprland And Interaction

- Use explicit focus, workspace, screenshot, and key-send operations instead of vague UI instructions.
- When a change touches layout or monitor behavior, capture the before/after state.

## Shader And Rice Iteration

- Inspect the current shader state before swapping or cycling.
- Screenshot visual changes after applying them.
- When reviewing rice quality, comment on palette, legibility, spacing, and motion together.

## Legacy Compression

The previous Claude skills now collapse into this one canonical surface:

- `ironbar`, `hyprland`, `shader-browse`, `screenshot`, `screenshot-review`, `screen-record`
- `rice-check`, `sway-rice`, `mapping`, `notify`, `clipboard`, `tmux`
