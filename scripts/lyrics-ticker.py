#!/usr/bin/env python3
"""lyrics-ticker.py — now-playing banner. v1 shows track metadata; future
revisions will sync to time-stamped .lrc files co-located with the audio.

Polls `playerctl metadata` every 2s; if the player is playing, displays:
    ♪ {title} — {artist}   [album]  pos/duration
Otherwise shows a muted IDLE sentinel. 28px layer-shell strip, same form
factor as keybind-ticker but without scroll or stream cycling.
"""

from __future__ import annotations

import argparse
import subprocess
import sys
import time

import gi

gi.require_version("Gtk", "4.0")
gi.require_version("Gtk4LayerShell", "1.0")
gi.require_version("Gdk", "4.0")
gi.require_version("Pango", "1.0")
gi.require_version("PangoCairo", "1.0")

from gi.repository import Gtk, Gtk4LayerShell, Gdk, GLib, Gio, Pango, PangoCairo  # noqa: E402

BAR_H = 28


def _playerctl(fmt: str) -> str:
    try:
        out = subprocess.run(
            ["playerctl", "metadata", "--format", fmt],
            capture_output=True, text=True, timeout=1,
        ).stdout.strip()
        return out
    except Exception:
        return ""


def _format_time(micros: str) -> str:
    try:
        s = int(micros) // 1_000_000
    except (ValueError, TypeError):
        return "--:--"
    mm, ss = divmod(s, 60)
    hh, mm = divmod(mm, 60)
    return f"{hh}:{mm:02d}:{ss:02d}" if hh else f"{mm}:{ss:02d}"


class LyricsWindow(Gtk.ApplicationWindow):
    def __init__(self, app, monitor_name: str):
        super().__init__(application=app)
        self.line = "\u266a IDLE"
        self.color = "#66708f"

        Gtk4LayerShell.init_for_window(self)
        Gtk4LayerShell.set_layer(self, Gtk4LayerShell.Layer.OVERLAY)
        for edge in (Gtk4LayerShell.Edge.TOP, Gtk4LayerShell.Edge.LEFT, Gtk4LayerShell.Edge.RIGHT):
            Gtk4LayerShell.set_anchor(self, edge, True)
        Gtk4LayerShell.set_exclusive_zone(self, 0)
        Gtk4LayerShell.set_namespace(self, "hg-lyrics")
        display = Gdk.Display.get_default()
        if display:
            for i in range(display.get_monitors().get_n_items()):
                mon = display.get_monitors().get_item(i)
                if mon and monitor_name in (mon.get_connector() or ""):
                    Gtk4LayerShell.set_monitor(self, mon)
                    break

        self.da = Gtk.DrawingArea()
        self.da.set_content_height(BAR_H)
        self.da.set_vexpand(True)
        self.da.set_hexpand(True)
        self.da.set_draw_func(self._draw)
        self.set_child(self.da)
        self.da.connect("realize", lambda w: w.get_frame_clock().begin_updating())
        self.present()

        GLib.timeout_add(2000, self._refresh)
        self._refresh()

    def _refresh(self) -> bool:
        status = _playerctl("{{status}}")
        if status != "Playing":
            self.line = "\u266a IDLE"
            self.color = "#66708f"
        else:
            title = _playerctl("{{title}}") or "—"
            artist = _playerctl("{{artist}}") or "—"
            pos = _format_time(_playerctl("{{position}}"))
            dur = _format_time(_playerctl("{{mpris:length}}"))
            self.line = f"\u266a {title} — {artist}   {pos}/{dur}"
            self.color = "#ff47d1"
        self.da.queue_draw()
        return True

    def _draw(self, widget, cr, w, h):
        cr.set_source_rgba(0.02, 0.03, 0.05, 0.85)
        cr.rectangle(0, 0, w, h)
        cr.fill()
        layout = PangoCairo.create_layout(cr)
        layout.set_font_description(Pango.FontDescription("Maple Mono NF CN Bold 11"))
        layout.set_markup(
            f'<span foreground="{self.color}">{GLib.markup_escape_text(self.line)}</span>',
            -1,
        )
        tw, th = layout.get_pixel_size()
        cr.move_to(max(8, (w - tw) / 2), max(0, (h - th) / 2))
        PangoCairo.show_layout(cr, layout)


class LyricsApp(Gtk.Application):
    def __init__(self, monitor: str):
        super().__init__(
            application_id="io.hairglasses.lyrics",
            flags=Gio.ApplicationFlags.NON_UNIQUE,
        )
        self.monitor = monitor

    def do_activate(self):
        LyricsWindow(self, self.monitor)


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument("--monitor", default="DP-2")
    args = ap.parse_args()
    app = LyricsApp(args.monitor)
    return app.run([sys.argv[0]])


if __name__ == "__main__":
    sys.exit(main())
