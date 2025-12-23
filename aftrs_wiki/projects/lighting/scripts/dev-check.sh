#!/usr/bin/env bash
    # Quick pre-show sanity check for AFTRS lighting repo.
    # - Validates DMX patch
    # - Verifies Mermaid diagram freshness (no stale SVGs)
    # - Optionally pings Art-Net nodes from configs/artnet_plan.yaml
    set -euo pipefail
    REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")"/.. && pwd)"
    cd "$REPO_ROOT"

    ok=1
    warn=0
    err=0

    say_ok()   { printf "\033[32m[OK]\033[0m %s\n" "$*"; }
    say_warn() { printf "\033[33m[WARN]\033[0m %s\n" "$*"; warn=1; }
    say_err()  { printf "\033[31m[ERR]\033[0m %s\n" "$*"; err=1; }

    echo "==> Validating DMX patch"
    if command -v python3 >/dev/null 2>&1; then
      if python3 tools/validate_patch.py; then
        say_ok "DMX patch valid"
      else
        say_err "DMX patch validation failed"
      fi
    else
      say_warn "Python3 not found; skipping DMX validation"
    fi

    echo "==> Checking Mermaid diagram freshness"
    if ./scripts/check_diagrams_fresh.sh; then
      say_ok "Mermaid outputs are fresh"
    else
      code=$?
      if [[ $code -eq 3 ]]; then
        say_warn "No Mermaid renderer (docker/podman/mmdc). Skipping freshness check"
      elif [[ $code -eq 4 ]]; then
        say_err "Mermaid outputs are stale. Run: ./scripts/render_diagrams.sh && git add diagrams/_rendered && git commit -m 'chore: update diagrams'"
      else
        say_err "Unexpected error from check_diagrams_fresh.sh (exit $code)"
      fi
    fi

    echo "==> Pinging Art-Net nodes (optional)"
    CFG="configs/artnet_plan.yaml"
    ips=""
    if command -v python3 >/dev/null 2>&1; then
      python3 - "$CFG" <<'PY' || true
import sys, yaml, os
p = sys.argv[1]
try:
  y = yaml.safe_load(open(p))
except Exception as e:
  sys.exit(1)
nodes = y.get("nodes", {})
ips = [str(v.get("ip")) for v in nodes.values() if v.get("ip")]
for ip in ips:
  print(ip)
PY
    else
      # crude grep fallback
      if [[ -f "$CFG" ]]; then
        grep -E 'ip:\s*[0-9]+' "$CFG" | sed -E 's/.*ip:\s*([0-9.]+).*/\1/'
      fi
    fi | while read -r ip; do
      [[ -z "$ip" ]] && continue
      if ping -c 1 -W 1 "$ip" >/dev/null 2>&1; then
        say_ok "Reachable: $ip"
      else
        say_warn "Unreachable: $ip"
      fi
    done

    echo "==> Summary"
    if [[ $err -eq 1 ]]; then
      echo "One or more errors detected."
      exit 2
    elif [[ $warn -eq 1 ]]; then
      echo "Completed with warnings."
      exit 1
    else
      echo "All checks passed."
      exit 0
    fi
