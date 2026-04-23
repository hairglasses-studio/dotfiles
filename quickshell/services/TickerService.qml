import Quickshell
import Quickshell.Io
import QtQuick

Item {
    id: root

    visible: false

    property string bridgePath: Quickshell.shellDir + "/../scripts/ticker-bridge.py"
    property var streams: ["keybinds"]
    property int streamIndex: 0
    property string stream: "keybinds"
    property string text: "loading ticker stream..."
    property string status: "KEYBINDS"
    property string preset: ""
    property int refreshMs: 300000
    property var segments: []
    property bool pendingRestart: false

    function compact(value, max) {
        const clean = (value || "").replace(/\s+/g, " ").trim();
        return clean.length > max ? clean.slice(0, max - 1) + "..." : clean;
    }

    function applyPayload(data) {
        try {
            const payload = JSON.parse(data);
            if (!payload.ok) {
                root.status = "TICKER ERR";
                root.text = payload.error || "ticker bridge failed";
                root.refreshMs = 30000;
                root.segments = [];
                return;
            }
            root.stream = payload.stream || root.stream;
            root.status = root.stream.toUpperCase();
            root.text = root.compact(payload.text || payload.markup || "", 5000);
            root.refreshMs = Math.max(5000, (payload.refresh || 300) * 1000);
            root.preset = payload.preset || "";
            root.segments = payload.segments || [];
        } catch (error) {
            root.status = "TICKER JSON";
            root.text = root.compact(String(error), 240);
            root.refreshMs = 30000;
            root.segments = [];
        }
    }

    function applyListPayload(data) {
        try {
            const payload = JSON.parse(data);
            if (payload.ok && payload.streams && payload.streams.length > 0) {
                root.streams = payload.streams;
                const idx = root.streams.indexOf(root.stream);
                root.streamIndex = idx >= 0 ? idx : 0;
                root.stream = root.streams[root.streamIndex];
            }
        } catch (_error) {
        }
    }

    function startTickerNow() {
        root.pendingRestart = false;
        tickerProc.exec(["python3", root.bridgePath, "--stream", root.stream, "--watch"]);
    }

    function restartTicker() {
        root.pendingRestart = true;
        if (tickerProc.running) {
            tickerProc.signal(15);
        } else {
            startTickerNow();
        }
    }

    function advance() {
        if (root.streams.length <= 1) return;
        root.streamIndex = (root.streamIndex + 1) % root.streams.length;
        root.stream = root.streams[root.streamIndex];
        restartTicker();
    }

    onTextChanged: sweepRequested()
    signal sweepRequested()

    Process {
        id: tickerProc
        stdout: SplitParser { onRead: data => root.applyPayload(data) }
        stderr: SplitParser {
            onRead: data => {
                root.status = "TICKER STDERR";
                root.text = root.compact(data, 240);
            }
        }
        onExited: {
            if (root.pendingRestart) root.startTickerNow();
        }
    }

    Process {
        id: tickerListProc
        command: ["python3", root.bridgePath, "--list"]
        stdout: SplitParser { onRead: data => root.applyListPayload(data) }
    }

    Timer {
        interval: 60000
        running: true
        repeat: true
        onTriggered: root.advance()
    }

    Timer {
        interval: 300000
        running: true
        repeat: true
        onTriggered: {
            if (!tickerListProc.running) tickerListProc.exec(tickerListProc.command);
        }
    }

    Component.onCompleted: {
        tickerListProc.exec(tickerListProc.command);
        startTickerNow();
    }
}
