import Quickshell
import Quickshell.Wayland
import QtQuick

PanelWindow {
    id: panel

    property var screenModel
    property var colors
    property var shellState
    property var ticker
    property var notifications

    screen: screenModel
    anchors { left: true; right: true; top: !shellState.tickerCutover; bottom: shellState.tickerCutover }
    margins { top: shellState.tickerCutover ? 0 : 32 }
    implicitHeight: 30
    exclusiveZone: shellState.tickerCutover ? 30 : 0
    color: "transparent"

    Connections {
        target: ticker
        function onSweepRequested() {
            tickerSweep.restart();
        }
    }

    Rectangle {
        anchors.fill: parent
        color: colors.bg
        opacity: 0.93
    }

    Rectangle {
        anchors { left: parent.left; right: parent.right; top: parent.top }
        height: 1
        color: ticker.preset === "cyberpunk" ? colors.secondary : colors.primary
        opacity: 0.74
    }

    Rectangle {
        id: streamBadge
        anchors {
            left: parent.left
            leftMargin: 8
            verticalCenter: parent.verticalCenter
        }
        width: Math.max(92, streamLabel.implicitWidth + 18)
        height: 22
        radius: 4
        color: ticker.preset === "cyberpunk" ? colors.secondary : colors.primary
        opacity: 0.95

        Text {
            id: streamLabel
            anchors.centerIn: parent
            text: ticker.status
            color: colors.bg
            font.family: "Maple Mono NF CN"
            font.pixelSize: 11
            font.bold: true
        }

        MouseArea {
            anchors.fill: parent
            cursorShape: Qt.PointingHandCursor
            onClicked: ticker.advance()
        }
    }

    Item {
        id: tickerClip
        anchors {
            left: streamBadge.right
            leftMargin: 12
            right: parent.right
            rightMargin: 8
            top: parent.top
            bottom: parent.bottom
        }
        clip: true

        Text {
            id: tickerLabel
            x: tickerClip.width
            anchors.verticalCenter: parent.verticalCenter
            text: ticker.text
            textFormat: Text.PlainText
            color: colors.fg
            font.family: "Maple Mono NF CN"
            font.pixelSize: 13
            font.bold: true
        }

        NumberAnimation {
            id: tickerSweep
            target: tickerLabel
            property: "x"
            from: tickerClip.width
            to: -tickerLabel.implicitWidth
            duration: Math.max(18000, tickerLabel.implicitWidth * 34)
            loops: Animation.Infinite
            running: true
            easing.type: Easing.Linear
        }
    }

    Text {
        visible: notifications.latestText.length > 0
        anchors {
            right: parent.right
            rightMargin: 14
            verticalCenter: parent.verticalCenter
        }
        width: 340
        horizontalAlignment: Text.AlignRight
        elide: Text.ElideRight
        text: notifications.latestText
        color: notifications.criticalCount > 0 ? colors.danger : colors.secondary
        font.family: "Maple Mono NF CN"
        font.pixelSize: 10
        font.bold: true
        opacity: 0.85
    }
}
