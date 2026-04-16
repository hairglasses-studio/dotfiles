# Wayland Status Bar Alternatives — GPU Capability Research

**Date:** 2026-04-16  
**Context:** Evaluating status bar alternatives to ironbar for a Hyprland setup with a cyberpunk aesthetic,
139+ GLSL shaders, Hairglasses Neon palette, and theme-sync.sh pipeline. Current setup: ironbar
(GTK4, TOML config, CSS styling, v0.18.0, 1.3k stars).

---

## 1. AGS v3 / Astal

**Repository:** https://github.com/aylur/ags  
**Stars:** ~3,000 (April 2026)  
**Latest release:** v3.1.2 (April 8, 2026) — actively maintained  
**Language / framework:** TypeScript (55%), Go (30%), Nix (13%). Runtime: GJS (GNOME JavaScript /
SpiderMonkey) on top of the Astal C/Vala library. Widgets are GTK3 or GTK4 depending on the Astal
build. JSX is supported via the Gnim library.  
**License:** GPL-3.0

**What it is:** AGS v3 is a scaffolding CLI and TypeScript/JSX toolkit layered on top of *Astal*, a
GLib/GObject-based suite that provides pre-built Wayland shell primitives (layer shell, workspaces,
notifications, Bluetooth, audio, network, etc.). You write the shell entirely as TypeScript with JSX
— Astal handles the GTK widget bindings.

**GPU / shader capability:** Indirect. GTK4's `GtkGLArea` widget exposes an OpenGL context inside
any GTK4 window; Astal surfaces are GTK4 windows, so you can embed a `GtkGLArea` child and run
arbitrary GLSL fragment shaders there. In GTK 4.16+ a `GtkGLArea` can emit dmabuf textures, enabling
zero-copy compositing via `GtkGraphicsOffload`. This is real but low-level: you write the GL setup
yourself in C (via a Vala/GIR wrapper) or through a JS binding — there is no high-level shader API.
A proof-of-concept GPU-accelerated shell exists (matshell on Astal/AGS, material-design-themed).

**Hyprland compatibility:** First-class. Deep IPC bindings (workspaces, focused window, gestures) are
built into Astal. Majority of the showcase community configs target Hyprland.

**Theming:** CSS via GTK's CSS engine. Because widgets are plain GTK, any GTK-compatible CSS class
works. Dynamic theming via Matugen (wallpaper → palette) is a common pattern. Color variables can be
exported from theme-sync.sh directly to a GTK CSS file.

**Community:** Largest of the scripting-shell frameworks. Thousands of Hyprland dotfiles on
r/unixporn use AGS. Active Discord, frequent releases.

**Advantages vs ironbar:**
- Full programmability — any widget, any layout, arbitrary logic in TypeScript.
- GTK4 GL area support allows genuine in-bar shader rendering (gradients, animated neon effects).
- Astal provides more Hyprland IPC surface than ironbar's built-in modules.
- Large ecosystem of ready-made components (end_4 dotfiles, etc.).

**Disadvantages vs ironbar:**
- No TOML config — requires writing TypeScript code. Migration effort is significant.
- GJS performance ceiling; very complex shaders will block the JS event loop.
- GTK4 GL area setup requires C/Vala glue or verbose GIR calls from JS — not a drop-in
  shader pipeline.
- Runtime dependency on SpiderMonkey + Astal native libraries.

---

## 2. Quickshell

**Repository:** https://github.com/outfoxxed/quickshell (primary: git.outfoxxed.me; GitHub mirror:
quickshell-mirror/quickshell — ~2.3k stars)  
**Latest release:** v0.2.1 (October 12, 2025)  
**Language / framework:** C++ (93%) runtime + QML (3.5%) for configuration. Qt6 / QtQuick scene
graph.  
**License:** LGPL-3.0 / GPL-3.0

**What it is:** A flexible toolkit for building Wayland (and X11) desktop shells — bars, lockscreens,
notification popups, OSD — entirely in QML. The C++ runtime handles layer shell integration; you
write your shell as `.qml` files that are hot-reloaded on save.

**GPU / shader capability:** **Best-in-class among bar frameworks.** Qt Quick's `ShaderEffect` type
is a first-class QML primitive that accepts arbitrary GLSL vertex and fragment shaders. Any property
on the QML item can be bound as a uniform, and Qt's animation system drives uniforms for animated
effects. This means: animated neon gradients, CRT scanlines, plasma backgrounds, and other Shadertoy-
style effects can run directly inside bar widgets with a few lines of QML — no C++ glue needed.
The QtQuick scene graph is GPU-rendered (OpenGL/Vulkan/Metal) end-to-end; the CPU never touches
pixel data. The Vulkan backend (available since Qt 6) provides dmabuf export. Several community
configs already use `ShaderEffect` for glassmorphism and liquid-glass effects. This is the closest
available analogue to "GLSL shaders in the bar" without building from scratch.

**Hyprland compatibility:** Supported. Multiple popular Hyprland dotfiles (end_4, vaxry) use
Quickshell. Hyprland IPC is available via the compositor's socket API from QML.

**Theming:** QML properties + Qt styling. No CSS. Color palettes are passed as QML properties or
loaded from a JSON/env file. Integration with theme-sync.sh requires a small shim that exports
palette env vars as a QML singleton.

**Community:** Rapidly growing. Active Arch Linux forum thread, multiple dotfiles showcases on
unixporn with Quickshell (April 2026). Still in alpha (minor breaking changes expected).

**Advantages vs ironbar:**
- `ShaderEffect` in QML = native GLSL shader support in bar widgets, no external process.
- The animated uniform binding aligns perfectly with the existing shader playlist/cycle workflow.
- Compositor-agnostic (Hyprland, Sway, Niri, etc.).
- Full QML widget model — any layout, responsive sizing, animations with zero extra deps.
- Hyprland live window preview integration.

**Disadvantages vs ironbar:**
- No TOML config — QML is a new language to learn (relatively approachable for JS/TS developers).
- Alpha software; API may break between releases.
- Qt6 dependency is heavier than GTK4 (larger install, different styling ecosystem).
- No CSS theming — GTK CSS stylesheets cannot be reused directly.
- Smaller community than Waybar/AGS today, though growing fast.

---

## 3. Waybar

**Repository:** https://github.com/Alexays/Waybar  
**Stars:** ~11,100 (largest of any Wayland bar project)  
**Latest release:** v0.15.0 (February 6, 2026)  
**Language / framework:** C++ (95%), CSS (0.5%). GTK3-based rendering.  
**License:** MIT

**What it is:** The de-facto standard Wayland status bar. Mature, widely deployed, extremely well
documented, with a large module library (Sway, Hyprland, network, Bluetooth, PulseAudio, MPD,
custom scripts, etc.). JSON config + GTK CSS.

**GPU / shader capability:** None. Waybar renders via GTK3's software (or GPU-composited but
shader-free) renderer. There is no mechanism to run GLSL shaders in Waybar widgets. Custom
"image" modules can display bitmaps but not procedurally generated GPU content. CSS transitions
and animations are CPU-composited.

**Hyprland compatibility:** Good — workspace, focused window, and window icon modules exist.
Not Hyprland-exclusive.

**Theming:** GTK3 CSS. Very flexible for typography, borders, and color. Hot-reload on file change.
Easy to pipe theme-sync.sh output to `~/.config/waybar/style.css`.

**Community:** Largest community by far. Massive library of community themes, dotfiles, and module
scripts. r/unixporn is full of Waybar setups.

**Advantages vs ironbar:**
- Most stable, most documented, most community-supported bar available.
- CSS theming is directly portable from ironbar's `style.css`.
- Widest module coverage out of the box.

**Disadvantages vs ironbar:**
- GTK3 (older toolkit, no GL areas, software-ish rendering).
- Zero GPU shader capability — hard blocker for the neon/cyberpunk aesthetic with live GLSL effects.
- Config is JSON (verbose); no TOML.
- No scripting layer — logic requires external shell scripts.

---

## 4. Fabric

**Repository:** https://github.com/Fabric-Development/fabric  
**Stars:** ~1,300 (April 2026)  
**Latest release:** v0.0.1 (November 2024)  
**Language / framework:** Python (97%). Uses GTK3 via Python-GObject (PyGObject) + gtk-layer-shell.  
**License:** AGPL-3.0

**What it is:** A Python-first desktop widgets framework for building Wayland shell components
(bars, notifications, dashboards). Emphasizes developer experience — full Python, strong typing,
pre-built widget collection. Native Hyprland and X11/Wayland support. Community project Ax-Shell
("hackable shell for Hyprland, powered by Fabric") is the flagship demo.

**GPU / shader capability:** Very limited. Fabric is built on GTK3 Python bindings, which use
Cairo for 2D rendering. Cairo can be hardware-accelerated by the compositor but has no shader API.
There is no built-in GLSL pathway. In theory you can embed a GtkGLArea in Python via PyGObject and
draw OpenGL into it, but this is verbose and unsupported by Fabric's widget API. Unlike AGS (GTK4
GL areas) or Quickshell (ShaderEffect), there is no practical route to live GLSL in Fabric widgets
without significant custom C extension work.

**Hyprland compatibility:** Native — Hyprland IPC is a first-class Fabric integration.

**Theming:** Python code drives widget properties. No CSS theming layer (unlike GTK4 apps). Color
variables can be set via Python constants loaded from an env file, so theme-sync.sh integration is
possible but requires a Python shim.

**Community:** Small but active. Discord community, ~43 forks. Version numbering (v0.0.1 after
224 commits) suggests the project does not follow semver closely; it's pre-1.0 with API churn.

**Advantages vs ironbar:**
- Full Python flexibility — MCP tool output can be piped into any widget trivially.
- Hyprland-native IPC.
- Good for dashboard-style panels with live data.

**Disadvantages vs ironbar:**
- GTK3, no GPU shader capability.
- AGPL license (more restrictive than ironbar MIT / Waybar MIT).
- v0.0.1 release tag despite months of development — signals instability.
- Python runtime overhead; not suitable for animation-heavy bars.
- No CSS theming.

---

## 5. nwg-panel / nwg-bar

**Repository:** https://github.com/nwg-piotr/nwg-panel  
**Stars:** ~769 (April 2026)  
**Latest release:** v0.10.13 (November 27, 2025)  
**Language / framework:** Python (97%). GTK3, gtk-layer-shell.  
**License:** MIT

**What it is:** A GTK3 Python panel with a graphical configuration UI. Designed for Sway and
Hyprland. 17+ modules: system tray, workspaces, taskbar, volume/brightness sliders, clock, custom
buttons, weather, media player.

**GPU / shader capability:** None. GTK3 / Cairo rendering only.

**Hyprland compatibility:** Supported — dedicated Hyprland taskbar and workspace modules.

**Theming:** GTK3 CSS via the graphical configurator or manual file editing.

**Community:** Moderate. Part of the nwg-shell ecosystem (nwg-look, nwg-bar, nwg-drawer).

**Advantages vs ironbar:**
- GUI configurator lowers the barrier for non-technical changes.
- Part of a maintained ecosystem of Wayland shell tools.
- MIT license.

**Disadvantages vs ironbar:**
- GTK3, no GPU/shader support.
- Less customizable than ironbar without editing Python source.
- Python GTK3 is on a slow deprecation path — gtk4 migration not started.
- Weaker MCP / programmatic integration story.

---

## 6. HyprPanel (maintenance mode) / Wayle (successor)

**HyprPanel**  
**Repository:** https://github.com/Jas-SinghFSU/HyprPanel  
**Stars:** ~2,200  
**Status:** Maintenance mode (bugs fixed, no new features). Developer transitioning to Wayle.  
**Language:** TypeScript (86%), SCSS. Built on AGS v2 / GTK3.

**Wayle**  
**Repository:** https://github.com/wayle-rs/wayle  
**Stars:** ~228 (early; v0.2.0 released April 14, 2026)  
**Language:** Rust (90%), GTK4, Relm4.  
**Status:** Very new — HyprPanel to be archived imminently.

**GPU / shader capability:** HyprPanel: none (GTK3). Wayle: GTK4 gives access to GtkGLArea
in principle; Relm4 is a reactive Rust GTK4 framework. No shader API is documented or exposed
in Wayle's current surface. Theming is TOML + CSS color palette. Wayle is too new to evaluate
shader support.

**Hyprland compatibility:** HyprPanel — Hyprland-exclusive. Wayle — compositor-agnostic
(Hyprland, Niri, Sway targeted).

**Theming (Wayle):** Reactive TOML with schema validation (Tombi LSP). TOML color palette
fields map to GTK4 CSS variables — familiar to the current ironbar TOML workflow.

**Community:** HyprPanel community is significant (~216 forks) but project is ending.
Wayle community is minimal (v0.2.0, first stable settings release April 2026).

**Advantages vs ironbar (Wayle):**
- TOML config maps naturally to the current ironbar workflow.
- Rust + GTK4 — same tech stack as ironbar; CSS from ironbar's style.css is partially portable.
- Compositor-agnostic.

**Disadvantages vs ironbar:**
- Wayle is extremely new — treat as pre-alpha for production use.
- No documented GPU/shader capability.
- HyprPanel (the predecessor) is dead; community will migrate slowly.

---

## 7. Custom Layer-Shell Bar (from scratch)

**Protocol:** `wlr-layer-shell-unstable-v1`  
**Rendering options:** Vulkan (wgpu / ash), OpenGL (wayland-egl), or GTK4/Qt6 widget toolkit.

**What it is:** Building a bar entirely from scratch using the Wayland layer shell protocol.
The surface is a normal Wayland client surface pinned to a screen edge; the application renders
into it with whatever graphics API it chooses. This is what any of the above toolkits are
doing under the hood.

**GPU / shader capability:** Full — this is the only option with unrestricted GPU access.
A bar built directly on wgpu or EGL can run arbitrary Vulkan compute shaders or GLSL fragment
shaders at the bar surface level (not just widget-level). The existing 139-shader GLSL library
could render directly into bar background regions. Relevant Rust crates: `wayland-client`,
`wayland-protocols`, `wayland-egl`, `smithay-client-toolkit`, `wgpu`. Layer-shika (Rust + Slint
UI) is emerging as a library for this use case (targeting first stable release late 2026).

**Hyprland compatibility:** Any compositor implementing wlr-layer-shell works.

**Theming:** Entirely custom — you implement whatever theming system you want.

**Community:** No off-the-shelf community; you build the community by open-sourcing the result.

**Development cost:** Very high. A minimal functional bar (workspaces, clock, system tray) would
require thousands of lines of Rust or C++ and months of work. Maintenance burden is significant.

**Advantages:**
- Unlimited GPU capability: full Vulkan/GLSL control, animate anything, zero compositor overhead.
- Custom hot-reload shader pipeline integration (same pattern as existing wallpaper-shaders/).
- Aligns with the "400+ MCP tools" architecture — bar surface can be an MCP-addressable target.

**Disadvantages:**
- Enormous build investment for something ironbar already provides.
- Custom module ecosystem (Bluetooth, network, tray) must be reimplemented.
- Maintenance burden on a solo developer.

---

## Comparison Matrix

| Bar | Stars | Lang | GTK ver | GPU / Shader | Hyprland | Config | CSS | Activity |
|-----|-------|------|---------|--------------|----------|--------|-----|----------|
| ironbar (current) | 1.3k | Rust | GTK4 | Indirect (GTK4 GL area) | First-class | TOML | Yes (hot-reload) | Active, daily-driver |
| Quickshell | ~2.3k | C++/QML | Qt6 | **Native ShaderEffect + GLSL** | Supported | QML | No | Active, alpha |
| AGS v3 / Astal | ~3k | TS/GJS | GTK4 | GLArea (manual setup) | First-class | TypeScript+JSX | Yes | Very active |
| Waybar | 11.1k | C++ | GTK3 | None | Good | JSON | Yes | Very active |
| Fabric | 1.3k | Python | GTK3 | None (Cairo) | Native | Python | No | Active, pre-1.0 |
| HyprPanel | 2.2k | TypeScript | GTK3 | None | Hyprland-only | JSON/SCSS | SCSS | Maintenance mode |
| Wayle | 228 | Rust | GTK4 | None documented | Supported | TOML+CSS | Partial | New, v0.2.0 |
| nwg-panel | 769 | Python | GTK3 | None | Supported | GUI/CSS | Yes | Maintained |
| Custom layer-shell | N/A | Rust/C++ | None | **Unlimited Vulkan/GL** | Any wlroots | Custom | Custom | N/A |

---

## Recommendation

### Primary recommendation: Stay on ironbar, evaluate Quickshell for shader widgets

**For immediate needs (shader aesthetics + theme-sync):**  
The current ironbar setup is solid and shares the same Rust + GTK4 + TOML stack as Wayle (the
emerging successor to HyprPanel). GTK4's `GtkGLArea` is available to ironbar via GTK4's widget
system — the gap is that ironbar does not expose a shader widget type today, but this could be
implemented as a custom ironbar module in Rust using `gtk4::GLArea`. A PR to ironbar adding a
`type = "gl_shader"` module that loads a `.frag` file and renders it as a bar background region
would be the lowest-friction path to live GLSL in the bar without migrating toolkits.

**For GPU-first shader integration (longer term):**  
**Quickshell is the strongest candidate.** It is the only bar framework with native, first-class
GLSL shader support via QML `ShaderEffect`. The workflow maps well to the existing infrastructure:

- Each wallpaper-shaders `.frag` file can be loaded as a `fragmentShader` property in a
  `ShaderEffect` item.
- The shader_cycle / shader_playlist MCP tool output can drive a `iTime`-style uniform via
  QML property bindings with `NumberAnimation`.
- The Hairglasses Neon palette (`THEME_PRIMARY`, `THEME_SECONDARY`, etc.) maps directly to
  QML properties exported from a singleton loaded from `theme/palette.env`.
- Hyprland IPC integration is available for workspaces, focused window, and keybind ticker.

The main migration cost is learning QML (approachable for anyone comfortable with JS/CSS) and
re-implementing ironbar modules in QML. Given that ironbar is already modular and well-understood,
a phased migration (prototype one bar on Quickshell while keeping ironbar on the other monitor)
is low-risk.

**AGS v3 / Astal** is the second-strongest option if the preference is to stay in the GTK/CSS
ecosystem. The TypeScript/JSX workflow is richer and more programmable than QML, and the GLArea
shader path is viable with some Vala/C glue. However, the shader integration story is more
complex than Quickshell's `ShaderEffect`, and the GJS runtime adds overhead.

**Wayle** is worth watching given its TOML config parity with ironbar, but it is too new (v0.2.0,
April 14 2026) to recommend for a production setup. Revisit in 3–6 months.

**Do not migrate to Waybar** — it provides zero GPU/shader capability, regresses to GTK3, and
the cyberpunk aesthetic would be impossible to achieve at the GLSL level.

### Decision matrix for this setup

| Priority | Best option |
|----------|------------|
| Minimum disruption | Stay on ironbar, add GLArea module |
| GPU shader in bar widgets | Quickshell (QML ShaderEffect) |
| Stay in GTK/CSS ecosystem | AGS v3 / Astal |
| TOML config migration | Wayle (too early, revisit Q3 2026) |
| Full GPU freedom | Custom layer-shell (high cost) |
| Largest community | Waybar (no GPU, not recommended) |

---

## Sources

- [AGS GitHub](https://github.com/aylur/ags) — v3.1.2, April 2026
- [Astal documentation](https://aylur.github.io/astal/)
- [Fabric GitHub](https://github.com/Fabric-Development/fabric) — v0.0.1, 1.3k stars
- [Fabric wiki](https://wiki.ffpy.org/)
- [Waybar GitHub](https://github.com/Alexays/Waybar) — 11.1k stars, v0.15.0
- [Quickshell GitHub mirror](https://github.com/outfoxxed/quickshell) — 2.3k stars, v0.2.1
- [Quickshell website](https://quickshell.org/)
- [Qt ShaderEffect docs](https://doc.qt.io/qt-6/qml-qtquick-shadereffect.html)
- [HyprPanel GitHub](https://github.com/Jas-SinghFSU/HyprPanel) — 2.2k stars, maintenance mode
- [Wayle GitHub](https://github.com/wayle-rs/wayle) — 228 stars, v0.2.0
- [nwg-panel GitHub](https://github.com/nwg-piotr/nwg-panel) — 769 stars, v0.10.13
- [Hyprland status bars wiki](https://wiki.hypr.land/Useful-Utilities/Status-Bars/)
- [Ironbar GitHub](https://github.com/JakeStanger/ironbar) — 1.3k stars, v0.18.0
- [Matshell (GPU-accelerated AGS/Astal shell)](https://github.com/Neurarian/matshell)
- [layer-shika Codeberg](https://codeberg.org/waydeer/layer-shika)
- [GTK4 GLArea docs](https://docs.gtk.org/gtk4/class.GLArea.html)
