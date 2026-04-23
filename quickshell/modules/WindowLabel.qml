import Quickshell
import Quickshell.Wayland
import QtQuick

PanelWindow {
    id: panel

    property var screenModel
    property var colors
    property var shellState
    property var windowFocus

    screen: screenModel
    anchors { top: true; left: true; right: true }
    margins { top: 34 }
    implicitHeight: 28
    exclusiveZone: 0
    color: "transparent"
    visible: shellState && shellState.companionCutover && windowFocus && windowFocus.active && windowFocus.labelVisible

    Rectangle {
        anchors.centerIn: parent
        width: Math.min(parent.width - 80, label.implicitWidth + 48)
        height: 24
        radius: 7
        color: colors.panelStrong
        border.color: colors.primary
        border.width: 1
        opacity: 0.92

        Rectangle {
            anchors {
                left: parent.left
                right: parent.right
                top: parent.top
            }
            height: 1
            color: colors.secondary
            opacity: 0.55
        }

        Text {
            id: label
            anchors.centerIn: parent
            text: windowFocus ? windowFocus.labelText : ""
            width: Math.min(920, parent.parent.width - 128)
            horizontalAlignment: Text.AlignHCenter
            elide: Text.ElideRight
            color: colors.fg
            font.family: "Maple Mono NF CN"
            font.pixelSize: 11
            font.bold: true
        }
    }
}
