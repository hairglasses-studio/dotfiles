import Quickshell
import Quickshell.Wayland
import QtQuick

PanelWindow {
    id: panel

    property var screenModel
    property var colors
    property var shellState
    property var lyrics
    property var windowFocus

    screen: screenModel
    anchors { top: true; left: true; right: true }
    margins { top: 34 }
    implicitHeight: 28
    exclusiveZone: 0
    color: "transparent"
    visible: shellState && shellState.companionCutover && lyrics && lyrics.playing && lyrics.hasText && !(windowFocus && windowFocus.labelVisible)

    Rectangle {
        anchors.centerIn: parent
        width: Math.min(parent.width - 140, lyricText.implicitWidth + 70)
        height: 24
        radius: 7
        color: colors.panelStrong
        border.color: lyrics && lyrics.synced ? colors.tertiary : colors.secondary
        border.width: 1
        opacity: 0.9

        Text {
            anchors {
                left: parent.left
                leftMargin: 12
                verticalCenter: parent.verticalCenter
            }
            text: lyrics && lyrics.synced ? "LRC" : "MPRIS"
            color: lyrics && lyrics.synced ? colors.tertiary : colors.secondary
            font.family: "Maple Mono NF CN"
            font.pixelSize: 9
            font.bold: true
        }

        Text {
            id: lyricText
            anchors.centerIn: parent
            text: lyrics ? lyrics.line : ""
            width: Math.min(940, parent.parent.width - 240)
            horizontalAlignment: Text.AlignHCenter
            elide: Text.ElideRight
            color: lyrics ? lyrics.color : colors.muted
            font.family: "Maple Mono NF CN"
            font.pixelSize: 11
            font.bold: true
        }
    }
}
