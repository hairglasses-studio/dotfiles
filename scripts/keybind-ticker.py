#!/usr/bin/env python3
"""keybind-ticker.py — Cyberpunk multi-stream ticker for Hyprland (v3).

GTK4 DrawingArea with PangoCairo at 240Hz frame-clock sync.

Features:
  - Multi-stream content (16 streams): keybinds, system, fleet, weather,
    github, notifications, music, updates, mx-battery, disk, load, workspace,
    claude-sessions, network, audio, shader
  - Visual effects: water caustic, neon glow (breathing), gradient text,
    scanlines, text outline (solid or color-cycling pulse), wave distortion,
    edge fade vignette, typewriter reveal, urgency glitch surge, phosphor
    decay trail, synthwave sweep top border, holo shimmer, ghost echo,
    progress indicator
  - Interactive: scroll (speed) / Shift+scroll (stream switch) /
    middle-click (pause) / right-click (menu) / left-click (copy/open URL),
    hover tooltip, adaptive speed on hover
  - 4 effect presets: ambient, cyberpunk, minimal, clean
  - Per-stream preset + per-stream refresh interval
  - Slow streams (github, music) run on background thread
  - Priority interrupts: critical urgency notifications jump the queue
  - Playlist persistence + order-preserving restore

Usage:
  keybind-ticker.py                     # regular window
  keybind-ticker.py --layer             # layer-shell bar (for systemd)
  keybind-ticker.py --preset minimal    # start with a preset
  keybind-ticker.py --monitor DP-2      # target a different output
  keybind-ticker.py --playlist focus    # load a non-default playlist
  keybind-ticker.py --state-dir ~/.local/state/keybind-ticker-dp3  # isolate state for multi-instance
"""

import gi
import subprocess
import json
import sys
import os
import re
import math
import random
import signal
import glob
import threading
import time as _time
import cairo as _cairo
import numpy as np
from collections import deque
from scipy.ndimage import uniform_filter1d
from html import escape

# Shared rendering helpers (badge / empty / dup / layer-shell / drawing-area).
# Import path handling mirrors the other ticker scripts.
sys.path.insert(0, os.path.join(os.path.dirname(os.path.realpath(__file__)), "lib"))
import ticker_render as tr  # noqa: E402

gi.require_version("Gtk", "4.0")
gi.require_version("Gdk", "4.0")
gi.require_version("Pango", "1.0")
gi.require_version("PangoCairo", "1.0")

from gi.repository import Gtk, Gdk, Pango, PangoCairo, GLib, Gio

# ── CLI arg parsing ───────────────────────────────
LAYER_MODE = "--layer" in sys.argv


def _cli_value(flag, default=None):
    for i, arg in enumerate(sys.argv):
        if arg == flag and i + 1 < len(sys.argv):
            return sys.argv[i + 1]
    return default


MONITOR_NAME = _cli_value("--monitor", "DP-2")
START_PRESET = _cli_value("--preset", "ambient")
START_PLAYLIST = _cli_value("--playlist", None)  # None → resolved at runtime via state file or "main"
STATE_DIR_OVERRIDE = _cli_value("--state-dir", None)

if LAYER_MODE:
    gi.require_version("Gtk4LayerShell", "1.0")
    from gi.repository import Gtk4LayerShell


# ══════════════════════════════════════════════════
# Configuration
# ══════════════════════════════════════════════════

BAR_H = 39
DEFAULT_REFRESH_S = 300
ERROR_BACKOFF_S = 30  # Streams that returned _empty() advance after this instead of their full refresh interval, so a broken stream cannot freeze the ticker.
MAX_DWELL_S = 75  # Cap on how long any single stream stays on screen. Sized so content at ~55 px/s can traverse a 3840px ultrawide (~70s) before rotating. Prevents long cache TTLs (weather=1800, arch-news=3600) from freezing rotation.
MIN_DWELL_S = 12  # Floor so real-time streams (pomodoro, recording) stay readable.
STATE_DIR = os.path.expanduser(STATE_DIR_OVERRIDE) if STATE_DIR_OVERRIDE else os.path.expanduser("~/.local/state/keybind-ticker")
PLAYLIST_DIR = os.path.expanduser(
    "~/hairglasses-studio/dotfiles/ticker/content-playlists")
QUOTES_DIR = os.path.expanduser(
    "~/hairglasses-studio/dotfiles/ticker/quotes")
DEFAULT_PLAYLIST = "main"

# Maple font weight cycle
FONTS = [
    "Maple Mono NF CN 15",
    "Maple Mono NF CN Bold 15",
    "Maple Mono NF CN Italic 15",
    "Maple Mono NF CN SemiBold 15",
    "Maple Mono NF CN Light 15",
    "Maple Mono NF CN Medium 15",
    "Maple Mono NF CN ExtraBold 15",
    "Maple Mono NF CN Thin 15",
    "Maple Mono NF CN ExtraLight 15",
    "Maple Mono NF CN Bold Italic 15",
]

# Hairglasses Neon gradient stops
GRADIENT_COLORS = [
    (0.161, 0.941, 1.000),  # #29f0ff cyan
    (1.000, 0.278, 0.820),  # #ff47d1 magenta
    (0.239, 1.000, 0.710),  # #3dffb5 green
    (1.000, 0.894, 0.369),  # #ffe45e yellow
    (1.000, 0.361, 0.541),  # #ff5c8a pink
    (0.290, 0.659, 1.000),  # #4aa8ff blue
    (0.161, 0.941, 1.000),  # #29f0ff cyan (wrap)
]

URL_RE = re.compile(r"https?://\S+")

# ── Effect presets ────────────────────────────────
PRESETS = {
    "ambient": {
        "speed": 55.0, "gradient_speed": 40.0, "gradient_span": 800.0,
        "glow_kernel": 17, "glow_base_alpha": 0.35, "glow_pulse_amp": 0.15,
        "glow_pulse_period": 3.5, "water_skip": 120, "scanline_opacity": 0.08,
        "shadow_offset": 2, "shadow_alpha": 0.30,
        "glitch_prob": 0.004, "glitch_frames": 4, "ca_offset": 3,
        "bg_alpha": 0.82, "outline_width": 0.8, "outline_pulse": True,
        "wave_amp": 1.5, "wave_freq": 0.015, "wave_speed": 2.0,
        "edge_fade": 120, "progress_bar": True, "phosphor_trail": 0.0,
        "ghost_echo": 0.0, "synthwave_border": True, "holo_shimmer": 0.05,
        "hue_sweep": False,
    },
    "cyberpunk": {
        "speed": 65.0, "gradient_speed": 55.0, "gradient_span": 600.0,
        "glow_kernel": 21, "glow_base_alpha": 0.50, "glow_pulse_amp": 0.25,
        "glow_pulse_period": 2.5, "water_skip": 80, "scanline_opacity": 0.14,
        "shadow_offset": 3, "shadow_alpha": 0.35,
        "glitch_prob": 0.008, "glitch_frames": 5, "ca_offset": 4,
        "bg_alpha": 0.78, "outline_width": 1.0, "outline_pulse": True,
        "wave_amp": 2.0, "wave_freq": 0.02, "wave_speed": 3.0,
        "edge_fade": 100, "progress_bar": True, "phosphor_trail": 0.20,
        "ghost_echo": 0.08, "synthwave_border": True, "holo_shimmer": 0.10,
        "hue_sweep": True,
    },
    "minimal": {
        "speed": 45.0, "gradient_speed": 30.0, "gradient_span": 1000.0,
        "glow_kernel": 13, "glow_base_alpha": 0.20, "glow_pulse_amp": 0.05,
        "glow_pulse_period": 5.0, "water_skip": 0, "scanline_opacity": 0.0,
        "shadow_offset": 1, "shadow_alpha": 0.20,
        "glitch_prob": 0.0, "glitch_frames": 0, "ca_offset": 0,
        "bg_alpha": 0.90, "outline_width": 0.5, "outline_pulse": False,
        "wave_amp": 0, "wave_freq": 0, "wave_speed": 0,
        "edge_fade": 60, "progress_bar": True, "phosphor_trail": 0.0,
        "ghost_echo": 0.0, "synthwave_border": False, "holo_shimmer": 0.0,
        "hue_sweep": False,
    },
    "clean": {
        "speed": 50.0, "gradient_speed": 35.0, "gradient_span": 900.0,
        "glow_kernel": 0, "glow_base_alpha": 0.0, "glow_pulse_amp": 0.0,
        "glow_pulse_period": 1.0, "water_skip": 0, "scanline_opacity": 0.0,
        "shadow_offset": 0, "shadow_alpha": 0.0,
        "glitch_prob": 0.0, "glitch_frames": 0, "ca_offset": 0,
        "bg_alpha": 0.92, "outline_width": 0.0, "outline_pulse": False,
        "wave_amp": 0, "wave_freq": 0, "wave_speed": 0,
        "edge_fade": 0, "progress_bar": True, "phosphor_trail": 0.0,
        "ghost_echo": 0.0, "synthwave_border": False, "holo_shimmer": 0.0,
        "hue_sweep": False,
    },
}


def load_preset(name):
    p = PRESETS.get(name, PRESETS["ambient"])
    return type("P", (), dict(p))()


# ══════════════════════════════════════════════════
# Content Streams
# ══════════════════════════════════════════════════

_badge = tr.badge
_empty = tr.empty
_dup = tr.dup


# ══════════════════════════════════════════════════
# Stream registry + metadata
# ══════════════════════════════════════════════════

STREAMS: dict = {}

# Per-stream metadata: effect preset override + refresh interval (seconds).
# Populated entirely by the plugin loader below (scripts/lib/ticker_streams/
# *.py) and the TOML catalogue (ticker/streams.toml). Slow-threaded streams
# declare `META["slow"] = True` in their plugin module.
STREAM_META: dict = {}
SLOW_STREAMS: set = set()

FALLBACK_ORDER = [
    "keybinds", "system", "fleet", "weather", "github",
    "notifications", "music", "updates", "mx-battery",
    "disk", "load", "cpu", "gpu", "top-procs", "uptime", "tmux", "workspace",
    "claude-sessions", "network", "audio", "shader", "ci", "hacker",
    "calendar", "pomodoro", "token-burn", "dirty-repos", "failed-units",
    "arch-news", "smart-disk", "wifi-quality", "container-status",
    "net-throughput", "kernel-errors", "recording",
    "hn-top", "github-prs", "weather-alerts", "cve-alerts",
]


# ══════════════════════════════════════════════════
# Declarative TOML catalogue (ticker/streams.toml)
#
# Streams declared there override same-named inline builders; new names
# get appended to FALLBACK_ORDER. Keeps the data-definition boilerplate
# out of this file for cache-fed streams that follow a fixed pattern.
# ══════════════════════════════════════════════════

def _load_toml_catalogue():
    import ticker_streams as _ts
    toml_path = os.path.expanduser(
        "~/hairglasses-studio/dotfiles/ticker/streams.toml")
    try:
        toml_builders, toml_meta, toml_order = _ts.load_toml_streams(toml_path)
    except Exception as _e:
        sys.stderr.write(f"streams.toml load failed: {_e}\n")
        return
    for name, fn in toml_builders.items():
        STREAMS[name] = fn
    for name, m in toml_meta.items():
        STREAM_META[name] = {"preset": m.get("preset"),
                             "refresh": int(m.get("refresh", 300))}
        if "dwell" in m:
            STREAM_META[name]["dwell"] = m["dwell"]
    for name in toml_order:
        if name not in FALLBACK_ORDER:
            FALLBACK_ORDER.append(name)


_load_toml_catalogue()


# ══════════════════════════════════════════════════
# Bundled plugin catalogue (scripts/lib/ticker_streams/*.py)
#
# Complex streams live as individual Python modules with a standard
# META + build() contract. The loader imports each and registers the
# resulting callable into the same STREAMS / STREAM_META / FALLBACK_ORDER
# / SLOW_STREAMS structures the inline builders use. Plugins override
# same-named inline builders to allow incremental extraction.
# ══════════════════════════════════════════════════

def _load_bundled_plugins():
    import ticker_streams as _ts
    pkg_dir = os.path.join(os.path.dirname(os.path.realpath(__file__)),
                           "lib", "ticker_streams")
    try:
        builders, meta, order, slow = _ts.load_plugin_streams(pkg_dir)
    except Exception as _e:
        sys.stderr.write(f"bundled plugin load failed: {_e}\n")
        return
    for name, fn in builders.items():
        STREAMS[name] = fn
    for name, m in meta.items():
        STREAM_META[name] = {"preset": m.get("preset"),
                             "refresh": int(m.get("refresh", 300))}
        if "dwell" in m:
            STREAM_META[name]["dwell"] = m["dwell"]
    for name in order:
        if name not in FALLBACK_ORDER:
            FALLBACK_ORDER.append(name)
    SLOW_STREAMS.update(slow)


_load_bundled_plugins()


# ══════════════════════════════════════════════════
# Plugin loader (Phase E1)
#
# Drop a file at ~/.config/keybind-ticker/plugins/<name>.py that defines:
#   META = {"preset": "ambient" | None, "refresh": <int seconds>}
#   def build_markup() -> tuple[str, list[str]]
# The stream auto-registers as "plugin:<name>". If the plugin import fails
# we log a warning and skip it — plugins must never be able to crash the
# ticker.
# ══════════════════════════════════════════════════

def _load_plugins():
    import importlib.util as _ilu
    plugin_dir = os.path.expanduser("~/.config/keybind-ticker/plugins")
    if not os.path.isdir(plugin_dir):
        return
    for entry in sorted(os.listdir(plugin_dir)):
        if not entry.endswith(".py") or entry.startswith("_"):
            continue
        path = os.path.join(plugin_dir, entry)
        name = f"plugin:{entry[:-3]}"
        try:
            spec = _ilu.spec_from_file_location(f"_kt_plugin_{entry[:-3]}", path)
            if spec is None or spec.loader is None:
                continue
            mod = _ilu.module_from_spec(spec)
            spec.loader.exec_module(mod)
            if not hasattr(mod, "build_markup"):
                continue
            STREAMS[name] = mod.build_markup
            STREAM_META[name] = {
                "preset": getattr(mod, "META", {}).get("preset"),
                "refresh": int(getattr(mod, "META", {}).get("refresh", 60)),
            }
            if name not in FALLBACK_ORDER:
                FALLBACK_ORDER.append(name)
        except Exception as _e:
            sys.stderr.write(f"plugin load failed: {entry}: {_e}\n")


_load_plugins()


# ══════════════════════════════════════════════════
# Per-stream click action dispatch (Phase A3)
#
# Each entry is a callable (win, segment_idx, segment_text) → bool. Return
# True to claim the click and skip the generic URL/clipboard fallback.
# Actions run on the main loop so they can freely call widget APIs.
# ══════════════════════════════════════════════════

def _spawn_detached(cmd):
    try:
        subprocess.Popen(cmd, stdout=subprocess.DEVNULL,
                         stderr=subprocess.DEVNULL, start_new_session=True)
        return True
    except Exception:
        return False


def _click_ci(win, idx, text):
    # CI segments look like "✓ <repo>" / "✗ <repo>" / "⏳ <repo>". If the
    # clicked token has a repo name, open its latest workflow runs in gh.
    for token in text.split():
        if "/" in token or token.isidentifier():
            repo = token.strip(":·")
            return _spawn_detached(["gh", "run", "list", "--repo",
                                    f"hairglasses-studio/{repo}", "--web"])
    return False


def _click_arch_news(win, idx, text):
    return _spawn_detached(["xdg-open", "https://archlinux.org/news/"])


def _click_calendar(win, idx, text):
    return _spawn_detached(["xdg-open", "https://calendar.google.com/"])


def _click_updates(win, idx, text):
    # Opening the pacman log is more useful than copying segment text.
    return _spawn_detached(["xdg-open", "/var/log/pacman.log"])


def _click_github(win, idx, text):
    # github stream already renders URL tokens — let the generic fallback
    # handle them. But if the segment is a bare notification title, open the
    # notifications dashboard.
    if text.startswith(("http://", "https://")):
        return False  # generic URL open handles it
    return _spawn_detached(["xdg-open", "https://github.com/notifications"])


def _click_hn(win, idx, text):
    # Filled in by Phase C1 (hn-top stream). Tries to extract the HN item id
    # from the segment prefix "#<id>"; falls back to HN front page.
    import re as _re
    m = _re.search(r"#(\d+)", text)
    if m:
        return _spawn_detached(
            ["xdg-open", f"https://news.ycombinator.com/item?id={m.group(1)}"]
        )
    return _spawn_detached(["xdg-open", "https://news.ycombinator.com/"])


STREAM_CLICK_ACTIONS = {
    "ci":        _click_ci,
    "arch-news": _click_arch_news,
    "calendar":  _click_calendar,
    "updates":   _click_updates,
    "github":    _click_github,
    "hn-top":    _click_hn,
}


def _playlist_path(name):
    return os.path.join(PLAYLIST_DIR, f"{name}.txt")


def list_playlists():
    """Enumerate available playlist names from PLAYLIST_DIR."""
    try:
        names = sorted(
            f[:-4] for f in os.listdir(PLAYLIST_DIR)
            if f.endswith(".txt") and not f.startswith(".")
        )
        return names or [DEFAULT_PLAYLIST]
    except OSError:
        return [DEFAULT_PLAYLIST]


def resolve_playlist_name():
    """Pick the active playlist name: CLI > state file > default."""
    if START_PLAYLIST:
        return START_PLAYLIST
    try:
        with open(os.path.join(STATE_DIR, "active-playlist")) as f:
            name = f.read().strip()
        if name:
            return name
    except OSError:
        pass
    return DEFAULT_PLAYLIST


def load_playlist(name=None):
    """Read playlist <name>.txt; fall back to FALLBACK_ORDER if missing/empty."""
    name = name or resolve_playlist_name()
    try:
        with open(_playlist_path(name)) as f:
            lines = [line.strip() for line in f]
        order = [l for l in lines
                 if l and not l.startswith("#") and l in STREAMS]
        if order:
            return order
    except OSError:
        pass
    return list(FALLBACK_ORDER)


# ══════════════════════════════════════════════════
# Visual Helpers
# ══════════════════════════════════════════════════

def make_gradient(x_start, total_width, phase):
    x0 = x_start - phase
    grad = _cairo.LinearGradient(x0, 0, x0 + total_width, 0)
    grad.set_extend(_cairo.Extend.REPEAT)
    n = len(GRADIENT_COLORS)
    for i, (r, g, b) in enumerate(GRADIENT_COLORS):
        grad.add_color_stop_rgb(i / (n - 1), r, g, b)
    return grad


# ── Water caustic ─────────────────────────────────
_WATER_SCALE = 8       # was 4 — larger stride = fewer numpy cells per frame
_WATER_MAX_ITER = 5    # was 6 — 1 fewer inner iter, effect visually identical
_WATER_TAU = 6.28318530718


def compute_water_caustic(width, height, time_s):
    # Ceil-divide so that after the final np.repeat upscale we have *at
    # least* `height` rows and `width` cols; `[:height, :width]` then trims
    # to the exact size. Floor-division undersized the output when BAR_H
    # wasn't a multiple of _WATER_SCALE (e.g. 39 / 4 → 9*4=36 < 39).
    sw = (width  + _WATER_SCALE - 1) // _WATER_SCALE
    sh = max((height + _WATER_SCALE - 1) // _WATER_SCALE, 2)
    t = time_s * 0.5 + 23.0
    ux = np.linspace(0, 1, sw, dtype=np.float32)
    uy = np.linspace(0, 1, sh, dtype=np.float32)
    u, v = np.meshgrid(ux, uy)
    px = np.mod(u * _WATER_TAU, _WATER_TAU) - 250.0
    py = np.mod(v * _WATER_TAU, _WATER_TAU) - 250.0
    ix, iy = px.copy(), py.copy()
    c = np.ones_like(px)
    inten = 0.005
    for n in range(_WATER_MAX_ITER):
        tn = t * (1.0 - 3.5 / (n + 1))
        ix, iy = (px + np.cos(tn - ix) + np.sin(tn + iy),
                  py + np.sin(tn - iy) + np.cos(tn + ix))
        denom = np.sqrt((px / (np.sin(ix + tn) / inten)) ** 2 +
                        (py / (np.cos(iy + tn) / inten)) ** 2)
        c += 1.0 / np.maximum(denom, 1e-6)
    c /= _WATER_MAX_ITER
    c = 1.17 - np.power(c, 1.4)
    b = np.clip(np.power(np.abs(c), 15.0) * 1.2, 0, 1)
    if _WATER_SCALE > 1:
        b = np.repeat(np.repeat(b, _WATER_SCALE, axis=0), _WATER_SCALE, axis=1)[:height, :width]
    return b


def water_caustic_surface(width, height, time_s):
    bright = compute_water_caustic(width, height, time_s)
    argb = np.zeros((height, width, 4), dtype=np.uint8)
    argb[:, :, 0] = (bright * 0.75 * 255).astype(np.uint8)  # B
    argb[:, :, 1] = (bright * 0.55 * 255).astype(np.uint8)  # G
    argb[:, :, 2] = (bright * 0.10 * 255).astype(np.uint8)  # R
    argb[:, :, 3] = (bright * 0.25 * 255).astype(np.uint8)  # A
    surf = _cairo.ImageSurface(_cairo.FORMAT_ARGB32, width, height)
    stride = surf.get_stride()
    # Vectorised row copy. Cairo pads rows to a 4-byte stride, which is
    # already satisfied at width*4 for ARGB32 so we can write a contiguous
    # block. Replaces a Python for-row loop + per-row .tobytes() that cost
    # ~10-15 ms per caustic frame at BAR_H=39.
    buf_view = np.frombuffer(surf.get_data(), dtype=np.uint8).reshape(height, stride)
    buf_view[:, :width * 4] = argb.reshape(height, width * 4)
    surf.mark_dirty()
    return surf


# ══════════════════════════════════════════════════
# Main Window
# ══════════════════════════════════════════════════

class TickerWindow(Gtk.ApplicationWindow):
    def __init__(self, preset_name="ambient", **kwargs):
        super().__init__(**kwargs)
        self.set_title("keybind-ticker")
        self._base_preset_name = preset_name
        self.preset = load_preset(preset_name)

        target_width = 1200  # windowed-mode fallback; layer-shell overrides below
        if LAYER_MODE:
            Gtk4LayerShell.init_for_window(self)
            Gtk4LayerShell.set_layer(self, Gtk4LayerShell.Layer.BOTTOM)
            for edge in (Gtk4LayerShell.Edge.BOTTOM, Gtk4LayerShell.Edge.LEFT, Gtk4LayerShell.Edge.RIGHT):
                Gtk4LayerShell.set_anchor(self, edge, True)
            Gtk4LayerShell.set_exclusive_zone(self, BAR_H)
            Gtk4LayerShell.set_namespace(self, "keybind-ticker")
            display = Gdk.Display.get_default()
            if display:
                for i in range(display.get_monitors().get_n_items()):
                    mon = display.get_monitors().get_item(i)
                    if mon and MONITOR_NAME in (mon.get_connector() or ""):
                        Gtk4LayerShell.set_monitor(self, mon)
                        geom = mon.get_geometry()
                        if geom:
                            target_width = max(geom.width, 1)
                        break
        self.set_default_size(target_width, BAR_H)

        self.da = Gtk.DrawingArea()
        self.da.set_content_height(BAR_H)
        self.da.set_vexpand(True)
        self.da.set_hexpand(True)
        self.da.set_draw_func(self._draw)
        self.set_child(self.da)

        # ── Interactive controls ──────────────────
        scroll_ctl = Gtk.EventControllerScroll.new(Gtk.EventControllerScrollFlags.VERTICAL)
        scroll_ctl.connect("scroll", self._on_scroll)
        self.da.add_controller(scroll_ctl)

        # Click: left button + middle button handled in same controller with button-check
        click_ctl = Gtk.GestureClick.new()
        click_ctl.set_button(0)  # all buttons
        click_ctl.connect("pressed", self._on_click)
        self.da.add_controller(click_ctl)

        motion_ctl = Gtk.EventControllerMotion.new()
        motion_ctl.connect("enter", self._on_enter)
        motion_ctl.connect("leave", self._on_leave)
        motion_ctl.connect("motion", self._on_motion)
        self.da.add_controller(motion_ctl)

        self.da.set_has_tooltip(True)
        self.da.connect("query-tooltip", self._on_tooltip)

        # ── State ─────────────────────────────────
        self.segments = []
        self.segment_bounds = []  # [(x_start, x_end, text), ...] in layout coords
        self.offset = 0.0
        self.gradient_phase = 0.0
        self.time_s = 0.0
        self.last_us = None
        self.layout = None
        self.half_w = 1.0
        # A8 glyph-mask caches. Rendered once per layout rebuild; reused
        # every frame via cr.mask_surface with a fresh gradient source so
        # the colour cycle animates without re-rasterising glyphs.
        self._glyph_mask_fill = None
        self._glyph_mask_outline = None

        self.paused = self._load_pause_state()
        self.pinned_stream = self._load_pin_state()
        self.urgent_mode = False
        self.urgent_timer_handle = None

        self.hover_x = 0
        self.is_hovering = False
        self.base_speed_cache = self.preset.speed

        self.reveal_chars = 99999  # no typewriter by default; set smaller to animate

        self.glitch_remaining = 0
        self.glitch_strips = []
        self.water_surf = None
        self.water_frame = 0
        self.glow_cache = None
        self.glow_frame = 0

        # Adaptive quality state. Tier 0 = base preset as configured; each
        # higher tier disables progressively-more-expensive effects when
        # the EMA of `_draw` frame time approaches the 30 Hz budget (33 ms).
        # Asymmetric EMA (fast up α=0.1, slow down α=0.05) reacts quickly
        # to overload and restores slowly so quiet periods don't flip the
        # tier back instantly on re-entry into a heavy stream.
        self._tier = 0
        self._ema_frame_ms = 0.0
        self._tier_stable_frames = 0

        self.stream_start_s = 0.0

        # Phosphor trail ring buffer (for Phase 6)
        self.phosphor_ring = deque(maxlen=4)
        self.phosphor_frame_ctr = 0

        # ── Content stream state ──────────────────
        self.playlist_name = resolve_playlist_name()
        self.stream_order = load_playlist(self.playlist_name)
        self.shuffle = self._load_shuffle_state()
        if self.shuffle:
            random.shuffle(self.stream_order)
        self.stream_idx = self._restore_stream_idx()

        # Background thread state for slow streams
        self._bg_lock = threading.Lock()
        self._bg_result = {}  # stream_name -> (markup, segments)
        self._bg_inflight = set()

        # Per-stream error flag → drives short-backoff scheduling.
        self._stream_errored = {}
        # Per-stream health counters (Phase A4). Populated on each absorb;
        # exported to /tmp/ticker-health.json every 30s for `hg ticker health`.
        self._stream_health = {}

        self._rebuild_stream()
        self._schedule_next_advance()
        self.da.add_tick_callback(self._tick)

        # Force the frame clock to tick continuously. Without this, Hyprland
        # stops delivering frame callbacks when the layer-shell surface is
        # "idle" (no interaction), which freezes every time-driven animation
        # (water caustic, glow breathing, phosphor trail, glitch, gradient)
        # until the user hovers or clicks. begin_updating keeps the clock live.
        self.da.connect("realize", self._on_da_realize)
        if self.da.get_realized():
            self._on_da_realize(self.da)

        # Priority-interrupt listener (watches notification history file)
        self._last_notif_mtime = 0.0
        GLib.timeout_add_seconds(3, self._check_priority_interrupt)

        # Phase A4: export per-stream health snapshot every 30s.
        GLib.timeout_add_seconds(30, self._write_health_snapshot)

    # ── Persistence helpers ─────────────────────

    def reload_from_state(self):
        """Re-read state files (active-playlist, pinned-stream, paused) and
        apply the resulting configuration without restarting. Intended to be
        called via GLib.idle_add from a SIGUSR1 handler so external tools
        (hg ticker playlist / pin / pause) can switch the ticker without
        flickering. Runs on the main loop so it's safe to call widget APIs."""
        # Playlist
        try:
            with open(os.path.join(STATE_DIR, "active-playlist")) as f:
                desired_playlist = f.read().strip()
        except OSError:
            desired_playlist = self.playlist_name
        if desired_playlist and desired_playlist != self.playlist_name:
            new_order = load_playlist(desired_playlist)
            if new_order:
                self.playlist_name = desired_playlist
                self.stream_order = new_order
                if self.shuffle:
                    random.shuffle(self.stream_order)
                self.stream_idx = 0
                self.pinned_stream = None
                self._save_pin_state()
                self._rebuild_stream()
                self._schedule_next_advance()
        # Shuffle toggle — reshuffle the current order if the flag flipped
        # to on, or reload the pristine playlist order if it flipped off.
        desired_shuffle = self._load_shuffle_state()
        if desired_shuffle != self.shuffle:
            self.shuffle = desired_shuffle
            current_name = self.stream_order[self.stream_idx] if self.stream_order else None
            fresh = load_playlist(self.playlist_name)
            if self.shuffle and fresh:
                random.shuffle(fresh)
            self.stream_order = fresh
            # Keep playing the currently-displayed stream if possible.
            if current_name in self.stream_order:
                self.stream_idx = self.stream_order.index(current_name)
            else:
                self.stream_idx = 0
                self._rebuild_stream()
                self._schedule_next_advance()
        # Pinned stream
        desired_pin = self._load_pin_state()
        if desired_pin != self.pinned_stream:
            self.pinned_stream = desired_pin
            if desired_pin and desired_pin in self.stream_order:
                self.stream_idx = self.stream_order.index(desired_pin)
                self._rebuild_stream()
                self._schedule_next_advance()
        # Paused
        desired_paused = self._load_pause_state()
        if desired_paused != self.paused:
            self.paused = desired_paused
        return False  # one-shot idle_add

    def _load_pause_state(self):
        return os.path.exists(os.path.join(STATE_DIR, "paused"))

    def _load_shuffle_state(self):
        return os.path.exists(os.path.join(STATE_DIR, "shuffle"))

    def _save_pause_state(self):
        os.makedirs(STATE_DIR, exist_ok=True)
        path = os.path.join(STATE_DIR, "paused")
        if self.paused:
            open(path, "w").close()
        else:
            try:
                os.remove(path)
            except FileNotFoundError:
                pass

    def _load_pin_state(self):
        try:
            with open(os.path.join(STATE_DIR, "pinned-stream")) as f:
                name = f.read().strip()
                if name in STREAMS:
                    return name
        except (OSError, ValueError):
            pass
        return None

    def _save_pin_state(self):
        os.makedirs(STATE_DIR, exist_ok=True)
        path = os.path.join(STATE_DIR, "pinned-stream")
        if self.pinned_stream:
            # Atomic write via temp file + os.replace so readers
            # (hg ticker status, _load_pin_state on SIGUSR1) never see
            # a half-written name during concurrent updates.
            tmp = path + ".tmp"
            with open(tmp, "w") as f:
                f.write(self.pinned_stream)
            os.replace(tmp, path)
        else:
            try:
                os.remove(path)
            except FileNotFoundError:
                pass

    def _restore_stream_idx(self):
        try:
            with open(os.path.join(STATE_DIR, "current-stream")) as f:
                name = f.read().strip()
            if name in self.stream_order:
                return self.stream_order.index(name)
        except (OSError, ValueError):
            pass
        return 0

    # ── Stream scheduling ─────────────────────

    def _current_interval(self, stream_name):
        if self._stream_errored.get(stream_name):
            return ERROR_BACKOFF_S
        meta = STREAM_META.get(stream_name, {})
        raw = meta.get("dwell", meta.get("refresh", DEFAULT_REFRESH_S))
        return max(MIN_DWELL_S, min(raw, MAX_DWELL_S))

    def _absorb_segments(self, stream_name, segments):
        """Strip control sentinels and update error/urgent state."""
        now = int(_time.time())
        health = self._stream_health.setdefault(
            stream_name,
            {"last_ok": 0, "last_err": 0, "consecutive_err": 0, "total_err": 0, "total_ok": 0},
        )
        if segments == ["__EMPTY__"]:
            self._stream_errored[stream_name] = True
            health["last_err"] = now
            health["consecutive_err"] += 1
            health["total_err"] += 1
            return []
        if segments == ["__URGENT__"]:
            self._stream_errored[stream_name] = False
            health["last_ok"] = now
            health["consecutive_err"] = 0
            health["total_ok"] += 1
            self._trigger_urgent_mode()
            return []
        self._stream_errored[stream_name] = False
        health["last_ok"] = now
        health["consecutive_err"] = 0
        health["total_ok"] += 1
        return segments

    def _write_health_snapshot(self):
        """Export stream health counters + perf tier to tmpfs for
        `hg ticker health`. The ``perf`` block surfaces adaptive-quality
        state and live runtime flags that are otherwise invisible to
        external tools (bg thread backlog, urgent mode, dwell remaining).
        """
        import json
        path = "/tmp/ticker-health.json"
        tmp = path + ".tmp"

        # Compute dwell remaining for the currently-displayed stream.
        current = (self.stream_order[self.stream_idx]
                   if self.stream_order else None)
        dwell_remaining = None
        if current and not self.paused and not self.pinned_stream:
            try:
                interval = self._current_interval(current)
                elapsed = max(0.0, self.time_s - self.stream_start_s)
                dwell_remaining = max(0, int(interval - elapsed))
            except Exception:
                dwell_remaining = None

        urgent_until = None
        if self.urgent_mode and self.urgent_timer_handle is not None:
            # Best-effort — no exact remaining time without tracking
            # start; expose the boolean state and an ack-by marker.
            urgent_until = int(_time.time()) + 10

        try:
            with open(tmp, "w") as f:
                json.dump({
                    "pid": os.getpid(),
                    "ts": int(_time.time()),
                    "playlist": self.playlist_name,
                    "streams": self._stream_health,
                    "perf": {
                        "tier":             self._tier,
                        "ema_frame_ms":     round(self._ema_frame_ms, 2),
                        "bg_inflight":      len(self._bg_inflight),
                        "current_stream":   current,
                        "dwell_remaining":  dwell_remaining,
                        "urgent_mode":      self.urgent_mode,
                        "urgent_until":     urgent_until,
                        "paused":           self.paused,
                        "shuffle":          self.shuffle,
                        "pinned":           self.pinned_stream,
                    },
                }, f)
            os.replace(tmp, path)
        except OSError:
            pass
        return True  # keep the timer alive

    def _schedule_next_advance(self):
        current = self.stream_order[self.stream_idx]
        interval = self._current_interval(current)
        GLib.timeout_add_seconds(interval, self._maybe_advance_stream,
                                 current)

    def _maybe_advance_stream(self, from_stream):
        # Only advance if we're still on this stream
        if self.stream_order[self.stream_idx] != from_stream:
            return GLib.SOURCE_REMOVE
        if self.pinned_stream:
            # Just rebuild, don't advance
            self._rebuild_stream()
        else:
            next_idx = (self.stream_idx + 1) % len(self.stream_order)
            # End-of-cycle reshuffle: on wrap-around in shuffle mode,
            # re-randomise so successive cycles aren't in the same order.
            if next_idx == 0 and self.shuffle and len(self.stream_order) > 1:
                pinned_last = self.stream_order[self.stream_idx]
                random.shuffle(self.stream_order)
                # Guarantee the new first stream isn't a repeat of the one
                # we just finished — swap index 0 with index 1 on collision.
                if self.stream_order[0] == pinned_last:
                    self.stream_order[0], self.stream_order[1] = (
                        self.stream_order[1], self.stream_order[0])
            self.stream_idx = next_idx
            self._rebuild_stream()
        self._schedule_next_advance()
        return GLib.SOURCE_REMOVE

    def _rebuild_stream(self, force_name=None):
        stream_name = force_name or (self.pinned_stream
                                     if self.pinned_stream
                                     else self.stream_order[self.stream_idx])
        # Persist current stream atomically; the _write_health_snapshot
        # path and external tools (hg ticker status) read this on demand
        # and must never see a half-written name.
        os.makedirs(STATE_DIR, exist_ok=True)
        try:
            path = os.path.join(STATE_DIR, "current-stream")
            tmp = path + ".tmp"
            with open(tmp, "w") as f:
                f.write(stream_name)
            os.replace(tmp, path)
        except OSError:
            pass
        # Apply per-stream preset override
        meta = STREAM_META.get(stream_name, {})
        target_preset = meta.get("preset") or self._base_preset_name
        self._apply_preset(target_preset)
        # Stream start timestamp for progress bar
        self.stream_start_s = self.time_s
        # Typewriter reveal animation: start from 0
        self.reveal_chars = 0

        if stream_name in SLOW_STREAMS:
            # Check if we have a cached result from last time
            with self._bg_lock:
                cached = self._bg_result.get(stream_name)
            if cached is not None:
                markup, segments = cached
            else:
                # Show placeholder while fetching
                markup, segments = _empty(
                    f" {stream_name.upper()}", "#66708f", "loading...")
            self.ticker_markup = markup
            self.segments = self._absorb_segments(stream_name, segments)
            self._dispatch_bg_fetch(stream_name)
        else:
            builder = STREAMS.get(stream_name)
            if builder is None:
                markup, segments = _empty(
                    stream_name.upper(), "#66708f", "unknown stream")
            else:
                try:
                    markup, segments = builder()
                except Exception:
                    markup, segments = _empty(
                        stream_name.upper(), "#ff5c8a", "builder error")
            self.ticker_markup = markup
            self.segments = self._absorb_segments(stream_name, segments)

        self.layout = None
        self.segment_bounds = []
        self.da.queue_draw()
        self._emit_app_signal("StreamChanged",
                              (GLib.Variant("s", stream_name or ""),))

    def _dispatch_bg_fetch(self, stream_name):
        with self._bg_lock:
            if stream_name in self._bg_inflight:
                return
            self._bg_inflight.add(stream_name)

        def worker():
            builder = STREAMS.get(stream_name)
            if builder is None:
                result = _empty(stream_name.upper(), "#66708f", "unknown stream")
            else:
                try:
                    result = builder()
                except Exception:
                    result = _empty(stream_name.upper(), "#ff5c8a", "builder error")
            with self._bg_lock:
                self._bg_result[stream_name] = result
                self._bg_inflight.discard(stream_name)
            GLib.idle_add(self._bg_apply, stream_name, result)

        t = threading.Thread(target=worker, daemon=True)
        t.start()

    def _bg_apply(self, stream_name, result):
        # Apply result only if we're still showing this stream
        current = (self.pinned_stream if self.pinned_stream
                   else self.stream_order[self.stream_idx])
        if current == stream_name:
            markup, segments = result
            self.ticker_markup = markup
            self.segments = self._absorb_segments(stream_name, segments)
            self.layout = None
            self.segment_bounds = []
            self.da.queue_draw()
        return False  # idle_add -> run once

    # ── Urgency surge ─────────────────────────

    def _trigger_urgent_mode(self):
        self.urgent_mode = True
        # Cancel existing timer if any
        if self.urgent_timer_handle is not None:
            GLib.source_remove(self.urgent_timer_handle)
        self.urgent_timer_handle = GLib.timeout_add_seconds(
            10, self._clear_urgent_mode)
        self._emit_app_signal("UrgentMode", (GLib.Variant("b", True),))

    def _clear_urgent_mode(self):
        was = self.urgent_mode
        self.urgent_mode = False
        self.urgent_timer_handle = None
        if was:
            self._emit_app_signal("UrgentMode", (GLib.Variant("b", False),))
        return GLib.SOURCE_REMOVE

    # ── DBus handlers ─────────────────────────
    # All methods below run on the GTK main thread via GLib.idle_add
    # from TickerApp._on_method_call. Each one should return False so
    # idle_add doesn't re-queue it.

    def _emit_app_signal(self, name, variant_tuple):
        """Thin wrapper: ask the Gtk.Application to emit a DBus signal."""
        app = self.get_application()
        if app is not None and hasattr(app, "_emit_signal"):
            app._emit_signal(name, variant_tuple)

    def _dbus_apply_pin(self, stream_or_none):
        """Pin (or unpin if None). Validates against current stream_order."""
        if stream_or_none:
            if stream_or_none not in STREAMS:
                sys.stderr.write(f"dbus Pin: unknown stream {stream_or_none}\n")
                return False
            if stream_or_none not in self.stream_order:
                # Mirror ticker-shot's auto-switch: if the requested pin
                # isn't in the active playlist, the external tool is
                # expected to have already switched; we don't reshuffle here.
                sys.stderr.write(f"dbus Pin: {stream_or_none} not in "
                                 f"active playlist {self.playlist_name}\n")
                return False
            self.pinned_stream = stream_or_none
            self.stream_idx = self.stream_order.index(stream_or_none)
        else:
            self.pinned_stream = None
        self._save_pin_state()
        self._rebuild_stream()
        self._schedule_next_advance()
        return False

    def _dbus_toggle_pin(self):
        if self.pinned_stream:
            self._dbus_apply_pin(None)
        elif self.stream_order:
            self._dbus_apply_pin(self.stream_order[self.stream_idx])
        return False

    def _dbus_advance(self, direction: int):
        """Advance or rewind by `direction` steps without pinning."""
        if not self.stream_order:
            return False
        step = 1 if direction >= 0 else -1
        new_idx = (self.stream_idx + step) % len(self.stream_order)
        # End-of-cycle reshuffle parity with _maybe_advance_stream
        if step > 0 and new_idx == 0 and self.shuffle and len(self.stream_order) > 1:
            last = self.stream_order[self.stream_idx]
            random.shuffle(self.stream_order)
            if self.stream_order[0] == last:
                self.stream_order[0], self.stream_order[1] = (
                    self.stream_order[1], self.stream_order[0])
        self.stream_idx = new_idx
        self._rebuild_stream()
        self._schedule_next_advance()
        return False

    def _dbus_toggle_pause(self):
        self.paused = not self.paused
        self._save_pause_state()
        return False

    def _dbus_set_shuffle(self, mode: str):
        mode = (mode or "").strip().lower()
        path = os.path.join(STATE_DIR, "shuffle")
        os.makedirs(STATE_DIR, exist_ok=True)
        if mode == "on":
            new_state = True
        elif mode == "off":
            new_state = False
        else:  # toggle / empty / anything else
            new_state = not self.shuffle
        if new_state:
            open(path, "w").close()
        else:
            try:
                os.remove(path)
            except FileNotFoundError:
                pass
        # Trigger the reload path so stream_order is reshuffled in place.
        self.reload_from_state()
        return False

    def _dbus_set_playlist(self, name: str):
        name = (name or "").strip()
        if not name:
            return False
        path = os.path.join(STATE_DIR, "active-playlist")
        os.makedirs(STATE_DIR, exist_ok=True)
        tmp = path + ".tmp"
        with open(tmp, "w") as f:
            f.write(name)
        os.replace(tmp, path)
        self.reload_from_state()
        return False

    def _dbus_set_preset(self, name: str):
        name = (name or "").strip()
        if name in PRESETS:
            self._base_preset_name = name
            self._apply_preset(name)
            self.da.queue_draw()
        return False

    def _dbus_show_banner(self, text: str, color: str):
        """Delegate to toast-ticker via its DBus surface.

        Keeps the banner overlay separate from the scrolling text so an
        external caller gets the same visual as `ShowToast` already
        provides for swaync → urgent notifications.
        """
        try:
            conn = Gio.bus_get_sync(Gio.BusType.SESSION, None)
            conn.call_sync(
                "io.hairglasses.toast",
                "/io/hairglasses/Toast",
                "io.hairglasses.Toast",
                "ShowToast",
                GLib.Variant("(ss)", (str(text), str(color or "#29f0ff"))),
                None, Gio.DBusCallFlags.NONE, 1000, None,
            )
        except Exception as e:
            sys.stderr.write(f"dbus ShowBanner → toast relay failed: {e}\n")
        return False

    def _dbus_set_urgent(self, active: bool):
        if active:
            self._trigger_urgent_mode()
        else:
            self._clear_urgent_mode()
        return False

    def _dbus_reload_plugins(self):
        """Re-import bundled plugin modules + reload TOML catalogue.

        Clears STREAMS/STREAM_META/SLOW_STREAMS entries populated by the
        plugin and TOML loaders, then re-runs both. Preserves the
        user-drop-in plugins path handled elsewhere.
        """
        import importlib
        try:
            # Drop anything the plugin loader / TOML loader populated;
            # user drop-ins are registered under `plugin:` prefix so we
            # leave those alone.
            stale = [k for k in STREAMS.keys() if not k.startswith("plugin:")]
            for k in stale:
                STREAMS.pop(k, None)
                STREAM_META.pop(k, None)
                SLOW_STREAMS.discard(k)
            # Reload each ticker_streams submodule so edits take effect.
            import ticker_streams as _ts
            pkg_dir = os.path.join(os.path.dirname(os.path.realpath(__file__)),
                                   "lib", "ticker_streams")
            for fname in sorted(os.listdir(pkg_dir)):
                if not fname.endswith(".py") or fname == "__init__.py":
                    continue
                mod_name = f"ticker_streams.{fname[:-3]}"
                if mod_name in sys.modules:
                    importlib.reload(sys.modules[mod_name])
            # Reload the package __init__ so FONTS / factory changes apply
            importlib.reload(_ts)
            # Re-register
            _load_toml_catalogue()
            _load_bundled_plugins()
            self._rebuild_stream()
            sys.stderr.write(
                f"dbus ReloadPlugins: {len(STREAMS)} streams registered\n"
            )
        except Exception as e:
            sys.stderr.write(f"dbus ReloadPlugins failed: {e}\n")
        return False

    def _check_priority_interrupt(self):
        history = os.path.expanduser(
            "~/.local/state/dotfiles/desktop-control/notifications/history.jsonl")
        try:
            mtime = os.stat(history).st_mtime
        except OSError:
            return GLib.SOURCE_CONTINUE
        if mtime <= self._last_notif_mtime:
            return GLib.SOURCE_CONTINUE
        self._last_notif_mtime = mtime
        # Peek at last entry
        try:
            with open(history) as f:
                lines = f.readlines()
            if not lines:
                return GLib.SOURCE_CONTINUE
            n = json.loads(lines[-1])
            if n.get("urgency") == "critical":
                # Jump to notifications stream
                if "notifications" in self.stream_order:
                    self.stream_idx = self.stream_order.index("notifications")
                    self._rebuild_stream()
        except Exception:
            pass
        return GLib.SOURCE_CONTINUE

    # ── Preset application ────────────────────

    def _apply_preset(self, name):
        # Preserve runtime-adjusted speed
        new_preset = load_preset(name)
        self.preset = new_preset
        self.base_speed_cache = new_preset.speed
        self._base_preset_name = name
        # Re-apply any active degradation — tier state persists across
        # stream changes so a loaded machine doesn't bounce back to full
        # effects on every rotation.
        self._apply_tier(self._tier)

    def _apply_tier(self, tier):
        """Mutate `self.preset` to disable effects based on degradation tier.

        Only reduces effect intensity; never promotes beyond the preset's
        own configured value. `_apply_preset` calls this after reloading
        the base preset so the cumulative effect is correct."""
        if tier >= 1:
            self.preset.phosphor_trail = 0.0
            self.preset.ghost_echo = 0.0
        if tier >= 2:
            self.preset.glow_kernel = min(self.preset.glow_kernel, 11)
        if tier >= 3:
            self.preset.water_skip = 0
        if tier >= 4:
            self.preset.glitch_prob = 0.0
            self.preset.synthwave_border = False
            self.preset.holo_shimmer = 0.0
            self.preset.wave_amp = 0

    def _update_adaptive_tier(self, frame_ms):
        """Fold this frame's draw time into the EMA and promote/demote the
        degradation tier if the smoothed frame time crosses a threshold.

        Thresholds are relative to the 33 ms budget at 30 Hz:
          tier 1 ≥ 20 ms (60 %)  — kill phosphor_trail + ghost_echo
          tier 2 ≥ 26 ms (80 %)  — clamp glow_kernel ≤ 11
          tier 3 ≥ 30 ms (90 %)  — disable water_caustic
          tier 4 ≥ 33 ms (100 %) — strip to minimal-like preset
        Upgrade fires on the next frame that trips the threshold; downgrade
        requires 120 consecutive frames (4 s at 30 Hz) under budget before
        restoring one tier — prevents oscillation on periodic heavy frames.
        """
        alpha = 0.1 if frame_ms > self._ema_frame_ms else 0.05
        self._ema_frame_ms += (frame_ms - self._ema_frame_ms) * alpha

        target = 0
        ema = self._ema_frame_ms
        if ema > 20: target = 1
        if ema > 26: target = 2
        if ema > 30: target = 3
        if ema > 33: target = 4

        if target > self._tier:
            self._tier = target
            self._apply_tier(self._tier)
            self._tier_stable_frames = 0
            self._emit_app_signal("TierChanged",
                                  (GLib.Variant("i", self._tier),))
        elif target < self._tier:
            self._tier_stable_frames += 1
            if self._tier_stable_frames >= 120:  # 4 s at 30 Hz
                self._tier = max(self._tier - 1, target)
                self._tier_stable_frames = 0
                # Reload base + reapply tier so effects we cleared at the
                # higher tier come back at the lower one.
                self._apply_preset(self._base_preset_name)
                self._emit_app_signal("TierChanged",
                                      (GLib.Variant("i", self._tier),))
        else:
            self._tier_stable_frames = 0

    # ── Input handlers ────────────────────────

    def _on_scroll(self, controller, dx, dy):
        # Shift+scroll: cycle streams instead of speed
        event = controller.get_current_event()
        shift = False
        if event is not None:
            state = event.get_modifier_state()
            shift = bool(state & Gdk.ModifierType.SHIFT_MASK)
        if shift:
            if dy > 0:
                self.stream_idx = (self.stream_idx + 1) % len(self.stream_order)
            elif dy < 0:
                self.stream_idx = (self.stream_idx - 1) % len(self.stream_order)
            self._rebuild_stream()
        else:
            p = self.preset
            p.speed = max(10.0, min(200.0, p.speed - dy * 5.0))
            self.base_speed_cache = p.speed
        return True

    def _on_enter(self, controller, x, y):
        self.is_hovering = True
        # Adaptive speed: slow to 20% on hover
        self.preset.speed = self.base_speed_cache * 0.2

    def _on_leave(self, controller):
        self.is_hovering = False
        self.preset.speed = self.base_speed_cache

    def _on_motion(self, controller, x, y):
        self.hover_x = x

    def _on_click(self, gesture, n_press, x, y):
        button = gesture.get_current_button()
        # Middle-click: toggle pause
        if button == 2:
            self.paused = not self.paused
            self._save_pause_state()
            gesture.set_state(Gtk.EventSequenceState.CLAIMED)
            return
        # Right-click: open context menu
        if button == 3:
            self._show_context_menu(x, y)
            gesture.set_state(Gtk.EventSequenceState.CLAIMED)
            return
        # Left-click: first try a per-stream custom action, then fall back
        # to URL-open / clipboard-copy.
        if button == 1:
            idx = self._segment_at_x(x)
            if idx is None or idx >= len(self.segments):
                return
            text = self.segments[idx]
            if not text:
                return
            stream = self.stream_order[self.stream_idx] if self.stream_order else None
            action = STREAM_CLICK_ACTIONS.get(stream) if stream else None
            if action is not None:
                try:
                    if action(self, idx, text):
                        return
                except Exception:
                    pass
            if URL_RE.match(text) or text.startswith(("http://", "https://")):
                try:
                    subprocess.Popen(["xdg-open", text],
                                     stdout=subprocess.DEVNULL,
                                     stderr=subprocess.DEVNULL)
                except Exception:
                    pass
            else:
                try:
                    subprocess.Popen(["wl-copy", text],
                                     stdout=subprocess.DEVNULL,
                                     stderr=subprocess.DEVNULL)
                except Exception:
                    pass

    def _segment_at_x(self, x):
        """Accurate segment lookup using Pango byte-index mapping."""
        if not self.segment_bounds or self.half_w <= 0:
            # Fall back to uniform estimate if bounds not computed
            if not self.segments:
                return None
            seg_width = self.half_w / max(len(self.segments), 1)
            abs_x = (x + self.offset) % self.half_w
            return min(int(abs_x / seg_width), len(self.segments) - 1)
        abs_x = (x + self.offset) % self.half_w
        for i, (xs, xe, _text) in enumerate(self.segment_bounds):
            if xs <= abs_x < xe:
                return i
        return None

    def _show_context_menu(self, x, y):
        # Remember which segment the right-click landed on so the per-segment
        # menu items (Copy, Open URL, Dismiss) can target it.
        self._menu_segment_idx = self._segment_at_x(x)
        menu = Gio.Menu()

        # Per-segment actions at the top of the menu.
        seg_section = Gio.Menu()
        cur_stream = self.stream_order[self.stream_idx] if self.stream_order else ""
        seg_text = ""
        if (self._menu_segment_idx is not None
                and self._menu_segment_idx < len(self.segments)):
            seg_text = self.segments[self._menu_segment_idx] or ""
        if seg_text:
            preview = seg_text if len(seg_text) <= 32 else seg_text[:30] + "…"
            seg_section.append(f"Copy: {preview}", "ticker.copy_segment")
            if URL_RE.search(seg_text) or seg_text.startswith(("http://", "https://")):
                seg_section.append("Open URL", "ticker.open_segment")
            menu.append_section("Segment", seg_section)

        stream_section = Gio.Menu()
        for name in self.stream_order:
            stream_section.append(name, f"ticker.stream::{name}")
        menu.append_section("Streams", stream_section)

        playlist_section = Gio.Menu()
        for name in list_playlists():
            label = f"{name} \u2713" if name == self.playlist_name else name
            playlist_section.append(label, f"ticker.playlist::{name}")
        menu.append_section("Playlists", playlist_section)

        preset_section = Gio.Menu()
        for name in PRESETS.keys():
            preset_section.append(name, f"ticker.preset::{name}")
        menu.append_section("Presets", preset_section)

        action_section = Gio.Menu()
        action_section.append("Pause" if not self.paused else "Resume",
                              "ticker.pause")
        if self.pinned_stream:
            action_section.append(f"Unpin ({self.pinned_stream})",
                                  "ticker.unpin")
        else:
            action_section.append("Pin current stream", "ticker.pin")
        action_section.append(
            "Shuffle \u2713" if self.shuffle else "Shuffle",
            "ticker.shuffle_toggle",
        )
        # Dismiss is an Actions-level item (not Segment-only) so right-
        # clicking an empty region still surfaces it.
        action_section.append(f"Dismiss {cur_stream}",
                              "ticker.dismiss_stream")
        action_section.append("Next stream",  "ticker.next_stream")
        action_section.append("Prev stream",  "ticker.prev_stream")
        if self.urgent_mode:
            action_section.append("Snooze urgent mode",
                                  "ticker.snooze_urgent")
        menu.append_section("Actions", action_section)

        self._ensure_actions()

        popover = Gtk.PopoverMenu.new_from_model(menu)
        popover.set_parent(self.da)
        popover.set_pointing_to(Gdk.Rectangle(int(x), int(y), 1, 1))
        popover.popup()

    def _ensure_actions(self):
        if getattr(self, "_actions_registered", False):
            return
        group = Gio.SimpleActionGroup()

        def on_stream(action, param):
            name = param.get_string()
            if name in STREAMS and name in self.stream_order:
                self.stream_idx = self.stream_order.index(name)
                self._rebuild_stream()

        def on_preset(action, param):
            name = param.get_string()
            if name in PRESETS:
                self._base_preset_name = name
                self._apply_preset(name)
                self.layout = None
                self.da.queue_draw()

        def on_pause(action, param):
            self.paused = not self.paused
            self._save_pause_state()

        def on_pin(action, param):
            self.pinned_stream = self.stream_order[self.stream_idx]
            self._save_pin_state()

        def on_unpin(action, param):
            self.pinned_stream = None
            self._save_pin_state()

        def on_copy_segment(action, param):
            idx = getattr(self, "_menu_segment_idx", None)
            if idx is None or idx >= len(self.segments):
                return
            text = self.segments[idx] or ""
            if text:
                try:
                    subprocess.Popen(["wl-copy", text],
                                     stdout=subprocess.DEVNULL,
                                     stderr=subprocess.DEVNULL)
                except Exception:
                    pass

        def on_open_segment(action, param):
            idx = getattr(self, "_menu_segment_idx", None)
            if idx is None or idx >= len(self.segments):
                return
            text = self.segments[idx] or ""
            m = URL_RE.search(text)
            url = m.group(0) if m else (text if text.startswith(("http://", "https://")) else None)
            if url:
                try:
                    subprocess.Popen(["xdg-open", url],
                                     stdout=subprocess.DEVNULL,
                                     stderr=subprocess.DEVNULL)
                except Exception:
                    pass

        def on_dismiss_stream(action, param):
            # Advance past the current stream once; the playlist rotation
            # will bring it back next cycle. Useful for noisy streams you
            # want to skip for now.
            if self.stream_order:
                self.stream_idx = (self.stream_idx + 1) % len(self.stream_order)
                self._rebuild_stream()
                self._schedule_next_advance()

        def on_playlist(action, param):
            new_name = param.get_string()
            if new_name == self.playlist_name:
                return
            new_order = load_playlist(new_name)
            if not new_order:
                return
            self.playlist_name = new_name
            self.stream_order = new_order
            self.stream_idx = 0
            self.pinned_stream = None
            self._save_pin_state()
            os.makedirs(STATE_DIR, exist_ok=True)
            try:
                with open(os.path.join(STATE_DIR, "active-playlist"), "w") as f:
                    f.write(new_name)
            except OSError:
                pass
            self._rebuild_stream()
            self._schedule_next_advance()

        def on_shuffle_toggle(action, param):
            # Toggle in-process then reuse the DBus state-file + reload
            # path so CLI / DBus / menu all stay in sync.
            mode = "off" if self.shuffle else "on"
            self._dbus_set_shuffle(mode)

        def on_snooze_urgent(action, param):
            self._clear_urgent_mode()

        def on_next_stream(action, param):
            self._dbus_advance(1)

        def on_prev_stream(action, param):
            self._dbus_advance(-1)

        for act_name, cb, ptype in (
            ("stream",          on_stream,         GLib.VariantType.new("s")),
            ("preset",          on_preset,         GLib.VariantType.new("s")),
            ("playlist",        on_playlist,       GLib.VariantType.new("s")),
            ("copy_segment",    on_copy_segment,   None),
            ("open_segment",    on_open_segment,   None),
            ("dismiss_stream",  on_dismiss_stream, None),
            ("pause",           on_pause,          None),
            ("pin",             on_pin,            None),
            ("unpin",           on_unpin,          None),
            ("shuffle_toggle",  on_shuffle_toggle, None),
            ("snooze_urgent",   on_snooze_urgent,  None),
            ("next_stream",     on_next_stream,    None),
            ("prev_stream",     on_prev_stream,    None),
        ):
            act = Gio.SimpleAction.new(act_name, ptype)
            act.connect("activate", cb)
            group.add_action(act)
        self.insert_action_group("ticker", group)
        self._actions_registered = True

    def _on_tooltip(self, widget, x, y, keyboard_mode, tooltip):
        stream = (self.pinned_stream if self.pinned_stream
                  else self.stream_order[self.stream_idx])
        seg_info = ""
        if self.segments:
            idx = self._segment_at_x(x)
            if idx is not None and idx < len(self.segments):
                seg_text = self.segments[idx]
                label = "URL" if URL_RE.match(seg_text) else "Segment"
                action = "click to open" if URL_RE.match(seg_text) else "click to copy"
                seg_info = f'\n<b>{label}:</b> {escape(seg_text)}\n<i>{action}</i>'
        status = []
        if self.paused:
            status.append("<b>PAUSED</b>")
        if self.pinned_stream:
            status.append(f"<b>PINNED:</b> {escape(self.pinned_stream)}")
        if self.urgent_mode:
            status.append("<b>URGENT</b>")
        status_line = " \u00b7 ".join(status) + "\n" if status else ""
        tooltip.set_markup(
            f'{status_line}'
            f'<b>Playlist:</b> {escape(self.playlist_name)}\n'
            f'<b>Stream:</b> {escape(stream)}\n'
            f'<b>Preset:</b> {escape(self._base_preset_name)}\n'
            f'<b>Speed:</b> {self.preset.speed:.0f} px/s\n'
            f'<b>Controls:</b> scroll=speed, shift+scroll=switch stream,\n'
            f'mid-click=pause, right-click=menu'
            f'{seg_info}'
        )
        return True

    # ── Tick / scroll ─────────────────────────

    def _on_da_realize(self, widget):
        clock = widget.get_frame_clock()
        if clock is not None:
            clock.begin_updating()

    def _tick(self, widget, frame_clock):
        now = frame_clock.get_frame_time()
        # DPMS-stall guard: when the monitor enters DPMS-off the compositor
        # stops firing frame callbacks, but the GLib tick callback keeps
        # running. `get_frame_time()` returns the same timestamp until the
        # clock resumes, so an early-exit here avoids all per-tick math
        # (glitch RNG, dt accumulation, gradient phase) while the display
        # is asleep. Costs one monotonic comparison on the hot path.
        if self.last_us is not None and now <= self.last_us:
            return GLib.SOURCE_CONTINUE
        p = self.preset
        if self.last_us is not None:
            dt = min((now - self.last_us) / 1_000_000.0, 0.05)
            if not self.paused:
                self.offset += p.speed * dt
                self.gradient_phase += p.gradient_speed * dt
            self.time_s += dt
            if self.half_w > 0 and self.offset >= self.half_w:
                self.offset -= self.half_w
            if self.gradient_phase >= p.gradient_span:
                self.gradient_phase -= p.gradient_span
            # Typewriter reveal: advance
            if self.reveal_chars < 99999:
                self.reveal_chars += 3
                if self.reveal_chars >= 600:
                    self.reveal_chars = 99999  # disable until next rebuild

            # Glitch
            glitch_prob = p.glitch_prob
            glitch_frames = p.glitch_frames
            ca_offset = p.ca_offset
            if self.urgent_mode:
                glitch_prob = max(glitch_prob, 0.05)
                glitch_frames = max(glitch_frames, 6)
                ca_offset = max(ca_offset, 6)
            if self.glitch_remaining > 0:
                self.glitch_remaining -= 1
            elif glitch_prob > 0 and random.random() < glitch_prob:
                self.glitch_remaining = glitch_frames
                self.glitch_strips = [
                    (random.randint(0, BAR_H - 4), random.randint(3, 8),
                     random.randint(-10, 10))
                    for _ in range(3)
                ]
            # Stash for _draw
            self._runtime_ca_offset = ca_offset
        self.last_us = now
        self.water_frame += 1
        # Throttle redraws to ~30 Hz. Offset / time / glitch state still
        # advances via dt every tick (animation stays smooth and high-rate);
        # only pixel rasterization is capped. 30 Hz is still perceptually
        # smooth for horizontally scrolling text (matches 24 fps film, with
        # headroom) and lets the 40 % enlarged effects fit in a comfortable
        # frame budget without the compositor dropping frames.
        if now - getattr(self, "_last_draw_us", 0) >= 33_333:
            widget.queue_draw()
            self._last_draw_us = now
        return GLib.SOURCE_CONTINUE

    # ── Rendering helpers ─────────────────────

    def _compute_segment_bounds(self):
        """Populate self.segment_bounds by finding each U+00B7 separator.

        Pango uses UTF-8 byte offsets; Python str uses code points. We convert
        via len(prefix.encode('utf-8')) to bridge the two.
        """
        self.segment_bounds = []
        if not self.segments or self.layout is None:
            return
        text = self.layout.get_text()
        if not text:
            return
        # Markup is duplicated for seamless scroll — only scan first half
        half_len = len(text) // 2
        cursor_char = 0
        bounds = []
        for seg in self.segments:
            sep_char = text.find("\u00b7", cursor_char, half_len)
            if sep_char == -1:
                break
            start_byte = len(text[:cursor_char].encode("utf-8"))
            end_byte = len(text[:sep_char + 1].encode("utf-8"))
            try:
                rect_start = self.layout.index_to_pos(start_byte)
                rect_end = self.layout.index_to_pos(end_byte)
                xs = rect_start.x / Pango.SCALE
                xe = rect_end.x / Pango.SCALE
            except Exception:
                break
            if xe < xs:
                xe = xs + 20
            bounds.append((xs, xe, seg))
            cursor_char = sep_char + 1
        self.segment_bounds = bounds

    def _rebuild_glyph_masks(self, height):
        """Rasterise the Pango layout once into FORMAT_A8 alpha masks.

        Produces up to two masks:
        - `_glyph_mask_fill`: show_layout at alpha=1 (the glyph shapes)
        - `_glyph_mask_outline`: layout_path + stroke at alpha=1 (the edge)

        Per-frame rendering uses `mask_surface` with a gradient source so
        the colour cycle animates while glyph rasterisation happens once
        per stream change. Cost moves from ~2–4 ms each frame to that same
        cost once per stream rotation.
        """
        w = max(int(math.ceil(self.half_w)), 1)
        y = (height - 19) / 2
        p = self.preset

        fill = _cairo.ImageSurface(_cairo.FORMAT_A8, w, height)
        tc = _cairo.Context(fill)
        tc.set_source_rgba(1, 1, 1, 1)
        tc.move_to(0, y)
        PangoCairo.update_layout(tc, self.layout)
        PangoCairo.show_layout(tc, self.layout)
        fill.flush()
        self._glyph_mask_fill = fill

        ow = getattr(p, "outline_width", 0.8)
        if ow > 0:
            outline = _cairo.ImageSurface(_cairo.FORMAT_A8, w, height)
            tc2 = _cairo.Context(outline)
            tc2.set_source_rgba(1, 1, 1, 1)
            tc2.set_line_width(ow)
            tc2.set_line_join(_cairo.LINE_JOIN_ROUND)
            tc2.move_to(0, y)
            PangoCairo.update_layout(tc2, self.layout)
            PangoCairo.layout_path(tc2, self.layout)
            tc2.stroke()
            outline.flush()
            self._glyph_mask_outline = outline
        else:
            self._glyph_mask_outline = None

    def _render_text_surface(self, width, height):
        if self._glyph_mask_fill is None:
            self._rebuild_glyph_masks(height)
        surf = _cairo.ImageSurface(_cairo.FORMAT_ARGB32, width, height)
        tc = _cairo.Context(surf)
        x0 = -self.offset
        p = self.preset
        ow = getattr(p, "outline_width", 0.8)
        pulse = getattr(p, "outline_pulse", False)

        # Tile the cached A8 masks at the same two x positions the old
        # per-frame `show_layout` loop used — preserves identical scroll
        # behaviour (including the intentional trailing gap when
        # `half_w * 2 < width`).
        xs = [x0, x0 + self.half_w]

        # Outline pass: dark static colour (non-pulse) or the cycling
        # gradient (pulse). The mask handles per-glyph shape; only the
        # source paint is per-frame.
        if ow > 0 and self._glyph_mask_outline is not None:
            if pulse:
                tc.set_source(make_gradient(x0, self.half_w, self.gradient_phase))
            else:
                tc.set_source_rgba(0.02, 0.03, 0.05, 0.6)
            for xi in xs:
                tc.mask_surface(self._glyph_mask_outline, xi, 0)

        # Fill pass: always the cycling gradient.
        tc.set_source(make_gradient(x0, self.half_w, self.gradient_phase))
        for xi in xs:
            tc.mask_surface(self._glyph_mask_fill, xi, 0)

        surf.flush()
        return surf

    def _apply_typewriter_clip(self, surf, width, height):
        """If reveal is active, clip to only the first N glyphs worth of pixels.

        Approximation: each glyph ~8px wide at 11pt Maple Mono. Use offset +
        reveal_chars*8 as the right edge.
        """
        if self.reveal_chars >= 99999:
            return surf
        glyph_w = 8
        reveal_px = int(self.reveal_chars * glyph_w - self.offset)
        if reveal_px >= width:
            return surf
        # Black out pixels past the reveal edge
        stride = surf.get_stride()
        data = np.frombuffer(surf.get_data(), dtype=np.uint8).reshape(height, stride)
        if reveal_px > 0:
            data[:, reveal_px * 4:] = 0
        else:
            data[:, :] = 0
        surf.mark_dirty()
        return surf

    def _apply_wave(self, surf, width, height):
        p = self.preset
        amp = getattr(p, 'wave_amp', 0)
        if amp <= 0:
            return surf
        stride = surf.get_stride()
        src = np.frombuffer(surf.get_data(), dtype=np.uint8).copy().reshape(height, stride)
        out = _cairo.ImageSurface(_cairo.FORMAT_ARGB32, width, height)
        out_stride = out.get_stride()
        dst = np.zeros((height, out_stride), dtype=np.uint8)
        freq = getattr(p, 'wave_freq', 0.015)
        speed = getattr(p, 'wave_speed', 2.0)
        strip_w = 8
        for x0 in range(0, width, strip_w):
            x1 = min(x0 + strip_w, width)
            cx = (x0 + x1) // 2
            dy = int(round(amp * math.sin(freq * cx + speed * self.time_s)))
            bx0, bx1 = x0 * 4, x1 * 4
            if dy == 0:
                dst[:, bx0:bx1] = src[:, bx0:bx1]
            elif dy > 0:
                dst[dy:, bx0:bx1] = src[:height - dy, bx0:bx1]
            else:
                dst[:height + dy, bx0:bx1] = src[-dy:, bx0:bx1]
        buf = out.get_data()
        buf[:] = dst.tobytes()
        out.mark_dirty()
        return out

    def _compute_glow(self, text_surf, width, height):
        p = self.preset
        if p.glow_kernel < 3:
            return None
        stride = text_surf.get_stride()
        argb = np.frombuffer(text_surf.get_data(), dtype=np.uint8).reshape(height, stride // 4, 4)
        alpha = argb[:, :width, 3].astype(np.float32)
        blurred = uniform_filter1d(alpha, size=p.glow_kernel, axis=1, mode='constant')
        blurred = np.clip(blurred, 0, 255).astype(np.uint8)
        glow = _cairo.ImageSurface(_cairo.FORMAT_A8, width, height)
        gs = glow.get_stride()
        gb = glow.get_data()
        gb[:] = np.hstack([blurred, np.zeros((height, gs - width), dtype=np.uint8)]).tobytes()
        glow.mark_dirty()
        return glow

    # ── Draw ──────────────────────────────────

    def _draw(self, da, cr, width, height):
        _draw_t0 = _time.monotonic_ns()
        p = self.preset

        if self.layout is None:
            self.layout = PangoCairo.create_layout(cr)
            self.layout.set_markup(self.ticker_markup, -1)
            _, logical = self.layout.get_pixel_extents()
            self.half_w = max(logical.width / 2.0, 1.0)
            self.offset = self.offset % self.half_w
            self._compute_segment_bounds()
            # Layout changed → glyph-mask caches are stale.
            self._glyph_mask_fill = None
            self._glyph_mask_outline = None

        # ═══ BACKGROUND ═══════════════════════════
        cr.set_source_rgba(0.020, 0.027, 0.051, p.bg_alpha)
        cr.paint()

        # Water caustic
        if p.water_skip > 0:
            if self.water_surf is None or self.water_frame % p.water_skip == 0:
                self.water_surf = water_caustic_surface(width, height, self.time_s)
            cr.set_source_surface(self.water_surf, 0, 0)
            cr.paint()

        # Scanlines
        if p.scanline_opacity > 0:
            cr.set_source_rgba(0, 0, 0, p.scanline_opacity)
            for row in range(0, height, 2):
                cr.rectangle(0, row, width, 1)
            cr.fill()

        # Holo shimmer (Phase 6) — every other row (stripe still reads as
        # shimmer to the eye at the current `shimmer_a` values) to halve the
        # per-frame cairo fill count. Rows are 2 px tall instead of 1.
        shimmer_a = getattr(p, "holo_shimmer", 0.0)
        if shimmer_a > 0:
            for i in range(0, height, 2):
                v = (math.sin(i * 4.2 + self.time_s * 3) *
                     math.sin(i * 0.7 - self.time_s * 1.5))
                v = (v + 1) / 2  # 0..1
                cr.set_source_rgba(0.161, 0.941, 1.000, shimmer_a * v)
                cr.rectangle(0, i, width, 2)
                cr.fill()

        # Top border (synthwave sweep or static gradient)
        if getattr(p, "synthwave_border", True):
            # Wider sweeping band with hot-pink to purple accent
            phase = self.time_s * 150
            grad = _cairo.LinearGradient(0, 0, width, 0)
            grad.set_extend(_cairo.Extend.REPEAT)
            # Offset colors by phase for sweeping feel
            sweep = (phase % 800) / 800.0
            grad.add_color_stop_rgb((0.0 + sweep) % 1.0, 1.000, 0.278, 0.820)  # magenta
            grad.add_color_stop_rgb((0.33 + sweep) % 1.0, 1.000, 0.361, 0.541)  # pink
            grad.add_color_stop_rgb((0.66 + sweep) % 1.0, 0.290, 0.659, 1.000)  # blue
            grad.add_color_stop_rgb((1.0 + sweep) % 1.0, 1.000, 0.278, 0.820)
            cr.set_source(grad)
        else:
            cr.set_source(make_gradient(0, width, self.gradient_phase))
        cr.rectangle(0, 0, width, 1)
        cr.fill()

        # ═══ FOREGROUND ═══════════════════════════
        text_surf = self._render_text_surface(width, height)
        text_surf = self._apply_typewriter_clip(text_surf, width, height)
        text_surf = self._apply_wave(text_surf, width, height)

        # Phosphor decay trail — composite old surfaces at diminishing alpha
        phosphor_a = getattr(p, "phosphor_trail", 0.0)
        if phosphor_a > 0:
            # Append current surface to ring buffer every 2 frames
            self.phosphor_frame_ctr += 1
            if self.phosphor_frame_ctr % 2 == 0:
                self.phosphor_ring.append(text_surf)
            for i, old in enumerate(list(self.phosphor_ring)[:-1]):
                age = len(self.phosphor_ring) - i
                alpha = phosphor_a / age
                cr.save()
                cr.set_source_rgba(0.161, 0.941, 1.000, alpha)
                cr.mask_surface(old, 0, 0)
                cr.restore()

        # Ghost echo (VHS double-image)
        echo_a = getattr(p, "ghost_echo", 0.0)
        if echo_a > 0:
            cr.save()
            cr.set_source_rgba(0.161, 0.941, 1.000, echo_a)
            cr.mask_surface(text_surf, -3 * getattr(p, "shadow_offset", 2), 0)
            cr.restore()

        # Glow (cached — recompute every 8 frames)
        self.glow_frame += 1
        if self.glow_cache is None or self.glow_frame % 8 == 0:
            self.glow_cache = self._compute_glow(text_surf, width, height)
        glow_a8 = self.glow_cache
        if glow_a8 is not None:
            glow_alpha = p.glow_base_alpha + p.glow_pulse_amp * math.sin(
                self.time_s * (2 * math.pi / max(p.glow_pulse_period, 0.1)))
            glow_alpha = max(0.0, min(1.0, glow_alpha))
            cr.save()
            cr.push_group()
            cr.set_source(make_gradient(0, width, self.gradient_phase))
            cr.mask_surface(glow_a8, 0, 0)
            pattern = cr.pop_group()
            cr.set_source(pattern)
            cr.paint_with_alpha(glow_alpha)
            cr.restore()

        # Shadow
        if p.shadow_alpha > 0:
            cr.set_source_rgba(0, 0, 0, p.shadow_alpha)
            cr.mask_surface(text_surf, p.shadow_offset, p.shadow_offset)

        # Sharp text
        cr.set_source_surface(text_surf, 0, 0)
        cr.paint()

        # Glitch + CA
        ca = getattr(self, "_runtime_ca_offset", p.ca_offset)
        if self.glitch_remaining > 0 and ca > 0:
            cr.save()
            cr.set_source_rgba(1, 0, 0, 0.3)
            cr.mask_surface(text_surf, ca, 0)
            cr.restore()
            cr.save()
            cr.set_source_rgba(0, 0.3, 1, 0.3)
            cr.mask_surface(text_surf, -ca, 0)
            cr.restore()
            for sy, sh, sx in self.glitch_strips:
                cr.save()
                cr.rectangle(0, sy, width, sh)
                cr.clip()
                cr.set_source_surface(text_surf, sx, 0)
                cr.paint()
                cr.restore()

        # ═══ POST-PASS ══════════════════════════
        # Edge fade vignette (FG mask)
        fade_px = getattr(p, "edge_fade", 0)
        if fade_px > 0 and width > 2 * fade_px:
            # Left fade
            lgrad = _cairo.LinearGradient(0, 0, fade_px, 0)
            lgrad.add_color_stop_rgba(0, 0.020, 0.027, 0.051, 1.0)
            lgrad.add_color_stop_rgba(1, 0.020, 0.027, 0.051, 0.0)
            cr.set_source(lgrad)
            cr.rectangle(0, 0, fade_px, height)
            cr.fill()
            # Right fade
            rgrad = _cairo.LinearGradient(width - fade_px, 0, width, 0)
            rgrad.add_color_stop_rgba(0, 0.020, 0.027, 0.051, 0.0)
            rgrad.add_color_stop_rgba(1, 0.020, 0.027, 0.051, 1.0)
            cr.set_source(rgrad)
            cr.rectangle(width - fade_px, 0, fade_px, height)
            cr.fill()

        # Progress indicator — 2px bar at bottom, filling as dwell elapses
        if getattr(p, "progress_bar", False):
            current = (self.pinned_stream if self.pinned_stream
                       else self.stream_order[self.stream_idx])
            interval = self._current_interval(current)
            elapsed = self.time_s - self.stream_start_s
            pct = min(elapsed / max(interval, 1), 1.0)
            prog_w = int(width * pct)
            # Dim rail spanning full width so the bar is always anchored
            cr.set_source_rgba(0.35, 0.40, 0.60, 0.22)
            cr.rectangle(0, height - 2, width, 2)
            cr.fill()
            if prog_w > 0:
                prog_grad = make_gradient(0, max(prog_w, 1), self.gradient_phase)
                cr.set_source(prog_grad)
                cr.rectangle(0, height - 2, prog_w, 2)
                cr.fill()

        # Stream-change wipe — brief gradient sweep the first 400ms after a
        # stream rotation, signalling new content has arrived.
        wipe_dur = 0.4
        wipe_age = self.time_s - self.stream_start_s
        if wipe_age < wipe_dur:
            progress = wipe_age / wipe_dur
            band_w = max(int(width * 0.18), 80)
            lead = int((width + band_w) * progress) - band_w
            alpha = 0.55 * (1.0 - progress)
            sweep = _cairo.LinearGradient(lead, 0, lead + band_w, 0)
            sweep.add_color_stop_rgba(0.0, 0.161, 0.941, 1.000, 0.0)
            sweep.add_color_stop_rgba(0.5, 1.000, 0.278, 0.820, alpha)
            sweep.add_color_stop_rgba(1.0, 0.161, 0.941, 1.000, 0.0)
            cr.set_source(sweep)
            cr.rectangle(0, 0, width, height)
            cr.fill()

        # Paused overlay
        if self.paused:
            cr.set_source_rgba(1, 0.278, 0.820, 0.08)
            cr.paint()

        # Adaptive quality: record this frame's render time; the EMA inside
        # `_update_adaptive_tier` decides whether to promote/demote tier.
        self._update_adaptive_tier((_time.monotonic_ns() - _draw_t0) / 1e6)


# ══════════════════════════════════════════════════
# DBus interface
# ══════════════════════════════════════════════════
#
# Exposed on the session bus so external tools can drive the ticker
# without racing on SIGUSR1 + state files. The primary instance
# (default STATE_DIR) owns `io.hairglasses.keybind_ticker`; secondary
# instances (e.g. DP-3 focus) own a suffixed name derived from their
# state-dir basename so both can coexist without collision.

DBUS_PATH = "/io/hairglasses/Ticker"
DBUS_IFACE = "io.hairglasses.Ticker"
DBUS_INTROSPECTION = """
<node>
  <interface name='io.hairglasses.Ticker'>
    <method name='Pin'>
      <arg name='stream' direction='in' type='s'/>
    </method>
    <method name='Unpin'/>
    <method name='PinToggle'/>
    <method name='NextStream'/>
    <method name='PrevStream'/>
    <method name='TogglePause'/>
    <method name='Shuffle'>
      <arg name='mode' direction='in' type='s'/>
    </method>
    <method name='SetPlaylist'>
      <arg name='name' direction='in' type='s'/>
    </method>
    <method name='SetPreset'>
      <arg name='name' direction='in' type='s'/>
    </method>
    <method name='ShowBanner'>
      <arg name='text'  direction='in' type='s'/>
      <arg name='color' direction='in' type='s'/>
    </method>
    <method name='SetUrgent'>
      <arg name='active' direction='in' type='b'/>
    </method>
    <method name='SnoozeUrgent'/>
    <method name='ReloadPlugins'/>
    <property name='CurrentStream' type='s' access='read'/>
    <property name='Playlist'      type='s' access='read'/>
    <property name='Pinned'        type='s' access='read'/>
    <property name='Paused'        type='b' access='read'/>
    <property name='Shuffle'       type='b' access='read'/>
    <property name='Tier'          type='i' access='read'/>
    <property name='EmaFrameMs'    type='d' access='read'/>
    <property name='Urgent'        type='b' access='read'/>
    <signal name='StreamChanged'>
      <arg name='name' type='s'/>
    </signal>
    <signal name='TierChanged'>
      <arg name='tier' type='i'/>
    </signal>
    <signal name='UrgentMode'>
      <arg name='active' type='b'/>
    </signal>
  </interface>
</node>
"""


def _dbus_bus_name_for_instance():
    """Return the well-known bus name for this instance.

    Primary (default state dir) → `io.hairglasses.keybind_ticker`.
    Secondary (via --state-dir) → suffix with sanitised basename so
    each instance owns a distinct name and `hg ticker` can target
    them separately when desired.
    """
    if not STATE_DIR_OVERRIDE:
        return "io.hairglasses.keybind_ticker"
    base = os.path.basename(os.path.normpath(STATE_DIR))
    # "keybind-ticker-DP-3_focus" → "DP_3_focus"
    suffix = base.replace("keybind-ticker-", "").replace("-", "_")
    suffix = re.sub(r"[^A-Za-z0-9_]", "_", suffix) or "secondary"
    return f"io.hairglasses.keybind_ticker.{suffix}"


class TickerApp(Gtk.Application):
    def __init__(self):
        super().__init__(
            application_id="io.hairglasses.keybind_ticker",
            flags=Gio.ApplicationFlags.NON_UNIQUE,
        )
        self._bus_conn = None
        self._bus_registration_id = 0
        self._bus_name_id = 0

    # ── DBus registration ─────────────────────────
    def _register_dbus(self):
        """Own the instance bus name and register the Ticker object.

        Uses `Gio.bus_own_name` rather than relying on Gtk.Application's
        auto-registration because NON_UNIQUE suppresses that path.
        """
        bus_name = _dbus_bus_name_for_instance()

        def on_bus_acquired(conn, name):
            self._bus_conn = conn
            try:
                node = Gio.DBusNodeInfo.new_for_xml(DBUS_INTROSPECTION)
                iface = node.lookup_interface(DBUS_IFACE)
                self._bus_registration_id = conn.register_object(
                    DBUS_PATH, iface,
                    self._on_method_call,
                    self._on_get_property,
                    None,  # no writable properties
                )
            except Exception as e:
                sys.stderr.write(f"dbus register failed: {e}\n")

        def on_name_lost(conn, name):
            sys.stderr.write(f"dbus: lost name {name}\n")

        self._bus_name_id = Gio.bus_own_name(
            Gio.BusType.SESSION,
            bus_name,
            Gio.BusNameOwnerFlags.NONE,
            on_bus_acquired,
            None,       # name acquired callback — we register on bus_acquired
            on_name_lost,
        )

    def _on_method_call(self, conn, sender, path, iface, method, params, invocation):
        win = getattr(self, "_window", None)
        if win is None:
            invocation.return_error_literal(
                Gio.dbus_error_quark(), Gio.DBusError.FAILED,
                "ticker window not ready")
            return

        def done():
            invocation.return_value(None)

        try:
            if method == "Pin":
                (stream,) = params.unpack()
                GLib.idle_add(win._dbus_apply_pin, stream)
            elif method == "Unpin":
                GLib.idle_add(win._dbus_apply_pin, None)
            elif method == "PinToggle":
                GLib.idle_add(win._dbus_toggle_pin)
            elif method == "NextStream":
                GLib.idle_add(win._dbus_advance, 1)
            elif method == "PrevStream":
                GLib.idle_add(win._dbus_advance, -1)
            elif method == "TogglePause":
                GLib.idle_add(win._dbus_toggle_pause)
            elif method == "Shuffle":
                (mode,) = params.unpack()
                GLib.idle_add(win._dbus_set_shuffle, mode)
            elif method == "SetPlaylist":
                (name,) = params.unpack()
                GLib.idle_add(win._dbus_set_playlist, name)
            elif method == "SetPreset":
                (name,) = params.unpack()
                GLib.idle_add(win._dbus_set_preset, name)
            elif method == "ShowBanner":
                text, color = params.unpack()
                GLib.idle_add(win._dbus_show_banner, text, color)
            elif method == "SetUrgent":
                (active,) = params.unpack()
                GLib.idle_add(win._dbus_set_urgent, bool(active))
            elif method == "SnoozeUrgent":
                GLib.idle_add(win._clear_urgent_mode)
            elif method == "ReloadPlugins":
                GLib.idle_add(win._dbus_reload_plugins)
            else:
                invocation.return_error_literal(
                    Gio.dbus_error_quark(), Gio.DBusError.UNKNOWN_METHOD,
                    f"unknown method: {method}")
                return
        except Exception as e:
            invocation.return_error_literal(
                Gio.dbus_error_quark(), Gio.DBusError.FAILED, str(e))
            return
        done()

    def _on_get_property(self, conn, sender, path, iface, prop):
        win = getattr(self, "_window", None)
        if win is None:
            return GLib.Variant("s", "") if prop in ("CurrentStream", "Playlist", "Pinned") else GLib.Variant("b", False)
        if prop == "CurrentStream":
            cur = (win.stream_order[win.stream_idx]
                   if win.stream_order else "")
            return GLib.Variant("s", cur or "")
        if prop == "Playlist":
            return GLib.Variant("s", win.playlist_name or "")
        if prop == "Pinned":
            return GLib.Variant("s", win.pinned_stream or "")
        if prop == "Paused":
            return GLib.Variant("b", bool(win.paused))
        if prop == "Shuffle":
            return GLib.Variant("b", bool(win.shuffle))
        if prop == "Tier":
            return GLib.Variant("i", int(win._tier))
        if prop == "EmaFrameMs":
            return GLib.Variant("d", float(win._ema_frame_ms))
        if prop == "Urgent":
            return GLib.Variant("b", bool(win.urgent_mode))
        return None

    def _emit_signal(self, signal_name, variant_tuple):
        """Emit a DBus signal (best-effort; silently skip if unregistered)."""
        if not self._bus_conn:
            return
        try:
            self._bus_conn.emit_signal(
                None, DBUS_PATH, DBUS_IFACE, signal_name,
                GLib.Variant.new_tuple(*variant_tuple),
            )
        except Exception:
            pass

    def do_activate(self):
        win = TickerWindow(preset_name=START_PRESET, application=self)
        win.present()
        self._window = win

        # Register DBus once the window is up. This runs on the next
        # idle tick so the window reference is stable before any method
        # call lands.
        GLib.idle_add(lambda: (self._register_dbus(), False)[1])

        # SIGUSR1 → reload state files inline (no restart, no flicker).
        # signal.signal handlers fire in the signal thread; marshal the work
        # onto the GTK main loop via GLib.idle_add so widget calls are safe.
        def _on_sigusr1(signum, frame):
            GLib.idle_add(win.reload_from_state)
        try:
            signal.signal(signal.SIGUSR1, _on_sigusr1)
        except (ValueError, OSError):
            # Not in main thread or platform mismatch — ignore silently;
            # `hg ticker` will fall back to `systemctl --user restart`.
            pass


if __name__ == "__main__":
    # Hide our custom CLI flags from Gtk.Application
    TickerApp().run([sys.argv[0]])
