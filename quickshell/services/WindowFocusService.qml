import Quickshell.Io
import QtQuick

Item {
    id: root

    visible: false

    property string windowClass: ""
    property string windowTitle: ""
    readonly property string labelText: {
        if (windowClass.length > 0 && windowTitle.length > 0) return windowClass + " - " + windowTitle;
        return windowClass.length > 0 ? windowClass : windowTitle;
    }
    readonly property bool active: labelText.length > 0
    property bool labelVisible: false

    function compact(text, max) {
        const clean = (text || "").replace(/\s+/g, " ").trim();
        return clean.length > max ? clean.slice(0, max - 1) + "..." : clean;
    }

    function refresh() {
        if (!focusedProc.running) focusedProc.exec(focusedProc.command);
    }

    function ingest(data) {
        const parts = (data || "").split("\t");
        const nextClass = compact(parts[0] || "", 42);
        const nextTitle = compact(parts.slice(1).join("\t") || "", 140);
        const changed = nextClass !== windowClass || nextTitle !== windowTitle;
        windowClass = nextClass;
        windowTitle = nextTitle;

        if (!active) {
            labelVisible = false;
            hideTimer.stop();
            return;
        }

        if (changed) {
            labelVisible = true;
            hideTimer.restart();
        }
    }

    Process {
        id: focusedProc
        command: ["bash", "-lc", "hyprctl -j activewindow 2>/dev/null | jq -r '[.class // \"\", .title // \"\"] | @tsv' 2>/dev/null || true"]
        stdout: SplitParser { onRead: data => root.ingest(data) }
    }

    Timer {
        interval: 750
        running: true
        repeat: true
        onTriggered: root.refresh()
    }

    Timer {
        id: hideTimer
        interval: 3600
        repeat: false
        onTriggered: root.labelVisible = false
    }

    Component.onCompleted: refresh()
}
