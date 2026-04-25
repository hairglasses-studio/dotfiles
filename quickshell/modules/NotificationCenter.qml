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
                id: entryRoot
                property var entry: modelData
                property var entryActions: (entry && entry.actions) || []
                property bool hasLive: entry && notifications.liveNotifications && notifications.liveNotifications.hasOwnProperty(entry.id)
                width: notificationColumn.width
                height: Math.max(50, bodyText.implicitHeight + 24) + (entryActions.length > 0 && hasLive ? 30 : 0)
                radius: 7
                color: Qt.rgba(colors.bg.r, colors.bg.g, colors.bg.b, 0.82)
                border.width: 1
                border.color: entry.urgency === "critical"
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
                    text: notifications.compact((entry.app ? entry.app + "  " : "") + (entry.summary || ""), 76)
                    color: entry.urgency === "critical" ? colors.danger : colors.secondary
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
                    text: notifications.compact(entry.body || entry.text || "", 210)
                    color: colors.fg
                    font.family: "Maple Mono NF CN"
                    font.pixelSize: 10
                    wrapMode: Text.WordWrap
                    maximumLineCount: 2
                    elide: Text.ElideRight
                }

                Row {
                    visible: entryRoot.entryActions.length > 0 && entryRoot.hasLive
                    anchors {
                        left: parent.left
                        leftMargin: 10
                        right: parent.right
                        rightMargin: 10
                        bottom: parent.bottom
                        bottomMargin: 5
                    }
                    spacing: 5

                    Repeater {
                        model: entryRoot.entryActions
                        delegate: Components.Badge {
                            colors: panel.colors
                            text: (modelData.text || modelData.id || "").toUpperCase()
                            accent: panel.colors.tertiary
                            fontSize: 10
                            onClicked: notifications.invokeAction(entryRoot.entry.id, modelData.id)
                        }
                    }
                }
            }
        }
    }
}
