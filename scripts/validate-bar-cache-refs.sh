#!/usr/bin/env bash
set -euo pipefail

# validate-bar-cache-refs.sh — guard the bar cache producer/consumer
# graph under /tmp/bar-*.txt.
#
# Bar/ticker cache files follow a three-layer convention:
#   - writer: scripts/bar-<name>-cache.sh (optionally scheduled by
#     systemd/bar-<name>.{service,timer})
#   - consumer: ironbar/config.toml (widget label) OR
#     ticker/streams.toml cache entry OR
#     scripts/lib/ticker_streams/<name>.py plugin
#
# A new widget pointing at /tmp/bar-new.txt with no writer renders as
# blank; a new cache file with no reader is dead work each refresh
# interval. Both are silent until someone notices the gap.
#
# Exit 0 if every /tmp/bar-*.txt read has a writer. Unused producers
# (write-only) surface as warnings but don't fail the gate — internal
# rowset caches like bar-prs-rows are legitimately write-read by the
# same script.

ROOT="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
cd "$ROOT"

python3 - <<'PY'
import re
import sys
from pathlib import Path

BAR_RE_READ = re.compile(r"/tmp/bar-([a-zA-Z0-9_-]+)\.txt")
BAR_RE_WRITE = re.compile(r"/tmp/bar-([a-zA-Z0-9_-]+)\.(?:txt|XXXXXX)")

consumers: dict[str, list[str]] = {}
# Scan config + ticker plugin surface for reads
consumer_sources = []
consumer_sources.extend(Path(".").rglob("*.toml"))
consumer_sources.extend(Path("scripts/lib/ticker_streams").rglob("*.py"))
consumer_sources.extend(Path(".").rglob("*.conf"))

for f in consumer_sources:
    if ".git" in f.parts or "node_modules" in f.parts:
        continue
    if not f.is_file():
        continue
    try:
        content = f.read_text(errors="ignore")
    except (OSError, UnicodeDecodeError):
        continue
    for name in BAR_RE_READ.findall(content):
        consumers.setdefault(name, []).append(str(f))

producers: dict[str, list[str]] = {}
for f in list(Path("scripts").rglob("*")) + list(Path("systemd").rglob("*")):
    if not f.is_file():
        continue
    try:
        content = f.read_text(errors="ignore")
    except (OSError, UnicodeDecodeError):
        continue
    for name in BAR_RE_WRITE.findall(content):
        producers.setdefault(name, []).append(str(f))

orphans = {k: v for k, v in consumers.items() if k not in producers}
unused = {k: v for k, v in producers.items() if k not in consumers}

if orphans:
    print(f"consumers={len(consumers)} producers={len(producers)} "
          f"orphan_consumers={len(orphans)}")
    for name, where in sorted(orphans.items()):
        print(f"  ORPHAN /tmp/bar-{name}.txt — read by:")
        for loc in where:
            print(f"    {loc}")
    sys.exit(1)

unused_note = ""
if unused:
    unused_note = f" unused_producers={len(unused)}"

print(
    f"consumers={len(consumers)} producers={len(producers)}{unused_note} errors=0"
)
PY
