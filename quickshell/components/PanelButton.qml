import QtQuick

Rectangle {
    id: button

    property var colors
    property alias text: label.text
    property color accent: colors ? colors.primary : "#29f0ff"
    signal clicked()

    height: 30
    radius: 8
    color: colors ? Qt.rgba(colors.bg.r, colors.bg.g, colors.bg.b, 0.72) : "#05070d"
    border.width: 1
    border.color: Qt.rgba(accent.r, accent.g, accent.b, 0.38)
    implicitWidth: label.implicitWidth + 24

    Text {
        id: label
        anchors.centerIn: parent
        color: button.accent
        font.family: "Maple Mono NF CN"
        font.pixelSize: 11
        font.bold: true
    }

    MouseArea {
        anchors.fill: parent
        cursorShape: Qt.PointingHandCursor
        onClicked: button.clicked()
    }
}
