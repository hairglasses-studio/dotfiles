// Quickshell pilot shell — staged bar and ticker replacement surface.
//
// Managed in parallel with ironbar/swaync/keybind-ticker so migration can
// advance without removing the stable desktop stack.

import Quickshell
import Quickshell.Io
import Quickshell.Wayland
import QtQuick
import QtQuick.Layouts
import "styles" as Theme

ShellRoot {
    id: root

    Theme.Colors { id: colors }

    property var targetScreen: Quickshell.screens.find(s => s.name === "DP-2") || Quickshell.screens[0]
    property string workspacesText: "..."
    property string focusedTitle: ""
    property string volumeText: ""
    property string networkText: ""
    property string gpuText: ""
    property string updatesText: ""
    property string weatherText: ""
    property string tickerBridgePath: Quickshell.shellDir + "/../scripts/ticker-bridge.py"
    property string notificationBridgePath: Quickshell.shellDir + "/../scripts/notification-bridge.py"
    property var tickerStreams: ["keybinds"]
    property int tickerStreamIndex: 0
    property string tickerStream: "keybinds"
    property string tickerText: "loading ticker stream..."
    property string tickerStatus: "KEYBINDS"
    property int tickerRefreshMs: 300000
    property int notificationCount: 0
    property int notificationCriticalCount: 0
    property bool notificationDnd: false
    property string notificationText: ""
    property var notificationEntries: []
    property bool notificationCenterVisible: false

    function compact(text, max) {
        const clean = (text || "").replace(/\s+/g, " ").trim();
        return clean.length > max ? clean.slice(0, max - 1) + "..." : clean;
    }

    function refresh(proc) {
        if (!proc.running) proc.exec(proc.command);
    }

    function applyTickerPayload(data) {
        try {
            const payload = JSON.parse(data);
            if (!payload.ok) {
                root.tickerStatus = "TICKER ERR";
                root.tickerText = payload.error || "ticker bridge failed";
                root.tickerRefreshMs = 30000;
                return;
            }
            root.tickerStream = payload.stream || root.tickerStream;
            root.tickerStatus = root.tickerStream.toUpperCase();
            root.tickerText = root.compact(payload.text || payload.markup || "", 5000);
            root.tickerRefreshMs = Math.max(5000, (payload.refresh || 300) * 1000);
        } catch (error) {
            root.tickerStatus = "TICKER JSON";
            root.tickerText = root.compact(String(error), 240);
            root.tickerRefreshMs = 30000;
        }
    }

    function applyTickerListPayload(data) {
        try {
            const payload = JSON.parse(data);
            if (payload.ok && payload.streams && payload.streams.length > 0) {
                root.tickerStreams = payload.streams;
                const idx = root.tickerStreams.indexOf(root.tickerStream);
                root.tickerStreamIndex = idx >= 0 ? idx : 0;
                root.tickerStream = root.tickerStreams[root.tickerStreamIndex];
            }
        } catch (_error) {
            // Keep the current stream if bridge discovery is temporarily noisy.
        }
    }

    function advanceTickerStream() {
        if (root.tickerStreams.length <= 1) return;
        root.tickerStreamIndex = (root.tickerStreamIndex + 1) % root.tickerStreams.length;
        root.tickerStream = root.tickerStreams[root.tickerStreamIndex];
        root.refresh(tickerProc);
    }

    function applyNotificationPayload(data) {
        try {
            const payload = JSON.parse(data);
            if (!payload.ok) {
                root.notificationText = "notification bridge failed";
                return;
            }
            root.notificationCount = payload.daemon_count === null || payload.daemon_count === undefined ? (payload.visible || 0) : payload.daemon_count;
            root.notificationCriticalCount = payload.critical || 0;
            root.notificationDnd = payload.dnd === true;
            root.notificationEntries = payload.entries || [];
            const latest = payload.latest || {};
            const app = latest.app ? latest.app + ": " : "";
            root.notificationText = root.compact(app + (latest.text || latest.summary || ""), 48);
        } catch (error) {
            root.notificationText = root.compact(String(error), 48);
        }
    }

    onTickerTextChanged: tickerSweep.restart()

    // ── Process-backed pilot modules ────────────────────────────────
    Process {
        id: workspacesProc
        command: ["bash", "-lc", "active=$(hyprctl -j activeworkspace 2>/dev/null | jq -r '.id // 0'); hyprctl -j workspaces 2>/dev/null | jq -r --arg active \"$active\" 'sort_by(.id) | map(if ((.id | tostring) == $active) then \"[\" + (.name | tostring) + \"]\" else (.name | tostring) end) | join(\"  \")'"]
        stdout: SplitParser { onRead: data => root.workspacesText = root.compact(data, 80) }
    }

    Process {
        id: focusedProc
        command: ["bash", "-lc", "hyprctl -j activewindow 2>/dev/null | jq -r '.title // empty'"]
        stdout: SplitParser { onRead: data => root.focusedTitle = root.compact(data, 48) }
    }

    Process {
        id: volumeProc
        command: ["bash", "-lc", "wpctl get-volume @DEFAULT_AUDIO_SINK@ 2>/dev/null | awk '{printf \"%d%%\", $2 * 100}'"]
        stdout: SplitParser { onRead: data => root.volumeText = root.compact(data, 12) }
    }

    Process {
        id: networkProc
        command: ["bash", "-lc", "ip -j route get 1.1.1.1 2>/dev/null | jq -r '.[0].dev // empty'"]
        stdout: SplitParser { onRead: data => root.networkText = root.compact(data, 16) }
    }

    Process {
        id: gpuProc
        command: ["bash", "-lc", "cat /tmp/bar-gpu.txt 2>/dev/null || true"]
        stdout: SplitParser { onRead: data => root.gpuText = root.compact(data, 18) }
    }

    Process {
        id: updatesProc
        command: ["bash", "-lc", "cat /tmp/bar-updates.txt 2>/dev/null || true"]
        stdout: SplitParser { onRead: data => root.updatesText = root.compact(data, 16) }
    }

    Process {
        id: weatherProc
        command: ["bash", "-lc", "cat /tmp/bar-weather.txt 2>/dev/null || true"]
        stdout: SplitParser { onRead: data => root.weatherText = root.compact(data, 24) }
    }

    Process {
        id: tickerProc
        command: ["python3", root.tickerBridgePath, "--stream", root.tickerStream, "--once"]
        stdout: SplitParser { onRead: data => root.applyTickerPayload(data) }
        stderr: SplitParser {
            onRead: data => {
                root.tickerStatus = "TICKER STDERR";
                root.tickerText = root.compact(data, 240);
            }
        }
    }

    Process {
        id: tickerListProc
        command: ["python3", root.tickerBridgePath, "--list"]
        stdout: SplitParser { onRead: data => root.applyTickerListPayload(data) }
    }

    Process {
        id: notificationProc
        command: ["python3", root.notificationBridgePath, "--limit", "12"]
        stdout: SplitParser { onRead: data => root.applyNotificationPayload(data) }
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
            root.refresh(gpuProc);
            root.refresh(updatesProc);
            root.refresh(weatherProc);
        }
    }

    Timer {
        interval: root.tickerRefreshMs
        running: true
        repeat: true
        onTriggered: root.refresh(tickerProc)
    }

    Timer {
        interval: 60000
        running: true
        repeat: true
        onTriggered: root.advanceTickerStream()
    }

    Timer {
        interval: 300000
        running: true
        repeat: true
        onTriggered: root.refresh(tickerListProc)
    }

    Timer {
        interval: 15000
        running: true
        repeat: true
        onTriggered: root.refresh(notificationProc)
    }

    Component.onCompleted: {
        refresh(workspacesProc);
        refresh(focusedProc);
        refresh(volumeProc);
        refresh(networkProc);
        refresh(gpuProc);
        refresh(updatesProc);
        refresh(weatherProc);
        refresh(tickerListProc);
        refresh(tickerProc);
        refresh(notificationProc);
    }

    component Badge: Rectangle {
        id: badge
        property alias text: label.text
        property color accent: colors.primary
        signal clicked()
        height: 22
        radius: 5
        border.width: 1
        border.color: Qt.rgba(accent.r, accent.g, accent.b, 0.42)
        color: Qt.rgba(colors.surface.r, colors.surface.g, colors.surface.b, 0.74)
        implicitWidth: label.implicitWidth + 18

        Text {
            id: label
            anchors.centerIn: parent
            color: badge.accent
            font.family: "Maple Mono NF CN"
            font.pixelSize: 11
            font.bold: true
        }

        MouseArea {
            anchors.fill: parent
            cursorShape: Qt.PointingHandCursor
            onClicked: badge.clicked()
        }
    }

    PanelWindow {
        id: topPanel
        screen: root.targetScreen
        anchors { top: true; left: true; right: true }
        implicitHeight: 28
        color: "transparent"

        Rectangle {
            anchors.fill: parent
            color: colors.panelStrong
            opacity: 0.88
        }

        // Qt 6 ShaderEffect requires precompiled .qsb files. Keep the service
        // bootstrap warning-free with a scene-graph gradient sweep for now.
        Rectangle {
            id: topSweep
            y: 0
            height: parent.height
            width: parent.width * 0.45
            opacity: 0.24
            gradient: Gradient {
                orientation: Gradient.Horizontal
                GradientStop { position: 0.00; color: "transparent" }
                GradientStop { position: 0.34; color: colors.primary }
                GradientStop { position: 0.66; color: colors.secondary }
                GradientStop { position: 1.00; color: "transparent" }
            }
            NumberAnimation on x {
                from: -topSweep.width
                to: topPanel.width
                duration: 8000
                loops: Animation.Infinite
                easing.type: Easing.InOutSine
            }
        }

        Row {
            anchors {
                left: parent.left
                leftMargin: 28
                verticalCenter: parent.verticalCenter
            }
            spacing: 6
            Badge { text: root.workspacesText; accent: colors.primary }
            Text {
                color: colors.blue
                text: root.focusedTitle
                width: 300
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
            Badge { visible: root.gpuText.length > 0; text: root.gpuText; accent: colors.warning }
            Badge { visible: root.updatesText.length > 0; text: root.updatesText; accent: colors.warning }
            Badge {
                visible: root.notificationCount > 0 || root.notificationDnd
                text: root.notificationDnd ? "DND" : ((root.notificationCriticalCount > 0 ? "!" : "N") + " " + root.notificationCount)
                accent: root.notificationDnd ? colors.muted : (root.notificationCriticalCount > 0 ? colors.danger : colors.secondary)
                onClicked: root.notificationCenterVisible = !root.notificationCenterVisible
            }
            Badge { visible: root.volumeText.length > 0; text: "VOL " + root.volumeText; accent: colors.primary }
            Badge { visible: root.weatherText.length > 0; text: root.weatherText; accent: colors.tertiary }
            Badge { visible: root.networkText.length > 0; text: root.networkText; accent: colors.blue }
        }

        Rectangle {
            id: pulse
            width: 8
            height: 8
            radius: 4
            anchors { verticalCenter: parent.verticalCenter; left: parent.left; leftMargin: 10 }
            color: colors.primary
            SequentialAnimation on opacity {
                loops: Animation.Infinite
                NumberAnimation { from: 1.0; to: 0.35; duration: 1200; easing.type: Easing.InOutSine }
                NumberAnimation { from: 0.35; to: 1.0; duration: 1200; easing.type: Easing.InOutSine }
            }
        }
    }

    PanelWindow {
        id: notificationPanel
        visible: root.notificationCenterVisible && root.notificationEntries.length > 0
        screen: root.targetScreen
        anchors { top: true; right: true }
        margins { top: 34; right: 10 }
        implicitWidth: 430
        implicitHeight: Math.min(360, notificationColumn.implicitHeight + 24)
        color: "transparent"

        Rectangle {
            anchors.fill: parent
            radius: 10
            color: Qt.rgba(colors.surface.r, colors.surface.g, colors.surface.b, 0.96)
            border.width: 1
            border.color: Qt.rgba(colors.secondary.r, colors.secondary.g, colors.secondary.b, 0.46)
        }

        Column {
            id: notificationColumn
            anchors {
                left: parent.left
                right: parent.right
                top: parent.top
                margins: 12
            }
            spacing: 8

            Row {
                spacing: 8
                Badge {
                    text: root.notificationDnd ? "DND ENABLED" : "NOTIFICATIONS"
                    accent: root.notificationDnd ? colors.muted : colors.secondary
                    onClicked: root.notificationCenterVisible = false
                }
                Text {
                    text: root.notificationEntries.length + " recent"
                    color: colors.muted
                    font.family: "Maple Mono NF CN"
                    font.pixelSize: 11
                    anchors.verticalCenter: parent.verticalCenter
                }
            }

            Repeater {
                model: root.notificationEntries.slice(0, 5)
                delegate: Rectangle {
                    width: notificationColumn.width
                    height: Math.max(48, bodyText.implicitHeight + 22)
                    radius: 7
                    color: Qt.rgba(colors.bg.r, colors.bg.g, colors.bg.b, 0.82)
                    border.width: 1
                    border.color: modelData.urgency === "critical"
                        ? Qt.rgba(colors.danger.r, colors.danger.g, colors.danger.b, 0.55)
                        : Qt.rgba(colors.primary.r, colors.primary.g, colors.primary.b, 0.20)

                    Text {
                        id: titleText
                        anchors {
                            left: parent.left
                            leftMargin: 10
                            right: parent.right
                            rightMargin: 10
                            top: parent.top
                            topMargin: 7
                        }
                        text: root.compact((modelData.app ? modelData.app + "  " : "") + (modelData.summary || ""), 72)
                        color: modelData.urgency === "critical" ? colors.danger : colors.secondary
                        font.family: "Maple Mono NF CN"
                        font.pixelSize: 11
                        font.bold: true
                        elide: Text.ElideRight
                    }

                    Text {
                        id: bodyText
                        anchors {
                            left: parent.left
                            leftMargin: 10
                            right: parent.right
                            rightMargin: 10
                            top: titleText.bottom
                            topMargin: 4
                        }
                        text: root.compact(modelData.body || modelData.text || "", 180)
                        color: colors.fg
                        font.family: "Maple Mono NF CN"
                        font.pixelSize: 10
                        wrapMode: Text.WordWrap
                        maximumLineCount: 2
                        elide: Text.ElideRight
                    }
                }
            }
        }
    }

    PanelWindow {
        id: tickerPanel
        screen: root.targetScreen
        // Top-offset during pilot mode so it does not collide with the live
        // Python keybind ticker on the bottom edge.
        anchors { top: true; left: true; right: true }
        margins { top: 30 }
        implicitHeight: 30
        color: "transparent"

        Rectangle {
            anchors.fill: parent
            color: colors.bg
            opacity: 0.92
        }

        Rectangle {
            anchors {
                left: parent.left
                right: parent.right
                top: parent.top
            }
            height: 1
            color: colors.primary
            opacity: 0.74
        }

        Rectangle {
            id: streamBadge
            anchors {
                left: parent.left
                leftMargin: 8
                verticalCenter: parent.verticalCenter
            }
            width: Math.max(92, streamLabel.implicitWidth + 18)
            height: 22
            radius: 4
            color: colors.primary
            opacity: 0.95

            Text {
                id: streamLabel
                anchors.centerIn: parent
                text: root.tickerStatus
                color: colors.bg
                font.family: "Maple Mono NF CN"
                font.pixelSize: 11
                font.bold: true
            }

            MouseArea {
                anchors.fill: parent
                cursorShape: Qt.PointingHandCursor
                onClicked: root.advanceTickerStream()
            }
        }

        Item {
            id: tickerClip
            anchors {
                left: streamBadge.right
                leftMargin: 12
                right: parent.right
                rightMargin: 8
                top: parent.top
                bottom: parent.bottom
            }
            clip: true

            Text {
                id: tickerLabel
                x: tickerClip.width
                anchors.verticalCenter: parent.verticalCenter
                text: root.tickerText
                textFormat: Text.PlainText
                color: colors.fg
                font.family: "Maple Mono NF CN"
                font.pixelSize: 13
                font.bold: true
            }

            NumberAnimation {
                id: tickerSweep
                target: tickerLabel
                property: "x"
                from: tickerClip.width
                to: -tickerLabel.implicitWidth
                duration: Math.max(18000, tickerLabel.implicitWidth * 34)
                loops: Animation.Infinite
                running: true
                easing.type: Easing.Linear
            }
        }

        Text {
            visible: root.notificationText.length > 0
            anchors {
                right: parent.right
                rightMargin: 14
                bottom: parent.top
                bottomMargin: 4
            }
            text: root.notificationText
            color: root.notificationCriticalCount > 0 ? colors.danger : colors.secondary
            font.family: "Maple Mono NF CN"
            font.pixelSize: 10
            font.bold: true
            opacity: 0.85
        }
    }
}
