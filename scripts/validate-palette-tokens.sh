#!/usr/bin/env bash
set -euo pipefail

# validate-palette-tokens.sh — guard that every $THEME_* reference in
# matugen/templates/ resolves against every palette in theme/palettes/.
#
# Catches the common drift where someone adds a new token reference to
# a matugen template (e.g. ${THEME_NEW_ACCENT}) but forgets to add it
# to every palette env. envsubst would render the template with that
# placeholder empty, silently breaking the palette swap.
#
# Exit 0 on clean, 1 if any template references a token missing from
# any palette.

ROOT="$(cd "$(dirname "$(readlink -f "${BASH_SOURCE[0]}")")/.." && pwd)"
cd "$ROOT"

python3 - <<'PY'
import re
import sys
from pathlib import Path

TEMPLATE_DIR = Path("matugen/templates")
PALETTE_DIR = Path("theme/palettes")

if not TEMPLATE_DIR.is_dir():
    print(f"template dir not found: {TEMPLATE_DIR}", file=sys.stderr)
    sys.exit(0)  # not a schema bug — just no matugen yet

if not PALETTE_DIR.is_dir():
    print(f"palette dir not found: {PALETTE_DIR}", file=sys.stderr)
    sys.exit(1)

token_pat = re.compile(r"\$\{?(THEME_[A-Z_]+)\}?")

template_refs: dict[str, set[str]] = {}
for t in sorted(TEMPLATE_DIR.iterdir()):
    if not t.is_file():
        continue
    refs = set(token_pat.findall(t.read_text(errors="ignore")))
    if refs:
        template_refs[t.name] = refs

palette_tokens: dict[str, set[str]] = {}
for p in sorted(PALETTE_DIR.glob("*.env")):
    tokens = set(re.findall(r"^(THEME_[A-Z_]+)", p.read_text(errors="ignore"), re.M))
    palette_tokens[p.stem] = tokens

if not palette_tokens:
    print("no palettes found", file=sys.stderr)
    sys.exit(1)

# Every palette must carry every token any template references.
all_referenced: set[str] = set()
for refs in template_refs.values():
    all_referenced |= refs

errors = 0
for palette, tokens in palette_tokens.items():
    missing = all_referenced - tokens
    if missing:
        print(f"PALETTE {palette}.env missing tokens: {sorted(missing)}")
        errors += 1

# Also surface which template references each missing token, for repair.
if errors:
    print()
    print("Token → template callers:")
    missing_any = set()
    for palette, tokens in palette_tokens.items():
        missing_any |= (all_referenced - tokens)
    for token in sorted(missing_any):
        callers = [t for t, refs in template_refs.items() if token in refs]
        print(f"  {token}: {', '.join(callers)}")
    sys.exit(1)

print(
    f"palettes={len(palette_tokens)} templates={len(template_refs)} "
    f"tokens_referenced={len(all_referenced)} errors=0"
)
PY
