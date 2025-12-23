#!/usr/bin/env bash
set -euo pipefail
here="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
src="$here/diagrams"
out="$src/_rendered"
mkdir -p "$out"

render_one() {
  local infile="$1"
  local base="$(basename "$infile" .mmd)"
  local outfile="_rendered/${base}.svg"
  if command -v docker >/dev/null 2>&1; then
    docker run --rm -u "$(id -u):$(id -g)" -v "$src:/work" minlag/mermaid-cli           -i "$(basename "$infile")" -o "$outfile"
  elif command -v podman >/dev/null 2>&1; then
    podman run --rm -v "$src:/work" docker.io/minlag/mermaid-cli           -i "$(basename "$infile")" -o "$outfile"
  elif command -v mmdc >/dev/null 2>&1; then
    mmdc -i "$infile" -o "$src/$outfile"
  else
    echo "No renderer found (docker/podman/mmdc). Aborting." >&2
    exit 2
  fi
}

shopt -s nullglob
for f in "$src"/*.mmd; do
  echo "Rendering $(basename "$f") ..."
  render_one "$f"
done
echo "Rendered SVGs to $out"
