#!/usr/bin/env python3
"""Populate RetroArch system/BIOS directories surfaced by retroarch-workstation-audit."""

from __future__ import annotations

import argparse
import json
import os
import shutil
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


sys.path.insert(0, os.path.join(os.path.dirname(os.path.realpath(__file__)), "lib"))

import retroarch_profile


PPSSPP_REPO = "https://github.com/hrydgard/ppsspp.git"
PPSSPP_BRANCH = "master"
PPSSPP_SPARSE_PATH = "assets"


def _expand(value: str | None, default: Path) -> Path:
    if not value:
        return default
    return Path(os.path.expandvars(os.path.expanduser(value)))


def _parse_ra_kv_file(path: Path) -> dict[str, str]:
    data: dict[str, str] = {}
    if not path.exists():
        return data
    for raw_line in path.read_text(errors="ignore").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        value = value.strip()
        if value.startswith('"') and value.endswith('"') and len(value) >= 2:
            value = value[1:-1]
        data[key.strip()] = value
    return data


def _resolve_system_dir(args: argparse.Namespace) -> Path:
    config_dir = _expand(
        args.config_dir or os.environ.get("RETROARCH_CONFIG_DIR"),
        Path.home() / ".config" / "retroarch",
    )
    settings = _parse_ra_kv_file(config_dir / "retroarch.cfg")
    return _expand(
        args.system_dir or os.environ.get("RETROARCH_SYSTEM_DIR") or settings.get("system_directory"),
        config_dir / "system",
    )


def _missing_required_rows(catalog: list[dict[str, Any]], system_dir: Path) -> list[dict[str, Any]]:
    rows: list[dict[str, Any]] = []
    for entry in catalog:
        if entry.get("requirement") != "required":
            continue
        kind = str(entry["kind"])
        relative_path = str(entry["relative_path"])
        absolute = system_dir / relative_path
        missing = False
        if kind == "file":
            missing = not absolute.is_file()
        elif kind == "dir":
            missing = not absolute.is_dir()
        elif kind == "dir_nonempty":
            missing = not (absolute.is_dir() and any(absolute.iterdir()))
        if missing:
            rows.append({**entry, "absolute_path": str(absolute)})
    return rows


def _plan_dreamcast(rows: list[dict[str, Any]], system_dir: Path, source_dir: Path | None) -> list[dict[str, Any]]:
    steps: list[dict[str, Any]] = []
    dc_dir = system_dir / "dc"
    if any(row["relative_path"] == "dc" for row in rows):
        steps.append({
            "system": "dreamcast",
            "kind": "mkdir",
            "target": str(dc_dir),
            "note": "Flycast system subdirectory",
        })
    if source_dir:
        for basename in ("dc_boot.bin", "dc_flash.bin"):
            src = source_dir / basename
            if src.is_file():
                steps.append({
                    "system": "dreamcast",
                    "kind": "copy",
                    "source": str(src),
                    "target": str(dc_dir / basename),
                    "note": f"Dreamcast {basename}",
                })
    return steps


def _plan_psp(rows: list[dict[str, Any]], system_dir: Path, source_dir: Path | None) -> list[dict[str, Any]]:
    if not any(row["relative_path"] == "PPSSPP" for row in rows):
        return []
    target = system_dir / "PPSSPP"
    if source_dir:
        return [{
            "system": "psp",
            "kind": "copy_tree",
            "source": str(source_dir),
            "target": str(target),
            "note": "PPSSPP helper assets (local source)",
        }]
    return [{
        "system": "psp",
        "kind": "sparse_clone",
        "source": f"{PPSSPP_REPO}#{PPSSPP_BRANCH}:{PPSSPP_SPARSE_PATH}",
        "target": str(target),
        "note": "PPSSPP helper assets via sparse clone",
    }]


def _sparse_clone_ppsspp(target: Path) -> None:
    target.parent.mkdir(parents=True, exist_ok=True)
    import tempfile
    tmp = Path(tempfile.mkdtemp(prefix="ppsspp-sparse-"))
    try:
        subprocess.run(["git", "-C", str(tmp), "init", "-q"], check=True)
        subprocess.run(["git", "-C", str(tmp), "remote", "add", "origin", PPSSPP_REPO], check=True)
        subprocess.run(["git", "-C", str(tmp), "config", "core.sparseCheckout", "true"], check=True)
        (tmp / ".git" / "info" / "sparse-checkout").write_text(PPSSPP_SPARSE_PATH + "\n")
        subprocess.run(
            ["git", "-C", str(tmp), "pull", "--depth=1", "origin", PPSSPP_BRANCH],
            check=True,
            stdout=subprocess.DEVNULL,
        )
        if target.exists():
            shutil.rmtree(target)
        shutil.copytree(tmp / PPSSPP_SPARSE_PATH, target)
    finally:
        shutil.rmtree(tmp, ignore_errors=True)


def _apply_step(step: dict[str, Any]) -> dict[str, Any]:
    kind = step["kind"]
    if kind == "mkdir":
        Path(step["target"]).mkdir(parents=True, exist_ok=True)
    elif kind == "copy":
        target = Path(step["target"])
        target.parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(step["source"], target)
    elif kind == "copy_tree":
        target = Path(step["target"])
        if target.exists():
            shutil.rmtree(target)
        shutil.copytree(step["source"], target)
    elif kind == "sparse_clone":
        _sparse_clone_ppsspp(Path(step["target"]))
    else:
        return {"step": step, "ok": False, "error": f"unknown-kind:{kind}"}
    return {"step": step, "ok": True, "error": None}


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Populate missing RetroArch BIOS/helper directories.")
    parser.add_argument(
        "--system",
        action="append",
        choices=["all", "dreamcast", "psp"],
        help="Which system(s) to handle. Default: all.",
    )
    parser.add_argument("--config-dir", help="RetroArch config directory.")
    parser.add_argument("--system-dir", help="RetroArch system/BIOS directory (overrides retroarch.cfg).")
    parser.add_argument("--source-dir", help="Optional local directory supplying Dreamcast BIOS files or PPSSPP assets.")
    parser.add_argument("--retroarch-profile", help="Optional path to the shared RomHub RetroArch profile YAML.")
    parser.add_argument("--dry-run", action="store_true", help="Print planned steps without writing.")
    parser.add_argument(
        "--output",
        help="JSON report destination. Defaults to $XDG_STATE_HOME/retroarch/bios-apply.json.",
    )
    args = parser.parse_args(argv)

    requested = set(args.system or ["all"])
    if "all" in requested:
        requested = {"dreamcast", "psp"}

    source_dir = _expand(args.source_dir, Path()) if args.source_dir else None
    system_dir = _resolve_system_dir(args)
    catalog = retroarch_profile.load_requirement_catalog(args.retroarch_profile)
    missing = _missing_required_rows(catalog, system_dir)

    planned: list[dict[str, Any]] = []
    if "dreamcast" in requested:
        dreamcast_rows = [row for row in missing if row["system"] == "dreamcast"]
        planned.extend(_plan_dreamcast(dreamcast_rows, system_dir, source_dir))
    if "psp" in requested:
        psp_rows = [row for row in missing if row["system"] == "psp"]
        planned.extend(_plan_psp(psp_rows, system_dir, source_dir))

    applied: list[dict[str, Any]] = []
    if not args.dry_run:
        for step in planned:
            applied.append(_apply_step(step))

    report = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "system_dir": str(system_dir),
        "dry_run": bool(args.dry_run),
        "requested_systems": sorted(requested),
        "missing_before": [
            {"system": row["system"], "relative_path": row["relative_path"], "kind": row["kind"]}
            for row in missing
        ],
        "planned_steps": planned,
        "applied_steps": applied,
    }

    output_path = _expand(
        args.output or os.environ.get("RETROARCH_BIOS_APPLY_OUTPUT"),
        Path.home() / ".local" / "state" / "retroarch" / "bios-apply.json",
    )
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(report, indent=2) + "\n")

    ok_count = sum(1 for entry in applied if entry.get("ok"))
    fail_count = len(applied) - ok_count
    print(f"{output_path}")
    print(
        "planned={planned} applied={applied} failed={failed} dry_run={dry_run}".format(
            planned=len(planned),
            applied=ok_count,
            failed=fail_count,
            dry_run="yes" if args.dry_run else "no",
        )
    )
    return 0 if fail_count == 0 else 1


if __name__ == "__main__":
    raise SystemExit(main())
