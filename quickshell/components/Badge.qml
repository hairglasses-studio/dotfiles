import QtQuick

Rectangle {
    id: badge

    property var colors
    property alias text: label.text
    property color accent: colors ? colors.primary : "#29f0ff"
    property color fill: colors ? Qt.rgba(colors.surface.r, colors.surface.g, colors.surface.b, 0.74) : "#0f1219"
    property int fontSize: 11
    property bool bold: true

    signal clicked()

    height: 22
    radius: 5
    border.width: 1
    border.color: Qt.rgba(accent.r, accent.g, accent.b, 0.42)
    color: fill
    implicitWidth: label.implicitWidth + 18

    Text {
        id: label
        anchors.centerIn: parent
        color: badge.accent
        font.family: "Maple Mono NF CN"
        font.pixelSize: badge.fontSize
        font.bold: badge.bold
        elide: Text.ElideRight
    }

    MouseArea {
        anchors.fill: parent
        cursorShape: Qt.PointingHandCursor
        onClicked: badge.clicked()
    }
}
