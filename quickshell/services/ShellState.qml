import Quickshell
import QtQuick

QtObject {
    id: root

    readonly property string mode: String(Quickshell.env("SHELL_STACK_MODE") || "pilot")
    readonly property bool barCutover: String(Quickshell.env("QS_BAR_CUTOVER") || "0") === "1"
    readonly property bool tickerCutover: String(Quickshell.env("QS_TICKER_CUTOVER") || "0") === "1"
    readonly property bool menuCutover: String(Quickshell.env("QS_MENU_CUTOVER") || "0") === "1"
    readonly property bool dockCutover: String(Quickshell.env("QS_DOCK_CUTOVER") || "0") === "1"
    readonly property bool companionCutover: String(Quickshell.env("QS_COMPANION_CUTOVER") || "0") === "1"
    readonly property bool notificationOwner: String(Quickshell.env("QUICKSHELL_NOTIFICATION_OWNER") || "0") === "1"
    readonly property string primaryMonitor: String(Quickshell.env("QS_MONITOR") || Quickshell.env("QS_PRIMARY_MONITOR") || "DP-2")
    property bool quickSettingsVisible: false

    function compact(text, max) {
        const clean = (text || "").replace(/\s+/g, " ").trim();
        return clean.length > max ? clean.slice(0, max - 1) + "..." : clean;
    }
}
