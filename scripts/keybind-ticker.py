#!/usr/bin/env python3
"""keybind-ticker.py — Pixel-smooth scrolling keybind ticker for Hyprland.

GTK4 DrawingArea with PangoCairo rendering at float pixel offsets,
synced to the display frame clock via add_tick_callback (240Hz on DP-3).
Each keybind entry cycles through Maple font weights. Text is painted
with an animated neon color gradient that flows across the bar.

Usage:
  keybind-ticker.py              # regular window (tiles with hy3)
  keybind-ticker.py --layer      # layer-shell bar (needs LD_PRELOAD)
"""

import gi
import subprocess
import json
import sys
import math
import cairo as _cairo  # for LinearGradient
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
SPEED = 55.0           # px/sec text scroll speed
GRADIENT_SPEED = 40.0  # px/sec gradient phase drift (independent of scroll)
GRADIENT_SPAN = 800.0  # px width of one full color cycle
REFRESH_S = 300        # rebuild keybind list every 5 min

# Maple font weight cycle — each keybind gets a different variant
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

# Hairglasses Neon gradient stops (looping: cyan → magenta → green → yellow → pink → cyan)
GRADIENT_COLORS = [
    (0.161, 0.941, 1.000),  # #29f0ff cyan
    (1.000, 0.278, 0.820),  # #ff47d1 magenta
    (0.239, 1.000, 0.710),  # #3dffb5 green
    (1.000, 0.894, 0.369),  # #ffe45e yellow
    (1.000, 0.361, 0.541),  # #ff5c8a pink
    (0.290, 0.659, 1.000),  # #4aa8ff blue
    (0.161, 0.941, 1.000),  # #29f0ff cyan (wrap)
]

BG = (0.020, 0.027, 0.051, 0.92)
CYAN = (0.161, 0.941, 1.0)


def fmt_mods(mask):
    out = ""
    if mask & 64: out += "Super+"
    if mask & 1:  out += "Shift+"
    if mask & 4:  out += "Ctrl+"
    if mask & 8:  out += "Alt+"
    return out


def build_ticker_markup():
    """Build Pango markup with cycling font weights — no foreground attrs
    so Cairo gradient controls all text color."""
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
                f'<span font_desc="{font}">'
                f'  {desc}  {key}  \u00b7'
                f'</span>'
            )
            i += 1

    single = "".join(parts)
    return single + single  # doubled for seamless wrap


def make_gradient(x_start, total_width, phase):
    """Create a Cairo LinearGradient cycling through the neon palette.
    `phase` shifts the gradient origin for animation."""
    x0 = x_start - phase
    x1 = x0 + total_width
    grad = _cairo.LinearGradient(x0, 0, x1, 0)
    grad.set_extend(_cairo.Extend.REPEAT)

    n = len(GRADIENT_COLORS)
    for i, (r, g, b) in enumerate(GRADIENT_COLORS):
        grad.add_color_stop_rgb(i / (n - 1), r, g, b)

    return grad


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

        self.offset = 0.0
        self.gradient_phase = 0.0
        self.last_us = None
        self.layout = None
        self.half_w = 1.0

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
            if self.half_w > 0 and self.offset >= self.half_w:
                self.offset -= self.half_w
            if self.gradient_phase >= GRADIENT_SPAN:
                self.gradient_phase -= GRADIENT_SPAN
        self.last_us = now
        widget.queue_draw()
        return GLib.SOURCE_CONTINUE

    def _draw(self, da, cr, width, height):
        # Background
        cr.set_source_rgba(*BG)
        cr.paint()

        # Top border — animated gradient matching text
        border_grad = make_gradient(0, width, self.gradient_phase)
        cr.set_source(border_grad)
        cr.rectangle(0, 0, width, 1)
        cr.fill()

        # Build layout once per markup rebuild
        if self.layout is None:
            self.layout = PangoCairo.create_layout(cr)
            self.layout.set_markup(self.ticker_markup, -1)
            _, logical = self.layout.get_pixel_extents()
            self.half_w = logical.width / 2.0
            if self.half_w > 0:
                self.offset = self.offset % self.half_w

        # Animated neon gradient as text paint source
        x = -self.offset
        y = (height - 14) / 2
        text_grad = make_gradient(x, self.half_w, self.gradient_phase)
        cr.set_source(text_grad)

        # Draw scrolling text (two copies for seamless wrap)
        cr.move_to(x, y)
        PangoCairo.show_layout(cr, self.layout)
        cr.move_to(x + self.half_w, y)
        PangoCairo.show_layout(cr, self.layout)


class TickerApp(Gtk.Application):
    def __init__(self):
        super().__init__(application_id="io.hairglasses.keybind_ticker")

    def do_activate(self):
        win = TickerWindow(application=self)
        win.present()


if __name__ == "__main__":
    TickerApp().run()
