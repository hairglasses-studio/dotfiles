#!/usr/bin/env bash
# validate-publishing-assets.sh -- public publishing asset consistency gate.
#
# Catches drift in README-linked media, publishing drafts, and MCP directory
# submission copy before a launch checklist is reused.

set -euo pipefail

ROOT_DIR="${1:-$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)}"
errors=0

fail() {
    printf 'DRIFT %s\n' "$*" >&2
    errors=$((errors + 1))
}

require_file() {
    local rel="$1"
    if [[ ! -f "$ROOT_DIR/$rel" ]]; then
        fail "missing required file: $rel"
        return 1
    fi
    return 0
}

require_file README.md || true
require_file docs/assets/ticker-demo.gif || true
require_file docs/publishing/hyprland-mcp-agent-blog.md || true
require_file docs/publishing/mcp-directories.md || true
require_file mcp/dotfiles-mcp/snapshots/contract/overview.json || true

if [[ -f "$ROOT_DIR/README.md" ]]; then
    if ! grep -Fq 'docs/assets/ticker-demo.gif' "$ROOT_DIR/README.md"; then
        fail "README.md does not reference docs/assets/ticker-demo.gif"
    fi
fi

if [[ -f "$ROOT_DIR/docs/assets/ticker-demo.gif" ]]; then
    header="$(LC_ALL=C head -c 6 "$ROOT_DIR/docs/assets/ticker-demo.gif" || true)"
    case "$header" in
        GIF87a|GIF89a) ;;
        *) fail "docs/assets/ticker-demo.gif does not have a GIF header" ;;
    esac

    size_bytes="$(wc -c < "$ROOT_DIR/docs/assets/ticker-demo.gif")"
    max_bytes=$((12 * 1024 * 1024))
    if (( size_bytes == 0 || size_bytes > max_bytes )); then
        fail "docs/assets/ticker-demo.gif size ${size_bytes} outside 1..${max_bytes} bytes"
    fi
fi

python3 - "$ROOT_DIR" <<'PY' || errors=$((errors + 1))
import json
import re
import sys
from pathlib import Path

root = Path(sys.argv[1])
failures: list[str] = []

def read_text(rel: str) -> str:
    path = root / rel
    try:
        return path.read_text(encoding="utf-8")
    except FileNotFoundError:
        failures.append(f"missing required file: {rel}")
        return ""

overview_path = root / "mcp/dotfiles-mcp/snapshots/contract/overview.json"
try:
    overview = json.loads(overview_path.read_text(encoding="utf-8"))
except FileNotFoundError:
    failures.append("missing required file: mcp/dotfiles-mcp/snapshots/contract/overview.json")
    overview = {}

directories = read_text("docs/publishing/mcp-directories.md")
blog = read_text("docs/publishing/hyprland-mcp-agent-blog.md")
module_readme = read_text("mcp/dotfiles-mcp/README.md")

expected_counts = {
    "Tool count": overview.get("total_tools"),
    "Module count": overview.get("module_count"),
    "Resources": overview.get("resource_count"),
    "Prompts": overview.get("prompt_count"),
}

for label, expected in expected_counts.items():
    if expected is None:
        failures.append(f"overview.json missing {label}")
        continue
    pattern = rf"\|\s*{re.escape(label)}\s*\|\s*{re.escape(str(expected))}\s*\|"
    if not re.search(pattern, directories):
        failures.append(f"mcp-directories.md {label} is not {expected}")

version = overview.get("version")
if version and not re.search(rf"\|\s*Version\s*\|\s*{re.escape(str(version))}\s*\|", directories):
    failures.append(f"mcp-directories.md Version is not {version}")

if "https://github.com/hairglasses-studio/dotfiles-mcp" not in directories:
    failures.append("mcp-directories.md missing standalone dotfiles-mcp repo URL")
if "github.com/hairglasses-studio/dotfiles-mcp/cmd/dotfiles-mcp@latest" not in directories:
    failures.append("mcp-directories.md missing standalone go install command")
if "v1.1.0" not in directories or "server contract version" not in directories:
    failures.append("mcp-directories.md missing verified standalone module release note")
if "https://github.com/hairglasses-studio/dotfiles-mcp/releases/tag/v1.1.0" not in directories:
    failures.append("mcp-directories.md missing standalone GitHub Release URL")
if "status=in_sync" not in directories:
    failures.append("mcp-directories.md missing cleared standalone projection status")
if "glama.json" not in directories:
    failures.append("mcp-directories.md missing committed Glama metadata note")
if "Official MCP Registry" not in directories or "supported package channel" not in directories:
    failures.append("mcp-directories.md missing official registry package-channel blocker")
if "browser submission" not in directories:
    failures.append("mcp-directories.md missing browser submission status")
if "raw.githubusercontent.com/hairglasses-studio/dotfiles/main/mcp/dotfiles-mcp/.well-known/mcp.json" not in directories:
    failures.append("mcp-directories.md missing externally crawlable well-known raw URL")

if not re.search(r"^---\n.*?^---\n", blog, re.M | re.S):
    failures.append("blog draft missing YAML frontmatter")
if not re.search(r"^published:\s*false\s*$", blog, re.M):
    failures.append("blog draft must remain published: false")
if not re.search(r"^tags:\s*\[.*hyprland.*mcp.*\]\s*$", blog, re.M | re.I):
    failures.append("blog draft tags should include hyprland and mcp")
if "## What is MCP?" not in module_readme:
    failures.append("mcp/dotfiles-mcp/README.md missing What is MCP? section")

leak_pattern = re.compile(
    r"(/home/hg|mixellburk|mitch@|gh[pousr]_[A-Za-z0-9_]{20,}|"
    r"sk-[A-Za-z0-9_-]{20,}|AKIA[0-9A-Z]{16}|"
    r"(?:API_KEY|SECRET|ACCESS_TOKEN)\s*[=:])",
    re.I,
)
for rel, text in {
    "docs/publishing/mcp-directories.md": directories,
    "docs/publishing/hyprland-mcp-agent-blog.md": blog,
}.items():
    match = leak_pattern.search(text)
    if match:
        failures.append(f"{rel} contains likely private material near {match.group(1)!r}")

if failures:
    for failure in failures:
        print(f"DRIFT {failure}", file=sys.stderr)
    sys.exit(1)

print(
    "publishing_assets=ok "
    f"tools={overview.get('total_tools')} "
    f"modules={overview.get('module_count')} "
    f"resources={overview.get('resource_count')} "
    f"prompts={overview.get('prompt_count')}"
)
PY

if (( errors > 0 )); then
    printf 'publishing_assets=fail errors=%d\n' "$errors"
    exit 1
fi

printf 'publishing_assets=ok errors=0\n'
