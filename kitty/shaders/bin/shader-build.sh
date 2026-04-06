#!/usr/bin/env bash
set -euo pipefail

# shader-build.sh — Transpile Ghostty shaders to CRTty and DarkWindow formats
# Reads from ghostty/shaders/*.glsl (canonical source)
# Outputs to kitty/shaders/crtty/ and kitty/shaders/darkwindow/

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
DOTFILES_DIR="$(cd "$SCRIPT_DIR/../../.." && pwd)"
GHOSTTY_SHADERS="$DOTFILES_DIR/ghostty/shaders"
CRTTY_OUT="$DOTFILES_DIR/kitty/shaders/crtty"
DARKWINDOW_OUT="$DOTFILES_DIR/kitty/shaders/darkwindow"
TRANSPILE_CRTTY="$DOTFILES_DIR/scripts/lib/shader-transpile-crtty.sh"
TRANSPILE_DARKWINDOW="$DOTFILES_DIR/scripts/lib/shader-transpile-darkwindow.sh"

source "$DOTFILES_DIR/scripts/lib/hg-core.sh"

mode="${1:-build}"

case "$mode" in
  --check|check)
    hg_info "Checking transpilation compatibility..."
    ok=0 warn=0 total=0
    for src in "$GHOSTTY_SHADERS"/*.glsl; do
      name="$(basename "$src")"
      total=$((total + 1))
      crtty_status="$("$TRANSPILE_CRTTY" --check "$src")"
      dw_status="$("$TRANSPILE_DARKWINDOW" --check "$src")"
      if [[ "$crtty_status" == "OK" && "$dw_status" == "OK" ]]; then
        ok=$((ok + 1))
      else
        warn=$((warn + 1))
        printf "  ${HG_YELLOW}%-45s${HG_RESET} CRTty: %-20s DarkWindow: %s\n" "$name" "$crtty_status" "$dw_status"
      fi
    done
    echo ""
    hg_ok "$ok/$total clean, $warn warnings"
    ;;

  build|"")
    mkdir -p "$CRTTY_OUT" "$DARKWINDOW_OUT"

    built=0 skipped=0 failed=0
    for src in "$GHOSTTY_SHADERS"/*.glsl; do
      name="$(basename "$src")"
      crtty_dst="$CRTTY_OUT/$name"
      dw_dst="$DARKWINDOW_OUT/$name"

      # Skip if outputs are newer than source (idempotent)
      if [[ -f "$crtty_dst" && -f "$dw_dst" && \
            "$crtty_dst" -nt "$src" && "$dw_dst" -nt "$src" ]]; then
        skipped=$((skipped + 1))
        continue
      fi

      # Transpile CRTty
      if ! "$TRANSPILE_CRTTY" "$src" "$crtty_dst" 2>/dev/null; then
        hg_warn "CRTty transpile failed: $name"
        failed=$((failed + 1))
        continue
      fi

      # Transpile DarkWindow
      if ! "$TRANSPILE_DARKWINDOW" "$src" "$dw_dst" 2>/dev/null; then
        hg_warn "DarkWindow transpile failed: $name"
        failed=$((failed + 1))
        continue
      fi

      built=$((built + 1))
    done

    hg_ok "Built: $built, Skipped (up-to-date): $skipped, Failed: $failed"
    hg_info "CRTty:      $CRTTY_OUT/ ($(ls "$CRTTY_OUT"/*.glsl 2>/dev/null | wc -l) shaders)"
    hg_info "DarkWindow: $DARKWINDOW_OUT/ ($(ls "$DARKWINDOW_OUT"/*.glsl 2>/dev/null | wc -l) shaders)"
    ;;

  clean)
    rm -f "$CRTTY_OUT"/*.glsl "$DARKWINDOW_OUT"/*.glsl
    hg_ok "Cleaned transpiled shaders"
    ;;

  *)
    cat <<EOF
Usage: shader-build.sh [command]

Commands:
  build   Transpile all Ghostty shaders to CRTty + DarkWindow (default)
  check   Report compatibility warnings without building
  clean   Remove all transpiled shaders
EOF
    ;;
esac
