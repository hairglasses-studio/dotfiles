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
    // Pilot-mode gate: hide entirely until QS_TICKER_CUTOVER=1. The
    // legacy keybind-ticker.service owns the bottom-of-primary-monitor
    // surface; rendering this panel in pilot (top with margin=32)
    // overlapped ironbar's DP-3 top bar, producing a visible
    // double-stack in screenshots.
    visible: shellState.tickerCutover
    anchors { left: true; right: true; top: !shellState.tickerCutover; bottom: shellState.tickerCutover }
    margins { top: shellState.tickerCutover ? 0 : 32 }
    implicitHeight: 30
    exclusiveZone: shellState.tickerCutover ? 30 : 0
    color: "transparent"

    Connections {
        target: ticker
        function onSweepRequested() {
            tickerSweep.restart();
            glitchBurst.restart();
        }
    }

    Rectangle {
        anchors.fill: parent
        color: colors.bg
        opacity: ticker.preset === "minimal" ? 0.88 : 0.95
    }

    // Scanline overlay. Cyberpunk preset gets the animated wave, every other
    // preset gets a static stripe pattern (no per-frame repaint cost). The
    // animated path renders to an FBO so rasterization happens on the GPU,
    // and the timer runs at ~5Hz instead of ~7Hz — the wave amplitude is
    // 1.4px so the human eye doesn't notice the difference, but it cuts the
    // continuous CPU cost by ~30%.
    Canvas {
        anchors.fill: parent
        visible: ticker.preset !== "clean"
        opacity: ticker.urgent ? 0.54 : 0.28
        renderTarget: Canvas.FramebufferObject
        renderStrategy: Canvas.Threaded
        property bool animated: ticker.preset === "cyberpunk" || ticker.urgent
        onAnimatedChanged: requestPaint()
        onPaint: {
            const ctx = getContext("2d");
            ctx.clearRect(0, 0, width, height);
            ctx.strokeStyle = ticker.urgent ? colors.danger : colors.primary;
            ctx.globalAlpha = 0.32;
            const animate = animated;
            const t = animate ? Date.now() / 260 : 0;
            for (let y = 2; y < height; y += 4) {
                const dy = animate ? Math.sin(t + y) * 1.4 : 0;
                ctx.beginPath();
                ctx.moveTo(0, y + dy);
                ctx.lineTo(width, y);
                ctx.stroke();
            }
        }

        Timer {
            interval: 200
            running: panel.visible && parent.animated
            repeat: true
            onTriggered: parent.requestPaint()
        }
    }

    Rectangle {
        id: urgentWash
        anchors.fill: parent
        color: colors.danger
        opacity: ticker.urgent ? 0.16 : 0
        Behavior on opacity { NumberAnimation { duration: 150 } }
    }

    Rectangle {
        anchors { left: parent.left; right: parent.right; top: parent.top }
        height: 1
        color: ticker.urgent ? colors.danger : (ticker.preset === "cyberpunk" ? colors.secondary : colors.primary)
        opacity: 0.82
    }

    Rectangle {
        id: progress
        anchors { left: parent.left; bottom: parent.bottom }
        height: 1
        width: tickerSweep.running ? parent.width * Math.max(0, Math.min(1, 1 - (tickerLabel.x + tickerLabel.implicitWidth) / (tickerClip.width + tickerLabel.implicitWidth))) : 0
        color: ticker.urgent ? colors.danger : colors.tertiary
        opacity: ticker.preset === "minimal" ? 0.35 : 0.7
    }

    Rectangle {
        id: streamBadge
        anchors {
            left: parent.left
            leftMargin: 8
            verticalCenter: parent.verticalCenter
        }
        width: Math.max(98, streamLabel.implicitWidth + 20)
        height: 22
        radius: 5
        color: ticker.urgent ? colors.danger : ticker.accent
        opacity: 0.96

        Text {
            id: streamLabel
            anchors.centerIn: parent
            text: ticker.paused ? ticker.status + " PAUSED" : ticker.status
            color: colors.bg
            font.family: "Maple Mono NF CN"
            font.pixelSize: 10
            font.bold: true
        }

        MouseArea {
            anchors.fill: parent
            acceptedButtons: Qt.LeftButton | Qt.RightButton | Qt.MiddleButton
            cursorShape: Qt.PointingHandCursor
            onClicked: mouse => {
                if (mouse.button === Qt.RightButton) ticker.contextVisible = !ticker.contextVisible;
                else if (mouse.button === Qt.MiddleButton) ticker.togglePause();
                else ticker.next();
            }
            onWheel: wheel => {
                if (wheel.angleDelta.y > 0) ticker.prev();
                else ticker.next();
            }
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
            id: ghostLabel
            x: tickerLabel.x + (ticker.urgent ? -2 : 1)
            y: tickerLabel.y + 1
            text: tickerLabel.text
            visible: ticker.preset === "cyberpunk" || ticker.urgent
            textFormat: Text.PlainText
            color: ticker.urgent ? colors.danger : colors.secondary
            opacity: ticker.urgent ? 0.55 : 0.25
            font.family: tickerLabel.font.family
            font.pixelSize: tickerLabel.font.pixelSize
            font.bold: true
        }

        Text {
            id: tickerLabel
            x: tickerClip.width
            anchors.verticalCenter: parent.verticalCenter
            text: ticker.bannerText.length > 0 ? ticker.bannerText : ticker.text
            textFormat: Text.PlainText
            color: ticker.bannerText.length > 0 ? ticker.bannerColor : colors.fg
            font.family: "Maple Mono NF CN"
            font.pixelSize: ticker.urgent ? 14 : 13
            font.bold: true
        }

        Rectangle {
            id: glitchSlice
            x: Math.max(0, tickerLabel.x + 40)
            y: 5
            width: ticker.urgent ? 130 : 90
            height: 3
            color: ticker.urgent ? colors.danger : colors.secondary
            opacity: 0
        }

        NumberAnimation {
            id: glitchBurst
            target: glitchSlice
            property: "opacity"
            from: ticker.preset === "clean" ? 0 : 0.65
            to: 0
            duration: ticker.urgent ? 520 : 280
        }

        NumberAnimation {
            id: tickerSweep
            target: tickerLabel
            property: "x"
            from: tickerClip.width
            to: -tickerLabel.implicitWidth
            duration: {
                const base = Math.max(18000, tickerLabel.implicitWidth * 34);
                if (ticker.paused) return 999999999;
                if (ticker.hovered) return base * 5;
                if (ticker.urgent) return Math.max(9000, base * 0.52);
                return base;
            }
            loops: Animation.Infinite
            running: !ticker.paused
            easing.type: Easing.Linear
        }

        MouseArea {
            anchors.fill: parent
            hoverEnabled: true
            acceptedButtons: Qt.LeftButton | Qt.RightButton | Qt.MiddleButton
            onEntered: ticker.hovered = true
            onExited: ticker.hovered = false
            onClicked: mouse => {
                if (mouse.button === Qt.RightButton) ticker.contextVisible = !ticker.contextVisible;
                else if (mouse.button === Qt.MiddleButton) ticker.togglePause();
                else ticker.copyCurrent();
            }
            onWheel: wheel => {
                if (wheel.modifiers & Qt.ShiftModifier) {
                    if (wheel.angleDelta.y > 0) ticker.prev();
                    else ticker.next();
                } else {
                    ticker.hovered = wheel.angleDelta.y > 0;
                }
            }
        }

        Rectangle {
            anchors { left: parent.left; top: parent.top; bottom: parent.bottom }
            width: 36
            gradient: Gradient {
                orientation: Gradient.Horizontal
                GradientStop { position: 0.0; color: colors.bg }
                GradientStop { position: 1.0; color: Qt.rgba(colors.bg.r, colors.bg.g, colors.bg.b, 0) }
            }
        }

        Rectangle {
            anchors { right: parent.right; top: parent.top; bottom: parent.bottom }
            width: 110
            gradient: Gradient {
                orientation: Gradient.Horizontal
                GradientStop { position: 0.0; color: Qt.rgba(colors.bg.r, colors.bg.g, colors.bg.b, 0) }
                GradientStop { position: 1.0; color: colors.bg }
            }
        }
    }

    Text {
        visible: ticker.bannerText.length === 0 && notifications.latestText.length > 0
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
