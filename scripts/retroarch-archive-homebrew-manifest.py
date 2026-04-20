#!/usr/bin/env python3
"""Build a conservative Internet Archive homebrew/public-domain manifest."""

from __future__ import annotations

import argparse
import json
import os
import re
import urllib.parse
import urllib.request
from collections import Counter, defaultdict
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


ARCHIVE_METADATA_URL = "https://archive.org/metadata/{identifier}"
ARCHIVE_DOWNLOAD_URL = "https://archive.org/download/{identifier}/{path}"


@dataclass(frozen=True)
class SourceItem:
    identifier: str
    title: str
    kind: str
    system_dirs: dict[str, str]


SOURCES: list[SourceItem] = [
    SourceItem(
        identifier="homebrew-romsets",
        title="Homebrew Romsets Collection",
        kind="romset",
        system_dirs={
            "Dreamcast": "dreamcast",
            "Nintendo 64": "n64",
            "Super Nintendo": "snes",
            "Genesis": "genesis",
        },
    ),
    SourceItem(
        identifier="homebrew-channels",
        title="The Homebrew Channel Custom Forwarders",
        kind="utility",
        system_dirs={
            "": "wii",
        },
    ),
]


CONTENT_EXTENSIONS = {
    ".zip",
    ".7z",
    ".rar",
    ".cdi",
    ".chd",
    ".cue",
    ".gdi",
    ".iso",
    ".wbfs",
    ".rvz",
    ".wad",
    ".z64",
    ".v64",
    ".n64",
    ".sfc",
    ".smc",
    ".fig",
    ".bin",
    ".md",
    ".gen",
}


PUBLIC_DOMAIN_PATTERNS = [
    r"\(pd\)",
    r"\bpublic domain\b",
]

EXCLUDED_PATTERNS = [
    r"\bhack\b",
    r"\btranslation\b",
    r"\bprototype\b",
    r"\bproto\b",
    r"\bpirate\b",
    r"\bdemo\b",
    r"\bbeta\b",
    r"\bsample\b",
    r"\btest\b",
]

FRANCHISE_RISK_PATTERNS = [
    r"\bmario\b",
    r"\bzelda\b",
    r"\bpokemon\b",
    r"\bmetroid\b",
    r"\bkirby\b",
    r"\bsonic\b",
    r"\bkong\b",
    r"\bflappy bird\b",
    r"\bpringles\b",
]


def _expand(path: str | None, default: Path) -> Path:
    if not path:
        return default
    return Path(os.path.expandvars(os.path.expanduser(path)))


def _fetch_metadata(identifier: str) -> dict[str, Any]:
    url = ARCHIVE_METADATA_URL.format(identifier=identifier)
    with urllib.request.urlopen(url, timeout=30) as response:
        return json.load(response)


def _top_level_dir(name: str) -> str:
    return name.split("/", 1)[0] if "/" in name else ""


def _is_content_file(name: str) -> bool:
    lower = name.lower()
    return Path(lower).suffix in CONTENT_EXTENSIONS


def _matches_any(patterns: list[str], text: str) -> bool:
    return any(re.search(pattern, text, re.IGNORECASE) for pattern in patterns)


def _classify_entry(source: SourceItem, system_id: str, name: str) -> tuple[str, list[str]]:
    lowered = name.lower()
    reasons: list[str] = []

    if _matches_any(EXCLUDED_PATTERNS, lowered):
        reasons.append("filename matches excluded pattern")
        return "excluded", reasons

    if source.kind == "utility":
        reasons.append("utility/homebrew support item, redistribution rights not auto-verified")
        return "utility_unverified", reasons

    if _matches_any(PUBLIC_DOMAIN_PATTERNS, lowered):
        if _matches_any(FRANCHISE_RISK_PATTERNS, lowered):
            reasons.append("filename says PD, but title appears to use an obvious protected game franchise")
            return "homebrew_unverified", reasons
        reasons.append("filename explicitly marked PD/public domain")
        return "public_domain", reasons

    if "homebrew" in lowered:
        reasons.append("filename marked homebrew but license not auto-verified")
        return "homebrew_unverified", reasons

    if "aftermarket" in lowered or "(unl)" in lowered or "unl)." in lowered:
        reasons.append("aftermarket/unlicensed marker without explicit public-domain grant")
        return "homebrew_unverified", reasons

    if system_id in {"dreamcast", "n64", "snes", "genesis"}:
        reasons.append("sourced from a homebrew collection, but no explicit public-domain marker")
        return "homebrew_unverified", reasons

    reasons.append("no explicit safe redistribution signal found")
    return "excluded", reasons


def build_manifest(selected_systems: set[str] | None = None) -> dict[str, Any]:
    entries: list[dict[str, Any]] = []
    summary_by_system: dict[str, Counter] = defaultdict(Counter)
    source_summaries: list[dict[str, Any]] = []

    for source in SOURCES:
        metadata = _fetch_metadata(source.identifier)
        files = metadata.get("files", [])
        source_count = 0
        for file_info in files:
            name = file_info.get("name", "")
            if not name or not _is_content_file(name):
                continue

            top = _top_level_dir(name)
            system_id = source.system_dirs.get(top)
            if system_id is None and "" in source.system_dirs and "/" not in name:
                system_id = source.system_dirs[""]
            if system_id is None:
                continue
            if selected_systems and system_id not in selected_systems:
                continue

            tier, reasons = _classify_entry(source, system_id, name)
            rel_path = name
            entry = {
                "system": system_id,
                "source_identifier": source.identifier,
                "source_title": source.title,
                "archive_path": rel_path,
                "file_name": Path(rel_path).name,
                "tier": tier,
                "default_selected": tier == "public_domain",
                "reasons": reasons,
                "download_url": ARCHIVE_DOWNLOAD_URL.format(
                    identifier=source.identifier,
                    path=urllib.parse.quote(rel_path, safe="/()[]!'+,.-_ "),
                ).replace(" ", "%20"),
                "size": file_info.get("size"),
            }
            entries.append(entry)
            source_count += 1
            summary_by_system[system_id][tier] += 1
        source_summaries.append(
            {
                "identifier": source.identifier,
                "title": source.title,
                "kind": source.kind,
                "candidate_count": source_count,
            }
        )

    summary = {
        system: {
            "count": sum(counter.values()),
            "public_domain_count": counter["public_domain"],
            "homebrew_unverified_count": counter["homebrew_unverified"],
            "utility_unverified_count": counter["utility_unverified"],
            "excluded_count": counter["excluded"],
        }
        for system, counter in sorted(summary_by_system.items())
    }

    return {
        "generated_at": datetime.now(timezone.utc).isoformat(),
        "sources": source_summaries,
        "summary": summary,
        "entries": entries,
    }


def main(argv: list[str] | None = None) -> int:
    parser = argparse.ArgumentParser(description="Build a conservative Archive.org homebrew/public-domain manifest.")
    parser.add_argument(
        "--system",
        action="append",
        choices=["dreamcast", "wii", "n64", "snes", "genesis"],
        help="Limit manifest generation to specific system ids.",
    )
    parser.add_argument(
        "--output",
        help="Write manifest JSON to a custom path.",
    )
    args = parser.parse_args(argv)

    selected_systems = set(args.system) if args.system else None
    manifest = build_manifest(selected_systems)
    output_path = _expand(
        args.output or os.environ.get("RETROARCH_ARCHIVE_HOMEBREW_MANIFEST"),
        Path.home() / ".local" / "state" / "retroarch-archive" / "homebrew-manifest.json",
    )
    output_path.parent.mkdir(parents=True, exist_ok=True)
    output_path.write_text(json.dumps(manifest, indent=2))

    summary_bits = []
    for system, counts in manifest["summary"].items():
        summary_bits.append(
            f"{system}:count={counts['count']},pd={counts['public_domain_count']},"
            f"unverified={counts['homebrew_unverified_count']},utility={counts['utility_unverified_count']}"
        )
    print(output_path)
    if summary_bits:
        print(" | ".join(summary_bits))
    else:
        print("no-matching-systems")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
