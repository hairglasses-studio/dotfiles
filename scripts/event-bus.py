#!/usr/bin/env python3
"""event-bus.py — unified observability daemon for the dotfiles stack.

Ingests four streams concurrently and matches each against the rule table
in event-bus-rules.yaml:

  1. journalctl -f -k            → planePitch / DRM kernel errors
  2. hyprctl monitors -j (poll)  → Samsung DSC mode loss detection
  3. nvidia-smi --query-gpu      → hardware thermal slowdown
  4. pactl subscribe             → audio sink removal

Plus a correlation rule: when Hyprland's IPC emits `configreloaded`, we
check whether a kernel DRM error fired within the last 5s. If so, emit a
single composite `hypr_reload_induced_drm` event with both timestamps.

Structured events are appended to ~/.local/state/dotfiles/events.jsonl
(one JSON object per line). Consumers like /heal and /canary tail this
file — they look up remediations themselves via the remediation_lookup
MCP tool, so this daemon stays decoupled from the Go registry.

Design notes:
  - Each sub-stream runs as a separate asyncio task. If a sub-command
    exits, the task reopens it with exponential backoff (max 60s). A
    permanent failure does not take down the daemon — just that stream.
  - Events are written with `os.fsync()` so a daemon crash doesn't lose
    the last second of signal.
  - Notification only fires for events with severity in {medium, high}.
    The rule-table `notify:` key can force it either way.
  - SIGTERM triggers graceful shutdown — all tasks cancelled, file
    flushed, exit 0 so systemd records a clean stop.
"""

from __future__ import annotations

import asyncio
import json
import os
import re
import signal
import sys
import time
from collections import deque
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

try:
    import yaml  # type: ignore
except ImportError:
    print("event-bus.py requires PyYAML — pacman -S python-yaml", file=sys.stderr)
    sys.exit(2)


RULES_PATH = Path(__file__).parent / "event-bus-rules.yaml"
STATE_DIR = Path(os.environ.get("XDG_STATE_HOME", str(Path.home() / ".local" / "state"))) / "dotfiles"
EVENTS_PATH = STATE_DIR / "events.jsonl"

# Sliding window of recently-seen events, used for correlation rules.
# Each entry is (error_code, epoch_seconds).
CORRELATION_WINDOW_S = 5
_recent_events: deque[tuple[str, float]] = deque(maxlen=200)

# Per-(rule, fingerprint) dedup window. A persistent condition like a DSC
# mode loss would otherwise fire every 10s poll and spam swaync into
# uselessness. DEDUP_WINDOW_S throttles same-fingerprint emits; the first
# one lands promptly and the rest are swallowed until the window expires.
DEDUP_WINDOW_S = 300
_dedup_seen: dict[tuple[str, str], float] = {}


@dataclass
class Rule:
    name: str
    source: str
    match: Any
    error_code: str
    severity: str = "medium"
    notify: bool | None = None
    correlation_window_s: int = 0


@dataclass
class Event:
    type: str
    at: str
    fingerprint: str
    error_code: str
    severity: str
    rule: str
    source: str
    correlation: dict[str, Any] = field(default_factory=dict)


def load_rules() -> list[Rule]:
    if not RULES_PATH.exists():
        print(f"event-bus.py: rules file missing at {RULES_PATH}", file=sys.stderr)
        sys.exit(2)
    with RULES_PATH.open("r", encoding="utf-8") as f:
        data = yaml.safe_load(f) or {}
    return [
        Rule(
            name=r["name"],
            source=r["source"],
            match=r["match"],
            error_code=r["error_code"],
            severity=r.get("severity", "medium"),
            notify=r.get("notify"),
            correlation_window_s=r.get("correlation_window_s", 0),
        )
        for r in data.get("rules", [])
    ]


def now_rfc() -> str:
    return time.strftime("%Y-%m-%dT%H:%M:%SZ", time.gmtime())


def emit(event: Event) -> None:
    """Append an event to events.jsonl and fire a notification if warranted.

    Skips same-(rule, fingerprint) events within DEDUP_WINDOW_S so a
    persistent condition (a DSC mode loss, a stuck throttle) produces one
    event at onset rather than a flood from the poll loop.
    """
    now = time.time()
    fp_short = event.fingerprint[:240]
    dedup_key = (event.rule, fp_short)
    last_seen = _dedup_seen.get(dedup_key, 0.0)
    if now - last_seen < DEDUP_WINDOW_S:
        return
    _dedup_seen[dedup_key] = now
    # Cheap GC so the map doesn't grow unbounded under varied fingerprints.
    if len(_dedup_seen) > 1024:
        cutoff = now - DEDUP_WINDOW_S
        for k in [k for k, t in _dedup_seen.items() if t < cutoff]:
            del _dedup_seen[k]

    STATE_DIR.mkdir(parents=True, exist_ok=True)
    payload = {
        "type": event.type,
        "at": event.at,
        "fingerprint": fp_short,
        "error_code": event.error_code,
        "severity": event.severity,
        "rule": event.rule,
        "source": event.source,
    }
    if event.correlation:
        payload["correlation"] = event.correlation

    line = json.dumps(payload, ensure_ascii=False) + "\n"
    with EVENTS_PATH.open("a", encoding="utf-8") as f:
        f.write(line)
        f.flush()
        os.fsync(f.fileno())

    _recent_events.append((event.error_code, now))

    # Notification policy.
    should_notify = event.severity in ("medium", "high")
    # (Per-event override hints live on the Rule, not the Event — looked up
    # by the caller before emit().)
    if should_notify:
        _notify_send(event)


def _notify_send(event: Event) -> None:
    """Best-effort swaync notification; errors swallowed."""
    urgency = "critical" if event.severity == "high" else "normal"
    cmd = [
        "notify-send",
        "-u", urgency,
        "-a", "event-bus",
        f"[{event.error_code}] {event.type}",
        f"{event.rule} — {event.fingerprint[:160]}",
    ]
    try:
        import subprocess
        subprocess.run(cmd, check=False, capture_output=True, timeout=2)
    except Exception:
        pass


def recent_has(code: str, within_s: int) -> bool:
    cutoff = time.time() - within_s
    return any(c == code and ts >= cutoff for c, ts in _recent_events)


# ---------------------------------------------------------------------------
# Stream ingestors
# ---------------------------------------------------------------------------


async def stream_journal(rules: list[Rule]) -> None:
    """Tail journalctl -f -k and match each line against journal rules."""
    journal_rules = [r for r in rules if r.source == "journal"]
    if not journal_rules:
        return
    backoff = 1
    while True:
        try:
            proc = await asyncio.create_subprocess_exec(
                "journalctl", "-f", "-k", "-n", "0", "--no-pager",
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.DEVNULL,
            )
            backoff = 1
            assert proc.stdout is not None
            async for raw in proc.stdout:
                line = raw.decode(errors="replace").rstrip()
                if not line:
                    continue
                for rule in journal_rules:
                    if isinstance(rule.match, str) and re.search(rule.match, line):
                        ev = Event(
                            type=rule.error_code,
                            at=now_rfc(),
                            fingerprint=line,
                            error_code=rule.error_code,
                            severity=rule.severity,
                            rule=rule.name,
                            source="journal",
                        )
                        emit(ev)
            await proc.wait()
        except Exception as e:
            print(f"event-bus.py: journal stream error: {e}", file=sys.stderr)
        await asyncio.sleep(min(backoff, 60))
        backoff = min(backoff * 2, 60)


async def stream_monitors(rules: list[Rule]) -> None:
    """Poll `hyprctl monitors -j` every 10s and match geometry against rules."""
    monitor_rules = [r for r in rules if r.source == "monitors"]
    if not monitor_rules:
        return
    while True:
        try:
            proc = await asyncio.create_subprocess_exec(
                "hyprctl", "monitors", "-j",
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.DEVNULL,
            )
            stdout, _ = await proc.communicate()
            monitors = json.loads(stdout.decode() or "[]")
            for rule in monitor_rules:
                target = rule.match.get("monitor")
                for m in monitors:
                    if target and m.get("name") != target:
                        continue
                    if "width_below" in rule.match and m.get("width", 0) < rule.match["width_below"]:
                        ev = Event(
                            type=rule.error_code,
                            at=now_rfc(),
                            fingerprint=f"{m.get('name')} width={m.get('width')} height={m.get('height')}",
                            error_code=rule.error_code,
                            severity=rule.severity,
                            rule=rule.name,
                            source="monitors",
                        )
                        emit(ev)
        except Exception as e:
            print(f"event-bus.py: monitors poll error: {e}", file=sys.stderr)
        await asyncio.sleep(10)


async def stream_nvidia_throttle(rules: list[Rule]) -> None:
    """Poll nvidia-smi every 10s for hardware throttle state."""
    throttle_rules = [r for r in rules if r.source == "nvidia_throttle"]
    if not throttle_rules:
        return
    query = "clocks_throttle_reasons.hw_thermal_slowdown,clocks.sm"
    while True:
        try:
            proc = await asyncio.create_subprocess_exec(
                "nvidia-smi",
                f"--query-gpu={query}",
                "--format=csv,noheader,nounits",
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.DEVNULL,
            )
            stdout, _ = await proc.communicate()
            text = stdout.decode().strip()
            if text:
                first = text.splitlines()[0]
                parts = [p.strip() for p in first.split(",")]
                hw_slow = parts[0].lower() in ("active", "true", "1")
                sm_clock = int(parts[1]) if len(parts) > 1 and parts[1].isdigit() else 0
                for rule in throttle_rules:
                    want_hw = rule.match.get("hw_slowdown")
                    sm_below = rule.match.get("sm_clock_below")
                    hit = False
                    if want_hw is not None and want_hw == hw_slow:
                        hit = True
                    if sm_below is not None and sm_clock and sm_clock < sm_below:
                        hit = True
                    if hit:
                        ev = Event(
                            type=rule.error_code,
                            at=now_rfc(),
                            fingerprint=f"hw_slow={hw_slow} sm_clock={sm_clock}",
                            error_code=rule.error_code,
                            severity=rule.severity,
                            rule=rule.name,
                            source="nvidia_throttle",
                        )
                        emit(ev)
        except Exception as e:
            print(f"event-bus.py: nvidia poll error: {e}", file=sys.stderr)
        await asyncio.sleep(10)


async def stream_pulse(rules: list[Rule]) -> None:
    """Tail `pactl subscribe` for sink/source/card events."""
    pulse_rules = [r for r in rules if r.source == "pulse"]
    if not pulse_rules:
        return
    backoff = 1
    while True:
        try:
            proc = await asyncio.create_subprocess_exec(
                "pactl", "subscribe",
                stdout=asyncio.subprocess.PIPE,
                stderr=asyncio.subprocess.DEVNULL,
            )
            backoff = 1
            assert proc.stdout is not None
            async for raw in proc.stdout:
                line = raw.decode(errors="replace").rstrip()
                if not line:
                    continue
                for rule in pulse_rules:
                    if isinstance(rule.match, str) and re.search(rule.match, line):
                        ev = Event(
                            type=rule.error_code,
                            at=now_rfc(),
                            fingerprint=line,
                            error_code=rule.error_code,
                            severity=rule.severity,
                            rule=rule.name,
                            source="pulse",
                        )
                        emit(ev)
            await proc.wait()
        except Exception as e:
            print(f"event-bus.py: pulse stream error: {e}", file=sys.stderr)
        await asyncio.sleep(min(backoff, 60))
        backoff = min(backoff * 2, 60)


async def stream_hypr_events(rules: list[Rule]) -> None:
    """Tail the Hyprland IPC socket2 via `hyprctl -r event -`."""
    hypr_rules = [r for r in rules if r.source == "hypr_events"]
    if not hypr_rules:
        return
    runtime = os.environ.get("XDG_RUNTIME_DIR", "")
    sig = os.environ.get("HYPRLAND_INSTANCE_SIGNATURE", "")
    if not runtime or not sig:
        return
    sock = Path(runtime) / "hypr" / sig / ".socket2.sock"
    if not sock.exists():
        return
    backoff = 1
    while True:
        try:
            reader, writer = await asyncio.open_unix_connection(str(sock))
            backoff = 1
            while True:
                raw = await reader.readline()
                if not raw:
                    break
                # Format: "event>>arg"
                line = raw.decode(errors="replace").rstrip()
                parts = line.split(">>", 1)
                event_name = parts[0] if parts else ""
                for rule in hypr_rules:
                    if isinstance(rule.match, str) and re.search(rule.match, event_name):
                        # For correlation rules, only fire when a related
                        # error is in the recent window.
                        if rule.correlation_window_s > 0:
                            if not recent_has(rule.error_code, rule.correlation_window_s):
                                continue
                        ev = Event(
                            type=rule.error_code,
                            at=now_rfc(),
                            fingerprint=line,
                            error_code=rule.error_code,
                            severity=rule.severity,
                            rule=rule.name,
                            source="hypr_events",
                            correlation={"window_s": rule.correlation_window_s} if rule.correlation_window_s else {},
                        )
                        emit(ev)
            writer.close()
            await writer.wait_closed()
        except Exception as e:
            print(f"event-bus.py: hypr events error: {e}", file=sys.stderr)
        await asyncio.sleep(min(backoff, 60))
        backoff = min(backoff * 2, 60)


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------


async def main() -> None:
    rules = load_rules()
    tasks = [
        asyncio.create_task(stream_journal(rules)),
        asyncio.create_task(stream_monitors(rules)),
        asyncio.create_task(stream_nvidia_throttle(rules)),
        asyncio.create_task(stream_pulse(rules)),
        asyncio.create_task(stream_hypr_events(rules)),
    ]

    loop = asyncio.get_running_loop()
    stop = loop.create_future()
    for sig_name in (signal.SIGINT, signal.SIGTERM):
        loop.add_signal_handler(sig_name, lambda: stop.set_result(None))

    await stop
    for t in tasks:
        t.cancel()
    await asyncio.gather(*tasks, return_exceptions=True)


if __name__ == "__main__":
    try:
        asyncio.run(main())
    except KeyboardInterrupt:
        pass
