import Quickshell
import Quickshell.Wayland
import QtQuick
import "../components" as Components

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
                id: popupRoot
                property var entry: modelData
                property int entryIndex: index
                property var entryActions: (entry && entry.actions) || []
                width: popupColumn.width
                height: Math.max(64, bodyText.implicitHeight + 34) + (entryActions.length > 0 ? 32 : 0)
                radius: 12
                color: Qt.rgba(colors.panel.r, colors.panel.g, colors.panel.b, 0.96)
                border.width: 1
                border.color: entry.urgency === "critical"
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
                    text: notifications.compact((entry.app ? entry.app + "  " : "") + (entry.summary || ""), 58)
                    color: entry.urgency === "critical" ? colors.danger : colors.secondary
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
                    text: notifications.compact(entry.body || entry.text || "", 160)
                    color: colors.fg
                    font.family: "Maple Mono NF CN"
                    font.pixelSize: 10
                    wrapMode: Text.WordWrap
                    maximumLineCount: 2
                    elide: Text.ElideRight
                }

                Row {
                    visible: popupRoot.entryActions.length > 0
                    anchors {
                        left: parent.left
                        leftMargin: 12
                        right: parent.right
                        rightMargin: 12
                        bottom: parent.bottom
                        bottomMargin: 6
                    }
                    spacing: 6

                    Repeater {
                        model: popupRoot.entryActions
                        delegate: Components.Badge {
                            colors: panel.colors
                            text: (modelData.text || modelData.id || "").toUpperCase()
                            accent: panel.colors.tertiary
                            onClicked: notifications.invokeAction(popupRoot.entry.id, modelData.id)
                        }
                    }
                }

                MouseArea {
                    anchors {
                        left: parent.left
                        right: parent.right
                        top: parent.top
                        bottom: popupRoot.entryActions.length > 0 ? bodyText.bottom : parent.bottom
                    }
                    cursorShape: Qt.PointingHandCursor
                    onClicked: {
                        const copy = notifications.popupEntries.slice();
                        copy.splice(popupRoot.entryIndex, 1);
                        notifications.popupEntries = copy;
                    }
                }
            }
        }
    }
}
