import Quickshell
import Quickshell.Services.SystemTray
import Quickshell.Widgets
import QtQuick

Row {
    id: root

    property var colors
    property var hostWindow

    spacing: 4
    visible: SystemTray.items.values.length > 0

    Repeater {
        model: SystemTray.items

        delegate: Rectangle {
            id: trayItem

            width: 22
            height: 22
            radius: 5
            color: Qt.rgba(root.colors.surface.r, root.colors.surface.g, root.colors.surface.b, 0.74)
            border.width: 1
            border.color: modelData.status === Status.NeedsAttention
                ? Qt.rgba(root.colors.danger.r, root.colors.danger.g, root.colors.danger.b, 0.55)
                : Qt.rgba(root.colors.border.r, root.colors.border.g, root.colors.border.b, 0.70)

            IconImage {
                anchors.centerIn: parent
                width: 16
                height: 16
                implicitSize: 16
                source: Quickshell.iconPath(modelData.icon || "", true)
            }

            MouseArea {
                anchors.fill: parent
                acceptedButtons: Qt.LeftButton | Qt.RightButton | Qt.MiddleButton
                cursorShape: Qt.PointingHandCursor
                onClicked: mouse => {
                    if (mouse.button === Qt.RightButton && modelData.hasMenu) {
                        modelData.display(root.hostWindow, trayItem.x, trayItem.y + trayItem.height);
                    } else if (mouse.button === Qt.MiddleButton) {
                        modelData.secondaryActivate();
                    } else {
                        modelData.activate();
                    }
                }
            }
        }
    }
}
