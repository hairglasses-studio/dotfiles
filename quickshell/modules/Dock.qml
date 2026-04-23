import Quickshell
import Quickshell.Wayland
import QtQuick

PanelWindow {
    id: panel

    property var screenModel
    property var colors
    property var shellState
    property var dock
    property int hoveredIndex: -1

    visible: shellState && shellState.dockCutover && dock && !dock.hidden
    screen: screenModel
    anchors { left: true; right: true; bottom: true }
    margins { bottom: shellState && shellState.tickerCutover ? 38 : 8 }
    implicitHeight: 62
    exclusiveZone: 0
    color: "transparent"
    WlrLayershell.layer: WlrLayer.Top

    Rectangle {
        id: capsule
        anchors.centerIn: parent
        width: Math.min(parent.width - 36, Math.max(340, dockRow.implicitWidth + 28))
        height: 54
        radius: 18
        color: Qt.rgba(colors.panelStrong.r, colors.panelStrong.g, colors.panelStrong.b, 0.9)
        border.width: 1
        border.color: Qt.rgba(colors.primary.r, colors.primary.g, colors.primary.b, 0.42)

        Rectangle {
            anchors.fill: parent
            anchors.margins: 1
            radius: 17
            color: "transparent"
            border.width: 1
            border.color: Qt.rgba(colors.secondary.r, colors.secondary.g, colors.secondary.b, 0.12)
        }

        Rectangle {
            id: scan
            y: 1
            height: 1
            width: parent.width * 0.38
            opacity: 0.65
            gradient: Gradient {
                orientation: Gradient.Horizontal
                GradientStop { position: 0.0; color: "transparent" }
                GradientStop { position: 0.45; color: colors.primary }
                GradientStop { position: 1.0; color: "transparent" }
            }
            NumberAnimation on x {
                from: -scan.width
                to: capsule.width
                duration: 5200
                loops: Animation.Infinite
                easing.type: Easing.InOutSine
            }
        }

        Row {
            id: dockRow
            anchors.centerIn: parent
            spacing: 7

            Repeater {
                model: dock ? dock.entries : []

                delegate: Rectangle {
                    id: item
                    width: 42
                    height: 42
                    radius: 13
                    color: modelData.active
                        ? Qt.rgba(colors.primary.r, colors.primary.g, colors.primary.b, 0.22)
                        : (hovered ? Qt.rgba(colors.secondary.r, colors.secondary.g, colors.secondary.b, 0.16)
                            : Qt.rgba(colors.bg.r, colors.bg.g, colors.bg.b, 0.68))
                    border.width: 1
                    border.color: modelData.active
                        ? Qt.rgba(colors.primary.r, colors.primary.g, colors.primary.b, 0.78)
                        : (modelData.running > 0
                            ? Qt.rgba(colors.tertiary.r, colors.tertiary.g, colors.tertiary.b, 0.54)
                            : Qt.rgba(colors.borderStrong.r, colors.borderStrong.g, colors.borderStrong.b, 0.42))

                    property bool hovered: panel.hoveredIndex === index

                    Behavior on color { ColorAnimation { duration: 110 } }
                    Behavior on border.color { ColorAnimation { duration: 110 } }

                    Text {
                        anchors.centerIn: parent
                        text: modelData.badge || "APP"
                        color: modelData.active ? colors.primary : (modelData.running > 0 ? colors.tertiary : colors.fg)
                        font.family: "Maple Mono NF CN"
                        font.pixelSize: (modelData.badge || "").length > 3 ? 9 : 11
                        font.bold: true
                    }

                    Rectangle {
                        anchors {
                            horizontalCenter: parent.horizontalCenter
                            bottom: parent.bottom
                            bottomMargin: 4
                        }
                        width: modelData.active ? 18 : (modelData.running > 0 ? 8 : 0)
                        height: 2
                        radius: 1
                        color: modelData.active ? colors.primary : colors.tertiary
                        opacity: modelData.running > 0 || modelData.active ? 0.95 : 0
                    }

                    MouseArea {
                        anchors.fill: parent
                        hoverEnabled: true
                        acceptedButtons: Qt.LeftButton | Qt.RightButton | Qt.MiddleButton
                        cursorShape: Qt.PointingHandCursor
                        onEntered: panel.hoveredIndex = index
                        onExited: if (panel.hoveredIndex === index) panel.hoveredIndex = -1
                        onClicked: mouse => {
                            if (mouse.button === Qt.RightButton || mouse.button === Qt.MiddleButton) dock.launch(modelData.id);
                            else dock.activate(modelData.id);
                        }
                    }
                }
            }
        }
    }

    Rectangle {
        visible: panel.hoveredIndex >= 0 && dock && panel.hoveredIndex < dock.entries.length
        anchors {
            horizontalCenter: capsule.horizontalCenter
            bottom: capsule.top
            bottomMargin: 7
        }
        width: Math.min(520, tooltipText.implicitWidth + 28)
        height: 28
        radius: 9
        color: Qt.rgba(colors.surface.r, colors.surface.g, colors.surface.b, 0.94)
        border.width: 1
        border.color: Qt.rgba(colors.secondary.r, colors.secondary.g, colors.secondary.b, 0.5)

        Text {
            id: tooltipText
            anchors.centerIn: parent
            width: Math.min(480, implicitWidth)
            elide: Text.ElideRight
            text: {
                const entry = dock.entries[panel.hoveredIndex] || {};
                const title = entry.title || "";
                const subtitle = entry.subtitle || "";
                return subtitle.length > 0 ? title + " - " + subtitle : title;
            }
            color: colors.fg
            font.family: "Maple Mono NF CN"
            font.pixelSize: 11
            font.bold: true
        }
    }
}
