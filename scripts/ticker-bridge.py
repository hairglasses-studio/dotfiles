#!/usr/bin/env python3
"""ticker-bridge.py — NDJSON bridge for Quickshell ticker migration.

Runs one existing keybind-ticker stream builder and emits line-delimited JSON
that QML can consume through Quickshell.Io.Process + SplitParser.
"""

from __future__ import annotations

import argparse
import html
import json
import re
import sys
import time
import traceback
from pathlib import Path


HERE = Path(__file__).resolve().parent
DOTFILES = HERE.parent
LIB = HERE / "lib"
PLUGIN_DIR = LIB / "ticker_streams"
DEFAULT_CATALOG = DOTFILES / "ticker" / "streams.toml"

sys.path.insert(0, str(LIB))
sys.path.insert(0, str(PLUGIN_DIR))


def _load_streams(catalog: Path):
    import ticker_streams as ts

    builders: dict[str, object] = {}
    meta: dict[str, dict] = {}
    order: list[str] = []

    toml_builders, toml_meta, toml_order = ts.load_toml_streams(str(catalog))
    builders.update(toml_builders)
    meta.update(toml_meta)
    order.extend(toml_order)

    plugin_builders, plugin_meta, plugin_order, _slow = ts.load_plugin_streams(str(PLUGIN_DIR))
    builders.update(plugin_builders)
    meta.update(plugin_meta)
    for name in plugin_order:
        if name not in order:
            order.append(name)

    return builders, meta, order


def _emit(payload: dict) -> None:
    print(json.dumps(payload, ensure_ascii=True, separators=(",", ":")), flush=True)


def _plain_text(markup: str) -> str:
    """Convert ticker Pango markup into compact display text for non-Pango UIs."""
    text = re.sub(r"<[^>]+>", "", markup)
    return re.sub(r"\s+", " ", html.unescape(text)).strip()


def _build_payload(stream: str, builder, meta: dict) -> dict:
    try:
        markup, segments = builder()
        markup = str(markup)
        return {
            "ok": True,
            "stream": stream,
            "ts": int(time.time()),
            "refresh": int(meta.get("refresh", 300)),
            "preset": meta.get("preset"),
            "markup": markup,
            "text": _plain_text(markup),
            "segments": list(segments or []),
        }
    except Exception as exc:  # pragma: no cover - exercised by runtime failures
        print(traceback.format_exc(), file=sys.stderr)
        return {
            "ok": False,
            "stream": stream,
            "ts": int(time.time()),
            "refresh": int(meta.get("refresh", 30)),
            "preset": meta.get("preset"),
            "markup": "",
            "segments": [],
            "error": str(exc),
        }


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--stream", help="Stream name to emit")
    parser.add_argument("--catalog", type=Path, default=DEFAULT_CATALOG)
    parser.add_argument("--interval", type=float, default=0, help="Override refresh interval in seconds")
    parser.add_argument("--once", action="store_true", help="Emit one payload and exit")
    parser.add_argument("--watch", action="store_true", help="Continuously emit payloads at the stream refresh interval")
    parser.add_argument("--list", action="store_true", help="List available stream names as JSON and exit")
    args = parser.parse_args()

    builders, meta, order = _load_streams(args.catalog)

    if args.list:
        _emit({"ok": True, "streams": order})
        return 0

    if not args.stream:
        parser.error("--stream is required unless --list is set")

    builder = builders.get(args.stream)
    if builder is None:
        _emit({"ok": False, "stream": args.stream, "error": "unknown stream"})
        return 2

    stream_meta = meta.get(args.stream, {})
    interval = args.interval if args.interval > 0 else int(stream_meta.get("refresh", 300))
    interval = max(1.0, float(interval))

    while True:
        _emit(_build_payload(args.stream, builder, stream_meta))
        if args.once:
            return 0
        time.sleep(interval)


if __name__ == "__main__":
    raise SystemExit(main())
