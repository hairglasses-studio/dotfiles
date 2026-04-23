// Quickshell shell migration root.
//
// This stays runnable as a parallel pilot while modules grow into the bar,
// ticker, and notification owners controlled by scripts/shell-stack-mode.sh.

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
                notificationsVisible: notificationServiceObj.centerVisible,
                quickSettingsVisible: shellStateObj.quickSettingsVisible,
                dnd: notificationServiceObj.dnd,
                notificationCount: notificationServiceObj.notificationCount,
                criticalCount: notificationServiceObj.criticalCount
            });
        }
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
