import Quickshell.Io
import QtQuick

Item {
    id: root

    visible: false

    property int maxSamples: 60
    property int sampleVersion: 0
    property real cpuPct: 0
    property real memPct: 0
    property real netKbps: 0
    property real gpuTemp: 0
    property string cpuText: "CPU --"
    property string memText: "MEM --"
    property string netText: "NET --"
    property string gpuText: "GPU --"
    property var cpuHistory: []
    property var memHistory: []
    property var netHistory: []
    property var gpuHistory: []

    function refresh() {
        if (!telemetryProc.running) telemetryProc.exec(telemetryProc.command);
    }

    function appendSample(history, value) {
        const next = (history || []).concat([Number(value) || 0]);
        return next.slice(Math.max(0, next.length - maxSamples));
    }

    function ingest(data) {
        let payload = {};
        try {
            payload = JSON.parse(data || "{}");
        } catch (error) {
            return;
        }
        if (!payload.ok) return;

        cpuPct = Number(payload.cpu_pct) || 0;
        memPct = Number(payload.mem_pct) || 0;
        netKbps = Number(payload.net_kbps) || 0;
        gpuTemp = Number(payload.gpu_temp_c) || 0;
        cpuText = payload.cpu_text || "CPU --";
        memText = payload.mem_text || "MEM --";
        netText = payload.net_text || "NET --";
        gpuText = payload.gpu_text || "GPU --";
        cpuHistory = appendSample(cpuHistory, cpuPct);
        memHistory = appendSample(memHistory, memPct);
        netHistory = appendSample(netHistory, netKbps);
        gpuHistory = appendSample(gpuHistory, gpuTemp);
        sampleVersion += 1;
    }

    Process {
        id: telemetryProc
        command: ["bash", "-lc", "python3 \"$HOME/hairglasses-studio/dotfiles/scripts/fleet-telemetry-bridge.py\" --once"]
        stdout: SplitParser { onRead: data => root.ingest(data) }
    }

    Timer {
        interval: 1000
        running: true
        repeat: true
        onTriggered: root.refresh()
    }

    Component.onCompleted: refresh()
}
