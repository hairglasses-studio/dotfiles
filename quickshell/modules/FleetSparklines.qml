import Quickshell
import Quickshell.Wayland
import QtQuick

PanelWindow {
    id: panel

    property var screenModel
    property var colors
    property var shellState
    property var telemetry

    screen: screenModel
    anchors { top: true; right: true }
    margins { top: shellState && shellState.barCutover ? 68 : 38; right: 10 }
    implicitWidth: 214
    implicitHeight: 88
    exclusiveZone: 0
    color: "transparent"
    visible: shellState && shellState.companionCutover

    Connections {
        target: telemetry
        function onSampleVersionChanged() {
            chart.requestPaint();
        }
    }

    Rectangle {
        anchors.fill: parent
        radius: 9
        color: colors.panelStrong
        border.color: colors.borderStrong
        border.width: 1
        opacity: 0.92
    }

    Canvas {
        id: chart
        anchors.fill: parent
        anchors.margins: 4

        function maxOf(history, hint) {
            let value = hint || 1;
            for (let i = 0; i < history.length; i++) value = Math.max(value, Number(history[i]) || 0);
            return Math.max(value, 1);
        }

        function drawSpark(ctx, history, x, y, w, h, accent, hint) {
            if (!history || history.length < 2) return;
            const maxValue = maxOf(history, hint);
            const step = w / Math.max(1, history.length - 1);
            ctx.beginPath();
            for (let i = 0; i < history.length; i++) {
                const px = x + i * step;
                const py = y + h - Math.min(Number(history[i]) || 0, maxValue) / maxValue * h;
                if (i === 0) ctx.moveTo(px, py);
                else ctx.lineTo(px, py);
            }
            ctx.strokeStyle = accent;
            ctx.lineWidth = 1.4;
            ctx.stroke();
        }

        function drawTile(ctx, x, y, w, h, title, value, history, accent, hint) {
            ctx.fillStyle = colors.surfaceAlt;
            ctx.globalAlpha = 0.94;
            ctx.fillRect(x + 1, y + 1, w - 2, h - 2);
            ctx.globalAlpha = 0.42;
            ctx.strokeStyle = accent;
            ctx.lineWidth = 1;
            ctx.strokeRect(x + 1.5, y + 1.5, w - 3, h - 3);
            ctx.globalAlpha = 1;
            ctx.font = "bold 9px 'Maple Mono NF CN'";
            ctx.fillStyle = colors.muted;
            ctx.fillText(title, x + 6, y + 13);
            ctx.fillStyle = accent;
            ctx.fillText(value, x + 36, y + 13);
            drawSpark(ctx, history || [], x + 6, y + 19, w - 12, h - 25, accent, hint);
        }

        onPaint: {
            const ctx = getContext("2d");
            ctx.clearRect(0, 0, width, height);
            const hw = width / 2;
            const hh = height / 2;
            drawTile(ctx, 0, 0, hw, hh, "CPU", telemetry ? telemetry.cpuText.replace("CPU ", "") : "--", telemetry ? telemetry.cpuHistory : [], colors.primary, 100);
            drawTile(ctx, hw, 0, hw, hh, "MEM", telemetry ? telemetry.memText.replace("MEM ", "") : "--", telemetry ? telemetry.memHistory : [], colors.secondary, 100);
            drawTile(ctx, 0, hh, hw, hh, "NET", telemetry ? telemetry.netText.replace("NET ", "") : "--", telemetry ? telemetry.netHistory : [], colors.tertiary, 64);
            drawTile(ctx, hw, hh, hw, hh, "GPU", telemetry ? telemetry.gpuText.replace("GPU ", "") : "--", telemetry ? telemetry.gpuHistory : [], colors.warning, 100);
        }

        Component.onCompleted: requestPaint()
        onWidthChanged: requestPaint()
        onHeightChanged: requestPaint()
    }
}
