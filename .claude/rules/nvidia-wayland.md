---
paths:
  - "hypr/**"
  - "hyprland/**"
  - "etc/modprobe.d/**"
---

# NVIDIA + Wayland tuning (2026 best practice)

Target stack: RTX desktop (GA102/Ampere+), Hyprland 0.54+, driver 570+
(nvidia-open-dkms), 240Hz Samsung LC49G95T + 180Hz XEC portrait monitor.

## Canonical env vars (hyprland.conf `env =` block)

| Var | Value | Why |
|---|---|---|
| `LIBVA_DRIVER_NAME` | `nvidia` | VA-API → NVDEC for hardware video decode |
| `__GLX_VENDOR_LIBRARY_NAME` | `nvidia` | Force NVIDIA libGL |
| `GBM_BACKEND` | `nvidia-drm` | GBM won; EGLStreams deprecated since driver 495 |
| `ELECTRON_OZONE_PLATFORM_HINT` | `auto` | Electron apps (discord, vscode) use native Wayland |
| `__GL_GSYNC_ALLOWED` | `1` | 240Hz panel supports G-Sync; enable |
| `__GL_VRR_ALLOWED` | `1` | Adaptive refresh on desktop |
| `NVD_BACKEND` | `direct` | NVDEC direct-mode for mpv/ffmpeg |
| `GSK_RENDERER` | `ngl` | GTK4 Vulkan-over-NGL renderer (ironbar). Ticker uses `gl` to work around cairo drawing-area quirks — keep it pinned at the systemd unit level. |
| `AQ_NO_MODIFIERS` | `1` | Resolves ~80% of Aquamarine rendering corruption on NVIDIA. Remove only if DMA-BUF modifiers are confirmed working cleanly. |

## VRR — per-monitor syntax (monitors.conf)

```
monitor = DP-3, 5120x1440@239.76, 1810x280, 2, vrr, 2
monitor = DP-2, 2560x1440@180,    4370x0,   2, transform, 3, vrr, 2
```

VRR mode `2` = fullscreen-only. NVIDIA-safe: prevents Samsung LC49G95T
backlight flicker that occurs when refresh rate swings under desktop use.
Mode `1` (always-on) can work but causes visible brightness drift on most
VA panels.

## Compositor blocks (hyprland.conf)

```ini
misc {
    vfr  = false  # load-bearing for Hypr-DarkWindow animated shaders
    vrr  = 2      # fullscreen-only (redundant with per-monitor, belt+braces)
}

render {
    direct_scanout      = 1  # zero-copy for fullscreen apps (games, video)
    explicit_sync       = 2  # force on; stable with driver 570+
    explicit_sync_kms   = 2  # keep on; flip to 0 if driver 560 regression returns
}

cursor {
    no_hardware_cursors = true  # NVIDIA Wayland HW cursors still broken in 2026
}

debug {
    damage_tracking = 2          # full — required for 240Hz perf
    overlay         = false      # toggle via `hyprctl keyword debug:overlay 1` during tuning
}
```

## Kernel-module settings (etc/modprobe.d/)

- `nvidia-nogsp.conf`: `options nvidia NVreg_EnableGpuFirmware=0` —
  disables GSP firmware. Required for multi-monitor detection on
  nvidia-open / GA102. Do **not** remove.
- `mhwd-gpu.conf`: `options nvidia NVreg_DynamicPowerManagement=0x00` —
  no runtime power management. Keep at `0x00` for desktops; laptops
  should use `0x02` for suspend/resume support.

## Kernel cmdline (GRUB / rEFInd)

```
nvidia_drm.modeset=1 nvidia_drm.fbdev=1 nvidia.NVreg_PreserveVideoMemoryAllocations=1
```

All three are required for clean suspend/resume + Wayland modesetting
on RTX cards.

## Known-bad settings to avoid

- `WLR_NO_HARDWARE_CURSORS=1` — **deprecated**. Use `cursor { no_hardware_cursors = true }` in hyprland.conf instead.
- `__GL_SYNC_TO_VBLANK=1` — no longer needed on Wayland (compositor governs vblank).
- `__NV_PRIME_RENDER_OFFLOAD` — only relevant for hybrid-GPU laptops; leave unset on desktop.
- `KWIN_DRM_NO_AMS` — KWin-specific, irrelevant to Hyprland.

## Runtime A/B testing pattern

```bash
# Baseline
hyprctl getoption misc:vrr -j | jq '.int'
nvtop                                    # note idle GPU %

# Test change
hyprctl keyword misc:vrr 2
# Wait 5-10s for driver to reassess, observe nvtop

# If good, persist to hyprland.conf and reload
hyprctl reload

# If bad, revert
hyprctl keyword misc:vrr 0
```

## Profiling tools (from scripts/hypr-perf-mode.sh and skills/perf_profile)

- `hyprctl debug:overlay 1` — compositor FPS/frametime overlay
- `nvtop` — NVIDIA engine utilization + VRAM + power
- `mangohud` — per-app Vulkan/OpenGL frame timing
- `gamescope -W W -H H --expose-wayland -- app` — microsecond frame timing via mangoapp
- `hyprctl monitors -j` — per-monitor state

## Perf-mode toggle

Use `hypr-perf-mode.sh` (keybind `$mod CTRL ALT Q`) to live-toggle between
`quality` (committed config) and `performance` (reduced blur, no shadow,
vfr on, overlay on). State persisted at `~/.local/state/hypr-perf-mode`.

## Source of record

All claims here trace back to: Hyprland wiki (`wiki.hypr.land/Nvidia/`,
`/Configuring/Monitors/`, `/Hypr-Ecosystem/aquamarine/`), Hyprland v0.52
release notes, NVIDIA driver 570 changelog, Phoronix HDR/color-management
reporting. When driver / compositor major version changes, re-validate
this file against current upstream recommendations.
