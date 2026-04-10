# Shader Pipeline Architecture

The shader pipeline in `hairglasses-studio` provides live, animated GLSL backgrounds and terminal effects using `kitty`, `ghostty`, and `shaderbg`. It is one of the largest curated terminal shader collections available.

## Core Components

1. **Shader Repository (`wallpaper-shaders/`, `kitty/shaders/`, `ghostty/shaders/`)**
   - Fragments of GLSL code mapped to specific visual aesthetics (e.g., CRT, Cyberpunk, Watercolor).
   - Each shader includes metadata defining its computational cost, intended category, and playlist associations.

2. **Playlist Management**
   - Shaders are organized into playlists (e.g., 'focus', 'chill', 'high-energy').
   - The `kitty-shader-playlist.sh` script (or related MCP tool) handles cycling through shaders in a playlist.
   - Playlist state is written to temporary files so the current visual theme can be tracked and restored.

3. **Rendering Engines**
   - **Terminal**: Native support in modern terminals like Kitty or Ghostty to render a background shader fragment.
   - **Desktop**: Using `shaderbg` (or `swaybg` for static fallbacks) to render GLSL shaders directly on the Wayland root surface.

4. **Integration with Desktop Control**
   - `hyprland` binds hotkeys to cycle shaders (`shader_cycle` MCP tool).
   - Changes to shaders can trigger updates in the `eww` bar to reflect the current aesthetic label or palette.
   - Matugen integration can optionally extract color palettes from the active shader/wallpaper to theme the broader UI.

## Performance Budgets

- Shaders are benchmarked to prevent excessive GPU usage. High-cost shaders are excluded from "battery" or "focus" playlists.
- The `mcp/shader-mcp` tools provide endpoints (`shader_build`, `shader_test`) to validate shader compilation before applying them.
