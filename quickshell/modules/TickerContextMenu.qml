import Quickshell
import Quickshell.Wayland
import QtQuick
import "../components" as Components

PanelWindow {
    id: panel

    property var screenModel
    property var colors
    property var shellState
    property var ticker

    visible: ticker.contextVisible
    screen: screenModel
    anchors { left: true; bottom: true }
    margins { left: 10; bottom: 38 }
    implicitWidth: 620
    implicitHeight: 324
    color: "transparent"
    focusable: true
    aboveWindows: true
    WlrLayershell.layer: WlrLayer.Overlay
    WlrLayershell.keyboardFocus: WlrKeyboardFocus.OnDemand

    Rectangle {
        anchors.fill: parent
        radius: 14
        color: Qt.rgba(colors.surface.r, colors.surface.g, colors.surface.b, 0.98)
        border.width: 1
        border.color: Qt.rgba(colors.primary.r, colors.primary.g, colors.primary.b, 0.42)
    }

    Column {
        anchors.fill: parent
        anchors.margins: 14
        spacing: 10

        Row {
            width: parent.width
            spacing: 8

            Components.Badge {
                colors: panel.colors
                text: ticker.status
                accent: ticker.urgent ? colors.danger : colors.primary
                onClicked: ticker.contextVisible = false
            }

            Components.Badge {
                colors: panel.colors
                text: ticker.paused ? "RESUME" : "PAUSE"
                accent: ticker.paused ? colors.tertiary : colors.warning
                onClicked: ticker.togglePause()
            }

            Components.Badge {
                colors: panel.colors
                text: ticker.shuffle ? "SHUFFLE ON" : "SHUFFLE OFF"
                accent: ticker.shuffle ? colors.secondary : colors.muted
                onClicked: ticker.setShuffleMode("toggle")
            }

            Components.Badge {
                colors: panel.colors
                text: ticker.pinnedStream ? "UNPIN" : "PIN"
                accent: ticker.pinnedStream ? colors.danger : colors.tertiary
                onClicked: ticker.pinToggle()
            }
        }

        Flow {
            width: parent.width
            spacing: 8

            Components.PanelButton { colors: panel.colors; text: "PREV"; accent: colors.primary; onClicked: ticker.prev() }
            Components.PanelButton { colors: panel.colors; text: "NEXT"; accent: colors.primary; onClicked: ticker.next() }
            Components.PanelButton { colors: panel.colors; text: "MAIN"; accent: ticker.playlist === "main" ? colors.tertiary : colors.muted; onClicked: ticker.setPlaylist("main") }
            Components.PanelButton { colors: panel.colors; text: "CODING"; accent: ticker.playlist === "coding" ? colors.tertiary : colors.muted; onClicked: ticker.setPlaylist("coding") }
            Components.PanelButton { colors: panel.colors; text: "FOCUS"; accent: ticker.playlist === "focus" ? colors.tertiary : colors.muted; onClicked: ticker.setPlaylist("focus") }
            Components.PanelButton { colors: panel.colors; text: "RECORDING"; accent: ticker.playlist === "recording" ? colors.tertiary : colors.muted; onClicked: ticker.setPlaylist("recording") }
            Components.PanelButton { colors: panel.colors; text: "LOCK"; accent: ticker.playlist === "lock" ? colors.tertiary : colors.muted; onClicked: ticker.setPlaylist("lock") }
        }

        ListView {
            width: parent.width
            height: 34
            orientation: ListView.Horizontal
            spacing: 8
            clip: true
            model: ticker.streams

            delegate: Rectangle {
                width: Math.max(82, streamName.implicitWidth + 18)
                height: 28
                radius: 8
                color: modelData === ticker.stream
                    ? Qt.rgba(colors.primary.r, colors.primary.g, colors.primary.b, 0.20)
                    : Qt.rgba(colors.bg.r, colors.bg.g, colors.bg.b, 0.72)
                border.width: 1
                border.color: modelData === ticker.stream ? colors.primary : Qt.rgba(colors.border.r, colors.border.g, colors.border.b, 0.48)

                Text {
                    id: streamName
                    anchors.centerIn: parent
                    text: String(modelData).toUpperCase()
                    color: modelData === ticker.stream ? colors.primary : colors.muted
                    font.family: "Maple Mono NF CN"
                    font.pixelSize: 10
                    font.bold: true
                }

                MouseArea {
                    anchors.fill: parent
                    cursorShape: Qt.PointingHandCursor
                    onClicked: ticker.selectStream(modelData)
                }
            }
        }

        Flow {
            width: parent.width
            spacing: 8

            Components.PanelButton { colors: panel.colors; text: "AMBIENT"; accent: ticker.preset === "ambient" ? colors.secondary : colors.muted; onClicked: ticker.setPreset("ambient") }
            Components.PanelButton { colors: panel.colors; text: "CYBERPUNK"; accent: ticker.preset === "cyberpunk" ? colors.secondary : colors.muted; onClicked: ticker.setPreset("cyberpunk") }
            Components.PanelButton { colors: panel.colors; text: "MINIMAL"; accent: ticker.preset === "minimal" ? colors.secondary : colors.muted; onClicked: ticker.setPreset("minimal") }
            Components.PanelButton { colors: panel.colors; text: "CLEAN"; accent: ticker.preset === "clean" ? colors.secondary : colors.muted; onClicked: ticker.setPreset("clean") }
            Components.PanelButton { colors: panel.colors; text: "COPY"; accent: colors.primary; onClicked: ticker.copyCurrent() }
            Components.PanelButton { colors: panel.colors; text: "URGENT OFF"; accent: colors.danger; onClicked: ticker.snoozeUrgent() }
        }

        Text {
            width: parent.width
            height: 72
            text: ticker.text
            color: colors.fg
            font.family: "Maple Mono NF CN"
            font.pixelSize: 11
            wrapMode: Text.WordWrap
            maximumLineCount: 4
            elide: Text.ElideRight
        }
    }
}
