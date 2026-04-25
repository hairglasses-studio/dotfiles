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
    Services.DockService { id: dockServiceObj }
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
                dockCutover: shellStateObj.dockCutover,
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

    IpcHandler {
        target: "dock"

        function status(): string { return dockServiceObj.status(); }
        function refresh(): string { dockServiceObj.refresh(); return "ok"; }
        function activate(id: string): string { dockServiceObj.activate(id); return "ok"; }
        function launch(id: string): string { dockServiceObj.launch(id); return "ok"; }
        function toggleHidden(): string { dockServiceObj.toggleHidden(); return "ok"; }
    }

    property var primaryScreen: Quickshell.screens.find(s => s.name === shellStateObj.primaryMonitor) || Quickshell.screens[0]

    // Pilot-mode gate: TopBar must NOT render concurrently with ironbar.
    // Pre-cutover (barCutover === false) ironbar owns the DP-3 top + DP-2
    // bottom bar slots; rendering Quickshell TopBar in addition produces
    // visible double-stacked bars (verified via screenshot). Empty model
    // = zero delegates, no surfaces. Flip QS_BAR_CUTOVER=1 via
    // `hg shell bar-cutover` to swap ironbar out and TopBar in.
    //
    // Two Variants blocks at cutover: one for top-anchored bars on every
    // screen, one for bottom-anchored bars on DP-2 (which ironbar pre-
    // cutover ran at the bottom while DP-3's bar was at the top). Both
    // gated by barCutover so pilot mode stays invisible.
    Variants {
        model: shellStateObj.barCutover ? Quickshell.screens : []
        delegate: Component {
            Modules.TopBar {
                primaryScreen: root.primaryScreen
                colors: palette
                shellState: shellStateObj
                barData: barDataObj
                notifications: notificationServiceObj
                anchor: "top"
            }
        }
    }

    Variants {
        // DP-2 bottom anchor — picks just that screen out of the screens
        // list when bar-cutover is active. Empty model in pilot.
        model: shellStateObj.barCutover
            ? Quickshell.screens.filter(s => s.name === "DP-2")
            : []
        delegate: Component {
            Modules.TopBar {
                primaryScreen: root.primaryScreen
                colors: palette
                shellState: shellStateObj
                barData: barDataObj
                notifications: notificationServiceObj
                anchor: "bottom"
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

    Modules.Dock {
        screenModel: root.primaryScreen
        colors: palette
        shellState: shellStateObj
        dock: dockServiceObj
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
