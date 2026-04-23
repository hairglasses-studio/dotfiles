import Quickshell.Services.Mpris
import QtQuick
import "." as Components

Row {
    id: root

    property var colors
    property string fallbackText: ""

    spacing: 6

    Repeater {
        model: Mpris.players

        delegate: Components.Badge {
            colors: root.colors
            visible: index < 1 && ((modelData.trackTitle || "").length > 0)
            text: ((modelData.isPlaying ? "PLAY " : "PAUSE ") + (modelData.trackTitle || "") + (modelData.trackArtist ? " -- " + modelData.trackArtist : "")).replace(/\s+/g, " ").slice(0, 42)
            accent: modelData.isPlaying ? root.colors.secondary : root.colors.muted
            onClicked: {
                if (modelData.canTogglePlaying) modelData.togglePlaying();
            }
        }
    }

    Components.Badge {
        colors: root.colors
        visible: fallbackText.length > 0 && Mpris.players.values.length === 0
        text: fallbackText
        accent: root.colors.secondary
    }
}
