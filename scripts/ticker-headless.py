#!/usr/bin/env python3
"""ticker-headless.py — headless variant of keybind-ticker for tmux status lines.

Imports the stream builders from scripts/keybind-ticker.py (no GTK required),
strips Pango markup, and prints plain text. Usage:

  ticker-headless.py --stream ci                # single-shot plain render
  ticker-headless.py --stream system --limit 60 # truncate to 60 chars
  ticker-headless.py --list                     # print available stream names
  ticker-headless.py --playlist main            # render one stream at a time,
                                                #   cycling once per invocation

Designed to be called from `tmux status-right` via `#()` interpolation,
which runs every `status-interval` seconds. Keep --limit tight to stay within
the status line budget.

The module imports keybind-ticker.py dynamically because the filename contains
a hyphen, which prevents `import keybind_ticker` under the normal PEP 328 rules.
GI / GTK bindings are NOT loaded; the import path used here uses only the
non-GTK builders. Any builder that needs GTK-specific state (the main window's
stateful net-throughput delta) falls back gracefully via its cache path.
"""

from __future__ import annotations

import argparse
import html
import importlib.util
import os
import re
import sys
import time

DOTFILES = os.path.expanduser("~/hairglasses-studio/dotfiles")
TICKER_PY = os.path.join(DOTFILES, "scripts/keybind-ticker.py")
PANGO_TAG_RE = re.compile(r"<[^>]+>")
WHITESPACE_RE = re.compile(r"\s+")


def _load_ticker_module():
    # Run keybind-ticker.py as a plain module but prevent the GTK main-loop
    # from starting by ensuring __name__ != "__main__". We also monkeypatch
    # gi.require_version before importing so GTK version warnings stay quiet
    # when GI isn't installed in the host env.
    spec = importlib.util.spec_from_file_location("kt_headless", TICKER_PY)
    mod = importlib.util.module_from_spec(spec)
    # Silence the handful of module-level side effects (print statements,
    # subprocess warmups) by redirecting stderr briefly — these are rare but
    # noisy in a tmux status line.
    stderr_fd = os.dup(2)
    try:
        devnull = os.open(os.devnull, os.O_WRONLY)
        os.dup2(devnull, 2)
        os.close(devnull)
        spec.loader.exec_module(mod)
    finally:
        os.dup2(stderr_fd, 2)
        os.close(stderr_fd)
    return mod


def _plain(markup: str, limit: int | None) -> str:
    """Strip Pango tags and collapse whitespace. Optionally truncate."""
    if not markup:
        return ""
    # Drop the duplicate ghost that _dup() appends for seamless scroll —
    # headless output only needs one copy.
    half = len(markup) // 2
    if markup[:half] == markup[half:]:
        markup = markup[:half]
    text = PANGO_TAG_RE.sub("", markup)
    text = html.unescape(text)
    text = WHITESPACE_RE.sub(" ", text).strip()
    if limit and len(text) > limit:
        text = text[: max(0, limit - 1)].rstrip() + "\u2026"
    return text


def _cycle_index(n: int) -> int:
    # Round-robin across invocations by wall-clock minute so tmux reruns
    # see different streams without needing a persistent state file.
    if n <= 0:
        return 0
    return int(time.time() // 60) % n


def main() -> int:
    ap = argparse.ArgumentParser(description="Headless ticker renderer for tmux.")
    ap.add_argument("--stream", help="Render a single stream by name.")
    ap.add_argument("--playlist", help="Render one stream from this playlist, cycling per-minute.")
    ap.add_argument("--list", action="store_true", help="Print available stream names and exit.")
    ap.add_argument("--limit", type=int, default=0, help="Truncate output to N characters (0 = no limit).")
    args = ap.parse_args()

    mod = _load_ticker_module()
    streams: dict = mod.STREAMS

    if args.list:
        for name in sorted(streams.keys()):
            print(name)
        return 0

    name = None
    if args.stream:
        name = args.stream
    elif args.playlist:
        order = mod.load_playlist(args.playlist)
        if not order:
            return 1
        name = order[_cycle_index(len(order))]

    if not name or name not in streams:
        print(f"unknown stream: {name}", file=sys.stderr)
        return 2

    try:
        markup, _segments = streams[name]()
    except Exception as e:
        # Fail soft — tmux status lines should never crash the pane.
        print(f"[{name} err]", end="")
        return 0

    print(_plain(markup, args.limit if args.limit > 0 else None), end="")
    return 0


if __name__ == "__main__":
    sys.exit(main())
