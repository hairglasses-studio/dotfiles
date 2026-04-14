#!/usr/bin/env bats

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export TEST_BIN="${BATS_TEST_TMPDIR}/bin"
    export AUX_BIN="${BATS_TEST_TMPDIR}/aux-bin"
    export PATH="${TEST_BIN}:${PATH}"
    mkdir -p "${TEST_BIN}"
    mkdir -p "${AUX_BIN}"
    export WOFI_LOG="${BATS_TEST_TMPDIR}/wofi.log"
    export WOFI_STDIN_LOG="${BATS_TEST_TMPDIR}/wofi.stdin.log"
    export ROFI_LOG="${BATS_TEST_TMPDIR}/rofi.log"
    export ROFI_STDIN_LOG="${BATS_TEST_TMPDIR}/rofi.stdin.log"
    export HYPRCTL_LOG="${BATS_TEST_TMPDIR}/hyprctl.log"
    export HYPRSHELL_LOG="${BATS_TEST_TMPDIR}/hyprshell.log"
    export PGREP_EXIT=1
    export WOFI_CHOICE=""
    export ROFI_CHOICE=""
    export RUNTIME_DIR_DEFAULT="${BATS_TEST_TMPDIR}/run-user"
    export DOTFILES_RUNTIME_SCAN_ROOT="${BATS_TEST_TMPDIR}/runtime-scan"
    export HYPRCTL_MONITORS_JSON='[
      {
        "name": "DP-3",
        "focused": true,
        "width": 5120,
        "height": 1440,
        "scale": 1
      }
    ]'
    export HYPRCTL_LAYERS_JSON='{
      "DP-3": {
        "levels": {
          "3": [
            { "namespace": "hyprshell_overview" },
            { "namespace": "hyprshell_switch" }
          ]
        }
      }
    }'
    export HYPRCTL_CLIENTS_JSON='[
      {
        "address": "0x001",
        "class": "kitty",
        "title": "shell"
      },
      {
        "address": "0x002",
        "class": "firefox",
        "title": "Docs"
      }
    ]'
    export HYPR_TEST_INSTANCE="hypr-test-instance"

    mkdir -p "${RUNTIME_DIR_DEFAULT}" "${DOTFILES_RUNTIME_SCAN_ROOT}"
    mkdir -p "${RUNTIME_DIR_DEFAULT}/hypr/${HYPR_TEST_INSTANCE}"
    touch "${RUNTIME_DIR_DEFAULT}/hypr/${HYPR_TEST_INSTANCE}/hyprshell.sock"

    for cmd in awk bash cat dirname find head id jq readlink sleep sort; do
        ln -s "$(command -v "$cmd")" "${AUX_BIN}/${cmd}"
    done

    cat > "${TEST_BIN}/hyprshell" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "${HYPRSHELL_LOG}"
exit 0
EOF
    chmod +x "${TEST_BIN}/hyprshell"

    cat > "${TEST_BIN}/wofi" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "${WOFI_LOG}"
if [[ " $* " == *" --dmenu "* ]]; then
    cat > "${WOFI_STDIN_LOG}"
    printf '%s\n' "${WOFI_CHOICE:-}"
else
    printf '%s\n' "$*"
fi
EOF
    chmod +x "${TEST_BIN}/wofi"

    cat > "${TEST_BIN}/rofi" <<'EOF'
#!/usr/bin/env bash
printf '%s\n' "$*" >> "${ROFI_LOG}"
if [[ " $* " == *" -dmenu "* ]]; then
    cat > "${ROFI_STDIN_LOG}"
    printf '%s\n' "${ROFI_CHOICE:-}"
else
    printf '%s\n' "$*"
fi
EOF
    chmod +x "${TEST_BIN}/rofi"

    cat > "${TEST_BIN}/hyprctl" <<'EOF'
#!/usr/bin/env bash
case "${1:-} ${2:-}" in
    "-j monitors")
        printf '%s\n' "${HYPRCTL_MONITORS_JSON:-[]}"
        ;;
    "layers -j")
        printf '%s\n' "${HYPRCTL_LAYERS_JSON:-{}}"
        ;;
    "clients -j")
        printf '%s\n' "${HYPRCTL_CLIENTS_JSON:-[]}"
        ;;
    "dispatch focuswindow")
        printf '%s\n' "$*" >> "${HYPRCTL_LOG}"
        ;;
    *)
        printf '%s\n' "$*" >> "${HYPRCTL_LOG}"
        ;;
esac
EOF
    chmod +x "${TEST_BIN}/hyprctl"

    cat > "${TEST_BIN}/pgrep" <<'EOF'
#!/usr/bin/env bash
exit "${PGREP_EXIT:-1}"
EOF
    chmod +x "${TEST_BIN}/pgrep"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

@test "app-launcher sizes wofi for a focused ultrawide monitor" {
    run bash "${SCRIPTS_DIR}/app-launcher.sh"
    assert_success
    assert_output --partial "--show drun"
    assert_output --partial "--monitor DP-3"
    assert_output --partial "--width 922"
    assert_output --partial "--height 835"
}

@test "app-launcher uses default geometry when monitor metadata is missing" {
    export HYPRCTL_MONITORS_JSON='[]'
    run bash "${SCRIPTS_DIR}/app-launcher.sh"
    assert_success
    assert_output --partial "--show drun"
    refute_output --partial "--monitor"
    assert_output --partial "--width 860"
    assert_output --partial "--height 620"
}

@test "app-launcher uses stable fallback launcher by default" {
    export PGREP_EXIT=0
    run env PATH="${TEST_BIN}:${AUX_BIN}" /usr/bin/bash "${SCRIPTS_DIR}/app-launcher.sh"
    assert_success
    assert_output --partial "--show drun"

    run cat "${HYPRSHELL_LOG}"
    assert_failure
}

@test "app-launcher can prefer hyprshell overview when explicitly requested" {
    export PGREP_EXIT=0
    run env PATH="${TEST_BIN}:${AUX_BIN}" DOTFILES_LAUNCHER_PREFER_HYPRSHELL=1 /usr/bin/bash "${SCRIPTS_DIR}/app-launcher.sh"
    assert_success

    run cat "${HYPRSHELL_LOG}"
    assert_success
    assert_output --partial 'socat "OpenOverview"'
}

@test "app-launcher skips hyprshell socket call when daemon socket is missing" {
    export PGREP_EXIT=0
    rm -f "${RUNTIME_DIR_DEFAULT}/hypr/${HYPR_TEST_INSTANCE}/hyprshell.sock"

    run env PATH="${TEST_BIN}:${AUX_BIN}" /usr/bin/bash "${SCRIPTS_DIR}/app-launcher.sh"
    assert_success
    assert_output --partial "--show drun"

    run cat "${HYPRSHELL_LOG}"
    assert_failure
}

@test "app-launcher falls back to rofi when wofi is unavailable" {
    rm -f "${TEST_BIN}/wofi"
    run env PATH="${TEST_BIN}:${AUX_BIN}" /usr/bin/bash "${SCRIPTS_DIR}/app-launcher.sh"
    assert_success
    assert_output --partial "-show drun"
}

@test "app-launcher resolves helper paths when invoked through a symlink" {
    ln -s "${SCRIPTS_DIR}/app-launcher.sh" "${TEST_BIN}/app-launcher"
    run "${TEST_BIN}/app-launcher"
    assert_success
    assert_output --partial "--show drun"
}

@test "app-switcher sizes wofi and focuses the selected client" {
    export WOFI_CHOICE="0x002 firefox — Docs"
    run bash "${SCRIPTS_DIR}/app-switcher.sh"
    assert_success
    assert_output ""

    run cat "${WOFI_LOG}"
    assert_success
    assert_output --partial "--dmenu"
    assert_output --partial "--monitor DP-3"
    assert_output --partial "--width 922"
    assert_output --partial "--height 835"

    run cat "${WOFI_STDIN_LOG}"
    assert_success
    assert_output --partial "0x001 kitty — shell"
    assert_output --partial "0x002 firefox — Docs"

    run cat "${HYPRCTL_LOG}"
    assert_success
    assert_output --partial "dispatch focuswindow address:0x002"
}

@test "app-switcher uses default geometry when monitor metadata is missing" {
    export HYPRCTL_MONITORS_JSON='[]'
    run bash "${SCRIPTS_DIR}/app-switcher.sh"
    assert_success
    assert_output ""

    run cat "${WOFI_LOG}"
    assert_success
    assert_output --partial "--dmenu"
    refute_output --partial "--monitor"
    assert_output --partial "--width 860"
    assert_output --partial "--height 620"
}

@test "app-switcher uses stable fallback switcher by default" {
    export PGREP_EXIT=0
    run env PATH="${TEST_BIN}:${AUX_BIN}" /usr/bin/bash "${SCRIPTS_DIR}/app-switcher.sh"
    assert_success
    assert_output ""

    run cat "${HYPRSHELL_LOG}"
    assert_failure
}

@test "app-switcher can prefer hyprshell switcher when explicitly requested" {
    export PGREP_EXIT=0
    run env PATH="${TEST_BIN}:${AUX_BIN}" DOTFILES_LAUNCHER_PREFER_HYPRSHELL=1 /usr/bin/bash "${SCRIPTS_DIR}/app-switcher.sh"
    assert_success

    run cat "${HYPRSHELL_LOG}"
    assert_success
    assert_output --partial 'socat {"OpenSwitch":{"reverse":false}}'
}

@test "app-switcher can send close and reverse commands to hyprshell" {
    export PGREP_EXIT=0

    run env PATH="${TEST_BIN}:${AUX_BIN}" DOTFILES_LAUNCHER_PREFER_HYPRSHELL=1 /usr/bin/bash "${SCRIPTS_DIR}/app-switcher.sh" reverse
    assert_success
    assert_output ""

    run env PATH="${TEST_BIN}:${AUX_BIN}" DOTFILES_LAUNCHER_PREFER_HYPRSHELL=1 /usr/bin/bash "${SCRIPTS_DIR}/app-switcher.sh" close
    assert_success
    assert_output ""

    run cat "${HYPRSHELL_LOG}"
    assert_success
    assert_output --partial 'socat {"OpenSwitch":{"reverse":true}}'
    assert_output --partial 'socat "CloseSwitch"'
}

@test "app-switcher falls back to rofi when wofi is unavailable" {
    rm -f "${TEST_BIN}/wofi"
    export ROFI_CHOICE="0x001 kitty — shell"
    run env PATH="${TEST_BIN}:${AUX_BIN}" /usr/bin/bash "${SCRIPTS_DIR}/app-switcher.sh"
    assert_success
    assert_output ""

    run cat "${ROFI_LOG}"
    assert_success
    assert_output --partial "-dmenu"
    assert_output --partial "-p Switch to"

    run cat "${HYPRCTL_LOG}"
    assert_success
    assert_output --partial "dispatch focuswindow address:0x001"
}

@test "install.sh print-link-specs includes managed launcher shims" {
    run bash "${DOTFILES_DIR}/install.sh" --print-link-specs
    assert_success
    assert_output --partial "scripts/kitty-shader-playlist.sh|${HOME}/.local/bin/kitty-shader-playlist"
    assert_output --partial "scripts/kitty-shell-launch.sh|${HOME}/.local/bin/kitty-shell-launch"
    assert_output --partial "scripts/kitty-dev-launch.sh|${HOME}/.local/bin/kitty-dev-launch"
    assert_output --partial "scripts/kitty-visual-launch.sh|${HOME}/.local/bin/kitty-visual-launch"
    assert_output --partial "scripts/app-launcher.sh|${HOME}/.local/bin/app-launcher"
    assert_output --partial "scripts/app-switcher.sh|${HOME}/.local/bin/app-switcher"
}
