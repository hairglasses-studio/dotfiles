#!/usr/bin/env python3
"""rsvp-ticker.py — Rapid Serial Visual Presentation reader.

Presents text one word at a time at a fixed focal point. Faster than scroll
because the eye doesn't have to saccade. Adjustable WPM via scroll wheel or
arrow keys. Consumes text from stdin or a file argument.

Usage:
  wl-paste | rsvp-ticker.py --wpm 400
  rsvp-ticker.py --wpm 250 path/to/file.txt
  rsvp-ticker.py --clipboard

Keyboard controls (while running):
  Space       : pause / resume
  ArrowUp     : +25 WPM
  ArrowDown   : -25 WPM
  ArrowRight  : skip forward 10 words
  ArrowLeft   : skip back 10 words
  Escape / q  : quit
"""

from __future__ import annotations

import argparse
import os
import subprocess
import sys

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


class RsvpWindow(Gtk.ApplicationWindow):
    def __init__(self, app, words: list[str], wpm: int, monitor_name: str):
        super().__init__(application=app)
        self.words = words
        self.idx = 0
        self.wpm = max(60, min(1200, wpm))
        self.paused = False
        self._timer_id: int | None = None

        tr.setup_layer_shell(
            self,
            (Gtk4LayerShell.Edge.TOP, Gtk4LayerShell.Edge.LEFT, Gtk4LayerShell.Edge.RIGHT),
            "hg-rsvp",
            monitor_name,
            layer="OVERLAY",
            exclusive_zone=0,
        )
        self.da = tr.make_drawing_area(BAR_H, self._draw)
        self.set_child(self.da)

        key = Gtk.EventControllerKey.new()
        key.connect("key-pressed", self._on_key)
        self.add_controller(key)
        scroll = Gtk.EventControllerScroll.new(Gtk.EventControllerScrollFlags.VERTICAL)
        scroll.connect("scroll", self._on_scroll)
        self.add_controller(scroll)

        self.present()
        self._reschedule()

    def _reschedule(self) -> None:
        if self._timer_id is not None:
            GLib.source_remove(self._timer_id)
            self._timer_id = None
        if self.paused or self.idx >= len(self.words):
            return
        interval_ms = max(30, int(60_000 / self.wpm))
        # Slow down 60% on punctuation (.,;:) to let the brain catch up
        if self.idx > 0 and self.words[self.idx - 1].rstrip("\"')]").endswith(tuple(".,;:!?")):
            interval_ms = int(interval_ms * 1.6)
        self._timer_id = GLib.timeout_add(interval_ms, self._advance)
        self.da.queue_draw()

    def _advance(self) -> bool:
        self.idx += 1
        if self.idx >= len(self.words):
            self.da.queue_draw()
            return False
        self._reschedule()
        self.da.queue_draw()
        return False

    def _on_key(self, ctl, keyval, keycode, state):
        kv = keyval
        if kv in (Gdk.KEY_Escape, Gdk.KEY_q):
            self.get_application().quit()
            return True
        if kv == Gdk.KEY_space:
            self.paused = not self.paused
            self._reschedule()
            return True
        if kv == Gdk.KEY_Up:
            self.wpm = min(1200, self.wpm + 25)
            self._reschedule()
            return True
        if kv == Gdk.KEY_Down:
            self.wpm = max(60, self.wpm - 25)
            self._reschedule()
            return True
        if kv == Gdk.KEY_Right:
            self.idx = min(len(self.words), self.idx + 10)
            self._reschedule()
            return True
        if kv == Gdk.KEY_Left:
            self.idx = max(0, self.idx - 10)
            self._reschedule()
            return True
        return False

    def _on_scroll(self, ctl, dx, dy):
        if dy < 0:
            self.wpm = min(1200, self.wpm + 25)
        elif dy > 0:
            self.wpm = max(60, self.wpm - 25)
        self._reschedule()
        return True

    def _draw(self, widget, cr, w, h):
        tr.fill_bg(cr, w, h, alpha=0.9)
        if self.idx >= len(self.words):
            word = "\u2713 done (Esc to close)"
            color = "#3dffb5"
        else:
            word = self.words[self.idx]
            color = "#29f0ff"
        layout = PangoCairo.create_layout(cr)
        layout.set_font_description(Pango.FontDescription("Maple Mono NF CN Bold 13"))
        layout.set_markup(
            f'<span foreground="{color}">{GLib.markup_escape_text(word)}</span>',
            -1,
        )
        tw, th = layout.get_pixel_size()
        cr.move_to((w - tw) / 2, max(0, (h - th) / 2))
        PangoCairo.show_layout(cr, layout)
        # Progress bar
        if self.words:
            pct = self.idx / len(self.words)
            cr.set_source_rgba(0.24, 1.0, 0.71, 0.7)
            cr.rectangle(0, h - 1, int(w * pct), 1)
            cr.fill()
        # WPM + pause indicator at top-right
        status = f"{self.wpm} WPM" + (" PAUSED" if self.paused else "")
        layout2 = PangoCairo.create_layout(cr)
        layout2.set_font_description(Pango.FontDescription("Maple Mono NF CN 9"))
        layout2.set_markup(
            f'<span foreground="#66708f">{status}</span>', -1,
        )
        sw, sh = layout2.get_pixel_size()
        cr.move_to(w - sw - 8, 4)
        PangoCairo.show_layout(cr, layout2)


class RsvpApp(Gtk.Application):
    def __init__(self, words: list[str], wpm: int, monitor: str):
        super().__init__(
            application_id="io.hairglasses.rsvp",
            flags=Gio.ApplicationFlags.NON_UNIQUE,
        )
        self.words = words
        self.wpm = wpm
        self.monitor = monitor

    def do_activate(self):
        RsvpWindow(self, self.words, self.wpm, self.monitor)


def _collect_words(args) -> list[str]:
    if args.clipboard:
        text = subprocess.run(["wl-paste", "--no-newline"], capture_output=True, text=True).stdout
    elif args.file:
        with open(args.file, "r") as f:
            text = f.read()
    elif not sys.stdin.isatty():
        text = sys.stdin.read()
    else:
        text = "Usage: pipe text, pass a file, or use --clipboard."
    return [w for w in text.split() if w]


def main() -> int:
    ap = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    ap.add_argument("file", nargs="?", help="Path to a text file.")
    ap.add_argument("--clipboard", action="store_true", help="Read from wl-paste.")
    ap.add_argument("--wpm", type=int, default=300)
    ap.add_argument("--monitor", default="DP-2")
    args = ap.parse_args()

    words = _collect_words(args)
    if not words:
        print("no text to present", file=sys.stderr)
        return 1
    app = RsvpApp(words, args.wpm, args.monitor)
    return app.run([sys.argv[0]])


if __name__ == "__main__":
    sys.exit(main())
