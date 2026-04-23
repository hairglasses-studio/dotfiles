// Quickshell shell migration root.
//
// This stays runnable as a parallel pilot while modules grow into the bar,
// ticker, and notification owners controlled by scripts/shell-stack-mode.sh.

import Quickshell
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
