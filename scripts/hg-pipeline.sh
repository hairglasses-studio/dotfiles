#!/usr/bin/env bash
# hg-pipeline.sh — Build+test pipeline for hairglasses-studio repos.
# Usage: hg-pipeline.sh [REPO_DIR] [--json] [--build-only] [--test-only]
#
# Supports Go, Node.js, Python, and hybrid (Go+Node) projects.
set -euo pipefail
# Detects Makefile targets and falls back to language-native commands.
# Sources hg-core.sh for Hairglasses Neon palette formatted output.
# Exit codes: 0=pass, 1=build fail, 2=test fail, 3=vet fail.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "$SCRIPT_DIR/lib/hg-core.sh"

# ── Defaults ─────────────────────────────────
REPO_DIR=""
JSON_OUTPUT=false
BUILD_ONLY=false
TEST_ONLY=false

# ── Argument parsing ─────────────────────────
for arg in "$@"; do
  case "$arg" in
    --json)       JSON_OUTPUT=true ;;
    --build-only) BUILD_ONLY=true ;;
    --test-only)  TEST_ONLY=true ;;
    -*)           hg_die "Unknown flag: $arg" ;;
    *)            REPO_DIR="$arg" ;;
  esac
done

if [[ -n "$REPO_DIR" ]]; then
  cd "$REPO_DIR"
fi

REPO_NAME="$(basename "$PWD")"

# ── Language detection ───────────────────────
HAS_GOMOD=false;    [[ -f go.mod ]]         && HAS_GOMOD=true
HAS_PACKAGE=false;  [[ -f package.json ]]   && HAS_PACKAGE=true
HAS_PYTHON=false;   [[ -f pyproject.toml ]] || [[ -f requirements.txt ]] || [[ -f setup.py ]] && HAS_PYTHON=true
HAS_MAKEFILE=false; [[ -f Makefile ]]       && HAS_MAKEFILE=true

# Determine project language(s)
LANG_GO=false;     $HAS_GOMOD && LANG_GO=true
LANG_NODE=false;   $HAS_PACKAGE && LANG_NODE=true
LANG_PYTHON=false; $HAS_PYTHON && LANG_PYTHON=true

# Hybrid: Go + Node (e.g., hg-mcp with web UI)
LANG_HYBRID=false
$LANG_GO && $LANG_NODE && LANG_HYBRID=true

if ! $LANG_GO && ! $LANG_NODE && ! $LANG_PYTHON; then
  hg_die "No buildable project detected (no go.mod, package.json, or pyproject.toml in $PWD)"
fi

# ── MCP detection ────────────────────────────
IS_MCP=false
DOTFILES_MCP="$HOME/hairglasses-studio/dotfiles/.mcp.json"

if [[ -f .mcp.json ]]; then
  IS_MCP=true
elif [[ "$REPO_NAME" == *-mcp ]]; then
  IS_MCP=true
elif [[ -d cmd ]] && ls cmd/ 2>/dev/null | grep -q mcp; then
  IS_MCP=true
elif [[ -f "$DOTFILES_MCP" ]] && grep -q "$REPO_NAME" "$DOTFILES_MCP" 2>/dev/null; then
  IS_MCP=true
fi

# ── Target detection ─────────────────────────
has_make_target() {
  $HAS_MAKEFILE && make -n "$1" &>/dev/null
}

# Check if package.json has a script
has_npm_script() {
  $HAS_PACKAGE && python3 -c "import json; d=json.load(open('package.json')); exit(0 if '$1' in d.get('scripts',{}) else 1)" 2>/dev/null
}

# ── Step runners ─────────────────────────────
STEP_RESULTS=()
TOTAL_START=$(date +%s%N)

run_step() {
  local name="$1"
  shift
  local start_ns=$(date +%s%N)

  hg_info "Step: $name"
  local exit_code=0
  "$@" 2>&1 || exit_code=$?

  local end_ns=$(date +%s%N)
  local duration_ms=$(( (end_ns - start_ns) / 1000000 ))

  if [[ $exit_code -eq 0 ]]; then
    hg_ok "$name passed (${duration_ms}ms)"
    STEP_RESULTS+=("{\"name\":\"$name\",\"status\":\"pass\",\"duration_ms\":$duration_ms}")
  else
    hg_error "$name failed (exit $exit_code, ${duration_ms}ms)"
    STEP_RESULTS+=("{\"name\":\"$name\",\"status\":\"fail\",\"exit_code\":$exit_code,\"duration_ms\":$duration_ms}")
  fi
  return $exit_code
}

# ── Go steps ─────────────────────────────────
do_build_go() {
  if has_make_target build; then
    make build
  else
    go build ./...
  fi
}

do_vet_go() {
  if has_make_target vet; then
    make vet
  else
    go vet ./...
  fi
}

do_test_go() {
  if has_make_target test; then
    make test
  else
    go test ./... -count=1
  fi
}

do_lint_go() {
  if has_make_target lint; then
    make lint
  elif command -v golangci-lint &>/dev/null; then
    golangci-lint run ./...
  else
    hg_warn "No Go linter available, skipping"
  fi
}

# ── Node.js steps ────────────────────────────
do_build_node() {
  if has_make_target build; then
    make build
  elif has_npm_script build; then
    npm install && npm run build
  else
    npm install
  fi
}

do_test_node() {
  if has_make_target test; then
    make test
  elif has_npm_script test; then
    npm test
  else
    hg_warn "No Node.js test script found, skipping"
  fi
}

do_lint_node() {
  if has_npm_script lint; then
    npm run lint
  else
    hg_warn "No Node.js lint script found, skipping"
  fi
}

# ── Python steps ─────────────────────────────
do_build_python() {
  if has_make_target build; then
    make build
  elif [[ -n "${VIRTUAL_ENV:-}" ]]; then
    # Inside a venv — safe to install
    if [[ -f requirements.txt ]]; then
      pip install -r requirements.txt --quiet
    elif [[ -f pyproject.toml ]]; then
      pip install -e . --quiet 2>/dev/null || pip install . --quiet
    fi
  else
    # No venv — check syntax only, don't install (PEP 668)
    hg_warn "No virtualenv active — skipping pip install (PEP 668)"
    if command -v python3 &>/dev/null; then
      python3 -m py_compile "$(find . -name '*.py' -not -path './.*' | head -1)" 2>&1 || true
    fi
  fi
}

do_test_python() {
  if has_make_target test; then
    make test
  elif command -v pytest &>/dev/null; then
    pytest
  elif python3 -m pytest --version &>/dev/null 2>&1; then
    python3 -m pytest
  else
    hg_warn "No Python test runner found, skipping"
  fi
}

do_lint_python() {
  if has_make_target lint; then
    make lint
  elif command -v ruff &>/dev/null; then
    ruff check .
  elif command -v mypy &>/dev/null; then
    mypy .
  else
    hg_warn "No Python linter available, skipping"
  fi
}

# ── Language label ───────────────────────────
LANG_LABEL=""
$LANG_HYBRID && LANG_LABEL="Go+Node" || {
  $LANG_GO && LANG_LABEL="Go"
  $LANG_NODE && [[ -z "$LANG_LABEL" ]] && LANG_LABEL="Node.js"
  $LANG_PYTHON && [[ -z "$LANG_LABEL" ]] && LANG_LABEL="Python"
}

# ── Header ───────────────────────────────────
echo ""
printf "%s━━━ Pipeline: %s ━━━%s\n" "$HG_CYAN" "$REPO_NAME" "$HG_RESET"
LANG_VERSION=""
if $LANG_GO; then
  LANG_VERSION="$(go version 2>/dev/null | awk '{print $3}')"
elif $LANG_NODE; then
  LANG_VERSION="node $(node --version 2>/dev/null)"
elif $LANG_PYTHON; then
  LANG_VERSION="python $(python3 --version 2>/dev/null | awk '{print $2}')"
fi
printf "%sLang: %s (%s) | Makefile: %s | MCP: %s%s\n" \
  "$HG_DIM" \
  "$LANG_LABEL" \
  "$LANG_VERSION" \
  "$($HAS_MAKEFILE && echo "yes" || echo "no")" \
  "$($IS_MCP && echo "yes" || echo "no")" \
  "$HG_RESET"
echo ""

# ── Run pipeline ─────────────────────────────
FAILED_STEP=""

# Build
if ! $TEST_ONLY; then
  if $LANG_GO; then
    # For hybrid repos, `make build` handles web+Go in order
    if ! run_step "build" do_build_go; then
      FAILED_STEP="build"
    fi
  elif $LANG_NODE; then
    if ! run_step "build" do_build_node; then
      FAILED_STEP="build"
    fi
  elif $LANG_PYTHON; then
    if ! run_step "build" do_build_python; then
      FAILED_STEP="build"
    fi
  fi
fi

# Vet (Go only)
if [[ -z "$FAILED_STEP" ]] && ! $BUILD_ONLY && $LANG_GO; then
  if ! run_step "vet" do_vet_go; then
    FAILED_STEP="vet"
  fi
fi

# Test
if [[ -z "$FAILED_STEP" ]] && ! $BUILD_ONLY; then
  if $LANG_GO; then
    if ! run_step "test" do_test_go; then
      FAILED_STEP="test"
    fi
  elif $LANG_NODE; then
    if ! run_step "test" do_test_node; then
      FAILED_STEP="test"
    fi
  elif $LANG_PYTHON; then
    if ! run_step "test" do_test_python; then
      FAILED_STEP="test"
    fi
  fi
fi

# Lint (non-blocking)
if [[ -z "$FAILED_STEP" ]] && ! $BUILD_ONLY && ! $TEST_ONLY; then
  if $LANG_GO; then
    run_step "lint" do_lint_go || hg_warn "Lint issues found (non-blocking)"
  elif $LANG_NODE; then
    run_step "lint" do_lint_node || hg_warn "Lint issues found (non-blocking)"
  elif $LANG_PYTHON; then
    run_step "lint" do_lint_python || hg_warn "Lint issues found (non-blocking)"
  fi
fi

# ── Summary ──────────────────────────────────
TOTAL_END=$(date +%s%N)
TOTAL_MS=$(( (TOTAL_END - TOTAL_START) / 1000000 ))

echo ""
if [[ -z "$FAILED_STEP" ]]; then
  printf "%s━━━ PASS ━━━%s %s(%s, %dms)%s\n" "$HG_GREEN" "$HG_RESET" "$HG_DIM" "$REPO_NAME" "$TOTAL_MS" "$HG_RESET"
else
  printf "%s━━━ FAIL at %s ━━━%s %s(%s, %dms)%s\n" "$HG_RED" "$FAILED_STEP" "$HG_RESET" "$HG_DIM" "$REPO_NAME" "$TOTAL_MS" "$HG_RESET"
fi

# ── JSON output ──────────────────────────────
if $JSON_OUTPUT; then
  STEPS_JSON=$(printf '%s,' "${STEP_RESULTS[@]}")
  STEPS_JSON="[${STEPS_JSON%,}]"
  printf '{"repo":"%s","lang":"%s","is_mcp":%s,"overall":"%s","failed_step":"%s","total_ms":%d,"steps":%s}\n' \
    "$REPO_NAME" \
    "$LANG_LABEL" \
    "$($IS_MCP && echo "true" || echo "false")" \
    "$([[ -z "$FAILED_STEP" ]] && echo "pass" || echo "fail")" \
    "$FAILED_STEP" \
    "$TOTAL_MS" \
    "$STEPS_JSON"
fi

# ── Exit code ────────────────────────────────
case "$FAILED_STEP" in
  build) exit 1 ;;
  test)  exit 2 ;;
  vet)   exit 3 ;;
  "")    exit 0 ;;
  *)     exit 1 ;;
esac
