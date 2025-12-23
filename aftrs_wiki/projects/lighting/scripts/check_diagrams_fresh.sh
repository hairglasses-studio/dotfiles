#!/usr/bin/env bash
# Renders Mermaid diagrams and returns:
#  - 0 if outputs are up to date (no working tree changes in diagrams/_rendered)
#  - 3 if renderer missing
#  - 4 if outputs changed (stale), prints a message
set -euo pipefail
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
SRC="$REPO_ROOT/diagrams"
OUT="$SRC/_rendered"
mkdir -p "$OUT"

# Renderer detection
renderer=""
if command -v docker >/dev/null 2>&1; then
  renderer="docker"
elif command -v podman >/dev/null 2>&1; then
  renderer="podman"
elif command -v mmdc >/dev/null 2>&1; then
  renderer="mmdc"
fi

if [[ -z "$renderer" ]]; then
  echo "[check] No Mermaid renderer found (docker/podman/mmdc)."
  exit 3
fi

# Render all .mmd (cheap; small repo), then check if anything changed
if [[ "$renderer" == "docker" ]]; then
  for f in "$SRC"/*.mmd; do
    base="$(basename "$f" .mmd)"
    docker run --rm -u "$(id -u):$(id -g)" -v "$SRC:/work" minlag/mermaid-cli           -i "$(basename "$f")" -o "_rendered/${base}.svg"
  done
elif [[ "$renderer" == "podman" ]]; then
  for f in "$SRC"/*.mmd; do
    base="$(basename "$f" .mmd)"
    podman run --rm -v "$SRC:/work" docker.io/minlag/mermaid-cli           -i "$(basename "$f")" -o "_rendered/${base}.svg"
  done
else
  for f in "$SRC"/*.mmd; do
    base="$(basename "$f" .mmd)"
    mmdc -i "$f" -o "$OUT/${base}.svg"
  done
fi

# See if working tree now has diffs for _rendered
if ! git diff --quiet -- "$OUT"; then
  echo "[check] Mermaid outputs changed. Commit updated SVGs from: $OUT"
  exit 4
fi

echo "[check] Mermaid outputs are fresh."
