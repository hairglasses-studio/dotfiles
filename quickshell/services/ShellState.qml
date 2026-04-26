import Quickshell
import QtQuick

QtObject {
    id: root

    readonly property string mode: String(Quickshell.env("SHELL_STACK_MODE") || "full-cutover")
    readonly property string primaryMonitor: String(Quickshell.env("QS_MONITOR") || Quickshell.env("QS_PRIMARY_MONITOR") || "DP-2")
    property bool quickSettingsVisible: false

    function compact(text, max) {
        const clean = (text || "").replace(/\s+/g, " ").trim();
        return clean.length > max ? clean.slice(0, max - 1) + "..." : clean;
    }
}
