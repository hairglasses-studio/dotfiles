#!/usr/bin/env bash
set -euo pipefail
# ── shader-build — inline shared GLSL libraries into shaders ──
# Ghostty has no #include support, so each .glsl must be self-contained.
# This script inlines lib/ files referenced by `// #include "lib/X.glsl"`
# directives, making them idempotent (re-running replaces existing inlines).
#
# Usage:
#   shader-build [shader.glsl ...]    # build specific shaders
#   shader-build --all                # build all shaders with #include directives
#   shader-build --check              # dry-run: show what would change
#   shader-build --strip <shader>     # remove inlined code, leave directives

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SHADERS_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
LIB_DIR="$SHADERS_DIR/lib"
source "$SCRIPT_DIR/../../../scripts/lib/notify.sh"

CHECK_ONLY=false
STRIP_MODE=false

# ── Parse args ────────────────────────────────────
FILES=()
while [[ $# -gt 0 ]]; do
  case "$1" in
    --all)
      while IFS= read -r f; do
        FILES+=("$f")
      done < <(grep -rl '// #include "lib/' "$SHADERS_DIR"/*.glsl 2>/dev/null || true)
      shift ;;
    --check)  CHECK_ONLY=true; shift ;;
    --strip)  STRIP_MODE=true; shift ;;
    -h|--help)
      echo "Usage: shader-build [--all | --check | --strip] [shader.glsl ...]"
      echo ""
      echo "Inlines lib/ files referenced by // #include \"lib/X.glsl\" directives."
      echo "Each .glsl remains valid standalone — build just keeps shared code in sync."
      echo ""
      echo "Options:"
      echo "  --all     Process all shaders with #include directives"
      echo "  --check   Dry-run: show what would change (exit 1 if changes needed)"
      echo "  --strip   Remove inlined code, leaving only the directives"
      exit 0 ;;
    *.glsl)
      if [[ -f "$SHADERS_DIR/$1" ]]; then
        FILES+=("$SHADERS_DIR/$1")
      elif [[ -f "$1" ]]; then
        FILES+=("$1")
      else
        echo "Not found: $1" >&2; exit 1
      fi
      shift ;;
    *) echo "Unknown option: $1" >&2; exit 1 ;;
  esac
done

if [[ ${#FILES[@]} -eq 0 ]]; then
  echo "No files to process. Use --all or specify .glsl files." >&2
  exit 0
fi

# ── Markers ───────────────────────────────────────
BEGIN_MARKER="// --- BEGIN"   # e.g. // --- BEGIN lib/hash.glsl ---
END_MARKER="// --- END"       # e.g. // --- END lib/hash.glsl ---

# ── Process one shader ────────────────────────────
process_shader() {
  local shader_path="$1"
  local shader_name
  shader_name="$(basename "$shader_path")"
  local changed=false
  local tmp
  tmp="$(/usr/bin/mktemp "${shader_path}.XXXXXX")"

  local in_block=false
  local current_lib=""

  while IFS= read -r line || [[ -n "$line" ]]; do
    # Check for end of inlined block
    if $in_block; then
      if [[ "$line" == "${END_MARKER}"* ]]; then
        in_block=false
        if ! $STRIP_MODE; then
          # Re-inline the library
          local lib_file="$LIB_DIR/${current_lib#lib/}"
          if [[ -f "$lib_file" ]]; then
            echo "$BEGIN_MARKER $current_lib ---" >> "$tmp"
            cat "$lib_file" >> "$tmp"
            echo "$END_MARKER $current_lib ---" >> "$tmp"
          else
            echo "WARNING: $current_lib not found for $shader_name" >&2
            echo "$line" >> "$tmp"
          fi
        fi
        # In strip mode, the block (including markers) is dropped
      fi
      # Skip all lines inside the block (they'll be replaced)
      continue
    fi

    # Check for begin marker (existing inlined block)
    if [[ "$line" == "${BEGIN_MARKER}"* ]]; then
      current_lib="${line#${BEGIN_MARKER} }"
      current_lib="${current_lib% ---}"
      in_block=true
      changed=true
      include_just_seen=""
      continue
    fi

    # Check for #include directive
    if [[ "$line" =~ ^'// #include "'(.+)'"' ]]; then
      local lib_ref="${BASH_REMATCH[1]}"
      local lib_file="$LIB_DIR/${lib_ref#lib/}"

      # Write the directive
      echo "$line" >> "$tmp"

      # If an existing inlined block follows, the BEGIN handler will re-inline it.
      # Only inline here if this is a fresh directive (no block follows yet).
      # We track this with a flag; the BEGIN handler checks it.
      include_just_seen="$lib_ref"
      continue
    fi

    # If we just saw an #include and the next line is a BEGIN marker for the same lib,
    # let the block handler take over (it will strip + re-inline). Otherwise, inline now.
    if [[ -n "${include_just_seen:-}" ]]; then
      if [[ "$line" == "${BEGIN_MARKER}"* ]]; then
        # Existing block follows — enter block mode to replace it
        current_lib="${line#${BEGIN_MARKER} }"
        current_lib="${current_lib% ---}"
        in_block=true
        changed=true
        include_just_seen=""
        continue
      else
        # No existing block — inline the lib now
        if ! $STRIP_MODE; then
          local ilib_file="$LIB_DIR/${include_just_seen#lib/}"
          if [[ -f "$ilib_file" ]]; then
            echo "$BEGIN_MARKER $include_just_seen ---" >> "$tmp"
            cat "$ilib_file" >> "$tmp"
            echo "$END_MARKER $include_just_seen ---" >> "$tmp"
            changed=true
          else
            echo "WARNING: $include_just_seen not found for $shader_name" >&2
          fi
        fi
        include_just_seen=""
        # Fall through to write the current line normally
      fi
    fi

    echo "$line" >> "$tmp"
  done < "$shader_path"

  if $changed || $STRIP_MODE; then
    if $CHECK_ONLY; then
      if ! diff -q "$shader_path" "$tmp" &>/dev/null; then
        echo "CHANGED: $shader_name"
        diff -u "$shader_path" "$tmp" | head -30
        rm -f "$tmp"
        return 1
      fi
      rm -f "$tmp"
      return 0
    else
      mv "$tmp" "$shader_path"
      echo "Built: $shader_name"
    fi
  else
    rm -f "$tmp"
    if ! $CHECK_ONLY; then
      echo "No change: $shader_name"
    fi
  fi
}

# ── Main ──────────────────────────────────────────
exit_code=0
for f in "${FILES[@]}"; do
  process_shader "$f" || exit_code=1
done

if (( exit_code != 0 )); then
  hg_notify "Shader" "Build failed — check output"
fi

exit $exit_code
