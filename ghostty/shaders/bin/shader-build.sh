#!/usr/bin/env bash
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
      continue
    fi

    # Check for #include directive
    if [[ "$line" =~ ^'// #include "'(.+)'"' ]]; then
      local lib_ref="${BASH_REMATCH[1]}"
      local lib_file="$LIB_DIR/${lib_ref#lib/}"

      # Write the directive
      echo "$line" >> "$tmp"

      if $STRIP_MODE; then
        : # Just keep the directive, no inline
      elif [[ -f "$lib_file" ]]; then
        echo "$BEGIN_MARKER $lib_ref ---" >> "$tmp"
        cat "$lib_file" >> "$tmp"
        echo "$END_MARKER $lib_ref ---" >> "$tmp"
        changed=true
      else
        echo "WARNING: $lib_ref not found for $shader_name" >&2
      fi
      continue
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

exit $exit_code
