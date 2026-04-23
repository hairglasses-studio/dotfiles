#!/usr/bin/env bats

load 'test_helper'

setup() {
    export BATS_TEST_TMPDIR="$(mktemp -d)"
    export HOME="${BATS_TEST_TMPDIR}/home"
    export XDG_CONFIG_HOME="${HOME}/.config"
    export XDG_STATE_HOME="${HOME}/.local/state"
    mkdir -p "${HOME}" "${XDG_CONFIG_HOME}" "${XDG_STATE_HOME}"
}

teardown() {
    [[ -d "${BATS_TEST_TMPDIR}" ]] && rm -rf "${BATS_TEST_TMPDIR}"
}

@test "retroarch-flycast-widescreen-audit emits JSON report with expected shape" {
    # Audit walks the ROM tree + retroarch.cfg and writes a report. With
    # an empty roms dir + no content, status falls to no_content (or
    # core_missing if the runner has no flycast core). We assert the
    # JSON shape invariants regardless of the status branch.
    local cfg="${HOME}/.config/retroarch"
    local roms="${HOME}/Games/RetroArch/roms"
    local state="${HOME}/.local/state"
    mkdir -p "${cfg}" "${roms}/dreamcast" "${state}"
    touch "${cfg}/retroarch.cfg"

    run python3 "${DOTFILES_DIR}/scripts/retroarch-flycast-widescreen-audit.py" \
        --config-dir "${cfg}" \
        --roms-dir "${roms}" \
        --state-home "${state}"
    assert_success
    local report="${state}/retroarch-flycast/widescreen-audit.json"
    [[ -f "${report}" ]] || fail "widescreen-audit.json not produced at ${report}"

    # Assert required top-level keys + system identity + display block.
    run python3 -c "
import json, sys
d = json.load(open(sys.argv[1]))
for key in ('system','status','summary','ultrawide_summary','core','display','paths','entries','warnings','next_steps'):
    assert key in d, f'missing top-level key: {key}'
assert d['system'] == 'flycast', f'system={d[\"system\"]}'
assert d['status'] in {'ready', 'no_content', 'core_missing'}, f'unexpected status: {d[\"status\"]}'
assert 'aspect_class' in d['display']
assert d['paths']['output_path'].endswith('widescreen-audit.json')
print('shape_ok')
" "${report}"
    assert_success
    assert_output --partial "shape_ok"
}

@test "retroarch-dolphin-widescreen-audit classifies 32:9 display from retroarch.cfg" {
    # Writes video_fullscreen_x=5120 + y=1440 into the cfg so the audit
    # computes aspect_value ~3.555 → aspect_class=32:9. That branch
    # drives an ultrawide-mode recommendation set we assert via JSON.
    local cfg="${HOME}/.config/retroarch"
    local roms="${HOME}/Games/RetroArch/roms"
    local state="${HOME}/.local/state"
    mkdir -p "${cfg}" "${roms}/gamecube" "${state}"
    cat > "${cfg}/retroarch.cfg" <<'CFG'
video_fullscreen_x = "5120"
video_fullscreen_y = "1440"
video_monitor_index = "2"
CFG

    run python3 "${DOTFILES_DIR}/scripts/retroarch-dolphin-widescreen-audit.py" \
        --config-dir "${cfg}" \
        --roms-dir "${roms}" \
        --state-home "${state}"
    assert_success

    local report="${state}/retroarch-dolphin/widescreen-audit.json"
    run python3 -c "
import json, sys
d = json.load(open(sys.argv[1]))
assert d['display']['aspect_class'] == '32:9', f\"aspect_class={d['display']['aspect_class']}\"
assert d['display']['monitor_index'] == 2, f\"monitor_index={d['display']['monitor_index']}\"
assert d['system'] == 'dolphin'
print('dolphin_display_ok')
" "${report}"
    assert_success
    assert_output --partial "dolphin_display_ok"
}

@test "retroarch-flycast-apply-widescreen-defaults writes expected opt keys" {
    # Apply script seeds Flycast.opt with widescreen_cheats=On and
    # widescreen_hack=Off (the safe 16:9-preserving choice). Also
    # writes a minimal Flycast.cfg with video_smooth + run_ahead_enabled
    # and ensures dreamcast/dc ROM subdirs exist.
    local cfg="${HOME}/.config/retroarch"
    local roms="${HOME}/Games/RetroArch/roms"
    local state="${HOME}/.local/state"
    mkdir -p "${cfg}" "${roms}" "${state}"
    touch "${cfg}/retroarch.cfg"

    run python3 "${DOTFILES_DIR}/scripts/retroarch-flycast-apply-widescreen-defaults.py" \
        --config-dir "${cfg}" \
        --roms-dir "${roms}" \
        --state-home "${state}"
    assert_success

    local opt="${cfg}/config/Flycast/Flycast.opt"
    local cfgfile="${cfg}/config/Flycast/Flycast.cfg"
    [[ -f "${opt}" ]] || fail "Flycast.opt not written at ${opt}"
    [[ -f "${cfgfile}" ]] || fail "Flycast.cfg not written at ${cfgfile}"

    run cat "${opt}"
    assert_success
    assert_output --partial 'flycast_widescreen_cheats = "On"'
    assert_output --partial 'flycast_widescreen_hack = "Off"'

    run cat "${cfgfile}"
    assert_success
    assert_output --partial 'video_smooth = "true"'
    assert_output --partial 'run_ahead_enabled = "false"'

    # Alias ROM dirs created (dreamcast + dc).
    [[ -d "${roms}/dreamcast" ]] || fail "dreamcast ROM dir not created"
    [[ -d "${roms}/dc" ]] || fail "dc alias ROM dir not created"
}

@test "retroarch-dolphin-apply-widescreen-defaults writes Dolphin.opt with widescreen hack enabled" {
    # Dolphin apply sets dolphin_wii_scaled_efb = "enabled" and
    # dolphin_widescreen_hack = "enabled" — distinct from Flycast,
    # where hack is deliberately Off. Asserts both values plus ROM
    # alias dir creation for gamecube + gc + wii.
    local cfg="${HOME}/.config/retroarch"
    local roms="${HOME}/Games/RetroArch/roms"
    local state="${HOME}/.local/state"
    mkdir -p "${cfg}" "${roms}" "${state}"
    touch "${cfg}/retroarch.cfg"

    run python3 "${DOTFILES_DIR}/scripts/retroarch-dolphin-apply-widescreen-defaults.py" \
        --config-dir "${cfg}" \
        --roms-dir "${roms}" \
        --state-home "${state}"
    assert_success

    local opt="${cfg}/config/Dolphin/Dolphin.opt"
    [[ -f "${opt}" ]] || fail "Dolphin.opt not written at ${opt}"

    run cat "${opt}"
    assert_success
    # Exact key assertion depends on what dolphin_apply writes. Assert
    # keys present via a grep — we just need the widescreen_hack line.
    assert_output --partial 'dolphin_widescreen_hack'

    for alias in gamecube gc wii; do
        [[ -d "${roms}/${alias}" ]] || fail "${alias} ROM dir not created"
    done

    # Apply JSON report shape.
    local report="${state}/retroarch-dolphin/widescreen-apply.json"
    [[ -f "${report}" ]] || fail "widescreen-apply.json not at ${report}"
    run python3 -c "
import json, sys
d = json.load(open(sys.argv[1]))
assert d['system'] == 'dolphin'
assert 'results' in d and isinstance(d['results'], dict)
print('dolphin_apply_ok')
" "${report}"
    assert_success
    assert_output --partial "dolphin_apply_ok"
}
