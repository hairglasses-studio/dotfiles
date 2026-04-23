import Quickshell
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

    screen: screenModel
    anchors { top: true; left: true; right: true }
    implicitHeight: 30
    exclusiveZone: shellState && shellState.barCutover ? 30 : 0
    color: "transparent"

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
            text: barData.workspacesText
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
        Components.Badge { visible: primary && barData.mediaText.length > 0; colors: panel.colors; text: barData.mediaText; accent: colors.secondary }
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
    }
}
