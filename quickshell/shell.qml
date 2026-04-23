// Quickshell shell migration root.
//
// This stays runnable as a parallel pilot while modules grow into the bar,
// ticker, menu, notification, and companion owners controlled by
// scripts/shell-stack-mode.sh.

import Quickshell
import Quickshell.Io
import QtQuick
import "modules" as Modules
import "services" as Services
import "styles" as Theme

ShellRoot {
    id: root

    Theme.Colors { id: palette }
    Services.ShellState { id: shellStateObj }
    Services.BarData { id: barDataObj }
    Services.TickerService { id: tickerServiceObj }
    Services.MenuService { id: menuServiceObj }
    Services.WindowFocusService { id: windowFocusObj }
    Services.FleetTelemetryService { id: fleetTelemetryObj }
    Services.LyricsService { id: lyricsServiceObj }
    Services.NotificationService {
        id: notificationServiceObj
        ownerEnabled: shellStateObj.notificationOwner
    }

    IpcHandler {
        target: "shell"

        function toggleNotifications(): string { notificationServiceObj.centerVisible = !notificationServiceObj.centerVisible; return "ok"; }
        function showNotifications(): string { notificationServiceObj.centerVisible = true; return "ok"; }
        function hideNotifications(): string { notificationServiceObj.centerVisible = false; return "ok"; }
        function toggleQuickSettings(): string { shellStateObj.quickSettingsVisible = !shellStateObj.quickSettingsVisible; return "ok"; }
        function setDnd(enabled: bool): string { notificationServiceObj.setDnd(enabled); return "ok"; }
        function toggleDnd(): string { notificationServiceObj.toggleDnd(); return "ok"; }
        function closeNotifications(): string { notificationServiceObj.closeAll(); return "ok"; }
        function clearNotifications(): string { notificationServiceObj.clearHistory(); return "ok"; }
        function status(): string {
            return JSON.stringify({
                mode: shellStateObj.mode,
                notificationOwner: shellStateObj.notificationOwner,
                barCutover: shellStateObj.barCutover,
                tickerCutover: shellStateObj.tickerCutover,
                menuCutover: shellStateObj.menuCutover,
                companionCutover: shellStateObj.companionCutover,
                notificationsVisible: notificationServiceObj.centerVisible,
                quickSettingsVisible: shellStateObj.quickSettingsVisible,
                dnd: notificationServiceObj.dnd,
                notificationCount: notificationServiceObj.notificationCount,
                criticalCount: notificationServiceObj.criticalCount
            });
        }
    }

    IpcHandler {
        target: "menus"

        function open(mode: string): string { menuServiceObj.open(mode); return "ok"; }
        function toggle(mode: string): string { menuServiceObj.toggle(mode); return "ok"; }
        function close(): string { menuServiceObj.close(); return "ok"; }
        function status(): string { return menuServiceObj.status(); }
    }

    IpcHandler {
        target: "ticker"

        function status(): string { return tickerServiceObj.statusJson(); }
        function listStreams(): string { return tickerServiceObj.listStreams(); }
        function next(): string { tickerServiceObj.next(); return "ok"; }
        function prev(): string { tickerServiceObj.prev(); return "ok"; }
        function pin(stream: string): string { tickerServiceObj.pin(stream); return "ok"; }
        function unpin(): string { tickerServiceObj.unpin(); return "ok"; }
        function pinToggle(): string { tickerServiceObj.pinToggle(); return "ok"; }
        function pauseToggle(): string { tickerServiceObj.togglePause(); return "ok"; }
        function shuffle(mode: string): string { tickerServiceObj.setShuffleMode(mode); return "ok"; }
        function playlist(name: string): string { tickerServiceObj.setPlaylist(name); return "ok"; }
        function preset(name: string): string { tickerServiceObj.setPreset(name); return "ok"; }
        function reload(): string { tickerServiceObj.reload(); return "ok"; }
        function banner(text: string, color: string): string { tickerServiceObj.showBanner(text, color); return "ok"; }
        function urgent(enabled: bool): string { tickerServiceObj.setUrgent(enabled); return "ok"; }
        function snoozeUrgent(): string { tickerServiceObj.snoozeUrgent(); return "ok"; }
    }

    property var primaryScreen: Quickshell.screens.find(s => s.name === shellStateObj.primaryMonitor) || Quickshell.screens[0]

    Variants {
        model: Quickshell.screens
        delegate: Component {
            Modules.TopBar {
                primaryScreen: root.primaryScreen
                colors: palette
                shellState: shellStateObj
                barData: barDataObj
                notifications: notificationServiceObj
            }
        }
    }

    Modules.TickerBar {
        screenModel: root.primaryScreen
        colors: palette
        shellState: shellStateObj
        ticker: tickerServiceObj
        notifications: notificationServiceObj
    }

    Modules.TickerContextMenu {
        screenModel: root.primaryScreen
        colors: palette
        shellState: shellStateObj
        ticker: tickerServiceObj
    }

    Modules.MenuOverlay {
        screenModel: root.primaryScreen
        colors: palette
        menus: menuServiceObj
    }

    Modules.WindowLabel {
        screenModel: root.primaryScreen
        colors: palette
        shellState: shellStateObj
        windowFocus: windowFocusObj
    }

    Modules.LyricsBanner {
        screenModel: root.primaryScreen
        colors: palette
        shellState: shellStateObj
        lyrics: lyricsServiceObj
        windowFocus: windowFocusObj
    }

    Modules.FleetSparklines {
        screenModel: root.primaryScreen
        colors: palette
        shellState: shellStateObj
        telemetry: fleetTelemetryObj
    }

    Modules.NotificationCenter {
        screenModel: root.primaryScreen
        colors: palette
        notifications: notificationServiceObj
    }

    Modules.NotificationPopups {
        screenModel: root.primaryScreen
        colors: palette
        notifications: notificationServiceObj
    }

    Modules.QuickSettings {
        screenModel: root.primaryScreen
        colors: palette
        shellState: shellStateObj
        barData: barDataObj
        notifications: notificationServiceObj
    }
}
