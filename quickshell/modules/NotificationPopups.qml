import Quickshell
import Quickshell.Wayland
import QtQuick

PanelWindow {
    id: panel

    property var screenModel
    property var colors
    property var notifications

    visible: notifications.ownerEnabled && !notifications.dnd && notifications.popupEntries.length > 0 && !notifications.centerVisible
    screen: screenModel
    anchors { top: true; right: true }
    margins { top: 42; right: 12 }
    implicitWidth: 360
    implicitHeight: popupColumn.implicitHeight
    color: "transparent"

    Column {
        id: popupColumn
        width: parent.width
        spacing: 8

        Repeater {
            model: notifications.popupEntries

            delegate: Rectangle {
                width: popupColumn.width
                height: Math.max(64, bodyText.implicitHeight + 34)
                radius: 12
                color: Qt.rgba(colors.panel.r, colors.panel.g, colors.panel.b, 0.96)
                border.width: 1
                border.color: modelData.urgency === "critical"
                    ? Qt.rgba(colors.danger.r, colors.danger.g, colors.danger.b, 0.72)
                    : Qt.rgba(colors.primary.r, colors.primary.g, colors.primary.b, 0.46)

                Text {
                    id: titleText
                    anchors {
                        left: parent.left
                        leftMargin: 12
                        right: parent.right
                        rightMargin: 12
                        top: parent.top
                        topMargin: 10
                    }
                    text: notifications.compact((modelData.app ? modelData.app + "  " : "") + (modelData.summary || ""), 58)
                    color: modelData.urgency === "critical" ? colors.danger : colors.secondary
                    font.family: "Maple Mono NF CN"
                    font.pixelSize: 12
                    font.bold: true
                    elide: Text.ElideRight
                }

                Text {
                    id: bodyText
                    anchors {
                        left: parent.left
                        leftMargin: 12
                        right: parent.right
                        rightMargin: 12
                        top: titleText.bottom
                        topMargin: 5
                    }
                    text: notifications.compact(modelData.body || modelData.text || "", 160)
                    color: colors.fg
                    font.family: "Maple Mono NF CN"
                    font.pixelSize: 10
                    wrapMode: Text.WordWrap
                    maximumLineCount: 2
                    elide: Text.ElideRight
                }

                MouseArea {
                    anchors.fill: parent
                    cursorShape: Qt.PointingHandCursor
                    onClicked: {
                        const copy = notifications.popupEntries.slice();
                        copy.splice(index, 1);
                        notifications.popupEntries = copy;
                    }
                }
            }
        }
    }
}
