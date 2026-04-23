import Quickshell
import Quickshell.Wayland
import QtQuick
import "../components" as Components

PanelWindow {
    id: panel

    property var screenModel
    property var colors
    property var shellState
    property var barData
    property var notifications

    visible: shellState.quickSettingsVisible
    screen: screenModel
    anchors { top: true; right: true }
    margins { top: 36; right: 470 }
    implicitWidth: 330
    implicitHeight: quickColumn.implicitHeight + 24
    color: "transparent"
    focusable: true

    Rectangle {
        anchors.fill: parent
        radius: 12
        color: Qt.rgba(colors.surface.r, colors.surface.g, colors.surface.b, 0.97)
        border.width: 1
        border.color: Qt.rgba(colors.primary.r, colors.primary.g, colors.primary.b, 0.42)
    }

    Column {
        id: quickColumn
        anchors {
            left: parent.left
            right: parent.right
            top: parent.top
            margins: 12
        }
        spacing: 10

        Row {
            spacing: 8
            Components.Badge {
                colors: panel.colors
                text: "SHELL " + shellState.mode.toUpperCase()
                accent: colors.primary
                onClicked: shellState.quickSettingsVisible = false
            }
            Components.Badge {
                colors: panel.colors
                text: notifications.dnd ? "DND" : "LOUD"
                accent: notifications.dnd ? colors.muted : colors.secondary
                onClicked: notifications.toggleDnd()
            }
        }

        Text {
            width: parent.width
            text: [barData.systemText, barData.fleetText, barData.claudeText].filter(v => v && v.length > 0).join("  /  ")
            color: colors.fg
            font.family: "Maple Mono NF CN"
            font.pixelSize: 11
            wrapMode: Text.WordWrap
        }

        Flow {
            width: parent.width
            spacing: 8

            Components.PanelButton {
                colors: panel.colors
                text: "RELOAD"
                accent: colors.primary
                onClicked: Quickshell.execDetached(["bash", "-lc", "$HOME/hairglasses-studio/dotfiles/scripts/rice-reload.sh"])
            }
            Components.PanelButton {
                colors: panel.colors
                text: "THEME"
                accent: colors.secondary
                onClicked: Quickshell.execDetached(["bash", "-lc", "$HOME/.local/bin/theme-sync"])
            }
            Components.PanelButton {
                colors: panel.colors
                text: "WALL"
                accent: colors.tertiary
                onClicked: Quickshell.execDetached(["bash", "-lc", "$HOME/hairglasses-studio/dotfiles/scripts/wallpaper-cycle.sh random"])
            }
            Components.PanelButton {
                colors: panel.colors
                text: "NIGHT"
                accent: colors.warning
                onClicked: Quickshell.execDetached(["bash", "-lc", "pkill hyprsunset || hyprsunset --temperature 3500"])
            }
            Components.PanelButton {
                colors: panel.colors
                text: "CLEAR NOTIFS"
                accent: colors.danger
                onClicked: notifications.closeAll()
            }
        }
    }
}
