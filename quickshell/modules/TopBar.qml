import Quickshell
import Quickshell.Io
import Quickshell.Wayland
import QtQuick
import "../components" as Components

PanelWindow {
    id: panel

    property var modelData
    property var primaryScreen
    property var screenModel: modelData
    property bool primary: screenModel === primaryScreen
    property var colors
    property var shellState
    property var barData
    property var notifications
    // Anchor edge — "top" (default) renders the bar above the workspace,
    // "bottom" renders below. ironbar's pre-cutover layout had a top bar
    // on DP-3 and a bottom bar on DP-2; once `bar-cutover` lands, the
    // shell.qml Variants block uses both anchors so the same surfaces
    // are owned by Quickshell.
    property string anchor: "top"

    // Per-monitor workspaces text. Each TopBar variant shows its own
    // monitor's slice (1-5 / 6-10 / 11-15 with split-monitor-workspaces),
    // not the global list barData.workspacesText would otherwise return.
    property string workspacesText: "..."

    screen: screenModel
    anchors {
        top: anchor === "top"
        bottom: anchor === "bottom"
        left: true
        right: true
    }
    implicitHeight: 30
    exclusiveZone: 30
    color: "transparent"

    // Filter workspaces by .monitor field so each TopBar instance only
    // shows its own monitor's set. Refreshed every 1s (workspace state
    // changes are user-driven, not high-frequency).
    Process {
        id: monitorWorkspacesProc
        command: ["bash", "-lc",
            "active=$(hyprctl -j activeworkspace 2>/dev/null | jq -r '.id // 0'); " +
            "hyprctl -j workspaces 2>/dev/null | jq -r --arg active \"$active\" --arg mon \"" + (screenModel ? screenModel.name : "") + "\" " +
            "'sort_by(.id) | map(select(.monitor == $mon)) | map(if ((.id | tostring) == $active) then \"[\" + (.name | tostring) + \"]\" else (.name | tostring) end) | join(\"  \")'"
        ]
        stdout: SplitParser {
            onRead: data => {
                const clean = (data || "").replace(/\s+/g, " ").trim();
                panel.workspacesText = clean.length > 0 ? clean : "—";
            }
        }
    }
    Timer {
        interval: 1000
        repeat: true
        running: true
        triggeredOnStart: true
        onTriggered: if (!monitorWorkspacesProc.running) monitorWorkspacesProc.exec(monitorWorkspacesProc.command);
    }

    Rectangle {
        anchors.fill: parent
        color: colors.panelStrong
        opacity: primary ? 0.91 : 0.82
    }

    Rectangle {
        id: topSweep
        y: 0
        height: parent.height
        width: parent.width * 0.45
        opacity: primary ? 0.25 : 0.12
        gradient: Gradient {
            orientation: Gradient.Horizontal
            GradientStop { position: 0.00; color: "transparent" }
            GradientStop { position: 0.34; color: colors.primary }
            GradientStop { position: 0.66; color: colors.secondary }
            GradientStop { position: 1.00; color: "transparent" }
        }
        NumberAnimation on x {
            from: -topSweep.width
            to: panel.width
            duration: 8000
            loops: Animation.Infinite
            easing.type: Easing.InOutSine
        }
    }

    Row {
        anchors {
            left: parent.left
            leftMargin: 10
            verticalCenter: parent.verticalCenter
        }
        spacing: 6

        Rectangle {
            width: 8
            height: 8
            radius: 4
            anchors.verticalCenter: parent.verticalCenter
            color: primary ? colors.primary : colors.muted
            SequentialAnimation on opacity {
                loops: Animation.Infinite
                NumberAnimation { from: 1.0; to: 0.35; duration: 1200; easing.type: Easing.InOutSine }
                NumberAnimation { from: 0.35; to: 1.0; duration: 1200; easing.type: Easing.InOutSine }
            }
        }

        Components.Badge {
            colors: panel.colors
            text: panel.workspacesText
            accent: colors.primary
        }

        Text {
            color: colors.blue
            text: barData.focusedTitle
            width: primary ? 360 : 220
            elide: Text.ElideRight
            font.family: "Maple Mono NF CN"
            font.pixelSize: 11
        }
    }

    Text {
        id: clock
        anchors.centerIn: parent
        color: colors.warning
        text: Qt.formatDateTime(new Date(), "HH:mm")
        font.family: "Maple Mono NF CN"
        font.pixelSize: 14
        font.bold: true

        MouseArea {
            anchors.fill: parent
            cursorShape: Qt.PointingHandCursor
            onClicked: shellState.quickSettingsVisible = !shellState.quickSettingsVisible
        }

        Timer {
            interval: 1000
            running: true
            repeat: true
            onTriggered: clock.text = Qt.formatDateTime(new Date(), "HH:mm")
        }
    }

    Row {
        anchors {
            right: parent.right
            rightMargin: 8
            verticalCenter: parent.verticalCenter
        }
        spacing: 6

        Components.Badge { visible: primary && barData.fleetText.length > 0; colors: panel.colors; text: barData.fleetText; accent: colors.secondary }
        Components.Badge { visible: primary && barData.claudeText.length > 0; colors: panel.colors; text: barData.claudeText; accent: colors.tertiary }
        Components.Badge { visible: primary && barData.shaderText.length > 0; colors: panel.colors; text: "SH " + barData.shaderText; accent: colors.blue }
        Components.MediaBadges { visible: primary; colors: panel.colors; fallbackText: barData.mediaText }
        Components.Badge { visible: primary && barData.systemText.length > 0; colors: panel.colors; text: barData.systemText; accent: colors.warning }
        Components.Badge { visible: barData.gpuText.length > 0; colors: panel.colors; text: barData.gpuText; accent: colors.warning }
        Components.Badge { visible: primary && barData.retroarchText.length > 0; colors: panel.colors; text: barData.retroarchText; accent: colors.tertiary }
        Components.Badge { visible: primary && barData.updatesText.length > 0; colors: panel.colors; text: barData.updatesText; accent: colors.warning }
        Components.Badge { visible: primary && barData.bluetoothText.length > 0; colors: panel.colors; text: barData.bluetoothText; accent: colors.blue }
        Components.Badge { visible: primary && barData.mxText.length > 0; colors: panel.colors; text: barData.mxText; accent: colors.warning }
        Components.Badge {
            visible: notifications.notificationCount > 0 || notifications.dnd
            colors: panel.colors
            text: notifications.dnd ? "DND" : ((notifications.criticalCount > 0 ? "!" : "N") + " " + notifications.notificationCount)
            accent: notifications.dnd ? colors.muted : (notifications.criticalCount > 0 ? colors.danger : colors.secondary)
            onClicked: notifications.centerVisible = !notifications.centerVisible
        }
        Components.Badge { visible: barData.volumeText.length > 0; colors: panel.colors; text: "VOL " + barData.volumeText; accent: colors.primary }
        Components.Badge { visible: primary && barData.weatherText.length > 0; colors: panel.colors; text: barData.weatherText; accent: colors.tertiary }
        Components.Badge { visible: barData.networkText.length > 0; colors: panel.colors; text: barData.networkText; accent: colors.blue }
        Components.Tray { visible: primary; colors: panel.colors; hostWindow: panel }
    }
}
