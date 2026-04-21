#!/usr/bin/env python3
"""Run the RetroArch Archive.org manifest/fetch/import/playlist/audit flow as one command."""

from __future__ import annotations

import argparse
import json
import os
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path


SCRIPT_DIR = Path(__file__).resolve().parent
STATE_HOME = Path.home() / ".local" / "state"


def _expand(path: str | None, default: Path) -> Path:
    if not path:
        return default
    return Path(os.path.expandvars(os.path.expanduser(path)))


def _tool_path(env_name: str, default_name: str) -> Path:
    override = os.environ.get(env_name)
    if override:
        return Path(os.path.expandvars(os.path.expanduser(override)))
    return SCRIPT_DIR / default_name


def _build_python_step(
    script: Path,
    *,
    args: list[str],
) -> list[str]:
    return [sys.executable, str(script), *args]


def _append_repeatable(argv: list[str], flag: str, values: list[str]) -> None:
    for value in values:
        argv.extend([flag, value])


def _run_step(name: str, argv: list[str]) -> dict[str, object]:
    completed = subprocess.run(argv, capture_output=True, text=True, check=False)
    return {
        "name": name,
        "argv": argv,
        "returncode": completed.returncode,
        "stdout": completed.stdout,
        "stderr": completed.stderr,
        "ok": completed.returncode == 0,
    }


def _load_json(path: Path) -> dict[str, object] | None:
    if not path.exists():
        return None
    try:
        return json.loads(path.read_text())
    except json.JSONDecodeError:
        return None


def _notify_runtime(message: str, audit_report: dict[str, object], timeout_seconds: float) -> dict[str, object] | None:
    runtime = audit_report.get("runtime")
    if not isinstance(runtime, dict):
        return None
    if runtime.get("process_running") is not True or runtime.get("network_cmd_enable") is not True:
        return None

    port = runtime.get("network_cmd_port")
    port_number = int(port) if isinstance(port, (int, float)) else 55355

    sys.path.insert(0, str(SCRIPT_DIR / "lib"))
    import retroarch_runtime  # type: ignore

    return retroarch_runtime.send_udp_command(
        f"SHOW_MSG {message}",
        port=port_number,
        timeout_seconds=timeout_seconds,
    )


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(
        description="Run the Archive.org homebrew/public-domain RetroArch sync flow end-to-end."
    )
    parser.add_argument("--manifest", help="Path to the generated manifest JSON.")
    parser.add_argument("--download-root", help="Root directory for mirrored Archive.org files.")
    parser.add_argument("--rom-root", help="RetroArch ROM root directory.")
    parser.add_argument("--playlist-root", help="RetroArch playlist directory.")
    parser.add_argument("--config-dir", help="RetroArch config directory used by the audit step.")
    parser.add_argument("--system-dir", help="RetroArch system/BIOS directory used by the audit step.")
    parser.add_argument("--retroarch-profile", help="Optional path to the shared RomHub RetroArch profile YAML.")
    parser.add_argument("--subdir", default="archive-homebrew", help="ROM subdirectory for imported Archive content.")
    parser.add_argument(
        "--tier",
        action="append",
        choices=["public_domain", "verified_redistributable", "homebrew_unverified", "utility_unverified"],
        help="Include one or more manifest tiers. Default includes public_domain and verified_redistributable.",
    )
    parser.add_argument("--system", action="append", help="Limit work to one or more RetroArch system ids.")
    parser.add_argument(
        "--transport",
        choices=["auto", "direct", "ia"],
        default="auto",
        help="Archive download transport for the fetch step.",
    )
    parser.add_argument("--import-mode", choices=["symlink", "copy"], default="symlink")
    parser.add_argument("--notify-runtime", action="store_true", help="Show an on-screen RetroArch message after a successful sync.")
    parser.add_argument(
        "--runtime-message",
        default="Archive homebrew sync complete",
        help="OSD message sent through RetroArch network commands when --notify-runtime is enabled.",
    )
    parser.add_argument(
        "--runtime-timeout-seconds",
        type=float,
        default=1.0,
        help="Timeout for RetroArch UDP runtime notification.",
    )
    parser.add_argument("--skip-manifest", action="store_true", help="Skip manifest regeneration.")
    parser.add_argument("--skip-fetch", action="store_true", help="Skip archive download fetches.")
    parser.add_argument("--skip-import", action="store_true", help="Skip staging mirrored files into RetroArch ROM dirs.")
    parser.add_argument("--skip-playlists", action="store_true", help="Skip playlist generation.")
    parser.add_argument("--skip-audit", action="store_true", help="Skip the final workstation audit.")
    parser.add_argument("--dry-run", action="store_true", help="Preview the flow without changing downloads, ROM staging, or playlists.")
    parser.add_argument("--output", help="Summary JSON output path.")
    args = parser.parse_args(argv)

    manifest_path = _expand(
        args.manifest or os.environ.get("RETROARCH_ARCHIVE_HOMEBREW_MANIFEST"),
        STATE_HOME / "retroarch-archive" / "homebrew-manifest.json",
    )
    audit_output_path = _expand(
        os.environ.get("RETROARCH_WORKSTATION_AUDIT_PATH"),
        STATE_HOME / "retroarch" / "workstation-audit.json",
    )
    summary_path = _expand(
        args.output or os.environ.get("RETROARCH_ARCHIVE_SYNC_SUMMARY"),
        STATE_HOME / "retroarch-archive" / "sync-summary.json",
    )

    manifest_tool = _tool_path("RETROARCH_ARCHIVE_MANIFEST_TOOL", "retroarch-archive-homebrew-manifest.py")
    fetch_tool = _tool_path("RETROARCH_ARCHIVE_FETCH_TOOL", "retroarch-archive-homebrew-fetch.py")
    import_tool = _tool_path("RETROARCH_ARCHIVE_IMPORT_TOOL", "retroarch-archive-homebrew-import.py")
    playlist_tool = _tool_path("RETROARCH_ARCHIVE_PLAYLIST_TOOL", "retroarch-archive-homebrew-playlists.py")
    audit_tool = _tool_path("RETROARCH_WORKSTATION_AUDIT_TOOL", "retroarch-workstation-audit.py")

    steps: list[dict[str, object]] = []

    if not args.skip_manifest:
        manifest_argv = _build_python_step(
            manifest_tool,
            args=["--output", str(manifest_path)],
        )
        _append_repeatable(manifest_argv, "--system", args.system or [])
        steps.append(_run_step("manifest", manifest_argv))
        if steps[-1]["ok"] is not True:
            return _finish(summary_path, manifest_path, audit_output_path, steps, None, None)

    tier_values = args.tier or []

    if not args.skip_fetch:
        fetch_argv = _build_python_step(
            fetch_tool,
            args=[
                "--manifest",
                str(manifest_path),
                "--transport",
                args.transport,
            ],
        )
        if args.download_root:
            fetch_argv.extend(["--download-root", args.download_root])
        _append_repeatable(fetch_argv, "--tier", tier_values)
        _append_repeatable(fetch_argv, "--system", args.system or [])
        if args.dry_run:
            fetch_argv.append("--dry-run")
        steps.append(_run_step("fetch", fetch_argv))
        if steps[-1]["ok"] is not True:
            return _finish(summary_path, manifest_path, audit_output_path, steps, None, None)

    if not args.skip_import:
        import_argv = _build_python_step(
            import_tool,
            args=[
                "--manifest",
                str(manifest_path),
                "--mode",
                args.import_mode,
                "--subdir",
                args.subdir,
            ],
        )
        if args.download_root:
            import_argv.extend(["--archive-root", args.download_root])
        if args.rom_root:
            import_argv.extend(["--rom-root", args.rom_root])
        _append_repeatable(import_argv, "--tier", tier_values)
        _append_repeatable(import_argv, "--system", args.system or [])
        if args.dry_run:
            import_argv.append("--dry-run")
        steps.append(_run_step("import", import_argv))
        if steps[-1]["ok"] is not True:
            return _finish(summary_path, manifest_path, audit_output_path, steps, None, None)

    if not args.skip_playlists:
        playlist_argv = _build_python_step(
            playlist_tool,
            args=[
                "--manifest",
                str(manifest_path),
                "--subdir",
                args.subdir,
            ],
        )
        if args.rom_root:
            playlist_argv.extend(["--rom-root", args.rom_root])
        if args.playlist_root:
            playlist_argv.extend(["--playlist-root", args.playlist_root])
        if args.retroarch_profile:
            playlist_argv.extend(["--retroarch-profile", args.retroarch_profile])
        _append_repeatable(playlist_argv, "--tier", tier_values)
        _append_repeatable(playlist_argv, "--system", args.system or [])
        if args.dry_run:
            playlist_argv.append("--dry-run")
        steps.append(_run_step("playlists", playlist_argv))
        if steps[-1]["ok"] is not True:
            return _finish(summary_path, manifest_path, audit_output_path, steps, None, None)

    audit_report = None
    if not args.skip_audit:
        audit_argv = _build_python_step(
            audit_tool,
            args=[
                "--output",
                str(audit_output_path),
            ],
        )
        if args.config_dir:
            audit_argv.extend(["--config-dir", args.config_dir])
        if args.system_dir:
            audit_argv.extend(["--system-dir", args.system_dir])
        if args.rom_root:
            audit_argv.extend(["--roms-dir", args.rom_root])
        if args.playlist_root:
            audit_argv.extend(["--playlist-dir", args.playlist_root])
        if args.retroarch_profile:
            audit_argv.extend(["--retroarch-profile", args.retroarch_profile])
        steps.append(_run_step("audit", audit_argv))
        if steps[-1]["ok"] is True:
            audit_report = _load_json(audit_output_path)
        else:
            return _finish(summary_path, manifest_path, audit_output_path, steps, None, None)

    runtime_notice = None
    if args.notify_runtime and not args.dry_run and audit_report:
        runtime_notice = _notify_runtime(args.runtime_message, audit_report, args.runtime_timeout_seconds)

    return _finish(summary_path, manifest_path, audit_output_path, steps, audit_report, runtime_notice)


def _finish(
    summary_path: Path,
    manifest_path: Path,
    audit_output_path: Path,
    steps: list[dict[str, object]],
    audit_report: dict[str, object] | None,
    runtime_notice: dict[str, object] | None,
) -> int:
    ok = all(step.get("ok") is True for step in steps)
    payload = {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "ok": ok,
        "paths": {
            "manifest": str(manifest_path),
            "audit_report": str(audit_output_path),
            "summary": str(summary_path),
        },
        "steps": steps,
        "audit": audit_report,
        "runtime_notice": runtime_notice,
    }
    summary_path.parent.mkdir(parents=True, exist_ok=True)
    summary_path.write_text(json.dumps(payload, indent=2) + "\n")
    print(summary_path)
    print(
        "steps={steps} failed={failed} runtime_notice={runtime_notice}".format(
            steps=len(steps),
            failed=sum(1 for step in steps if step.get("ok") is not True),
            runtime_notice="sent" if runtime_notice and runtime_notice.get("ok") else "none",
        )
    )
    return 0 if ok else 1


if __name__ == "__main__":
    raise SystemExit(main())
