#!/usr/bin/env python3
"""Render a review report for Archive.org homebrew/public-domain candidates."""

from __future__ import annotations

import argparse
import json
import os
from collections import Counter, defaultdict
from pathlib import Path


def _expand(path: str | None, default: Path) -> Path:
    if not path:
        return default
    return Path(os.path.expandvars(os.path.expanduser(path)))


def _human_size(raw: str | int | None) -> str:
    if raw in (None, ""):
        return "-"
    size = int(raw)
    units = ["B", "KB", "MB", "GB"]
    value = float(size)
    for unit in units:
        if value < 1024.0 or unit == units[-1]:
            if unit == "B":
                return f"{int(value)} {unit}"
            return f"{value:.1f} {unit}"
        value /= 1024.0
    return f"{size} B"


def _markdown_escape(text: str) -> str:
    return text.replace("|", r"\|").replace("\n", " ")


def _filter_entries(manifest: dict, allowed_tiers: set[str], allowed_systems: set[str]) -> list[dict]:
    filtered = []
    for entry in manifest.get("entries", []):
        if entry["tier"] not in allowed_tiers:
            continue
        if allowed_systems and entry["system"] not in allowed_systems:
            continue
        filtered.append(entry)
    return filtered


def _render_markdown(entries: list[dict], manifest: dict, limit_per_system: int) -> str:
    reason_counts: Counter[str] = Counter()
    by_system: dict[str, list[dict]] = defaultdict(list)
    for entry in entries:
        for reason in entry.get("reasons", []):
            reason_counts[reason] += 1
        by_system[entry["system"]].append(entry)

    lines = [
        "# Archive Homebrew Review",
        "",
        f"Generated from manifest: `{manifest.get('generated_at', '-')}`",
        f"Filtered candidates: `{len(entries)}`",
        "",
        "## System Summary",
        "",
        "| System | Candidates |",
        "| --- | ---: |",
    ]
    for system in sorted(by_system):
        lines.append(f"| `{system}` | {len(by_system[system])} |")

    lines.extend(
        [
            "",
            "## Reason Summary",
            "",
            "| Reason | Count |",
            "| --- | ---: |",
        ]
    )
    for reason, count in reason_counts.most_common():
        lines.append(f"| {_markdown_escape(reason)} | {count} |")

    for system in sorted(by_system):
        lines.extend(
            [
                "",
                f"## {system}",
                "",
                "| File | Size | Reason | Source |",
                "| --- | ---: | --- | --- |",
            ]
        )
        for entry in sorted(by_system[system], key=lambda item: item["file_name"].lower())[:limit_per_system]:
            reason = "; ".join(entry.get("reasons", []))
            source = f"`{entry['source_identifier']}:{entry['archive_path']}`"
            lines.append(
                "| "
                f"`{_markdown_escape(entry['file_name'])}` | "
                f"{_human_size(entry.get('size'))} | "
                f"{_markdown_escape(reason)} | "
                f"{source} |"
            )
        remaining = len(by_system[system]) - min(len(by_system[system]), limit_per_system)
        if remaining > 0:
            lines.append(f"\n`... {remaining} more entries omitted for {system}`")

    return "\n".join(lines) + "\n"


def _render_tsv(entries: list[dict]) -> str:
    lines = ["system\tfile_name\ttier\tsize\treasons\tsource_identifier\tarchive_path"]
    for entry in sorted(entries, key=lambda item: (item["system"], item["file_name"].lower())):
        lines.append(
            "\t".join(
                [
                    entry["system"],
                    entry["file_name"],
                    entry["tier"],
                    str(entry.get("size") or ""),
                    "; ".join(entry.get("reasons", [])),
                    entry["source_identifier"],
                    entry["archive_path"],
                ]
            )
        )
    return "\n".join(lines) + "\n"


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Render a review report for Archive.org homebrew candidates.")
    parser.add_argument("--manifest", help="Path to the manifest JSON.")
    parser.add_argument(
        "--tier",
        action="append",
        choices=["public_domain", "verified_redistributable", "homebrew_unverified", "utility_unverified", "excluded"],
        help="Filter to one or more tiers. Default is homebrew_unverified.",
    )
    parser.add_argument("--system", action="append", help="Limit review output to specific system ids.")
    parser.add_argument(
        "--format",
        choices=["markdown", "json", "tsv"],
        default="markdown",
        help="Report format.",
    )
    parser.add_argument(
        "--limit-per-system",
        type=int,
        default=50,
        help="Maximum entries rendered per system for markdown output.",
    )
    parser.add_argument("--output", help="Path to write the report.")
    args = parser.parse_args(argv)

    manifest_path = _expand(
        args.manifest or os.environ.get("RETROARCH_ARCHIVE_HOMEBREW_MANIFEST"),
        Path.home() / ".local" / "state" / "retroarch-archive" / "homebrew-manifest.json",
    )
    manifest = json.loads(manifest_path.read_text())
    allowed_tiers = set(args.tier or ["homebrew_unverified"])
    allowed_systems = set(args.system or [])
    entries = _filter_entries(manifest, allowed_tiers, allowed_systems)

    default_suffix = {
        "markdown": "md",
        "json": "json",
        "tsv": "tsv",
    }[args.format]
    output_path = _expand(
        args.output or os.environ.get("RETROARCH_ARCHIVE_HOMEBREW_REVIEW"),
        Path.home() / ".local" / "state" / "retroarch-archive" / f"homebrew-review.{default_suffix}",
    )
    output_path.parent.mkdir(parents=True, exist_ok=True)

    if args.format == "json":
        payload = {
            "generated_at": manifest.get("generated_at"),
            "entry_count": len(entries),
            "entries": entries,
        }
        output_path.write_text(json.dumps(payload, indent=2))
    elif args.format == "tsv":
        output_path.write_text(_render_tsv(entries))
    else:
        output_path.write_text(_render_markdown(entries, manifest, args.limit_per_system))

    counts = Counter(entry["system"] for entry in entries)
    summary = " | ".join(f"{system}:{count}" for system, count in sorted(counts.items())) if counts else "no-matching-entries"
    print(output_path)
    print(summary)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
