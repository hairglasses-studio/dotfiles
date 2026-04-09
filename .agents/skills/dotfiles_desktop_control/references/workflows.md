# Dotfiles Desktop Control Workflow Reference

This skill is the read-first path for live desktop control after the desktop config is already in place.

## Capability Gate

- Confirm `dotfiles_desktop_status` before any OCR, click, or text-entry path.
- Use `dotfiles_rice_check` alongside it so desktop runtime issues and visual-service issues are separated early.

## Targeting And Visibility

- Use Hyprland reads (`hypr_list_windows`, `hypr_get_monitors`) to locate the right surface before acting.
- Use `screen_screenshot`, `desktop_screenshot_ocr`, and `desktop_find_text` to prove what is visible and where it lives.
- Do not click or type into a surface you have not positively identified.

## Write Discipline

- Prefer the smallest write path that matches the task: focus, type, click, or one targeted reload.
- Reserve `dotfiles_cascade_reload` for layered desktop refreshes after the stale layer is already clear.
- Re-check visible state after any write, not just command success.

## Terminal Visual Pipeline

- Kitty remains the terminal shader write target for the managed visual pipeline.
- Ghostty remains a state-aware companion config surface and compatibility view, not a parallel shader controller.
- Use shader reads before changing visuals, and capture screenshots when a shader change is part of the task.
