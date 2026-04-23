import Quickshell.Io
import QtQuick

Item {
    id: root

    visible: false

    property string status: "Idle"
    property string line: ""
    property string color: "#66708f"
    property string title: ""
    property string artist: ""
    property bool synced: false
    readonly property bool playing: status === "Playing"
    readonly property bool hasText: line.length > 0

    function refresh() {
        if (!lyricsProc.running) lyricsProc.exec(lyricsProc.command);
    }

    function compact(text, max) {
        const clean = (text || "").replace(/\s+/g, " ").trim();
        return clean.length > max ? clean.slice(0, max - 1) + "..." : clean;
    }

    function ingest(data) {
        let payload = {};
        try {
            payload = JSON.parse(data || "{}");
        } catch (error) {
            return;
        }
        if (!payload.ok) return;
        status = payload.status || "Idle";
        line = compact(payload.line || "", 160);
        color = payload.color || "#66708f";
        title = compact(payload.title || "", 90);
        artist = compact(payload.artist || "", 60);
        synced = !!payload.synced;
    }

    Process {
        id: lyricsProc
        command: ["bash", "-lc", "python3 \"$HOME/hairglasses-studio/dotfiles/scripts/lyrics-bridge.py\" --once"]
        stdout: SplitParser { onRead: data => root.ingest(data) }
    }

    Timer {
        interval: 1500
        running: true
        repeat: true
        onTriggered: root.refresh()
    }

    Component.onCompleted: refresh()
}
