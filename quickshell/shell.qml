// Quickshell prototype — GPU-native bar on DP-3 portrait.
//
// Phase 0 of the "evaluate quickshell as a future bar surface" item
// in ROADMAP.md's Future Considerations. Not enabled as a service yet;
// launch manually with `quickshell --config ~/.config/quickshell/shell.qml`
// or via the `scripts/quickshell-try.sh` helper.
//
// Goals of the prototype:
//   1. Prove layer-shell anchoring + PanelWindow works on a portrait
//      monitor alongside the existing Hyprland layer stack (ironbar
//      top, ticker bottom on DP-2; nothing on DP-3 yet besides the
//      secondary ticker instance).
//   2. Exercise the GPU shader pipeline — a ShaderEffect node with a
//      minimal GLSL fragment so we can measure "is this actually
//      faster than the Python/cairo ticker on the same content?"
//   3. Share the Hairglasses Neon palette (cyan / magenta / green /
//      amber) so it doesn't look foreign next to ironbar + ticker.
//
// Non-goals: replacing ironbar. Quickshell stays opt-in on DP-3 until
// the prototype shows clear wins (or doesn't, in which case the code
// is removed and the decision documented in ROADMAP).

import Quickshell
import Quickshell.Io
import Quickshell.Wayland
import QtQuick
import QtQuick.Layouts

PanelWindow {
    id: panel
    screen: Quickshell.screens.find(s => s.name === "DP-3")

    anchors {
        top:   true
        left:  true
        right: true
    }
    // Portrait monitor, 28 px thin — matches the ticker's pre-bump size
    // so Quickshell looks consistent next to the existing stack.
    implicitHeight: 28
    color: "transparent"

    // ── Background: transparent dark panel + subtle shader ──────────
    Rectangle {
        anchors.fill: parent
        color: "#05070d"
        opacity: 0.88
    }

    // GPU shader: a minimal horizontal gradient sweep in GLSL so we
    // can benchmark the GPU path vs Cairo. The sweep phase comes from
    // an animated number property driven by a QML Timer — same idea
    // as the ticker's `self.gradient_phase` accumulator but all the
    // per-pixel work happens on the GPU.
    ShaderEffect {
        anchors.fill: parent
        property real phase: 0.0
        NumberAnimation on phase {
            from: 0.0
            to:   1.0
            duration: 8000
            loops: Animation.Infinite
        }
        fragmentShader: "
            #version 440
            layout(location = 0) in vec2 qt_TexCoord0;
            layout(location = 0) out vec4 fragColor;
            layout(std140, binding = 0) uniform buf {
                mat4 qt_Matrix;
                float qt_Opacity;
                float phase;
            };
            void main() {
                float t = fract(qt_TexCoord0.x + phase);
                // Hairglasses Neon cycle — cyan → magenta → green → amber.
                vec3 c0 = vec3(0.161, 0.941, 1.000);  // #29f0ff
                vec3 c1 = vec3(1.000, 0.278, 0.820);  // #ff47d1
                vec3 c2 = vec3(0.239, 1.000, 0.710);  // #3dffb5
                vec3 c3 = vec3(1.000, 0.894, 0.369);  // #ffe45e
                vec3 col;
                if (t < 0.333)       col = mix(c0, c1, t / 0.333);
                else if (t < 0.666)  col = mix(c1, c2, (t - 0.333) / 0.333);
                else                 col = mix(c2, c3, (t - 0.666) / 0.334);
                fragColor = vec4(col, qt_Opacity * 0.18);
            }
        "
    }

    // ── Content: clock ──────────────────────────────────────────────
    Text {
        id: clock
        anchors.centerIn: parent
        color: "#f7fbff"
        text: Qt.formatDateTime(new Date(), "ddd  HH:mm:ss")
        font.family: "Maple Mono NF CN"
        font.pixelSize: 14
        font.bold: true

        Timer {
            interval: 1000
            running: true
            repeat: true
            onTriggered: clock.text = Qt.formatDateTime(new Date(), "ddd  HH:mm:ss")
        }
    }

    // ── Left indicator — pulse dot marking "Quickshell active" ──────
    // Helps differentiate this prototype from the ironbar surface on
    // DP-3 at a glance.
    Rectangle {
        id: pulse
        width: 8; height: 8; radius: 4
        anchors { verticalCenter: parent.verticalCenter; left: parent.left; leftMargin: 10 }
        color: "#29f0ff"
        SequentialAnimation on opacity {
            loops: Animation.Infinite
            NumberAnimation { from: 1.0; to: 0.35; duration: 1200; easing.type: Easing.InOutSine }
            NumberAnimation { from: 0.35; to: 1.0; duration: 1200; easing.type: Easing.InOutSine }
        }
    }
}
