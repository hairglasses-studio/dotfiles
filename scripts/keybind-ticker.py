#!/usr/bin/env python3
"""keybind-ticker.py — Cyberpunk keybind ticker for Hyprland.

GTK4 DrawingArea with PangoCairo at 240Hz frame-clock sync.
Effect stack (background → foreground for readability):
  water caustic → scanlines → glow → shadow → crisp gradient text
  + periodic glitch/chromatic aberration

Usage:
  keybind-ticker.py              # regular window (tiles with hy3)
  keybind-ticker.py --layer      # layer-shell bar (needs LD_PRELOAD)
"""

import gi
import subprocess
import json
import sys
import math
import random
import cairo as _cairo
import numpy as np
from scipy.ndimage import uniform_filter1d
from html import escape

gi.require_version("Gtk", "4.0")
gi.require_version("Gdk", "4.0")
gi.require_version("Pango", "1.0")
gi.require_version("PangoCairo", "1.0")

from gi.repository import Gtk, Gdk, Pango, PangoCairo, GLib

LAYER_MODE = "--layer" in sys.argv

if LAYER_MODE:
    gi.require_version("Gtk4LayerShell", "1.0")
    from gi.repository import Gtk4LayerShell

# ── Config ────────────────────────────────────────
BAR_H = 28
SPEED = 55.0            # px/sec text scroll
GRADIENT_SPEED = 40.0   # px/sec gradient phase drift
GRADIENT_SPAN = 800.0   # px per full color cycle
REFRESH_S = 300         # rebuild keybinds every 5 min
GLOW_KERNEL = 17        # horizontal blur kernel size
GLOW_BASE_ALPHA = 0.35  # glow base intensity
GLOW_PULSE_AMP = 0.15   # glow pulse amplitude (±)
GLOW_PULSE_PERIOD = 3.5 # seconds per breathe cycle
WATER_SKIP = 4          # compute water caustic every N frames
SCANLINE_OPACITY = 0.08 # subtle CRT scanlines (background only)
SHADOW_OFFSET = 2       # drop shadow px offset
SHADOW_ALPHA = 0.30     # drop shadow opacity
GLITCH_PROB = 0.004     # per-frame probability (~1/sec at 240Hz)
GLITCH_FRAMES = 4       # duration in frames
GLITCH_STRIPS = 3       # number of displaced strips
GLITCH_MAX_SHIFT = 10   # max horizontal displacement px
CA_OFFSET = 3           # chromatic aberration px offset

# Maple font weight cycle
FONTS = [
    "Maple Mono NF CN 11",
    "Maple Mono NF CN Bold 11",
    "Maple Mono NF CN Italic 11",
    "Maple Mono NF CN SemiBold 11",
    "Maple Mono NF CN Light 11",
    "Maple Mono NF CN Medium 11",
    "Maple Mono NF CN ExtraBold 11",
    "Maple Mono NF CN Thin 11",
    "Maple Mono NF CN ExtraLight 11",
    "Maple Mono NF CN Bold Italic 11",
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

BG = (0.020, 0.027, 0.051, 0.82)


# ── Helpers ───────────────────────────────────────

def fmt_mods(mask):
    out = ""
    if mask & 64: out += "Super+"
    if mask & 1:  out += "Shift+"
    if mask & 4:  out += "Ctrl+"
    if mask & 8:  out += "Alt+"
    return out


def build_ticker_markup():
    try:
        raw = subprocess.run(
            ["hyprctl", "binds", "-j"],
            capture_output=True, text=True, timeout=5,
        ).stdout
        binds = json.loads(raw)
    except Exception:
        return '<span font_desc="Maple Mono NF CN 11">  No keybinds loaded  \u00b7</span>'

    parts = []
    font_count = len(FONTS)
    i = 0
    for b in binds:
        if b.get("has_description") and not b.get("submap") and not b.get("mouse"):
            mods = fmt_mods(b["modmask"])
            desc = escape(b["description"])
            key = escape(f"{mods}{b['key']}")
            font = FONTS[i % font_count]
            parts.append(
                f'<span font_desc="{font}">  {desc}  {key}  \u00b7</span>'
            )
            i += 1

    single = "".join(parts)
    return single + single


def make_gradient(x_start, total_width, phase):
    x0 = x_start - phase
    grad = _cairo.LinearGradient(x0, 0, x0 + total_width, 0)
    grad.set_extend(_cairo.Extend.REPEAT)
    n = len(GRADIENT_COLORS)
    for i, (r, g, b) in enumerate(GRADIENT_COLORS):
        grad.add_color_stop_rgb(i / (n - 1), r, g, b)
    return grad


# ── Water caustic (ported from kitty/shaders/darkwindow/water.glsl) ──

_WATER_SCALE = 4
_WATER_MAX_ITER = 6
_WATER_TAU = 6.28318530718


def compute_water_caustic(width, height, time_s):
    sw = width // _WATER_SCALE
    sh = max(height // _WATER_SCALE, 2)
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
        ix_new = px + np.cos(tn - ix) + np.sin(tn + iy)
        iy_new = py + np.sin(tn - iy) + np.cos(tn + ix)
        ix, iy = ix_new, iy_new
        denom = np.sqrt(
            (px / (np.sin(ix + tn) / inten)) ** 2 +
            (py / (np.cos(iy + tn) / inten)) ** 2
        )
        c += 1.0 / np.maximum(denom, 1e-6)

    c /= _WATER_MAX_ITER
    c = 1.17 - np.power(c, 1.4)
    brightness = np.power(np.abs(c), 15.0)
    brightness = np.clip(brightness * 1.2, 0, 1)

    if _WATER_SCALE > 1:
        brightness = np.repeat(np.repeat(brightness, _WATER_SCALE, axis=0), _WATER_SCALE, axis=1)
        brightness = brightness[:height, :width]

    return brightness


def water_caustic_surface(width, height, time_s):
    bright = compute_water_caustic(width, height, time_s)

    r = (bright * 0.10 * 255).astype(np.uint8)
    g = (bright * 0.55 * 255).astype(np.uint8)
    b = (bright * 0.75 * 255).astype(np.uint8)
    a = (bright * 0.25 * 255).astype(np.uint8)

    argb = np.zeros((height, width, 4), dtype=np.uint8)
    argb[:, :, 0] = b
    argb[:, :, 1] = g
    argb[:, :, 2] = r
    argb[:, :, 3] = a

    surf = _cairo.ImageSurface(_cairo.FORMAT_ARGB32, width, height)
    stride = surf.get_stride()
    buf = surf.get_data()
    for row in range(height):
        buf[row * stride: row * stride + width * 4] = argb[row].tobytes()
    surf.mark_dirty()
    return surf


# ── Main Window ───────────────────────────────────

class TickerWindow(Gtk.ApplicationWindow):
    def __init__(self, **kwargs):
        super().__init__(**kwargs)
        self.set_title("keybind-ticker")
        self.set_default_size(1200, BAR_H)

        if LAYER_MODE:
            Gtk4LayerShell.init_for_window(self)
            Gtk4LayerShell.set_layer(self, Gtk4LayerShell.Layer.BOTTOM)
            Gtk4LayerShell.set_anchor(self, Gtk4LayerShell.Edge.BOTTOM, True)
            Gtk4LayerShell.set_anchor(self, Gtk4LayerShell.Edge.LEFT, True)
            Gtk4LayerShell.set_anchor(self, Gtk4LayerShell.Edge.RIGHT, True)
            Gtk4LayerShell.set_exclusive_zone(self, BAR_H)
            Gtk4LayerShell.set_namespace(self, "keybind-ticker")
            display = Gdk.Display.get_default()
            if display:
                monitors = display.get_monitors()
                for i in range(monitors.get_n_items()):
                    mon = monitors.get_item(i)
                    if mon and "DP-3" in (mon.get_connector() or ""):
                        Gtk4LayerShell.set_monitor(self, mon)
                        break

        self.da = Gtk.DrawingArea()
        self.da.set_content_height(BAR_H)
        self.da.set_vexpand(True)
        self.da.set_hexpand(True)
        self.da.set_draw_func(self._draw)
        self.set_child(self.da)

        # Scroll / animation state
        self.offset = 0.0
        self.gradient_phase = 0.0
        self.time_s = 0.0
        self.last_us = None
        self.layout = None
        self.half_w = 1.0

        # Glitch state machine
        self.glitch_remaining = 0
        self.glitch_strips = []

        # Water caustic cache
        self.water_surf = None
        self.water_frame = 0

        self._rebuild()
        GLib.timeout_add_seconds(REFRESH_S, self._rebuild)
        self.da.add_tick_callback(self._tick)

    def _rebuild(self):
        self.ticker_markup = build_ticker_markup()
        self.layout = None
        self.da.queue_draw()
        return GLib.SOURCE_CONTINUE

    def _tick(self, widget, frame_clock):
        now = frame_clock.get_frame_time()
        if self.last_us is not None:
            dt = min((now - self.last_us) / 1_000_000.0, 0.05)
            self.offset += SPEED * dt
            self.gradient_phase += GRADIENT_SPEED * dt
            self.time_s += dt
            if self.half_w > 0 and self.offset >= self.half_w:
                self.offset -= self.half_w
            if self.gradient_phase >= GRADIENT_SPAN:
                self.gradient_phase -= GRADIENT_SPAN

            # Glitch trigger
            if self.glitch_remaining > 0:
                self.glitch_remaining -= 1
            elif random.random() < GLITCH_PROB:
                self.glitch_remaining = GLITCH_FRAMES
                self.glitch_strips = [
                    (random.randint(0, BAR_H - 4),
                     random.randint(3, 8),
                     random.randint(-GLITCH_MAX_SHIFT, GLITCH_MAX_SHIFT))
                    for _ in range(GLITCH_STRIPS)
                ]

        self.last_us = now
        self.water_frame += 1
        widget.queue_draw()
        return GLib.SOURCE_CONTINUE

    def _render_text_surface(self, width, height):
        surf = _cairo.ImageSurface(_cairo.FORMAT_ARGB32, width, height)
        tc = _cairo.Context(surf)

        x = -self.offset
        y = (height - 14) / 2
        text_grad = make_gradient(x, self.half_w, self.gradient_phase)
        tc.set_source(text_grad)

        tc.move_to(x, y)
        PangoCairo.show_layout(tc, self.layout)
        tc.move_to(x + self.half_w, y)
        PangoCairo.show_layout(tc, self.layout)

        surf.flush()
        return surf

    def _compute_glow(self, text_surf, width, height):
        stride = text_surf.get_stride()
        argb = np.frombuffer(text_surf.get_data(), dtype=np.uint8).reshape(height, stride // 4, 4)
        alpha = argb[:, :width, 3].astype(np.float32)

        blurred = uniform_filter1d(alpha, size=GLOW_KERNEL, axis=1, mode='constant')
        blurred = np.clip(blurred, 0, 255).astype(np.uint8)

        glow = _cairo.ImageSurface(_cairo.FORMAT_A8, width, height)
        glow_stride = glow.get_stride()
        glow_buf = glow.get_data()
        for row in range(height):
            glow_buf[row * glow_stride: row * glow_stride + width] = blurred[row].tobytes()
        glow.mark_dirty()
        return glow

    def _draw(self, da, cr, width, height):
        # Build layout once per markup rebuild
        if self.layout is None:
            self.layout = PangoCairo.create_layout(cr)
            self.layout.set_markup(self.ticker_markup, -1)
            _, logical = self.layout.get_pixel_extents()
            self.half_w = logical.width / 2.0
            if self.half_w > 0:
                self.offset = self.offset % self.half_w

        # ═══════════════════════════════════════════
        # BACKGROUND LAYERS (effects live here)
        # ═══════════════════════════════════════════

        # ── BG: Dark panel ───────────────────────────
        cr.set_source_rgba(*BG)
        cr.paint()

        # ── BG: Water caustic ────────────────────────
        if self.water_surf is None or self.water_frame % WATER_SKIP == 0:
            self.water_surf = water_caustic_surface(width, height, self.time_s)
        cr.set_source_surface(self.water_surf, 0, 0)
        cr.paint()

        # ── BG: CRT scanlines (under text) ───────────
        cr.set_source_rgba(0, 0, 0, SCANLINE_OPACITY)
        for row in range(0, height, 2):
            cr.rectangle(0, row, width, 1)
        cr.fill()

        # ── BG: Top border (animated gradient) ───────
        border_grad = make_gradient(0, width, self.gradient_phase)
        cr.set_source(border_grad)
        cr.rectangle(0, 0, width, 1)
        cr.fill()

        # ═══════════════════════════════════════════
        # FOREGROUND LAYERS (text stays crisp here)
        # ═══════════════════════════════════════════

        # ── Render text to offscreen surface ─────────
        text_surf = self._render_text_surface(width, height)

        # ── FG: Neon glow (breathing halo) ───────────
        glow_a8 = self._compute_glow(text_surf, width, height)
        glow_alpha = GLOW_BASE_ALPHA + GLOW_PULSE_AMP * math.sin(
            self.time_s * (2 * math.pi / GLOW_PULSE_PERIOD)
        )
        glow_grad = make_gradient(0, width, self.gradient_phase)
        cr.save()
        cr.set_source(glow_grad)
        cr.mask_surface(glow_a8, 0, 0)
        cr.restore()

        # ── FG: Drop shadow ──────────────────────────
        cr.set_source_rgba(0, 0, 0, SHADOW_ALPHA)
        cr.mask_surface(text_surf, SHADOW_OFFSET, SHADOW_OFFSET)

        # ── FG: Sharp gradient text (topmost) ────────
        cr.set_source_surface(text_surf, 0, 0)
        cr.paint()

        # ── FG: Glitch effects (brief, periodic) ─────
        if self.glitch_remaining > 0:
            # Chromatic aberration
            cr.save()
            cr.set_source_rgba(1, 0, 0, 0.3)
            cr.mask_surface(text_surf, CA_OFFSET, 0)
            cr.restore()
            cr.save()
            cr.set_source_rgba(0, 0.3, 1, 0.3)
            cr.mask_surface(text_surf, -CA_OFFSET, 0)
            cr.restore()
            # Strip displacement
            for strip_y, strip_h, shift in self.glitch_strips:
                cr.save()
                cr.rectangle(0, strip_y, width, strip_h)
                cr.clip()
                cr.set_source_surface(text_surf, shift, 0)
                cr.paint()
                cr.restore()


class TickerApp(Gtk.Application):
    def __init__(self):
        super().__init__(application_id="io.hairglasses.keybind_ticker")

    def do_activate(self):
        win = TickerWindow(application=self)
        win.present()


if __name__ == "__main__":
    TickerApp().run()
