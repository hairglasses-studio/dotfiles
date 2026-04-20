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
import os
import subprocess
import sys
import time

sys.path.insert(0, os.path.join(os.path.dirname(os.path.realpath(__file__)), "lib"))
import ticker_render as tr  # noqa: E402

import gi

gi.require_version("Gtk", "4.0")
gi.require_version("Gtk4LayerShell", "1.0")
gi.require_version("Gdk", "4.0")
gi.require_version("Pango", "1.0")
gi.require_version("PangoCairo", "1.0")

from gi.repository import Gtk, Gtk4LayerShell, Gdk, GLib, Gio, Pango, PangoCairo  # noqa: E402

BAR_H = tr.BAR_H


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

        tr.setup_layer_shell(
            self,
            (Gtk4LayerShell.Edge.TOP, Gtk4LayerShell.Edge.LEFT, Gtk4LayerShell.Edge.RIGHT),
            "hg-lyrics",
            monitor_name,
            layer="OVERLAY",
            exclusive_zone=0,
        )
        self.da = tr.make_drawing_area(BAR_H, self._draw)
        self.set_child(self.da)
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
        tr.fill_bg(cr, w, h, alpha=0.85)
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
