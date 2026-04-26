import Quickshell
import Quickshell.Io
import Quickshell.Services.Notifications
import QtQuick

Item {
    id: root

    property bool ownerEnabled: false
    property string bridgePath: Quickshell.shellDir + "/../scripts/notification-bridge.py"
    property int notificationCount: 0
    property int criticalCount: 0
    property bool dnd: false
    property string latestText: ""
    property var entries: []
    property var popupEntries: []
    property bool centerVisible: false
    // Map: entry.id (string) -> live Notification object. Only populated for
    // notifications we own (ownerEnabled=true) so action buttons can call
    // invoke() on the original NotificationAction. Persisted entries from
    // history.jsonl have no live ref and render without action buttons.
    property var liveNotifications: ({})

    function compact(value, max) {
        const clean = (value || "").replace(/\s+/g, " ").trim();
        return clean.length > max ? clean.slice(0, max - 1) + "..." : clean;
    }

    function urgencyName(value) {
        if (value === NotificationUrgency.Critical) return "critical";
        if (value === NotificationUrgency.Low) return "low";
        return "normal";
    }

    function applyPayload(data) {
        try {
            const payload = JSON.parse(data);
            if (!payload.ok) {
                root.latestText = "notification bridge failed";
                return;
            }
            root.notificationCount = payload.daemon_count === null || payload.daemon_count === undefined ? (payload.visible || 0) : payload.daemon_count;
            root.criticalCount = payload.critical || 0;
            root.dnd = payload.dnd === true;
            root.entries = payload.entries || [];
            const latest = payload.latest || {};
            const app = latest.app ? latest.app + ": " : "";
            root.latestText = root.compact(app + (latest.text || latest.summary || ""), 54);
        } catch (error) {
            root.latestText = root.compact(String(error), 54);
        }
    }

    function refresh() {
        if (!refreshProc.running) refreshProc.exec(refreshProc.command);
    }

    function runAction(args) {
        if (!actionProc.running) actionProc.exec(["python3", root.bridgePath].concat(args));
    }

    function setDnd(enabled) {
        root.dnd = enabled;
        runAction(["--dnd", root.dnd ? "true" : "false"]);
    }

    function toggleDnd() {
        setDnd(!root.dnd);
    }

    function closeAll() {
        root.popupEntries = [];
        runAction(["--close-all"]);
    }

    function clearHistory() {
        root.entries = [];
        root.popupEntries = [];
        runAction(["--clear-history"]);
    }

    function appendEntry(entry) {
        if (!appendProc.running) appendProc.exec(["python3", root.bridgePath, "--append-entry-json", JSON.stringify(entry)]);
    }

    function acceptOwned(notification) {
        const entryId = String(notification.id);
        const actions = [];
        const rawActions = notification.actions || [];
        for (let i = 0; i < rawActions.length; i++) {
            const a = rawActions[i];
            if (!a) continue;
            const id = a.identifier || a.id || "";
            const text = a.text || a.label || id;
            if (id === "default" && !text) continue;
            actions.push({ id: id, text: text });
        }
        const entry = {
            id: entryId,
            app: notification.appName || "",
            summary: notification.summary || "",
            body: notification.body || "",
            urgency: urgencyName(notification.urgency),
            visible: true,
            dismissed: false,
            ts: Math.floor(Date.now() / 1000),
            actions: actions
        };
        entry.text = entry.body && entry.body !== entry.summary ? entry.summary + ": " + entry.body : (entry.summary || entry.body);
        notification.tracked = true;
        // Hold a live ref so action buttons can invoke later. Drop on close.
        // Cap at 60 entries — only entries inside the visible history (40)
        // ever show buttons, and we keep a small extra margin for in-flight
        // popups. An app spamming notifications without close() events
        // would otherwise leak refs indefinitely; evict oldest when we hit
        // the cap.
        const liveCopy = root.liveNotifications;
        liveCopy[entryId] = notification;
        const liveIds = Object.keys(liveCopy);
        if (liveIds.length > 60) {
            const overflow = liveIds.length - 60;
            for (let i = 0; i < overflow; i++) {
                delete liveCopy[liveIds[i]];
            }
        }
        root.liveNotifications = liveCopy;
        notification.closed.connect(() => root.dropLive(entryId));
        appendEntry(entry);
        root.entries = [entry].concat(root.entries).slice(0, 40);
        root.notificationCount = root.entries.length;
        root.criticalCount = root.entries.filter(e => e.urgency === "critical").length;
        root.latestText = root.compact((entry.app ? entry.app + ": " : "") + entry.text, 54);
        if (root.dnd) {
            notification.dismiss();
            return;
        }
        root.popupEntries = [entry].concat(root.popupEntries).slice(0, 3);
    }

    function dropLive(entryId) {
        if (!root.liveNotifications.hasOwnProperty(entryId)) return;
        const liveCopy = root.liveNotifications;
        delete liveCopy[entryId];
        root.liveNotifications = liveCopy;
    }

    function invokeAction(entryId, actionId) {
        const live = root.liveNotifications[entryId];
        if (!live || !live.actions) return false;
        for (let i = 0; i < live.actions.length; i++) {
            const a = live.actions[i];
            if (!a) continue;
            const id = a.identifier || a.id || "";
            if (id === actionId) {
                a.invoke();
                root.dropLive(entryId);
                // Remove from popups (action click implies dismissal).
                root.popupEntries = root.popupEntries.filter(e => e.id !== entryId);
                return true;
            }
        }
        return false;
    }

    Process {
        id: refreshProc
        command: ["python3", root.bridgePath, "--limit", "18"]
        stdout: SplitParser { onRead: data => root.applyPayload(data) }
    }

    Process {
        id: actionProc
        stdout: SplitParser { onRead: _data => root.refresh() }
    }

    Process {
        id: appendProc
    }

    Timer {
        interval: 15000
        running: true
        repeat: true
        onTriggered: root.refresh()
    }

    Loader {
        active: root.ownerEnabled
        sourceComponent: Component {
            NotificationServer {
                id: server
                keepOnReload: false
                persistenceSupported: true
                bodySupported: true
                bodyMarkupSupported: false
                actionsSupported: true
                actionIconsSupported: true
                imageSupported: true

                onNotification: notification => root.acceptOwned(notification)
            }
        }
    }

    Component.onCompleted: refresh()
}
