import Quickshell
import Quickshell.Wayland
import QtQuick
import "../components" as Components

PanelWindow {
    id: panel

    property var screenModel
    property var colors
    property var notifications

    visible: notifications.centerVisible
    screen: screenModel
    anchors { top: true; right: true }
    margins { top: 36; right: 10 }
    implicitWidth: 450
    implicitHeight: Math.min(520, notificationColumn.implicitHeight + 24)
    color: "transparent"
    focusable: true

    Rectangle {
        anchors.fill: parent
        radius: 12
        color: Qt.rgba(colors.surface.r, colors.surface.g, colors.surface.b, 0.97)
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

            Components.Badge {
                colors: panel.colors
                text: notifications.dnd ? "DND ENABLED" : "NOTIFICATIONS"
                accent: notifications.dnd ? colors.muted : colors.secondary
                onClicked: notifications.centerVisible = false
            }

            Components.Badge {
                colors: panel.colors
                text: notifications.dnd ? "DND OFF" : "DND ON"
                accent: notifications.dnd ? colors.tertiary : colors.muted
                onClicked: notifications.toggleDnd()
            }

            Components.Badge {
                colors: panel.colors
                text: "CLEAR"
                accent: colors.danger
                onClicked: notifications.clearHistory()
            }

            Text {
                text: notifications.entries.length + " recent"
                color: colors.muted
                font.family: "Maple Mono NF CN"
                font.pixelSize: 11
                anchors.verticalCenter: parent.verticalCenter
            }
        }

        Repeater {
            model: notifications.entries.slice(0, 8)

            delegate: Rectangle {
                width: notificationColumn.width
                height: Math.max(50, bodyText.implicitHeight + 24)
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
                    text: notifications.compact((modelData.app ? modelData.app + "  " : "") + (modelData.summary || ""), 76)
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
                    text: notifications.compact(modelData.body || modelData.text || "", 210)
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
