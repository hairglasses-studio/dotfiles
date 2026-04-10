#!/usr/bin/env bats

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export TEST_BIN="${BATS_TEST_TMPDIR}/bin"
    export HOME="${BATS_TEST_TMPDIR}/home"
    export PATH="${TEST_BIN}:${PATH}"
    mkdir -p "${TEST_BIN}" "${HOME}" "${HOME}/projects/demo"

    cat > "${TEST_BIN}/hyprctl" <<'EOF'
#!/usr/bin/env bash
case "$1 $2 $3" in
  "monitors all -j"|"monitors -j ")
    cat <<'JSON'
[
  {
    "id": 0,
    "name": "DP-1",
    "width": 2560,
    "height": 1440,
    "refreshRate": 144.0,
    "x": 0,
    "y": 0,
    "scale": 1,
    "transform": 0,
    "description": "Primary"
  }
]
JSON
    ;;
  "clients -j ")
    cat <<'JSON'
[
  {
    "address": "0xabc",
    "class": "kitty",
    "title": "Editor",
    "initialClass": "kitty",
    "initialTitle": "Editor",
    "workspace": {"id": 1, "name": "1"},
    "monitor": 0,
    "floating": false,
    "mapped": true,
    "pinned": false,
    "fullscreen": false,
    "fullscreenMode": 0,
    "at": [10, 20],
    "size": [1200, 900]
  }
]
JSON
    ;;
  "workspaces -j ")
    cat <<'JSON'
[
  {"id": 1, "name": "1", "monitor": "DP-1"}
]
JSON
    ;;
  *)
    printf '%s\n' "$*" >> "${BATS_TEST_TMPDIR}/hyprctl.log"
    printf '{"ok":true}\n'
    ;;
esac
EOF
    chmod +x "${TEST_BIN}/hyprctl"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

_run_hg_module() {
    local module="$1"
    shift
    env \
        HOME="${HOME}" \
        PATH="${PATH}" \
        DOTFILES_DIR="${DOTFILES_DIR}" \
        HG_DOTFILES="${DOTFILES_DIR}" \
        bash -lc '
            source "'"${DOTFILES_DIR}"'/scripts/lib/hg-core.sh"
            source "'"${DOTFILES_DIR}"'/scripts/hg-modules/mod-'"${module}"'.sh"
            '"${module}"'_run "$@"
        ' bash "$@"
}

@test "hg desktop preset save and list use the shared desktop-control state root" {
    run _run_hg_module desktop preset save "Desk Scene"
    assert_success
    assert_output --partial "Saved monitor preset Desk Scene"
    assert_output --partial "${HOME}/.local/state/dotfiles/desktop-control/hypr/monitor-presets/desk-scene.json"

    run _run_hg_module desktop preset list
    assert_success
    assert_output --partial "Desk Scene"
    assert_output --partial "desk-scene.json"
}

@test "hg desktop layout save and project open dry-run reuse the saved scene state" {
    run _run_hg_module desktop preset save "Desk Scene"
    assert_success

    run _run_hg_module desktop layout save "Code Scene" --launch kitty='kitty --class kitty'
    assert_success
    assert_output --partial "Saved layout Code Scene"

    run _run_hg_module desktop project open "${HOME}/projects/demo" --monitor "Desk Scene" --layout "Code Scene" --tmux-session demo --dry-run
    assert_success
    assert_output --partial "restore monitor preset"
    assert_output --partial "restore layout"
    assert_output --partial "dry-run tmux session demo"
}
