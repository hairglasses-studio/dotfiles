#!/usr/bin/env python3
"""lyrics-ticker.py — now-playing banner with time-synced lyrics.

v2 behaviour: poll ``playerctl`` for MPRIS metadata; on track change,
fetch synced lyrics from LRCLib (https://lrclib.net — free, no auth).
Parse the standard [mm:ss.cc] LRC format into (seconds, line) tuples
and render the active line centered, advancing via a 250ms timer that
binary-searches the timestamp list against the current player
position. Falls back to the v1 "title — artist" banner when LRCLib
has no match for the track.

Lyrics are cached per-track at ~/.cache/lyrics-ticker/<sha1>.json so
repeat plays and offline reconnects don't re-hit the API.

Controls: nothing interactive — this is a passive display. Kill the
service to hide it.
"""

from __future__ import annotations

import argparse
import bisect
import hashlib
import json
import os
import re
import subprocess
import sys
import threading
import time
import urllib.parse
import urllib.request

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
LRCLIB_URL = "https://lrclib.net/api/get"
CACHE_DIR = os.path.expanduser("~/.cache/lyrics-ticker")
LRC_RE = re.compile(r"\[(\d+):(\d+(?:\.\d+)?)\](.*)")


def _playerctl(fmt: str) -> str:
    try:
        return subprocess.run(
            ["playerctl", "metadata", "--format", fmt],
            capture_output=True, text=True, timeout=1,
        ).stdout.strip()
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


def _track_key(title: str, artist: str, album: str) -> str:
    raw = f"{artist}|{album}|{title}".encode("utf-8", "replace")
    return hashlib.sha1(raw).hexdigest()


def _parse_lrc(text: str) -> list[tuple[float, str]]:
    """Parse a standard LRC blob into (seconds, line) tuples, sorted by
    timestamp. Skips metadata tags and empty lines."""
    out: list[tuple[float, str]] = []
    for raw_line in text.splitlines():
        line = raw_line.strip()
        if not line:
            continue
        # A single source line may have multiple [mm:ss] stamps (shared line);
        # capture all of them.
        stamps: list[float] = []
        rest = line
        while True:
            m = LRC_RE.match(rest)
            if not m:
                break
            mm, ss, rest = m.group(1), m.group(2), m.group(3)
            try:
                stamps.append(int(mm) * 60 + float(ss))
            except ValueError:
                pass
        if not stamps:
            continue
        body = rest.strip()
        for t in stamps:
            out.append((t, body))
    out.sort(key=lambda x: x[0])
    return out


def _fetch_lrclib(title: str, artist: str, album: str, duration_s: int) -> dict | None:
    """Fetch LRCLib entry. Returns parsed JSON or None on any failure."""
    if not title or not artist:
        return None
    params = {
        "track_name": title,
        "artist_name": artist,
        "album_name": album,
        "duration": str(max(0, int(duration_s))),
    }
    url = f"{LRCLIB_URL}?{urllib.parse.urlencode(params)}"
    req = urllib.request.Request(
        url,
        headers={"User-Agent": "lyrics-ticker/2 (+https://github.com/hairglasses-studio)"},
    )
    try:
        with urllib.request.urlopen(req, timeout=4) as r:
            if r.status != 200:
                return None
            return json.loads(r.read().decode("utf-8", "replace"))
    except Exception:
        return None


def _cached_lookup(key: str) -> dict | None:
    path = os.path.join(CACHE_DIR, f"{key}.json")
    try:
        with open(path) as f:
            return json.load(f)
    except (OSError, ValueError):
        return None


def _cache_write(key: str, data: dict) -> None:
    try:
        os.makedirs(CACHE_DIR, exist_ok=True)
        path = os.path.join(CACHE_DIR, f"{key}.json")
        with open(path, "w") as f:
            json.dump(data, f)
    except OSError:
        pass


class LyricsWindow(Gtk.ApplicationWindow):
    def __init__(self, app, monitor_name: str):
        super().__init__(application=app)
        self.line = "\u266a IDLE"
        self.color = "#66708f"
        self._track_key: str | None = None
        self._lrc: list[tuple[float, str]] = []
        self._fallback_line = ""
        self._last_idx = -1

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

        # Two timers:
        #   - every 1.5s: check track change & status
        #   - every 250ms: advance the active LRC line against position
        GLib.timeout_add(1500, self._poll_track)
        GLib.timeout_add(250, self._advance_lrc)
        self._poll_track()

    def _poll_track(self) -> bool:
        status = _playerctl("{{status}}")
        if status != "Playing":
            self.line = "\u266a IDLE"
            self.color = "#66708f"
            self._track_key = None
            self._lrc = []
            self.da.queue_draw()
            return True
        title = _playerctl("{{title}}")
        artist = _playerctl("{{artist}}")
        album = _playerctl("{{album}}")
        try:
            dur_s = int(_playerctl("{{mpris:length}}") or "0") // 1_000_000
        except ValueError:
            dur_s = 0
        key = _track_key(title, artist, album)
        if key != self._track_key:
            # New track — switch to fallback immediately, kick off async fetch.
            self._track_key = key
            self._lrc = []
            self._fallback_line = f"\u266a {title} — {artist}"
            self.line = self._fallback_line
            self.color = "#ff47d1"
            self._last_idx = -1
            threading.Thread(
                target=self._fetch_and_apply,
                args=(key, title, artist, album, dur_s),
                daemon=True,
            ).start()
        self.da.queue_draw()
        return True

    def _fetch_and_apply(self, key: str, title: str, artist: str,
                          album: str, dur_s: int) -> None:
        data = _cached_lookup(key)
        if data is None:
            data = _fetch_lrclib(title, artist, album, dur_s) or {}
            if data:
                _cache_write(key, data)
        synced = (data.get("syncedLyrics") or "") if isinstance(data, dict) else ""
        lrc = _parse_lrc(synced) if synced else []
        # Marshal back to the main loop.
        GLib.idle_add(self._set_lrc, key, lrc)

    def _set_lrc(self, key: str, lrc: list[tuple[float, str]]) -> bool:
        if key == self._track_key:
            self._lrc = lrc
            self._last_idx = -1
        return False

    def _advance_lrc(self) -> bool:
        if not self._lrc or self._track_key is None:
            return True
        try:
            pos_us = int(_playerctl("{{position}}") or "0")
        except ValueError:
            pos_us = 0
        pos_s = pos_us / 1_000_000
        # Binary-search the active line: largest timestamp <= pos_s.
        idx = bisect.bisect_right([t for t, _ in self._lrc], pos_s) - 1
        if idx < 0:
            idx = 0
        if idx != self._last_idx:
            self._last_idx = idx
            self.line = self._lrc[idx][1] or self._fallback_line
            self.color = "#3dffb5"
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
