import Quickshell
import Quickshell.Io
import QtQuick

Item {
    id: root

    property bool active: false
    property string mode: "apps"
    property string query: ""
    property var entries: []
    property var filteredEntries: []
    property int selectedIndex: 0
    property bool confirmPower: false
    property string dataScript: Quickshell.shellDir + "/../scripts/quickshell-menu-data.sh"

    readonly property string title: ({
        apps: "Launch",
        windows: "Switch",
        power: "Power",
        emoji: "Emoji",
        agents: "Agents",
        clipboard: "Clipboard"
    })[mode] || "Menu"

    readonly property string hint: ({
        apps: "Type to filter desktop apps",
        windows: "Type title, class, or workspace",
        power: "Select once, press Enter again to confirm",
        emoji: "Search aliases, tags, or unicode names",
        agents: "Focus a running agent terminal",
        clipboard: "Search recent clipboard history"
    })[mode] || "Search"

    function compact(value, max) {
        const clean = String(value || "").replace(/\s+/g, " ").trim();
        return clean.length > max ? clean.slice(0, max - 1) + "..." : clean;
    }

    function score(entry, needle) {
        if (!needle) return 1;
        const haystack = [
            entry.title || "",
            entry.subtitle || "",
            entry.badge || "",
            entry.search || ""
        ].join(" ").toLowerCase();
        const parts = needle.toLowerCase().split(/\s+/).filter(Boolean);
        for (let i = 0; i < parts.length; i++) {
            if (haystack.indexOf(parts[i]) < 0) return 0;
        }
        return 1;
    }

    function applyFilter() {
        const needle = query.trim();
        filteredEntries = entries.filter(e => score(e, needle) > 0).slice(0, 80);
        if (selectedIndex >= filteredEntries.length) selectedIndex = Math.max(0, filteredEntries.length - 1);
        if (selectedIndex < 0) selectedIndex = 0;
        confirmPower = false;
    }

    function open(nextMode) {
        mode = nextMode || "apps";
        query = "";
        selectedIndex = 0;
        confirmPower = false;
        active = true;
        refresh();
    }

    function toggle(nextMode) {
        if (active && mode === nextMode) {
            close();
        } else {
            open(nextMode);
        }
    }

    function close() {
        active = false;
        query = "";
        selectedIndex = 0;
        confirmPower = false;
    }

    function refresh() {
        if (dataProc.running) return;
        dataProc.exec(["bash", dataScript, "list", mode]);
    }

    function move(delta) {
        if (filteredEntries.length === 0) return;
        selectedIndex = (selectedIndex + delta + filteredEntries.length) % filteredEntries.length;
        confirmPower = false;
    }

    function page(delta) {
        move(delta * 8);
    }

    function activateCurrent() {
        activate(selectedIndex);
    }

    function activate(index) {
        if (index < 0 || index >= filteredEntries.length) return;
        const entry = filteredEntries[index];
        if (mode === "power" && !confirmPower) {
            selectedIndex = index;
            confirmPower = true;
            return;
        }
        close();
        Quickshell.execDetached(["bash", dataScript, "exec", mode, String(entry.id || "")]);
    }

    function status() {
        return JSON.stringify({
            visible: active,
            mode: mode,
            query: query,
            count: filteredEntries.length,
            selected: selectedIndex
        });
    }

    onQueryChanged: applyFilter()

    Process {
        id: dataProc
        stdout: SplitParser {
            onRead: data => {
                try {
                    const payload = JSON.parse(data);
                    root.entries = payload.entries || [];
                    root.applyFilter();
                } catch (error) {
                    root.entries = [{
                        id: "error",
                        title: "Menu data error",
                        subtitle: String(error),
                        badge: "ERR",
                        danger: true
                    }];
                    root.applyFilter();
                }
            }
        }
        stderr: SplitParser {
            onRead: data => {
                root.entries = [{
                    id: "stderr",
                    title: "Menu command failed",
                    subtitle: root.compact(data, 180),
                    badge: "ERR",
                    danger: true
                }];
                root.applyFilter();
            }
        }
    }
}
