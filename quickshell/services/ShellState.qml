import Quickshell
import QtQuick

QtObject {
    id: root

    // Post-2026-04-26: Quickshell is the sole shell stack. The cutover
    // properties are kept as scaffolding so module bindings don't all
    // churn in this PR; they're inlined and removed in the PR 4 cleanup.
    readonly property string mode: String(Quickshell.env("SHELL_STACK_MODE") || "full-cutover")
    readonly property bool barCutover: true
    readonly property bool tickerCutover: true
    readonly property bool menuCutover: true
    readonly property bool dockCutover: true
    readonly property bool companionCutover: true
    readonly property bool notificationOwner: true
    readonly property string primaryMonitor: String(Quickshell.env("QS_MONITOR") || Quickshell.env("QS_PRIMARY_MONITOR") || "DP-2")
    property bool quickSettingsVisible: false

    function compact(text, max) {
        const clean = (text || "").replace(/\s+/g, " ").trim();
        return clean.length > max ? clean.slice(0, max - 1) + "..." : clean;
    }
}
