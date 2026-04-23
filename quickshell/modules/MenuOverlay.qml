import Quickshell
import Quickshell.Wayland
import QtQuick
import "../components" as Components

PanelWindow {
    id: panel

    property var screenModel
    property var colors
    property var menus

    visible: menus.active
    screen: screenModel
    anchors { left: true; right: true; top: true; bottom: true }
    exclusiveZone: 0
    color: "transparent"
    focusable: true
    aboveWindows: true
    WlrLayershell.layer: WlrLayer.Overlay
    WlrLayershell.keyboardFocus: WlrKeyboardFocus.Exclusive

    onVisibleChanged: {
        if (visible) queryInput.forceActiveFocus();
    }

    Rectangle {
        anchors.fill: parent
        color: Qt.rgba(0, 0, 0, 0.46)

        MouseArea {
            anchors.fill: parent
            onClicked: menus.close()
        }
    }

    Rectangle {
        id: card
        width: Math.min(940, Math.max(640, panel.width * 0.52))
        height: Math.min(680, Math.max(420, panel.height * 0.62))
        anchors {
            horizontalCenter: parent.horizontalCenter
            top: parent.top
            topMargin: Math.max(70, panel.height * 0.12)
        }
        radius: 18
        color: Qt.rgba(colors.surface.r, colors.surface.g, colors.surface.b, 0.98)
        border.width: 1
        border.color: menus.mode === "power"
            ? Qt.rgba(colors.danger.r, colors.danger.g, colors.danger.b, 0.58)
            : Qt.rgba(colors.primary.r, colors.primary.g, colors.primary.b, 0.44)

        Rectangle {
            anchors.fill: parent
            anchors.margins: 1
            radius: 17
            color: "transparent"
            border.width: 1
            border.color: Qt.rgba(colors.secondary.r, colors.secondary.g, colors.secondary.b, 0.14)
        }

        Column {
            anchors.fill: parent
            anchors.margins: 18
            spacing: 12

            Row {
                width: parent.width
                height: 32
                spacing: 10

                Text {
                    text: menus.title.toUpperCase()
                    color: menus.mode === "power" ? colors.danger : colors.primary
                    font.family: "Maple Mono NF CN"
                    font.pixelSize: 20
                    font.bold: true
                    anchors.verticalCenter: parent.verticalCenter
                }

                Text {
                    text: menus.filteredEntries.length + " results"
                    color: colors.muted
                    font.family: "Maple Mono NF CN"
                    font.pixelSize: 11
                    anchors.verticalCenter: parent.verticalCenter
                }

                Item { width: Math.max(1, parent.width - 260); height: 1 }
            }

            Rectangle {
                width: parent.width
                height: 48
                radius: 12
                color: Qt.rgba(colors.bg.r, colors.bg.g, colors.bg.b, 0.86)
                border.width: 1
                border.color: Qt.rgba(colors.borderStrong.r, colors.borderStrong.g, colors.borderStrong.b, 0.62)

                Text {
                    anchors {
                        left: parent.left
                        leftMargin: 14
                        verticalCenter: parent.verticalCenter
                    }
                    text: ">"
                    color: colors.secondary
                    font.family: "Maple Mono NF CN"
                    font.pixelSize: 18
                    font.bold: true
                }

                TextInput {
                    id: queryInput
                    anchors {
                        left: parent.left
                        leftMargin: 42
                        right: parent.right
                        rightMargin: 14
                        verticalCenter: parent.verticalCenter
                    }
                    text: menus.query
                    color: colors.fg
                    selectionColor: colors.secondary
                    selectedTextColor: colors.bg
                    font.family: "Maple Mono NF CN"
                    font.pixelSize: 16
                    focus: panel.visible
                    clip: true
                    onTextChanged: menus.query = text

                    Keys.onPressed: event => {
                        if (event.key === Qt.Key_Escape) {
                            menus.close();
                            event.accepted = true;
                        } else if (event.key === Qt.Key_Down || event.key === Qt.Key_Tab) {
                            menus.move(1);
                            event.accepted = true;
                        } else if (event.key === Qt.Key_Up || event.key === Qt.Key_Backtab) {
                            menus.move(-1);
                            event.accepted = true;
                        } else if (event.key === Qt.Key_PageDown) {
                            menus.page(1);
                            event.accepted = true;
                        } else if (event.key === Qt.Key_PageUp) {
                            menus.page(-1);
                            event.accepted = true;
                        } else if (event.key === Qt.Key_Return || event.key === Qt.Key_Enter) {
                            menus.activateCurrent();
                            event.accepted = true;
                        }
                    }
                }

                Text {
                    visible: queryInput.text.length === 0
                    anchors {
                        left: queryInput.left
                        verticalCenter: parent.verticalCenter
                    }
                    text: menus.hint
                    color: colors.muted
                    font.family: "Maple Mono NF CN"
                    font.pixelSize: 13
                }
            }

            ListView {
                id: list
                width: parent.width
                height: parent.height - 104
                clip: true
                spacing: 7
                model: menus.filteredEntries
                currentIndex: menus.selectedIndex
                interactive: true
                boundsBehavior: Flickable.StopAtBounds
                onCurrentIndexChanged: positionViewAtIndex(currentIndex, ListView.Contain)

                delegate: Rectangle {
                    width: list.width
                    height: 58
                    radius: 11
                    color: index === menus.selectedIndex
                        ? Qt.rgba(colors.primary.r, colors.primary.g, colors.primary.b, 0.16)
                        : Qt.rgba(colors.bg.r, colors.bg.g, colors.bg.b, 0.62)
                    border.width: 1
                    border.color: modelData.danger
                        ? Qt.rgba(colors.danger.r, colors.danger.g, colors.danger.b, index === menus.selectedIndex ? 0.82 : 0.38)
                        : (index === menus.selectedIndex
                            ? Qt.rgba(colors.primary.r, colors.primary.g, colors.primary.b, 0.66)
                            : Qt.rgba(colors.border.r, colors.border.g, colors.border.b, 0.46))

                    Text {
                        id: badgeText
                        anchors {
                            left: parent.left
                            leftMargin: 12
                            verticalCenter: parent.verticalCenter
                        }
                        width: 82
                        elide: Text.ElideRight
                        text: modelData.badge || menus.mode
                        color: modelData.danger ? colors.danger : colors.secondary
                        font.family: "Maple Mono NF CN"
                        font.pixelSize: 10
                        font.bold: true
                    }

                    Column {
                        anchors {
                            left: parent.left
                            leftMargin: 106
                            right: confirmLabel.left
                            rightMargin: 12
                            verticalCenter: parent.verticalCenter
                        }
                        spacing: 3

                        Text {
                            width: parent.width
                            text: modelData.title || ""
                            color: colors.fg
                            font.family: "Maple Mono NF CN"
                            font.pixelSize: 14
                            font.bold: true
                            elide: Text.ElideRight
                        }

                        Text {
                            width: parent.width
                            text: modelData.subtitle || ""
                            color: colors.muted
                            font.family: "Maple Mono NF CN"
                            font.pixelSize: 11
                            elide: Text.ElideRight
                        }
                    }

                    Text {
                        id: confirmLabel
                        anchors {
                            right: parent.right
                            rightMargin: 14
                            verticalCenter: parent.verticalCenter
                        }
                        width: 130
                        horizontalAlignment: Text.AlignRight
                        text: menus.mode === "power" && index === menus.selectedIndex && menus.confirmPower ? "ENTER CONFIRMS" : ""
                        color: colors.warning
                        font.family: "Maple Mono NF CN"
                        font.pixelSize: 10
                        font.bold: true
                    }

                    MouseArea {
                        anchors.fill: parent
                        hoverEnabled: true
                        cursorShape: Qt.PointingHandCursor
                        onEntered: menus.selectedIndex = index
                        onClicked: menus.activate(index)
                    }
                }
            }
        }
    }
}
