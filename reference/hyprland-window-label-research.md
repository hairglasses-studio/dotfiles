# Hyprland Window Label Research

Research into rendering custom text labels on/near Hyprland windows without standard titlebars.
Conducted 2026-04-02 against Hyprland 0.54.2.

## Approach 1: Hyprbars Plugin (BEST — fork/modify)

**What it is:** Official Hyprland plugin that renders title bars with text above windows.
**Repo:** https://github.com/hyprwm/hyprland-plugins/tree/main/hyprbars
**Already loaded:** No (but borders-plus-plus from same repo IS loaded)

### Why this is the best approach

- Uses `IHyprWindowDecoration` — the official API for adding visual elements to windows
- Already renders text via Cairo/Pango onto OpenGL textures
- Supports per-window rules: `hyprbars:no_bar`, `hyprbars:bar_color`, `hyprbars:title_color`
- Can be positioned ABOVE or BELOW the window via `getDecorationLayer()`
- Positions can be STICKY (reserves space) or ABSOLUTE (overlays without affecting layout)

### What to modify for a "window label" plugin

1. Instead of rendering `PWINDOW->m_title`, render a custom label string
2. Use `DECORATION_POSITION_ABSOLUTE` instead of STICKY to overlay without shifting window
3. Position at bottom edge instead of top (change `edges` in `getPositioningInfo()`)
4. Make the label text configurable per-window via dynamic windowrules
5. Strip out button handling, drag handling — just render text

### Key architecture (from hyprbars source)

Plugin lifecycle:
```
PLUGIN_INIT() →
  Listen to window.open event →
  For each window: create CHyprBar(window), call HyprlandAPI::addWindowDecoration()

IHyprWindowDecoration interface methods:
  getPositioningInfo()  → declare size/position/edge
  draw()                → queue a CBarPassElement to render pass
  renderPass()          → actual OpenGL rendering (called by pass element)
  updateWindow()        → respond to geometry changes
  damageEntire()        → mark area for redraw
  getDecorationLayer()  → DECORATION_LAYER_UNDER or DECORATION_LAYER_OVER
  getDecorationFlags()  → behavior flags
```

### Text rendering pipeline (Cairo/Pango → OpenGL)

```cpp
// 1. Create Cairo surface
const auto CAIROSURFACE = cairo_image_surface_create(CAIRO_FORMAT_ARGB32, bufferSize.x, bufferSize.y);
const auto CAIRO = cairo_create(CAIROSURFACE);

// 2. Clear surface
cairo_save(CAIRO);
cairo_set_operator(CAIRO, CAIRO_OPERATOR_CLEAR);
cairo_paint(CAIRO);
cairo_restore(CAIRO);

// 3. Layout text with Pango
PangoLayout* layout = pango_cairo_create_layout(CAIRO);
pango_layout_set_text(layout, text.c_str(), -1);
PangoFontDescription* fontDesc = pango_font_description_from_string("sans");
pango_font_description_set_size(fontDesc, fontSize * scale * PANGO_SCALE);
pango_layout_set_font_description(layout, fontDesc);
pango_font_description_free(fontDesc);

pango_layout_set_width(layout, maxWidth * PANGO_SCALE);
pango_layout_set_ellipsize(layout, PANGO_ELLIPSIZE_END);

cairo_set_source_rgba(CAIRO, color.r, color.g, color.b, color.a);

// 4. Center text
int layoutWidth, layoutHeight;
pango_layout_get_size(layout, &layoutWidth, &layoutHeight);
const double xOffset = (bufferSize.x / 2.0 - layoutWidth / PANGO_SCALE / 2.0);
const double yOffset = (bufferSize.y / 2.0 - layoutHeight / PANGO_SCALE / 2.0);

cairo_move_to(CAIRO, xOffset, yOffset);
pango_cairo_show_layout(CAIRO, layout);
g_object_unref(layout);
cairo_surface_flush(CAIROSURFACE);

// 5. Copy to OpenGL texture
const auto DATA = cairo_image_surface_get_data(CAIROSURFACE);
out->allocate();
glBindTexture(GL_TEXTURE_2D, out->m_texID);
glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MAG_FILTER, GL_NEAREST);
glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_MIN_FILTER, GL_NEAREST);
#ifndef GLES2
glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_SWIZZLE_R, GL_BLUE);
glTexParameteri(GL_TEXTURE_2D, GL_TEXTURE_SWIZZLE_B, GL_RED);
#endif
glTexImage2D(GL_TEXTURE_2D, 0, GL_RGBA, bufferSize.x, bufferSize.y, 0, GL_RGBA, GL_UNSIGNED_BYTE, DATA);

// 6. Cleanup
cairo_destroy(CAIRO);
cairo_surface_destroy(CAIROSURFACE);
```

### Positioning system

```cpp
SDecorationPositioningInfo info;
// STICKY = reserves space (pushes window down), ABSOLUTE = overlays
info.policy = DECORATION_POSITION_STICKY;
// Which edge the decoration attaches to
info.edges = DECORATION_EDGE_TOP;
// Priority for ordering among decorations
info.priority = 5000;
info.reserved = true;
// Pixel extents: {{left, top}, {right, bottom}}
info.desiredExtents = {{0, barHeight}, {0, 0}};
```

### Per-window rule registration

```cpp
// In PLUGIN_INIT:
g_pGlobalState->nobarRuleIdx = Desktop::Rule::windowEffects()->registerEffect("hyprbars:no_bar");
g_pGlobalState->barColorRuleIdx = Desktop::Rule::windowEffects()->registerEffect("hyprbars:bar_color");

// Usage in hyprland.conf:
// windowrule = plugin:hyprbars:no_bar, ^(firefox)$
// windowrule = plugin:hyprbars:bar_color rgb(2a2a2a), ^(kitty)$
```

### Build dependencies

- hyprland headers (must match running version exactly)
- pango, pangocairo
- OpenGL (GL/GLES2)
- hyprpm handles version matching automatically


## Approach 2: Hyprbars AS-IS (Quick hack, no code changes)

**Feasibility:** Medium — works TODAY but limited control over label text.

### How it works

1. Install hyprbars via `hyprpm add https://github.com/hyprwm/hyprland-plugins && hyprpm enable hyprbars`
2. Configure a minimal bar that shows the window title:
   ```
   plugin:hyprbars {
       bar_height = 20
       bar_color = rgba(00000000)     # transparent background
       col.text = rgba(57c7ffee)      # cyan text
       bar_text_size = 10
       bar_text_font = Maple Mono NF CN
       bar_text_align = left
       bar_title_enabled = true
       bar_precedence_over_border = true
       enabled = true
   }
   ```
3. Use per-window rules to hide it on non-target windows:
   ```
   windowrule = plugin:hyprbars:no_bar, class:^(?!com\.mitchellh\.ghostty)
   ```

### Limitation

The title bar shows `PWINDOW->m_title` — the actual window title. You can't set arbitrary
label text. BUT Ghostty's window title IS the shell prompt title, so if the shell sets the
title to include the branch name (which starship/zsh already does), this effectively shows
the CWD/branch as a label.

To force a specific title, you could use `hyprctl setprop address:0xADDR title "my label"`
but this is NOT a real Hyprland feature — setprop only supports a limited set of properties.


## Approach 3: Layer-shell overlay (external process)

**What it is:** A separate process that creates a transparent wlr-layer-shell surface
positioned over a specific window, rendering text.

### Architecture

```
[hyprland IPC] → subscribe to window move/resize events
                 ↓
[overlay process] → update layer-shell surface position to match target window
                 ↓
[wlr-layer-shell] → render text label on overlay layer
```

### Implementation options

| Tool | Language | Complexity | Notes |
|------|----------|-----------|-------|
| gtk4-layer-shell | Python/C | Medium | GTK4 + layer-shell bindings, easy text rendering |
| gtk-layer-shell | Python/C | Medium | GTK3 version, more examples available |
| Quickshell | QML/JS | Medium | QtQuick-based, has native Hyprland IPC integration |
| AGS/Astal | TypeScript | Medium | GTK shell framework with Hyprland service |
| Custom C | C | Hard | Direct wlr-layer-shell-unstable-v1 protocol |

### Key limitation

Layer-shell surfaces are screen-anchored, NOT window-anchored. To follow a window:
1. Query window position via `hyprctl clients -j`
2. Subscribe to Hyprland socket2 events (windowtitle, movewindow, etc.)
3. Reposition the layer surface on every window geometry change

This adds latency and complexity. The overlay will visually lag behind fast window moves.

### Relevant libraries

- https://github.com/wmww/gtk-layer-shell (GTK3, C with GObject introspection for Python)
- https://github.com/wmww/gtk4-layer-shell (GTK4 version)
- https://quickshell.org/ (QtQuick, native Hyprland module)
- https://github.com/Aylur/ags (TypeScript/Astal framework)


## Approach 4: eww floating widget

**What it is:** Use eww (ElKowar's Wacky Widgets) to create a small label widget.

### How it works

```lisp
(defwindow window-label
  :monitor 0
  :geometry (geometry :x "100" :y "100" :width "200" :height "24" :anchor "top left")
  :stacking "overlay"
  (label :text "---improve-cwd-visibility---"))
```

### Limitation

- eww windows are layer-shell surfaces anchored to screen edges, not windows
- No native "follow window" capability
- Would need a script polling `hyprctl clients -j` and calling `eww update` to reposition
- Already running eww bar, so infrastructure exists


## Approach 5: Hyprdecor (ninepatch decorations)

**What it is:** Plugin using Android-style 9-patch bitmaps for window decorations.
**Repo:** https://github.com/CapsAdmin/hyprdecor

### Assessment

- Very early stage (3 commits, 5 stars)
- Uses ninepatch images, not text rendering
- Would need significant modification to render dynamic text
- Not recommended for this use case


## Recommendation

**Primary: Fork hyprbars into "hypr-window-label"**

The hyprbars codebase is the perfect starting point because:
1. It already implements `IHyprWindowDecoration` correctly
2. It already renders text with Cairo/Pango
3. It already supports per-window rules
4. It's well-maintained (official Hyprland plugin)

Key modifications needed:
- Replace title text with configurable label text (via windowrule or IPC)
- Use ABSOLUTE positioning to overlay without affecting layout
- Position at bottom edge (or configurable edge)
- Strip ~60% of the code (buttons, drag, touch handling)
- Add IPC command to set/update label text dynamically
- Make bar height auto-size to text

**Secondary: Use hyprbars as-is** with transparent background, relying on Ghostty's
window title (set by shell/starship) as the label. This works TODAY with zero code changes.

**Tertiary: gtk4-layer-shell Python script** that subscribes to Hyprland IPC and
positions a transparent label over the target window. More flexible but adds a separate
process and visual lag.
