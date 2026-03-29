#!/usr/bin/env bash
# ── shader-meta — query the shader manifest ──────
# Reads shaders.toml and provides metadata lookups.
# No external dependencies (pure bash/awk).
#
# Usage:
#   shader-meta get bloom-soft description
#   shader-meta list [--category CRT] [--cost LOW]
#   shader-meta fzf-lines
#   shader-meta generate-playlists
#   shader-meta generate-sources-md
#   shader-meta validate

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SHADERS_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
MANIFEST="$SHADERS_DIR/shaders.toml"
PLAYLISTS_DIR="$SHADERS_DIR/playlists"

# ── TOML parser (flat sections only) ─────────────
# Extracts field from a [shaders."name"] section.
_toml_get() {
  local name="$1" field="$2"
  awk -v name="$name" -v field="$field" '
    BEGIN { in_section = 0 }
    /^\[shaders\."/ {
      # Extract section name
      gsub(/^\[shaders\."/, "")
      gsub(/"\].*/, "")
      in_section = ($0 == name) ? 1 : 0
      next
    }
    in_section && $0 ~ "^" field " *= *" {
      val = $0
      sub(/^[^=]*= */, "", val)
      # Strip quotes
      gsub(/^"/, "", val)
      gsub(/"$/, "", val)
      print val
      exit
    }
  ' "$MANIFEST"
}

# Extract array field (playlists = ["a", "b"])
_toml_get_array() {
  local name="$1" field="$2"
  awk -v name="$name" -v field="$field" '
    BEGIN { in_section = 0 }
    /^\[shaders\."/ {
      gsub(/^\[shaders\."/, "")
      gsub(/"\].*/, "")
      in_section = ($0 == name) ? 1 : 0
      next
    }
    in_section && $0 ~ "^" field " *= *" {
      val = $0
      sub(/^[^=]*= *\[/, "", val)
      gsub(/\].*/, "", val)
      gsub(/"/, "", val)
      gsub(/, */, "\n", val)
      print val
      exit
    }
  ' "$MANIFEST"
}

# List all shader names in manifest
_toml_list_names() {
  awk '/^\[shaders\."/ {
    name = $0
    gsub(/^\[shaders\."/, "", name)
    gsub(/"\].*/, "", name)
    print name
  }' "$MANIFEST"
}

# Dump all entries as tab-separated: name\tcategory\tcost\tsource\tdescription
_toml_dump_all() {
  awk '
    /^\[shaders\."/ {
      if (name != "") print name "\t" cat "\t" cost "\t" src "\t" desc
      name = $0
      gsub(/^\[shaders\."/, "", name)
      gsub(/"\].*/, "", name)
      cat = ""; cost = ""; src = ""; desc = ""
    }
    /^category *= */ { val=$0; sub(/^[^=]*= *"/, "", val); gsub(/"$/, "", val); cat=val }
    /^cost *= */     { val=$0; sub(/^[^=]*= *"/, "", val); gsub(/"$/, "", val); cost=val }
    /^source *= */   { val=$0; sub(/^[^=]*= *"/, "", val); gsub(/"$/, "", val); src=val }
    /^description *= */ { val=$0; sub(/^[^=]*= *"/, "", val); gsub(/"$/, "", val); desc=val }
    END { if (name != "") print name "\t" cat "\t" cost "\t" src "\t" desc }
  ' "$MANIFEST"
}

# ── Commands ─────────────────────────────────────

cmd_get() {
  local name="${1%.glsl}" field="$2"
  case "$field" in
    playlists) _toml_get_array "$name" "$field" ;;
    *)         _toml_get "$name" "$field" ;;
  esac
}

cmd_list() {
  local filter_cat="" filter_cost=""
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --category) filter_cat="$2"; shift 2 ;;
      --cost)     filter_cost="$2"; shift 2 ;;
      *)          echo "Unknown option: $1" >&2; return 1 ;;
    esac
  done

  _toml_dump_all | while IFS=$'\t' read -r name cat cost src desc; do
    [[ -n "$filter_cat" && "$cat" != "$filter_cat" ]] && continue
    [[ -n "$filter_cost" && "$cost" != "$filter_cost" ]] && continue
    echo "${name}.glsl"
  done
}

cmd_fzf_lines() {
  # Output: "name.glsl\tCategory\tDescription" for fzf --with-nth
  _toml_dump_all | while IFS=$'\t' read -r name cat cost src desc; do
    printf '%s.glsl\t%-12s\t%s\n' "$name" "$cat" "$desc"
  done | sort -t$'\t' -k2,2 -k1,1
}

cmd_generate_playlists() {
  local tmpdir
  tmpdir=$(/usr/bin/mktemp -d)

  # Collect shaders per playlist
  awk '
    /^\[shaders\."/ {
      name = $0
      gsub(/^\[shaders\."/, "", name)
      gsub(/"\].*/, "", name)
    }
    /^playlists *= *\[/ {
      val = $0
      sub(/^[^=]*= *\[/, "", val)
      gsub(/\].*/, "", val)
      gsub(/"/, "", val)
      n = split(val, arr, /, */)
      for (i = 1; i <= n; i++) {
        gsub(/ /, "", arr[i])
        if (arr[i] != "") print name ".glsl" > (TMPDIR "/" arr[i] ".txt")
      }
    }
  ' TMPDIR="$tmpdir" "$MANIFEST"

  # Sort and write each playlist
  local changed=0
  for pfile in "$tmpdir"/*.txt; do
    local pname
    pname="$(basename "$pfile")"
    sort "$pfile" > "$pfile.sorted"
    local dest="$PLAYLISTS_DIR/$pname"
    if [[ -f "$dest" ]]; then
      local existing_sorted
      existing_sorted=$(/usr/bin/mktemp)
      sort "$dest" > "$existing_sorted"
      if ! diff -q "$pfile.sorted" "$existing_sorted" &>/dev/null; then
        /bin/cp "$pfile.sorted" "$dest"
        echo "Updated: $pname"
        changed=$((changed + 1))
      fi
      rm -f "$existing_sorted"
    else
      /bin/cp "$pfile.sorted" "$dest"
      echo "Created: $pname"
      changed=$((changed + 1))
    fi
  done

  rm -rf "$tmpdir"
  if [[ $changed -eq 0 ]]; then
    echo "All playlists up to date."
  fi
}

cmd_generate_sources_md() {
  local outfile="$SHADERS_DIR/SOURCES.md"
  local total
  total=$(_toml_list_names | wc -l | tr -d ' ')

  {
    echo "# Shader Sources"
    echo ""
    echo "**${total} shaders** from the following sources:"
    echo ""

    # Group by source
    _toml_dump_all | sort -t$'\t' -k4,4 -k1,1 | awk -F'\t' '
      BEGIN { prev_src = "" }
      {
        name = $1; cat = $2; src = $4; desc = $5
        if (src != prev_src) {
          if (prev_src != "") print ""
          print "## " src
          print ""
          prev_src = src
        }
        print "- **" name ".glsl** (" cat ") — " desc
      }
    '
  } > "$outfile"
  echo "Generated: SOURCES.md ($total shaders)"
}

cmd_validate() {
  local errors=0

  # Check every .glsl file has a manifest entry
  for f in "$SHADERS_DIR"/*.glsl; do
    local name
    name="$(basename "$f" .glsl)"
    if ! grep -q "^\[shaders\\.\"${name}\"\]" "$MANIFEST"; then
      echo "MISSING in manifest: ${name}.glsl"
      errors=$((errors + 1))
    fi
  done

  # Check every manifest entry has a .glsl file
  while read -r name; do
    if [[ ! -f "$SHADERS_DIR/${name}.glsl" ]]; then
      echo "ORPHAN in manifest: ${name} (no .glsl file)"
      errors=$((errors + 1))
    fi
  done < <(_toml_list_names)

  if [[ $errors -eq 0 ]]; then
    echo "OK: All $(wc -l < <(_toml_list_names) | tr -d ' ') manifest entries match .glsl files."
  else
    echo "ERRORS: $errors mismatch(es) found."
    return 1
  fi
}

# ── Main ─────────────────────────────────────────
case "${1:-help}" in
  get)                shift; cmd_get "$@" ;;
  list)               shift; cmd_list "$@" ;;
  fzf-lines)          cmd_fzf_lines ;;
  generate-playlists) cmd_generate_playlists ;;
  generate-sources-md) cmd_generate_sources_md ;;
  validate)           cmd_validate ;;
  help|--help|-h)
    echo "Usage: shader-meta <command> [args]"
    echo ""
    echo "Commands:"
    echo "  get <shader> <field>       Get a field (category, cost, source, description, playlists)"
    echo "  list [--category X] [--cost X]  List shaders matching filters"
    echo "  fzf-lines                  Output formatted lines for fzf picker"
    echo "  generate-playlists         Regenerate playlist .txt files from manifest"
    echo "  generate-sources-md        Regenerate SOURCES.md from manifest"
    echo "  validate                   Check manifest ↔ .glsl file consistency"
    ;;
  *)
    echo "Unknown command: $1" >&2
    exit 1 ;;
esac
