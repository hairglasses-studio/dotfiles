---
paths:
  - "hypr/**"
  - "hyprland/**"
  - "etc/modprobe.d/**"
---

# NVIDIA + Wayland tuning

Target stack: RTX 3090 (GA102/Ampere), Hyprland 0.54.2, Samsung LC49G95T on
DP-2 @ 5120x1440 (DSC-required) + XEC ES-G32C1Q on DP-3 portrait @
2560x1440x180 (transform=3).

## Driver state

- **Pinned stack (working)**: `linux612 6.12.77-1` + `linux612-nvidia 590.48.01-15` (proprietary) + Hyprland `0.54.2` **without** VRR / explicit_sync / aggressive env config. Verified 2026-04-16.
- **Real root cause of the 2026-04-16 regression**: the VRR + sync config block (`misc:vrr=2`, per-monitor `vrr,2`, `AQ_NO_MODIFIERS=1`, `GBM_BACKEND=nvidia-drm`, `__GL_GSYNC_ALLOWED=1`, `__GL_VRR_ALLOWED=1`) triggers `nvidia-modeset: ERROR: ... planePitch ...` when applied at **compositor startup**. Applying the same settings at runtime via `hyprctl keyword` does NOT reproduce it — different allocation path. This makes live A/B testing unreliable for the VRR config; the only honest signal is a clean session spawn after a login cycle.
- **Known-bad driver variant**: `linux*-nvidia-open 590.48.01` — surface-registration bug upstream. Swap to proprietary `linux*-nvidia` at the same version.
- **linux619 is installed but unused.** It was tested with the VRR config still active and reproduced planePitch. Because VRR is the confirmed startup trigger and we never retested 6.19 without it, we cannot attribute the 6.19 failure to the kernel alone. Treat 6.19 as "untested clean; revisit only when NVIDIA 595.x lands in Manjaro."
- **Pin GRUB saved default to the 6.12 entry** (`grub-set-default "Advanced options for Manjaro Linux>Manjaro Linux (Kernel: 6.12.77-1-MANJARO x64)"`) so routine updates don't silently boot you into 6.19.
- The aspirational "240Hz VRR" stack stays deferred. Re-attempt after NVIDIA 595.x ships, and A/B test **only via fresh login cycles**, not `hyprctl keyword`.

## Safe baseline (what's committed and working)

### `hyprland.conf` — `env` block

```ini
env = LIBVA_DRIVER_NAME,nvidia
env = __GLX_VENDOR_LIBRARY_NAME,nvidia
env = __GL_GSYNC_ALLOWED,0
env = __GL_VRR_ALLOWED,0
env = NVD_BACKEND,direct
env = GSK_RENDERER,ngl
```

Notable **omissions** vs. the old aggressive baseline:
`GBM_BACKEND=nvidia-drm`, `ELECTRON_OZONE_PLATFORM_HINT=auto`,
`AQ_NO_MODIFIERS=1`. Re-introducing them individually is the next step —
see "Re-introduction protocol".

### `hyprland.conf` — compositor blocks

```ini
misc {
    vfr = false     # load-bearing for Hypr-DarkWindow animated shaders
    # vrr intentionally omitted — defaults to 0 (off)
}

render {
    direct_scanout = 1
    # explicit_sync / explicit_sync_kms intentionally omitted — defaults to auto.
    # Forcing either to 2 triggers nvidia-modeset planePitch ERROR on 590.48.01.
}

cursor {
    no_hardware_cursors = 0   # keep at 0; flipping to true did not help and
                              # the deprecated WLR_NO_HARDWARE_CURSORS env is
                              # still exported by the session start script.
}

debug {
    damage_tracking = 2
    overlay = false
}
```

### `monitors.conf`

```
monitor = DP-2, 5120x1440@239.76, 4596x271, 2
monitor = DP-3, 2560x1440@180,    7156x0,  2, transform, 3
```

Positions align both monitors' logical bottom edges so bottom-anchored
layer-shell surfaces line up. No `vrr` flag on either monitor until the
driver-level regression is resolved.

## Known regression — 2026-04-16, driver 590.48.01

Commit `5c119fb feat(hyprland): NVIDIA + 240Hz frame pacing foundation`
introduced **five** surface-allocation changes simultaneously (explicit_sync=2,
explicit_sync_kms=2, misc:vrr=2, per-monitor vrr flags, plus new env vars).
First boot after the commit produced `nvidia-modeset: ERROR: Invalid request
parameters, planePitch or rmObjectSizeInBytes, passed during surface
registration` at Wayland session start. After the first failure, the Samsung
LC49G95T dropped out of its DSC-dependent EDID modes — `hyprctl monitors`
reported both displays stuck at `0x0@60Hz` and `availableModes` no longer
listed `5120x1440`. Symptoms:

1. Black screen on Hyprland session start.
2. `hyprctl` IPC still responsive — compositor itself was alive, just no surface.
3. Recovery required BOTH a config revert AND a physical power-cycle of the
   Samsung (hold power for ~8s, unplug 10s, replug) to re-handshake DSC.
   Software alone could not recover a monitor that had already failed DSC.

Softening only `explicit_sync_kms = 0` was not enough — the other changes
still triggered planePitch on their own. Full revert + monitor power-cycle
was the minimum repro-free fix.

## Re-introduction protocol

Before committing any change below to `hyprland.conf`, A/B test at runtime:

```bash
# snapshot the current value
hyprctl getoption render:explicit_sync -j | jq

# toggle it and watch for errors
hyprctl keyword render:explicit_sync 2
sleep 5
journalctl -b 0 -k --since '10 sec ago' | grep -c 'nvidia-modeset: ERROR'
# expect 0; if non-zero, revert:
hyprctl keyword render:explicit_sync 0
```

If the Samsung goes to 0x0 during an A/B test, stop and physically power-cycle
before trying the next setting — leaving the monitor in a failed-DSC state
will mask the cause of subsequent tests.

Suggested re-introduction order (least risk first):

1. `__GL_GSYNC_ALLOWED=1`, `__GL_VRR_ALLOWED=1` — pure env vars, app-side effect.
2. `GBM_BACKEND=nvidia-drm` — env var, affects newly-launched clients only.
3. `AQ_NO_MODIFIERS=1` — documented upstream to *reduce* NVIDIA corruption.
4. Per-monitor `vrr, 2` on DP-3 (XEC portrait, less DSC-sensitive than the Samsung).
5. Per-monitor `vrr, 2` on DP-2 (Samsung ultrawide — highest risk).
6. `misc:vrr = 2` — global (redundant with per-monitor, belt+braces).
7. `render:explicit_sync = 2` — requires driver-level handshake.
8. `render:explicit_sync_kms = 2` — **leave last; confirmed regression on 590.48.01**.

## Kernel-module settings (etc/modprobe.d/)

- `nvidia-nogsp.conf`: `options nvidia NVreg_EnableGpuFirmware=0` —
  disables GSP firmware. Required for multi-monitor detection on nvidia-open / GA102.
- `mhwd-gpu.conf`: `options nvidia NVreg_DynamicPowerManagement=0x00` —
  no runtime power management. Desktop value; laptops use `0x02`.

## Kernel cmdline (rEFInd — `/boot/refind_linux.conf`)

```
nvidia_drm.modeset=1 nvidia_drm.fbdev=1 nvidia.NVreg_PreserveVideoMemoryAllocations=1
```

All three required for clean suspend/resume + Wayland modesetting.

## Monitor recovery — DSC failure

The Samsung LC49G95T (and similar VA ultrawides with DSC-only native modes)
can get stuck exposing only fallback EDID modes once DSC negotiation fails.
`hyprctl dispatch dpms off/on` does NOT recover this — only hardware power
removal does.

Recovery procedure:
1. Exit Hyprland (`hyprctl dispatch exit` or kill the process).
2. Hold monitor power button ~8 seconds until fully off.
3. Unplug monitor power for 10+ seconds.
4. Replug, let the monitor fully boot.
5. Log back in through greetd → fresh Hyprland reads EDID with DSC restored.

## Known-bad settings to avoid

- `WLR_NO_HARDWARE_CURSORS=1` — **deprecated** by Hyprland; use
  `cursor { no_hardware_cursors = true }` in hyprland.conf instead. However,
  on driver 590.48.01 even the config-form flip triggered downstream issues,
  so keep at `0` until tested.
- `__GL_SYNC_TO_VBLANK=1` — no longer needed on Wayland (compositor governs vblank).
- `__NV_PRIME_RENDER_OFFLOAD` — hybrid-GPU laptops only; unset on desktop.
- `KWIN_DRM_NO_AMS` — KWin-specific, irrelevant to Hyprland.
- **`render:explicit_sync_kms = 2` on nvidia-open 590.48.01** — confirmed
  triggers planePitch surface-registration errors. Revisit on next driver bump.

## Profiling tools

- `hyprctl debug:overlay 1` — compositor FPS/frametime overlay
- `nvtop` — NVIDIA engine utilization + VRAM + power
- `mangohud` — per-app Vulkan/OpenGL frame timing
- `gamescope -W W -H H --expose-wayland -- app` — microsecond frame timing via mangoapp
- `hyprctl monitors -j` — per-monitor state, DSC/EDID sanity check

## Perf-mode toggle

`hypr-perf-mode.sh` (keybind `$mod CTRL ALT Q`) live-toggles between
`quality` and `performance` (reduced blur, no shadow, vfr on, overlay on).
State persisted at `~/.local/state/hypr-perf-mode`.

## Source of record

- Hyprland wiki: `wiki.hypr.land/Nvidia/`, `/Configuring/Monitors/`,
  `/Hypr-Ecosystem/aquamarine/`
- Hyprland v0.54.2 release notes
- NVIDIA driver 590 release notes
- Session logs 2026-04-16 `investigate-boot-issues-linear-wombat` documenting
  the planePitch + DSC failure chain

Re-validate against upstream on every driver or Hyprland major-version bump —
this file describes a constrained safe state, not the long-term target.
