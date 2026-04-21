#!/usr/bin/env python3
"""Toggle RetroArch network command support in retroarch.cfg with an atomic backup."""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


sys.path.insert(0, os.path.join(os.path.dirname(os.path.realpath(__file__)), "lib"))

import retroarch_cfg_writer
import retroarch_runtime


DEFAULT_PORT = 55355


def _expand(path: str | None, default: Path) -> Path:
    if not path:
        return default
    return Path(os.path.expandvars(os.path.expanduser(path)))


def _retroarch_running() -> bool:
    completed = subprocess.run(["pgrep", "-x", "retroarch"], capture_output=True, check=False)
    return completed.returncode == 0


def _probe_version(port: int) -> dict[str, Any] | None:
    if not _retroarch_running():
        return None
    return retroarch_runtime.send_udp_command("VERSION", port=port, expect_response=True)


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(
        description="Enable or revert RetroArch network command support in retroarch.cfg.",
    )
    parser.add_argument("--config-dir", help="RetroArch config directory (default: ~/.config/retroarch).")
    parser.add_argument("--port", type=int, default=DEFAULT_PORT, help="network_cmd_port value (default 55355).")
    parser.add_argument("--revert", action="store_true", help="Set network_cmd_enable to false (leaves port alone).")
    parser.add_argument("--no-backup", action="store_true", help="Skip the retroarch.cfg.bak.<stamp> copy.")
    parser.add_argument("--dry-run", action="store_true", help="Print the diff summary without writing.")
    parser.add_argument("--output", help="Optional path for a JSON report of the applied action.")
    args = parser.parse_args(argv)

    config_dir = _expand(args.config_dir or os.environ.get("RETROARCH_CONFIG_DIR"), Path.home() / ".config" / "retroarch")
    cfg_path = config_dir / "retroarch.cfg"

    if not cfg_path.is_file():
        print(f"retroarch.cfg not found at {cfg_path}", file=sys.stderr)
        return 2

    if args.revert:
        updates = {"network_cmd_enable": "false"}
    else:
        updates = {
            "network_cmd_enable": "true",
            "network_cmd_port": str(args.port),
        }

    diff = retroarch_cfg_writer.apply_settings(
        cfg_path,
        updates,
        backup=not args.no_backup,
        dry_run=args.dry_run,
    )

    probe = None
    if diff["applied"] and not args.revert:
        probe = _probe_version(args.port)

    report = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "cfg_path": str(cfg_path),
        "requested": updates,
        "dry_run": bool(args.dry_run),
        "revert": bool(args.revert),
        "diff": diff,
        "retroarch_running": _retroarch_running(),
        "version_probe": probe,
    }

    if args.output:
        output_path = _expand(args.output, Path.cwd() / "retroarch-apply-network-cmd.json")
        output_path.parent.mkdir(parents=True, exist_ok=True)
        output_path.write_text(json.dumps(report, indent=2) + "\n")

    status_bits = [
        f"applied={'yes' if diff['applied'] else 'no'}",
        f"set={len(diff['set'])}",
        f"added={len(diff['added'])}",
        f"unchanged={len(diff['unchanged'])}",
    ]
    if diff["backup_path"]:
        status_bits.append(f"backup={diff['backup_path']}")
    if probe is not None:
        status_bits.append(f"probe={'ok' if probe.get('ok') else 'fail'}")
    elif diff["applied"] and not args.revert:
        status_bits.append("probe=retroarch-not-running")
    print(" ".join(status_bits))

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
