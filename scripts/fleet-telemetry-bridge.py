#!/usr/bin/env python3
"""Emit one Quickshell-friendly fleet telemetry sample as JSON."""

from __future__ import annotations

import argparse
import json
import os
import re
import time
from pathlib import Path
from typing import Any

GPU_TEMP_RE = re.compile(r"(\d+(?:\.\d+)?)\s*(?:deg|[cC]|" + re.escape(chr(176)) + r")")


def state_path() -> Path:
    root = Path(os.environ.get("XDG_STATE_HOME", Path.home() / ".local/state"))
    return root / "dotfiles" / "quickshell" / "fleet-telemetry-state.json"


def read_json(path: Path) -> dict[str, Any]:
    try:
        with path.open() as handle:
            data = json.load(handle)
        return data if isinstance(data, dict) else {}
    except (OSError, ValueError):
        return {}


def write_json(path: Path, payload: dict[str, Any]) -> None:
    try:
        path.parent.mkdir(parents=True, exist_ok=True)
        tmp = path.with_suffix(".tmp")
        with tmp.open("w") as handle:
            json.dump(payload, handle, separators=(",", ":"))
        tmp.replace(path)
    except OSError:
        pass


def read_cpu_idle_total() -> tuple[int, int]:
    try:
        with open("/proc/stat") as handle:
            for line in handle:
                if line.startswith("cpu "):
                    nums = [int(part) for part in line.split()[1:8]]
                    idle = nums[3] + nums[4]
                    return idle, sum(nums)
    except (OSError, ValueError, IndexError):
        pass
    return 0, 0


def read_mem_pct() -> float:
    total = available = 0
    try:
        with open("/proc/meminfo") as handle:
            for line in handle:
                if line.startswith("MemTotal:"):
                    total = int(line.split()[1])
                elif line.startswith("MemAvailable:"):
                    available = int(line.split()[1])
                if total and available:
                    break
    except (OSError, ValueError, IndexError):
        return 0.0
    if not total:
        return 0.0
    return max(0.0, min(100.0, 100.0 * (total - available) / total))


def read_net_bytes() -> int:
    total = 0
    try:
        with open("/proc/net/dev") as handle:
            next(handle)
            next(handle)
            for line in handle:
                iface, rest = line.split(":", 1)
                if iface.strip() == "lo":
                    continue
                cols = rest.split()
                total += int(cols[0]) + int(cols[8])
    except (OSError, ValueError, IndexError, StopIteration):
        return 0
    return total


def read_gpu_temp() -> float:
    path = Path(os.environ.get("FLEET_TELEMETRY_GPU_FILE", "/tmp/bar-gpu.txt"))
    try:
        text = path.read_text(errors="replace")
    except OSError:
        return 0.0
    match = GPU_TEMP_RE.search(text)
    if not match:
        return 0.0
    try:
        return max(0.0, min(120.0, float(match.group(1))))
    except ValueError:
        return 0.0


def fmt_rate(kbps: float) -> str:
    if kbps >= 1024:
        return f"{kbps / 1024:.1f}M"
    return f"{kbps:.0f}K"


def sample() -> dict[str, Any]:
    now = time.monotonic()
    idle, total = read_cpu_idle_total()
    net = read_net_bytes()
    previous = read_json(state_path())

    prev_total = int(previous.get("cpu_total") or 0)
    prev_idle = int(previous.get("cpu_idle") or 0)
    d_total = total - prev_total
    d_idle = idle - prev_idle
    cpu_pct = 0.0
    if d_total > 0 and d_idle >= 0:
        cpu_pct = max(0.0, min(100.0, 100.0 * (1.0 - (d_idle / d_total))))

    prev_net = int(previous.get("net_bytes") or 0)
    prev_time = float(previous.get("time") or now)
    dt = max(0.001, now - prev_time)
    net_kbps = 0.0
    if prev_net > 0 and net >= prev_net:
        net_kbps = (net - prev_net) / 1024.0 / dt

    mem_pct = read_mem_pct()
    gpu_temp = read_gpu_temp()

    write_json(
        state_path(),
        {
            "time": now,
            "cpu_idle": idle,
            "cpu_total": total,
            "net_bytes": net,
        },
    )

    return {
        "ok": True,
        "timestamp": time.time(),
        "cpu_pct": round(cpu_pct, 1),
        "mem_pct": round(mem_pct, 1),
        "net_kbps": round(net_kbps, 1),
        "gpu_temp_c": round(gpu_temp, 1),
        "cpu_text": f"CPU {cpu_pct:.0f}%",
        "mem_text": f"MEM {mem_pct:.0f}%",
        "net_text": f"NET {fmt_rate(net_kbps)}",
        "gpu_text": f"GPU {gpu_temp:.0f}C" if gpu_temp else "GPU --",
    }


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--once", action="store_true", help="emit one sample and exit")
    parser.parse_args()
    print(json.dumps(sample(), separators=(",", ":")))
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
