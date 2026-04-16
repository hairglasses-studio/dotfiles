#!/usr/bin/env python3
"""keybind-ticker.py — Pixel-smooth scrolling keybind ticker for Hyprland.

GTK4 DrawingArea with PangoCairo rendering at float pixel offsets,
synced to the display frame clock via add_tick_callback (240Hz on DP-3).

Usage:
  keybind-ticker.py              # regular window (tiles with hy3)
  keybind-ticker.py --layer      # layer-shell bar (needs LD_PRELOAD)
"""

import gi
import subprocess
import json
import sys

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
SPEED = 55.0        # px/sec scroll speed
FONT = "Maple Mono NF CN 11"
REFRESH_S = 300     # rebuild keybind list every 5 min

# Hairglasses Neon palette
BG      = (0.020, 0.027, 0.051, 0.92)
CYAN    = (0.161, 0.941, 1.0)
MAGENTA = (1.0,   0.278, 0.820)
GREEN   = (0.239, 1.0,   0.710)
DIM     = (0.400, 0.440, 0.561)


def fmt_mods(mask):
    out = ""
    if mask & 64: out += "Super+"
    if mask & 1:  out += "Shift+"
    if mask & 4:  out += "Ctrl+"
    if mask & 8:  out += "Alt+"
    return out


def build_ticker_text():
    try:
        raw = subprocess.run(
            ["hyprctl", "binds", "-j"],
            capture_output=True, text=True, timeout=5,
        ).stdout
        binds = json.loads(raw)
    except Exception:
        return "  No keybinds loaded  \u00b7"

    parts = []
    for b in binds:
        if b.get("has_description") and not b.get("submap") and not b.get("mouse"):
            mods = fmt_mods(b["modmask"])
            parts.append(f"  {b['description']}  {mods}{b['key']}  \u00b7")
    text = "".join(parts)
    return text + text  # doubled for seamless wrap


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
        self.last_us = None
        self.layout = None
        self.half_w = 1.0

        self._rebuild_text()
        GLib.timeout_add_seconds(REFRESH_S, self._rebuild_text)
        self.da.add_tick_callback(self._tick)

    def _rebuild_text(self):
        self.ticker_text = build_ticker_text()
        self.layout = None
        self.da.queue_draw()
        return GLib.SOURCE_CONTINUE

    def _tick(self, widget, frame_clock):
        now = frame_clock.get_frame_time()
        if self.last_us is not None:
            dt = min((now - self.last_us) / 1_000_000.0, 0.05)
            self.offset += SPEED * dt
            if self.half_w > 0 and self.offset >= self.half_w:
                self.offset -= self.half_w
        self.last_us = now
        widget.queue_draw()
        return GLib.SOURCE_CONTINUE

    def _draw(self, da, cr, width, height):
        # Background
        cr.set_source_rgba(*BG)
        cr.paint()

        # Top border
        cr.set_source_rgba(*CYAN, 0.30)
        cr.rectangle(0, 0, width, 1)
        cr.fill()

        # Build layout once per text rebuild
        if self.layout is None:
            self.layout = PangoCairo.create_layout(cr)
            fd = Pango.FontDescription.from_string(FONT)
            self.layout.set_font_description(fd)
            self.layout.set_text(self.ticker_text, -1)
            _, logical = self.layout.get_pixel_extents()
            self.half_w = logical.width / 2.0
            if self.half_w > 0:
                self.offset = self.offset % self.half_w

        # Scrolling text at sub-pixel offset
        cr.set_source_rgb(*CYAN)
        x = -self.offset
        y = (height - 14) / 2  # vertically center 11px font
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
