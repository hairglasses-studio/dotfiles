import Quickshell
import Quickshell.Io
import QtQuick

Item {
    id: root

    visible: false

    property string workspacesText: "..."
    property string focusedTitle: ""
    property string volumeText: ""
    property string networkText: ""
    property string gpuText: ""
    property string updatesText: ""
    property string weatherText: ""
    property string retroarchText: ""
    property string mxText: ""
    property string fleetText: ""
    property string claudeText: ""
    property string shaderText: ""
    property string mediaText: ""
    property string bluetoothText: ""
    property string systemText: ""

    function compact(text, max) {
        const clean = (text || "").replace(/\s+/g, " ").trim();
        return clean.length > max ? clean.slice(0, max - 1) + "..." : clean;
    }

    function refresh(proc) {
        if (!proc.running) proc.exec(proc.command);
    }

    Process {
        id: workspacesProc
        command: ["bash", "-lc", "active=$(hyprctl -j activeworkspace 2>/dev/null | jq -r '.id // 0'); hyprctl -j workspaces 2>/dev/null | jq -r --arg active \"$active\" 'sort_by(.id) | map(if ((.id | tostring) == $active) then \"[\" + (.name | tostring) + \"]\" else (.name | tostring) end) | join(\"  \")'"]
        stdout: SplitParser { onRead: data => root.workspacesText = root.compact(data, 120) }
    }

    Process {
        id: focusedProc
        command: ["bash", "-lc", "hyprctl -j activewindow 2>/dev/null | jq -r '.title // empty'"]
        stdout: SplitParser { onRead: data => root.focusedTitle = root.compact(data, 64) }
    }

    Process {
        id: volumeProc
        command: ["bash", "-lc", "wpctl get-volume @DEFAULT_AUDIO_SINK@ 2>/dev/null | awk '{muted=index($0,\"MUTED\")?\" MUTED\":\"\"; printf \"%d%%\" muted, $2 * 100}'"]
        stdout: SplitParser { onRead: data => root.volumeText = root.compact(data, 18) }
    }

    Process {
        id: networkProc
        command: ["bash", "-lc", "ip -j route get 1.1.1.1 2>/dev/null | jq -r '.[0].dev // empty'"]
        stdout: SplitParser { onRead: data => root.networkText = root.compact(data, 16) }
    }

    Process {
        id: cacheProc
        command: ["bash", "-lc", "printf '%s\\t%s\\t%s\\t%s\\t%s\\n' \"$(cat /tmp/bar-gpu.txt 2>/dev/null)\" \"$(cat /tmp/bar-updates.txt 2>/dev/null)\" \"$(cat /tmp/bar-weather.txt 2>/dev/null)\" \"$(cat /tmp/bar-retroarch.txt 2>/dev/null)\" \"$(cat /tmp/bar-mx.txt 2>/dev/null)\""]
        stdout: SplitParser {
            onRead: data => {
                const parts = data.split("\t");
                root.gpuText = root.compact(parts[0] || "", 18);
                root.updatesText = root.compact(parts[1] || "", 18);
                root.weatherText = root.compact(parts[2] || "", 24);
                root.retroarchText = root.compact(parts[3] || "", 26);
                root.mxText = root.compact(parts[4] || "", 18);
            }
        }
    }

    Process {
        id: fleetProc
        command: ["bash", "-lc", "jq -r '\"RG \\(.fleet.running // 0)/\\(.fleet.total // 0)  \\(.loops.total_runs // 0)L  $\\(.cost.total_spend_usd // 0)\"' /tmp/rg-status.json 2>/dev/null || true"]
        stdout: SplitParser { onRead: data => root.fleetText = root.compact(data, 36) }
    }

    Process {
        id: claudeProc
        command: ["bash", "-lc", "f=${XDG_STATE_HOME:-$HOME/.local/state}/claude/bar-state; [ -f \"$f\" ] && awk -F '\\t' '{printf \"CC %s\", $2; if ($3 != \"\" && $3 != \"0\") printf \" $\" $3}' \"$f\" || true"]
        stdout: SplitParser { onRead: data => root.claudeText = root.compact(data, 34) }
    }

    Process {
        id: shaderProc
        command: ["bash", "-lc", "cat ${XDG_STATE_HOME:-$HOME/.local/state}/kitty-shaders/current-label 2>/dev/null || cat ${XDG_STATE_HOME:-$HOME/.local/state}/kitty-shaders/current 2>/dev/null | sed 's/[.]glsl$//' || true"]
        stdout: SplitParser { onRead: data => root.shaderText = root.compact(data, 26) }
    }

    Process {
        id: mediaProc
        command: ["bash", "-lc", "timeout 1 playerctl metadata --format '{{title}} -- {{artist}}' 2>/dev/null || true"]
        stdout: SplitParser { onRead: data => root.mediaText = root.compact(data, 36) }
    }

    Process {
        id: bluetoothProc
        command: ["bash", "-lc", "n=$(bluetoothctl devices Connected 2>/dev/null | wc -l); [ \"$n\" -gt 0 ] && printf 'BT %s' \"$n\" || true"]
        stdout: SplitParser { onRead: data => root.bluetoothText = root.compact(data, 12) }
    }

    Process {
        id: systemProc
        command: ["bash", "-lc", "load=$(cut -d' ' -f1 /proc/loadavg); mem=$(free | awk '/Mem:/ {printf \"%d%%\", $3 * 100 / $2}'); printf 'LOAD %s MEM %s' \"$load\" \"$mem\""]
        stdout: SplitParser { onRead: data => root.systemText = root.compact(data, 24) }
    }

    Timer {
        interval: 1000
        running: true
        repeat: true
        onTriggered: {
            root.refresh(workspacesProc);
            root.refresh(focusedProc);
            root.refresh(volumeProc);
            root.refresh(networkProc);
        }
    }

    Timer {
        interval: 10000
        running: true
        repeat: true
        onTriggered: {
            root.refresh(cacheProc);
            root.refresh(fleetProc);
            root.refresh(claudeProc);
            root.refresh(shaderProc);
            root.refresh(mediaProc);
            root.refresh(bluetoothProc);
            root.refresh(systemProc);
        }
    }

    Component.onCompleted: {
        refresh(workspacesProc);
        refresh(focusedProc);
        refresh(volumeProc);
        refresh(networkProc);
        refresh(cacheProc);
        refresh(fleetProc);
        refresh(claudeProc);
        refresh(shaderProc);
        refresh(mediaProc);
        refresh(bluetoothProc);
        refresh(systemProc);
    }
}
