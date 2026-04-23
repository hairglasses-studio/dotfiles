import Quickshell
import Quickshell.Io
import QtQuick

Item {
    id: root

    visible: false

    property var entries: []
    property bool hidden: false
    property string dataScript: Quickshell.shellDir + "/../scripts/quickshell-dock-data.sh"
    property string lastError: ""

    function refresh() {
        if (!dataProc.running) dataProc.exec(["bash", dataScript, "list"]);
    }

    function activate(id) {
        if (!id) return;
        Quickshell.execDetached(["bash", dataScript, "activate", String(id)]);
        settleTimer.restart();
    }

    function launch(id) {
        if (!id) return;
        Quickshell.execDetached(["bash", dataScript, "launch", String(id)]);
        settleTimer.restart();
    }

    function toggleHidden() {
        hidden = !hidden;
    }

    function status() {
        return JSON.stringify({
            hidden: hidden,
            count: entries.length,
            error: lastError
        });
    }

    Process {
        id: dataProc
        stdout: SplitParser {
            onRead: data => {
                try {
                    const payload = JSON.parse(data);
                    root.entries = payload.entries || [];
                    root.lastError = "";
                } catch (error) {
                    root.lastError = String(error);
                    root.entries = [];
                }
            }
        }
        stderr: SplitParser {
            onRead: data => root.lastError = String(data || "").trim()
        }
    }

    Timer {
        id: refreshTimer
        interval: 2500
        running: true
        repeat: true
        onTriggered: root.refresh()
    }

    Timer {
        id: settleTimer
        interval: 450
        repeat: false
        onTriggered: root.refresh()
    }

    Component.onCompleted: refresh()
}
